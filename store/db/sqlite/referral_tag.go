package sqlite

import (
	"context"
	"fmt"
	"reftrail/store"
)

func (d *Driver) CreateReferralTag(ctx context.Context, create *store.CreateReferralTag) (*store.ReferralTag, error) {
	query := `INSERT INTO tag_definition (name, description, created_ts) 
              VALUES (?, ?, datetime('now')) RETURNING id, name, description, created_ts`

	var tag store.ReferralTag
	err := d.db.QueryRowContext(ctx, query, create.Name, create.Description).Scan(
		&tag.ID, &tag.Name, &tag.Description, &tag.CreatedTS,
	)
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

func (d *Driver) ListReferralTags(ctx context.Context) ([]*store.ReferralTag, error) {
	query := `SELECT id, name, description, created_ts 
              FROM tag_definition 
              ORDER BY name ASC`

	rows, err := d.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tags := []*store.ReferralTag{}
	for rows.Next() {
		var tag store.ReferralTag
		if err := rows.Scan(&tag.ID, &tag.Name, &tag.Description, &tag.CreatedTS); err != nil {
			return nil, err
		}
		tags = append(tags, &tag)
	}
	return tags, nil
}

func (d *Driver) DeleteReferralTag(ctx context.Context, delete *store.DeleteReferralTag) error {
	// Deleting from the definition table triggers the cascade in the junction table
	query := `DELETE FROM tag_definition WHERE id = ?`

	result, err := d.db.ExecContext(ctx, query, delete.ID)
	if err != nil {
		return err
	}

	// Optional: Check if we actually deleted something
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("tag with ID %d not found", delete.ID)
	}

	return nil
}
