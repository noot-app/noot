# Database Management System

This document describes the dual database management system that supports both SQLite (for development) and PostgreSQL/Supabase (for production and local testing).

## Overview

The Noot application now supports two database providers:

- **SQLite**: Default for local development - simple, no setup required
- **Supabase/PostgreSQL**: For production and local Supabase development stack

## Environment Variables

| Variable | Description | Default | Examples |
|----------|-------------|---------|----------|
| `DATABASE_PROVIDER` | Database provider to use | `sqlite` | `sqlite`, `supabase` |
| `DATABASE_PATH` | SQLite database file path | `./noot.db` | `./dev.db`, `:memory:` |
| `SUPABASE_DB_URL` | PostgreSQL connection string | - | `postgresql://user:pass@host:5432/db?sslmode=disable` |
| `PUBLIC_SUPABASE_URL` | Supabase API URL | - | `http://localhost:54321` (local) or `https://xxx.supabase.co` (cloud) |
| `ENV` | Application environment | `production` | `development`, `production` |

## Database Manager Script

The `script/db` provides unified management for both database types with **server-free execution** for fast operations.

### Usage

```bash
script/db [provider] {command}
```

**Note:** All database commands run without starting the HTTP server, making them fast and port-conflict free.

### Providers

- `sqlite` - SQLite operations (default when no provider specified)
- `supabase` - Supabase PostgreSQL operations (works with both local and cloud)

### Commands

| Command | Description | SQLite | Supabase | Performance |
|---------|-------------|--------|----------|-------------|
| `reset` | Drop all tables and re-run migrations | ✅ | ✅ | Fast (migration-only) |
| `migrate` | Run migrations only | ✅ | ✅ | ⚡ Very Fast (migration-only) |
| `seed` | Add development seed data | ✅ | ⚠️ | Fast |
| `dump` | Show database content | ✅ | ✅ | ⚡ Very Fast (direct SQL) |
| `validate` | Check schema compatibility | ✅ | ✅ | ⚡ Very Fast (direct SQL) |
| `setup` | Initial setup/connection test | - | ✅ | ⚡ Very Fast (direct SQL) |
| `compare` | Compare SQLite and PostgreSQL schemas | ✅ (cross-provider) | | Fast |

**Performance Note:** All commands use optimized execution paths (direct SQL or `--migrate-only` flag) to avoid HTTP server startup.

### Examples

```bash
# SQLite operations (default provider)
script/db reset              # Reset SQLite database
script/db sqlite migrate     # Explicitly use SQLite for migrations
script/db sqlite seed        # Add development data
script/db sqlite dump        # Show database content

# Supabase operations (local or cloud)
script/db supabase setup     # Test Supabase connection
script/db supabase migrate   # Run PostgreSQL migrations (fast, no server)
script/db supabase dump      # Show Supabase content
script/db supabase validate  # Validate schema (fast, no server)

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

### Local Supabase Development

For production-like environment testing with local Supabase stack.

#### Quick Start

1. **Start Supabase** (automatically runs migrations + seeds):

   ```bash
   supabase start
   # Services available at:
   # - API: http://localhost:54321
   # - DB: postgresql://postgres:postgres@localhost:54322/postgres  
   # - Studio: http://localhost:54323
   ```

2. **Configure Environment** (`.env`):

   ```ini
   DATABASE_PROVIDER=supabase
   PUBLIC_SUPABASE_URL=http://localhost:54321
   SUPABASE_DB_URL=postgresql://postgres:postgres@localhost:54322/postgres?sslmode=disable
   ```

3. **Start Development Server**:

   ```bash
   script/server                 # Start server (database already seeded)
   # OR reset if needed:
   script/server --clean         # Quick reset & start server  
   script/server --clean-full    # Full Supabase restart & start server
   ```

#### Seeding Process

The system now uses Supabase's built-in `supabase/seed.sql` file which automatically runs after migrations. This provides a much cleaner and more maintainable seeding approach compared to the previous manual Admin API method.

**Benefits of the new approach:**

- Uses official Supabase seeding mechanism
- Automatically runs with `supabase start` and `supabase db reset`
- No complex Admin API calls or error handling needed
- Consistent with Supabase best practices

#### Development Commands

| Command | Speed | Description |
|---------|-------|-------------|
| `script/server --clean` | ⚡ Fast | Quick table reset, reseed, start server |
| `script/server --clean-full` | 🐌 Slow | Full Supabase restart, reseed, start server |
| `script/supabase-seed --quick` | ⚡ Fast | Clear tables, reseed users only |
| `script/supabase-seed --clean` | 🐌 Slow | Stop/restart Supabase, reseed users |

#### Seeded Test Users

Both reset commands create these test users:

- `monalisa@birki.io` / `password123` (pro tier)
- `alice@birki.io` / `password123` (free tier)

#### Key Benefits

- **Quick iteration**: `--clean` keeps services running (5-10x faster)
- **Full reset**: `--clean-full` when you need completely fresh state
- **Auth integration**: Proper Supabase auth + application user sync
- **Production-like**: PostgreSQL features, RLS, auth flows

### Cloud Supabase Production Testing

For testing against your production Supabase instance:

1. **Get Connection String**: From Supabase Dashboard → Settings → Database
2. **Set Environment**:

   ```bash
   export SUPABASE_DB_URL="postgresql://postgres.xxx:[password]@aws-0-region.pooler.supabase.com:6543/postgres"
   export DATABASE_PROVIDER=supabase
   ```

3. **Initialize**:

   ```bash
   script/db supabase setup      # Test connection
   script/db supabase migrate    # Run migrations
   script/server                 # Start with cloud database
   ```

**⚠️ Warning**: Cloud databases should use migration-based seeding, not the development seed scripts.

## Migration System

Migrations are organized by database type:

```text
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

1. **Port conflicts**: Kill processes on ports 3001, 54321-54324  
2. **SSL errors with local Supabase**: Add `?sslmode=disable` to connection string
3. **Auth/App user mismatch**: Check `subject` field matches Supabase auth user `id`
4. **Connection issues**: Verify `SUPABASE_DB_URL` format and network access
5. **Supabase won't start**: Ensure Docker is running, ports are free
6. **"psql not found"**: Install PostgreSQL client (`brew install postgresql`)

### Quick Diagnostics

```bash
# Test connections
script/db supabase setup

# Check database content
script/db supabase dump

# Verify schema
script/db supabase validate

# Nuclear reset options
script/server --clean       # Quick reset (recommended)  
script/server --clean-full  # Full restart (slow but thorough)
```

### Getting Help

1. Check connection: `script/db [provider] validate`
2. Check content: `script/db [provider] dump`
3. Compare schemas: `script/db compare`
4. Reset everything: `script/db [provider] reset`
