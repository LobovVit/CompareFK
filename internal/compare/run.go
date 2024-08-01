package compare

import (
	"context"
	"fmt"
	"sync"
	"time"

	"Compare/internal/config"
	"Compare/internal/files"
	"Compare/pkg/db"
	"Compare/pkg/logger"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type Comparator struct {
	masterSQL   string
	slaveSQL    string
	masterGuids []string
	slaveGuids  []string
	resultGuids []string
	mutex       sync.Mutex
	config      *config.Config
}

func NewComparator(cfg *config.Config) (*Comparator, error) {
	mSQL, err := files.ReadFile(cfg.MasterSQL)
	if err != nil {
		logger.Log.Error("Ошибка ReadFile", zap.Error(err))
	}
	sSQL, err := files.ReadFile(cfg.SlaveSQL)
	if err != nil {
		logger.Log.Error("Ошибка ReadFile", zap.Error(err))
	}
	return &Comparator{
			masterSQL:   mSQL,
			slaveSQL:    sSQL,
			masterGuids: make([]string, 0),
			slaveGuids:  make([]string, 0),
			resultGuids: make([]string, 0),
			config:      cfg},
		nil
}

func (c *Comparator) Run(ctx context.Context) error {
	err := c.getMasterData(ctx)
	if err != nil {
		logger.Log.Error("Get master data", zap.Error(err))
		return fmt.Errorf("get master data: %w", err)
	}
	logger.Log.Info("Get master data OK", zap.Int("master count:", len(c.masterGuids)))
	err = c.getSlaveDataParallel(ctx)
	if err != nil {
		logger.Log.Error("Get slave data", zap.Error(err))
		return fmt.Errorf("get slave data: %w", err)
	}

	switch c.config.Мode {
	case "intersection":
		c.resultGuids = intersection(c.masterGuids, c.slaveGuids)
		logger.Log.Info("result=", zap.Int(c.config.Мode, len(c.resultGuids)))
	case "difference":
		c.resultGuids = difference(c.masterGuids, c.slaveGuids)
		logger.Log.Info("result=", zap.Int(c.config.Мode, len(c.resultGuids)))
	default:
		logger.Log.Info("mode is incorrect")
		return fmt.Errorf("mode is incorrect")
	}

	err = files.WriteFile("./"+time.Now().Format(time.DateTime)+c.config.ResFile+c.config.Мode+".txt", c.resultGuids)
	if err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}

func (c *Comparator) getMasterData(ctx context.Context) error {
	//get master data
	mastedDB, err := db.NweConn(c.config.Masterdsn)
	if err != nil {
		logger.Log.Error("Ошибка connMaster", zap.Error(err))
	}

	masterRows, err := mastedDB.QueryContext(ctx, c.masterSQL)
	if err != nil {
		logger.Log.Error("Select master", zap.Error(err))
		return fmt.Errorf("select master: %w", err)
	}
	if err = masterRows.Err(); err != nil {
		logger.Log.Error("Select master", zap.Error(err))
		return fmt.Errorf("select master: %w", err)
	}
	defer masterRows.Close()

	for masterRows.Next() {
		var guid string
		err = masterRows.Scan(&guid)
		if err != nil {
			logger.Log.Error("Select slave", zap.Error(err))
			return fmt.Errorf("select slave: %w", err)
		}
		c.masterGuids = append(c.masterGuids, guid)
	}
	logger.Log.Info("masterGuids=", zap.Int("cnt", len(c.masterGuids)))
	return nil
}

func (c *Comparator) getSlaveData(ctx context.Context, val []string) ([]string, error) {
	//get slave data part
	result := []string{}
	slaveDB, err := db.NweConn(c.config.Slavedsn)
	if err != nil {
		logger.Log.Error("Ошибка connSlave", zap.Error(err))
	}
	slaveRows, err := slaveDB.QueryContext(ctx, c.slaveSQL, val)
	if err != nil {
		logger.Log.Error("Select slave", zap.Error(err))
		return nil, fmt.Errorf("select slave: %w", err)
	}
	if err = slaveRows.Err(); err != nil {
		logger.Log.Error("Select slave", zap.Error(err))
		return nil, fmt.Errorf("select slave: %w", err)
	}
	defer slaveRows.Close()
	var guid string
	for slaveRows.Next() {
		err = slaveRows.Scan(&guid)
		if err != nil {
			logger.Log.Error("Scan rows", zap.Error(err))
			return nil, fmt.Errorf("scan rows: %w", err)
		}
		result = append(result, guid)
	}
	//logger.Log.Info("Get slave data goroutine OK", zap.Int("goroutine count:", len(result)))
	//logger.Log.Info("slaveGuids=", zap.Int("cnt", len(c.slaveGuids)))
	return result, nil
}

func (c *Comparator) getSlaveDataParallel(ctx context.Context) error {
	var maxPart = len(c.masterGuids) / c.config.Limit
	logger.Log.Info("maxPart:", zap.Int("maxPart", maxPart+1))
	for part := 0; part <= maxPart; part++ {
		startPos := part * c.config.Limit
		endPos := part*c.config.Limit + c.config.Limit - 1
		if endPos > len(c.masterGuids) {
			endPos = len(c.masterGuids)
		}
		logger.Log.Info("Iter:", zap.Int("part", part+1), zap.Int("startPos", startPos), zap.Int("endPos", endPos))
		val := c.masterGuids[startPos:endPos]
		g := errgroup.Group{}
		g.SetLimit(c.config.RateLimit)
		g.Go(func() error {
			res, err := c.getSlaveData(ctx, val)
			if err != nil {
				return fmt.Errorf("getSlaveData: %w", err)
			}
			c.mutex.Lock()
			c.slaveGuids = append(c.slaveGuids, res...)
			c.mutex.Unlock()
			logger.Log.Info("Get slave data goroutine OK",
				zap.Int("goroutine count:", len(res)),
				zap.Int("slaveGuids count:", len(c.slaveGuids)))
			return nil
		})
		if err := g.Wait(); err != nil {
			logger.Log.Info("getSlaveDataParallel", zap.Error(err))
			return fmt.Errorf("getSlaveDataParallel: %w", err)
		}
	}
	logger.Log.Info("Get slave data full OK", zap.Int(fmt.Sprintf("%v part count:", maxPart+1), len(c.resultGuids)))
	return nil
}
