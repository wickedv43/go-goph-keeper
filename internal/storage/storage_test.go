package storage

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestStorage_Shutdown(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	// Ожидаем, что будет вызван Close()
	mock.ExpectClose()

	gdb, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	require.NoError(t, err)

	s := &Storage{
		db:  gdb,
		log: zap.NewNop().Sugar(),
	}

	require.NoError(t, s.Shutdown())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestStorage_HealthCheck(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	mock.ExpectPing()

	gdb, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	require.NoError(t, err)

	s := &Storage{
		db:  gdb,
		log: zap.NewNop().Sugar(),
	}

	err = s.HealthCheck()
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}
