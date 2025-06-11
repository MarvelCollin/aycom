/**
 * Standardize data fields to maintain consistent naming across the application
 * This helps bridge gaps between backend and frontend field naming conventions
 */

import type { ICommunity } from '../interfaces/ICommunity';
import type { IThread } from '../interfaces/IThread';
import type { IUser } from '../interfaces/IUser';
import type { IPagination } from '../interfaces/ICommon';
import { createLoggerWithPrefix } from './logger';

const logger = createLoggerWithPrefix('standardize-api-data');

/**
 * Standardize pagination fields to match the IPagination interface
 * @param pagination Pagination data from API response
 * @returns Standardized pagination object
 */
export function standardizePagination(pagination: any): IPagination {
  if (!pagination) return {
    total_count: 0,
    current_page: 1,
    per_page: 10,
    total_pages: 0,
    has_more: false
  };

  return {
    total_count: pagination.total_count || pagination.totalCount || pagination.total || 0,
    current_page: pagination.current_page || pagination.currentPage || pagination.page || 1,
    per_page: pagination.per_page || pagination.perPage || pagination.limit || 10,
    total_pages: pagination.total_pages || pagination.totalPages || 
      Math.ceil((pagination.total_count || pagination.totalCount || pagination.total || 0) / 
      (pagination.per_page || pagination.perPage || pagination.limit || 10)),
    has_more: pagination.has_more || pagination.hasMore || false
  };
}

/**
 * Standardize community data to match the ICommunity interface
 * @param community Community data from API response
 * @returns Standardized community object
 */
export function standardizeCommunity(community: any): ICommunity {
  if (!community) return null;

  return {
    id: community.id || community.community_id || '',
    name: community.name || '',
    description: community.description || '',
    logo_url: community.logo_url || community.logoUrl || community.logo || '',
    banner_url: community.banner_url || community.bannerUrl || community.banner || '',
    creator_id: community.creator_id || community.creatorId || '',
    is_approved: typeof community.is_approved !== 'undefined' ? community.is_approved : 
      (typeof community.isApproved !== 'undefined' ? community.isApproved : false),
    categories: community.categories || [],
    created_at: community.created_at || community.createdAt || new Date().toISOString(),
    member_count: community.member_count || community.memberCount || 0
  };
}

/**
 * Standardize thread data to match the IThread interface
 * @param thread Thread data from API response
 * @returns Standardized thread object
 */
export function standardizeThread(thread: any): IThread {
  if (!thread) return null;

  return {
    id: thread.id || thread.thread_id || '',
    content: thread.content || '',
    user_id: thread.user_id || thread.userId || '',
    created_at: thread.created_at || thread.createdAt || new Date().toISOString(),
    updated_at: thread.updated_at || thread.updatedAt || thread.created_at || thread.createdAt || new Date().toISOString(),
    likes_count: thread.likes_count || thread.likesCount || thread.like_count || 0,
    replies_count: thread.replies_count || thread.repliesCount || thread.reply_count || 0,
    reposts_count: thread.reposts_count || thread.repostsCount || thread.repost_count || 0,
    is_liked: typeof thread.is_liked !== 'undefined' ? thread.is_liked : 
      (typeof thread.isLiked !== 'undefined' ? thread.isLiked : false),
    is_reposted: typeof thread.is_reposted !== 'undefined' ? thread.is_reposted : 
      (typeof thread.isReposted !== 'undefined' ? thread.isReposted : false),
    is_bookmarked: typeof thread.is_bookmarked !== 'undefined' ? thread.is_bookmarked : 
      (typeof thread.isBookmarked !== 'undefined' ? thread.isBookmarked : false),
    media: thread.media || [],
    user: thread.user || null,
    community_id: thread.community_id || thread.communityId || null,
    is_pinned: typeof thread.is_pinned !== 'undefined' ? thread.is_pinned : 
      (typeof thread.isPinned !== 'undefined' ? thread.isPinned : false)
  };
}

/**
 * Standardize user data to match the IUser interface
 * @param user User data from API response
 * @returns Standardized user object
 */
export function standardizeUser(user: any): IUser {
  if (!user) return null;

  return {
    id: user.id || user.user_id || '',
    username: user.username || '',
    name: user.name || user.display_name || user.displayName || '',
    bio: user.bio || '',
    profile_picture_url: user.profile_picture_url || user.profilePictureUrl || user.profile_picture || user.profilePicture || '',
    banner_url: user.banner_url || user.bannerUrl || user.banner || '',
    is_verified: typeof user.is_verified !== 'undefined' ? user.is_verified : 
      (typeof user.isVerified !== 'undefined' ? user.isVerified : false),
    is_admin: typeof user.is_admin !== 'undefined' ? user.is_admin : 
      (typeof user.isAdmin !== 'undefined' ? user.isAdmin : false),
    follower_count: user.follower_count || user.followerCount || 0,
    following_count: user.following_count || user.followingCount || 0,
    created_at: user.created_at || user.createdAt || new Date().toISOString(),
    is_following: typeof user.is_following !== 'undefined' ? user.is_following : 
      (typeof user.isFollowing !== 'undefined' ? user.isFollowing : false)
  };
}

/**
 * Standardize arrays of data
 * @param items Array of items to standardize
 * @param standardizeFn The standardization function to apply to each item
 * @returns Array of standardized items
 */
export function standardizeArray<T>(items: any[], standardizeFn: (item: any) => T): T[] {
  if (!items || !Array.isArray(items)) return [];
  
  return items.map(item => standardizeFn(item));
}

/**
 * Standardize report request data
 * @param request Report request data from API
 * @returns Standardized report request
 */
export function standardizeReportRequest(request: any): any {
  if (!request) return null;
  
  return {
    id: request.id || '',
    reporter_id: request.reporter_id || request.reporterId || '',
    reported_user_id: request.reported_user_id || request.reportedUserId || request.reported_id || '',
    reason: request.reason || '',
    status: request.status || 'pending',
    created_at: request.created_at || request.createdAt || new Date().toISOString(),
    updated_at: request.updated_at || request.updatedAt || request.created_at || request.createdAt || new Date().toISOString(),
    reporter: request.reporter ? standardizeUser(request.reporter) : null,
    reported_user: request.reported_user || request.reportedUser ? standardizeUser(request.reported_user || request.reportedUser) : null
  };
}

/**
 * Standardize community request data
 * @param request Community request data from API
 * @returns Standardized community request
 */
export function standardizeCommunityRequest(request: any): any {
  if (!request) return null;
  
  return {
    id: request.id || '',
    user_id: request.user_id || request.userId || '',
    name: request.name || request.community_name || '',
    description: request.description || '',
    status: request.status || 'pending',
    logo_url: request.logo_url || request.logoUrl || '',
    banner_url: request.banner_url || request.bannerUrl || '',
    category_id: request.category_id || request.categoryId || null,
    created_at: request.created_at || request.createdAt || new Date().toISOString(),
    updated_at: request.updated_at || request.updatedAt || request.created_at || request.createdAt || new Date().toISOString(),
    requester: request.requester ? standardizeUser(request.requester) : null
  };
}

/**
 * Standardize premium request data
 * @param request Premium request data from API
 * @returns Standardized premium request
 */
export function standardizePremiumRequest(request: any): any {
  if (!request) return null;
  
  return {
    id: request.id || '',
    user_id: request.user_id || request.userId || '',
    payment_id: request.payment_id || request.paymentId || '',
    amount: request.amount || 0,
    status: request.status || 'pending',
    created_at: request.created_at || request.createdAt || new Date().toISOString(),
    updated_at: request.updated_at || request.updatedAt || request.created_at || request.createdAt || new Date().toISOString(),
    requester: request.requester ? standardizeUser(request.requester) : null
  };
} 