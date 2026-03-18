package auth

import (
	"time"
	"wl/store"

	"github.com/golang-jwt/jwt/v5"
)

const (
	// In a real app, move this to an Environment Variable!
	SecretKey = "your-secret-key-123"
)

type Claims struct {
	ID   int32  `json:"id"`
	Role string `json:"role"`
	jwt.RegisteredClaims
}

func GenerateToken(user *store.User) (string, error) {
	claims := &Claims{
		ID:   user.ID,
		Role: user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(72 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(SecretKey))
}
