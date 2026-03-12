package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"
	"time"

	"github.com/LobovVit/CompareFK/internal/result"
	"github.com/LobovVit/CompareFK/internal/web"
	"go.uber.org/zap"

	"github.com/LobovVit/CompareFK/internal/app"
	"github.com/LobovVit/CompareFK/internal/config"
	"github.com/LobovVit/CompareFK/pkg/db"
	"github.com/LobovVit/CompareFK/pkg/logger"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)
	if err := run(context.Background()); err != nil {
		panic(err)
	}
}

func run(ctx context.Context) error {
	if err := config.Initialize(); err != nil {
		return fmt.Errorf("config initialize: %w", err)
	}
	if err := result.Initialize(config.Cfg.ConfigFile); err != nil {
		return fmt.Errorf("result initialize: %w", err)
	}
	if err := logger.Initialize(config.Cfg); err != nil {
		return fmt.Errorf("log initialize: %w", err)
	}
	showConfig(config.Cfg)

	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGABRT)
	defer cancel()

	if config.Cfg.WebEnabled {
		if _, err := web.Start(ctx); err != nil {
			return fmt.Errorf("web initialize: %w", err)
		}
	}

	application, err := app.NewComparator(ctx)
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

func showConfig(c *config.Config) {
	logger.Log.Info("--------------------------------------------")
	logger.Log.Info(fmt.Sprintf("Build version: %s\n", buildVersion))
	logger.Log.Info(fmt.Sprintf("Build date: %s\n", buildDate))
	logger.Log.Info(fmt.Sprintf("Build commit: %s\n", buildCommit))
	logger.Log.Info("--------------------------------------------")
	logger.Log.Info("--------" + time.Now().Format(time.DateTime) + "------")
	logger.Log.Info("--------------------------------------------")
	logger.Log.Info(fmt.Sprintf("config---Mode: %v", c.Mode))
	logger.Log.Info(fmt.Sprintf("config---MasterDB: %v", db.SafeDSNInfo(c.MasterDSN)))
	logger.Log.Info(fmt.Sprintf("config---SlaveDB: %v", db.SafeDSNInfo(c.SlaveDSN)))
	logger.Log.Info(fmt.Sprintf("config---LogLevel: %v", c.LogLevel))
	logger.Log.Info(fmt.Sprintf("config---Limit: %v", c.Limit))
	logger.Log.Info(fmt.Sprintf("config---RateLimit: %v", c.RateLimit))
	logger.Log.Info(fmt.Sprintf("config---MasterSQLDir: %v", c.MasterSQLDir))
	logger.Log.Info(fmt.Sprintf("config---MasterSQLGlob: %v", c.MasterSQLGlob))
	logger.Log.Info(fmt.Sprintf("config---MasterSQLFiles: %v", c.MasterSQLFiles))
	logger.Log.Info(fmt.Sprintf("config---SlaveSQLFile: %v", c.SlaveSQLFile))
	logger.Log.Info(fmt.Sprintf("config---Storage: %v", c.Storage))
	logger.Log.Info(fmt.Sprintf("config---SQLitePath: %v", c.SQLitePath))
	logger.Log.Info(fmt.Sprintf("config---SQLiteBusyTimeoutMs: %v", c.SQLiteBusyTimeoutMs))
	logger.Log.Info(fmt.Sprintf("config---SQLiteCacheSizeKB: %v", c.SQLiteCacheSizeKB))
	logger.Log.Info(fmt.Sprintf("config---SQLiteMmapSizeMB: %v", c.SQLiteMmapSizeMB))
	logger.Log.Info(fmt.Sprintf("config---SQLiteWriteBatch: %v", c.SQLiteWriteBatch))
	logger.Log.Info(fmt.Sprintf("config---OutputDir: %v", c.OutputDir))
	logger.Log.Info(fmt.Sprintf("config---WebEnabled: %v", c.WebEnabled))
	logger.Log.Info(fmt.Sprintf("config---WebListen: %v", c.WebListen))
	logger.Log.Info(fmt.Sprintf("config---WebRefreshSec: %v", c.WebRefreshSec))
	logger.Log.Info("--------------------------------------------")
}
