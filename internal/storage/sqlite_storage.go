package storage

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/LobovVit/CompareFK/internal/config"
	"github.com/LobovVit/CompareFK/internal/result"
	"github.com/LobovVit/CompareFK/pkg/files"
	"github.com/LobovVit/CompareFK/pkg/logger"
	"go.uber.org/zap"
	_ "modernc.org/sqlite"
)

type SQLiteStorage struct {
	db         *sql.DB
	sqlitePath string
}

func GetSQLiteStorage(ctx context.Context) (*SQLiteStorage, error) {
	sqlitePath := config.Cfg.SQLitePath
	if !filepath.IsAbs(sqlitePath) {
		sqlitePath = filepath.Join(result.Res.DateTimeFolder, "work", sqlitePath)
	}
	if err := os.MkdirAll(filepath.Dir(sqlitePath), os.ModePerm); err != nil {
		return nil, fmt.Errorf("mkdir sqlite path: %w", err)
	}

	db, err := sql.Open("sqlite", sqlitePath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	storage := &SQLiteStorage{db: db, sqlitePath: sqlitePath}
	if err := storage.initSchema(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return storage, nil
}

func (s *SQLiteStorage) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	closeErr := s.db.Close()
	if !config.Cfg.SQLiteKeepDB {
		if err := os.Remove(s.sqlitePath); err != nil && !os.IsNotExist(err) {
			logger.Log.Warn("remove sqlite db", zap.Error(err), zap.String("path", s.sqlitePath))
		}
		for _, suffix := range []string{"-wal", "-shm"} {
			if err := os.Remove(s.sqlitePath + suffix); err != nil && !os.IsNotExist(err) {
				logger.Log.Warn("remove sqlite side file", zap.Error(err), zap.String("path", s.sqlitePath+suffix))
			}
		}
	}
	return closeErr
}

func (s *SQLiteStorage) initSchema(ctx context.Context) error {
	pragmas := []string{
		fmt.Sprintf("PRAGMA journal_mode = %s;", config.Cfg.SQLiteJournalMode),
		fmt.Sprintf("PRAGMA synchronous = %s;", config.Cfg.SQLiteSynchronous),
		fmt.Sprintf("PRAGMA temp_store = %s;", config.Cfg.SQLiteTempStore),
		"PRAGMA foreign_keys = OFF;",
		fmt.Sprintf("PRAGMA cache_size = -%d;", config.Cfg.SQLiteCacheSizeKB),
		fmt.Sprintf("PRAGMA mmap_size = %d;", config.Cfg.SQLiteMmapSizeMB*1024*1024),
		fmt.Sprintf("PRAGMA busy_timeout = %d;", config.Cfg.SQLiteBusyTimeoutMs),
	}
	pragmas = append(pragmas, config.Cfg.SQLiteExtraPragmas...)

	for _, pragma := range pragmas {
		if _, err := s.db.ExecContext(ctx, pragma); err != nil {
			return fmt.Errorf("apply pragma %q: %w", pragma, err)
		}
	}

	const schema = `
CREATE TABLE IF NOT EXISTS master_guids (
	guid TEXT PRIMARY KEY
);
CREATE TABLE IF NOT EXISTS matched_guids (
	guid TEXT PRIMARY KEY
);
`
	if _, err := s.db.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("init sqlite schema: %w", err)
	}
	return nil
}

func (s *SQLiteStorage) GetMaster(ctx context.Context, i int, query string, srcDB *sql.DB) error {
	executeFileName := strconv.Itoa(i) + "_master.sql"
	if err := files.WriteSQLFile(executeFileName, query); err != nil {
		logger.Log.Error("write SQL file error", zap.Error(err))
	}

	startTime := time.Now()
	rows, err := srcDB.QueryContext(ctx, query)
	if err != nil {
		return fmt.Errorf("select master query: %w", err)
	}
	defer rows.Close()

	count, err := s.insertIntoTable(ctx, rows, "master_guids")
	if err != nil {
		return fmt.Errorf("load master into sqlite: %w", err)
	}

	result.Res.AddStat(executeFileName, result.ScriptStat{StartTime: startTime, EndTime: time.Now(), Count: count})
	logger.Log.Info("master chunk loaded", zap.String("script", executeFileName), zap.Int("count", count))
	return nil
}

func (s *SQLiteStorage) GetSlave(ctx context.Context, query string, slaveDB *sql.DB) error {
	executeFileName := "slave.sql"
	if err := files.WriteSQLFile(executeFileName, query); err != nil {
		logger.Log.Error("write SQL file error", zap.Error(err))
	}

	chunks, err := s.masterChunkCount(ctx)
	if err != nil {
		return err
	}
	logger.Log.Info("slave compare started", zap.Int("parts", chunks), zap.Int("limit", config.Cfg.Limit))

	for part := 0; part < chunks; part++ {
		startPos := part * config.Cfg.Limit
		endPos := startPos + config.Cfg.Limit

		masterChunk, err := s.loadMasterChunk(ctx, config.Cfg.Limit, startPos)
		if err != nil {
			return err
		}
		if len(masterChunk) == 0 {
			continue
		}
		if endPos > startPos+len(masterChunk) {
			endPos = startPos + len(masterChunk)
		}

		if err := s.addPartSlave(ctx, query, masterChunk, startPos, endPos, slaveDB); err != nil {
			return err
		}
	}

	if config.Cfg.SQLiteRunAnalyze {
		if _, err := s.db.ExecContext(ctx, `ANALYZE`); err != nil {
			logger.Log.Warn("sqlite analyze failed", zap.Error(err))
		}
	}
	if config.Cfg.SQLiteVacuumOnFinish {
		if _, err := s.db.ExecContext(ctx, `VACUUM`); err != nil {
			logger.Log.Warn("sqlite vacuum failed", zap.Error(err))
		}
	}
	return nil
}

func (s *SQLiteStorage) masterChunkCount(ctx context.Context) (int, error) {
	var total int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM master_guids`).Scan(&total); err != nil {
		return 0, fmt.Errorf("count master guids: %w", err)
	}
	if total == 0 {
		return 0, nil
	}
	chunks := total / config.Cfg.Limit
	if total%config.Cfg.Limit != 0 {
		chunks++
	}
	return chunks, nil
}

func (s *SQLiteStorage) loadMasterChunk(ctx context.Context, limit, offset int) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT guid FROM master_guids ORDER BY guid LIMIT ? OFFSET ?`, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("load master chunk: %w", err)
	}
	defer rows.Close()

	res := make([]string, 0, limit)
	for rows.Next() {
		var guid string
		if err := rows.Scan(&guid); err != nil {
			return nil, fmt.Errorf("scan master chunk: %w", err)
		}
		res = append(res, guid)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate master chunk: %w", err)
	}
	return res, nil
}

func (s *SQLiteStorage) addPartSlave(ctx context.Context, query string, chunk []string, startPos, endPos int, slaveDB *sql.DB) error {
	executeFileName := "slave_" + strconv.Itoa(startPos) + "_" + strconv.Itoa(endPos) + ".sql"
	startTime := time.Now()

	rows, err := slaveDB.QueryContext(ctx, query, chunk)
	if err != nil {
		return fmt.Errorf("select slave query: %w", err)
	}
	defer rows.Close()

	count, err := s.insertIntoTable(ctx, rows, "matched_guids")
	if err != nil {
		return fmt.Errorf("load slave chunk into sqlite: %w", err)
	}

	result.Res.AddStat(executeFileName, result.ScriptStat{StartTime: startTime, EndTime: time.Now(), Count: count})
	logger.Log.Info("slave chunk processed",
		zap.String("script", executeFileName),
		zap.Int("chunk_size", len(chunk)),
		zap.Int("matches", count))
	return nil
}

func (s *SQLiteStorage) insertIntoTable(ctx context.Context, rows *sql.Rows, table string) (int, error) {
	batchSize := config.Cfg.SQLiteWriteBatch
	count := 0
	currentBatch := 0

	tx, stmt, err := s.beginInsert(ctx, table)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	commitAndReopen := func() error {
		if err := stmt.Close(); err != nil {
			return fmt.Errorf("close stmt: %w", err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit tx: %w", err)
		}
		tx, stmt, err = s.beginInsert(ctx, table)
		if err != nil {
			return err
		}
		currentBatch = 0
		return nil
	}

	for rows.Next() {
		var guid string
		if err := rows.Scan(&guid); err != nil {
			_ = tx.Rollback()
			return 0, fmt.Errorf("scan row: %w", err)
		}
		guid = strings.TrimSpace(guid)
		if guid == "" {
			continue
		}
		if _, err := stmt.ExecContext(ctx, guid); err != nil {
			_ = tx.Rollback()
			return 0, fmt.Errorf("insert guid into %s: %w", table, err)
		}
		count++
		currentBatch++

		if currentBatch >= batchSize {
			if err := commitAndReopen(); err != nil {
				return 0, err
			}
		}
	}
	if err := rows.Err(); err != nil {
		_ = tx.Rollback()
		return 0, fmt.Errorf("iterate rows: %w", err)
	}
	if err := stmt.Close(); err != nil {
		_ = tx.Rollback()
		return 0, fmt.Errorf("close stmt: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit tx: %w", err)
	}
	return count, nil
}

func (s *SQLiteStorage) beginInsert(ctx context.Context, table string) (*sql.Tx, *sql.Stmt, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("begin sqlite tx: %w", err)
	}
	stmt, err := tx.PrepareContext(ctx, fmt.Sprintf(`INSERT OR IGNORE INTO %s(guid) VALUES (?)`, table))
	if err != nil {
		_ = tx.Rollback()
		return nil, nil, fmt.Errorf("prepare insert %s: %w", table, err)
	}
	return tx, stmt, nil
}

func (s *SQLiteStorage) WriteResult(ctx context.Context, outputFile string) error {
	executeStep := "z_compute_" + config.Cfg.Mode
	startTime := time.Now()

	query := s.resultQuery()
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return fmt.Errorf("query result: %w", err)
	}
	defer rows.Close()

	path := filepath.Join(result.Res.DateTimeFolder, outputFile)
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create result file: %w", err)
	}
	defer file.Close()

	writer := bufio.NewWriterSize(file, 1024*1024)
	count := 0
	for rows.Next() {
		var guid string
		if err := rows.Scan(&guid); err != nil {
			return fmt.Errorf("scan result guid: %w", err)
		}
		if _, err := writer.WriteString(guid); err != nil {
			return fmt.Errorf("write result guid: %w", err)
		}
		if _, err := writer.WriteString("\r\n"); err != nil {
			return fmt.Errorf("write result delimiter: %w", err)
		}
		count++
		if config.Cfg.SQLiteResultFetchSize > 0 && count%config.Cfg.SQLiteResultFetchSize == 0 {
			if err := writer.Flush(); err != nil {
				return fmt.Errorf("flush result writer: %w", err)
			}
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("result rows: %w", err)
	}
	if err := writer.Flush(); err != nil {
		return fmt.Errorf("flush result writer: %w", err)
	}

	result.Res.AddStat(executeStep, result.ScriptStat{StartTime: startTime, EndTime: time.Now(), Count: count})
	logger.Log.Info("result written", zap.String("file", path), zap.Int("count", count))
	return nil
}

func (s *SQLiteStorage) resultQuery() string {
	switch config.Cfg.Mode {
	case "intersection":
		return `SELECT guid FROM matched_guids ORDER BY guid`
	default:
		return `
SELECT m.guid
FROM master_guids m
LEFT JOIN matched_guids mt ON mt.guid = m.guid
WHERE mt.guid IS NULL
ORDER BY m.guid`
	}
}
