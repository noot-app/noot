-- Migration 001: Initial schema baseline using auth.users.id directly as primary key (PostgreSQL)
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY, -- This will be auth.users.id directly
    handle TEXT NOT NULL UNIQUE,
    full_name TEXT DEFAULT NULL,
    email TEXT NOT NULL UNIQUE,
    subscription_tier TEXT NOT NULL DEFAULT 'free',
    active_goal_name TEXT DEFAULT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    avatar_url TEXT DEFAULT NULL,
    
    -- Add foreign key constraint to auth.users
    CONSTRAINT fk_users_auth FOREIGN KEY (id) REFERENCES auth.users(id) ON DELETE CASCADE DEFERRABLE INITIALLY DEFERRED
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_handle ON users(handle);
CREATE INDEX IF NOT EXISTS idx_users_active_goal ON users(active_goal_name);

-- Create trigger for automatic user creation when auth users are created
CREATE OR REPLACE FUNCTION public.handle_new_user()
RETURNS trigger AS $$
BEGIN
  -- Set search_path to empty string for security
  SET search_path = '';
  
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

-- Create the trigger
CREATE TRIGGER on_auth_user_created
  AFTER INSERT ON auth.users
  FOR EACH ROW EXECUTE FUNCTION public.handle_new_user();

-- Enable RLS on users table for security
ALTER TABLE users ENABLE ROW LEVEL SECURITY;

-- Create RLS policies
CREATE POLICY "Users can view own profile" ON users
  FOR SELECT USING (auth.uid() = id);

-- Users can only update certain profile fields (NOT subscription_tier)
CREATE POLICY "Users can update own profile" ON users  
  FOR UPDATE USING (auth.uid() = id)
  WITH CHECK (
    auth.uid() = id AND
    -- Prevent users from modifying subscription_tier or id
    subscription_tier = (SELECT subscription_tier FROM users WHERE id = auth.uid()) AND
    id = auth.uid()
  );

-- Admin/system can insert users (for the trigger)
CREATE POLICY "System can insert users" ON users
  FOR INSERT WITH CHECK (true);
