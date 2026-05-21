package sqlite

import (
	"context"
	"reftrail/internal/domain"
	"reftrail/store"
)

func (d *Driver) CreateReferralTag(ctx context.Context, create *store.CreateReferralTag) (*store.ReferralTag, error) {
	query := `INSERT INTO referral_tag_definition (name, description) 
              VALUES (?, ?) RETURNING id, name, description`

	var tag store.ReferralTag
	err := d.conn(ctx).QueryRowContext(ctx, query, create.Name, create.Description).Scan(
		&tag.ID, &tag.Name, &tag.Description,
	)
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

func (d *Driver) ListReferralTags(ctx context.Context) ([]*store.ReferralTag, error) {
	query := `SELECT id, name, description
              FROM referral_tag_definition 
              ORDER BY name ASC`

	rows, err := d.conn(ctx).QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tags := []*store.ReferralTag{}
	for rows.Next() {
		var tag store.ReferralTag
		if err := rows.Scan(&tag.ID, &tag.Name, &tag.Description); err != nil {
			return nil, err
		}
		tags = append(tags, &tag)
	}
	return tags, nil
}

func (d *Driver) ListAllLinkedReferralTags(ctx context.Context) ([]*store.LinkedReferralTagRow, error) {
	// This joins the link table with the string names table
	query := `
		SELECT rt.referral_id, rtd.name 
		FROM referral_tag rt
		JOIN referral_tag_definition rtd ON rt.tag_id = rtd.id
	`

	rows, err := d.conn(ctx).QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*store.LinkedReferralTagRow
	for rows.Next() {
		var row store.LinkedReferralTagRow
		// Scan both the patient's UUID and the tag's string name
		if err := rows.Scan(&row.ReferralID, &row.TagName); err != nil {
			return nil, err
		}
		results = append(results, &row)
	}
	return results, nil
}

func (d *Driver) DeleteReferralTag(ctx context.Context, delete *store.DeleteReferralTag) error {
	// Deleting from the definition table triggers the cascade in the junction table
	query := `DELETE FROM referral_tag_definition WHERE id = ?`

	result, err := d.conn(ctx).ExecContext(ctx, query, delete.ID)
	if err != nil {
		return err
	}

	// Optional: Check if we actually deleted something
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrTagNotFound
	}

	return nil
}

func (d *Driver) AssignTagToReferral(ctx context.Context, referralID domain.ReferralID, tagID int64) error {
	query := `INSERT OR IGNORE INTO referral_tag (referral_id, tag_id) 
              VALUES (?, ?)`
	_, err := d.conn(ctx).ExecContext(ctx, query, referralID, tagID)
	return err
}

func (d *Driver) RemoveTagFromReferral(ctx context.Context, referralID domain.ReferralID, tagID int64) error {
	query := `DELETE FROM referral_tag WHERE referral_id = ? AND tag_id = ?`
	_, err := d.conn(ctx).ExecContext(ctx, query, referralID, tagID)
	return err
}
