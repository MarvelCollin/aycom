import type { IMedia } from './IMedia';
import type { ITrend } from './ITrend';
import type { IPagination } from './ICommon';

export interface ITweet {
  id: string;
  content: string;
  created_at: string;
  updated_at?: string;
  
  // User info
  user_id: string;
  username: string;
  name: string;
  profile_picture_url: string;
  
  // Metrics
  likes_count: number;
  replies_count: number;
  reposts_count: number;
  bookmark_count: number;
  views_count?: number;
  
  // Status flags
  is_liked: boolean;
  is_reposted: boolean;
  is_bookmarked: boolean;
  is_pinned: boolean;
  
  // Relations
  parent_id: string | null;
  
  // Media
  media: IMedia[];
  
  // Community
  community_id?: string | null;
  community_name?: string | null;
  
  // Other
  is_advertisement?: boolean;
}

export interface ISuggestedFollow {
  id: string;
  username: string;
  name: string;
  profile_picture_url: string;
  bio?: string;
  is_verified?: boolean;
  is_following?: boolean;
}

// Backend Thread API response types
export interface IThreadApiResponse {
  id: string;
  content: string;
  created_at: string;
  updated_at?: string;
  user_id: string;
  username: string;
  name: string;
  profile_picture_url: string;
  likes_count: number;
  replies_count: number;
  reposts_count: number;
  bookmark_count: number;
  is_liked: boolean;
  is_reposted: boolean;
  is_bookmarked: boolean;
  is_pinned: boolean;
  parent_id: string | null;
  media: IMedia[];
}

// Standard thread interaction interface for consistent naming
export interface IThreadInteraction {
  likes_count: number;
  is_liked: boolean;
  replies_count: number;
  reposts_count: number;
  is_reposted: boolean;
  bookmark_count: number; 
  is_bookmarked: boolean;
}

/**
 * Thread API response interfaces
 */
export interface IThreadResponse {
  success: boolean;
  data: ITweet;
}

export interface IThreadsResponse {
  success: boolean;
  data: {
    threads: ITweet[];
    pagination: IPagination;
  };
}

export interface ICreateThreadRequest {
  content: string;
  media?: Array<string>;
  scheduled_at?: string;
  community_id?: string;
  who_can_reply?: 'everyone' | 'following' | 'mentioned';
}

export interface ICreateReplyRequest {
  content: string;
  media?: Array<string>;
  parent_id?: string;
}

export interface ILikeResponse {
  success: boolean;
  data: {
    message: string;
  };
}

export interface IRepostResponse {
  success: boolean;
  data: {
    message: string;
  };
}

export interface IBookmarkResponse {
  success: boolean;
  data: {
    message: string;
  };
}

export interface IThreadMediaUpdateRequest {
  media_urls: string[];
}

export interface IPinResponse {
  success: boolean;
  data: {
    message: string;
  };
} 