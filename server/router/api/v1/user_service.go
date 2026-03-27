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

	// Now we use userCtx.ID directly
	user, err := s.Store.GetUser(c.Request().Context(), &store.FindUser{ID: &userCtx.ID})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, user)
}

func (s *APIV1Service) ChangePasswordHandler(c *echo.Context) error {
	// 1. Define what we expect from React
	var req struct {
		OldPassword string `json:"oldPassword"`
		NewPassword string `json:"newPassword"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid request format")
	}

	// 2. Get the UserID from the context (the "Badge" we fixed)
	userID, ok := c.Request().Context().Value("user-id").(int32)
	if !ok {
		return c.JSON(http.StatusUnauthorized, "Not logged in")
	}

	// 3. Fetch the current user to get their CURRENT hash
	user, err := s.Store.GetUser(c.Request().Context(), &store.FindUser{ID: &userID})
	if err != nil {
		return c.JSON(http.StatusNotFound, "User not found")
	}

	// 4. Verify: Does the 'OldPassword' match the one in the DB?
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.OldPassword))
	if err != nil {
		return c.JSON(http.StatusUnauthorized, "Incorrect old password")
	}

	// 5. Hash the NEW password
	newHash, _ := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)

	// 6. Tell the Store to save it
	if err := s.Store.ChangeUserPassword(c.Request().Context(), userID, string(newHash)); err != nil {
		return c.JSON(http.StatusInternalServerError, "Failed to save new password")
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
