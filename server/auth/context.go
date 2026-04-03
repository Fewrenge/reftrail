package auth

import (
	"wl/internal/types"
)

// ContextKey is a custom type to prevent name collisions
type ContextKey string

const (
	UserContextKey ContextKey = "user"
)

// UserContext holds the data we actually care about during a request
type UserContext struct {
	ID   types.UserID
	Role types.UserRole
}
