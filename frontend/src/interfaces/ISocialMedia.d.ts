// Social media related interfaces
import type { IUserProfile } from './IUser';

// Media interface
export interface IMedia {
  id?: string;
  url: string;
  type: string;
  alt?: string;
}

// Tweet/Post interface
export interface ITweet {
  // Core fields
  id: string;
  thread_id: string;
  content: string;
  created_at: string;        // ISO timestamp
  
  // User-related fields
  user_id: string;
  username: string;
  name: string;              // Was display_name
  profile_picture_url: string;  // Was avatar
  
  // Interaction metrics
  likes_count: number;       // Was likes
  replies_count: number;     // Was replies
  reposts_count: number;     // Was reposts
  bookmarks_count: number;   // Was bookmarks
  views_count: number;       // Was views
  
  // Media
  media?: IMedia[];
  
  // Interaction states
  is_liked: boolean;         // Was is_liked
  is_reposted: boolean;      // Was is_reposted
  is_bookmarked: boolean;    // Was is_bookmarked
  is_pinned: boolean;        // Was is_pinned
  
  // Relations
  reply_to?: ITweet | null;  // Was reply_to
  
  // Community-related fields
  community_id?: string | null; // Was community_id
  community_name?: string | null; // Was community_name
  
  // Additional metadata
  is_advertisement?: boolean; // Was is_advertisement
  
  // Legacy fields for backward compatibility (optional)
  _original_data?: any;      // Was _original_data
}

// Trend interface
export interface ITrend {
  id?: string;
  category: string;
  title: string;
  post_count: number;       // Was post_count
}

// Suggested follow interface
export interface ISuggestedFollow {
  user_id: string;          // Was user_id
  username: string;
  name: string;             // Was display_name
  profile_picture_url: string | null; // Was avatar
  is_verified: boolean;     // Was verified
  follower_count: number;   // Was follower_count
  is_following?: boolean;   // Was is_following
}

// Community interface
export interface ICommunity {
  id: string;
  name: string;
  member_count: number;     // Was member_count
  logo_url: string;         // Was avatar
}