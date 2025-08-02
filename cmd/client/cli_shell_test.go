package main

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/wickedv43/go-goph-keeper/cmd/client/internal/kv"
	"github.com/wickedv43/go-goph-keeper/cmd/client/internal/mocks"
	"github.com/wickedv43/go-goph-keeper/internal/config"
	"go.uber.org/mock/gomock"
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
