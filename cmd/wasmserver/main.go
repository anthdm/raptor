package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/anthdm/ffaas/pkg/actrs"
	"github.com/anthdm/ffaas/pkg/config"
	"github.com/anthdm/ffaas/pkg/storage"
	"github.com/anthdm/hollywood/actor"
	"github.com/anthdm/hollywood/cluster"
	"github.com/anthdm/hollywood/remote"
)

func main() {
	var configFile string
	flagSet := flag.NewFlagSet("ffaas", flag.ExitOnError)
	flagSet.StringVar(&configFile, "config", "config.toml", "")
	flagSet.Parse(os.Args[1:])

	err := config.Parse(configFile)
	if err != nil {
		log.Fatal(err)
	}

	store, err := storage.NewRedisStore()
	if err != nil {
		log.Fatal(err)
	}
	var (
		modCache    = storage.NewDefaultModCache()
		metricStore = storage.NewMemoryMetricStore()
	)

	remote := remote.New(config.Get().WASMClusterAddr, nil)
	engine, err := actor.NewEngine(&actor.EngineConfig{
		Remote: remote,
	})
	if err != nil {
		log.Fatal(err)
	}
	// TODO: Get these values from the config.
	c, err := cluster.New(cluster.Config{
		Region:          "f",
		Engine:          engine,
		ID:              "member1",
		ClusterProvider: cluster.NewSelfManagedProvider(),
	})
	c.RegisterKind(actrs.KindRuntime, actrs.NewRuntime(store, modCache), &cluster.KindConfig{})
	c.Start()

	server := actrs.NewWasmServer(config.Get().WASMServerAddr, c, store, metricStore, modCache)
	c.Engine().Spawn(server, actrs.KindWasmServer)
	fmt.Printf("wasm server running\t%s\n", config.Get().WASMServerAddr)

	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, syscall.SIGINT, syscall.SIGTERM)
	<-sigch
}
