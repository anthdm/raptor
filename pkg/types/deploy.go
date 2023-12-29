package types

import (
	"crypto/md5"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
)

type Deploy struct {
	ID        uuid.UUID `json:"id"`
	AppID     uuid.UUID `json:"app_id"`
	Hash      string    `json:"hash"`
	Blob      []byte    `json:"-"`
	CreatedAT time.Time `json:"created_at"`
}

func NewDeploy(app *Application, blob []byte) *Deploy {
	hashBytes := md5.Sum(blob)
	hashstr := hex.EncodeToString(hashBytes[:])
	deployID := uuid.New()
	return &Deploy{
		ID:        deployID,
		AppID:     app.ID,
		Blob:      blob,
		Hash:      hashstr,
		CreatedAT: time.Now(),
	}
}
