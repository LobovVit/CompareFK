package db

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/LobovVit/CompareFK/internal/config"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/sijms/go-ora/v2"
)

func NweConn(dsn string, maxOpenConns int) (*sql.DB, error) {
	rawDSN := strings.TrimSpace(dsn)
	lowerDSN := strings.ToLower(rawDSN)
	switch {
	case strings.HasPrefix(lowerDSN, "postgresql"), strings.HasPrefix(lowerDSN, "postgres"):
		return newConnPG(rawDSN, maxOpenConns)
	case strings.HasPrefix(lowerDSN, "oracle"):
		return newConnOra(rawDSN, maxOpenConns)
	default:
		return nil, fmt.Errorf("dsn error, must start with postgresql or oracle")
	}
}

func newConnPG(dsn string, maxOpenConns int) (*sql.DB, error) {
	conn, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("DB open: %w", err)
	}
	applyPool(conn, maxOpenConns)
	return conn, nil
}

func newConnOra(dsn string, maxOpenConns int) (*sql.DB, error) {
	conn, err := sql.Open("oracle", dsn)
	if err != nil {
		return nil, fmt.Errorf("DB open: %w", err)
	}
	applyPool(conn, maxOpenConns)
	return conn, nil
}

func applyPool(conn *sql.DB, maxOpenConns int) {
	if maxOpenConns <= 0 {
		maxOpenConns = 1
	}
	conn.SetMaxOpenConns(maxOpenConns)
	conn.SetMaxIdleConns(min(maxOpenConns, 2))
	conn.SetConnMaxIdleTime(0)
	conn.SetConnMaxLifetime(0)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

var _ = config.Cfg
