package server

import (
	"context"
	"testing"

	"github.com/pkg/errors"
	"github.com/samber/do/v2"
	"github.com/stretchr/testify/require"
	"github.com/wickedv43/go-goph-keeper/internal/api"
	pb "github.com/wickedv43/go-goph-keeper/internal/api"
	"github.com/wickedv43/go-goph-keeper/internal/logger"
	"github.com/wickedv43/go-goph-keeper/internal/mocks"
	"github.com/wickedv43/go-goph-keeper/internal/storage"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestServer_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// зависимости
	mockService := mocks.NewMockGophKeeper(ctrl)
	log, err := logger.NewLogger(do.New())
	require.NoError(t, err)

	s := &Server{
		service: mockService,
		log:     log.SugaredLogger,
	}

	t.Run("success: user created", func(t *testing.T) {
		mockService.
			EXPECT().
			NewUser(gomock.Any(), gomock.Any()).
			Return(storage.User{ID: 42, Login: "test"}, nil)

		req := &api.RegisterRequest{
			Login:    "test",
			Password: "pass123",
		}

		resp, err := s.Register(context.Background(), req)
		require.NoError(t, err)
		require.Equal(t, uint64(42), resp.UserId)
	})

	t.Run("error: login already used", func(t *testing.T) {
		mockService.
			EXPECT().
			NewUser(gomock.Any(), gomock.Any()).
			Return(storage.User{}, storage.ErrLoginUsed)

		req := &api.RegisterRequest{
			Login:    "test",
			Password: "pass123",
		}

		resp, err := s.Register(context.Background(), req)
		require.ErrorIs(t, err, storage.ErrLoginUsed)
		require.Nil(t, resp)
	})

	t.Run("error: internal error", func(t *testing.T) {
		mockService.
			EXPECT().
			NewUser(gomock.Any(), gomock.Any()).
			Return(storage.User{}, errors.New("db down"))

		req := &api.RegisterRequest{
			Login:    "test",
			Password: "pass123",
		}

		resp, err := s.Register(context.Background(), req)
		require.Error(t, err)
		require.Contains(t, err.Error(), "db down")
		require.Nil(t, resp)
	})
}

func TestServer_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockGophKeeper(ctrl)
	log := zap.NewNop().Sugar()

	s := &Server{
		service: mockService,
		log:     log,
	}

	t.Run("success: valid login and password", func(t *testing.T) {
		mockService.
			EXPECT().
			UserByLogin(gomock.Any(), "tester").
			Return(storage.User{ID: 101, Login: "tester", PasswordHash: "qwerty"}, nil)

		req := &pb.LoginRequest{
			Login:    "tester",
			Password: "qwerty",
		}

		resp, err := s.Login(context.Background(), req)
		require.NoError(t, err)
		require.NotEmpty(t, resp.Token)
	})

	t.Run("error: user not found", func(t *testing.T) {
		mockService.
			EXPECT().
			UserByLogin(gomock.Any(), "unknown").
			Return(storage.User{}, errors.New("not found"))

		req := &pb.LoginRequest{
			Login:    "unknown",
			Password: "pass",
		}

		resp, err := s.Login(context.Background(), req)
		require.Error(t, err)
		require.Nil(t, resp)
		st, _ := status.FromError(err)
		require.Equal(t, codes.NotFound, st.Code())
	})

	t.Run("error: wrong password", func(t *testing.T) {
		mockService.
			EXPECT().
			UserByLogin(gomock.Any(), "tester").
			Return(storage.User{ID: 101, Login: "tester", PasswordHash: "correct"}, nil)

		req := &pb.LoginRequest{
			Login:    "tester",
			Password: "wrong",
		}

		resp, err := s.Login(context.Background(), req)
		require.Error(t, err)
		require.Nil(t, resp)
		st, _ := status.FromError(err)
		require.Equal(t, codes.Unauthenticated, st.Code())
	})
}

func TestServer_CreateVault(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockGophKeeper(ctrl)
	log := zap.NewNop().Sugar()

	s := &Server{
		service: mockService,
		log:     log,
	}

	t.Run("success: vault created", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), userIDKey, uint64(42))

		mockService.
			EXPECT().
			CreateVault(gomock.Any(), &storage.VaultRecord{
				UserID:        42,
				Type:          "note",
				Title:         "My Note",
				Metadata:      "meta",
				EncryptedData: []byte("secret"),
			}).
			Return(nil)

		req := &pb.CreateVaultRequest{
			Record: &pb.VaultRecord{
				Type:          "note",
				Title:         "My Note",
				Metadata:      "meta",
				EncryptedData: []byte("secret"),
			},
		}

		resp, err := s.CreateVault(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
	})

	t.Run("error: unauthenticated", func(t *testing.T) {
		req := &pb.CreateVaultRequest{
			Record: &pb.VaultRecord{
				Type: "login",
			},
		}

		resp, err := s.CreateVault(context.Background(), req)
		require.Error(t, err)
		require.Nil(t, resp)

		st, _ := status.FromError(err)
		require.Equal(t, codes.Unauthenticated, st.Code())
	})

	t.Run("error: internal failure", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), userIDKey, uint64(99))

		mockService.
			EXPECT().
			CreateVault(gomock.Any(), gomock.Any()).
			Return(errors.New("db error"))

		req := &pb.CreateVaultRequest{
			Record: &pb.VaultRecord{
				Type:          "card",
				Title:         "My Card",
				Metadata:      "",
				EncryptedData: []byte("1234"),
			},
		}

		resp, err := s.CreateVault(ctx, req)
		require.Error(t, err)
		require.Nil(t, resp)

		st, _ := status.FromError(err)
		require.Equal(t, codes.Internal, st.Code())
	})
}

func TestServer_GetVault(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockGophKeeper(ctrl)
	log := zap.NewNop().Sugar()

	s := &Server{
		service: mockService,
		log:     log,
	}

	t.Run("success: vault found", func(t *testing.T) {
		mockService.
			EXPECT().
			GetVault(gomock.Any(), uint64(1)).
			Return(storage.VaultRecord{
				ID:            1,
				UserID:        42,
				Type:          "note",
				Title:         "Test note",
				Metadata:      "meta",
				EncryptedData: []byte("secret"),
			}, nil)

		req := &pb.GetVaultRequest{VaultId: 1}
		resp, err := s.GetVault(context.Background(), req)
		require.NoError(t, err)
		require.Equal(t, uint64(1), resp.Id)
		require.Equal(t, "note", resp.Type)
		require.Equal(t, "Test note", resp.Title)
		require.Equal(t, "meta", resp.Metadata)
		require.Equal(t, []byte("secret"), resp.EncryptedData)
	})

	t.Run("error: vault not found", func(t *testing.T) {
		mockService.
			EXPECT().
			GetVault(gomock.Any(), uint64(999)).
			Return(storage.VaultRecord{}, errors.New("not found"))

		req := &pb.GetVaultRequest{VaultId: 999}
		resp, err := s.GetVault(context.Background(), req)
		require.Error(t, err)
		require.Nil(t, resp)

		st, _ := status.FromError(err)
		require.Equal(t, codes.NotFound, st.Code())
		require.Contains(t, st.Message(), "запись не найдена")
	})
}

func TestServer_UpdateVault(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockGophKeeper(ctrl)
	log := zap.NewNop().Sugar()

	s := &Server{
		service: mockService,
		log:     log,
	}

	t.Run("success: vault updated", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), userIDKey, uint64(42))

		mockService.
			EXPECT().
			UpdateVault(gomock.Any(), &storage.VaultRecord{
				ID:            1,
				UserID:        42,
				Type:          "login",
				Title:         "Email",
				Metadata:      "gmail",
				EncryptedData: []byte("abc123"),
			}).
			Return(nil)

		req := &pb.VaultRecord{
			Id:            1,
			Type:          "login",
			Title:         "Email",
			Metadata:      "gmail",
			EncryptedData: []byte("abc123"),
		}

		resp, err := s.UpdateVault(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
	})

	t.Run("error: unauthenticated", func(t *testing.T) {
		req := &pb.VaultRecord{Id: 1}

		resp, err := s.UpdateVault(context.Background(), req)
		require.Error(t, err)
		require.Nil(t, resp)

		st, _ := status.FromError(err)
		require.Equal(t, codes.Unauthenticated, st.Code())
	})

	t.Run("error: service failure", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), userIDKey, uint64(1))

		mockService.
			EXPECT().
			UpdateVault(gomock.Any(), gomock.Any()).
			Return(errors.New("update failed"))

		req := &pb.VaultRecord{
			Id:            2,
			Type:          "note",
			Title:         "Note 2",
			Metadata:      "",
			EncryptedData: []byte("xyz"),
		}

		resp, err := s.UpdateVault(ctx, req)
		require.Error(t, err)
		require.Nil(t, resp)

		st, _ := status.FromError(err)
		require.Equal(t, codes.Internal, st.Code())
		require.Contains(t, st.Message(), "не удалось обновить запись")
	})
}

func TestServer_DeleteVault(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockGophKeeper(ctrl)
	log := zap.NewNop().Sugar()

	s := &Server{
		service: mockService,
		log:     log,
	}

	t.Run("success: deleted", func(t *testing.T) {
		mockService.
			EXPECT().
			DeleteVault(gomock.Any(), uint64(10)).
			Return(nil)

		req := &pb.DeleteVaultRequest{VaultId: 10}

		resp, err := s.DeleteVault(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, resp)
	})

	t.Run("error: failed to delete", func(t *testing.T) {
		mockService.
			EXPECT().
			DeleteVault(gomock.Any(), uint64(999)).
			Return(errors.New("db error"))

		req := &pb.DeleteVaultRequest{VaultId: 999}

		resp, err := s.DeleteVault(context.Background(), req)
		require.Error(t, err)
		require.Nil(t, resp)

		st, _ := status.FromError(err)
		require.Equal(t, codes.Internal, st.Code())
		require.Contains(t, st.Message(), "не удалось удалить запись")
	})
}

func TestServer_ListVaults(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockGophKeeper(ctrl)
	log := zap.NewNop().Sugar()

	s := &Server{
		service: mockService,
		log:     log,
	}

	t.Run("success: vaults returned", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), userIDKey, uint64(42))

		mockService.
			EXPECT().
			ListVaults(gomock.Any(), uint64(42)).
			Return([]storage.VaultRecord{
				{ID: 1, UserID: 42, Type: "note", Title: "Note1", Metadata: "", EncryptedData: []byte("123")},
				{ID: 2, UserID: 42, Type: "login", Title: "Login1", Metadata: "m", EncryptedData: []byte("456")},
			}, nil)

		resp, err := s.ListVaults(ctx, &pb.ListVaultsRequest{})
		require.NoError(t, err)
		require.Len(t, resp.Vaults, 2)
		require.Equal(t, "note", resp.Vaults[0].Type)
		require.Equal(t, "Login1", resp.Vaults[1].Title)
	})

	t.Run("error: unauthenticated", func(t *testing.T) {
		resp, err := s.ListVaults(context.Background(), &pb.ListVaultsRequest{})
		require.Error(t, err)
		require.Nil(t, resp)

		st, _ := status.FromError(err)
		require.Equal(t, codes.Unauthenticated, st.Code())
	})

	t.Run("error: service failure", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), userIDKey, uint64(99))

		mockService.
			EXPECT().
			ListVaults(gomock.Any(), uint64(99)).
			Return(nil, errors.New("db error"))

		resp, err := s.ListVaults(ctx, &pb.ListVaultsRequest{})
		require.Error(t, err)
		require.Nil(t, resp)

		st, _ := status.FromError(err)
		require.Equal(t, codes.Internal, st.Code())
		require.Contains(t, st.Message(), "не удалось получить список")
	})
}
