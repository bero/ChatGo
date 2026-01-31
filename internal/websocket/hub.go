// Package websocket manages WebSocket connections for real-time chat.
package websocket

import (
	"encoding/json"
	"log"
	"sync"
)

// Hub maintains the set of active clients and broadcasts messages.
type Hub struct {
	// clients maps user ID to their connection.
	// A user can only have one active connection.
	clients map[string]*Client

	// mutex protects the clients map from concurrent access.
	// Go maps are not thread-safe, so we need this.
	mutex sync.RWMutex

	// register channel for new client connections.
	register chan *Client

	// unregister channel for client disconnections.
	unregister chan *Client

	// broadcast channel for messages to send to specific users.
	broadcast chan *OutgoingMessage
}

// OutgoingMessage is a message to send to a specific user.
type OutgoingMessage struct {
	RecipientID string
	Data        []byte
}

// NewHub creates a new Hub instance.
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *OutgoingMessage, 256), // Buffered channel
	}
}

// Run starts the hub's main loop.
// This should be run in a goroutine.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			log.Printf("Register request for: %s (%s)", client.Username, client.UserID)
			h.mutex.Lock()
			// If user already has a connection, close the old one.
			if oldClient, exists := h.clients[client.UserID]; exists {
				log.Printf("Replacing existing client for: %s", client.UserID)
				oldClient.Close() // Use safe Close method
			}
			h.clients[client.UserID] = client
			h.mutex.Unlock()
			log.Printf("Client connected: %s (%s)", client.Username, client.UserID)

		case client := <-h.unregister:
			log.Printf("Unregister request for: %s (%s)", client.Username, client.UserID)
			h.mutex.Lock()
			// Only remove and close if this client is still the active one
			if existingClient, exists := h.clients[client.UserID]; exists && existingClient == client {
				log.Printf("Removing active client: %s", client.UserID)
				delete(h.clients, client.UserID)
				client.Close() // Use safe Close method
				log.Printf("Client disconnected: %s (%s)", client.Username, client.UserID)
			} else {
				log.Printf("Skipping unregister - client already replaced: %s", client.UserID)
			}
			h.mutex.Unlock()

		case message := <-h.broadcast:
			h.mutex.RLock()
			if client, exists := h.clients[message.RecipientID]; exists {
				select {
				case client.send <- message.Data:
					// Message sent successfully
				default:
					// Client's send buffer is full, skip this message
					log.Printf("Failed to send message to %s: buffer full", message.RecipientID)
				}
			}
			h.mutex.RUnlock()
		}
	}
}

// SendToUser sends a message to a specific user by their ID.
func (h *Hub) SendToUser(userID string, message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	h.broadcast <- &OutgoingMessage{
		RecipientID: userID,
		Data:        data,
	}
	return nil
}

// IsUserOnline checks if a user is currently connected.
func (h *Hub) IsUserOnline(userID string) bool {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	_, exists := h.clients[userID]
	return exists
}
