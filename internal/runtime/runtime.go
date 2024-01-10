package runtime

import (
	"context"
	"fmt"
	"io"
	"os"

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

	mod, err := r.runtime.CompileModule(ctx, args.Blob)
	if err != nil {
		return nil, fmt.Errorf("runtime failed to compile module: %s", err)
	}
	r.mod = mod

	return r, nil
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
