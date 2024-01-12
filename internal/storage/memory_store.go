package storage

import (
	"fmt"
	"sync"

	"github.com/anthdm/raptor/internal/types"
	"github.com/google/uuid"
)

type MemoryStore struct {
	mu        sync.RWMutex
	endpoints map[uuid.UUID]*types.Endpoint
	deploys   map[uuid.UUID]*types.Deployment
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		endpoints: make(map[uuid.UUID]*types.Endpoint),
		deploys:   make(map[uuid.UUID]*types.Deployment),
	}
}

func (s *MemoryStore) CreateEndpoint(e *types.Endpoint) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.endpoints[e.ID] = e
	return nil
}

func (s *MemoryStore) GetEndpoint(id uuid.UUID) (*types.Endpoint, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.endpoints[id]
	if !ok {
		return nil, fmt.Errorf("could not find endpoint with id (%s)", id)
	}
	return e, nil
}

func (s *MemoryStore) UpdateEndpoint(id uuid.UUID, params UpdateEndpointParams) error {
	endpoint, err := s.GetEndpoint(id)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if params.ActiveDeployID.String() != "00000000-0000-0000-0000-000000000000" {
		endpoint.ActiveDeploymentID = params.ActiveDeployID
	}
	if params.Environment != nil {
		for key, val := range params.Environment {
			endpoint.Environment[key] = val
		}
	}
	return nil
}

func (s *MemoryStore) CreateDeployment(deploy *types.Deployment) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.deploys[deploy.ID] = deploy
	return nil
}

func (s *MemoryStore) GetDeployment(id uuid.UUID) (*types.Deployment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	deploy, ok := s.deploys[id]
	if !ok {
		return nil, fmt.Errorf("could not find deployment with id (%s)", id)
	}
	return deploy, nil
}

func (s *MemoryStore) GetDeployments() ([]*types.Deployment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	deployments := make([]*types.Deployment, 0, len(s.deploys))
	for _, deploy := range s.deploys {
		deployments = append(deployments, deploy)
	}

	return deployments, nil
}

func (s *MemoryStore) CreateRuntimeMetric(_ *types.RuntimeMetric) error {
	return nil
}

func (s *MemoryStore) GetRuntimeMetrics(_ uuid.UUID) ([]types.RuntimeMetric, error) {
	return nil, nil
}
