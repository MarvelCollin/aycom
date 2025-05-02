import { getAuthToken } from '../utils/auth';
import appConfig from '../config/appConfig';
import type { ISuggestedFollow } from '../interfaces/ISocialMedia';

const API_BASE_URL = appConfig.api.baseUrl;

export async function getSuggestedUsers(limit: number = 3): Promise<ISuggestedFollow[]> {
  try {
    const token = getAuthToken();
    
    const response = await fetch(`${API_BASE_URL}/users/suggestions?limit=${limit}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': token ? `Bearer ${token}` : ''
      },
      credentials: 'include'
    });
    
    if (!response.ok) {
      try {
        const errorData = await response.json();
        throw new Error(errorData.message || `Failed to fetch suggested users: ${response.status}`);
      } catch (parseError) {
        throw new Error(`Failed to fetch suggested users: ${response.status}`);
      }
    }
    
    const data = await response.json();
    
    if (data && data.users) {
      return data.users.map((user: any) => ({
        username: user.username,
        displayName: user.display_name || user.username,
        avatar: user.avatar_url || null,
        verified: user.verified || false,
        followerCount: user.follower_count || 0
      }));
    }
    
    return [];
  } catch (error) {
    throw error;
  }
} 