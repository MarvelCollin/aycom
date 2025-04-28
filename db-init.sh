#!/bin/bash
set -e

# Create databases
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
  -- Check if auth_db exists, if not create it
  SELECT 'CREATE DATABASE auth_db'
  WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'auth_db');
  
  -- Check if user_db exists, if not create it
  SELECT 'CREATE DATABASE user_db'
  WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'user_db');
  
  -- Set up users with proper permissions
  CREATE USER $AUTH_DB_USER WITH PASSWORD '$AUTH_DB_PASSWORD';
  GRANT ALL PRIVILEGES ON DATABASE auth_db TO $AUTH_DB_USER;
  
  CREATE USER $USER_DB_USER WITH PASSWORD '$USER_DB_PASSWORD';
  GRANT ALL PRIVILEGES ON DATABASE user_db TO $USER_DB_USER;
EOSQL

echo "Database initialization completed" 