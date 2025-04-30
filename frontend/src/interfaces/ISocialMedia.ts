export interface ITweet {
  id: number;
  username: string;
  displayName: string;
  avatar: string;
  content: string;
  timestamp: string;
  likes: number;
  replies: number;
  reposts: number;
  views: string;
}

export interface ITrend {
  category: string;
  title: string;
  postCount: string;
}

export interface ISuggestedFollow {
  username: string;
  displayName: string;
  avatar: string;
  verified: boolean;
  followerCount: number;
}

export interface ICommunity {
  id: string;
  name: string;
  description?: string;
  memberCount?: number;
  avatar?: string;
} 