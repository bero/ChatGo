// Package api - user management handlers (admin only)
package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"chatgo/internal/auth"
	"chatgo/internal/db"
	"chatgo/internal/models"
)

// CreateUserHandler handles POST /api/users (admin only)
func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Parse the request body.
	var req models.UserCreateRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
		return
	}

	// Validate input.
	if req.Username == "" || req.Password == "" {
		http.Error(w, `{"error": "Username and password required"}`, http.StatusBadRequest)
		return
	}

	// Check if username already exists.
	existingUser, err := db.GetUserByUsername(req.Username)
	if err != nil {
		http.Error(w, `{"error": "Database error"}`, http.StatusInternalServerError)
		return
	}
	if existingUser != nil {
		http.Error(w, `{"error": "Username already taken"}`, http.StatusConflict)
		return
	}

	// Hash the password.
	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		http.Error(w, `{"error": "Failed to hash password"}`, http.StatusInternalServerError)
		return
	}

	// Create the user.
	user, err := db.CreateUser(req.Username, passwordHash, req.IsAdmin)
	if err != nil {
		http.Error(w, `{"error": "Failed to create user"}`, http.StatusInternalServerError)
		return
	}

	// Return the created user (without password hash).
	json.NewEncoder(w).Encode(user.ToResponse())
}

// DeleteUserHandler handles DELETE /api/users/{id} (admin only)
func DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract user ID from URL path.
	// URL format: /api/users/{id}
	path := r.URL.Path
	parts := strings.Split(path, "/")

	// Expected: ["", "api", "users", "{id}"]
	if len(parts) != 4 {
		http.Error(w, `{"error": "Invalid URL format"}`, http.StatusBadRequest)
		return
	}

	userID := parts[3]
	if userID == "" {
		http.Error(w, `{"error": "User ID required"}`, http.StatusBadRequest)
		return
	}

	// Get current user from context (set by middleware).
	currentUser := GetUserFromContext(r)
	if currentUser != nil && currentUser.UserID == userID {
		http.Error(w, `{"error": "Cannot delete yourself"}`, http.StatusBadRequest)
		return
	}

	// Delete the user.
	deleted, err := db.DeleteUser(userID)
	if err != nil {
		http.Error(w, `{"error": "Failed to delete user"}`, http.StatusInternalServerError)
		return
	}

	if !deleted {
		http.Error(w, `{"error": "User not found"}`, http.StatusNotFound)
		return
	}

	// Return success message.
	json.NewEncoder(w).Encode(map[string]string{
		"message": "User deleted successfully",
	})
}

// UpdateUserHandler handles PUT /api/users/{id} (admin only)
func UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract user ID from URL path.
	path := r.URL.Path
	parts := strings.Split(path, "/")

	if len(parts) != 4 {
		http.Error(w, `{"error": "Invalid URL format"}`, http.StatusBadRequest)
		return
	}

	userID := parts[3]
	if userID == "" {
		http.Error(w, `{"error": "User ID required"}`, http.StatusBadRequest)
		return
	}

	// Parse request body.
	var req models.UserUpdateRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
		return
	}

	// Validate username is not empty.
	if req.Username == "" {
		http.Error(w, `{"error": "Username required"}`, http.StatusBadRequest)
		return
	}

	// Check if username is taken by another user.
	existingUser, err := db.GetUserByUsername(req.Username)
	if err != nil {
		http.Error(w, `{"error": "Database error"}`, http.StatusInternalServerError)
		return
	}
	if existingUser != nil && existingUser.ID != userID {
		http.Error(w, `{"error": "Username already taken"}`, http.StatusConflict)
		return
	}

	// Hash new password if provided.
	var passwordHash string
	if req.Password != "" {
		passwordHash, err = auth.HashPassword(req.Password)
		if err != nil {
			http.Error(w, `{"error": "Failed to hash password"}`, http.StatusInternalServerError)
			return
		}
	}

	// Update the user.
	user, err := db.UpdateUser(userID, req.Username, passwordHash, req.IsAdmin)
	if err != nil {
		http.Error(w, `{"error": "Failed to update user"}`, http.StatusInternalServerError)
		return
	}

	if user == nil {
		http.Error(w, `{"error": "User not found"}`, http.StatusNotFound)
		return
	}

	// Return the updated user.
	json.NewEncoder(w).Encode(user.ToResponse())
}
