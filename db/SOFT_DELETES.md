# Soft Delete Implementation Guide

## Overview

User accounts use a **tiered soft delete system** to preserve data integrity while allowing for account removal and GDPR compliance.

## Why Soft Deletes?

### Problems with Hard Deletes (Cascading)

❌ **Broken comment threads**: Deleting a user removes all their comments, creating holes in discussions
❌ **Rating analytics corruption**: Movie ratings change retroactively when users delete accounts
❌ **Loss of context**: Can't see historical discussions or moderation history
❌ **Irreversible**: Can't undo accidental deletions

### Benefits of Soft Deletes

✅ **Preserved content**: Comments and ratings remain visible as "Deleted User"
✅ **Stable analytics**: Movie ratings don't fluctuate from account deletions
✅ **Moderation history**: Can review banned user's activity
✅ **Reversible**: Can restore accounts if needed
✅ **GDPR compliant**: Can fully anonymize and delete after retention period

## Architecture

### Three-Tier Deletion Process

```
┌─────────────────┐
│  Tier 1: Soft   │  Immediate (user request or ban)
│     Delete      │  → Mark account as deleted
└────────┬────────┘  → Keep all data intact
         │
         ▼
┌─────────────────┐
│ Tier 2: Anon-   │  After 30 days (automated)
│  ymization      │  → Remove PII (email, username, etc.)
└────────┬────────┘  → Delete private data (sessions, OAuth)
         │            → Keep public content (comments, ratings)
         ▼
┌─────────────────┐
│  Tier 3: Hard   │  On demand (GDPR/legal compliance)
│     Delete      │  → Permanently delete user record
└─────────────────┘  → Comments/ratings show as "Deleted User"
```

### Database Schema

```sql
-- Users table additions
users.deleted_at      TIMESTAMPTZ  -- When account was soft deleted
users.anonymized_at   TIMESTAMPTZ  -- When PII was removed
users.status          user_status  -- Includes 'DELETED' status

-- Constraints
CHECK (anonymized_at IS NULL OR deleted_at IS NOT NULL)
```

### Foreign Key Strategy

| Table | FK Behavior | Reason |
|-------|-------------|--------|
| **comments** | `ON DELETE SET NULL` | Preserve comment threads |
| **ratings** | `ON DELETE SET NULL` | Maintain rating analytics |
| **activities** | `ON DELETE CASCADE` | User-specific, not public |
| **notifications** | `ON DELETE CASCADE` | User-specific, ephemeral |
| **sessions** | `ON DELETE CASCADE` | Auth data, no longer needed |
| **accounts** | `ON DELETE CASCADE` | OAuth data, no longer needed |
| **follows** | `ON DELETE CASCADE` | Relationship data |
| **watchlists** | `ON DELETE CASCADE` | Personal data |

## Usage

### Tier 1: Soft Delete a User

```sql
-- Mark user as deleted (immediate)
UPDATE users
SET
    status = 'DELETED',
    deleted_at = NOW()
WHERE id = ?;
```

**What happens:**
- Account is marked as deleted
- User cannot log in
- All data remains intact
- Can be reversed by setting `deleted_at = NULL, status = 'ACTIVE'`

### Tier 2: Anonymize Deleted Users

Run the anonymization script (recommended: via cron job):

```bash
# Dry run (preview)
DRY_RUN=true ./db/scripts/anonymize-deleted-users.sh 30

# Actual anonymization (users deleted >30 days ago)
./db/scripts/anonymize-deleted-users.sh 30
```

**What happens:**
- Email → `deleted_<id>@filmophilia.local`
- Username → `deleted_user_<id>`
- Display name, avatar, bio → `NULL`
- Password hash → `NULL`
- Sessions and OAuth accounts → **deleted**
- Comments and ratings → **preserved** (shown as "Deleted User")

**Automated setup** (crontab):
```bash
# Run anonymization daily at 2 AM
0 2 * * * cd /path/to/filmophilia && ./db/scripts/anonymize-deleted-users.sh 30
```

### Tier 3: Hard Delete (GDPR Compliance)

```sql
-- Permanently delete user record
DELETE FROM users WHERE id = ?;
```

**What happens:**
- User record deleted permanently
- Comments remain with `user_id = NULL` → UI shows "Deleted User"
- Ratings remain with `user_id = NULL` → count in stats but no attribution
- All cascading data deleted (sessions, follows, etc.)

## Application Integration

### Querying Active Users

Always filter out deleted users in application queries:

```go
// Good: Filter out deleted users
func (r *UserRepository) FindByID(id int) (*User, error) {
    return r.db.Where("id = ? AND deleted_at IS NULL", id).First()
}

// Good: List active users
func (r *UserRepository) List() ([]*User, error) {
    return r.db.Where("deleted_at IS NULL").Find()
}
```

### Displaying Deleted User Content

```go
// Comment display logic
func (c *Comment) AuthorName() string {
    if c.UserID == nil {
        return "Deleted User"
    }
    return c.User.DisplayName
}

// Rating attribution
func (r *Rating) IsFromDeletedUser() bool {
    return r.UserID == nil
}
```

### Unique Constraints with Nullable user_id

The `ratings` table uses a **partial unique index** to handle deleted users:

```sql
-- Only enforces uniqueness when user_id is NOT NULL
CREATE UNIQUE INDEX ratings_user_movie_unique_idx
    ON ratings (user_id, movie_id)
    WHERE user_id IS NOT NULL;
```

This allows:
- ✅ Multiple ratings from deleted users (NULL user_id) on same movie
- ✅ One rating per active user per movie (enforced uniqueness)

## Data Retention Policy

### Recommended Policy

1. **Soft delete**: Immediate (user request or moderation)
2. **Anonymization**: 30 days after soft delete
3. **Hard delete**: Only on legal/GDPR request

### Compliance Considerations

**GDPR Right to Erasure:**
- Tier 1 (soft delete) satisfies most user deletion requests
- Tier 2 (anonymization) removes PII while preserving analytics
- Tier 3 (hard delete) for explicit GDPR requests

**Data retention:**
- Comments/ratings: Indefinite (anonymized)
- User profile data: 30 days (then anonymized)
- Session/auth data: Deleted during anonymization

## Monitoring

### Check Deletion Pipeline

```sql
-- Users pending anonymization
SELECT
    COUNT(*) as pending_anonymization,
    MIN(deleted_at) as oldest_deletion
FROM users
WHERE deleted_at < NOW() - INTERVAL '30 days'
  AND deleted_at IS NOT NULL
  AND anonymized_at IS NULL;

-- Recently anonymized users
SELECT
    COUNT(*) as recently_anonymized,
    MAX(anonymized_at) as last_anonymization
FROM users
WHERE anonymized_at > NOW() - INTERVAL '7 days';

-- Orphaned content (for verification)
SELECT
    (SELECT COUNT(*) FROM comments WHERE user_id IS NULL) as orphaned_comments,
    (SELECT COUNT(*) FROM ratings WHERE user_id IS NULL) as orphaned_ratings;
```

### Audit Deleted Users

```sql
-- Deletion timeline
SELECT
    DATE_TRUNC('day', deleted_at) as deletion_date,
    COUNT(*) as users_deleted,
    COUNT(CASE WHEN anonymized_at IS NOT NULL THEN 1 END) as anonymized
FROM users
WHERE deleted_at IS NOT NULL
GROUP BY DATE_TRUNC('day', deleted_at)
ORDER BY deletion_date DESC
LIMIT 30;
```

## Migration History

- `20260325070000_implement_soft_deletes.up.sql`: Soft delete implementation

## Troubleshooting

### Orphaned Content Not Displaying

Ensure your application handles NULL user_id:

```go
// Template example
{{ if .Comment.UserID }}
    <span>{{ .Comment.User.Username }}</span>
{{ else }}
    <span class="deleted-user">Deleted User</span>
{{ end }}
```

### Rating Analytics Changed

After implementing soft deletes, existing deletion cascade won't affect ratings. New deletions preserve ratings with NULL user_id. Your analytics queries should handle this:

```sql
-- Still count ratings from deleted users
SELECT
    movie_id,
    AVG(score) as average_rating,
    COUNT(*) as total_ratings,
    COUNT(user_id) as ratings_from_active_users
FROM ratings
GROUP BY movie_id;
```

### Reverting a Deletion

```sql
-- Restore a soft-deleted user (if not yet anonymized)
UPDATE users
SET
    status = 'ACTIVE',
    deleted_at = NULL
WHERE id = ? AND anonymized_at IS NULL;
```

Cannot restore after anonymization (PII is permanently removed).
