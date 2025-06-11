const axios = require('axios');

// Base URL for API
const BASE_URL = 'http://localhost:8083';

// Try different user credentials to find which ones work
const testCredentials = [
  { email: "admin@aycom.com", password: "Miawmiaw123@" },
  { email: "admin@example.com", password: "password123" },
  { email: "john@example.com", password: "Miawmiaw123@" },
  { email: "jane@example.com", password: "Miawmiaw123@" },
  { email: "sam@example.com", password: "Miawmiaw123@" },
  { email: "techguru@example.com", password: "Miawmiaw123@" }
];

// Function to login and get token
async function login(credentials) {
  try {
    console.log(`\nTrying to login as ${credentials.email}`);
    const response = await axios.post(`${BASE_URL}/api/v1/auth/login`, credentials);
    console.log('✅ Login successful');
    return { success: true, token: response.data.token };
  } catch (error) {
    console.log('❌ Login failed:', error.response ? error.response.data : error.message);
    return { success: false };
  }
}

// Test all credentials
async function testAllCredentials() {
  console.log("Testing user credentials...");
  
  for (const creds of testCredentials) {
    const result = await login(creds);
    if (result.success) {
      console.log(`User ${creds.email} authentication successful!`);
      
      // Decode token to get user info
      if (result.token) {
        try {
          const tokenPayload = JSON.parse(Buffer.from(result.token.split('.')[1], 'base64').toString());
          console.log("User ID:", tokenPayload.sub || tokenPayload.user_id);
          console.log("Is Admin:", tokenPayload.is_admin === true);
        } catch (e) {
          console.log("Could not decode token:", e.message);
        }
      }
    }
  }
  
  console.log("\nCredential testing complete!");
}

// Run the test
testAllCredentials();
