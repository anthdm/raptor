package act

import (
	"log/slog"

	"github.com/anthdm/ffaas/pkg/storage"
	"github.com/anthdm/ffaas/pkg/types"
	"github.com/anthdm/hollywood/actor"
)

type Metric struct {
	store storage.MetricStore
}

func NewMetric(store storage.MetricStore) actor.Producer {
	return func() actor.Receiver {
		return &Metric{
			store: store,
		}
	}
}

func (m *Metric) Receive(c *actor.Context) {
	switch msg := c.Message().(type) {
	case actor.Started:
	case actor.Stopped:
	case *types.RuntimeMetric:
		if err := m.store.CreateRuntimeMetric(msg); err != nil {
			slog.Warn("failed to store RuntimeMetric", "err", err)
		}
	}
}
