package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"github.com/wickedv43/go-goph-keeper/cmd/client/internal/kv"
	"github.com/wickedv43/go-goph-keeper/cmd/client/internal/mocks"
	pb "github.com/wickedv43/go-goph-keeper/internal/api"
	"github.com/wickedv43/go-goph-keeper/internal/config"
	"go.uber.org/mock/gomock"
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
