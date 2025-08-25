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

The `script/db-manager` provides unified management for both database types.

### Usage

```bash
script/db-manager [provider] {command}
```

### Providers

- `sqlite` - SQLite operations (default)
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
script/db-manager reset              # Reset SQLite database
script/db-manager sqlite migrate     # Run SQLite migrations
script/db-manager sqlite seed        # Add development data
script/db-manager sqlite dump        # Show database content

# Supabase operations
script/db-manager supabase setup     # Test Supabase connection
script/db-manager supabase migrate   # Run PostgreSQL migrations
script/db-manager supabase dump      # Show Supabase content
script/db-manager supabase validate  # Validate schema

# Cross-database operations
script/db-manager compare            # Compare schemas
```

## Development Workflow

### Local Development (SQLite - Default)

1. **Setup**: No additional setup required
   ```bash
   script/db-manager reset    # Creates and migrates database
   script/db-manager seed     # Adds sample data
   ```

2. **Daily use**:
   ```bash
   script/db-manager dump     # View data
   script/db-manager reset    # Fresh start
   ```

### Production-like Testing (Supabase)

1. **Setup**: Set your Supabase connection
   ```bash
   export SUPABASE_DB_URL="postgresql://user:pass@host:5432/db"
   script/db-manager supabase setup    # Test connection
   ```

2. **Initialize**:
   ```bash
   script/db-manager supabase migrate  # Run PostgreSQL migrations
   script/db-manager supabase validate # Verify schema
   ```

3. **Development**:
   ```bash
   export DATABASE_PROVIDER=supabase
   # Your app now uses PostgreSQL
   ```

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
   - Use `script/db-manager validate` to check schema

4. **Supabase connection issues**
   - Verify `SUPABASE_DB_URL` format
   - Check firewall/network access
   - Use `script/db-manager supabase setup` to test

### Getting Help

1. Check the database connection: `script/db-manager [provider] validate`
2. Compare schemas if you suspect drift: `script/db-manager compare`
3. Reset and start fresh: `script/db-manager [provider] reset`

## Future Enhancements

- [ ] Complete PostgreSQL Store implementation (remaining methods)
- [ ] Supabase-specific seeding with auth integration
- [ ] Database migration versioning/tracking
- [ ] Automated schema drift detection
- [ ] Docker Compose setup for local PostgreSQL testing