-- Supabase seed file for local development
-- This file is automatically run after migrations when using `supabase db reset`

-- Create test users using Supabase's auth functions
-- These users will be created in auth.users and can login immediately

-- Insert test users directly into auth.users (simpler approach)
-- Password is hashed version of "password123"
INSERT INTO auth.users (
    instance_id,
    id,
    aud,
    role,
    email,
    encrypted_password,
    email_confirmed_at,
    confirmation_sent_at,
    last_sign_in_at,
    raw_app_meta_data,
    raw_user_meta_data,
    created_at,
    updated_at,
    confirmation_token,
    email_change,
    email_change_token_new,
    recovery_token
) VALUES
-- monalisa@birki.io (pro user) 
(
    '00000000-0000-0000-0000-000000000000'::uuid,
    'a1b2c3d4-e5f6-7890-abcd-ef1234567890'::uuid,
    'authenticated',
    'authenticated', 
    'monalisa@birki.io',
    '$2a$10$jPUaHh/VmsO8EZFYnCeTIupjE86hdgJqZoaXglwhN5nzsPvBd4xWy',
    NOW(),
    NOW(),
    NOW(),
    '{"provider": "email", "providers": ["email"]}',
    '{"handle": "monalisa", "full_name": "Mona Lisa"}',
    NOW(),
    NOW(),
    '',
    '',
    '',
    ''
),
-- alice@birki.io (free user)
(
    '00000000-0000-0000-0000-000000000000'::uuid,
    'b2c3d4e5-f6a7-8901-bcde-f23456789abc'::uuid,
    'authenticated', 
    'authenticated',
    'alice@birki.io',
    '$2a$10$jPUaHh/VmsO8EZFYnCeTIupjE86hdgJqZoaXglwhN5nzsPvBd4xWy',
    NOW(),
    NOW(),
    NOW(),
    '{"provider": "email", "providers": ["email"]}',
    '{"handle": "alice", "full_name": "Alice Smith"}',
    NOW(),
    NOW(),
    '',
    '',
    '',
    ''
);

-- Insert corresponding identities
INSERT INTO auth.identities (
    provider_id,
    user_id,
    provider,
    identity_data,
    created_at,
    updated_at
) VALUES 
(
    'monalisa@birki.io',
    'a1b2c3d4-e5f6-7890-abcd-ef1234567890'::uuid,
    'email',
    '{"sub": "a1b2c3d4-e5f6-7890-abcd-ef1234567890", "email": "monalisa@birki.io", "email_verified": true}',
    NOW(),
    NOW()
),
(
    'alice@birki.io',
    'b2c3d4e5-f6a7-8901-bcde-f23456789abc'::uuid,
    'email',
    '{"sub": "b2c3d4e5-f6a7-8901-bcde-f23456789abc", "email": "alice@birki.io", "email_verified": true}',
    NOW(),
    NOW()
);

-- Update subscription tiers for test users
UPDATE public.profiles 
SET subscription_tier = 'pro' 
WHERE email = 'monalisa@birki.io';

-- Note: Alice remains 'free' tier as set by the trigger
-- Note: Application users will be created automatically by the authentication trigger
-- when these auth.users are inserted, so they should now appear in public.profiles table.
