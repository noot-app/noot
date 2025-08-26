-- Migration 004: Create item_aliases table for alternative names/spellings
CREATE TABLE IF NOT EXISTS item_aliases (
    id TEXT PRIMARY KEY,
    alias_name TEXT NOT NULL,
    alias_brand TEXT NOT NULL DEFAULT '',
    canonical_name TEXT NOT NULL,
    canonical_brand TEXT NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL,
    UNIQUE(alias_name, alias_brand)
);

CREATE INDEX IF NOT EXISTS idx_item_aliases_alias ON item_aliases(alias_name, alias_brand);
CREATE INDEX IF NOT EXISTS idx_item_aliases_canonical ON item_aliases(canonical_name, canonical_brand);