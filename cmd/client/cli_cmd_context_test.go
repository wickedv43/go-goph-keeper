package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/wickedv43/go-goph-keeper/cmd/client/internal/kv"
	"github.com/wickedv43/go-goph-keeper/cmd/client/internal/mocks"
	"go.uber.org/mock/gomock"
)

func TestContextUseCMD(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)

	g := &GophKeeper{storage: mockStorage}
	cmd := g.ContextUseCMD()
	cmd.SetArgs([]string{"work"})

	cfg := kv.Config{
		Current:  "default",
		Contexts: map[string]kv.Context{"default": {}, "work": {}},
	}

	mockStorage.EXPECT().GetConfig().Return(cfg, nil).AnyTimes()
	mockStorage.EXPECT().SetConfig(kv.Config{
		Current:  "work",
		Contexts: cfg.Contexts,
	}).Return(nil)

	// Мокаем shellLoop если он вызывает ввод

	err := cmd.Execute()
	require.NoError(t, err)
}

func TestContextListCMD(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)

	cfg := kv.Config{
		Current: "dev",
		Contexts: map[string]kv.Context{
			"dev":  {},
			"prod": {},
		},
	}

	mockStorage.EXPECT().GetConfig().Return(cfg, nil).AnyTimes()

	g := &GophKeeper{storage: mockStorage}
	cmd := g.ContextListCMD()

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	err := cmd.Execute()
	require.NoError(t, err)

	out := buf.String()
	require.Contains(t, out, "dev (in use)")
	require.Contains(t, out, "prod")
}
