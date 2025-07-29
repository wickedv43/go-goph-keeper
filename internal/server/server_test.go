package server

import (
	"testing"

	"github.com/samber/do/v2"
	"github.com/stretchr/testify/require"
	"github.com/wickedv43/go-goph-keeper/internal/config"
	"github.com/wickedv43/go-goph-keeper/internal/logger"
	"github.com/wickedv43/go-goph-keeper/internal/mocks"
	"github.com/wickedv43/go-goph-keeper/internal/service"
)

func TestNewServer(t *testing.T) {
	// 1. Готовим зависимости
	i := do.New()

	// Config с портом
	cfg := &config.Config{
		Server: config.Server{Port: "9090"},
	}

	// Мок-сервис
	mockService := &mocks.MockGophKeeper{}

	// Логгер
	log, err := logger.NewLogger(i)
	require.NoError(t, err)

	// 2. Регистрируем зависимости в do
	do.ProvideValue(i, cfg)
	do.ProvideValue[service.GophKeeper](i, mockService)
	do.ProvideValue[*logger.Logger](i, log)

	// 3. Вызываем
	srv, err := NewServer(i)
	require.NoError(t, err)
	require.NotNil(t, srv)
	require.NotNil(t, srv.GRPC)
	require.Same(t, cfg, srv.cfg)
	require.Same(t, mockService, srv.service)
}
