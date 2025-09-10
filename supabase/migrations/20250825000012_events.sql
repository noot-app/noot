-- Migration 012: Create events table for manual health/symptom/activity tracking (PostgreSQL)
-- Events are user-owned objects that can be linked to consumptions/items for correlation analysis
CREATE TABLE IF NOT EXISTS public.events (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES public.profiles(id) ON DELETE CASCADE,
    name text NOT NULL,
    event_type_id uuid REFERENCES public.event_types(id) ON DELETE SET NULL,
    started_at timestamptz NOT NULL,
    ended_at timestamptz,
    level smallint,
    note text,
    color text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT events_name_length CHECK (length(name) <= 100),
    CONSTRAINT events_level_range CHECK (level IS NULL OR (level >= 0 AND level <= 10)),
    CONSTRAINT events_note_length CHECK (note IS NULL OR length(note) <= 1000),
    CONSTRAINT events_color_hex CHECK (color IS NULL OR color ~* '^#?[0-9a-f]{6}$'),
    CONSTRAINT events_time_order CHECK (ended_at IS NULL OR ended_at >= started_at)
);

-- Indexes for efficient lookups
CREATE INDEX IF NOT EXISTS idx_events_user_id ON public.events(user_id);
CREATE INDEX IF NOT EXISTS idx_events_user_started ON public.events(user_id, started_at);
CREATE INDEX IF NOT EXISTS idx_events_user_event_type ON public.events(user_id, event_type_id) WHERE event_type_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_events_user_level ON public.events(user_id, level) WHERE level IS NOT NULL;
-- Add index for time-based queries
CREATE INDEX IF NOT EXISTS idx_events_time_range ON public.events(user_id, started_at, ended_at);

-- Updated_at trigger (reuse existing function)
CREATE TRIGGER set_events_updated_at
    BEFORE UPDATE ON public.events
    FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();

-- Enable RLS
ALTER TABLE public.events ENABLE ROW LEVEL SECURITY;

-- RLS policies: users can only CRUD their own events
CREATE POLICY "Users can view own events" ON public.events
    FOR SELECT USING (auth.uid() = user_id);
    
CREATE POLICY "Users can insert own events" ON public.events
    FOR INSERT WITH CHECK (auth.uid() = user_id);
    
CREATE POLICY "Users can update own events" ON public.events
    FOR UPDATE USING (auth.uid() = user_id);
    
CREATE POLICY "Users can delete own events" ON public.events
    FOR DELETE USING (auth.uid() = user_id);

-- Enforce per-user event limit = 1000 (more generous than labels since these are time-bound entries)
CREATE OR REPLACE FUNCTION public.events_limit_enforce()
RETURNS trigger AS $$
BEGIN
    IF (SELECT COUNT(*) FROM public.events WHERE user_id = NEW.user_id) >= 1000 THEN
        RAISE EXCEPTION 'event_limit_exceeded';
    END IF;
    RETURN NEW;
END; 
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Apply limit trigger on insert
CREATE TRIGGER events_limit_before_insert
    BEFORE INSERT ON public.events
    FOR EACH ROW EXECUTE FUNCTION public.events_limit_enforce();
