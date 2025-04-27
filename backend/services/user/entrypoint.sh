#!/bin/sh

set -e

echo "Waiting for PostgreSQL to be ready..."
# Wait for user_db to be ready
until PGPASSWORD=$DATABASE_PASSWORD psql -h $DATABASE_HOST -U $DATABASE_USER -c '\q'; do
  echo "PostgreSQL is unavailable - sleeping"
  sleep 1
done

echo "PostgreSQL is up - starting application"

# Run migrations
echo "Running migrations..."
./user-service migrate

# Start the application
echo "Starting user service..."
exec ./user-service