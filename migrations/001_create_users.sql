-- Migration: Create users table
-- This table stores all user accounts for the chat application.

CREATE TABLE IF NOT EXISTS users (
    -- UUID as primary key (better than auto-increment for distributed systems)
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Username must be unique - no two users can have the same name
    username VARCHAR(50) UNIQUE NOT NULL,

    -- We store a hash of the password, never the plain password!
    password_hash VARCHAR(255) NOT NULL,

    -- Admin users can create/edit/delete other users
    is_admin BOOLEAN DEFAULT FALSE,

    -- Automatically set when the row is created
    created_at TIMESTAMP DEFAULT NOW()
);

-- Create an index on username for faster lookups during login
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);

-- Insert a default admin user (password: "admin")
-- The hash below is bcrypt hash of "admin" - we'll generate real hashes in Go
INSERT INTO users (username, password_hash, is_admin)
VALUES ('admin', '$2a$10$N9qo8uLOickgx2ZMRZoMye.IjqQBrkHx3PLHiLqf4KCwIq.OMwqK.', true)
ON CONFLICT (username) DO NOTHING;
