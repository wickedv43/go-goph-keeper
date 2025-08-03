package main

import (
	"testing"

	"github.com/wickedv43/go-goph-keeper/internal/config"
)

type Suite struct {
	*testing.T                // Управление тестами
	Cfg        *config.Config // Конфиг
}

func New(t *testing.T) *Suite {
	t.Helper()
	t.Parallel()

	return &Suite{
		T: t,
	}
}
