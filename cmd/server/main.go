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

var configPath = flag.String("c", "config.yaml", "path to config file")

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
	do.Provide(i, service.NewService)
	do.Provide(i, storage.NewStorage)

	log := do.MustInvoke[*logger.Logger](i).Named("GophKeeper")
	log.Info("starting app...")

	go do.MustInvoke[*server.Server](i).Start()

	signals := []os.Signal{syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT, os.Interrupt}

	_, err := i.ShutdownOnSignalsWithContext(ctx, signals...)
	if err != nil {
		log.Error(err)
	}

	log.Info("grace shutdown")
}
