package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"reftrail/store"

	uuid "github.com/google/uuid"
)

func (d *Driver) CreateReferralPhysician(ctx context.Context, p *store.ReferralPhysician) (*store.ReferralPhysician, error) {
	newID, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	idStr := newID.String()

	query := `INSERT INTO referral_physician (id, cpso_number, first_name, last_name, emr_physician_id) 
	          VALUES (?, ?, ?, ?, ?)`

	_, err = d.conn(ctx).ExecContext(ctx, query, idStr, p.CPSONumber, p.FirstName, p.LastName, p.EMRPhysicianID)
	if err != nil {
		return nil, fmt.Errorf("failed to insert physician: %w", err)
	}

	// Return the saved physician including its generated ID
	return &store.ReferralPhysician{
		ID:             idStr,
		CPSONumber:     p.CPSONumber,
		FirstName:      p.FirstName,
		LastName:       p.LastName,
		EMRPhysicianID: p.EMRPhysicianID,
	}, nil
}

func (d *Driver) ListReferralPhysicians(ctx context.Context) ([]*store.ReferralPhysician, error) {
	query := `SELECT id, cpso_number, first_name, last_name, emr_physician_id FROM referral_physician`

	rows, err := d.conn(ctx).QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute list physicians query: %w", err)
	}
	defer rows.Close()

	var list []*store.ReferralPhysician
	for rows.Next() {
		var p store.ReferralPhysician
		var cpsoTarget any
		var emrTarget any

		err := rows.Scan(&p.ID, &cpsoTarget, &p.FirstName, &p.LastName, &emrTarget)
		if err != nil {
			return nil, fmt.Errorf("failed to scan physician row: %w", err)
		}

		// Inline helper to resolve standard SQLite null pointer strings cleanly
		getStringPointer := func(val any) *string {
			if val == nil {
				return nil
			}
			if str, ok := val.(string); ok {
				if str == "" {
					return nil
				}
				return &str
			}
			return nil
		}

		p.CPSONumber = getStringPointer(cpsoTarget)
		p.EMRPhysicianID = getStringPointer(emrTarget)
		list = append(list, &p)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("physician row streaming error iteration failure: %w", err)
	}

	return list, nil
}

// GetPhysicianByID fetches an individual doctor record by its unique database ID.
func (d *Driver) GetReferralPhysicianByID(ctx context.Context, id string) (*store.ReferralPhysician, error) {
	query := `SELECT id, cpso_number, first_name, last_name, emr_physician_id 
	          FROM referral_physician WHERE id = ?`

	var p store.ReferralPhysician
	var cpsoTarget any
	var emrTarget any

	err := d.conn(ctx).QueryRowContext(ctx, query, id).Scan(&p.ID, &cpsoTarget, &p.FirstName, &p.LastName, &emrTarget)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("physician not found for id %s", id)
	} else if err != nil {
		return nil, fmt.Errorf("failed to query physician by id: %w", err)
	}

	getStringPointer := func(val any) *string {
		if val == nil {
			return nil
		}
		if str, ok := val.(string); ok {
			if str == "" {
				return nil
			}
			return &str
		}
		return nil
	}

	p.CPSONumber = getStringPointer(cpsoTarget)
	p.EMRPhysicianID = getStringPointer(emrTarget)

	return &p, nil
}

// UpdatePhysician allows managing a physician's record (e.g., adding an EMR link or CPSO number).
func (d *Driver) UpdateReferralPhysician(ctx context.Context, p *store.ReferralPhysician) error {
	query := `UPDATE referral_physician 
	          SET cpso_number = ?, first_name = ?, last_name = ?, emr_physician_id = ? 
	          WHERE id = ?`

	res, err := d.conn(ctx).ExecContext(ctx, query, p.CPSONumber, p.FirstName, p.LastName, p.EMRPhysicianID, p.ID)
	if err != nil {
		return fmt.Errorf("failed to execute update physician query: %w", err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("cannot update physician: record with id %s not found", p.ID)
	}

	return nil
}

// FindReferralPhysicians searches and filters the physicians table dynamically
func (d *Driver) FindReferralPhysicians(ctx context.Context, find *store.FindReferralPhysician) ([]*store.ReferralPhysician, error) {
	query := `SELECT id, cpso_number, first_name, last_name, emr_physician_id FROM referral_physician WHERE 1 = 1`
	var args []any

	if find.ID != nil && *find.ID != "" {
		query += " AND id = ?"
		args = append(args, *find.ID)
	}
	if find.CPSONumber != nil && *find.CPSONumber != "" {
		query += " AND cpso_number = ?"
		args = append(args, *find.CPSONumber)
	}
	if find.FirstName != nil && *find.FirstName != "" {
		query += " AND first_name LIKE ?"
		args = append(args, *find.FirstName+"%")
	}
	if find.LastName != nil && *find.LastName != "" {
		query += " AND last_name LIKE ?"
		args = append(args, *find.LastName+"%")
	}
	if find.EMRPhysicianID != nil && *find.EMRPhysicianID != "" {
		query += " AND emr_physician_id = ?"
		args = append(args, *find.EMRPhysicianID)
	}

	// Dynamic multi-column fuzzy search group
	if find.GeneralSearch != nil && *find.GeneralSearch != "" {
		term := "%" + *find.GeneralSearch + "%"
		query += " AND (first_name LIKE ? OR last_name LIKE ? OR cpso_number LIKE ?)"
		args = append(args, term, term, term)
	}

	rows, err := d.conn(ctx).QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed executing find referral physicians query: %w", err)
	}
	defer rows.Close()

	var list []*store.ReferralPhysician
	for rows.Next() {
		var p store.ReferralPhysician
		var cpsoTarget any
		var emrTarget any

		err := rows.Scan(&p.ID, &cpsoTarget, &p.FirstName, &p.LastName, &emrTarget)
		if err != nil {
			return nil, fmt.Errorf("failed to scan referral physician row: %w", err)
		}

		getStringPointer := func(val any) *string {
			if val == nil {
				return nil
			}
			if str, ok := val.(string); ok {
				if str == "" {
					return nil
				}
				return &str
			}
			return nil
		}

		p.CPSONumber = getStringPointer(cpsoTarget)
		p.EMRPhysicianID = getStringPointer(emrTarget)
		list = append(list, &p)
	}

	return list, nil
}

func (d *Driver) DeleteReferralPhysician(ctx context.Context, delete *store.DeleteReferralPhysician) error {
	// Guard against empty structures or missing IDs
	if delete == nil || delete.ID == "" {
		return fmt.Errorf("cannot delete referral physician: target ID is missing")
	}

	query := `DELETE FROM referral_physician WHERE id = ?`

	res, err := d.conn(ctx).ExecContext(ctx, query, delete.ID)
	if err != nil {
		return fmt.Errorf("failed to execute delete referral physician operation: %w", err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("cannot delete referral physician: record with id %s not found", delete.ID)
	}

	return nil
}
