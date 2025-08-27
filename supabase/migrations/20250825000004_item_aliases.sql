-- Migration 004: Create item_aliases table for alternative names/spellings (PostgreSQL)
CREATE TABLE IF NOT EXISTS item_aliases (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    alias_name TEXT NOT NULL,
    alias_brand TEXT NOT NULL DEFAULT '',
    canonical_name TEXT NOT NULL,
    canonical_brand TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(alias_name, alias_brand)
);

CREATE INDEX IF NOT EXISTS idx_item_aliases_alias ON item_aliases(alias_name, alias_brand);
CREATE INDEX IF NOT EXISTS idx_item_aliases_canonical ON item_aliases(canonical_name, canonical_brand);

-- Enable RLS to restrict public access
ALTER TABLE item_aliases ENABLE ROW LEVEL SECURITY;

-- No public policies - only service role can access this table
-- This ensures item_aliases table is only accessible from the backend API
