-- Migration 001: Initial schema baseline with ULID primary keys
CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    provider TEXT NOT NULL,
    subject TEXT NOT NULL,
    email TEXT NOT NULL,
    subscription_tier TEXT NOT NULL DEFAULT 'free',
    created_at DATETIME NOT NULL,
    UNIQUE(provider, subject)
);

CREATE INDEX IF NOT EXISTS idx_users_provider_subject ON users(provider, subject);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
