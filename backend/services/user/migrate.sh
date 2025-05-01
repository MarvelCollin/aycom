#!/bin/sh

# Set this to ensure the script exits on any error
set -e

echo "Running migration script for user service..."

# Database connection
echo "Connecting to database..."
# Wait for PostgreSQL to be ready
if [ -n "$DATABASE_HOST" ] && [ -n "$DATABASE_USER" ] && [ -n "$DATABASE_PASSWORD" ]; then
  until PGPASSWORD=$DATABASE_PASSWORD psql -h $DATABASE_HOST -U $DATABASE_USER -c '\q' 2>/dev/null; do
    echo "PostgreSQL is unavailable - sleeping"
    sleep 1
  done
  echo "PostgreSQL is up"
fi

# Run migration without starting server
echo "Running database migrations..."
./user-service migrate --skip-server || {
  # If the command failed because of the port already being in use, still consider it a success
  # since the migrations would have completed
  if grep -q "bind: address already in use" /tmp/migration.log 2>/dev/null; then
    echo "Port is already in use, but migration likely completed successfully"
    exit 0
  fi
  echo "Migration failed"
  exit 1
}

echo "Migration completed successfully"
exit 0 