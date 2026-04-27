package store

import (
	"context"
	"errors"
	"fmt"
	"reftrail/internal/domain"
)

type ReferralEntry struct {
	ID        int32         `json:"id"`
	CreatorID domain.UserID `json:"creatorId"`
	CreatedTs int64         `json:"createdTs"`
	UpdatedTs int64         `json:"updatedTs"`

	// 2. Patient Info (Matches your requirement #1)
	PatientName string `json:"patientName"`
	PatientDOB  string `json:"patientDob"`

	// 3. Juvonno Integration (Matches your requirements #2 & #3)
	// We use string for TxtCustomerID because Juvonno IDs can sometimes be alphanumeric
	TxtCustomerID    string `json:"txtCustomerId"`
	IntCustomerDocID int32  `json:"intCustomerDocId"`

	// 4. Clinical Details (Matches #4, #6, #7, #10)
	ReferringPhysician string `json:"referringPhysician"`
	Complaint          string `json:"complaint"` // e.g., "Left Knee"
	TriageNote         string `json:"triageNote"`
	XRayClinic         string `json:"xrayClinic"`

	// 5. Workflow & Urgency (Matches #8, #9, #11)
	Urgency string `json:"urgency"` // Elective, Urgent, ASAP
	State   string `json:"state"`   // Ready to book, 1st call, etc.

	// Appointment Info (If state is "Booked")
	ApptDate      string `json:"apptDate"`
	ApptTime      string `json:"apptTime"`
	Practitioner  string `json:"practitioner"`
	JuvonnoApptID string `json:"juvonnoApptId"` // e.g., #18752
}

type CreateReferralEntry struct {
	// Patient & Juvonno Info
	PatientName      string `json:"patientName"`
	PatientDOB       string `json:"patientDob"`
	TxtCustomerID    string `json:"txtCustomerId"`
	IntCustomerDocID int32  `json:"intCustomerDocId"`

	// Clinical Info
	ReferringPhysician string `json:"referringPhysician"`
	Complaint          string `json:"complaint"`
	TriageNote         string `json:"triageNote"`
	// XRayClinic         string `json:"xrayClinic"`

	// Status
	Urgency string `json:"urgency"`
	State   string `json:"state"` // Usually defaults to "READY_TO_BOOK"

	// Accountability
	CreatorID domain.UserID `json:"creatorId"`
}

// FindReferralEntry is the "Search Filter" for your referrals.
type FindReferralEntry struct {
	// 1. Basic Filters
	ID        *int32 `json:"id"`
	CreatorID *int32 `json:"creatorId"`

	// 2. Clinical Filters (Requirement #8 & #9)
	// We use pointers (*) so we can tell the difference between
	// "Filter by this" and "Don't filter at all" (nil).
	Urgency *string `json:"urgency"`
	State   *string `json:"state"`

	// 3. Search Filters (For Fuzzy Physician matching)
	PatientName        *string `json:"patientName"`
	ReferringPhysician *string `json:"referringPhysician"`

	// 4. Pagination (For when your list gets huge)
	Limit  *int `json:"limit"`
	Offset *int `json:"offset"`
}

// UpdateReferralEntry defines which fields are allowed to be changed.
type UpdateReferralEntry struct {
	ID int32 `json:"id"`

	// Fields that change during the workflow
	State      *string `json:"state"`
	TriageNote *string `json:"triageNote"`
	Urgency    *string `json:"urgency"`

	// Appt details (Requirement #11)
	ApptDate     *string `json:"apptDate"`
	ApptTime     *string `json:"apptTime"`
	Practitioner *string `json:"practitioner"`

	// Force flag
	Force bool `json:"force"`
}

type DeleteReferralEntry struct {
	ID int32 `json:"id"`
}

// 1. Create: The "Guard"
func (s *Store) CreateReferralEntry(ctx context.Context, create *CreateReferralEntry) (*ReferralEntry, error) {
	// Logic Check: Don't let someone create a referral without a patient name
	if create.PatientName == "" {
		return nil, errors.New("patient name is required")
	}

	userCtx, ok := domain.GetUserContext(ctx)

	if !ok {
		return nil, errors.New("unauthorized: creator context missing")
	}

	// 2. Set the ID onto the form
	create.CreatorID = userCtx.ID

	// Pass it to the worker (driver)
	return s.driver.CreateReferralEntry(ctx, create)
}

// 2. List: The "Broadcaster"
func (s *Store) ListReferralEntries(ctx context.Context, find *FindReferralEntry) ([]*ReferralEntry, error) {
	// This just asks the driver for a list based on your filters (Urgent, etc.)
	return s.driver.ListReferralEntries(ctx, find)
}

// 3. Get: The "Sniper"
func (s *Store) GetReferralEntry(ctx context.Context, find *FindReferralEntry) (*ReferralEntry, error) {
	// Instead of writing new SQL, it just reuses "List"
	list, err := s.ListReferralEntries(ctx, find)
	if err != nil {
		return nil, err
	}
	// If the list is empty, return "nil" (nothing found)
	if len(list) == 0 {
		return nil, nil
	}

	// Just return the first one found
	return list[0], nil
}

// 4. Update: The "Editor"
func (s *Store) UpdateReferralEntry(ctx context.Context, update *UpdateReferralEntry) error {
	// 1. Get the CURRENT state before it changes
	current, err := s.GetReferralEntry(ctx, &FindReferralEntry{ID: &update.ID})
	if err != nil || current == nil {
		return errors.New("entry not found")
	}

	// 2. ONLY create a log if the state is actually changing
	if update.State != nil && *update.State != current.State {
		// Grab UserID from the context "mailbox" (set by the Bouncer)
		userCtx, ok := domain.GetUserContext(ctx)
		if !ok {
			return errors.New("unauthorized")
		}

		// 3. Tell the Worker to write the history
		_, err := s.driver.CreateReferralLog(ctx, &ReferralLog{
			EntryID:  update.ID,
			UserID:   int32(userCtx.ID),
			OldState: current.State,
			NewState: *update.State,
			Note:     "Status updated via dashboard",
		})
		if err != nil {
			return err // Stop if we can't record history!
		}
	}

	// 4. Finally, update the actual patient record
	return s.driver.UpdateReferralEntry(ctx, update)
}

// 5. Delete: The "Janitor"
func (s *Store) DeleteReferralEntry(ctx context.Context, delete *DeleteReferralEntry) error {

	// Logic Check: Don't try to delete nothing
	if delete.ID == 0 {
		return errors.New("valid ID is required for deletion")
	}

	// Optional: Check if user has permission (Admin role)
	userCtx, ok := domain.GetUserContext(ctx)
	// -----DEBUG-----
	fmt.Printf("Value: %+v, Type: %T\n", ctx.Value("user-role"), ctx.Value("user-role"))
	fmt.Printf("Looking for key: %T(%v)\n", domain.UserKey, domain.UserKey)
	fmt.Printf("Actually in context: %+v\n", ctx)

	if !ok || userCtx.Role != domain.RoleReftrailAdmin {
		if !ok {
			return errors.New("unauthorized: only admins can delete entries, not ok!")
		}
		return errors.New("unauthorized: only admins can delete entries, but ok!")
	}

	// Pass the whole struct to the worker (driver)
	// Before deleting the entry, clean up related logs/comments
	// So call driver.DeleteReferralLogs here later
	return s.driver.DeleteReferralEntry(ctx, delete)
}
