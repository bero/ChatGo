// Package db - conversation database operations
package db

import (
	"database/sql"
	"fmt"

	"chatgo/internal/models"
)

// GetOrCreateConversation finds an existing conversation between two users,
// or creates a new one if it doesn't exist.
func GetOrCreateConversation(userID1, userID2 string) (*models.Conversation, error) {
	// First, try to find an existing conversation between these two users.
	query := `
		SELECT c.id, c.created_at
		FROM conversations c
		JOIN conversation_participants cp1 ON c.id = cp1.conversation_id
		JOIN conversation_participants cp2 ON c.id = cp2.conversation_id
		WHERE cp1.user_id = $1 AND cp2.user_id = $2
		LIMIT 1
	`

	var conv models.Conversation
	err := DB.QueryRow(query, userID1, userID2).Scan(&conv.ID, &conv.CreatedAt)

	if err == nil {
		// Found existing conversation
		return &conv, nil
	}

	if err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to find conversation: %w", err)
	}

	// No existing conversation - create a new one.
	// Use a transaction to ensure both inserts succeed or fail together.
	tx, err := DB.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // Rollback if we don't commit

	// Create the conversation
	err = tx.QueryRow(
		`INSERT INTO conversations DEFAULT VALUES RETURNING id, created_at`,
	).Scan(&conv.ID, &conv.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create conversation: %w", err)
	}

	// Add both users as participants
	_, err = tx.Exec(
		`INSERT INTO conversation_participants (conversation_id, user_id) VALUES ($1, $2), ($1, $3)`,
		conv.ID, userID1, userID2,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to add participants: %w", err)
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &conv, nil
}

// GetUserConversations returns all conversations for a user.
func GetUserConversations(userID string) ([]models.ConversationWithParticipants, error) {
	query := `
		SELECT c.id, c.created_at, u.id, u.username
		FROM conversations c
		JOIN conversation_participants cp1 ON c.id = cp1.conversation_id
		JOIN conversation_participants cp2 ON c.id = cp2.conversation_id
		JOIN users u ON cp2.user_id = u.id
		WHERE cp1.user_id = $1 AND cp2.user_id != $1
		ORDER BY c.created_at DESC
	`

	rows, err := DB.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query conversations: %w", err)
	}
	defer rows.Close()

	var conversations []models.ConversationWithParticipants
	for rows.Next() {
		var conv models.ConversationWithParticipants
		err := rows.Scan(&conv.ID, &conv.CreatedAt, &conv.OtherUserID, &conv.OtherUsername)
		if err != nil {
			return nil, fmt.Errorf("failed to scan conversation: %w", err)
		}
		conversations = append(conversations, conv)
	}

	return conversations, nil
}

// IsUserInConversation checks if a user is a participant in a conversation.
func IsUserInConversation(userID, conversationID string) (bool, error) {
	query := `SELECT 1 FROM conversation_participants WHERE user_id = $1 AND conversation_id = $2`

	var exists int
	err := DB.QueryRow(query, userID, conversationID).Scan(&exists)

	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check participation: %w", err)
	}

	return true, nil
}
