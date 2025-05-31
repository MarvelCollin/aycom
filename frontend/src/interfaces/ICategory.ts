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