import { getAuthToken } from '../utils/auth';
import appConfig from '../config/appConfig';
import type { ITrend } from '../interfaces/ISocialMedia';
import { createLoggerWithPrefix } from '../utils/logger';

const API_BASE_URL = appConfig.api.baseUrl;
const logger = createLoggerWithPrefix('trends-api');

/**
 * Get trending topics from the API or database
 * @param limit Number of trends to fetch
 * @returns Array of trending topics
 */
export async function getTrends(limit: number = 5): Promise<ITrend[]> {
  try {
    // Try to get from API first
    const apiTrends = await getTrendsFromAPI(limit);
    if (apiTrends.length > 0) {
      return apiTrends;
    }
    
    // Fallback to mock data if API returns no data
    logger.info('API returned no trends, using mock data instead');
    return getMockTrends(limit);
  } catch (error) {
    logger.error('Error fetching trends:', error);
    // Fallback to mock data on API error
    try {
      logger.info('Falling back to mock trends');
      return getMockTrends(limit);
    } catch (mockError) {
      logger.error('Mock data fallback also failed:', mockError);
      return [];
    }
  }
}

/**
 * Get mock trends data
 */
function getMockTrends(limit: number): ITrend[] {
  // Return mock data based on current trends
  const mockTrends = [
    { id: 'trend-1', category: 'Technology', title: '#AI', postCount: '125K' },
    { id: 'trend-2', category: 'Entertainment', title: '#Music', postCount: '98K' },
    { id: 'trend-3', category: 'News', title: '#BreakingNews', postCount: '87K' },
    { id: 'trend-4', category: 'Sports', title: '#Basketball', postCount: '76K' },
    { id: 'trend-5', category: 'Politics', title: '#Election', postCount: '65K' },
    { id: 'trend-6', category: 'Technology', title: '#Blockchain', postCount: '54K' },
    { id: 'trend-7', category: 'Entertainment', title: '#Movies', postCount: '43K' },
    { id: 'trend-8', category: 'Sports', title: '#Football', postCount: '32K' },
    { id: 'trend-9', category: 'Technology', title: '#Programming', postCount: '21K' },
    { id: 'trend-10', category: 'Lifestyle', title: '#Fitness', postCount: '10K' }
  ];
  
  return mockTrends.slice(0, limit);
}

/**
 * Get trends from the API
 */
async function getTrendsFromAPI(limit: number): Promise<ITrend[]> {
  const token = getAuthToken();
  
  logger.debug('Fetching trends from API', { limit });
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
    postCount: trend.post_count || 0
  }));
} 