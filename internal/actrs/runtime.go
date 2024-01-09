package actrs

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/anthdm/hollywood/actor"
	"github.com/anthdm/raptor/internal/shared"
	"github.com/anthdm/raptor/internal/spidermonkey"
	"github.com/anthdm/raptor/internal/storage"
	"github.com/anthdm/raptor/internal/types"
	"github.com/anthdm/raptor/proto"
	"github.com/google/uuid"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"

	prot "google.golang.org/protobuf/proto"
)

const KindRuntime = "runtime"

var (
	runtimeKeepAlive = time.Second
)

type shutdown struct{}

// Runtime is an actor that can execute compiled WASM blobs in a distributed cluster.
type Runtime struct {
	store        storage.Store
	cache        storage.ModCacher
	started      time.Time
	deploymentID uuid.UUID

	managerPID *actor.PID
	runtime    wazero.Runtime
	mod        wazero.CompiledModule
	blob       []byte
	repeat     actor.SendRepeater
}

func NewRuntime(store storage.Store, cache storage.ModCacher) actor.Producer {
	return func() actor.Receiver {
		return &Runtime{
			store: store,
			cache: cache,
		}
	}
}

func (r *Runtime) Receive(c *actor.Context) {
	switch msg := c.Message().(type) {
	case actor.Started:
		r.repeat = c.SendRepeat(c.PID(), shutdown{}, runtimeKeepAlive)
		r.started = time.Now()
		r.managerPID = c.Engine().Registry.GetPID(KindRuntimeManager, "1")
	case actor.Stopped:
		c.Send(r.managerPID, removeRuntime{key: r.deploymentID.String()})
		r.runtime.Close(context.Background())
		// Releasing this mod will invalidate the cache for some reason.
		// r.mod.Close(context.TODO())
	case *proto.HTTPRequest:
		// Refresh the keepAlive timer
		r.repeat = c.SendRepeat(c.PID(), shutdown{}, runtimeKeepAlive)
		if r.runtime == nil {
			r.initialize(msg)
		}
		// Handle the HTTP request that is forwarded from the WASM server actor.
		r.handleHTTPRequest(c, msg)

	case shutdown:
		c.Engine().Poison(c.PID())
	}
}

func (r *Runtime) initialize(msg *proto.HTTPRequest) error {
	ctx := context.Background()
	r.deploymentID = uuid.MustParse(msg.DeploymentID)

	// TODO: this could be coming from a Redis cache instead of Postres.
	// Maybe only the blob. Not sure...
	deploy, err := r.store.GetDeployment(r.deploymentID)
	if err != nil {
		slog.Warn("runtime could not find deploy from store", "err", err, "id", r.deploymentID)
		return fmt.Errorf("runtime: could not find deployment (%s)", r.deploymentID)
	}
	r.blob = deploy.Blob // can be optimized

	modCache, ok := r.cache.Get(r.deploymentID)
	if !ok {
		slog.Warn("no cache hit", "endpoint", r.deploymentID)
		modCache = wazero.NewCompilationCache()
	}

	config := wazero.NewRuntimeConfigCompiler().WithCompilationCache(modCache)
	r.runtime = wazero.NewRuntimeWithConfig(ctx, config)
	wasi_snapshot_preview1.MustInstantiate(ctx, r.runtime)

	var blob []byte
	if msg.Runtime == "js" {
		blob = spidermonkey.WasmBlob
	} else if msg.Runtime == "go" {
		blob = deploy.Blob
	}

	mod, err := r.runtime.CompileModule(ctx, blob)
	if err != nil {
		return fmt.Errorf("failed to compile module: %s", err)
	}

	r.cache.Put(deploy.ID, modCache)

	r.mod = mod
	return nil
}

func (r *Runtime) handleHTTPRequest(ctx *actor.Context, msg *proto.HTTPRequest) {
	b, err := prot.Marshal(msg)
	if err != nil {
		slog.Warn("failed to marshal incoming HTTP request", "err", err)
		respondError(ctx, http.StatusInternalServerError, "internal server error", msg.ID)
		return
	}

	in := bytes.NewReader(b)
	out := &bytes.Buffer{} // TODO: pool this bad boy

	args := []string{}
	if msg.Runtime == "js" {
		args = []string{"", "-e", string(r.blob)}
	}

	modConf := wazero.NewModuleConfig().
		WithStdin(in).
		WithStdout(out).
		WithStderr(os.Stderr).
		WithArgs(args...)
	for k, v := range msg.Env {
		modConf = modConf.WithEnv(k, v)
	}
	_, err = r.runtime.InstantiateModule(context.Background(), r.mod, modConf)
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

	// only send metrics when its a request on LIVE
	if !msg.Preview {
		metric := types.RuntimeMetric{
			ID:           uuid.New(),
			StartTime:    r.started,
			Duration:     time.Since(r.started),
			DeploymentID: r.deploymentID,
			// EndpointID:   deploy.EndpointID,
			RequestURL: msg.URL,
			StatusCode: status,
		}
		pid := ctx.Engine().Registry.GetPID(KindMetric, "1")
		ctx.Send(pid, metric)
	}
}

func respondError(ctx *actor.Context, code int32, msg string, id string) {
	ctx.Respond(&proto.HTTPResponse{
		Response:   []byte(msg),
		StatusCode: code,
		RequestID:  id,
	})
}
