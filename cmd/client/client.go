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
	tui   *tview.Application
	pages *tview.Pages

	client api.GophKeeperClient

	cfg *config.Config

	rootCtx   context.Context
	cancelCtx func()

	token string
}

func NewGophKeeper(i do.Injector) (*GophKeeper, error) {
	cfg := do.MustInvoke[*config.Config](i)

	target := fmt.Sprintf("localhost:%s", cfg.Server.Port)
	cc, err := grpc.Dial(
		target,
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	client := pb.NewGophKeeperClient(cc)
	ctx, cancel := context.WithCancel(context.Background())

	tui := tview.NewApplication()
	pages := tview.NewPages()

	g := &GophKeeper{
		cfg:       cfg,
		client:    client,
		tui:       tui,
		rootCtx:   ctx,
		cancelCtx: cancel,
		pages:     pages,
		token:     "",
	}

	pages.AddPage("Login", g.LoginPage(), true, true)
	pages.AddPage("Register", g.RegisterPage(), true, false)
	pages.AddPage("VaultCreate", g.NewVaultPage(), true, false)

	return g, nil
}

func (g *GophKeeper) Start() {
	if err := g.tui.SetRoot(g.pages, true).Run(); err != nil {
		panic(err)
	}
}
