package main

import (
	"context"

	"github.com/samber/do/v2"
	"github.com/wickedv43/go-goph-keeper/internal/logger"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background()) //root ctx
	defer cancel()

	i := do.New()

	do.ProvideNamedValue(i, "root.context", ctx)
	do.Provide(i, logger.NewLogger)

	log := do.MustInvoke[*logger.Logger](i)
	log.Info("starting app...")

	_ = i.ShutdownWithContext(ctx)
	log.Info("grace shutdown")
}
