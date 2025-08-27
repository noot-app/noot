# Database Management System

This document describes the database management system that uses Supabase/PostgreSQL for all environments (development and production).

## Overview

The Noot application uses **Supabase/PostgreSQL** exclusively:

- **Local Development**: Use Supabase CLI to run a local PostgreSQL instance
- **Production**: Connect to hosted Supabase/PostgreSQL database

## Environment Variables

| Variable | Description | Default | Examples |
|----------|-------------|---------|----------|
| `DATABASE_PROVIDER` | Database provider to use | `supabase` | `supabase`, `postgres` |
| `SUPABASE_DB_URL` | PostgreSQL connection string | - | `postgresql://user:pass@host:5432/db?sslmode=disable` |
| `PUBLIC_SUPABASE_URL` | Supabase API URL | - | `http://localhost:54321` (local) or `https://xxx.supabase.co` (cloud) |
| `ENV` | Application environment | `production` | `development`, `production` |

## Database Manager Script

The `script/db` provides unified management for Supabase/PostgreSQL databases with **server-free execution** for fast operations.

### Usage

```bash
script/db {command}
```

**Note:** All database commands run without starting the HTTP server, making them fast and port-conflict free.

### Commands

| Command | Description | Supabase | Performance |
|---------|-------------|----------|-------------|
| `reset` | Drop all tables and re-run migrations | ✅ | Fast (migration-only) |
| `--clean` | Full clean reset: stop Supabase, restart, and run all migrations + seed | ✅ | Fast |
| `--quick` | Quick reset: reset database without restarting services, run migrations + seed | ✅ | ⚡ Very Fast |

**Performance Note:** All commands use optimized execution paths to avoid HTTP server startup.

### Examples

```bash
# Supabase operations (local or cloud)
script/db --clean            # Full reset: stop Supabase, restart, and run migrations + seed
script/db --quick            # Quick reset without restarting services
script/db reset              # Run supabase db reset on local development database
script/db reset --production # Run supabase db reset on linked production database (with warning)
```

## Security Model

### Authentication vs Profile Data

The Noot application implements Supabase's recommended security pattern:

- **Authentication data**: Managed by Supabase Auth in `auth.users`
- **Profile data**: Stored in `public.users` with foreign key to `auth.users(id)`
- **Automatic creation**: Database trigger creates profile when user signs up
- **Access control**: Row Level Security (RLS) policies protect data

### Why We Don't Grant Permissions to `supabase_auth_admin`

**❌ Problematic approach:**
```sql
GRANT SELECT, INSERT, UPDATE, DELETE ON public.users TO supabase_auth_admin;
```

**Issues with this approach:**
- Creates security risks by over-privileging the auth system
- Violates trust boundaries between auth and app data
- Allows uncontrolled modifications to app-specific data

**✅ Our secure approach:**
- Use `SECURITY DEFINER` trigger function for controlled elevated privileges
- Enable RLS policies for fine-grained access control
- Maintain separation between auth system and app data

See [`docs/database-security.md`](database-security.md) for detailed security documentation.

## Development Workflow

### Local Development with Supabase CLI

1. **Setup**: Install Supabase CLI and start local stack

   ```bash
   # Install Supabase CLI (if not already installed)
   brew install supabase/tap/supabase
   
   # Start local Supabase stack (includes PostgreSQL)
   supabase start
   ```

2. **Daily use**:

   ```bash
   script/db --quick   # Quick reset without restarting services
   script/db --clean   # Full reset: stop Supabase, restart, and run migrations + seed
   
   # Existing workflow still works:
   script/server --clean  # Reset, seed, and start server
   ```

### Supabase Configuration

Configure your environment variables (`.env`):

```ini
DATABASE_PROVIDER=supabase
PUBLIC_SUPABASE_URL=http://localhost:54321
SUPABASE_DB_URL=postgresql://postgres:postgres@localhost:54322/postgres?sslmode=disable
```

### Running the Development Server

```bash
script/server                 # Start server (database already seeded)
# OR reset if needed:
script/server --clean         # Reset database and start server  
```

### Seeding Process

The system uses Supabase's built-in `supabase/seed.sql` file which automatically runs after migrations. This provides a clean and maintainable seeding approach.

**Benefits of this approach:**

- Uses official Supabase seeding mechanism
- Automatically runs with `supabase start` and `supabase db reset`
- Consistent with Supabase best practices

### Development Commands

| Command | Speed | Description |
|---------|-------|-------------|
| `script/server --clean` | ⚡ Fast | Reset database and start server |
| `script/db --quick` | ⚡ Fast | Quick database reset without restarting services |
| `script/db --clean` | 🐌 Slow | Full Supabase restart and database reset |

### Seeded Test Users

Database reset commands create these test users:

- `monalisa@birki.io` / `password123` (pro tier)
- `alice@birki.io` / `password123` (free tier)

### Key Benefits

- **Quick iteration**: `--quick` keeps services running (5-10x faster)
- **Full reset**: `--clean` when you need completely fresh state
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
