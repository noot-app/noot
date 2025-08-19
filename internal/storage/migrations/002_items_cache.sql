-- Migration 002: Items cache with soft TTL (30-day expiration, refresh in-place)
CREATE TABLE IF NOT EXISTS items_cache (
    id TEXT PRIMARY KEY,
    normalized_name TEXT NOT NULL,
    normalized_brand TEXT NOT NULL DEFAULT '',
    original_name TEXT NOT NULL,
    original_brand TEXT NOT NULL DEFAULT '',
    nutrient_data_json TEXT NOT NULL, -- JSON serialized nutrition data in canonical units (grams/mg/kcal)
    fetched_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at DATETIME NOT NULL,
    UNIQUE(normalized_name, normalized_brand)
);

CREATE INDEX IF NOT EXISTS idx_items_cache_normalized ON items_cache(normalized_name, normalized_brand);
CREATE INDEX IF NOT EXISTS idx_items_cache_expires_at ON items_cache(expires_at);

-- Item aliases table for alternative names
CREATE TABLE IF NOT EXISTS item_aliases (
    id TEXT PRIMARY KEY,
    cached_item_id TEXT NOT NULL,
    alias_name TEXT NOT NULL,
    normalized_alias TEXT NOT NULL,
    UNIQUE(normalized_alias, cached_item_id)
);

CREATE INDEX IF NOT EXISTS idx_item_aliases_cached_item_id ON item_aliases(cached_item_id);
CREATE INDEX IF NOT EXISTS idx_item_aliases_normalized ON item_aliases(normalized_alias);