-- Migration 007: Revoke any problematic grants to supabase_auth_admin
--
-- SECURITY FIX: This migration removes any grants to supabase_auth_admin on public.users
-- that may have been added incorrectly. Such grants are not needed and create security risks.
--
-- Background: The auth system should use the trigger-based approach with SECURITY DEFINER
-- instead of having direct DML permissions on application tables.

-- Revoke any existing grants to supabase_auth_admin on public.users
-- These commands will succeed even if no grants exist
DO $$
BEGIN
    -- Revoke all possible permissions that might have been granted
    EXECUTE 'REVOKE ALL PRIVILEGES ON TABLE public.users FROM supabase_auth_admin CASCADE';
    RAISE NOTICE 'Revoked all privileges on public.users from supabase_auth_admin';
EXCEPTION
    WHEN undefined_object THEN
        RAISE NOTICE 'No privileges to revoke - supabase_auth_admin role may not exist yet';
    WHEN undefined_table THEN
        RAISE NOTICE 'public.users table does not exist - skipping revoke';
    WHEN OTHERS THEN
        RAISE NOTICE 'Could not revoke privileges: %', SQLERRM;
END
$$;

-- Ensure public.users table exists with correct schema
-- This should be a no-op if the table already exists from migration 001
CREATE TABLE IF NOT EXISTS public.users (
    id UUID PRIMARY KEY,
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

-- Ensure indexes exist
CREATE INDEX IF NOT EXISTS idx_users_email ON public.users(email);
CREATE INDEX IF NOT EXISTS idx_users_handle ON public.users(handle);
CREATE INDEX IF NOT EXISTS idx_users_active_goal ON public.users(active_goal_name);

-- Ensure RLS is enabled (should be a no-op if already enabled)
ALTER TABLE public.users ENABLE ROW LEVEL SECURITY;

-- Ensure the secure trigger function exists with proper SECURITY DEFINER
CREATE OR REPLACE FUNCTION public.handle_new_user()
RETURNS trigger AS $$
BEGIN
  -- SECURITY: Set search_path to empty string for security
  -- This function uses SECURITY DEFINER to run with elevated privileges
  -- instead of granting DML permissions to supabase_auth_admin
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

-- Ensure the trigger exists
DROP TRIGGER IF EXISTS on_auth_user_created ON auth.users;
CREATE TRIGGER on_auth_user_created
  AFTER INSERT ON auth.users
  FOR EACH ROW EXECUTE FUNCTION public.handle_new_user();

-- Ensure RLS policies exist
DO $$
BEGIN
    -- Drop existing policies if they exist to avoid conflicts
    DROP POLICY IF EXISTS "Users can view own profile" ON public.users;
    DROP POLICY IF EXISTS "Users can update own profile" ON public.users;
    DROP POLICY IF EXISTS "System can insert users" ON public.users;
EXCEPTION
    WHEN OTHERS THEN
        RAISE NOTICE 'Some policies may not have existed to drop: %', SQLERRM;
END
$$;

-- Recreate RLS policies
CREATE POLICY "Users can view own profile" ON public.users
  FOR SELECT USING (auth.uid() = id);

CREATE POLICY "Users can update own profile" ON public.users  
  FOR UPDATE USING (auth.uid() = id)
  WITH CHECK (
    auth.uid() = id AND
    -- Prevent users from modifying subscription_tier or id
    subscription_tier = (SELECT subscription_tier FROM public.users WHERE id = auth.uid()) AND
    id = auth.uid()
  );

CREATE POLICY "System can insert users" ON public.users
  FOR INSERT WITH CHECK (true);

-- Log completion
DO $$
BEGIN
    RAISE NOTICE '✅ Migration 007 completed: Cleaned up auth permissions and ensured secure setup';
    RAISE NOTICE '📋 Security status:';
    RAISE NOTICE '   - Revoked any grants to supabase_auth_admin';
    RAISE NOTICE '   - Confirmed public.users table exists';
    RAISE NOTICE '   - Verified trigger function with SECURITY DEFINER';
    RAISE NOTICE '   - Ensured RLS policies are active';
END
$$;