package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"reftrail/internal/domain"
	"reftrail/store"
	"strings"
)

func (d *Driver) CreateUser(ctx context.Context, create *store.CreateUser) (*store.User, error) {
	query := `INSERT INTO user (username, password_hash, role, user_first_name, user_last_name, is_archived) VALUES (?, ?, ?, ?, ?, ?)`
	_, err := d.conn(ctx).ExecContext(ctx, query, create.Username, create.Password, create.Role, create.UserFirstName, create.UserLastName, false)
	if err != nil {
		return nil, err
	}
	return &store.User{
		Username:      create.Username,
		Role:          create.Role,
		UserFirstName: create.UserFirstName,
		UserLastName:  create.UserLastName,
		IsArchived:    false,
	}, nil
}

func (d *Driver) CountUsers(ctx context.Context) (int, error) {
	var count int
	err := d.conn(ctx).QueryRowContext(ctx, "SELECT COUNT(*) FROM user").Scan(&count) // FROM user, not users
	return count, err
}

func (d *Driver) ListUsers(ctx context.Context, find *store.FindUser) ([]*store.User, error) {
	var args []any
	where := []string{"1 = 1"}

	if find.Username != "" {
		where = append(where, "username = ?")
		args = append(args, find.Username)
	}

	query := `SELECT username, password_hash, role, user_first_name, user_last_name, is_archived FROM user WHERE ` + strings.Join(where, " AND ")
	rows, err := d.conn(ctx).QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*store.User
	for rows.Next() {
		var user store.User
		if err := rows.Scan(&user.Username, &user.PasswordHash, &user.Role, &user.UserFirstName, &user.UserLastName, &user.IsArchived); err != nil {
			return nil, err
		}
		users = append(users, &user)
	}
	return users, nil
}

func (d *Driver) UpdateUserInfo(ctx context.Context, update *store.UpdateUserInfo) (*store.User, error) {
	var updates []string
	var args []any

	// Support renaming the primary key username string! (ON UPDATE CASCADE triggers automatically)
	if update.UpdatedUsername != nil {
		updates = append(updates, "username = ?")
		args = append(args, *update.UpdatedUsername)
	}
	if update.UserFirstName != nil {
		updates = append(updates, "user_first_name = ?")
		args = append(args, *update.UserFirstName)
	}
	if update.UserLastName != nil {
		updates = append(updates, "user_last_name = ?")
		args = append(args, *update.UserLastName)
	}

	// If no patch modifications were passed, return the unchanged user record
	if len(updates) == 0 {
		// return d.conn(ctx).Driver().(*Driver).GetUserByUsername(ctx, update.CurrentUsername)
	}

	// Complete the dynamic query string matching against currentUsername string
	query := fmt.Sprintf("UPDATE user SET %s WHERE username = ?", strings.Join(updates, ", "))
	args = append(args, update.CurrentUsername)

	_, err := d.conn(ctx).ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	// Determine the lookup identity handle following potential rename sequence
	targetUsername := update.CurrentUsername
	if update.UpdatedUsername != nil {
		targetUsername = *update.UpdatedUsername
	}

	return d.GetUserByUsername(ctx, targetUsername)
}

func (d *Driver) DeleteUser(ctx context.Context, delete *store.DeleteUser) error {
	query := `DELETE FROM user WHERE username = ?`
	result, err := d.conn(ctx).ExecContext(ctx, query, delete.Username)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}

func (d *Driver) UpdateUserPassword(ctx context.Context, username domain.Username, newHash string) error {
	_, err := d.conn(ctx).ExecContext(ctx, `
		UPDATE user 
		SET password_hash = ? 
		WHERE username = ?
	`, newHash, username)
	return err
}

func (d *Driver) GetUserByUsername(ctx context.Context, username domain.Username) (*store.User, error) {
	query := `SELECT username, password_hash, role, user_first_name, user_last_name, is_archived FROM user WHERE username = ?`
	var user store.User
	err := d.conn(ctx).QueryRowContext(ctx, query, username).Scan(
		&user.Username, &user.PasswordHash, &user.Role, &user.UserFirstName, &user.UserLastName, &user.IsArchived,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &user, err
}

func (d *Driver) ArchiveUser(ctx context.Context, username domain.Username) error {
	query := `UPDATE user SET is_archived = 1 WHERE username = ?`
	result, err := d.conn(ctx).ExecContext(ctx, query, username)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return err
	}
	return nil
}

// PLAN: UnarchiveUser method if we want to support that in the future

// CountActiveAdmins returns the number of active, non-archived administrators in SQLite.
func (d *Driver) CountActiveAdmins(ctx context.Context) (int, error) {
	var count int

	query := `
		SELECT COUNT(*) 
		FROM user 
		WHERE role = 'REFTRAIL_ADMIN' AND is_archived = 0
	`

	err := d.conn(ctx).QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}
