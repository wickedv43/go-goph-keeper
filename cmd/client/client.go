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
}

func NewGophKeeper(i do.Injector) (*GophKeeper, error) {
	gophKeeper := &GophKeeper{}

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

	//pages
	pages := tview.NewPages()
	pages.AddPage("Login", gophKeeper.LoginPage(), true, true)
	pages.AddPage("Register", gophKeeper.RegisterPage(), true, false)
	//pages.AddPage("Main")

	//set
	gophKeeper.rootCtx = ctx
	gophKeeper.cancelCtx = cancel
	gophKeeper.pages = pages
	gophKeeper.client = client
	gophKeeper.cfg = cfg
	gophKeeper.tui = tview.NewApplication()

	return gophKeeper, nil
}

func (g *GophKeeper) Start() {
	if err := g.tui.SetRoot(g.pages, true).Run(); err != nil {
		panic(err)
	}
}
