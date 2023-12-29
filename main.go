package main

import (
	"context"
	"fmt"
	_ "net/http/pprof"
	"os"

	"github.com/anthdm/ffaas/pkg/runtime"
	"github.com/stealthrocket/wasi-go"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/sys"
)

func main() {
	if err := run("examples/go/app.wasm", []string{}); err != nil {
		if exitErr, ok := err.(*sys.ExitError); ok {
			os.Exit(int(exitErr.ExitCode()))
		}
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

type Runtime struct {
	r           wazero.Runtime
	compiledMod wazero.CompiledModule
	system      wasi.System
}

func run(wasmFile string, args []string) error {
	b, err := os.ReadFile(wasmFile)
	if err != nil {
		return err
	}
	ctx := context.Background()
	cache := wazero.NewCompilationCache()
	mod, _ := runtime.NewRequestModule(nil)
	runtime.Run(ctx, cache, b, mod)
	fmt.Println(mod)
	return nil
}
