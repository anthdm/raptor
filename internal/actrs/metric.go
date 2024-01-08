package actrs

import (
	"github.com/anthdm/hollywood/actor"
	"github.com/anthdm/raptor/internal/types"
)

// The metric actor is responsible for handling metrics that are being
// sent from the runtimes locally from the same machine.

const KindMetric = "runtime_metric"

type Metric struct{}

func NewMetric() actor.Receiver {
	return &Metric{}
}

// TODO: Store metrics where they belong
func (m *Metric) Receive(c *actor.Context) {
	switch msg := c.Message().(type) {
	case actor.Started:
	case actor.Stopped:
	case types.RuntimeMetric:
		_ = msg
	}
}
