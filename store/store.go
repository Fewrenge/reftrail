package store

// Store is the "Manager" that handles the database and cache.
type Store struct {
	driver Driver // This points to the interface in driver.go

	// Later, you'll add caches here, like:
	// userCache *cache.Cache
}

// New creates a new Manager (Store) and gives them a Worker (Driver).
func New(driver Driver) *Store {
	return &Store{
		driver: driver,
	}
}
