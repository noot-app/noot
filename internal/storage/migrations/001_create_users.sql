-- Migration 001: Initial schema for users and meals
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    provider TEXT NOT NULL,
    subject TEXT NOT NULL,
    email TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(provider, subject)
);

CREATE INDEX IF NOT EXISTS idx_users_provider_subject ON users(provider, subject);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);