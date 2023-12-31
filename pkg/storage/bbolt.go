package storage

import (
	"os"

	"github.com/anthdm/ffaas/pkg/types"
	"github.com/google/uuid"
	"github.com/vmihailenco/msgpack/v5"
	"go.etcd.io/bbolt"
)

type BoltConfig struct {
	path     string
	readonly bool
}

func NewBoltConfig() BoltConfig {
	return BoltConfig{
		path:     ".db",
		readonly: false,
	}
}

func (config BoltConfig) WithPath(path string) BoltConfig {
	config.path = path
	return config
}

func (config BoltConfig) WithReadOnly(b bool) BoltConfig {
	config.readonly = true
	return config
}

type BoltStore struct {
	config BoltConfig
	db     *bbolt.DB
}

func NewBoltStore(config BoltConfig) (*BoltStore, error) {
	var init bool
	if _, err := os.Stat(config.path); err != nil {
		init = true
	}
	db, err := bbolt.Open(config.path, 0600, &bbolt.Options{
		ReadOnly: config.readonly,
	})
	if err != nil {
		return nil, err
	}

	if init {
		tx, err := db.Begin(true)
		if err != nil {
			return nil, err
		}
		defer tx.Rollback()

		_, err = tx.CreateBucket([]byte("endpoint"))
		if err != nil {
			return nil, err
		}
		_, err = tx.CreateBucket([]byte("deploy"))
		if err != nil {
			return nil, err
		}
		if err := tx.Commit(); err != nil {
			return nil, err
		}
	}

	return &BoltStore{
		config: config,
		db:     db,
	}, nil
}

func (s *BoltStore) CreateEndpoint(e *types.Endpoint) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("endpoint"))
		b, err := msgpack.Marshal(e)
		if err != nil {
			return err
		}
		return bucket.Put([]byte(e.ID.String()), b)
	})
}

func (s *BoltStore) GetEndpoint(id uuid.UUID) (*types.Endpoint, error) {
	var endpoint types.Endpoint
	err := s.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("endpoint"))
		v := bucket.Get([]byte(id.String()))
		return msgpack.Unmarshal(v, &endpoint)
	})
	return &endpoint, err
}

func (s *BoltStore) UpdateEndpoint(id uuid.UUID, params UpdateEndpointParams) error {
	var endpoint types.Endpoint
	err := s.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("endpoint"))
		bid := []byte(id.String())
		v := bucket.Get(bid)
		if err := msgpack.Unmarshal(v, &endpoint); err != nil {
			return err
		}
		if params.ActiveDeployID.String() != "00000000-0000-0000-0000-000000000000" {
			endpoint.ActiveDeployID = params.ActiveDeployID
		}
		if params.Environment != nil {
			for key, val := range params.Environment {
				endpoint.Environment[key] = val
			}
		}
		if len(params.Deploys) > 0 {
			endpoint.DeployHistory = append(endpoint.DeployHistory, params.Deploys...)
		}
		b, err := msgpack.Marshal(endpoint)
		if err != nil {
			return err
		}
		return bucket.Put(bid, b)
	})
	return err
}

func (s *BoltStore) CreateDeploy(deploy *types.Deploy) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("deploy"))
		b, err := msgpack.Marshal(deploy)
		if err != nil {
			return err
		}
		return bucket.Put([]byte(deploy.ID.String()), b)
	})
}

func (s *BoltStore) GetDeploy(id uuid.UUID) (*types.Deploy, error) {
	var deploy types.Deploy
	err := s.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("deploy"))
		v := bucket.Get([]byte(id.String()))
		return msgpack.Unmarshal(v, &deploy)
	})
	return &deploy, err
}
