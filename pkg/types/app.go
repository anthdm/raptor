package types

import (
	"time"

	"github.com/google/uuid"
)

type Application struct {
	ID             uuid.UUID         `json:"id"`
	Name           string            `json:"name"`
	Endpoint       string            `json:"endpoint"`
	ActiveDeployID uuid.UUID         `json:"active_deploy_id"`
	Environment    map[string]string `json:"-"`
	DeployHistory  []Deploy          `json:"deploy_history"`
	CreatedAT      time.Time         `json:"created_at"`
}

func (app Application) HasActiveDeploy() bool {
	return app.ActiveDeployID.String() != "00000000-0000-0000-0000-000000000000"
}

func NewApplication(name string, env map[string]string) *Application {
	if env == nil {
		env = make(map[string]string)
	}
	id := uuid.New()
	return &Application{
		ID:            id,
		Name:          name,
		Environment:   env,
		Endpoint:      "",
		DeployHistory: []Deploy{},
		CreatedAT:     time.Now(),
	}
}
