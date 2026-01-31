# ChatGo Tutorial - Complete Guide

## Overview

ChatGo is a real-time chat application with:
- **Backend**: Go (REST API + WebSocket)
- **Frontend**: TypeScript
- **Database**: PostgreSQL
- **Auth**: JWT + bcrypt

---

## 1. Architecture

```
┌─────────────────┐         ┌──────────────────┐
│    Frontend     │◄───────►│   Go Backend     │
│  (TypeScript)   │  HTTP   │                  │
│                 │◄───────►│  ┌────────────┐  │
│                 │   WS    │  │ WebSocket  │  │
└─────────────────┘         │  │    Hub     │  │
                            │  └────────────┘  │
                            │        ↓         │
                            │  ┌────────────┐  │
                            │  │ PostgreSQL │  │
                            │  └────────────┘  │
                            └──────────────────┘
```

---

## 2. Server Startup (`cmd/server/main.go`)

1. **Connect to database**
2. **Create WebSocket Hub** and run in goroutine
3. **Register routes**:
   - `POST /api/login` - Authentication
   - `GET /api/users` - List users
   - `POST /api/conversations` - Create/get conversation
   - `GET /api/conversations/{id}/messages` - Load history
   - `GET /ws?token=xxx` - WebSocket connection
4. **Serve static files** from `frontend/public/`
5. **Listen on port 8080**

---

## 3. Authentication Flow

### Login Process

```
Frontend                    Backend
   │                           │
   │ POST /api/login           │
   │ {username, password}      │
   │──────────────────────────►│
   │                           │ 1. Find user by username
   │                           │ 2. bcrypt.Compare(password, hash)
   │                           │ 3. Generate JWT (24h expiry)
   │◄──────────────────────────│
   │ {token, username}         │
   │                           │
   │ Save to localStorage      │
```

### JWT Token Contents

```json
{
  "user_id": "uuid",
  "username": "admin",
  "is_admin": true,
  "exp": 1706800000,
  "iat": 1706713600
}
```

### Protected Endpoints

All requests include: `Authorization: Bearer <token>`

The `AuthMiddleware` validates the token and extracts user info into request context.

---

## 4. Real-Time Messaging (WebSocket)

### Connection

```typescript
// Frontend connects after login
websocket = new WebSocket(`ws://localhost:8080/ws?token=${authToken}`);
```

### Hub Architecture

```
                    ┌─────────────────────┐
                    │        Hub          │
                    │                     │
   register ───────►│  clients map:       │
  unregister ──────►│    userID → Client  │
   broadcast ──────►│                     │
                    └─────────────────────┘
                              │
              ┌───────────────┼───────────────┐
              ▼               ▼               ▼
         ┌────────┐      ┌────────┐      ┌────────┐
         │Client A│      │Client B│      │Client C│
         │ send ◄─┼──────┤        │      │        │
         └────────┘      └────────┘      └────────┘
```

### Message Flow

**Sending a message:**

```
Alice (Frontend)           Backend                    Bob (Frontend)
      │                       │                            │
      │ WS: {type: "message", │                            │
      │  conversation_id,     │                            │
      │  content: "Hello"}    │                            │
      │──────────────────────►│                            │
      │                       │ 1. Verify participant      │
      │                       │ 2. Save to database        │
      │                       │ 3. Send to both users      │
      │◄──────────────────────│                            │
      │                       │───────────────────────────►│
      │ Display message       │              Display message│
```

### Typing Indicator

```typescript
// Frontend sends when typing
websocket.send(JSON.stringify({
  type: "typing",
  conversation_id: id,
  is_typing: true
}));

// Auto-stops after 2 seconds of no input
```

---

## 5. Database Schema

```sql
-- Users
CREATE TABLE users (
    id UUID PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    is_admin BOOLEAN DEFAULT FALSE
);

-- Conversations (supports future group chat)
CREATE TABLE conversations (
    id UUID PRIMARY KEY
);

CREATE TABLE conversation_participants (
    conversation_id UUID REFERENCES conversations(id),
    user_id UUID REFERENCES users(id),
    PRIMARY KEY (conversation_id, user_id)
);

-- Messages
CREATE TABLE messages (
    id UUID PRIMARY KEY,
    conversation_id UUID REFERENCES conversations(id),
    sender_id UUID REFERENCES users(id),
    content TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);
```

---

## 6. Frontend Structure (`frontend/src/main.ts`)

### State Variables

```typescript
let authToken: string | null;        // JWT token
let currentUserId: string | null;    // Current user
let websocket: WebSocket | null;     // WS connection
let currentConversationId: string;   // Active chat
```

### Key Functions

| Function | Purpose |
|----------|---------|
| `init()` | Check localStorage, setup listeners |
| `handleLogin()` | POST to /api/login, store token |
| `connectWebSocket()` | Establish WS, handle messages |
| `sendMessage()` | Send via WebSocket |
| `loadMessages()` | Fetch history via REST |
| `handleIncomingMessage()` | Display received messages |

---

## 7. Admin Panel

Admin users can manage users:

| Action | Endpoint | Notes |
|--------|----------|-------|
| Create | `POST /api/users` | Set username, password, is_admin |
| Edit | `PUT /api/users/{id}` | Can't edit self |
| Delete | `DELETE /api/users/{id}` | Can't delete self |

Protected by `AdminMiddleware` which checks `claims.IsAdmin`.

---

## 8. File Structure

```
ChatGo/
├── cmd/
│   ├── server/main.go       # Entry point
│   └── genhash/main.go      # Password hash utility
├── internal/
│   ├── api/
│   │   ├── handlers.go      # Health check
│   │   ├── auth_handlers.go # Login
│   │   ├── user_handlers.go # User CRUD
│   │   ├── conversation_handlers.go
│   │   └── middleware.go    # Auth/Admin middleware
│   ├── auth/
│   │   ├── jwt.go           # Token generation/validation
│   │   └── password.go      # bcrypt helpers
│   ├── db/
│   │   ├── db.go            # Connection
│   │   ├── users.go         # User queries
│   │   ├── conversations.go
│   │   └── messages.go
│   ├── models/
│   │   ├── user.go
│   │   └── conversation.go
│   └── websocket/
│       ├── hub.go           # Connection manager
│       ├── client.go        # Read/Write pumps
│       └── handler.go       # WS upgrade handler
├── frontend/
│   ├── src/main.ts          # TypeScript source
│   ├── public/
│   │   ├── index.html       # UI with CSS
│   │   └── js/main.js       # Compiled JS
│   ├── package.json
│   └── tsconfig.json
├── migrations/
│   ├── 001_create_users.sql
│   └── 002_create_chat_tables.sql
├── go.mod
└── go.sum
```

---

## 9. Running the Application

```bash
# 1. Setup PostgreSQL database
psql -U postgres -c "CREATE DATABASE chatgo"
psql -U postgres -d chatgo < migrations/001_create_users.sql
psql -U postgres -d chatgo < migrations/002_create_chat_tables.sql

# 2. Build frontend
cd frontend
npm install
npx tsc

# 3. Run server
cd ..
go run cmd/server/main.go
# Server at http://localhost:8080

# 4. Login with: admin / admin
```

---

## 10. Security Notes

**Implemented:**
- bcrypt password hashing
- JWT with expiration
- Parameterized SQL queries
- Authorization middleware

**Production TODO:**
- Move JWT secret to environment variable
- Enable HTTPS/WSS
- Validate WebSocket origin
- Rate limit login attempts
