-- Migration 001: Initial schema baseline with UUID primary keys (PostgreSQL)
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    handle TEXT NOT NULL UNIQUE,
    full_name TEXT DEFAULT NULL,
    provider TEXT NOT NULL,
    subject TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    subscription_tier TEXT NOT NULL DEFAULT 'free',
    active_goal_name TEXT DEFAULT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    avatar_url TEXT DEFAULT NULL,
);

CREATE INDEX IF NOT EXISTS idx_users_provider_subject ON users(provider, subject);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_active_goal ON users(active_goal_name);
