package main

import (
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
	err := config.Parse("ffaas.toml")
	if err != nil {
		log.Fatal(err)
	}

	cfg := storage.NewBoltConfig().
		WithPath(config.Get().StoragePath).
		WithReadOnly(true)
	store, err := storage.NewBoltStore(cfg)
	if err != nil {
		log.Fatal(err)
	}
	var (
		modCache    = storage.NewDefaultModCache()
		metricStore = storage.NewMemoryMetricStore()
	)

	remote := remote.New("localhost:6666", nil)
	engine, err := actor.NewEngine(&actor.EngineConfig{
		Remote: remote,
	})
	if err != nil {
		log.Fatal(err)
	}
	c, err := cluster.New(cluster.Config{
		Region:          "f",
		Engine:          engine,
		ID:              "member1",
		ClusterProvider: cluster.NewSelfManagedProvider(),
	})
	c.RegisterKind(actrs.KindRuntime, actrs.NewRuntime(store), &cluster.KindConfig{})
	c.Start()

	server := actrs.NewWasmServer(config.Get().WASMServerAddr, c, store, metricStore, modCache)
	c.Engine().Spawn(server, actrs.KindWasmServer)
	fmt.Printf("wasm server running\t%s\n", config.Get().WASMServerAddr)

	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, syscall.SIGINT, syscall.SIGTERM)
	<-sigch
}
