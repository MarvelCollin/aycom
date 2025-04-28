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
  bookmarked?: boolean;
  liked?: boolean;
  reposted?: boolean;
}

// User profile interface
export interface IUserProfile {
  name: string;
  username: string;
  bio?: string;
  location?: string;
  website?: string;
  joined_date: string;
  following_count: number;
  followers_count: number;
  profile_picture?: string[];
}

// Trend interface
export interface ITrend {
  category: string;
  title: string;
  postCount: string;
}

// Suggested follow interface
export interface ISuggestedFollow {
  displayName: string;
  username: string;
  avatar: string;
  verified: boolean;
  followerCount: number;
  isFollowing?: boolean;
}

// Navigation item interface
export interface INavigationItem {
  label: string;
  icon: string; // This will contain SVG component reference as a string
  path: string;
} 