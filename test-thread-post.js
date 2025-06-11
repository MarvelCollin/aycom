// Simple script to test thread posting with detailed logs
// Run with: node --no-warnings test-thread-post.js

// Configuration
const API_URL = 'http://localhost:8083/api/v1';
const EMAIL = 'kolina@gmail.com';
const PASSWORD = 'Miawmiaw123@';

async function testThreadPosting() {
  try {
    console.log('1. Attempting to login...');
    
    // Step 1: Login to get token
    const loginResponse = await fetch(`${API_URL}/auth/login`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        email: EMAIL,
        password: PASSWORD
      })
    });
    
    if (!loginResponse.ok) {
      throw new Error(`Login failed with status ${loginResponse.status}`);
    }
    
    const loginData = await loginResponse.json();
    const token = loginData.token;
    
    console.log('Login successful!');
    console.log(`Token: ${token.substring(0, 20)}...${token.substring(token.length - 20)}`);
    console.log(`Token length: ${token.length} characters`);
    console.log(`User ID: ${loginData.user?.id || 'unknown'}`);
    
    // Step 2: Post a thread
    console.log('\n2. Attempting to post a thread...');
    
    // Log the exact request we're about to make
    console.log('Request details:');
    console.log(`URL: ${API_URL}/threads`);
    console.log('Headers:');
    console.log(`  Authorization: Bearer ${token.substring(0, 10)}...`);
    console.log('  Content-Type: application/json');
    console.log('Body:');
    console.log('  {"content":"Test thread from diagnostic script"}');
    
    const threadResponse = await fetch(`${API_URL}/threads`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        content: 'Test thread from diagnostic script'
      })
    });
    
    console.log(`\nThread post response status: ${threadResponse.status} ${threadResponse.statusText}`);
    
    const responseHeaders = {};
    threadResponse.headers.forEach((value, name) => {
      responseHeaders[name] = value;
    });
    
    console.log('Response headers:', responseHeaders);
    
    const responseText = await threadResponse.text();
    console.log('Response body:', responseText);
    
    if (!threadResponse.ok) {
      console.error('Thread post failed!');
      
      // Try to parse as JSON if possible
      try {
        const errorData = JSON.parse(responseText);
        console.error('Error details:', errorData);
      } catch (e) {
        // Text wasn't valid JSON
      }
    } else {
      console.log('Thread posted successfully!');
    }
    
  } catch (error) {
    console.error('Error in test script:', error);
  }
}

// Run the test
testThreadPosting(); 