package storage

import "time"

type RecordType string

const (
	RecordTypeLogin  RecordType = "login"
	RecordTypeNote   RecordType = "note"
	RecordTypeBinary RecordType = "binary"
	RecordTypeCard   RecordType = "card"
)

type User struct {
	ID           uint64    `gorm:"primaryKey"`
	Login        string    `gorm:"uniqueIndex;size:255;not null"`
	PasswordHash string    `gorm:"size:255;not null"` // bcrypt hash
	CreatedAt    time.Time `gorm:"autoCreateTime"`
}
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
