# Database Partitioning Guide

## Overview

The `notifications` and `activities` tables are partitioned using **monthly range partitioning** on the `created_at` column. This improves query performance and simplifies maintenance operations on high-growth tables.

## Partitioned Tables

- **notifications**: Partitioned by month on `created_at`
- **activities**: Partitioned by month on `created_at`

## Current Partition Coverage

Partitions are pre-created for:
- **2026**: All 12 months (January - December)
- **2027**: All 12 months (January - December)
- **Default partition**: Catches data outside defined ranges

## Benefits

1. **Query Performance**: Queries with `WHERE created_at` filters only scan relevant partitions
2. **Maintenance Efficiency**: Vacuum, analyze, and index operations work on smaller partition chunks
3. **Easy Archival**: Drop old partitions to delete historical data instantly
4. **Unbounded Growth**: Tables can grow indefinitely without performance degradation

## Partition Maintenance

### Creating New Partitions

Before the start of a new year, create partitions for upcoming months:

```bash
# Create partitions for a specific month
./db/scripts/create-monthly-partitions.sh 2028-01
./db/scripts/create-monthly-partitions.sh 2028-02
# ... continue for all 12 months
```

**Recommendation**: Set up a cron job or scheduled task to create partitions 2-3 months in advance.

### Dropping Old Partitions

To delete historical data (e.g., data older than 2 years), drop the corresponding partition:

```bash
# Drop partitions from 2024
./db/scripts/drop-old-partitions.sh 2024-01
./db/scripts/drop-old-partitions.sh 2024-02
# ... etc
```

**Warning**: Dropping a partition permanently deletes all data in that month!

### Listing Existing Partitions

```sql
-- List notification partitions
SELECT
    schemaname,
    tablename
FROM pg_tables
WHERE schemaname = 'public'
  AND tablename LIKE 'notifications_%'
ORDER BY tablename;

-- List activity partitions
SELECT
    schemaname,
    tablename
FROM pg_tables
WHERE schemaname = 'public'
  AND tablename LIKE 'activities_%'
ORDER BY tablename;
```

### Viewing Partition Details

```sql
-- View notification partition ranges
\d+ notifications

-- View activity partition ranges
\d+ activities
```

## Important Notes

### Primary Key Change

Due to PostgreSQL partitioning requirements, the primary keys for these tables now include `created_at`:

- **Before**: `PRIMARY KEY (id)`
- **After**: `PRIMARY KEY (id, created_at)`

This change is transparent to the application since:
1. No foreign keys reference these tables
2. `id` values are still unique across all partitions
3. Queries by `id` still work normally (Postgres scans all partitions automatically)

### Query Best Practices

For optimal performance, always include `created_at` filters when querying:

```sql
-- ✅ GOOD: Partition pruning occurs, only scans relevant partition(s)
SELECT * FROM notifications
WHERE user_id = 123
  AND created_at >= '2026-03-01'
  AND created_at < '2026-04-01'
ORDER BY created_at DESC
LIMIT 20;

-- ⚠️ WORKS BUT SLOWER: Scans all partitions
SELECT * FROM notifications
WHERE user_id = 123
ORDER BY created_at DESC
LIMIT 20;
```

The existing index `(user_id, is_read, created_at DESC)` supports efficient queries with partition pruning.

## Automation Options

### Option 1: Cron Job

Add to your server's crontab:

```bash
# Create next year's partitions every December 1st
0 0 1 12 * /path/to/filmophilia/db/scripts/create-monthly-partitions.sh $(date -d "+1 year" +\%Y)-01
# ... repeat for all 12 months
```

### Option 2: pg_partman Extension

For fully automated partition management, consider [pg_partman](https://github.com/pgpartman/pg_partman):

```sql
CREATE EXTENSION pg_partman;

-- Configure automatic partition creation
SELECT partman.create_parent('public.notifications', 'created_at', 'native', 'monthly');
SELECT partman.create_parent('public.activities', 'created_at', 'native', 'monthly');
```

## Migration History

- `20260325050000_partition_high_growth_tables.up.sql`: Initial partitioning setup
- `20260325060000_fix_partition_constraint_names.up.sql`: Clean up constraint naming

## Troubleshooting

### Error: "no partition of relation for row"

This occurs when inserting data outside the defined partition ranges. The default partition should catch this, but if you've dropped it:

```sql
-- Recreate default partitions
CREATE TABLE notifications_default PARTITION OF notifications DEFAULT;
CREATE TABLE activities_default PARTITION OF activities DEFAULT;
```

### Query Performance Issues

1. Check if partition pruning is happening:
   ```sql
   EXPLAIN SELECT * FROM notifications WHERE created_at >= '2026-03-01';
   ```
   Look for "Partitions removed" in the output.

2. Ensure indexes exist on partitions:
   ```sql
   \di+ notifications_*
   ```

3. Run ANALYZE on parent table:
   ```sql
   ANALYZE notifications;
   ANALYZE activities;
   ```
