-- Migration 013: Create event join tables for label assignments and consumption linking (PostgreSQL)
-- These tables enable many-to-many relationships between events and labels/consumptions/items

-- Join table for event labels (reuse existing labels)
CREATE TABLE IF NOT EXISTS public.event_labels (
    event_id uuid NOT NULL REFERENCES public.events(id) ON DELETE CASCADE,
    label_id uuid NOT NULL REFERENCES public.labels(id) ON DELETE CASCADE,
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (event_id, label_id)
);

-- Indexes for efficient queries
CREATE INDEX IF NOT EXISTS idx_event_labels_label ON public.event_labels(label_id);
CREATE INDEX IF NOT EXISTS idx_event_labels_event ON public.event_labels(event_id);

-- Enable RLS
ALTER TABLE public.event_labels ENABLE ROW LEVEL SECURITY;

-- RLS policies: user must own both the event and the label
CREATE POLICY "Owner can read event_labels" ON public.event_labels
    FOR SELECT USING (
        EXISTS (SELECT 1 FROM public.events e WHERE e.id = event_id AND e.user_id = auth.uid())
        AND EXISTS (SELECT 1 FROM public.labels l WHERE l.id = label_id AND l.user_id = auth.uid())
    );

CREATE POLICY "Owner can insert event_labels" ON public.event_labels
    FOR INSERT WITH CHECK (
        EXISTS (SELECT 1 FROM public.events e WHERE e.id = event_id AND e.user_id = auth.uid())
        AND EXISTS (SELECT 1 FROM public.labels l WHERE l.id = label_id AND l.user_id = auth.uid())
    );

CREATE POLICY "Owner can delete event_labels" ON public.event_labels
    FOR DELETE USING (
        EXISTS (SELECT 1 FROM public.events e WHERE e.id = event_id AND e.user_id = auth.uid())
        AND EXISTS (SELECT 1 FROM public.labels l WHERE l.id = label_id AND l.user_id = auth.uid())
    );

-- Join table for manual event-consumption/item links
CREATE TABLE IF NOT EXISTS public.event_links (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id uuid NOT NULL REFERENCES public.events(id) ON DELETE CASCADE,
    consumption_id uuid REFERENCES public.consumptions(id) ON DELETE CASCADE,
    consumption_item_id uuid REFERENCES public.consumption_items(id) ON DELETE CASCADE,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT event_links_target_check CHECK (
        (consumption_id IS NOT NULL AND consumption_item_id IS NULL) OR
        (consumption_id IS NULL AND consumption_item_id IS NOT NULL)
    )
);

-- Indexes for efficient queries
CREATE INDEX IF NOT EXISTS idx_event_links_event ON public.event_links(event_id);
CREATE INDEX IF NOT EXISTS idx_event_links_consumption ON public.event_links(consumption_id) WHERE consumption_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_event_links_item ON public.event_links(consumption_item_id) WHERE consumption_item_id IS NOT NULL;

-- Unique constraint: prevent duplicate links
CREATE UNIQUE INDEX IF NOT EXISTS ux_event_links_event_consumption ON public.event_links(event_id, consumption_id) WHERE consumption_id IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS ux_event_links_event_item ON public.event_links(event_id, consumption_item_id) WHERE consumption_item_id IS NOT NULL;

-- Enable RLS
ALTER TABLE public.event_links ENABLE ROW LEVEL SECURITY;

-- RLS policies: user must own the event and the linked consumption/item
CREATE POLICY "Owner can read event_links" ON public.event_links
    FOR SELECT USING (
        EXISTS (SELECT 1 FROM public.events e WHERE e.id = event_id AND e.user_id = auth.uid())
        AND (
            (consumption_id IS NOT NULL AND EXISTS (SELECT 1 FROM public.consumptions c WHERE c.id = consumption_id AND c.user_id = auth.uid()))
            OR
            (consumption_item_id IS NOT NULL AND EXISTS (
                SELECT 1 
                FROM public.consumption_items ci 
                JOIN public.consumptions c ON c.id = ci.consumption_id 
                WHERE ci.id = consumption_item_id AND c.user_id = auth.uid()
            ))
        )
    );

CREATE POLICY "Owner can insert event_links" ON public.event_links
    FOR INSERT WITH CHECK (
        EXISTS (SELECT 1 FROM public.events e WHERE e.id = event_id AND e.user_id = auth.uid())
        AND (
            (consumption_id IS NOT NULL AND EXISTS (SELECT 1 FROM public.consumptions c WHERE c.id = consumption_id AND c.user_id = auth.uid()))
            OR
            (consumption_item_id IS NOT NULL AND EXISTS (
                SELECT 1 
                FROM public.consumption_items ci 
                JOIN public.consumptions c ON c.id = ci.consumption_id 
                WHERE ci.id = consumption_item_id AND c.user_id = auth.uid()
            ))
        )
    );

CREATE POLICY "Owner can delete event_links" ON public.event_links
    FOR DELETE USING (
        EXISTS (SELECT 1 FROM public.events e WHERE e.id = event_id AND e.user_id = auth.uid())
        AND (
            (consumption_id IS NOT NULL AND EXISTS (SELECT 1 FROM public.consumptions c WHERE c.id = consumption_id AND c.user_id = auth.uid()))
            OR
            (consumption_item_id IS NOT NULL AND EXISTS (
                SELECT 1 
                FROM public.consumption_items ci 
                JOIN public.consumptions c ON c.id = ci.consumption_id 
                WHERE ci.id = consumption_item_id AND c.user_id = auth.uid()
            ))
        )
    );
