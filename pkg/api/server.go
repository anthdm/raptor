package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/anthdm/ffaas/pkg/storage"
	"github.com/anthdm/ffaas/pkg/types"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// Server serves the public ffaas API.
type Server struct {
	router *chi.Mux
	store  storage.Store
	cache  storage.ModCacher
}

// NewServer returns a new server given a Store interface.
func NewServer(store storage.Store, cache storage.ModCacher) *Server {
	return &Server{
		store: store,
		cache: cache,
	}
}

// Listen starts listening on the given address.
func (s *Server) Listen(addr string) error {
	s.initRouter()
	return http.ListenAndServe(addr, s.router)
}

func (s *Server) initRouter() {
	s.router = chi.NewRouter()
	s.router.Get("/status", handleStatus)
	s.router.Post("/app", makeAPIHandler(s.handleCreateApp))
	s.router.Post("/app/{id}", makeAPIHandler(s.handleCreateDeploy))
}

func handleStatus(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// CreateAppParams holds all the necessary fields to create a new ffaas application.
type CreateAppParams struct {
	Name string `json:"name"`
}

func (p CreateAppParams) validate() error {
	if len(p.Name) < 3 || len(p.Name) > 20 {
		return fmt.Errorf("name of the application should be longer than 3 and less than 20 characters")
	}
	return nil
}

func (s *Server) handleCreateApp(w http.ResponseWriter, r *http.Request) error {
	var params CreateAppParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		return writeJSON(w, http.StatusBadRequest, ErrorResponse(ErrDecodeRequestBody))
	}
	defer r.Body.Close()
	if err := params.validate(); err != nil {
		return writeJSON(w, http.StatusBadRequest, ErrorResponse(err))
	}
	app := types.App{
		ID:        uuid.New(),
		Name:      params.Name,
		CreatedAT: time.Now(),
	}
	if err := s.store.CreateApp(&app); err != nil {
		return writeJSON(w, http.StatusBadRequest, ErrorResponse(err))
	}
	return writeJSON(w, http.StatusOK, app)
}

// CreateDeployParams holds all the necessary fields to deploy a new function.
type CreateDeployParams struct{}

func (s *Server) handleCreateDeploy(w http.ResponseWriter, r *http.Request) error {
	appID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		return writeJSON(w, http.StatusBadRequest, ErrorResponse(err))
	}
	app, err := s.store.GetAppByID(appID)
	if err != nil {
		return writeJSON(w, http.StatusNotFound, ErrorResponse(err))
	}

	b, err := io.ReadAll(r.Body)
	if err != nil {
		return writeJSON(w, http.StatusNotFound, ErrorResponse(err))
	}

	deploy := types.Deploy{
		ID:        uuid.New(),
		AppID:     app.ID,
		Blob:      b,
		CreatedAT: time.Now(),
	}
	if err := s.store.CreateDeploy(&deploy); err != nil {
		return writeJSON(w, http.StatusUnprocessableEntity, ErrorResponse(err))
	}

	return writeJSON(w, http.StatusOK, deploy)
}
