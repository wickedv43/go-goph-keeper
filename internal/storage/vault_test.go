package storage

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupVaultDB(t *testing.T) (*Storage, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)

	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	require.NoError(t, err)

	return &Storage{db: gormDB}, mock
}

func TestVaultCRUD(t *testing.T) {
	t.Run("CreateVault/success", func(t *testing.T) {
		store, mock := setupVaultDB(t)
		ctx := context.Background()

		vault := &VaultRecord{
			ID:            1,
			UserID:        42,
			Type:          RecordTypeNote,
			Title:         "Test Note",
			Metadata:      `{"tag":"secret"}`,
			EncryptedData: []byte("encrypted"),
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		mock.ExpectBegin()
		mock.ExpectQuery(`INSERT INTO "vault_records"`).
			WithArgs(vault.UserID, vault.Type, vault.Title, vault.Metadata, vault.EncryptedData, sqlmock.AnyArg(), sqlmock.AnyArg(), vault.ID).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(vault.ID))
		mock.ExpectCommit()

		err := store.CreateVault(ctx, vault)
		require.NoError(t, err)
	})

	t.Run("GetVault/success", func(t *testing.T) {
		store, mock := setupVaultDB(t)
		ctx := context.Background()

		vault := VaultRecord{
			ID:            1,
			UserID:        42,
			Type:          RecordTypeNote,
			Title:         "Test Note",
			Metadata:      `{"tag":"secret"}`,
			EncryptedData: []byte("encrypted"),
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		rows := sqlmock.NewRows([]string{"id", "user_id", "type", "title", "metadata", "encrypted_data", "created_at", "updated_at"}).
			AddRow(vault.ID, vault.UserID, vault.Type, vault.Title, vault.Metadata, vault.EncryptedData, vault.CreatedAt, vault.UpdatedAt)

		mock.ExpectQuery(`SELECT \* FROM "vault_records" WHERE id = \$1 ORDER BY "vault_records"\."id" LIMIT \$2`).WithArgs(vault.ID, 1).WillReturnRows(rows)

		res, err := store.GetVault(ctx, vault.ID)
		require.NoError(t, err)
		require.Equal(t, vault.ID, res.ID)
	})

	t.Run("UpdateVault/success", func(t *testing.T) {
		store, mock := setupVaultDB(t)
		ctx := context.Background()

		vault := &VaultRecord{
			ID:            1,
			UserID:        42,
			Type:          RecordTypeNote,
			Title:         "Updated Title",
			Metadata:      `{"tag":"secret"}`,
			EncryptedData: []byte("encrypted"),
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		mock.ExpectBegin()
		mock.ExpectExec(`UPDATE "vault_records"`).
			WithArgs(vault.UserID, vault.Type, vault.Title, vault.Metadata, vault.EncryptedData, sqlmock.AnyArg(), sqlmock.AnyArg(), vault.ID).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := store.UpdateVault(ctx, vault)
		require.NoError(t, err)
	})

	t.Run("ListVaults/success", func(t *testing.T) {
		store, mock := setupVaultDB(t)
		ctx := context.Background()

		vault := VaultRecord{
			ID:            1,
			UserID:        42,
			Type:          RecordTypeNote,
			Title:         "Test Note",
			Metadata:      `{"tag":"secret"}`,
			EncryptedData: []byte("encrypted"),
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		rows := sqlmock.NewRows([]string{"id", "user_id", "type", "title", "metadata", "encrypted_data", "created_at", "updated_at"}).
			AddRow(vault.ID, vault.UserID, vault.Type, vault.Title, vault.Metadata, vault.EncryptedData, vault.CreatedAt, vault.UpdatedAt)

		mock.ExpectQuery(`SELECT \* FROM "vault_records" WHERE user_id = \$1`).WithArgs(vault.UserID).WillReturnRows(rows)

		res, err := store.ListVaults(ctx, vault.UserID)
		require.NoError(t, err)
		require.Len(t, res, 1)
	})

	t.Run("DeleteVault/success", func(t *testing.T) {
		store, mock := setupVaultDB(t)
		ctx := context.Background()

		mock.ExpectBegin()
		mock.ExpectExec(`DELETE FROM "vault_records" WHERE "vault_records"\."id" = \$1`).WithArgs(uint64(1)).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := store.DeleteVault(ctx, 1)
		require.NoError(t, err)
	})
}
