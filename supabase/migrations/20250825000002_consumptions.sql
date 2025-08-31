-- Migration 002: Create consumptions table with UUID primary keys and denormalized nutrition totals (PostgreSQL)
-- A consumption represents a meal or food intake event logged by a user. For RLS purposes, users should only be able to access their own consumptions.
-- This means that they should be able to read, create, update, and delete their own consumptions, but not those of other users.
CREATE TABLE IF NOT EXISTS consumptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
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
    omega3_ala_g REAL NOT NULL DEFAULT 0,
    omega3_epa_g REAL NOT NULL DEFAULT 0,
    omega3_dha_g REAL NOT NULL DEFAULT 0,
    omega6_g REAL NOT NULL DEFAULT 0,
    creatine_mg REAL NOT NULL DEFAULT 0,
    caffeine_mg REAL NOT NULL DEFAULT 0,
    alcohol_g REAL NOT NULL DEFAULT 0,
    polyunsaturated_fat_g REAL NOT NULL DEFAULT 0,
    monounsaturated_fat_g REAL NOT NULL DEFAULT 0,
    -- Additional metadata fields
    note TEXT CONSTRAINT consumptions_note_length_check CHECK (LENGTH(note) <= 1000),
    is_public BOOLEAN NOT NULL DEFAULT FALSE, -- Allow users to make consumptions publicly viewable
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE, -- When the consumption was last modified
    FOREIGN KEY (user_id) REFERENCES profiles(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_consumptions_user_id ON consumptions(user_id);
CREATE INDEX IF NOT EXISTS idx_consumptions_created_at ON consumptions(created_at);
CREATE INDEX IF NOT EXISTS idx_consumptions_user_created ON consumptions(user_id, created_at);
CREATE INDEX IF NOT EXISTS idx_consumptions_public ON consumptions (is_public) WHERE is_public = TRUE;

-- Enable RLS for consumptions table
ALTER TABLE consumptions ENABLE ROW LEVEL SECURITY;

-- RLS policies for consumptions with public sharing support
CREATE POLICY "Users can view own consumptions and public ones" ON consumptions
  FOR SELECT USING (auth.uid() = user_id OR is_public = TRUE);

CREATE POLICY "Users can insert own consumptions" ON consumptions  
  FOR INSERT WITH CHECK (auth.uid() = user_id);

CREATE POLICY "Users can update own consumptions" ON consumptions
  FOR UPDATE USING (auth.uid() = user_id) WITH CHECK (auth.uid() = user_id);

CREATE POLICY "Users can delete own consumptions" ON consumptions
  FOR DELETE USING (auth.uid() = user_id);
