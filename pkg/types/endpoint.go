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
	ID             uuid.UUID         `json:"id"`
	Name           string            `json:"name"`
	URL            string            `json:"url"`
	Runtime        string            `json:"runtime"`
	ActiveDeployID uuid.UUID         `json:"active_deploy_id"`
	Environment    map[string]string `json:"environment"`
	DeployHistory  []*Deploy         `json:"deploy_history"`
	CreatedAT      time.Time         `json:"created_at"`
}

func (e Endpoint) HasActiveDeploy() bool {
	return e.ActiveDeployID.String() != "00000000-0000-0000-0000-000000000000"
}

func NewEndpoint(name string, runtime string, env map[string]string) *Endpoint {
	if env == nil {
		env = make(map[string]string)
	}
	id := uuid.New()
	return &Endpoint{
		ID:            id,
		Name:          name,
		Environment:   env,
		Runtime:       runtime,
		URL:           "",
		DeployHistory: []*Deploy{},
		CreatedAT:     time.Now(),
	}
}
