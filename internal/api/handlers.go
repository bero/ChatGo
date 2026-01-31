// Package api contains HTTP handlers for the ChatGO API.
package api

import (
	"encoding/json"
	"net/http"

	"chatgo/internal/db"
	"chatgo/internal/models"
)

// HomeHandler handles requests to the root path "/".
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello from ChatGO!"))
}

// HealthHandler returns the server health status as JSON.
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := map[string]string{
		"status":  "ok",
		"message": "ChatGO is running",
	}

	json.NewEncoder(w).Encode(response)
}

// ListUsersHandler returns all users from the database.
// This is a real endpoint that queries the database!
func ListUsersHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get all users from the database.
	users, err := db.GetAllUsers()
	if err != nil {
		// Return an error response.
		// http.StatusInternalServerError = 500
		http.Error(w, `{"error": "Failed to get users"}`, http.StatusInternalServerError)
		return
	}

	// Convert each user to a safe response (without password hash).
	var responses []models.UserResponse
	for _, user := range users {
		responses = append(responses, user.ToResponse())
	}

	json.NewEncoder(w).Encode(responses)
}
