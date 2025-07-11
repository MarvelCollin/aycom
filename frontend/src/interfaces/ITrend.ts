export interface ITrend {
  id: string;
  title?: string;     
  name?: string;      
  category?: string;
  post_count?: number; 
  tweet_count?: number; 
  query?: string;
}

export interface ITrendsResponse {
  success: boolean;
  data: {
    trends: ITrend[];
  };
}