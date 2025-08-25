-- Migration 007: Add handle and full_name columns to users table (SQLite)
ALTER TABLE users ADD COLUMN handle TEXT;
ALTER TABLE users ADD COLUMN full_name TEXT DEFAULT NULL;
ALTER TABLE users ADD COLUMN avatar_url TEXT DEFAULT NULL;

-- Since we cannot modify existing UNIQUE constraints in SQLite easily,
-- we'll add an index for handle uniqueness
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_handle ON users(handle);