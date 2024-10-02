package main

import (
	"Compare/internal/result"
	"context"
	"fmt"
	"go.uber.org/zap"
	"os/signal"
	"syscall"
	"time"

	"Compare/internal/app"
	"Compare/internal/config"
	"Compare/pkg/logger"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	if err := result.Initialize(); err != nil {
		return fmt.Errorf("result initialize: %w", err)
	}
	if err := config.Initialize(); err != nil {
		return fmt.Errorf("config initialize: %w", err)
	}
	if err := logger.Initialize(config.Cfg); err != nil {
		return fmt.Errorf("log initialize: %w", err)
	}
	ShowConfig(config.Cfg)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGABRT)
	defer cancel()
	application, err := app.NewComparator()
	if err != nil {
		logger.Log.Error("new application", zap.Error(err))
		return fmt.Errorf("new application: %w", err)
	}
	if err := application.Run(ctx); err != nil {
		logger.Log.Error("run application", zap.Error(err))
		return fmt.Errorf("run application: %w", err)
	}
	return nil
}

func ShowConfig(c *config.Config) {
	logger.Log.Info("--------------------------------------------")
	logger.Log.Info("--------" + time.Now().Format(time.DateTime) + "------")
	logger.Log.Info("--------------------------------------------")
	logger.Log.Info(fmt.Sprintf("config---Мode: %v", c.Мode))
	logger.Log.Info(fmt.Sprintf("config---Masterdsn: %v", c.Masterdsn))
	logger.Log.Info(fmt.Sprintf("config---Slavedsn: %v", c.Slavedsn))
	logger.Log.Info(fmt.Sprintf("config---LogLevel: %v", c.LogLevel))
	logger.Log.Info(fmt.Sprintf("config---Limit: %v", c.Limit))
	logger.Log.Info(fmt.Sprintf("config---RateLimit: %v", c.RateLimit))
	logger.Log.Info(fmt.Sprintf("config---MasterSQL: %v", c.MasterSQL))
	logger.Log.Info(fmt.Sprintf("config---SlaveSQL: %v", c.SlaveSQL))
	logger.Log.Info("--------------------------------------------")
}
