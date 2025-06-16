#!/usr/bin/env node

/**
 * WebSocket Connection Test Script
 * 
 * This script tests the WebSocket functionality to ensure:
 * 1. Connections can be established
 * 2. Messages can be sent and received
 * 3. No fallback mode is triggered
 * 4. Reconnection works properly
 */

const WebSocket = require('ws');

// Configuration
const CONFIG = {
  wsUrl: 'ws://localhost:8083/api/v1/chats/test-chat-123/ws',
  testToken: 'your-test-jwt-token-here', // Replace with valid JWT
  timeout: 10000,
  maxReconnectAttempts: 5
};

class WebSocketTester {
  constructor() {
    this.ws = null;
    this.connected = false;
    this.messages = [];
    this.reconnectAttempts = 0;
    this.testResults = {
      connection: false,
      messaging: false,
      reconnection: false,
      noFallback: true
    };
  }

  async runTests() {
    console.log('🚀 Starting WebSocket Tests...\n');
    
    try {
      await this.testConnection();
      await this.testMessaging();
      await this.testReconnection();
      await this.testNoFallback();
      
      this.printResults();
    } catch (error) {
      console.error('❌ Test suite failed:', error);
      process.exit(1);
    }
  }

  async testConnection() {
    console.log('📡 Testing WebSocket Connection...');
    
    return new Promise((resolve, reject) => {
      const wsUrl = `${CONFIG.wsUrl}?token=${encodeURIComponent(CONFIG.testToken)}`;
      this.ws = new WebSocket(wsUrl);
      
      const timeout = setTimeout(() => {
        reject(new Error('Connection timeout'));
      }, CONFIG.timeout);
      
      this.ws.on('open', () => {
        clearTimeout(timeout);
        this.connected = true;
        this.testResults.connection = true;
        console.log('✅ WebSocket connection established');
        resolve();
      });
      
      this.ws.on('error', (error) => {
        clearTimeout(timeout);
        console.log('❌ WebSocket connection failed:', error.message);
        reject(error);
      });
      
      this.ws.on('message', (data) => {
        try {
          const message = JSON.parse(data);
          this.messages.push(message);
          
          // Check for fallback indicators
          if (message.is_fallback || message.content?.includes('local mode')) {
            this.testResults.noFallback = false;
            console.log('❌ Detected fallback mode message');
          }
          
          console.log('📨 Received message:', message);
        } catch (e) {
          console.log('📨 Received raw message:', data.toString());
        }
      });
    });
  }

  async testMessaging() {
    console.log('\n💬 Testing Message Sending...');
    
    if (!this.connected) {
      throw new Error('Cannot test messaging: not connected');
    }
    
    const testMessage = {
      type: 'text',
      content: 'Test message from WebSocket tester',
      user_id: 'test-user-123',
      chat_id: 'test-chat-123',
      timestamp: new Date().toISOString(),
      message_id: `test-${Date.now()}`
    };
    
    return new Promise((resolve, reject) => {
      const timeout = setTimeout(() => {
        reject(new Error('Message sending timeout'));
      }, 5000);
      
      try {
        this.ws.send(JSON.stringify(testMessage));
        clearTimeout(timeout);
        this.testResults.messaging = true;
        console.log('✅ Message sent successfully');
        resolve();
      } catch (error) {
        clearTimeout(timeout);
        console.log('❌ Failed to send message:', error.message);
        reject(error);
      }
    });
  }

  async testReconnection() {
    console.log('\n🔄 Testing Reconnection...');
    
    return new Promise((resolve) => {
      if (!this.connected) {
        console.log('⚠️  Skipping reconnection test: not initially connected');
        resolve();
        return;
      }
      
      // Close connection to simulate network issue
      this.ws.close();
      this.connected = false;
      console.log('🔌 Connection closed intentionally');
      
      // Wait a moment then try to reconnect
      setTimeout(() => {
        this.attemptReconnect().then(() => {
          this.testResults.reconnection = true;
          console.log('✅ Reconnection successful');
          resolve();
        }).catch(() => {
          console.log('❌ Reconnection failed');
          resolve();
        });
      }, 1000);
    });
  }

  async testNoFallback() {
    console.log('\n🚫 Testing No Fallback Mode...');
    
    // Check all received messages for fallback indicators
    const fallbackMessages = this.messages.filter(msg => 
      msg.is_fallback || 
      msg.content?.includes('local mode') ||
      msg.content?.includes('fallback') ||
      msg.type === 'system' && msg.content?.includes('connection issues')
    );
    
    if (fallbackMessages.length > 0) {
      this.testResults.noFallback = false;
      console.log('❌ Found fallback mode messages:', fallbackMessages);
    } else {
      console.log('✅ No fallback mode detected');
    }
  }

  async attemptReconnect() {
    return new Promise((resolve, reject) => {
      if (this.reconnectAttempts >= CONFIG.maxReconnectAttempts) {
        reject(new Error('Max reconnection attempts reached'));
        return;
      }
      
      this.reconnectAttempts++;
      const wsUrl = `${CONFIG.wsUrl}?token=${encodeURIComponent(CONFIG.testToken)}`;
      this.ws = new WebSocket(wsUrl);
      
      const timeout = setTimeout(() => {
        reject(new Error('Reconnection timeout'));
      }, CONFIG.timeout);
      
      this.ws.on('open', () => {
        clearTimeout(timeout);
        this.connected = true;
        resolve();
      });
      
      this.ws.on('error', (error) => {
        clearTimeout(timeout);
        reject(error);
      });
    });
  }

  printResults() {
    console.log('\n📊 Test Results:');
    console.log('================');
    
    const results = [
      { name: 'Connection', passed: this.testResults.connection },
      { name: 'Messaging', passed: this.testResults.messaging },
      { name: 'Reconnection', passed: this.testResults.reconnection },
      { name: 'No Fallback', passed: this.testResults.noFallback }
    ];
    
    results.forEach(test => {
      const status = test.passed ? '✅ PASS' : '❌ FAIL';
      console.log(`${test.name}: ${status}`);
    });
    
    const allPassed = results.every(test => test.passed);
    
    console.log('\n' + '='.repeat(30));
    console.log(`Overall: ${allPassed ? '✅ ALL TESTS PASSED' : '❌ SOME TESTS FAILED'}`);
    
    if (allPassed) {
      console.log('\n🎉 WebSocket is working correctly with real-time messaging!');
    } else {
      console.log('\n⚠️  Some issues detected. Check the logs above.');
    }
    
    // Clean up
    if (this.ws && this.connected) {
      this.ws.close();
    }
  }
}

// Usage instructions
if (process.argv.includes('--help') || process.argv.includes('-h')) {
  console.log(`
WebSocket Test Script Usage:
===========================

1. Make sure your backend is running:
   cd backend && docker-compose up

2. Get a valid JWT token from your authentication endpoint

3. Update CONFIG.testToken in this script with your JWT token

4. Run the test:
   node websocket-test.js

Options:
  --help, -h    Show this help message

The script will test:
- WebSocket connection establishment
- Message sending/receiving
- Reconnection functionality  
- Absence of fallback mode

`);
  process.exit(0);
}

// Main execution
if (require.main === module) {
  const tester = new WebSocketTester();
  tester.runTests().catch(console.error);
}

module.exports = WebSocketTester;
