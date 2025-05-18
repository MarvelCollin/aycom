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
}

// Map to track interaction states by tweet ID
const interactionMap = new Map<string, InteractionStatus>();

// Create the writable store
const tweetStore = writable({
  interactions: interactionMap,
  
  // Method to update a tweet's like state
  updateLike: (id: string, isLiked: boolean) => {
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
    const likes = isLiked 
      ? currentStatus.likes + 1
      : Math.max(0, currentStatus.likes - 1);
    
    interactionMap.set(id, {
      ...currentStatus,
      isLiked,
      likes
    });
    
    // Update the store to trigger reactivity
    tweetStore.update(store => ({
      ...store,
      interactions: new Map(interactionMap)
    }));
  },
  
  // Method to update a tweet's bookmark state
  updateBookmark: (id: string, isBookmarked: boolean) => {
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
    const bookmarks = isBookmarked 
      ? currentStatus.bookmarks + 1
      : Math.max(0, currentStatus.bookmarks - 1);
    
    interactionMap.set(id, {
      ...currentStatus,
      isBookmarked,
      bookmarks
    });
    
    // Update the store to trigger reactivity
    tweetStore.update(store => ({
      ...store,
      interactions: new Map(interactionMap)
    }));
  },
  
  // Method to update a tweet's repost state
  updateRepost: (id: string, isReposted: boolean) => {
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
    const reposts = isReposted 
      ? currentStatus.reposts + 1
      : Math.max(0, currentStatus.reposts - 1);
    
    interactionMap.set(id, {
      ...currentStatus,
      isReposted,
      reposts
    });
    
    // Update the store to trigger reactivity
    tweetStore.update(store => ({
      ...store,
      interactions: new Map(interactionMap)
    }));
  },
  
  // Method to update a tweet's reply count
  updateReplyCount: (id: string, count: number) => {
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
    
    interactionMap.set(id, {
      ...currentStatus,
      replies: count
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
    
    // If we already have data for this tweet, don't overwrite it
    if (!interactionMap.has(id)) {
      interactionMap.set(id, {
        isLiked: tweet.isLiked || tweet.is_liked || false,
        isBookmarked: tweet.isBookmarked || false,
        isReposted: tweet.isReposted || false,
        likes: tweet.likes || tweet.likesCount || 0,
        bookmarks: tweet.bookmarks || 0,
        reposts: tweet.reposts || tweet.repostsCount || 0,
        replies: tweet.replies || tweet.commentsCount || 0
      });
      
      // Update the store to trigger reactivity
      tweetStore.update(store => ({
        ...store,
        interactions: new Map(interactionMap)
      }));
    }
  }
});

export const tweetInteractionStore = {
  subscribe: tweetStore.subscribe,
  updateLike: (id: string, isLiked: boolean) => {
    tweetStore.update(store => {
      store.updateLike(id, isLiked);
      return store;
    });
  },
  updateBookmark: (id: string, isBookmarked: boolean) => {
    tweetStore.update(store => {
      store.updateBookmark(id, isBookmarked);
      return store;
    });
  },
  updateRepost: (id: string, isReposted: boolean) => {
    tweetStore.update(store => {
      store.updateRepost(id, isReposted);
      return store;
    });
  },
  updateReplyCount: (id: string, count: number) => {
    tweetStore.update(store => {
      store.updateReplyCount(id, count);
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