package auth

import "context"

// ContextKey is a custom type to prevent name collisions
type ContextKey string

const (
	UserContextKey ContextKey = "user"
)

// UserContext holds the data we actually care about during a request
type UserContext struct {
	ID   int32
	Role string
}

func GetUserContext(ctx context.Context) (*UserContext, bool) {
	user, ok := ctx.Value(UserContextKey).(*UserContext)
	return user, ok
}
