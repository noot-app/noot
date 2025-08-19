-- Migration 001: Initial schema with ULID primary keys and no foreign keys (PlanetScale-friendly)
CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    provider TEXT NOT NULL,
    subject TEXT NOT NULL,
    email TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(provider, subject)
);

CREATE INDEX IF NOT EXISTS idx_users_provider_subject ON users(provider, subject);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- Meals table with ULID primary key and no foreign key constraint
CREATE TABLE IF NOT EXISTS meals (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    transcript TEXT NOT NULL,
    items_json TEXT NOT NULL, -- JSON array of meal items with full nutrition data
    total_calories REAL NOT NULL DEFAULT 0,
    total_protein_g REAL NOT NULL DEFAULT 0,
    total_fat_g REAL NOT NULL DEFAULT 0,
    total_carbs_g REAL NOT NULL DEFAULT 0,
    total_fiber_g REAL NOT NULL DEFAULT 0,
    total_sodium_mg REAL NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_meals_user_id ON meals(user_id);
CREATE INDEX IF NOT EXISTS idx_meals_created_at ON meals(created_at);
CREATE INDEX IF NOT EXISTS idx_meals_user_created ON meals(user_id, created_at);