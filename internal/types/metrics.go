package types

import (
	"time"

	"github.com/google/uuid"
)

type RuntimeMetric struct {
	ID     uuid.UUID     `json:"id"`
	Uptime time.Duration `json:"uptime"`
}

// RequestMetric holds information about a single HTTP request
// invoked by the runtime.
type RequestMetric struct {
	ID           uuid.UUID     `json:"id"`
	EndpointID   uuid.UUID     `json:"endpoint_id"`
	DeploymentID uuid.UUID     `json:"deployment_id"`
	RequestURL   string        `json:"request_url"`
	Duration     time.Duration `json:"duration"`
	StatusCode   int           `json:"status_code"`
}
