// Package service implements the business logic layer for user and vault management.
package service

import (
	"github.com/pkg/errors"
	"github.com/samber/do/v2"

	"github.com/wickedv43/go-goph-keeper/internal/config"
	"github.com/wickedv43/go-goph-keeper/internal/logger"
	"github.com/wickedv43/go-goph-keeper/internal/storage"
	"go.uber.org/zap"
)

// Service provides business logic operations using configuration, logging, and storage.
type Service struct {
	cfg     *config.Config     // Configuration settings.
	logger  *zap.SugaredLogger // Structured logger.
	storage storage.DataKeeper // Interface to storage layer.
}

// NewService constructs a new Service instance using dependency injection.
func NewService(i do.Injector) (*Service, error) {
	u, err := do.InvokeStruct[Service](i)
	if err != nil {
		return nil, errors.Wrap(err, "invoke struct error")
	}

	u.cfg = do.MustInvoke[*config.Config](i)
	u.logger = do.MustInvoke[*logger.Logger](i).Named("service")
	u.storage = do.MustInvoke[storage.DataKeeper](i)

	return u, nil
}
