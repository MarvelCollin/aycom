import { logger } from './logger';

export interface StandardUser {
  id: string;
  username: string;
  displayName: string;
  avatar: string | null;
  bio?: string;
  isVerified: boolean;
  isFollowing?: boolean;
  followerCount?: number;
}

export function transformApiUser(user: any): StandardUser {
  if (!user || !user.id) {
    logger.warn('Invalid user object provided to transform', { user });
    throw new Error('Invalid user object');
  }

  return {
    id: user.id,
    username: user.username || '',
    displayName: user.display_name || user.name || user.username || 'User',
    avatar: user.avatar_url || user.profile_picture_url || user.avatar || null,
    bio: user.bio || '',
    isVerified: !!user.is_verified,
    isFollowing: !!user.is_following,
    followerCount: user.follower_count || 0
  };
}

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