package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	echo "github.com/labstack/echo/v5"
)

func JWTMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error { // Pointer * is required
		authHeader := c.Request().Header.Get("Authorization")

		// 1. Check if the badge exists
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			// If this returns 400, it's because Echo v5 is strict about response types
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Missing badge"})
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// 2. Parse the badge
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
			return []byte(SecretKey), nil
		})

		if err != nil || !token.Valid {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid badge"})
		}

		// 3. Pin the UserID to the context "memory"
		ctx := context.WithValue(c.Request().Context(), "user-id", claims.ID)
		ctx = context.WithValue(ctx, "user-role", claims.Role)
		c.SetRequest(c.Request().WithContext(ctx))

		// 4. IMPORTANT: Pass the pointer 'c' to the next room
		return next(c)
	}
}
