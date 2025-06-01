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

/**
 * Category response from the API
 */
export interface ICategoriesResponse {
  success: boolean;
  data: {
    categories: ICategory[];
  };
}

/**
 * Category creation/update request
 */
export interface ICategoryRequest {
  name: string;
  description?: string;
}

/**
 * Single category response
 */
export interface ICategoryResponse {
  success: boolean;
  category: ICategory;
} 