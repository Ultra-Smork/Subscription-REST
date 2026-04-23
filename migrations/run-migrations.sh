#!/bin/sh

set -e

echo "Waiting for database..."
until pg_isready -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER"; do
    echo "Database is unavailable - sleeping"
    sleep 2
done

echo "Database is up - running migrations"

for migration in migrations/*.up.sql; do
    if [ -f "$migration" ]; then
        echo "Running migration: $migration"
        PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f "$migration"
    fi
done

echo "Migrations completed successfully"

