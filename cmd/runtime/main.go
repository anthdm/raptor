package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/anthdm/hollywood/actor"
	"github.com/anthdm/hollywood/cluster"
	"github.com/anthdm/raptor/internal/actrs"
	"github.com/anthdm/raptor/internal/config"
	"github.com/anthdm/raptor/internal/storage"
)

func main() {
	var (
		configFile string
		address    string
		id         string
		region     string
	)

	flagSet := flag.NewFlagSet("runtime", flag.ExitOnError)
	flagSet.StringVar(&configFile, "config", "config.toml", "")
	flagSet.StringVar(&address, "cluster-addr", "127.0.0.1:8134", "")
	flagSet.StringVar(&id, "id", "runtime", "")
	flagSet.StringVar(&region, "region", "default", "")
	flagSet.Parse(os.Args[1:])

	if err := config.Parse(configFile); err != nil {
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
		modCache = storage.NewDefaultModCache()
		// metricStore = store
	)
	clusterConfig := cluster.NewConfig().
		WithListenAddr(address).
		WithRegion(region).
		WithID(id)
	c, err := cluster.New(clusterConfig)
	if err != nil {
		log.Fatal(err)
	}
	c.RegisterKind(actrs.KindRuntime, actrs.NewRuntime(store, modCache), &cluster.KindConfig{})
	c.Engine().Spawn(actrs.NewMetric, actrs.KindMetric, actor.WithID("1"))
	c.Engine().Spawn(actrs.NewRuntimeLog, actrs.KindRuntimeLog, actor.WithID("1"))
	c.Start()

	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, syscall.SIGINT, syscall.SIGTERM)
	<-sigch
}
