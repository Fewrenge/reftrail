package sqlite

import (
	"context"
	"reftrail/internal/domain"
	"reftrail/store"
	"time"

	uuid "github.com/google/uuid"
)

func (d *Driver) CreateReferralLog(ctx context.Context, create *store.ReferralLog) (*store.ReferralLog, error) {
	newID, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	idStr := newID.String()

	ts := time.Now().Format(time.RFC3339)
	query := `INSERT INTO referral_log (id, referral_id, user_id, old_status, new_status, note, created_ts) 
			 VALUES (?, ?, ?, ?, ?, ?, ?)`

	_, err = d.conn(ctx).ExecContext(ctx, query,
		idStr, create.EntryID, create.UserID, create.OldStatus, create.NewStatus, create.Note, ts,
	)
	if err != nil {
		return nil, err
	}

	create.ID = domain.ReferralLogID(idStr)
	create.CreatedTs = ts

	return create, nil
}

func (d *Driver) ListReferralLogs(ctx context.Context, entryID domain.ReferralID) ([]*store.ReferralLog, error) {
	query := `
		SELECT l.id, l.referral_id, l.user_id, u.username, u.user_first_name, u.user_last_name, l.old_status, l.new_status, l.note, l.created_ts 
		FROM referral_log l
		INNER JOIN user u ON l.user_id = u.id
		WHERE l.referral_id = ? 
		ORDER BY l.created_ts DESC`

	rows, err := d.conn(ctx).QueryContext(ctx, query, entryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*store.ReferralLog
	for rows.Next() {
		var l store.ReferralLog
		if err := rows.Scan(&l.ID, &l.EntryID, &l.UserID, &l.Username, &l.UserFirstName, &l.UserLastName, &l.OldStatus, &l.NewStatus, &l.Note, &l.CreatedTs); err != nil {
			return nil, err
		}
		logs = append(logs, &l)
	}
	return logs, nil
}
