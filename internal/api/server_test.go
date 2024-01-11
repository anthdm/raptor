package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/anthdm/raptor/internal/shared"
	"github.com/anthdm/raptor/internal/storage"
	"github.com/anthdm/raptor/internal/types"
	"github.com/stretchr/testify/require"
)

func TestCreateEndpoint(t *testing.T) {
	s := createServer()

	name := "My endpoint"
	runtime := "go"
	environment := map[string]string{"FOO": "BAR"}

	params := CreateEndpointParams{
		Name:        name,
		Runtime:     runtime,
		Environment: environment,
	}
	b, err := json.Marshal(params)
	require.Nil(t, err)

	req := httptest.NewRequest("POST", "/endpoint", bytes.NewReader(b))
	resp := httptest.NewRecorder()
	s.router.ServeHTTP(resp, req)

	var endpoint types.Endpoint
	err = json.NewDecoder(resp.Body).Decode(&endpoint)
	require.Nil(t, err)

	require.Equal(t, http.StatusOK, resp.Result().StatusCode)
	require.Equal(t, name, endpoint.Name)
	require.Equal(t, runtime, endpoint.Runtime)
	require.Equal(t, environment, endpoint.Environment)
	require.True(t, shared.IsZeroUUID(endpoint.ActiveDeploymentID))
}

func TestGetEndpoint(t *testing.T) {
	s := createServer()
	endpoint := seedEndpoint(t, s)

	req := httptest.NewRequest("GET", "/endpoint/"+endpoint.ID.String(), nil)
	resp := httptest.NewRecorder()
	s.router.ServeHTTP(resp, req)

	require.Equal(t, http.StatusOK, resp.Result().StatusCode)

	var other types.Endpoint
	err := json.NewDecoder(resp.Body).Decode(&other)
	require.Nil(t, err)
	endpoint.CreatedAT = time.Time{}
	other.CreatedAT = time.Time{}
	require.Equal(t, *endpoint, other)
}

func TestCreateDeploy(t *testing.T) {
	s := createServer()
	endpoint := seedEndpoint(t, s)

	req := httptest.NewRequest("POST", "/endpoint/"+endpoint.ID.String()+"/deployment", bytes.NewReader([]byte("a")))
	req.Header.Set("content-type", "application/octet-stream")
	resp := httptest.NewRecorder()
	s.router.ServeHTTP(resp, req)

	require.Equal(t, http.StatusOK, resp.Result().StatusCode)

	var deploy types.Deployment
	require.Nil(t, json.NewDecoder(resp.Body).Decode(&deploy))

	require.Equal(t, endpoint.ID, deploy.EndpointID)
	require.Equal(t, 32, len(deploy.Hash))
}

func TestPublish(t *testing.T) {
	s := createServer()
	endpoint := seedEndpoint(t, s)
	deployment := types.NewDeployment(endpoint, []byte("somefakeblob"))

	require.Nil(t, s.store.CreateDeployment(deployment))
	require.True(t, shared.IsZeroUUID(endpoint.ActiveDeploymentID))

	params := PublishParams{
		DeploymentID: deployment.ID,
	}
	b, err := json.Marshal(params)
	require.Nil(t, err)

	req := httptest.NewRequest("POST", "/publish", bytes.NewReader(b))
	resp := httptest.NewRecorder()
	s.router.ServeHTTP(resp, req)

	var publishResp PublishResponse
	require.Nil(t, json.NewDecoder(resp.Body).Decode(&publishResp))

	require.Equal(t, http.StatusOK, resp.Result().StatusCode)
	require.Equal(t, deployment.ID, endpoint.ActiveDeploymentID)
	require.Equal(t, deployment.ID, publishResp.DeploymentID)
	require.Equal(t, "http://0.0.0.0:80/live/"+endpoint.ID.String(), publishResp.URL)
}

func seedEndpoint(t *testing.T, s *Server) *types.Endpoint {
	e := types.NewEndpoint("My endpoint", "go", map[string]string{"FOO": "BAR"})
	require.Nil(t, s.store.CreateEndpoint(e))
	return e
}

func createServer() *Server {
	cache := storage.NewDefaultModCache()
	store := storage.NewMemoryStore()
	s := NewServer(store, store, cache)
	s.initRouter()
	return s
}
