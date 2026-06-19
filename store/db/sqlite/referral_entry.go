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

func nullString(s string) *string {
	cleaned := strings.TrimSpace(s)
	if cleaned == "" {
		return nil
	}
	return &cleaned
}

func (d *Driver) CreateReferralComplaint(ctx context.Context, referralID domain.ReferralID, complaint *store.ReferralComplaint) error {
	newID, err := uuid.NewV7()
	if err != nil {
		return err
	}
	idStr := newID.String()
	query := `INSERT INTO referral_complaint (id, referral_id, body_part, side, details) VALUES (?, ?, ?, ?, ?)`
	_, err = d.conn(ctx).ExecContext(ctx, query, idStr, referralID, complaint.BodyPart, complaint.Side, complaint.Details)
	return err
}

func (d *Driver) DeleteReferralComplaint(ctx context.Context, referralID domain.ReferralID) error {
	query := `DELETE FROM referral_complaint WHERE referral_id = ?`
	_, err := d.conn(ctx).ExecContext(ctx, query, referralID)
	return err
}

func (d *Driver) CreateReferralEntry(ctx context.Context, create *store.CreateReferralEntry) (*store.ReferralEntry, error) {
	newID, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	idStr := newID.String()
	ts := time.Now().Format(time.RFC3339)

	query := `INSERT INTO referral_entry (
		id, created_ts, updated_ts, creator_id, 
		patient_last_name, patient_first_name, patient_dob, patient_healthcard_number, patient_healthcard_version_code, patient_phone_number, patient_email,
		emr_patient_id, emr_referral_doc_id,
		referring_physician_id, triage_note, urgency, status, source, referral_date, consult_type, consult_type_detail
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	// Direct injection: Go automatically converts nil pointers into real database NULL values
	_, err = d.conn(ctx).ExecContext(ctx, query,
		idStr, ts, ts, string(create.CreatorUsername),
		create.PatientLastName, create.PatientFirstName, create.PatientDOB,
		create.PatientHealthcardNumber, create.PatientHealthcardVersionCode,
		create.PatientPhoneNumber, create.PatientEmail,
		create.EMRPatientID, create.EMRReferralDocID,
		create.ReferringPhysicianID, create.TriageNote, create.Urgency, create.Status, create.Source, create.ReferralDate, create.ConsultType, create.ConsultTypeDetail,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert referral entry: %w", err)
	}

	return &store.ReferralEntry{
		ID:                           domain.ReferralID(idStr),
		CreatedTs:                    ts,
		UpdatedTs:                    ts,
		CreatorUsername:              create.CreatorUsername,
		PatientLastName:              create.PatientLastName,
		PatientFirstName:             create.PatientFirstName,
		PatientDOB:                   create.PatientDOB,
		PatientHealthcardNumber:      create.PatientHealthcardNumber,
		PatientHealthcardVersionCode: create.PatientHealthcardVersionCode,
		PatientPhoneNumber:           create.PatientPhoneNumber,
		PatientEmail:                 create.PatientEmail,
		EMRPatientID:                 create.EMRPatientID,
		EMRReferralDocID:             create.EMRReferralDocID,
		ReferringPhysician:           nil, // Hydrate it later
		TriageNote:                   create.TriageNote,
		Urgency:                      create.Urgency,
		Status:                       create.Status,
		Source:                       create.Source,
		ReferralDate:                 create.ReferralDate,
		ConsultType:                  create.ConsultType,
		ConsultTypeDetail:            create.ConsultTypeDetail,
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
		re.id, re.creator_id, re.created_ts, re.updated_ts, 
		re.patient_last_name, re.patient_first_name, re.patient_dob, 
		re.patient_healthcard_number, re.patient_healthcard_version_code, re.patient_phone_number, re.patient_email,
		re.emr_patient_id, re.emr_referral_doc_id,
		re.referring_physician_id, re.triage_note, re.urgency, re.status, re.source, re.referral_date, 
		re.consult_type, re.consult_type_detail,
		p.id, p.cpso_number, p.first_name, p.last_name, p.emr_physician_id
	FROM referral_entry re
	LEFT JOIN physicians p ON re.referring_physician_id = p.id
	WHERE 1 = 1`
	var args []any

	if find.ID != nil {
		query += " AND re.id = ?"
		args = append(args, *find.ID)
	}
	if find.CreatorUsername != nil {
		query += " AND creator_id = ?"
		args = append(args, *find.CreatorUsername)
	}
	if find.PatientDOB != nil && *find.PatientDOB != "" {
		query += " AND patient_dob = ?"
		args = append(args, *find.PatientDOB)
	}

	// 1. Array Parameter Slices
	if len(find.Statuses) > 0 {
		placeholders := make([]string, len(find.Statuses))
		for i, s := range find.Statuses {
			placeholders[i] = "?"
			args = append(args, string(s))
		}
		query += fmt.Sprintf(" AND status IN (%s)", strings.Join(placeholders, ", "))
	}

	if len(find.Urgencies) > 0 {
		placeholders := make([]string, len(find.Urgencies))
		for i, u := range find.Urgencies {
			placeholders[i] = "?"
			args = append(args, string(u))
		}
		query += fmt.Sprintf(" AND urgency IN (%s)", strings.Join(placeholders, ", "))
	}

	if len(find.Sources) > 0 {
		placeholders := make([]string, len(find.Sources))
		for i, src := range find.Sources {
			placeholders[i] = "?"
			args = append(args, string(src))
		}
		query += fmt.Sprintf(" AND source IN (%s)", strings.Join(placeholders, ", "))
	}

	// 2. Clinical Workflow Checkbox Filters (Consult Types)
	if len(find.ConsultTypes) > 0 {
		placeholders := make([]string, len(find.ConsultTypes))
		for i, ct := range find.ConsultTypes {
			placeholders[i] = "?"
			args = append(args, string(ct))
		}
		query += fmt.Sprintf(" AND consult_type IN (%s)", strings.Join(placeholders, ", "))
	}

	if len(find.BodyParts) > 0 {
		placeholders := make([]string, len(find.BodyParts))
		for i, bp := range find.BodyParts {
			placeholders[i] = "?"
			args = append(args, bp)
		}
		query += fmt.Sprintf(
			" AND re.id IN (SELECT referral_id FROM referral_complaint WHERE body_part IN (%s))",
			strings.Join(placeholders, ", "),
		)
	}

	// 3. Filter by Associated Junction Tags (Strict AND matching)
	if len(find.TagNames) > 0 {
		placeholders := make([]string, len(find.TagNames))
		for i, t := range find.TagNames {
			placeholders[i] = "?"
			args = append(args, t)
		}
		query += fmt.Sprintf(` AND re.id IN (
			SELECT referral_id FROM referral_tag 
			WHERE tag_name IN (%s) 
			GROUP BY referral_id 
			HAVING COUNT(DISTINCT tag_name) = ?
		)`, strings.Join(placeholders, ", "))
		args = append(args, len(find.TagNames))
	}

	// Patient Directory Lookups & Searches
	// Left-anchored fuzzy searches preserve B-Tree index optimization
	if find.PatientLastName != nil && find.PatientFirstName != nil && *find.PatientLastName == *find.PatientFirstName {
		// If both pointers hold the identical search query string, run an OR group lookup
		searchTerm := *find.PatientLastName + "%"
		query += " AND (patient_last_name LIKE ? OR patient_first_name LIKE ?)"
		args = append(args, searchTerm, searchTerm)
	} else {
		// Fallback to standalone isolated filters if inputs are distinct
		if find.PatientLastName != nil && *find.PatientLastName != "" {
			query += " AND patient_last_name LIKE ?"
			args = append(args, *find.PatientLastName+"%")
		}
		if find.PatientFirstName != nil && *find.PatientFirstName != "" {
			query += " AND patient_first_name LIKE ?"
			args = append(args, *find.PatientFirstName+"%")
		}
	}

	// Left Anchored
	if find.PatientHealthcardNumber != nil && *find.PatientHealthcardNumber != "" {
		query += " AND patient_healthcard_number LIKE ?"
		args = append(args, *find.PatientHealthcardNumber+"%")
	}

	if find.ReferringPhysicianID != nil && *find.ReferringPhysicianID != "" {
		query += " AND re.referring_physician_id = ?"
		args = append(args, *find.ReferringPhysicianID)
	}

	// Fuzzy name matching (Checks across first name, last name, and CPSO number)
	if find.ReferringPhysicianName != nil && *find.ReferringPhysicianName != "" {
		term := "%" + *find.ReferringPhysicianName + "%"
		query += " AND (p.first_name LIKE ? OR p.last_name LIKE ? OR p.cpso_number LIKE ?)"
		args = append(args, term, term, term)
	}

	// Phone Number Search (Advanced Filter Component)
	if find.PatientPhoneNumber != nil && *find.PatientPhoneNumber != "" {
		query += " AND patient_phone_number LIKE ?"
		args = append(args, "%"+*find.PatientPhoneNumber+"%")
	}

	// Date Bounds
	if find.ReferralDateFrom != nil && *find.ReferralDateFrom != "" {
		query += " AND referral_date >= ?"
		args = append(args, *find.ReferralDateFrom)
	}
	if find.ReferralDateTo != nil && *find.ReferralDateTo != "" {
		query += " AND referral_date <= ?"
		args = append(args, *find.ReferralDateTo)
	}

	// Sorting
	query += " ORDER BY created_ts DESC"

	// Pagination
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

		// 1. Declare targets as 'any' so Go's SQL driver can natively accept both strings and NULLs
		var (
			patientHealthcardNumberTarget      any
			patientHealthcardVersionCodeTarget any
			patientPhoneNumberTarget           any
			patientEmailTarget                 any
			emrPatientIDTarget                 any
			emrReferralDocIDTarget             any
			referringPhysicianIDTarget         any
			triageNoteTarget                   any
			consultTypeDetailTarget            any

			physicianIDTarget         any
			physicianCPSONumberTarget any
			physicianFirstNameTarget  any
			physicianLastNameTarget   any
			physicianEMRIDTarget      any
		)

		// 2. Scan directly into the memory locations of our open target interfaces
		err := rows.Scan(
			&entry.ID, &entry.CreatorUsername, &entry.CreatedTs, &entry.UpdatedTs,
			&entry.PatientLastName, &entry.PatientFirstName, &entry.PatientDOB,
			&patientHealthcardNumberTarget,
			&patientHealthcardVersionCodeTarget,
			&patientPhoneNumberTarget,
			&patientEmailTarget,
			&emrPatientIDTarget,
			&emrReferralDocIDTarget,
			&referringPhysicianIDTarget,
			&triageNoteTarget,
			&entry.Urgency, &entry.Status, &entry.Source, &entry.ReferralDate,
			&entry.ConsultType,
			&consultTypeDetailTarget,

			&physicianIDTarget,
			&physicianCPSONumberTarget,
			&physicianFirstNameTarget,
			&physicianLastNameTarget,
			&physicianEMRIDTarget,
		)
		if err != nil {
			return nil, err
		}

		// 3. Inline type-assertion helper to extract the string value safely.
		// If the database column was NULL, the target is nil, so this returns an empty string "".
		getStringValue := func(val any) string {
			if val == nil {
				return ""
			}
			if str, ok := val.(string); ok {
				return str
			}
			return ""
		}

		// 4. Process values through our type assertions and nullString helper
		entry.PatientHealthcardNumber = nullString(getStringValue(patientHealthcardNumberTarget))
		entry.PatientHealthcardVersionCode = nullString(getStringValue(patientHealthcardVersionCodeTarget))
		entry.PatientPhoneNumber = nullString(getStringValue(patientPhoneNumberTarget))
		entry.PatientEmail = nullString(getStringValue(patientEmailTarget))
		entry.EMRPatientID = nullString(getStringValue(emrPatientIDTarget))
		entry.EMRReferralDocID = nullString(getStringValue(emrReferralDocIDTarget))
		entry.ReferringPhysicianID = nullString(getStringValue(referringPhysicianIDTarget))
		entry.ConsultTypeDetail = nullString(getStringValue(consultTypeDetailTarget))
		entry.TriageNote = getStringValue(triageNoteTarget)

		// 5. Populate structured nested Physician object if a relation ID was matched
		physicianID := getStringValue(physicianIDTarget)
		if physicianID != "" {
			entry.ReferringPhysician = &store.ReferralPhysician{
				ID:             physicianID,
				CPSONumber:     nullString(getStringValue(physicianCPSONumberTarget)),
				FirstName:      getStringValue(physicianFirstNameTarget),
				LastName:       getStringValue(physicianLastNameTarget),
				EMRPhysicianID: nullString(getStringValue(physicianEMRIDTarget)),
			}
		}

		list = append(list, &entry)
	}

	return list, nil
}

func (d *Driver) GetReferralEntriesCount(ctx context.Context, find *store.FindReferralEntry) (int, error) {
	// 1. Target the base rows using a COUNT query
	query := `SELECT COUNT(1) 
	          FROM referral_entry re
	          LEFT JOIN physicians p ON re.referring_physician_id = p.id 
	          WHERE 1 = 1`
	var args []any

	// Exact matches (High performance index hits)
	if find.ID != nil {
		query += " AND re.id = ?"
		args = append(args, *find.ID)
	}
	if find.CreatorUsername != nil {
		query += " AND creator_id = ?"
		args = append(args, *find.CreatorUsername)
	}
	if find.PatientDOB != nil && *find.PatientDOB != "" {
		query += " AND patient_dob = ?"
		args = append(args, *find.PatientDOB)
	}

	// 2. Multi-Select Enums
	if len(find.Statuses) > 0 {
		placeholders := make([]string, len(find.Statuses))
		for i, s := range find.Statuses {
			placeholders[i] = "?"
			args = append(args, string(s))
		}
		query += fmt.Sprintf(" AND status IN (%s)", strings.Join(placeholders, ", "))
	}

	if len(find.Urgencies) > 0 {
		placeholders := make([]string, len(find.Urgencies))
		for i, u := range find.Urgencies {
			placeholders[i] = "?"
			args = append(args, string(u))
		}
		query += fmt.Sprintf(" AND urgency IN (%s)", strings.Join(placeholders, ", "))
	}

	if len(find.Sources) > 0 {
		placeholders := make([]string, len(find.Sources))
		for i, src := range find.Sources {
			placeholders[i] = "?"
			args = append(args, string(src))
		}
		query += fmt.Sprintf(" AND source IN (%s)", strings.Join(placeholders, ", "))
	}

	// 3. Clinical Workflow Checkbox Filters (Consult Types)
	if len(find.ConsultTypes) > 0 {
		placeholders := make([]string, len(find.ConsultTypes))
		for i, ct := range find.ConsultTypes {
			placeholders[i] = "?"
			args = append(args, string(ct))
		}
		query += fmt.Sprintf(" AND consult_type IN (%s)", strings.Join(placeholders, ", "))
	}

	// 4. Filter by Associated Junction Tags (Strict AND matching)
	if len(find.TagNames) > 0 {
		placeholders := make([]string, len(find.TagNames))
		for i, t := range find.TagNames {
			placeholders[i] = "?"
			args = append(args, t)
		}
		query += fmt.Sprintf(` AND re.id IN (
			SELECT referral_id FROM referral_tag 
			WHERE tag_name IN (%s) 
			GROUP BY referral_id 
			HAVING COUNT(DISTINCT tag_name) = ?
		)`, strings.Join(placeholders, ", "))
		args = append(args, len(find.TagNames))
	}

	// Patient Directory Lookups & Searches
	if find.PatientLastName != nil && find.PatientFirstName != nil && *find.PatientLastName == *find.PatientFirstName {
		// If both pointers hold the identical search query string, run an OR group lookup
		searchTerm := *find.PatientLastName + "%"
		query += " AND (patient_last_name LIKE ? OR patient_first_name LIKE ?)"
		args = append(args, searchTerm, searchTerm)
	} else {
		// Fallback to standalone isolated filters if inputs are distinct
		if find.PatientLastName != nil && *find.PatientLastName != "" {
			query += " AND patient_last_name LIKE ?"
			args = append(args, *find.PatientLastName+"%")
		}
		if find.PatientFirstName != nil && *find.PatientFirstName != "" {
			query += " AND patient_first_name LIKE ?"
			args = append(args, *find.PatientFirstName+"%")
		}
	}

	if find.ReferringPhysicianID != nil && *find.ReferringPhysicianID != "" {
		query += " AND re.referring_physician_id = ?"
		args = append(args, *find.ReferringPhysicianID)
	}

	// Fuzzy match criteria against relational physician first name, last name, or CPSO
	if find.ReferringPhysicianName != nil && *find.ReferringPhysicianName != "" {
		term := "%" + *find.ReferringPhysicianName + "%"
		query += " AND (p.first_name LIKE ? OR p.last_name LIKE ? OR p.cpso_number LIKE ?)"
		args = append(args, term, term, term)
	}

	if find.PatientHealthcardNumber != nil && *find.PatientHealthcardNumber != "" {
		query += " AND patient_healthcard_number LIKE ?"
		args = append(args, *find.PatientHealthcardNumber+"%")
	}

	// 6. Date Ranges
	if find.ReferralDateFrom != nil && *find.ReferralDateFrom != "" {
		query += " AND referral_date >= ?"
		args = append(args, *find.ReferralDateFrom)
	}
	if find.ReferralDateTo != nil && *find.ReferralDateTo != "" {
		query += " AND referral_date <= ?"
		args = append(args, *find.ReferralDateTo)
	}

	// 7. Execute the query
	var count int
	err := d.conn(ctx).QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed executing count context lookup: %w", err)
	}

	return count, nil
}

// For miscellaneous updates (e.g., correcting a typo, changing urgency, etc.)
func (d *Driver) UpdateReferralEntry(ctx context.Context, update *store.UpdateReferralEntry) error {
	// 1. Build the "SET" part of our SQL dynamically
	set, args := []string{}, []any{}

	// --- Workflow & Core Triage ---
	if v := update.Status; v != nil {
		set = append(set, "status = ?")
		args = append(args, *v)
	}
	if v := update.Urgency; v != nil {
		set = append(set, "urgency = ?")
		args = append(args, *v)
	}
	if v := update.Source; v != nil {
		set = append(set, "source = ?")
		args = append(args, *v)
	}
	if v := update.TriageNote; v != nil {
		set = append(set, "triage_note = ?")
		args = append(args, *v)
	}

	// --- Clinical Data ---
	if v := update.ReferringPhysicianID; v != nil {
		set = append(set, "referring_physician_id = ?")
		args = append(args, *v)
	}

	if v := update.ConsultType; v != nil {
		set = append(set, "consult_type = ?")
		args = append(args, *v)
	}
	if v := update.ReferralDate; v != nil {
		set = append(set, "referral_date = ?")
		args = append(args, *v)
	}

	// --- EMR Integration Links ---
	if v := update.EMRPatientID; v != nil {
		set = append(set, "emr_patient_id = ?")
		args = append(args, *v)
	}
	if v := update.EMRReferralDocID; v != nil {
		set = append(set, "emr_referral_doc_id = ?")
		args = append(args, *v)
	}

	// Safety Guard: If no columns are targeted for updates, skip to avoid executing bad SQL syntax
	if len(set) == 0 {
		return nil
	}

	// Update the timestamp automatically every single time
	// set = append(set, "updated_ts = ?")
	// args = append(args, time.Now().Format(time.RFC3339))

	// Add the referral entry ID for the WHERE clause
	args = append(args, update.ID)

	// Execute query
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
