package main

import (
	"context"
	"log"
	"wl/server"
	"wl/store"
	sqlite "wl/store/db/sqlite"

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

	// 2.5 Bootstrap admin
	err = bootstrapAdmin(context.Background(), st)
	if err != nil {
		log.Printf("Bootstrap error: %v", err)
	}

	// 3. Open the Front Desk (Create Server)
	srv := server.NewServer(st)

	// 4. Turn on the "Open" sign
	log.Println("Waitlist System started on http://localhost:8080")
	if err := srv.Start(":8080"); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func bootstrapAdmin(ctx context.Context, s *store.Store) error {
	// 1. Check if any users exist
	users, err := s.ListUsers(ctx, &store.FindUser{})
	if err != nil {
		return err
	}

	// 2. If no users, create the 'admin'
	if len(users) == 0 {
		log.Println("No users found. Creating default admin account...")
		_, err := s.CreateUser(ctx, &store.CreateUser{
			Username: "admin",
			Password: "password123", // In real life, use a secure password!
			Role:     "ADMIN",
		})
		return err
	}

	return nil
}
