import { getAuthToken } from '../utils/auth';
import appConfig from '../config/appConfig';
import { createLoggerWithPrefix } from '../utils/logger';
import type { ICategory } from './categories';

const API_BASE_URL = appConfig.api.baseUrl;
const logger = createLoggerWithPrefix('CommunityAPI');

export async function getCommunities(params = {}) {
  try {

    const queryParams = new URLSearchParams();
    Object.entries(params).forEach(([key, value]) => {
      if (Array.isArray(value)) {
        value.forEach(v => queryParams.append(key, v));
      } else if (value !== null && value !== undefined) {
        queryParams.append(key, value.toString());
      }
    });

    const token = getAuthToken();

    const response = await fetch(`${API_BASE_URL}/communities?${queryParams.toString()}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': token ? `Bearer ${token}` : ''
      },
      credentials: 'include'
    });

    if (!response.ok) {

      const text = await response.text();
      if (!text) {
        throw new Error(`HTTP error ${response.status}: Empty response`);
      }

      try {

        const errorData = JSON.parse(text);
        throw new Error(errorData.message || `Failed to list communities (${response.status})`);
      } catch (parseError) {

        throw new Error(`Failed to list communities (${response.status}): ${text.substring(0, 100)}`);
      }
    }

    const text = await response.text();
    if (!text) {
      return { 
        success: true,
        communities: [],
        pagination: {
          total_count: 0,
          current_page: 1,
          per_page: 25,
          total_pages: 0
        },
        limit_options: [25, 30, 35]
      };
    }

    try {
      const data = JSON.parse(text);

      const normalizedCommunities = (data.communities || []).map(community => ({
        id: community.id,
        name: community.name,
        description: community.description || '',
        logo_url: community.logo_url || community.logoUrl || community.avatar || '',
        banner_url: community.banner_url || community.bannerUrl || '',
        creator_id: community.creator_id || community.creatorId || '',
        is_approved: community.is_approved || community.isApproved || false,
        member_count: community.member_count || community.memberCount || 0,
        created_at: community.created_at || community.createdAt || ''
      }));

      const pagination = {
        total_count: data.pagination?.total_count || data.pagination?.total || data.pagination?.totalCount || 0,
        current_page: data.pagination?.current_page || data.pagination?.page || data.pagination?.currentPage || 1,
        per_page: data.pagination?.per_page || data.pagination?.limit || data.pagination?.perPage || 25,
        total_pages: data.pagination?.total_pages || data.pagination?.totalPages || 0
      };

      return {
        success: data.success,
        communities: normalizedCommunities,
        pagination: pagination,
        limit_options: data.limitOptions || data.limit_options || [25, 30, 35]
      };
    } catch (parseError) {
      logger.error('Failed to parse JSON response:', parseError);
      throw new Error('Invalid JSON response from server');
    }
  } catch (error) {
    logger.error('Get communities failed:', error);
    throw error;
  }
}

export async function getCommunityCategories(): Promise<ICategory[]> {
  try {
    const token = getAuthToken();
    const response = await fetch(`${API_BASE_URL}/communities/categories`, {
      headers: {
        'Authorization': `Bearer ${token}`,
      },
    });

    if (!response.ok) {
      throw new Error(`Error fetching community categories: ${response.statusText}`);
    }

    const data = await response.json();
    return data.categories || [];
  } catch (error) {
    console.error('Failed to fetch community categories:', error);
    return [];
  }
}

export const getCategories = getCommunityCategories;

export async function checkUserCommunityMembership(communityId) {
  try {
    const token = getAuthToken();

    const response = await fetch(`${API_BASE_URL}/communities/${communityId}/membership`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': token ? `Bearer ${token}` : ''
      },
      credentials: 'include'
    });

    if (!response.ok) {
      if (response.status === 404) {

        return { success: true, status: 'none' };
      }
      throw new Error(`Failed to check membership (${response.status})`);
    }

    const data = await response.json();
    return {
      success: true,
      status: data.status || 'none'
    };
  } catch (error) {
    logger.error('Check membership failed:', error);
    return {
      success: false,
      status: 'none',
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

      const text = await response.text();
      if (!text) {
        throw new Error(`HTTP error ${response.status}: Empty response`);
      }

      try {

        const errorData = JSON.parse(text);
        throw new Error(errorData.message || `Failed to list communities (${response.status})`);
      } catch (parseError) {

        throw new Error(`Failed to list communities (${response.status}): ${text.substring(0, 100)}`);
      }
    }

    const text = await response.text();
    if (!text) {
      return { communities: [] };
    }

    try {
      return JSON.parse(text);
    } catch (parseError) {
      logger.error('Failed to parse JSON response:', parseError);
      throw new Error('Invalid JSON response from server');
    }
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
      const text = await response.text();
      let errorData;
      try {
        errorData = JSON.parse(text);
      } catch (e) {
        throw new Error(`Failed to get community: HTTP ${response.status} ${response.statusText}`);
      }
      throw new Error(errorData.message || 'Failed to get community');
    }

    const text = await response.text();
    if (!text) {

      logger.warn(`Empty response received from server for community ID: ${id}`);
      return {
        success: true,
        community: {
          id: id,
          name: "Unknown Community",
          description: "Community information is not available",
          logo: "",
          banner: "",
          creatorId: "",
          isApproved: true,
          categories: [],
          createdAt: new Date(),
          memberCount: 0
        }
      };
    }

    try {
      const data = JSON.parse(text);
      return data;
    } catch (parseError) {
      logger.error('JSON parse error:', parseError);

      return {
        success: true,
        community: {
          id: id,
          name: "Unknown Community",
          description: "Community information is not available",
          logo: "",
          banner: "",
          creatorId: "",
          isApproved: true,
          categories: [],
          createdAt: new Date(),
          memberCount: 0
        }
      };
    }
  } catch (error) {

    logger.warn(`Error fetching community ${id}:`, error);

    return {
      success: true,
      community: {
        id: id,
        name: "Unknown Community",
        description: "Error loading community information",
        logo: "",
        banner: "",
        creatorId: "",
        isApproved: true,
        categories: [],
        createdAt: new Date(),
        memberCount: 0
      }
    };
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
      const text = await response.text();
      let errorData;
      try {
        errorData = JSON.parse(text);
      } catch (e) {
        logger.warn(`Failed to parse error response when listing members: ${e}`);

        return { success: true, members: [] };
      }
      logger.warn(`Failed to list members: ${errorData?.message || 'Unknown error'}`);
      return { success: true, members: [] };
    }

    const text = await response.text();
    if (!text) {
      logger.warn(`Empty response received when listing members for community: ${communityId}`);
      return { success: true, members: [] };
    }

    try {
      const data = JSON.parse(text);
      return data;
    } catch (parseError) {
      logger.warn('JSON parse error when listing members:', parseError);
      return { success: true, members: [] };
    }
  } catch (error) {
    logger.warn(`List members for community ${communityId} failed:`, error);
    return { success: true, members: [] };
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
      const text = await response.text();
      let errorData;
      try {
        errorData = JSON.parse(text);
      } catch (e) {
        logger.warn(`Failed to parse error response when listing rules: ${e}`);

        return { success: true, rules: [] };
      }
      logger.warn(`Failed to list rules: ${errorData?.message || 'Unknown error'}`);
      return { success: true, rules: [] };
    }

    const text = await response.text();
    if (!text) {
      logger.warn(`Empty response received when listing rules for community: ${communityId}`);
      return { success: true, rules: [] };
    }

    try {
      const data = JSON.parse(text);
      return data;
    } catch (parseError) {
      logger.warn('JSON parse error when listing rules:', parseError);
      return { success: true, rules: [] };
    }
  } catch (error) {
    logger.warn(`List rules for community ${communityId} failed:`, error);
    return { success: true, rules: [] };
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

      if (response.status === 401) {
        logger.warn('Unauthorized when searching communities - returning empty results');
        return { communities: [], total_count: 0, page, limit };
      }

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

    return {
      communities: [],
      total_count: 0,
      page,
      limit
    };
  }
}