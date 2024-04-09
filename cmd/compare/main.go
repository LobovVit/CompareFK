package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"
	"time"

	"Compare/internal/compare"
	"Compare/internal/config"
	"Compare/pkg/logger"
	"go.uber.org/zap"
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
	logger.Log.Info("--------------------------------------------")
	logger.Log.Info("--------" + time.Now().Format(time.DateTime) + "------")
	logger.Log.Info("--------------------------------------------")
	logger.Log.Info("config", zap.String("---Мode", cfg.Мode))
	logger.Log.Info("config", zap.String("---Masterdsn", cfg.Masterdsn))
	logger.Log.Info("config", zap.String("---Slavedsn", cfg.Slavedsn))
	logger.Log.Info("config", zap.String("---Attrs", cfg.Attrs))
	logger.Log.Info("config", zap.String("---LogLevel", cfg.LogLevel))
	logger.Log.Info("----------------------------")
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGABRT)
	defer cancel()
	app, err := compare.NewComparator(cfg)
	if err != nil {
		return err
	}
	return app.Run(ctx)
}
