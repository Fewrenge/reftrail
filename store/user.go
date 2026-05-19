package store

import (
	"context"
	"errors" // Needed for errors.New
	"log/slog"
	"reftrail/internal/domain"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID            domain.UserID   `json:"id"`
	Username      string          `json:"username"`
	PasswordHash  string          `json:"-"`
	Role          domain.UserRole `json:"role"`
	UserFirstName string          `json:"userFirstName"`
	UserLastName  string          `json:"userLastName"`
}

// The "Form" for logging in
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type CreateUser struct {
	Username      string          `json:"username"`
	Password      string          `json:"password"`
	Role          domain.UserRole `json:"role"`
	UserFirstName string          `json:"userFirstName"`
	UserLastName  string          `json:"userLastName"`
}

type FindUser struct {
	ID       *domain.UserID `json:"id"`
	Username *string        `json:"username"`
}

type UpdateUser struct {
	ID       domain.UserID    `json:"id"`
	Username *string          `json:"username"`
	Password *string          `json:"password"`
	Role     *domain.UserRole `json:"role"`
}

type DeleteUser struct {
	ID domain.UserID `json:"id"`
}

// --- THE MANAGER LOGIC ---

func (s *Store) Login(ctx context.Context, req *LoginRequest) (*User, error) {
	// 1. Find the user by username
	user, err := s.GetUser(ctx, &FindUser{Username: &req.Username})
	if err != nil || user == nil {
		return nil, errors.New("invalid username or password")
	}

	// 2. Compare passwords
	// Integrated the method of comparing Hash
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		return nil, errors.New("invalid username or password")
	}

	return user, nil
}

func (s *Store) SeedAdminUser(ctx context.Context) error {
	count, err := s.driver.CountUsers(ctx)
	if err != nil {
		return err
	}

	if count == 0 {
		hashed, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)

		// Fix: We need to use CreateUser struct because s.driver.CreateUser
		// likely expects the "Create" form, not the final "User" form.
		admin := &CreateUser{
			Username:      "admin",
			Password:      string(hashed), // This will be saved to password_hash in SQLite
			Role:          domain.RoleReftrailAdmin,
			UserFirstName: "Admin",
			UserLastName:  "User",
		}

		// Fix: s.driver.CreateUser returns (*User, error),
		// but SeedAdminUser only wants to return (error).
		_, err := s.driver.CreateUser(ctx, admin)
		if err == nil {
			slog.Info("SEED SUCCESS: Created admin/admin123")
		}
		return err
	}
	return nil
}

func (s *Store) CreateUser(ctx context.Context, create *CreateUser) (*User, error) {
	return s.driver.CreateUser(ctx, create)
}

func (s *Store) ListUsers(ctx context.Context, find *FindUser) ([]*User, error) {
	return s.driver.ListUsers(ctx, find)
}

func (s *Store) GetUser(ctx context.Context, find *FindUser) (*User, error) {
	list, err := s.ListUsers(ctx, find)
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, nil
	}
	return list[0], nil
}

func (s *Store) UpdateUser(ctx context.Context, update *UpdateUser) (*User, error) {
	return s.driver.UpdateUser(ctx, update)
}

func (s *Store) DeleteUser(ctx context.Context, delete *DeleteUser) error {
	if delete.ID == 1 {
		return errors.New("cannot delete the system administrator")
	}
	return s.driver.DeleteUser(ctx, delete)
}

func (s *Store) ChangeUserPassword(ctx context.Context, userID domain.UserID, newHash string) error {
	// Relay the command to the driver (the stove)
	return s.driver.ChangeUserPassword(ctx, userID, newHash)
}
