// Package websocket - HTTP handler for WebSocket connections
package websocket

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"

	"chatgo/internal/auth"
)

// upgrader configures the WebSocket upgrade.
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Allow connections from any origin (for development).
	// In production, you should check the origin!
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Handler handles WebSocket connection requests.
// It authenticates the user via JWT token in query parameter.
func Handler(hub *Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get token from query parameter.
		// WebSocket connections can't use Authorization header easily,
		// so we use a query parameter: /ws?token=xxx
		token := r.URL.Query().Get("token")
		if token == "" {
			http.Error(w, "Token required", http.StatusUnauthorized)
			return
		}

		// Validate the token.
		claims, err := auth.ValidateToken(token)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Upgrade HTTP connection to WebSocket.
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("WebSocket upgrade error: %v", err)
			return
		}

		// Create a new client.
		client := NewClient(hub, conn, claims.UserID, claims.Username)

		// Register the client with the hub.
		hub.register <- client

		// Start the read and write pumps in goroutines.
		// These handle all communication for this client.
		go client.WritePump()
		go client.ReadPump()
	}
}
