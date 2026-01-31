// Package api - middleware for authentication
package api

import (
	"context"
	"net/http"
	"strings"

	"chatgo/internal/auth"
)

// ContextKey is a type for context keys to avoid collisions.
type ContextKey string

const (
	// UserContextKey is the key for storing user claims in the request context.
	UserContextKey ContextKey = "user"
)

// AuthMiddleware checks for a valid JWT token in the Authorization header.
// If valid, it adds the user claims to the request context.
// Usage: wrap your handler with AuthMiddleware(yourHandler)
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Get the Authorization header.
		// Format: "Bearer <token>"
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error": "Authorization header required"}`, http.StatusUnauthorized)
			return
		}

		// Split "Bearer <token>" into parts.
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, `{"error": "Invalid authorization format. Use: Bearer <token>"}`, http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]

		// Validate the token.
		claims, err := auth.ValidateToken(tokenString)
		if err != nil {
			http.Error(w, `{"error": "Invalid or expired token"}`, http.StatusUnauthorized)
			return
		}

		// Add the claims to the request context.
		// This lets the handler access user info via r.Context().
		ctx := context.WithValue(r.Context(), UserContextKey, claims)

		// Call the next handler with the updated context.
		next(w, r.WithContext(ctx))
	}
}

// AdminMiddleware checks that the user is an admin.
// Must be used AFTER AuthMiddleware.
func AdminMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Get user claims from context (set by AuthMiddleware).
		claims, ok := r.Context().Value(UserContextKey).(*auth.Claims)
		if !ok {
			http.Error(w, `{"error": "User not authenticated"}`, http.StatusUnauthorized)
			return
		}

		// Check if user is admin.
		if !claims.IsAdmin {
			http.Error(w, `{"error": "Admin access required"}`, http.StatusForbidden)
			return
		}

		// User is admin, continue to the handler.
		next(w, r)
	}
}

// GetUserFromContext retrieves the user claims from the request context.
// Returns nil if no user is in the context.
func GetUserFromContext(r *http.Request) *auth.Claims {
	claims, ok := r.Context().Value(UserContextKey).(*auth.Claims)
	if !ok {
		return nil
	}
	return claims
}
