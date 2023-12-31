package actrs

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/anthdm/ffaas/pkg/storage"
	"github.com/anthdm/ffaas/proto"
	"github.com/anthdm/hollywood/actor"
	"github.com/google/uuid"
	"github.com/stealthrocket/wasi-go"
	"github.com/stealthrocket/wasi-go/imports"
	"github.com/tetratelabs/wazero"
	wapi "github.com/tetratelabs/wazero/api"
	"github.com/vmihailenco/msgpack/v5"
)

const KindRuntime = "runtime"

// Runtime is an actor that can execute compiled WASM blobs in a distributed cluster.
type Runtime struct {
	store storage.Store
}

func NewRuntime(store storage.Store) actor.Producer {
	return func() actor.Receiver {
		return &Runtime{
			store: store,
		}
	}
}

func (r *Runtime) Receive(c *actor.Context) {
	switch msg := c.Message().(type) {
	case actor.Started:
	case actor.Stopped:
	case *proto.HTTPRequest:
		endpoint, err := r.store.GetEndpoint(uuid.MustParse(msg.EndpointID))
		if err != nil {
			slog.Warn("runtime could not find endpoint from store", "err", err)
			return
		}
		deploy, err := r.store.GetDeploy(endpoint.ActiveDeployID)
		if err != nil {
			slog.Warn("runtime could not find the endpoint's active deploy from store", "err", err)
			return
		}
		cache := wazero.NewCompilationCache()
		httpmod, _ := NewRequestModule(msg)
		r.exec(context.TODO(), deploy.Blob, cache, endpoint.Environment, httpmod)
		resp := &proto.HTTPResponse{
			Response:   httpmod.responseBytes,
			RequestID:  msg.ID,
			StatusCode: http.StatusOK,
		}
		c.Respond(resp)
		c.Engine().Poison(c.PID())
	}
}

func (r *Runtime) exec(ctx context.Context, blob []byte, cache wazero.CompilationCache, env map[string]string, httpmod *RequestModule) {
	config := wazero.NewRuntimeConfig().WithCompilationCache(cache)
	runtime := wazero.NewRuntimeWithConfig(ctx, config)
	defer runtime.Close(ctx)

	mod, err := runtime.CompileModule(ctx, blob)
	if err != nil {
		slog.Warn("compiling module failed", "err", err)
		return
	}
	fd := -1
	builder := imports.NewBuilder().
		WithName("ffaas").
		WithArgs().
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
		slog.Warn("failed to instanciate wasi module", "err", err)
		return
	}
	defer system.Close(ctx)

	httpmod.Instanciate(ctx, runtime)

	_, err = runtime.InstantiateModule(ctx, mod, wazero.NewModuleConfig())
	if err != nil {
		slog.Warn("failed to instanciate guest module", "err", err)
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
	b, err := msgpack.Marshal(req)
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

func (r *RequestModule) Instanciate(ctx context.Context, runtime wazero.Runtime) error {
	_, err := runtime.NewHostModuleBuilder("env").
		NewFunctionBuilder().
		WithGoModuleFunction(r.moduleMalloc(), []wapi.ValueType{}, []wapi.ValueType{wapi.ValueTypeI32}).
		Export("malloc").
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

func (r *RequestModule) moduleMalloc() wapi.GoModuleFunc {
	return func(ctx context.Context, module wapi.Module, stack []uint64) {
		size := uint64(len(r.requestBytes))
		stack[0] = uint64(wapi.DecodeU32(size))
	}
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
