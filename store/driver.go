package store

import (
	"context"
	"database/sql"
	"reftrail/internal/types"
)

// Driver is the interface that any database must implement.
type Driver interface {
	// 1. Connection Management
	GetDB() *sql.DB
	Close() error

	// 2. Waitlist Entry Methods
	// Notice we use the "Form" structs we just created!
	CreateWLEntry(ctx context.Context, create *CreateWLEntry) (*WLEntry, error)
	ListWLEntries(ctx context.Context, find *FindWLEntry) ([]*WLEntry, error)
	UpdateWLEntry(ctx context.Context, update *UpdateWLEntry) error
	DeleteWLEntry(ctx context.Context, delete *DeleteWLEntry) error

	// 3. Accountability (Optional but recommended for your logs)
	CreateWLLog(ctx context.Context, create *WLLog) (*WLLog, error)
	ListWLLogs(ctx context.Context, entryID int32) ([]*WLLog, error)

	// 4. User/Account Methods (For your Login/Privileges)
	CreateUser(ctx context.Context, create *CreateUser) (*User, error)
	CountUsers(ctx context.Context) (int, error)
	ListUsers(ctx context.Context, find *FindUser) ([]*User, error)
	UpdateUser(ctx context.Context, update *UpdateUser) (*User, error)
	DeleteUser(ctx context.Context, delete *DeleteUser) error
	ChangeUserPassword(ctx context.Context, userID types.UserID, newHash string) error
}
