import appConfig from '../config/appConfig';
import type { ISuggestedFollow } from '../interfaces/ISocialMedia';
import { createLoggerWithPrefix } from '../utils/logger';

const API_BASE_URL = appConfig.api.baseUrl;
const logger = createLoggerWithPrefix('suggestions-api');

export async function getSuggestedUsers(limit: number = 3): Promise<ISuggestedFollow[]> {
  try {
    const url = `${API_BASE_URL}/users/all?limit=${limit}`;
    logger.debug(`Fetching suggested users from ${url}`);
    
    const response = await fetch(url, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json'
      },
      credentials: 'include'
    });
    
    if (!response.ok) {
      logger.error(`Failed to fetch suggestions: ${response.status}`);
      return [];
    }
    
    const data = await response.json();
    
    if (!data || !data.users || !Array.isArray(data.users)) {
      logger.warn('Invalid data format from suggestions API');
      return [];
    }

    logger.info(`Successfully fetched ${data.users.length} user suggestions from API`);

    return data.users.map((user: any) => ({
      id: user.id, 
      username: user.username,
      name: user.name || user.display_name, 
      profile_picture_url: user.profile_picture_url,
      is_verified: user.is_verified || false,
      follower_count: user.follower_count || 0,
      is_following: user.is_following || false
    }));

  } catch (error: any) {
    logger.error('Failed to fetch suggested users', { error: error.message });
    return [];
  }
}

