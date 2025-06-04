/**
 * Trend-related interfaces
 */

/**
 * Trend interface
 */
export interface ITrend {
  id: string;
  title?: string;     // Some APIs use title
  name?: string;      // Some APIs use name
  category?: string;
  post_count?: number; // Some APIs use post_count
  tweet_count?: number; // Some APIs use tweet_count
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