package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"github.com/wickedv43/go-goph-keeper/cmd/client/internal/crypto"
	"github.com/wickedv43/go-goph-keeper/cmd/client/internal/kv"
	"github.com/wickedv43/go-goph-keeper/cmd/client/internal/mocks"
	pb "github.com/wickedv43/go-goph-keeper/internal/api"
	"github.com/wickedv43/go-goph-keeper/internal/config"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestLoginCMD(t *testing.T) {
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

	// Создаём pipe
	r, w, _ := os.Pipe()

	origStdin := os.Stdin

	os.Stdin = r
	defer func() {
		os.Stdin = origStdin
	}()

	t.Run("success login", func(t *testing.T) {
		//emulate user's input
		go func() {
			fmt.Fprintln(w, "login")
			fmt.Fprintln(w, "pass")
		}()

		mockClient.EXPECT().
			Login(gomock.Any(), &pb.LoginRequest{Login: "login", Password: gk.hashPassword("pass")}).
			Return(&pb.LoginResponse{Token: "token123"}, nil)

		mockStorage.EXPECT().
			SaveContext("login", "token123").
			Return(nil)

		mockStorage.EXPECT().
			GetCurrentKey().
			Return("already-there", nil)

		mockStorage.EXPECT().
			GetConfig().
			Return(kv.Config{Current: "testctx"}, nil).
			AnyTimes()

		cmd := gk.LoginCMD()

		err := cmd.RunE(cmd, nil)
		require.NoError(t, err)
	})

	t.Run("error bad pass", func(t *testing.T) {
		//emulate user's input
		go func() {
			fmt.Fprintln(w, "login")
			fmt.Fprintln(w, "pass")
		}()

		mockClient.EXPECT().
			Login(gomock.Any(), gomock.Any()).
			Return(&pb.LoginResponse{}, errors.New("bad password"))

		cmd := gk.LoginCMD()

		err := cmd.RunE(cmd, nil)
		require.ErrorContains(t, err, "bad password")
	})

	t.Run("error not found login", func(t *testing.T) {
		//emulate user's input
		go func() {
			fmt.Fprintln(w, "login")
			fmt.Fprintln(w, "pass")
		}()

		mockClient.EXPECT().
			Login(gomock.Any(), &pb.LoginRequest{Login: "login", Password: gk.hashPassword("pass")}).
			Return(&pb.LoginResponse{}, errors.New("not found"))

		cmd := gk.LoginCMD()

		err := cmd.RunE(cmd, nil)
		require.Error(t, err)
	})

	t.Run("enter mnemo", func(t *testing.T) {
		//emulate user's input
		go func() {
			fmt.Fprintln(w, "login")
			fmt.Fprintln(w, "pass")

			// mnemo input)))
			for i := 0; i < 12; i++ {
				fmt.Fprintln(w, "apple")
			}
		}()

		mockClient.EXPECT().
			Login(gomock.Any(), &pb.LoginRequest{Login: "login", Password: gk.hashPassword("pass")}).
			Return(&pb.LoginResponse{Token: "token123"}, nil)

		mockStorage.EXPECT().
			SaveContext("login", "token123").
			Return(nil)

		mockStorage.EXPECT().
			GetCurrentKey().
			Return("", kv.ErrEmptyKey)

		mockStorage.EXPECT().
			GetConfig().
			Return(kv.Config{Current: "login"}, nil).
			AnyTimes()

		mockStorage.EXPECT().
			SaveKey("login", "82d946efc257129275a0da26c2be39b149a7eaa2cb0d020468fba95850733b39f1c5d8e8328f4950573e94a075f84f64034631c09fd95a26aa8f8362df209afe").
			Return(nil)

		cmd := gk.LoginCMD()

		err := cmd.RunE(cmd, nil)
		require.NoError(t, err)
	})

	t.Run("error cannot save ctx", func(t *testing.T) {
		//emulate user's input
		go func() {
			fmt.Fprintln(w, "login")
			fmt.Fprintln(w, "pass")
		}()

		mockClient.EXPECT().
			Login(gomock.Any(), gomock.Any()).
			Return(&pb.LoginResponse{Token: "token123"}, nil)

		mockStorage.EXPECT().
			SaveContext(gomock.Any(), gomock.Any()).
			Return(errors.New("cannot save context"))

		cmd := gk.LoginCMD()

		err := cmd.RunE(cmd, nil)
		require.ErrorContains(t, err, "cannot save context")
	})

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

	// Создаём pipe
	r, w, _ := os.Pipe()

	origStdin := os.Stdin

	os.Stdin = r
	defer func() {
		os.Stdin = origStdin
	}()

	t.Run("register_success", func(t *testing.T) {
		//emulate user's input
		go func() {
			fmt.Fprintln(w, "login")
			fmt.Fprintln(w, "pass")
		}()

		mockClient.EXPECT().
			Register(gomock.Any(), gomock.Any()).
			Return(&pb.RegisterResponse{}, nil)

		mockStorage.EXPECT().
			SaveKey("login", gomock.Any()).
			Return(nil)

		mockStorage.EXPECT().
			GetConfig().
			Return(kv.Config{Current: "login"}, nil).
			AnyTimes()

		var buf bytes.Buffer

		cmd := gk.RegisterCMD()
		cmd.SetOut(&buf)

		err := cmd.RunE(cmd, nil)
		require.NoError(t, err)

		require.Contains(t, buf.String(), "💾 Save this phrase:")
	})

	t.Run("register_error_already_used", func(t *testing.T) {
		//emulate user's input
		go func() {
			fmt.Fprintln(w, "login")
			fmt.Fprintln(w, "pass")
		}()

		mockClient.EXPECT().
			Register(gomock.Any(), gomock.Any()).
			Return(&pb.RegisterResponse{}, errors.New("login already exists"))

		cmd := gk.RegisterCMD()

		err := cmd.RunE(cmd, nil)
		require.Error(t, err)
	})
}

func TestNewVaultsCMD(t *testing.T) {
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

	// Создаём pipe
	r, w, _ := os.Pipe()

	origStdin := os.Stdin

	os.Stdin = r
	defer func() {
		os.Stdin = origStdin
	}()

	t.Run("new_vault_login_success", func(t *testing.T) {
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
			Return("6368616e676520746869732070617373", nil)

		// Ожидаемый вызов получения ключа
		mockStorage.EXPECT().
			GetCurrentToken().
			Return("6368616e676520746869732070617373", nil)

		cmd := gk.NewVaultCMD()

		err := cmd.RunE(cmd, nil)
		require.NoError(t, err)
	})

	t.Run("new_vault_login_error", func(t *testing.T) {
		go func() {
			fmt.Fprintln(w, "TestTitle")
			fmt.Fprintln(w, "login")
			fmt.Fprintln(w, "log")
			fmt.Fprintln(w, "pass")
		}()

		mockClient.EXPECT().
			CreateVault(gomock.Any(), gomock.Any()).
			Return(&emptypb.Empty{}, errors.New("bad request"))

		// Ожидаемый вызов получения ключа
		mockStorage.EXPECT().
			GetCurrentKey().
			Return("6368616e676520746869732070617373", nil)

		// Ожидаемый вызов получения ключа
		mockStorage.EXPECT().
			GetCurrentToken().
			Return("6368616e676520746869732070617373", nil)

		cmd := gk.NewVaultCMD()

		err := cmd.RunE(cmd, nil)
		require.Error(t, err)
	})

	t.Run("new_vault_note_success", func(t *testing.T) {
		go func() {
			fmt.Fprintln(w, "TestTitle")
			fmt.Fprintln(w, "note")
			fmt.Fprintln(w, "noteasdasdasd")
		}()

		mockClient.EXPECT().
			CreateVault(gomock.Any(), gomock.Any()).
			Return(&emptypb.Empty{}, nil)

		// Ожидаемый вызов получения ключа
		mockStorage.EXPECT().
			GetCurrentKey().
			Return("6368616e676520746869732070617373", nil)

		// Ожидаемый вызов получения ключа
		mockStorage.EXPECT().
			GetCurrentToken().
			Return("6368616e676520746869732070617373", nil)

		cmd := gk.NewVaultCMD()

		err := cmd.RunE(cmd, nil)
		require.NoError(t, err)
	})

	t.Run("new_vault_note_error", func(t *testing.T) {
		go func() {
			fmt.Fprintln(w, "TestTitle")
			fmt.Fprintln(w, "note")
			fmt.Fprintln(w, "noteasdasdasd")
		}()

		mockClient.EXPECT().
			CreateVault(gomock.Any(), gomock.Any()).
			Return(&emptypb.Empty{}, errors.New("bad request"))

		// Ожидаемый вызов получения ключа
		mockStorage.EXPECT().
			GetCurrentKey().
			Return("6368616e676520746869732070617373", nil)

		// Ожидаемый вызов получения ключа
		mockStorage.EXPECT().
			GetCurrentToken().
			Return("6368616e676520746869732070617373", nil)

		cmd := gk.NewVaultCMD()

		err := cmd.RunE(cmd, nil)
		require.Error(t, err)
	})

	t.Run("new_vault_note_success", func(t *testing.T) {
		go func() {
			fmt.Fprintln(w, "TestTitle")
			fmt.Fprintln(w, "note")
			fmt.Fprintln(w, "noteasdasdasd")
		}()

		mockClient.EXPECT().
			CreateVault(gomock.Any(), gomock.Any()).
			Return(&emptypb.Empty{}, nil)

		// Ожидаемый вызов получения ключа
		mockStorage.EXPECT().
			GetCurrentKey().
			Return("6368616e676520746869732070617373", nil)

		// Ожидаемый вызов получения ключа
		mockStorage.EXPECT().
			GetCurrentToken().
			Return("6368616e676520746869732070617373", nil)

		cmd := gk.NewVaultCMD()

		err := cmd.RunE(cmd, nil)
		require.NoError(t, err)
	})

	t.Run("new_vault_card_success", func(t *testing.T) {
		go func() {
			fmt.Fprintln(w, "TestTitle")
			fmt.Fprintln(w, "card")
			fmt.Fprintln(w, "cardnumber")
			fmt.Fprintln(w, "carddate")
			fmt.Fprintln(w, "cvv")
		}()

		mockClient.EXPECT().
			CreateVault(gomock.Any(), gomock.Any()).
			Return(&emptypb.Empty{}, nil)

		// Ожидаемый вызов получения ключа
		mockStorage.EXPECT().
			GetCurrentKey().
			Return("6368616e676520746869732070617373", nil)

		// Ожидаемый вызов получения ключа
		mockStorage.EXPECT().
			GetCurrentToken().
			Return("6368616e676520746869732070617373", nil)

		cmd := gk.NewVaultCMD()

		err := cmd.RunE(cmd, nil)
		require.NoError(t, err)
	})

	t.Run("new_vault_card_error", func(t *testing.T) {
		go func() {
			fmt.Fprintln(w, "TestTitle")
			fmt.Fprintln(w, "card")
			fmt.Fprintln(w, "cardnumber")
			fmt.Fprintln(w, "carddate")
			fmt.Fprintln(w, "cvv")
		}()

		mockClient.EXPECT().
			CreateVault(gomock.Any(), gomock.Any()).
			Return(&emptypb.Empty{}, errors.New("bad request"))

		// Ожидаемый вызов получения ключа
		mockStorage.EXPECT().
			GetCurrentKey().
			Return("6368616e676520746869732070617373", nil)

		// Ожидаемый вызов получения ключа
		mockStorage.EXPECT().
			GetCurrentToken().
			Return("6368616e676520746869732070617373", nil)

		cmd := gk.NewVaultCMD()

		err := cmd.RunE(cmd, nil)
		require.Error(t, err)
	})

	t.Run("new_vault_binary_success", func(t *testing.T) {
		go func() {
			fmt.Fprintln(w, "TestTitle")
			fmt.Fprintln(w, "binary")
		}()

		// Моки
		mockStorage.EXPECT().
			GetCurrentKey().
			Return("6368616e676520746869732070617373", nil)

		mockStorage.EXPECT().
			GetCurrentToken().
			Return("6368616e676520746869732070617373", nil)

		mockClient.EXPECT().
			CreateVault(gomock.Any(), gomock.Any()).
			Return(&emptypb.Empty{}, errors.New("bad request")) // проверим поведение при ошибке

		cmd := gk.NewVaultCMD()
		cmd.SetOut(io.Discard)
		cmd.SetErr(io.Discard)

		err := cmd.RunE(cmd, nil)
		require.ErrorContains(t, err, "bad request")
	})
}

func TestVaultTypes(t *testing.T) {
	// Создаём pipe
	r, w, _ := os.Pipe()
	// Сохраняем оригинальный Stdin
	origStdin := os.Stdin
	// Подменяем Stdin
	os.Stdin = r
	defer func() {
		os.Stdin = origStdin
	}()

	t.Run("type_login_success", func(t *testing.T) {
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
	})
	t.Run("type_note_success", func(t *testing.T) {
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
	})

	t.Run("type_card_success", func(t *testing.T) {
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
	})
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

	t.Run("new_vault_list_success", func(t *testing.T) {
		// Ожидаемый вызов получения ключа
		mockStorage.EXPECT().
			GetCurrentKey().
			Return("6368616e676520746869732070617373", nil).AnyTimes()

		mockClient.EXPECT().
			ListVaults(gomock.Any(), gomock.Any()).
			Return(&pb.ListVaultsResponse{
				Vaults: []*pb.VaultRecord{
					{Id: 1, Title: "Test1", Type: "note"},
					{Id: 2, Title: "Test2", Type: "login"},
				},
			}, nil)

		// Ожидаемый вызов получения ключа
		mockStorage.EXPECT().
			GetCurrentToken().
			Return("6368616e676520746869732070617373", nil)

		cmd := gk.VaultListCMD()

		err := cmd.RunE(cmd, nil)
		require.NoError(t, err)
	})

	t.Run("new_vault_list_error", func(t *testing.T) {
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

		var buf bytes.Buffer
		cmd := gk.VaultListCMD()
		cmd.SetOut(&buf)

		err := cmd.RunE(cmd, nil)
		require.NoError(t, err)

		list := buf.String()
		require.Contains(t, list, "пусто")

	})
}

func TestVaultShowCMD(t *testing.T) {
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

	// Создаём pipe
	r, _, _ := os.Pipe()

	origStdin := os.Stdin

	os.Stdin = r
	defer func() {
		os.Stdin = origStdin
	}()

	t.Run("show_note_success", func(t *testing.T) {
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
			Return(key, nil)

		mockStorage.EXPECT().GetCurrentToken().
			Return(key, nil)

		args := []string{"get", "1"}
		cmd := gk.VaultShowCMD()
		cmd.SetArgs(args)

		var b bytes.Buffer
		cmd.SetOut(&b)

		err = cmd.RunE(cmd, args)
		require.NoError(t, err)

		vaults := b.String()
		require.Contains(t, vaults, "Test1")
		require.Contains(t, vaults, "text note")
	})

	t.Run("show_note_error_key", func(t *testing.T) {
		key := "6368616e676520746869732070617373"
		key1 := "6368616e676520746869732070617374"

		var note kv.Note
		note.Text = "text note"
		data, err := json.Marshal(note)
		require.NoError(t, err)

		crypted, err := crypto.EncryptWithSeed(data, key1)
		require.NoError(t, err)

		mockClient.EXPECT().GetVault(gomock.Any(), gomock.Any()).Return(&pb.VaultRecord{
			Id:            1,
			Type:          "note",
			Title:         "Test1",
			EncryptedData: crypted,
		}, nil)

		mockStorage.EXPECT().GetCurrentKey().
			Return(key, nil)

		mockStorage.EXPECT().GetCurrentToken().
			Return(key, nil)

		args := []string{"get", "1"}
		cmd := gk.VaultShowCMD()
		cmd.SetArgs(args)

		var b bytes.Buffer
		cmd.SetOut(&b)

		err = cmd.RunE(cmd, args)
		require.Error(t, err)
	})

	t.Run("show_login_success", func(t *testing.T) {
		key := "6368616e676520746869732070617373"

		var log kv.LoginPass
		log.Login = "test login"
		log.Password = "test password"
		data, err := json.Marshal(log)
		require.NoError(t, err)

		crypted, err := crypto.EncryptWithSeed(data, key)
		require.NoError(t, err)

		mockClient.EXPECT().GetVault(gomock.Any(), gomock.Any()).Return(&pb.VaultRecord{
			Id:            1,
			Type:          "login",
			Title:         "Test1",
			EncryptedData: crypted,
		}, nil)

		mockStorage.EXPECT().GetCurrentKey().
			Return(key, nil)

		mockStorage.EXPECT().GetCurrentToken().
			Return(key, nil)

		args := []string{"get", "1"}
		cmd := gk.VaultShowCMD()
		cmd.SetArgs(args)

		var b bytes.Buffer
		cmd.SetOut(&b)

		err = cmd.RunE(cmd, args)
		require.NoError(t, err)

		vaults := b.String()
		require.Contains(t, vaults, "Test1")
		require.Contains(t, vaults, "test login")
		require.Contains(t, vaults, "test password")
	})

	t.Run("show_login_error_key", func(t *testing.T) {
		key := "6368616e676520746869732070617373"
		key1 := "6368616e676520746869732070617374"

		var log kv.LoginPass
		log.Login = "test login"
		log.Password = "test password"
		data, err := json.Marshal(log)
		require.NoError(t, err)

		crypted, err := crypto.EncryptWithSeed(data, key1)
		require.NoError(t, err)

		mockClient.EXPECT().GetVault(gomock.Any(), gomock.Any()).Return(&pb.VaultRecord{
			Id:            1,
			Type:          "login",
			Title:         "Test1",
			EncryptedData: crypted,
		}, nil)

		mockStorage.EXPECT().GetCurrentKey().
			Return(key, nil)

		mockStorage.EXPECT().GetCurrentToken().
			Return(key, nil)

		args := []string{"get", "1"}
		cmd := gk.VaultShowCMD()
		cmd.SetArgs(args)

		var b bytes.Buffer
		cmd.SetOut(&b)

		err = cmd.RunE(cmd, args)
		require.Error(t, err)
	})

	t.Run("show_card_success", func(t *testing.T) {
		key := "6368616e676520746869732070617373"

		var card kv.Card
		card.Number = "test number"
		card.Date = "test date"
		card.CVV = "test cvv"
		data, err := json.Marshal(card)
		require.NoError(t, err)

		crypted, err := crypto.EncryptWithSeed(data, key)
		require.NoError(t, err)

		mockClient.EXPECT().GetVault(gomock.Any(), gomock.Any()).Return(&pb.VaultRecord{
			Id:            1,
			Type:          "card",
			Title:         "Test1",
			EncryptedData: crypted,
		}, nil)

		mockStorage.EXPECT().GetCurrentKey().
			Return(key, nil)

		mockStorage.EXPECT().GetCurrentToken().
			Return(key, nil)

		args := []string{"get", "1"}
		cmd := gk.VaultShowCMD()
		cmd.SetArgs(args)

		var b bytes.Buffer
		cmd.SetOut(&b)

		err = cmd.RunE(cmd, args)
		require.NoError(t, err)

		vaults := b.String()
		require.Contains(t, vaults, "Test1")
		require.Contains(t, vaults, "test number")
		require.Contains(t, vaults, "test date")
		require.Contains(t, vaults, "test cvv")

	})

	t.Run("show_card_error_key", func(t *testing.T) {
		key := "6368616e676520746869732070617373"
		key1 := "6368616e676520746869732070617373"

		var card kv.Card
		card.Number = "test number"
		card.Date = "test date"
		card.CVV = "test cvv"
		data, err := json.Marshal(card)
		require.NoError(t, err)

		crypted, err := crypto.EncryptWithSeed(data, key1)
		require.NoError(t, err)

		mockClient.EXPECT().GetVault(gomock.Any(), gomock.Any()).Return(&pb.VaultRecord{
			Id:            1,
			Type:          "card",
			Title:         "Test1",
			EncryptedData: crypted,
		}, nil)

		mockStorage.EXPECT().GetCurrentKey().
			Return(key, nil)

		mockStorage.EXPECT().GetCurrentToken().
			Return(key, nil)

		args := []string{"get", "1"}
		cmd := gk.VaultShowCMD()
		cmd.SetArgs(args)

		var b bytes.Buffer
		cmd.SetOut(&b)

		err = cmd.RunE(cmd, args)
		require.NoError(t, err)

	})
}

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

	t.Run("delete_success", func(t *testing.T) {
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
	})

	t.Run("delete_error", func(t *testing.T) {
		cmd := gk.VaultDeleteCMD()
		vaultID := uint64(123)

		// токен + авторизация
		mockStorage.EXPECT().GetCurrentToken().Return("token123", nil)
		mockClient.EXPECT().
			DeleteVault(gomock.Any(), &pb.DeleteVaultRequest{
				VaultId: vaultID,
			}).
			Return(&emptypb.Empty{}, errors.New("test error"))

		mockStorage.EXPECT().GetConfig().Return(kv.Config{Current: "testctx"}, nil).AnyTimes()

		// передаём фейковые аргументы (id как строка)
		args := []string{"delete", strconv.FormatUint(vaultID, 10)}
		cmd.SetArgs(args)

		err := cmd.RunE(cmd, args)
		require.Error(t, err)
	})
}
