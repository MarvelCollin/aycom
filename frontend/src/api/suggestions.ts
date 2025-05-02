import { getAuthToken } from '../utils/auth';
import type { ISuggestedFollow } from '../interfaces/ISocialMedia';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8081/api/v1';

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
      throw new Error(`Failed to fetch suggested users: ${response.status}`);
    }
    
    const data = await response.json();
    
    if (data && data.users) {
      return data.users.map((user: any) => ({
        username: user.username,
        displayName: user.display_name || user.username,
        avatar: user.avatar_url || '👤',
        verified: user.verified || false,
        followerCount: user.follower_count || 0
      }));
    }
    
    return [];
  } catch (error) {
    console.error('Error fetching suggested users:', error);
    throw error;
  }
} 