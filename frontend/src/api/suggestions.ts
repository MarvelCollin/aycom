import { getAuthToken } from '../utils/auth';
import appConfig from '../config/appConfig';
import type { ISuggestedFollow } from '../interfaces/ISocialMedia';
import { createLoggerWithPrefix } from '../utils/logger';

const API_BASE_URL = appConfig.api.baseUrl;
const logger = createLoggerWithPrefix('suggestions-api');

const USERS_ENDPOINT = `${API_BASE_URL}/users/all`;

export async function getSuggestedUsers(limit: number = 3): Promise<ISuggestedFollow[]> {
  try {

    const token = getAuthToken();
    const response = await fetch(`${API_BASE_URL}/users/recommendations?limit=${limit}`, {
      method: 'GET',
      headers: { 
        'Content-Type': 'application/json',
        'Authorization': token ? `Bearer ${token}` : ''
      },
      credentials: 'include'
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      logger.error(`Failed to fetch user recommendations: ${response.status}`, errorData);
      throw new Error(errorData.message || `API returned status ${response.status}`);
    }

    const data = await response.json();

    if (!data || !data.users || !Array.isArray(data.users)) {
      logger.error('Invalid data format from user recommendations API');
      throw new Error('Invalid response format from server');
    }

    logger.info(`Successfully fetched ${data.users.length} user recommendations from API`);

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
    throw error; 
  }
}

