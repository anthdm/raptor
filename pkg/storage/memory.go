package storage

import (
	"fmt"
	"sync"

	"github.com/anthdm/ffaas/pkg/types"
	"github.com/google/uuid"
)

type MemoryStore struct {
	mu        sync.RWMutex
	endpoints map[uuid.UUID]*types.Endpoint
	deploys   map[uuid.UUID]*types.Deploy
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		endpoints: make(map[uuid.UUID]*types.Endpoint),
		deploys:   make(map[uuid.UUID]*types.Deploy),
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
		endpoint.ActiveDeployID = params.ActiveDeployID
	}
	if params.Environment != nil {
		for key, val := range params.Environment {
			endpoint.Environment[key] = val
		}
	}
	return nil
}

func (s *MemoryStore) CreateDeploy(deploy *types.Deploy) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.deploys[deploy.ID] = deploy
	return nil
}

func (s *MemoryStore) GetDeploy(id uuid.UUID) (*types.Deploy, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	deploy, ok := s.deploys[id]
	if !ok {
		return nil, fmt.Errorf("could not find deployment with id (%s)", id)
	}
	return deploy, nil
}
