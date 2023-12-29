package runtime

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/stealthrocket/wasi-go"
	"github.com/stealthrocket/wasi-go/imports"
	"github.com/stealthrocket/wasi-go/imports/wasi_http"
	"github.com/tetratelabs/wazero"
	wapi "github.com/tetratelabs/wazero/api"
	"github.com/vmihailenco/msgpack/v5"
)

type request struct {
	Body   []byte
	Method string
	URL    string
}

type RequestPlugin interface {
	Instanciate(context.Context, wazero.Runtime) error
	WriteResponse(io.Writer) (int, error)
	Close(context.Context) error
}

type RequestModule struct {
	requestBytes  []byte
	responseBytes []byte
}

// TODO: could probably do more optimized stuff for larger bodies.
// We want to limit the body size though...
func NewRequestModule(r *http.Request) (*RequestModule, error) {
	if r == nil {
		return nil, fmt.Errorf("http.request is nil")
	}
	b, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	req := request{
		Body:   b,
		Method: r.Method,
		URL:    r.URL.RequestURI(),
	}

	b, err = msgpack.Marshal(req)
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

func Compile(ctx context.Context, cache wazero.CompilationCache, blob []byte) error {
	config := wazero.NewRuntimeConfig().WithCompilationCache(cache)
	runtime := wazero.NewRuntimeWithConfig(ctx, config)
	defer runtime.Close(ctx)

	_, err := runtime.CompileModule(ctx, blob)
	if err != nil {
		return err
	}
	return nil
}

// TODO: add some kind of log capture...
type Args struct {
	Blob          []byte
	Cache         wazero.CompilationCache
	RequestPlugin RequestPlugin
	Env           map[string]string
}

func Run(ctx context.Context, args Args) error {
	config := wazero.NewRuntimeConfig().WithCompilationCache(args.Cache)
	runtime := wazero.NewRuntimeWithConfig(ctx, config)
	defer runtime.Close(ctx)

	if err := args.RequestPlugin.Instanciate(ctx, runtime); err != nil {
		return err
	}

	wasmModule, err := runtime.CompileModule(ctx, args.Blob)
	if err != nil {
		return err
	}
	// TODO: Can't close cause it will invalidate the cache.
	// defer wasmModule.Close(ctx)

	// TODO: Open with append mode or configure it like that ok!
	// f, err := os.Create("foo")
	// if err != nil {
	// 	return err
	// }
	//f.Seek(0, io.SeekStart)
	// fd := int(f.Fd())
	fd := -1

	builder := imports.NewBuilder().
		WithName("ffaas").
		WithArgs().
		WithStdio(fd, fd, fd).
		// TODO: env...
		WithEnv().
		// TODO: we want to mount this to some virtual folder?
		WithDirs("/").
		WithListens().
		WithDials().
		WithNonBlockingStdio(false).
		WithSocketsExtension("auto", wasmModule).
		WithMaxOpenFiles(10).
		WithMaxOpenDirs(10)

	var system wasi.System
	ctx, system, err = builder.Instantiate(ctx, runtime)
	if err != nil {
		return err
	}
	defer system.Close(ctx)

	wasiHTTP := wasi_http.MakeWasiHTTP()
	if err := wasiHTTP.Instantiate(ctx, runtime); err != nil {
		return err
	}

	_, err = runtime.InstantiateModule(ctx, wasmModule, wazero.NewModuleConfig())

	return err
}
