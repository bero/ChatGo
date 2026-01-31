// Package db - user database operations
package db

import (
	"database/sql"
	"fmt"

	"chatgo/internal/models"
)

// GetUserByUsername finds a user by their username.
// Returns the user and nil error if found.
// Returns nil user and nil error if not found.
// Returns nil user and error if something went wrong.
func GetUserByUsername(username string) (*models.User, error) {
	query := `SELECT id, username, password_hash, is_admin, created_at
	          FROM users WHERE username = $1`

	row := DB.QueryRow(query, username)

	var user models.User
	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.IsAdmin,
		&user.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// GetUserByID finds a user by their ID.
func GetUserByID(id string) (*models.User, error) {
	query := `SELECT id, username, password_hash, is_admin, created_at
	          FROM users WHERE id = $1`

	row := DB.QueryRow(query, id)

	var user models.User
	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.IsAdmin,
		&user.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// GetAllUsers returns all users from the database.
func GetAllUsers() ([]models.User, error) {
	query := `SELECT id, username, password_hash, is_admin, created_at
	          FROM users ORDER BY created_at`

	rows, err := DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.PasswordHash,
			&user.IsAdmin,
			&user.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	return users, nil
}

// CreateUser inserts a new user into the database.
// Returns the created user with its generated ID.
func CreateUser(username, passwordHash string, isAdmin bool) (*models.User, error) {
	query := `INSERT INTO users (username, password_hash, is_admin)
	          VALUES ($1, $2, $3)
	          RETURNING id, username, password_hash, is_admin, created_at`

	row := DB.QueryRow(query, username, passwordHash, isAdmin)

	var user models.User
	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.IsAdmin,
		&user.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &user, nil
}

// DeleteUser removes a user from the database.
// Returns true if a user was deleted, false if no user found.
func DeleteUser(id string) (bool, error) {
	query := `DELETE FROM users WHERE id = $1`

	result, err := DB.Exec(query, id)
	if err != nil {
		return false, fmt.Errorf("failed to delete user: %w", err)
	}

	// RowsAffected tells us how many rows were deleted.
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected > 0, nil
}

// UpdateUser updates a user's username, password (optional), and admin status.
// If passwordHash is empty, the password is not changed.
// Returns the updated user, or nil if user not found.
func UpdateUser(id, username, passwordHash string, isAdmin bool) (*models.User, error) {
	var query string
	var row *sql.Row

	if passwordHash == "" {
		// Update without changing password.
		query = `UPDATE users SET username = $1, is_admin = $2
		         WHERE id = $3
		         RETURNING id, username, password_hash, is_admin, created_at`
		row = DB.QueryRow(query, username, isAdmin, id)
	} else {
		// Update including new password.
		query = `UPDATE users SET username = $1, password_hash = $2, is_admin = $3
		         WHERE id = $4
		         RETURNING id, username, password_hash, is_admin, created_at`
		row = DB.QueryRow(query, username, passwordHash, isAdmin, id)
	}

	var user models.User
	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.IsAdmin,
		&user.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // User not found
	}
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return &user, nil
}
