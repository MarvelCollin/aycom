// Social media related interfaces

// Tweet/Post interface
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

// User profile interface
export interface IUserProfile {
  id: string;
  username: string;
  displayName: string;
  avatar: string;
  bio?: string;
  followers: number;
  following: number;
  joined: string;
  location?: string;
  website?: string;
  banner?: string;
}

// Trend interface
export interface ITrend {
  category: string;
  title: string;
  postCount: string;
}

// Suggested follow interface
export interface ISuggestedFollow {
  username: string;
  displayName: string;
  avatar: string;
  verified?: boolean;
  followerCount?: number;
  isFollowing?: boolean;
}

// Navigation item interface
export interface INavigationItem {
  label: string;
  icon: string;
  path: string;
} 