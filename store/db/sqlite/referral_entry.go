package sqlite

import (
	"context"
	"fmt"
	"reftrail/internal/domain"
	"reftrail/store"
	"strings"
	"time"

	uuid "github.com/google/uuid"
)

func (d *Driver) CreateReferralComplaint(ctx context.Context, referralID domain.ReferralID, complaint *store.ReferralComplaint) error {
	query := `INSERT INTO referral_complaint (referral_id, body_part, side, details) VALUES (?, ?, ?, ?)`
	_, err := d.conn(ctx).ExecContext(ctx, query, referralID, complaint.BodyPart, complaint.Side, complaint.Details)
	return err
}

func (d *Driver) CreateReferralEntry(ctx context.Context, create *store.CreateReferralEntry) (*store.ReferralEntry, error) {
	// Get the current time for our timestamps
	newID, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	idStr := newID.String()
	ts := time.Now().Format(time.RFC3339)

	query := `INSERT INTO referral_entry (
		id, created_ts, updated_ts, creator_id, 
		patient_last_name, patient_first_name, patient_dob, patient_healthcard_number, patient_healthcard_version_code,
		 txt_customer_id, int_customer_doc_id,
		referring_physician, triage_note, urgency, status, source
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	// Execute the command
	_, err = d.conn(ctx).ExecContext(ctx, query,
		idStr, ts, ts, int64(create.CreatorID),
		create.PatientLastName, create.PatientFirstName, create.PatientDOB,
		create.PatientHealthcardNumber, create.PatientHealthcardVersionCode,
		create.TxtCustomerID, create.IntCustomerDocID,
		create.ReferringPhysician, create.TriageNote, create.Urgency, create.Status, create.Source,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert referral entry for patient %s, %s (creator_id: %d): %w",
			create.PatientLastName, create.PatientFirstName, create.CreatorID, err)
	}
	return &store.ReferralEntry{
		ID:                           domain.ReferralID(idStr), // Cast to custom type
		CreatedTs:                    ts,
		UpdatedTs:                    ts,
		CreatorID:                    create.CreatorID,
		PatientLastName:              create.PatientLastName,
		PatientFirstName:             create.PatientFirstName,
		PatientDOB:                   create.PatientDOB,
		PatientHealthcardNumber:      create.PatientHealthcardNumber,
		PatientHealthcardVersionCode: create.PatientHealthcardVersionCode,
		TxtCustomerID:                create.TxtCustomerID,
		IntCustomerDocID:             create.IntCustomerDocID,
		ReferringPhysician:           create.ReferringPhysician,
		TriageNote:                   create.TriageNote,
		Urgency:                      create.Urgency,
		Status:                       create.Status,
		Source:                       create.Source,
	}, nil
}

func (d *Driver) ListAllComplaints(ctx context.Context) ([]*store.ReferralComplaint, error) {
	query := `SELECT id, referral_id, body_part, side, details FROM referral_complaint`
	rows, err := d.conn(ctx).QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*store.ReferralComplaint
	for rows.Next() {
		var c store.ReferralComplaint
		if err := rows.Scan(&c.ID, &c.ReferralID, &c.BodyPart, &c.Side, &c.Details); err != nil {
			return nil, err
		}
		list = append(list, &c)
	}
	return list, nil
}

func (d *Driver) ListReferralEntries(ctx context.Context, find *store.FindReferralEntry) ([]*store.ReferralEntry, error) {
	// 1. The Base Query
	query := `SELECT 
		id, creator_id, created_ts, updated_ts, 
		patient_last_name, patient_first_name, patient_dob, patient_healthcard_number, patient_healthcard_version_code,
		txt_customer_id, int_customer_doc_id,
		referring_physician, triage_note, urgency, status, source
	FROM referral_entry WHERE 1 = 1`

	// 2. The "Arguments" list
	// This stores the values we will plug into the "?" placeholders
	var args []any

	// 3. Add Dynamic Filters (Requirement #8 & #9)
	if find.ID != nil {
		query += " AND id = ?"
		args = append(args, *find.ID)
	}
	if find.Urgency != nil {
		query += " AND urgency = ?"
		args = append(args, *find.Urgency)
	}
	if find.Status != nil {
		query += " AND status = ?"
		args = append(args, *find.Status)
	}

	// Fuzzy Matching for Patient Name (Requirement #1)
	if find.PatientLastName != nil && *find.PatientLastName != "" {
		query += " AND patient_last_name LIKE ?"
		args = append(args, "%"+*find.PatientLastName+"%")
	}
	if find.PatientFirstName != nil && *find.PatientFirstName != "" {
		query += " AND patient_first_name LIKE ?"
		args = append(args, "%"+*find.PatientFirstName+"%")
	}

	// 4. Sorting (Always show newest or most urgent first)
	query += " ORDER BY created_ts DESC"

	// 5. Run the Query
	rows, err := d.conn(ctx).QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// 6. The "Bucket" for our results
	var list []*store.ReferralEntry

	// 7. Loop through the database rows
	for rows.Next() {
		var entry store.ReferralEntry
		// Scan matches the columns in our SELECT statusment to our Go struct
		err := rows.Scan(
			&entry.ID, &entry.CreatorID, &entry.CreatedTs, &entry.UpdatedTs,
			&entry.PatientLastName, &entry.PatientFirstName, &entry.PatientDOB,
			&entry.PatientHealthcardNumber, &entry.PatientHealthcardVersionCode,
			&entry.TxtCustomerID, &entry.IntCustomerDocID,
			&entry.ReferringPhysician, &entry.TriageNote, &entry.Urgency, &entry.Status, &entry.Source,
		)
		if err != nil {
			return nil, err
		}
		list = append(list, &entry)
	}

	return list, nil
}

// For miscellaneous updates
func (d *Driver) UpdateReferralEntry(ctx context.Context, update *store.UpdateReferralEntry) error {
	// 1. Build the "SET" part of our SQL dynamically
	set, args := []string{}, []any{}

	if v := update.Status; v != nil {
		set = append(set, "status = ?")
		args = append(args, *v)
	}
	if v := update.TriageNote; v != nil {
		set = append(set, "triage_note = ?")
		args = append(args, *v)
	}
	// Update the timestamp automatically
	set = append(set, "updated_ts = ?")
	args = append(args, time.Now().Format(time.RFC3339))

	// 2. Add the ID for the WHERE clause
	args = append(args, update.ID)

	// 3. Execute: UPDATE referral_entry SET status = ?, updated_ts = ? WHERE id = ?
	query := `UPDATE referral_entry SET ` + strings.Join(set, ", ") + ` WHERE id = ?`
	_, err := d.conn(ctx).ExecContext(ctx, query, args...)
	return err
}

func (d *Driver) GetReferralEntryStatusByID(ctx context.Context, id domain.ReferralID) (domain.ReferralStatus, error) {
	var status domain.ReferralStatus
	err := d.conn(ctx).QueryRowContext(ctx, "SELECT status FROM referral_entry WHERE id = $1", id).Scan(&status)
	return status, err
}

// Only updates referral entry status
func (d *Driver) UpdateReferralEntryStatus(ctx context.Context, id domain.ReferralID, status domain.ReferralStatus) error {
	query := `UPDATE referral_entry SET status = ?, updated_ts = ? WHERE id = ?`
	_, err := d.conn(ctx).ExecContext(ctx, query, string(status), time.Now().Format(time.RFC3339), id)
	return err
}

func (d *Driver) DeleteReferralEntry(ctx context.Context, delete *store.DeleteReferralEntry) error {
	// We pull the ID out of the struct's ID field
	query := `DELETE FROM referral_entry WHERE id = ?`
	_, err := d.conn(ctx).ExecContext(ctx, query, delete.ID)
	return err
}

/*
func (d *Driver) DeleteReferralEntries(ctx context.Context, ids []int32) error {
	if len(ids) == 0 {
		return nil
	}

	// Create the (?, ?, ?) string based on how many IDs we have
	placeholders := make([]string, len(ids))
	args := make([]any, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}

	query := "DELETE FROM referral_entry WHERE id IN (%s)"
	return err
}
*/
