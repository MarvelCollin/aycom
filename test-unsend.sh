#!/bin/bash

echo "=== Testing Unsend Message Feature ==="
echo ""

# Configuration
API_URL="http://localhost:8080/api/v1"
AUTH_TOKEN="${AUTH_TOKEN:-your_token_here}"
CHAT_ID="${CHAT_ID:-your_chat_id_here}"

if [ "$AUTH_TOKEN" = "your_token_here" ]; then
    echo "❌ Please set AUTH_TOKEN environment variable"
    echo "Usage: AUTH_TOKEN=your_token CHAT_ID=your_chat_id ./test-unsend.sh"
    exit 1
fi

if [ "$CHAT_ID" = "your_chat_id_here" ]; then
    echo "❌ Please set CHAT_ID environment variable"
    echo "Usage: AUTH_TOKEN=your_token CHAT_ID=your_chat_id ./test-unsend.sh"
    exit 1
fi

echo "🌐 API URL: $API_URL"
echo "💬 Chat ID: $CHAT_ID"
echo ""

# Test 1: Send a test message
echo "📤 Test 1: Sending test message..."
RESPONSE=$(curl -s -X POST "$API_URL/messages" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $AUTH_TOKEN" \
  -d "{
    \"chat_id\": \"$CHAT_ID\",
    \"content\": \"Test message for unsend - $(date)\",
    \"message_type\": \"text\"
  }")

echo "Response: $RESPONSE"

# Extract message ID from response
MESSAGE_ID=$(echo "$RESPONSE" | grep -o '"message_id":"[^"]*"' | cut -d'"' -f4)

if [ -z "$MESSAGE_ID" ]; then
    echo "❌ Failed to send test message or extract message ID"
    exit 1
fi

echo "✅ Test message sent with ID: $MESSAGE_ID"
echo ""

# Wait a moment for the message to be processed
sleep 2

# Test 2: Unsend the message
echo "🗑️ Test 2: Unsending message $MESSAGE_ID..."
UNSEND_RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X DELETE "$API_URL/messages/$MESSAGE_ID?chat_id=$CHAT_ID" \
  -H "Authorization: Bearer $AUTH_TOKEN")

HTTP_STATUS=$(echo "$UNSEND_RESPONSE" | grep "HTTP_STATUS:" | cut -d: -f2)
RESPONSE_BODY=$(echo "$UNSEND_RESPONSE" | sed '/HTTP_STATUS:/d')

echo "HTTP Status: $HTTP_STATUS"
echo "Response: $RESPONSE_BODY"

if [ "$HTTP_STATUS" = "200" ]; then
    echo "✅ Test 2 PASSED: Message unsent successfully"
else
    echo "❌ Test 2 FAILED: Expected HTTP 200, got $HTTP_STATUS"
fi
echo ""

# Test 3: Try to unsend a temp message
echo "🔄 Test 3: Testing temp message unsend..."
TEMP_MESSAGE_ID="temp_$(date +%s)"
TEMP_RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X DELETE "$API_URL/messages/$TEMP_MESSAGE_ID?chat_id=$CHAT_ID" \
  -H "Authorization: Bearer $AUTH_TOKEN")

TEMP_HTTP_STATUS=$(echo "$TEMP_RESPONSE" | grep "HTTP_STATUS:" | cut -d: -f2)
TEMP_RESPONSE_BODY=$(echo "$TEMP_RESPONSE" | sed '/HTTP_STATUS:/d')

echo "HTTP Status: $TEMP_HTTP_STATUS"
echo "Response: $TEMP_RESPONSE_BODY"

if [ "$TEMP_HTTP_STATUS" = "200" ] || [ "$TEMP_HTTP_STATUS" = "404" ]; then
    echo "✅ Test 3 PASSED: Temp message handled gracefully"
else
    echo "❌ Test 3 FAILED: Expected HTTP 200 or 404, got $TEMP_HTTP_STATUS"
fi
echo ""

# Test 4: Try to unsend non-existent message
echo "🚫 Test 4: Testing non-existent message..."
FAKE_MESSAGE_ID="nonexistent_message_99999"
FAKE_RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X DELETE "$API_URL/messages/$FAKE_MESSAGE_ID?chat_id=$CHAT_ID" \
  -H "Authorization: Bearer $AUTH_TOKEN")

FAKE_HTTP_STATUS=$(echo "$FAKE_RESPONSE" | grep "HTTP_STATUS:" | cut -d: -f2)
FAKE_RESPONSE_BODY=$(echo "$FAKE_RESPONSE" | sed '/HTTP_STATUS:/d')

echo "HTTP Status: $FAKE_HTTP_STATUS"
echo "Response: $FAKE_RESPONSE_BODY"

if [ "$FAKE_HTTP_STATUS" = "404" ]; then
    echo "✅ Test 4 PASSED: Non-existent message returned 404"
else
    echo "❌ Test 4 FAILED: Expected HTTP 404, got $FAKE_HTTP_STATUS"
fi
echo ""

echo "=== Test Summary ==="
echo "All tests completed. Check individual test results above."
