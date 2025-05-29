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
  name: string;              // Was displayName
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
  is_liked: boolean;         // Was isLiked
  is_reposted: boolean;      // Was isReposted
  is_bookmarked: boolean;    // Was isBookmarked
  is_pinned: boolean;        // Was isPinned
  
  // Relations
  reply_to?: ITweet | null;  // Was replyTo
  
  // Community-related fields
  community_id?: string | null; // Was communityId
  community_name?: string | null; // Was communityName
  
  // Additional metadata
  is_advertisement?: boolean; // Was isAdvertisement
  
  // Legacy fields for backward compatibility (optional)
  _original_data?: any;      // Was _originalData
}

// Trend interface
export interface ITrend {
  id?: string;
  category: string;
  title: string;
  post_count: number;       // Was postCount
}

// Suggested follow interface
export interface ISuggestedFollow {
  user_id: string;          // Was userId
  username: string;
  name: string;             // Was displayName
  profile_picture_url: string | null; // Was avatar
  is_verified: boolean;     // Was verified
  follower_count: number;   // Was followerCount
  is_following?: boolean;   // Was isFollowing
}

// Community interface
export interface ICommunity {
  id: string;
  name: string;
  member_count: number;     // Was memberCount
  logo_url: string;         // Was avatar
}