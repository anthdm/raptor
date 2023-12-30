package types

import (
	"crypto/md5"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
)

type Deploy struct {
	ID         uuid.UUID `json:"id"`
	EndpointID uuid.UUID `json:"endpoint_id"`
	Hash       string    `json:"hash"`
	Blob       []byte    `json:"-"`
	CreatedAT  time.Time `json:"created_at"`
}

func NewDeploy(endpoint *Endpoint, blob []byte) *Deploy {
	hashBytes := md5.Sum(blob)
	hashstr := hex.EncodeToString(hashBytes[:])
	deployID := uuid.New()
	return &Deploy{
		ID:         deployID,
		EndpointID: endpoint.ID,
		Blob:       blob,
		Hash:       hashstr,
		CreatedAT:  time.Now(),
	}
}
