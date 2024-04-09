package compare

import (
	"context"
	"fmt"
	"time"

	"Compare/internal/config"
	"Compare/internal/files"
	"Compare/pkg/db"
	"Compare/pkg/logger"
	"go.uber.org/zap"
)

type Comparator struct {
	mastedDSN   string
	slavedDSN   string
	masterSQL   string
	slaveSQL    string
	attrsSQL    string
	masterGuids []string
	slaveGuids  []string
	attrsSource string
	attrsData   []string
	mode        string
	resultGuids []string
}

func NewComparator(cfg *config.Config) (*Comparator, error) {
	mSQL, err := files.ReadFile("Master.sql")
	if err != nil {
		logger.Log.Error("Ошибка ReadFile", zap.Error(err))
	}
	sSQL, err := files.ReadFile("Slave.sql")
	if err != nil {
		logger.Log.Error("Ошибка ReadFile", zap.Error(err))
	}
	aSQL, err := files.ReadFile("Attrs.sql")
	if err != nil {
		logger.Log.Error("Ошибка ReadFile", zap.Error(err))
	}
	return &Comparator{mastedDSN: cfg.Masterdsn,
		slavedDSN:   cfg.Slavedsn,
		masterSQL:   mSQL,
		slaveSQL:    sSQL,
		attrsSQL:    aSQL,
		masterGuids: make([]string, 0),
		slaveGuids:  make([]string, 0),
		attrsSource: cfg.Attrs,
		attrsData:   make([]string, 0),
		mode:        cfg.Мode,
		resultGuids: make([]string, 0)}, nil
}

func (c *Comparator) Run(ctx context.Context) error {
	err := c.getMasterData(ctx)
	if err != nil {
		logger.Log.Error("Get master data", zap.Error(err))
		return fmt.Errorf("get master data: %w", err)
	}
	err = c.getSlaveData(ctx)
	if err != nil {
		logger.Log.Error("Get slave data", zap.Error(err))
		return fmt.Errorf("get slave data: %w", err)
	}

	switch c.mode {
	case "intersection":
		c.resultGuids = intersection(c.masterGuids, c.slaveGuids)
		logger.Log.Info("result=", zap.Int(c.mode, len(c.resultGuids)))
	case "difference":
		c.resultGuids = difference(c.masterGuids, c.slaveGuids)
		logger.Log.Info("result=", zap.Int(c.mode, len(c.resultGuids)))
	default:
		logger.Log.Info("mode is incorrect")
		return fmt.Errorf("mode is incorrect")
	}

	err = files.WriteFile(time.Now().Format(time.DateTime)+"_result_guids_"+c.mode+".txt", c.resultGuids)
	if err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	if len(c.attrsSQL) > 20 {
		err = c.getAttrsData(ctx)
		if err != nil {
			logger.Log.Error("get attrs data", zap.Error(err))
		}
	}

	return nil
}

func (c *Comparator) getMasterData(ctx context.Context) error {
	//get master data
	mastedDB, err := db.NweConn(c.mastedDSN)
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

func (c *Comparator) getSlaveData(ctx context.Context) error {
	//get slave data
	slaveDB, err := db.NweConn(c.slavedDSN)
	if err != nil {
		logger.Log.Error("Ошибка connSlave", zap.Error(err))
	}

	slaveRows, err := slaveDB.QueryContext(ctx, c.slaveSQL, c.masterGuids)
	if err != nil {
		logger.Log.Error("Select slave", zap.Error(err))
		return fmt.Errorf("select slave: %w", err)
	}
	if err = slaveRows.Err(); err != nil {
		logger.Log.Error("Select slave", zap.Error(err))
		return fmt.Errorf("select slave: %w", err)
	}
	defer slaveRows.Close()
	var guid string
	for slaveRows.Next() {
		err = slaveRows.Scan(&guid)
		if err != nil {
			logger.Log.Error("Scan rows", zap.Error(err))
			return fmt.Errorf("scan rows: %w", err)
		}
		c.slaveGuids = append(c.slaveGuids, guid)
	}
	logger.Log.Info("slaveGuids=", zap.Int("cnt", len(c.slaveGuids)))
	return nil
}

func (c *Comparator) getAttrsData(ctx context.Context) error {
	var attrsDSN string
	switch c.attrsSource {
	case "slave":
		attrsDSN = c.slavedDSN
	case "master":
		attrsDSN = c.mastedDSN
	default:
		return fmt.Errorf("no need attrs")
	}

	attrsDB, err := db.NweConn(attrsDSN)
	if err != nil {
		logger.Log.Error("Conn attrs", zap.Error(err))
		return fmt.Errorf("conn attrs: %w", err)
	}

	attrsRows, err := attrsDB.QueryContext(ctx, c.attrsSQL, c.resultGuids)
	if err != nil {
		logger.Log.Error("Select slave", zap.Error(err))
		return fmt.Errorf("select slave: %w", err)
	}
	if err = attrsRows.Err(); err != nil {
		logger.Log.Error("Select slave", zap.Error(err))
		return fmt.Errorf("select slave: %w", err)
	}
	defer attrsRows.Close()

	var item string
	for attrsRows.Next() {
		err = attrsRows.Scan(&item)
		if err != nil {
			logger.Log.Error("Scan rows", zap.Error(err))
			return fmt.Errorf("scan rows: %w", err)
		}
		c.attrsData = append(c.attrsData, item)
	}
	logger.Log.Info("attrsData=", zap.Int("cnt", len(c.attrsData)))

	if len(c.attrsData) > 0 {
		err = files.WriteFile(time.Now().Format(time.DateTime)+"_result_attrs_"+c.mode+".txt", c.attrsData)
		if err != nil {
			logger.Log.Info("Write file result attrs", zap.Error(err))
			return fmt.Errorf("write file: %w", err)
		}
	}
	return nil
}
