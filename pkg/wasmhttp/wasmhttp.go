package wasmhttp

import (
	"context"
	"net/http"
	"strings"

	"github.com/anthdm/ffaas/pkg/act"
	"github.com/anthdm/ffaas/pkg/runtime"
	"github.com/anthdm/ffaas/pkg/storage"
	"github.com/anthdm/hollywood/actor"
	"github.com/google/uuid"
	"github.com/tetratelabs/wazero"
)

// Server is an HTTP server that will proxy and route the request to the corresponding function.
type Server struct {
	server      *http.Server
	store       storage.Store
	metricStore storage.MetricStore
	cache       storage.ModCacher
	engine      *actor.Engine
}

// NewServer return a new wasm server given a storage and a mod cache.
func NewServer(addr string, engine *actor.Engine, store storage.Store, metricStore storage.MetricStore, cache storage.ModCacher) *Server {
	s := &Server{
		store:       store,
		metricStore: metricStore,
		cache:       cache,
		engine:      engine,
	}
	server := &http.Server{
		Handler: s,
		Addr:    addr,
	}
	s.server = server
	return s
}

func (s *Server) Listen() error {
	return s.server.ListenAndServe()
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/")
	pathParts := strings.Split(path, "/")
	if len(pathParts) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid endpoint id given"))
		return
	}
	id := pathParts[0]
	endpointID, err := uuid.Parse(id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(err.Error()))
		return
	}
	endpoint, err := s.store.GetEndpoint(endpointID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(err.Error()))
		return
	}
	if !endpoint.HasActiveDeploy() {
		w.WriteHeader(http.StatusNotFound)
		// TODO: might want to render something decent?
		w.Write([]byte("endpoint does not have an active deploy yet"))
		return
	}
	deploy, err := s.store.GetDeploy(endpoint.ActiveDeployID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(err.Error()))
		return
	}
	compCache, ok := s.cache.Get(endpoint.ID)
	if !ok {
		compCache = wazero.NewCompilationCache()
		s.cache.Put(endpoint.ID, compCache)
	}
	reqPlugin, err := runtime.NewRequestModule(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	args := runtime.Args{
		Blob:          deploy.Blob,
		Cache:         compCache,
		RequestPlugin: reqPlugin,
		Env:           endpoint.Environment,
	}
	// Trim the endpoint id from the path as the "actual" requestURL
	requestURL := "/" + strings.Join(pathParts[1:], "/")

	ctx, cancel := context.WithCancel(r.Context())
	rt := act.NewRuntime(w, args, endpointID, deploy.ID, s.metricStore, requestURL, cancel)
	s.engine.Spawn(rt, act.KindRuntime)
	<-ctx.Done()
}
