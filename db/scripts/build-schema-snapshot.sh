#!/bin/bash
# ============================================================
# Build Schema Snapshot
# Generates a single schema snapshot SQL file from all up migrations
# ============================================================

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DB_DIR="$(dirname "$SCRIPT_DIR")"
REPO_DIR="$(dirname "$DB_DIR")"
BUILD_DIR="$DB_DIR/.build"
OUTPUT_FILE="${1:-$DB_DIR/schema.snapshot.sql}"
TEMP_DB_NAME="${SCHEMA_SNAPSHOT_DB_NAME:-filmophilia_schema_snapshot_tmp}"
SOURCE_DB_NAME="${SCHEMA_SNAPSHOT_SOURCE_DB:-}"
CONTAINER_NAME="${SCHEMA_SNAPSHOT_CONTAINER:-filmophilia_db}"
POSTGRES_USER="${SCHEMA_SNAPSHOT_USER:-user}"

mkdir -p "$(dirname "$OUTPUT_FILE")"

cleanup_local() {
  dropdb --if-exists "$TEMP_DB_NAME" >/dev/null 2>&1 || true
}

dump_local_database() {
  pg_dump \
    --schema-only \
    --no-owner \
    --no-privileges \
    "$SOURCE_DB_NAME" >"$OUTPUT_FILE"
}

generate_with_local_postgres() {
  "$SCRIPT_DIR/build-migrations.sh" >/dev/null

  UP_MIGRATIONS=()
  while IFS= read -r file; do
    UP_MIGRATIONS+=("$file")
  done < <(find "$BUILD_DIR" -maxdepth 1 \( -type f -o -type l \) -name "*.up.sql" | sort)

  if [ "${#UP_MIGRATIONS[@]}" -eq 0 ]; then
    echo "No up migrations found in $BUILD_DIR" >&2
    exit 1
  fi

  trap cleanup_local EXIT

  createdb "$TEMP_DB_NAME"

  for file in "${UP_MIGRATIONS[@]}"; do
    psql -v ON_ERROR_STOP=1 -d "$TEMP_DB_NAME" -f "$file" >/dev/null
  done

  pg_dump \
    --schema-only \
    --no-owner \
    --no-privileges \
    "$TEMP_DB_NAME" >"$OUTPUT_FILE"
}

cleanup_docker() {
  docker exec "$CONTAINER_NAME" dropdb -U "$POSTGRES_USER" --if-exists "$TEMP_DB_NAME" >/dev/null 2>&1 || true
}

dump_docker_database() {
  if ! docker ps --format '{{.Names}}' | grep -qx "$CONTAINER_NAME"; then
    (cd "$REPO_DIR" && docker compose up -d db >/dev/null)
  fi

  docker exec "$CONTAINER_NAME" pg_dump \
    --schema-only \
    --no-owner \
    --no-privileges \
    -U "$POSTGRES_USER" \
    "$SOURCE_DB_NAME" >"$OUTPUT_FILE"
}

generate_with_docker_postgres() {
  "$SCRIPT_DIR/build-migrations.sh" >/dev/null

  UP_MIGRATIONS=()
  while IFS= read -r file; do
    UP_MIGRATIONS+=("$file")
  done < <(find "$BUILD_DIR" -maxdepth 1 \( -type f -o -type l \) -name "*.up.sql" | sort)

  if [ "${#UP_MIGRATIONS[@]}" -eq 0 ]; then
    echo "No up migrations found in $BUILD_DIR" >&2
    exit 1
  fi

  if ! docker ps --format '{{.Names}}' | grep -qx "$CONTAINER_NAME"; then
    (cd "$REPO_DIR" && docker compose up -d db >/dev/null)
  fi

  trap cleanup_docker EXIT

  docker exec "$CONTAINER_NAME" dropdb -U "$POSTGRES_USER" --if-exists "$TEMP_DB_NAME" >/dev/null 2>&1 || true
  docker exec "$CONTAINER_NAME" createdb -U "$POSTGRES_USER" "$TEMP_DB_NAME"

  for file in "${UP_MIGRATIONS[@]}"; do
    docker exec -i "$CONTAINER_NAME" psql -U "$POSTGRES_USER" -d "$TEMP_DB_NAME" -v ON_ERROR_STOP=1 <"$file" >/dev/null
  done

  docker exec "$CONTAINER_NAME" pg_dump \
    --schema-only \
    --no-owner \
    --no-privileges \
    -U "$POSTGRES_USER" \
    "$TEMP_DB_NAME" >"$OUTPUT_FILE"
}

if command -v createdb >/dev/null 2>&1 && command -v psql >/dev/null 2>&1 && command -v pg_dump >/dev/null 2>&1; then
  if [ -n "$SOURCE_DB_NAME" ]; then
    dump_local_database
  else
    generate_with_local_postgres
  fi
elif command -v docker >/dev/null 2>&1; then
  if [ -n "$SOURCE_DB_NAME" ]; then
    dump_docker_database
  else
    generate_with_docker_postgres
  fi
else
  echo "Missing PostgreSQL CLI tools and Docker; cannot build schema snapshot." >&2
  exit 1
fi

echo "Schema snapshot written to $OUTPUT_FILE"
