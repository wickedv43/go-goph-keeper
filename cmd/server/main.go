// GophKeeper server entry point. Initializes dependencies and starts the gRPC server.
package main

import (
	"context"
	"flag"
	"os"
	"syscall"

	"github.com/samber/do/v2"
	"github.com/wickedv43/go-goph-keeper/internal/config"
	"github.com/wickedv43/go-goph-keeper/internal/logger"
	"github.com/wickedv43/go-goph-keeper/internal/server"
	"github.com/wickedv43/go-goph-keeper/internal/service"
	"github.com/wickedv43/go-goph-keeper/internal/storage"
)

// configPath defines the path to the server configuration file.
var configPath = flag.String("c", "config.server.yaml", "path to config file")

// main initializes the dependency graph, starts the gRPC server, and waits for termination signals.
func main() {
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background()) //root ctx
	defer cancel()

	i := do.New()

	do.ProvideNamedValue(i, "root.context", ctx)
	do.ProvideNamedValue(i, "config.path", *configPath) //cfg
	do.Provide(i, config.NewConfig)
	do.Provide(i, logger.NewLogger)
	do.Provide(i, server.NewServer)

	log := do.MustInvoke[*logger.Logger](i).Named("GophKeeper")
	log.Debug("starting app...")

	do.Provide(i, func(i do.Injector) (service.GophKeeper, error) {
		return service.NewService(i)
	})

	do.Provide(i, func(i do.Injector) (storage.DataKeeper, error) {
		return storage.NewStorage(i)
	})

	go do.MustInvoke[*server.Server](i).Start()

	signals := []os.Signal{syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT, os.Interrupt}

	_, err := i.ShutdownOnSignalsWithContext(ctx, signals...)
	if err != nil {
		log.Error(err)
	}

	log.Debug("grace shutdown")
}
