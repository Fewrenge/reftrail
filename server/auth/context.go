package auth

import "context"

// ContextKey is a custom type to prevent name collisions
type ContextKey string

const (
	UserIDContextKey ContextKey = "user-id"
)

// SetUserID puts the ID into the request context
func SetUserID(ctx context.Context, userID int32) context.Context {
	return context.WithValue(ctx, UserIDContextKey, userID)
}

// GetUserID pulls the ID back out
func GetUserID(ctx context.Context) (int32, bool) {
	userID, ok := ctx.Value(UserIDContextKey).(int32)
	return userID, ok
}
