-- Migration 001: Initial schema baseline with UUID primary keys (SQLite)
CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY, -- This will be auth.users.id directly (UUID as text in SQLite)
    handle TEXT NOT NULL UNIQUE,
    full_name TEXT DEFAULT NULL,
    email TEXT NOT NULL UNIQUE,
    subscription_tier TEXT NOT NULL DEFAULT 'free',
    active_goal_name TEXT DEFAULT NULL,
    created_at DATETIME NOT NULL,
    avatar_url TEXT DEFAULT NULL
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_handle ON users(handle);
CREATE INDEX IF NOT EXISTS idx_users_active_goal ON users(active_goal_name);
