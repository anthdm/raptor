package runtime

import (
	"context"
	"io"
	"net/http"
	"os"

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

type Runtime struct {
	wazero.Runtime

	compiledMod   wazero.CompiledModule
	requestBytes  []byte
	responseBytes []byte
}

func New(cache wazero.CompilationCache, blob []byte) (*Runtime, error) {
	ctx := context.Background()
	config := wazero.NewRuntimeConfig().WithCompilationCache(cache)
	runtime := wazero.NewRuntimeWithConfig(ctx, config)

	compiledMod, err := runtime.CompileModule(ctx, blob)
	if err != nil {
		return nil, err
	}
	builder := imports.NewBuilder().
		WithName("ffaas").
		WithArgs().
		WithEnv().
		WithDirs("/").
		WithDials().
		WithNonBlockingStdio(false).
		WithSocketsExtension("auto", compiledMod).
		//WithTracer(false, os.Stderr, wasi.WithTracerStringSize(tracerStringSize)).
		WithMaxOpenFiles(1024).
		WithMaxOpenDirs(1024)

	_, _, err = builder.Instantiate(ctx, runtime)
	if err != nil {
		return nil, err
	}
	wasiHTTP := wasi_http.MakeWasiHTTP()
	if err := wasiHTTP.Instantiate(ctx, runtime); err != nil {
		return nil, err
	}

	r := &Runtime{
		Runtime:     runtime,
		compiledMod: compiledMod,
	}
	if err := r.initHostModule(context.Background()); err != nil {
		return nil, err
	}
	return r, err
}

func (r *Runtime) moduleMalloc() wapi.GoModuleFunc {
	return func(ctx context.Context, module wapi.Module, stack []uint64) {
		size := uint64(len(r.requestBytes))
		stack[0] = uint64(wapi.DecodeU32(size))
	}
}

func (r *Runtime) moduleWriteRequest() wapi.GoModuleFunc {
	return func(ctx context.Context, module wapi.Module, stack []uint64) {
		offset := wapi.DecodeU32(stack[0])
		module.Memory().Write(offset, r.requestBytes)
	}
}

func (r *Runtime) moduleWriteResponse() wapi.GoModuleFunc {
	return func(ctx context.Context, module wapi.Module, stack []uint64) {
		offset := wapi.DecodeU32(stack[0])
		size := wapi.DecodeU32(stack[1])
		resp, _ := module.Memory().Read(offset, size)
		r.responseBytes = resp
	}
}

func (r *Runtime) initHostModule(ctx context.Context) error {
	_, err := r.NewHostModuleBuilder("env").
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

func (r *Runtime) Exec(ctx context.Context, req *http.Request) error {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return err
	}
	defer req.Body.Close()
	b, err := msgpack.Marshal(request{
		Method: req.Method,
		URL:    req.URL.Path,
		Body:   body,
	})
	if err != nil {
		return err
	}
	r.requestBytes = b

	modConfig := wazero.NewModuleConfig().
		WithStdout(os.Stdout)
	_, err = r.InstantiateModule(ctx, r.compiledMod, modConfig)
	if err != nil {
		return err
	}
	// _, err = mod.ExportedFunction("_start").Call(ctx)
	// if !strings.Contains(err.Error(), "closed with exit_code(0)") {
	// 	return err
	// }
	return nil
}

func (r *Runtime) Response() []byte {
	b := make([]byte, len(r.responseBytes))
	copy(b, r.responseBytes)
	r.responseBytes = nil
	return b
}

func (runtime *Runtime) Close(ctx context.Context) error {
	runtime.requestBytes = nil
	runtime.responseBytes = nil
	return runtime.Runtime.Close(ctx)
}
