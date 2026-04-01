package v1

import (
	"fmt"
	"net/http"
	"wl/server/auth"
	"wl/store"

	echo "github.com/labstack/echo/v5"
	"golang.org/x/crypto/bcrypt"
)

// CreateUserHandler handles POST /api/v1/users
func (s *APIV1Service) CreateUserHandler(c *echo.Context) error {
	ctx := c.Request().Context()
	create := &store.CreateUser{}

	if err := c.Bind(create); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	user, err := s.Store.CreateUser(ctx, create)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, user)
}

// GET /api/v1/users/me
func (s *APIV1Service) GetCurrentUserHandler(c *echo.Context) error {
	userCtx, ok := auth.GetUserContext(c.Request().Context())
	if !ok {
		return c.JSON(http.StatusUnauthorized, "Not logged in")
	}

	user, err := s.Store.GetUser(c.Request().Context(), &store.FindUser{ID: &userCtx.ID})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, user)
}

// PATCH /api/v1/users/password
func (s *APIV1Service) ChangePasswordHandler(c *echo.Context) error {
	ctx := c.Request().Context()

	userCtx, ok := auth.GetUserContext(ctx)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "User context not found"})
	}

	userID := userCtx.ID

	var req struct {
		OldPassword string `json:"oldPassword"`
		NewPassword string `json:"newPassword"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request format"})
	}

	user, err := s.Store.GetUser(ctx, &store.FindUser{ID: &userID})
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "User not found"})
	}

	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.OldPassword)); err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Incorrect old password"})
	}

	// Hash new password
	newHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to hash password"})
	}

	// Save
	if err := s.Store.ChangeUserPassword(ctx, userID, string(newHash)); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Database update failed"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Password updated successfully"})
}

// GET /api/v1/users
func (s *APIV1Service) ListUsersHandler(c *echo.Context) error {
	ctx := c.Request().Context()

	// Passing an empty FindUser gets everyone
	users, err := s.Store.ListUsers(ctx, &store.FindUser{})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, users)
}

// DELETE /api/v1/users/:id
func (s *APIV1Service) DeleteUserHandler(c *echo.Context) error {
	ctx := c.Request().Context()

	// Get ID from URL /api/v1/users/5
	idParam := c.Param("id")
	// Convert string "5" to int32
	var id int32
	fmt.Sscanf(idParam, "%d", &id)

	err := s.Store.DeleteUser(ctx, &store.DeleteUser{ID: id})
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "User deleted"})
}
