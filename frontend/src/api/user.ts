import appConfig from '../config/appConfig';
import { getAuthToken, getUserId as getUserIdFromAuth } from '../utils/auth';
import { uploadProfilePicture as supabaseUploadProfilePicture, uploadBanner as supabaseUploadBanner } from '../utils/supabase';

const API_BASE_URL = appConfig.api.baseUrl;

export interface FollowUserResponse {
  success: boolean;
  message: string;
  was_already_following: boolean;
  is_now_following: boolean;
}

export interface UnfollowUserResponse {
  success: boolean;
  message: string;
  was_following: boolean;
  is_now_following: boolean;
}

export async function getProfile() {
  const token = getAuthToken();
  const authState = localStorage.getItem('auth');
  let userId = null;

  console.log('Getting user profile, token exists:', !!token);

  try {
    if (authState) {
      const parsedAuth = JSON.parse(authState);

      userId = parsedAuth.user_id || parsedAuth.userId;
      console.log('Found user ID in auth state:', userId);
    }
  } catch (err) {
    console.error('Failed to parse auth state:', err);
  }

  if (!userId) {
    console.error('No user ID available, cannot fetch profile');
    throw new Error('User not logged in');
  }

  try {
    return getUserById(userId);
  } catch (error) {
    console.error('Profile fetch exception:', error);
    throw error;
  }
}

export async function updateProfile(data: Record<string, any>) {
  const token = getAuthToken();

  console.log('Updating profile with data:', data);

  const formattedData = {
    name: data.name,
    bio: data.bio,
    date_of_birth: data.date_of_birth,
    profile_picture_url: data.profile_picture_url,
    banner_url: data.banner_url,

    ...Object.keys(data)
      .filter(key => !['name', 'bio', 'date_of_birth', 'profile_picture_url', 'banner_url'].includes(key))
      .reduce((obj, key) => {
        obj[key] = data[key];
        return obj;
      }, {} as Record<string, any>)
  };

  console.log('Formatted profile update data:', formattedData);

  const response = await fetch(`${API_BASE_URL}/users/profile`, {
    method: "PUT",
    headers: { 
      "Content-Type": "application/json",
      "Authorization": token ? `Bearer ${token}` : ''
    },
    body: JSON.stringify(formattedData),
    credentials: "include",
  });
  if (!response.ok) {
    try {
      const errorData = await response.json();
      throw new Error(errorData.message || "Failed to update user profile");
    } catch (parseError) {
      throw new Error("Failed to update user profile");
    }
  }
  return response.json();
}

export async function checkUsernameAvailability(username: string): Promise<boolean> {
  try {
    const response = await fetch(`${API_BASE_URL}/users/check-username?username=${encodeURIComponent(username)}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json'
      }
    });

    if (!response.ok) {
      throw new Error(`Failed to check username: ${response.status}`);
    }

    const data = await response.json();
    return data.available;
  } catch (err) {
    console.error('Failed to check username availability:', err);
    return false;
  }
}

export async function followUser(userId: string): Promise<FollowUserResponse> {
  try {
    console.log(`Attempting to follow user: ${userId}`);
    const token = getAuthToken();

    if (!token) {
      console.error('Cannot follow user: No authentication token available');
      return { success: false, message: 'No authentication token available', was_already_following: false, is_now_following: false };
    }

    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), 10000); 

    const response = await fetch(`${API_BASE_URL}/users/${userId}/follow`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      signal: controller.signal
    });

    clearTimeout(timeoutId);

    const responseText = await response.text();
    console.log(`Follow API raw response: ${responseText}`);

    let responseData: any;
    try {
      responseData = JSON.parse(responseText);
    } catch (parseError) {
      console.error(`Failed to parse follow response as JSON: ${responseText}`);
      return { success: false, message: 'Invalid response format', was_already_following: false, is_now_following: false };
    }

    console.log('Follow API parsed response:', responseData);

    if (!response.ok) {
      console.error(`Failed to follow user: ${response.status}`, responseData);
      return { 
        success: false, 
        message: responseData.message || `Request failed with status ${response.status}`,
        was_already_following: responseData.was_already_following || responseData.wasAlreadyFollowing || false,
        is_now_following: responseData.is_now_following || responseData.isNowFollowing || false
      };
    }

    const standardizedResponse: FollowUserResponse = {
      success: responseData.success === true,
      message: responseData.message || 'Follow operation processed',
      was_already_following: responseData.was_already_following || responseData.wasAlreadyFollowing || false,
      is_now_following: responseData.is_now_following || responseData.isNowFollowing || true
    };

    console.log(`Successfully processed follow request for user ${userId}:`, standardizedResponse);
    return standardizedResponse;
  } catch (err) {
    if (err instanceof DOMException && err.name === 'AbortError') {
      console.error("Follow user request timed out after 10 seconds");
      return { success: false, message: 'Request timed out after 10 seconds', was_already_following: false, is_now_following: false };
    } else {
      console.error('Failed to follow user:', err);
      return { success: false, message: err instanceof Error ? err.message : 'Unknown error', was_already_following: false, is_now_following: false };
    }
  }
}

export async function unfollowUser(userId: string): Promise<UnfollowUserResponse> {
  try {
    console.log(`Attempting to unfollow user: ${userId}`);
    const token = getAuthToken();

    if (!token) {
      console.error('Cannot unfollow user: No authentication token available');
      return { success: false, message: 'No authentication token available', was_following: false, is_now_following: false };
    }

    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), 10000); 

    const response = await fetch(`${API_BASE_URL}/users/${userId}/unfollow`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      signal: controller.signal
    });

    clearTimeout(timeoutId);

    const responseText = await response.text();
    console.log(`Unfollow API raw response: ${responseText}`);

    let responseData: any;
    try {
      responseData = JSON.parse(responseText);
    } catch (parseError) {
      console.error(`Failed to parse unfollow response as JSON: ${responseText}`);
      return { success: false, message: 'Invalid response format', was_following: false, is_now_following: false };
    }

    console.log('Unfollow API parsed response:', responseData);

    if (!response.ok) {
      console.error(`Failed to unfollow user: ${response.status}`, responseData);
      return { 
        success: false, 
        message: responseData.message || `Request failed with status ${response.status}`,
        was_following: responseData.was_following || responseData.wasFollowing || false,
        is_now_following: responseData.is_now_following || responseData.isNowFollowing || false
      };
    }

    const standardizedResponse: UnfollowUserResponse = {
      success: responseData.success === true,
      message: responseData.message || 'Unfollow operation processed',
      was_following: responseData.was_following || responseData.wasFollowing || true,
      is_now_following: responseData.is_now_following || responseData.isNowFollowing || false
    };

    console.log(`Successfully processed unfollow request for user ${userId}:`, standardizedResponse);
    return standardizedResponse;
  } catch (err) {
    if (err instanceof DOMException && err.name === 'AbortError') {
      console.error("Unfollow user request timed out after 10 seconds");
      return { success: false, message: 'Request timed out after 10 seconds', was_following: false, is_now_following: false };
    } else {
      console.error('Failed to unfollow user:', err);
      return { success: false, message: err instanceof Error ? err.message : 'Unknown error', was_following: false, is_now_following: false };
    }
  }
}

export async function getFollowers(userId: string, page = 1, limit = 20): Promise<any> {
  try {
    const token = getAuthToken();

    const response = await fetch(`${API_BASE_URL}/users/${userId}/followers?page=${page}&limit=${limit}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': token ? `Bearer ${token}` : ''
      }
    });

    if (!response.ok) {
      throw new Error(`Failed to get followers: ${response.status}`);
    }

    const rawData = await response.json();
    console.log('Followers API response:', rawData);

    let data = rawData;
    if (rawData.data) {
      data = rawData.data;
    }

    if (data && data.followers) {
      const followersList = data.followers.map((follower: any) => ({
        id: follower.id,
        name: follower.name || follower.display_name,
        username: follower.username,
        profile_picture_url: follower.profile_picture_url || '👤',
        is_verified: follower.is_verified || follower.verified || false,
        is_following: follower.is_following || false,
        bio: follower.bio || ''
      }));

      if (rawData.data) {
        return {
          data: {
            followers: followersList,
            pagination: data.pagination || {
              total_count: followersList.length,
              current_page: page,
              per_page: limit
            }
          },
          success: rawData.success
        };
      }

      return {
        followers: followersList,
        pagination: data.pagination || {
          total_count: followersList.length,
          current_page: page,
          per_page: limit
        }
      };
    }

    if (rawData.data) {
      return {
        data: { 
          followers: [], 
          pagination: { total_count: 0, current_page: page, per_page: limit }
        },
        success: rawData.success
      };
    }

    return { followers: [], pagination: { total_count: 0, current_page: page, per_page: limit } };
  } catch (err) {
    console.error('Failed to get followers:', err);
    throw err;
  }
}

export async function getFollowing(userId: string, page = 1, limit = 20): Promise<any> {
  try {
    const token = getAuthToken();

    const response = await fetch(`${API_BASE_URL}/users/${userId}/following?page=${page}&limit=${limit}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': token ? `Bearer ${token}` : ''
      }
    });

    if (!response.ok) {
      throw new Error(`Failed to get following: ${response.status}`);
    }

    const rawData = await response.json();
    console.log('Following API response:', rawData);

    let data = rawData;
    if (rawData.data) {

      data = rawData.data;
    }

    if (data && data.following) {
      const followingList = data.following.map((following: any) => ({
        id: following.id,
        name: following.name || following.display_name,
        username: following.username,
        profile_picture_url: following.profile_picture_url || '👤',
        is_verified: following.is_verified || following.verified || false,
        is_following: true,
        bio: following.bio || ''
      }));

      if (rawData.data) {
        return {
          data: {
            following: followingList,
            pagination: data.pagination || {
              total_count: followingList.length,
              current_page: page,
              per_page: limit
            }
          },
          success: rawData.success
        };
      }

      return {
        following: followingList,
        pagination: data.pagination || {
          total_count: followingList.length,
          current_page: page,
          per_page: limit
        }
      };
    }

    if (rawData.data) {
      return {
        data: { 
          following: [], 
          pagination: { total_count: 0, current_page: page, per_page: limit }
        },
        success: rawData.success
      };
    }

    return { following: [], pagination: { total_count: 0, current_page: page, per_page: limit } };
  } catch (err) {
    console.error('Failed to get following:', err);
    throw err;
  }
}

function getUserId(): string {
  const userId = getUserIdFromAuth();
  return userId || '';
}

export async function getUserById(userId: string): Promise<any> {
  try {
    console.log('Fetching user by ID:', userId);
    const token = getAuthToken();

    const response = await fetch(`${API_BASE_URL}/users/${userId}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': token ? `Bearer ${token}` : ''
      }
    });

    console.log('Get user by ID response status:', response.status);

    if (!response.ok) {
      throw new Error(`Failed to get user: ${response.status}`);
    }

    const data = await response.json();
    console.log('User data received:', data);

    let is_admin = false;
    try {
      const authState = localStorage.getItem('auth');
      if (authState) {
        const auth = JSON.parse(authState);
        if (auth.is_admin === true) {
          is_admin = true;
          console.log('User is admin according to auth state');
        }
      }
    } catch (e) {
      console.error('Error getting admin status from auth state:', e);
    }

    if (data && data.data && data.data.user) {
      const userData = data.data.user;

      if (userData.is_admin === true) {
        is_admin = true;
        console.log('User is admin according to API response');

        try {
          const authState = localStorage.getItem('auth');
          if (authState) {
            const auth = JSON.parse(authState);
            auth.is_admin = true;
            localStorage.setItem('auth', JSON.stringify(auth));
            console.log('Updated auth state with admin status from API');
          }
        } catch (e) {
          console.error('Error updating auth state with admin status:', e);
        }
      }

      const isFollowing = userData.is_following === true;
      console.log(`User is_following status from API: ${isFollowing}`);

      return {
        success: true,
        user: {
          id: userData.id,
          username: userData.username,
          name: userData.display_name || userData.name,
          display_name: userData.display_name,
          profile_picture_url: userData.profile_picture_url,
          banner_url: userData.banner_url,
          bio: userData.bio,
          is_verified: userData.is_verified,
          is_admin: userData.is_admin === true || is_admin, 
          follower_count: userData.follower_count || 0,
          following_count: userData.following_count || 0,
          is_following: isFollowing
        }
      };
    }

    return data;
  } catch (err) {
    console.error('Failed to get user by ID:', err);
    throw err;
  }
}

export async function getUserByUsername(username: string): Promise<any> {
  try {
    const token = getAuthToken();

    console.log(`Fetching user with username: ${username}`);
    const response = await fetch(`${API_BASE_URL}/users/username/${username}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': token ? `Bearer ${token}` : ''
      }
    });
      if (!response.ok) {
      console.error(`Failed to fetch user by username: ${response.status}`);
      throw new Error(`Failed to fetch user by username: ${response.status} ${response.statusText}`);
    }

    const data = await response.json();

    if (data && data.user) {
      return {
        success: true,
        user: data.user
      };
    } else if (data && data.data && data.data.user) {
      return {
        success: true,
        user: data.data.user
      };
    } else if (data) {

      return {
        success: true,
        user: data
      };
    }

    throw new Error('Unrecognized API response format');
  } catch (err) {
    console.error('Failed to fetch user by username:', err);
    throw err;
  }
}

export async function checkFollowStatus(userId: string): Promise<boolean> {
  try {
    const token = getAuthToken();

    if (!token || !userId) {
      return false;
    }

    if (!/^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i.test(userId)) {

      const userData = await getUserByUsername(userId);
      if (userData?.data?.user?.id) {

        if (userData.data.user.is_following !== undefined) {
          return userData.data.user.is_following === true;
        }
        userId = userData.data.user.id;
      } else {
        return false;
      }
    }

    const response = await fetch(`${API_BASE_URL}/users/${userId}/follow-status`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      }
    });

    if (!response.ok) {
      return false;
    }

    const data = await response.json();

    if (data?.is_following !== undefined) {
      return data.is_following === true;
    } else if (data?.data?.is_following !== undefined) {
      return data.data.is_following === true;
    }

    return false;
  } catch (err) {
    console.error('Error checking follow status:', err);
    return false;
  }
}

export const getUserFollowers = getFollowers;
export const getUserFollowing = getFollowing; 

export async function checkAdminStatus(): Promise<boolean> {
  try {
    const token = getAuthToken();
    if (!token) return false;

    const userId = getUserId();
    if (!userId) return false;

    const response = await fetch(`${API_BASE_URL}/auth/check-admin`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      }
    });

    if (!response.ok) {
      return false;
    }

    const data = await response.json();

    let is_admin = false;

    if (data && typeof data.is_admin === 'boolean') {
      is_admin = data.is_admin;
    } else if (data && data.data && typeof data.data.is_admin === 'boolean') {
      is_admin = data.data.is_admin;
    }

    if (is_admin) {
      try {
        const authData = localStorage.getItem('auth');
        if (authData) {
          const auth = JSON.parse(authData);
          auth.is_admin = true;
          localStorage.setItem('auth', JSON.stringify(auth));
        }
      } catch (e) {

      }
    }

    return is_admin;
  } catch (error) {
    console.error('Admin status check failed:', error);
    return false;
  }
}

export async function getAllUsers(
  page: number = 1,
  limit: number = 10, 
  sortBy: string = 'created_at',
  ascending: boolean = false,
  searchQuery?: string
): Promise<any> {
  try {
    const token = getAuthToken();

    const params = new URLSearchParams({
      page: page.toString(),
      limit: limit.toString(),
      sort_by: sortBy,
      ascending: ascending.toString()
    });

    if (searchQuery) {
      params.append('search', searchQuery);
    }

    const response = await fetch(`${API_BASE_URL}/users?${params.toString()}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      }
    });

    if (!response.ok) {
      throw new Error(`Failed to get users (${response.status})`);
    }

    return await response.json();
  } catch (error) {
    console.error('Get all users failed:', error);
    return {
      success: false,
      users: [],
      total: 0,
      error: error instanceof Error ? error.message : 'Unknown error'
    };
  }
}

export async function searchUsers(
  query: string, 
  page: number = 1, 
  limit: number = 10, 
  options?: any
): Promise<any> {
  try {
    const token = getAuthToken();

    // Validate and truncate search query if too long
    const MAX_QUERY_LENGTH = 30;
    let validatedQuery = query;
    if (query.length > MAX_QUERY_LENGTH) {
      console.warn(`Search query too long (${query.length} chars). Truncating to ${MAX_QUERY_LENGTH} characters.`);
      validatedQuery = query.substring(0, MAX_QUERY_LENGTH);
    }

    // Function to make a search request with or without auth
    const makeSearchRequest = async (withAuth: boolean) => {
      const params = new URLSearchParams({
        query: validatedQuery,
        page: page.toString(),
        limit: limit.toString()
      });
  
      if (options) {
        Object.keys(options).forEach(key => {
          if (options[key] !== undefined) {
            params.append(key, options[key].toString());
          }
        });
      }
  
      const headers: Record<string, string> = {
        'Content-Type': 'application/json'
      };
  
      if (withAuth && token) {
        headers['Authorization'] = `Bearer ${token}`;
      }
  
      return await fetch(`${API_BASE_URL}/users/search?${params.toString()}`, {
        method: 'GET',
        headers,
        credentials: 'include'
      });
    };

    // First try with authentication if we have a token
    let response = await makeSearchRequest(!!token);
    
    // If we get a 401 error and tried with auth, retry without auth
    if (response.status === 401 && token) {
      console.log('Search API returned 401, retrying without authentication');
      response = await makeSearchRequest(false);
    }

    if (!response.ok) {
      // For certain error codes, return empty results instead of throwing
      if (response.status === 401 || response.status === 403 || response.status === 400) {
        console.warn(`Search users returned ${response.status}, returning empty results`);
        return {
          success: true,
          users: [],
          total: 0
        };
      }
      throw new Error(`Search users failed (${response.status})`);
    }

    const data = await response.json();
    return {
      success: true,
      users: data.users || [],
      total: data.total || (data.users ? data.users.length : 0)
    };
  } catch (error) {
    console.error('Search users failed:', error);
    // Always return a valid response object even when an error occurs
    return {
      success: false,
      users: [],
      total: 0,
      error: error instanceof Error ? error.message : 'Unknown error'
    };
  }
}

export async function uploadProfilePicture(file: File): Promise<string> {
  try {
    console.log('Uploading profile picture:', file.name);

    if (!file.type.match(/^image\/(jpeg|png|gif|jpg|webp)$/)) {
      throw new Error('Invalid file type. Please upload an image file.');
    }

    if (file.size > 5 * 1024 * 1024) { 
      throw new Error('File size exceeds the limit of 5MB.');
    }

    const userId = getUserId();
    if (!userId) {
      throw new Error('Cannot upload profile picture: User is not authenticated');
    }

    const url = await supabaseUploadProfilePicture(file, userId);

    if (!url) {
      throw new Error('Failed to get URL from upload service');
    }

    console.log('Profile picture uploaded successfully:', url);
    return url;
  } catch (error) {
    console.error('Failed to upload profile picture:', error);
    throw error;
  }
}

export async function uploadBanner(file: File): Promise<string> {
  try {
    console.log('Uploading banner:', file.name);

    if (!file.type.match(/^image\/(jpeg|png|gif|jpg|webp)$/)) {
      throw new Error('Invalid file type. Please upload an image file.');
    }

    if (file.size > 5 * 1024 * 1024) { 
      throw new Error('File size exceeds the limit of 5MB.');
    }

    const userId = getUserId();
    if (!userId) {
      throw new Error('Cannot upload banner: User is not authenticated');
    }

    const url = await supabaseUploadBanner(file, userId);

    if (!url) {
      throw new Error('Failed to get URL from upload service');
    }

    console.log('Banner uploaded successfully:', url);
    return url;
  } catch (error) {
    console.error('Failed to upload banner:', error);
    throw error;
  }
}

export async function updateUserAdminStatus(
  userId: string, 
  is_admin: boolean,
  isDebugRequest: boolean = false
): Promise<{ success: boolean, message: string }> {
  try {
    const token = getAuthToken();

    if (!token) {
      console.error('Cannot update admin status: No authentication token available');
      return { success: false, message: 'Authentication required' };
    }

    const debugParam = isDebugRequest ? '?debug=true' : '';

    const response = await fetch(`${API_BASE_URL}/users/${userId}/admin-status${debugParam}`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      body: JSON.stringify({
        is_admin
      })
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.message || `Failed to update admin status: ${response.status}`);
    }

    const result = await response.json();
    return { 
      success: result.success || true, 
      message: result.message || `User admin status updated to ${is_admin ? 'admin' : 'regular user'}`
    };
  } catch (error) {
    console.error('Failed to update admin status:', error);
    return { 
      success: false, 
      message: error instanceof Error ? error.message : 'Unknown error' 
    };
  }
}

export async function reportUser(userId: string, reason: string): Promise<boolean> {
  try {
    const token = getAuthToken();

    const response = await fetch(`${API_BASE_URL}/users/${userId}/report`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': token ? `Bearer ${token}` : ''
      },
      body: JSON.stringify({ reason })
    });

    return response.ok;
  } catch (err) {
    console.error('Failed to report user:', err);
    return false;
  }
}

export async function blockUser(userId: string): Promise<boolean> {
  try {
    const token = getAuthToken();

    const response = await fetch(`${API_BASE_URL}/users/${userId}/block`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': token ? `Bearer ${token}` : ''
      }
    });

    return response.ok;
  } catch (err) {
    console.error('Failed to block user:', err);
    return false;
  }
}

export async function unblockUser(userId: string): Promise<boolean> {
  try {
    const token = getAuthToken();

    const response = await fetch(`${API_BASE_URL}/users/${userId}/unblock`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': token ? `Bearer ${token}` : ''
      }
    });

    return response.ok;
  } catch (err) {
    console.error('Failed to unblock user:', err);
    return false;
  }
}

export async function getBlockedUsers(page = 1, limit = 20): Promise<any[]> {
  try {
    const token = getAuthToken();

    const response = await fetch(`${API_BASE_URL}/users/blocked?page=${page}&limit=${limit}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': token ? `Bearer ${token}` : ''
      }
    });

    if (!response.ok) {
      throw new Error(`Failed to get blocked users: ${response.status}`);
    }

    const data = await response.json();

    if (data && data.blocked_users) {
      return data.blocked_users;
    } else if (data && data.data && data.data.blocked_users) {
      return data.data.blocked_users;
    }

    return [];
  } catch (err) {
    console.error('Failed to get blocked users:', err);
    return [];
  }
}

export async function pinThread(threadId: string): Promise<any> {
  try {
    const token = getAuthToken();

    const response = await fetch(`${API_BASE_URL}/threads/${threadId}/pin`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': token ? `Bearer ${token}` : ''
      }
    });

    if (!response.ok) {
      throw new Error(`Failed to pin thread: ${response.status}`);
    }

    return response.json();
  } catch (err) {
    console.error('Failed to pin thread:', err);
    throw err;
  }
}

export async function unpinThread(threadId: string): Promise<any> {
  try {
    const token = getAuthToken();

    const response = await fetch(`${API_BASE_URL}/threads/${threadId}/unpin`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': token ? `Bearer ${token}` : ''
      }
    });

    if (!response.ok) {
      throw new Error(`Failed to unpin thread: ${response.status}`);
    }

    return response.json();
  } catch (err) {
    console.error('Failed to unpin thread:', err);
    throw err;
  }
}

export async function pinReply(replyId: string): Promise<any> {
  try {
    const token = getAuthToken();

    const response = await fetch(`${API_BASE_URL}/replies/${replyId}/pin`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': token ? `Bearer ${token}` : ''
      }
    });

    if (!response.ok) {
      throw new Error(`Failed to pin reply: ${response.status}`);
    }

    return response.json();
  } catch (err) {
    console.error('Failed to pin reply:', err);
    throw err;
  }
}

export async function unpinReply(replyId: string): Promise<any> {
  try {
    const token = getAuthToken();

    const response = await fetch(`${API_BASE_URL}/replies/${replyId}/unpin`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': token ? `Bearer ${token}` : ''
      }
    });

    if (!response.ok) {
      throw new Error(`Failed to unpin reply: ${response.status}`);
    }

    return response.json();
  } catch (err) {
    console.error('Failed to unpin reply:', err);
    throw err;
  }
}

export async function submitPremiumRequest(reason: string, identityCardNumber: string, facePhotoURL: string): Promise<boolean> {
  try {
    const token = getAuthToken();
    if (!token) return false;

    const response = await fetch(`${API_BASE_URL}/users/premium-request`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      body: JSON.stringify({
        reason,
        identity_card_number: identityCardNumber,
        face_photo_url: facePhotoURL
      })
    });

    if (!response.ok) {
      const errorData = await response.json();
      console.error('Error submitting premium request:', errorData);
      return false;
    }

    const data = await response.json();
    return data.success;
  } catch (error) {
    console.error('Error submitting premium request:', error);
    return false;
  }
}