#!/bin/sh

set -e

echo "Waiting for PostgreSQL to be ready..."
# Wait for auth_db to be ready
until PGPASSWORD=$DATABASE_PASSWORD psql -h $DATABASE_HOST -U $DATABASE_USER -c '\q'; do
  echo "PostgreSQL is unavailable - sleeping"
  sleep 1
done

echo "PostgreSQL is up - starting application"

# Start the application
echo "Starting auth service..."
exec ./auth-service