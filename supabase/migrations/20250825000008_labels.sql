-- Migration 008: Create labels table for GitHub-style labels (PostgreSQL)
-- Labels are user-owned objects with name, description, and color that can be attached to consumptions and items
CREATE TABLE IF NOT EXISTS public.labels (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES public.profiles(id) ON DELETE CASCADE,
    name text NOT NULL,
    description text,
    color text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT labels_description_len CHECK (description IS NULL OR length(description) <= 250),
    CONSTRAINT labels_color_hex CHECK (color ~* '^#?[0-9a-f]{6}$'),
    CONSTRAINT labels_name_format CHECK (
        length(name) <= 39 AND
        name ~ '^[a-zA-Z0-9]([a-zA-Z0-9]|-(?=[a-zA-Z0-9]))*$'
    )
);

-- Unique constraint: case-insensitive name uniqueness per user
CREATE UNIQUE INDEX IF NOT EXISTS ux_labels_user_name ON public.labels (user_id, lower(name));

-- Index for efficient lookups
CREATE INDEX IF NOT EXISTS idx_labels_user_id ON public.labels(user_id);
CREATE INDEX IF NOT EXISTS idx_labels_user_created ON public.labels(user_id, created_at);

-- Updated_at trigger (reuse existing function)
CREATE TRIGGER set_labels_updated_at
    BEFORE UPDATE ON public.labels
    FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();

-- Enable RLS
ALTER TABLE public.labels ENABLE ROW LEVEL SECURITY;

-- RLS policies: users can only CRUD their own labels
CREATE POLICY "Users can view own labels" ON public.labels
    FOR SELECT USING (auth.uid() = user_id);
    
CREATE POLICY "Users can insert own labels" ON public.labels
    FOR INSERT WITH CHECK (auth.uid() = user_id);
    
CREATE POLICY "Users can update own labels" ON public.labels
    FOR UPDATE USING (auth.uid() = user_id);
    
CREATE POLICY "Users can delete own labels" ON public.labels
    FOR DELETE USING (auth.uid() = user_id);

-- Enforce per-user label limit = 100
CREATE OR REPLACE FUNCTION public.labels_limit_enforce()
RETURNS trigger AS $$
BEGIN
    IF (SELECT COUNT(*) FROM public.labels WHERE user_id = NEW.user_id) >= 100 THEN
        RAISE EXCEPTION 'label_limit_exceeded';
    END IF;
    RETURN NEW;
END; 
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Apply limit trigger on insert
CREATE TRIGGER labels_limit_before_insert
    BEFORE INSERT ON public.labels
    FOR EACH ROW EXECUTE FUNCTION public.labels_limit_enforce();
