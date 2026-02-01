# CLAUDE.md - ChatGo Project

## Build Commands

```bash
# Build Go backend
cd /c/Attracs/ChatGo && go build ./...

# Build frontend TypeScript
cd /c/Attracs/ChatGo/frontend && npm run build

# Run server
cd /c/Attracs/ChatGo && go run ./cmd/server
```

## Linting

```bash
# Lint Go code (install first: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
cd /c/Attracs/ChatGo && golangci-lint run

# Lint TypeScript
cd /c/Attracs/ChatGo/frontend && npm run lint

# Fix TypeScript lint issues automatically
cd /c/Attracs/ChatGo/frontend && npm run lint:fix
```

## Database Migrations

Migrations are in `migrations/` directory. Apply them manually to PostgreSQL:
```bash
psql -U postgres -d chatgo -f migrations/001_create_users.sql
psql -U postgres -d chatgo -f migrations/002_create_chat_tables.sql
psql -U postgres -d chatgo -f migrations/003_add_conversation_name.sql
```
