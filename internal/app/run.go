package app

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/LobovVit/CompareFK/internal/config"
	"github.com/LobovVit/CompareFK/internal/result"
	"github.com/LobovVit/CompareFK/internal/storage"
	"github.com/LobovVit/CompareFK/pkg/db"
	"github.com/LobovVit/CompareFK/pkg/files"
	"github.com/LobovVit/CompareFK/pkg/logger"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type Storage interface {
	GetMaster(ctx context.Context, i int, query string, db *sql.DB) error
	GetSlave(ctx context.Context, query string, db *sql.DB) error
	WriteResult(ctx context.Context, outputFile string) error
	Close() error
}

type Comparator struct {
	masterSQL []string
	slaveSQL  string
	Storage
}

func NewComparator(ctx context.Context) (*Comparator, error) {
	mSQL, err := files.ReadSQLSources(config.Cfg.MasterSQLDir, config.Cfg.MasterSQLGlob, config.Cfg.MasterSQLFiles)
	if err != nil {
		return nil, fmt.Errorf("read master sql sources: %w", err)
	}
	if len(mSQL) == 0 {
		return nil, fmt.Errorf("master sql files not found: dir=%s glob=%s files=%v", config.Cfg.MasterSQLDir, config.Cfg.MasterSQLGlob, config.Cfg.MasterSQLFiles)
	}
	sSQL, err := files.ReadFile(config.Cfg.SlaveSQLFile)
	if err != nil {
		return nil, fmt.Errorf("read slave sql file: %w", err)
	}

	store, err := storage.GetSQLiteStorage(ctx)
	if err != nil {
		return nil, fmt.Errorf("init sqlite storage: %w", err)
	}

	return &Comparator{masterSQL: mSQL, slaveSQL: sSQL, Storage: store}, nil
}

func (c *Comparator) Run(ctx context.Context) error {
	defer func() {
		if err := c.Close(); err != nil {
			logger.Log.Error("close storage", zap.Error(err))
		}
	}()

	if err := c.getMasterData(ctx); err != nil {
		return fmt.Errorf("get master data: %w", err)
	}
	logger.Log.Info("Get master data OK")

	if err := c.getSlaveData(ctx); err != nil {
		return fmt.Errorf("get slave data: %w", err)
	}
	logger.Log.Info("Get slave data OK")

	if err := c.WriteResult(ctx, config.Cfg.Mode+".txt"); err != nil {
		return fmt.Errorf("write result: %w", err)
	}

	statistic := []string{
		"--------------------------------------------",
		fmt.Sprintf("Mode: %v", config.Cfg.Mode),
		fmt.Sprintf("MasterDSN: %v", config.Cfg.MasterDSN),
		fmt.Sprintf("SlaveDSN: %v", config.Cfg.SlaveDSN),
		fmt.Sprintf("LogLevel: %v", config.Cfg.LogLevel),
		fmt.Sprintf("Limit: %v", config.Cfg.Limit),
		fmt.Sprintf("RateLimit: %v", config.Cfg.RateLimit),
		fmt.Sprintf("MasterSQLDir: %v", config.Cfg.MasterSQLDir),
		fmt.Sprintf("MasterSQLGlob: %v", config.Cfg.MasterSQLGlob),
		fmt.Sprintf("MasterSQLFiles: %v", config.Cfg.MasterSQLFiles),
		fmt.Sprintf("SlaveSQLFile: %v", config.Cfg.SlaveSQLFile),
		fmt.Sprintf("Storage: %v", config.Cfg.Storage),
		fmt.Sprintf("SQLitePath: %v", config.Cfg.SQLitePath),
		fmt.Sprintf("SQLiteWriteBatch: %v", config.Cfg.SQLiteWriteBatch),
		fmt.Sprintf("SQLiteCacheSizeKB: %v", config.Cfg.SQLiteCacheSizeKB),
		fmt.Sprintf("SQLiteMmapSizeMB: %v", config.Cfg.SQLiteMmapSizeMB),
		fmt.Sprintf("OutputDir: %v", config.Cfg.OutputDir),
		"--------------------------------------------",
	}
	statistic = append(statistic, result.Res.GetResult()...)
	if err := files.WriteFile("stat.txt", statistic); err != nil {
		logger.Log.Error("write file stat.txt", zap.Error(err))
	}
	return nil
}

func (c *Comparator) getMasterData(ctx context.Context) error {
	masterDB, err := db.NweConn(config.Cfg.MasterDSN, config.Cfg.MaxOpenConnsMaster)
	if err != nil {
		return fmt.Errorf("conn master: %w", err)
	}
	defer masterDB.Close()

	if err := masterDB.PingContext(ctx); err != nil {
		return fmt.Errorf("ping master: %w", err)
	}

	g := errgroup.Group{}
	g.SetLimit(config.Cfg.RateLimit)
	for i, script := range c.masterSQL {
		i := i
		script := script
		g.Go(func() error {
			if err := c.Storage.GetMaster(ctx, i, script, masterDB); err != nil {
				return fmt.Errorf("get master %d: %w", i, err)
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return fmt.Errorf("get master data: %w", err)
	}
	return nil
}

func (c *Comparator) getSlaveData(ctx context.Context) error {
	slaveDB, err := db.NweConn(config.Cfg.SlaveDSN, config.Cfg.MaxOpenConnsSlave)
	if err != nil {
		return fmt.Errorf("conn slave: %w", err)
	}
	defer slaveDB.Close()

	if err := slaveDB.PingContext(ctx); err != nil {
		return fmt.Errorf("ping slave: %w", err)
	}
	if err := c.Storage.GetSlave(ctx, c.slaveSQL, slaveDB); err != nil {
		return fmt.Errorf("get slave data: %w", err)
	}
	return nil
}
