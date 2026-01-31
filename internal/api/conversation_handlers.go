// Package api - conversation handlers
package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"chatgo/internal/db"
	"chatgo/internal/models"
)

// CreateConversationRequest is the request body for creating/getting a conversation.
type CreateConversationRequest struct {
	OtherUserID string `json:"other_user_id"`
}

// CreateConversationHandler handles POST /api/conversations
// Gets or creates a conversation between the current user and another user.
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

	if req.OtherUserID == "" {
		http.Error(w, `{"error": "other_user_id required"}`, http.StatusBadRequest)
		return
	}

	// Get or create the conversation
	conversation, err := db.GetOrCreateConversation(user.UserID, req.OtherUserID)
	if err != nil {
		http.Error(w, `{"error": "Failed to create conversation"}`, http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(conversation)
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
