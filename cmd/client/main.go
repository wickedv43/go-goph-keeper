package main

import (
	"flag"
	"os"
	"syscall"

	"github.com/samber/do/v2"
	"github.com/wickedv43/go-goph-keeper/internal/config"
	"github.com/wickedv43/go-goph-keeper/internal/logger"
)

var configPath = flag.String("c", "config.yaml", "path to config file")

func main() {
	flag.Parse()

	i := do.New()

	do.ProvideNamedValue(i, "config.path", *configPath) //cfg
	do.Provide(i, config.NewConfig)
	do.Provide(i, logger.NewLogger)
	do.Provide(i, NewGophKeeper)

	log := do.MustInvoke[*logger.Logger](i).Named("GophKeeper")

	do.MustInvoke[*GophKeeper](i).Start()

	signals := []os.Signal{syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT, os.Interrupt}

	_, err := i.ShutdownOnSignals(signals...)
	if err != nil {
		log.Error(err)
	}
}
