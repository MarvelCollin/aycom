// Social media related interfaces
import type { IUserProfile } from './IUser';

// Media interface
export interface IMedia {
  id: string;
  url: string;
  type: string;
  thumbnail?: string;
  alt?: string;
}

// Thread/Tweet interface
export interface ITweet {
  // Core fields
  id: string;
  content: string;
  created_at: string;        // ISO timestamp
  updated_at?: string;
  
  // User-related fields
  user_id: string;
  username: string;
  name: string;             
  profile_picture_url: string;
  
  // Interaction metrics
  likes_count: number;       
  replies_count: number;     
  reposts_count: number;     
  bookmark_count: number;    
  views_count?: number;       
  
  // Media
  media?: IMedia[];
  
  // Interaction states
  is_liked: boolean;         
  is_reposted: boolean;      
  is_bookmarked: boolean;    
  is_pinned: boolean;        
  
  // Relations
  parent_id?: string | null; 
  thread_id?: string;
  reply_to?: ITweet | null;
  
  // Community-related fields
  community_id?: string | null;
  community_name?: string | null;
  
  // Additional metadata
  is_advertisement?: boolean;
}

// Trend interface
export interface ITrend {
  id?: string;
  category: string;
  title: string;
  post_count: number;
}

// Suggested follow interface
export interface ISuggestedFollow {
  user_id: string;
  username: string;
  name: string;
  profile_picture_url: string | null;
  is_verified: boolean;
  follower_count: number;
  is_following?: boolean;
}

// Community interface
export interface ICommunity {
  id: string;
  name: string;
  member_count: number;
  logo_url: string;
}

// Follower/Following interface
export interface IFollowUser {
  id: string;
  username: string;
  name: string;
  display_name?: string;
  profile_picture_url: string;
  is_following: boolean;
  bio: string;
}