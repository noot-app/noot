-- Migration 002: Create meals table with ULID primary keys and denormalized nutrition totals
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
    created_at DATETIME NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_meals_user_id ON meals(user_id);
CREATE INDEX IF NOT EXISTS idx_meals_created_at ON meals(created_at);
CREATE INDEX IF NOT EXISTS idx_meals_user_created ON meals(user_id, created_at);