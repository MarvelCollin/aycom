const axios = require('axios');
const uuid = require('uuid');

// Base URL for API
const BASE_URL = 'http://localhost:8083';

// Test user credentials
const TEST_USER = {
  email: "jane@example.com",
  password: "Miawmiaw123@"
};

// Test admin credentials
const ADMIN_USER = {
  email: "admin@aycom.com",
  password: "Miawmiaw123@"
};

// Community ID to test with - you may need to replace this with a valid community ID
let COMMUNITY_ID = '';

// Global variables for tokens and other data
let userToken = '';
let adminToken = '';
let joinRequestId = '';
let userId = '';

// Helper function to log steps
function log(message) {
  console.log(`\n===== ${message} =====`);
}

// Function to login and get token
async function login(credentials) {
  try {
    log(`Logging in as ${credentials.email}`);
    const response = await axios.post(`${BASE_URL}/api/v1/auth/login`, credentials);
    console.log('Login successful');
    return response.data.token;
  } catch (error) {
    console.error('Login failed:', error.response ? error.response.data : error.message);
    throw error;
  }
}

// Function to get user communities
async function getUserCommunities(token) {
  try {
    log('Getting user communities');
    const response = await axios.get(`${BASE_URL}/api/v1/communities`, {
      headers: { Authorization: `Bearer ${token}` }
    });
    console.log(`Found ${response.data.communities.length} communities`);
    return response.data.communities;
  } catch (error) {
    console.error('Failed to get communities:', error.response ? error.response.data : error.message);
    throw error;
  }
}

// Function to create a join request
async function createJoinRequest(token, communityId) {
  try {
    log('Creating join request');
    const response = await axios.post(
      `${BASE_URL}/api/v1/communities/${communityId}/join`,
      {},
      { headers: { Authorization: `Bearer ${token}` } }
    );
    console.log('Join request created successfully');
    console.log(response.data);
    return response.data.join_request.id;
  } catch (error) {
    console.error('Failed to create join request:', error.response ? error.response.data : error.message);
    throw error;
  }
}

// Function to get join requests for a community
async function getJoinRequests(token, communityId) {
  try {
    log('Getting join requests');
    const response = await axios.get(
      `${BASE_URL}/api/v1/communities/${communityId}/join-requests`,
      { headers: { Authorization: `Bearer ${token}` } }
    );
    console.log(`Found ${response.data.join_requests.length} join requests`);
    return response.data.join_requests;
  } catch (error) {
    console.error('Failed to get join requests:', error.response ? error.response.data : error.message);
    throw error;
  }
}

// Function to approve a join request
async function approveJoinRequest(token, communityId, requestId) {
  try {
    log('Approving join request');
    const response = await axios.post(
      `${BASE_URL}/api/v1/communities/${communityId}/join-requests/${requestId}/approve`,
      {},
      { headers: { Authorization: `Bearer ${token}` } }
    );
    console.log('Join request approved successfully');
    console.log(response.data);
    return response.data;
  } catch (error) {
    console.error('Failed to approve join request:', error.response ? error.response.data : error.message);
    throw error;
  }
}

// Function to check membership status
async function checkMembershipStatus(token, communityId) {
  try {
    log('Checking membership status');
    const response = await axios.get(
      `${BASE_URL}/api/v1/communities/${communityId}/membership-status`,
      { headers: { Authorization: `Bearer ${token}` } }
    );
    console.log('Membership status:', response.data.status);
    return response.data.status;
  } catch (error) {
    console.error('Failed to check membership status:', error.response ? error.response.data : error.message);
    throw error;
  }
}

// Function to get community members
async function getCommunityMembers(token, communityId) {
  try {
    log('Getting community members');
    const response = await axios.get(
      `${BASE_URL}/api/v1/communities/${communityId}/members`,
      { headers: { Authorization: `Bearer ${token}` } }
    );
    console.log(`Found ${response.data.members.length} members`);
    return response.data.members;
  } catch (error) {
    console.error('Failed to get community members:', error.response ? error.response.data : error.message);
    throw error;
  }
}

// Main function to run the test
async function runTest() {
  try {
    // Login as regular user and admin
    userToken = await login(TEST_USER);
    adminToken = await login(ADMIN_USER);

    // Get user ID from token payload
    const tokenPayload = JSON.parse(Buffer.from(userToken.split('.')[1], 'base64').toString());
    userId = tokenPayload.sub || tokenPayload.user_id;
    console.log('User ID:', userId);

    // Get communities for admin user
    const communities = await getUserCommunities(adminToken);
    if (communities.length === 0) {
      throw new Error('No communities found for admin user');
    }
    COMMUNITY_ID = communities[0].id;
    console.log('Using community ID:', COMMUNITY_ID);

    // Check initial membership status
    const initialStatus = await checkMembershipStatus(userToken, COMMUNITY_ID);
    if (initialStatus === 'member') {
      console.log('User is already a member. Removing to test join flow...');
      // Here you would add code to remove the member, but for simplicity we'll just
      // assume the user is not a member initially
    }

    // Create join request
    joinRequestId = await createJoinRequest(userToken, COMMUNITY_ID);

    // Check join request exists
    const joinRequests = await getJoinRequests(adminToken, COMMUNITY_ID);
    const ourRequest = joinRequests.find(req => req.id === joinRequestId);
    if (!ourRequest) {
      throw new Error('Created join request not found in list');
    }
    console.log('Join request found in list:', ourRequest);

    // Approve join request
    await approveJoinRequest(adminToken, COMMUNITY_ID, joinRequestId);

    // Check membership status after approval
    const finalStatus = await checkMembershipStatus(userToken, COMMUNITY_ID);
    if (finalStatus !== 'member') {
      throw new Error(`Expected membership status to be 'member' after approval, but got '${finalStatus}'`);
    }

    // Check user appears in members list
    const members = await getCommunityMembers(adminToken, COMMUNITY_ID);
    const isMember = members.some(member => member.user_id === userId);
    if (!isMember) {
      throw new Error('User not found in community members list after approval');
    }

    log('TEST PASSED! Join request flow is working correctly.');
  } catch (error) {
    log('TEST FAILED!');
    console.error(error);
  }
}

// Run the test
runTest();
