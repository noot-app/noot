-- Migration 007: Refactor items structure from cache to permanent storage with relational consumption_items
-- This migration transforms the current JSON-embedded items approach to a proper relational structure

-- Step 1: Create the new items table (evolved from items_cache)
-- Remove TTL fields and add updated_at for 30-day refresh logic
CREATE TABLE IF NOT EXISTS items (
    id TEXT PRIMARY KEY,
    normalized_name TEXT NOT NULL,
    normalized_brand TEXT NOT NULL DEFAULT '',
    display_name TEXT NOT NULL,
    display_brand TEXT DEFAULT '',
    -- Nutrition data in canonical units (grams/mg/kcal per 100g)
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
    -- Timestamps for 30-day refresh logic (no more TTL)
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    UNIQUE(normalized_name, normalized_brand)
);

-- Step 2: Migrate data from items_cache to items (if items_cache exists)
INSERT OR IGNORE INTO items (
    id, normalized_name, normalized_brand, display_name, display_brand,
    calories_per_100g, protein_g_per_100g, total_fat_g_per_100g,
    saturated_fat_g_per_100g, trans_fat_g_per_100g, cholesterol_mg_per_100g,
    sodium_mg_per_100g, total_carbs_g_per_100g, dietary_fiber_g_per_100g,
    total_sugars_g_per_100g, added_sugars_g_per_100g,
    vitamin_a_mcg_per_100g, vitamin_c_mg_per_100g, vitamin_d_mcg_per_100g,
    vitamin_e_mg_per_100g, vitamin_k_mcg_per_100g, thiamine_mg_per_100g,
    riboflavin_mg_per_100g, niacin_mg_per_100g, vitamin_b6_mg_per_100g,
    folate_mcg_per_100g, vitamin_b12_mcg_per_100g, biotin_mcg_per_100g,
    pantothenic_acid_mg_per_100g, choline_mg_per_100g, calcium_mg_per_100g,
    iron_mg_per_100g, magnesium_mg_per_100g, phosphorus_mg_per_100g,
    potassium_mg_per_100g, zinc_mg_per_100g, copper_mg_per_100g,
    manganese_mg_per_100g, selenium_mcg_per_100g, iodine_mcg_per_100g,
    molybdenum_mcg_per_100g, chromium_mcg_per_100g, fluoride_mg_per_100g,
    chloride_mg_per_100g, created_at, updated_at
)
SELECT 
    id, normalized_name, normalized_brand, display_name, display_brand,
    calories_per_100g, protein_g_per_100g, total_fat_g_per_100g,
    saturated_fat_g_per_100g, trans_fat_g_per_100g, cholesterol_mg_per_100g,
    sodium_mg_per_100g, total_carbs_g_per_100g, dietary_fiber_g_per_100g,
    total_sugars_g_per_100g, added_sugars_g_per_100g,
    vitamin_a_mcg_per_100g, vitamin_c_mg_per_100g, vitamin_d_mcg_per_100g,
    vitamin_e_mg_per_100g, vitamin_k_mcg_per_100g, thiamine_mg_per_100g,
    riboflavin_mg_per_100g, niacin_mg_per_100g, vitamin_b6_mg_per_100g,
    folate_mcg_per_100g, vitamin_b12_mcg_per_100g, biotin_mcg_per_100g,
    pantothenic_acid_mg_per_100g, choline_mg_per_100g, calcium_mg_per_100g,
    iron_mg_per_100g, magnesium_mg_per_100g, phosphorus_mg_per_100g,
    potassium_mg_per_100g, zinc_mg_per_100g, copper_mg_per_100g,
    manganese_mg_per_100g, selenium_mcg_per_100g, iodine_mcg_per_100g,
    molybdenum_mcg_per_100g, chromium_mcg_per_100g, fluoride_mg_per_100g,
    chloride_mg_per_100g, created_at, updated_at
FROM items_cache
WHERE EXISTS (SELECT 1 FROM sqlite_master WHERE type='table' AND name='items_cache');

-- Step 3: Create consumption_items junction table
CREATE TABLE IF NOT EXISTS consumption_items (
    id TEXT PRIMARY KEY,
    consumption_id TEXT NOT NULL,
    item_id TEXT NOT NULL,
    grams REAL NOT NULL, -- Actual grams consumed (already normalized internally)
    user_quantity REAL, -- Original user input quantity for display (e.g., 3.0 for "3 sticks")
    user_unit TEXT,     -- Original user input unit for display (e.g., "sticks")
    created_at DATETIME NOT NULL,
    FOREIGN KEY (consumption_id) REFERENCES consumptions(id) ON DELETE CASCADE,
    FOREIGN KEY (item_id) REFERENCES items(id) ON DELETE CASCADE,
    UNIQUE(consumption_id, item_id)
);

-- Step 4: Create new consumptions table without items_json and global user_quantity/unit
CREATE TABLE IF NOT EXISTS consumptions_new (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    transcript TEXT NOT NULL,
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
    created_at DATETIME NOT NULL,
    updated_at DATETIME, -- When the consumption was last modified
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Step 5: Migrate existing consumption data (exclude items_json, user_quantity, user_unit)
INSERT INTO consumptions_new (
    id, user_id, transcript, total_calories, total_protein_g,
    total_fat_g, total_carbs_g, dietary_fiber_g, total_sodium_mg,
    saturated_fat_g, trans_fat_g, cholesterol_mg, total_sugars_g, added_sugars_g,
    vitamin_a_mcg, vitamin_c_mg, vitamin_d_mcg, vitamin_e_mg, vitamin_k_mcg,
    thiamine_mg, riboflavin_mg, niacin_mg, vitamin_b6_mg, folate_mcg, vitamin_b12_mcg,
    biotin_mcg, pantothenic_acid_mg, choline_mg, calcium_mg, iron_mg, magnesium_mg,
    phosphorus_mg, potassium_mg, zinc_mg, copper_mg, manganese_mg, selenium_mcg,
    iodine_mcg, molybdenum_mcg, chromium_mcg, fluoride_mg, chloride_mg,
    created_at, updated_at
)
SELECT 
    id, user_id, transcript, total_calories, total_protein_g,
    total_fat_g, total_carbs_g, dietary_fiber_g, total_sodium_mg,
    saturated_fat_g, trans_fat_g, cholesterol_mg, total_sugars_g, added_sugars_g,
    vitamin_a_mcg, vitamin_c_mg, vitamin_d_mcg, vitamin_e_mg, vitamin_k_mcg,
    thiamine_mg, riboflavin_mg, niacin_mg, vitamin_b6_mg, folate_mcg, vitamin_b12_mcg,
    biotin_mcg, pantothenic_acid_mg, choline_mg, calcium_mg, iron_mg, magnesium_mg,
    phosphorus_mg, potassium_mg, zinc_mg, copper_mg, manganese_mg, selenium_mcg,
    iodine_mcg, molybdenum_mcg, chromium_mcg, fluoride_mg, chloride_mg,
    created_at, updated_at
FROM consumptions;

-- Step 6: Replace old consumptions table
DROP TABLE consumptions;
ALTER TABLE consumptions_new RENAME TO consumptions;

-- Step 7: Drop items_cache table if it exists (data already migrated to items)
DROP TABLE IF EXISTS items_cache;

-- Step 8: Create indexes for the new tables
CREATE INDEX IF NOT EXISTS idx_items_normalized ON items(normalized_name, normalized_brand);
CREATE INDEX IF NOT EXISTS idx_items_updated_at ON items(updated_at);
CREATE INDEX IF NOT EXISTS idx_items_display_name ON items(display_name);

CREATE INDEX IF NOT EXISTS idx_consumption_items_consumption_id ON consumption_items(consumption_id);
CREATE INDEX IF NOT EXISTS idx_consumption_items_item_id ON consumption_items(item_id);
CREATE INDEX IF NOT EXISTS idx_consumption_items_created_at ON consumption_items(created_at);

CREATE INDEX IF NOT EXISTS idx_consumptions_user_id ON consumptions(user_id);
CREATE INDEX IF NOT EXISTS idx_consumptions_created_at ON consumptions(created_at);
CREATE INDEX IF NOT EXISTS idx_consumptions_user_created ON consumptions(user_id, created_at);