// Package models - conversation and message data structures
package models

import "time"

// Conversation represents a chat between users.
type Conversation struct {
	ID        string    `json:"id"`
	Name      string    `json:"name,omitempty"` // Optional name for group chats
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

// Participant represents a user in a conversation.
type Participant struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

// ConversationWithParticipants includes all participants in the conversation.
// Used when listing conversations for a user.
type ConversationWithParticipants struct {
	ID           string        `json:"id"`
	Name         string        `json:"name,omitempty"`
	IsGroup      bool          `json:"is_group"`
	Participants []Participant `json:"participants"`
	CreatedAt    time.Time     `json:"created_at"`
}
