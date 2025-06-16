import { getAuthToken } from '../utils/auth';
import appConfig from '../config/appConfig';
import type { ISearchUsersResponse } from '../interfaces/ISearch';

const API_BASE_URL = appConfig.api.baseUrl;

export const searchUsers = async (query: string, filter: string = 'all', page: number = 1, limit: number = 10): Promise<ISearchUsersResponse> => {
  try {
    // For search terms that might be typos (short terms with 4-6 chars), add a second query parameter
    // This is a workaround for the backend fuzzy search limitation
    let searchUrl = '';
    
    if (query && query.length >= 4 && query.length <= 6) {
      // For short queries that might be typos, also include a prefix search
      // e.g. if searching for "kolnb", also look for "kolin" 
      const prefixQuery = query.substring(0, query.length - 1);
      searchUrl = `/search/users?q=${encodeURIComponent(query)}&alt_q=${encodeURIComponent(prefixQuery)}&filter=${filter}&page=${page}&limit=${limit}`;
      console.debug(`Using expanded search with alt_q=${prefixQuery} to catch potential typos`);
    } else {
      // Standard search
      searchUrl = `/search/users?q=${encodeURIComponent(query)}&filter=${filter}&page=${page}&limit=${limit}`;
    }

    const response = await fetch(`${API_BASE_URL}${searchUrl}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': getAuthToken() ? `Bearer ${getAuthToken()}` : ''
      },
      credentials: 'include'
    });
    
    if (!response.ok) {
      throw new Error(`Search request failed with status ${response.status}`);
    }
    
    return await response.json();
  } catch (error) {
    console.error('Error searching users:', error);
    throw error;
  }
}; 