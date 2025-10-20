# smauth

A Go-based authentication service with PostgreSQL database migrations.

## Database Migrations

This project includes a comprehensive PostgreSQL migration CLI tool that helps you manage database schema changes.

### Features

- **Create migrations**: Generate timestamped up and down migration files
- **Run migrations**: Execute pending migrations in order
- **Rollback migrations**: Safely rollback the most recent migration
- **Migration tracking**: Automatic tracking of applied migrations in `tbl_migration` table
- **Status reporting**: View current migration status and pending migrations

### Quick Start

1. **Set up environment variables**:
   ```bash
   export DATABASE_URL="postgres://username:password@localhost:5432/dbname?sslmode=disable"
   export MIGRATIONS_DIR="database/migrations"  # Optional, defaults to database/migrations
   ```

2. **Build the migrator**:
   ```bash
   make build-migrator
   ```

3. **Create your first migration**:
   ```bash
   make migrate-create name=create_users_table
   # Or directly: ./bin/migrator create create_users_table
   ```

4. **Edit the generated SQL files** in `database/migrations/` directory

5. **Run migrations**:
   ```bash
   make migrate-up
   # Or directly: ./bin/migrator up
   ```

### Migration Commands

| Command | Description | Example |
|---------|-------------|---------|
| `create <name>` | Create new up and down migration files | `migrator create add_users_table` |
| `up` | Run all pending migrations | `migrator up` |
| `down` | Rollback the most recent migration | `migrator down` |
| `status` | Show migration status | `migrator status` |
| `init` | Initialize the migration tracking table | `migrator init` |

### Using Make Targets

```bash
# Build the migrator
make build-migrator

# Create a new migration
make migrate-create name=your_migration_name

# Run pending migrations
make migrate-up

# Rollback the latest migration
make migrate-down

# Check migration status
make migrate-status
```

### Migration File Structure

Migration files are created with timestamps to ensure proper ordering:

```
database/migrations/
├── 20251020204307_create_users_table_up.sql
├── 20251020204307_create_users_table_down.sql
├── 20251020204512_add_sessions_table_up.sql
└── 20251020204512_add_sessions_table_down.sql
```

**Up migration example** (`*_up.sql`):
```sql
-- Migration: 20251020204307_create_users_table
-- Created: 2025-10-20T20:43:07+02:00
-- Description: create_users_table

CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);
```

**Down migration example** (`*_down.sql`):
```sql
-- Migration: 20251020204307_create_users_table
-- Created: 2025-10-20T20:43:07+02:00
-- Description: create_users_table (rollback)

DROP INDEX IF EXISTS idx_users_email;
DROP TABLE IF EXISTS users;
```

### Migration Tracking

The system automatically creates and manages a `tbl_migration` table:

```sql
CREATE TABLE tbl_migration (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    applied_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

### Environment Variables

- `DATABASE_URL`: PostgreSQL connection string (required for up/down/status/init commands)
- `MIGRATIONS_DIR`: Directory containing migration files (optional, defaults to `database/migrations`)

### Best Practices

1. **Always test migrations**: Test both up and down migrations in a development environment
2. **Make migrations reversible**: Ensure every up migration has a corresponding down migration
3. **Use transactions**: The migrator automatically wraps each migration in a transaction
4. **Keep migrations small**: Create focused migrations that do one thing well
5. **Never edit applied migrations**: Once a migration is applied, create a new one for changes

### Error Handling

- Each migration runs in a transaction and rolls back automatically on error
- Failed migrations are not recorded in the tracking table
- The system provides detailed error messages for troubleshooting

### Development Workflow

1. Create a new migration: `make migrate-create name=descriptive_name`
2. Edit the generated SQL files with your schema changes
3. Test the migration: `make migrate-up`
4. If issues occur, rollback: `make migrate-down`
5. Fix and re-apply as needed

### Handy commands

env $(cat .env | xargs) your-command
export $(grep -v '^#' .env | xargs) && docker compose -f deploy/docker-compose.production.yaml up


<!-- // User login
auditService.LogEvent(ctx, audit.LogEventRequest{
    Event:     audit.EventUserLogin,
    AccountID: user.ID,
    ActorType: audit.AccountTypeUser,
    Resource:  "authentication",
    Action:    "login",
    Success:   true,
    IPAddress: req.RemoteAddr,
    UserAgent: req.UserAgent(),
})

// Client accessing EHR data
auditService.LogEvent(ctx, audit.LogEventRequest{
    Event:     audit.EventEHRRead,
    AccountID: client.ID,
    ActorType: audit.AccountTypeClient,
    Resource:  fmt.Sprintf("ehr/%s", ehrID),
    Action:    "read",
    Success:   true,
    Details: map[string]interface{}{
        "client_name": client.Name,
        "ehr_id":      ehrID,
        "composition_count": compositionCount,
    },
})

// System automated process
auditService.LogEvent(ctx, audit.LogEventRequest{
    Event:     audit.EventDataDelete,
    AccountID: "system",
    ActorType: audit.AccountTypeSystem,
    Resource:  "expired_sessions",
    Action:    "cleanup",
    Success:   true,
    Details: map[string]interface{}{
        "deleted_count": deletedCount,
        "automated":     true,
    },
}) -->