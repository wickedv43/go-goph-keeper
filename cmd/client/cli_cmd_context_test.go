package main

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"github.com/wickedv43/go-goph-keeper/cmd/client/internal/kv"
	"github.com/wickedv43/go-goph-keeper/cmd/client/internal/mocks"
	"github.com/wickedv43/go-goph-keeper/internal/config"
	"go.uber.org/mock/gomock"
)

func TestContextUseCMD(t *testing.T) {
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

	t.Run("context_use_success", func(t *testing.T) {

		cfg := kv.Config{
			Current:  "default",
			Contexts: map[string]kv.Context{"default": {}, "work": {}},
		}

		mockStorage.EXPECT().GetConfig().Return(cfg, nil).AnyTimes()

		mockStorage.EXPECT().SetConfig(kv.Config{
			Current:  "work",
			Contexts: cfg.Contexts,
		}).Return(nil)

		cmd := gk.ContextUseCMD()

		args := []string{"use", "work"}
		err := cmd.RunE(cmd, args)
		require.NoError(t, err)
	})

	t.Run("context_use_error_noctx", func(t *testing.T) {

		cfg := kv.Config{
			Current:  "default",
			Contexts: map[string]kv.Context{"default": {}, "work": {}},
		}

		mockStorage.EXPECT().GetConfig().Return(cfg, errors.New("контекст не найден")).AnyTimes()

		cmd := gk.ContextUseCMD()

		args := []string{"use", "vetersuka"}
		err := cmd.RunE(cmd, args)
		require.Error(t, err)
	})

	t.Run("context_use_error_save_ctx", func(t *testing.T) {
		cfg := kv.Config{
			Current:  "default",
			Contexts: map[string]kv.Context{"default": {}, "work": {}},
		}

		mockStorage.EXPECT().GetConfig().Return(cfg, nil).AnyTimes()

		mockStorage.EXPECT().SetConfig(kv.Config{
			Current:  "work",
			Contexts: cfg.Contexts,
		}).Return(errors.New("save config error"))

		cmd := gk.ContextUseCMD()

		args := []string{"use", "work"}
		err := cmd.RunE(cmd, args)
		require.Error(t, err)
	})

}

func TestContextListCMD(t *testing.T) {
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

	t.Run("context_list_success", func(t *testing.T) {
		cfg := kv.Config{
			Current: "dev",
			Contexts: map[string]kv.Context{
				"dev":  {},
				"prod": {},
			},
		}

		mockStorage.EXPECT().GetConfig().Return(cfg, nil).AnyTimes()

		cmd := gk.ContextListCMD()

		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetErr(&buf)
		err := cmd.Execute()
		require.NoError(t, err)

		out := buf.String()
		require.Contains(t, out, "dev (in use)")
		require.Contains(t, out, "prod")
	})

	t.Run("context_list_error", func(t *testing.T) {
		mockStorage.EXPECT().GetConfig().Return(kv.Config{}, errors.New("get cfg err")).AnyTimes()

		cmd := gk.ContextListCMD()

		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetErr(&buf)
		err := cmd.RunE(cmd, nil)
		require.NoError(t, err)
	})

}
