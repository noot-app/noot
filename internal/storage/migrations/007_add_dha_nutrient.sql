-- Migration 007: Add DHA (Docosahexaenoic acid) nutrient tracking
-- DHA is an omega-3 fatty acid important for brain and heart health

-- Add DHA column to consumptions table for tracking totals (if not exists)
ALTER TABLE consumptions ADD COLUMN dha_mg REAL NOT NULL DEFAULT 0;

-- Add DHA column to items_cache table for per-100g nutrition data (if not exists)
ALTER TABLE items_cache ADD COLUMN dha_mg_per_100g REAL NOT NULL DEFAULT 0;