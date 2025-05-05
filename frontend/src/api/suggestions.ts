import { getAuthToken } from '../utils/auth';
import appConfig from '../config/appConfig';
import type { ISuggestedFollow } from '../interfaces/ISocialMedia';
import { createLoggerWithPrefix } from '../utils/logger';

const API_BASE_URL = appConfig.api.baseUrl;
const logger = createLoggerWithPrefix('suggestions-api');

/**
 * Get suggested users to follow from the API
 * @param limit Number of suggestions to fetch
 * @returns Array of suggested users
 */
export async function getSuggestedUsers(limit: number = 3): Promise<ISuggestedFollow[]> {
  try {
    const token = getAuthToken();
    
    logger.debug('Fetching suggested users from API', { limit });
    const response = await fetch(`${API_BASE_URL}/users/suggestions?limit=${limit}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': token ? `Bearer ${token}` : ''
      },
      credentials: 'include'
    });
    
    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      const errorMessage = errorData.message || 
        `Error ${response.status}: ${response.statusText}`;
      logger.error(`Failed to fetch suggested users: ${errorMessage}`);
      throw new Error(errorMessage);
    }
    
    const data = await response.json();
    
    if (!data || !data.users || !Array.isArray(data.users)) {
      logger.warn('API returned invalid users data format');
      return [];
    }
    
    logger.info('Successfully fetched suggested users from API', { count: data.users.length });
    return data.users.map((user: any) => ({
      username: user.username,
      displayName: user.display_name || user.username,
      avatar: user.avatar_url || null,
      verified: user.verified || false,
      followerCount: user.follower_count || 0
    }));
  } catch (error) {
    logger.error('Error fetching suggested users:', error);
    throw error;
  }
}