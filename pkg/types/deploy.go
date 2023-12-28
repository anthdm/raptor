package types

import (
	"crypto/md5"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
)

type Deploy struct {
	ID        uuid.UUID `json:"id"`
	AppID     uuid.UUID `json:"appId"`
	Hash      string    `json:"hash"`
	Blob      []byte    `json:"-"`
	CreatedAT time.Time `json:"createdAt"`
}

func NewDeploy(appID uuid.UUID, blob []byte) *Deploy {
	hashBytes := md5.Sum(blob)
	hashstr := hex.EncodeToString(hashBytes[:])
	return &Deploy{
		ID:        uuid.New(),
		AppID:     appID,
		Blob:      blob,
		Hash:      hashstr,
		CreatedAT: time.Now(),
	}
}
