package actrs

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/anthdm/hollywood/actor"
	"github.com/anthdm/run/pkg/runtime"
	"github.com/anthdm/run/pkg/storage"
	"github.com/anthdm/run/pkg/types"
	"github.com/anthdm/run/pkg/util"
	"github.com/anthdm/run/proto"
	"github.com/bananabytelabs/wazero"
	wapi "github.com/bananabytelabs/wazero/api"
	"github.com/google/uuid"

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
	start := time.Now()

	r.deployID = uuid.MustParse(msg.ActiveDeployID)
	deploy, err := r.store.GetDeploy(r.deployID)
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
		Blob:  deploy.Blob,
		Env:   msg.Env,
		In:    in,
		Out:   out,
		Cache: modCache,
	}

	switch msg.Runtime {
	case "go":
		err = runtime.Invoke(context.Background(), args)
	case "js":
	default:
		err = fmt.Errorf("invalid runtime: %s", msg.Runtime)
	}
	if err != nil {
		respondError(ctx, http.StatusInternalServerError, "internal server error", msg.ID)
		return
	}

	res, status, err := util.ParseRuntimeHTTPResponse(out.String())
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

	fmt.Println("runtine handle HTTP took: ", time.Since(start))

	r.cache.Put(deploy.EndpointID, modCache)

	ctx.Engine().Poison(ctx.PID())
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

func (r *Runtime) invokeJSRuntime(ctx context.Context, blob []byte, buffer io.Writer, env map[string]string) {
	// modcache, ok := r.cache.Get(r.deployID)
	// if !ok {
	// 	modcache = wazero.NewCompilationCache()
	// 	slog.Warn("no cache hit", "endpoint", r.endpointID)
	// 	r.cache.Put(r.endpointID, modcache)
	// }
	// config := wazero.NewRuntimeConfig().WithCompilationCache(modcache)
	// runtime := wazero.NewRuntimeWithConfig(ctx, config)
	// defer runtime.Close(ctx)

	// mod, err := runtime.CompileModule(ctx, spidermonkey.WasmBlob)
	// if err != nil {
	// 	panic(err)
	// }

	// wasi_snapshot_preview1.MustInstantiate(ctx, runtime)
	// modConfig := wazero.NewModuleConfig().
	// 	WithStdin(os.Stdin).
	// 	WithStdout(buffer).
	// 	WithArgs("", "-e", string(blob))
	// _, err = runtime.InstantiateModule(ctx, mod, modConfig)
	// if err != nil {
	// 	panic(err)
	// }
}

func envMapToSlice(env map[string]string) []string {
	slice := make([]string, len(env))
	i := 0
	for k, v := range env {
		s := fmt.Sprintf("%s=%s", k, v)
		slice[i] = s
		i++
	}
	return slice
}

type RequestModule struct {
	requestBytes  []byte
	responseBytes []byte
}

func NewRequestModule(req *proto.HTTPRequest) (*RequestModule, error) {
	b, err := prot.Marshal(req)
	if err != nil {
		return nil, err
	}

	return &RequestModule{
		requestBytes: b,
	}, nil
}

func (r *RequestModule) WriteResponse(w io.Writer) (int, error) {
	return w.Write(r.responseBytes)
}

func (r *RequestModule) Instantiate(ctx context.Context, runtime wazero.Runtime) error {
	_, err := runtime.NewHostModuleBuilder("env").
		NewFunctionBuilder().
		WithGoModuleFunction(r.moduleWriteRequest(), []wapi.ValueType{wapi.ValueTypeI32}, []wapi.ValueType{}).
		Export("write_request").
		NewFunctionBuilder().
		WithGoModuleFunction(r.moduleWriteResponse(), []wapi.ValueType{wapi.ValueTypeI32, wapi.ValueTypeI32}, []wapi.ValueType{}).
		Export("write_response").
		Instantiate(ctx)
	return err
}

func (r *RequestModule) Close(ctx context.Context) error {
	r.responseBytes = nil
	r.requestBytes = nil
	return nil
}

func (r *RequestModule) moduleWriteRequest() wapi.GoModuleFunc {
	return func(ctx context.Context, module wapi.Module, stack []uint64) {
		offset := wapi.DecodeU32(stack[0])
		module.Memory().Write(offset, r.requestBytes)
	}
}

func (r *RequestModule) moduleWriteResponse() wapi.GoModuleFunc {
	return func(ctx context.Context, module wapi.Module, stack []uint64) {
		offset := wapi.DecodeU32(stack[0])
		size := wapi.DecodeU32(stack[1])
		resp, _ := module.Memory().Read(offset, size)
		r.responseBytes = resp
	}
}

func respondError(ctx *actor.Context, code int32, msg string, id string) {
	ctx.Respond(&proto.HTTPResponse{
		Response:   []byte(msg),
		StatusCode: code,
		RequestID:  id,
	})
}
