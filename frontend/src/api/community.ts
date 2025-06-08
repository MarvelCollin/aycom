import { getAuthToken } from '../utils/auth';
import appConfig from '../config/appConfig';
import { createLoggerWithPrefix } from '../utils/logger';
import type { ICategory } from '../interfaces/ICategory';
import { uploadFile, SUPABASE_BUCKETS } from '../utils/supabase';

const API_BASE_URL = appConfig.api.baseUrl;
const logger = createLoggerWithPrefix('CommunityAPI');

// Define the params interface for better type checking
interface CommunitiesParams {
  page?: number;
  limit?: number;
  filter?: string;
  q?: string;
  category?: string[];
  is_approved?: boolean;
  [key: string]: any;
}

export async function getCommunities(params: CommunitiesParams = {}) {
  try {
    const token = getAuthToken();
    console.log(`Getting communities with token: ${token ? 'present' : 'missing'}`);

    // Only use pagination and search parameters for the backend query
    // We'll filter by is_approved on the frontend
    const queryParams = new URLSearchParams();
    Object.entries(params).forEach(([key, value]) => {
      // Skip is_approved parameter as we'll filter in the frontend
      if (key === 'is_approved') {
        return;
      } 
      
      if (Array.isArray(value)) {
        value.forEach(v => queryParams.append(key, v));
      } else if (value !== null && value !== undefined) {
        queryParams.append(key, value.toString());
      }
    });

    const response = await fetch(`${API_BASE_URL}/communities?${queryParams.toString()}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': token ? `Bearer ${token}` : ''
      },
      credentials: 'include'
    });

    if (!response.ok) {
      logger.error(`Communities API responded with ${response.status}: ${response.statusText}`);

      const text = await response.text();
      if (!text) {
        throw new Error(`HTTP error ${response.status}: Empty response`);
      }

      try {
        const errorData = JSON.parse(text);
        logger.error('Communities API error details:', errorData);
        throw new Error(errorData.message || `Failed to list communities (${response.status})`);
      } catch (parseError) {
        logger.error('Failed to parse error response:', parseError);
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
      console.log('Raw API response data:', data);
      
      // Handle different response formats
      let success = true;
      let communities = [];
      let pagination: any = {}; // Use any type for now to avoid TypeScript errors
      let limitOptions = [25, 30, 35];
      
      if (data.data && data.data.communities) {
        // Format: { data: { communities: [...], pagination: {...} } }
        communities = data.data.communities;
        pagination = data.data.pagination || {};
        success = data.success !== false;
        limitOptions = data.data.limit_options || limitOptions;
      } else if (data.communities) {
        // Format: { communities: [...], pagination: {...} }
        communities = data.communities;
        pagination = data.pagination || {};
        success = data.success !== false;
        limitOptions = data.limit_options || limitOptions;
      } else {
        console.error('Unexpected API response format:', data);
        success = false;
      }
      
      console.log('Extracted data:', { success, communities, pagination, limitOptions });
      
      // Ensure communities is always an array
      const communitiesArray = Array.isArray(communities) ? communities : [];
      console.log('Communities array to normalize:', communitiesArray);
      
      // Define a type for the community objects
      interface ApiCommunity {
        id?: string;
        name?: string;
        description?: string;
        logo_url?: string;
        logoUrl?: string;
        avatar?: string;
        banner_url?: string;
        bannerUrl?: string;
        creator_id?: string;
        creatorId?: string;
        is_approved?: boolean;
        isApproved?: boolean;
        member_count?: number;
        memberCount?: number;
        created_at?: string;
        createdAt?: string;
        is_private?: boolean;
        isPrivate?: boolean;
        [key: string]: any;
      }
      
      const normalizedCommunities = communitiesArray.map((community: ApiCommunity) => {
        console.log('Normalizing community:', community);
        return {
          id: community.id || '',
          name: community.name || '',
          description: community.description || '',
          logo_url: community.logo_url || community.logoUrl || community.avatar || '',
          banner_url: community.banner_url || community.bannerUrl || '',
          creator_id: community.creator_id || community.creatorId || '',
          is_approved: community.is_approved != null ? community.is_approved : (community.isApproved || false),
          member_count: community.member_count || community.memberCount || 0,
          created_at: community.created_at || community.createdAt || '',
          isPrivate: community.is_private || community.isPrivate || false
        };
      });
      
      // Apply frontend filtering for is_approved if the parameter was provided
      let filteredCommunities = normalizedCommunities;
      console.log('is_approved parameter:', params.is_approved);
      
      if (params.is_approved !== undefined) {
        filteredCommunities = normalizedCommunities.filter(
          community => community.is_approved === params.is_approved
        );
        console.log('Filtered communities after is_approved check:', filteredCommunities);
      }
      
      // Define a type for the pagination object
      interface ApiPagination {
        current_page?: number;
        page?: number;
        currentPage?: number;
        per_page?: number;
        limit?: number;
        perPage?: number;
        total_count?: number;
        totalCount?: number;
        total_pages?: number;
        totalPages?: number;
        [key: string]: any;
      }
      
      // Cast pagination to the correct type
      const paginationData = pagination as ApiPagination;
      
      const paginationResult = {
        total_count: filteredCommunities.length,
        current_page: paginationData.current_page || paginationData.page || paginationData.currentPage || 1,
        per_page: paginationData.per_page || paginationData.limit || paginationData.perPage || 25,
        total_pages: Math.ceil(filteredCommunities.length / (paginationData.per_page || 25))
      };
      
      return {
        success: success,
        communities: filteredCommunities,
        pagination: paginationResult,
        limit_options: limitOptions
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
    
    logger.debug('Creating community with data:', {
      name: data.name,
      description: data.description,
      icon: data.icon ? `[File: ${data.icon.name}]` : null,
      banner: data.banner ? `[File: ${data.banner.name}]` : null,
      categories: data.categories
    });
    
    const payload = {
      name: data.name,
      description: data.description || '',
      logo_url: '',
      banner_url: '',
      rules: data.rules?.toString() || '',
      categories: Array.isArray(data.categories) ? data.categories : []
    };
    
    if (data.icon instanceof File) {
      try {
        const iconUrl = await uploadFile(data.icon, SUPABASE_BUCKETS.FALLBACK, '1kolknj_1');
        if (iconUrl) {
          payload.logo_url = iconUrl;
        }
      } catch (uploadError) {
        logger.error('Failed to upload community icon:', uploadError);
        throw new Error('Failed to upload community icon. Please try again.');
      }
    }
    
    if (data.banner instanceof File) {
      try {
        const bannerUrl = await uploadFile(data.banner, SUPABASE_BUCKETS.FALLBACK, '1kolknj_1');
        if (bannerUrl) {
          payload.banner_url = bannerUrl;
        }
      } catch (uploadError) {
        logger.error('Failed to upload community banner:', uploadError);
        throw new Error('Failed to upload community banner. Please try again.');
      }
    }
    
    const response = await fetch(`${API_BASE_URL}/communities`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': token ? `Bearer ${token}` : ''
      },
      body: JSON.stringify(payload),
      credentials: 'include'
    });
    
    if (!response.ok) {
      const errorText = await response.text();
      let errorMessage = 'Failed to create community';
      try {
        const errorData = JSON.parse(errorText);
        errorMessage = errorData.message || errorMessage;
      } catch(e) {
        logger.error('Error parsing error response:', { error: e, text: errorText });
      }
      throw new Error(errorMessage);
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