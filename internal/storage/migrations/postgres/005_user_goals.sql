-- Migration 005: User goals for Pro-tier custom overrides (PostgreSQL)
-- For RLS purposes, users should only be able to access their own goals if they are pro-tier subscribers.
-- This means that they should be able to read, create, update, and delete their own goals, but not those of other users.
CREATE TABLE IF NOT EXISTS user_goals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    name TEXT NOT NULL DEFAULT 'custom',
    overrides_json TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE(user_id, name)
);

CREATE INDEX IF NOT EXISTS idx_user_goals_user_id ON user_goals(user_id);
CREATE INDEX IF NOT EXISTS idx_user_goals_updated_at ON user_goals(updated_at);
