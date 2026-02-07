# Database

## Structure

```
db/
├── migrations/          # Source migrations (organized by year/month)
│   └── YYYY/
│       └── MM/
│           ├── YYYYMMDDHHMMSS_description.up.sql
│           └── YYYYMMDDHHMMSS_description.down.sql
├── scripts/
│   └── build-migrations.sh   # Flattens migrations for golang-migrate
└── .build/              # Generated flat structure (git-ignored)
```

## Creating Migrations

```bash
# From apps/api directory
make migrate-create name=add_user_preferences
```

This creates files in the current year/month folder:
```
db/migrations/2026/02/20260207143000_add_user_preferences.up.sql
db/migrations/2026/02/20260207143000_add_user_preferences.down.sql
```

## Running Migrations

```bash
# Apply all pending migrations
make migrate-up

# Rollback last migration
make migrate-down

# Check current version
make migrate-version
```

## How It Works

1. **Source files** live in `migrations/YYYY/MM/` for organization
2. **Build step** symlinks all `.sql` files into `.build/` (flat)
3. **golang-migrate** reads from the flat `.build/` directory

The build step runs automatically before any migrate command.

## Best Practices

- One logical change per migration
- Always write both up and down migrations
- Test rollbacks: `migrate-up` → `migrate-down` → `migrate-up`
- Never modify applied migrations—create a new one instead
