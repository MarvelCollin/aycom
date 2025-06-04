import { getAuthToken } from '../utils/auth';
import appConfig from '../config/appConfig';
import type { ITrend } from '../interfaces/ITrend';
import { createLoggerWithPrefix } from '../utils/logger';

const API_BASE_URL = appConfig.api.baseUrl;
const logger = createLoggerWithPrefix('trends-api');

export async function getTrends(limit: number = 5): Promise<ITrend[]> {
  try {
    logger.debug('Fetching trends from API', { limit });
    const token = getAuthToken();

    const response = await fetch(`${API_BASE_URL}/trends?limit=${limit}`, {
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
      logger.error(`Failed to fetch trends: ${errorMessage}`);

      if (response.status === 401) {
        logger.warn('Unauthorized access to trends API');
        return [];
    }

      throw new Error(errorMessage);
    }

    const data = await response.json();

    if (!data || !data.trends || !Array.isArray(data.trends)) {
      logger.warn('API returned invalid trends data format');
      return [];
    }

    logger.info('Successfully fetched trends from API', { count: data.trends.length });
    return data.trends.map((trend: any) => ({
      id: trend.id || `trend-${Math.random().toString(36).substring(2, 9)}`,
      category: trend.category || 'Trending',
      title: trend.title,
      post_count: trend.post_count || 0
    }));
  } catch (error) {
    logger.error('Error fetching trends:', error);
    throw error;
  }
}

