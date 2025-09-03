-- Migration 011: Create event_types table for user-defined event categories (PostgreSQL)
-- Event types are user-owned templates that define default values for creating events
CREATE TABLE IF NOT EXISTS public.event_types (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES public.profiles(id) ON DELETE CASCADE,
    name text NOT NULL,
    description text,
    default_name text,
    color text NOT NULL,
    icon text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT event_types_name_format CHECK (
        length(name) <= 39 AND
        name ~ '^[a-zA-Z0-9]([a-zA-Z0-9 ]|-(?=[a-zA-Z0-9]))*$'
    ),
    CONSTRAINT event_types_description_len CHECK (description IS NULL OR length(description) <= 250),
    CONSTRAINT event_types_default_name_len CHECK (default_name IS NULL OR length(default_name) <= 100),
    CONSTRAINT event_types_color_hex CHECK (color ~* '^#?[0-9a-f]{6}$'),
    CONSTRAINT event_types_icon_len CHECK (icon IS NULL OR length(icon) <= 50)
);

-- Unique constraint: case-insensitive name uniqueness per user
CREATE UNIQUE INDEX IF NOT EXISTS ux_event_types_user_name ON public.event_types (user_id, lower(name));

-- Index for efficient lookups
CREATE INDEX IF NOT EXISTS idx_event_types_user_id ON public.event_types(user_id);
CREATE INDEX IF NOT EXISTS idx_event_types_user_created ON public.event_types(user_id, created_at);

-- Updated_at trigger (reuse existing function)
CREATE TRIGGER set_event_types_updated_at
    BEFORE UPDATE ON public.event_types
    FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();

-- Enable RLS
ALTER TABLE public.event_types ENABLE ROW LEVEL SECURITY;

-- RLS policies: users can only CRUD their own event types
CREATE POLICY "Users can view own event types" ON public.event_types
    FOR SELECT USING (auth.uid() = user_id);
    
CREATE POLICY "Users can insert own event types" ON public.event_types
    FOR INSERT WITH CHECK (auth.uid() = user_id);
    
CREATE POLICY "Users can update own event types" ON public.event_types
    FOR UPDATE USING (auth.uid() = user_id);
    
CREATE POLICY "Users can delete own event types" ON public.event_types
    FOR DELETE USING (auth.uid() = user_id);

-- Enforce per-user event type limit = 50
CREATE OR REPLACE FUNCTION public.event_types_limit_enforce()
RETURNS trigger AS $$
BEGIN
    IF (SELECT COUNT(*) FROM public.event_types WHERE user_id = NEW.user_id) >= 50 THEN
        RAISE EXCEPTION 'event_type_limit_exceeded';
    END IF;
    RETURN NEW;
END; 
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Apply limit trigger on insert
CREATE TRIGGER event_types_limit_before_insert
    BEFORE INSERT ON public.event_types
    FOR EACH ROW EXECUTE FUNCTION public.event_types_limit_enforce();

-- Insert default event types for all existing users
INSERT INTO public.event_types (user_id, name, description, default_name, color, icon)
SELECT 
    p.id,
    'Symptom',
    'Physical symptoms and discomfort',
    'Symptom',
    'FF0000',
    '🤒'
FROM public.profiles p
WHERE NOT EXISTS (
    SELECT 1 FROM public.event_types et WHERE et.user_id = p.id AND lower(et.name) = 'symptom'
);

INSERT INTO public.event_types (user_id, name, description, default_name, color, icon)
SELECT 
    p.id,
    'Activity',
    'Physical activities and exercise',
    'Activity',
    'FF8C00',
    '🏃'
FROM public.profiles p
WHERE NOT EXISTS (
    SELECT 1 FROM public.event_types et WHERE et.user_id = p.id AND lower(et.name) = 'activity'
);

INSERT INTO public.event_types (user_id, name, description, default_name, color, icon)
SELECT 
    p.id,
    'Measurement',
    'Health measurements and vitals',
    'Measurement',
    'FFD700',
    '📊'
FROM public.profiles p
WHERE NOT EXISTS (
    SELECT 1 FROM public.event_types et WHERE et.user_id = p.id AND lower(et.name) = 'measurement'
);

INSERT INTO public.event_types (user_id, name, description, default_name, color, icon)
SELECT 
    p.id,
    'Medication',
    'Medications and supplements',
    'Medication',
    '1E90FF',
    '💊'
FROM public.profiles p
WHERE NOT EXISTS (
    SELECT 1 FROM public.event_types et WHERE et.user_id = p.id AND lower(et.name) = 'medication'
);

INSERT INTO public.event_types (user_id, name, description, default_name, color, icon)
SELECT 
    p.id,
    'Sleep',
    'Sleep patterns and quality',
    'Sleep',
    '9B59B6',
    '😴'
FROM public.profiles p
WHERE NOT EXISTS (
    SELECT 1 FROM public.event_types et WHERE et.user_id = p.id AND lower(et.name) = 'sleep'
);

INSERT INTO public.event_types (user_id, name, description, default_name, color, icon)
SELECT 
    p.id,
    'Mood',
    'Emotional state and mental health',
    'Mood',
    '2DD4BF',
    '😊'
FROM public.profiles p
WHERE NOT EXISTS (
    SELECT 1 FROM public.event_types et WHERE et.user_id = p.id AND lower(et.name) = 'mood'
);

INSERT INTO public.event_types (user_id, name, description, default_name, color, icon)
SELECT 
    p.id,
    'Other',
    'Miscellaneous health events',
    'Other',
    '9CA3AF',
    '📝'
FROM public.profiles p
WHERE NOT EXISTS (
    SELECT 1 FROM public.event_types et WHERE et.user_id = p.id AND lower(et.name) = 'other'
);

-- Function to create default event types for new users
CREATE OR REPLACE FUNCTION public.create_default_event_types_for_user()
RETURNS trigger AS $$
BEGIN
    -- Insert default event types for the new user
    INSERT INTO public.event_types (user_id, name, description, default_name, color, icon) VALUES
        (NEW.id, 'Symptom', 'Physical symptoms and discomfort', 'Symptom', 'FF0000', '🤒'),
        (NEW.id, 'Activity', 'Physical activities and exercise', 'Activity', 'FF8C00', '🏃'),
        (NEW.id, 'Measurement', 'Health measurements and vitals', 'Measurement', 'FFD700', '📊'),
        (NEW.id, 'Medication', 'Medications and supplements', 'Medication', '1E90FF', '💊'),
        (NEW.id, 'Sleep', 'Sleep patterns and quality', 'Sleep', '9B59B6', '😴'),
        (NEW.id, 'Mood', 'Emotional state and mental health', 'Mood', '2DD4BF', '😊'),
        (NEW.id, 'Other', 'Miscellaneous health events', 'Other', '9CA3AF', '📝');
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Trigger to create default event types when a new user profile is created
CREATE TRIGGER create_default_event_types_on_profile_insert
    AFTER INSERT ON public.profiles
    FOR EACH ROW EXECUTE FUNCTION public.create_default_event_types_for_user();
