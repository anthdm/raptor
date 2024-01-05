package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/anthdm/run/pkg/api"
	"github.com/anthdm/run/pkg/config"
	"github.com/anthdm/run/pkg/storage"
	"github.com/anthdm/run/pkg/types"
	"github.com/bananabytelabs/wazero"
	"github.com/google/uuid"
)

func main() {
	var (
		modCache   = storage.NewDefaultModCache()
		configFile string
		seed       bool
	)
	flagSet := flag.NewFlagSet("run", flag.ExitOnError)
	flagSet.StringVar(&configFile, "config", "config.toml", "")
	flagSet.BoolVar(&seed, "seed", false, "")
	flagSet.Parse(os.Args[1:])

	err := config.Parse(configFile)
	if err != nil {
		log.Fatal(err)
	}

	store, err := storage.NewRedisStore()
	if err != nil {
		log.Fatal(err)
	}

	if seed {
		seedEndpoint(store, modCache)
	}

	server := api.NewServer(store, store, modCache)
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
		Runtime:     "go",
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
