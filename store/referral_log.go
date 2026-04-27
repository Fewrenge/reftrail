package store

import "context"

type ReferralLog struct {
	ID        int32  `json:"id"`
	EntryID   int32  `json:"entryId"`
	UserID    int32  `json:"userId"`
	OldState  string `json:"oldState"`
	NewState  string `json:"newState"`
	Note      string `json:"note"`
	CreatedTs int64  `json:"createdTs"`
}

// Manager Logic: Notice we don't use a "Find" struct here.
// We just ask for the ID of the patient we care about.
func (s *Store) ListReferralLogs(ctx context.Context, entryID int32) ([]*ReferralLog, error) {
	return s.driver.ListReferralLogs(ctx, entryID)
}

// We don't even need a CreateReferralLog function here!
// The Manager calls s.driver.CreateReferralLog directly inside the UpdateWLEntry function.
