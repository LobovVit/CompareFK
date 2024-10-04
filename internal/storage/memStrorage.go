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
		logger.Log.Error("select master query", zap.Error(err))
		return fmt.Errorf("select master query: %w", err)
	}
	if err = masterRows.Err(); err != nil {
		logger.Log.Error("select master rows", zap.Error(err))
		return fmt.Errorf("select master rows: %w", err)
	}
	defer masterRows.Close()

	var tmpMasterGuids []string
	for masterRows.Next() {
		var guid string
		err = masterRows.Scan(&guid)
		if err != nil {
			logger.Log.Error("select master (Scan)", zap.Error(err))
			return fmt.Errorf("select master (Scan): %w", err)
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
	logger.Log.Info("max part:", zap.Int("maxPart", maxPart+1))
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
				logger.Log.Error("add part slave", zap.Error(err))
				return fmt.Errorf("add part slave: %w", err)
			}
			m.mutex.Lock()
			m.slaveGuids = append(m.slaveGuids, res...)
			m.mutex.Unlock()
			logger.Log.Info("Get slave data goroutine OK",
				zap.Int("goroutine count:", len(res)),
				zap.Int("slave guids count:", len(m.slaveGuids)))
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		logger.Log.Error("get slave data parallel", zap.Error(err))
		return fmt.Errorf("get slave data parallel: %w", err)
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
		logger.Log.Error("select slave query", zap.Error(err))
		return nil, fmt.Errorf("select slave query: %w", err)
	}
	if err = slaveRows.Err(); err != nil {
		logger.Log.Error("select slave rows", zap.Error(err))
		return nil, fmt.Errorf("select slave rows: %w", err)
	}
	defer slaveRows.Close()
	var guid string
	for slaveRows.Next() {
		err = slaveRows.Scan(&guid)
		if err != nil {
			logger.Log.Error("select slave (Scan)", zap.Error(err))
			return nil, fmt.Errorf("select slave (Scan): %w", err)
		}
		res = append(res, guid)
	}
	endTime := time.Now()
	count := len(res)
	result.Res.StatRows[executeFileName] = result.ScriptStat{StartTime: startTime, EndTime: endTime, Count: count}
	logger.Log.Debug(result.Res.GetResultString("add part slave"))
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
		logger.Log.Error("mode is incorrect")
		return nil
	}
	endTime := time.Now()
	count := len(resultGuids)
	result.Res.StatRows[executeStep] = result.ScriptStat{StartTime: startTime, EndTime: endTime, Count: count}
	logger.Log.Debug(result.Res.GetResultString("get result"))
	return resultGuids
}
