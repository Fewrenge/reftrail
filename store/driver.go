package store

import (
	"context"
	"database/sql"
	"reftrail/internal/domain"
)

// Driver is the interface that any database must implement.
type Driver interface {
	// 1. Connection Management
	GetDB() *sql.DB
	Close() error

	// 2. Referral Entry Methods
	// Notice we use the "Form" structs we just created!
	CreateReferralComplaint(ctx context.Context, referralID int32, c *ReferralComplaint) error
	ListAllComplaints(ctx context.Context) ([]*ReferralComplaint, error)
	CreateReferralEntry(ctx context.Context, create *CreateReferralEntry) (int32, error)
	ListReferralEntries(ctx context.Context, find *FindReferralEntry) ([]*ReferralEntry, error)
	UpdateReferralEntry(ctx context.Context, update *UpdateReferralEntry) error
	DeleteReferralEntry(ctx context.Context, delete *DeleteReferralEntry) error
	GetReferralEntryStatusByID(ctx context.Context, id int32) (domain.ReferralStatus, error)
	UpdateReferralEntryStatus(ctx context.Context, id int32, status domain.ReferralStatus) error

	// 3. Accountability (Optional but recommended for your logs)
	CreateReferralLog(ctx context.Context, create *ReferralLog) (*ReferralLog, error)
	ListReferralLogs(ctx context.Context, entryID int32) ([]*ReferralLog, error)

	// 4. User/Account Methods (For your Login/Privileges)
	CreateUser(ctx context.Context, create *CreateUser) (*User, error)
	CountUsers(ctx context.Context) (int, error)
	ListUsers(ctx context.Context, find *FindUser) ([]*User, error)
	UpdateUser(ctx context.Context, update *UpdateUser) (*User, error)
	DeleteUser(ctx context.Context, delete *DeleteUser) error
	ChangeUserPassword(ctx context.Context, userID domain.UserID, newHash string) error

	// 5. Transaction methods
	RunInTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}
