// Simple test script to check if the API Gateway is correctly handling authentication
const token = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NDYzNTY3ODIsImlhdCI6MTc0NjM1MzE4MiwidHlwZSI6ImFjY2VzcyIsInVzZXJfaWQiOiJjZTA4MDViZS1lNjVkLTRhMzEtYmExOS1kOTg5ZDY0NzUwMTkifQ.OGrTghvpSJnztP0anCBBOe1351j47-uYVmQ4GBmfDUE';

// Function to make a request to the /users/me endpoint
async function testUsersMeEndpoint() {
  try {
    console.log('Testing /users/me endpoint...');
    const response = await fetch('http://localhost:8081/api/v1/users/me', {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      }
    });

    console.log(`Response status: ${response.status}`);
    
    if (response.ok) {
      const data = await response.json();
      console.log('Response data:', data);
    } else {
      try {
        const errorData = await response.json();
        console.error('Error response:', errorData);
      } catch (e) {
        console.error('Failed to parse error response');
        const text = await response.text();
        console.error('Error response text:', text);
      }
    }
  } catch (error) {
    console.error('Fetch error:', error);
  }
}

// Function to test the /users/profile endpoint
async function testUsersProfileEndpoint() {
  try {
    console.log('Testing /users/profile endpoint...');
    const response = await fetch('http://localhost:8081/api/v1/users/profile', {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      }
    });

    console.log(`Response status: ${response.status}`);
    
    if (response.ok) {
      const data = await response.json();
      console.log('Response data:', data);
    } else {
      try {
        const errorData = await response.json();
        console.error('Error response:', errorData);
      } catch (e) {
        console.error('Failed to parse error response');
        const text = await response.text();
        console.error('Error response text:', text);
      }
    }
  } catch (error) {
    console.error('Fetch error:', error);
  }
}

// Function to test token refresh
async function testTokenRefresh() {
  try {
    console.log('Testing token refresh...');
    const refreshToken = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NDY5NTc5ODIsImlhdCI6MTc0NjM1MzE4MiwidHlwZSI6InJlZnJlc2giLCJ1c2VyX2lkIjoiY2UwODA1YmUtZTY1ZC00YTMxLWJhMTktZDk4OWQ2NDc1MDE5In0.hn6OjHHk32OlPgspMa8bpvJkbEWMzYBeqz8E93XtWKw';
    
    const response = await fetch('http://localhost:8081/api/v1/auth/refresh-token', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({ refresh_token: refreshToken })
    });

    console.log(`Response status: ${response.status}`);
    
    if (response.ok) {
      const data = await response.json();
      console.log('Response data:', data);
    } else {
      try {
        const errorData = await response.json();
        console.error('Error response:', errorData);
      } catch (e) {
        console.error('Failed to parse error response');
        const text = await response.text();
        console.error('Error response text:', text);
      }
    }
  } catch (error) {
    console.error('Fetch error:', error);
  }
}

// Run all tests
async function runTests() {
  await testUsersMeEndpoint();
  console.log('\n----------------------------\n');
  await testUsersProfileEndpoint();
  console.log('\n----------------------------\n');
  await testTokenRefresh();
}

runTests(); 