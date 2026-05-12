package sqlite

import (
	"context"
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
