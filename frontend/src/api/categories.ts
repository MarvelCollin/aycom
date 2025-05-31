import appConfig from '../config/appConfig';
import { createLoggerWithPrefix } from '../utils/logger';
import { getAuthToken } from '../utils/auth';

/**
 * Interface representing a category
 */
export interface ICategory {
  id: string;
  name: string;
  description?: string;
  slug?: string;
  thread_count?: number;
  icon?: string;
  color?: string;
  created_at?: string;
  updated_at?: string;
}

const API_BASE_URL = appConfig.api.baseUrl;
const logger = createLoggerWithPrefix('CategoriesAPI');

// Default categories for fallback
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

/**
 * Get thread categories from the API
 * 
 * NOTE: For community categories, use getCommunityCategories from community.ts
 * 
 * @returns Promise with categories
 */
export async function getThreadCategories(): Promise<ICategory[]> {
  try {
    const token = getAuthToken();
    const response = await fetch(`${API_BASE_URL}/categories`, {
      headers: {
        'Authorization': `Bearer ${token}`,
      },
    });

    if (!response.ok) {
      throw new Error(`Error fetching categories: ${response.statusText}`);
    }

    const data = await response.json();
    return data.categories || [];
  } catch (error) {
    console.error('Failed to fetch categories:', error);
    return [];
  }
}

// Maintain backwards compatibility
export const getCategories = getThreadCategories; 