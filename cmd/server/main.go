// Package main is the entry point for the ChatGO server application.
package main

import (
	"fmt"
	"log"
	"net/http"

	"chatgo/internal/api"
	"chatgo/internal/db"
	"chatgo/internal/websocket"
)

func main() {
	// Connect to PostgreSQL.
	connectionString := "postgres://postgres:postgres@localhost:5432/chatgo?sslmode=disable"

	err := db.Connect(connectionString)
	if err != nil {
		log.Fatal("Database connection failed: ", err)
	}
	defer db.Close()

	// Create and start the WebSocket hub.
	hub := websocket.NewHub()
	go hub.Run()

	// Public endpoints (no auth required).
	http.HandleFunc("/api/health", api.HealthHandler)
	http.HandleFunc("/api/login", api.LoginHandler)

	// WebSocket endpoint.
	http.HandleFunc("/ws", websocket.Handler(hub))

	// User endpoints.
	http.HandleFunc("/api/users", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			// GET - any authenticated user can list users (for chat)
			api.AuthMiddleware(api.ListUsersHandler)(w, r)
		case http.MethodPost:
			// POST - only admin can create users
			api.AuthMiddleware(api.AdminMiddleware(api.CreateUserHandler))(w, r)
		default:
			http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/api/users/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodDelete:
			api.AuthMiddleware(api.AdminMiddleware(api.DeleteUserHandler))(w, r)
		case http.MethodPut:
			api.AuthMiddleware(api.AdminMiddleware(api.UpdateUserHandler))(w, r)
		default:
			http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		}
	})

	// Conversation endpoints (authenticated users).
	http.HandleFunc("/api/conversations", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			api.AuthMiddleware(api.CreateConversationHandler)(w, r)
		} else {
			http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		}
	})

	// Messages endpoint: /api/conversations/{id}/messages
	http.HandleFunc("/api/conversations/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			api.AuthMiddleware(api.GetMessagesHandler)(w, r)
		} else {
			http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		}
	})

	// Serve static files from frontend/public directory.
	fs := http.FileServer(http.Dir("frontend/public"))
	http.Handle("/", fs)

	fmt.Println("Server starting on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
