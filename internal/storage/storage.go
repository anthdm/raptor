package storage

import (
	"github.com/anthdm/raptor/internal/types"
	"github.com/google/uuid"
)

type Store interface {
	CreateEndpoint(*types.Endpoint) error
	UpdateEndpoint(uuid.UUID, UpdateEndpointParams) error
	GetEndpoint(uuid.UUID) (*types.Endpoint, error)
	CreateDeployment(*types.Deployment) error
	GetDeployment(uuid.UUID) (*types.Deployment, error)
}

type MetricStore interface {
	CreateRuntimeMetric(*types.RuntimeMetric) error
	GetRuntimeMetrics(uuid.UUID) ([]types.RuntimeMetric, error)
}

type UpdateEndpointParams struct {
	Environment       map[string]string
	ActiveDeployID    uuid.UUID
	DeploymentHistory *types.DeploymentHistory
}
