package sqlite

import (
	"context"
	"strings"
	"time"
	"wl/store" // Import your main store for the object definitions
)

func (d *Driver) CreateWLEntry(ctx context.Context, create *store.CreateWLEntry) (*store.WLEntry, error) {
	// 1. Get the current time for our timestamps
	ts := time.Now().Unix()

	// 2. Write the SQL command
	// We use "?" as placeholders to prevent "SQL Injection" (Hacking)
	stmt := `INSERT INTO wl_entry (
		creator_id, created_ts, updated_ts, 
		patient_name, patient_dob, txt_customer_id, int_customer_doc_id,
		referring_physician, complaint, triage_note, urgency, state
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	// 3. Execute the command
	result, err := d.db.ExecContext(ctx, stmt,
		create.CreatorID, ts, ts,
		create.PatientName, create.PatientDOB, create.TxtCustomerID, create.IntCustomerDocID,
		create.ReferringPhysician, create.Complaint, create.TriageNote, create.Urgency, create.State,
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
	return &store.WLEntry{
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
		State:              create.State,
	}, nil
}

func (d *Driver) ListWLEntries(ctx context.Context, find *store.FindWLEntry) ([]*store.WLEntry, error) {
	// 1. The Base Query
	// "WHERE 1 = 1" is a classic trick. It does nothing, but lets us
	// safely add "AND ..." to the end of the string later.
	query := `SELECT 
		id, creator_id, created_ts, updated_ts, 
		patient_name, patient_dob, txt_customer_id, int_customer_doc_id,
		referring_physician, complaint, triage_note, urgency, state,
		IFNULL(appt_date, ''),
		IFNULL(appt_time,''),
		IFNULL(practitioner, ''), 
		IFNULL(juvonno_appt_id,'')
	FROM wl_entry WHERE 1 = 1`

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
	if find.State != nil {
		query += " AND state = ?"
		args = append(args, *find.State)
	}

	// Fuzzy Matching for Patient Name (Requirement #1)
	if find.PatientName != nil {
		query += " AND patient_name LIKE ?"
		args = append(args, "%"+*find.PatientName+"%")
	}

	// 4. Sorting (Always show newest or most urgent first)
	query += " ORDER BY created_ts DESC"

	// 5. Run the Query
	rows, err := d.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// 6. The "Bucket" for our results
	var list []*store.WLEntry

	// 7. Loop through the database rows
	for rows.Next() {
		var entry store.WLEntry
		// Scan matches the columns in our SELECT statement to our Go struct
		err := rows.Scan(
			&entry.ID, &entry.CreatorID, &entry.CreatedTs, &entry.UpdatedTs,
			&entry.PatientName, &entry.PatientDOB, &entry.TxtCustomerID, &entry.IntCustomerDocID,
			&entry.ReferringPhysician, &entry.Complaint, &entry.TriageNote, &entry.Urgency, &entry.State,
			&entry.ApptDate, &entry.ApptTime, &entry.Practitioner, &entry.JuvonnoApptID,
		)
		if err != nil {
			return nil, err
		}
		list = append(list, &entry)
	}

	return list, nil
}

func (d *Driver) UpdateWLEntry(ctx context.Context, update *store.UpdateWLEntry) error {
	// 1. Build the "SET" part of our SQL dynamically
	set, args := []string{}, []any{}

	if v := update.State; v != nil {
		set = append(set, "state = ?")
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

	// 3. Execute: UPDATE wl_entry SET state = ?, updated_ts = ? WHERE id = ?
	query := `UPDATE wl_entry SET ` + strings.Join(set, ", ") + ` WHERE id = ?`
	_, err := d.db.ExecContext(ctx, query, args...)
	return err
}

func (d *Driver) DeleteWLEntry(ctx context.Context, delete *store.DeleteWLEntry) error {
	return nil
}
