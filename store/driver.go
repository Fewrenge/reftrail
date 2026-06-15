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
	CreateReferralEntry(ctx context.Context, create *CreateReferralEntry) (*ReferralEntry, error)
	ListReferralEntries(ctx context.Context, find *FindReferralEntry) ([]*ReferralEntry, error)
	GetReferralEntriesCount(ctx context.Context, find *FindReferralEntry) (int, error)
	UpdateReferralEntry(ctx context.Context, update *UpdateReferralEntry) error // Admin use only
	DeleteReferralEntry(ctx context.Context, delete *DeleteReferralEntry) error

	GetReferralEntryStatusByID(ctx context.Context, id domain.ReferralID) (domain.ReferralStatus, error)
	UpdateReferralEntryStatus(ctx context.Context, id domain.ReferralID, status domain.ReferralStatus) error

	CreateReferralComplaint(ctx context.Context, referralID domain.ReferralID, c *ReferralComplaint) error
	DeleteReferralComplaint(ctx context.Context, referralID domain.ReferralID) error // Only used for UpdateReferralEntry
	ListAllComplaints(ctx context.Context) ([]*ReferralComplaint, error)

	// 3. Accountability (Optional but recommended for your logs)
	CreateReferralLog(ctx context.Context, create *ReferralLog) (*ReferralLog, error)
	ListReferralLogs(ctx context.Context, referralID domain.ReferralID) ([]*ReferralLogWithUser, error)

	// 4. User/Account Methods (For your Login/Privileges)
	CreateUser(ctx context.Context, create *CreateUser) (*User, error)
	CountUsers(ctx context.Context) (int, error)
	ListUsers(ctx context.Context, find *FindUser) ([]*User, error)
	DeleteUser(ctx context.Context, delete *DeleteUser) error
	UpdateUserInfo(ctx context.Context, update *UpdateUserInfo) (*User, error)
	UpdateUserPassword(ctx context.Context, username domain.Username, newHash string) error
	GetUserByUsername(ctx context.Context, username domain.Username) (*User, error)
	ArchiveUser(ctx context.Context, username domain.Username) error
	CountActiveAdmins(ctx context.Context) (int, error)
	UpdateUserRole(ctx context.Context, username domain.Username, role domain.UserRole) error

	// 5. Transaction methods
	RunInTransaction(ctx context.Context, fn func(ctx context.Context) error) error

	// 6. Tag methods
	CreateReferralTag(ctx context.Context, create *CreateReferralTag) (*ReferralTag, error)
	ListReferralTags(ctx context.Context) ([]*ReferralTag, error)
	ListAllLinkedReferralTags(ctx context.Context) ([]*LinkedReferralTagRow, error)
	DeleteReferralTag(ctx context.Context, delete *DeleteReferralTag) error
	AssignTagToReferral(ctx context.Context, referralID domain.ReferralID, tagName string) error
	RemoveTagFromReferral(ctx context.Context, referralID domain.ReferralID, tagName string) error

	// 7. Analytics methods
	GetUrgencyDistribution(ctx context.Context, find *FindReferralEntry) (*UrgencyDistributionResponse, error)
	GetReferralVolume(ctx context.Context, find *FindReferralEntry) (*ReferralVolumeResponse, error)
}
