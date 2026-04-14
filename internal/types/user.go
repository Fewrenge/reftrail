package types

import "context"

type UserID int32
type UserRole string

const (
	RoleWLSystemAdmin = "WL_SYSTEM_ADMIN"
	RoleBookingTeam   = "BOOKING_TEAM"
)

type UserContext struct {
	ID   UserID
	Role UserRole
}

type contextKey string

const UserKey contextKey = "user"

func GetUserContext(ctx context.Context) (*UserContext, bool) {
	u, ok := ctx.Value(UserKey).(*UserContext)
	return u, ok
}
