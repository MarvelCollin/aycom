import { writable, get } from 'svelte/store';
import type { ITweet, ITrend, ISuggestedFollow } from '../interfaces/ISocialMedia';

export function useSocialMedia() {
  // Stores for social media data
  const tweets = writable<ITweet[]>([]);
  const trends = writable<ITrend[]>([]);
  const suggestedUsers = writable<ISuggestedFollow[]>([]);
  const loading = writable(false);
  const error = writable<string | null>(null);

  // Initialize with mock data
  function initMockData() {
    // Mock tweets
    const mockTweets: ITweet[] = [
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
      },
      {
        id: 2,
        username: 'AYCOM',
        displayName: 'AYCOM Official',
        avatar: 'AY',
        content: 'Welcome to our new social media platform! #AYCOM #Launch',
        timestamp: '5h',
        likes: 1548,
        replies: 246,
        reposts: 567,
        views: '820K'
      },
      {
        id: 3,
        username: 'tech_news',
        displayName: 'Tech News',
        avatar: '📱',
        content: 'Breaking: New advancements in AI technology have researchers excited about future applications. #AI #Technology',
        timestamp: '12h',
        likes: 982,
        replies: 124,
        reposts: 325,
        views: '456K'
      }
    ];
    tweets.set(mockTweets);

    // Mock trends
    const mockTrends: ITrend[] = [
      { category: 'Trending', title: '#Trump', postCount: '1.95M' },
      { category: 'Trending', title: '#nct2d', postCount: '145K' },
      { category: 'Trending in Indonesia', title: 'Pagi', postCount: '6,974' },
      { category: 'Trending in Indonesia', title: '#reddysoekinspirasi', postCount: '5.4K' },
      { category: 'Trending in Indonesia', title: '#AYCOMLaunch', postCount: '12.8K' }
    ];
    trends.set(mockTrends);

    // Mock suggested users to follow
    const mockSuggestedUsers: ISuggestedFollow[] = [
      { 
        displayName: 'Brainwalla', 
        username: 'brainwalla', 
        avatar: '🧠', 
        verified: true,
        followerCount: 12300000
      },
      { 
        displayName: 'Peach', 
        username: 'peach', 
        avatar: '🍑', 
        verified: true,
        followerCount: 8500000
      },
      { 
        displayName: 'YTuber', 
        username: 'ytuber', 
        avatar: '▶️', 
        verified: false,
        followerCount: 5700000
      }
    ];
    suggestedUsers.set(mockSuggestedUsers);
  }

  // Initialize data
  initMockData();

  // Post a new tweet
  async function postTweet(content: string): Promise<boolean> {
    loading.set(true);
    error.set(null);

    try {
      const newTweet: ITweet = {
        id: Date.now(),
        username: 'johndoe', // Should come from auth
        displayName: 'John Doe',
        avatar: '👤',
        content,
        timestamp: 'now',
        likes: 0,
        replies: 0,
        reposts: 0,
        views: '0'
      };

      tweets.update(currentTweets => [newTweet, ...currentTweets]);
      return true;
    } catch (err) {
      console.error('Failed to post tweet:', err);
      error.set('Failed to post tweet');
      return false;
    } finally {
      loading.set(false);
    }
  }

  // Like a tweet
  function likeTweet(tweetId: number): void {
    tweets.update(currentTweets => 
      currentTweets.map(tweet => 
        tweet.id === tweetId 
          ? { ...tweet, likes: tweet.liked ? tweet.likes - 1 : tweet.likes + 1, liked: !tweet.liked } 
          : tweet
      )
    );
  }

  // Repost a tweet
  function repostTweet(tweetId: number): void {
    tweets.update(currentTweets => 
      currentTweets.map(tweet => 
        tweet.id === tweetId 
          ? { ...tweet, reposts: tweet.reposted ? tweet.reposts - 1 : tweet.reposts + 1, reposted: !tweet.reposted } 
          : tweet
      )
    );
  }

  // Reply to a tweet
  function replyToTweet(tweetId: number, content: string): void {
    tweets.update(currentTweets => 
      currentTweets.map(tweet => 
        tweet.id === tweetId 
          ? { ...tweet, replies: tweet.replies + 1 } 
          : tweet
      )
    );
    // In a real app, this would create a new tweet linked to the original
  }

  // Follow a suggested user
  function followUser(username: string): void {
    suggestedUsers.update(users => 
      users.map(user => 
        user.username === username 
          ? { ...user, isFollowing: true } 
          : user
      )
    );
  }

  // Unfollow a user
  function unfollowUser(username: string): void {
    suggestedUsers.update(users => 
      users.map(user => 
        user.username === username 
          ? { ...user, isFollowing: false } 
          : user
      )
    );
  }

  // Toggle bookmark status of a tweet
  function toggleBookmark(tweetId: number): void {
    tweets.update(currentTweets => 
      currentTweets.map(tweet => 
        tweet.id === tweetId 
          ? { ...tweet, bookmarked: !tweet.bookmarked } 
          : tweet
      )
    );
  }

  return {
    tweets,
    trends,
    suggestedUsers,
    loading,
    error,
    postTweet,
    likeTweet,
    repostTweet,
    replyToTweet,
    followUser,
    unfollowUser,
    toggleBookmark
  };
} 