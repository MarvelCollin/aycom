import { getAuthToken } from '../utils/auth';
import appConfig from '../config/appConfig';
import type { ITrend } from '../interfaces/ISocialMedia';
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
      
      // If 401 unauthorized for a logged-in feature, return empty array
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

function getMockTrends(limit: number): ITrend[] {
  const mockTrends = [
    { id: 'trend-1', category: 'Technology', title: '#AI', post_count: 125000 },
    { id: 'trend-2', category: 'Entertainment', title: '#Music', post_count: 98000 },
    { id: 'trend-3', category: 'News', title: '#BreakingNews', post_count: 87000 },
    { id: 'trend-4', category: 'Sports', title: '#Basketball', post_count: 76000 },
    { id: 'trend-5', category: 'Politics', title: '#Election', post_count: 65000 },
    { id: 'trend-6', category: 'Technology', title: '#Blockchain', post_count: 54000 },
    { id: 'trend-7', category: 'Entertainment', title: '#Movies', post_count: 43000 },
    { id: 'trend-8', category: 'Sports', title: '#Football', post_count: 32000 },
    { id: 'trend-9', category: 'Technology', title: '#Programming', post_count: 21000 },
    { id: 'trend-10', category: 'Lifestyle', title: '#Fitness', post_count: 10000 }
  ];
  
  return mockTrends.slice(0, limit);
}

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
    post_count: trend.post_count || 0
  }));
}