-- Migration 010: Seed default labels for new users
-- Update the handle_new_user function to create default labels when a user signs up

-- Replace the handle_new_user function to include default label seeding
create or replace function public.handle_new_user()
returns trigger
language plpgsql
security definer
set search_path = public
as $$
declare
  default_handle text;
begin
  -- Generate a valid GitHub-style handle from user ID
  default_handle := 'user' || replace(substring(new.id::text, 1, 8), '-', '');
  
  -- Insert the user profile
  insert into public.profiles (id, handle, full_name, email, subscription_tier, avatar_url)
  values (
    new.id,
    coalesce(new.raw_user_meta_data->>'handle', default_handle),
    new.raw_user_meta_data->>'full_name',
    new.email,
    'free', -- Default tier
    new.raw_user_meta_data->>'avatar_url'
  );
  
  -- Insert default labels for the new user
  insert into public.labels (user_id, name, description, color) values
    (new.id, 'breakfast', '', 'FFD700'),
    (new.id, 'lunch', '', '74B986'),
    (new.id, 'dinner', '', '1E90FF'),
    (new.id, 'snack', 'A small snack or light bite', '9B59B6'),
    (new.id, 'drink', 'A beverage', '9CA3AF'),
    (new.id, 'trigger-food', 'The consumption contained a known trigger food', 'FF0000'),
    (new.id, 'high-protein', 'High protein foods', '2DD4BF'),
    (new.id, 'meal-prep', 'Pre-prepared meals', 'A8E6CF'),
    (new.id, 'restaurant', 'Restaurant or takeout meal', 'F59E0B');
  
  return new;
end;
$$;