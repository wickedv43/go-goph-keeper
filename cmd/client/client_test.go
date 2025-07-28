package main

import (
	"testing"

	"github.com/samber/do/v2"
	"github.com/stretchr/testify/require"
	"github.com/wickedv43/go-goph-keeper/cmd/client/internal/kv"
	"github.com/wickedv43/go-goph-keeper/internal/config"
	"github.com/wickedv43/go-goph-keeper/internal/logger"
)

func setupTestClient(tb testing.TB) *GophKeeper {
	tb.Helper()

	container := do.New()

	// Минимальная конфигурация
	cfg := &config.Config{
		Server: config.Server{
			Port: "50051", // укажи нужный тестовый порт gRPC
		},
		KV: config.KV{
			DirPath: tb.TempDir(), // временное хранилище
		},
	}
	do.ProvideValue(container, cfg)

	// Логгер
	log, err := logger.NewLogger(container)
	require.NoError(tb, err)
	do.ProvideValue(container, log)

	// KV-хранилище
	kv, err := kv.NewRoseDB(container)
	require.NoError(tb, err)
	tb.Cleanup(func() { kv.Shutdown() })
	do.ProvideValue(container, kv)

	// Клиент GophKeeper
	client, err := NewGophKeeper(container)
	require.NoError(tb, err)

	return client
}
