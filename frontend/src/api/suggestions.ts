import { getAuthToken } from '../utils/auth';
import appConfig from '../config/appConfig';
import type { ISuggestedFollow } from '../interfaces/ISocialMedia';
import { createLoggerWithPrefix } from '../utils/logger';

const API_BASE_URL = appConfig.api.baseUrl;
const logger = createLoggerWithPrefix('suggestions-api');

export async function getSuggestedUsers(limit: number = 3): Promise<ISuggestedFollow[]> {
  try {
    const apiUsers = await getSuggestedUsersFromAPI(limit);
    if (apiUsers.length > 0) {
      return apiUsers;
    }
    
    logger.info('API returned no suggestions, using mock data instead');
    return getMockSuggestedUsers(limit);
  } catch (error) {
    logger.error('Error fetching suggested users from API:', error);
    try {
      logger.info('Falling back to mock data for user suggestions');
      return getMockSuggestedUsers(limit);
    } catch (mockError) {
      logger.error('Mock data fallback also failed:', mockError);
      return [];
    }
  }
}

function getMockSuggestedUsers(limit: number): ISuggestedFollow[] {
  const mockUsers = [
    { 
      username: 'tech_insider', 
      displayName: 'Tech Insider', 
      avatar: 'https://i.pravatar.cc/150?u=tech_insider', 
      verified: true,
      followerCount: 1240000,
      isFollowing: false
    },
    { 
      username: 'travel_adventures', 
      displayName: 'Travel Adventures', 
      avatar: 'https://i.pravatar.cc/150?u=travel_adventures', 
      verified: true,
      followerCount: 890000,
      isFollowing: false
    },
    { 
      username: 'photo_daily', 
      displayName: 'Photography Daily', 
      avatar: 'https://i.pravatar.cc/150?u=photo_daily', 
      verified: false,
      followerCount: 625000,
      isFollowing: false
    },
    { 
      username: 'food_lovers', 
      displayName: 'Food Lovers', 
      avatar: 'https://i.pravatar.cc/150?u=food_lovers', 
      verified: true,
      followerCount: 520000,
      isFollowing: false
    },
    { 
      username: 'fitness_coach', 
      displayName: 'Fitness Coach', 
      avatar: 'https://i.pravatar.cc/150?u=fitness_coach', 
      verified: false,
      followerCount: 480000,
      isFollowing: false
    }
  ];
  
  return mockUsers.slice(0, limit);
}

async function getSuggestedUsersFromAPI(limit: number): Promise<ISuggestedFollow[]> {
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
    followerCount: user.follower_count || 0,
    isFollowing: user.is_following || false
  }));
}