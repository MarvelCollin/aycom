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
  id: number | string;
  threadId?: string;
  username: string;
  displayName: string;
  avatar: string | null;
  content: string;
  timestamp: string;
  likes: number;
  replies: number;
  reposts: number;
  bookmarks: number;
  views: string;
  media?: IMedia[];
  isLiked?: boolean;
  is_liked?: boolean;
  isReposted?: boolean;
  isBookmarked?: boolean;
  is_pinned?: boolean;
  replyTo?: ITweet | null;
  isAdvertisement?: boolean;
  communityId?: string | null;
  communityName?: string | null;
  
  createdAt?: string;
  authorId?: string;
  authorName?: string;
  authorUsername?: string;
  authorAvatar?: string | null;
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