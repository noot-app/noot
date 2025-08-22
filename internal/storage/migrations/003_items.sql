-- Migration 003: Create items table for reusable nutrition data with soft TTL
CREATE TABLE IF NOT EXISTS items (
    id TEXT PRIMARY KEY,
    normalized_name TEXT NOT NULL,
    normalized_brand TEXT NOT NULL DEFAULT '',
    display_name TEXT NOT NULL,
    display_brand TEXT DEFAULT '',
    -- Nutrition data in canonical units (grams/mg/kcal)
    calories_per_100g REAL NOT NULL DEFAULT 0,
    protein_g_per_100g REAL NOT NULL DEFAULT 0,
    total_fat_g_per_100g REAL NOT NULL DEFAULT 0,
    saturated_fat_g_per_100g REAL NOT NULL DEFAULT 0,
    trans_fat_g_per_100g REAL NOT NULL DEFAULT 0,
    cholesterol_mg_per_100g REAL NOT NULL DEFAULT 0,
    sodium_mg_per_100g REAL NOT NULL DEFAULT 0,
    total_carbs_g_per_100g REAL NOT NULL DEFAULT 0,
    dietary_fiber_g_per_100g REAL NOT NULL DEFAULT 0,
    total_sugars_g_per_100g REAL NOT NULL DEFAULT 0,
    added_sugars_g_per_100g REAL NOT NULL DEFAULT 0,
    -- Additional nutrition fields
    vitamin_a_mcg_per_100g REAL NOT NULL DEFAULT 0,
    vitamin_c_mg_per_100g REAL NOT NULL DEFAULT 0,
    vitamin_d_mcg_per_100g REAL NOT NULL DEFAULT 0,
    vitamin_e_mg_per_100g REAL NOT NULL DEFAULT 0,
    vitamin_k_mcg_per_100g REAL NOT NULL DEFAULT 0,
    thiamine_mg_per_100g REAL NOT NULL DEFAULT 0,
    riboflavin_mg_per_100g REAL NOT NULL DEFAULT 0,
    niacin_mg_per_100g REAL NOT NULL DEFAULT 0,
    vitamin_b6_mg_per_100g REAL NOT NULL DEFAULT 0,
    folate_mcg_per_100g REAL NOT NULL DEFAULT 0,
    vitamin_b12_mcg_per_100g REAL NOT NULL DEFAULT 0,
    biotin_mcg_per_100g REAL NOT NULL DEFAULT 0,
    pantothenic_acid_mg_per_100g REAL NOT NULL DEFAULT 0,
    choline_mg_per_100g REAL NOT NULL DEFAULT 0,
    calcium_mg_per_100g REAL NOT NULL DEFAULT 0,
    iron_mg_per_100g REAL NOT NULL DEFAULT 0,
    magnesium_mg_per_100g REAL NOT NULL DEFAULT 0,
    phosphorus_mg_per_100g REAL NOT NULL DEFAULT 0,
    potassium_mg_per_100g REAL NOT NULL DEFAULT 0,
    zinc_mg_per_100g REAL NOT NULL DEFAULT 0,
    copper_mg_per_100g REAL NOT NULL DEFAULT 0,
    manganese_mg_per_100g REAL NOT NULL DEFAULT 0,
    selenium_mcg_per_100g REAL NOT NULL DEFAULT 0,
    iodine_mcg_per_100g REAL NOT NULL DEFAULT 0,
    molybdenum_mcg_per_100g REAL NOT NULL DEFAULT 0,
    chromium_mcg_per_100g REAL NOT NULL DEFAULT 0,
    fluoride_mg_per_100g REAL NOT NULL DEFAULT 0,
    chloride_mg_per_100g REAL NOT NULL DEFAULT 0,
    -- Soft TTL fields
    fetched_at DATETIME NOT NULL,
    expires_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    UNIQUE(normalized_name, normalized_brand)
);

CREATE INDEX IF NOT EXISTS idx_items_normalized ON items(normalized_name, normalized_brand);
CREATE INDEX IF NOT EXISTS idx_items_expires_at ON items(expires_at);
CREATE INDEX IF NOT EXISTS idx_items_display_name ON items(display_name);
