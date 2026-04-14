package auth

import (
	"log"
	"os"
	"time"
	"wl/internal/types"
	"wl/store"

	"github.com/joho/godotenv"

	"github.com/golang-jwt/jwt/v5"
)

func getSecret() []byte {
	// Load .env file if it exists (for local development)
	_ = godotenv.Load()

	// Get the secret from the environment
	secret := os.Getenv("JWT_SECRET")

	// Fallback for safety (though you should ideally fail if it's missing)
	if secret == "" {
		log.Fatal("CRITICAL ERROR: JWT_SECRET environment variable is not set.")
	}

	return []byte(secret)
}

type Claims struct {
	ID   types.UserID   `json:"id"`
	Role types.UserRole `json:"role"`
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
	return token.SignedString(getSecret())
}
