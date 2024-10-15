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
	GetMaster(ctx context.Context, i int, sql string, db *sql.DB) error
	GetSlave(ctx context.Context, sql string, db *sql.DB) error
	GetResult(ctx context.Context) []string
}

type Comparator struct {
	masterSQL []string
	slaveSQL  string
	Storage
}

func NewComparator() (*Comparator, error) {
	mSQL, err := files.ReadCatalog(config.Cfg.MasterSQL)
	if err != nil {
		return nil, fmt.Errorf("readCatalog MasterSQL: %w", err)
	}
	sSQL, err := files.ReadFile(config.Cfg.SlaveSQL)
	if err != nil {
		return nil, fmt.Errorf("readFile SlaveSQL: %w", err)
	}
	store := storage.GetMemStorage() //todo - storage.getSQLLightStorage (from config)
	return &Comparator{
			masterSQL: mSQL,
			slaveSQL:  sSQL,
			Storage:   store},
		nil
}

func (c *Comparator) Run(ctx context.Context) error {
	err := c.getMasterData(ctx)
	if err != nil {
		return fmt.Errorf("get master data: %w", err)
	}
	logger.Log.Info("Get master data OK")
	err = c.getSlaveData(ctx)
	if err != nil {
		return fmt.Errorf("get slave data: %w", err)
	}

	resultGuids := c.GetResult(ctx)

	logger.Log.Info("write results", zap.Int("count", len(resultGuids)))
	err = files.WriteFile(config.Cfg.Мode+".txt", resultGuids)
	if err != nil {
		return fmt.Errorf("write file: %w", err)
	}
	logger.Log.Info("write statistic")
	var statistic = make([]string, 0)
	statistic = append(statistic,
		"--------------------------------------------",
		fmt.Sprintf("Мode: %v", config.Cfg.Мode),
		fmt.Sprintf("Masterdsn: %v", config.Cfg.Masterdsn),
		fmt.Sprintf("Slavedsn: %v", config.Cfg.Slavedsn),
		fmt.Sprintf("LogLevel: %v", config.Cfg.LogLevel),
		fmt.Sprintf("Limit: %v", config.Cfg.Limit),
		fmt.Sprintf("RateLimit: %v", config.Cfg.RateLimit),
		fmt.Sprintf("MasterSQL: %v", config.Cfg.MasterSQL),
		fmt.Sprintf("SlaveSQL: %v", config.Cfg.SlaveSQL),
		"--------------------------------------------")
	statistic = append(statistic, result.Res.GetResult()...)
	err = files.WriteFile("stat.txt", statistic)
	if err != nil {
		logger.Log.Error("write file stat.txt", zap.Error(err))
	}
	return nil
}

func (c *Comparator) getMasterData(ctx context.Context) error {
	//get master data
	mastedDB, err := db.NweConn(config.Cfg.Masterdsn)
	if err != nil {
		logger.Log.Error("conn master", zap.Error(err))
		return fmt.Errorf("conn master: %w", err)
	}
	if err := mastedDB.PingContext(ctx); err != nil {
		logger.Log.Error("ping master", zap.Error(err))
		return fmt.Errorf("ping master: %w", err)
	}
	g := errgroup.Group{}
	g.SetLimit(config.Cfg.RateLimit)
	for i, script := range c.masterSQL {
		g.Go(func() error {
			err = c.Storage.GetMaster(ctx, i, script, mastedDB)
			if err != nil {
				logger.Log.Error("get master", zap.Error(err))
				return fmt.Errorf("get master: %w", err)
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		logger.Log.Error("get master data", zap.Error(err))
		return fmt.Errorf("get master data: %w", err)
	}
	return nil
}

func (c *Comparator) getSlaveData(ctx context.Context) error {
	//get slave data
	slaveDB, err := db.NweConn(config.Cfg.Slavedsn)
	if err != nil {
		logger.Log.Error("conn slave", zap.Error(err))
		return fmt.Errorf("conn slave: %w", err)
	}
	if err := slaveDB.PingContext(ctx); err != nil {
		logger.Log.Error("ping slave", zap.Error(err))
		return fmt.Errorf("ping slave: %w", err)
	}
	err = c.Storage.GetSlave(ctx, c.slaveSQL, slaveDB)
	if err != nil {
		logger.Log.Error("get slave data", zap.Error(err))
		return fmt.Errorf("get slave data: %w", err)
	}
	return nil
}
