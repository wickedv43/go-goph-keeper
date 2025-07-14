package main

import (
	"context"
	"fmt"

	"github.com/rivo/tview"
	"github.com/samber/do/v2"
	"github.com/wickedv43/go-goph-keeper/internal/api"
	"github.com/wickedv43/go-goph-keeper/internal/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/wickedv43/go-goph-keeper/internal/api"
)

type GophKeeper struct {
	tui    *tview.Application
	client api.GophKeeperClient

	cfg *config.Config

	rootCtx   context.Context
	cancelCtx func()
}

func NewGophKeeper(i do.Injector) (*GophKeeper, error) {
	cfg := do.MustInvoke[*config.Config](i)

	target := fmt.Sprintf("localhost:%s", cfg.Server.Port)

	cc, err := grpc.NewClient(
		target,
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	// gRPC-клиент сервера Auth
	client := pb.NewGophKeeperClient(cc)

	ctx, cancel := context.WithCancel(context.Background())

	return &GophKeeper{
		tui:    tview.NewApplication(),
		client: client,
		cfg:    cfg,

		rootCtx:   ctx,
		cancelCtx: cancel,
	}, nil
}

func (g *GophKeeper) Start() {
	menu := g.startMenu()

	menu.SetBorder(true).SetTitle("GophKeeper TUI")

	if err := g.tui.SetRoot(menu, true).Run(); err != nil {
		panic(err)
	}
}
