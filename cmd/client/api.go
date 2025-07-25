package main

import (
	"fmt"
	"strings"

	"github.com/wickedv43/go-goph-keeper/cmd/client/internal/crypto"
	"github.com/wickedv43/go-goph-keeper/cmd/client/kv"
	pb "github.com/wickedv43/go-goph-keeper/internal/api"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (g *GophKeeper) Login(login, password string) error {
	resp, err := g.client.Login(g.rootCtx, &pb.LoginRequest{
		Login:    login,
		Password: g.hashPassword(password),
	})
	if err != nil {
		return err
	}

	err = g.storage.SaveContext(login, resp.Token)
	if err != nil {
		return err
	}

	key, err := g.storage.GetCurrentKey()
	if err != nil {
		return kv.ErrEmptyKey
	}

	fmt.Println(key)

	return nil
}

func (g *GophKeeper) Register(login, password string) ([]string, error) {
	_, err := g.client.Register(g.rootCtx, &pb.RegisterRequest{
		Login:    login,
		Password: g.hashPassword(password),
	})
	if err != nil {
		return nil, err
	}

	words, err := crypto.GenerateMnemonic()
	if err != nil {
		return nil, err
	}

	key := crypto.GenerateSeed(words, password)

	err = g.storage.SaveKey(login, key)
	if err != nil {
		return nil, err
	}
	mnemonic := strings.Split(words, " ")

	return mnemonic, nil
}

func (g *GophKeeper) VaultList() (*pb.ListVaultsResponse, error) {
	return g.client.ListVaults(g.authCtx(), &pb.ListVaultsRequest{})
}

func (g *GophKeeper) VaultCreate(v *pb.VaultRecord) (*emptypb.Empty, error) {
	return g.client.CreateVault(g.authCtx(), &pb.CreateVaultRequest{
		Record: v,
	})
}

func (g *GophKeeper) VaultGet(id uint64) (*pb.VaultRecord, error) {
	return g.client.GetVault(g.authCtx(), &pb.GetVaultRequest{
		VaultId: id,
	})
}

func (g *GophKeeper) VaultDelete(id uint64) (*emptypb.Empty, error) {
	return g.client.DeleteVault(g.authCtx(), &pb.DeleteVaultRequest{
		VaultId: id,
	})
}
