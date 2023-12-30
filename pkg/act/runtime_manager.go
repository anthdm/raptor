package act

import (
	"fmt"

	"github.com/anthdm/hollywood/actor"
)

const (
	KindRuntimeManager = "runtime_manager"
	RuntimeManagerID   = "rtm"
)

type RuntimeStart struct {
	pid *actor.PID
}

type RuntimeManager struct {
	runtimes map[string]*actor.PID
}

func NewRuntimeManager() actor.Producer {
	return func() actor.Receiver {
		return &RuntimeManager{
			runtimes: make(map[string]*actor.PID),
		}
	}
}

func (rm *RuntimeManager) Receive(c *actor.Context) {
	switch msg := c.Message().(type) {
	case RuntimeStart:
		fmt.Println("need to start", msg)
		// pid := s.engine.Spawn(act.NewRuntime(w, args, endpointID, deploy.ID, s.metricStore, requestURL), act.KindRuntime)
	case actor.Started:
	case actor.Stopped:

	}
}
