package types

import (
	"time"

	"github.com/google/uuid"
)

var Runtimes = map[string]bool{
	"js": true,
	"go": true,
}

func ValidRuntime(runtime string) bool {
	_, ok := Runtimes[runtime]
	return ok
}

type Endpoint struct {
	ID                 uuid.UUID            `json:"id"`
	Name               string               `json:"name"`
	URL                string               `json:"url"`
	Runtime            string               `json:"runtime"`
	ActiveDeploymentID uuid.UUID            `json:"active_deployment_id"`
	Environment        map[string]string    `json:"environment"`
	DeploymentHistory  []*DeploymentHistory `json:"deployment_history"`
	CreatedAT          time.Time            `json:"created_at"`
}

func (e Endpoint) HasActiveDeploy() bool {
	return e.ActiveDeploymentID.String() != "00000000-0000-0000-0000-000000000000"
}

func NewEndpoint(name string, runtime string, env map[string]string) *Endpoint {
	if env == nil {
		env = make(map[string]string)
	}
	id := uuid.New()
	return &Endpoint{
		ID:                id,
		Name:              name,
		Environment:       env,
		Runtime:           runtime,
		URL:               "",
		DeploymentHistory: []*DeploymentHistory{},
		CreatedAT:         time.Now(),
	}
}

type DeploymentHistory struct {
	ID        uuid.UUID `json:"id"`
	CreatedAT time.Time `json:"created_at"`
}
