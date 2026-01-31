// Package models - conversation and message data structures
package models

import "time"

// Conversation represents a chat between users.
type Conversation struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
}

// Message represents a single chat message.
type Message struct {
	ID             string    `json:"id"`
	ConversationID string    `json:"conversation_id"`
	SenderID       string    `json:"sender_id"`
	SenderUsername string    `json:"sender_username,omitempty"` // Populated when fetching messages
	Content        string    `json:"content"`
	CreatedAt      time.Time `json:"created_at"`
}

// ConversationWithParticipants includes the other user in the conversation.
// Used when listing conversations for a user.
type ConversationWithParticipants struct {
	ID            string    `json:"id"`
	CreatedAt     time.Time `json:"created_at"`
	OtherUserID   string    `json:"other_user_id"`
	OtherUsername string    `json:"other_username"`
}
