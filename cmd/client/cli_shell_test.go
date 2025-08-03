package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/wickedv43/go-goph-keeper/cmd/client/internal/crypto"
	"github.com/wickedv43/go-goph-keeper/cmd/client/internal/kv"
	"github.com/wickedv43/go-goph-keeper/cmd/client/internal/mocks"
	pb "github.com/wickedv43/go-goph-keeper/internal/api"
	"github.com/wickedv43/go-goph-keeper/internal/config"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestGophKeeper_ShellCMD_Success(t *testing.T) {
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

	// Подмена stdout для проверки вывода
	var buf bytes.Buffer
	cmd := gk.ShellCMD()
	cmd.SetOut(&buf)

	err := cmd.RunE(cmd, []string{})
	require.NoError(t, err)
}

func TestProcessShellCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockGophKeeperClient(ctrl)
	mockStorage := mocks.NewMockStorage(ctrl)

	gk := &GophKeeper{
		rootCmd: &cobra.Command{}, // пустышка, чтобы не nil
		client:  mockClient,
		storage: mockStorage,
		rootCtx: context.Background(),
		cfg:     &config.Config{},
	}

	// Создаём pipe
	r, w, _ := os.Pipe()

	origStdin := os.Stdin

	os.Stdin = r
	defer func() {
		os.Stdin = origStdin
	}()

	t.Run("processShellCommand_login", func(t *testing.T) {
		go func() {
			fmt.Fprintln(w, "login")
			fmt.Fprintln(w, "pass")
		}()

		mockClient.EXPECT().
			Login(gomock.Any(), gomock.Any()).
			Return(&pb.LoginResponse{}, errors.New("bad password"))

		err := gk.processShellCommand([]string{"login"})
		require.Error(t, err)
	})

	t.Run("processShellCommand_register", func(t *testing.T) {
		go func() {
			fmt.Fprintln(w, "login")
			fmt.Fprintln(w, "pass")
		}()

		mockClient.EXPECT().
			Register(gomock.Any(), gomock.Any()).
			Return(&pb.RegisterResponse{}, errors.New("login already exists"))

		err := gk.processShellCommand([]string{"register"})
		require.Error(t, err)
	})

	t.Run("processShellCommand_contexts", func(t *testing.T) {
		mockStorage.EXPECT().GetConfig().Return(kv.Config{}, errors.New("get cfg err")).AnyTimes()

		err := gk.processShellCommand([]string{"contexts"})
		require.Error(t, err)
	})

	t.Run("processShellCommand_use", func(t *testing.T) {
		cfg := kv.Config{
			Current:  "default",
			Contexts: map[string]kv.Context{"default": {}, "work": {}},
		}

		mockStorage.EXPECT().GetConfig().Return(cfg, errors.New("контекст не найден")).AnyTimes()

		err := gk.processShellCommand([]string{"use"})
		require.Error(t, err)
	})

	t.Run("processShellCommand_list", func(t *testing.T) {
		// Ожидаемый вызов получения ключа
		mockStorage.EXPECT().
			GetCurrentKey().
			Return("6368616e676520746869732070617373", nil).AnyTimes()

		mockClient.EXPECT().
			ListVaults(gomock.Any(), gomock.Any()).
			Return(&pb.ListVaultsResponse{
				Vaults: []*pb.VaultRecord{},
			}, nil)

		// Ожидаемый вызов получения ключа
		mockStorage.EXPECT().
			GetCurrentToken().
			Return("6368616e676520746869732070617373", nil)

		err := gk.processShellCommand([]string{"list"})
		require.NoError(t, err)
	})

	t.Run("processShellCommand_create", func(t *testing.T) {
		go func() {
			fmt.Fprintln(w, "TestTitle")
			fmt.Fprintln(w, "login")
			fmt.Fprintln(w, "log")
			fmt.Fprintln(w, "pass")
		}()

		mockClient.EXPECT().
			CreateVault(gomock.Any(), gomock.Any()).
			Return(&emptypb.Empty{}, nil)

		// Ожидаемый вызов получения ключа
		mockStorage.EXPECT().
			GetCurrentKey().
			Return("6368616e676520746869732070617373", nil).AnyTimes()

		// Ожидаемый вызов получения ключа
		mockStorage.EXPECT().
			GetCurrentToken().
			Return("6368616e676520746869732070617373", nil).AnyTimes()

		err := gk.processShellCommand([]string{"create"})
		require.NoError(t, err)
	})

	t.Run("processShellCommand_get", func(t *testing.T) {
		key := "6368616e676520746869732070617373"

		var note kv.Note
		note.Text = "text note"
		data, err := json.Marshal(note)
		require.NoError(t, err)

		crypted, err := crypto.EncryptWithSeed(data, key)
		require.NoError(t, err)

		mockClient.EXPECT().GetVault(gomock.Any(), gomock.Any()).Return(&pb.VaultRecord{
			Id:            1,
			Type:          "note",
			Title:         "Test1",
			EncryptedData: crypted,
		}, nil)

		mockStorage.EXPECT().GetCurrentKey().
			Return(key, nil).AnyTimes()

		mockStorage.EXPECT().GetCurrentToken().
			Return(key, nil).AnyTimes()

		err = gk.processShellCommand([]string{"get", "1"})
		require.NoError(t, err)
	})

	t.Run("processShellCommand_delete", func(t *testing.T) {
		vaultID := uint64(123)

		// токен + авторизация
		mockStorage.EXPECT().GetCurrentToken().Return("token123", nil).AnyTimes()
		mockClient.EXPECT().
			DeleteVault(gomock.Any(), &pb.DeleteVaultRequest{
				VaultId: vaultID,
			}).
			Return(&emptypb.Empty{}, errors.New("test error"))

		mockStorage.EXPECT().GetConfig().Return(kv.Config{Current: "testctx"}, nil).AnyTimes()

		// передаём фейковые аргументы (id как строка)
		args := []string{"delete", strconv.FormatUint(vaultID, 10)}

		err := gk.processShellCommand(args)
		require.Error(t, err)
	})

	t.Run("processShellCommand_help", func(t *testing.T) {
		mockStorage.EXPECT().GetCurrentKey().Return("", nil).AnyTimes()

		err := gk.processShellCommand([]string{"help"})
		require.NoError(t, err)
	})

	t.Run("processShellCommand_unknown", func(t *testing.T) {
		err := gk.processShellCommand([]string{"unknown"})
		require.Error(t, err)
	})

}
