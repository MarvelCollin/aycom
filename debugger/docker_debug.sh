#!/bin/bash

# Docker Debugging Script for AYCOM Microservices
# This script helps with Docker-related debugging tasks

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}AYCOM Docker Debugger${NC}"
echo "----------------------"

# Function to check if a command exists
command_exists() {
  command -v "$1" >/dev/null 2>&1
}

# Check Docker installation
echo -e "\n${GREEN}Checking Docker installation:${NC}"
if command_exists docker; then
  docker --version
  echo -e "${GREEN}Docker is properly installed.${NC}"
else
  echo -e "${RED}Docker is not installed or not in PATH.${NC}"
  exit 1
fi

# Check Docker Compose installation
echo -e "\n${GREEN}Checking Docker Compose installation:${NC}"
if command_exists docker-compose; then
  docker-compose --version
  echo -e "${GREEN}Docker Compose is properly installed.${NC}"
else
  echo -e "${RED}Docker Compose is not installed or not in PATH.${NC}"
  exit 1
fi

# Check if docker-compose.yml exists
echo -e "\n${GREEN}Checking for docker-compose.yml:${NC}"
if [ -f "../docker-compose.yml" ]; then
  echo -e "${GREEN}docker-compose.yml found.${NC}"
  echo -e "\n${YELLOW}Services defined in docker-compose.yml:${NC}"
  grep -E "^\s*[a-zA-Z0-9_-]+:" ../docker-compose.yml | sed 's/://'
else
  echo -e "${RED}docker-compose.yml not found in parent directory.${NC}"
  exit 1
fi

# Check Docker container status
echo -e "\n${GREEN}Checking Docker container status:${NC}"
docker ps -a

# Function to check and display logs for a service
check_service_logs() {
  local service=$1
  local lines=${2:-50}
  
  echo -e "\n${GREEN}Fetching logs for $service (last $lines lines):${NC}"
  docker-compose -f ../docker-compose.yml logs --tail="$lines" "$service"
}

# Function to check Docker network
check_docker_network() {
  echo -e "\n${GREEN}Docker Networks:${NC}"
  docker network ls
  
  echo -e "\n${GREEN}Detailed Network Information:${NC}"
  networks=$(docker network ls --format "{{.Name}}" | grep -i aycom)
  
  if [ -z "$networks" ]; then
    echo -e "${RED}No AYCOM-related networks found.${NC}"
  else
    for network in $networks; do
      echo -e "\n${YELLOW}Network: $network${NC}"
      docker network inspect "$network"
    done
  fi
}

# Check Docker volumes
check_docker_volumes() {
  echo -e "\n${GREEN}Docker Volumes:${NC}"
  docker volume ls
}

# Check for resource usage
check_resource_usage() {
  echo -e "\n${GREEN}Container Resource Usage:${NC}"
  docker stats --no-stream
}

# Display menu for debugging options
while true; do
  echo -e "\n${BLUE}AYCOM Docker Debugging Menu:${NC}"
  echo "1. Check container status"
  echo "2. Check service logs"
  echo "3. Check Docker networks"
  echo "4. Check Docker volumes"
  echo "5. Check resource usage"
  echo "6. Restart a service"
  echo "7. Rebuild a service"
  echo "8. Run docker-compose up -d"
  echo "9. Run docker-compose down"
  echo "10. Exit"
  
  read -p "Enter your choice (1-10): " choice
  
  case $choice in
    1)
      echo -e "\n${GREEN}Container status:${NC}"
      docker ps -a
      ;;
    2)
      echo -e "\n${YELLOW}Available services:${NC}"
      grep -E "^\s*[a-zA-Z0-9_-]+:" ../docker-compose.yml | sed 's/://'
      read -p "Enter service name: " service_name
      read -p "Enter number of lines (default: 50): " lines
      lines=${lines:-50}
      check_service_logs "$service_name" "$lines"
      ;;
    3)
      check_docker_network
      ;;
    4)
      check_docker_volumes
      ;;
    5)
      check_resource_usage
      ;;
    6)
      echo -e "\n${YELLOW}Available services:${NC}"
      grep -E "^\s*[a-zA-Z0-9_-]+:" ../docker-compose.yml | sed 's/://'
      read -p "Enter service name to restart: " service_name
      echo -e "\n${GREEN}Restarting $service_name...${NC}"
      docker-compose -f ../docker-compose.yml restart "$service_name"
      ;;
    7)
      echo -e "\n${YELLOW}Available services:${NC}"
      grep -E "^\s*[a-zA-Z0-9_-]+:" ../docker-compose.yml | sed 's/://'
      read -p "Enter service name to rebuild: " service_name
      echo -e "\n${GREEN}Rebuilding $service_name...${NC}"
      docker-compose -f ../docker-compose.yml up -d --build "$service_name"
      ;;
    8)
      echo -e "\n${GREEN}Starting all services...${NC}"
      docker-compose -f ../docker-compose.yml up -d
      ;;
    9)
      echo -e "\n${GREEN}Stopping all services...${NC}"
      docker-compose -f ../docker-compose.yml down
      ;;
    10)
      echo -e "\n${YELLOW}Exiting...${NC}"
      exit 0
      ;;
    *)
      echo -e "${RED}Invalid choice. Please try again.${NC}"
      ;;
  esac
done