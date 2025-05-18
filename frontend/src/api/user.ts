import appConfig from '../config/appConfig';
import { getAuthToken } from '../utils/auth';
import { uploadProfilePicture as supabaseUploadProfilePicture, uploadBanner as supabaseUploadBanner } from '../utils/supabase';

const API_BASE_URL = appConfig.api.baseUrl;

export async function getProfile() {
  const token = getAuthToken();
  
  const response = await fetch(`${API_BASE_URL}/users/profile`, {
    method: "GET",
    headers: { 
      "Content-Type": "application/json",
      "Authorization": token ? `Bearer ${token}` : ''
    },
    credentials: "include",
  });
  if (!response.ok) {
    try {
      const errorData = await response.json();
      throw new Error(errorData.message || "Failed to fetch user profile");
    } catch (parseError) {
      throw new Error("Failed to fetch user profile");
    }
  }
  return response.json();
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

export async function followUser(userId: string): Promise<boolean> {
  try {
    const token = getAuthToken();
    
    const response = await fetch(`${API_BASE_URL}/users/${userId}/follow`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': token ? `Bearer ${token}` : ''
      }
    });
    
    if (!response.ok) {
      throw new Error(`Failed to follow user: ${response.status}`);
    }
    
    return true;
  } catch (err) {
    console.error('Failed to follow user:', err);
    return false;
  }
}

export async function unfollowUser(userId: string): Promise<boolean> {
  try {
    const token = getAuthToken();
    
    const response = await fetch(`${API_BASE_URL}/users/${userId}/unfollow`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': token ? `Bearer ${token}` : ''
      }
    });
    
    if (!response.ok) {
      throw new Error(`Failed to unfollow user: ${response.status}`);
    }
    
    return true;
  } catch (err) {
    console.error('Failed to unfollow user:', err);
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
    
    const data = await response.json();
    
    // Apply client-side fuzzy search if requested and API doesn't support it
    if (options?.clientFuzzy && query.trim()) {
      // Import the fuzzy search function
      const { fuzzySearch } = await import('../utils/helpers');
      
      // If the API doesn't indicate it did fuzzy searching, do it client-side
      if (!data.fuzzy_search_applied) {
        // Apply fuzzy search on both username and display_name
        const allUsers = [...(data.users || [])];
        const fuzzyMatches = fuzzySearch(
          query,
          allUsers,
          'username',
          0.5 // Lower threshold for username matches
        );
        
        const nameMatches = fuzzySearch(
          query,
          allUsers.filter(user => !fuzzyMatches.includes(user)), // Remove already matched users
          'display_name',
          0.6
        );
        
        // Combine and keep original order if possible
        data.users = [...new Set([...fuzzyMatches, ...nameMatches])];
        console.log(`Applied client-side fuzzy search, found ${data.users.length} matches`);
      }
    }
    
    // Standardize the response format
    return {
      users: data.users || [],
      totalCount: data.total_count || 0,
      page: data.page || page,
      totalPages: data.total_pages || 1
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

export async function getUserById(userId: string) {
  try {
    const token = getAuthToken();
    
    // Check if the userId looks like a UUID (basic check)
    const uuidPattern = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;
    const isUuid = uuidPattern.test(userId);
    
    // If it doesn't look like a UUID, assume it's a username and try that endpoint
    if (!isUuid && !userId.match(/^\d+$/)) {
      console.log(`Input '${userId}' doesn't look like an ID, trying to fetch by username instead`);
      return getUserByUsername(userId);
    }
    
    console.log(`Fetching user with ID: ${userId}`);
    const response = await fetch(`${API_BASE_URL}/users/${userId}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': token ? `Bearer ${token}` : ''
      }
    });
    
    if (!response.ok) {
      if (response.status === 404) {
        // If ID lookup fails with 404, try username as fallback
        console.log(`User ID ${userId} not found, trying as username`);
        return getUserByUsername(userId);
      }
      throw new Error(`Failed to fetch user: ${response.status}`);
    }
    
    return response.json();
  } catch (err) {
    console.error('Failed to fetch user:', err);
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

export async function getAllUsers(limit: number = 20, page: number = 1, sortBy: string = 'created_at', ascending: boolean = false): Promise<any> {
  try {
    const url = new URL(`${API_BASE_URL}/users/all`);
    
    // Add parameters
    url.searchParams.append('limit', limit.toString());
    url.searchParams.append('page', page.toString());
    url.searchParams.append('sort_by', sortBy);
    url.searchParams.append('ascending', ascending.toString());
    
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
    
    // Return the users in the standardized format
    return {
      users: data.users || [],
      totalCount: data.total_count || 0,
      page: data.page || 1,
      totalPages: data.total_pages || 1
    };
  } catch (err) {
    console.error('Failed to fetch users:', err);
    return { users: [], totalCount: 0, page: 1, totalPages: 0 };
  }
}

/**
 * Report a user for review by administrators
 * @param userId The ID of the user to report
 * @param reason The reason for reporting the user
 * @returns Promise resolving to an object containing success status
 */
export async function reportUser(userId: string, reason: string): Promise<{ success: boolean, message?: string }> {
  try {
    const token = getAuthToken();
    
    const response = await fetch(`${API_BASE_URL}/users/${userId}/report`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': token ? `Bearer ${token}` : ''
      },
      body: JSON.stringify({ reason }),
      credentials: 'include'
    });
    
    if (!response.ok) {
      try {
        const errorData = await response.json();
        throw new Error(errorData.message || `Failed to report user: ${response.status}`);
      } catch (parseError) {
        throw new Error(`Failed to report user: ${response.status}`);
      }
    }
    
    const data = await response.json();
    return { success: true, message: data.message || 'User reported successfully' };
  } catch (err) {
    console.error('Failed to report user:', err);
    throw err;
  }
}

/**
 * Block a user to prevent them from seeing your content and vice versa
 * @param userId The ID of the user to block
 * @returns Promise resolving to an object containing success status
 */
export async function blockUser(userId: string): Promise<{ success: boolean, message?: string }> {
  try {
    const token = getAuthToken();
    
    const response = await fetch(`${API_BASE_URL}/users/${userId}/block`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': token ? `Bearer ${token}` : ''
      },
      credentials: 'include'
    });
    
    if (!response.ok) {
      try {
        const errorData = await response.json();
        throw new Error(errorData.message || `Failed to block user: ${response.status}`);
      } catch (parseError) {
        throw new Error(`Failed to block user: ${response.status}`);
      }
    }
    
    const data = await response.json();
    return { success: true, message: data.message || 'User blocked successfully' };
  } catch (err) {
    console.error('Failed to block user:', err);
    throw err;
  }
}

/**
 * Unblock a previously blocked user
 * @param userId The ID of the user to unblock
 * @returns Promise resolving to an object containing success status
 */
export async function unblockUser(userId: string): Promise<{ success: boolean, message?: string }> {
  try {
    const token = getAuthToken();
    
    const response = await fetch(`${API_BASE_URL}/users/${userId}/block`, {
      method: 'DELETE',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': token ? `Bearer ${token}` : ''
      },
      credentials: 'include'
    });
    
    if (!response.ok) {
      try {
        const errorData = await response.json();
        throw new Error(errorData.message || `Failed to unblock user: ${response.status}`);
      } catch (parseError) {
        throw new Error(`Failed to unblock user: ${response.status}`);
      }
    }
    
    const data = await response.json();
    return { success: true, message: data.message || 'User unblocked successfully' };
  } catch (err) {
    console.error('Failed to unblock user:', err);
    throw err;
  }
}

/**
 * Get a list of users that the current user has blocked
 * @returns Promise resolving to an array of blocked users
 */
export async function getBlockedUsers(): Promise<any[]> {
  try {
    const token = getAuthToken();
    
    const response = await fetch(`${API_BASE_URL}/users/blocked`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': token ? `Bearer ${token}` : ''
      },
      credentials: 'include'
    });
    
    if (!response.ok) {
      throw new Error(`Failed to get blocked users: ${response.status}`);
    }
    
    const data = await response.json();
    return data.blocked_users || [];
  } catch (err) {
    console.error('Failed to get blocked users:', err);
    return [];
  }
} 