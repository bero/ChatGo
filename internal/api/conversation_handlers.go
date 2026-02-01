// Package api - conversation handlers
package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"chatgo/internal/db"
	"chatgo/internal/models"
	"chatgo/internal/websocket"
)

// CreateConversationRequest is the request body for creating/getting a conversation.
type CreateConversationRequest struct {
	OtherUserID    string   `json:"other_user_id,omitempty"`    // For 1:1 chat
	ParticipantIDs []string `json:"participant_ids,omitempty"`  // For group chat
	Name           string   `json:"name,omitempty"`             // Group name (required for groups)
}

// CreateConversationHandler handles POST /api/conversations
// Gets or creates a conversation between users (1:1 or group).
func CreateConversationHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get current user from context
	user := GetUserFromContext(r)
	if user == nil {
		http.Error(w, `{"error": "User not authenticated"}`, http.StatusUnauthorized)
		return
	}

	// Parse request
	var req CreateConversationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
		return
	}

	// Determine if this is a group or 1:1 conversation
	if len(req.ParticipantIDs) > 0 {
		// Group conversation
		if req.Name == "" {
			http.Error(w, `{"error": "name required for group conversations"}`, http.StatusBadRequest)
			return
		}

		// Ensure current user is included in participant list
		participantSet := make(map[string]bool)
		participantSet[user.UserID] = true
		for _, id := range req.ParticipantIDs {
			participantSet[id] = true
		}

		// Convert set to slice
		participants := make([]string, 0, len(participantSet))
		for id := range participantSet {
			participants = append(participants, id)
		}

		if len(participants) < 2 {
			http.Error(w, `{"error": "group requires at least 2 participants"}`, http.StatusBadRequest)
			return
		}

		conversation, err := db.CreateGroupConversation(req.Name, participants)
		if err != nil {
			http.Error(w, `{"error": "Failed to create group conversation"}`, http.StatusInternalServerError)
			return
		}

		// Notify all participants about the new conversation
		websocket.NotifyNewConversation(conversation.ID, participants)

		json.NewEncoder(w).Encode(conversation)
		return
	}

	// 1:1 conversation
	if req.OtherUserID == "" {
		http.Error(w, `{"error": "other_user_id or participant_ids required"}`, http.StatusBadRequest)
		return
	}

	// Get or create the conversation
	conversation, err := db.GetOrCreateConversation(user.UserID, req.OtherUserID)
	if err != nil {
		http.Error(w, `{"error": "Failed to create conversation"}`, http.StatusInternalServerError)
		return
	}

	// Notify both users about the conversation (harmless if it already existed)
	websocket.NotifyNewConversation(conversation.ID, []string{user.UserID, req.OtherUserID})

	json.NewEncoder(w).Encode(conversation)
}

// GetConversationsHandler handles GET /api/conversations
// Returns all conversations for the current user.
func GetConversationsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get current user from context
	user := GetUserFromContext(r)
	if user == nil {
		http.Error(w, `{"error": "User not authenticated"}`, http.StatusUnauthorized)
		return
	}

	// Get user's conversations
	conversations, err := db.GetUserConversations(user.UserID)
	if err != nil {
		http.Error(w, `{"error": "Failed to get conversations"}`, http.StatusInternalServerError)
		return
	}

	// Return empty array instead of null
	if conversations == nil {
		conversations = []models.ConversationWithParticipants{}
	}

	json.NewEncoder(w).Encode(conversations)
}

// GetMessagesHandler handles GET /api/conversations/{id}/messages
func GetMessagesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get current user from context
	user := GetUserFromContext(r)
	if user == nil {
		http.Error(w, `{"error": "User not authenticated"}`, http.StatusUnauthorized)
		return
	}

	// Extract conversation ID from URL
	// URL: /api/conversations/{id}/messages
	path := r.URL.Path
	parts := strings.Split(path, "/")

	// Expected: ["", "api", "conversations", "{id}", "messages"]
	if len(parts) < 5 {
		http.Error(w, `{"error": "Invalid URL"}`, http.StatusBadRequest)
		return
	}

	conversationID := parts[3]

	// Verify user is in this conversation
	isParticipant, err := db.IsUserInConversation(user.UserID, conversationID)
	if err != nil {
		http.Error(w, `{"error": "Database error"}`, http.StatusInternalServerError)
		return
	}
	if !isParticipant {
		http.Error(w, `{"error": "Not authorized"}`, http.StatusForbidden)
		return
	}

	// Get messages (limit to 100)
	messages, err := db.GetConversationMessages(conversationID, 100)
	if err != nil {
		http.Error(w, `{"error": "Failed to get messages"}`, http.StatusInternalServerError)
		return
	}

	// Return empty array instead of null
	if messages == nil {
		messages = []models.Message{}
	}

	json.NewEncoder(w).Encode(messages)
}
