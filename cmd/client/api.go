package main

import (
	"strings"

	"github.com/wickedv43/go-goph-keeper/cmd/client/internal/crypto"
	pb "github.com/wickedv43/go-goph-keeper/internal/api"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Login performs user authentication and stores the received access token in the current context.
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

	return nil
}

// Register creates a new user account and returns the generated mnemonic for local key storage.
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

// VaultList retrieves the list of vault records for the authenticated user.
func (g *GophKeeper) VaultList() (*pb.ListVaultsResponse, error) {
	return g.client.ListVaults(g.authCtx(), &pb.ListVaultsRequest{})
}

// VaultCreate creates a new vault record using the provided data.
func (g *GophKeeper) VaultCreate(v *pb.VaultRecord) (*emptypb.Empty, error) {
	return g.client.CreateVault(g.authCtx(), &pb.CreateVaultRequest{
		Record: v,
	})
}

// VaultGet retrieves a specific vault record by its ID.
func (g *GophKeeper) VaultGet(id uint64) (*pb.VaultRecord, error) {
	return g.client.GetVault(g.authCtx(), &pb.GetVaultRequest{
		VaultId: id,
	})
}

// VaultDelete deletes a vault record by its ID.
func (g *GophKeeper) VaultDelete(id uint64) (*emptypb.Empty, error) {
	return g.client.DeleteVault(g.authCtx(), &pb.DeleteVaultRequest{
		VaultId: id,
	})
}
