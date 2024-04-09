package db

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/sijms/go-ora/v2"
)

func NweConn(dsn string) (*sql.DB, error) {
	switch dsn[0:6] {
	case "postgr":
		return newConnPG(dsn)
	case "oracle":
		return newConnOra(dsn)
	default:
		return nil, fmt.Errorf("dsn error, must be postgresql or oracle ")
	}
}

func newConnPG(dsn string) (*sql.DB, error) {
	conn, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("DB open: %w", err)
	}
	return conn, nil
}

func newConnOra(dsn string) (*sql.DB, error) {
	conn, err := sql.Open("oracle", dsn)
	if err != nil {
		return nil, fmt.Errorf("DB open: %w", err)
	}
	return conn, nil
}
