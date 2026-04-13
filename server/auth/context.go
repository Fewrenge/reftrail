package auth

// ContextKey is a custom type to prevent name collisions
type ContextKey string

const (
	UserContextKey ContextKey = "user"
)
