package storage

import (
	"github.com/anthdm/ffaas/pkg/types"
	"github.com/google/uuid"
)

type Store interface {
	CreateApp(*types.Application) error
	UpdateApp(uuid.UUID, UpdateAppParams) error
	CreateDeploy(*types.Deploy) error
	GetDeployByID(uuid.UUID) (*types.Deploy, error)
	GetAppByID(uuid.UUID) (*types.Application, error)
}
