package storage

import "database/sql"

type SQLLightStorage struct {
	db *sql.DB
}
