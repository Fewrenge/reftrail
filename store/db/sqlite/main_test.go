package sqlite

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"reftrail/store"
	"testing"
)

func setupTestStore(t *testing.T) *store.Store {
	// 1. Open a fresh in-memory database with random name
	b := make([]byte, 4)
	rand.Read(b)
	dbName := hex.EncodeToString(b)

	// Use the unique name in the DSN
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared&_fk=1", dbName)
	db, err := sql.Open("sqlite3", dsn)

	t.Cleanup(func() {
		db.Close()
	})

	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	db.SetMaxOpenConns(1)

	// 3. Initialize real Driver and Store using this memory DB
	driver := NewWithDB(db)

	if err := driver.Migrate(context.Background()); err != nil {
		t.Fatalf("failed to run production schema migration sequence on test DB: %v", err)
	}

	// 4. Initialize Store using the freshly migrated driver
	return store.NewStore(driver)
}
