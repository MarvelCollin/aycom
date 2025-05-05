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

// Search users functionality
export async function searchUsers(query: string, page: number = 1, limit: number = 10, options?: any) {
  try {
    const url = new URL(`${import.meta.env.VITE_API_BASE_URL || 'http://localhost:8083/api/v1'}/users/search`);
    
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
      throw new Error(`Failed to search users: ${response.status}`);
    }
    
    return await response.json();
  } catch (error) {
    console.error('Error searching users:', error);
    // Mock data for development
    return {
      users: [
        {
          id: '1',
          username: 'johndoe',
          name: 'John Doe',
          profile_picture_url: null,
          bio: 'Web developer and tech enthusiast',
          is_verified: true,
          follower_count: 1542,
          is_following: false
        },
        {
          id: '2',
          username: 'janedoe',
          name: 'Jane Doe',
          profile_picture_url: null,
          bio: 'Designer and creator',
          is_verified: false,
          follower_count: 876,
          is_following: true
        },
        {
          id: '3',
          username: 'alexsmith',
          name: 'Alex Smith',
          profile_picture_url: null,
          bio: 'Software engineer',
          is_verified: false,
          follower_count: 342,
          is_following: false
        }
      ]
    };
  }
} 