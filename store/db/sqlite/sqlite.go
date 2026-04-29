package sqlite

import (
	"context"
	"database/sql"
	"os"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

type contextKey string // Private
const txKey contextKey = "tx"

// commonExec allows us to use either *sql.DB or *sql.Tx interchangeably
type commonExec interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// conn is the "magic" helper. It checks if there is a transaction in the context.
func (d *Driver) conn(ctx context.Context) commonExec {
	if tx, ok := ctx.Value(txKey).(*sql.Tx); ok {
		return tx
	}
	return d.db
}

// RunInTransaction is your safety container
func (d *Driver) RunInTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	_, _ = tx.Exec("BEGIN IMMEDIATE")

	txCtx := context.WithValue(ctx, txKey, tx)

	if err := fn(txCtx); err != nil {
		// Only rollback if the context is still "alive"
		// If ctx.Err() is not nil, the driver already rolled back for us
		if ctx.Err() == nil {
			tx.Rollback()
		}
		return err
	}

	// Only commit if the context hasn't been cancelled/timed out
	if err := ctx.Err(); err != nil {
		return err
	}

	return tx.Commit()
}

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

	// _, _ = db.Exec("PRAGMA journal_mode=WAL;")

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
