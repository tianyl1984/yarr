package storage

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

type Storage struct {
	db *sql.DB
}

func New(dsn string) (*Storage, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	if err = migrate(db); err != nil {
		return nil, err
	}
	return &Storage{db: db}, nil
}
