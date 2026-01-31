// Package auth handles authentication: passwords and JWT tokens.
package auth

import (
	"golang.org/x/crypto/bcrypt"
)

// HashPassword takes a plain text password and returns a bcrypt hash.
// The hash is safe to store in the database.
// bcrypt automatically includes a random "salt" to prevent rainbow table attacks.
func HashPassword(password string) (string, error) {
	// bcrypt.DefaultCost = 10 - this controls how slow the hashing is.
	// Slower = more secure against brute force, but uses more CPU.
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// CheckPassword compares a plain text password with a bcrypt hash.
// Returns true if they match, false otherwise.
func CheckPassword(password, hash string) bool {
	// CompareHashAndPassword returns nil if they match, error if not.
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
