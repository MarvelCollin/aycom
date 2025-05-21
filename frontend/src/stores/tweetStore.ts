import { writable } from 'svelte/store';
import type { ITweet } from '../interfaces/ISocialMedia';

// Type for tracking interaction status
interface InteractionStatus {
  isLiked: boolean;
  isBookmarked: boolean;
  isReposted: boolean;
  likes: number;
  bookmarks: number;
  reposts: number;
  replies: number;
  pendingLike?: boolean;
  pendingBookmark?: boolean;
  pendingRepost?: boolean;
}

// Map to track interaction states by tweet ID
const interactionMap = new Map<string, InteractionStatus>();

// Create the writable store
const tweetStore = writable({
  interactions: interactionMap,
  
  // Method to update multiple tweet interaction properties at once
  updateTweetInteraction: (id: string, updates: Partial<InteractionStatus>) => {
    if (!interactionMap.has(id)) {
      // Initialize if not exists
      interactionMap.set(id, {
        isLiked: false,
        isBookmarked: false,
        isReposted: false,
        likes: 0,
        bookmarks: 0,
        reposts: 0,
        replies: 0
      });
    }
    
    const currentStatus = interactionMap.get(id)!;
    
    // Apply all updates at once
    interactionMap.set(id, {
      ...currentStatus,
      ...updates
    });
    
    // Update the store to trigger reactivity
    tweetStore.update(store => ({
      ...store,
      interactions: new Map(interactionMap)
    }));
  },
  
  // Method to initialize a tweet's interaction state
  initTweet: (tweet: ITweet) => {
    const id = typeof tweet.id === 'number' ? String(tweet.id) : tweet.id;
    
    // If we already have data for this tweet, merge with existing data
    const existingData = interactionMap.get(id);
    const newData = {
      isLiked: tweet.isLiked || tweet.is_liked || false,
      isBookmarked: tweet.isBookmarked || false,
      isReposted: tweet.isReposted || false,
      likes: tweet.likes || tweet.likesCount || 0,
      bookmarks: tweet.bookmarks || 0,
      reposts: tweet.reposts || tweet.repostsCount || 0,
      replies: tweet.replies || tweet.commentsCount || 0
    };

    if (existingData) {
      // Only update values that are not in a pending state
      interactionMap.set(id, {
        ...existingData,
        ...newData,
        isLiked: existingData.pendingLike ? existingData.isLiked : newData.isLiked,
        isBookmarked: existingData.pendingBookmark ? existingData.isBookmarked : newData.isBookmarked,
        isReposted: existingData.pendingRepost ? existingData.isReposted : newData.isReposted
      });
    } else {
      interactionMap.set(id, newData);
    }
    
    // Update the store to trigger reactivity
    tweetStore.update(store => ({
      ...store,
      interactions: new Map(interactionMap)
    }));
  }
});

export const tweetInteractionStore = {
  subscribe: tweetStore.subscribe,
  updateTweetInteraction: (id: string, updates: Partial<InteractionStatus>) => {
    tweetStore.update(store => {
      store.updateTweetInteraction(id, updates);
      return store;
    });
  },
  initTweet: (tweet: ITweet) => {
    tweetStore.update(store => {
      store.initTweet(tweet);
      return store;
    });
  },
  getInteractionStatus: (id: string): InteractionStatus | undefined => {
    let result: InteractionStatus | undefined;
    tweetStore.update(store => {
      result = store.interactions.get(id);
      return store;
    });
    return result;
  }
}; 