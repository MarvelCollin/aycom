import appConfig from '../config/appConfig';
import { getAuthToken } from '../utils/auth';

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
  
  const response = await fetch(`${API_BASE_URL}/users/profile`, {
    method: "PUT",
    headers: { 
      "Content-Type": "application/json",
      "Authorization": token ? `Bearer ${token}` : ''
    },
    body: JSON.stringify(data),
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

// Added functions from useProfile.ts
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

// Search users functionality
export async function searchUsers(query: string, page: number = 1, limit: number = 10, options?: any) {
  try {
    const url = new URL(`${API_BASE_URL}/users/search`);
    
    // Set query parameters
    url.searchParams.append('q', query);
    url.searchParams.append('page', page.toString());
    url.searchParams.append('limit', limit.toString());
    
    // Add optional filters
    if (options?.filter === 'following') {
      url.searchParams.append('filter', 'following');
    } else if (options?.filter === 'verified') {
      url.searchParams.append('filter', 'verified');
    }
    
    // Add sorting if provided
    if (options?.sortBy) {
      url.searchParams.append('sort', options.sortBy);
    }
    
    // Get token
    const token = getAuthToken();
    
    // Make request
    const response = await fetch(url.toString(), {
      method: 'GET',
      headers: {
        'Authorization': token ? `Bearer ${token}` : '',
        'Content-Type': 'application/json'
      }
    });
    
    if (!response.ok) {
      throw new Error(`Failed to search users: ${response.status}`);
    }
    
    return await response.json();
  } catch (error) {
    console.error('Error searching users:', error);
    // Return empty results instead of mock data in production
    return {
      users: [],
      total_count: 0
    };
  }
}

// Upload profile picture
export async function uploadProfilePicture(file: File) {
  const token = getAuthToken();
  
  const formData = new FormData();
  formData.append('file', file);
  formData.append('type', 'profile_picture');
  
  const response = await fetch(`${API_BASE_URL}/users/media`, {
    method: 'POST',
    headers: { 
      "Authorization": token ? `Bearer ${token}` : ''
    },
    body: formData,
    credentials: "include",
  });
  
  if (!response.ok) {
    try {
      const errorData = await response.json();
      throw new Error(errorData.message || "Failed to upload profile picture");
    } catch (parseError) {
      throw new Error("Failed to upload profile picture");
    }
  }
  return response.json();
}

// Upload banner
export async function uploadBanner(file: File) {
  const token = getAuthToken();
  
  const formData = new FormData();
  formData.append('file', file);
  formData.append('type', 'banner');
  
  const response = await fetch(`${API_BASE_URL}/users/media`, {
    method: 'POST',
    headers: { 
      "Authorization": token ? `Bearer ${token}` : ''
    },
    body: formData,
    credentials: "include",
  });
  
  if (!response.ok) {
    try {
      const errorData = await response.json();
      throw new Error(errorData.message || "Failed to upload banner");
    } catch (parseError) {
      throw new Error("Failed to upload banner");
    }
  }
  return response.json();
}

// Pin and unpin threads/replies
export async function pinThread(threadId: string) {
  try {
    const token = getAuthToken();
    if (!token) {
      throw new Error('Authentication required');
    }
    
    const response = await fetch(`${API_BASE_URL}/threads/${threadId}/pin`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      }
    });
    
    if (!response.ok) {
      // Try to get more details from the error response
      try {
        const errorData = await response.json();
        throw new Error(errorData.message || `Failed to pin thread: ${response.status}`);
      } catch (parseError) {
        // If we can't parse the error, just use the status code
        throw new Error(`Failed to pin thread: ${response.status}`);
      }
    }
    
    // Return a consistent response format even if the server returns empty
    try {
      return await response.json();
    } catch (e) {
      // If the response can't be parsed as JSON, return a default success object
      return { success: true, message: "Thread pinned successfully" };
    }
  } catch (err) {
    console.error('Failed to pin thread:', err);
    throw err;
  }
}

export async function unpinThread(threadId: string) {
  try {
    const token = getAuthToken();
    if (!token) {
      throw new Error('Authentication required');
    }
    
    const response = await fetch(`${API_BASE_URL}/threads/${threadId}/pin`, {
      method: 'DELETE',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      }
    });
    
    if (!response.ok) {
      // Try to get more details from the error response
      try {
        const errorData = await response.json();
        throw new Error(errorData.message || `Failed to unpin thread: ${response.status}`);
      } catch (parseError) {
        // If we can't parse the error, just use the status code
        throw new Error(`Failed to unpin thread: ${response.status}`);
      }
    }
    
    // Return a consistent response format even if the server returns empty
    try {
      return await response.json();
    } catch (e) {
      // If the response can't be parsed as JSON, return a default success object
      return { success: true, message: "Thread unpinned successfully" };
    }
  } catch (err) {
    console.error('Failed to unpin thread:', err);
    throw err;
  }
}

export async function pinReply(replyId: string) {
  try {
    const token = getAuthToken();
    if (!token) {
      throw new Error('Authentication required');
    }
    
    const response = await fetch(`${API_BASE_URL}/replies/${replyId}/pin`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      }
    });
    
    if (!response.ok) {
      // Try to get more details from the error response
      try {
        const errorData = await response.json();
        throw new Error(errorData.message || `Failed to pin reply: ${response.status}`);
      } catch (parseError) {
        // If we can't parse the error, just use the status code
        throw new Error(`Failed to pin reply: ${response.status}`);
      }
    }
    
    // Return a consistent response format even if the server returns empty
    try {
      return await response.json();
    } catch (e) {
      // If the response can't be parsed as JSON, return a default success object
      return { success: true, message: "Reply pinned successfully" };
    }
  } catch (err) {
    console.error('Failed to pin reply:', err);
    throw err;
  }
}

export async function unpinReply(replyId: string) {
  try {
    const token = getAuthToken();
    if (!token) {
      throw new Error('Authentication required');
    }
    
    const response = await fetch(`${API_BASE_URL}/replies/${replyId}/pin`, {
      method: 'DELETE',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      }
    });
    
    if (!response.ok) {
      // Try to get more details from the error response
      try {
        const errorData = await response.json();
        throw new Error(errorData.message || `Failed to unpin reply: ${response.status}`);
      } catch (parseError) {
        // If we can't parse the error, just use the status code
        throw new Error(`Failed to unpin reply: ${response.status}`);
      }
    }
    
    // Return a consistent response format even if the server returns empty
    try {
      return await response.json();
    } catch (e) {
      // If the response can't be parsed as JSON, return a default success object
      return { success: true, message: "Reply unpinned successfully" };
    }
  } catch (err) {
    console.error('Failed to unpin reply:', err);
    throw err;
  }
} 