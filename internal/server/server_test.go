package server

import (
	"testing"
	"time"

	"github.com/samber/do/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wickedv43/go-goph-keeper/internal/config"
	"github.com/wickedv43/go-goph-keeper/internal/logger"
	"github.com/wickedv43/go-goph-keeper/internal/mocks"
	"github.com/wickedv43/go-goph-keeper/internal/service"
	"go.uber.org/zap"
	"google.golang.org/grpc"
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

func TestServer_StartAndShutdown(t *testing.T) {
	logger := zap.NewNop().Sugar()

	// Заменяем реальные зависимости на заглушки
	cfg := config.Config{
		Server: config.Server{
			Port: "5051", // свободный порт
		},
	}

	grpcServer := grpc.NewServer()

	srv := &Server{
		log:  logger,
		GRPC: grpcServer,
		cfg:  &cfg,
	}

	// Запускаем сервер в фоне
	go func() {
		srv.Start()
	}()

	// Дожидаемся, пока порт откроется (до 1 секунды)
	require.Eventually(t, func() bool {
		conn, err := grpc.Dial("localhost:"+cfg.Server.Port, grpc.WithInsecure())
		if err != nil {
			return false
		}
		_ = conn.Close()
		return true
	}, time.Second, 10*time.Millisecond)

	// Завершаем сервер
	srv.Shutdown()

	// После shutdown: попытка подключиться должна фейлиться
	time.Sleep(50 * time.Millisecond) // дать время закрыться
	_, err := grpc.Dial("localhost:"+cfg.Server.Port, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(time.Second))
	assert.Error(t, err, "ожидалась ошибка подключения после Shutdown()")
}
