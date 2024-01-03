package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/anthdm/run/pkg/config"
	"github.com/anthdm/run/pkg/storage"
	"github.com/anthdm/run/pkg/types"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// Server serves the public run API.
type Server struct {
	router      *chi.Mux
	store       storage.Store
	metricStore storage.MetricStore
	cache       storage.ModCacher
}

// NewServer returns a new server given a Store interface.
func NewServer(store storage.Store, metricStore storage.MetricStore, cache storage.ModCacher) *Server {
	return &Server{
		store:       store,
		cache:       cache,
		metricStore: metricStore,
	}
}

// Listen starts listening on the given address.
func (s *Server) Listen(addr string) error {
	s.initRouter()
	return http.ListenAndServe(addr, s.router)
}

func (s *Server) initRouter() {
	s.router = chi.NewRouter()
	if config.Get().Authorization {
		s.router.Use(s.withAPIToken)
	}
	s.router.Get("/status", handleStatus)
	s.router.Get("/endpoint/{id}", makeAPIHandler(s.handleGetEndpoint))
	s.router.Get("/endpoint", makeAPIHandler(s.handleGetEndpoints))
	s.router.Get("/endpoint/{id}/metrics", makeAPIHandler(s.handleGetEndpointMetrics))
	s.router.Post("/endpoint", makeAPIHandler(s.handleCreateEndpoint))
	s.router.Post("/endpoint/{id}/deploy", makeAPIHandler(s.handleCreateDeploy))
	s.router.Post("/endpoint/{id}/rollback", makeAPIHandler(s.handleCreateRollback))
}

func handleStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	status := map[string]string{
		"status": "ok",
	}
	json.NewEncoder(w).Encode(status)
}

// CreateEndpointParams holds all the necessary fields to create a new run application.
type CreateEndpointParams struct {
	Name        string            `json:"name"`
	Environment map[string]string `json:"environment"`
}

func (p CreateEndpointParams) validate() error {
	minlen, maxlen := 3, 50
	if len(p.Name) < minlen {
		return fmt.Errorf("endpoint name should be at least %d characters long", minlen)
	}
	if len(p.Name) > maxlen {
		return fmt.Errorf("endpoint name can be maximum %d characters long", maxlen)
	}
	return nil
}

func (s *Server) handleCreateEndpoint(w http.ResponseWriter, r *http.Request) error {
	var params CreateEndpointParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		return writeJSON(w, http.StatusBadRequest, ErrorResponse(ErrDecodeRequestBody))
	}
	defer r.Body.Close()

	if err := params.validate(); err != nil {
		return writeJSON(w, http.StatusBadRequest, ErrorResponse(err))
	}

	endpoint := types.NewEndpoint(params.Name, params.Environment)
	endpoint.URL = config.GetWasmUrl() + "/" + endpoint.ID.String()
	if err := s.store.CreateEndpoint(endpoint); err != nil {
		return writeJSON(w, http.StatusBadRequest, ErrorResponse(err))
	}
	return writeJSON(w, http.StatusOK, endpoint)
}

// CreateDeployParams holds all the necessary fields to deploy a new function.
type CreateDeployParams struct{}

func (s *Server) handleCreateDeploy(w http.ResponseWriter, r *http.Request) error {
	endpointID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		return writeJSON(w, http.StatusBadRequest, ErrorResponse(err))
	}
	endpoint, err := s.store.GetEndpoint(endpointID)
	if err != nil {
		return writeJSON(w, http.StatusNotFound, ErrorResponse(err))
	}

	b, err := io.ReadAll(r.Body)
	if err != nil {
		return writeJSON(w, http.StatusNotFound, ErrorResponse(err))
	}
	deploy := types.NewDeploy(endpoint, b)
	if err := s.store.CreateDeploy(deploy); err != nil {
		return writeJSON(w, http.StatusUnprocessableEntity, ErrorResponse(err))
	}
	// Each new deploy will be the endpoint's active deploy
	err = s.store.UpdateEndpoint(endpointID, storage.UpdateEndpointParams{
		ActiveDeployID: deploy.ID,
		Deploys:        []*types.Deploy{deploy},
	})
	if err != nil {
		return writeJSON(w, http.StatusUnprocessableEntity, ErrorResponse(err))
	}
	return writeJSON(w, http.StatusOK, deploy)
}

func (s *Server) handleGetEndpoint(w http.ResponseWriter, r *http.Request) error {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		return writeJSON(w, http.StatusBadRequest, ErrorResponse(err))
	}
	endpoint, err := s.store.GetEndpoint(id)
	if err != nil {
		return writeJSON(w, http.StatusNotFound, ErrorResponse(err))
	}
	return writeJSON(w, http.StatusOK, endpoint)
}

func (s *Server) handleGetEndpoints(w http.ResponseWriter, r *http.Request) error {
	endpoints, err := s.store.GetEndpoints()
	if err != nil {
		return writeJSON(w, http.StatusNotFound, ErrorResponse(err))
	}
	return writeJSON(w, http.StatusOK, endpoints)
}

// CreateRollbackParams holds all the necessary fields to rollback your application
// to a specific deploy id (version).
type CreateRollbackParams struct {
	DeployID uuid.UUID `json:"deploy_id"`
}

type CreateRollbackResponse struct {
	DeployID uuid.UUID `json:"deploy_id"`
}

func (s *Server) handleCreateRollback(w http.ResponseWriter, r *http.Request) error {
	endpointID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		return writeJSON(w, http.StatusBadRequest, ErrorResponse(err))
	}
	endpoint, err := s.store.GetEndpoint(endpointID)
	if err != nil {
		return writeJSON(w, http.StatusBadRequest, ErrorResponse(err))
	}

	currentDeployID := endpoint.ActiveDeployID

	var params CreateRollbackParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		return writeJSON(w, http.StatusBadRequest, ErrorResponse(err))
	}

	if currentDeployID.String() == params.DeployID.String() {
		err := fmt.Errorf("deploy %s already active", params.DeployID)
		return writeJSON(w, http.StatusBadRequest, ErrorResponse(err))
	}

	deploy, err := s.store.GetDeploy(params.DeployID)
	if err != nil {
		return writeJSON(w, http.StatusNotFound, ErrorResponse(err))
	}

	updateParams := storage.UpdateEndpointParams{
		ActiveDeployID: deploy.ID,
		Deploys:        []*types.Deploy{deploy},
	}
	if err := s.store.UpdateEndpoint(endpointID, updateParams); err != nil {
		return writeJSON(w, http.StatusBadRequest, ErrorResponse(err))
	}

	s.cache.Delete(currentDeployID)

	resp := CreateRollbackResponse{DeployID: deploy.ID}
	return writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleGetEndpointMetrics(w http.ResponseWriter, r *http.Request) error {
	endpointID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		return writeJSON(w, http.StatusBadRequest, ErrorResponse(err))
	}
	metrics, err := s.metricStore.GetRuntimeMetrics(endpointID)
	if err != nil {
		return writeJSON(w, http.StatusNotFound, ErrorResponse(err))
	}
	return writeJSON(w, http.StatusOK, metrics)
}

var errUnauthorized = errors.New("unauthorized")

func (s *Server) withAPIToken(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if len(authHeader) < 10 {
			writeJSON(w, http.StatusUnauthorized, ErrorResponse(errUnauthorized))
			return
		}
		apiToken := authHeader[7:]
		if apiToken != config.Get().APIToken {
			writeJSON(w, http.StatusUnauthorized, ErrorResponse(errUnauthorized))
			return
		}
		h.ServeHTTP(w, r)
	})
}
