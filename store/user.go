package store

import (
	"context"
	"errors" // Needed for errors.New
	"log/slog"
	"reftrail/internal/domain"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Username      domain.Username `json:"username"`
	PasswordHash  string          `json:"-"`
	Role          domain.UserRole `json:"role"`
	UserFirstName string          `json:"userFirstName"`
	UserLastName  string          `json:"userLastName"`
	IsArchived    bool            `json:"isArchived"`
}

type UserPublicInfo struct {
	Username      domain.Username `json:"username"`
	UserFirstName string          `json:"userFirstName"`
	UserLastName  string          `json:"userLastName"`
}

// The "Form" for logging in
type LoginRequest struct {
	Username domain.Username `json:"username"`
	Password string          `json:"password"`
}

type CreateUser struct {
	Username      domain.Username `json:"username"`
	Password      string          `json:"password"`
	Role          domain.UserRole `json:"role"`
	UserFirstName string          `json:"userFirstName"`
	UserLastName  string          `json:"userLastName"`
}

type FindUser struct {
	Username domain.Username `json:"username"`

	// Optional fields
	Role          *domain.UserRole `json:"role,omitempty"`
	UserFirstName *string          `json:"userFirstName,omitempty"`
	UserLastName  *string          `json:"userLastName,omitempty"`
}

type UpdateUserInfo struct {
	CurrentUsername domain.Username  `json:"currentUsername"` // Used to find the user to update
	UpdatedUsername *domain.Username `json:"updatedUsername"`
	UserFirstName   *string          `json:"userFirstName"`
	UserLastName    *string          `json:"userLastName"`
	// Password        *string          `json:"password"`
	// Role            *domain.UserRole `json:"role"`
}

type DeleteUser struct {
	Username domain.Username `json:"username"`
}

// --- THE MANAGER LOGIC ---

func (s *Store) Login(ctx context.Context, req *LoginRequest) (*User, error) {
	// 1. Find the user by username
	user, err := s.GetUser(ctx, &FindUser{Username: req.Username}) // req.Username is a direct value converted by implicit dereferencing
	if err != nil || user == nil {
		return nil, domain.ErrInvalidCredentials
	}

	// 2. Compare passwords
	// Integrated the method of comparing Hash
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	// 3. Check if the user is archived
	// Do this after comparing passwords to avoid giving away information about which usernames are valid
	if user.IsArchived {
		return nil, domain.ErrUserArchived
	}

	return user, nil
}

func (s *Store) SeedAdminUser(ctx context.Context) error {
	count, err := s.driver.CountUsers(ctx)
	if err != nil {
		return err
	}

	if count == 0 {
		hashed, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost) // TODO: put default password in .env file

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

func (s *Store) UpdateUserInfo(ctx context.Context, update *UpdateUserInfo) (*User, error) {
	return s.driver.UpdateUserInfo(ctx, update)
}

func (s *Store) DeleteUser(ctx context.Context, delete *DeleteUser) error {
	if delete.Username == "admin" {
		return errors.New("cannot delete the system administrator")
	}
	return s.driver.DeleteUser(ctx, delete)
}

func (s *Store) UpdateUserPassword(ctx context.Context, username domain.Username, newHash string) error {
	// Relay the command to the driver (the stove)
	return s.driver.UpdateUserPassword(ctx, username, newHash)
}

func (s *Store) ArchiveUser(ctx context.Context, username domain.Username) error {
	return s.driver.ArchiveUser(ctx, username)
}
