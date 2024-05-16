package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

const dbVersion = 1

func openDB(cfg config) (*sql.DB, error) {

	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.db.maxOpenConns)
	db.SetMaxIdleConns(cfg.db.maxIdleConns)
	db.SetConnMaxIdleTime(cfg.db.maxIdleTime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var version int

	err = db.QueryRowContext(ctx, "SELECT version FROM migrations LIMIT 1").Scan(&version)
	if err != nil {
		db.Close()
		return nil, err
	}

	if version != dbVersion {
		db.Close()
		return nil, fmt.Errorf("db version mismatch: expected %d, got %d", dbVersion, version)
	}

	return db, nil
}
