package store

import (
	"context"
	"errors" // Needed for errors.New
)

type User struct {
	ID           int32  `json:"id"`
	Username     string `json:"username"`
	PasswordHash string `json:"-"`
	Role         string `json:"role"`
}

// The "Form" for logging in
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type CreateUser struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type FindUser struct {
	ID       *int32  `json:"id"`
	Username *string `json:"username"`
}

type UpdateUser struct {
	ID       int32   `json:"id"`
	Username *string `json:"username"`
	Password *string `json:"password"`
	Role     *string `json:"role"`
}

type DeleteUser struct {
	ID int32 `json:"id"`
}

// --- THE MANAGER LOGIC ---

func (s *Store) Login(ctx context.Context, req *LoginRequest) (*User, error) {
	// 1. Find the user by username
	user, err := s.GetUser(ctx, &FindUser{Username: &req.Username})
	if err != nil || user == nil {
		return nil, errors.New("invalid username or password")
	}

	// 2. Compare passwords
	// (Note: In a real medical app, we would use bcrypt.CompareHashAndPassword here)
	if req.Password != user.PasswordHash {
		return nil, errors.New("invalid username or password")
	}

	return user, nil
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
