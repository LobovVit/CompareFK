package config

import (
	"errors"
	"flag"
	"log"
	"path/filepath"
	"strings"

	"github.com/ilyakaznacheev/cleanenv"
)

var Cfg *Config

type Config struct {
	Mode            string `yaml:"mode"`
	LogLevel        string `yaml:"loglevel"`
	Limit           int    `yaml:"limit"`
	RateLimit       int    `yaml:"ratelimit"`
	Storage         string `yaml:"storage"`
	OutputDir       string `yaml:"output_dir"`
	AllowedEmptyDSN bool   `yaml:"allowed_empty_dsn"`
	ConfigFile      string

	MasterDSN          string `yaml:"masterdsn"`
	SlaveDSN           string `yaml:"slavedsn"`
	MaxOpenConnsMaster int    `yaml:"max_open_conns_master"`
	MaxOpenConnsSlave  int    `yaml:"max_open_conns_slave"`

	MasterSQLDir   string   `yaml:"master_sql_dir"`
	MasterSQLGlob  string   `yaml:"master_sql_glob"`
	MasterSQLFiles []string `yaml:"master_sql_files"`
	SlaveSQLFile   string   `yaml:"slave_sql_file"`

	SQLitePath            string   `yaml:"sqlite_path"`
	SQLiteBusyTimeoutMs   int      `yaml:"sqlite_busy_timeout_ms"`
	SQLiteCacheSizeKB     int      `yaml:"sqlite_cache_size_kb"`
	SQLiteMmapSizeMB      int      `yaml:"sqlite_mmap_size_mb"`
	SQLiteWriteBatch      int      `yaml:"sqlite_write_batch"`
	SQLiteResultFetchSize int      `yaml:"sqlite_result_fetch_size"`
	SQLiteTempStore       string   `yaml:"sqlite_temp_store"`
	SQLiteSynchronous     string   `yaml:"sqlite_synchronous"`
	SQLiteJournalMode     string   `yaml:"sqlite_journal_mode"`
	SQLiteKeepDB          bool     `yaml:"sqlite_keep_db"`
	SQLiteRunAnalyze      bool     `yaml:"sqlite_run_analyze"`
	SQLiteVacuumOnFinish  bool     `yaml:"sqlite_vacuum_on_finish"`
	SQLiteExtraPragmas    []string `yaml:"sqlite_extra_pragmas"`

	WebEnabled      bool   `yaml:"web_enabled"`
	WebListen       string `yaml:"web_listen"`
	WebRefreshSec   int    `yaml:"web_refresh_sec"`
	WebReadTimeout  int    `yaml:"web_read_timeout_sec"`
	WebWriteTimeout int    `yaml:"web_write_timeout_sec"`
}

func Initialize() error {
	cfg, err := getConfig()
	Cfg = cfg
	return err
}

func getConfig() (*Config, error) {
	log.Print("read config")
	cfgFile := flag.String("c", "config.yml", "путь к конфиг файлу")
	flag.Parse()

	cfg := &Config{}
	if err := cleanenv.ReadConfig(*cfgFile, cfg); err != nil {
		return nil, err
	}
	cfg.ConfigFile = *cfgFile

	cfg.Mode = strings.TrimSpace(strings.ToLower(cfg.Mode))
	cfg.Storage = strings.TrimSpace(strings.ToLower(cfg.Storage))
	cfg.OutputDir = strings.TrimSpace(cfg.OutputDir)
	cfg.MasterDSN = strings.TrimSpace(cfg.MasterDSN)
	cfg.SlaveDSN = strings.TrimSpace(cfg.SlaveDSN)
	cfg.MasterSQLDir = strings.TrimSpace(cfg.MasterSQLDir)
	cfg.MasterSQLGlob = strings.TrimSpace(cfg.MasterSQLGlob)
	cfg.SlaveSQLFile = strings.TrimSpace(cfg.SlaveSQLFile)
	cfg.SQLitePath = strings.TrimSpace(cfg.SQLitePath)
	cfg.SQLiteTempStore = strings.TrimSpace(strings.ToUpper(cfg.SQLiteTempStore))
	cfg.SQLiteSynchronous = strings.TrimSpace(strings.ToUpper(cfg.SQLiteSynchronous))
	cfg.SQLiteJournalMode = strings.TrimSpace(strings.ToUpper(cfg.SQLiteJournalMode))
	cfg.WebListen = strings.TrimSpace(cfg.WebListen)
	for i := range cfg.MasterSQLFiles {
		cfg.MasterSQLFiles[i] = strings.TrimSpace(cfg.MasterSQLFiles[i])
	}

	if cfg.Mode != "difference" && cfg.Mode != "intersection" {
		return nil, errors.New("укажите mode: \"difference\" или \"intersection\"")
	}
	if !cfg.AllowedEmptyDSN && cfg.MasterDSN == "" {
		return nil, errors.New("укажите masterdsn")
	}
	if !cfg.AllowedEmptyDSN && cfg.SlaveDSN == "" {
		return nil, errors.New("укажите slavedsn")
	}
	if cfg.LogLevel == "" {
		cfg.LogLevel = "info"
	}
	if cfg.Limit <= 0 {
		cfg.Limit = 50000
	}
	if cfg.RateLimit <= 0 {
		cfg.RateLimit = 4
	}
	if cfg.Storage == "" {
		cfg.Storage = "sqlite"
	}
	if cfg.Storage != "sqlite" {
		return nil, errors.New("поддерживается только storage: \"sqlite\"")
	}
	if cfg.OutputDir == "" {
		cfg.OutputDir = "./runs"
	}
	if cfg.MasterSQLDir == "" && cfg.MasterSQLGlob == "" && len(nonEmpty(cfg.MasterSQLFiles)) == 0 {
		cfg.MasterSQLDir = "./sql/master"
	}
	cfg.MasterSQLFiles = nonEmpty(cfg.MasterSQLFiles)
	if cfg.SlaveSQLFile == "" {
		cfg.SlaveSQLFile = "./sql/slave/Slave.sql"
	}
	if cfg.SQLitePath == "" {
		cfg.SQLitePath = filepath.Join("work", "compare.db")
	}
	if cfg.SQLiteBusyTimeoutMs <= 0 {
		cfg.SQLiteBusyTimeoutMs = 5000
	}
	if cfg.SQLiteCacheSizeKB <= 0 {
		cfg.SQLiteCacheSizeKB = 65536
	}
	if cfg.SQLiteMmapSizeMB <= 0 {
		cfg.SQLiteMmapSizeMB = 256
	}
	if cfg.SQLiteWriteBatch <= 0 {
		cfg.SQLiteWriteBatch = 20000
	}
	if cfg.SQLiteResultFetchSize <= 0 {
		cfg.SQLiteResultFetchSize = 50000
	}
	if cfg.SQLiteTempStore == "" {
		cfg.SQLiteTempStore = "FILE"
	}
	if cfg.SQLiteSynchronous == "" {
		cfg.SQLiteSynchronous = "NORMAL"
	}
	if cfg.SQLiteJournalMode == "" {
		cfg.SQLiteJournalMode = "WAL"
	}
	if cfg.MaxOpenConnsMaster <= 0 {
		cfg.MaxOpenConnsMaster = cfg.RateLimit
	}
	if cfg.MaxOpenConnsSlave <= 0 {
		cfg.MaxOpenConnsSlave = 1
	}
	if cfg.WebListen == "" {
		cfg.WebListen = ":8080"
	}
	if cfg.WebRefreshSec <= 0 {
		cfg.WebRefreshSec = 2
	}
	if cfg.WebReadTimeout <= 0 {
		cfg.WebReadTimeout = 5
	}
	if cfg.WebWriteTimeout <= 0 {
		cfg.WebWriteTimeout = 30
	}

	return cfg, nil
}

func nonEmpty(items []string) []string {
	res := make([]string, 0, len(items))
	for _, item := range items {
		if strings.TrimSpace(item) != "" {
			res = append(res, strings.TrimSpace(item))
		}
	}
	return res
}
