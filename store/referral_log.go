package store

import (
	"context"
	"reftrail/internal/domain"
)

type ReferralLog struct {
	ID        domain.ReferralLogID `json:"id"`
	EntryID   domain.ReferralID    `json:"entryId"`
	UserID    domain.UserID        `json:"userId"`
	OldStatus string               `json:"oldStatus"`
	NewStatus string               `json:"newStatus"`
	Note      string               `json:"note"`
	CreatedTs string               `json:"createdTs"`
}

// Manager Logic: Notice we don't use a "Find" struct here.
// We just ask for the ID of the patient we care about.
func (s *Store) ListReferralLogs(ctx context.Context, entryID domain.ReferralID) ([]*ReferralLog, error) {
	return s.driver.ListReferralLogs(ctx, entryID)
}

// We don't even need a CreateReferralLog function here!
// The Manager calls s.driver.CreateReferralLog directly inside the UpdateReferralEntry function.
