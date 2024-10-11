package storage

import (
	"context"
	"database/sql"
)

// TODO SQLLightStorage
type SQLLightStorage struct {
	db *sql.DB
}

func GetSQLLightStorage() *SQLLightStorage {
	return &SQLLightStorage{db: nil}
}
func (s *SQLLightStorage) GetMaster(ctx context.Context, i int, sql string, db *sql.DB) error {
	_, err := s.db.ExecContext(ctx, sql)
	return err
}

func (s *SQLLightStorage) GetSlave(ctx context.Context, sql string, db *sql.DB) error {
	_, err := s.db.ExecContext(ctx, sql)
	return err
}
func (s *SQLLightStorage) GetResult(ctx context.Context) []string {
	return make([]string, 0)
}
