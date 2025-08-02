package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

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

func TestVaultListCMD(t *testing.T) {
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

	// Конфиг и токен
	mockStorage.EXPECT().
		GetCurrentToken().
		Return("token123", nil)

	mockStorage.EXPECT().
		GetConfig().
		Return(kv.Config{Current: "testctx"}, nil).
		AnyTimes()

	mockClient.EXPECT().
		ListVaults(gomock.Any(), gomock.Any()).
		Return(&pb.ListVaultsResponse{
			Vaults: []*pb.VaultRecord{
				{Id: 1, Title: "Test1", Type: "note"},
				{Id: 2, Title: "Test2", Type: "login"},
			},
		}, nil)

	cmd := gk.VaultListCMD()
	cmd.SetArgs([]string{"list"})
	err := cmd.RunE(cmd, []string{"list"})
	require.NoError(t, err)
}

func TestNewVaultCMD_Login(t *testing.T) {
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

	//mock for shellLoop()
	mockStorage.EXPECT().
		GetConfig().
		Return(kv.Config{Current: "testctx"}, nil).
		AnyTimes()

	// Ожидаемый вызов получения ключа
	mockStorage.EXPECT().
		GetCurrentKey().
		Return("6368616e676520746869732070617373", nil)

	// Ожидаемый вызов получения ключа
	mockStorage.EXPECT().
		GetCurrentToken().
		Return("6368616e676520746869732070617373", nil)

	// Ожидаемый вызов VaultCreate
	args := []string{"My Login Title", "login", "testdata"} //rec type rec title

	mockClient.EXPECT().
		CreateVault(gomock.Any(), gomock.Any()).
		Return(&emptypb.Empty{}, nil)

	cmd := gk.NewVaultCMD()
	err := cmd.RunE(cmd, args)
	require.NoError(t, err)
}

func TestVaultListCMD_Success(t *testing.T) {
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

	mockStorage.EXPECT().
		GetConfig().
		Return(kv.Config{Current: "testctx"}, nil).
		AnyTimes()

	// Ожидаемый вызов получения ключа
	mockStorage.EXPECT().
		GetCurrentToken().
		Return("6368616e676520746869732070617373", nil)

	// Подмена stdout для проверки вывода
	var buf bytes.Buffer
	cmd := gk.VaultListCMD()
	cmd.SetOut(&buf)

	now := time.Now().Format(time.RFC3339)
	mockClient.EXPECT().
		ListVaults(gomock.Any(), gomock.Any()).
		Return(&pb.ListVaultsResponse{
			Vaults: []*pb.VaultRecord{
				{
					Id:        1,
					Type:      "note",
					Title:     "Test Note",
					UpdatedAt: now,
					Metadata:  `{"tag":"important"}`,
				},
			},
		}, nil)

	err := cmd.RunE(cmd, []string{})
	require.NoError(t, err)

	output := buf.String()
	require.Contains(t, output, "Test Note")
	require.Contains(t, output, "note")
	require.Contains(t, output, "important")
}

func TestVaultLoginPass(t *testing.T) {
	// Создаём pipe
	r, w, _ := os.Pipe()
	// Сохраняем оригинальный Stdin
	origStdin := os.Stdin
	// Подменяем Stdin
	os.Stdin = r
	defer func() {
		os.Stdin = origStdin
	}()

	// Пишем в pipe то, что будет "введено" пользователем
	go func() {
		fmt.Fprintln(w, "mylogin") // Имитация ввода логина
		fmt.Fprintln(w, "mypass")  // Имитация ввода пароля
	}()

	v := &pb.VaultRecord{}
	res, err := vaultLoginPass(v)
	require.NoError(t, err)

	var data kv.LoginPass
	err = json.Unmarshal(res.EncryptedData, &data)
	require.NoError(t, err)
	require.Equal(t, "mylogin", data.Login)
	require.Equal(t, "mypass", data.Password)
}

func TestVaultNote(t *testing.T) {
	// Создаём pipe
	r, w, _ := os.Pipe()
	// Сохраняем оригинальный Stdin
	origStdin := os.Stdin
	// Подменяем Stdin
	os.Stdin = r
	defer func() {
		os.Stdin = origStdin
	}()

	// Пишем в pipe то, что будет "введено" пользователем
	go func() {
		fmt.Fprintln(w, "testtext")
	}()

	v := &pb.VaultRecord{}
	res, err := vaultNote(v)
	require.NoError(t, err)

	var data kv.Note
	err = json.Unmarshal(res.EncryptedData, &data)
	require.NoError(t, err)
	require.Equal(t, "testtext", data.Text)

}

func TestVaultCard(t *testing.T) {
	// Создаём pipe
	r, w, _ := os.Pipe()
	// Сохраняем оригинальный Stdin
	origStdin := os.Stdin
	// Подменяем Stdin
	os.Stdin = r
	defer func() {
		os.Stdin = origStdin
	}()

	// Пишем в pipe то, что будет "введено" пользователем
	go func() {
		fmt.Fprintln(w, "number")
		fmt.Fprintln(w, "date")
		fmt.Fprintln(w, "cvv")
	}()

	v := &pb.VaultRecord{}
	res, err := vaultCard(v)
	require.NoError(t, err)

	var data kv.Card
	err = json.Unmarshal(res.EncryptedData, &data)
	require.NoError(t, err)
	require.Equal(t, "number", data.Number)
	require.Equal(t, "date", data.Date)
	require.Equal(t, "cvv", data.CVV)

}
