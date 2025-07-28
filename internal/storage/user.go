package storage

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// User represents an application user with a unique login and hashed password.
type User struct {
	ID           uint64    `gorm:"primaryKey"`
	Login        string    `gorm:"uniqueIndex;size:255;not null"`
	PasswordHash string    `gorm:"size:255;not null"` // bcrypt hash
	CreatedAt    time.Time `gorm:"autoCreateTime"`
}

// NewUser creates a new user if the login is not already taken.
func (s *Storage) NewUser(ctx context.Context, u *User) (User, error) {
	var existing User

	err := s.db.WithContext(ctx).Where("login = ?", u.Login).First(&existing).Error
	if err == nil {
		return User{}, ErrLoginUsed
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return User{}, err
	}

	if err = s.db.WithContext(ctx).Create(u).Error; err != nil {
		return User{}, err
	}

	return *u, nil
}

// User retrieves a user by their ID.
func (s *Storage) User(ctx context.Context, id uint64) (User, error) {
	var user User

	result := s.db.WithContext(ctx).First(&user, id)
	if result.Error != nil {
		return User{}, result.Error
	}

	return user, nil
}

// UserByLogin retrieves a user by their login.
func (s *Storage) UserByLogin(ctx context.Context, login string) (User, error) {
	var user User
	result := s.db.WithContext(ctx).Where("login = ?", login).First(&user)
	if result.Error != nil {
		return User{}, result.Error
	}
	return user, nil
}
