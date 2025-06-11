import { getAuthToken } from '../utils/auth';
import appConfig from '../config/appConfig';
import type { ITrend } from '../interfaces/ITrend';
import { createLoggerWithPrefix } from '../utils/logger';
import { useAuth } from '../hooks/useAuth';

const API_BASE_URL = appConfig.api.baseUrl;
const logger = createLoggerWithPrefix('trends-api');

export async function getTrends(limit: number = 5): Promise<ITrend[]> {
  try {
    const { checkAndRefreshTokenIfNeeded } = useAuth();
    await checkAndRefreshTokenIfNeeded();
    
    const token = getAuthToken();
    const url = `${API_BASE_URL}/trends?limit=${limit}`;
    
    logger.debug(`Fetching trends from API: ${url}`);
    
    const response = await fetch(url, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': token ? `Bearer ${token}` : ''
      },
      credentials: 'include'
    });
    
    // For public endpoints, if we get a 401 error when a token was provided,
    // try again without the token
    if (!response.ok && response.status === 401 && token) {
      logger.debug('Got 401, retrying without auth token');
      
      const publicResponse = await fetch(url, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json'
        },
        credentials: 'include'
      });
      
      if (publicResponse.ok) {
        const data = await publicResponse.json();
        if (!data || !data.trends || !Array.isArray(data.trends)) {
          logger.warn('Invalid data format from trends API (public response)');
          return [];
        }
        
        return data.trends;
      }
      
      // If the second request also fails, fall through to the empty array return
      logger.error(`Failed to fetch trends without auth: ${publicResponse.status}`);
      return [];
    }
    
    // Handle any other error case
    if (!response.ok) {
      logger.error(`Failed to fetch trends: ${response.status}`);
      return [];
    }
    
    const data = await response.json();
    
    // Handle different API response formats
    if (data && data.data && Array.isArray(data.data.trends)) {
      // Format: { data: { trends: [...] } }
      return data.data.trends;
    } else if (data && Array.isArray(data.trends)) {
      // Format: { trends: [...] }
      return data.trends;
    } else {
      logger.warn('Invalid data format from trends API', { data });
      return [];
    }
    
    logger.info(`Successfully fetched ${data.trends.length} trends from API`);
    
    return data.trends;
  } catch (error: any) {
    logger.error('Failed to fetch trends', { error: error.message });
    return [];
  }
}

