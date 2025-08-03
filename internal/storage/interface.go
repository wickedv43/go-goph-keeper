package storage

import (
	"context"

	"github.com/pkg/errors"
)

// ErrLoginUsed indicates that the login is already taken by another user.
var (
	ErrLoginUsed = errors.New("login already used")
)

// DataKeeper defines the storage interface for users and their encrypted vault records.
type DataKeeper interface {
	// NewUser creates a new user in the storage.
	NewUser(ctx context.Context, u *User) (User, error)

	// User retrieves a user by their unique ID.
	User(ctx context.Context, uID uint64) (User, error)

	// UserByLogin retrieves a user by their login.
	UserByLogin(ctx context.Context, login string) (User, error)

	// CreateVault stores a new encrypted vault record.
	CreateVault(ctx context.Context, v *VaultRecord) error

	// GetVault retrieves a vault record by its ID.
	GetVault(ctx context.Context, vID uint64) (VaultRecord, error)

	// UpdateVault updates an existing vault record.
	UpdateVault(ctx context.Context, v *VaultRecord) error

	// ListVaults lists all vault records for the specified user.
	ListVaults(ctx context.Context, uID uint64) ([]VaultRecord, error)

	// DeleteVault removes a vault record by its ID.
	DeleteVault(ctx context.Context, vID uint64) error

	// Shutdown gracefully closes the storage and releases resources.
	Shutdown() error
}
