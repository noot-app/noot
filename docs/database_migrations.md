# Database Migrations Guide

This guide explains how to create and apply database migrations using the `script/db` helper script with Supabase.

## What are Migrations?

Migrations are versioned database schema changes that allow you to:

- Track database schema changes over time
- Apply changes consistently across environments (local, staging, production)
- Roll back changes if needed
- Collaborate with team members on database changes

## What are "Pending Migrations"?

A **pending migration** is a migration file that exists in your `supabase/migrations/` directory but hasn't been applied to the database yet. Supabase tracks which migrations have been applied using an internal history table.

## Creating a Migration

### 1. Create a New Migration File

```bash
# Create a new empty migration file
supabase migration new your_migration_name
```

This creates a file like `supabase/migrations/20240911123456_your_migration_name.sql` with a timestamp prefix.

### 2. Write Your Migration

Edit the generated file and add your SQL changes:

```sql
-- Example: Add a new table
CREATE TABLE IF NOT EXISTS public.users (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Example: Add RLS (Row Level Security)
ALTER TABLE public.users ENABLE ROW LEVEL SECURITY;

-- Example: Create a policy
CREATE POLICY "Users can view own profile" ON public.users
    FOR SELECT USING (auth.uid() = id);
```

## Applying Migrations

### Local Development

```bash
# Apply all pending migrations to your local database
script/db migrate
```

### Production

```bash
# Apply all pending migrations to production (requires confirmation)
script/db migrate --production
```

**⚠️ Warning**: Always test migrations locally first!

## Migration Workflow

### Recommended Development Flow

1. **Create migration locally**:

   ```bash
   supabase migration new add_user_preferences_table
   ```

2. **Write your migration** in the generated SQL file

3. **Apply to local database**:

   ```bash
   script/db migrate
   ```

4. **Test your changes** thoroughly in local development

5. **Check for differences** between local and production:

   ```bash
   script/db diff
   ```

6. **Apply to production** when ready:

   ```bash
   script/db migrate --production
   ```

## Common Scenarios

### Starting Fresh

If you want to reset your local database and apply all migrations from scratch:

```bash
script/db reset  # This will reset and apply all migrations + seed data
```

### Checking Migration Status

To see which migrations have been applied:

```bash
supabase migration list
```

### Creating Data Migrations

For migrations that modify data (not just schema):

```sql
-- Always use IF EXISTS and proper error handling
UPDATE public.users 
SET updated_at = NOW() 
WHERE updated_at IS NULL;

-- Add constraints after data is cleaned up
ALTER TABLE public.users 
ALTER COLUMN updated_at SET NOT NULL;
```

## Edge Cases and Troubleshooting

### 1. Migration Fails During Application

**Symptoms**: Migration command exits with error, database is in inconsistent state.

**Solutions**:

- Check the error message carefully
- Fix the SQL in your migration file
- For local: `script/db reset` to start fresh
- For production: Contact your DBA or use `supabase migration repair`

### 2. Out-of-Order Migrations

**Problem**: Someone else created a migration with an earlier timestamp.

**Solution**:

- Rename your migration file to have a later timestamp
- Or create a new migration and delete the old one (if not yet applied to production)

### 3. Migration Applied Locally but Not in Production

**Symptoms**: `script/db diff` shows differences even after migrating.

**Causes**:

- Migration file not committed to git
- Production hasn't been updated with latest migrations
- Migration failed in production but succeeded locally

**Solutions**:

```bash
# Check what migrations are pending for production
supabase migration list --linked

# Apply missing migrations
script/db migrate --production
```

### 4. Schema Drift

**Problem**: Production database has manual changes not reflected in migrations.

**Detection**:

```bash
script/db diff  # Shows differences between local migrations and production
```

**Solutions**:

- Create a new migration to match production changes
- Or reset production to match migrations (⚠️ **DATA LOSS**)

### 5. Large Data Migrations

**Problem**: Migration takes too long or times out.

**Solutions**:

- Break large migrations into smaller chunks
- Use `LIMIT` clauses and run multiple times
- Consider using background jobs for data transformations

### 6. Rollback Needed

**Problem**: Need to undo a migration that was already applied.

**Solutions**:

```bash
# Create a new migration that reverses the changes
supabase migration new rollback_previous_change

# Write SQL to undo the previous migration
# Then apply it normally
script/db migrate
```

## Best Practices

### ✅ Do's

1. **Always test locally first**
2. **Keep migrations small and focused**
3. **Use IF EXISTS and IF NOT EXISTS** for safety
4. **Write reversible migrations** when possible
5. **Include comments** explaining complex changes
6. **Backup production** before major migrations

### ❌ Don'ts

1. **Don't edit existing migration files** once applied to production
2. **Don't skip testing** migrations locally
3. **Don't run raw SQL** directly on production
4. **Don't assume migrations are atomic** - plan for partial failures
5. **Don't ignore migration warnings**

## Safety Checklist

Before running `script/db migrate --production`:

- [ ] Migration tested locally
- [ ] Database backup exists
- [ ] Team is aware of the change
- [ ] Migration is small and focused
- [ ] Rollback plan is ready
- [ ] Off-peak hours (if significant change)

## Emergency Procedures

### If Production Migration Fails

1. **Don't panic** - assess the damage first
2. **Check application logs** for errors
3. **Run `script/db diff`** to see current state
4. **Create hotfix migration** if needed
5. **Document the incident** for future reference

### If Database is Corrupted

1. **Restore from backup** immediately
2. **Review failed migration** for issues
3. **Fix migration file** and test locally
4. **Re-apply once confident**

## Getting Help

- **Supabase Documentation**: <https://supabase.com/docs/guides/database/migrations>
- **Check migration status**: `supabase migration list`
- **View migration history**: Check `supabase_migrations.schema_migrations` table
- **Community Support**: Supabase Discord/GitHub
