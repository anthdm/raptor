package actrs

import (
	"github.com/anthdm/hollywood/actor"
	"github.com/anthdm/hollywood/cluster"
)

const KindRuntimeManager = "runtime_manager"

type (
	requestRuntime struct {
		key string
	}
	addRuntime struct {
		key string
		pid *actor.PID
	}
	removeRuntime struct {
		key string
	}
)

type RuntimeManager struct {
	runtimes map[string]*actor.PID
	cluster  *cluster.Cluster
}

func NewRuntimeManager(c *cluster.Cluster) actor.Producer {
	return func() actor.Receiver {
		return &RuntimeManager{
			runtimes: make(map[string]*actor.PID),
			cluster:  c,
		}
	}
}

func (rm *RuntimeManager) Receive(c *actor.Context) {
	switch msg := c.Message().(type) {
	case requestRuntime:
		pid := rm.runtimes[msg.key]
		if pid == nil {
			pid = rm.cluster.Activate(KindRuntime, cluster.NewActivationConfig())
			rm.runtimes[msg.key] = pid
		}
		c.Respond(pid)
	case addRuntime:
		rm.runtimes[msg.key] = msg.pid
	case removeRuntime:
		delete(rm.runtimes, msg.key)
	case actor.Started:
	case actor.Stopped:
	case actor.Initialized:
	}
}
