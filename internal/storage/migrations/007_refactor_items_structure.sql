-- Migration 007: Remove items_json, user_quantity, user_unit from consumptions table
-- This is a simplified version that just removes the problematic fields

-- First create a new consumptions table without the problematic fields
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

-- Migrate existing consumption data (exclude items_json, user_quantity, user_unit)
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
FROM consumptions
WHERE EXISTS (SELECT 1 FROM sqlite_master WHERE type='table' AND name='consumptions');

-- Replace old consumptions table
DROP TABLE IF EXISTS consumptions;
ALTER TABLE consumptions_new RENAME TO consumptions;

-- Create indexes for the new table
CREATE INDEX IF NOT EXISTS idx_consumptions_user_id ON consumptions(user_id);
CREATE INDEX IF NOT EXISTS idx_consumptions_created_at ON consumptions(created_at);
CREATE INDEX IF NOT EXISTS idx_consumptions_user_created ON consumptions(user_id, created_at);