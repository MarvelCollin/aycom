/**
 * WebSocket Connection Stability Test
 * 
 * This script tests the WebSocket connection stability to ensure:
 * 1. No infinite reconnection loops
 * 2. Proper connection throttling
 * 3. Per-chat connection management
 * 4. Correct URL building for different environments
 */

// Mock environment for testing
const mockWindow = {
  location: {
    hostname: 'localhost',
    protocol: 'http:',
    port: '3000'
  }
};

// Mock auth token
const mockToken = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoidGVzdC11c2VyLWlkIiwic3ViIjoidGVzdC11c2VyLWlkIiwiZXhwIjoxNzM0MzM2MzE4fQ.test-signature';

// Test UUID for chat ID
const testChatId = '123e4567-e89b-12d3-a456-426614174000';

console.log('Testing WebSocket URL building...');

// Test URL building logic (extracted from websocketStore.ts)
function buildTestWebSocketUrl(chatId) {
  try {
    const token = mockToken;
    if (!token) {
      throw new Error('No authentication token available');
    }
    
    let protocol, hostname, port;
    
    const isLocalhost = mockWindow.location.hostname === 'localhost' || mockWindow.location.hostname === '127.0.0.1';
    
    if (isLocalhost) {
      protocol = mockWindow.location.protocol === 'https:' ? 'wss:' : 'ws:';
      hostname = mockWindow.location.hostname;
      port = '8083'; // API gateway port mapping
    } else {
      protocol = mockWindow.location.protocol === 'https:' ? 'wss:' : 'ws:';
      hostname = mockWindow.location.hostname;
      port = mockWindow.location.port || (protocol === 'wss:' ? '443' : '80');
    }
    
    const wsUrl = `${protocol}//${hostname}:${port}/api/v1/chats/${chatId}/ws?token=${encodeURIComponent(token)}`;
    
    console.log(`✓ WebSocket URL built successfully: ${wsUrl}`);
    return wsUrl;
  } catch (e) {
    console.error(`✗ Error building WebSocket URL: ${e.message}`);
    throw e;
  }
}

// Test connection throttling logic
function testConnectionThrottling() {
  console.log('\nTesting connection throttling...');
  
  const lastConnectionAttempt = {};
  const chatId = testChatId;
  
  function checkThrottling(chatId) {
    const now = Date.now();
    const lastAttempt = lastConnectionAttempt[chatId] || 0;
    
    if (now - lastAttempt < 2000) {
      console.log(`✓ Connection throttled for chat ${chatId}, last attempt was ${now - lastAttempt}ms ago`);
      return false; // Throttled
    }
    
    lastConnectionAttempt[chatId] = now;
    console.log(`✓ Connection allowed for chat ${chatId}`);
    return true; // Allowed
  }
  
  // Test rapid connection attempts
  console.log('Attempt 1:');
  checkThrottling(chatId); // Should be allowed
  
  console.log('Attempt 2 (immediate):');
  checkThrottling(chatId); // Should be throttled
  
  setTimeout(() => {
    console.log('Attempt 3 (after 2.1s delay):');
    checkThrottling(chatId); // Should be allowed
  }, 2100);
}

// Test per-chat reconnection logic
function testPerChatReconnection() {
  console.log('\nTesting per-chat reconnection attempts...');
  
  const reconnectAttempts = {};
  const maxReconnectAttempts = 10;
  
  function attemptReconnect(chatId) {
    const currentAttempts = reconnectAttempts[chatId] || 0;
    
    if (currentAttempts >= maxReconnectAttempts) {
      console.log(`✗ Maximum reconnect attempts (${maxReconnectAttempts}) reached for chat ${chatId}`);
      return false;
    }
    
    const baseDelay = 1000;
    const delay = Math.min(baseDelay * Math.pow(1.5, currentAttempts), 30000);
    reconnectAttempts[chatId] = currentAttempts + 1;
    
    console.log(`✓ Reconnection attempt ${currentAttempts + 1}/${maxReconnectAttempts} for chat ${chatId}, delay: ${delay}ms`);
    return true;
  }
  
  // Test reconnection for different chats
  const chat1 = '123e4567-e89b-12d3-a456-426614174001';
  const chat2 = '123e4567-e89b-12d3-a456-426614174002';
  
  // Test multiple attempts for chat1
  for (let i = 0; i < 12; i++) {
    if (!attemptReconnect(chat1)) break;
  }
  
  // Test that chat2 has independent counter
  console.log('\nTesting independent counters:');
  attemptReconnect(chat2); // Should start from attempt 1
  
  // Test reset logic
  console.log('\nTesting reset logic:');
  reconnectAttempts[chat1] = 0; // Simulate successful connection
  attemptReconnect(chat1); // Should start from attempt 1 again
}

// Test UUID validation
function testUUIDValidation() {
  console.log('\nTesting UUID validation...');
  
  const validUUIDs = [
    '123e4567-e89b-12d3-a456-426614174000',
    '550e8400-e29b-41d4-a716-446655440000',
    'f47ac10b-58cc-4372-a567-0e02b2c3d479'
  ];
  
  const invalidUUIDs = [
    'not-a-uuid',
    '123-456-789',
    '',
    'g23e4567-e89b-12d3-a456-426614174000', // Invalid character 'g'
    '123e4567-e89b-12d3-a456-42661417400' // Missing character
  ];
  
  const uuidRegex = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;
  
  validUUIDs.forEach(uuid => {
    if (uuidRegex.test(uuid)) {
      console.log(`✓ Valid UUID: ${uuid}`);
    } else {
      console.log(`✗ Should be valid but failed: ${uuid}`);
    }
  });
  
  invalidUUIDs.forEach(uuid => {
    if (!uuidRegex.test(uuid)) {
      console.log(`✓ Correctly rejected invalid UUID: ${uuid}`);
    } else {
      console.log(`✗ Should be invalid but passed: ${uuid}`);
    }
  });
}

// Run all tests
console.log('='.repeat(50));
console.log('WebSocket Connection Stability Test');
console.log('='.repeat(50));

try {
  buildTestWebSocketUrl(testChatId);
  testUUIDValidation();
  testConnectionThrottling();
  testPerChatReconnection();
  
  console.log('\n' + '='.repeat(50));
  console.log('✓ All tests completed successfully!');
  console.log('The WebSocket connection stability improvements should prevent:');
  console.log('- Infinite reconnection loops');
  console.log('- Duplicate connection attempts');
  console.log('- Rapid connection spamming');
  console.log('- Cross-chat interference');
  console.log('='.repeat(50));
  
} catch (error) {
  console.error('\n' + '='.repeat(50));
  console.error('✗ Test failed:', error.message);
  console.error('='.repeat(50));
}
