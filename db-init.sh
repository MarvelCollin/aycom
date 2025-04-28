#!/bin/bash
set -e

# Create databases
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    CREATE DATABASE auth_db;
    CREATE DATABASE user_db;
    GRANT ALL PRIVILEGES ON DATABASE auth_db TO $POSTGRES_USER;
    GRANT ALL PRIVILEGES ON DATABASE user_db TO $POSTGRES_USER;
    
    -- Connect to auth_db and create extensions if needed
    \c auth_db
    CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
    
    -- Connect to user_db and create extensions if needed
    \c user_db
    CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
EOSQL

echo "Database initialization completed" 