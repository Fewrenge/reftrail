package sqlite

import (
	"context"
	"reftrail/internal/domain"
	"reftrail/store" // Import your main store for the object definitions
	"strings"
	"time"
)

func (d *Driver) CreateReferralComplaint(ctx context.Context, referralID int32, complaint *store.ReferralComplaint) error {
	stmt := `INSERT INTO referral_complaint (referral_id, body_part, side, details) VALUES (?, ?, ?, ?)`
	_, err := d.conn(ctx).ExecContext(ctx, stmt, referralID, complaint.BodyPart, complaint.Side, complaint.Details)
	return err
}

func (d *Driver) CreateReferralEntry(ctx context.Context, create *store.CreateReferralEntry) (int32, error) {
	// Get the current time for our timestamps
	ts := time.Now().Unix()

	stmt := `INSERT INTO referral_entry (
		creator_id, created_ts, updated_ts, 
		patient_name, patient_dob, txt_customer_id, int_customer_doc_id,
		referring_physician, triage_note, urgency, status, source
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	// Execute the command
	result, err := d.conn(ctx).ExecContext(ctx, stmt,
		int32(create.CreatorID), ts, ts,
		create.PatientName, create.PatientDOB, create.TxtCustomerID, create.IntCustomerDocID,
		create.ReferringPhysician, create.TriageNote, create.Urgency, create.Status, create.Source,
	)
	if err != nil {
		return 0, err
	}

	// Get the ID that SQLite just generated automatically
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int32(id), err
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
		patient_name, patient_dob, txt_customer_id, int_customer_doc_id,
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
	if find.PatientName != nil {
		query += " AND patient_name LIKE ?"
		args = append(args, "%"+*find.PatientName+"%")
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
			&entry.PatientName, &entry.PatientDOB, &entry.TxtCustomerID, &entry.IntCustomerDocID,
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
	args = append(args, time.Now().Unix())

	// 2. Add the ID for the WHERE clause
	args = append(args, update.ID)

	// 3. Execute: UPDATE referral_entry SET status = ?, updated_ts = ? WHERE id = ?
	query := `UPDATE referral_entry SET ` + strings.Join(set, ", ") + ` WHERE id = ?`
	_, err := d.conn(ctx).ExecContext(ctx, query, args...)
	return err
}

func (d *Driver) GetReferralEntryStatusByID(ctx context.Context, id int32) (domain.ReferralStatus, error) {
	var status domain.ReferralStatus
	err := d.conn(ctx).QueryRowContext(ctx, "SELECT status FROM referral_entry WHERE id = $1", id).Scan(&status)
	return status, err
}

// Only updates referral entry status
func (d *Driver) UpdateReferralEntryStatus(ctx context.Context, id int32, status domain.ReferralStatus) error {
	query := `UPDATE referral_entry SET status = ?, updated_ts = ? WHERE id = ?`
	_, err := d.conn(ctx).ExecContext(ctx, query, string(status), time.Now().Unix(), id)
	return err
}

func (d *Driver) DeleteReferralEntry(ctx context.Context, delete *store.DeleteReferralEntry) error {
	// We pull the ID out of the struct's ID field
	stmt := `DELETE FROM referral_entry WHERE id = ?`
	_, err := d.conn(ctx).ExecContext(ctx, stmt, delete.ID)
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

	query := fmt.Sprintf("DELETE FROM referral_entry WHERE id IN (%s)", strings.Join(placeholders, ","))
	_, err := d.conn(ctx).ExecContext(ctx, query, args...)
	return err
}
*/
