package storage

import (
	"context"
	"fmt"
	"time"
)

// RecordType defines the type of vault record, such as login credentials or notes.
type RecordType string

const (
	// RecordTypeLogin represents a login/password record.
	RecordTypeLogin RecordType = "login"

	// RecordTypeNote represents a plain text note.
	RecordTypeNote RecordType = "note"

	// RecordTypeBinary represents binary data.
	RecordTypeBinary RecordType = "binary"

	// RecordTypeCard represents credit card information.
	RecordTypeCard RecordType = "card"
)

// VaultRecord represents an encrypted data entry belonging to a user.
type VaultRecord struct {
	ID            uint64     `gorm:"primaryKey"`
	UserID        uint64     `gorm:"index;not null"`   // Foreign key to User
	Type          RecordType `gorm:"size:32;not null"` // One of: "login", "note", "card", "binary"
	Title         string     `gorm:"size:255;not null"`
	Metadata      string     `gorm:"type:jsonb"` // Optional metadata, stored as JSON
	EncryptedData []byte     `gorm:"not null"`   // Encrypted content, handled on the client side
	CreatedAt     time.Time  `gorm:"autoCreateTime"`
	UpdatedAt     time.Time  `gorm:"autoUpdateTime"`
}

// CreateVault stores a new vault record in the database.
func (s *Storage) CreateVault(ctx context.Context, v *VaultRecord) error {
	return s.db.WithContext(ctx).Create(v).Error
}

// GetVault retrieves a vault record by its ID.
func (s *Storage) GetVault(ctx context.Context, vID uint64) (VaultRecord, error) {
	var v VaultRecord
	err := s.db.WithContext(ctx).First(&v, "id = ?", vID).Error
	return v, err
}

// UpdateVault updates an existing vault record.
func (s *Storage) UpdateVault(ctx context.Context, v *VaultRecord) error {
	// Убедись, что v.ID задан
	return s.db.WithContext(ctx).Save(v).Error
}

// ListVaults returns all vault records associated with the specified user.
func (s *Storage) ListVaults(ctx context.Context, userID uint64) ([]VaultRecord, error) {
	var list []VaultRecord
	err := s.db.WithContext(ctx).Where("user_id = ?", userID).Find(&list).Error
	return list, err
}

// DeleteVault deletes a vault record by its ID.
func (s *Storage) DeleteVault(ctx context.Context, vID uint64) error {
	res := s.db.WithContext(ctx).Delete(&VaultRecord{}, vID)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("запись с id=%d не найдена", vID)
	}
	return nil
}
