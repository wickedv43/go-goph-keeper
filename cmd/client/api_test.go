package main

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/wickedv43/go-goph-keeper/internal/config"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/wickedv43/go-goph-keeper/cmd/client/internal/mocks"
	pb "github.com/wickedv43/go-goph-keeper/internal/api"
)

func TestGophKeeper_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	login := "testuser"
	password := "pass123"

	t.Run("success", func(t *testing.T) {
		mockClient := mocks.NewMockGophKeeperClient(ctrl)
		mockStorage := mocks.NewMockStorage(ctrl)
		gk := &GophKeeper{
			client:  mockClient,
			storage: mockStorage,
			cfg:     &config.Config{Master: "test"},
			rootCtx: context.Background(),
		}

		expectedToken := "some-token"

		mockClient.EXPECT().
			Login(gomock.Any(), &pb.LoginRequest{
				Login:    login,
				Password: gk.hashPassword(password),
			}).
			Return(&pb.LoginResponse{Token: expectedToken}, nil)

		mockStorage.EXPECT().
			SaveContext(login, expectedToken).
			Return(nil)

		err := gk.Login(login, password)
		require.NoError(t, err)
	})

	t.Run("client error", func(t *testing.T) {
		mockClient := mocks.NewMockGophKeeperClient(ctrl)
		mockStorage := mocks.NewMockStorage(ctrl)
		gk := &GophKeeper{
			client:  mockClient,
			storage: mockStorage,
			cfg:     &config.Config{Master: "test"},
			rootCtx: context.Background(),
		}

		mockClient.EXPECT().
			Login(gomock.Any(), gomock.Any()).
			Return(nil, errors.New("login failed"))

		err := gk.Login(login, password)
		require.ErrorContains(t, err, "login failed")
	})

	t.Run("save context error", func(t *testing.T) {
		mockClient := mocks.NewMockGophKeeperClient(ctrl)
		mockStorage := mocks.NewMockStorage(ctrl)
		gk := &GophKeeper{
			client:  mockClient,
			storage: mockStorage,
			cfg:     &config.Config{Master: "test"},
			rootCtx: context.Background(),
		}

		expectedToken := "some-token"

		mockClient.EXPECT().
			Login(gomock.Any(), gomock.Any()).
			Return(&pb.LoginResponse{Token: expectedToken}, nil)

		mockStorage.EXPECT().
			SaveContext(login, expectedToken).
			Return(errors.New("save error"))

		err := gk.Login(login, password)
		require.ErrorContains(t, err, "save error")
	})
}

func TestGophKeeper_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	login := "newuser"
	password := "pass123"

	t.Run("success", func(t *testing.T) {
		mockClient := mocks.NewMockGophKeeperClient(ctrl)
		mockStorage := mocks.NewMockStorage(ctrl)
		gk := &GophKeeper{
			client:  mockClient,
			storage: mockStorage,
			cfg:     &config.Config{Master: "test"},
			rootCtx: context.Background(),
		}

		mockClient.EXPECT().
			Register(gomock.Any(), gomock.Any()).
			Return(&pb.RegisterResponse{}, nil)

		mockStorage.EXPECT().
			SaveKey(gomock.Eq(login), gomock.Any()).
			Return(nil)

		mnemonic, err := gk.Register(login, password)
		require.NoError(t, err)
		require.Len(t, mnemonic, 12)
	})

	t.Run("client error", func(t *testing.T) {
		mockClient := mocks.NewMockGophKeeperClient(ctrl)
		mockStorage := mocks.NewMockStorage(ctrl)
		gk := &GophKeeper{
			client:  mockClient,
			storage: mockStorage,
			cfg:     &config.Config{Master: "test"},
			rootCtx: context.Background(),
		}

		mockClient.EXPECT().
			Register(gomock.Any(), gomock.Any()).
			Return(nil, errors.New("register fail"))

		_, err := gk.Register(login, password)
		require.ErrorContains(t, err, "register fail")
	})

	t.Run("save key error", func(t *testing.T) {
		mockClient := mocks.NewMockGophKeeperClient(ctrl)
		mockStorage := mocks.NewMockStorage(ctrl)
		gk := &GophKeeper{
			client:  mockClient,
			storage: mockStorage,
			cfg:     &config.Config{Master: "test"},
			rootCtx: context.Background(),
		}

		mockClient.EXPECT().
			Register(gomock.Any(), gomock.Any()).
			Return(&pb.RegisterResponse{}, nil)

		mockStorage.EXPECT().
			SaveKey(gomock.Eq(login), gomock.Any()).
			Return(errors.New("save failed"))

		_, err := gk.Register(login, password)
		require.ErrorContains(t, err, "save failed")
	})
}

func TestGophKeeper_VaultList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("success with auth token", func(t *testing.T) {
		mockClient := mocks.NewMockGophKeeperClient(ctrl)
		mockStorage := mocks.NewMockStorage(ctrl)

		expectedToken := "test-token"
		expectedResp := &pb.ListVaultsResponse{
			Vaults: []*pb.VaultRecord{
				{Id: 1, Type: "note", Title: "Note 1"},
			},
		}

		gk := &GophKeeper{
			client:  mockClient,
			storage: mockStorage,
			cfg:     &config.Config{},
			rootCtx: context.Background(),
		}

		mockStorage.EXPECT().
			GetCurrentToken().
			Return(expectedToken, nil)

		mockClient.EXPECT().
			ListVaults(gomock.Any(), &pb.ListVaultsRequest{}).
			DoAndReturn(func(ctx context.Context, req *pb.ListVaultsRequest, _ ...grpc.CallOption) (*pb.ListVaultsResponse, error) {
				md, ok := metadata.FromOutgoingContext(ctx)
				require.True(t, ok)
				require.Equal(t, []string{"Bearer " + expectedToken}, md["authorization"])
				return expectedResp, nil
			})

		resp, err := gk.VaultList()
		require.NoError(t, err)
		require.Equal(t, expectedResp, resp)
	})

	t.Run("storage error falls back to rootCtx", func(t *testing.T) {
		mockClient := mocks.NewMockGophKeeperClient(ctrl)
		mockStorage := mocks.NewMockStorage(ctrl)

		gk := &GophKeeper{
			client:  mockClient,
			storage: mockStorage,
			cfg:     &config.Config{},
			rootCtx: context.WithValue(context.Background(), "key", "value"),
		}

		mockStorage.EXPECT().
			GetCurrentToken().
			Return("", errors.New("no token"))

		mockClient.EXPECT().
			ListVaults(gomock.Any(), &pb.ListVaultsRequest{}).
			DoAndReturn(func(ctx context.Context, req *pb.ListVaultsRequest, _ ...grpc.CallOption) (*pb.ListVaultsResponse, error) {
				// Здесь нет authorization
				md, _ := metadata.FromOutgoingContext(ctx)
				require.Nil(t, md)
				return &pb.ListVaultsResponse{}, nil
			})

		_, err := gk.VaultList()
		require.NoError(t, err)
	})
}

func TestGophKeeper_VaultCreate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("success with token", func(t *testing.T) {
		mockClient := mocks.NewMockGophKeeperClient(ctrl)
		mockStorage := mocks.NewMockStorage(ctrl)

		expectedToken := "secure-token"
		testRecord := &pb.VaultRecord{
			Id:    123,
			Type:  "login",
			Title: "GitHub",
		}

		gk := &GophKeeper{
			client:  mockClient,
			storage: mockStorage,
			cfg:     &config.Config{},
			rootCtx: context.Background(),
		}

		mockStorage.EXPECT().
			GetCurrentToken().
			Return(expectedToken, nil)

		mockClient.EXPECT().
			CreateVault(gomock.Any(), &pb.CreateVaultRequest{
				Record: testRecord,
			}).
			DoAndReturn(func(ctx context.Context, req *pb.CreateVaultRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
				md, ok := metadata.FromOutgoingContext(ctx)
				require.True(t, ok)
				require.Equal(t, []string{"Bearer " + expectedToken}, md["authorization"])
				return &emptypb.Empty{}, nil
			})

		resp, err := gk.VaultCreate(testRecord)
		require.NoError(t, err)
		require.NotNil(t, resp)
	})

	t.Run("fallback to rootCtx on token error", func(t *testing.T) {
		mockClient := mocks.NewMockGophKeeperClient(ctrl)
		mockStorage := mocks.NewMockStorage(ctrl)

		gk := &GophKeeper{
			client:  mockClient,
			storage: mockStorage,
			cfg:     &config.Config{},
			rootCtx: context.Background(),
		}

		testRecord := &pb.VaultRecord{
			Id:    456,
			Type:  "note",
			Title: "Secret Note",
		}

		mockStorage.EXPECT().
			GetCurrentToken().
			Return("", errors.New("missing"))

		mockClient.EXPECT().
			CreateVault(gomock.Any(), &pb.CreateVaultRequest{
				Record: testRecord,
			}).
			DoAndReturn(func(ctx context.Context, req *pb.CreateVaultRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
				md, _ := metadata.FromOutgoingContext(ctx)
				require.Nil(t, md)
				return &emptypb.Empty{}, nil
			})

		_, err := gk.VaultCreate(testRecord)
		require.NoError(t, err)
	})
}

func TestGophKeeper_VaultGet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("success with token", func(t *testing.T) {
		mockClient := mocks.NewMockGophKeeperClient(ctrl)
		mockStorage := mocks.NewMockStorage(ctrl)

		expectedToken := "secure-token"
		vaultID := uint64(42)
		expectedRecord := &pb.VaultRecord{
			Id:    vaultID,
			Title: "My Vault Record",
			Type:  "note",
		}

		gk := &GophKeeper{
			client:  mockClient,
			storage: mockStorage,
			cfg:     &config.Config{},
			rootCtx: context.Background(),
		}

		mockStorage.EXPECT().
			GetCurrentToken().
			Return(expectedToken, nil)

		mockClient.EXPECT().
			GetVault(gomock.Any(), &pb.GetVaultRequest{
				VaultId: vaultID,
			}).
			DoAndReturn(func(ctx context.Context, req *pb.GetVaultRequest, _ ...grpc.CallOption) (*pb.VaultRecord, error) {
				md, ok := metadata.FromOutgoingContext(ctx)
				require.True(t, ok)
				require.Equal(t, []string{"Bearer " + expectedToken}, md["authorization"])
				return expectedRecord, nil
			})

		record, err := gk.VaultGet(vaultID)
		require.NoError(t, err)
		require.Equal(t, expectedRecord, record)
	})

	t.Run("fallback to rootCtx", func(t *testing.T) {
		mockClient := mocks.NewMockGophKeeperClient(ctrl)
		mockStorage := mocks.NewMockStorage(ctrl)

		gk := &GophKeeper{
			client:  mockClient,
			storage: mockStorage,
			cfg:     &config.Config{},
			rootCtx: context.Background(),
		}

		vaultID := uint64(99)
		expectedRecord := &pb.VaultRecord{Id: vaultID}

		mockStorage.EXPECT().
			GetCurrentToken().
			Return("", errors.New("no token"))

		mockClient.EXPECT().
			GetVault(gomock.Any(), &pb.GetVaultRequest{
				VaultId: vaultID,
			}).
			DoAndReturn(func(ctx context.Context, req *pb.GetVaultRequest, _ ...grpc.CallOption) (*pb.VaultRecord, error) {
				md, _ := metadata.FromOutgoingContext(ctx)
				require.Nil(t, md)
				return expectedRecord, nil
			})

		record, err := gk.VaultGet(vaultID)
		require.NoError(t, err)
		require.Equal(t, expectedRecord, record)
	})
}

func TestGophKeeper_VaultDelete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("success with token", func(t *testing.T) {
		mockClient := mocks.NewMockGophKeeperClient(ctrl)
		mockStorage := mocks.NewMockStorage(ctrl)

		expectedToken := "secure-token"
		vaultID := uint64(777)

		gk := &GophKeeper{
			client:  mockClient,
			storage: mockStorage,
			cfg:     &config.Config{},
			rootCtx: context.Background(),
		}

		mockStorage.EXPECT().
			GetCurrentToken().
			Return(expectedToken, nil)

		mockClient.EXPECT().
			DeleteVault(gomock.Any(), &pb.DeleteVaultRequest{
				VaultId: vaultID,
			}).
			DoAndReturn(func(ctx context.Context, req *pb.DeleteVaultRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
				md, ok := metadata.FromOutgoingContext(ctx)
				require.True(t, ok)
				require.Equal(t, []string{"Bearer " + expectedToken}, md["authorization"])
				return &emptypb.Empty{}, nil
			})

		resp, err := gk.VaultDelete(vaultID)
		require.NoError(t, err)
		require.NotNil(t, resp)
	})

	t.Run("fallback to rootCtx", func(t *testing.T) {
		mockClient := mocks.NewMockGophKeeperClient(ctrl)
		mockStorage := mocks.NewMockStorage(ctrl)

		gk := &GophKeeper{
			client:  mockClient,
			storage: mockStorage,
			cfg:     &config.Config{},
			rootCtx: context.Background(),
		}

		vaultID := uint64(888)

		mockStorage.EXPECT().
			GetCurrentToken().
			Return("", errors.New("no token"))

		mockClient.EXPECT().
			DeleteVault(gomock.Any(), &pb.DeleteVaultRequest{
				VaultId: vaultID,
			}).
			DoAndReturn(func(ctx context.Context, req *pb.DeleteVaultRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
				md, _ := metadata.FromOutgoingContext(ctx)
				require.Nil(t, md)
				return &emptypb.Empty{}, nil
			})

		resp, err := gk.VaultDelete(vaultID)
		require.NoError(t, err)
		require.NotNil(t, resp)
	})
}
