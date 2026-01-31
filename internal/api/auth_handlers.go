// Package api - authentication handlers
package api

import (
	"encoding/json"
	"net/http"

	"chatgo/internal/auth"
	"chatgo/internal/db"
)

// LoginRequest is the expected JSON body for login.
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse is what we send back after successful login.
type LoginResponse struct {
	Token    string `json:"token"`
	Username string `json:"username"`
	IsAdmin  bool   `json:"is_admin"`
}

// LoginHandler handles POST /api/login
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Only accept POST requests.
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	// Parse the JSON body.
	var req LoginRequest
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

	// Find the user in the database.
	user, err := db.GetUserByUsername(req.Username)
	if err != nil {
		http.Error(w, `{"error": "Database error"}`, http.StatusInternalServerError)
		return
	}
	if user == nil {
		// User not found - but don't reveal this! Say "invalid credentials" instead.
		http.Error(w, `{"error": "Invalid credentials"}`, http.StatusUnauthorized)
		return
	}

	// Check the password.
	if !auth.CheckPassword(req.Password, user.PasswordHash) {
		http.Error(w, `{"error": "Invalid credentials"}`, http.StatusUnauthorized)
		return
	}

	// Generate a JWT token.
	token, err := auth.GenerateToken(user.ID, user.Username, user.IsAdmin)
	if err != nil {
		http.Error(w, `{"error": "Failed to generate token"}`, http.StatusInternalServerError)
		return
	}

	// Send the response.
	response := LoginResponse{
		Token:    token,
		Username: user.Username,
		IsAdmin:  user.IsAdmin,
	}

	json.NewEncoder(w).Encode(response)
}
