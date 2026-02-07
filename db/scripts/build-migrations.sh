#!/bin/bash
# ============================================================
# Build Migrations
# Flattens year/month structure into .build/ for golang-migrate
# ============================================================

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DB_DIR="$(dirname "$SCRIPT_DIR")"
MIGRATIONS_DIR="$DB_DIR/migrations"
BUILD_DIR="$DB_DIR/.build"

# Clean and recreate build directory
rm -rf "$BUILD_DIR"
mkdir -p "$BUILD_DIR"

# Find all migration files and symlink to build directory
find "$MIGRATIONS_DIR" -type f -name "*.sql" | while read -r file; do
    filename=$(basename "$file")
    ln -sf "$file" "$BUILD_DIR/$filename"
done

echo "Built $(find "$BUILD_DIR" -name "*.sql" | wc -l | tr -d ' ') migration files to $BUILD_DIR"
