package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/anthdm/raptor/internal/api"
	"github.com/anthdm/raptor/internal/config"
	"github.com/anthdm/raptor/internal/storage"
	"github.com/anthdm/raptor/internal/types"
	"github.com/google/uuid"
	"github.com/tetratelabs/wazero"
)

func main() {
	var (
		modCache   = storage.NewDefaultModCache()
		configFile string
		seed       bool
	)
	flagSet := flag.NewFlagSet("raptor", flag.ExitOnError)
	flagSet.StringVar(&configFile, "config", "config.toml", "")
	flagSet.BoolVar(&seed, "seed", false, "")
	flagSet.Parse(os.Args[1:])

	err := config.Parse(configFile)
	if err != nil {
		log.Fatal(err)
	}

	var (
		user    = config.Get().Storage.User
		pw      = config.Get().Storage.Password
		dbname  = config.Get().Storage.Name
		host    = config.Get().Storage.Host
		port    = config.Get().Storage.Port
		sslmode = config.Get().Storage.SSLMode
	)
	store, err := storage.NewSQLStore(user, pw, dbname, host, port, sslmode)
	if err != nil {
		log.Fatal(err)
	}

	if seed {
		seedEndpoint(store, modCache)
	}

	server := api.NewServer(store, store, modCache)
	fmt.Printf("api server running\t%s\n", config.ApiUrl())
	log.Fatal(server.Listen(config.Get().HTTPAPIAddr))
}

func seedEndpoint(store storage.Store, cache storage.ModCacher) {
	b, err := os.ReadFile("examples/js/index.js")
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

	deploy := types.NewDeployment(endpoint, b)
	endpoint.ActiveDeploymentID = deploy.ID
	endpoint.DeploymentHistory = append(endpoint.DeploymentHistory, &types.DeploymentHistory{
		ID:        deploy.ID,
		CreatedAT: deploy.CreatedAT,
	})
	store.CreateEndpoint(endpoint)
	store.CreateDeployment(deploy)
	err = store.UpdateEndpoint(endpoint.ID, storage.UpdateEndpointParams{
		ActiveDeployID: deploy.ID,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("endpoint seeded: %s/live/%s\n", config.IngressUrl(), endpoint.ID)
}

func compile(ctx context.Context, cache wazero.CompilationCache, blob []byte) {
	config := wazero.NewRuntimeConfig().WithCompilationCache(cache)
	runtime := wazero.NewRuntimeWithConfig(ctx, config)
	defer runtime.Close(ctx)
	runtime.CompileModule(ctx, blob)
}
