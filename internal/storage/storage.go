// Package storage provides persistent storage implementations for users and their encrypted vault records.
package storage

import (
	"time"

	"github.com/pkg/errors"
	"github.com/samber/do/v2"
	"github.com/wickedv43/go-goph-keeper/internal/config"
	"github.com/wickedv43/go-goph-keeper/internal/logger"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Storage provides access to the database and handles data persistence.
type Storage struct {
	db *gorm.DB

	log *zap.SugaredLogger
}

// NewStorage initializes the Storage instance with database connection and logger.
func NewStorage(i do.Injector) (*Storage, error) {
	cfg := do.MustInvoke[*config.Config](i)

	postgresDB, err := do.InvokeStruct[Storage](i)
	if err != nil {
		return nil, errors.Wrap(err, "invoke struct")
	}

	db, err := gorm.Open(postgres.Open(cfg.Database.DSN), &gorm.Config{})
	if err != nil {
		return nil, errors.Wrap(err, "Open postgres")
	}

	//
	postgresDB.db = db
	postgresDB.log = do.MustInvoke[*logger.Logger](i).Named("postgres")

	//?
	sqlDB, err := postgresDB.db.DB()
	if err != nil {
		return nil, errors.Wrap(err, "connection to db")
	}

	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	pStore := postgresDB

	err = pStore.Migrate()
	if err != nil {
		return nil, errors.Wrap(err, "migration failed")
	}

	return pStore, nil
}

// Shutdown closes the database connection and releases resources.
func (s *Storage) Shutdown() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return errors.Wrap(err, "close db")
	}
	if err = sqlDB.Close(); err != nil {
		return errors.Wrap(err, "close db")
	}

	s.log.Debug("database connection closed")
	return nil
}

// HealthCheck verifies the database connection and logs the result.
func (s *Storage) HealthCheck() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		s.log.Debug("failed to get sql.DB from gorm.DB")
	}

	if err = sqlDB.Ping(); err != nil {
		s.log.Debug("database connection is unhealthy")
	}

	s.log.Debug("Database health check: [ok]")

	return nil
}

// Migrate runs database schema migrations for all required models.
func (s *Storage) Migrate() error {
	if err := s.db.AutoMigrate(
		&User{},
		&VaultRecord{},
	); err != nil {
		s.log.Errorf("migration plan error: %v", err)
		return err
	}

	s.log.Debug("successfully migrated")
	return nil
}
