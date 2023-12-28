package proxy

import (
	"context"
	"net/http"

	"github.com/anthdm/ffaas/pkg/runtime"
	"github.com/anthdm/ffaas/pkg/storage"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/tetratelabs/wazero"
)

// Server is an HTTP server that will proxy and route the request to the corresponding function.
type Server struct {
	router *chi.Mux
	store  storage.Store
	cache  storage.ModCacher
}

// NewServer return a new (proxy) server given a storage.
func NewServer(store storage.Store, cache storage.ModCacher) *Server {
	return &Server{
		router: chi.NewRouter(),
		store:  store,
		cache:  cache,
	}
}

// Listen starts listening on the given address.
func (s *Server) Listen(addr string) error {
	s.initRoutes()
	return http.ListenAndServe(addr, s.router)
}

func (s *Server) initRoutes() {
	s.router.Handle("/{id}", http.HandlerFunc(s.handleRequest))
}

func (s *Server) handleRequest(w http.ResponseWriter, r *http.Request) {
	appID, err := uuid.Parse(chi.URLParam(r, ("id")))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(err.Error()))
		return
	}
	app, err := s.store.GetAppByID(appID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(err.Error()))
		return
	}
	if !app.HasActiveDeploy() {
		w.WriteHeader(http.StatusNotFound)
		// TODO: might want to render something decent?
		w.Write([]byte("application does not have an active deploy yet"))
		return

	}
	deploy, err := s.store.GetDeployByID(app.ActiveDeploy)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(err.Error()))
		return
	}
	compCache, ok := s.cache.Get(app.ID)
	if !ok {
		compCache = wazero.NewCompilationCache()
		s.cache.Put(app.ID, compCache)
	}
	run, err := runtime.New(compCache, deploy.Blob)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	if err := run.Exec(context.Background(), r); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Write(run.Response())
	run.Close(r.Context())
}
