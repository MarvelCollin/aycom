/**
 * AI-related interfaces
 */

/**
 * Predict category request
 */
export interface IPredictCategoryRequest {
  content: string;
}

/**
 * Predict category response
 */
export interface IPredictCategoryResponse {
  success: boolean;
  data: {
    category: string;
    confidence: number;
  };
} 