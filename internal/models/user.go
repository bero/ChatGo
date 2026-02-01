// Package models contains data structures used throughout the application.
package models

import "time"

// User represents a user in the chat system.
// struct is Go's way to define a custom data type with multiple fields.
type User struct {
	// Each field has a name, type, and an optional "tag" (the `json:"..."` part).
	// Tags tell the JSON encoder what name to use when converting to/from JSON.

	ID           string    `json:"id"`         // Unique identifier
	Username     string    `json:"username"`   // Display name / login name
	PasswordHash string    `json:"-"`          // "-" means: never include in JSON output (security!)
	IsAdmin      bool      `json:"is_admin"`   // Can this user manage other users?
	CreatedAt    time.Time `json:"created_at"` // When the user was created
}

// UserCreateRequest is the data needed to create a new user.
// We use separate structs for requests to control what fields are accepted.
type UserCreateRequest struct {
	Username string `json:"username"`
	Password string `json:"password"` // Plain password - we'll hash it before storing
	IsAdmin  bool   `json:"is_admin"`
}

// UserUpdateRequest is the data for updating a user.
// Password is optional - empty string means don't change it.
type UserUpdateRequest struct {
	Username string `json:"username"`
	Password string `json:"password"` // Optional: empty = keep current password
	IsAdmin  bool   `json:"is_admin"`
}

// UserResponse is what we send back to the client.
// Notice: no password field at all - we never send passwords back.
type UserResponse struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	IsAdmin   bool      `json:"is_admin"`
	CreatedAt time.Time `json:"created_at"`
}

// ToResponse converts a User to a UserResponse.
// This is a "method" - a function attached to a type.
// (u User) means this method can be called on any User value.
func (u User) ToResponse() UserResponse {
	return UserResponse{
		ID:        u.ID,
		Username:  u.Username,
		IsAdmin:   u.IsAdmin,
		CreatedAt: u.CreatedAt,
	}
}
