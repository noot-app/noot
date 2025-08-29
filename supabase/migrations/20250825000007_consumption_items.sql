-- Migration 007: Create consumption_items table for persistent meal ingredient breakdowns
-- Each consumption can have multiple items with nutrition snapshots at serving level
CREATE TABLE IF NOT EXISTS consumption_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    consumption_id UUID NOT NULL,
    item_id UUID NULL, -- Optional link to global items cache
    name TEXT NOT NULL,
    brand TEXT NOT NULL DEFAULT '',
    grams REAL NOT NULL DEFAULT 0,
    user_quantity REAL NULL,
    user_unit TEXT NULL,
    label TEXT NULL CONSTRAINT consumption_items_label_length CHECK (LENGTH(label) <= 63),
    note TEXT NULL CONSTRAINT consumption_items_note_length CHECK (LENGTH(note) <= 1000),
    -- Snapshot nutrition for THIS SERVING
    calories REAL NOT NULL DEFAULT 0,
    protein_g REAL NOT NULL DEFAULT 0,
    total_fat_g REAL NOT NULL DEFAULT 0,
    saturated_fat_g REAL NOT NULL DEFAULT 0,
    trans_fat_g REAL NOT NULL DEFAULT 0,
    cholesterol_mg REAL NOT NULL DEFAULT 0,
    sodium_mg REAL NOT NULL DEFAULT 0,
    total_carbs_g REAL NOT NULL DEFAULT 0,
    dietary_fiber_g REAL NOT NULL DEFAULT 0,
    total_sugars_g REAL NOT NULL DEFAULT 0,
    added_sugars_g REAL NOT NULL DEFAULT 0,
    vitamin_a_mcg REAL NOT NULL DEFAULT 0,
    vitamin_c_mg REAL NOT NULL DEFAULT 0,
    vitamin_d_mcg REAL NOT NULL DEFAULT 0,
    vitamin_e_mg REAL NOT NULL DEFAULT 0,
    vitamin_k_mcg REAL NOT NULL DEFAULT 0,
    thiamine_mg REAL NOT NULL DEFAULT 0,
    riboflavin_mg REAL NOT NULL DEFAULT 0,
    niacin_mg REAL NOT NULL DEFAULT 0,
    vitamin_b6_mg REAL NOT NULL DEFAULT 0,
    folate_mcg REAL NOT NULL DEFAULT 0,
    vitamin_b12_mcg REAL NOT NULL DEFAULT 0,
    biotin_mcg REAL NOT NULL DEFAULT 0,
    pantothenic_acid_mg REAL NOT NULL DEFAULT 0,
    choline_mg REAL NOT NULL DEFAULT 0,
    calcium_mg REAL NOT NULL DEFAULT 0,
    iron_mg REAL NOT NULL DEFAULT 0,
    magnesium_mg REAL NOT NULL DEFAULT 0,
    phosphorus_mg REAL NOT NULL DEFAULT 0,
    potassium_mg REAL NOT NULL DEFAULT 0,
    zinc_mg REAL NOT NULL DEFAULT 0,
    copper_mg REAL NOT NULL DEFAULT 0,
    manganese_mg REAL NOT NULL DEFAULT 0,
    selenium_mcg REAL NOT NULL DEFAULT 0,
    iodine_mcg REAL NOT NULL DEFAULT 0,
    molybdenum_mcg REAL NOT NULL DEFAULT 0,
    chromium_mcg REAL NOT NULL DEFAULT 0,
    fluoride_mg REAL NOT NULL DEFAULT 0,
    chloride_mg REAL NOT NULL DEFAULT 0,
    omega3_ala_g REAL NOT NULL DEFAULT 0,
    omega3_epa_g REAL NOT NULL DEFAULT 0,
    omega3_dha_g REAL NOT NULL DEFAULT 0,
    omega6_g REAL NOT NULL DEFAULT 0,
    creatine_mg REAL NOT NULL DEFAULT 0,
    caffeine_mg REAL NOT NULL DEFAULT 0,
    alcohol_g REAL NOT NULL DEFAULT 0,
    polyunsaturated_fat_g REAL NOT NULL DEFAULT 0,
    monounsaturated_fat_g REAL NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    FOREIGN KEY (consumption_id) REFERENCES consumptions(id) ON DELETE CASCADE,
    FOREIGN KEY (item_id) REFERENCES items(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_consumption_items_consumption_id ON consumption_items(consumption_id);
CREATE INDEX IF NOT EXISTS idx_consumption_items_item_id ON consumption_items(item_id);
CREATE INDEX IF NOT EXISTS idx_consumption_items_created ON consumption_items(consumption_id, created_at);

-- Trigger to update updated_at column
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS set_consumption_items_updated_at ON consumption_items;
CREATE TRIGGER set_consumption_items_updated_at
    BEFORE UPDATE ON consumption_items
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Enable RLS for consumption_items table
ALTER TABLE consumption_items ENABLE ROW LEVEL SECURITY;

-- RLS policies for consumption_items (users can only access items through their own consumptions)
CREATE POLICY "Users can view own consumption_items" ON consumption_items
    FOR SELECT USING (
        EXISTS (
            SELECT 1 FROM consumptions c
            WHERE c.id = consumption_id AND c.user_id = auth.uid()
        )
    );

CREATE POLICY "Users can insert consumption_items into own consumptions" ON consumption_items
    FOR INSERT WITH CHECK (
        EXISTS (
            SELECT 1 FROM consumptions c
            WHERE c.id = consumption_id AND c.user_id = auth.uid()
        )
    );

CREATE POLICY "Users can update own consumption_items" ON consumption_items
    FOR UPDATE USING (
        EXISTS (
            SELECT 1 FROM consumptions c
            WHERE c.id = consumption_id AND c.user_id = auth.uid()
        )
    );

CREATE POLICY "Users can delete own consumption_items" ON consumption_items
    FOR DELETE USING (
        EXISTS (
            SELECT 1 FROM consumptions c
            WHERE c.id = consumption_id AND c.user_id = auth.uid()
        )
    );