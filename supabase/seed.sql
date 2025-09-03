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

-- ================================================================================================
-- Dev seed: API keys
-- Create a test API key for the pro user (monalisa)
-- ================================================================================================
-- $ curl -H "X-API-Key: noot_3eb35a4c_36cd5f4d802c8ab41d3d7e4f6b7302a09c61067c" http://localhost:3001/api/v1/consumptions
-- Insert API key for monalisa (pro user) with read_write permissions and no expiration
INSERT INTO api_keys (
    id,
    user_id,
    name,
    prefix,
    hash,
    scope,
    expires_at,
    created_at
) VALUES (
    '550e8400-e29b-41d4-a716-446655440001'::uuid,
    'a1b2c3d4-e5f6-7890-abcd-ef1234567890'::uuid, -- monalisa's user_id
    'dev-key',
    'noot_3eb35a4c',
    '$2a$12$/NUFGftTB7DAY6I2hPIIjufqbbj9vEn7PRK3ELLcQjDoJ.wtS9YTO',
    'read_write',
    NULL, -- no expiration
    NOW()
);

-- Show API key creation result
SELECT 
    'API Key Created' as status,
    ak.name,
    ak.prefix,
    ak.scope,
    CASE WHEN ak.expires_at IS NULL THEN 'Never' ELSE ak.expires_at::text END as expires,
    p.email as user_email
FROM api_keys ak
JOIN profiles p ON p.id = ak.user_id
WHERE p.email = 'monalisa@birki.io';

-- ================================================================================================
-- Dev seed: consumptions
-- Generate realistic consumption data for both test users with varied meal patterns across time
-- ================================================================================================

-- Use reproducible randomness for consistent dev data
SELECT setseed(0.42);

-- Delete existing consumptions for idempotent reseeding
DELETE FROM consumptions WHERE user_id IN (
    'a1b2c3d4-e5f6-7890-abcd-ef1234567890', -- monalisa@birki.io
    'b2c3d4e5-f6a7-8901-bcde-f23456789abc'  -- alice@birki.io
);

-- Insert consumption data using a CTE with meal templates and date generation
WITH meal_templates AS (
    SELECT * FROM (VALUES
        -- Breakfast options
        ('breakfast', 'Breakfast: oatmeal with blueberries and almond milk', 320, 8.5, 6.2, 58.3, 8.1, 95, 0),
        ('breakfast', 'Morning latte with whole milk and a banana', 350, 14.3, 9.3, 54.7, 3.1, 138, 76),
        ('breakfast', 'Scrambled eggs with avocado toast and orange juice', 485, 18.2, 22.1, 45.6, 12.3, 284, 0),
        ('breakfast', 'Greek yogurt with granola and fresh strawberries', 380, 16.8, 11.4, 52.9, 6.7, 125, 0),
        ('breakfast', 'Whole grain cereal with 2% milk and sliced banana', 295, 12.1, 4.8, 54.2, 7.3, 189, 0),
        ('breakfast', 'Coffee with cream and a blueberry muffin from Starbucks', 425, 6.9, 16.8, 65.4, 2.8, 392, 142),
        ('breakfast', 'Smoothie bowl with spinach, mango, protein powder, and chia seeds', 340, 28.4, 8.1, 38.7, 11.2, 156, 0),
        
        -- Lunch options  
        ('lunch', 'Lunch: grilled chicken salad with mixed greens and balsamic dressing', 385, 32.4, 18.6, 22.1, 6.8, 456, 0),
        ('lunch', 'Turkey and cheese sandwich on sourdough with an apple', 520, 28.9, 18.2, 65.8, 8.4, 892, 0),
        ('lunch', 'Chipotle burrito bowl with chicken, brown rice, and guacamole', 680, 42.1, 28.3, 58.7, 12.9, 1240, 0),
        ('lunch', 'Trader Joes Mediterranean wrap with hummus and vegetables', 390, 14.2, 16.8, 48.3, 9.1, 684, 0),
        ('lunch', 'Leftover pasta with marinara sauce and a side salad', 445, 16.7, 12.4, 68.9, 7.2, 758, 0),
        ('lunch', 'Quinoa bowl with roasted vegetables and tahini dressing', 425, 15.8, 19.2, 52.6, 8.9, 345, 0),
        ('lunch', 'Poke bowl with salmon, brown rice, and edamame from Sweetgreen', 520, 28.3, 16.7, 62.4, 6.8, 892, 0),
        
        -- Dinner options
        ('dinner', 'Dinner: baked salmon with roasted Brussels sprouts and quinoa', 585, 38.2, 24.1, 48.7, 9.8, 432, 0),
        ('dinner', 'Spaghetti with meat sauce and garlic bread', 720, 32.8, 22.4, 89.6, 6.3, 1156, 0),
        ('dinner', 'Grilled chicken breast with sweet potato and steamed broccoli', 465, 42.3, 8.9, 48.2, 8.7, 198, 0),
        ('dinner', 'Takeout pad thai with shrimp from local Thai restaurant', 650, 28.4, 18.9, 89.3, 4.2, 1890, 0),
        ('dinner', 'Homemade pizza with mozzarella, tomatoes, and basil', 580, 24.6, 22.8, 68.4, 4.9, 1024, 0),
        ('dinner', 'Beef stir-fry with mixed vegetables and brown rice', 520, 35.1, 16.2, 52.8, 5.4, 896, 0),
        ('dinner', 'Fish tacos with cabbage slaw and black beans', 485, 28.7, 14.6, 58.9, 12.1, 742, 0),
        
        -- Snack options
        ('snack', 'Afternoon snack: handful of almonds and an apple', 285, 8.4, 18.2, 32.1, 8.9, 2, 0),
        ('snack', 'Greek yogurt with honey and walnuts', 245, 15.6, 12.8, 18.4, 2.1, 68, 0),
        ('snack', 'Clif Bar energy bar and sparkling water', 250, 9.0, 5.0, 45.0, 5.0, 150, 0),
        ('snack', 'Hummus with baby carrots and cucumber slices', 180, 6.8, 8.4, 18.7, 6.2, 284, 0),
        ('snack', 'Dark chocolate square and green tea', 85, 1.2, 5.8, 8.4, 2.1, 2, 25),
        ('snack', 'Trail mix with dried fruit and nuts', 320, 8.9, 18.7, 28.4, 4.8, 156, 0),
        
        -- Coffee/drink options with caffeine
        ('drink', 'Large iced coffee with oat milk from Blue Bottle', 65, 2.1, 2.8, 8.4, 1.2, 15, 185),
        ('drink', 'Cappuccino with whole milk', 150, 8.1, 8.2, 12.3, 0, 95, 154),
        ('drink', 'Cold brew coffee with almond milk', 25, 1.2, 1.8, 2.1, 0.5, 8, 200),
        ('drink', 'Matcha latte with coconut milk', 180, 4.2, 6.8, 24.1, 2.1, 45, 70),
        ('drink', 'Energy drink and protein bar', 380, 22.4, 8.9, 42.6, 3.2, 245, 160)
    ) AS t(meal_type, transcript, calories, protein_g, fat_g, carbs_g, fiber_g, sodium_mg, caffeine_mg)
),
date_schedules AS (
    -- Generate date offsets for realistic meal distribution
    SELECT user_id, day_offset, meal_count FROM (VALUES
        -- Today: 3-4 meals per user
        ('a1b2c3d4-e5f6-7890-abcd-ef1234567890', 0, 4), -- monalisa today
        ('b2c3d4e5-f6a7-8901-bcde-f23456789abc', 0, 3), -- alice today
        
        -- Yesterday: 2-3 meals per user  
        ('a1b2c3d4-e5f6-7890-abcd-ef1234567890', 1, 3), -- monalisa yesterday
        ('b2c3d4e5-f6a7-8901-bcde-f23456789abc', 1, 2), -- alice yesterday
        
        -- Last 7 days: distributed meals
        ('a1b2c3d4-e5f6-7890-abcd-ef1234567890', 2, 2),
        ('a1b2c3d4-e5f6-7890-abcd-ef1234567890', 3, 3),
        ('a1b2c3d4-e5f6-7890-abcd-ef1234567890', 4, 2),
        ('a1b2c3d4-e5f6-7890-abcd-ef1234567890', 5, 4),
        ('a1b2c3d4-e5f6-7890-abcd-ef1234567890', 6, 3),
        ('b2c3d4e5-f6a7-8901-bcde-f23456789abc', 2, 3),
        ('b2c3d4e5-f6a7-8901-bcde-f23456789abc', 3, 2),
        ('b2c3d4e5-f6a7-8901-bcde-f23456789abc', 4, 3),
        ('b2c3d4e5-f6a7-8901-bcde-f23456789abc', 5, 2),
        ('b2c3d4e5-f6a7-8901-bcde-f23456789abc', 6, 4),
        
        -- Weeks 2-8: scattered meals (remaining ~15 per user)
        ('a1b2c3d4-e5f6-7890-abcd-ef1234567890', 8, 2),
        ('a1b2c3d4-e5f6-7890-abcd-ef1234567890', 12, 3),
        ('a1b2c3d4-e5f6-7890-abcd-ef1234567890', 15, 2),
        ('a1b2c3d4-e5f6-7890-abcd-ef1234567890', 18, 2),
        ('a1b2c3d4-e5f6-7890-abcd-ef1234567890', 22, 3),
        ('a1b2c3d4-e5f6-7890-abcd-ef1234567890', 26, 2),
        ('a1b2c3d4-e5f6-7890-abcd-ef1234567890', 30, 1),
        ('a1b2c3d4-e5f6-7890-abcd-ef1234567890', 35, 2),
        ('a1b2c3d4-e5f6-7890-abcd-ef1234567890', 42, 2),
        ('a1b2c3d4-e5f6-7890-abcd-ef1234567890', 48, 2),
        ('a1b2c3d4-e5f6-7890-abcd-ef1234567890', 55, 1),
        
        ('b2c3d4e5-f6a7-8901-bcde-f23456789abc', 9, 2),
        ('b2c3d4e5-f6a7-8901-bcde-f23456789abc', 13, 2),
        ('b2c3d4e5-f6a7-8901-bcde-f23456789abc', 16, 3),
        ('b2c3d4e5-f6a7-8901-bcde-f23456789abc', 20, 2),
        ('b2c3d4e5-f6a7-8901-bcde-f23456789abc', 24, 2),
        ('b2c3d4e5-f6a7-8901-bcde-f23456789abc', 28, 3),
        ('b2c3d4e5-f6a7-8901-bcde-f23456789abc', 32, 1),
        ('b2c3d4e5-f6a7-8901-bcde-f23456789abc', 38, 2),
        ('b2c3d4e5-f6a7-8901-bcde-f23456789abc', 44, 2),
        ('b2c3d4e5-f6a7-8901-bcde-f23456789abc', 50, 3),
        ('b2c3d4e5-f6a7-8901-bcde-f23456789abc', 58, 1)
    ) AS schedule(user_id, day_offset, meal_count)
),
consumption_rows AS (
    -- Generate individual consumption rows by expanding meal_count per date
    SELECT 
        ds.user_id,
        ds.day_offset,
        generate_series(1, ds.meal_count) as meal_number
    FROM date_schedules ds
),
consumption_with_meals AS (
    -- Join with random meal templates
    SELECT 
        cr.user_id,
        cr.day_offset,
        cr.meal_number,
        mt.meal_type,
        mt.transcript,
        mt.calories,
        mt.protein_g,
        mt.fat_g,
        mt.carbs_g,
        mt.fiber_g,
        mt.sodium_mg,
        mt.caffeine_mg,
        -- Add row number for deterministic randomness
        row_number() OVER (PARTITION BY cr.user_id ORDER BY cr.day_offset, cr.meal_number) as rn
    FROM consumption_rows cr
    CROSS JOIN meal_templates mt
),
selected_meals AS (
    -- Select one random meal per consumption row
    SELECT DISTINCT ON (user_id, day_offset, meal_number)
        user_id,
        day_offset,
        meal_type,
        transcript,
        calories,
        protein_g,
        fat_g,
        carbs_g,
        fiber_g,
        sodium_mg,
        caffeine_mg
    FROM consumption_with_meals
    ORDER BY user_id, day_offset, meal_number, random()
),
timed_consumptions AS (
    -- Add realistic timestamps based on meal type and day
    SELECT 
        user_id,
        meal_type,
        transcript,
        calories,
        protein_g,
        fat_g,
        carbs_g,
        fiber_g,
        sodium_mg,
        caffeine_mg,
        -- Calculate realistic timestamp for meal type
        CASE 
            WHEN meal_type = 'breakfast' THEN 
                (CURRENT_DATE - day_offset)::timestamp + INTERVAL '7 hours' + (random() * INTERVAL '3 hours')
            WHEN meal_type = 'lunch' THEN 
                (CURRENT_DATE - day_offset)::timestamp + INTERVAL '11.5 hours' + (random() * INTERVAL '2.5 hours')
            WHEN meal_type = 'dinner' THEN 
                (CURRENT_DATE - day_offset)::timestamp + INTERVAL '18 hours' + (random() * INTERVAL '3 hours')
            WHEN meal_type = 'snack' THEN 
                (CURRENT_DATE - day_offset)::timestamp + INTERVAL '15 hours' + (random() * INTERVAL '8 hours')
            WHEN meal_type = 'drink' THEN 
                (CURRENT_DATE - day_offset)::timestamp + INTERVAL '9 hours' + (random() * INTERVAL '10 hours')
            ELSE 
                (CURRENT_DATE - day_offset)::timestamp + INTERVAL '12 hours' + (random() * INTERVAL '8 hours')
        END AS created_at
    FROM selected_meals
)
INSERT INTO consumptions (
    user_id,
    transcript,
    total_calories,
    total_protein_g,
    total_fat_g,
    total_carbs_g,
    dietary_fiber_g,
    total_sodium_mg,
    caffeine_mg,
    created_at
)
SELECT 
    user_id::uuid,
    transcript,
    calories,
    protein_g,
    fat_g,
    carbs_g,
    fiber_g,
    sodium_mg,
    caffeine_mg,
    created_at
FROM timed_consumptions
ORDER BY user_id, created_at;

-- Show seeding results
SELECT 
    p.email,
    COUNT(*) as total_consumptions,
    COUNT(CASE WHEN c.created_at::date = CURRENT_DATE THEN 1 END) as today_meals,
    COUNT(CASE WHEN c.created_at::date = CURRENT_DATE - 1 THEN 1 END) as yesterday_meals,
    COUNT(CASE WHEN c.created_at >= CURRENT_DATE - INTERVAL '7 days' THEN 1 END) as last_week_meals,
    SUM(c.caffeine_mg)::int as total_caffeine_mg
FROM consumptions c 
JOIN profiles p ON p.id = c.user_id 
WHERE p.email IN ('monalisa@birki.io', 'alice@birki.io')
GROUP BY p.email 
ORDER BY p.email;

-- ================================================================================================
-- Dev seed: consumption labels
-- Assign labels to ~70% of Mona's consumptions with intelligent categorization
-- ================================================================================================

-- Get label IDs for Mona (monalisa@birki.io)
WITH mona_labels AS (
    SELECT l.id, l.name 
    FROM labels l 
    JOIN profiles p ON p.id = l.user_id 
    WHERE p.email = 'monalisa@birki.io'
),
mona_consumptions AS (
    SELECT c.id, c.transcript, c.total_protein_g, c.caffeine_mg,
           row_number() OVER (ORDER BY c.created_at) as rn
    FROM consumptions c 
    JOIN profiles p ON p.id = c.user_id 
    WHERE p.email = 'monalisa@birki.io'
),
consumption_labels AS (
    SELECT 
        mc.id as consumption_id,
        ml.id as label_id,
        ml.name as label_name,
        mc.transcript
    FROM mona_consumptions mc
    CROSS JOIN mona_labels ml
    WHERE 
        -- Apply labels based on transcript content and nutritional profile
        (
            -- Meal type labels (applied to ~90% of relevant meals)
            (ml.name = 'breakfast' AND (mc.transcript ILIKE '%breakfast%' OR mc.transcript ILIKE '%morning%' OR mc.transcript ILIKE '%oatmeal%' OR mc.transcript ILIKE '%yogurt%' OR mc.transcript ILIKE '%cereal%') AND mc.rn % 10 < 9)
            OR (ml.name = 'lunch' AND mc.transcript ILIKE '%lunch%' AND mc.rn % 10 < 9)
            OR (ml.name = 'dinner' AND mc.transcript ILIKE '%dinner%' AND mc.rn % 10 < 9)
            OR (ml.name = 'snack' AND (mc.transcript ILIKE '%snack%' OR mc.transcript ILIKE '%handful%' OR mc.transcript ILIKE '%chocolate%' OR mc.transcript ILIKE '%trail mix%' OR mc.transcript ILIKE '%clif bar%') AND mc.rn % 10 < 9)
            OR (ml.name = 'drink' AND (mc.transcript ILIKE '%coffee%' OR mc.transcript ILIKE '%latte%' OR mc.transcript ILIKE '%cappuccino%' OR mc.transcript ILIKE '%tea%') AND mc.rn % 10 < 9)
            
            -- Trigger food labels (applied to all potential triggers)
            OR (ml.name = 'trigger-food' AND (
                mc.transcript ILIKE '%pizza%' OR 
                mc.transcript ILIKE '%onion%' OR 
                mc.transcript ILIKE '%garlic%' OR 
                mc.transcript ILIKE '%tomato%' OR 
                mc.transcript ILIKE '%citrus%' OR 
                mc.transcript ILIKE '%orange%' OR 
                mc.transcript ILIKE '%spicy%' OR 
                mc.transcript ILIKE '%pad thai%' OR 
                mc.transcript ILIKE '%marinara%' OR 
                mc.transcript ILIKE '%chocolate%' OR
                mc.transcript ILIKE '%cheese%'
            ))
            
            -- High protein foods (>25g protein)
            OR (ml.name = 'high-protein' AND mc.total_protein_g > 25 AND mc.rn % 3 < 2)
            
            -- Restaurant meals
            OR (ml.name = 'restaurant' AND (
                mc.transcript ILIKE '%chipotle%' OR 
                mc.transcript ILIKE '%takeout%' OR 
                mc.transcript ILIKE '%restaurant%' OR 
                mc.transcript ILIKE '%thai%' OR 
                mc.transcript ILIKE '%starbucks%' OR
                mc.transcript ILIKE '%sweetgreen%'
            ) AND mc.rn % 4 < 3)
            
            -- Meal prep items (apply to ~30% randomly)
            OR (ml.name = 'meal-prep' AND (
                mc.transcript ILIKE '%leftover%' OR 
                mc.transcript ILIKE '%sheet pan%' OR 
                mc.transcript ILIKE '%homemade%'
            ) AND mc.rn % 5 < 2)
        )
        -- Only label ~70% of Mona's consumptions overall
        AND mc.rn % 10 < 7
)
INSERT INTO consumption_labels (consumption_id, label_id)
SELECT DISTINCT consumption_id, label_id 
FROM consumption_labels
ON CONFLICT DO NOTHING;

-- Show labeling results
SELECT 
    'Label Assignment Summary' as summary,
    COUNT(DISTINCT cl.consumption_id) as labeled_consumptions,
    COUNT(*) as total_label_assignments,
    ROUND(COUNT(DISTINCT cl.consumption_id)::numeric / 
          (SELECT COUNT(*) FROM consumptions c JOIN profiles p ON p.id = c.user_id WHERE p.email = 'monalisa@birki.io')::numeric * 100, 1) as percentage_labeled
FROM consumption_labels cl
JOIN consumptions c ON c.id = cl.consumption_id
JOIN profiles p ON p.id = c.user_id
WHERE p.email = 'monalisa@birki.io';

-- Show label distribution for Mona
SELECT 
    l.name as label_name,
    l.description,
    COUNT(*) as usage_count,
    STRING_AGG(LEFT(c.transcript, 60), '; ' ORDER BY c.created_at DESC) as sample_transcripts
FROM consumption_labels cl
JOIN labels l ON l.id = cl.label_id
JOIN consumptions c ON c.id = cl.consumption_id
JOIN profiles p ON p.id = c.user_id
WHERE p.email = 'monalisa@birki.io'
GROUP BY l.id, l.name, l.description
ORDER BY usage_count DESC, l.name;
