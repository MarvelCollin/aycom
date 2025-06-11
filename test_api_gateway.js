const axios = require('axios');
const uuid = require('uuid');

// Constants
const API_URL = 'http://localhost:8083';

// Helper function to log steps
function log(message) {
  console.log(`\n===== ${message} =====`);
}

// Helper to make requests to API gateway
async function apiRequest(method, path, data = null, token = null) {
  try {
    const config = {
      method,
      url: `${API_URL}${path}`,
      headers: {
        'Content-Type': 'application/json'
      }
    };
    
    if (token) {
      config.headers['Authorization'] = `Bearer ${token}`;
    }
    
    if (data) {
      config.data = data;
    }
    
    const response = await axios(config);
    return response.data;
  } catch (error) {
    console.error('API request failed:', error.response ? error.response.data : error.message);
    throw error;
  }
}

// Function to get all communities
async function getCommunities() {
  log('Getting all communities');
  try {
    const response = await apiRequest('get', '/api/v1/communities');
    return response.data.communities;
  } catch (error) {
    console.error('Failed to get communities:', error);
    return [];
  }
}

// Function to create a fake token with admin permissions
function createFakeToken(userId, isAdmin = false) {
  // Create a simple JWT-like structure (not a real JWT, just for testing)
  const header = { alg: "HS256", typ: "JWT" };
  const payload = {
    sub: userId,
    is_admin: isAdmin,
    exp: Math.floor(Date.now() / 1000) + 3600
  };
  
  // Convert to base64
  const base64Header = Buffer.from(JSON.stringify(header)).toString('base64');
  const base64Payload = Buffer.from(JSON.stringify(payload)).toString('base64');
  const signature = "fake_signature";
  
  return `${base64Header}.${base64Payload}.${signature}`;
}

// Main test function to check if the database update works
async function testJoinRequestFlow() {
  try {
    // Step 1: Get a community to test with
    const communities = await getCommunities();
    if (communities.length === 0) {
      throw new Error('No communities found to test with');
    }
    
    const testCommunity = communities[0];
    log(`Using community: ${testCommunity.name} (${testCommunity.id})`);
    
    // Step 2: Create user IDs for testing
    const adminUserId = uuid.v4();
    const regularUserId = uuid.v4();
    log(`Created admin user ID: ${adminUserId}`);
    log(`Created regular user ID: ${regularUserId}`);
    
    // Step 3: Create fake tokens
    const adminToken = createFakeToken(adminUserId, true);
    const userToken = createFakeToken(regularUserId, false);
    
    // Step 4: Have the regular user request to join the community
    log('Regular user requesting to join community');
    try {
      const joinResponse = await apiRequest(
        'post', 
        `/api/v1/communities/${testCommunity.id}/join`,
        {},
        userToken
      );
      console.log('Join request response:', joinResponse);
    } catch (error) {
      console.log('Note: Join request might have failed due to authorization issues. We will proceed with the test anyway.');
    }
    
    // Step 5: Get join requests for this community
    log('Getting join requests for community');
    try {
      const joinRequestsResponse = await apiRequest(
        'get',
        `/api/v1/communities/${testCommunity.id}/join-requests`,
        null,
        adminToken
      );
      console.log('Join requests:', joinRequestsResponse);
      
      if (joinRequestsResponse && 
          joinRequestsResponse.join_requests && 
          joinRequestsResponse.join_requests.length > 0) {
        
        const joinRequest = joinRequestsResponse.join_requests[0];
        const joinRequestId = joinRequest.id;
        
        // Step 6: Approve a join request
        log(`Approving join request: ${joinRequestId}`);
        try {
          const approveResponse = await apiRequest(
            'post',
            `/api/v1/communities/${testCommunity.id}/join-requests/${joinRequestId}/approve`,
            {},
            adminToken
          );
          console.log('Approval response:', approveResponse);
          
          // Step 7: Check if the user is now a member
          log('Checking if user is now a member');
          try {
            const membersResponse = await apiRequest(
              'get',
              `/api/v1/communities/${testCommunity.id}/members`,
              null,
              adminToken
            );
            console.log('Members:', membersResponse);
            
            const userIsMember = membersResponse.members.some(
              member => member.user_id === joinRequest.user_id
            );
            
            if (userIsMember) {
              log('TEST PASSED! ✅ User was successfully added as a member after join request approval');
            } else {
              log('TEST FAILED! ❌ User was not added as a member after join request approval');
            }
          } catch (error) {
            log('Failed to check members');
          }
        } catch (error) {
          log('Failed to approve join request');
        }
      } else {
        log('No join requests found to approve');
      }
    } catch (error) {
      log('Failed to get join requests');
    }
    
  } catch (error) {
    log('TEST FAILED WITH ERROR! ❌');
    console.error(error);
  }
}

// Run the test
testJoinRequestFlow(); 