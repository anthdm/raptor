package actrs

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/anthdm/ffaas/pkg/storage"
	"github.com/anthdm/ffaas/proto"
	"github.com/anthdm/hollywood/actor"
	"github.com/anthdm/hollywood/cluster"
	"github.com/google/uuid"
)

const KindWasmServer = "wasm_server"

type requestWithCancel struct {
	request *proto.HTTPRequest
	cancel  context.CancelFunc
}

// WasmServer is an HTTP server that will proxy and route the request to the corresponding function.
type WasmServer struct {
	server      *http.Server
	self        *actor.PID
	store       storage.Store
	metricStore storage.MetricStore
	cache       storage.ModCacher
	cluster     *cluster.Cluster

	requests map[string]context.CancelFunc
}

// NewWasmServer return a new wasm server given a storage and a mod cache.
func NewWasmServer(addr string, cluster *cluster.Cluster, store storage.Store, metricStore storage.MetricStore, cache storage.ModCacher) actor.Producer {
	return func() actor.Receiver {
		s := &WasmServer{
			store:       store,
			metricStore: metricStore,
			cache:       cache,
			cluster:     cluster,
			requests:    make(map[string]context.CancelFunc),
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
	case requestWithCancel:
		s.requests[msg.request.ID] = msg.cancel
		fmt.Println("got from myself", msg)
		s.sendRequestToRuntime(msg.request)
	case *proto.HTTPResponse:
		fmt.Println("received resposne from runtime", msg)
		// if cancel, ok := s.requests[msg.RequestID]; ok {
		// 	cancel()
		// }
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
	req, err := makeProtoRequest(r)
	if err != nil {
		writeResponse(w, http.StatusInternalServerError, []byte(err.Error()))
		return
	}
	req.EndpointID = endpointID.String()
	ctx, cancel := context.WithCancel(r.Context())
	s.cluster.Engine().Send(s.self, requestWithCancel{
		request: req,
		cancel:  cancel,
	})
	<-ctx.Done()
}

func makeProtoRequest(r *http.Request) (*proto.HTTPRequest, error) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	return &proto.HTTPRequest{
		ID:     uuid.NewString(),
		Body:   b,
		Method: r.Method,
		URL:    r.URL.String(),
	}, nil
}

func writeResponse(w http.ResponseWriter, code int, b []byte) {
	w.WriteHeader(http.StatusNotFound)
	w.Write(b)
}
