# Database Migration Quick Reference

Quick reference for common migration tasks using `script/db`.

## Essential Commands

```bash
# Create new migration
supabase migration new migration_name

# Apply migrations locally
script/db migrate

# Apply migrations to production
script/db migrate --production

# Check differences between local and production
script/db diff

# See migration status
supabase migration list
```

## Complete Workflow Example

```bash
# 1. Create migration
supabase migration new add_user_table

# 2. Edit the SQL file in supabase/migrations/

# 3. Apply locally
script/db migrate

# 4. Test your changes

# 5. Check production differences
script/db diff

# 6. Apply to production
script/db migrate --production
```

## Common Migration Patterns

### Add Table

```sql
CREATE TABLE IF NOT EXISTS public.users (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

### Add Column

```sql
ALTER TABLE public.users 
ADD COLUMN IF NOT EXISTS full_name TEXT;
```

### Drop Column (Safe)

```sql
-- Step 1: Make nullable and drop constraints
ALTER TABLE public.users ALTER COLUMN old_column DROP NOT NULL;

-- Step 2: (In separate migration) Drop the column
ALTER TABLE public.users DROP COLUMN IF EXISTS old_column;
```
