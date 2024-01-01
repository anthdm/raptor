package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/anthdm/ffaas/pkg/api"
	"github.com/anthdm/ffaas/pkg/config"
	"github.com/anthdm/ffaas/pkg/storage"
	"github.com/anthdm/ffaas/pkg/types"
	"github.com/google/uuid"
	"github.com/tetratelabs/wazero"
)

func main() {
	var (
		modCache    = storage.NewDefaultModCache()
		metricStore = storage.NewMemoryMetricStore()
		configFile  string
		seed        bool
	)
	flagSet := flag.NewFlagSet("ffaas", flag.ExitOnError)
	flagSet.StringVar(&configFile, "config", "config.toml", "")
	flagSet.BoolVar(&seed, "seed", false, "")
	flagSet.Parse(os.Args[1:])

	err := config.Parse(configFile)
	if err != nil {
		log.Fatal(err)
	}

	cfg := storage.NewBoltConfig().
		WithPath(config.Get().BoltStoragePath).
		WithReadOnly(false)
	store, err := storage.NewBoltStore(cfg)
	if err != nil {
		log.Fatal(err)
	}

	if seed {
		seedEndpoint(store, modCache)
	}

	server := api.NewServer(store, metricStore, modCache)
	fmt.Printf("api server running\t%s\n", config.GetApiUrl())
	log.Fatal(server.Listen(config.Get().APIServerAddr))
}

func seedEndpoint(store storage.Store, cache storage.ModCacher) {
	b, err := os.ReadFile("examples/go/app.wasm")
	if err != nil {
		log.Fatal(err)
	}
	endpoint := &types.Endpoint{
		ID:          uuid.MustParse("09248ef6-c401-4601-8928-5964d61f2c61"),
		Name:        "Catfact parser",
		Environment: map[string]string{"FOO": "bar"},
		CreatedAT:   time.Now(),
	}

	deploy := types.NewDeploy(endpoint, b)
	endpoint.ActiveDeployID = deploy.ID
	endpoint.URL = config.GetWasmUrl() + "/" + endpoint.ID.String()
	endpoint.DeployHistory = append(endpoint.DeployHistory, deploy)
	store.CreateEndpoint(endpoint)
	store.CreateDeploy(deploy)
	fmt.Printf("endpoint seeded: %s\n", endpoint.URL)
}

func compile(ctx context.Context, cache wazero.CompilationCache, blob []byte) {
	config := wazero.NewRuntimeConfig().WithCompilationCache(cache)
	runtime := wazero.NewRuntimeWithConfig(ctx, config)
	defer runtime.Close(ctx)
	runtime.CompileModule(ctx, blob)
}
