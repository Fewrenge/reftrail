package v1

import (
	"errors"
	"log/slog"
	"net/http"
	"reftrail/internal/domain"
	"reftrail/store"

	echo "github.com/labstack/echo/v5"
	"golang.org/x/crypto/bcrypt"
)

// A mix of admin's user management and regular user self-service endpoints. Could be split into separate files if it gets too big

// CreateUserHandler handles POST /api/v1/users
func (s *APIV1Service) CreateUserHandler(c *echo.Context) error {
	ctx := c.Request().Context()
	create := &store.CreateUser{}

	if err := c.Bind(create); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request format"})
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(create.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to hash password"})
	}
	create.Password = string(hashedPassword)

	user, err := s.Store.CreateUser(ctx, create)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create user"})
	}

	return c.JSON(http.StatusOK, user)
}

// GET /api/v1/users/me
func (s *APIV1Service) GetCurrentUserHandler(c *echo.Context) error {
	ctx, ok := domain.GetUserContext(c.Request().Context())
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "User context not found"})
	}

	user, err := s.Store.GetUser(c.Request().Context(), &store.FindUser{Username: ctx.Username})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get current user"})
	}

	if user == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "User profile not found"})
	}

	return c.JSON(http.StatusOK, user)
}

// PATCH /api/v1/users/me/password
func (s *APIV1Service) ChangeOwnPasswordHandler(c *echo.Context) error {
	ctx := c.Request().Context()

	userCtx, ok := domain.GetUserContext(ctx)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "User context not found"})
	}

	var req struct {
		OldPassword string `json:"oldPassword"`
		NewPassword string `json:"newPassword"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request format"})
	}

	// Delegate all business logic and crypto to the Store layer
	err := s.Store.ChangeOwnPassword(ctx, userCtx.Username, req.OldPassword, req.NewPassword)
	if err != nil {
		// Switch case to catch specific domain errors
		switch {
		case errors.Is(err, domain.ErrUserNotFound):
			return c.JSON(http.StatusNotFound, map[string]string{"error": "User not found"})
		case errors.Is(err, domain.ErrPasswordMismatch):
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Incorrect old password"})
		default:
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Database update failed"})
		}
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Password updated successfully"})
}

// GET /api/v1/users
func (s *APIV1Service) ListUsersHandler(c *echo.Context) error {
	ctx := c.Request().Context()

	// Passing an empty FindUser gets everyone
	users, err := s.Store.ListUsers(ctx, &store.FindUser{})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to list users"})
	}
	return c.JSON(http.StatusOK, users)
}

// PATCH /api/v1/users/:username
func (s *APIV1Service) UpdateUserHandler(c *echo.Context) error {
	ctx := c.Request().Context()
	usernameParam := c.Param("username")
	if usernameParam == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Missing username parameter"})
	}
	username := domain.Username(usernameParam)

	// Dynamic inbound request form maps pointer options
	var req struct {
		UpdatedUsername *domain.Username `json:"updatedUsername"`
		UserFirstName   *string          `json:"userFirstName"`
		UserLastName    *string          `json:"userLastName"`
		// Password        *string          `json:"password"`
		// Role            *domain.UserRole `json:"role"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request format"})
	}

	// Hash password automatically if the administrator provided a fresh value
	/*
		if req.Password != nil && *req.Password != "" {
			hashed, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to hash password"})
			}
			strHash := string(hashed)
			req.Password = &strHash
		}*/

	updatePayload := &store.UpdateUserInfo{
		CurrentUsername: username,
		UpdatedUsername: req.UpdatedUsername,
		UserFirstName:   req.UserFirstName,
		UserLastName:    req.UserLastName,
		// Password:        req.Password,
		// Role:            req.Role,
	}

	updatedUser, err := s.Store.UpdateUserInfo(ctx, updatePayload)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, updatedUser)
}

// DELETE /api/v1/users/:username
func (s *APIV1Service) DeleteUserHandler(c *echo.Context) error {
	ctx := c.Request().Context()
	usernameParam := c.Param("username")
	if usernameParam == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Missing username parameter"})
	}

	username := domain.Username(usernameParam)

	err := s.Store.DeleteUser(ctx, &store.DeleteUser{Username: username})
	if err != nil {
		slog.Error("Failed to delete user from database",
			"username", usernameParam,
			"error", err.Error(),
		)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "User deleted"})
}

// Hard-reset
// PATCH /api/v1/users/:username/password
func (s *APIV1Service) ResetUserPasswordHandler(c *echo.Context) error {
	ctx := c.Request().Context()
	usernameParam := c.Param("username")

	if usernameParam == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Missing username parameter"})
	}

	username := domain.Username(usernameParam)

	// Fetch the acting admin from context
	userCtx, ok := domain.GetUserContext(ctx)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "User context not found"})
	}

	// CRITICAL: Prevent self-reset via the admin route
	if userCtx.Username == username {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Please use the self-service route to change your own password"})
	}

	var req struct {
		Password string `json:"password"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request format"})
	}

	// Hash the new password
	err := s.Store.ResetUserPassword(ctx, userCtx.Username, domain.Username(usernameParam), req.Password)
	if err != nil {
		// Map domain errors to explicit HTTP statuses
		if errors.Is(err, domain.ErrSelfResetBlocked) {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		if errors.Is(err, domain.ErrUserNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to reset password"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "User password reset successfully"})
}

// PUT /api/v1/users/:username/archive
func (s *APIV1Service) ArchiveUserHandler(c *echo.Context) error {
	ctx := c.Request().Context()
	usernameParam := c.Param("username")

	// Input Validation
	if usernameParam == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Missing username parameter"})
	}

	username := domain.Username(usernameParam)

	// Security Check: Protect master root account
	// TODO: implement a more flexible role-based access control system and use that here instead of hardcoding "admin"
	if usernameParam == "admin" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Cannot archive the master admin account"})
	}

	if err := s.Store.ArchiveUser(ctx, username); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to archive user"})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message":  "User archived successfully",
		"username": usernameParam,
	})
}
