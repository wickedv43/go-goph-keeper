package main

import (
	"context"
	"fmt"
	"net"
	"testing"

	"github.com/samber/do/v2"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/wickedv43/go-goph-keeper/cmd/client/internal/kv"
	"github.com/wickedv43/go-goph-keeper/cmd/client/internal/mocks"
	pb "github.com/wickedv43/go-goph-keeper/internal/api"
	"github.com/wickedv43/go-goph-keeper/internal/config"
	"github.com/wickedv43/go-goph-keeper/internal/logger"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc"
)

func TestNewGophKeeper(t *testing.T) {
	// Шаг 1: поднимаем временный gRPC-сервер на локальном порту
	lis, err := net.Listen("tcp", "localhost:0") // динамический порт
	require.NoError(t, err)
	port := lis.Addr().(*net.TCPAddr).Port

	grpcServer := grpc.NewServer()
	pb.RegisterGophKeeperServer(grpcServer, nil) // можно nil, если методы не вызываем

	go grpcServer.Serve(lis)
	defer grpcServer.Stop()

	// Шаг 2: создаём injector и регистрируем зависимости
	injector := do.New()

	cfg := &config.Config{
		Server: config.Server{
			Port: fmt.Sprintf("%d", port),
		},
	}
	do.ProvideValue(injector, cfg)

	log, err := logger.NewLogger(injector) // если нельзя — замокай
	do.ProvideValue(injector, log)

	memKV, err := kv.NewRoseDB(injector) // или мок
	do.ProvideValue(injector, memKV)

	// Шаг 3: создаём объект
	gk, err := NewGophKeeper(injector)
	require.NoError(t, err)
	require.NotNil(t, gk)

	// Шаг 4: проверка, что всё инициализировалось
	require.Equal(t, cfg, gk.cfg)
	require.Equal(t, log, gk.log)
	require.Equal(t, memKV, gk.storage)
	require.NotNil(t, gk.client)
	require.NotNil(t, gk.rootCtx)
	require.NotNil(t, gk.cancelCtx)
}

func TestGophKeeper_Run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockGophKeeperClient(ctrl)
	mockStorage := mocks.NewMockStorage(ctrl)

	gk := &GophKeeper{
		rootCmd: &cobra.Command{},
		client:  mockClient,
		storage: mockStorage,
		rootCtx: context.Background(),
		cfg:     &config.Config{},
	}

	gk.Start()

}
