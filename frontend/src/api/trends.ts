import { getAuthToken } from '../utils/auth';
import type { ITrend } from '../interfaces/ISocialMedia';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8081/api/v1';

export async function getTrends(limit: number = 5): Promise<ITrend[]> {
  try {
    // Get token for authenticated requests
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
      throw new Error(`Failed to fetch trends: ${response.status}`);
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
    console.error('Error fetching trends:', error);
    throw error;
  }
} 