package types

import (
	"time"

	"github.com/google/uuid"
)

type App struct {
	ID          uuid.UUID         `json:"id"`
	Name        string            `json:"name"`
	Environment map[string]string `json:"-"`
	CreatedAT   time.Time         `json:"createAt"`
}
