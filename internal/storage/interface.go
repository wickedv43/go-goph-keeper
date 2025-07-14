package storage

import (
	"context"

	"github.com/pkg/errors"
)

var (
	ErrLoginUsed = errors.New("login already used")
)

type DataKeeper interface {
	//users
	NewUser(ctx context.Context, u *User) (User, error)
	User(ctx context.Context, uID uint64) (User, error)
	UserByLogin(ctx context.Context, login string) (User, error)

	CreateVault(ctx context.Context, v *VaultRecord) error
	GetVault(ctx context.Context, vID uint64) (VaultRecord, error)
	UpdateVault(ctx context.Context, v *VaultRecord) error
	ListVaults(ctx context.Context, uID uint64) ([]VaultRecord, error)
	DeleteVault(ctx context.Context, vID uint64) error

	Shutdown() error
}
