package types

import (
	"time"

	"github.com/google/uuid"
)

type Endpoint struct {
	ID             uuid.UUID         `json:"id"`
	Name           string            `json:"name"`
	URL            string            `json:"url"`
	ActiveDeployID uuid.UUID         `json:"active_deploy_id"`
	Environment    map[string]string `json:"-"`
	DeployHistory  []*Deploy         `json:"deploy_history"`
	CreatedAT      time.Time         `json:"created_at"`
}

func (e Endpoint) HasActiveDeploy() bool {
	return e.ActiveDeployID.String() != "00000000-0000-0000-0000-000000000000"
}

func NewEndpoint(name string, env map[string]string) *Endpoint {
	if env == nil {
		env = make(map[string]string)
	}
	id := uuid.New()
	return &Endpoint{
		ID:            id,
		Name:          name,
		Environment:   env,
		URL:           "",
		DeployHistory: []*Deploy{},
		CreatedAT:     time.Now(),
	}
}
