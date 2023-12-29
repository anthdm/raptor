package storage

import (
	"fmt"
	"sync"

	"github.com/anthdm/ffaas/pkg/types"
	"github.com/google/uuid"
)

type MemoryStore struct {
	mu      sync.RWMutex
	apps    map[uuid.UUID]*types.Application
	deploys map[uuid.UUID]*types.Deploy
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		apps:    make(map[uuid.UUID]*types.Application),
		deploys: make(map[uuid.UUID]*types.Deploy),
	}
}

func (s *MemoryStore) CreateApp(app *types.Application) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.apps[app.ID] = app
	return nil
}

func (s *MemoryStore) GetAppByID(id uuid.UUID) (*types.Application, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	app, ok := s.apps[id]
	if !ok {
		return nil, fmt.Errorf("could not find app with id (%s)", id)
	}
	return app, nil
}

type UpdateAppParams struct {
	Environment  map[string]string
	ActiveDeploy uuid.UUID
}

func (s *MemoryStore) UpdateApp(id uuid.UUID, params UpdateAppParams) error {
	app, err := s.GetAppByID(id)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if params.ActiveDeploy.String() != "00000000-0000-0000-0000-000000000000" {
		app.ActiveDeployID = params.ActiveDeploy
	}
	if params.Environment != nil {
		for key, val := range params.Environment {
			app.Environment[key] = val
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

func (s *MemoryStore) GetDeployByID(id uuid.UUID) (*types.Deploy, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	deploy, ok := s.deploys[id]
	if !ok {
		return nil, fmt.Errorf("could not find deployment with id (%s)", id)
	}
	return deploy, nil
}
