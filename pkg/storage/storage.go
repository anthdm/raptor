package storage

import (
	"github.com/anthdm/ffaas/pkg/types"
	"github.com/google/uuid"
)

type Store interface {
	CreateApplication(*types.Application) error
	UpdateApplication(uuid.UUID, UpdateAppParams) error
	GetApplication(uuid.UUID) (*types.Application, error)
	CreateDeploy(*types.Deploy) error
	GetDeploy(uuid.UUID) (*types.Deploy, error)
}
