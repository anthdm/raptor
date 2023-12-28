package storage

import (
	"github.com/anthdm/ffaas/pkg/types"
	"github.com/google/uuid"
)

type Store interface {
	CreateApp(*types.App) error
	UpdateApp(uuid.UUID, UpdateAppParams) error
	CreateDeploy(*types.Deploy) error
	GetDeployByID(uuid.UUID) (*types.Deploy, error)
	GetAppByID(uuid.UUID) (*types.App, error)
}
