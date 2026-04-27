package main

import (
	"context"
	"log"
	"reftrail/server"
	"reftrail/store"
	sqlite "reftrail/store/db/sqlite"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// 1. Hire the Worker (Open SQLite)
	// This will create 'wl.db' in your folder if it doesn't exist
	driver, err := sqlite.New("wl.db")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	if err := driver.Migrate(context.Background()); err != nil {
		log.Fatalf("Failed to run migration: %v", err)
	}

	// 2. Hire the Manager (Create Store)
	st := store.New(driver)

	// 3. Open the Front Desk (Create Server)
	srv := server.NewServer(st)

	// 4. Turn on the "Open" sign
	log.Println("Waitlist System started on http://localhost:8080")
	if err := srv.Start(":8080"); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
