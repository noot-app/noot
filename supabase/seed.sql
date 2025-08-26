-- Supabase seed file for local development
-- This file is automatically run after migrations when using `supabase db reset`

-- Create test users using Supabase's auth functions
-- These users will be created in auth.users and can login immediately

-- Create monalisa@birki.io (pro user)
SELECT auth.uid() as current_user_id;

-- Insert test users directly into auth.users (simpler approach)
-- Password is hashed version of "password123"
INSERT INTO auth.users (
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
    'a1b2c3d4-e5f6-7890-abcd-ef1234567890'::uuid,
    'authenticated',
    'authenticated', 
    'monalisa@birki.io',
    '$2a$12$IoqPwkut0FJKRU5hd1Hqd.Us/Vp9KD/OUxFRCZrYCqG5XtaIaNGlG',
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
    'b2c3d4e5-f6a7-8901-bcde-f23456789abc'::uuid,
    'authenticated', 
    'authenticated',
    'alice@birki.io',
    '$2a$12$IoqPwkut0FJKRU5hd1Hqd.Us/Vp9KD/OUxFRCZrYCqG5XtaIaNGlG',
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

-- Note: Application users will be created automatically by the authentication middleware
-- when these users first authenticate, so we don't need to manually create them here.
