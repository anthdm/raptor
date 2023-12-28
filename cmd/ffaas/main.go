package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/anthdm/ffaas/pkg/api"
	"github.com/anthdm/ffaas/pkg/proxy"
	"github.com/anthdm/ffaas/pkg/storage"
	"github.com/anthdm/ffaas/pkg/types"
	"github.com/google/uuid"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

const (
	httpListenAddr = ":3000"
	httpProxyAddr  = ":5000"
)

func main() {
	var (
		memstore = storage.NewMemoryStore()
		modCache = storage.NewDefaultModCache()
	)
	seed(memstore, modCache)

	server := api.NewServer(memstore, modCache)
	go func() {
		slog.Info("api server running", "port", httpListenAddr)
		log.Fatal(server.Listen(httpListenAddr))
	}()

	proxy := proxy.NewServer(memstore, modCache)
	slog.Info("app proxy server running", "port", httpProxyAddr)
	log.Fatal(proxy.Listen(httpProxyAddr))

	log.Fatal(proxy.Listen(":5000"))
}

func seed(store storage.Store, cache storage.ModCacher) {
	b, err := os.ReadFile("examples/go/app.wasm")
	if err != nil {
		log.Fatal(err)
	}
	app := types.App{
		ID:          uuid.New(),
		Name:        "My first ffaas app",
		Environment: map[string]string{"FOO": "fooenv"},
		CreatedAT:   time.Now(),
	}
	store.CreateApp(&app)
	deploy := types.Deploy{
		ID:        uuid.MustParse("09248ef6-c401-4601-8928-5964d61f2c61"),
		AppID:     app.ID,
		Blob:      b,
		CreatedAT: time.Now(),
	}
	store.CreateDeploy(&deploy)

	compCache := wazero.NewCompilationCache()
	var (
		ctx    = context.Background()
		config = wazero.NewRuntimeConfig().
			WithDebugInfoEnabled(true).
			WithCompilationCache(compCache)
		runtime = wazero.NewRuntimeWithConfig(ctx, config)
	)

	wasi_snapshot_preview1.MustInstantiate(ctx, runtime)

	_, err = runtime.CompileModule(ctx, b)
	if err != nil {
		log.Fatal(err)
	}

	cache.Put(deploy.ID, compCache)

	fmt.Printf("My first ffaas app available localhost:5000/%s\n", deploy.ID)
}
