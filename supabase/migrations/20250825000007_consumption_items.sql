-- Migration 007: Create consumption_items table for meal ingredient breakdowns (PostgreSQL)
-- This table enables historic viewing of individual items within each consumption,
-- preserving nutrition snapshots at time of logging for permanent reference.
create table if not exists public.consumption_items (
  id uuid primary key default gen_random_uuid(),
  consumption_id uuid not null references public.consumptions(id) on delete cascade,
  item_id uuid null references public.items(id) on delete set null,
  name text not null,
  brand text not null default '',
  grams real not null default 0,
  user_quantity real null,
  user_unit text null,
  note text null constraint consumption_items_note_length check (length(note) <= 1000),
  -- Snapshot nutrition for THIS SERVING
  calories real not null default 0,
  protein_g real not null default 0,
  total_fat_g real not null default 0,
  saturated_fat_g real not null default 0,
  trans_fat_g real not null default 0,
  cholesterol_mg real not null default 0,
  sodium_mg real not null default 0,
  total_carbs_g real not null default 0,
  dietary_fiber_g real not null default 0,
  total_sugars_g real not null default 0,
  added_sugars_g real not null default 0,
  vitamin_a_mcg real not null default 0,
  vitamin_c_mg real not null default 0,
  vitamin_d_mcg real not null default 0,
  vitamin_e_mg real not null default 0,
  vitamin_k_mcg real not null default 0,
  thiamine_mg real not null default 0,
  riboflavin_mg real not null default 0,
  niacin_mg real not null default 0,
  vitamin_b6_mg real not null default 0,
  folate_mcg real not null default 0,
  vitamin_b12_mcg real not null default 0,
  biotin_mcg real not null default 0,
  pantothenic_acid_mg real not null default 0,
  choline_mg real not null default 0,
  calcium_mg real not null default 0,
  iron_mg real not null default 0,
  magnesium_mg real not null default 0,
  phosphorus_mg real not null default 0,
  potassium_mg real not null default 0,
  zinc_mg real not null default 0,
  copper_mg real not null default 0,
  manganese_mg real not null default 0,
  selenium_mcg real not null default 0,
  iodine_mcg real not null default 0,
  molybdenum_mcg real not null default 0,
  chromium_mcg real not null default 0,
  fluoride_mg real not null default 0,
  chloride_mg real not null default 0,
  omega3_ala_g real not null default 0,
  omega3_epa_g real not null default 0,
  omega3_dha_g real not null default 0,
  omega6_g real not null default 0,
  creatine_mg real not null default 0,
  caffeine_mg real not null default 0,
  alcohol_g real not null default 0,
  polyunsaturated_fat_g real not null default 0,
  monounsaturated_fat_g real not null default 0,
  -- Ingredient and URL data for historical tracking and correlation analysis
  ingredients JSONB,
  url TEXT,
  created_at timestamp with time zone not null default now(),
  updated_at timestamp with time zone not null default now()
);

create index if not exists idx_consumption_items_consumption_id on public.consumption_items(consumption_id);
create index if not exists idx_consumption_items_item_id on public.consumption_items(item_id);
create index if not exists idx_consumption_items_created on public.consumption_items(consumption_id, created_at);
create index if not exists idx_consumption_items_ingredients on public.consumption_items using gin (ingredients);
create index if not exists idx_consumption_items_url on public.consumption_items(url) where url is not null;

-- Create trigger for updated_at column (no notice if trigger doesn't exist)
do $$ 
begin
  -- Drop trigger if it exists
  if exists (select 1 from pg_trigger where tgname = 'set_consumption_items_updated_at') then
    drop trigger set_consumption_items_updated_at on public.consumption_items;
  end if;
  
  -- Create the trigger
  create trigger set_consumption_items_updated_at
    before update on public.consumption_items
    for each row execute function public.update_updated_at_column();
end $$;

alter table public.consumption_items enable row level security;
create policy "Users can view own consumption_items"
on public.consumption_items
for select using (
  exists (
    select 1 from public.consumptions c
    where c.id = consumption_id and c.user_id = auth.uid()
  )
);
create policy "Users can insert consumption_items into own consumptions"
on public.consumption_items
for insert with check (
  exists (
    select 1 from public.consumptions c
    where c.id = consumption_id and c.user_id = auth.uid()
  )
);
create policy "Users can update own consumption_items"
on public.consumption_items
for update using (
  exists (
    select 1 from public.consumptions c
    where c.id = consumption_id and c.user_id = auth.uid()
  )
);
create policy "Users can delete own consumption_items"
on public.consumption_items
for delete using (
  exists (
    select 1 from public.consumptions c
    where c.id = consumption_id and c.user_id = auth.uid()
  )
);
