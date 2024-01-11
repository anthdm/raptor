package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/anthdm/raptor/internal/config"
	"github.com/anthdm/raptor/internal/storage"
	"github.com/anthdm/raptor/internal/types"
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
	s.router.Post("/endpoint/{id}/deployment", makeAPIHandler(s.handleCreateDeployment))
	s.router.Post("/publish", makeAPIHandler(s.handlePublish))
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
	// Name of the endpoint
	Name string `json:"name"`
	// Runtime on which the code will be invoked. (go or js for now)
	Runtime string `json:"runtime"`
	// A map of environment variables
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
	if _, ok := types.Runtimes[p.Runtime]; !ok {
		return fmt.Errorf("invalid runtime given: %s", p.Runtime)
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

	endpoint := types.NewEndpoint(params.Name, params.Runtime, params.Environment)
	if err := s.store.CreateEndpoint(endpoint); err != nil {
		return writeJSON(w, http.StatusBadRequest, ErrorResponse(err))
	}
	return writeJSON(w, http.StatusOK, endpoint)
}

// CreateDeploymentParams holds all the necessary fields to deploy a new function.
type CreateDeploymentParams struct{}

func (s *Server) handleCreateDeployment(w http.ResponseWriter, r *http.Request) error {
	endpointID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		return writeJSON(w, http.StatusBadRequest, ErrorResponse(err))
	}
	endpoint, err := s.store.GetEndpoint(endpointID)
	if err != nil {
		return writeJSON(w, http.StatusNotFound, ErrorResponse(err))
	}

	// TODO:
	// 1. validate the contents of the blob.
	// 2. make sure we have a limit on the maximum blob size.
	b, err := io.ReadAll(r.Body)
	if err != nil {
		return writeJSON(w, http.StatusBadRequest, ErrorResponse(err))
	}
	if len(b) == 0 {
		err := fmt.Errorf("no blob")
		return writeJSON(w, http.StatusBadRequest, ErrorResponse(err))
	}
	deploy := types.NewDeployment(endpoint, b)
	if err := s.store.CreateDeployment(deploy); err != nil {
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
	return nil
	// endpoints, err := s.store.GetEndpoints()
	// if err != nil {
	// 	return writeJSON(w, http.StatusNotFound, ErrorResponse(err))
	// }
	// return writeJSON(w, http.StatusOK, endpoints)
}

// PublishParams holds all the necessary fields to publish a specific
// deployment LIVE to your application.
type PublishParams struct {
	DeploymentID uuid.UUID `json:"deployment_id"`
}

type PublishResponse struct {
	DeploymentID uuid.UUID `json:"deployment_id"`
	URL          string    `json:"url"`
}

func (s *Server) handlePublish(w http.ResponseWriter, r *http.Request) error {
	var params PublishParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		err := fmt.Errorf("failed to parse the response body: %s", err)
		return writeJSON(w, http.StatusBadRequest, ErrorResponse(err))
	}
	deploy, err := s.store.GetDeployment(params.DeploymentID)
	if err != nil {
		return writeJSON(w, http.StatusBadRequest, ErrorResponse(err))
	}
	endpoint, err := s.store.GetEndpoint(deploy.EndpointID)
	if err != nil {
		return writeJSON(w, http.StatusBadRequest, ErrorResponse(err))
	}

	currentDeploymentID := endpoint.ActiveDeploymentID

	if currentDeploymentID.String() == deploy.ID.String() {
		err := fmt.Errorf("deploy %s already active", deploy.ID)
		return writeJSON(w, http.StatusBadRequest, ErrorResponse(err))
	}

	updateParams := storage.UpdateEndpointParams{
		ActiveDeployID: deploy.ID,
	}
	if err := s.store.UpdateEndpoint(deploy.EndpointID, updateParams); err != nil {
		return writeJSON(w, http.StatusBadRequest, ErrorResponse(err))
	}

	s.cache.Delete(currentDeploymentID)

	resp := PublishResponse{
		DeploymentID: deploy.ID,
		URL:          fmt.Sprintf("%s/live/%s", config.IngressUrl(), endpoint.ID),
	}
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
