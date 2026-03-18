package sqlite

import (
	"context"
	"strings"
	"wl/store"
)

func (d *Driver) CreateUser(ctx context.Context, create *store.CreateUser) (*store.User, error) {
	stmt := `INSERT INTO user (username, password_hash, role) VALUES (?, ?, ?)`
	result, err := d.db.ExecContext(ctx, stmt, create.Username, create.Password, create.Role)
	if err != nil {
		return nil, err
	}
	id, _ := result.LastInsertId()
	return &store.User{
		ID:       int32(id),
		Username: create.Username,
		Role:     create.Role,
	}, nil
}

func (d *Driver) ListUsers(ctx context.Context, find *store.FindUser) ([]*store.User, error) {
	var args []any
	where := []string{"1 = 1"}

	if find.ID != nil {
		where = append(where, "id = ?")
		args = append(args, *find.ID)
	}
	if find.Username != nil {
		where = append(where, "username = ?")
		args = append(args, *find.Username)
	}

	query := `SELECT id, username, password_hash, role FROM user WHERE ` + strings.Join(where, " AND ")
	rows, err := d.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*store.User
	for rows.Next() {
		var user store.User
		if err := rows.Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Role); err != nil {
			return nil, err
		}
		users = append(users, &user)
	}
	return users, nil
}

func (d *Driver) UpdateUser(ctx context.Context, update *store.UpdateUser) (*store.User, error) {
	// Stub for now
	return nil, nil
}

func (d *Driver) DeleteUser(ctx context.Context, delete *store.DeleteUser) error {
	// Stub for now
	return nil
}
