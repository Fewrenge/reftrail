package domain

import "context"

type UserRole string
type Username string

const (
	RoleReftrailAdmin = "REFTRAIL_ADMIN"
	RoleBookingTeam   = "BOOKING_TEAM"
)

type UserContext struct {
	Username Username `json:"username"`
	Role     UserRole `json:"role"`
}

type contextKey string

const UserKey contextKey = "user"

func GetUserContext(ctx context.Context) (*UserContext, bool) {
	u, ok := ctx.Value(UserKey).(*UserContext)
	return u, ok
}
