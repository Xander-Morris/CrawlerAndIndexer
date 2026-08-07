package database

import (
	"context"
	"database/sql"
	"fmt"
)

var cachedDb *sql.DB = nil 

func getDb() (*sql.DB, error) {
	if cachedDb != nil {
		return cachedDb, nil
	}

	db, err := sql.Open("sqlite", databaseFileName)

	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	cachedDb = db

	return cachedDb, nil
}

func Ping(ctx context.Context) error {
	db, err := getDb()

	if err != nil {
		return err
	}

	return db.PingContext(ctx)
}
