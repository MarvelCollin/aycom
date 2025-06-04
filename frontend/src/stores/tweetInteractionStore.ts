import { writable } from 'svelte/store';
import type { ITweet } from '../interfaces/ISocialMedia';
import { likeThread, unlikeThread } from '../api/thread';

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
  retry_count?: number;
  last_interaction?: number;
}

// Maximum number of retries for failed operations
const MAX_RETRIES = 3;

const createTweetInteractionStore = () => {
  const interactions = new Map<string, TweetInteractionState>();
  const { subscribe, update, set } = writable(interactions);

  // Initialize from localStorage
  try {
    const savedLikes = JSON.parse(localStorage.getItem('likedThreads') || '{}');
    Object.entries(savedLikes).forEach(([id, timestamp]) => {
      interactions.set(id, {
        is_liked: true,
        is_reposted: false,
        is_bookmarked: false,
        likes: 1,
        reposts: 0,
        replies: 0,
        bookmarks: 0,
        pending_like: false,
        pending_repost: false,
        pending_bookmark: false,
        pending_reply: false,
        last_interaction: timestamp as number
      });
    });
  } catch (e) {
    console.error('Failed to load saved likes from localStorage', e);
  }

  // Function to sync pending likes with the server
  const syncPendingInteractions = async () => {
    if (!navigator.onLine) return;

    let syncPromises: Promise<void>[] = [];

    update(map => {
      map.forEach((state, tweetId) => {
        // Only process items that are pending and haven't exceeded retry limit
        if ((state.pending_like || state.pending_bookmark || state.pending_repost) && 
            (!state.retry_count || state.retry_count < MAX_RETRIES)) {
          
          // Increment retry count
          const newState = { ...state, retry_count: (state.retry_count || 0) + 1 };
          map.set(tweetId, newState);

          // Create a promise for this sync operation
          const syncPromise = (async () => {
            try {
              // Process like operations
              if (state.pending_like) {
                // Call the appropriate API based on the current state
                if (state.is_liked) {
                  await likeThread(tweetId);
                  console.log(`✅ Successfully liked tweet ${tweetId} on server`);
                } else {
                  await unlikeThread(tweetId);
                  console.log(`✅ Successfully unliked tweet ${tweetId} on server`);
                }
              
                // Success - clear pending flag
                update(innerMap => {
                  const currentState = innerMap.get(tweetId);
                  if (currentState) {
                    const updatedState: TweetInteractionState = { 
                      ...currentState,
                      pending_like: false,
                      retry_count: 0
                    };
                    innerMap.set(tweetId, updatedState);
                  }
                  return innerMap;
                });
              }
              
              // TODO: Handle pending_bookmark and pending_repost similarly
              // Left commented to focus on like/unlike functionality
              /*
              if (state.pending_bookmark) {
                // Handle bookmark operations
              }
              
              if (state.pending_repost) {
                // Handle repost operations
              }
              */
            } catch (error) {
              console.error(`Failed to sync interaction state for tweet ${tweetId}:`, error);
              // Only increment retry count on server errors, not client errors
              const errorMsg = String(error).toLowerCase();
              if (errorMsg.includes('network') || errorMsg.includes('timeout') || 
                  errorMsg.includes('failed to fetch')) {
                // Leave pending flag for next sync attempt on network errors
              } else {
                // For other errors (like already liked/unliked), clear the pending flag
                update(innerMap => {
                  const currentState = innerMap.get(tweetId);
                  if (currentState) {
                    const updatedState: TweetInteractionState = { 
                      ...currentState,
                      pending_like: false,
                      retry_count: 0
                    };
                    innerMap.set(tweetId, updatedState);
                  }
                  return innerMap;
                });
              }
            }
          })();
          
          syncPromises.push(syncPromise);
        }
      });
      return map;
    });

    // Wait for all sync operations to complete
    try {
      await Promise.allSettled(syncPromises);
      console.log(`✓ Completed syncing ${syncPromises.length} pending interactions with server`);
    } catch (error) {
      console.error('Error during interaction sync batch:', error);
    }
  };

  // Listen for online/offline events
  if (typeof window !== 'undefined') {
    window.addEventListener('online', syncPendingInteractions);
  }

  // Try to sync every minute when the app is active
  const syncInterval = setInterval(syncPendingInteractions, 60000);

  return {
    subscribe,

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
            pending_reply: false,
            last_interaction: Date.now()
          });
        }
        return map;
      });
    },

    updateTweetInteraction: (tweetId: string, changes: Partial<TweetInteractionState>) => {
      update(map => {
        if (map.has(tweetId)) {
          const current = map.get(tweetId)!;
          const newState: TweetInteractionState = { 
            ...current, 
            ...changes,
            // Ensure all required properties are defined (not undefined)
            is_liked: changes.is_liked ?? current.is_liked,
            is_reposted: changes.is_reposted ?? current.is_reposted,
            is_bookmarked: changes.is_bookmarked ?? current.is_bookmarked,
            likes: changes.likes ?? current.likes,
            reposts: changes.reposts ?? current.reposts,
            replies: changes.replies ?? current.replies,
            bookmarks: changes.bookmarks ?? current.bookmarks,
            last_interaction: Date.now()
          };
          map.set(tweetId, newState);
          
          // Update localStorage if like status changed and not pending
          if (changes.is_liked !== undefined && !changes.pending_like) {
            try {
              const likedItems = JSON.parse(localStorage.getItem('likedThreads') || '{}');
              if (changes.is_liked) {
                likedItems[tweetId] = Date.now();
              } else {
                delete likedItems[tweetId];
              }
              localStorage.setItem('likedThreads', JSON.stringify(likedItems));
            } catch (e) {
              console.error('Failed to update localStorage', e);
            }
          }
        } else {
          map.set(tweetId, {
            is_liked: changes.is_liked ?? false,
            is_reposted: changes.is_reposted ?? false,
            is_bookmarked: changes.is_bookmarked ?? false,
            likes: changes.likes ?? 0,
            reposts: changes.reposts ?? 0,
            replies: changes.replies ?? 0,
            bookmarks: changes.bookmarks ?? 0,
            pending_like: changes.pending_like ?? false,
            pending_repost: changes.pending_repost ?? false,
            pending_bookmark: changes.pending_bookmark ?? false,
            pending_reply: changes.pending_reply ?? false,
            retry_count: 0,
            last_interaction: Date.now()
          });
        }
        return map;
      });
    },

    getInteractionStatus: (tweetId: string): TweetInteractionState | undefined => {
      let result: TweetInteractionState | undefined;
      update(map => {
        result = map.get(tweetId);
        return map;
      });
      return result;
    },

    removeTweet: (tweetId: string) => {
      update(map => {
        map.delete(tweetId);
        return map;
      });
    },

    syncWithServer: () => {
      syncPendingInteractions();
    },

    reset: () => {
      set(new Map());
      try {
        localStorage.removeItem('likedThreads');
      } catch (e) {
        console.error('Failed to clear localStorage', e);
      }
    },

    // Clean up resources when the app unmounts
    destroy: () => {
      clearInterval(syncInterval);
      if (typeof window !== 'undefined') {
        window.removeEventListener('online', syncPendingInteractions);
      }
    }
  };
};

export const tweetInteractionStore = createTweetInteractionStore();