package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

// WithGoModuleFunction(fn api.GoModuleFunction, params, results []api.ValueType) HostFunctionBuilder

type ResponseWriter struct {
	buffer     bytes.Buffer
	statusCode int
}

func (w *ResponseWriter) Header() http.Header {
	return http.Header{}
}

func (w *ResponseWriter) Write(b []byte) (n int, err error) {
	return w.buffer.Write(b)
}

func (w *ResponseWriter) WriteHeader(status int) {
	w.statusCode = status
}

func foo(w http.ResponseWriter, r *http.Request) {

}

func bar(ctx context.Context, mod api.Module, stack []uint64) {

}

func main() {
	b, err := os.ReadFile("examples/go/app.wasm")
	if err != nil {
		log.Fatal(err)
	}
	var (
		ctx     = context.Background()
		config  = wazero.NewRuntimeConfig().WithDebugInfoEnabled(true)
		runtime = wazero.NewRuntimeWithConfig(ctx, config)
	)

	// resp := &ResponseWriter{}
	// req, _ := http.NewRequest("GET", "/", nil)
	bodySize := 69

	wasi_snapshot_preview1.MustInstantiate(ctx, runtime)
	runtime.NewHostModuleBuilder("env").
		NewFunctionBuilder().
		WithGoModuleFunction(api.GoModuleFunc(func(ctx context.Context, m api.Module, stack []uint64) {
			stack[0] = uint64(api.DecodeU32(uint64(bodySize)))
		}), []api.ValueType{}, []api.ValueType{api.ValueTypeI32}).
		Export("alloc").
		NewFunctionBuilder().
		WithGoModuleFunction(api.GoModuleFunc(func(ctx context.Context, m api.Module, stack []uint64) {
			offset := api.DecodeU32(stack[0])
			// size := api.DecodeU32(stack[1])
			m.Memory().Write(offset, []byte("anthony"))
		}), []api.ValueType{api.ValueTypeI32, api.ValueTypeI32}, []api.ValueType{api.ValueTypeI32}).
		Export("write_body").
		NewFunctionBuilder().
		WithGoModuleFunction(api.GoModuleFunc(func(ctx context.Context, m api.Module, stack []uint64) {
			offset := api.DecodeU32(stack[0])
			n := api.DecodeU32(stack[1])
			fmt.Println(offset)
			fmt.Println(n)

			b, _ := m.Memory().Read(offset, n)
			fmt.Println("response", string(b))
		}), []api.ValueType{api.ValueTypeI32, api.ValueTypeI32}, []api.ValueType{api.ValueTypeI32}).
		Export("write_response").Instantiate(ctx)

	compiledMod, err := runtime.CompileModule(ctx, b)
	if err != nil {
		log.Fatal(err)
	}
	modConfig := wazero.NewModuleConfig().WithStdout(os.Stdout).WithArgs("100").WithStartFunctions()
	mod, err := runtime.InstantiateModule(ctx, compiledMod, modConfig)
	if err != nil {
		log.Fatal(err)
	}
	mod.ExportedFunction("_start").Call(ctx)

}
