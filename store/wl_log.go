package store

import "context"

type WLLog struct {
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
func (s *Store) ListWLLogs(ctx context.Context, entryID int32) ([]*WLLog, error) {
	return s.driver.ListWLLogs(ctx, entryID)
}

// We don't even need a CreateWLLog function here!
// The Manager calls s.driver.CreateWLLog directly inside the UpdateWLEntry function.
