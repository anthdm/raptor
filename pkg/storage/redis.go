package storage

import (
	"context"
	"fmt"

	"github.com/anthdm/ffaas/pkg/types"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/vmihailenco/msgpack/v5"
)

type RedisStore struct {
	client *redis.Client
}

func NewRedisStore() (*RedisStore, error) {
	client := redis.NewClient(&redis.Options{})
	err := client.Ping(context.Background()).Err()
	if err != nil {
		err = fmt.Errorf("failed to connect to the Redis server: %s", err)
		return nil, err
	}
	return &RedisStore{
		client: client,
	}, nil
}

func (s *RedisStore) CreateEndpoint(endpoint *types.Endpoint) error {
	b, err := msgpack.Marshal(endpoint)
	if err != nil {
		return err
	}
	return s.client.Set(context.Background(), endpoint.ID.String(), b, 0).Err()
}

func (s *RedisStore) GetEndpoint(id uuid.UUID) (*types.Endpoint, error) {
	var endpoint types.Endpoint
	b, err := s.client.Get(context.Background(), id.String()).Bytes()
	if err != nil {
		return nil, err
	}
	err = msgpack.Unmarshal(b, &endpoint)
	return &endpoint, err
}

func (s *RedisStore) UpdateEndpoint(id uuid.UUID, params UpdateEndpointParams) error {
	endpoint, err := s.GetEndpoint(id)
	if err != nil {
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
	return s.client.Set(context.Background(), endpoint.ID.String(), b, 0).Err()
}

func (s *RedisStore) GetDeploy(id uuid.UUID) (*types.Deploy, error) {
	var deploy types.Deploy
	b, err := s.client.Get(context.Background(), id.String()).Bytes()
	if err != nil {
		return nil, err
	}
	err = msgpack.Unmarshal(b, &deploy)
	return &deploy, err
}

func (s *RedisStore) CreateDeploy(deploy *types.Deploy) error {
	b, err := msgpack.Marshal(deploy)
	if err != nil {
		return err
	}
	return s.client.Set(context.Background(), deploy.ID.String(), b, 0).Err()
}

func (s *RedisStore) CreateRuntimeMetric(metric *types.RuntimeMetric) error {
	b, err := msgpack.Marshal(metric)
	if err != nil {
		return err
	}
	return s.client.Set(context.Background(), metric.ID.String(), b, 0).Err()
}

func (s *RedisStore) GetRuntimeMetrics(id uuid.UUID) ([]types.RuntimeMetric, error) {
	var metrics []types.RuntimeMetric
	b, err := s.client.Get(context.Background(), id.String()).Bytes()
	if err != nil {
		return nil, err
	}
	err = msgpack.Unmarshal(b, &metrics)
	return metrics, err
}
