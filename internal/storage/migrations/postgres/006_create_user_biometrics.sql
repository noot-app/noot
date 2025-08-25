-- Migration 006: Create user biometrics table (PostgreSQL)
CREATE TABLE IF NOT EXISTS user_biometrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL UNIQUE,
    
    -- Basic Demographics (for DRI)
    birth_date DATE NULL,
    sex TEXT CHECK (sex IN ('male','female','other','prefer_not_to_say')) DEFAULT 'prefer_not_to_say',
    
    -- Physical Measurements
    height_cm REAL NULL,
    weight_kg REAL NULL,
    
    -- Activity & Lifestyle
    activity_level TEXT CHECK (activity_level IN ('sedentary','lightly_active','moderately_active','very_active','extra_active')) DEFAULT 'lightly_active',
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_user_biometrics_user_id ON user_biometrics(user_id);