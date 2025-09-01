-- Migration 013: Add ingredients and url fields to items and consumption_items tables
-- This migration adds support for storing ingredient data and Open Food Facts URLs
-- for historical tracking and correlation analysis.

-- Add ingredients and url to items table
ALTER TABLE items 
ADD COLUMN IF NOT EXISTS ingredients JSONB,
ADD COLUMN IF NOT EXISTS url TEXT;

-- Add ingredients and url to consumption_items table  
ALTER TABLE consumption_items
ADD COLUMN IF NOT EXISTS ingredients JSONB,
ADD COLUMN IF NOT EXISTS url TEXT;

-- Add indexes for ingredient queries (GIN index for JSONB)
CREATE INDEX IF NOT EXISTS idx_items_ingredients ON items USING GIN (ingredients);
CREATE INDEX IF NOT EXISTS idx_consumption_items_ingredients ON consumption_items USING GIN (ingredients);

-- Add indexes for OFF URL queries
CREATE INDEX IF NOT EXISTS idx_items_url ON items(url) WHERE url IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_consumption_items_url ON consumption_items(url) WHERE url IS NOT NULL;
