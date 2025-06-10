import appConfig from '../config/appConfig';
import { createLoggerWithPrefix } from '../utils/logger';
import { getAuthToken } from '../utils/auth';
import type { ICategory } from '../interfaces/ICategory';

const API_BASE_URL = appConfig.api.baseUrl;
const logger = createLoggerWithPrefix('CategoriesAPI');

export async function getThreadCategories(): Promise<ICategory[]> {
  try {
    const token = getAuthToken();
    const response = await fetch(`${API_BASE_URL}/categories`, {
      headers: {
        'Authorization': `Bearer ${token}`,
      },
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      const errorMessage = errorData.message || `Error ${response.status}: ${response.statusText}`;
      logger.error(`Failed to fetch categories: ${errorMessage}`);
      throw new Error(errorMessage);
    }

    const data = await response.json();

    if (!data || !data.categories || !Array.isArray(data.categories)) {
      logger.warn('API returned invalid categories data format');
      return [];
    }

    logger.info('Successfully fetched categories from API', { count: data.categories.length });
    return data.categories;
  } catch (error) {
    logger.error('Failed to fetch categories:', error);
    return [];
  }
}

export const getCategories = getThreadCategories;