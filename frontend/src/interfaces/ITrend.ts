/**
 * Trend-related interfaces
 */

/**
 * Trend interface
 */
export interface ITrend {
  id: string;
  title: string;
  category: string;
  post_count: number;
  query?: string;
}

/**
 * Trends response
 */
export interface ITrendsResponse {
  success: boolean;
  data: {
    trends: ITrend[];
  };
} 