package actrs

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/anthdm/hollywood/actor"
	"github.com/anthdm/hollywood/cluster"
	"github.com/anthdm/run/pkg/storage"
	"github.com/anthdm/run/proto"
	"github.com/google/uuid"
)

const KindWasmServer = "wasm_server"

type requestWithResponse struct {
	request  *proto.HTTPRequest
	response chan *proto.HTTPResponse
}

func newRequestWithResponse(request *proto.HTTPRequest) requestWithResponse {
	return requestWithResponse{
		request:  request,
		response: make(chan *proto.HTTPResponse, 1),
	}
}

// WasmServer is an HTTP server that will proxy and route the request to the corresponding function.
type WasmServer struct {
	server      *http.Server
	self        *actor.PID
	store       storage.Store
	metricStore storage.MetricStore
	cache       storage.ModCacher
	cluster     *cluster.Cluster
	responses   map[string]chan *proto.HTTPResponse
}

// NewWasmServer return a new wasm server given a storage and a mod cache.
func NewWasmServer(addr string, cluster *cluster.Cluster, store storage.Store, metricStore storage.MetricStore, cache storage.ModCacher) actor.Producer {
	return func() actor.Receiver {
		s := &WasmServer{
			store:       store,
			metricStore: metricStore,
			cache:       cache,
			cluster:     cluster,
			responses:   make(map[string]chan *proto.HTTPResponse),
		}
		server := &http.Server{
			Handler: s,
			Addr:    addr,
		}
		s.server = server
		return s
	}
}

func (s *WasmServer) Receive(c *actor.Context) {
	switch msg := c.Message().(type) {
	case actor.Started:
		s.initialize(c)
	case actor.Stopped:
	case requestWithResponse:
		s.responses[msg.request.ID] = msg.response
		s.sendRequestToRuntime(msg.request)
	case *proto.HTTPResponse:
		if resp, ok := s.responses[msg.RequestID]; ok {
			resp <- msg
			delete(s.responses, msg.RequestID)
		}
	}
}

func (s *WasmServer) initialize(c *actor.Context) {
	s.self = c.PID()
	go func() {
		log.Fatal(s.server.ListenAndServe())
	}()
}

func (s *WasmServer) sendRequestToRuntime(req *proto.HTTPRequest) {
	pid := s.cluster.Activate(KindRuntime, &cluster.ActivationConfig{})
	s.cluster.Engine().SendWithSender(pid, req, s.self)
}

func (s *WasmServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/")
	pathParts := strings.Split(path, "/")
	if len(pathParts) == 0 {
		writeResponse(w, http.StatusBadRequest, []byte("invalid endpoint id given"))
		return
	}
	id := pathParts[0]
	endpointID, err := uuid.Parse(id)
	if err != nil {
		writeResponse(w, http.StatusBadRequest, []byte(err.Error()))
		return
	}
	endpoint, err := s.store.GetEndpoint(endpointID)
	if err != nil {
		writeResponse(w, http.StatusNotFound, []byte(err.Error()))
		return
	}
	if !endpoint.HasActiveDeploy() {
		writeResponse(w, http.StatusNotFound, []byte("endpoint does not have an active deploy yet"))
		return
	}
	requestID := uuid.NewString()
	r.Header.Set("x-request-id", requestID)
	req, err := makeProtoRequest(requestID, r)
	if err != nil {
		writeResponse(w, http.StatusInternalServerError, []byte(err.Error()))
		return
	}
	req.Runtime = endpoint.Runtime
	req.EndpointID = endpointID.String()
	req.ActiveDeployID = endpoint.ActiveDeployID.String()
	req.Env = endpoint.Environment
	reqres := newRequestWithResponse(req)

	s.cluster.Engine().Send(s.self, reqres)

	resp := <-reqres.response

	w.WriteHeader(int(resp.StatusCode))
	w.Write(resp.Response)
}

func makeProtoRequest(id string, r *http.Request) (*proto.HTTPRequest, error) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	return &proto.HTTPRequest{
		Header: makeProtoHeader(r.Header),
		ID:     id,
		Body:   b,
		Method: r.Method,
		URL:    trimmedEndpointFromURL(r.URL),
	}, nil
}

func writeResponse(w http.ResponseWriter, code int, b []byte) {
	w.WriteHeader(http.StatusNotFound)
	w.Write(b)
}

func trimmedEndpointFromURL(url *url.URL) string {
	path := strings.TrimPrefix(url.Path, "/")
	pathParts := strings.Split(path, "/")
	if len(pathParts) == 0 {
		return "/"
	}
	return "/" + strings.Join(pathParts[1:], "/")
}

func makeProtoHeader(header http.Header) map[string]*proto.HeaderFields {
	m := make(map[string]*proto.HeaderFields, len(header))
	for k, v := range header {
		m[k] = &proto.HeaderFields{
			Fields: v,
		}
	}
	return m
}
