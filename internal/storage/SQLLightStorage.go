package storage

import "database/sql"

// TODO SQLLightStorage
type SQLLightStorage struct {
	db *sql.DB
}
