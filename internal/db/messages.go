// Package db - message database operations
package db

import (
	"fmt"

	"chatgo/internal/models"
)

// CreateMessage inserts a new message into the database.
func CreateMessage(conversationID, senderID, content string) (*models.Message, error) {
	query := `
		INSERT INTO messages (conversation_id, sender_id, content)
		VALUES ($1, $2, $3)
		RETURNING id, conversation_id, sender_id, content, created_at
	`

	var msg models.Message
	err := DB.QueryRow(query, conversationID, senderID, content).Scan(
		&msg.ID,
		&msg.ConversationID,
		&msg.SenderID,
		&msg.Content,
		&msg.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	return &msg, nil
}

// GetConversationMessages returns all messages in a conversation.
// Includes the sender's username for display purposes.
func GetConversationMessages(conversationID string, limit int) ([]models.Message, error) {
	query := `
		SELECT m.id, m.conversation_id, m.sender_id, u.username, m.content, m.created_at
		FROM messages m
		JOIN users u ON m.sender_id = u.id
		WHERE m.conversation_id = $1
		ORDER BY m.created_at ASC
		LIMIT $2
	`

	rows, err := DB.Query(query, conversationID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query messages: %w", err)
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var msg models.Message
		err := rows.Scan(
			&msg.ID,
			&msg.ConversationID,
			&msg.SenderID,
			&msg.SenderUsername,
			&msg.Content,
			&msg.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		messages = append(messages, msg)
	}

	return messages, nil
}
