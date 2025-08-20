-- Migration 002: Create consumptions table with ULID primary keys and denormalized nutrition totals
CREATE TABLE IF NOT EXISTS consumptions (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    transcript TEXT NOT NULL,
    items_json TEXT NOT NULL, -- JSON array of consumption items with full nutrition data
    total_calories REAL NOT NULL DEFAULT 0,
    total_protein_g REAL NOT NULL DEFAULT 0,
    total_fat_g REAL NOT NULL DEFAULT 0,
    total_carbs_g REAL NOT NULL DEFAULT 0,
    dietary_fiber_g REAL NOT NULL DEFAULT 0,
    total_sodium_mg REAL NOT NULL DEFAULT 0,
    -- Additional micronutrients for complete nutrition tracking
    saturated_fat_g REAL NOT NULL DEFAULT 0,
    trans_fat_g REAL NOT NULL DEFAULT 0,
    cholesterol_mg REAL NOT NULL DEFAULT 0,
    total_sugars_g REAL NOT NULL DEFAULT 0,
    added_sugars_g REAL NOT NULL DEFAULT 0,
    vitamin_a_mcg REAL NOT NULL DEFAULT 0,
    vitamin_c_mg REAL NOT NULL DEFAULT 0,
    vitamin_d_mcg REAL NOT NULL DEFAULT 0,
    vitamin_e_mg REAL NOT NULL DEFAULT 0,
    vitamin_k_mcg REAL NOT NULL DEFAULT 0,
    thiamine_mg REAL NOT NULL DEFAULT 0,
    riboflavin_mg REAL NOT NULL DEFAULT 0,
    niacin_mg REAL NOT NULL DEFAULT 0,
    vitamin_b6_mg REAL NOT NULL DEFAULT 0,
    folate_mcg REAL NOT NULL DEFAULT 0,
    vitamin_b12_mcg REAL NOT NULL DEFAULT 0,
    biotin_mcg REAL NOT NULL DEFAULT 0,
    pantothenic_acid_mg REAL NOT NULL DEFAULT 0,
    choline_mg REAL NOT NULL DEFAULT 0,
    calcium_mg REAL NOT NULL DEFAULT 0,
    iron_mg REAL NOT NULL DEFAULT 0,
    magnesium_mg REAL NOT NULL DEFAULT 0,
    phosphorus_mg REAL NOT NULL DEFAULT 0,
    potassium_mg REAL NOT NULL DEFAULT 0,
    zinc_mg REAL NOT NULL DEFAULT 0,
    copper_mg REAL NOT NULL DEFAULT 0,
    manganese_mg REAL NOT NULL DEFAULT 0,
    selenium_mcg REAL NOT NULL DEFAULT 0,
    iodine_mcg REAL NOT NULL DEFAULT 0,
    molybdenum_mcg REAL NOT NULL DEFAULT 0,
    chromium_mcg REAL NOT NULL DEFAULT 0,
    fluoride_mg REAL NOT NULL DEFAULT 0,
    chloride_mg REAL NOT NULL DEFAULT 0,
    user_quantity REAL, -- Original user input quantity for display (e.g., 3.0 for "3 sticks")
    user_unit TEXT,     -- Original user input unit for display (e.g., "sticks")
    created_at DATETIME NOT NULL,
    updated_at DATETIME, -- When the consumption was last modified
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_consumptions_user_id ON consumptions(user_id);
CREATE INDEX IF NOT EXISTS idx_consumptions_created_at ON consumptions(created_at);
CREATE INDEX IF NOT EXISTS idx_consumptions_user_created ON consumptions(user_id, created_at);
