-- Migration 008: Refactor users table to use UUIDs directly as primary key (SQLite version)
-- This migration transforms the users table from subject/provider lookup to direct UUID reference
-- Note: SQLite doesn't have auth.users table, so we simulate the same structure for development

-- Step 1: Create new users table structure with UUID primary keys
CREATE TABLE IF NOT EXISTS users_new (
    id TEXT PRIMARY KEY, -- UUID as text in SQLite
    handle TEXT NOT NULL UNIQUE,
    full_name TEXT DEFAULT NULL,
    email TEXT NOT NULL UNIQUE,
    subscription_tier TEXT NOT NULL DEFAULT 'free',
    active_goal_name TEXT DEFAULT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    avatar_url TEXT DEFAULT NULL
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_users_new_email ON users_new(email);
CREATE INDEX IF NOT EXISTS idx_users_new_handle ON users_new(handle);
CREATE INDEX IF NOT EXISTS idx_users_new_active_goal ON users_new(active_goal_name);

-- Step 2: Generate UUIDs for existing users and migrate data
-- For SQLite, we'll generate new UUIDs for existing users
-- This simulates what would happen with real auth.users.id values

-- First, let's add a temp column to the old users table to store the generated UUIDs
ALTER TABLE users ADD COLUMN temp_uuid TEXT;

-- Generate UUIDs for existing users (in a real migration, you'd use proper UUID generation)
-- For simplicity in SQLite, we'll create deterministic UUIDs based on existing data
UPDATE users 
SET temp_uuid = 
    CASE 
        WHEN provider = 'email' AND subject = 'monalisa' 
            THEN '550e8400-e29b-41d4-a716-446655440000'  -- Fixed UUID for test user
        WHEN provider = 'email' AND subject = 'alice'
            THEN '550e8400-e29b-41d4-a716-446655440001'  -- Fixed UUID for alice
        ELSE 
            '550e8400-e29b-41d4-a716-' || printf('%012d', abs(random() % 1000000000000))
    END;

-- Insert migrated data into new table
INSERT INTO users_new (id, handle, full_name, email, subscription_tier, active_goal_name, created_at, avatar_url)
SELECT 
    temp_uuid as id,
    COALESCE(handle, 'user_' || substr(temp_uuid, 1, 8)) as handle,
    full_name,
    email,
    subscription_tier,
    active_goal_name,
    created_at,
    avatar_url
FROM users
WHERE temp_uuid IS NOT NULL;

-- Step 3: Update all tables that reference users.id to use the new UUIDs
-- First, let's create a mapping of old IDs to new UUIDs
CREATE TEMP TABLE user_id_mapping AS
SELECT 
    id as old_user_id,
    temp_uuid as new_user_id
FROM users
WHERE temp_uuid IS NOT NULL;

-- Update consumptions table
-- Add new column
ALTER TABLE consumptions ADD COLUMN new_user_id TEXT;

-- Update with mapped UUIDs
UPDATE consumptions 
SET new_user_id = (
    SELECT new_user_id 
    FROM user_id_mapping 
    WHERE user_id_mapping.old_user_id = consumptions.user_id
);

-- Update user_goals table
ALTER TABLE user_goals ADD COLUMN new_user_id TEXT;

UPDATE user_goals 
SET new_user_id = (
    SELECT new_user_id 
    FROM user_id_mapping 
    WHERE user_id_mapping.old_user_id = user_goals.user_id
);

-- Update user_biometrics table
ALTER TABLE user_biometrics ADD COLUMN new_user_id TEXT;

UPDATE user_biometrics 
SET new_user_id = (
    SELECT new_user_id 
    FROM user_id_mapping 
    WHERE user_id_mapping.old_user_id = user_biometrics.user_id
);

-- Step 4: Replace the old foreign key columns with new ones
-- SQLite doesn't support dropping columns easily, so we'll recreate the tables

-- Recreate consumptions table
CREATE TABLE consumptions_new (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    transcript TEXT NOT NULL,
    total_calories REAL NOT NULL DEFAULT 0,
    total_protein_g REAL NOT NULL DEFAULT 0,
    total_fat_g REAL NOT NULL DEFAULT 0,
    total_carbs_g REAL NOT NULL DEFAULT 0,
    dietary_fiber_g REAL NOT NULL DEFAULT 0,
    total_sodium_mg REAL NOT NULL DEFAULT 0,
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
    updated_at DATETIME,
    FOREIGN KEY (user_id) REFERENCES users_new(id) ON DELETE CASCADE
);

-- Copy data to new consumptions table
INSERT INTO consumptions_new 
SELECT id, new_user_id, transcript, total_calories, total_protein_g, total_fat_g, total_carbs_g,
       dietary_fiber_g, total_sodium_mg, saturated_fat_g, trans_fat_g, cholesterol_mg, total_sugars_g,
       added_sugars_g, vitamin_a_mcg, vitamin_c_mg, vitamin_d_mcg, vitamin_e_mg, vitamin_k_mcg,
       thiamine_mg, riboflavin_mg, niacin_mg, vitamin_b6_mg, folate_mcg, vitamin_b12_mcg,
       biotin_mcg, pantothenic_acid_mg, choline_mg, calcium_mg, iron_mg, magnesium_mg,
       phosphorus_mg, potassium_mg, zinc_mg, copper_mg, manganese_mg, selenium_mcg,
       iodine_mcg, molybdenum_mcg, chromium_mcg, fluoride_mg, chloride_mg, created_at, updated_at
FROM consumptions
WHERE new_user_id IS NOT NULL;

-- Recreate user_goals table
CREATE TABLE user_goals_new (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    name TEXT NOT NULL,
    overrides_json TEXT NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users_new(id) ON DELETE CASCADE,
    UNIQUE(user_id, name)
);

-- Copy data
INSERT INTO user_goals_new 
SELECT id, new_user_id, name, overrides_json, created_at, updated_at
FROM user_goals
WHERE new_user_id IS NOT NULL;

-- Recreate user_biometrics table
CREATE TABLE user_biometrics_new (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL UNIQUE,
    birth_date DATE,
    sex TEXT,
    height_cm REAL,
    weight_kg REAL,
    activity_level TEXT,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users_new(id) ON DELETE CASCADE
);

-- Copy data
INSERT INTO user_biometrics_new 
SELECT id, new_user_id, birth_date, sex, height_cm, weight_kg, activity_level, created_at, updated_at
FROM user_biometrics
WHERE new_user_id IS NOT NULL;

-- Step 5: Replace old tables with new ones
DROP TABLE users;
DROP TABLE consumptions;
DROP TABLE user_goals;
DROP TABLE user_biometrics;

ALTER TABLE users_new RENAME TO users;
ALTER TABLE consumptions_new RENAME TO consumptions;
ALTER TABLE user_goals_new RENAME TO user_goals;
ALTER TABLE user_biometrics_new RENAME TO user_biometrics;

-- Recreate indexes with correct names
DROP INDEX IF EXISTS idx_users_new_email;
DROP INDEX IF EXISTS idx_users_new_handle; 
DROP INDEX IF EXISTS idx_users_new_active_goal;

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_handle ON users(handle);
CREATE INDEX IF NOT EXISTS idx_users_active_goal ON users(active_goal_name);

CREATE INDEX IF NOT EXISTS idx_consumptions_user_id ON consumptions(user_id);
CREATE INDEX IF NOT EXISTS idx_consumptions_created_at ON consumptions(created_at);
CREATE INDEX IF NOT EXISTS idx_consumptions_user_created ON consumptions(user_id, created_at);

-- Clean up temp table
DROP TABLE IF EXISTS user_id_mapping;