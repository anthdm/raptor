package storage

import (
	"github.com/anthdm/raptor/pkg/types"
	"github.com/google/uuid"
)

type Store interface {
	CreateEndpoint(*types.Endpoint) error
	UpdateEndpoint(uuid.UUID, UpdateEndpointParams) error
	GetEndpoint(uuid.UUID) (*types.Endpoint, error)
	GetEndpoints() ([]types.Endpoint, error)
	CreateDeploy(*types.Deploy) error
	GetDeploy(uuid.UUID) (*types.Deploy, error)
}

type MetricStore interface {
	CreateRuntimeMetric(*types.RuntimeMetric) error
	GetRuntimeMetrics(uuid.UUID) ([]types.RuntimeMetric, error)
}

type UpdateEndpointParams struct {
	Environment    map[string]string
	ActiveDeployID uuid.UUID
	DeployHistory  *types.DeployHistory
}
