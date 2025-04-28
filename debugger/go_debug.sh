#!/bin/bash

# Go Debugging Script for AYCOM Microservices
# This script helps with common Go debugging tasks

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}AYCOM Go Microservices Debugger${NC}"
echo "------------------------------------"

# Function to check if a command exists
command_exists() {
  command -v "$1" >/dev/null 2>&1
}

# Check Go installation
echo -e "\n${GREEN}Checking Go installation:${NC}"
if command_exists go; then
  go version
  echo -e "${GREEN}Go is properly installed.${NC}"
else
  echo -e "${RED}Go is not installed or not in PATH. Please install Go 1.21 or higher.${NC}"
fi

# Function to check a specific service
check_service() {
  local service_path=$1
  local service_name=$2
  
  echo -e "\n${GREEN}Checking $service_name service:${NC}"
  if [ -d "../backend/services/$service_path" ]; then
    echo "Service directory exists"
    
    if [ -f "../backend/services/$service_path/go.mod" ]; then
      echo "go.mod exists"
      cd "../backend/services/$service_path" || return
      
      echo -e "\n${GREEN}Running go mod verify for $service_name:${NC}"
      go mod verify
      
      echo -e "\n${GREEN}Checking for compilation errors in $service_name:${NC}"
      go build -o /dev/null ./... 2>&1
      
      cd - > /dev/null || return
    else
      echo -e "${RED}go.mod not found for $service_name service${NC}"
    fi
  else
    echo -e "${RED}$service_name service directory not found${NC}"
  fi
}

# Check API Gateway
echo -e "\n${GREEN}Checking API Gateway:${NC}"
if [ -d "../backend/api-gateway" ]; then
  echo "API Gateway directory exists"
  
  if [ -f "../backend/api-gateway/go.mod" ]; then
    echo "go.mod exists"
    cd "../backend/api-gateway" || exit
    
    echo -e "\n${GREEN}Running go mod verify for API Gateway:${NC}"
    go mod verify
    
    echo -e "\n${GREEN}Checking for compilation errors in API Gateway:${NC}"
    go build -o /dev/null ./... 2>&1
    
    cd - > /dev/null || exit
  else
    echo -e "${RED}go.mod not found for API Gateway${NC}"
  fi
else
  echo -e "${RED}API Gateway directory not found${NC}"
fi

# Check Auth Service
check_service "auth" "Auth"

# Check User Service
check_service "user" "User"

# Check Proto Files
echo -e "\n${GREEN}Checking Proto files:${NC}"
if [ -d "../backend/proto" ]; then
  echo "Proto directory exists"
  ls -la "../backend/proto"
else
  echo -e "${RED}Proto directory not found${NC}"
fi

# Offer additional debugging options
echo -e "\n${GREEN}Additional debugging options:${NC}"
echo "1. Run 'go vet' on all services"
echo "2. Check for race conditions"
echo "3. Run tests for all services"
echo "4. Exit"

read -p "Enter your choice (1-4): " choice

case $choice in
  1)
    echo -e "\n${GREEN}Running go vet on API Gateway:${NC}"
    cd "../backend/api-gateway" && go vet ./...
    echo -e "\n${GREEN}Running go vet on Auth Service:${NC}"
    cd "../backend/services/auth" && go vet ./...
    echo -e "\n${GREEN}Running go vet on User Service:${NC}"
    cd "../backend/services/user" && go vet ./...
    ;;
  2)
    echo -e "\n${GREEN}Checking for race conditions on API Gateway:${NC}"
    cd "../backend/api-gateway" && go build -race
    echo -e "\n${GREEN}Checking for race conditions on Auth Service:${NC}"
    cd "../backend/services/auth" && go build -race
    echo -e "\n${GREEN}Checking for race conditions on User Service:${NC}"
    cd "../backend/services/user" && go build -race
    ;;
  3)
    echo -e "\n${GREEN}Running tests for API Gateway:${NC}"
    cd "../backend/api-gateway" && go test -v ./...
    echo -e "\n${GREEN}Running tests for Auth Service:${NC}"
    cd "../backend/services/auth" && go test -v ./...
    echo -e "\n${GREEN}Running tests for User Service:${NC}"
    cd "../backend/services/user" && go test -v ./...
    ;;
  4)
    echo "Exiting..."
    exit 0
    ;;
  *)
    echo -e "${RED}Invalid choice${NC}"
    ;;
esac

echo -e "\n${YELLOW}Debugging completed!${NC}"