package actrs

import (
	"log"
	"net/http"
	"strings"

	"github.com/anthdm/hollywood/actor"
	"github.com/anthdm/hollywood/cluster"
	"github.com/anthdm/raptor/pkg/shared"
	"github.com/anthdm/raptor/pkg/storage"
	"github.com/anthdm/raptor/proto"
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

// TODO(anthdm): Handle the favicon.ico
func (s *WasmServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/")
	path = strings.TrimSuffix(path, "/")
	pathParts := strings.Split(path, "/")

	if len(pathParts) < 2 {
		writeResponse(w, http.StatusBadRequest, []byte("invalid request url"))
		return
	}
	if pathParts[0] != "live" && pathParts[0] != "preview" {
		writeResponse(w, http.StatusBadRequest, []byte("invalid request url"))
		return
	}

	requestID := uuid.NewString()
	r.Header.Set("x-request-id", requestID)
	req, err := shared.MakeProtoRequest(requestID, r)
	if err != nil {
		writeResponse(w, http.StatusInternalServerError, []byte(err.Error()))
		return
	}

	if err != nil {
		writeResponse(w, http.StatusInternalServerError, []byte(err.Error()))
		return
	}
	if pathParts[0] == "live" {
		endpointID, err := uuid.Parse(pathParts[1])
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
			writeResponse(w, http.StatusNotFound, []byte("endpoint does not have any published deploy"))
			return
		}
		req.Runtime = endpoint.Runtime
		req.EndpointID = endpointID.String()
		// When serving LIVE endpoints we use the active deployment id.
		req.DeploymentID = endpoint.ActiveDeploymentID.String()
		req.Env = endpoint.Environment
		req.Preview = false
	}
	if pathParts[0] == "preview" {
		deployID, err := uuid.Parse(pathParts[1])
		if err != nil {
			writeResponse(w, http.StatusBadRequest, []byte(err.Error()))
			return
		}
		deploy, err := s.store.GetDeployment(deployID)
		if err != nil {
			writeResponse(w, http.StatusBadRequest, []byte(err.Error()))
			return
		}
		endpoint, err := s.store.GetEndpoint(deploy.EndpointID)
		if err != nil {
			writeResponse(w, http.StatusBadRequest, []byte(err.Error()))
			return
		}
		req.Runtime = endpoint.Runtime
		req.EndpointID = endpoint.ID.String()
		// When serving PREVIEW endpoints, we just use the deployment id from the
		// request.
		req.DeploymentID = deploy.ID.String()
		req.Env = endpoint.Environment
		req.Preview = true
	}

	reqres := newRequestWithResponse(req)
	s.cluster.Engine().Send(s.self, reqres)

	resp := <-reqres.response

	w.WriteHeader(int(resp.StatusCode))
	w.Write(resp.Response)
}

func writeResponse(w http.ResponseWriter, code int, b []byte) {
	w.WriteHeader(http.StatusNotFound)
	w.Write(b)
}
