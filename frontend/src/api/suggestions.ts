import { getAuthToken } from '../utils/auth';
import appConfig from '../config/appConfig';
import type { ISuggestedFollow } from '../interfaces/ISocialMedia';
import { createLoggerWithPrefix } from '../utils/logger';

const API_BASE_URL = appConfig.api.baseUrl;
const logger = createLoggerWithPrefix('suggestions-api');

// Always use this endpoint for fetching real user data
const USERS_ENDPOINT = `${API_BASE_URL}/users/all`;

export async function getSuggestedUsers(limit: number = 3): Promise<ISuggestedFollow[]> {
  try {
    // Use the dedicated recommendations endpoint instead of the general users endpoint
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
      // Map the API data to our interface with consistent field names
    return data.users.map((user: any) => ({
      id: user.id, // Use 'id' as expected by ISuggestedFollow interface
      username: user.username,
      name: user.name || user.display_name, // Prefer 'name' over 'display_name'
      profile_picture_url: user.profile_picture_url,
      is_verified: user.is_verified || false,
      follower_count: user.follower_count || 0,
      is_following: user.is_following || false
    }));
    
  } catch (error: any) {
    logger.error('Failed to fetch suggested users', { error: error.message });
    throw error; // Remove fallback, let error bubble up
  }
}

// This function has been removed as part of removing mock data implementations
// All user suggestions now come from real API endpoints