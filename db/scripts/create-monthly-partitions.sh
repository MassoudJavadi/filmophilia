#!/bin/bash
# ============================================================
# Create Monthly Partitions for High-Growth Tables
# Usage: ./create-monthly-partitions.sh YYYY-MM
# Example: ./create-monthly-partitions.sh 2028-01
# ============================================================

set -euo pipefail

if [ $# -ne 1 ]; then
    echo "Usage: $0 YYYY-MM"
    echo "Example: $0 2028-01"
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

# Calculate next month using PostgreSQL
PARTITION_START="${YEAR}-${MONTH}-01"

# Use PostgreSQL to calculate the next month (more reliable)
PARTITION_END=$(docker exec "$CONTAINER_NAME" psql -U "$POSTGRES_USER" -d "$DB_NAME" -t -c "SELECT ('${PARTITION_START}'::date + interval '1 month')::date;" | tr -d ' ')

if [ -z "$PARTITION_END" ]; then
    echo "Error: Failed to calculate next month"
    exit 1
fi

# Format partition name (remove hyphen: 2028-01 -> 2028_01)
PARTITION_NAME="${YEAR}_${MONTH}"

echo "Creating partitions for $YEAR_MONTH..."
echo "  Range: [$PARTITION_START, $PARTITION_END)"

# Create notifications partition
echo "Creating notifications_${PARTITION_NAME}..."
docker exec "$CONTAINER_NAME" psql -U "$POSTGRES_USER" -d "$DB_NAME" -c \
    "CREATE TABLE IF NOT EXISTS notifications_${PARTITION_NAME} PARTITION OF notifications FOR VALUES FROM ('$PARTITION_START') TO ('$PARTITION_END');"

# Create activities partition
echo "Creating activities_${PARTITION_NAME}..."
docker exec "$CONTAINER_NAME" psql -U "$POSTGRES_USER" -d "$DB_NAME" -c \
    "CREATE TABLE IF NOT EXISTS activities_${PARTITION_NAME} PARTITION OF activities FOR VALUES FROM ('$PARTITION_START') TO ('$PARTITION_END');"

echo ""
echo "✓ Successfully created partitions for $YEAR_MONTH"
echo ""
echo "To verify, run:"
echo "  docker exec $CONTAINER_NAME psql -U $POSTGRES_USER -d $DB_NAME -c \"\\d+ notifications\""
