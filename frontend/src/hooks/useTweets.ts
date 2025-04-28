import { writable } from 'svelte/store';

// Define tweet interface
export interface Tweet {
  id: number;
  username: string;
  displayName: string;
  avatar: string;
  content: string;
  timestamp: string;
  likes: number;
  replies: number;
  reposts: number;
  views: string;
}

// Sample data for initial tweets
const initialTweets: Tweet[] = [
  {
    id: 1,
    username: 'elonmusk',
    displayName: 'Elon Musk',
    avatar: '👨‍🚀',
    content: 'Just launched another rocket! 🚀 #SpaceX',
    timestamp: '2h',
    likes: 3240,
    replies: 421,
    reposts: 892,
    views: '1.2M'
  }
];

// Create a writable store for tweets
const tweetStore = writable<Tweet[]>(initialTweets);

// Counter for generating new tweet IDs
let nextId = initialTweets.length + 1;

export function useTweets() {
  // Get all tweets
  const getTweets = () => {
    return tweetStore;
  };

  // Add a new tweet
  const addTweet = (content: string, user: { username: string; displayName: string; avatar: string }) => {
    const newTweet: Tweet = {
      id: nextId++,
      username: user.username,
      displayName: user.displayName,
      avatar: user.avatar,
      content,
      timestamp: 'now',
      likes: 0,
      replies: 0,
      reposts: 0,
      views: '0'
    };
    
    tweetStore.update(tweets => [newTweet, ...tweets]);
    return newTweet;
  };

  // Like a tweet
  const likeTweet = (id: number) => {
    tweetStore.update(tweets => 
      tweets.map(tweet => 
        tweet.id === id 
          ? { ...tweet, likes: tweet.likes + 1 } 
          : tweet
      )
    );
  };

  // Repost a tweet
  const repostTweet = (id: number) => {
    tweetStore.update(tweets => 
      tweets.map(tweet => 
        tweet.id === id 
          ? { ...tweet, reposts: tweet.reposts + 1 } 
          : tweet
      )
    );
  };

  // Delete a tweet
  const deleteTweet = (id: number) => {
    tweetStore.update(tweets => tweets.filter(tweet => tweet.id !== id));
  };

  return {
    tweets: getTweets(),
    addTweet,
    likeTweet,
    repostTweet,
    deleteTweet
  };
} 