import { writable } from 'svelte/store';
import type { ITweet } from '../interfaces/ISocialMedia';

// Define the interaction state type
export interface TweetInteractionState {
  is_liked: boolean;
  is_reposted: boolean;
  is_bookmarked: boolean;
  likes: number;
  reposts: number;
  replies: number;
  bookmarks: number;
  pending_like?: boolean;
  pending_repost?: boolean;
  pending_bookmark?: boolean;
  pending_reply?: boolean;
}

// Store to keep track of tweet interaction states
const createTweetInteractionStore = () => {
  const interactions = new Map<string, TweetInteractionState>();
  const { subscribe, update, set } = writable(interactions);

  return {
    subscribe,
    
    // Initialize a tweet in the store
    initTweet: (tweet: ITweet) => {
      update(map => {
        if (!map.has(tweet.id)) {
          map.set(tweet.id, {
            is_liked: tweet.is_liked || false,
            is_reposted: tweet.is_reposted || false,
            is_bookmarked: tweet.is_bookmarked || false,
            likes: tweet.likes_count || 0,
            reposts: tweet.reposts_count || 0,
            replies: tweet.replies_count || 0,
            bookmarks: tweet.bookmark_count || 0,
            pending_like: false,
            pending_repost: false,
            pending_bookmark: false,
            pending_reply: false
          });
        }
        return map;
      });
    },
    
    // Update a tweet's interaction state
    updateTweetInteraction: (tweetId: string, changes: Partial<TweetInteractionState>) => {
      update(map => {
        if (map.has(tweetId)) {
          const current = map.get(tweetId);
          map.set(tweetId, { ...current, ...changes });
        } else {
          map.set(tweetId, {
            is_liked: changes.is_liked === true,
            is_reposted: changes.is_reposted === true,
            is_bookmarked: changes.is_bookmarked === true,
            likes: changes.likes || 0,
            reposts: changes.reposts || 0,
            replies: changes.replies || 0,
            bookmarks: changes.bookmarks || 0,
            pending_like: changes.pending_like === true,
            pending_repost: changes.pending_repost === true,
            pending_bookmark: changes.pending_bookmark === true,
            pending_reply: changes.pending_reply === true
          });
        }
        return map;
      });
    },
    
    // Get a tweet's interaction state
    getInteractionStatus: (tweetId: string): TweetInteractionState | undefined => {
      let result: TweetInteractionState | undefined;
      update(map => {
        result = map.get(tweetId);
        return map;
      });
      return result;
    },
    
    // Remove a tweet from the store
    removeTweet: (tweetId: string) => {
      update(map => {
        map.delete(tweetId);
        return map;
      });
    },
    
    // Reset the store
    reset: () => {
      set(new Map());
    }
  };
};

export const tweetInteractionStore = createTweetInteractionStore(); 