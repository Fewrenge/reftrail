package store

import (
	"context"
	"fmt"
	"reftrail/internal/domain"
)

type ReferralEntry struct {
	ID        domain.ReferralID `json:"id"`
	CreatorID domain.UserID     `json:"-"`
	CreatedTs string            `json:"createdTs"`
	UpdatedTs string            `json:"updatedTs"`

	// 2. Patient Info
	PatientLastName              string `json:"patientLastName"`
	PatientFirstName             string `json:"patientFirstName"`
	PatientDOB                   string `json:"patientDob"`
	PatientHealthcardNumber      string `json:"patientHealthcardNumber"`
	PatientHealthcardVersionCode string `json:"patientHealthcardVersionCode"`

	Complaints []*ReferralComplaint `json:"complaints"`

	// 3. EMR Integration
	// Use string in case EMR updates their ID format in the future
	TxtCustomerID    string `json:"txtCustomerId"`
	IntCustomerDocID int64  `json:"intCustomerDocId"`

	// 4. Clinical Details
	ReferringPhysician string `json:"referringPhysician"`
	TriageNote         string `json:"triageNote"`
	XRayClinic         string `json:"xrayClinic"`

	// 5. Workflow & Urgency
	Urgency string `json:"urgency"` // Elective, Urgent, ASAP
	Status  string `json:"status"`  // Ready to book, 1st call, etc.
	Source  string `json:"source"`

	// Appointment Info (If status is "Booked")
	ApptDateAndTime string `json:"apptDateAndTime"`
	Practitioner    string `json:"practitioner"`
	JuvonnoApptID   string `json:"juvonnoApptId"` // e.g., #18752
}

type ReferralComplaint struct {
	ID         int64             `json:"id"`
	ReferralID domain.ReferralID `json:"referralId"`
	BodyPart   string            `json:"bodyPart" validate:"required,oneof=SHOULDER KNEE HIP ELBOW WRIST ANKLE FOOT OTHER"`
	Side       string            `json:"side"     validate:"required,oneof=LEFT RIGHT BILATERAL OTHER"`
	Details    string            `json:"details"`
}

type CreateReferralEntry struct {
	// Patient & Juvonno Info
	PatientLastName              string `json:"patientLastName" validate:"required"`
	PatientFirstName             string `json:"patientFirstName" validate:"required"`
	PatientDOB                   string `json:"patientDob"`
	PatientHealthcardNumber      string `json:"patientHealthcardNumber"`
	PatientHealthcardVersionCode string `json:"patientHealthcardVersionCode"`
	TxtCustomerID                string `json:"txtCustomerId"`
	IntCustomerDocID             int64  `json:"intCustomerDocId"`

	// Clinical Info
	ReferringPhysician string              `json:"referringPhysician"`
	Complaints         []ReferralComplaint `json:"complaints" validate:"required,min=1,dive"`
	TriageNote         string              `json:"triageNote"`
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
	ID        *domain.ReferralID `json:"id"`
	CreatorID *domain.UserID     `json:"creatorId"`

	// 2. Clinical Filters (Requirement #8 & #9)
	// We use pointers (*) so we can tell the difference between
	// "Filter by this" and "Don't filter at all" (nil).
	Urgency *string `json:"urgency"`
	Status  *string `json:"status"`

	// 3. Search Filters (For Fuzzy Physician matching)
	PatientLastName         *string `json:"patientLastName"`
	PatientFirstName        *string `json:"patientFirstName"`
	ReferringPhysician      *string `json:"referringPhysician"`
	PatientHealthcardNumber *string `json:"patientHealthcardNumber"`

	// 4. Pagination (For when your list gets huge)
	Limit  *int `json:"limit"`
	Offset *int `json:"offset"`
}

// Admin use only, for arbiturary updates (e.g., correcting a typo, changing urgency, etc.)
// TODO: Determine what can be updated
type UpdateReferralEntry struct {
	ID domain.ReferralID `json:"id"`

	// Fields that change during the workflow
	Status     *string `json:"status"`
	TriageNote *string `json:"triageNote"`
	Urgency    *string `json:"urgency"`

	Note *string `json:"note"`

	// Force flag
	Force bool `json:"force"`
}

type UpdateReferralEntryStatus struct {
	ID        domain.ReferralID     `json:"id"`
	NewStatus domain.ReferralStatus `json:"newStatus"`
	Note      string                `json:"note"`
}

// Only records initial appointment, not for rescheduling (which is the EMR's job)
type UpdateReferralEntryAppointment struct {
	// Appt details (Requirement #11)
	ApptDateAndTime *string `json:"apptDateAndTime"`
	Practitioner    *string `json:"practitioner"`
	EMRApptID       *string `json:"emrApptId"`
}

type DeleteReferralEntry struct {
	ID domain.ReferralID `json:"id"`
}

// 1. Create: The "Guard"
// Implement log creation logic at the store level so that it can be reused across different handlers
func (s *Store) CreateReferralEntry(ctx context.Context, create *CreateReferralEntry) (*ReferralEntry, error) {
	var referralEntry *ReferralEntry

	err := s.driver.RunInTransaction(ctx, func(txCtx context.Context) error {
		// 1. Get User
		user, ok := domain.GetUserContext(txCtx)
		if !ok {
			return domain.ErrUnauthorized
		}
		create.CreatorID = domain.UserID(user.ID)

		// 2. Insert Main Entry
		var err error
		entry, err := s.driver.CreateReferralEntry(txCtx, create)
		if err != nil {
			return err
		}
		referralEntry = entry

		// 3. Insert Complaints
		for _, c := range create.Complaints {
			if err := s.driver.CreateReferralComplaint(txCtx, referralEntry.ID, &c); err != nil {
				return err
			}
		}

		// 4. Create log for creation
		var creationLog *ReferralLog
		creationLog = &ReferralLog{
			EntryID:   referralEntry.ID,
			UserID:    domain.UserID(user.ID),
			OldStatus: "",
			NewStatus: referralEntry.Status,
			Note:      "Referral entry created",
		}

		if _, err := s.driver.CreateReferralLog(txCtx, creationLog); err != nil {
			return fmt.Errorf("failed to create initial audit log: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// 4. Return the full object that the Handler expects
	return referralEntry, nil
}

func (s *Store) BatchCreateReferralEntries(ctx context.Context, batch *BatchCreateReferralEntries) error {
	// 1. Run everything in one transaction
	return s.driver.RunInTransaction(ctx, func(txCtx context.Context) error {
		_, ok := domain.GetUserContext(txCtx)
		if !ok {
			return domain.ErrUnauthorized
		}

		// 2. Loop through the entries
		for _, create := range batch.ReferralEntries {
			_, err := s.CreateReferralEntry(txCtx, &create)
			if err != nil {
				return fmt.Errorf("failed to create referral entry for patient %s, %s: %w",
					create.PatientLastName, create.PatientFirstName, err)
			}
		}

		return nil
	})
}

// 2. List: The "Broadcaster"
func (s *Store) ListReferralEntries(ctx context.Context, find *FindReferralEntry) ([]*ReferralEntry, error) {
	// 1. Get the list of referrals (Your existing query)
	entries, err := s.driver.ListReferralEntries(ctx, find)
	if err != nil {
		return nil, err
	}

	if len(entries) == 0 {
		return entries, nil
	}

	// 2. Get EVERY complaint
	allComplaints, err := s.driver.ListAllComplaints(ctx)
	if err != nil {
		return nil, err
	}

	// 3. Group complaints by ReferralID using a Map
	// Key: ReferralID, Value: Slice of complaints
	complaintMap := make(map[domain.ReferralID][]*ReferralComplaint)
	for _, c := range allComplaints {
		complaintMap[c.ReferralID] = append(complaintMap[c.ReferralID], c)
	}

	// 4. Attach complaints to each entry
	for _, entry := range entries {
		if comps, found := complaintMap[entry.ID]; found {
			entry.Complaints = comps
		} else {
			entry.Complaints = []*ReferralComplaint{} // Return empty array instead of null JSON
		}
	}

	return entries, nil
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

func (s *Store) UpdateReferralEntry(ctx context.Context, update *UpdateReferralEntry) error {
	// Wrap the whole operation in a transaction
	return s.driver.RunInTransaction(ctx, func(txCtx context.Context) error {
		// 1. Get current record using the transaction context txCtx
		current, err := s.GetReferralEntry(txCtx, &FindReferralEntry{ID: &update.ID})
		if err != nil || current == nil {
			return domain.ErrReferralEntryNotFound
		}

		// Grab UserID from the context "mailbox" (set by the Bouncer)
		userCtx, ok := domain.GetUserContext(ctx)
		if !ok {
			return domain.ErrUnauthorized
		}

		// 2. Tell the Worker to write the history
		logPayload := &ReferralLog{
			EntryID:   update.ID,
			UserID:    domain.UserID(userCtx.ID),
			OldStatus: current.Status,
			NewStatus: *update.Status,
			Note:      *update.Note,
		}

		if _, err := s.driver.CreateReferralLog(txCtx, logPayload); err != nil {
			return fmt.Errorf("failed to create referral history log during record update: %w", err)
		}

		// 3. Commit the changes to the primary referral entity record
		if err := s.driver.UpdateReferralEntry(txCtx, update); err != nil {
			return fmt.Errorf("failed to execute referral entry update: %w", err)
		}

		return nil
	})
}

func (s *Store) GetReferralEntryStatusByID(ctx context.Context, id domain.ReferralID) (domain.ReferralStatus, error) {
	return s.driver.GetReferralEntryStatusByID(ctx, id)
}

func (s *Store) UpdateReferralEntryStatus(ctx context.Context, update *UpdateReferralEntryStatus) error {
	// Anonymous function here
	return s.driver.RunInTransaction(ctx, func(txCtx context.Context) error {
		// 1. Get the "Who" (User Context)
		user, ok := domain.GetUserContext(txCtx)
		if !ok {
			return domain.ErrUnauthorized
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
			return fmt.Errorf("illegal status transition from %s to %s for role %s: %w",
				oldStatus, newStatus, user.Role, domain.ErrIllegalTransition)
		}

		// 4. Update the Status
		if err := s.driver.UpdateReferralEntryStatus(txCtx, update.ID, newStatus); err != nil {
			return fmt.Errorf("failed to update status in database: %w", err)
		}

		// 5. Create the Log
		logPayload := &ReferralLog{
			EntryID:   update.ID,
			UserID:    domain.UserID(user.ID),
			OldStatus: string(oldStatus),
			NewStatus: string(newStatus),
			Note:      update.Note,
		}

		if _, err := s.driver.CreateReferralLog(txCtx, logPayload); err != nil {
			return fmt.Errorf("failed to write audit log entry: %w", err)
		}

		return nil // If this is nil, transaction commits!
	})
}

// 5. Delete: The "Janitor"
func (s *Store) DeleteReferralEntry(ctx context.Context, delete *DeleteReferralEntry) error {

	// Logic Check: Don't try to delete nothing
	if delete.ID == "" {
		return domain.ErrDataValidationFailed
	}

	// Optional: Check if user has permission (Admin role)
	userCtx, ok := domain.GetUserContext(ctx)
	if !ok || userCtx.Role != domain.RoleReftrailAdmin {
		return domain.ErrUnauthorized
	}

	// Pass the whole struct to the worker (driver)
	// ON DELETE CASCADE activated
	if err := s.driver.DeleteReferralEntry(ctx, delete); err != nil {
		return fmt.Errorf("failed to delete referral entry with ID %s: %w", delete.ID, err)
	}

	return nil
}
