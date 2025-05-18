import { getAuthToken } from '../utils/auth';
import appConfig from '../config/appConfig';
import { createLoggerWithPrefix } from '../utils/logger';

const API_BASE_URL = appConfig.api.baseUrl;
const logger = createLoggerWithPrefix('CommunityAPI');

/**
 * Get a formatted list of communities for UI components
 * @returns Object with success status and communities array (id, name)
 */
export async function getCommunities() {
  try {
    const response = await listCommunities();
    
    if (response && Array.isArray(response.communities)) {
      return {
        success: true,
        communities: response.communities.map((community: any) => ({
          id: community.id,
          name: community.name
        }))
      };
    }
    
    logger.warn('Unexpected response format from listCommunities', response);
    return {
      success: false,
      communities: [],
      error: 'Failed to parse communities data'
    };
  } catch (error) {
    logger.error('Failed to get communities:', error);
    return {
      success: false,
      communities: [],
      error: error instanceof Error ? error.message : 'Unknown error'
    };
  }
}

export async function createCommunity(data: Record<string, any>) {
  try {
    const token = getAuthToken();
    
    const response = await fetch(`${API_BASE_URL}/communities`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': token ? `Bearer ${token}` : ''
      },
      body: JSON.stringify(data),
      credentials: 'include'
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
    const token = getAuthToken();
    
    const response = await fetch(`${API_BASE_URL}/communities`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': token ? `Bearer ${token}` : ''
      },
      credentials: 'include'
    });
    
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
    const token = getAuthToken();
    
    const response = await fetch(`${API_BASE_URL}/communities/${id}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': token ? `Bearer ${token}` : ''
      },
      credentials: 'include'
    });
    
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
    const response = await fetch(`${API_BASE_URL}/communities/${id}`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': getAuthToken() ? `Bearer ${getAuthToken()}` : ''
      },
      body: JSON.stringify(data),
      credentials: 'include'
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
    const response = await fetch(`${API_BASE_URL}/communities/${id}`, {
      method: 'DELETE',
      headers: {
        'Authorization': getAuthToken() ? `Bearer ${getAuthToken()}` : ''
      },
      credentials: 'include'
    });
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
    const response = await fetch(`${API_BASE_URL}/communities/${id}/approve`, {
      method: 'POST',
      headers: {
        'Authorization': getAuthToken() ? `Bearer ${getAuthToken()}` : ''
      },
      credentials: 'include'
    });
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
    const response = await fetch(`${API_BASE_URL}/communities/${communityId}/members`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': getAuthToken() ? `Bearer ${getAuthToken()}` : ''
      },
      body: JSON.stringify(data),
      credentials: 'include'
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
    const response = await fetch(`${API_BASE_URL}/communities/${communityId}/members`, {
      method: 'GET',
      headers: {
        'Authorization': getAuthToken() ? `Bearer ${getAuthToken()}` : ''
      },
      credentials: 'include'
    });
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
    const response = await fetch(`${API_BASE_URL}/communities/${communityId}/members/${userId}`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': getAuthToken() ? `Bearer ${getAuthToken()}` : ''
      },
      body: JSON.stringify(data),
      credentials: 'include'
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
    const response = await fetch(`${API_BASE_URL}/communities/${communityId}/members/${userId}`, {
      method: 'DELETE',
      headers: {
        'Authorization': getAuthToken() ? `Bearer ${getAuthToken()}` : ''
      },
      credentials: 'include'
    });
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
    const response = await fetch(`${API_BASE_URL}/communities/${communityId}/rules`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': getAuthToken() ? `Bearer ${getAuthToken()}` : ''
      },
      body: JSON.stringify(data),
      credentials: 'include'
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
    const response = await fetch(`${API_BASE_URL}/communities/${communityId}/rules`, {
      method: 'GET',
      headers: {
        'Authorization': getAuthToken() ? `Bearer ${getAuthToken()}` : ''
      },
      credentials: 'include'
    });
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
    const response = await fetch(`${API_BASE_URL}/communities/${communityId}/rules/${ruleId}`, {
      method: 'DELETE',
      headers: {
        'Authorization': getAuthToken() ? `Bearer ${getAuthToken()}` : ''
      },
      credentials: 'include'
    });
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
    const response = await fetch(`${API_BASE_URL}/communities/${communityId}/join-requests`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': getAuthToken() ? `Bearer ${getAuthToken()}` : ''
      },
      body: JSON.stringify(data),
      credentials: 'include'
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
    const response = await fetch(`${API_BASE_URL}/communities/${communityId}/join-requests`, {
      method: 'GET',
      headers: {
        'Authorization': getAuthToken() ? `Bearer ${getAuthToken()}` : ''
      },
      credentials: 'include'
    });
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
    const response = await fetch(`${API_BASE_URL}/communities/${communityId}/join-requests/${requestId}/approve`, {
      method: 'POST',
      headers: {
        'Authorization': getAuthToken() ? `Bearer ${getAuthToken()}` : ''
      },
      credentials: 'include'
    });
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
    const response = await fetch(`${API_BASE_URL}/communities/${communityId}/join-requests/${requestId}/reject`, {
      method: 'POST',
      headers: {
        'Authorization': getAuthToken() ? `Bearer ${getAuthToken()}` : ''
      },
      credentials: 'include'
    });
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
    const response = await fetch(`${API_BASE_URL}/communities/${communityId}/check-membership`, {
      method: 'GET',
      headers: {
        'Authorization': getAuthToken() ? `Bearer ${getAuthToken()}` : ''
      },
      credentials: 'include'
    });
    
    if (!response.ok) {
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
    const url = new URL(`${API_BASE_URL}/communities/search`);
    
    url.searchParams.append('q', query);
    url.searchParams.append('page', page.toString());
    url.searchParams.append('limit', limit.toString());
    
    const token = getAuthToken();
    
    logger.debug('Searching communities', { query, page, limit });
    
    const response = await fetch(url.toString(), {
      method: 'GET',
      headers: {
        'Authorization': token ? `Bearer ${token}` : '',
        'Content-Type': 'application/json'
      },
      credentials: 'include'
    });
    
    if (!response.ok) {
      // Handle 401 Unauthorized by returning empty results instead of throwing
      if (response.status === 401) {
        logger.warn('Unauthorized when searching communities - returning empty results');
        return { communities: [], total_count: 0, page, limit };
      }
      
      // For 500 server errors, return empty results with a log message instead of throwing
      if (response.status === 500) {
        logger.error(`Server error (500) when searching communities - returning empty results`);
        return { communities: [], total_count: 0, page, limit };
      }
      
      const errorMessage = `Failed to search communities: ${response.status}`;
      logger.error(errorMessage);
      throw new Error(errorMessage);
    }
    
    const data = await response.json();
    logger.debug('Communities search results', { 
      count: data.communities?.length || 0,
      totalCount: data.total_count || 0
    });
    
    return data;
  } catch (error) {
    logger.error('Error searching communities:', error);
    // Return empty results instead of throwing to avoid breaking the UI
    return {
      communities: [],
      total_count: 0,
      page,
      limit
    };
  }
}
