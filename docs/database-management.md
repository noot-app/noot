# Database Management System

This document describes the dual database management system that supports both SQLite (for development) and PostgreSQL/Supabase (for production).

## Overview

The Noot application now supports two database providers:

- **SQLite**: Default for local development - simple, no setup required
- **Supabase/PostgreSQL**: For production and when you need production-like testing

## Environment Variables

| Variable | Description | Default | Examples |
|----------|-------------|---------|----------|
| `DATABASE_PROVIDER` | Database provider to use | `sqlite` | `sqlite`, `supabase` |
| `DATABASE_PATH` | SQLite database file path | `./noot.db` | `./dev.db`, `:memory:` |
| `SUPABASE_DB_URL` | PostgreSQL connection string | - | `postgresql://user:pass@host:5432/db` |
| `ENV` | Application environment | `production` | `development`, `production` |

## Database Manager Script

The `script/db` provides unified management for both database types.

### Usage

```bash
script/db [provider] {command}
```

**Note:** Existing workflows are preserved - `script/server --clean` continues to work exactly the same.

### Providers

- `sqlite` - SQLite operations (default when no provider specified)
- `supabase` - Supabase PostgreSQL operations

### Commands

| Command | Description | SQLite | Supabase |
|---------|-------------|--------|----------|
| `reset` | Drop all tables and re-run migrations | ✅ | ✅ |
| `migrate` | Run migrations only | ✅ | ✅ |
| `seed` | Add development seed data | ✅ | ⚠️ |
| `dump` | Show database content | ✅ | ✅ |
| `validate` | Check schema compatibility | ✅ | ✅ |
| `setup` | Initial setup/connection test | - | ✅ |
| `compare` | Compare SQLite and PostgreSQL schemas | ✅ (cross-provider) | |

### Examples

```bash
# SQLite operations (default provider)
script/db reset              # Reset SQLite database
script/db sqlite migrate     # Explicitly use SQLite for migrations
script/db sqlite seed        # Add development data
script/db sqlite dump        # Show database content

# Supabase operations
script/db supabase setup     # Test Supabase connection
script/db supabase migrate   # Run PostgreSQL migrations
script/db supabase dump      # Show Supabase content
script/db supabase validate  # Validate schema

# Cross-database operations
script/db compare            # Compare schemas
```

## Development Workflow

### Local Development (SQLite - Default)

1. **Setup**: No additional setup required
   ```bash
   script/db reset    # Creates and migrates database
   script/db seed     # Adds sample data
   ```

2. **Daily use**:
   ```bash
   script/db dump     # View data
   script/db reset    # Fresh start
   
   # Existing workflow still works:
   script/server --clean  # Reset, seed, and start server
   ```

### Production-like Testing (Supabase)

1. **Setup**: Set your Supabase connection
   ```bash
   export SUPABASE_DB_URL="postgresql://user:pass@host:5432/db"
   script/db supabase setup    # Test connection
   ```

2. **Initialize**:
   ```bash
   script/db supabase migrate  # Run PostgreSQL migrations
   script/db supabase validate # Verify schema
   ```

3. **Development**:
   ```bash
   export DATABASE_PROVIDER=supabase
   # Your app now uses PostgreSQL
   ```

## Supabase Authentication & First-Time Setup

### Understanding Supabase Auth vs Application Users

Supabase provides two user systems that work together:

1. **Supabase Auth Users** (`auth.users` table)
   - Managed by Supabase's authentication system
   - Handles login/logout, email verification, password reset
   - Accessed via `https://supabase.com/dashboard/project/<project-id>/auth/users`
   - Contains authentication metadata (email, provider, etc.)

2. **Application Users** (`public.users` table) 
   - Your application's user profiles
   - Contains app-specific data (handle, full_name, subscription_tier, etc.)
   - Links to Supabase Auth via `subject` field (maps to `auth.users.id`)
   - The `provider` and `subject` fields are **required** for this linking

### First-Time Supabase Setup

#### 1. Get Your Connection String

From your Supabase dashboard:
1. Go to Settings → Database
2. Find your connection string under "Connection parameters" 
3. Use the "postgres" format, not the URI format
4. Example: `postgresql://postgres.xxx:[password]@aws-0-[region].pooler.supabase.com:6543/postgres`

#### 2. Set Up Your Environment

```bash
# Set your connection string
export SUPABASE_DB_URL="postgresql://postgres.xxx:[password]@aws-0-region.pooler.supabase.com:6543/postgres"

# Test the connection
script/db supabase setup
```

#### 3. Run Initial Migrations

```bash
# Create all application tables (users, consumptions, etc.)
script/db supabase migrate

# Verify everything was created
script/db supabase validate
```

#### 4. Create Your First User

You have two options for creating users:

**Option A: Via Supabase Auth Dashboard**
1. Go to `https://supabase.com/dashboard/project/<project-id>/auth/users`
2. Click "Add user" 
3. Enter email and temporary password
4. Note the user's `id` (this becomes your `subject`)

**Option B: Via SQL (for development)**
```sql
-- Insert directly into your app's users table
INSERT INTO users (provider, subject, email, handle) 
VALUES ('supabase', 'actual-auth-user-id-from-supabase', 'test@example.com', 'testuser');
```

#### 5. Production Seeding Strategy

For production databases, avoid traditional "seeding" scripts. Instead:

**Store Everything as Code:**
```bash
# Create additional migration files for initial data
# internal/storage/migrations/postgres/007_seed_initial_data.sql
```

**Example seed migration:**
```sql
-- 007_seed_initial_data.sql
-- Insert essential data that should exist in production

-- Essential nutrition data
INSERT INTO items (name, category, default_calories_per_100g) VALUES 
('Apple', 'fruit', 52),
('Banana', 'fruit', 89)
ON CONFLICT (name) DO NOTHING;

-- Do NOT insert test users - those should be created via Supabase Auth
```

**For user management:**
- Production users: Always create via Supabase Auth dashboard/API
- Development users: Can use `script/db supabase seed` (when implemented)
- Staging users: Create via Auth, then link in your app

## Migration System

Migrations are organized by database type:

```
internal/storage/migrations/
├── sqlite/          # SQLite-specific migrations
│   ├── 001_create_users.sql
│   ├── 002_create_consumptions.sql
│   └── ...
└── postgres/        # PostgreSQL-specific migrations  
    ├── 001_create_users.sql
    ├── 002_create_consumptions.sql
    └── ...
```

### Key Differences

| Feature | SQLite | PostgreSQL |
|---------|--------|------------|
| Primary Keys | `TEXT` (ULID) | `UUID` |
| Timestamps | `DATETIME` | `TIMESTAMP WITH TIME ZONE` |
| Auto-generation | Go-side ULIDs | `gen_random_uuid()` |
| Schema validation | File existence | Connection + table checks |

## Application Integration

### Code Changes Required

**Minimal!** The existing `Store` interface remains unchanged. Only configuration differs:

```go
// SQLite (existing)
config := &storage.Config{
    Type:     "sqlite",
    Database: "./noot.db",
}

// PostgreSQL/Supabase (new)
config := &storage.Config{
    Type:     "supabase", 
    Database: "postgresql://...",
}

store, err := storage.NewStore(config)
// Same Store interface for both!
```

### Environment-based Switching

The application automatically selects the database provider based on `DATABASE_PROVIDER`:

```bash
# Development (default)
DATABASE_PROVIDER=sqlite ./noot

# Production/testing
DATABASE_PROVIDER=supabase SUPABASE_DB_URL="..." ./noot
```

## Testing

The test suite supports both database types:

```bash
# SQLite tests (always run)
go test ./internal/storage

# PostgreSQL tests (require connection)
TEST_POSTGRESQL_URL="postgresql://..." go test ./internal/storage -v
```

## Troubleshooting

### Common Issues

1. **"postgres support not yet implemented"**
   - Old error - update your code from the latest version

2. **"psql command not found"**
   - Install PostgreSQL client tools
   - On Ubuntu/Debian: `apt-get install postgresql-client`
   - On macOS: `brew install postgresql`

3. **Migration failures**
   - Check database permissions
   - Verify connection string format
   - Use `script/db validate` to check schema

4. **Supabase connection issues**
   - Verify `SUPABASE_DB_URL` format
   - Check firewall/network access  
   - Use `script/db supabase setup` to test

5. **Auth user vs App user mismatch**
   - Check that your app user's `subject` field matches the Supabase auth user's `id`
   - Verify the `provider` field is set correctly (`supabase`)
   - Use `script/db supabase dump` to inspect user data

6. **"script/server --clean doesn't work with Supabase"**
   - The `--clean` flag works with whatever `DATABASE_PROVIDER` is set to
   - For Supabase: `DATABASE_PROVIDER=supabase script/server --clean`
   - Note: Supabase seeding may need manual user creation via Auth dashboard

### Getting Help

1. Check the database connection: `script/db [provider] validate`
2. Compare schemas if you suspect drift: `script/db compare`
3. Reset and start fresh: `script/db [provider] reset`
4. Check what's in your database: `script/db [provider] dump`

## Future Enhancements

- [ ] Complete Supabase seeding implementation with auth integration
- [ ] Database migration versioning/tracking
- [ ] Automated schema drift detection  
- [ ] Docker Compose setup for local PostgreSQL testing
- [ ] Row Level Security (RLS) policies for production
- [ ] Integration with Supabase Realtime features