package service

import (
	"context"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"github.com/wickedv43/go-goph-keeper/internal/mocks"
	"github.com/wickedv43/go-goph-keeper/internal/storage"
	"go.uber.org/mock/gomock"
)

func TestService_User(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockDataKeeper(ctrl)
	s := &Service{storage: mockStorage}

	t.Run("successfully retrieves user by ID", func(t *testing.T) {
		expected := storage.User{ID: 42, Login: "jdoe", PasswordHash: "secret"}

		mockStorage.
			EXPECT().
			User(gomock.Any(), uint64(42)).
			Return(expected, nil)

		result, err := s.User(context.Background(), 42)
		require.NoError(t, err)
		require.Equal(t, expected, result)
	})

	t.Run("fails to retrieve user by ID", func(t *testing.T) {
		expectedErr := errors.New("user not found")

		mockStorage.
			EXPECT().
			User(gomock.Any(), uint64(99)).
			Return(storage.User{}, expectedErr)

		result, err := s.User(context.Background(), 99)
		require.Error(t, err)
		require.Equal(t, expectedErr, err)
		require.Equal(t, storage.User{}, result)
	})
}

func TestService_UserByLogin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockDataKeeper(ctrl)
	s := &Service{storage: mockStorage}

	t.Run("successfully retrieves user by login", func(t *testing.T) {
		expected := storage.User{ID: 10, Login: "alice", PasswordHash: "hash123"}

		mockStorage.
			EXPECT().
			UserByLogin(gomock.Any(), "alice").
			Return(expected, nil)

		result, err := s.UserByLogin(context.Background(), "alice")
		require.NoError(t, err)
		require.Equal(t, expected, result)
	})

	t.Run("fails to retrieve user by login", func(t *testing.T) {
		expectedErr := errors.New("login not found")

		mockStorage.
			EXPECT().
			UserByLogin(gomock.Any(), "missing").
			Return(storage.User{}, expectedErr)

		result, err := s.UserByLogin(context.Background(), "missing")
		require.Error(t, err)
		require.Equal(t, expectedErr, err)
		require.Equal(t, storage.User{}, result)
	})
}

func TestService_CreateVault(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockDataKeeper(ctrl)
	s := &Service{storage: mockStorage}

	vault := &storage.VaultRecord{
		UserID:        1,
		Type:          "note",
		Title:         "My Note",
		Metadata:      "some meta",
		EncryptedData: []byte("encrypted content"),
	}

	t.Run("successfully creates vault record", func(t *testing.T) {
		mockStorage.
			EXPECT().
			CreateVault(gomock.Any(), vault).
			Return(nil)

		err := s.CreateVault(context.Background(), vault)
		require.NoError(t, err)
	})

	t.Run("fails to create vault record", func(t *testing.T) {
		expectedErr := errors.New("db failure")

		mockStorage.
			EXPECT().
			CreateVault(gomock.Any(), vault).
			Return(expectedErr)

		err := s.CreateVault(context.Background(), vault)
		require.Error(t, err)
		require.Equal(t, expectedErr, err)
	})
}

func TestService_GetVault(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockDataKeeper(ctrl)
	s := &Service{storage: mockStorage}

	expected := storage.VaultRecord{
		ID:            42,
		UserID:        1,
		Type:          "note",
		Title:         "Test",
		Metadata:      "meta",
		EncryptedData: []byte("data"),
	}

	t.Run("successfully retrieves vault record", func(t *testing.T) {
		mockStorage.
			EXPECT().
			GetVault(gomock.Any(), expected.ID).
			Return(expected, nil)

		got, err := s.GetVault(context.Background(), expected.ID)
		require.NoError(t, err)
		require.Equal(t, expected, got)
	})

	t.Run("fails to retrieve vault record", func(t *testing.T) {
		expectedErr := errors.New("not found")

		mockStorage.
			EXPECT().
			GetVault(gomock.Any(), expected.ID).
			Return(storage.VaultRecord{}, expectedErr)

		_, err := s.GetVault(context.Background(), expected.ID)
		require.Error(t, err)
		require.Equal(t, expectedErr, err)
	})
}

func TestService_UpdateVault(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockDataKeeper(ctrl)
	s := &Service{storage: mockStorage}

	vault := &storage.VaultRecord{
		ID:            42,
		UserID:        1,
		Type:          "note",
		Title:         "Updated Title",
		Metadata:      "updated-meta",
		EncryptedData: []byte("new data"),
	}

	t.Run("successfully updates vault record", func(t *testing.T) {
		mockStorage.
			EXPECT().
			UpdateVault(gomock.Any(), vault).
			Return(nil)

		err := s.UpdateVault(context.Background(), vault)
		require.NoError(t, err)
	})

	t.Run("fails to update vault record", func(t *testing.T) {
		expectedErr := errors.New("update failed")

		mockStorage.
			EXPECT().
			UpdateVault(gomock.Any(), vault).
			Return(expectedErr)

		err := s.UpdateVault(context.Background(), vault)
		require.Error(t, err)
		require.Equal(t, expectedErr, err)
	})
}

func TestService_ListVaults(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockDataKeeper(ctrl)
	s := &Service{storage: mockStorage}

	userID := uint64(1)

	expectedRecords := []storage.VaultRecord{
		{ID: 1, UserID: userID, Type: "login", Title: "Email", Metadata: "gmail", EncryptedData: []byte("data1")},
		{ID: 2, UserID: userID, Type: "note", Title: "Secrets", Metadata: "notes", EncryptedData: []byte("data2")},
	}

	t.Run("successfully returns list of vaults", func(t *testing.T) {
		mockStorage.
			EXPECT().
			ListVaults(gomock.Any(), userID).
			Return(expectedRecords, nil)

		got, err := s.ListVaults(context.Background(), userID)
		require.NoError(t, err)
		require.Equal(t, expectedRecords, got)
	})

	t.Run("fails to return vaults", func(t *testing.T) {
		expectedErr := errors.New("storage error")

		mockStorage.
			EXPECT().
			ListVaults(gomock.Any(), userID).
			Return(nil, expectedErr)

		got, err := s.ListVaults(context.Background(), userID)
		require.Error(t, err)
		require.Nil(t, got)
		require.Equal(t, expectedErr, err)
	})
}

func TestService_DeleteVault(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockDataKeeper(ctrl)
	s := &Service{storage: mockStorage}

	vaultID := uint64(1)

	t.Run("successfully deletes vault", func(t *testing.T) {
		mockStorage.
			EXPECT().
			DeleteVault(gomock.Any(), vaultID).
			Return(nil)

		err := s.DeleteVault(context.Background(), vaultID)
		require.NoError(t, err)
	})

	t.Run("fails to delete vault", func(t *testing.T) {
		expectedErr := errors.New("delete failed")

		mockStorage.
			EXPECT().
			DeleteVault(gomock.Any(), vaultID).
			Return(expectedErr)

		err := s.DeleteVault(context.Background(), vaultID)
		require.Error(t, err)
		require.Equal(t, expectedErr, err)
	})
}

func TestService_NewUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockDataKeeper(ctrl)
	s := &Service{storage: mockStorage}

	expected := storage.User{ID: 1, Login: "test", PasswordHash: "hash"}
	input := &storage.User{Login: "test", PasswordHash: "hash"}

	t.Run("successfully creates a new user", func(t *testing.T) {
		mockStorage.
			EXPECT().
			NewUser(gomock.Any(), input).
			Return(expected, nil)

		got, err := s.NewUser(context.Background(), input)
		require.NoError(t, err)
		require.Equal(t, expected, got)
	})

	t.Run("fails to create new user", func(t *testing.T) {
		mockErr := errors.New("db error")

		mockStorage.
			EXPECT().
			NewUser(gomock.Any(), input).
			Return(storage.User{}, mockErr)

		got, err := s.NewUser(context.Background(), input)
		require.Error(t, err)
		require.Equal(t, mockErr, err)
		require.Equal(t, storage.User{}, got)
	})
}
