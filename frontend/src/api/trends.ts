import { getAuthToken } from '../utils/auth';
import appConfig from '../config/appConfig';
import type { ITrend } from '../interfaces/ISocialMedia';
import { createLoggerWithPrefix } from '../utils/logger';

const API_BASE_URL = appConfig.api.baseUrl;
const logger = createLoggerWithPrefix('trends-api');

export async function getTrends(limit: number = 5): Promise<ITrend[]> {
  try {
    const token = getAuthToken();
    
    try {
      logger.debug('Attempting to fetch trends from API');
      const response = await fetch(`${API_BASE_URL}/trends?limit=${limit}`, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': token ? `Bearer ${token}` : ''
        },
        credentials: 'include'
      });
      
      if (response.ok) {
        const data = await response.json();
        
        if (data && data.trends && Array.isArray(data.trends)) {
          logger.info('Successfully fetched trends from API');
          return data.trends.map((trend: any) => ({
            id: trend.id,
            category: trend.category || 'Trending',
            title: trend.title,
            postCount: trend.post_count || 0
          }));
        }
      }
      
      // If API call fails with 401 or other error, fall back to mock data
      logger.warn('API call failed, falling back to mock data');
      throw new Error('Falling back to mock data');
    } catch (apiError) {
      // Attempt to load mock data
      logger.debug('Loading mock trends data');
      const mockResponse = await fetch('/mock-data/trends.json');
      
      if (mockResponse.ok) {
        const mockData = await mockResponse.json();
        
        if (mockData && mockData.trends && Array.isArray(mockData.trends)) {
          logger.info('Successfully loaded mock trends data');
          return mockData.trends.map((trend: any) => ({
            id: trend.id,
            category: trend.category || 'Trending',
            title: trend.title,
            postCount: trend.post_count || 0
          }));
        }
      }
      
      logger.error('Failed to load mock trends data');
      return [];
    }
  } catch (error) {
    logger.error('Unexpected error in getTrends', error);
    return [];
  }
} 