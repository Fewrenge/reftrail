package sqlite

import (
	"context"
	"reftrail/internal/domain"
	"reftrail/store" // Import your main store for the object definitions
	"strings"
	"time"
)

func (d *Driver) CreateReferralEntry(ctx context.Context, create *store.CreateReferralEntry) (*store.ReferralEntry, error) {
	// 1. Get the current time for our timestamps
	ts := time.Now().Unix()

	// 2. Write the SQL command
	// We use "?" as placeholders to prevent "SQL Injection" (Hacking)
	stmt := `INSERT INTO referral_entry (
		creator_id, created_ts, updated_ts, 
		patient_name, patient_dob, txt_customer_id, int_customer_doc_id,
		referring_physician, complaint, triage_note, urgency, status
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	// 3. Execute the command
	result, err := d.conn(ctx).ExecContext(ctx, stmt,
		create.CreatorID, ts, ts,
		create.PatientName, create.PatientDOB, create.TxtCustomerID, create.IntCustomerDocID,
		create.ReferringPhysician, create.Complaint, create.TriageNote, create.Urgency, create.Status,
	)
	if err != nil {
		return nil, err
	}

	// 4. Get the ID that SQLite just generated automatically
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	// 5. Return the "Finished" object back to the Manager
	return &store.ReferralEntry{
		ID:                 int32(id),
		CreatorID:          create.CreatorID,
		CreatedTs:          ts,
		UpdatedTs:          ts,
		PatientName:        create.PatientName,
		PatientDOB:         create.PatientDOB,
		TxtCustomerID:      create.TxtCustomerID,
		IntCustomerDocID:   create.IntCustomerDocID,
		ReferringPhysician: create.ReferringPhysician,
		Complaint:          create.Complaint,
		TriageNote:         create.TriageNote,
		Urgency:            create.Urgency,
		Status:             create.Status,
	}, nil
}

func (d *Driver) ListReferralEntries(ctx context.Context, find *store.FindReferralEntry) ([]*store.ReferralEntry, error) {
	// 1. The Base Query
	// "WHERE 1 = 1" is a classic trick. It does nothing, but lets us
	// safely add "AND ..." to the end of the string later.
	query := `SELECT 
		id, creator_id, created_ts, updated_ts, 
		patient_name, patient_dob, txt_customer_id, int_customer_doc_id,
		referring_physician, complaint, triage_note, urgency, status,
		IFNULL(appt_date, ''),
		IFNULL(appt_time,''),
		IFNULL(practitioner, ''), 
		IFNULL(juvonno_appt_id,'')
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
			&entry.ReferringPhysician, &entry.Complaint, &entry.TriageNote, &entry.Urgency, &entry.Status,
			&entry.ApptDate, &entry.ApptTime, &entry.Practitioner, &entry.JuvonnoApptID,
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
	_, err := d.conn(ctx).ExecContext(ctx, query, string(status), time.Now().Unix, id)
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
