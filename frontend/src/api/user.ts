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

export async function uploadBanner(file: File) {
  try {
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
    
    return await response.json();
  } catch (err) {
    console.error('Failed to upload banner:', err);
    throw err;
  }
}

export async function pinThread(threadId: string) {
  try {
    const token = getAuthToken();
    
    const response = await fetch(`${API_BASE_URL}/threads/${threadId}/pin`, {
      method: 'PUT',
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
    
    const response = await fetch(`${API_BASE_URL}/threads/${threadId}/unpin`, {
      method: 'PUT',
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
      method: 'PUT',
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
    
    const response = await fetch(`${API_BASE_URL}/replies/${replyId}/unpin`, {
      method: 'PUT',
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
    
    const response = await fetch(`${API_BASE_URL}/users/${userId}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': token ? `Bearer ${token}` : ''
      }
    });
    
    if (!response.ok) {
      throw new Error(`Failed to fetch user: ${response.status}`);
    }
    
    return response.json();
  } catch (err) {
    console.error('Failed to fetch user:', err);
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