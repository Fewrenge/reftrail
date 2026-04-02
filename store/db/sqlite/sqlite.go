package sqlite

import (
	"context"
	"database/sql"
	"os"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

// Driver is the actual worker.
type Driver struct {
	db *sql.DB

	// mu is a "Lock." Since SQLite is a single file, we use this
	// to make sure two people don't try to write at the exact same millisecond.
	mu sync.Mutex
}

func (d *Driver) Migrate(ctx context.Context) error {
	// 1. Read your SQL file (Make sure the path is correct!)
	script, err := os.ReadFile("store/migration/sqlite/00__init.sql")
	if err != nil {
		return err
	}

	// 2. Execute the SQL commands
	_, err = d.db.ExecContext(ctx, string(script))
	return err
}

// New opens the connection to the .db file.
func New(dbPath string) (*Driver, error) {
	// 1. Tell Go to open (or create) the file
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1) // SQLite works best if only one person writes at a time

	return &Driver{
		db: db,
	}, nil
}

// GetDB lets the Manager see the raw connection if needed.
func (d *Driver) GetDB() *sql.DB {
	return d.db
}

// Close safely shuts down the worker.
func (d *Driver) Close() error {
	return d.db.Close()
}
