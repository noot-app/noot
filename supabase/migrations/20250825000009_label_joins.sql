-- Migration 009: Create label join tables for consumption and item associations (PostgreSQL)
-- These tables enable many-to-many relationships between labels and consumptions/items

-- Join table for consumption labels
CREATE TABLE IF NOT EXISTS public.consumption_labels (
    consumption_id uuid NOT NULL REFERENCES public.consumptions(id) ON DELETE CASCADE,
    label_id uuid NOT NULL REFERENCES public.labels(id) ON DELETE CASCADE,
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (consumption_id, label_id)
);

-- Indexes for efficient queries
CREATE INDEX IF NOT EXISTS idx_consumption_labels_label ON public.consumption_labels(label_id);
CREATE INDEX IF NOT EXISTS idx_consumption_labels_consumption ON public.consumption_labels(consumption_id);

-- Enable RLS
ALTER TABLE public.consumption_labels ENABLE ROW LEVEL SECURITY;

-- RLS policies: user must own both the consumption and the label
CREATE POLICY "Owner can read consumption_labels" ON public.consumption_labels
    FOR SELECT USING (
        EXISTS (SELECT 1 FROM public.consumptions c WHERE c.id = consumption_id AND c.user_id = auth.uid())
        AND EXISTS (SELECT 1 FROM public.labels l WHERE l.id = label_id AND l.user_id = auth.uid())
    );

CREATE POLICY "Owner can insert consumption_labels" ON public.consumption_labels
    FOR INSERT WITH CHECK (
        EXISTS (SELECT 1 FROM public.consumptions c WHERE c.id = consumption_id AND c.user_id = auth.uid())
        AND EXISTS (SELECT 1 FROM public.labels l WHERE l.id = label_id AND l.user_id = auth.uid())
    );

CREATE POLICY "Owner can delete consumption_labels" ON public.consumption_labels
    FOR DELETE USING (
        EXISTS (SELECT 1 FROM public.consumptions c WHERE c.id = consumption_id AND c.user_id = auth.uid())
        AND EXISTS (SELECT 1 FROM public.labels l WHERE l.id = label_id AND l.user_id = auth.uid())
    );

-- Join table for consumption item labels
CREATE TABLE IF NOT EXISTS public.consumption_item_labels (
    consumption_item_id uuid NOT NULL REFERENCES public.consumption_items(id) ON DELETE CASCADE,
    label_id uuid NOT NULL REFERENCES public.labels(id) ON DELETE CASCADE,
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (consumption_item_id, label_id)
);

-- Indexes for efficient queries
CREATE INDEX IF NOT EXISTS idx_consumption_item_labels_label ON public.consumption_item_labels(label_id);
CREATE INDEX IF NOT EXISTS idx_consumption_item_labels_item ON public.consumption_item_labels(consumption_item_id);

-- Enable RLS
ALTER TABLE public.consumption_item_labels ENABLE ROW LEVEL SECURITY;

-- RLS policies: user must own the parent consumption of the item and the label
CREATE POLICY "Owner can read consumption_item_labels" ON public.consumption_item_labels
    FOR SELECT USING (
        EXISTS (
            SELECT 1
            FROM public.consumption_items ci
            JOIN public.consumptions c ON c.id = ci.consumption_id
            WHERE ci.id = consumption_item_id AND c.user_id = auth.uid()
        )
        AND EXISTS (SELECT 1 FROM public.labels l WHERE l.id = label_id AND l.user_id = auth.uid())
    );

CREATE POLICY "Owner can insert consumption_item_labels" ON public.consumption_item_labels
    FOR INSERT WITH CHECK (
        EXISTS (
            SELECT 1
            FROM public.consumption_items ci
            JOIN public.consumptions c ON c.id = ci.consumption_id
            WHERE ci.id = consumption_item_id AND c.user_id = auth.uid()
        )
        AND EXISTS (SELECT 1 FROM public.labels l WHERE l.id = label_id AND l.user_id = auth.uid())
    );

CREATE POLICY "Owner can delete consumption_item_labels" ON public.consumption_item_labels
    FOR DELETE USING (
        EXISTS (
            SELECT 1
            FROM public.consumption_items ci
            JOIN public.consumptions c ON c.id = ci.consumption_id
            WHERE ci.id = consumption_item_id AND c.user_id = auth.uid()
        )
        AND EXISTS (SELECT 1 FROM public.labels l WHERE l.id = label_id AND l.user_id = auth.uid())
    );
