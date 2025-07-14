package service

import (
	"context"

	"github.com/wickedv43/go-goph-keeper/internal/storage"
)

type GophKeeper interface {
	//users
	NewUser(ctx context.Context, u *storage.User) (storage.User, error)
	User(ctx context.Context, id uint64) (storage.User, error)
	UserByLogin(ctx context.Context, login string) (storage.User, error)

	// Vaults
	CreateVault(ctx context.Context, v *storage.VaultRecord) error
	GetVault(ctx context.Context, vID uint64) (storage.VaultRecord, error)
	UpdateVault(ctx context.Context, v *storage.VaultRecord) error
	ListVaults(ctx context.Context, uID uint64) ([]storage.VaultRecord, error)
	DeleteVault(ctx context.Context, vID uint64) error
}

func (s *Service) NewUser(ctx context.Context, u *storage.User) (storage.User, error) {
	return s.storage.NewUser(ctx, u)
}

func (s *Service) User(ctx context.Context, id uint64) (storage.User, error) {
	return s.storage.User(ctx, id)
}

func (s *Service) UserByLogin(ctx context.Context, login string) (storage.User, error) {
	return s.storage.UserByLogin(ctx, login)
}

func (s *Service) CreateVault(ctx context.Context, v *storage.VaultRecord) error {
	return s.storage.CreateVault(ctx, v)
}

func (s *Service) GetVault(ctx context.Context, vID uint64) (storage.VaultRecord, error) {
	return s.storage.GetVault(ctx, vID)
}

func (s *Service) UpdateVault(ctx context.Context, v *storage.VaultRecord) error {
	return s.storage.UpdateVault(ctx, v)
}

func (s *Service) ListVaults(ctx context.Context, uID uint64) ([]storage.VaultRecord, error) {
	return s.storage.ListVaults(ctx, uID)
}

func (s *Service) DeleteVault(ctx context.Context, vID uint64) error {
	return s.storage.DeleteVault(ctx, vID)
}
