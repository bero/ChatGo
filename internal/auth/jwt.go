// Package auth - JWT token handling
package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTSecret is the key used to sign tokens.
// In production, this should come from an environment variable!
var JWTSecret = []byte("your-secret-key-change-in-production")

// Claims contains the data we store in the JWT token.
// jwt.RegisteredClaims includes standard fields like expiration time.
type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	IsAdmin  bool   `json:"is_admin"`
	jwt.RegisteredClaims
}

// GenerateToken creates a new JWT token for a user.
// The token expires after 24 hours.
func GenerateToken(userID, username string, isAdmin bool) (string, error) {
	// Set expiration time to 24 hours from now.
	expirationTime := time.Now().Add(24 * time.Hour)

	// Create the claims (the data inside the token).
	claims := &Claims{
		UserID:   userID,
		Username: username,
		IsAdmin:  isAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	// Create the token with HS256 signing method.
	// HS256 = HMAC with SHA-256 (symmetric encryption).
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with our secret key.
	tokenString, err := token.SignedString(JWTSecret)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// ValidateToken checks if a token is valid and returns the claims.
// Returns nil and error if the token is invalid or expired.
func ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	// Parse the token and validate the signature.
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Verify the signing method is what we expect.
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return JWTSecret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("token is not valid")
	}

	return claims, nil
}
