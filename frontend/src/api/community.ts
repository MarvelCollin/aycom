import { apiRequest } from '../utils/apiClient';
import { createLoggerWithPrefix } from '../utils/logger';

const logger = createLoggerWithPrefix('CommunityAPI');

export async function createCommunity(data: Record<string, any>) {
  try {
    const response = await apiRequest('/communities', {
      method: 'POST',
      body: JSON.stringify(data)
    });
    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.message || 'Failed to create community');
    }
    return response.json();
  } catch (error) {
    logger.error('Create community failed:', error);
    throw error;
  }
}

export async function listCommunities() {
  try {
    const response = await apiRequest('/communities', { method: 'GET' });
    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.message || 'Failed to list communities');
    }
    return response.json();
  } catch (error) {
    logger.error('List communities failed:', error);
    throw error;
  }
}

export async function getCommunityById(id: string) {
  try {
    const response = await apiRequest(`/communities/${id}`, { method: 'GET' });
    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.message || 'Failed to get community');
    }
    return response.json();
  } catch (error) {
    logger.error(`Get community ${id} failed:`, error);
    throw error;
  }
}

export async function updateCommunity(id: string, data: Record<string, any>) {
  try {
    const response = await apiRequest(`/communities/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data)
    });
    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.message || 'Failed to update community');
    }
    return response.json();
  } catch (error) {
    logger.error(`Update community ${id} failed:`, error);
    throw error;
  }
}

export async function deleteCommunity(id: string) {
  try {
    const response = await apiRequest(`/communities/${id}`, { method: 'DELETE' });
    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.message || 'Failed to delete community');
    }
    return response.json();
  } catch (error) {
    logger.error(`Delete community ${id} failed:`, error);
    throw error;
  }
}

export async function approveCommunity(id: string) {
  try {
    const response = await apiRequest(`/communities/${id}/approve`, { method: 'POST' });
    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.message || 'Failed to approve community');
    }
    return response.json();
  } catch (error) {
    logger.error(`Approve community ${id} failed:`, error);
    throw error;
  }
}

export async function addMember(communityId: string, data: Record<string, any>) {
  try {
    const response = await apiRequest(`/communities/${communityId}/members`, {
      method: 'POST',
      body: JSON.stringify(data)
    });
    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.message || 'Failed to add member');
    }
    return response.json();
  } catch (error) {
    logger.error(`Add member to community ${communityId} failed:`, error);
    throw error;
  }
}

export async function listMembers(communityId: string) {
  try {
    const response = await apiRequest(`/communities/${communityId}/members`, { method: 'GET' });
    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.message || 'Failed to list members');
    }
    return response.json();
  } catch (error) {
    logger.error(`List members for community ${communityId} failed:`, error);
    throw error;
  }
}

export async function updateMemberRole(communityId: string, userId: string, data: Record<string, any>) {
  try {
    const response = await apiRequest(`/communities/${communityId}/members/${userId}`, {
      method: 'PUT',
      body: JSON.stringify(data)
    });
    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.message || 'Failed to update member role');
    }
    return response.json();
  } catch (error) {
    logger.error(`Update member role for user ${userId} in community ${communityId} failed:`, error);
    throw error;
  }
}

export async function removeMember(communityId: string, userId: string) {
  try {
    const response = await apiRequest(`/communities/${communityId}/members/${userId}`, { method: 'DELETE' });
    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.message || 'Failed to remove member');
    }
    return response.json();
  } catch (error) {
    logger.error(`Remove member ${userId} from community ${communityId} failed:`, error);
    throw error;
  }
}

export async function addRule(communityId: string, data: Record<string, any>) {
  try {
    const response = await apiRequest(`/communities/${communityId}/rules`, {
      method: 'POST',
      body: JSON.stringify(data)
    });
    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.message || 'Failed to add rule');
    }
    return response.json();
  } catch (error) {
    logger.error(`Add rule to community ${communityId} failed:`, error);
    throw error;
  }
}

export async function listRules(communityId: string) {
  try {
    const response = await apiRequest(`/communities/${communityId}/rules`, { method: 'GET' });
    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.message || 'Failed to list rules');
    }
    return response.json();
  } catch (error) {
    logger.error(`List rules for community ${communityId} failed:`, error);
    throw error;
  }
}

export async function removeRule(communityId: string, ruleId: string) {
  try {
    const response = await apiRequest(`/communities/${communityId}/rules/${ruleId}`, { method: 'DELETE' });
    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.message || 'Failed to remove rule');
    }
    return response.json();
  } catch (error) {
    logger.error(`Remove rule ${ruleId} from community ${communityId} failed:`, error);
    throw error;
  }
}

export async function requestToJoin(communityId: string, data: Record<string, any>) {
  try {
    const response = await apiRequest(`/communities/${communityId}/join-requests`, {
      method: 'POST',
      body: JSON.stringify(data)
    });
    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.message || 'Failed to request to join');
    }
    return response.json();
  } catch (error) {
    logger.error(`Request to join community ${communityId} failed:`, error);
    throw error;
  }
}

export async function listJoinRequests(communityId: string) {
  try {
    const response = await apiRequest(`/communities/${communityId}/join-requests`, { method: 'GET' });
    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.message || 'Failed to list join requests');
    }
    return response.json();
  } catch (error) {
    logger.error(`List join requests for community ${communityId} failed:`, error);
    throw error;
  }
}

export async function approveJoinRequest(communityId: string, requestId: string) {
  try {
    const response = await apiRequest(`/communities/${communityId}/join-requests/${requestId}/approve`, { method: 'POST' });
    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.message || 'Failed to approve join request');
    }
    return response.json();
  } catch (error) {
    logger.error(`Approve join request ${requestId} for community ${communityId} failed:`, error);
    throw error;
  }
}

export async function rejectJoinRequest(communityId: string, requestId: string) {
  try {
    const response = await apiRequest(`/communities/${communityId}/join-requests/${requestId}/reject`, { method: 'POST' });
    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.message || 'Failed to reject join request');
    }
    return response.json();
  } catch (error) {
    logger.error(`Reject join request ${requestId} for community ${communityId} failed:`, error);
    throw error;
  }
}

/**
 * Check if the current user is a member of a specific community
 * @param communityId The ID of the community to check membership for
 * @returns An object containing a boolean indicating membership status
 */
export async function checkUserCommunityMembership(communityId: string) {
  try {
    const response = await apiRequest(`/communities/${communityId}/check-membership`, { 
      method: 'GET'
    });
    
    if (!response.ok) {
      // If response status is 404, user is not a member
      if (response.status === 404) {
        return { isMember: false };
      }
      
      const errorData = await response.json();
      throw new Error(errorData.message || 'Failed to check community membership');
    }
    
    const data = await response.json();
    return { isMember: data.isMember || false };
  } catch (error) {
    logger.error(`Check membership for community ${communityId} failed:`, error);
    // Default to non-member in case of error
    return { isMember: false };
  }
}

// Search communities based on query
export async function searchCommunities(
  query: string, 
  page: number = 1, 
  limit: number = 10
) {
  try {
    const url = new URL(`${import.meta.env.VITE_API_BASE_URL || 'http://localhost:8083/api/v1'}/communities/search`);
    
    // Set query parameters
    url.searchParams.append('q', query);
    url.searchParams.append('page', page.toString());
    url.searchParams.append('limit', limit.toString());
    
    // Get token
    const token = localStorage.getItem('aycom_access_token');
    
    // Make request
    const response = await fetch(url.toString(), {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      }
    });
    
    if (!response.ok) {
      throw new Error(`Failed to search communities: ${response.status}`);
    }
    
    return await response.json();
  } catch (error) {
    console.error('Error searching communities:', error);
    // Mock data for development
    return {
      communities: [
        {
          id: 'c1',
          name: 'Tech Enthusiasts',
          description: 'A community for technology lovers and early adopters. We discuss the latest gadgets, software releases, and tech trends.',
          logo: null,
          member_count: 1247,
          is_joined: false,
          is_pending: false
        },
        {
          id: 'c2',
          name: 'Travel Adventures',
          description: 'Share your travel experiences, photos, tips, and recommendations. Connect with fellow travelers around the world!',
          logo: null,
          member_count: 3768,
          is_joined: true,
          is_pending: false
        },
        {
          id: 'c3',
          name: 'Coding Experts',
          description: 'A community dedicated to programming, software development, and coding best practices. Join to share knowledge and learn from others.',
          logo: null,
          member_count: 829,
          is_joined: false,
          is_pending: true
        }
      ]
    };
  }
}
