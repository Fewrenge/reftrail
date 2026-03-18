package v1

import (
	"net/http"
	"wl/server/auth" // Import your new auth package
	"wl/store"

	echo "github.com/labstack/echo/v5"
)

func (s *APIV1Service) LoginHandler(c *echo.Context) error {
	req := &store.LoginRequest{}
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	user, err := s.Store.Login(c.Request().Context(), req)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, err.Error())
	}

	// USE THE NEW AUTH PACKAGE TO GENERATE TOKEN
	token, err := auth.GenerateToken(user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "Token error")
	}

	return c.JSON(http.StatusOK, map[string]any{
		"token": token,
		"user":  user,
	})
}
