-- Migration 013: Add ingredients and off_url fields to items and consumption_items tables
-- This migration adds support for storing ingredient data and Open Food Facts URLs
-- for historical tracking and correlation analysis.

-- Add ingredients and off_url to items table
ALTER TABLE items 
ADD COLUMN IF NOT EXISTS ingredients JSONB,
ADD COLUMN IF NOT EXISTS off_url TEXT;

-- Add ingredients and off_url to consumption_items table  
ALTER TABLE consumption_items
ADD COLUMN IF NOT EXISTS ingredients JSONB,
ADD COLUMN IF NOT EXISTS off_url TEXT;

-- Add indexes for ingredient queries (GIN index for JSONB)
CREATE INDEX IF NOT EXISTS idx_items_ingredients ON items USING GIN (ingredients);
CREATE INDEX IF NOT EXISTS idx_consumption_items_ingredients ON consumption_items USING GIN (ingredients);

-- Add indexes for OFF URL queries
CREATE INDEX IF NOT EXISTS idx_items_off_url ON items(off_url) WHERE off_url IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_consumption_items_off_url ON consumption_items(off_url) WHERE off_url IS NOT NULL;