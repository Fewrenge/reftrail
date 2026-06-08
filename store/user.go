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
}

type UpdateUserRole struct {
	TargetUsername domain.Username `json:"targetUsername"`
	Role           domain.UserRole `json:"role"`
}

type DeleteUser struct {
	ActingAdmin domain.Username `json:"actingAdmin"`
	TargetUser  domain.Username `json:"targetUser"`
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

// TODO: implement a better seeding mechanism
func (s *Store) SeedAdminUser(ctx context.Context) error {
	count, err := s.driver.CountUsers(ctx)
	if err != nil {
		return err
	}

	if count == 0 {
		hashed, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)

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
	// Admin cannot delete themselves
	if delete.ActingAdmin == delete.TargetUser {
		return domain.ErrCannotDeleteSelf
	}

	// Fetch current target user to assess safety context
	user, err := s.GetUser(ctx, &FindUser{Username: delete.TargetUser})
	if err != nil {
		return domain.ErrUserNotFound
	}

	// RULE 5: Check structural safety bounds if the target is an admin
	if user.Role == domain.RoleReftrailAdmin {
		activeAdminCount, err := s.driver.CountActiveAdmins(ctx)
		if err != nil {
			return err
		}
		if activeAdminCount <= 1 {
			return domain.ErrLastAdminLockout
		}
	}

	// Hand off to the secure raw driver layer
	return s.driver.DeleteUser(ctx, delete)
}

func (s *Store) ChangeOwnPassword(ctx context.Context, username domain.Username, oldPassword, newPassword string) error {
	// 1. Guard against empty inputs
	if oldPassword == "" || newPassword == "" {
		return errors.New("passwords cannot be empty")
	}

	// 2. Fetch the user state using existing Store capability
	user, err := s.GetUser(ctx, &FindUser{Username: username})
	if err != nil {
		return domain.ErrUserNotFound // Or map your internal not found error
	}

	// 3. Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return domain.ErrPasswordMismatch
	}

	// 4. Hash new password
	newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// 5. Execute via raw driver
	return s.driver.UpdateUserPassword(ctx, username, string(newHash))
}

func (s *Store) ResetUserPassword(ctx context.Context, actingAdmin, targetUser domain.Username, newPassword string) error {
	// RULE: Prevent an admin from using the override path on themselves
	if actingAdmin == targetUser {
		return domain.ErrSelfResetBlocked
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.driver.UpdateUserPassword(ctx, targetUser, string(newHash))
}

func (s *Store) ArchiveUser(ctx context.Context, actingAdmin, targetUser domain.Username) error {
	// Admin cannot archive themselves
	if actingAdmin == targetUser {
		return domain.ErrCannotArchiveSelf
	}

	// Fetch target user to check their role status
	user, err := s.GetUser(ctx, &FindUser{Username: targetUser})
	if err != nil {
		return domain.ErrUserNotFound
	}

	// If target is an admin, ensure they aren't the last active one
	if user.Role == domain.RoleReftrailAdmin {
		activeAdminCount, err := s.CountActiveAdmins(ctx)
		if err != nil {
			return err
		}
		if activeAdminCount <= 1 {
			return domain.ErrLastAdminLockout
		}
	}

	// Delegate to your clean driver layer
	return s.driver.ArchiveUser(ctx, targetUser)
}

func (s *Store) CountActiveAdmins(ctx context.Context) (int, error) {
	return s.driver.CountActiveAdmins(ctx)
}

func (s *Store) UpdateUserRole(ctx context.Context, actingAdmin, targetUser domain.Username, newRole domain.UserRole) error {
	// RULE 1: An admin cannot demote themselves
	if actingAdmin == targetUser && newRole != domain.RoleReftrailAdmin {
		return domain.ErrCannotDemoteSelf
	}

	// Fetch current state of the target user
	user, err := s.GetUser(ctx, &FindUser{Username: targetUser})
	if err != nil {
		return domain.ErrUserNotFound
	}

	// If the target is already in the requested role, do nothing
	if user.Role == newRole {
		return nil
	}

	// RULE 5: If demoting an existing admin, protect against the last-admin lockout
	if user.Role == domain.RoleReftrailAdmin && newRole == domain.RoleBookingTeam {
		activeAdminCount, err := s.driver.CountActiveAdmins(ctx)
		if err != nil {
			return err
		}
		if activeAdminCount <= 1 {
			return domain.ErrLastAdminLockout
		}
	}

	// Delegate the direct SQL command to the driver layer
	return s.driver.UpdateUserRole(ctx, targetUser, newRole)
}
