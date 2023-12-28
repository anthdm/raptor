package types

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type App struct {
	ID           uuid.UUID         `json:"id"`
	Name         string            `json:"name"`
	Endpoint     string            `json:"endpoint"`
	ActiveDeploy uuid.UUID         `json:"activeDeploy"`
	Environment  map[string]string `json:"-"`
	Deploys      []Deploy          `json:"deploys"`
	CreatedAT    time.Time         `json:"createdAt"`
}

func (app App) HasActiveDeploy() bool {
	return app.ActiveDeploy.String() != "00000000-0000-0000-0000-000000000000"
}

func NewApp(name string, env map[string]string) *App {
	if env == nil {
		env = make(map[string]string)
	}
	id := uuid.New()
	return &App{
		ID:          id,
		Name:        name,
		Environment: env,
		// TODO: This is hardcoded AF :(
		Endpoint:  fmt.Sprintf("http://localhost:5000/%s", id),
		Deploys:   []Deploy{},
		CreatedAT: time.Now(),
	}
}
