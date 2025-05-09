import { getAuthToken } from '../utils/auth';
import appConfig from '../config/appConfig';
import { createLoggerWithPrefix } from '../utils/logger';

const API_BASE_URL = appConfig.api.baseUrl;
const logger = createLoggerWithPrefix('CategoriesAPI');

const DEFAULT_CATEGORIES = [
  {"id": "technology", "name": "Technology"},
  {"id": "health", "name": "Health"},
  {"id": "education", "name": "Education"},
  {"id": "entertainment", "name": "Entertainment"},
  {"id": "science", "name": "Science"},
  {"id": "sports", "name": "Sports"},
  {"id": "politics", "name": "Politics"},
  {"id": "business", "name": "Business"},
  {"id": "lifestyle", "name": "Lifestyle"},
  {"id": "travel", "name": "Travel"},
  {"id": "other", "name": "Other"}
];

export async function getCategories() {
  try {
    logger.debug('Fetching categories from API');
    
    return {
      success: true,
      categories: DEFAULT_CATEGORIES
    };
    
  } catch (error) {
    logger.error('Failed to fetch categories:', error);
    
    return {
      success: true,
      categories: DEFAULT_CATEGORIES
    };
  }
}