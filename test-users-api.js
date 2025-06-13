// Test script to verify the API call
// Run this in browser console to test the getAllUsers functionality

import { getAllUsers } from './src/api/user.ts';

async function testGetAllUsers() {
  try {
    console.log('Testing getAllUsers API call...');
    const result = await getAllUsers(1, 50);
    console.log('Result:', result);
    
    if (result && result.users && Array.isArray(result.users)) {
      console.log(`Success! Found ${result.users.length} users`);
      console.log('Sample user:', result.users[0]);
    } else {
      console.error('Invalid response format:', result);
    }
  } catch (error) {
    console.error('Test failed:', error);
  }
}

// For browser console testing
if (typeof window !== 'undefined') {
  window.testGetAllUsers = testGetAllUsers;
}

export { testGetAllUsers };
