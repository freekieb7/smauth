package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNoRows = pgx.ErrNoRows
)

type Database struct {
	Conn *pgxpool.Pool
}

func New() Database {
	return Database{
		Conn: nil,
	}
}

func (d *Database) ConnectWithRetry(connString string, maxRetries int, delay time.Duration) error {
	var err error
	for i := range maxRetries {
		err = d.Connect(connString)
		if err == nil {
			return nil
		}

		if i < maxRetries-1 {
			fmt.Printf("Failed to connect to database (attempt %d/%d): %v\n", i+1, maxRetries, err)
			time.Sleep(delay)
		}
	}
	return fmt.Errorf("failed to connect after %d attempts: %w", maxRetries, err)
}

func (db *Database) Connect(connString string) error {
	pool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		return err
	}
	db.Conn = pool
	return nil
}

func (db *Database) Close() {
	db.Conn.Close()
}

func (db *Database) Ping(ctx context.Context) error {
	return db.Conn.Ping(ctx)
}
