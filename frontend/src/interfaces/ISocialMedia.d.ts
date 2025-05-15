// Social media related interfaces
import type { IUserProfile } from './IUser';

// Media interface
export interface IMedia {
  url: string;
  type: string;
  media_id?: string;
  alt?: string;
}

// Tweet/Post interface
export interface ITweet {
  // Core fields
  id: number | string;
  threadId?: string;
  thread_id?: string;          // Snake case variant
  tweetId?: string;             // String-only alias for id
  content: string;
  timestamp: string;            // ISO timestamp from backend
  
  // User-related fields
  username: string;
  displayName: string;
  avatar: string | null;
  
  // User data sometimes included in different formats
  user_data?: {
    id?: string;
    username?: string;
    name?: string;
    email?: string;
    profile_picture_url?: string;
  };
  
  // Interaction metrics
  likes: number;
  replies: number;
  reposts: number;
  bookmarks: number;
  views: string;                // Some views come as formatted strings
  
  // Media
  media?: IMedia[];
  
  // Interaction states
  isLiked?: boolean;
  is_liked?: boolean;           // Backend sometimes uses snake_case
  isReposted?: boolean;
  isBookmarked?: boolean;
  is_pinned?: boolean;
  
  // Relations
  replyTo?: ITweet | null;
  
  // Author objects from different API responses
  thread?: {
    id?: string;
    user_id?: string;
    author?: {
      id?: string;
      username?: string;
      name?: string;
      profile_picture_url?: string;
    }
  };
  
  threadInfo?: any;
  
  // Complete user object sometimes included
  user?: {
    id?: string;
    username?: string;
    name?: string;
    display_name?: string;
    profile_picture_url?: string;
    email?: string;
  };
  
  author?: {
    id?: string;
    username?: string;
    name?: string;
    profile_picture_url?: string;
  };
  
  // Metadata
  isAdvertisement?: boolean;
  communityId?: string | null;
  communityName?: string | null;
  
  // Backend field mappings (all string-typed to avoid conversion issues)
  userId?: string;               // Standard user ID field
  user_id?: string;              // Alternate user ID field
  authorId?: string;             // Legacy user ID field
  author_id?: string;            // Snake case variant
  authorName?: string;           // Legacy display name field
  author_name?: string;          // Snake case variant
  authorUsername?: string;       // Legacy username field 
  author_username?: string;      // Snake case variant
  authorAvatar?: string | null;  // Legacy avatar field
  author_avatar?: string | null; // Snake case variant
  name?: string;                 // Backend user name field
  display_name?: string;         // Backend display name field
  profile_picture_url?: string;  // Backend avatar URL field
  created_at?: string;           // ISO timestamp sometimes used
  
  // Other properties
  verified?: boolean;
  likesCount?: number;
  commentsCount?: number;
  repostsCount?: number;
  viewsCount?: string;
  poll?: any | null;
  quoteTweet?: any | null;
  replyingTo?: any | null;
}

// Trend interface
export interface ITrend {
  id?: string;
  category: string;
  title: string;
  postCount: string | number;
}

// Suggested follow interface
export interface ISuggestedFollow {
  username: string;
  displayName: string;
  avatar: string;
  verified: boolean;
  followerCount: number;
  isFollowing?: boolean;
}

// Community interface
export interface ICommunity {
  id: string;
  name: string;
  memberCount: number;
  avatar: string;
} 