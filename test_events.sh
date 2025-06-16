#!/bin/bash

echo "🧪 Testing RabbitMQ Event Publishing"
echo "======================================"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}1. Testing Thread Like Event${NC}"
curl -X POST "http://localhost:8083/api/v1/threads/test-thread-123/like" \
  -H "Authorization: Bearer test-token" \
  -H "Content-Type: application/json" \
  -w "\nStatus: %{http_code}\n" || echo -e "${RED}❌ API Gateway not running${NC}"

echo -e "\n${YELLOW}2. Testing Thread Bookmark Event${NC}"  
curl -X POST "http://localhost:8083/api/v1/threads/test-thread-123/bookmark" \
  -H "Authorization: Bearer test-token" \
  -H "Content-Type: application/json" \
  -w "\nStatus: %{http_code}\n" || echo -e "${RED}❌ API Gateway not running${NC}"

echo -e "\n${YELLOW}3. Testing User Follow Event${NC}"
curl -X POST "http://localhost:8083/api/v1/users/test-user-456/follow" \
  -H "Authorization: Bearer test-token" \
  -H "Content-Type: application/json" \
  -w "\nStatus: %{http_code}\n" || echo -e "${RED}❌ API Gateway not running${NC}"

echo -e "\n${GREEN}✅ Test complete! Check event bus logs for events:${NC}"
echo "docker-compose logs event_bus --tail=20"
