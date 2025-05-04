// Test script to simulate the frontend authentication flow
// This is a simplified version of what's happening in the useAuth hook

const API_URL = 'http://localhost:8081/api/v1';
const token = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NDYzNTY3ODIsImlhdCI6MTc0NjM1MzE4MiwidHlwZSI6ImFjY2VzcyIsInVzZXJfaWQiOiJjZTA4MDViZS1lNjVkLTRhMzEtYmExOS1kOTg5ZDY0NzUwMTkifQ.OGrTghvpSJnztP0anCBBOe1351j47-uYVmQ4GBmfDUE';

// Simulates the getCurrentUser method in useAuth.ts
async function getCurrentUser() {
  try {
    console.log('Fetching current user profile...');
    
    const response = await fetch(`${API_URL}/users/me`, {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      }
    });
    
    console.log(`Response status: ${response.status}`);
    
    // If unauthorized, we would try to refresh the token here
    if (response.status === 401) {
      console.log('Received 401 Unauthorized response');
      // In a real implementation, we would try to refresh the token
      return { success: false, message: 'Not authenticated' };
    }
    
    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      console.error('Error data:', errorData);
      throw new Error(errorData.message || `Failed to get user profile: ${response.status}`);
    }
    
    const data = await response.json();
    console.log('User profile data:', data);
    return data;
  } catch (error) {
    console.error('Get current user error:', error);
    return { success: false, message: error.message };
  }
}

// Simulates the validateAuth method in useAuth.ts
async function validateAuth() {
  try {
    console.log('Validating current authentication with backend...');
    
    // Make a direct call to check the token
    const response = await fetch(`${API_URL}/users/me`, {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      }
    });
    
    console.log(`Auth validation response status: ${response.status}`);
    
    if (response.status === 401) {
      console.log('Auth validation failed with 401 Unauthorized');
      return false;
    }
    
    if (!response.ok) {
      const errorText = await response.text().catch(() => '');
      console.error(`Auth validation failed with status ${response.status}`, { errorText });
      return false;
    }
    
    const data = await response.json();
    console.log('Auth validation successful:', data);
    return true;
  } catch (error) {
    console.error('Auth validation error:', error);
    return false;
  }
}

// Run both tests
async function runTests() {
  console.log('Running getCurrentUser test...');
  const userResult = await getCurrentUser();
  console.log('getCurrentUser result:', userResult);
  
  console.log('\n----------------------------\n');
  
  console.log('Running validateAuth test...');
  const authResult = await validateAuth();
  console.log('validateAuth result:', authResult);
}

runTests(); 