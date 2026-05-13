package store

import (
	"context"
	"log/slog"
)

// Store is the "Manager" that handles the database and cache.
type Store struct {
	driver Driver // This points to the interface in driver.go

	// Later, you'll add caches here, like:
	// userCache *cache.Cache
}

// New creates a new Manager (Store) and gives them a Worker (Driver).
func NewStore(driver Driver) *Store {
	s := &Store{driver: driver}
	// Fire the seeding logic right when the manager starts
	if err := s.SeedAdminUser(context.Background()); err != nil {
		slog.Warn("Warning: Failed to seed admin user: ", "error", err.Error())
	}
	return s
}
