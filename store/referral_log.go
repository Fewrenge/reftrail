package store

import (
	"context"
	"reftrail/internal/domain"
)

type ReferralLog struct {
	ID         domain.ReferralLogID  `json:"id"`
	ReferralID domain.ReferralID     `json:"referralId"`
	UserID     domain.UserID         `json:"-"`
	OldStatus  domain.ReferralStatus `json:"oldStatus"`
	NewStatus  domain.ReferralStatus `json:"newStatus"`
	Note       string                `json:"note"`
	CreatedTs  string                `json:"createdTs"`
}

type ReferralLogWithUser struct {
	ReferralLog
	UserPublicInfo
}

func (s *Store) CreateReferralLog(ctx context.Context, create *ReferralLog) (*ReferralLog, error) {
	return s.driver.CreateReferralLog(ctx, create)
}

// Manager Logic: Notice we don't use a "Find" struct here.
// We just ask for the ID of the referral we care about.
func (s *Store) ListReferralLogs(ctx context.Context, referralID domain.ReferralID) ([]*ReferralLogWithUser, error) {
	return s.driver.ListReferralLogs(ctx, referralID)
}
