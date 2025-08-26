-- Migration 008: Refactor users table to use auth.users.id directly as primary key
-- This migration transforms the users table from subject/provider lookup to direct UUID reference

-- Step 1: Create new users table structure that directly references auth.users.id
CREATE TABLE IF NOT EXISTS users_new (
    id UUID PRIMARY KEY, -- This will be auth.users.id directly
    handle TEXT NOT NULL UNIQUE,
    full_name TEXT DEFAULT NULL,
    email TEXT NOT NULL UNIQUE,
    subscription_tier TEXT NOT NULL DEFAULT 'free',
    active_goal_name TEXT DEFAULT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    avatar_url TEXT DEFAULT NULL,
    
    -- Add foreign key constraint to auth.users if we want to enforce it
    -- Note: This might not be possible in all Supabase setups, so we'll make it optional
    CONSTRAINT fk_users_auth FOREIGN KEY (id) REFERENCES auth.users(id) ON DELETE CASCADE DEFERRABLE INITIALLY DEFERRED
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_users_new_email ON users_new(email);
CREATE INDEX IF NOT EXISTS idx_users_new_handle ON users_new(handle);
CREATE INDEX IF NOT EXISTS idx_users_new_active_goal ON users_new(active_goal_name);

-- Step 2: Migrate existing data
-- For existing users, we need to map their current app IDs to their auth.users.id
-- This requires matching by email since that's the common field between auth.users and current users table

-- Insert migrated users, matching by email to get the correct auth.users.id
INSERT INTO users_new (id, handle, full_name, email, subscription_tier, active_goal_name, created_at, avatar_url)
SELECT 
    au.id as id, -- Use auth.users.id as the primary key
    COALESCE(u.handle, 'user_' || SUBSTRING(au.id::text, 1, 8)) as handle, -- Generate handle if missing
    u.full_name,
    u.email,
    u.subscription_tier,
    u.active_goal_name,
    COALESCE(u.created_at, au.created_at) as created_at,
    u.avatar_url
FROM auth.users au
LEFT JOIN users u ON u.email = au.email
WHERE au.email IS NOT NULL
ON CONFLICT (id) DO NOTHING; -- Skip duplicates

-- Step 3: Create temporary table to store the user ID mapping for updating foreign keys
CREATE TEMP TABLE user_id_mapping AS
SELECT 
    u.id as old_user_id,
    au.id as new_user_id
FROM users u
JOIN auth.users au ON au.email = u.email;

-- Step 4: Update all tables that reference users.id to use the new UUIDs
-- Update consumptions table
ALTER TABLE consumptions ADD COLUMN new_user_id UUID;

UPDATE consumptions 
SET new_user_id = um.new_user_id
FROM user_id_mapping um
WHERE consumptions.user_id = um.old_user_id;

-- Update consumption_items (if it exists and has user references)
-- This table doesn't directly reference users, so no update needed

-- Update user_goals table
ALTER TABLE user_goals ADD COLUMN new_user_id UUID;

UPDATE user_goals 
SET new_user_id = um.new_user_id
FROM user_id_mapping um
WHERE user_goals.user_id = um.old_user_id;

-- Update user_biometrics table
ALTER TABLE user_biometrics ADD COLUMN new_user_id UUID;

UPDATE user_biometrics 
SET new_user_id = um.new_user_id
FROM user_id_mapping um
WHERE user_biometrics.user_id = um.old_user_id;

-- Step 5: Drop old foreign key constraints and rename columns
-- Consumptions
ALTER TABLE consumptions DROP CONSTRAINT IF EXISTS consumptions_user_id_fkey;
ALTER TABLE consumptions DROP COLUMN user_id;
ALTER TABLE consumptions RENAME COLUMN new_user_id TO user_id;
ALTER TABLE consumptions ALTER COLUMN user_id SET NOT NULL;
ALTER TABLE consumptions ADD CONSTRAINT consumptions_user_id_fkey 
    FOREIGN KEY (user_id) REFERENCES users_new(id) ON DELETE CASCADE;

-- User Goals
ALTER TABLE user_goals DROP CONSTRAINT IF EXISTS user_goals_user_id_fkey;
ALTER TABLE user_goals DROP COLUMN user_id;
ALTER TABLE user_goals RENAME COLUMN new_user_id TO user_id;
ALTER TABLE user_goals ALTER COLUMN user_id SET NOT NULL;
ALTER TABLE user_goals ADD CONSTRAINT user_goals_user_id_fkey 
    FOREIGN KEY (user_id) REFERENCES users_new(id) ON DELETE CASCADE;

-- User Biometrics
ALTER TABLE user_biometrics DROP CONSTRAINT IF EXISTS user_biometrics_user_id_fkey;
ALTER TABLE user_biometrics DROP COLUMN user_id;
ALTER TABLE user_biometrics RENAME COLUMN new_user_id TO user_id;
ALTER TABLE user_biometrics ALTER COLUMN user_id SET NOT NULL;
ALTER TABLE user_biometrics ADD CONSTRAINT user_biometrics_user_id_fkey 
    FOREIGN KEY (user_id) REFERENCES users_new(id) ON DELETE CASCADE;

-- Step 6: Replace the old users table with the new one
DROP TABLE users CASCADE;
ALTER TABLE users_new RENAME TO users;

-- Recreate indexes with correct names
DROP INDEX IF EXISTS idx_users_new_email;
DROP INDEX IF EXISTS idx_users_new_handle; 
DROP INDEX IF EXISTS idx_users_new_active_goal;

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_handle ON users(handle);
CREATE INDEX IF NOT EXISTS idx_users_active_goal ON users(active_goal_name);

-- Step 7: Create trigger for automatic user creation when auth users are created
CREATE OR REPLACE FUNCTION public.handle_new_user()
RETURNS trigger AS $$
BEGIN
  INSERT INTO public.users (id, handle, full_name, email, subscription_tier, created_at, avatar_url)
  VALUES (
    NEW.id,
    COALESCE(NEW.raw_user_meta_data->>'handle', 'user_' || SUBSTRING(NEW.id::text, 1, 8)),
    NEW.raw_user_meta_data->>'full_name',
    NEW.email,
    'free', -- Default tier
    NEW.created_at,
    NEW.raw_user_meta_data->>'avatar_url'
  )
  ON CONFLICT (id) DO UPDATE SET
    email = EXCLUDED.email,
    full_name = EXCLUDED.full_name,
    avatar_url = EXCLUDED.avatar_url;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Drop existing trigger if it exists
DROP TRIGGER IF EXISTS on_auth_user_created ON auth.users;

-- Create the trigger
CREATE TRIGGER on_auth_user_created
  AFTER INSERT ON auth.users
  FOR EACH ROW EXECUTE FUNCTION public.handle_new_user();

-- Step 8: Enable RLS on users table for future security
ALTER TABLE users ENABLE ROW LEVEL SECURITY;

-- Create basic RLS policies
CREATE POLICY "Users can view own profile" ON users
  FOR SELECT USING (auth.uid() = id);

CREATE POLICY "Users can update own profile" ON users  
  FOR UPDATE USING (auth.uid() = id);

-- Admin/system can insert users (for the trigger)
CREATE POLICY "System can insert users" ON users
  FOR INSERT WITH CHECK (true);

-- Clean up
DROP TABLE IF EXISTS user_id_mapping;