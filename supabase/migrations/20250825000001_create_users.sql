-- Migration 001: Complete user profiles setup following Supabase best practices
-- This migration creates the profiles table, trigger function, and RLS policies all in one place

-- 1. Create the profiles table
create table if not exists public.profiles (
  id uuid primary key references auth.users on delete cascade,
  handle text unique,
  full_name text,
  email text not null unique,
  subscription_tier text not null default 'free',
  active_goal_name text,
  avatar_url text,
  created_at timestamp with time zone default now(),
  updated_at timestamp with time zone default now()
);

-- Create indexes for performance
create index if not exists idx_profiles_email on profiles(email);
create index if not exists idx_profiles_handle on profiles(handle);
create index if not exists idx_profiles_active_goal on profiles(active_goal_name);

-- 2. Keep updated_at current
create or replace function public.update_updated_at_column()
returns trigger as $$
begin
  new.updated_at = now();
  return new;
end;
$$ language plpgsql;

create trigger set_profiles_updated_at
before update on public.profiles
for each row
execute function public.update_updated_at_column();

-- 3. Create function to insert profile when user signs up
create or replace function public.handle_new_user()
returns trigger
language plpgsql
security definer set search_path = ''
as $$
begin
  insert into public.profiles (id, handle, full_name, email, subscription_tier, avatar_url)
  values (
    new.id,
    coalesce(new.raw_user_meta_data->>'handle', 'user_' || substring(new.id::text, 1, 8)),
    new.raw_user_meta_data->>'full_name',
    new.email,
    'free', -- Default tier
    new.raw_user_meta_data->>'avatar_url'
  );
  return new;
end;
$$;

-- Attach trigger to auth.users table
create trigger on_auth_user_created
after insert on auth.users
for each row execute procedure public.handle_new_user();

-- 4. Enable row level security
alter table public.profiles enable row level security;

-- 5. RLS Policies

-- Each user can view their own profile
create policy "Users can view their own profile"
on public.profiles
for select
using (auth.uid() = id);

-- Each user can update their own profile (but not subscription_tier)
create policy "Users can update their own profile"
on public.profiles
for update
using (auth.uid() = id)
with check (
  auth.uid() = id and
  -- Prevent users from modifying subscription_tier or id
  subscription_tier = (select subscription_tier from profiles where id = auth.uid()) and
  id = auth.uid()
);

-- Prevent client-side inserts (only trigger inserts allowed)
create policy "No client inserts into profiles"
on public.profiles
for insert
with check (false);
