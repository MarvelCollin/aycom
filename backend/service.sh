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
    echo -e "  ${GREEN}status${NC}     - Check migration status for specified service(s)"
    echo -e "  ${GREEN}refresh${NC}    - Restart specified service(s)"
    echo -e "  ${GREEN}rebuild${NC}    - Rebuild and restart specified service(s)"
    echo -e "  ${GREEN}logs${NC}       - Show logs for specified service(s)"
    echo -e "  ${GREEN}stop${NC}       - Stop specified service(s)"
    echo -e "  ${GREEN}start${NC}      - Start specified service(s)"
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
    echo -e "  $0 refresh user    # Restart user service"
    echo -e "  $0 logs gateway    # Show logs for API Gateway"
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
        status)
            echo -e "${YELLOW}Checking migration status for Auth Service...${NC}"
            docker-compose exec -T auth ./auth-service status
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
        status)
            echo -e "${YELLOW}Checking migration status for User Service...${NC}"
            docker-compose exec -T user ./user-service status
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
    echo -e "  ${GREEN}3)${NC} Check migration status"
    echo -e "  ${GREEN}4)${NC} Refresh service"
    echo -e "  ${GREEN}5)${NC} Rebuild service"
    echo -e "  ${GREEN}6)${NC} View service logs"
    echo -e "  ${GREEN}7)${NC} Stop service"
    echo -e "  ${GREEN}8)${NC} Start service"
    echo -e "  ${GREEN}9)${NC} Exit"
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
        echo -e "Enter your choice [1-9]: "
        read -r choice
        
        case $choice in
            1) 
                command="migrate"
                ;;
            2)
                command="seed"
                ;;
            3)
                command="status"
                ;;
            4)
                command="refresh"
                ;;
            5)
                command="rebuild"
                ;;
            6)
                command="logs"
                ;;
            7)
                command="stop"
                ;;
            8)
                command="start"
                ;;
            9)
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