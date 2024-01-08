package actrs

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/anthdm/hollywood/actor"
	"github.com/anthdm/raptor/pkg/runtime"
	"github.com/anthdm/raptor/pkg/shared"
	"github.com/anthdm/raptor/pkg/spidermonkey"
	"github.com/anthdm/raptor/pkg/storage"
	"github.com/anthdm/raptor/pkg/types"
	"github.com/anthdm/raptor/proto"
	"github.com/google/uuid"
	"github.com/tetratelabs/wazero"

	prot "google.golang.org/protobuf/proto"
)

const KindRuntime = "runtime"

// Runtime is an actor that can execute compiled WASM blobs in a distributed cluster.
type Runtime struct {
	store       storage.Store
	metricStore storage.MetricStore
	cache       storage.ModCacher
	started     time.Time
	deployID    uuid.UUID
}

func NewRuntime(store storage.Store, metricStore storage.MetricStore, cache storage.ModCacher) actor.Producer {
	return func() actor.Receiver {
		return &Runtime{
			store:       store,
			metricStore: metricStore,
			cache:       cache,
		}
	}
}

func (r *Runtime) Receive(c *actor.Context) {
	switch msg := c.Message().(type) {
	case actor.Started:
		r.started = time.Now()
	case actor.Stopped:
	case *proto.HTTPRequest:
		// Handle the HTTP request that is forwarded from the WASM server actor.
		r.handleHTTPRequest(c, msg)
	}
}

func (r *Runtime) handleHTTPRequest(ctx *actor.Context, msg *proto.HTTPRequest) {
	r.deployID = uuid.MustParse(msg.DeploymentID)
	deploy, err := r.store.GetDeployment(r.deployID)
	if err != nil {
		slog.Warn("runtime could not find the endpoint's active deploy from store", "err", err)
		respondError(ctx, http.StatusInternalServerError, "internal server error", msg.ID)
		return
	}

	modCache, ok := r.cache.Get(deploy.EndpointID)
	if !ok {
		modCache = wazero.NewCompilationCache()
		slog.Warn("no cache hit", "endpoint", deploy.EndpointID)
	}

	b, err := prot.Marshal(msg)
	if err != nil {
		slog.Warn("failed to marshal incoming HTTP request", "err", err)
		respondError(ctx, http.StatusInternalServerError, "internal server error", msg.ID)
		return
	}

	in := bytes.NewReader(b)
	out := &bytes.Buffer{}
	args := runtime.InvokeArgs{
		Env:   msg.Env,
		In:    in,
		Out:   out,
		Cache: modCache,
	}

	switch msg.Runtime {
	case "go":
		args.Blob = deploy.Blob
	case "js":
		args.Blob = spidermonkey.WasmBlob
		args.Args = []string{"", "-e", string(deploy.Blob)}
	default:
		err = fmt.Errorf("invalid runtime: %s", msg.Runtime)
	}

	err = runtime.Invoke(context.Background(), args)
	if err != nil {
		slog.Error("runtime invoke error", "err", err)
		respondError(ctx, http.StatusInternalServerError, "internal server error", msg.ID)
		return
	}

	res, status, err := shared.ParseRuntimeHTTPResponse(out.String())
	if err != nil {
		respondError(ctx, http.StatusInternalServerError, "internal server error", msg.ID)
		return
	}
	resp := &proto.HTTPResponse{
		Response:   []byte(res),
		RequestID:  msg.ID,
		StatusCode: int32(status),
	}

	ctx.Respond(resp)

	r.cache.Put(deploy.EndpointID, modCache)

	ctx.Engine().Poison(ctx.PID())

	// only store metrics when its a request on LIVE
	if !msg.Preview {
		metric := types.RuntimeMetric{
			ID:         uuid.New(),
			StartTime:  r.started,
			Duration:   time.Since(r.started),
			DeployID:   deploy.ID,
			EndpointID: deploy.EndpointID,
			RequestURL: msg.URL,
		}
		if err := r.metricStore.CreateRuntimeMetric(&metric); err != nil {
			slog.Warn("failed to create runtime metric", "err", err)
		}
	}
}

func respondError(ctx *actor.Context, code int32, msg string, id string) {
	ctx.Respond(&proto.HTTPResponse{
		Response:   []byte(msg),
		StatusCode: code,
		RequestID:  id,
	})
}
