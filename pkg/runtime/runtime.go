package runtime

import (
	"context"
	"io"
	"net/http"
	"os"

	"github.com/tetratelabs/wazero"
	wapi "github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
	"github.com/vmihailenco/msgpack/v5"
)

func moduleMalloc(size uint32) wapi.GoModuleFunc {
	return func(ctx context.Context, module wapi.Module, stack []uint64) {
		stack[0] = uint64(wapi.DecodeU32(uint64(size)))
	}
}

func moduleWriteRequest(b []byte) wapi.GoModuleFunc {
	return func(ctx context.Context, module wapi.Module, stack []uint64) {
		offset := wapi.DecodeU32(stack[0])
		module.Memory().Write(offset, b)
	}
}

func moduleWriteResponse(w io.Writer) wapi.GoModuleFunc {
	return func(ctx context.Context, module wapi.Module, stack []uint64) {
		offset := wapi.DecodeU32(stack[0])
		size := wapi.DecodeU32(stack[1])
		resp, _ := module.Memory().Read(offset, size)
		w.Write(resp)
	}
}

type request struct {
	Body   []byte
	Method string
	URL    string
}

type Runtime struct {
	wasmBlob    []byte
	compiledMod wazero.CompiledModule
	module      wapi.Module
	wruntime    wazero.Runtime
}

func New(blob []byte, env map[string]string) (*Runtime, error) {
	var (
		ctx     = context.Background()
		config  = wazero.NewRuntimeConfig().WithDebugInfoEnabled(true)
		runtime = wazero.NewRuntimeWithConfig(ctx, config)
	)

	wasi_snapshot_preview1.MustInstantiate(ctx, runtime)

	compiledMod, err := runtime.CompileModule(ctx, blob)
	if err != nil {
		return nil, err
	}

	return &Runtime{
		wasmBlob:    blob,
		compiledMod: compiledMod,
		wruntime:    runtime,
	}, nil
}

func (runtime *Runtime) HandleHTTP(w http.ResponseWriter, r *http.Request) error {
	ctx := context.Background()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	// TODO: maybe close the body
	req := request{
		Method: r.Method,
		URL:    r.URL.Path,
		Body:   body,
	}

	b, err := msgpack.Marshal(req)
	if err != nil {
		return err
	}

	rsize := uint32(len(b))
	_, err = runtime.wruntime.NewHostModuleBuilder("env").
		NewFunctionBuilder().
		WithGoModuleFunction(moduleMalloc(rsize), []wapi.ValueType{}, []wapi.ValueType{wapi.ValueTypeI32}).
		Export("malloc").
		NewFunctionBuilder().
		WithGoModuleFunction(moduleWriteRequest(b), []wapi.ValueType{wapi.ValueTypeI32}, []wapi.ValueType{}).
		Export("write_request").
		NewFunctionBuilder().
		WithGoModuleFunction(moduleWriteResponse(w), []wapi.ValueType{wapi.ValueTypeI32, wapi.ValueTypeI32}, []wapi.ValueType{}).
		Export("write_response").
		Instantiate(ctx)
	if err != nil {
		return err
	}

	modConfig := wazero.NewModuleConfig().
		WithStdout(os.Stdout).
		WithStartFunctions()
	mod, err := runtime.wruntime.InstantiateModule(ctx, runtime.compiledMod, modConfig)
	if err != nil {
		return err
	}

	mod.ExportedFunction("_start").Call(ctx)

	return err
}

func (runtime *Runtime) Close(ctx context.Context) error {
	return runtime.wruntime.Close(ctx)
}
