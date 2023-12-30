package types

import (
	"time"

	"github.com/google/uuid"
)

// RuntimeMetric holds information about a single runtime execution.
type RuntimeMetric struct {
	ID         uuid.UUID     `json:"id"`
	EndpointID uuid.UUID     `json:"endpoint_id"`
	DeployID   uuid.UUID     `json:"deploy_id"`
	RequestURL string        `json:"request_url"`
	Duration   time.Duration `json:"duration"`
	StartTime  time.Time     `json:"start_time"`
}
