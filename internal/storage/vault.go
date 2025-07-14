package storage

import (
	"context"
	"time"
)

type RecordType string

const (
	RecordTypeLogin  RecordType = "login"
	RecordTypeNote   RecordType = "note"
	RecordTypeBinary RecordType = "binary"
	RecordTypeCard   RecordType = "card"
)

type VaultRecord struct {
	ID            uint64     `gorm:"primaryKey"`
	UserID        uint64     `gorm:"index;not null"`   // ForeignKey на User
	Type          RecordType `gorm:"size:32;not null"` // "login", "note", "card", "binary"
	Title         string     `gorm:"size:255;not null"`
	Metadata      string     `gorm:"type:jsonb"` // если Postgres — jsonb (можно просто text если SQLite)
	EncryptedData []byte     `gorm:"not null"`   // все данные шифруются на клиенте → сервер хранит EncryptedData
	CreatedAt     time.Time  `gorm:"autoCreateTime"`
	UpdatedAt     time.Time  `gorm:"autoUpdateTime"`
}

func (s *Storage) CreateVault(ctx context.Context, v *VaultRecord) error {
	return s.db.WithContext(ctx).Create(v).Error
}

func (s *Storage) GetVault(ctx context.Context, vID uint64) (VaultRecord, error) {
	var v VaultRecord
	err := s.db.WithContext(ctx).First(&v, "id = ?", vID).Error
	return v, err
}

func (s *Storage) UpdateVault(ctx context.Context, v *VaultRecord) error {
	// Убедись, что v.ID задан
	return s.db.WithContext(ctx).Save(v).Error
}

func (s *Storage) ListVaults(ctx context.Context, userID uint64) ([]VaultRecord, error) {
	var list []VaultRecord
	err := s.db.WithContext(ctx).Where("user_id = ?", userID).Find(&list).Error
	return list, err
}

func (s *Storage) DeleteVault(ctx context.Context, vID uint64) error {
	return s.db.WithContext(ctx).Delete(&VaultRecord{}, vID).Error
}
