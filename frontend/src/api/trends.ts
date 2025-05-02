import { getAuthToken } from '../utils/auth';
import appConfig from '../config/appConfig';
import type { ITrend } from '../interfaces/ISocialMedia';

const API_BASE_URL = appConfig.api.baseUrl;

export async function getTrends(limit: number = 5): Promise<ITrend[]> {
  try {
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
      try {
        const errorData = await response.json();
        throw new Error(errorData.message || `Failed to fetch trends: ${response.status}`);
      } catch (parseError) {
        throw new Error(`Failed to fetch trends: ${response.status}`);
      }
    }
    
    const data = await response.json();
    
    if (data && data.trends) {
      return data.trends.map((trend: any) => ({
        category: trend.category || 'Trending',
        title: trend.title,
        postCount: trend.post_count ? `${trend.post_count}` : '0'
      }));
    }
    
    return [];
  } catch (error) {
    throw error;
  }
} 