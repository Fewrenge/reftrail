package store

import (
	"context"
	"errors"
	"fmt"
	"reftrail/internal/domain"
	"time"
)

type ReferralEntry struct {
	ID        int32         `json:"id"`
	CreatorID domain.UserID `json:"-"`
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
	TriageNote         string `json:"triageNote"`
	XRayClinic         string `json:"xrayClinic"`

	// 5. Workflow & Urgency (Matches #8, #9, #11)
	Urgency string `json:"urgency"` // Elective, Urgent, ASAP
	Status  string `json:"status"`  // Ready to book, 1st call, etc.
	Source  string `json:"source"`

	// Appointment Info (If status is "Booked")
	ApptDate      string `json:"apptDate"`
	ApptTime      string `json:"apptTime"`
	Practitioner  string `json:"practitioner"`
	JuvonnoApptID string `json:"juvonnoApptId"` // e.g., #18752
}

type CreateReferralComplaint struct {
	BodyPart string `json:"bodyPart" validate:"required,oneof=SHOULDER KNEE HIP ELBOW WRIST ANKLE FOOT OTHER"`
	Side     string `json:"side"     validate:"required,oneof=LEFT RIGHT BILATERAL"`
	Details  string `json:"details"`
}

type CreateReferralEntry struct {
	// Patient & Juvonno Info
	PatientName      string `json:"patientName" validate:"required,min=2"`
	PatientDOB       string `json:"patientDob"`
	TxtCustomerID    string `json:"txtCustomerId"`
	IntCustomerDocID int32  `json:"intCustomerDocId"`

	// Clinical Info
	ReferringPhysician string                    `json:"referringPhysician"`
	Complaints         []CreateReferralComplaint `json:"complaints" validate:"required,min=1,dive"`
	TriageNote         string                    `json:"triageNote"`
	// XRayClinic         string `json:"xrayClinic"`

	// Status
	Urgency string `json:"urgency"`
	Status  string `json:"status"` // Usually defaults to "READY_TO_BOOK"
	Source  string `json:"source"`

	// Accountability
	CreatorID domain.UserID `json:"creatorId"`
}

type BatchCreateReferralEntries struct {
	ReferralEntries []CreateReferralEntry `json:"referralEntries"`
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
	Status  *string `json:"status"`

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
	Status     *string `json:"status"`
	TriageNote *string `json:"triageNote"`
	Urgency    *string `json:"urgency"`

	// Appt details (Requirement #11)
	ApptDate     *string `json:"apptDate"`
	ApptTime     *string `json:"apptTime"`
	Practitioner *string `json:"practitioner"`

	// Force flag
	Force bool `json:"force"`
}

type UpdateReferralEntryStatus struct {
	ID        int32                 `json:"id"`
	NewStatus domain.ReferralStatus `json:"newStatus"`
	Note      string                `json:"note"`
}

type DeleteReferralEntry struct {
	ID int32 `json:"id"`
}

// 1. Create: The "Guard"
func (s *Store) CreateReferralEntry(ctx context.Context, create *CreateReferralEntry) (*ReferralEntry, error) {
	var newID int32
	ts := time.Now().Unix()

	err := s.driver.RunInTransaction(ctx, func(txCtx context.Context) error {
		// 1. Get User
		user, ok := domain.GetUserContext(txCtx)
		if !ok {
			return errors.New("unauthorized")
		}
		create.CreatorID = domain.UserID(user.ID)

		// 2. Insert Main Entry
		id, err := s.driver.CreateReferralEntry(txCtx, create)
		if err != nil {
			return err
		}
		newID = id

		// 3. Insert Complaints
		for _, c := range create.Complaints {
			if err := s.driver.CreateReferralComplaint(txCtx, newID, &c); err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	// 4. Return the full object that the Handler expects
	return &ReferralEntry{
		ID:                 newID,
		CreatorID:          create.CreatorID,
		CreatedTs:          ts,
		UpdatedTs:          ts,
		PatientName:        create.PatientName,
		PatientDOB:         create.PatientDOB,
		TxtCustomerID:      create.TxtCustomerID,
		IntCustomerDocID:   create.IntCustomerDocID,
		ReferringPhysician: create.ReferringPhysician,
		TriageNote:         create.TriageNote,
		Urgency:            create.Urgency,
		Status:             create.Status,
		Source:             create.Source,
		// Note: Usually we'd fetch complaints back here if the UI needs them immediately
	}, nil
}

func (s *Store) BatchCreateReferralEntries(ctx context.Context, batch *BatchCreateReferralEntries) error {
	// 1. Run everything in one transaction
	return s.driver.RunInTransaction(ctx, func(txCtx context.Context) error {
		user, ok := domain.GetUserContext(txCtx)
		if !ok {
			return errors.New("unauthorized: user context missing")
		}

		// 2. Loop through the entries
		for _, create := range batch.ReferralEntries {
			create.CreatorID = domain.UserID(user.ID)
			// 3. Reuse your existing driver method!
			_, err := s.driver.CreateReferralEntry(txCtx, &create)
			if err != nil {
				// If one fails, the whole transaction returns an error and rolls back
				return fmt.Errorf("batch failed at entry for %s: %w", create.PatientName, err)
			}
		}

		return nil
	})
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
	// Wrap the whole operation in a transaction
	return s.driver.RunInTransaction(ctx, func(txCtx context.Context) error {
		// 1. Get the CURRENT status before it changes
		current, err := s.GetReferralEntry(ctx, &FindReferralEntry{ID: &update.ID})
		if err != nil || current == nil {
			return errors.New("entry not found")
		}

		// 2. ONLY create a log if the status is actually changing
		if update.Status != nil && *update.Status != current.Status {
			// Grab UserID from the context "mailbox" (set by the Bouncer)
			userCtx, ok := domain.GetUserContext(ctx)
			if !ok {
				return errors.New("unauthorized")
			}

			// 3. Tell the Worker to write the history
			_, err := s.driver.CreateReferralLog(ctx, &ReferralLog{
				EntryID:   update.ID,
				UserID:    int32(userCtx.ID),
				OldStatus: current.Status,
				NewStatus: *update.Status,
				Note:      "Status updated via dashboard",
			})
			if err != nil {
				return err // Stop if we can't record history!
			}
		}

		// 4. Finally, update the actual patient record
		return s.driver.UpdateReferralEntry(ctx, update)
	})
}

func (s *Store) GetReferralEntryStatusByID(ctx context.Context, id int32) (domain.ReferralStatus, error) {
	return s.driver.GetReferralEntryStatusByID(ctx, id)
}

func (s *Store) UpdateReferralEntryStatus(ctx context.Context, update *UpdateReferralEntryStatus) error {
	// You are looking at an anonymous function
	return s.driver.RunInTransaction(ctx, func(txCtx context.Context) error {
		// 1. Get the "Who" (User Context)
		user, ok := domain.GetUserContext(txCtx)
		if !ok {
			return errors.New("unauthorized: user context missing")
		}

		// 2. Get the "Where we are" (Old Status)
		// We use the driver directly because we are already inside a transaction
		oldStatus, err := s.driver.GetReferralEntryStatusByID(txCtx, update.ID)
		if err != nil {
			return fmt.Errorf("failed to fetch current status: %w", err)
		}

		// 3. THE RULE CHECK (Calling your domain code)
		// We convert update.NewStatus to domain.ReferralStatus type
		newStatus := domain.ReferralStatus(update.NewStatus)

		if !domain.CanTransition(oldStatus, newStatus, user.Role) {
			return fmt.Errorf("illegal status transition from %s to %s for role %s",
				oldStatus, newStatus, user.Role)
		}

		// 4. Update the Status
		if err := s.driver.UpdateReferralEntryStatus(txCtx, update.ID, newStatus); err != nil {
			return err
		}

		// 5. Create the Log
		_, err = s.driver.CreateReferralLog(txCtx, &ReferralLog{
			EntryID:   update.ID,
			UserID:    int32(user.ID),
			OldStatus: string(oldStatus),
			NewStatus: string(newStatus),
			Note:      update.Note,
		})

		return err // If this is nil, transaction commits!
	})
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
