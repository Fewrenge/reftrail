package domain

import "context"

type UserID int64
type UserRole string
type Username string

const (
	RoleReftrailAdmin = "REFTRAIL_ADMIN"
	RoleBookingTeam   = "BOOKING_TEAM"
)

type UserContext struct {
	ID       UserID
	Role     UserRole
	Username Username
}

type contextKey string

const UserKey contextKey = "user"

func GetUserContext(ctx context.Context) (*UserContext, bool) {
	u, ok := ctx.Value(UserKey).(*UserContext)
	return u, ok
}
