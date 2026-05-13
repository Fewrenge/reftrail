package main

import (
	"context"
	"log/slog"
	"os"
	"reftrail/server"
	"reftrail/store"
	sqlite "reftrail/store/db/sqlite"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Initialize the global JSON structure logger for production
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	// Open SQLite
	// This will create 'reftrail.db' if it doesn't exist
	driver, err := sqlite.New("reftrail.db")
	if err != nil {
		slog.Error("Failed to open database during startup", "error", err.Error())
		os.Exit(1)
	}

	if err := driver.Migrate(context.Background()); err != nil {
		slog.Error("Database schema migration sequence failed", "error", err.Error())
		os.Exit(1)
	}

	// 2. Hire the Manager (Create Store)
	st := store.NewStore(driver)

	// 3. Open the Front Desk (Create Server)
	srv := server.NewServer(st)

	// 4. Turn on the "Open" sign
	slog.Info("Referral system started",
		"address", "http://localhost:8080",
		"port", 8080,
	)
	if err := srv.Start(":8080"); err != nil {
		slog.Error("Web server execution crashed", "error", err.Error())
		os.Exit(1)
	}
}
