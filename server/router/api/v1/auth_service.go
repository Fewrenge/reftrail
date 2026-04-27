package v1

import (
	"net/http"
	"reftrail/server/auth" // Import your new auth package
	"reftrail/store"
	"time"

	echo "github.com/labstack/echo/v5"
)

func (s *APIV1Service) LoginHandler(c *echo.Context) error {
	req := &store.LoginRequest{}
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request format"})
	}

	user, err := s.Store.Login(c.Request().Context(), req)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
	}

	// USE THE NEW AUTH PACKAGE TO GENERATE TOKEN
	token, err := auth.GenerateToken(user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate authentication token"})
	}

	cookie := &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,                 // JavaScript can't read it
		Secure:   false,                // Set to TRUE in production (HTTPS)
		SameSite: http.SameSiteLaxMode, // Prevents some CSRF attacks
		MaxAge:   3600 * 24 * 3,        // 3 days
	}
	c.SetCookie(cookie) // Send the cookie header
	return c.JSON(http.StatusOK, map[string]string{"message": "Welcome!"})
}

func (s *APIV1Service) LogoutHandler(c *echo.Context) error {
	cookie := &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0), // Expire it immediately
		HttpOnly: true,
		MaxAge:   -1, // Alternative way to tell browser to delete it
	}
	c.SetCookie(cookie)
	return c.JSON(http.StatusOK, map[string]string{"message": "Logged out"})
}
