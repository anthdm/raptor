package actrs

import (
	"github.com/anthdm/hollywood/actor"
	"github.com/anthdm/raptor/internal/types"
)

const KindRuntimeLog = "runtime_log"

type RuntimeLog struct{}

func NewRuntimeLog() actor.Receiver {
	return &RuntimeLog{}
}

func (rl *RuntimeLog) Receive(c *actor.Context) {
	switch msg := c.Message().(type) {
	case actor.Started:
	case actor.Stopped:
	case types.RuntimeLogEvent:
		_ = msg
	}
}
