package runtime

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

type InvokeArgs struct {
	Blob  []byte
	Cache wazero.CompilationCache
	Out   io.Writer
	In    io.Reader
	Env   map[string]string
	Debug bool
	Args  []string
}

func Invoke(ctx context.Context, args InvokeArgs) error {
	start := time.Now()
	// only arm64
	// config := opt.NewRuntimeConfigOptimizingCompiler().WithCompilationCache(args.Cache)
	config := wazero.NewRuntimeConfigCompiler().WithCompilationCache(args.Cache)
	runtime := wazero.NewRuntimeWithConfig(ctx, config)
	defer runtime.Close(ctx)

	wasi_snapshot_preview1.MustInstantiate(ctx, runtime)
	if args.Debug {
		fmt.Println("runtime new: ", time.Since(start))
	}

	start = time.Now()
	mod, err := runtime.CompileModule(ctx, args.Blob)
	if err != nil {
		slog.Warn("compiling module failed", "err", err)
		return err
	}
	if args.Debug {
		fmt.Println("runtime compile module: ", time.Since(start))
	}

	start = time.Now()
	modConf := wazero.NewModuleConfig().
		WithStdin(args.In).
		WithStdout(args.Out).
		WithStderr(os.Stderr).
		WithArgs(args.Args...)
	for k, v := range args.Env {
		modConf = modConf.WithEnv(k, v)
	}
	_, err = runtime.InstantiateModule(ctx, mod, modConf)
	if args.Debug {
		fmt.Println("runtime instantiate: ", time.Since(start))
	}

	return err
}
