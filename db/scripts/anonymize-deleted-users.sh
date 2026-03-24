#!/bin/bash
# ============================================================
# Anonymize Deleted Users
#
# Runs anonymization on users who have been soft-deleted for
# more than the specified retention period (default: 30 days)
#
# Usage: ./anonymize-deleted-users.sh [retention_days]
# Example: ./anonymize-deleted-users.sh 30
# ============================================================

set -euo pipefail

RETENTION_DAYS="${1:-30}"
CONTAINER_NAME="${ANONYMIZE_CONTAINER:-filmophilia_db}"
POSTGRES_USER="${ANONYMIZE_USER:-user}"
DB_NAME="${ANONYMIZE_DB:-filmophilia}"
DRY_RUN="${DRY_RUN:-false}"

echo "============================================================"
echo "User Anonymization Script"
echo "============================================================"
echo "Retention period: $RETENTION_DAYS days"
echo "Database: $DB_NAME"
echo "Dry run: $DRY_RUN"
echo ""

# Check for users eligible for anonymization
echo "Checking for users eligible for anonymization..."
ELIGIBLE_COUNT=$(docker exec "$CONTAINER_NAME" psql -U "$POSTGRES_USER" -d "$DB_NAME" -t -c "
    SELECT COUNT(*)
    FROM users
    WHERE deleted_at < NOW() - INTERVAL '$RETENTION_DAYS days'
      AND deleted_at IS NOT NULL
      AND anonymized_at IS NULL;
" | tr -d ' ')

if [ "$ELIGIBLE_COUNT" -eq 0 ]; then
    echo "No users eligible for anonymization."
    exit 0
fi

echo "Found $ELIGIBLE_COUNT user(s) eligible for anonymization."
echo ""

if [ "$DRY_RUN" = "true" ]; then
    echo "DRY RUN MODE - Showing users that would be anonymized:"
    docker exec "$CONTAINER_NAME" psql -U "$POSTGRES_USER" -d "$DB_NAME" -c "
        SELECT
            id,
            username,
            email,
            deleted_at,
            AGE(NOW(), deleted_at) as deleted_for
        FROM users
        WHERE deleted_at < NOW() - INTERVAL '$RETENTION_DAYS days'
          AND deleted_at IS NOT NULL
          AND anonymized_at IS NULL
        ORDER BY deleted_at;
    "
    echo ""
    echo "Run without DRY_RUN=true to actually anonymize these users."
    exit 0
fi

echo "⚠️  WARNING: This will anonymize $ELIGIBLE_COUNT user(s) by:"
echo "  - Replacing email with 'deleted_<id>@filmophilia.local'"
echo "  - Replacing username with 'deleted_user_<id>'"
echo "  - Clearing display_name, avatar_url, bio, password_hash"
echo "  - Deleting associated sessions and OAuth accounts"
echo ""
read -p "Continue? (yes/no): " CONFIRM

if [ "$CONFIRM" != "yes" ]; then
    echo "Aborted."
    exit 0
fi

echo ""
echo "Starting anonymization process..."

# Perform anonymization in a transaction
docker exec "$CONTAINER_NAME" psql -U "$POSTGRES_USER" -d "$DB_NAME" <<SQL
BEGIN;

-- Store user IDs for cleanup
CREATE TEMP TABLE users_to_anonymize AS
SELECT id
FROM users
WHERE deleted_at < NOW() - INTERVAL '$RETENTION_DAYS days'
  AND deleted_at IS NOT NULL
  AND anonymized_at IS NULL;

-- Anonymize user data
UPDATE users
SET
    email = 'deleted_' || id || '@filmophilia.local',
    username = 'deleted_user_' || id,
    display_name = NULL,
    avatar_url = NULL,
    bio = NULL,
    password_hash = NULL,
    is_verified = FALSE,
    anonymized_at = NOW()
WHERE id IN (SELECT id FROM users_to_anonymize);

-- Delete associated sessions (cascade handles this, but explicit for clarity)
DELETE FROM sessions
WHERE user_id IN (SELECT id FROM users_to_anonymize);

-- Delete OAuth accounts (cascade handles this, but explicit for clarity)
DELETE FROM accounts
WHERE user_id IN (SELECT id FROM users_to_anonymize);

-- Report results
SELECT
    COUNT(*) as anonymized_users,
    MIN(deleted_at) as oldest_deletion,
    MAX(deleted_at) as newest_deletion
FROM users
WHERE id IN (SELECT id FROM users_to_anonymize);

COMMIT;

-- Cleanup temp table
DROP TABLE users_to_anonymize;
SQL

echo ""
echo "✓ Anonymization complete!"
echo ""
echo "Notes:"
echo "  - User comments and ratings are preserved (displayed as 'Deleted User')"
echo "  - User follows, watchlists, activities, notifications are cascade deleted"
echo "  - Sessions and OAuth accounts are deleted"
echo ""
echo "To hard delete anonymized users (GDPR compliance):"
echo "  DELETE FROM users WHERE anonymized_at IS NOT NULL;"
echo ""
