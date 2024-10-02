package storage

import (
	"Compare/internal/config"
	"Compare/internal/result"
	"Compare/pkg/compare"
	"Compare/pkg/files"
	"Compare/pkg/logger"
	"context"
	"database/sql"
	"fmt"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"strconv"
	"sync"
	"time"
)

type MemStorage struct {
	masterGuids []string
	slaveGuids  []string
	resultGuids []string
	mutex       sync.Mutex
}

func GetMemStorage() *MemStorage {
	return &MemStorage{masterGuids: make([]string, 0),
		slaveGuids:  make([]string, 0),
		resultGuids: make([]string, 0)}
}

func (m *MemStorage) GetMaster(ctx context.Context, i int, sql string, db *sql.DB) error {
	executeFileName := strconv.Itoa(i) + "_master" + ".sql"
	if err := files.WriteSQLFile(executeFileName, sql); err != nil {
		logger.Log.Error("write SQL file error", zap.Error(err))
	}
	startTime := time.Now()
	masterRows, err := db.QueryContext(ctx, sql)
	if err != nil {
		return fmt.Errorf("select master: %w", err)
	}
	if err = masterRows.Err(); err != nil {
		return fmt.Errorf("select master: %w", err)
	}
	defer masterRows.Close()

	var tmpMasterGuids []string
	for masterRows.Next() {
		var guid string
		err = masterRows.Scan(&guid)
		if err != nil {
			return fmt.Errorf("select slave: %w", err)
		}
		tmpMasterGuids = append(tmpMasterGuids, guid)
	}
	m.mutex.Lock()
	m.masterGuids = append(m.masterGuids, tmpMasterGuids...)
	m.mutex.Unlock()
	endTime := time.Now()
	count := len(tmpMasterGuids)
	result.Res.StatRows[executeFileName] = result.ScriptStat{StartTime: startTime, EndTime: endTime, Count: count}
	logger.Log.Debug(result.Res.GetResultString("GetMaster_" + strconv.Itoa(i)))
	return nil
}
func (m *MemStorage) GetSlave(ctx context.Context, sql string, db *sql.DB) error {
	executeFileName := "slave" + ".sql"
	if err := files.WriteSQLFile(executeFileName, sql); err != nil {
		logger.Log.Error("write SQL file error", zap.Error(err))
	}
	var maxPart = len(m.masterGuids) / config.Cfg.Limit
	logger.Log.Info("maxPart:", zap.Int("maxPart", maxPart+1))
	g := errgroup.Group{}
	g.SetLimit(config.Cfg.RateLimit)
	for part := 0; part <= maxPart; part++ {
		startPos := part * config.Cfg.Limit
		endPos := part*config.Cfg.Limit + config.Cfg.Limit /*- 1*/
		if endPos > len(m.masterGuids) {
			endPos = len(m.masterGuids)
		}
		logger.Log.Info("Iter:", zap.Int("part", part+1), zap.Int("startPos", startPos), zap.Int("endPos", endPos))
		val := m.masterGuids[startPos:endPos]
		g.Go(func() error {
			res, err := m.addPartSlave(ctx, sql, val, startPos, endPos, db)
			if err != nil {
				return fmt.Errorf("addPartSlave: %w", err)
			}
			m.mutex.Lock()
			m.slaveGuids = append(m.slaveGuids, res...)
			m.mutex.Unlock()
			logger.Log.Info("Get slave data goroutine OK",
				zap.Int("goroutine count:", len(res)),
				zap.Int("slaveGuids count:", len(m.slaveGuids)))
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		logger.Log.Info("getSlaveDataParallel", zap.Error(err))
		return fmt.Errorf("getSlaveDataParallel: %w", err)
	}
	logger.Log.Info("Get slave data full OK", zap.Int(fmt.Sprintf("%v part count:", maxPart+1), len(m.slaveGuids)))
	return nil
}
func (m *MemStorage) addPartSlave(ctx context.Context, sql string, val []string, startPos int, endPos int, db *sql.DB) ([]string, error) {
	executeFileName := "slave_" + strconv.Itoa(startPos) + "_" + strconv.Itoa(endPos) + ".sql"
	startTime := time.Now()
	res := []string{}
	slaveRows, err := db.QueryContext(ctx, sql, val)
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
		res = append(res, guid)
	}
	endTime := time.Now()
	count := len(res)
	result.Res.StatRows[executeFileName] = result.ScriptStat{StartTime: startTime, EndTime: endTime, Count: count}
	logger.Log.Debug(result.Res.GetResultString("addPartSlave"))
	return res, nil
}

func (m *MemStorage) GetResult(ctx context.Context) []string {
	executeStep := "z_compute_" + config.Cfg.Мode
	startTime := time.Now()
	var resultGuids []string
	switch config.Cfg.Мode {
	case "intersection":
		resultGuids = compare.Intersection(m.masterGuids, m.slaveGuids)
	case "difference":
		resultGuids = compare.Difference(m.masterGuids, m.slaveGuids)
	default:
		logger.Log.Info("mode is incorrect")
		return nil
	}
	endTime := time.Now()
	count := len(resultGuids)
	result.Res.StatRows[executeStep] = result.ScriptStat{StartTime: startTime, EndTime: endTime, Count: count}
	logger.Log.Debug(result.Res.GetResultString("GetResult"))
	return resultGuids
}
