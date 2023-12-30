package act

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/anthdm/ffaas/pkg/runtime"
	"github.com/anthdm/hollywood/actor"
)

const KindRuntime = "runtime"

type Runtime struct {
	args    runtime.Args
	w       http.ResponseWriter
	started time.Time
	uptime  time.Duration
}

func NewRuntime(w http.ResponseWriter, args runtime.Args) actor.Producer {
	return func() actor.Receiver {
		return &Runtime{
			args: args,
			w:    w,
		}
	}
}

func (r *Runtime) Receive(c *actor.Context) {
	switch c.Message().(type) {
	case actor.Started:
		r.start(c)
		c.Engine().Poison(c.PID())
	case actor.Stopped:
		r.uptime = time.Since(r.started)
		fmt.Println("stopped uptime", r.uptime)
	}
}

func (r *Runtime) start(c *actor.Context) {
	r.started = time.Now()
	if err := runtime.Run(context.Background(), r.args); err != nil {
		r.w.WriteHeader(http.StatusInternalServerError)
		r.w.Write([]byte(err.Error()))
		return
	}
	if _, err := r.args.RequestPlugin.WriteResponse(r.w); err != nil {
		r.w.WriteHeader(http.StatusInternalServerError)
		r.w.Write([]byte(err.Error()))
		return
	}
}
