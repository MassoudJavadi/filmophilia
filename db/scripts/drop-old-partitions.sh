#!/bin/bash
# ============================================================
# Drop Old Partitions for High-Growth Tables
# Usage: ./drop-old-partitions.sh YYYY-MM
# Example: ./drop-old-partitions.sh 2024-01
# ============================================================

set -euo pipefail

if [ $# -ne 1 ]; then
    echo "Usage: $0 YYYY-MM"
    echo "Example: $0 2024-01"
    echo ""
    echo "WARNING: This will permanently delete data from the specified partition!"
    exit 1
fi

YEAR_MONTH=$1
CONTAINER_NAME="${PARTITION_CONTAINER:-filmophilia_db}"
POSTGRES_USER="${PARTITION_USER:-user}"
DB_NAME="${PARTITION_DB:-filmophilia}"

# Validate format
if ! [[ $YEAR_MONTH =~ ^[0-9]{4}-[0-9]{2}$ ]]; then
    echo "Error: Invalid format. Expected YYYY-MM, got: $YEAR_MONTH"
    exit 1
fi

# Extract year and month
YEAR=$(echo "$YEAR_MONTH" | cut -d'-' -f1)
MONTH=$(echo "$YEAR_MONTH" | cut -d'-' -f2)

# Format partition name (remove hyphen: 2024-01 -> 2024_01)
PARTITION_NAME="${YEAR}_${MONTH}"

echo "⚠️  WARNING: You are about to DROP the following partitions:"
echo "  - notifications_${PARTITION_NAME}"
echo "  - activities_${PARTITION_NAME}"
echo ""
echo "This will PERMANENTLY DELETE all data from $YEAR_MONTH!"
echo ""
read -p "Are you sure? Type 'yes' to confirm: " CONFIRM

if [ "$CONFIRM" != "yes" ]; then
    echo "Aborted."
    exit 0
fi

# Drop notifications partition
echo "Dropping notifications_${PARTITION_NAME}..."
docker exec "$CONTAINER_NAME" psql -U "$POSTGRES_USER" -d "$DB_NAME" <<SQL
DROP TABLE IF EXISTS notifications_${PARTITION_NAME};
SQL

# Drop activities partition
echo "Dropping activities_${PARTITION_NAME}..."
docker exec "$CONTAINER_NAME" psql -U "$POSTGRES_USER" -d "$DB_NAME" <<SQL
DROP TABLE IF EXISTS activities_${PARTITION_NAME};
SQL

echo ""
echo "✓ Successfully dropped partitions for $YEAR_MONTH"
