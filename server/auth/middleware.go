package auth

import (
	"context"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	echo "github.com/labstack/echo/v5"
)

func JWTMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {

		cookie, err := c.Cookie("auth_token")

		if err != nil {
			// If the cookie isn't there, the user isn't logged in
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Missing auth cookie"})
		}

		// 2. The token string is now inside the cookie value
		tokenString := cookie.Value

		// 3. Parse the token exactly like before
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
			return getSecret(), nil
		})

		// If the token is fake or expired, reject it
		if err != nil || !token.Valid {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid or expired badge"})
		}

		// 4. Pin the UserID and Role to the context memory
		ctx := context.WithValue(c.Request().Context(), "user-id", claims.ID)
		ctx = context.WithValue(ctx, "user-role", claims.Role)
		c.SetRequest(c.Request().WithContext(ctx))

		return next(c)
	}
}
