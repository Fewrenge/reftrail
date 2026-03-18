package v1

import (
	"net/http"
	"wl/store"

	echo "github.com/labstack/echo/v5"
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

// GetCurrentUserHandler handles GET /api/v1/users/me
// (Useful for the frontend to check its own "Badge")
func (s *APIV1Service) GetCurrentUserHandler(c *echo.Context) error {
	// We'll use our GetUserID helper we wrote in auth/context.go
	// But since store shouldn't import auth, we look at the context key directly
	userID, ok := c.Request().Context().Value("user-id").(int32)
	if !ok {
		return c.JSON(http.StatusUnauthorized, "Not logged in")
	}

	user, err := s.Store.GetUser(c.Request().Context(), &store.FindUser{ID: &userID})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, user)
}
