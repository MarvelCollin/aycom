// Social media related interfaces
import type { IUserProfile } from './IUser';

// Media interface
export interface IMedia {
  url: string;
  type: string;
  id?: string;
  alt?: string;
}

// Tweet/Post interface
export interface ITweet {
  // Core fields
  id: string;
  threadId: string;
  content: string;
  timestamp: string;            // ISO timestamp
  createdAt?: string;           // Alternative timestamp field
  
  // User-related fields
  userId: string;
  username: string;
  displayName: string;
  avatar: string;
  
  // Interaction metrics
  likes: number;
  replies: number;
  reposts: number;
  bookmarks: number;
  views: number;
  
  // Media
  media?: IMedia[];
  
  // Interaction states
  isLiked: boolean;
  isReposted: boolean;
  isBookmarked: boolean;
  isPinned: boolean;
  
  // Relations
  replyTo?: ITweet | null;
  
  // Community-related fields
  communityId?: string | null;
  communityName?: string | null;
  
  // Additional metadata
  isAdvertisement?: boolean;
  
  // Legacy fields for backward compatibility (optional)
  _originalData?: any;
}

// Trend interface
export interface ITrend {
  id?: string;
  category: string;
  title: string;
  postCount: number;
}

// Suggested follow interface
export interface ISuggestedFollow {
  userId: string; 
  username: string;
  displayName: string;
  avatar: string | null;
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