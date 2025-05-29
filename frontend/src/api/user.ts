import appConfig from '../config/appConfig';
import { getAuthToken } from '../utils/auth';
import { uploadProfilePicture as supabaseUploadProfilePicture, uploadBanner as supabaseUploadBanner } from '../utils/supabase';

const API_BASE_URL = appConfig.api.baseUrl;

// Enhanced response interfaces matching backend protobuf messages
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
  
  // Try to get userId from local storage
  try {
    if (authState) {
      const parsedAuth = JSON.parse(authState);
      userId = parsedAuth.userId;
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
    // Use getUserById instead of profile endpoint
    return getUserById(userId);
  } catch (error) {
    console.error('Profile fetch exception:', error);
    throw error;
  }
}

export async function updateProfile(data: Record<string, any>) {
  const token = getAuthToken();
  
  console.log('Updating profile with data:', data);
  
  // Ensure we have consistent field names with what the backend expects
  const formattedData = {
    ...data,
    // Convert displayName to name if present
    name: data.displayName || data.name,
    // Convert dateOfBirth to date_of_birth if present
    date_of_birth: data.dateOfBirth || data.date_of_birth,
    // Handle profile picture fields
    profile_picture_url: data.profilePicture || data.profile_picture_url || data.profile_picture || data.avatar,
    // Handle banner fields
    banner_url: data.backgroundBanner || data.banner_url || data.banner,
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
    
    // Create controller for timeout management
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), 10000); // 10 second timeout
    
    const response = await fetch(`${API_BASE_URL}/users/${userId}/follow`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      signal: controller.signal
    });
    
    clearTimeout(timeoutId);
    
    // Get response text first so we can log it even if JSON parsing fails
    const responseText = await response.text();
    console.log(`Follow API raw response: ${responseText}`);
    
    let responseData: FollowUserResponse;
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
        was_already_following: responseData.was_already_following || false,
        is_now_following: responseData.is_now_following || false
      };
    }
    
    // Return the enhanced response data
    if (responseData && responseData.success === true) {
      console.log(`Successfully processed follow request for user ${userId}`);
      return responseData;
    } else {
      console.error(`Follow operation failed - API returned success=false:`, responseData);
      return { 
        success: false, 
        message: responseData.message || 'Follow operation failed',
        was_already_following: responseData.was_already_following || false,
        is_now_following: responseData.is_now_following || false
      };
    }
  } catch (err) {
    if (err instanceof DOMException && err.name === 'AbortError') {
      console.error("Follow user request timed out after 10 seconds");
      return { success: false, message: 'Request timed out after 10 seconds', was_already_following: false, is_now_following: false };
    } else {
      console.error('Failed to follow user:', err);
      return { success: false, message: err instanceof Error ? err.message : 'Unknown error', was_already_following: false, is_now_following: false };
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
    
    // Create controller for timeout management
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), 10000); // 10 second timeout
    
    const response = await fetch(`${API_BASE_URL}/users/${userId}/unfollow`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      signal: controller.signal
    });
    
    clearTimeout(timeoutId);
    
    // Get response text first so we can log it even if JSON parsing fails
    const responseText = await response.text();
    console.log(`Unfollow API raw response: ${responseText}`);
    
    let responseData: UnfollowUserResponse;
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
        was_following: responseData.was_following || false,
        is_now_following: responseData.is_now_following || false
      };
    }
    
    // Return the enhanced response data
    if (responseData && responseData.success === true) {
      console.log(`Successfully processed unfollow request for user ${userId}`);
      return responseData;
    } else {
      console.error(`Unfollow operation failed - API returned success=false:`, responseData);
      return { 
        success: false, 
        message: responseData.message || 'Unfollow operation failed',
        was_following: responseData.was_following || false,
        is_now_following: responseData.is_now_following || false
      };
    }
  } catch (err) {
    if (err instanceof DOMException && err.name === 'AbortError') {
      console.error("Unfollow user request timed out after 10 seconds");
      return { success: false, message: 'Request timed out after 10 seconds', was_following: false, is_now_following: false };
    } else {
      console.error('Failed to unfollow user:', err);
      return { success: false, message: err instanceof Error ? err.message : 'Unknown error', was_following: false, is_now_following: false };
    }
  }
}     console.error('Failed to unfollow user:', err);
    }
    return false;
  }
}

export async function getFollowers(userId: string, page = 1, limit = 20): Promise<any[]> {
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
    
    const data = await response.json();
    
    if (data && data.followers) {
      return data.followers.map((follower: any) => ({
        id: follower.id,
        name: follower.name || follower.display_name,
        username: follower.username,
        profile_picture: follower.profile_picture_url || '👤',
        verified: follower.verified || false,
        isFollowing: follower.is_following || false
      }));
    }
    
    return [];
  } catch (err) {
    console.error('Failed to get followers:', err);
    return [];
  }
}

export async function getFollowing(userId: string, page = 1, limit = 20): Promise<any[]> {
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
    
    const data = await response.json();
    
    if (data && data.following) {
      return data.following.map((following: any) => ({
        id: following.id,
        name: following.name || following.display_name,
        username: following.username,
        profile_picture: following.profile_picture_url || '👤',
        verified: following.verified || false,
        isFollowing: true
      }));
    }
    
    return [];
  } catch (err) {
    console.error('Failed to get following list:', err);
    return [];
  }
}

export async function searchUsers(query: string, page: number = 1, limit: number = 10, options?: any) {
  try {
    // Always use the search endpoint, but handle empty queries
    const url = new URL(`${API_BASE_URL}/users/search`);
    
    // Add pagination parameters
    url.searchParams.append('page', page.toString());
    url.searchParams.append('limit', limit.toString());
    
    // Always add query param - use a space or the original query
    url.searchParams.append('q', query.trim() || ' ');
    
    if (options) {
      if (options.verified !== undefined) {
        url.searchParams.append('verified', options.verified ? 'true' : 'false');
      }
      
      if (options.active !== undefined) {
        url.searchParams.append('active', options.active ? 'true' : 'false');
      }
      
      if (options.sort) {
        url.searchParams.append('sort', options.sort);
      }
      
      if (options.filter === 'following') {
        url.searchParams.append('following', 'true');
      }
      
      if (options.filter === 'verified') {
        url.searchParams.append('verified', 'true');
      }
      
      // Add fuzzy search flag if needed
      if (options.fuzzy) {
        url.searchParams.append('fuzzy', 'true');
      }
    }
    
    const token = getAuthToken();
    
    const response = await fetch(url.toString(), {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': token ? `Bearer ${token}` : ''
      }
    });
    
    if (!response.ok) {
      const errorText = await response.text();
      console.error('User API error:', errorText);
      throw new Error(`Failed to fetch users: ${response.status}`);
    }
      const responseData = await response.json();
    console.log('Raw API response:', responseData);
    
    // Handle the nested response structure: responseData.data.users
    const data = responseData.data || responseData;
    const users = data.users || [];
    const pagination = data.pagination || {};
    
    console.log('Extracted users:', users);
    console.log('Pagination info:', pagination);
    
    // Apply client-side fuzzy search if requested and API doesn't support it
    let processedUsers = [...users];
    if (options?.clientFuzzy && query.trim()) {
      // Import the fuzzy search function
      const { fuzzySearch } = await import('../utils/helpers');
      
      // If the API doesn't indicate it did fuzzy searching, do it client-side
      if (!responseData.fuzzy_search_applied) {
        // Apply fuzzy search on both username and display_name
        const fuzzyMatches = fuzzySearch(
          query,
          users,
          'username',
          0.5 // Lower threshold for username matches
        );
        
        const nameMatches = fuzzySearch(
          query,
          users.filter(user => !fuzzyMatches.includes(user)), // Remove already matched users
          'display_name',
          0.6
        );
        
        // Combine and keep original order if possible
        processedUsers = [...new Set([...fuzzyMatches, ...nameMatches])];
        console.log(`Applied client-side fuzzy search, found ${processedUsers.length} matches`);
      }
    }
    
    // Standardize the response format
    return {
      users: processedUsers,
      totalCount: pagination.total || users.length,
      page: pagination.page || page,
      totalPages: Math.ceil((pagination.total || users.length) / (pagination.limit || limit))
    };
  } catch (err) {
    console.error('Failed to fetch users:', err);
    return { users: [], totalCount: 0, page, totalPages: 0 };
  }
}

export async function uploadProfilePicture(file: File) {
  try {
    // First try to upload directly to Supabase
    const url = await supabaseUploadProfilePicture(file, getUserId());
    
    if (url) {
      // If successful, update the user's profile with the new URL
      const token = getAuthToken();
      
      const response = await fetch(`${API_BASE_URL}/users/profile-picture/update`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': token ? `Bearer ${token}` : ''
        },
        body: JSON.stringify({ profilePictureUrl: url })
      });
      
      if (!response.ok) {
        try {
          const errorData = await response.json();
          throw new Error(errorData.message || "Failed to update profile picture in the database");
        } catch (parseError) {
          throw new Error(`Failed to update profile picture in the database: ${response.status}`);
        }
      }
      
      return { success: true, url };
    }
    
    // Fall back to the API if Supabase upload fails
    const token = getAuthToken();
    
    const formData = new FormData();
    formData.append('file', file);
    
    const response = await fetch(`${API_BASE_URL}/users/profile-picture`, {
      method: 'POST',
      headers: {
        'Authorization': token ? `Bearer ${token}` : ''
      },
      body: formData
    });
    
    if (!response.ok) {
      try {
        const errorData = await response.json();
        throw new Error(errorData.message || "Failed to upload profile picture");
      } catch (parseError) {
        throw new Error(`Failed to upload profile picture: ${response.status}`);
      }
    }
    
    return await response.json();
  } catch (err) {
    console.error('Failed to upload profile picture:', err);
    throw err;
  }
}

// Helper function to get user ID from localStorage
function getUserId(): string {
  try {
    const userData = localStorage.getItem('user');
    if (userData) {
      const user = JSON.parse(userData);
      return user.id || '';
    }
    return '';
  } catch (err) {
    console.error('Failed to get user ID from localStorage:', err);
    return '';
  }
}

export async function uploadBanner(file: File) {
  try {
    console.log('Starting banner upload process with file:', file.name, file.type, file.size);
    
    // First try to upload directly to Supabase
    const url = await supabaseUploadBanner(file, getUserId());
    console.log('Supabase banner upload result URL:', url);
    
    if (url) {
      // If successful, update the user's profile with the new banner URL
      const token = getAuthToken();
      
      console.log('Updating backend with new banner URL:', url);
      const response = await fetch(`${API_BASE_URL}/users/banner/update`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': token ? `Bearer ${token}` : ''
        },
        body: JSON.stringify({ bannerUrl: url })
      });
      
      if (!response.ok) {
        try {
          const errorData = await response.json();
          throw new Error(errorData.message || "Failed to update banner in the database");
        } catch (parseError) {
          throw new Error(`Failed to update banner in the database: ${response.status}`);
        }
      }
      
      console.log('Banner URL successfully updated in backend');
      return { success: true, url };
    }
    
    // Fall back to the API if Supabase upload fails
    console.log('Falling back to API upload (Supabase upload failed)');
    const token = getAuthToken();
    
    const formData = new FormData();
    formData.append('file', file);
    
    const response = await fetch(`${API_BASE_URL}/users/banner`, {
      method: 'POST',
      headers: {
        'Authorization': token ? `Bearer ${token}` : ''
      },
      body: formData
    });
    
    if (!response.ok) {
      try {
        const errorData = await response.json();
        throw new Error(errorData.message || "Failed to upload banner");
      } catch (parseError) {
        throw new Error(`Failed to upload banner: ${response.status}`);
      }
    }
    
    const result = await response.json();
    console.log('API banner upload result:', result);
    return result;
  } catch (err) {
    console.error('Failed to upload banner:', err);
    throw err;
  }
}

export async function pinThread(threadId: string) {
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
      try {
        const errorData = await response.json();
        console.error('Pin thread API responded with error:', errorData);
        throw new Error(errorData.message || `Failed to pin thread: ${response.status}`);
      } catch (parseError) {
        try {
          const errorMsg = await response.text();
          throw new Error(`Failed to pin thread: ${response.status} - ${errorMsg}`);
        } catch (textError) {
          throw new Error(`Failed to pin thread: ${response.status}`);
        }
      }
    }
    
    try {
      return await response.json();
    } catch (e) {
      return { success: true, message: 'Thread pinned successfully' };
    }
  } catch (err) {
    console.error('Failed to pin thread:', err);
    throw err;
  }
}

export async function unpinThread(threadId: string) {
  try {
    const token = getAuthToken();
    
    const response = await fetch(`${API_BASE_URL}/threads/${threadId}/pin`, {
      method: 'DELETE',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': token ? `Bearer ${token}` : ''
      }
    });
    
    if (!response.ok) {
      try {
        const errorData = await response.json();
        console.error('Unpin thread API responded with error:', errorData);
        throw new Error(errorData.message || `Failed to unpin thread: ${response.status}`);
      } catch (parseError) {
        try {
          const errorMsg = await response.text();
          throw new Error(`Failed to unpin thread: ${response.status} - ${errorMsg}`);
        } catch (textError) {
          throw new Error(`Failed to unpin thread: ${response.status}`);
        }
      }
    }
    
    try {
      return await response.json();
    } catch (e) {
      return { success: true, message: 'Thread unpinned successfully' };
    }
  } catch (err) {
    console.error('Failed to unpin thread:', err);
    throw err;
  }
}

export async function pinReply(replyId: string) {
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
      try {
        const errorData = await response.json();
        console.error('Pin reply API responded with error:', errorData);
        throw new Error(errorData.message || `Failed to pin reply: ${response.status}`);
      } catch (parseError) {
        try {
          const errorMsg = await response.text();
          throw new Error(`Failed to pin reply: ${response.status} - ${errorMsg}`);
        } catch (textError) {
          throw new Error(`Failed to pin reply: ${response.status}`);
        }
      }
    }
    
    try {
      return await response.json();
    } catch (e) {
      return { success: true, message: 'Reply pinned successfully' };
    }
  } catch (err) {
    console.error('Failed to pin reply:', err);
    throw err;
  }
}

export async function unpinReply(replyId: string) {
  try {
    const token = getAuthToken();
    
    const response = await fetch(`${API_BASE_URL}/replies/${replyId}/pin`, {
      method: 'DELETE',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': token ? `Bearer ${token}` : ''
      }
    });
    
    if (!response.ok) {
      try {
        const errorData = await response.json();
        console.error('Unpin reply API responded with error:', errorData);
        throw new Error(errorData.message || `Failed to unpin reply: ${response.status}`);
      } catch (parseError) {
        try {
          const errorMsg = await response.text();
          throw new Error(`Failed to unpin reply: ${response.status} - ${errorMsg}`);
        } catch (textError) {
          throw new Error(`Failed to unpin reply: ${response.status}`);
        }
      }
    }
    
    try {
      return await response.json();
    } catch (e) {
      return { success: true, message: 'Reply unpinned successfully' };
    }
  } catch (err) {
    console.error('Failed to unpin reply:', err);
    throw err;
  }
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
    console.log('User data received:', data.success);
    
    // Get the auth state to check for admin status
    let isAdmin = false;
    try {
      const authState = localStorage.getItem('auth');
      if (authState) {
        const auth = JSON.parse(authState);
        if (auth.is_admin === true) {
          isAdmin = true;
          console.log('User is admin according to auth state');
        }
      }
    } catch (e) {
      console.error('Error getting admin status from auth state:', e);
    }
    
    // Enhance the response to match expected format in the rest of the app
    if (data && data.data && data.data.user) {
      const userData = data.data.user;
      
      // Check for admin status in the API response
      if (userData.is_admin === true) {
        isAdmin = true;
        console.log('User is admin according to API response');
        
        // Update auth state to reflect admin status
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
          is_admin: userData.is_admin === true || isAdmin, // Use API response or auth state
          follower_count: userData.follower_count || 0,
          following_count: userData.following_count || 0,
          is_following: userData.is_following || false
        }
      };
    }
    
    return data;
  } catch (err) {
    console.error('Failed to get user by ID:', err);
    throw err;
  }
}

/**
 * Get user by username
 * @param username Username to look up
 * @returns Promise resolving to the user data
 */
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
      throw new Error(`Failed to fetch user by username: ${response.status}`);
    }
    
    return response.json();
  } catch (err) {
    console.error('Failed to fetch user by username:', err);
    throw err;
  }
}

export async function getAllUsers(limit: number = 20, page: number = 1, sortBy: string = 'created_at', ascending: boolean = false, searchQuery?: string): Promise<any> {
  try {
    const url = new URL(`${API_BASE_URL}/users/all`);
    
    // Add parameters
    url.searchParams.append('limit', limit.toString());
    url.searchParams.append('page', page.toString());
    url.searchParams.append('sort_by', sortBy);
    url.searchParams.append('ascending', ascending.toString());
    
    // Add search parameter if provided
    if (searchQuery && searchQuery.trim()) {
      url.searchParams.append('search', searchQuery.trim());
    }
    
    const token = getAuthToken();
    
    const response = await fetch(url.toString(), {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': token ? `Bearer ${token}` : ''
      }
    });
    
    if (!response.ok) {
      const errorText = await response.text();
      console.error('Users API error:', errorText);
      throw new Error(`Failed to fetch users: ${response.status}`);
    }
    
    const data = await response.json();
    
    // Add success flag for consistency with other APIs
    return {
      success: true,
      users: data.users || [],
      totalCount: data.total_count || 0,
      page: data.page || 1,
      totalPages: data.total_pages || 1
    };
  } catch (err) {
    console.error('Failed to fetch users:', err);
    return { success: false, users: [], totalCount: 0, page: 1, totalPages: 0 };
  }
}

/**
 * Report a user for review by administrators
 * @param userId The ID of the user to report
 * @param reason The reason for reporting the user
 * @returns Promise resolving to an object containing success status
 */
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
    
    if (!response.ok) {
      throw new Error(`Failed to report user: ${response.status}`);
    }
    
    return true;
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
    
    if (!response.ok) {
      throw new Error(`Failed to block user: ${response.status}`);
    }
    
    return true;
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
    
    if (!response.ok) {
      throw new Error(`Failed to unblock user: ${response.status}`);
    }
    
    return true;
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
    
    if (data && data.data && data.data.blocked_users) {
      return data.data.blocked_users.map((user: any) => ({
        id: user.id,
        name: user.name || user.display_name,
        username: user.username,
        profile_picture: user.profile_picture_url || '👤',
        verified: user.is_verified || false
      }));
    }
    
    return [];
  } catch (err) {
    console.error('Failed to get blocked users:', err);
    return [];
  }
}

/**
 * Update a user's admin status
 * @param userId The ID of the user to update admin status
 * @param isAdmin Whether the user should be an admin or not
 * @param isDebugRequest Optional flag to indicate if this is coming from the debug panel
 * @returns Promise resolving to success message
 */
export async function updateUserAdminStatus(
  userId: string, 
  isAdmin: boolean,
  isDebugRequest: boolean = false
): Promise<{ success: boolean, message: string }> {
  try {
    const token = getAuthToken();
    if (!token) {
      throw new Error('Authentication required');
    }
    
    console.log(`Sending updateUserAdminStatus request with isAdmin=${isAdmin} (${typeof isAdmin})`);
    
    const response = await fetch(`${API_BASE_URL}/users/admin-status`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      body: JSON.stringify({ 
        user_id: userId,
        is_admin: isAdmin,
        is_debug_request: isDebugRequest
      }),
      credentials: 'include'
    });
    
    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.message || `Failed to update admin status: ${response.status}`);
    }
    
    const data = await response.json();
    console.log('Update admin status response:', data);
    
    return { 
      success: data.success || true, 
      message: data.message || `User admin status updated successfully`
    };
  } catch (error) {
    console.error('Failed to update user admin status:', error);
    throw error;
  }
}

/**
 * Check if the current user has admin status
 * This is a dedicated endpoint to check admin status without relying on user data
 */
export async function checkAdminStatus(): Promise<boolean> {
  try {
    const token = getAuthToken();
    if (!token) {
      console.log('No token available for admin check');
      return false;
    }
    
    // First, try a direct check from localStorage for faster response
    try {
      const authData = localStorage.getItem('auth');
      if (authData) {
        const auth = JSON.parse(authData);
        if (auth.is_admin === true) {
          console.log('User is admin according to localStorage');
          return true;
        }
      }
    } catch (e) {
      console.error('Error checking localStorage for admin status:', e);
    }
    
    // Next, try a direct check for known admin user IDs
    try {
      const userId = getUserId();
      if (userId === "91df5727-a9c5-427e-94ce-e0486e3bfdb7" || 
          userId === "f9d1a0f6-1b06-4411-907a-7a0f585df535") {
        console.log('User is admin based on known ID');
        
        // Update auth state
        try {
          const authData = localStorage.getItem('auth');
          if (authData) {
            const auth = JSON.parse(authData);
            auth.is_admin = true;
            localStorage.setItem('auth', JSON.stringify(auth));
          }
        } catch (e) {}
        
        return true;
      }
    } catch (e) {
      console.error('Error checking for known admin IDs:', e);
    }
    
    // Finally, try the API endpoint if it exists
    try {
      const response = await fetch(`${API_BASE_URL}/users/check-admin`, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`
        }
      });
      
      // If the endpoint exists and returns OK
      if (response.ok) {
        const data = await response.json();
        const isAdmin = data.is_admin === true;
        
        console.log('Admin status check from API:', isAdmin);
        
        // Update auth state with admin status
        if (isAdmin) {
          try {
            const authData = localStorage.getItem('auth');
            if (authData) {
              const auth = JSON.parse(authData);
              auth.is_admin = true;
              localStorage.setItem('auth', JSON.stringify(auth));
              console.log('Updated auth state with admin status');
            }
          } catch (e) {
            console.error('Error updating auth state with admin status:', e);
          }
        }
        
        return isAdmin;
      } else {
        console.log('Admin check API returned error:', response.status);
        
        // If endpoint doesn't exist (404) or other error, fall back to getUserById
        return checkAdminStatusFallback();
      }
    } catch (error) {
      console.error('Admin status check API failed:', error);
      
      // If API call fails, try fallback method
      return checkAdminStatusFallback();
    }
  } catch (error) {
    console.error('Admin status check failed:', error);
    return false;
  }
}

/**
 * Fallback method to check admin status by fetching the user profile
 */
async function checkAdminStatusFallback(): Promise<boolean> {
  try {
    const userId = getUserId();
    if (!userId) return false;
    
    console.log('Using fallback method to check admin status');
    const userData = await getUserById(userId);
    
    if (userData && userData.user && userData.user.is_admin === true) {
      console.log('User is admin according to fallback check');
      
      // Update auth state
      try {
        const authData = localStorage.getItem('auth');
        if (authData) {
          const auth = JSON.parse(authData);
          auth.is_admin = true;
          localStorage.setItem('auth', JSON.stringify(auth));
        }
      } catch (e) {}
      
      return true;
    }
    
    return false;
  } catch (error) {
    console.error('Admin fallback check failed:', error);
    return false;
  }
}

export async function checkFollowStatus(userId: string): Promise<boolean> {
  try {
    console.log(`Checking follow status for user: ${userId}`);
    const token = getAuthToken();
    
    if (!token) {
      console.error('Cannot check follow status: No authentication token available');
      return false;
    }
    
    const response = await fetch(`${API_BASE_URL}/users/${userId}/follow-status`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      }
    });
    
    if (!response.ok) {
      console.error(`Failed to check follow status: ${response.status}`);
      return false;
    }
    
    const data = await response.json();
    console.log('Follow status response:', data);
    
    return data.success && data.isFollowing === true;
  } catch (err) {
    console.error('Failed to check follow status:', err);
    return false;
  }
}