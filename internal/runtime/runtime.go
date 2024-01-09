package runtime

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/anthdm/raptor/internal/spidermonkey"
	"github.com/google/uuid"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

type Args struct {
	Stdout       io.Writer
	DeploymentID uuid.UUID
	Engine       string
	Blob         []byte
	Cache        wazero.CompilationCache
}

type Runtime struct {
	stdout       io.Writer
	ctx          context.Context
	deploymentID uuid.UUID
	engine       string
	blob         []byte
	mod          wazero.CompiledModule
	runtime      wazero.Runtime
}

func New(ctx context.Context, args Args) (*Runtime, error) {
	config := wazero.NewRuntimeConfigCompiler().WithCompilationCache(args.Cache)
	r := &Runtime{
		runtime:      wazero.NewRuntimeWithConfig(ctx, config),
		ctx:          ctx,
		deploymentID: args.DeploymentID,
		engine:       args.Engine,
		stdout:       args.Stdout,
	}
	wasi_snapshot_preview1.MustInstantiate(ctx, r.runtime)

	switch args.Engine {
	case "js":
		r.blob = spidermonkey.WasmBlob
	default:
		r.blob = args.Blob
	}

	mod, err := r.runtime.CompileModule(ctx, r.blob)
	if err != nil {
		return nil, fmt.Errorf("runtime failed to compile module: %s", err)
	}
	r.mod = mod

	return r, nil
}

func (r *Runtime) Blob() []byte {
	return r.blob
}

func (r *Runtime) Invoke(stdin io.Reader, env map[string]string, args ...string) error {
	modConf := wazero.NewModuleConfig().
		WithStdin(stdin).
		WithStdout(r.stdout).
		WithStderr(os.Stderr).
		WithArgs(args...)
	for k, v := range env {
		modConf = modConf.WithEnv(k, v)
	}
	_, err := r.runtime.InstantiateModule(r.ctx, r.mod, modConf)
	return err
}

func (r *Runtime) Close() error {
	return r.runtime.Close(r.ctx)
}

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
