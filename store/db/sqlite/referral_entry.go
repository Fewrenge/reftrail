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
		referring_physician, triage_note, urgency, status, source, referral_date
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	// Execute the command
	_, err = d.conn(ctx).ExecContext(ctx, query,
		idStr, ts, ts, string(create.CreatorUsername),
		create.PatientLastName, create.PatientFirstName, create.PatientDOB,
		create.PatientHealthcardNumber, create.PatientHealthcardVersionCode,
		create.TxtCustomerID, create.IntCustomerDocID,
		create.ReferringPhysician, create.TriageNote, create.Urgency, create.Status, create.Source, create.ReferralDate,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert referral entry for patient %s, %s (creator_username: %s): %w",
			create.PatientLastName, create.PatientFirstName, create.CreatorUsername, err)
	}
	return &store.ReferralEntry{
		ID:                           domain.ReferralID(idStr), // Cast to custom type
		CreatedTs:                    ts,
		UpdatedTs:                    ts,
		CreatorUsername:              create.CreatorUsername,
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
		ReferralDate:                 create.ReferralDate,
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
	// Base query targeting strictly singular top-level records
	query := `SELECT 
		id, creator_id, created_ts, updated_ts, 
		patient_last_name, patient_first_name, patient_dob, patient_healthcard_number, patient_healthcard_version_code,
		txt_customer_id, int_customer_doc_id,
		referring_physician, triage_note, urgency, status, source, referral_date
	FROM referral_entry WHERE 1 = 1`

	var args []any

	if find.ID != nil {
		query += " AND id = ?"
		args = append(args, *find.ID)
	}
	if find.CreatorUsername != nil {
		query += " AND creator_id = ?"
		args = append(args, *find.CreatorUsername)
	}

	// 1. Array Parameter Slices
	if len(find.Statuses) > 0 {
		query += " AND status IN ("
		for i, s := range find.Statuses {
			if i > 0 {
				query += ", "
			}
			query += "?"
			args = append(args, string(s))
		}
		query += ")"
	}

	if len(find.Urgencies) > 0 {
		query += " AND urgency IN ("
		for i, u := range find.Urgencies {
			if i > 0 {
				query += ", "
			}
			query += "?"
			args = append(args, string(u))
		}
		query += ")"
	}

	if len(find.Sources) > 0 {
		query += " AND source IN ("
		for i, src := range find.Sources {
			if i > 0 {
				query += ", "
			}
			query += "?"
			args = append(args, string(src))
		}
		query += ")"
	}

	// 2. Filter by Associated Complaints (Body Parts) via isolated lookup subquery
	if len(find.BodyParts) > 0 {
		query += " AND id IN (SELECT referral_id FROM referral_complaint WHERE body_part IN ("
		for i, bp := range find.BodyParts {
			if i > 0 {
				query += ", "
			}
			query += "?"
			args = append(args, bp)
		}
		query += "))"
	}

	// 3. Filter by Associated Junction Tags via isolated lookup subquery
	if len(find.TagNames) > 0 {
		query += " AND id IN (SELECT referral_id FROM referral_tag WHERE tag_name IN ("
		for i, t := range find.TagNames {
			if i > 0 {
				query += ", "
			}
			query += "?"
			args = append(args, t)
		}
		query += "))"
	}

	// 4. Fuzzy Searches
	if find.PatientLastName != nil && *find.PatientLastName != "" {
		query += " AND patient_last_name LIKE ?"
		args = append(args, "%"+*find.PatientLastName+"%")
	}
	if find.PatientFirstName != nil && *find.PatientFirstName != "" {
		query += " AND patient_first_name LIKE ?"
		args = append(args, "%"+*find.PatientFirstName+"%")
	}
	if find.ReferringPhysician != nil && *find.ReferringPhysician != "" {
		query += " AND referring_physician LIKE ?"
		args = append(args, "%"+*find.ReferringPhysician+"%")
	}
	if find.PatientHealthcardNumber != nil && *find.PatientHealthcardNumber != "" {
		query += " AND patient_healthcard_number LIKE ?"
		args = append(args, "%"+*find.PatientHealthcardNumber+"%")
	}

	// 5. Date Bounds
	if find.ReferralDateFrom != nil && *find.ReferralDateFrom != "" {
		query += " AND referral_date >= ?"
		args = append(args, *find.ReferralDateFrom)
	}
	if find.ReferralDateTo != nil && *find.ReferralDateTo != "" {
		query += " AND referral_date <= ?"
		args = append(args, *find.ReferralDateTo)
	}

	query += " ORDER BY created_ts DESC"

	if find.Limit != nil {
		query += " LIMIT ?"
		args = append(args, *find.Limit)
	}
	if find.Offset != nil {
		query += " OFFSET ?"
		args = append(args, *find.Offset)
	}

	rows, err := d.conn(ctx).QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*store.ReferralEntry
	for rows.Next() {
		var entry store.ReferralEntry
		err := rows.Scan(
			&entry.ID, &entry.CreatorUsername, &entry.CreatedTs, &entry.UpdatedTs,
			&entry.PatientLastName, &entry.PatientFirstName, &entry.PatientDOB,
			&entry.PatientHealthcardNumber, &entry.PatientHealthcardVersionCode,
			&entry.TxtCustomerID, &entry.IntCustomerDocID,
			&entry.ReferringPhysician, &entry.TriageNote, &entry.Urgency, &entry.Status, &entry.Source, &entry.ReferralDate,
		)
		if err != nil {
			return nil, err
		}
		list = append(list, &entry)
	}

	return list, nil
}

// For miscellaneous updates (e.g., correcting a typo, changing urgency, etc.)
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
	err := d.conn(ctx).QueryRowContext(ctx, "SELECT status FROM referral_entry WHERE id = ?", id).Scan(&status)
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
