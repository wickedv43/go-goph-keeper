package service

import (
	"testing"

	"github.com/samber/do/v2"
	"github.com/stretchr/testify/require"
	"github.com/wickedv43/go-goph-keeper/internal/config"
	"github.com/wickedv43/go-goph-keeper/internal/logger"
	"github.com/wickedv43/go-goph-keeper/internal/mocks"
	"github.com/wickedv43/go-goph-keeper/internal/storage"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func TestNewService(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	i := do.New()

	cfg := &config.Config{}
	log := zap.NewNop().Sugar() // тихий логгер

	do.Provide(i, func(_ do.Injector) (*config.Config, error) {
		return cfg, nil
	})

	do.Provide(i, func(_ do.Injector) (*logger.Logger, error) {
		return &logger.Logger{SugaredLogger: log}, nil
	})

	mockStorage := mocks.NewMockDataKeeper(ctrl)
	do.Provide(i, func(_ do.Injector) (storage.DataKeeper, error) {
		return mockStorage, nil
	})

	svc, err := NewService(i)
	require.NoError(t, err)
	require.NotNil(t, svc)
}
