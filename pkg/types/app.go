package types

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type App struct {
	ID          uuid.UUID         `json:"id"`
	Name        string            `json:"name"`
	Endpoint    string            `json:"endpoint"`
	Active      uuid.UUID         `json:"active"`
	Environment map[string]string `json:"-"`
	Deploys     []Deploy          `json:"builds"`
	CreatedAT   time.Time         `json:"createAt"`
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
		Endpoint:    fmt.Sprintf("http://localhost:3000/%s", id),
		Deploys:     []Deploy{},
		CreatedAT:   time.Now(),
	}
}
