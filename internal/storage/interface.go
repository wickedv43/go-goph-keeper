package storage

import (
	"context"
)

type DataKeeper interface {
	//users
	NewUser(ctx context.Context, u *User) (User, error)
	User(ctx context.Context, id uint64) (User, error)

	Shutdown() error
}

// users
func (s *Storage) NewUser(ctx context.Context, u *User) (User, error) {
	var user User

	result := s.db.WithContext(ctx).FirstOrCreate(&user, u)
	if result.Error != nil {
		return User{}, result.Error
	}

	return user, nil
}

func (s *Storage) User(ctx context.Context, id uint64) (User, error) {
	var user User

	result := s.db.WithContext(ctx).First(&user, id)
	if result.Error != nil {
		return User{}, result.Error
	}

	return user, nil
}
