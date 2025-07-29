package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/samber/do/v2"
	"github.com/spf13/cobra"
	"github.com/wickedv43/go-goph-keeper/cmd/client/internal/kv"
	"github.com/wickedv43/go-goph-keeper/internal/api"
	"github.com/wickedv43/go-goph-keeper/internal/config"
	"github.com/wickedv43/go-goph-keeper/internal/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/wickedv43/go-goph-keeper/internal/api"
)

type GophKeeper struct {
	rootCmd *cobra.Command

	client api.GophKeeperClient

	storage kv.Storage

	cfg *config.Config
	log *logger.Logger

	rootCtx   context.Context
	cancelCtx func()
}

func NewGophKeeper(i do.Injector) (*GophKeeper, error) {
	cfg := do.MustInvoke[*config.Config](i)
	log := do.MustInvoke[*logger.Logger](i)
	kv := do.MustInvoke[*kv.KV](i)

	target := fmt.Sprintf("localhost:%s", cfg.Server.Port)
	cc, err := grpc.Dial(
		target,
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	client := pb.NewGophKeeperClient(cc)
	ctx, cancel := context.WithCancel(context.Background())

	g := &GophKeeper{
		storage: kv,
		cfg:     cfg,
		log:     log,
		client:  client,

		rootCtx:   ctx,
		cancelCtx: cancel,
	}

	return g, nil
}

func (g *GophKeeper) Start() {
	args := os.Args[1:]

	if len(args) == 0 {
		// если не передано ни одной команды — запускаем shell
		if err := g.ShellCMD().RunE(g.rootCmd, nil); err != nil {
			fmt.Println("❌ Ошибка в shell:", err)
		}
		return
	}

	err := g.rootCmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}
