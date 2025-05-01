#!/bin/bash

# Set up colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Starting Auth Service Migration Script${NC}"

# Check if auth_service container is running
if ! docker ps | grep -q "aycom-auth_service-1"; then
    echo -e "${RED}Auth Service container is not running!${NC}"
    echo -e "${YELLOW}Please start the container first with: docker-compose up -d auth_service${NC}"
    exit 1
fi

# Check if auth_db container is running
if ! docker ps | grep -q "aycom-auth_db-1"; then
    echo -e "${RED}Auth DB container is not running!${NC}"
    echo -e "${YELLOW}Please start the container first with: docker-compose up -d auth_db${NC}"
    exit 1
fi

echo -e "${GREEN}Creating schema if it doesn't exist...${NC}"
docker exec aycom-auth_db-1 psql -U kolin -d auth_db -c "
CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";

-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    username VARCHAR(255) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    password_salt VARCHAR(64) NOT NULL,
    gender VARCHAR(10),
    date_of_birth TIMESTAMP WITH TIME ZONE,
    security_question VARCHAR(255),
    security_answer VARCHAR(255),
    google_id VARCHAR(255) UNIQUE,
    is_activated BOOLEAN DEFAULT TRUE,
    is_banned BOOLEAN DEFAULT FALSE,
    is_deactivated BOOLEAN DEFAULT FALSE,
    is_admin BOOLEAN DEFAULT FALSE,
    newsletter_subscription BOOLEAN DEFAULT FALSE,
    last_login_at TIMESTAMP WITH TIME ZONE,
    joined_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    verification_code VARCHAR(64),
    verification_code_expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create sessions table
CREATE TABLE IF NOT EXISTS sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    access_token VARCHAR(500) NOT NULL,
    refresh_token VARCHAR(500) NOT NULL UNIQUE,
    ip_address VARCHAR(45),
    user_agent TEXT,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
"

# Check if the migration succeeded
if [ $? -eq 0 ]; then
    echo -e "${GREEN}Auth service database migration completed successfully!${NC}"
    
    # Create an admin user for testing if needed
    echo -e "${YELLOW}Do you want to create a test admin user? (y/n)${NC}"
    read -r create_admin
    
    if [[ "$create_admin" == "y" || "$create_admin" == "Y" ]]; then
        # Note: In production you would hash this password
        echo -e "${GREEN}Creating test admin user...${NC}"
        docker exec aycom-auth_db-1 psql -U kolin -d auth_db -c "
        INSERT INTO users 
          (name, username, email, password_hash, password_salt, is_admin, is_activated) 
        VALUES 
          ('Admin User', 'admin', 'admin@example.com', 'password123', 'salt', TRUE, TRUE)
        ON CONFLICT (username) DO NOTHING;
        "
        echo -e "${GREEN}Test admin user created!${NC}"
    fi
    
    exit 0
else
    echo -e "${RED}Auth service database migration failed!${NC}"
    exit 1
fi 