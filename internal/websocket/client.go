// Package websocket - client connection handling
package websocket

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"chatgo/internal/db"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 4096
)

// Client represents a single WebSocket connection.
type Client struct {
	hub *Hub

	// The WebSocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte

	// User information (from JWT token).
	UserID   string
	Username string

	// closeOnce ensures we only close the send channel once.
	closeOnce sync.Once
}

// IncomingMessage is the format of messages from the client.
type IncomingMessage struct {
	Type           string `json:"type"`            // "message" or "typing"
	ConversationID string `json:"conversation_id"` // Target conversation
	Content        string `json:"content"`         // Message content (for "message" type)
	IsTyping       bool   `json:"is_typing"`       // Typing status (for "typing" type)
}

// ChatMessage is sent when a new message is created.
type ChatMessage struct {
	Type           string `json:"type"` // "message"
	ID             string `json:"id"`
	ConversationID string `json:"conversation_id"`
	SenderID       string `json:"sender_id"`
	SenderUsername string `json:"sender_username"`
	Content        string `json:"content"`
	CreatedAt      string `json:"created_at"`
}

// TypingMessage is sent when a user starts/stops typing.
type TypingMessage struct {
	Type           string `json:"type"` // "typing"
	ConversationID string `json:"conversation_id"`
	UserID         string `json:"user_id"`
	Username       string `json:"username"`
	IsTyping       bool   `json:"is_typing"`
}

// NewClient creates a new client instance.
func NewClient(hub *Hub, conn *websocket.Conn, userID, username string) *Client {
	return &Client{
		hub:      hub,
		conn:     conn,
		send:     make(chan []byte, 256),
		UserID:   userID,
		Username: username,
	}
}

// Close safely closes the client's send channel (only once).
func (c *Client) Close() {
	c.closeOnce.Do(func() {
		log.Printf("Closing send channel for client: %s", c.UserID)
		close(c.send)
	})
}

// ReadPump pumps messages from the WebSocket connection to the hub.
// Runs in its own goroutine.
func (c *Client) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Parse the incoming message.
		var msg IncomingMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			log.Printf("Invalid message format: %v", err)
			continue
		}

		// Handle the message based on type.
		switch msg.Type {
		case "message":
			c.handleChatMessage(msg)
		case "typing":
			c.handleTypingMessage(msg)
		default:
			log.Printf("Unknown message type: %s", msg.Type)
		}
	}
}

// WritePump pumps messages from the hub to the WebSocket connection.
// Runs in its own goroutine.
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleChatMessage processes an incoming chat message.
func (c *Client) handleChatMessage(msg IncomingMessage) {
	// Verify user is in this conversation.
	isParticipant, err := db.IsUserInConversation(c.UserID, msg.ConversationID)
	if err != nil || !isParticipant {
		log.Printf("User %s not in conversation %s", c.UserID, msg.ConversationID)
		return
	}

	// Save message to database.
	savedMsg, err := db.CreateMessage(msg.ConversationID, c.UserID, msg.Content)
	if err != nil {
		log.Printf("Failed to save message: %v", err)
		return
	}

	// Create the outgoing message.
	chatMsg := ChatMessage{
		Type:           "message",
		ID:             savedMsg.ID,
		ConversationID: savedMsg.ConversationID,
		SenderID:       savedMsg.SenderID,
		SenderUsername: c.Username,
		Content:        savedMsg.Content,
		CreatedAt:      savedMsg.CreatedAt.Format(time.RFC3339),
	}

	// Send to all participants in the conversation.
	c.sendToConversationParticipants(msg.ConversationID, chatMsg)
}

// handleTypingMessage processes a typing indicator.
func (c *Client) handleTypingMessage(msg IncomingMessage) {
	// Verify user is in this conversation.
	isParticipant, err := db.IsUserInConversation(c.UserID, msg.ConversationID)
	if err != nil || !isParticipant {
		return
	}

	// Create typing notification.
	typingMsg := TypingMessage{
		Type:           "typing",
		ConversationID: msg.ConversationID,
		UserID:         c.UserID,
		Username:       c.Username,
		IsTyping:       msg.IsTyping,
	}

	// Send to all other participants.
	c.sendToConversationParticipants(msg.ConversationID, typingMsg)
}

// sendToConversationParticipants sends a message to all users in a conversation.
func (c *Client) sendToConversationParticipants(conversationID string, message interface{}) {
	// Get all conversations for this conversation to find participants.
	// This is a simple approach - in production you might cache this.
	conversations, err := db.GetUserConversations(c.UserID)
	if err != nil {
		log.Printf("Failed to get conversations: %v", err)
		return
	}

	// Find the other user in this conversation and send to them.
	for _, conv := range conversations {
		if conv.ID == conversationID {
			// Send to the other user.
			c.hub.SendToUser(conv.OtherUserID, message)
			break
		}
	}

	// Also send to self (so message appears in sender's chat).
	c.hub.SendToUser(c.UserID, message)
}
