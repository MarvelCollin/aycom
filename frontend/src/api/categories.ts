import { getAuthToken } from '../utils/auth';
import appConfig from '../config/appConfig';
import { createLoggerWithPrefix } from '../utils/logger';

const API_BASE_URL = appConfig.api.baseUrl;
const logger = createLoggerWithPrefix('CategoriesAPI');

/**
 * Fetch available categories from the API
 * @returns Array of category objects
 */
export async function getCategories() {
  try {
    logger.debug('Fetching categories from API');
    const response = await fetch(`${API_BASE_URL}/categories`, {
      method: "GET",
      headers: { 
        "Content-Type": "application/json",
        "Authorization": `Bearer ${getAuthToken()}`
      },
      credentials: "include",
    });
    
    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      const errorMessage = errorData.message || `Error ${response.status}: ${response.statusText}`;
      logger.error(`Failed to fetch categories: ${errorMessage}`);
      throw new Error(errorMessage);
    }
    
    const data = await response.json();
    logger.debug('Categories fetched successfully', { count: data.categories?.length || 0 });
    
    return {
      success: true,
      categories: data.categories || []
    };
    
  } catch (error) {
    logger.error('Failed to fetch categories:', error);
    throw error;
  }
} 