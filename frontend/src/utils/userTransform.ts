/**
 * Utility functions for transforming user data from API responses
 */

import { logger } from './logger';

/**
 * Standard user interface used throughout the application
 */
export interface StandardUser {
  id: string;
  username: string;
  displayName: string;
  avatar: string;
  bio?: string;
  isVerified?: boolean;
  isFollowing?: boolean;
  followerCount?: number;
}

/**
 * Transforms a user object from API response to standardized format
 * Handles various field naming patterns that might exist in the API
 */
export function transformApiUser(user: any): StandardUser {
  if (!user || !user.id) {
    logger.warn('Invalid user object provided to transform', { user });
    throw new Error('Invalid user object');
  }

  return {
    id: user.id,
    username: user.username || '',
    displayName: user.display_name || user.name || user.username || 'User',
    avatar: user.avatar_url || user.profile_picture_url || user.avatar || '',
    bio: user.bio || '',
    isVerified: !!user.is_verified,
    isFollowing: !!user.is_following,
    followerCount: user.follower_count || 0
  };
}

/**
 * Transforms an array of users from API response to standardized format
 */
export function transformApiUsers(users: any[]): StandardUser[] {
  if (!Array.isArray(users)) {
    logger.warn('Invalid users array provided to transform', { users });
    return [];
  }
  
  return users.map(user => {
    try {
      return transformApiUser(user);
    } catch (error) {
      logger.warn('Error transforming user', { user, error });
      return null;
    }
  }).filter(Boolean) as StandardUser[];
} 