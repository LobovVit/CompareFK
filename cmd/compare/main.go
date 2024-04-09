package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"Compare/internal/compare"
	"Compare/internal/config"
	"Compare/pkg/logger"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	cfg, err := config.GetConfig()
	if err != nil {
		return fmt.Errorf("get config: %w", err)
	}
	if err = logger.Initialize(cfg.LogLevel); err != nil {
		return fmt.Errorf("log initialize: %w", err)
	}
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGABRT)
	defer cancel()
	app, err := compare.NewComparator(cfg)
	if err != nil {
		return err
	}
	return app.Run(ctx)
}
