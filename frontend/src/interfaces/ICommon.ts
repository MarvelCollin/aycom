/**
 * Common interfaces used across the application
 */

/**
 * Standard pagination metadata - use this interface for all paginated responses
 */
export interface IPagination {
  total_count: number;
  current_page: number;
  per_page: number;
  total_pages: number;
  has_more?: boolean;
}

/**
 * Standard API success response wrapper
 */
export interface IApiResponse<T> {
  success: boolean;
  data: T;
  message?: string;
}

/**
 * Standard API error response
 */
export interface IErrorResponse {
  success: false;
  error: {
    code: string;
    message: string;
    fields?: Record<string, string>;
  };
}

/**
 * Generic API response type (success or error)
 */
export type ApiResponse<T> = IApiResponse<T> | IErrorResponse;