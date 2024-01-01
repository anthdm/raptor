package storage

import (
	"fmt"
	"sync"

	"github.com/anthdm/ffaas/pkg/types"
	"github.com/google/uuid"
)

type MemoryMetricStore struct {
	mu   sync.RWMutex
	data map[uuid.UUID][]types.RuntimeMetric
}

func NewMemoryMetricStore() *MemoryMetricStore {
	return &MemoryMetricStore{
		data: make(map[uuid.UUID][]types.RuntimeMetric),
	}
}

func (s *MemoryMetricStore) CreateRuntimeMetric(metric *types.RuntimeMetric) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	var (
		metrics []types.RuntimeMetric
		ok      bool
	)
	metrics, ok = s.data[metric.EndpointID]
	if !ok {
		metrics = make([]types.RuntimeMetric, 0)
	}
	metrics = append(metrics, *metric)
	s.data[metric.EndpointID] = metrics
	return nil
}

func (s *MemoryMetricStore) GetRuntimeMetrics(endpointID uuid.UUID) ([]types.RuntimeMetric, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	metrics, ok := s.data[endpointID]
	if !ok {
		return nil, fmt.Errorf("could not find metrics for endpoint (%s)", endpointID)
	}
	return metrics, nil
}
