package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/anthdm/hollywood/cluster"
	"github.com/anthdm/raptor/internal/actrs"
	"github.com/anthdm/raptor/internal/config"
	"github.com/anthdm/raptor/internal/storage"
)

func main() {
	if err := config.Parse("config.toml"); err != nil {
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
		WithListenAddr("127.0.0.1:6000").
		WithRegion("eu-west").
		WithID("member_2")
	c, err := cluster.New(clusterConfig)
	if err != nil {
		log.Fatal(err)
	}
	c.RegisterKind(actrs.KindRuntime, actrs.NewRuntime(store, modCache), &cluster.KindConfig{})
	// c.Engine().Spawn(actrs.NewMetric, actrs.KindMetric, actor.WithID("1"))
	// c.Engine().Spawn(actrs.NewRuntimeManager(c), actrs.KindRuntimeManager, actor.WithID("1"))
	// c.Engine().Spawn(actrs.NewRuntimeLog, actrs.KindRuntimeLog, actor.WithID("1"))
	c.Start()

	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, syscall.SIGINT, syscall.SIGTERM)
	<-sigch
}
