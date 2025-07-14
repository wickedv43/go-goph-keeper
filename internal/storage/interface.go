package storage

import (
	"context"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

var (
	ErrLoginUsed = errors.New("login already used")
)

type DataKeeper interface {
	//users
	NewUser(ctx context.Context, u *User) (User, error)
	User(ctx context.Context, id uint64) (User, error)

	Shutdown() error
}

// users
func (s *Storage) NewUser(ctx context.Context, u *User) (User, error) {
	var existing User

	err := s.db.WithContext(ctx).Where("login = ?", u.Login).First(&existing).Error
	if err == nil {
		return User{}, ErrLoginUsed
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return User{}, err
	}

	// Логин свободен, создаём нового пользователя
	if err := s.db.WithContext(ctx).Create(u).Error; err != nil {
		return User{}, err
	}

	return *u, nil
}

func (s *Storage) User(ctx context.Context, id uint64) (User, error) {
	var user User

	result := s.db.WithContext(ctx).First(&user, id)
	if result.Error != nil {
		return User{}, result.Error
	}

	return user, nil
}
