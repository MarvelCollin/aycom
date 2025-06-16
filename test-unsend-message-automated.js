/**
 * Automated test script for the unsend message feature
 * Run with: node test-unsend-message-automated.js
 */

const https = require('https');
const http = require('http');

class UnsendMessageTester {
    constructor(baseUrl, authToken) {
        this.baseUrl = baseUrl.replace(/\/$/, ''); // Remove trailing slash
        this.authToken = authToken;
        this.testResults = [];
    }

    async makeRequest(endpoint, method = 'GET', body = null) {
        return new Promise((resolve, reject) => {
            const url = `${this.baseUrl}${endpoint}`;
            const isHttps = url.startsWith('https');
            const requestLib = isHttps ? https : http;
            
            const options = {
                method,
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${this.authToken}`
                }
            };

            const req = requestLib.request(url, options, (res) => {
                let data = '';
                res.on('data', chunk => data += chunk);
                res.on('end', () => {
                    try {
                        const parsedData = data ? JSON.parse(data) : {};
                        resolve({ status: res.statusCode, data: parsedData, headers: res.headers });
                    } catch (e) {
                        resolve({ status: res.statusCode, data: data, headers: res.headers });
                    }
                });
            });

            req.on('error', reject);
            
            if (body) {
                req.write(JSON.stringify(body));
            }
            
            req.end();
        });
    }

    log(message, type = 'INFO') {
        const timestamp = new Date().toISOString();
        console.log(`[${timestamp}] ${type}: ${message}`);
    }

    async testConnection() {
        this.log('Testing API connection...');
        try {
            const result = await this.makeRequest('/health');
            if (result.status === 200) {
                this.log('✅ API connection successful', 'SUCCESS');
                return true;
            } else {
                this.log(`❌ API connection failed with status ${result.status}`, 'ERROR');
                return false;
            }
        } catch (error) {
            this.log(`❌ API connection error: ${error.message}`, 'ERROR');
            return false;
        }
    }

    async sendTestMessage(chatId) {
        this.log('Sending test message...');
        const messageBody = {
            chat_id: chatId,
            content: `Test message for unsend - ${Date.now()}`,
            message_type: 'text'
        };

        try {
            const result = await this.makeRequest('/messages', 'POST', messageBody);
            if (result.status === 200 || result.status === 201) {
                this.log(`✅ Test message sent with ID: ${result.data.message_id}`, 'SUCCESS');
                return result.data.message_id;
            } else {
                this.log(`❌ Failed to send test message: ${result.status} - ${JSON.stringify(result.data)}`, 'ERROR');
                return null;
            }
        } catch (error) {
            this.log(`❌ Error sending test message: ${error.message}`, 'ERROR');
            return null;
        }
    }

    async testUnsendRealMessage(chatId) {
        this.log('=== Test 1: Unsend Real Message ===');
        
        // Send a message first
        const messageId = await this.sendTestMessage(chatId);
        if (!messageId) {
            this.testResults.push({ test: 'Unsend Real Message', status: 'FAILED', reason: 'Could not send test message' });
            return false;
        }

        // Wait for message to be processed
        await new Promise(resolve => setTimeout(resolve, 1000));

        // Try to unsend it
        this.log(`Attempting to unsend message ${messageId}...`);
        try {
            const result = await this.makeRequest(`/messages/${messageId}?chat_id=${chatId}`, 'DELETE');
            
            if (result.status === 200) {
                this.log('✅ Real message unsend test PASSED', 'SUCCESS');
                this.testResults.push({ test: 'Unsend Real Message', status: 'PASSED' });
                return true;
            } else {
                this.log(`❌ Real message unsend test FAILED: ${result.status} - ${JSON.stringify(result.data)}`, 'ERROR');
                this.testResults.push({ test: 'Unsend Real Message', status: 'FAILED', reason: `Status ${result.status}` });
                return false;
            }
        } catch (error) {
            this.log(`❌ Real message unsend test ERROR: ${error.message}`, 'ERROR');
            this.testResults.push({ test: 'Unsend Real Message', status: 'ERROR', reason: error.message });
            return false;
        }
    }

    async testUnsendTempMessage(chatId) {
        this.log('=== Test 2: Unsend Temp Message ===');
        
        const tempMessageId = `temp_${Date.now()}`;
        this.log(`Attempting to unsend temp message ${tempMessageId}...`);
        
        try {
            const result = await this.makeRequest(`/messages/${tempMessageId}?chat_id=${chatId}`, 'DELETE');
            
            if (result.status === 200 || result.status === 404) {
                this.log('✅ Temp message unsend test PASSED (handled gracefully)', 'SUCCESS');
                this.testResults.push({ test: 'Unsend Temp Message', status: 'PASSED' });
                return true;
            } else {
                this.log(`❌ Temp message unsend test FAILED: ${result.status} - ${JSON.stringify(result.data)}`, 'ERROR');
                this.testResults.push({ test: 'Unsend Temp Message', status: 'FAILED', reason: `Status ${result.status}` });
                return false;
            }
        } catch (error) {
            this.log(`❌ Temp message unsend test ERROR: ${error.message}`, 'ERROR');
            this.testResults.push({ test: 'Unsend Temp Message', status: 'ERROR', reason: error.message });
            return false;
        }
    }

    async testUnsendNonExistentMessage(chatId) {
        this.log('=== Test 3: Unsend Non-existent Message ===');
        
        const fakeMessageId = 'nonexistent_message_99999';
        this.log(`Attempting to unsend non-existent message ${fakeMessageId}...`);
        
        try {
            const result = await this.makeRequest(`/messages/${fakeMessageId}?chat_id=${chatId}`, 'DELETE');
            
            if (result.status === 404) {
                this.log('✅ Non-existent message unsend test PASSED (returned 404)', 'SUCCESS');
                this.testResults.push({ test: 'Unsend Non-existent Message', status: 'PASSED' });
                return true;
            } else {
                this.log(`❌ Non-existent message unsend test FAILED: ${result.status} - ${JSON.stringify(result.data)}`, 'ERROR');
                this.testResults.push({ test: 'Unsend Non-existent Message', status: 'FAILED', reason: `Status ${result.status}` });
                return false;
            }
        } catch (error) {
            this.log(`❌ Non-existent message unsend test ERROR: ${error.message}`, 'ERROR');
            this.testResults.push({ test: 'Unsend Non-existent Message', status: 'ERROR', reason: error.message });
            return false;
        }
    }

    async runAllTests(chatId) {
        this.log('=== Starting Comprehensive Unsend Message Tests ===');
        
        // Test connection first
        const connectionOk = await this.testConnection();
        if (!connectionOk) {
            this.log('❌ Cannot proceed with tests - API connection failed', 'ERROR');
            return;
        }

        await this.testUnsendRealMessage(chatId);
        await new Promise(resolve => setTimeout(resolve, 1000));
        
        await this.testUnsendTempMessage(chatId);
        await new Promise(resolve => setTimeout(resolve, 1000));
        
        await testUnsendNonExistentMessage(chatId);
        
        this.log('=== Test Results Summary ===');
        this.testResults.forEach(result => {
            const status = result.status === 'PASSED' ? '✅' : (result.status === 'FAILED' ? '❌' : '⚠️');
            const reason = result.reason ? ` (${result.reason})` : '';
            this.log(`${status} ${result.test}: ${result.status}${reason}`);
        });
        
        const passedTests = this.testResults.filter(r => r.status === 'PASSED').length;
        const totalTests = this.testResults.length;
        this.log(`Overall: ${passedTests}/${totalTests} tests passed`);
    }
}

// Configuration
const CONFIG = {
    API_BASE_URL: process.env.API_BASE_URL || 'http://localhost:8080/api/v1',
    AUTH_TOKEN: process.env.AUTH_TOKEN || '',
    CHAT_ID: process.env.CHAT_ID || ''
};

// Main execution
async function main() {
    console.log('Unsend Message Feature Test Suite');
    console.log('==================================');
    
    if (!CONFIG.AUTH_TOKEN) {
        console.log('❌ AUTH_TOKEN environment variable is required');
        console.log('Usage: AUTH_TOKEN=your_token CHAT_ID=your_chat_id node test-unsend-message-automated.js');
        process.exit(1);
    }
    
    if (!CONFIG.CHAT_ID) {
        console.log('❌ CHAT_ID environment variable is required');
        console.log('Usage: AUTH_TOKEN=your_token CHAT_ID=your_chat_id node test-unsend-message-automated.js');
        process.exit(1);
    }
    
    console.log(`API Base URL: ${CONFIG.API_BASE_URL}`);
    console.log(`Chat ID: ${CONFIG.CHAT_ID}`);
    console.log('');
    
    const tester = new UnsendMessageTester(CONFIG.API_BASE_URL, CONFIG.AUTH_TOKEN);
    await tester.runAllTests(CONFIG.CHAT_ID);
}

if (require.main === module) {
    main().catch(console.error);
}

module.exports = UnsendMessageTester;
