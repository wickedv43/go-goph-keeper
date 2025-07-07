package service

import (
	"context"

	"github.com/wickedv43/go-goph-keeper/internal/storage"
)

type GophKeeper interface {
	//users
	NewUser(ctx context.Context, u *storage.User) (storage.User, error)
	User(ctx context.Context, id uint64) (storage.User, error)
}

func (s *Service) NewUser(ctx context.Context, u *storage.User) (storage.User, error) {
	return s.storage.NewUser(ctx, u)
}

func (s *Service) User(ctx context.Context, id uint64) (storage.User, error) {
	return s.storage.User(ctx, id)
}
