// Extended Tweet interfaces to be used with TweetCard and other components
import type { ITweet } from './ISocialMedia';

// Extended tweet interface for additional properties that might be in the data
export interface ExtendedTweet extends ITweet {
  // Fields for retweets and bookmarks
  retweet_id?: string;
  threadId?: string;
  thread_id?: string;
  tweetId?: string;
  userId?: string;
  authorId?: string;
  // Additional fields that might be present
  display_name?: string;
  avatar?: string;
  [key: string]: any;
}

// Define the IMedia interface here since it's needed but not exported from ISocialMedia
interface IMedia {
  id: string;
  url: string;
  type: string;
  thumbnail?: string;
  alt_text?: string;
}

// Helper functions for compatibility
export function ensureTweetFormat(thread: any): ExtendedTweet {
  try {
    if (!thread || typeof thread !== 'object') {
      return createEmptyTweet();
    }
    
    console.log("Raw thread data in ensureTweetFormat:", thread);
  
    // More robust username extraction with better fallbacks
    let username = thread.username;
    
    // Handle deeply nested author structures
    if (!username && thread.author) {
      username = thread.author.username || thread.author_username;
    }
    
    // Handle other possible locations
    if (!username) {
      username = thread.author_username || 
                thread.authorUsername || 
                thread.user?.username || 
                thread.author?.username || 
                thread.user_data?.username;
    }
    
    // Final fallback - use a unique placeholder
    if (!username || username === 'anonymous') {
      username = `user_${thread.user_id || thread.userId || thread.authorId || thread.id || Math.random().toString(36).substring(2, 9)}`;
    }
    
    // Display name with similar logic
    const name = thread.name || 
               thread.author_name || 
               thread.authorName || 
               thread.display_name || 
               thread.displayName ||
               thread.author?.name || 
               thread.author?.display_name ||
               thread.user?.name ||
               thread.user_data?.name ||
               username;
    
    // Profile picture with similar robust fallbacks
    const profile_picture_url = thread.profile_picture_url || 
                        thread.profilePictureUrl ||
                        thread.author_avatar || 
                        thread.authorAvatar || 
                        thread.avatar || 
                        thread.author?.profile_picture_url ||
                        thread.user?.profile_picture_url ||
                        thread.user_data?.profile_picture_url ||
                        'https://secure.gravatar.com/avatar/0?d=mp';
    
    let created_at = thread.created_at || thread.createdAt || thread.timestamp || new Date().toISOString();
    if (typeof created_at === 'string' && !created_at.includes('T')) {
      created_at = new Date(created_at).toISOString();
    }
    
    const likes_count = Number(thread.likes_count || thread.like_count || thread.metrics?.likes || 0);
    const replies_count = Number(thread.replies_count || thread.reply_count || thread.metrics?.replies || 0);
    const reposts_count = Number(thread.reposts_count || thread.repost_count || thread.metrics?.reposts || 0);
    const bookmark_count = Number(thread.bookmarks_count || thread.bookmark_count || thread.metrics?.bookmarks || 0);
    const views_count = Number(thread.views || thread.views_count || 0);
    
    const is_liked = Boolean(thread.is_liked || thread.isLiked || false);
    const is_reposted = Boolean(thread.is_repost || thread.isReposted || false);
    const is_bookmarked = Boolean(thread.is_bookmarked || thread.isBookmarked || false);
    const is_pinned = Boolean(
      thread.is_pinned === true || 
      thread.is_pinned === 'true' || 
      thread.is_pinned === 1 || 
      thread.is_pinned === '1' || 
      thread.is_pinned === 't' || 
      thread.IsPinned === true || 
      false
    );
    
    // Handle verification status
    const is_verified = Boolean(
      thread.is_verified || 
      thread.verified || 
      thread.user?.is_verified || 
      thread.author?.is_verified || 
      false
    );
    
    const media = Array.isArray(thread.media) ? thread.media : [];
      
    const id = thread.id || `thread-${Math.random().toString(36).substring(2, 9)}`;
    const user_id = thread.user_id || thread.userId || thread.author_id || thread.authorId || thread.author?.id || thread.user?.id || '';
        
    return {
      id,
      content: thread.content || '',
      created_at: typeof created_at === 'string' ? created_at : new Date(created_at).toISOString(),
      updated_at: thread.updated_at,
      
      // User info with consistent values
      user_id,
      username,
      name,
      profile_picture_url,
      
      // Metrics with consistent values
      likes_count,
      replies_count,
      reposts_count,
      bookmark_count,
      views_count,
      
      // Status flags
      is_liked,
      is_reposted,
      is_bookmarked,
      is_pinned,
      is_verified,
      
      // Relations
      parent_id: thread.parent_id || null,
      
      // Media
      media,
      
      // Additional compatibility fields
      thread_id: thread.thread_id || thread.threadId || id,
      
      // Pass through any other properties
      ...thread
    };
  } catch (error) {
    console.error('Error formatting tweet:', error);
    return createEmptyTweet();
  }
}

function createEmptyTweet(): ExtendedTweet {
  const id = `error-${Math.random().toString(36).substring(2, 9)}`;
  return {
    id,
    content: 'Error loading tweet',
    created_at: new Date().toISOString(),
    updated_at: undefined,
    
    // User info
    user_id: '',
    username: 'error',
    name: 'Error',
    profile_picture_url: '',
    
    // Metrics
    likes_count: 0,
    replies_count: 0,
    reposts_count: 0,
    bookmark_count: 0,
    views_count: 0,
    
    // Status flags
    is_liked: false,
    is_reposted: false,
    is_bookmarked: false,
    is_pinned: false,
    is_verified: false,
    
    // Relations
    parent_id: null,
    
    // Media
    media: [],
    
    // Additional compatibility fields
    thread_id: id
  };
}
