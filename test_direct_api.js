const axios = require('axios');
const uuid = require('uuid');

// Constants
const COMMUNITY_SERVICE_URL = 'http://localhost:9093'; // Direct to the community service, not API gateway
const API_GATEWAY_URL = 'http://localhost:8083';

// Helper function to log steps
function log(message) {
  console.log(`\n===== ${message} =====`);
}

// Helper to make requests to community service directly
async function communityServiceRequest(method, path, data = null) {
  try {
    const config = {
      method,
      url: `${COMMUNITY_SERVICE_URL}${path}`,
      headers: {
        'Content-Type': 'application/json'
      }
    };
    
    if (data) {
      config.data = data;
    }
    
    const response = await axios(config);
    return response.data;
  } catch (error) {
    console.error('Request failed:', error.response ? error.response.data : error.message);
    throw error;
  }
}

// Helper to make requests to API gateway
async function apiGatewayRequest(method, path) {
  try {
    const config = {
      method,
      url: `${API_GATEWAY_URL}${path}`,
      headers: {
        'Content-Type': 'application/json'
      }
    };
    
    const response = await axios(config);
    return response.data;
  } catch (error) {
    console.error('API Gateway request failed:', error.response ? error.response.data : error.message);
    throw error;
  }
}

// Function to get all communities
async function getCommunities() {
  log('Getting communities from API Gateway');
  try {
    const response = await apiGatewayRequest('get', '/api/v1/communities');
    return response.data.communities;
  } catch (error) {
    console.error('Failed to get communities:', error);
    return [];
  }
}

// Main test function
async function testJoinRequestApproval() {
  try {
    // Step 1: Get a community to test with
    const communities = await getCommunities();
    if (communities.length === 0) {
      throw new Error('No communities found to test with');
    }
    
    const testCommunity = communities[0];
    log(`Using community: ${testCommunity.name} (${testCommunity.id})`);
    
    // Step 2: Create a new user ID to test with
    const testUserId = uuid.v4();
    log(`Created test user ID: ${testUserId}`);
    
    // Step 3: Create a join request
    log('Creating join request directly through community service');
    const joinRequestResponse = await communityServiceRequest('post', '/community/request-to-join', {
      community_id: testCommunity.id,
      user_id: testUserId
    });
    
    console.log('Join request created:', joinRequestResponse);
    const joinRequestId = joinRequestResponse.join_request.id;
    
    // Step 4: Approve the join request
    log(`Approving join request (ID: ${joinRequestId})`);
    const approveResponse = await communityServiceRequest('post', '/community/approve-join-request', {
      join_request_id: joinRequestId
    });
    
    console.log('Approval response:', approveResponse);
    
    // Step 5: Verify the user is now a member
    log('Checking if user is now a member');
    const isMemberResponse = await communityServiceRequest('post', '/community/is-member', {
      community_id: testCommunity.id,
      user_id: testUserId
    });
    
    console.log('Is member response:', isMemberResponse);
    
    if (isMemberResponse.is_member) {
      log('TEST PASSED! ✅ User was successfully added as a member after join request approval');
    } else {
      log('TEST FAILED! ❌ User was not added as a member after join request approval');
    }
  } catch (error) {
    log('TEST FAILED WITH ERROR! ❌');
    console.error(error);
  }
}

// Run the test
testJoinRequestApproval(); 