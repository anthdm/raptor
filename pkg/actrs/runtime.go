package actrs

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/anthdm/hollywood/actor"
	"github.com/anthdm/run/pkg/spidermonkey"
	"github.com/anthdm/run/pkg/storage"
	"github.com/anthdm/run/pkg/types"
	"github.com/anthdm/run/proto"
	"github.com/google/uuid"
	"github.com/stealthrocket/wasi-go"
	"github.com/stealthrocket/wasi-go/imports"
	"github.com/tetratelabs/wazero"
	wapi "github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"

	prot "google.golang.org/protobuf/proto"
)

const KindRuntime = "runtime"

// Runtime is an actor that can execute compiled WASM blobs in a distributed cluster.
type Runtime struct {
	store       storage.Store
	metricStore storage.MetricStore
	cache       storage.ModCacher
	started     time.Time
	endpointID  uuid.UUID
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
		r.endpointID = uuid.MustParse(msg.ActiveDeployID)
		deploy, err := r.store.GetDeploy(r.endpointID)
		if err != nil {
			// NOTE: Make sure we always respond to the message with an HTTPResponse or the
			// the request will never be completed.
			slog.Warn("runtime could not find the endpoint's active deploy from store", "err", err)
			c.Respond(&proto.HTTPResponse{
				Response:   []byte("internal server error"),
				StatusCode: http.StatusInternalServerError,
				RequestID:  msg.ID,
			})
			return
		}
		switch msg.Runtime {
		case "js":
			buffer := &bytes.Buffer{}
			r.invokeJSRuntime(context.TODO(), deploy.Blob, buffer, msg.Env)
			c.Respond(&proto.HTTPResponse{
				Response:   buffer.Bytes(),
				StatusCode: http.StatusOK,
				RequestID:  msg.ID,
			})
			buffer = nil
		case "go":
			httpmod, _ := NewRequestModule(msg)
			modcache, ok := r.cache.Get(deploy.EndpointID)
			if !ok {
				modcache = wazero.NewCompilationCache()
				slog.Warn("no cache hit", "endpoint", deploy.EndpointID)
			}
			r.invokeGORuntime(context.TODO(), deploy.Blob, modcache, msg.Env, httpmod)
			resp := &proto.HTTPResponse{
				Response:   httpmod.responseBytes,
				RequestID:  msg.ID,
				StatusCode: http.StatusOK,
			}
			c.Respond(resp)
		default:
			slog.Warn("invalid runtime", "runtime", msg.Runtime)
			c.Respond(&proto.HTTPResponse{
				Response:   []byte("internal server error"),
				StatusCode: http.StatusInternalServerError,
				RequestID:  msg.ID,
			})
		}

		c.Engine().Poison(c.PID())
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

func (r *Runtime) invokeJSRuntime(ctx context.Context, blob []byte, buffer io.Writer, env map[string]string) {
	modcache, ok := r.cache.Get(r.endpointID)
	if !ok {
		modcache = wazero.NewCompilationCache()
		slog.Warn("no cache hit", "endpoint", r.endpointID)
		r.cache.Put(r.endpointID, modcache)
	}
	config := wazero.NewRuntimeConfig().WithCompilationCache(modcache)
	runtime := wazero.NewRuntimeWithConfig(ctx, config)
	defer runtime.Close(ctx)

	mod, err := runtime.CompileModule(ctx, spidermonkey.WasmBlob)
	if err != nil {
		panic(err)
	}

	wasi_snapshot_preview1.MustInstantiate(ctx, runtime)
	modConfig := wazero.NewModuleConfig().
		WithStdin(os.Stdin).
		WithStdout(buffer).
		WithArgs("", "-e", string(blob))
	_, err = runtime.InstantiateModule(ctx, mod, modConfig)
	if err != nil {
		panic(err)
	}
}

func (r *Runtime) invokeGORuntime(ctx context.Context,
	blob []byte,
	cache wazero.CompilationCache,
	env map[string]string,
	httpmod *RequestModule) {
	config := wazero.NewRuntimeConfig().WithCompilationCache(cache)
	runtime := wazero.NewRuntimeWithConfig(ctx, config)
	defer runtime.Close(ctx)

	mod, err := runtime.CompileModule(ctx, blob)
	if err != nil {
		slog.Warn("compiling module failed", "err", err)
		return
	}
	fd := -1 // TODO: for capturing logs..
	requestLen := strconv.Itoa(len(httpmod.requestBytes))
	builder := imports.NewBuilder().
		WithName("run").
		WithArgs(requestLen).
		WithStdio(fd, fd, fd).
		WithEnv(envMapToSlice(env)...).
		// TODO: we want to mount this to some virtual folder?
		WithDirs("/").
		WithListens().
		WithDials().
		WithNonBlockingStdio(false).
		WithSocketsExtension("auto", mod).
		WithMaxOpenFiles(10).
		WithMaxOpenDirs(10)

	var system wasi.System
	ctx, system, err = builder.Instantiate(ctx, runtime)
	if err != nil {
		slog.Warn("failed to instantiate wasi module", "err", err)
		return
	}
	defer system.Close(ctx)

	httpmod.Instantiate(ctx, runtime)

	_, err = runtime.InstantiateModule(ctx, mod, wazero.NewModuleConfig())
	if err != nil {
		slog.Warn("failed to instantiate guest module", "err", err)
	}
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
