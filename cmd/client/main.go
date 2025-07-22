//go:generate go run generate/main.go
package main

import (
	"os"
	"syscall"

	"github.com/samber/do/v2"
	"github.com/spf13/cobra"
	"github.com/wickedv43/go-goph-keeper/internal/config"
	"github.com/wickedv43/go-goph-keeper/internal/logger"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

var configPath string

func main() {
	rootCmd := &cobra.Command{
		Use:   "gk",
		Short: "GophKeeper CLI",
	}
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "./config/config.local.yaml", "Путь до config.yaml")

	// 👇 Парсим флаги вручную, чтобы configPath был доступен
	rootCmd.ParseFlags(os.Args[1:])

	i := do.New()
	do.ProvideNamedValue(i, "config.path", configPath)
	do.Provide(i, config.NewConfig)
	do.Provide(i, logger.NewLogger)
	do.Provide(i, NewGophKeeper)

	log := do.MustInvoke[*logger.Logger](i).Named("GophKeeper")

	gophKeeper := do.MustInvoke[*GophKeeper](i)
	gophKeeper.rootCmd = rootCmd

	gophKeeper.rootCmd.AddCommand(gophKeeper.LoginCMD())

	gophKeeper.Start()

	signals := []os.Signal{syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT, os.Interrupt}

	_, err := i.ShutdownOnSignals(signals...)
	if err != nil {
		log.Error(err)
	}
}
