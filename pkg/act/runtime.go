package act

import (
	"context"
	"net/http"
	"time"

	"github.com/anthdm/ffaas/pkg/runtime"
	"github.com/anthdm/ffaas/pkg/storage"
	"github.com/anthdm/ffaas/pkg/types"
	"github.com/anthdm/hollywood/actor"
	"github.com/google/uuid"
)

const KindRuntime = "runtime"

type Runtime struct {
	terminateFn func()
	args        runtime.Args
	w           http.ResponseWriter
	endpointID  uuid.UUID
	deployID    uuid.UUID
	store       storage.MetricStore
	path        string
	started     time.Time
	cancel      context.CancelFunc
}

func NewRuntime(w http.ResponseWriter,
	args runtime.Args,
	endpointID uuid.UUID,
	deployID uuid.UUID,
	store storage.MetricStore,
	path string,
	cancel context.CancelFunc) actor.Producer {
	return func() actor.Receiver {
		return &Runtime{
			args:       args,
			w:          w,
			path:       path,
			store:      store,
			deployID:   deployID,
			endpointID: endpointID,
			cancel:     cancel,
		}
	}
}

func (r *Runtime) Receive(c *actor.Context) {
	switch c.Message().(type) {
	case actor.Started:
		r.terminateFn = func() {
			c.Engine().Poison(c.PID())
		}
		r.exec()
	case actor.Stopped:
		r.cancel()
		metric := types.RuntimeMetric{
			ID:         uuid.New(),
			StartTime:  r.started,
			Duration:   time.Since(r.started),
			EndpointID: r.endpointID,
			DeployID:   r.deployID,
			RequestURL: r.path,
		}
		r.store.CreateRuntimeMetric(&metric)
	}
}

func (r *Runtime) exec() {
	r.started = time.Now()
	go func() {
		defer r.terminateFn()
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
	}()
}
