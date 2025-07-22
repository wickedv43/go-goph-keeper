package main

import (
	pb "github.com/wickedv43/go-goph-keeper/internal/api"
)

func (g *GophKeeper) Login(email, password string) error {
	resp, err := g.client.Login(g.rootCtx, &pb.LoginRequest{
		Login:    email,
		Password: g.hashPassword(password),
	})
	if err != nil {
		return err
	}

	g.token = resp.GetToken()

	return nil
}

func (g *GophKeeper) Register(email, password string) error {
	_, err := g.client.Register(g.rootCtx, &pb.RegisterRequest{
		Login:    email,
		Password: g.hashPassword(password),
	})
	if err != nil {
		return err
	}

	return nil
}

func (g *GophKeeper) VaultList() (*pb.ListVaultsResponse, error) {
	return g.client.ListVaults(g.authCtx(), &pb.ListVaultsRequest{})
}
