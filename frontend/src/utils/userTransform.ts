import { logger } from "./logger";

export interface StandardUser {
  id: string;
  username: string;
  name: string;
  profile_picture_url: string | null;
  bio?: string;
  is_verified: boolean;
  is_following?: boolean;
  follower_count?: number;

  avatar?: string | null;
  displayName?: string;
  display_name?: string;
  role?: string;
}

export function transformApiUser(user: any): StandardUser {
  if (!user || !user.id) {
    logger.warn("Invalid user object provided to transform", { user });
    throw new Error("Invalid user object");
  }

  return {
    id: user.id,
    username: user.username || "",
    name: user.name || user.display_name || user.username || "User",
    profile_picture_url: user.profile_picture_url || user.avatar || null,
    bio: user.bio || "",
    is_verified: !!user.is_verified,
    is_following: !!user.is_following,
    follower_count: user.follower_count || 0,

    avatar: user.avatar || user.profile_picture_url || null,
    displayName: user.display_name || user.name || user.username || "User",
    display_name: user.display_name || user.name || user.username || "User",
    role: user.role || "user"
  };
}

export function transformApiUsers(users: any[]): StandardUser[] {
  if (!Array.isArray(users)) {
    logger.warn("Invalid users array provided to transform", { users });
    return [];
  }

  return users.map(user => {
    try {
      return transformApiUser(user);
    } catch (error) {
      logger.warn("Error transforming user", { user, error });
      return null;
    }
  }).filter(Boolean) as StandardUser[];
}