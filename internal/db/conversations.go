// Package db - conversation database operations
package db

import (
	"database/sql"
	"fmt"
	"strings"

	"chatgo/internal/models"
)

// GetOrCreateConversation finds an existing 1:1 conversation between two users,
// or creates a new one if it doesn't exist.
func GetOrCreateConversation(userID1, userID2 string) (*models.Conversation, error) {
	// First, try to find an existing 1:1 conversation between these two users.
	// A 1:1 conversation has exactly 2 participants and no name.
	query := `
		SELECT c.id, c.created_at
		FROM conversations c
		JOIN conversation_participants cp1 ON c.id = cp1.conversation_id
		JOIN conversation_participants cp2 ON c.id = cp2.conversation_id
		WHERE cp1.user_id = $1 AND cp2.user_id = $2
		AND c.name IS NULL
		AND (SELECT COUNT(*) FROM conversation_participants WHERE conversation_id = c.id) = 2
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

	// Create the conversation (no name for 1:1 chats)
	err = tx.QueryRow(
		`INSERT INTO conversations (name) VALUES (NULL) RETURNING id, created_at`,
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

// CreateGroupConversation creates a new group conversation with the given name and participants.
func CreateGroupConversation(name string, userIDs []string) (*models.Conversation, error) {
	if len(userIDs) < 2 {
		return nil, fmt.Errorf("group conversation requires at least 2 participants")
	}

	tx, err := DB.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var conv models.Conversation
	err = tx.QueryRow(
		`INSERT INTO conversations (name) VALUES ($1) RETURNING id, created_at`,
		name,
	).Scan(&conv.ID, &conv.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create group conversation: %w", err)
	}
	conv.Name = name

	// Build the insert statement for all participants
	valueStrings := make([]string, len(userIDs))
	valueArgs := make([]interface{}, len(userIDs)+1)
	valueArgs[0] = conv.ID
	for i, userID := range userIDs {
		valueStrings[i] = fmt.Sprintf("($1, $%d)", i+2)
		valueArgs[i+1] = userID
	}

	insertQuery := fmt.Sprintf(
		`INSERT INTO conversation_participants (conversation_id, user_id) VALUES %s`,
		strings.Join(valueStrings, ", "),
	)
	_, err = tx.Exec(insertQuery, valueArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to add participants: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &conv, nil
}

// GetConversationParticipants returns all participants in a conversation.
func GetConversationParticipants(conversationID string) ([]models.Participant, error) {
	query := `
		SELECT u.id, u.username
		FROM users u
		JOIN conversation_participants cp ON u.id = cp.user_id
		WHERE cp.conversation_id = $1
	`

	rows, err := DB.Query(query, conversationID)
	if err != nil {
		return nil, fmt.Errorf("failed to query participants: %w", err)
	}
	defer rows.Close()

	var participants []models.Participant
	for rows.Next() {
		var p models.Participant
		if err := rows.Scan(&p.ID, &p.Username); err != nil {
			return nil, fmt.Errorf("failed to scan participant: %w", err)
		}
		participants = append(participants, p)
	}

	return participants, nil
}

// GetUserConversations returns all conversations for a user with full participant lists.
func GetUserConversations(userID string) ([]models.ConversationWithParticipants, error) {
	// First, get all conversations the user is part of
	convQuery := `
		SELECT c.id, COALESCE(c.name, ''), c.created_at,
			(SELECT COUNT(*) FROM conversation_participants WHERE conversation_id = c.id) as participant_count
		FROM conversations c
		JOIN conversation_participants cp ON c.id = cp.conversation_id
		WHERE cp.user_id = $1
		ORDER BY c.created_at DESC
	`

	rows, err := DB.Query(convQuery, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query conversations: %w", err)
	}
	defer rows.Close()

	var conversations []models.ConversationWithParticipants
	for rows.Next() {
		var conv models.ConversationWithParticipants
		var participantCount int
		err := rows.Scan(&conv.ID, &conv.Name, &conv.CreatedAt, &participantCount)
		if err != nil {
			return nil, fmt.Errorf("failed to scan conversation: %w", err)
		}
		// A group has more than 2 participants OR has a name
		conv.IsGroup = participantCount > 2 || conv.Name != ""
		conversations = append(conversations, conv)
	}

	// For each conversation, get participants
	for i := range conversations {
		participants, err := GetConversationParticipants(conversations[i].ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get participants for conversation %s: %w", conversations[i].ID, err)
		}
		conversations[i].Participants = participants
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
