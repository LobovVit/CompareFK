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
	if err = logger.Initialize(cfg); err != nil {
		return fmt.Errorf("log initialize: %w", err)
	}
	ShowConfig(cfg)
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGABRT)
	defer cancel()
	app, err := compare.NewComparator(cfg)
	if err != nil {
		return err
	}
	return app.Run(ctx)
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
	logger.Log.Info(fmt.Sprintf("config---LogFile: %v", c.LogFile))
	logger.Log.Info(fmt.Sprintf("config---ResFilePrefix: %v", c.ResFile))
	logger.Log.Info("--------------------------------------------")
}
