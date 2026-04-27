package sqlite

import (
	"context"
	"reftrail/store"
	"time"
)

func (d *Driver) CreateReferralLog(ctx context.Context, create *store.ReferralLog) (*store.ReferralLog, error) {
	ts := time.Now().Unix()
	stmt := `INSERT INTO referral_log (entry_id, user_id, old_state, new_state, note, created_ts) 
			 VALUES (?, ?, ?, ?, ?, ?)`

	result, err := d.db.ExecContext(ctx, stmt,
		create.EntryID, create.UserID, create.OldState, create.NewState, create.Note, ts,
	)
	if err != nil {
		return nil, err
	}

	id, _ := result.LastInsertId()
	create.ID = int32(id)
	create.CreatedTs = ts
	return create, nil
}

func (d *Driver) ListReferralLogs(ctx context.Context, entryID int32) ([]*store.ReferralLog, error) {
	query := `SELECT id, entry_id, user_id, old_state, new_state, note, created_ts 
			  FROM referral_log WHERE entry_id = ? ORDER BY created_ts DESC`

	rows, err := d.db.QueryContext(ctx, query, entryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*store.ReferralLog
	for rows.Next() {
		var l store.ReferralLog
		if err := rows.Scan(&l.ID, &l.EntryID, &l.UserID, &l.OldState, &l.NewState, &l.Note, &l.CreatedTs); err != nil {
			return nil, err
		}
		logs = append(logs, &l)
	}
	return logs, nil
}
