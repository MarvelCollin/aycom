import { writable } from "svelte/store";
import type { ITweet } from "../interfaces/ISocialMedia";

export interface InteractionStatus {
  is_liked: boolean;
  is_bookmarked: boolean;
  is_reposted: boolean;
  likes: number;
  bookmarks: number;
  reposts: number;
  replies: number;
  pending_like?: boolean;
  pending_bookmark?: boolean;
  pending_repost?: boolean;
}

const interactionMap = new Map<string, InteractionStatus>();

const tweetStore = writable({
  interactions: interactionMap,

  updateTweetInteraction: (id: string, updates: Partial<InteractionStatus>) => {
    if (!interactionMap.has(id)) {

      interactionMap.set(id, {
        is_liked: false,
        is_bookmarked: false,
        is_reposted: false,
        likes: 0,
        bookmarks: 0,
        reposts: 0,
        replies: 0
      });
    }

    const currentStatus = interactionMap.get(id)!;

    interactionMap.set(id, {
      ...currentStatus,
      ...updates
    });

    tweetStore.update(store => ({
      ...store,
      interactions: new Map(interactionMap)
    }));
  },

  initTweet: (tweet: ITweet) => {
    const id = typeof tweet.id === "number" ? String(tweet.id) : tweet.id;

    const existingData = interactionMap.get(id);
    const newData = {
      is_liked: tweet.is_liked || false,
      is_bookmarked: tweet.is_bookmarked || false,
      is_reposted: tweet.is_reposted || false,
      likes: tweet.likes_count || 0,
      bookmarks: tweet.bookmark_count || 0,
      reposts: tweet.reposts_count || 0,
      replies: tweet.replies_count || 0
    };

    if (existingData) {

      interactionMap.set(id, {
        ...existingData,
        ...newData,
        is_liked: existingData.pending_like ? existingData.is_liked : newData.is_liked,
        is_bookmarked: existingData.pending_bookmark ? existingData.is_bookmarked : newData.is_bookmarked,
        is_reposted: existingData.pending_repost ? existingData.is_reposted : newData.is_reposted
      });
    } else {
      interactionMap.set(id, newData);
    }

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