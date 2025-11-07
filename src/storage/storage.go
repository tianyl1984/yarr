package storage

import (
	"database/sql"
	"time"

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
	db.SetConnMaxIdleTime(time.Minute * 20)
	db.SetConnMaxLifetime(time.Hour)
	db.SetMaxIdleConns(2)
	db.SetMaxOpenConns(5)
	if err = migrate(db); err != nil {
		return nil, err
	}
	return &Storage{db: db}, nil
}
