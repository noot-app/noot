-- Migration: Add consumed_at field to consumptions table
-- This field represents when the food was actually consumed, separate from when the record was created/updated

ALTER TABLE consumptions 
ADD COLUMN consumed_at TIMESTAMP WITH TIME ZONE;

-- Set default value for consumed_at to be the same as created_at for existing records
UPDATE consumptions 
SET consumed_at = created_at 
WHERE consumed_at IS NULL;

-- Make consumed_at NOT NULL and set default to current timestamp for new records
ALTER TABLE consumptions 
ALTER COLUMN consumed_at SET NOT NULL,
ALTER COLUMN consumed_at SET DEFAULT NOW();

-- Add index for consumed_at for better query performance
CREATE INDEX IF NOT EXISTS idx_consumptions_consumed_at ON consumptions(consumed_at);
CREATE INDEX IF NOT EXISTS idx_consumptions_user_consumed ON consumptions(user_id, consumed_at);

-- Update the existing user_created index to use consumed_at instead of created_at for consumption queries
DROP INDEX IF EXISTS idx_consumptions_user_created;
CREATE INDEX IF NOT EXISTS idx_consumptions_user_consumed_at ON consumptions(user_id, consumed_at);
