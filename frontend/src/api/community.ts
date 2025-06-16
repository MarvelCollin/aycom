import { getAuthToken } from "../utils/auth";
import appConfig from "../config/appConfig";
import { createLoggerWithPrefix } from "../utils/logger";
import type { ICategory } from "../interfaces/ICategory";
import { uploadFile, SUPABASE_BUCKETS, uploadCommunityLogo, uploadCommunityBanner } from "../utils/supabase";

const API_BASE_URL = appConfig.api.baseUrl;
const logger = createLoggerWithPrefix("CommunityAPI");

async function safeApiCall<T>(
  apiFunction: (...args: any[]) => Promise<T>,
  defaultValue: T,
  ...args: any[]
): Promise<T> {
  try {
    const result = await apiFunction(...args);
    return result || defaultValue;
  } catch (error) {
    logger.error(`API call failed: ${error instanceof Error ? error.message : String(error)}`);
    return defaultValue;
  }
}

interface CommunitiesParams {
  page?: number;
  limit?: number;
  filter?: string;
  q?: string;
  category?: string[];
  is_approved?: boolean;
  [key: string]: any;
}

// Helper function to clean empty values from parameters
function cleanParams(params: CommunitiesParams): CommunitiesParams {
  const cleaned: CommunitiesParams = {};
  
  Object.entries(params).forEach(([key, value]) => {
    if (key === 'category' && Array.isArray(value)) {
      const validCategories = value.filter(cat => cat && cat.trim());
      if (validCategories.length > 0) {
        cleaned[key] = validCategories;
      }
    } else if (key === 'q' && value && typeof value === 'string' && value.trim()) {
      cleaned[key] = value.trim();
    } else if (value !== null && value !== undefined && value !== '') {
      cleaned[key] = value;
    }
  });
  
  return cleaned;
}

export async function getUserCommunities(params: CommunitiesParams = {}) {
  try {
    const token = getAuthToken();
    console.log(`Getting user communities with token: ${token ? "present" : "missing"}`);

    // Clean parameters to remove empty values
    const cleanedParams = cleanParams(params);

    const queryParams = new URLSearchParams();
    Object.entries(cleanedParams).forEach(([key, value]) => {
      if (Array.isArray(value)) {
        value.forEach(v => queryParams.append(key, v));
      } else if (value !== null && value !== undefined) {
        queryParams.append(key, value.toString());
      }
    });

    try {
      if (!token) {
        logger.warn("No auth token available for getUserCommunities, returning empty results");
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

      const response = await fetch(`${API_BASE_URL}/communities/user?${queryParams.toString()}`, {
        method: "GET",
        headers: {
          "Content-Type": "application/json",
          "Authorization": `Bearer ${token}`
        },
        credentials: "include"
      });

      if (response.ok) {
        const data = await response.json();
        console.log("User Communities API response:", data);

        return {
          success: true,
          communities: data.communities || [],
          pagination: data.pagination || {
            total_count: 0,
            current_page: 1,
            per_page: 25,
            total_pages: 0
          },
          limit_options: data.limit_options || [25, 30, 35]
        };
      } else {
        logger.warn(`New endpoint failed with ${response.status}, falling back to old endpoint`);

        if (response.status === 401 || response.status === 500) {
          logger.warn(`Server returned ${response.status}, returning empty results`);
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

        if (params.filter === "joined") {
          return await getCommunities({
            ...params,
            filter: "joined"
          });
        } else if (params.filter === "pending") {
          return await getCommunities({
            ...params,
            filter: "pending"
          });
        } else {
          return await getCommunities({
            ...params,
            is_approved: true
          });
        }
      }
    } catch (error) {
      logger.warn("Error using new endpoint, returning empty results:", error);

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
  } catch (error) {
    logger.error("Get user communities failed:", error);
    throw error;
  }
}

export async function getCommunities(params: CommunitiesParams = {}) {
  try {
    const token = getAuthToken();
    console.log(`Getting communities with token: ${token ? "present" : "missing"}`);

    // Clean parameters to remove empty values
    const cleanedParams = cleanParams(params);
    
    const queryParams = new URLSearchParams();
    Object.entries(cleanedParams).forEach(([key, value]) => {
      if (key === "is_approved") {
        return;
      }

      if (Array.isArray(value)) {
        value.forEach(v => queryParams.append(key, v));
      } else if (value !== null && value !== undefined) {
        queryParams.append(key, value.toString());
      }
    });

    const response = await fetch(`${API_BASE_URL}/communities?${queryParams.toString()}`, {
      method: "GET",
      headers: {
        "Content-Type": "application/json",
        "Authorization": token ? `Bearer ${token}` : ""
      },
      credentials: "include"
    });

    if (!response.ok) {
      logger.error(`Communities API responded with ${response.status}: ${response.statusText}`);

      if (response.status === 401) {
        logger.warn("Unauthorized access to communities API - returning empty results");
        return {
          success: true,
          communities: [],
          pagination: {
            total_count: 0,
            current_page: params.page || 1,
            per_page: params.limit || 25,
            total_pages: 0
          },
          limit_options: [25, 30, 35]
        };
      }

      const text = await response.text();
      if (!text) {
        throw new Error(`HTTP error ${response.status}: Empty response`);
      }

      try {
        const errorData = JSON.parse(text);
        logger.error("Communities API error details:", errorData);
        throw new Error(errorData.message || `Failed to list communities (${response.status})`);
      } catch (parseError) {
        logger.error("Failed to parse error response:", parseError);
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
      console.log("Raw API response data:", data);

      let success = true;
      let communities = [];
      let pagination: any = {};
      let limitOptions = [25, 30, 35];

      if (data.data && data.data.communities) {
        communities = data.data.communities;
        pagination = data.data.pagination || {};
        success = data.success !== false;
        limitOptions = data.data.limit_options || limitOptions;
      } else if (data.communities) {
        communities = data.communities;
        pagination = data.pagination || {};
        success = data.success !== false;
        limitOptions = data.limit_options || limitOptions;
      } else {
        console.error("Unexpected API response format:", data);
        success = false;
      }

      console.log("Extracted data:", { success, communities, pagination, limitOptions });

      const communitiesArray = Array.isArray(communities) ? communities : [];
      console.log("Communities array to normalize:", communitiesArray);

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
        console.log("Normalizing community:", community);
        return {
          id: community.id || "",
          name: community.name || "",
          description: community.description || "",
          logo_url: community.logo_url || community.logoUrl || community.avatar || "",
          banner_url: community.banner_url || community.bannerUrl || "",
          creator_id: community.creator_id || community.creatorId || "",
          is_approved: community.is_approved != null ? community.is_approved : (community.isApproved || false),
          member_count: community.member_count || community.memberCount || 0,
          created_at: community.created_at || community.createdAt || "",
          isPrivate: community.is_private || community.isPrivate || false
        };
      });

      let filteredCommunities = normalizedCommunities;
      console.log("is_approved parameter:", params.is_approved);

      if (params.is_approved !== undefined) {
        filteredCommunities = normalizedCommunities.filter(
          community => community.is_approved === params.is_approved
        );
        console.log("Filtered communities after is_approved check:", filteredCommunities);
      }

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
      logger.error("Failed to parse JSON response:", parseError);
      throw new Error("Invalid JSON response from server");
    }
  } catch (error) {
    logger.error("Get communities failed:", error);
    throw error;
  }
}

export async function getCommunityCategories(): Promise<ICategory[]> {
  try {
    const response = await fetch(`${API_BASE_URL}/communities/categories`, {
      method: "GET",
      headers: {
        "Content-Type": "application/json"
      }
    });

    if (!response.ok) {
      throw new Error(`Error fetching community categories: ${response.statusText}`);
    }

    const data = await response.json();
    return data.categories || [];
  } catch (error) {
    console.error("Failed to fetch community categories:", error);
    return [];
  }
}

export const getCategories = getCommunityCategories;

export async function checkUserCommunityMembership(communityId) {
  try {
    const token = getAuthToken();

    if (!token) {
      return {
        success: true,
        status: "none",
        is_member: false,
        user_role: null
      };
    }

    try {
      console.log(`Checking membership for community ${communityId}`);

      const response = await fetch(`${API_BASE_URL}/communities/${communityId}/membership`, {
        method: "GET",
        headers: {
          "Content-Type": "application/json",
          "Authorization": `Bearer ${token}`
        },
        credentials: "include"
      });

      console.log(`Membership API response status: ${response.status}`);

      if (!response.ok) {
        if (response.status === 404) {
          return { success: true, status: "none", is_member: false };
        } else if (response.status === 401 || response.status === 403) {
          return { success: true, status: "none", is_member: false };
        } else if (response.status >= 500) {
          console.warn(`Server error (${response.status}) checking membership for community ${communityId}`);
          return { success: true, status: "none", is_member: false };
        }

        throw new Error(`Failed to check membership (${response.status})`);
      }

      const text = await response.text();

      if (!text || text.trim() === "") {
        return { success: true, status: "none", is_member: false };
      }

      try {
        const data = JSON.parse(text);
        console.log("Parsed membership data:", data);

        if (data.data) {
          const membershipData = data.data;

          if (membershipData.is_member === true || membershipData.status === "member") {
            return {
              success: true,
              status: "member",
              is_member: true,
              user_role: membershipData.role || membershipData.user_role || "member"
            };
          } else if (membershipData.status === "pending") {
            return {
              success: true,
              status: "pending",
              is_member: false
            };
          } else {
            return {
              success: true,
              status: "none",
              is_member: false
            };
          }
        } else {
          if (data.is_member === true || data.status === "member") {
            return {
              success: true,
              status: "member",
              is_member: true,
              user_role: data.role || data.user_role || "member"
            };
          } else if (data.status === "pending") {
            return {
              success: true,
              status: "pending",
              is_member: false
            };
          }
        }

        return {
          success: true,
          status: "none",
          is_member: false
        };
      } catch (parseError) {
        console.error("Error parsing membership response:", parseError);
        return {
          success: true,
          status: "none",
          is_member: false
        };
      }
    } catch (fetchError) {
      console.warn(`Error fetching membership status: ${fetchError instanceof Error ? fetchError.message : String(fetchError)}`);
      return {
        success: true,
        status: "none",
        is_member: false
      };
    }
  } catch (error) {
    logger.error("Check membership failed:", error);
    return {
      success: true,
      status: "none",
      is_member: false,
      error: error instanceof Error ? error.message : "Unknown error"
    };
  }
}

export async function createCommunity(data: Record<string, any>) {
  try {
    const token = getAuthToken();

    if (!token) {
      throw new Error("Authentication required");
    }

    if (data.icon && data.icon instanceof File) {
      const logoUrl = await uploadCommunityLogo(data.icon, "");
      if (logoUrl) {
        data.logo_url = logoUrl;
      }
      delete data.icon;
    }

    if (data.banner && data.banner instanceof File) {
      const bannerUrl = await uploadCommunityBanner(data.banner, "");
      if (bannerUrl) {
        data.banner_url = bannerUrl;
      }
      delete data.banner;
    }

    const response = await fetch(`${API_BASE_URL}/communities`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Authorization": `Bearer ${token}`
      },
      body: JSON.stringify(data),
      credentials: "include"
    });

    if (!response.ok) {
      const errorText = await response.text();
      let errorMessage = "Failed to create community";
      try {
        const errorData = JSON.parse(errorText);
        errorMessage = errorData.message || errorMessage;
      } catch(e) {
        logger.error("Error parsing error response:", { error: e, text: errorText });
      }
      throw new Error(errorMessage);
    }

    return response.json();
  } catch (error) {
    logger.error("Create community failed:", error);
    throw error;
  }
}

export async function listCommunities() {
  try {
    const response = await fetch(`${API_BASE_URL}/communities`, {
      method: "GET",
      headers: {
        "Content-Type": "application/json",
        "Authorization": getAuthToken() ? `Bearer ${getAuthToken()}` : ""
      },
      credentials: "include"
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
      logger.error("Failed to parse JSON response:", parseError);
      throw new Error("Invalid JSON response from server");
    }
  } catch (error) {
    logger.error("List communities failed:", error);
    throw error;
  }
}

export async function getCommunityById(id: string) {
  try {
    const token = getAuthToken();
    console.log(`Fetching community with ID: ${id}`);

    const response = await fetch(`${API_BASE_URL}/communities/${id}`, {
      method: "GET",
      headers: {
        "Content-Type": "application/json",
        "Authorization": token ? `Bearer ${token}` : ""
      },
      credentials: "include"
    });

    console.log(`API response status: ${response.status}`);

    if (!response.ok) {
      if (response.status === 401) {
        console.warn(`Unauthorized access to community ${id} - trying to fetch as public resource`);

        const publicResponse = await fetch(`${API_BASE_URL}/communities/${id}`, {
          method: "GET",
          headers: {
            "Content-Type": "application/json"
          }
        });

        if (!publicResponse.ok) {
          const text = await publicResponse.text();
          console.error(`Error response from public API: ${text}`);
          let errorData;
          try {
            errorData = JSON.parse(text);
            console.error("Parsed error data:", errorData);
          } catch (e) {
            throw new Error(`Failed to get community: HTTP ${publicResponse.status} ${publicResponse.statusText}`);
          }
          throw new Error(errorData.message || "Failed to get community");
        }

        const text = await publicResponse.text();
        try {
          const data = JSON.parse(text);
          console.log("Successfully parsed community data from public endpoint:", data);
          return data;
        } catch (parseError) {
          console.error("JSON parse error:", parseError);
          throw new Error("Failed to parse community data");
        }
      }

      const text = await response.text();
      console.error(`Error response from API: ${text}`);
      let errorData;
      try {
        errorData = JSON.parse(text);
        console.error("Parsed error data:", errorData);
      } catch (e) {
        throw new Error(`Failed to get community: HTTP ${response.status} ${response.statusText}`);
      }
      throw new Error(errorData.message || "Failed to get community");
    }

    const text = await response.text();
    console.log(`API response text length: ${text?.length || 0}`);

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
      console.log("Successfully parsed community data:", data);
      return data;
    } catch (parseError) {
      logger.error("JSON parse error:", parseError);

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
    const token = getAuthToken();

    if (!token) {
      throw new Error("Authentication required");
    }

    if (data.icon && data.icon instanceof File) {
      const logoUrl = await uploadCommunityLogo(data.icon, "");
      if (logoUrl) {
        data.logo_url = logoUrl;
      }
      delete data.icon;
    }

    if (data.banner && data.banner instanceof File) {
      const bannerUrl = await uploadCommunityBanner(data.banner, "");
      if (bannerUrl) {
        data.banner_url = bannerUrl;
      }
      delete data.banner;
    }

    const response = await fetch(`${API_BASE_URL}/communities/${id}`, {
      method: "PUT",
      headers: {
        "Content-Type": "application/json",
        "Authorization": `Bearer ${token}`
      },
      body: JSON.stringify(data),
      credentials: "include"
    });

    if (!response.ok) {
      const errorText = await response.text();
      let errorMessage = "Failed to update community";
      try {
        const errorData = JSON.parse(errorText);
        errorMessage = errorData.message || errorMessage;
      } catch(e) {
        logger.error("Error parsing error response:", { error: e, text: errorText });
      }
      throw new Error(errorMessage);
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
      method: "DELETE",
      headers: {
        "Authorization": getAuthToken() ? `Bearer ${getAuthToken()}` : ""
      },
      credentials: "include"
    });
    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.message || "Failed to delete community");
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
      method: "POST",
      headers: {
        "Authorization": getAuthToken() ? `Bearer ${getAuthToken()}` : ""
      },
      credentials: "include"
    });
    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.message || "Failed to approve community");
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
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Authorization": getAuthToken() ? `Bearer ${getAuthToken()}` : ""
      },
      body: JSON.stringify(data),
      credentials: "include"
    });
    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.message || "Failed to add member");
    }
    return response.json();
  } catch (error) {
    logger.error(`Add member to community ${communityId} failed:`, error);
    throw error;
  }
}

export async function listMembers(communityId: string) {
  try {
    const token = getAuthToken();

    const response = await fetch(`${API_BASE_URL}/communities/${communityId}/members`, {
      method: "GET",
      headers: {
        "Authorization": token ? `Bearer ${token}` : ""
      },
      credentials: "include"
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
      throw new Error(errorData.message || "Failed to list members");
    }

    const data = await response.json();
    console.log("Raw members data:", data);
    const members = data.members || [];

    if (members.length === 0) {
      console.log("No members found in response");
      return data;
    }

    console.log(`Processing ${members.length} community members`);
    console.log("Sample member data from backend:", members[0]);

    const processedMembers = members.map((member) => {
      if (member.username && member.username !== "user_" + member.user_id) {
        console.log(`Using backend user data for member ${member.user_id}: ${member.username}`);
        return {
          ...member,
          name: member.name || member.username || "Unknown User",
          avatar_url: member.avatar_url || member.profile_picture_url || ""
        };
      }

      const shortId = member.user_id ? member.user_id.substring(0, 8) : "unknown";
      return {
        ...member,
        username: member.username || `user_${shortId}`,
        name: member.name || `User ${shortId}`,
        avatar_url: member.avatar_url || member.profile_picture_url || "",
        needs_enrichment: true
      };
    });

    console.log("Processed members:", processedMembers);
    data.members = processedMembers;

    return data;
  } catch (error) {
    logger.error(`List members for community ${communityId} failed:`, error);
    return { success: true, members: [] };
  }
}

export async function updateMemberRole(communityId: string, userId: string, data: Record<string, any>) {
  try {
    const response = await fetch(`${API_BASE_URL}/communities/${communityId}/members/${userId}`, {
      method: "PUT",
      headers: {
        "Content-Type": "application/json",
        "Authorization": getAuthToken() ? `Bearer ${getAuthToken()}` : ""
      },
      body: JSON.stringify(data),
      credentials: "include"
    });
    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.message || "Failed to update member role");
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
      method: "DELETE",
      headers: {
        "Authorization": getAuthToken() ? `Bearer ${getAuthToken()}` : ""
      },
      credentials: "include"
    });
    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.message || "Failed to remove member");
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
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Authorization": getAuthToken() ? `Bearer ${getAuthToken()}` : ""
      },
      body: JSON.stringify(data),
      credentials: "include"
    });
    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.message || "Failed to add rule");
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
      method: "GET",
      headers: {
        "Authorization": getAuthToken() ? `Bearer ${getAuthToken()}` : ""
      },
      credentials: "include"
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
      throw new Error(errorData.message || "Failed to list rules");
    }
    return response.json();
  } catch (error) {
    logger.error(`List rules for community ${communityId} failed:`, error);
    return { success: true, rules: [] };
  }
}

export async function removeRule(communityId: string, ruleId: string) {
  try {
    const response = await fetch(`${API_BASE_URL}/communities/${communityId}/rules/${ruleId}`, {
      method: "DELETE",
      headers: {
        "Authorization": getAuthToken() ? `Bearer ${getAuthToken()}` : ""
      },
      credentials: "include"
    });
    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.message || "Failed to remove rule");
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
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Authorization": getAuthToken() ? `Bearer ${getAuthToken()}` : ""
      },
      body: JSON.stringify(data),
      credentials: "include"
    });
    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.message || "Failed to request to join community");
    }
    return response.json();
  } catch (error) {
    logger.error(`Request to join community ${communityId} failed:`, error);
    throw error;
  }
}

export async function listJoinRequests(communityId: string) {
  try {
    const token = getAuthToken();

    if (!token) {
      throw new Error("Authentication required");
    }

    const response = await fetch(`${API_BASE_URL}/communities/${communityId}/join-requests`, {
      method: "GET",
      headers: {
        "Authorization": `Bearer ${token}`
      },
      credentials: "include"
    });

    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.message || "Failed to list join requests");
    }

    const data = await response.json();
    console.log("Raw join requests data:", data);
    const joinRequests = data.join_requests || [];

    if (joinRequests.length === 0) {
      console.log("No join requests found in response");
      return data;
    }

    console.log(`Processing ${joinRequests.length} join requests`);
    console.log("Sample join request data from backend:", joinRequests[0]);

    const processedRequests = joinRequests.map((request) => {
      if (request.username && request.username !== "user_" + request.user_id) {
        console.log(`Using backend user data for join request ${request.user_id}: ${request.username}`);
        return {
          ...request,
          name: request.name || request.username || "Unknown User",
          avatar_url: request.avatar_url || request.profile_picture_url || ""
        };
      }

      const shortId = request.user_id ? request.user_id.substring(0, 8) : "unknown";
      return {
        ...request,
        username: request.username || `user_${shortId}`,
        name: request.name || `User ${shortId}`,
        avatar_url: request.avatar_url || request.profile_picture_url || "",
        needs_enrichment: true
      };
    });

    console.log("Processed join requests:", processedRequests);
    data.join_requests = processedRequests;

    return data;
  } catch (error) {
    logger.error(`List join requests for community ${communityId} failed:`, error);
    throw error;
  }
}

export async function approveJoinRequest(communityId: string, requestId: string) {
  try {
    const response = await fetch(`${API_BASE_URL}/communities/${communityId}/join-requests/${requestId}/approve`, {
      method: "POST",
      headers: {
        "Authorization": getAuthToken() ? `Bearer ${getAuthToken()}` : ""
      },
      credentials: "include"
    });
    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.message || "Failed to approve join request");
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
      method: "POST",
      headers: {
        "Authorization": getAuthToken() ? `Bearer ${getAuthToken()}` : ""
      },
      credentials: "include"
    });
    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.message || "Failed to reject join request");
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
  limit: number = 10,
  options: any = {}
) {
  try {
    logger.info(`Searching communities with query: ${query}, page: ${page}, limit: ${limit}, options:`, options);

    const cleanQuery = query ? query.trim() : "";

    const params = new URLSearchParams();
    params.append("page", page.toString());
    params.append("limit", limit.toString());

    if (cleanQuery) {
      params.append("q", cleanQuery);
    }

    if (options && typeof options === "object") {
      Object.entries(options).forEach(([key, value]) => {
        if (key === 'categories' && Array.isArray(value)) {
          // Filter out empty categories
          const validCategories = value.filter(cat => cat && cat.trim());
          validCategories.forEach(cat => params.append("category", cat));
        } else if (value !== undefined && value !== null && value !== '') {
          params.append(key, value.toString());
        }
      });
    }

    logger.debug(`Search params: ${params.toString()}`);

    const response = await fetch(`${API_BASE_URL}/communities/search?${params.toString()}`, {
      method: "GET",
      headers: {
        "Content-Type": "application/json",
        "Authorization": getAuthToken() ? `Bearer ${getAuthToken()}` : ""
      },
      credentials: "include"
    });

    if (!response.ok) {
      logger.error(`Communities API error (${response.status})`);
      return getEmptyCommunityResult(page, limit);
    }

    const data = await response.json();
    logger.debug("Communities search API response:", data);

    let communities = [];
    let totalCount = 0;

    if (data.data && data.data.communities) {
      communities = data.data.communities || [];
      totalCount = data.data.pagination?.total_count || 0;
    } else if (data.communities) {
      communities = data.communities || [];
      totalCount = data.total_count || 0;
    }

    logger.debug(`Found ${communities.length} communities${cleanQuery ? " matching search" : ""}`);

    return {
      communities,
      total_count: totalCount,
      pagination: {
        total_count: totalCount,
        current_page: page,
        per_page: limit,
        total_pages: Math.ceil(totalCount / limit) || 1
      }
    };
  } catch (error) {
    logger.error("Search communities failed:", error);
    return getEmptyCommunityResult(page, limit);
  }
}

async function handleCommunityResponse(response: Response, page: number, limit: number) {
  try {
    const text = await response.text();
    if (!text || text.trim() === "") {
      logger.warn("Communities endpoint returned empty response");
      return getEmptyCommunityResult(page, limit);
    }

    const data = JSON.parse(text);

    let communities = [];
    let totalCount = 0;

    if (data.data && data.data.communities) {
      communities = data.data.communities || [];
      totalCount = data.data.pagination?.total_count || 0;
    } else if (data.communities) {
      communities = data.communities || [];
      totalCount = data.total_count || 0;
    }

    return {
      communities: communities,
      total_count: totalCount,
      pagination: {
        total_count: totalCount,
        current_page: page,
        per_page: limit,
        total_pages: Math.ceil(totalCount / limit) || 1
      }
    };
  } catch (parseError) {
    logger.error("Failed to parse communities response:", parseError);
    return getEmptyCommunityResult(page, limit);
  }
}

function getEmptyCommunityResult(page: number, limit: number) {
  return {
    success: true,
    communities: [],
    total: 0,
    page: page,
    limit: limit,
    total_pages: 1
  };
}

export async function getJoinedCommunities(userId: string, params: CommunitiesParams = {}) {
  try {
    console.log(`Getting communities joined by user: ${userId}`);

    const queryParams = new URLSearchParams();
    Object.entries(params).forEach(([key, value]) => {
      if (Array.isArray(value)) {
        value.forEach(v => queryParams.append(key, v));
      } else if (value !== null && value !== undefined) {
        queryParams.append(key, value.toString());
      }
    });

    let attempts = 0;
    const maxAttempts = 2;
    let response;

    while (attempts < maxAttempts) {
      try {
        response = await fetch(`${API_BASE_URL}/communities/user/${userId}/joined?${queryParams.toString()}`, {
          method: "GET",
          headers: {
            "Content-Type": "application/json",
            "Authorization": getAuthToken() ? `Bearer ${getAuthToken()}` : ""
          },
          credentials: "include"
        });

        if (response.status === 401) {
          console.log("Unauthorized access to joined communities, returning empty data set");
          return {
            success: true,
            communities: [],
            total: 0,
            page: params.page || 1,
            limit: params.limit || 25,
            total_pages: 1
          };
        }

        if (response.ok) {
          break;
        }

        logger.warn(`Joined communities request failed (${response.status}), attempt ${attempts + 1}/${maxAttempts}`);
        attempts++;

        if (attempts < maxAttempts) {
          await new Promise(r => setTimeout(r, 1000));
        }
      } catch (err) {
        logger.error("Network error in getJoinedCommunities:", err);
        attempts++;

        if (attempts < maxAttempts) {
          await new Promise(r => setTimeout(r, 1000));
        }
      }
    }

    if (response && response.ok) {
      const result = await response.json();
      console.log("Joined communities raw response:", result);

      // Return the entire response structure directly
      return result;
    }

    logger.warn("All attempts to fetch joined communities failed, returning empty data set");
    return {
      success: true,
      communities: [],
      total: 0,
      page: params.page || 1,
      limit: params.limit || 25,
      total_pages: 1
    };
  } catch (error: any) {
    logger.error("Failed to fetch joined communities:", error);
    return {
      success: true,
      communities: [],
      total: 0,
      page: params.page || 1,
      limit: params.limit || 25,
      total_pages: 1
    };
  }
}

export async function getPendingCommunities(userId: string, params: CommunitiesParams = {}) {
  try {
    console.log(`Getting communities with pending requests by user: ${userId}`);

    const queryParams = new URLSearchParams();
    Object.entries(params).forEach(([key, value]) => {
      if (Array.isArray(value)) {
        value.forEach(v => queryParams.append(key, v));
      } else if (value !== null && value !== undefined) {
        queryParams.append(key, value.toString());
      }
    });

    let attempts = 0;
    const maxAttempts = 3;
    let response;

    while (attempts < maxAttempts) {
      try {
        response = await fetch(`${API_BASE_URL}/communities/user/${userId}/pending?${queryParams.toString()}`, {
          method: "GET",
          headers: {
            "Content-Type": "application/json",
            "Authorization": getAuthToken() ? `Bearer ${getAuthToken()}` : ""
          },
          credentials: "include"
        });

        if (response.status === 401) {
          console.log("Unauthorized access to pending communities, returning empty data set");
          return {
            success: true,
            communities: [],
            total: 0,
            page: params.page || 1,
            limit: params.limit || 25,
            total_pages: 1
          };
        }

        if (response.ok) {
          break;
        }

        logger.warn(`Pending communities request failed (${response.status}), attempt ${attempts + 1}/${maxAttempts}`);
        attempts++;

        if (attempts < maxAttempts) {
          await new Promise(r => setTimeout(r, 1000));
        }
      } catch (err) {
        logger.error("Network error in getPendingCommunities:", err);
        attempts++;

        if (attempts < maxAttempts) {
          await new Promise(r => setTimeout(r, 1000));
        }
      }
    }

    if (response && response.ok) {
      const result = await response.json();
      console.log("Pending communities raw response:", result);

      // Return the entire response structure directly
      return result;
    }

    logger.warn("All attempts to fetch pending communities failed, returning empty data set");
    return {
      success: true,
      communities: [],
      total: 0,
      page: params.page || 1,
      limit: params.limit || 25,
      total_pages: 1
    };
  } catch (error: any) {
    logger.error("Failed to fetch pending communities:", error);
    return {
      success: true,
      communities: [],
      total: 0,
      page: params.page || 1,
      limit: params.limit || 25,
      total_pages: 1
    };
  }
}

export async function getDiscoverCommunities(userId: string, params: CommunitiesParams = {}) {
  try {
    console.log(`Getting discover communities for user: ${userId}`);
    const paramsWithApproval = {
      ...params,
      is_approved: true,
      userId: userId
    };

    const queryParams = new URLSearchParams();
    Object.entries(paramsWithApproval).forEach(([key, value]) => {
      if (Array.isArray(value)) {
        value.forEach(v => queryParams.append(key, v));
      } else if (value !== null && value !== undefined) {
        queryParams.append(key, value.toString());
      }
    });

    let attempts = 0;
    const maxAttempts = 2;
    let response;

    while (attempts < maxAttempts) {
      try {
        response = await fetch(`${API_BASE_URL}/communities/user/${userId}/discover?${queryParams.toString()}`, {
          method: "GET",
          headers: {
            "Content-Type": "application/json",
            "Authorization": getAuthToken() ? `Bearer ${getAuthToken()}` : ""
          },
          credentials: "include"
        });

        if (response.status === 401) {
          console.log("Unauthorized access to discover communities, returning empty data set");
          return {
            success: true,
            communities: [],
            total: 0,
            page: params.page || 1,
            limit: params.limit || 25,
            total_pages: 1
          };
        }

        if (response.ok) {
          break;
        }

        logger.warn(`Discover communities request failed (${response.status}), attempt ${attempts + 1}/${maxAttempts}`);
        attempts++;

        if (attempts < maxAttempts) {
          await new Promise(r => setTimeout(r, 1000));
        }
      } catch (err) {
        logger.error("Network error in getDiscoverCommunities:", err);
        attempts++;

        if (attempts < maxAttempts) {
          await new Promise(r => setTimeout(r, 1000));
        }
      }
    }

    if (response && response.ok) {
      const result = await response.json();
      console.log("Discover communities raw response:", result);

      // Return the entire response structure directly
      return result;
    }

    logger.warn("All attempts to fetch discover communities failed, returning empty data set");
    return {
      success: true,
      communities: [],
      total: 0,
      page: params.page || 1,
      limit: params.limit || 25,
      total_pages: 1
    };
  } catch (error: any) {
    logger.error("Failed to fetch discover communities:", error);
    return {
      success: true,
      communities: [],
      total: 0,
      page: params.page || 1,
      limit: params.limit || 25,
      total_pages: 1
    };
  }
}