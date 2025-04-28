#!/bin/bash

# AYCOM Service Management Script
# This script manages backend services (auth, user, etc.)
# Allows operations like migrate, seed, refresh, status, and more

# Color definitions for better readability
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
AUTH_SERVICE_DIR="services/auth"
USER_SERVICE_DIR="services/user"
API_GATEWAY_DIR="api-gateway"
EVENT_BUS_DIR="event-bus"

# Function to display usage information
show_usage() {
    echo -e "${BLUE}AYCOM Service Management Script${NC}"
    echo -e "\nUsage: $0 [command] [service]"
    echo -e "\nCommands:"
    echo -e "  ${GREEN}migrate${NC}    - Run database migrations for specified service(s)"
    echo -e "  ${GREEN}seed${NC}       - Run database seeders for specified service(s)"
    echo -e "  ${GREEN}direct_seed${NC}- Seed database directly (works when services are not running)"
    echo -e "  ${GREEN}status${NC}     - Check migration status for specified service(s)"
    echo -e "  ${GREEN}refresh${NC}    - Restart specified service(s)"
    echo -e "  ${GREEN}rebuild${NC}    - Rebuild and restart specified service(s)"
    echo -e "  ${GREEN}logs${NC}       - Show logs for specified service(s)"
    echo -e "  ${GREEN}stop${NC}       - Stop specified service(s)"
    echo -e "  ${GREEN}start${NC}      - Start specified service(s)"
    echo -e "  ${GREEN}check_db${NC}   - Check database values in specified service(s)"
    echo -e "  ${GREEN}help${NC}       - Show this help message"
    echo -e "\nServices:"
    echo -e "  ${YELLOW}auth${NC}       - Auth service"
    echo -e "  ${YELLOW}user${NC}       - User service"
    echo -e "  ${YELLOW}gateway${NC}    - API Gateway"
    echo -e "  ${YELLOW}eventbus${NC}   - Event Bus"
    echo -e "  ${YELLOW}all${NC}        - All services"
    echo -e "\nExamples:"
    echo -e "  $0 migrate all     # Run migrations for all services"
    echo -e "  $0 seed auth       # Run seeders for auth service"
    echo -e "  $0 direct_seed all # Seed all databases directly"
    echo -e "  $0 refresh user    # Restart user service"
    echo -e "  $0 logs gateway    # Show logs for API Gateway"
    echo -e "  $0 check_db auth   # Check database values in auth service"
}

# Function to run a command for the auth service
run_auth_command() {
    local command=$1
    echo -e "\n${BLUE}Running '$command' for Auth Service${NC}"
    
    cd "$(dirname "$0")/$AUTH_SERVICE_DIR" || { echo -e "${RED}Cannot access Auth Service directory${NC}"; return 1; }
    
    case $command in
        migrate)
            echo -e "${YELLOW}Running migrations for Auth Service...${NC}"
            docker-compose exec -T auth ./auth-service migrate
            ;;
        seed)
            echo -e "${YELLOW}Running seeders for Auth Service...${NC}"
            docker-compose exec -T auth ./auth-service seed
            ;;
        direct_seed)
            seed_auth_directly
            ;;
        status)
            echo -e "${YELLOW}Checking migration status for Auth Service...${NC}"
            docker-compose exec -T auth ./auth-service status
            ;;
        check_db)
            check_auth_db_values
            ;;
        refresh)
            echo -e "${YELLOW}Restarting Auth Service...${NC}"
            cd ../../ && docker-compose restart auth
            ;;
        rebuild)
            echo -e "${YELLOW}Rebuilding and restarting Auth Service...${NC}"
            cd ../../ && docker-compose up -d --build auth
            ;;
        logs)
            echo -e "${YELLOW}Showing logs for Auth Service...${NC}"
            cd ../../ && docker-compose logs --tail=100 -f auth
            ;;
        stop)
            echo -e "${YELLOW}Stopping Auth Service...${NC}"
            cd ../../ && docker-compose stop auth
            ;;
        start)
            echo -e "${YELLOW}Starting Auth Service...${NC}"
            cd ../../ && docker-compose start auth
            ;;
        *)
            echo -e "${RED}Unknown command: $command${NC}"
            return 1
            ;;
    esac
    
    echo -e "${GREEN}Completed '$command' for Auth Service${NC}"
    return 0
}

# Function to run a command for the user service
run_user_command() {
    local command=$1
    echo -e "\n${BLUE}Running '$command' for User Service${NC}"
    
    cd "$(dirname "$0")/$USER_SERVICE_DIR" || { echo -e "${RED}Cannot access User Service directory${NC}"; return 1; }
    
    case $command in
        migrate)
            echo -e "${YELLOW}Running migrations for User Service...${NC}"
            docker-compose exec -T user ./user-service migrate
            ;;
        seed)
            echo -e "${YELLOW}Running seeders for User Service...${NC}"
            docker-compose exec -T user ./user-service seed
            ;;
        direct_seed)
            seed_user_directly
            ;;
        status)
            echo -e "${YELLOW}Checking migration status for User Service...${NC}"
            docker-compose exec -T user ./user-service status
            ;;
        check_db)
            check_user_db_values
            ;;
        refresh)
            echo -e "${YELLOW}Restarting User Service...${NC}"
            cd ../../ && docker-compose restart user
            ;;
        rebuild)
            echo -e "${YELLOW}Rebuilding and restarting User Service...${NC}"
            cd ../../ && docker-compose up -d --build user
            ;;
        logs)
            echo -e "${YELLOW}Showing logs for User Service...${NC}"
            cd ../../ && docker-compose logs --tail=100 -f user
            ;;
        stop)
            echo -e "${YELLOW}Stopping User Service...${NC}"
            cd ../../ && docker-compose stop user
            ;;
        start)
            echo -e "${YELLOW}Starting User Service...${NC}"
            cd ../../ && docker-compose start user
            ;;
        *)
            echo -e "${RED}Unknown command: $command${NC}"
            return 1
            ;;
    esac
    
    echo -e "${GREEN}Completed '$command' for User Service${NC}"
    return 0
}

# Function to directly seed user database (without requiring the service to be running)
seed_user_directly() {
    echo -e "${YELLOW}Directly seeding User database...${NC}"
    
    # Change to project root directory
    cd "$(dirname "$0")/../" || { echo -e "${RED}Cannot access project root directory${NC}"; return 1; }
    
    # First, check if any users already exist
    USER_COUNT=$(docker-compose exec -T postgres psql -U ${POSTGRES_USER:-postgres} -d user_db -t -c "SELECT COUNT(*) FROM users")
    
    # Remove whitespace from result
    USER_COUNT=$(echo $USER_COUNT | tr -d ' ')
    
    if [[ "$USER_COUNT" -gt "0" ]]; then
        echo -e "${YELLOW}User profiles already exist (count: $USER_COUNT), skipping seeding${NC}"
        return 0
    fi
    
    NOW=$(date +"%Y-%m-%d %H:%M:%S")
    DOB_ADMIN='1990-01-01 00:00:00'
    DOB_JOHN='1995-05-15 00:00:00'
    DOB_JANE='1997-08-22 00:00:00'
    
    # Create admin user profile - using user_id to match actual schema
    docker-compose exec -T postgres psql -U ${POSTGRES_USER:-postgres} -d user_db -c "
    INSERT INTO users (user_id, username, name, email, profile_picture_url, banner_url, bio, gender, date_of_birth, joined_at, is_banned, is_deactivated, is_private, is_premium, newsletter_subscription, created_at, updated_at) 
    VALUES ('550e8400-e29b-41d4-a716-446655440000', 'admin', 'Admin User', 'admin@aycom.com', 'https://via.placeholder.com/150', 'https://via.placeholder.com/1200x300', 'I am the administrator of this platform.', 'Other', '$DOB_ADMIN', '$NOW', false, false, false, false, false, '$NOW', '$NOW');
    "
    
    # Create John Doe user profile - using user_id to match actual schema
    docker-compose exec -T postgres psql -U ${POSTGRES_USER:-postgres} -d user_db -c "
    INSERT INTO users (user_id, username, name, email, profile_picture_url, banner_url, bio, gender, date_of_birth, joined_at, is_banned, is_deactivated, is_private, is_premium, newsletter_subscription, created_at, updated_at) 
    VALUES ('550e8400-e29b-41d4-a716-446655440001', 'johndoe', 'John Doe', 'kolin@example.com', 'https://via.placeholder.com/150', 'https://via.placeholder.com/1200x300', 'Hello, I''m John Doe. I love coding and connecting with people.', 'Male', '$DOB_JOHN', '$NOW', false, false, false, false, true, '$NOW', '$NOW');
    "
    
    # Create Jane Doe user profile - using user_id to match actual schema
    docker-compose exec -T postgres psql -U ${POSTGRES_USER:-postgres} -d user_db -c "
    INSERT INTO users (user_id, username, name, email, profile_picture_url, banner_url, bio, gender, date_of_birth, joined_at, is_banned, is_deactivated, is_private, is_premium, newsletter_subscription, created_at, updated_at) 
    VALUES ('550e8400-e29b-41d4-a716-446655440002', 'janedoe', 'Jane Doe', 'jane@example.com', 'https://via.placeholder.com/150', 'https://via.placeholder.com/1200x300', 'Designer, photographer, and tech enthusiast.', 'Female', '$DOB_JANE', '$NOW', false, false, false, false, true, '$NOW', '$NOW');
    "
    
    echo -e "${GREEN}Successfully seeded default user profiles${NC}"
    return 0
}

# Function to directly seed auth database (without requiring the service to be running)
seed_auth_directly() {
    echo -e "${YELLOW}Directly seeding Auth database...${NC}"
    
    # Change to project root directory
    cd "$(dirname "$0")/../" || { echo -e "${RED}Cannot access project root directory${NC}"; return 1; }
    
    # Check if the users table exists
    TABLE_EXISTS=$(docker-compose exec -T postgres psql -U ${POSTGRES_USER:-postgres} -d auth_db -t -c "SELECT to_regclass('public.users');")
    
    # Create the table if it doesn't exist
    if [[ -z "$TABLE_EXISTS" || "$TABLE_EXISTS" == *"NULL"* ]]; then
        echo -e "${BLUE}Creating users table in auth database...${NC}"
        docker-compose exec -T postgres psql -U ${POSTGRES_USER:-postgres} -d auth_db -c "
        CREATE TABLE users (
            id UUID PRIMARY KEY,
            email VARCHAR(255) NOT NULL UNIQUE,
            username VARCHAR(50) NOT NULL UNIQUE,
            name VARCHAR(100) NOT NULL,
            hashed_password VARCHAR(255) NOT NULL,
            is_verified BOOLEAN DEFAULT false,
            gender VARCHAR(20),
            date_of_birth DATE,
            profile_picture TEXT,
            banner TEXT,
            security_question TEXT,
            security_answer TEXT,
            subscribe_to_newsletter BOOLEAN DEFAULT false,
            created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
            updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
            verification_code VARCHAR(64),
            verification_code_expiry TIMESTAMP WITH TIME ZONE DEFAULT NOW()
        );
        "
        echo -e "${GREEN}Users table created successfully.${NC}"
    else
        echo -e "${BLUE}Users table already exists in auth database.${NC}"
    fi
    
    # First, check if any users already exist
    USER_COUNT=$(docker-compose exec -T postgres psql -U ${POSTGRES_USER:-postgres} -d auth_db -t -c "SELECT COUNT(*) FROM users")
    
    # Remove whitespace from result
    USER_COUNT=$(echo $USER_COUNT | tr -d ' ')
    
    if [[ "$USER_COUNT" -gt "0" ]]; then
        echo -e "${YELLOW}Auth users already exist (count: $USER_COUNT), skipping seeding${NC}"
        return 0
    fi
    
    NOW=$(date +"%Y-%m-%d %H:%M:%S")
    
    # Create admin user
    ADMIN_HASH='$2a$10$KgGZ2GNjdAj8LqoLwpJCaeEuNpZgRqy2KMM.aPXIUi7h3B4kxzLj2'  # Hash for 'admin123'
    docker-compose exec -T postgres psql -U ${POSTGRES_USER:-postgres} -d auth_db -c "
    INSERT INTO users (id, email, name, username, hashed_password, is_verified, gender, date_of_birth, profile_picture, banner, security_question, security_answer, subscribe_to_newsletter, created_at, updated_at, verification_code, verification_code_expiry) 
    VALUES ('550e8400-e29b-41d4-a716-446655440000', 'admin@aycom.com', 'Admin User', 'admin', '$ADMIN_HASH', true, 'Other', '1990-01-01', 'https://via.placeholder.com/150', 'https://via.placeholder.com/1200x300', 'What is your first pet''s name?', 'Admin', false, '$NOW', '$NOW', '', '$NOW');
    "
    
    # Create John Doe user
    JOHN_HASH='$2a$10$lMXtDHODM6mUoBSW1wZzve8EQjQqNmLIg8Y9/0psKDwILTmpnJ3w.'  # Hash for 'kolin123'
    docker-compose exec -T postgres psql -U ${POSTGRES_USER:-postgres} -d auth_db -c "
    INSERT INTO users (id, email, name, username, hashed_password, is_verified, gender, date_of_birth, profile_picture, banner, security_question, security_answer, subscribe_to_newsletter, created_at, updated_at, verification_code, verification_code_expiry) 
    VALUES ('550e8400-e29b-41d4-a716-446655440001', 'kolin@example.com', 'John Doe', 'johndoe', '$JOHN_HASH', true, 'Male', '1995-05-15', 'https://via.placeholder.com/150', 'https://via.placeholder.com/1200x300', 'What is your mother''s maiden name?', 'Doe', true, '$NOW', '$NOW', '', '$NOW');
    "
    
    # Create Jane Doe user
    JANE_HASH='$2a$10$eZlZJXu0i8F7XFOw/Gh4G.d9w9CpFKHEcbDKW0UkSAt1jXYWcNZXO'  # Hash for 'securePass456!'
    docker-compose exec -T postgres psql -U ${POSTGRES_USER:-postgres} -d auth_db -c "
    INSERT INTO users (id, email, name, username, hashed_password, is_verified, gender, date_of_birth, profile_picture, banner, security_question, security_answer, subscribe_to_newsletter, created_at, updated_at, verification_code, verification_code_expiry) 
    VALUES ('550e8400-e29b-41d4-a716-446655440002', 'jane@example.com', 'Jane Doe', 'janedoe', '$JANE_HASH', true, 'Female', '1997-08-22', 'https://via.placeholder.com/150', 'https://via.placeholder.com/1200x300', 'What city were you born in?', 'New York', true, '$NOW', '$NOW', '', '$NOW');
    "
    
    echo -e "${GREEN}Successfully seeded default auth users${NC}"
    return 0
}

# Function to check auth database values
check_auth_db_values() {
    echo -e "${YELLOW}Checking Auth Service Database Values...${NC}"
    
    # Change directory to root for proper docker-compose execution
    cd "$(dirname "$0")/../" || { echo -e "${RED}Cannot access project root directory${NC}"; return 1; }
    
    echo -e "\n${BLUE}Available tables in auth_db:${NC}"
    docker-compose exec postgres psql -U ${POSTGRES_USER:-postgres} -d auth_db -c "\dt"
    
    # Show menu to select a table to view
    echo -e "\n${YELLOW}Select a table to view or enter 'all' to see all tables:${NC}"
    read -r table_name
    
    if [[ $table_name == "all" ]]; then
        # Get all table names
        tables=$(docker-compose exec postgres psql -U ${POSTGRES_USER:-postgres} -d auth_db -t -c "SELECT tablename FROM pg_tables WHERE schemaname='public'")
        
        # Display the content of each table
        for table in $tables; do
            table=$(echo $table | tr -d ' ')
            if [[ ! -z $table ]]; then
                echo -e "\n${GREEN}Table: $table${NC}"
                docker-compose exec postgres psql -U ${POSTGRES_USER:-postgres} -d auth_db -c "SELECT * FROM $table"
            fi
        done
    elif [[ ! -z $table_name ]]; then
        # Display the content of the selected table
        echo -e "\n${GREEN}Table: $table_name${NC}"
        docker-compose exec postgres psql -U ${POSTGRES_USER:-postgres} -d auth_db -c "SELECT * FROM $table_name"
        
        # Offer to search for specific values
        echo -e "\n${YELLOW}Would you like to search for specific values? [y/N]:${NC}"
        read -r search_option
        
        if [[ $search_option == "y" || $search_option == "Y" ]]; then
            echo -e "Enter column name:"
            read -r column_name
            
            echo -e "Enter search value:"
            read -r search_value
            
            echo -e "\n${GREEN}Search results:${NC}"
            docker-compose exec postgres psql -U ${POSTGRES_USER:-postgres} -d auth_db -c "SELECT * FROM $table_name WHERE $column_name::text LIKE '%$search_value%'"
        fi
    fi
    
    return 0
}

# Function to check user database values
check_user_db_values() {
    echo -e "${YELLOW}Checking User Service Database Values...${NC}"
    
    # Change directory to root for proper docker-compose execution
    cd "$(dirname "$0")/../" || { echo -e "${RED}Cannot access project root directory${NC}"; return 1; }
    
    echo -e "\n${BLUE}Available tables in user_db:${NC}"
    docker-compose exec postgres psql -U ${POSTGRES_USER:-postgres} -d user_db -c "\dt"
    
    # Show menu to select a table to view
    echo -e "\n${YELLOW}Select a table to view or enter 'all' to see all tables:${NC}"
    read -r table_name
    
    if [[ $table_name == "all" ]]; then
        # Get all table names
        tables=$(docker-compose exec postgres psql -U ${POSTGRES_USER:-postgres} -d user_db -t -c "SELECT tablename FROM pg_tables WHERE schemaname='public'")
        
        # Display the content of each table
        for table in $tables; do
            table=$(echo $table | tr -d ' ')
            if [[ ! -z $table ]]; then
                echo -e "\n${GREEN}Table: $table${NC}"
                docker-compose exec postgres psql -U ${POSTGRES_USER:-postgres} -d user_db -c "SELECT * FROM $table"
            fi
        done
    elif [[ ! -z $table_name ]]; then
        # Display the content of the selected table
        echo -e "\n${GREEN}Table: $table_name${NC}"
        docker-compose exec postgres psql -U ${POSTGRES_USER:-postgres} -d user_db -c "SELECT * FROM $table_name"
        
        # Offer to search for specific values
        echo -e "\n${YELLOW}Would you like to search for specific values? [y/N]:${NC}"
        read -r search_option
        
        if [[ $search_option == "y" || $search_option == "Y" ]]; then
            echo -e "Enter column name:"
            read -r column_name
            
            echo -e "Enter search value:"
            read -r search_value
            
            echo -e "\n${GREEN}Search results:${NC}"
            docker-compose exec postgres psql -U ${POSTGRES_USER:-postgres} -d user_db -c "SELECT * FROM $table_name WHERE $column_name::text LIKE '%$search_value%'"
        fi
    fi
    
    return 0
}

# Function to run a command for the API Gateway
run_gateway_command() {
    local command=$1
    echo -e "\n${BLUE}Running '$command' for API Gateway${NC}"
    
    cd "$(dirname "$0")/$API_GATEWAY_DIR" || { echo -e "${RED}Cannot access API Gateway directory${NC}"; return 1; }
    
    case $command in
        refresh)
            echo -e "${YELLOW}Restarting API Gateway...${NC}"
            cd ../../ && docker-compose restart api-gateway
            ;;
        rebuild)
            echo -e "${YELLOW}Rebuilding and restarting API Gateway...${NC}"
            cd ../../ && docker-compose up -d --build api-gateway
            ;;
        logs)
            echo -e "${YELLOW}Showing logs for API Gateway...${NC}"
            cd ../../ && docker-compose logs --tail=100 -f api-gateway
            ;;
        stop)
            echo -e "${YELLOW}Stopping API Gateway...${NC}"
            cd ../../ && docker-compose stop api-gateway
            ;;
        start)
            echo -e "${YELLOW}Starting API Gateway...${NC}"
            cd ../../ && docker-compose start api-gateway
            ;;
        *)
            echo -e "${RED}Command '$command' not applicable for API Gateway${NC}"
            return 1
            ;;
    esac
    
    echo -e "${GREEN}Completed '$command' for API Gateway${NC}"
    return 0
}

# Function to run a command for the Event Bus
run_eventbus_command() {
    local command=$1
    echo -e "\n${BLUE}Running '$command' for Event Bus${NC}"
    
    cd "$(dirname "$0")/$EVENT_BUS_DIR" || { echo -e "${RED}Cannot access Event Bus directory${NC}"; return 1; }
    
    case $command in
        refresh)
            echo -e "${YELLOW}Restarting Event Bus...${NC}"
            cd ../../ && docker-compose restart event-bus
            ;;
        rebuild)
            echo -e "${YELLOW}Rebuilding and restarting Event Bus...${NC}"
            cd ../../ && docker-compose up -d --build event-bus
            ;;
        logs)
            echo -e "${YELLOW}Showing logs for Event Bus...${NC}"
            cd ../../ && docker-compose logs --tail=100 -f event-bus
            ;;
        stop)
            echo -e "${YELLOW}Stopping Event Bus...${NC}"
            cd ../../ && docker-compose stop event-bus
            ;;
        start)
            echo -e "${YELLOW}Starting Event Bus...${NC}"
            cd ../../ && docker-compose start event-bus
            ;;
        *)
            echo -e "${RED}Command '$command' not applicable for Event Bus${NC}"
            return 1
            ;;
    esac
    
    echo -e "${GREEN}Completed '$command' for Event Bus${NC}"
    return 0
}

# Function to run a command for all services
run_all_command() {
    local command=$1
    echo -e "\n${BLUE}Running '$command' for all services${NC}"
    
    local success=true
    
    # Run command for auth service
    run_auth_command "$command" || success=false
    
    # Run command for user service
    run_user_command "$command" || success=false
    
    # Run applicable commands for API Gateway
    case $command in
        refresh|rebuild|logs|stop|start)
            run_gateway_command "$command" || success=false
            ;;
    esac
    
    # Run applicable commands for Event Bus
    case $command in
        refresh|rebuild|logs|stop|start)
            run_eventbus_command "$command" || success=false
            ;;
    esac
    
    if [ "$success" = true ]; then
        echo -e "\n${GREEN}Successfully completed '$command' for all applicable services${NC}"
        return 0
    else
        echo -e "\n${RED}Some operations failed while running '$command' for all services${NC}"
        return 1
    fi
}

# Function to display the interactive menu
show_menu() {
    clear
    echo -e "${BLUE}=====================================${NC}"
    echo -e "${BLUE}   AYCOM Service Management Menu     ${NC}"
    echo -e "${BLUE}=====================================${NC}"
    echo -e ""
    echo -e "${YELLOW}Select a command:${NC}"
    echo -e ""
    echo -e "  ${GREEN}1)${NC} Migrate database"
    echo -e "  ${GREEN}2)${NC} Seed database"
    echo -e "  ${GREEN}3)${NC} Direct seed database (works without running services)"
    echo -e "  ${GREEN}4)${NC} Check migration status"
    echo -e "  ${GREEN}5)${NC} Refresh service"
    echo -e "  ${GREEN}6)${NC} Rebuild service"
    echo -e "  ${GREEN}7)${NC} View service logs"
    echo -e "  ${GREEN}8)${NC} Stop service"
    echo -e "  ${GREEN}9)${NC} Start service"
    echo -e "  ${GREEN}10)${NC} Check database values"
    echo -e "  ${GREEN}0)${NC} Exit"
    echo -e ""
    echo -e "${BLUE}=====================================${NC}"
}

# Function to display service selection menu
show_service_menu() {
    clear
    echo -e "${BLUE}=====================================${NC}"
    echo -e "${BLUE}      Select Service to Manage       ${NC}"
    echo -e "${BLUE}=====================================${NC}"
    echo -e ""
    echo -e "${YELLOW}Selected command: $1${NC}"
    echo -e ""
    echo -e "  ${GREEN}1)${NC} Auth Service"
    echo -e "  ${GREEN}2)${NC} User Service"
    echo -e "  ${GREEN}3)${NC} API Gateway"
    echo -e "  ${GREEN}4)${NC} Event Bus"
    echo -e "  ${GREEN}5)${NC} All Services"
    echo -e "  ${GREEN}6)${NC} Back to main menu"
    echo -e ""
    echo -e "${BLUE}=====================================${NC}"
}

# Function to execute the interactive menu
run_interactive_menu() {
    local exit_requested=false
    
    while [ "$exit_requested" = false ]; do
        show_menu
        echo -e "Enter your choice [0-10]: "
        read -r choice
        
        case $choice in
            1) 
                command="migrate"
                ;;
            2)
                command="seed"
                ;;
            3)
                command="direct_seed"
                ;;
            4)
                command="status"
                ;;
            5)
                command="refresh"
                ;;
            6)
                command="rebuild"
                ;;
            7)
                command="logs"
                ;;
            8)
                command="stop"
                ;;
            9)
                command="start"
                ;;
            10)
                command="check_db"
                ;;
            0)
                exit_requested=true
                continue
                ;;
            *)
                echo -e "${RED}Invalid option. Press any key to continue...${NC}"
                read -n 1
                continue
                ;;
        esac
        
        # Show service selection menu
        local back_requested=false
        while [ "$back_requested" = false ]; do
            show_service_menu "$command"
            echo -e "Enter your choice [1-6]: "
            read -r service_choice
            
            case $service_choice in
                1)
                    run_auth_command "$command"
                    ;;
                2)
                    run_user_command "$command"
                    ;;
                3)
                    run_gateway_command "$command"
                    ;;
                4)
                    run_eventbus_command "$command"
                    ;;
                5)
                    run_all_command "$command"
                    ;;
                6)
                    back_requested=true
                    continue
                    ;;
                *)
                    echo -e "${RED}Invalid option. Press any key to continue...${NC}"
                    read -n 1
                    continue
                    ;;
            esac
            
            echo -e "\n${GREEN}Operation completed. Press any key to continue...${NC}"
            read -n 1
            back_requested=true
        done
    done
    
    echo -e "${GREEN}Thank you for using AYCOM Service Management. Goodbye!${NC}"
}

# Main execution starts here
if [ $# -gt 0 ]; then
    # If arguments are provided, run in command-line mode (preserve original functionality)
    command=$1
    service=${2:-"all"}  # Default to "all" if no service specified
    
    # Handle help command
    if [ "$command" = "help" ]; then
        show_usage
        
        # For Windows: pause to see the output before closing
        if [[ "$OSTYPE" == "msys"* || "$OSTYPE" == "cygwin"* || "$OSTYPE" == "win"* ]]; then
            read -p "Press Enter to continue..."
        fi
        
        exit 0
    fi
    
    # Execute command for specified service
    case $service in
        auth)
            run_auth_command "$command"
            result=$?
            ;;
        user)
            run_user_command "$command"
            result=$?
            ;;
        gateway)
            run_gateway_command "$command"
            result=$?
            ;;
        eventbus)
            run_eventbus_command "$command"
            result=$?
            ;;
        all)
            run_all_command "$command"
            result=$?
            ;;
        *)
            echo -e "${RED}Unknown service: $service${NC}"
            show_usage
            result=1
            ;;
    esac
    
    # For Windows: pause to see the output before closing
    if [[ "$OSTYPE" == "msys"* || "$OSTYPE" == "cygwin"* || "$OSTYPE" == "win"* ]]; then
        read -p "Press Enter to continue..."
    fi
    
    exit $result
else
    # No arguments provided, run in interactive menu mode
    run_interactive_menu
fi