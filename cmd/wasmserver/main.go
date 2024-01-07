package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/anthdm/hollywood/actor"
	"github.com/anthdm/hollywood/cluster"
	"github.com/anthdm/hollywood/remote"
	"github.com/anthdm/raptor/pkg/actrs"
	"github.com/anthdm/raptor/pkg/config"
	"github.com/anthdm/raptor/pkg/storage"
)

func main() {
	var configFile string
	flagSet := flag.NewFlagSet("raptor", flag.ExitOnError)
	flagSet.StringVar(&configFile, "config", "config.toml", "")
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
	var (
		modCache    = storage.NewDefaultModCache()
		metricStore = store
	)

	remote := remote.New(config.Get().Cluster.WasmMemberAddr, nil)
	engine, err := actor.NewEngine(&actor.EngineConfig{
		Remote: remote,
	})
	if err != nil {
		log.Fatal(err)
	}
	c, err := cluster.New(cluster.Config{
		Region:          config.Get().Cluster.Region,
		Engine:          engine,
		ID:              config.Get().Cluster.ID,
		ClusterProvider: cluster.NewSelfManagedProvider(),
	})
	c.RegisterKind(actrs.KindRuntime, actrs.NewRuntime(store, metricStore, modCache), &cluster.KindConfig{})
	c.Start()

	server := actrs.NewWasmServer(config.Get().WASMServerAddr, c, store, metricStore, modCache)
	c.Engine().Spawn(server, actrs.KindWasmServer)
	fmt.Printf("wasm server running\t%s\n", config.Get().WASMServerAddr)

	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, syscall.SIGINT, syscall.SIGTERM)
	<-sigch
}
