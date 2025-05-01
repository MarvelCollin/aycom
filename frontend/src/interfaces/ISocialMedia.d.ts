// Social media related interfaces
import type { IUserProfile } from './IUser';

// Media interface
export interface IMedia {
  url: string;
  type: string;
  media_id?: string;
}

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
  media?: IMedia[];
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
  verified: boolean;
  followerCount: number;
}

// Community interface
export interface ICommunity {
  id: string;
  name: string;
  memberCount: number;
  avatar: string;
} 