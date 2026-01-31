// Package db handles database connections and queries.
package db

import (
	"database/sql"
	"fmt"

	// The underscore import means: import for side effects only.
	// The pq package registers itself as a PostgreSQL driver when imported.
	// We don't call any pq functions directly - we use the standard database/sql interface.
	_ "github.com/lib/pq"
)

// DB is the global database connection.
// In a larger app, you might pass this around instead of using a global.
var DB *sql.DB

// Connect establishes a connection to the PostgreSQL database.
// It takes the connection string and returns an error if connection fails.
func Connect(connectionString string) error {
	var err error

	// sql.Open doesn't actually connect - it just validates the arguments.
	// "postgres" is the driver name (registered by the pq package).
	DB, err = sql.Open("postgres", connectionString)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Ping actually tries to connect to the database.
	// This verifies that the connection string is correct and the database is reachable.
	err = DB.Ping()
	if err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	fmt.Println("Connected to database successfully!")
	return nil
}

// Close closes the database connection.
// Call this when shutting down the application.
func Close() {
	if DB != nil {
		DB.Close()
	}
}
