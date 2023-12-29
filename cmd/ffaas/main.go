package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/anthdm/ffaas/pkg/api"
	"github.com/anthdm/ffaas/pkg/runtime"
	"github.com/anthdm/ffaas/pkg/storage"
	"github.com/anthdm/ffaas/pkg/types"
	"github.com/anthdm/ffaas/pkg/version"
	"github.com/anthdm/ffaas/pkg/wasm"
	"github.com/google/uuid"
	"github.com/pelletier/go-toml/v2"
	"github.com/tetratelabs/wazero"
)

const defaultConfig = `
wasmServerAddr 	= ":5000"
apiServerAddr 	= ":3000"
`

type Config struct {
	APIServerAddr  string
	WASMServerAddr string
}

func parseConfig(path string) (Config, error) {
	var config Config
	_, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		if err := os.WriteFile("ffaas.toml", []byte(defaultConfig), os.ModePerm); err != nil {
			return config, err
		}
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return config, err
	}
	err = toml.Unmarshal(b, &config)
	return config, err
}

func main() {
	var (
		memstore   = storage.NewMemoryStore()
		modCache   = storage.NewDefaultModCache()
		configFile string
		seed       bool
	)
	flagSet := flag.NewFlagSet("ffaas", flag.ExitOnError)
	flagSet.StringVar(&configFile, "config", "ffaas.toml", "")
	flagSet.BoolVar(&seed, "seed", false, "")
	flagSet.Parse(os.Args[1:])

	config, err := parseConfig(configFile)
	if err != nil {
		log.Fatal(err)
	}

	if seed {
		seedApplication(memstore, modCache)
	}

	fmt.Println(banner())
	fmt.Println("The opensource faas platform powered by WASM")
	fmt.Println()
	server := api.NewServer(memstore, modCache)
	go func() {
		fmt.Printf("api server running\t0.0.0.0%s\n", config.APIServerAddr)
		log.Fatal(server.Listen(config.APIServerAddr))
	}()

	wasmServer := wasm.NewServer(memstore, modCache)
	fmt.Printf("wasm server running\t0.0.0.0%s\n", config.WASMServerAddr)
	log.Fatal(wasmServer.Listen(config.WASMServerAddr))
}

func seedApplication(store storage.Store, cache storage.ModCacher) {
	b, err := os.ReadFile("examples/go/app.wasm")
	if err != nil {
		log.Fatal(err)
	}
	app := types.App{
		ID:          uuid.MustParse("09248ef6-c401-4601-8928-5964d61f2c61"),
		Name:        "My first ffaas app",
		Environment: map[string]string{"FOO": "fooenv"},
		CreatedAT:   time.Now(),
	}

	deploy := types.NewDeploy(app.ID, b)
	app.ActiveDeploy = deploy.ID
	app.Endpoint = "http://localhost:5000/" + app.ID.String()
	app.Deploys = append(app.Deploys, *deploy)
	store.CreateApp(&app)
	store.CreateDeploy(deploy)

	compCache := wazero.NewCompilationCache()
	runtime.Compile(context.Background(), compCache, deploy.Blob)
	cache.Put(app.ID, compCache)
}

func banner() string {
	return fmt.Sprintf(`
  __  __                
 / _|/ _|               
| |_| |_ __ _  __ _ ___ 
|  _|  _/ _  |/ _  / __|
| | | || (_| | (_| \__ \
|_| |_| \__,_|\__,_|___/ V%s
	`, version.Version)
}
