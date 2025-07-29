package main

import (
	"context"
	"strconv"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"github.com/wickedv43/go-goph-keeper/cmd/client/internal/kv"
	"github.com/wickedv43/go-goph-keeper/cmd/client/internal/mocks"
	pb "github.com/wickedv43/go-goph-keeper/internal/api"
	"github.com/wickedv43/go-goph-keeper/internal/config"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestVaultDeleteCMD(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockGophKeeperClient(ctrl)
	mockStorage := mocks.NewMockStorage(ctrl)

	gk := &GophKeeper{
		client:  mockClient,
		storage: mockStorage,
		rootCtx: context.Background(),
		cfg:     &config.Config{},
	}

	cmd := gk.VaultDeleteCMD()

	vaultID := uint64(123)

	// токен + авторизация
	mockStorage.EXPECT().GetCurrentToken().Return("token123", nil)
	mockClient.EXPECT().
		DeleteVault(gomock.Any(), &pb.DeleteVaultRequest{
			VaultId: vaultID,
		}).
		Return(&emptypb.Empty{}, nil)

	mockStorage.EXPECT().GetConfig().Return(kv.Config{Current: "testctx"}, nil).AnyTimes()

	// передаём фейковые аргументы (id как строка)
	args := []string{"delete", strconv.FormatUint(vaultID, 10)}
	cmd.SetArgs(args)

	err := cmd.RunE(cmd, args)
	require.NoError(t, err)
}

func TestLoginCMD_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockGophKeeperClient(ctrl)
	mockStorage := mocks.NewMockStorage(ctrl)

	gk := &GophKeeper{
		client:  mockClient,
		storage: mockStorage,
		rootCtx: context.Background(),
		cfg:     &config.Config{},
	}

	mockClient.EXPECT().
		Login(gomock.Any(), &pb.LoginRequest{
			Login:    "testuser",
			Password: gk.hashPassword("superpass"),
		}).
		Return(&pb.LoginResponse{Token: "token123"}, nil)

	mockStorage.EXPECT().
		SaveContext("testuser", "token123").
		Return(nil)

	mockStorage.EXPECT().
		GetCurrentKey().
		Return("already-there", nil)

	mockStorage.EXPECT().
		GetConfig().
		Return(kv.Config{Current: "testctx"}, nil).
		AnyTimes()

	cmd := gk.LoginCMD()

	args := []string{"testuser", "superpass"}
	cmd.SetArgs(args)
	err := cmd.RunE(cmd, args)
	require.NoError(t, err)
}
func TestLoginCMD_SaveContextError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockGophKeeperClient(ctrl)
	mockStorage := mocks.NewMockStorage(ctrl)

	gk := &GophKeeper{
		client:  mockClient,
		storage: mockStorage,
		rootCtx: context.Background(),
		cfg:     &config.Config{},
	}

	mockClient.EXPECT().
		Login(gomock.Any(), gomock.Any()).
		Return(&pb.LoginResponse{Token: "token123"}, nil)

	mockStorage.EXPECT().
		SaveContext("testuser", "token123").
		Return(errors.New("cannot save context"))

	cmd := gk.LoginCMD()
	cmd.SetArgs([]string{"testuser", "superpass"})

	err := cmd.RunE(cmd, []string{"testuser", "superpass"})
	require.ErrorContains(t, err, "cannot save context")
}

func TestLoginCMD_LoginError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockGophKeeperClient(ctrl)

	gk := &GophKeeper{
		client:  mockClient,
		storage: mocks.NewMockStorage(ctrl),
		rootCtx: context.Background(),
		cfg:     &config.Config{},
	}

	mockClient.EXPECT().
		Login(gomock.Any(), &pb.LoginRequest{
			Login:    "testuser",
			Password: gk.hashPassword("badpass"),
		}).
		Return(nil, errors.New("unauthorized"))

	cmd := gk.LoginCMD()
	cmd.SetArgs([]string{"testuser", "badpass"})

	err := cmd.RunE(cmd, []string{"testuser", "badpass"})
	require.ErrorContains(t, err, "unauthorized")
}

func TestRegisterCMD(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockGophKeeperClient(ctrl)
	mockStorage := mocks.NewMockStorage(ctrl)

	gk := &GophKeeper{
		client:  mockClient,
		storage: mockStorage,
		rootCtx: context.Background(),
		cfg:     &config.Config{},
	}

	mockClient.EXPECT().
		Register(gomock.Any(), &pb.RegisterRequest{
			Login:    "testuser",
			Password: gk.hashPassword("superpass"),
		}).
		Return(&pb.RegisterResponse{}, nil)

	mockStorage.EXPECT().
		SaveKey("testuser", gomock.Any()).
		Return(nil)

	mockStorage.EXPECT().
		GetConfig().
		Return(kv.Config{Current: "testctx"}, nil).
		AnyTimes()

	cmd := gk.RegisterCMD()

	args := []string{"testuser", "superpass"}
	cmd.SetArgs(args)
	err := cmd.RunE(cmd, args)
	require.NoError(t, err)
}
