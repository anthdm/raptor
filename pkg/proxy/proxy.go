package proxy

import (
	"context"
	"net/http"

	"github.com/anthdm/ffaas/pkg/runtime"
	"github.com/anthdm/ffaas/pkg/storage"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// Server is an HTTP server that will proxy and route the request to the corresponding function.
type Server struct {
	router *chi.Mux
	store  storage.Store
}

// NewServer return a new (proxy) server given a storage.
func NewServer(store storage.Store) *Server {
	return &Server{
		router: chi.NewRouter(),
		store:  store,
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
	deployID, err := uuid.Parse(chi.URLParam(r, ("id")))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(err.Error()))
		return
	}
	deploy, err := s.store.GetDeployByID(deployID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(err.Error()))
		return
	}
	app, err := s.store.GetAppByID(deploy.AppID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(err.Error()))
		return
	}
	run, err := runtime.New(deploy.Blob, app.Environment)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(err.Error()))
		return
	}
	defer run.Close(context.Background())

	if err := run.HandleHTTP(w, r); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
}
