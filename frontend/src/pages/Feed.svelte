<script lang="ts">
  import MainLayout from '../components/layout/MainLayout.svelte';
  import ComposeTweet from '../components/social/ComposeTweet.svelte';
  import TweetCard from '../components/social/TweetCard.svelte';
  import Toast from '../components/common/Toast.svelte';
  import DebugPanel from '../components/common/DebugPanel.svelte';
  import { onMount } from 'svelte';
  import { useAuth } from '../hooks/useAuth';
  import { useTheme } from '../hooks/useTheme';
  import type { ITweet, ITrend, ISuggestedFollow } from '../interfaces/ISocialMedia';
  import type { IAuthStore } from '../interfaces/IAuth';
  import { getThreadsByUser, likeThread, unlikeThread, repostThread, bookmarkThread, removeBookmark, getAllThreads, getThreadReplies, getFollowingThreads, removeRepost } from '../api/thread';
  import { getTrends } from '../api/trends';
  import { getSuggestedUsers } from '../api/suggestions';
  import { createLoggerWithPrefix } from '../utils/logger';
  import { toastStore } from '../stores/toastStore';
  import { getProfile } from '../api/user';
  
  import ImageIcon from 'svelte-feather-icons/src/icons/ImageIcon.svelte';
  import FileIcon from 'svelte-feather-icons/src/icons/FileIcon.svelte';
  import BarChartIcon from 'svelte-feather-icons/src/icons/BarChartIcon.svelte';
  import SmileIcon from 'svelte-feather-icons/src/icons/SmileIcon.svelte';

  const logger = createLoggerWithPrefix('Feed');

  const { getAuthState } = useAuth();
  const { theme } = useTheme();

  $: authState = getAuthState ? (getAuthState() as IAuthStore) : { userId: null, isAuthenticated: false, accessToken: null, refreshToken: null };
  $: isDarkMode = $theme === 'dark';

  let username = '';
  let displayName = '';
  let avatar = 'https://secure.gravatar.com/avatar/0?d=mp'; 
  let isLoadingProfile = true;

  let tweetsForYou: ITweet[] = [];
  let tweetsFollowing: ITweet[] = [];
  let isLoadingForYou = true;
  let isLoadingFollowing = true;
  let errorForYou: string | null = null;
  let errorFollowing: string | null = null;
  let showComposeModal: boolean = false;
  let selectedTweet: ITweet | null = null;
  
  let activeTab: 'for-you' | 'following' = 'for-you';
  
  let pageForYou = 1;
  let pageFollowing = 1;
  let limit = 10;
  let hasMoreForYou = true;
  let hasMoreFollowing = true;
  
  let trends: ITrend[] = [];
  let isTrendsLoading = true;
  
  // Suggested users to follow
  let suggestedUsers: ISuggestedFollow[] = [];
  let isSuggestedUsersLoading = true;

  // Add nestedRepliesMap to track replies at different levels
  let repliesMap = new Map();
  let nestedRepliesMap = new Map(); // For storing replies to replies

  // Add isMobile variable near the top of the script
  let isMobile = false;

  // Check for mobile view on mount
  onMount(() => {
    // Simple check for mobile screens - can be replaced with a more sophisticated check
    isMobile = window.innerWidth < 768;
    
    // Add resize listener
    const handleResize = () => {
      isMobile = window.innerWidth < 768;
    };
    
    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  });

  // Convert thread data to tweet format
  function threadToTweet(thread: any): ITweet {
    // Check if we have debugging enabled
    const debug = false;
    if (debug) {
      console.log('Converting thread to tweet:', thread);
    }
    
    // Default values
    let username = 'anonymous';
    let displayName = 'User';
    let profilePicture = 'https://secure.gravatar.com/avatar/0?d=mp'; // Default avatar
    let content = thread.content || '';
    
    // Get author data from all possible locations
    // First try direct author fields
    if (thread.author_username) {
      username = thread.author_username;
    } else if (thread.authorUsername) {
      username = thread.authorUsername;
    } else if (thread.username) {
      username = thread.username;
    }
    
    if (thread.author_name) {
      displayName = thread.author_name;
    } else if (thread.authorName) {
      displayName = thread.authorName;
    } else if (thread.display_name) {
      displayName = thread.display_name;
    } else if (thread.displayName) {
      displayName = thread.displayName;
    }
    
    // Handle avatar URLs from Supabase
    if (thread.author_avatar) {
      profilePicture = formatSupabaseImageUrl(thread.author_avatar);
    } else if (thread.authorAvatar) {
      profilePicture = formatSupabaseImageUrl(thread.authorAvatar);
    } else if (thread.profile_picture_url) {
      profilePicture = formatSupabaseImageUrl(thread.profile_picture_url);
    } else if (thread.avatar) {
      profilePicture = formatSupabaseImageUrl(thread.avatar);
    }
    
    // Fallback: if user data is not directly in the thread, check for embedded content format
    if (username === 'anonymous' && typeof content === 'string') {
      // Look for enhanced user metadata that includes profile picture
      // Format: [USER:username@displayName@profileUrl]content
      const enhancedMetadataRegex = /^\[USER:([^@\]]+)@([^@\]]+)@([^\]]+)\](.*)/;
      const match = enhancedMetadataRegex.exec(content);
      
      if (match) {
        username = match[1] || username;
        displayName = match[2] || displayName;
        profilePicture = match[3] || profilePicture;
        content = match[4] || '';
      } else {
        // Try the old format without profile picture
        const userMetadataRegex = /^\[USER:([^@\]]+)(?:@([^\]]+))?\](.*)/;
        const basicMatch = content.match(userMetadataRegex);
        
        if (basicMatch) {
          username = basicMatch[1] || username;
          displayName = basicMatch[2] || displayName;
          content = basicMatch[3] || '';
        }
      }
    }

    // Safe date conversion with fallback
    let timestamp = new Date().toISOString();
    try {
      if (thread.created_at) {
        const date = new Date(thread.created_at);
        // Check if date is valid before converting to ISO string
        if (!isNaN(date.getTime())) {
          timestamp = date.toISOString();
        }
      } else if (thread.timestamp) {
        const date = new Date(thread.timestamp);
        if (!isNaN(date.getTime())) {
          timestamp = date.toISOString();
        }
      }
    } catch (error) {
      console.warn("Invalid date format in thread:", thread.created_at || thread.timestamp);
    }
    
    return {
      id: thread.id,
      threadId: thread.thread_id || thread.id,
      username: username,
      displayName: displayName,
      content: content,
      timestamp: timestamp,
      avatar: profilePicture,
      likes: thread.like_count || thread.metrics?.likes || 0,
      replies: thread.reply_count || thread.metrics?.replies || 0,
      reposts: thread.repost_count || thread.metrics?.reposts || 0,
      bookmarks: thread.bookmark_count || (thread.view_count > 0 ? thread.view_count : 0) || thread.metrics?.bookmarks || 0,
      views: '0', 
      media: thread.media || [],
      isLiked: thread.is_liked || false,
      isReposted: thread.is_repost || false,
      isBookmarked: thread.is_bookmarked || false,
      replyTo: null,
      isAdvertisement: thread.is_advertisement || false,
      communityId: thread.community_id || null,
      communityName: thread.community_name || null,
      authorId: thread.author_id || thread.authorId,
      authorName: thread.author_name || thread.authorName || displayName,
      authorUsername: thread.author_username || thread.authorUsername || username,
      authorAvatar: thread.author_avatar || thread.authorAvatar || profilePicture
    };
  }

  // Helper function to format Supabase image URLs
  function formatSupabaseImageUrl(url: string): string {
    if (!url) return 'https://secure.gravatar.com/avatar/0?d=mp';
    
    // If already a full URL, return as is
    if (url.startsWith('http')) return url;
    
    // Otherwise, construct the Supabase URL
    const supabaseUrl = import.meta.env.VITE_SUPABASE_URL || 'https://your-supabase-url.supabase.co';
    return `${supabaseUrl}/storage/v1/object/public/tpaweb/${url}`;
  }

  // Authentication check
  function checkAuth() {
    if (!authState.isAuthenticated) {
      logger.info('User not authenticated, redirecting to login page');
      
      // Only redirect if we're not already on the login page
      const currentPath = window.location.pathname;
      if (currentPath !== '/login') {
        window.location.href = '/login';
      }
      return false;
    }
    return true;
  }
  
  // Fetch user profile data using the API directly
  async function fetchUserProfile() {
    isLoadingProfile = true;
    try {
      const response = await getProfile();
      if (response && response.user) {
        username = response.user.username || '';
        displayName = response.user.name || response.user.display_name || username;
        
        // Use direct Supabase URL for profile picture if available
        if (response.user.profile_picture_url && response.user.profile_picture_url.startsWith('http')) {
          avatar = response.user.profile_picture_url;
        } else if (response.user.profile_picture_url) {
          // If it's a relative path or filename, construct proper Supabase URL
          const supabaseUrl = import.meta.env.VITE_SUPABASE_URL || 'https://your-supabase-url.supabase.co';
          avatar = `${supabaseUrl}/storage/v1/object/public/tpaweb/${response.user.profile_picture_url}`;
        } else {
          avatar = 'https://secure.gravatar.com/avatar/0?d=mp';
        }
        
        logger.debug('Profile loaded successfully', { username });
      } else {
        logger.warn('No user data received from API');
        // Set default values
        username = 'user';
        displayName = 'Guest User';
        avatar = 'https://secure.gravatar.com/avatar/0?d=mp';
      }
    } catch (error) {
      logger.error('Error fetching user profile:', error);
      // Set default values
      username = 'user';
      displayName = 'Guest User';
      avatar = 'https://secure.gravatar.com/avatar/0?d=mp';
    } finally {
      isLoadingProfile = false;
    }
  }

  // Function to fetch tweets for the "For You" tab
  async function fetchTweetsForYou(resetPage = false) {
    logger.info('Fetching tweets for the "For You" tab', { resetPage, page: pageForYou });
    
    if (resetPage) {
      pageForYou = 1;
      tweetsForYou = [];
    }
    
    isLoadingForYou = true;
    errorForYou = null;
    
    try {
      if (!checkAuth()) return;
      
      logger.debug('Fetching personalized feed');
      const response = await getAllThreads(pageForYou, limit);
      
      if (response && response.threads) {
        logger.info(`Received ${response.threads.length} threads from API`);
        
        // Process threads to identify replies and link them to parent threads
        const threadsMap = new Map();
        
        // First, convert all threads to tweets and create a map
        let convertedThreads = response.threads.map(thread => {
          const tweet = threadToTweet(thread);
          threadsMap.set(tweet.threadId, tweet);
          return tweet;
        });
        
        // Filter for community threads - only show if user is in that community
        convertedThreads = convertedThreads.filter(tweet => {
          // Show thread if not from a community or user is in that community
          return !tweet.communityId || (tweet.communityId && true); // Replace true with check if user is in community
        });
        
        // Insert advertisements
        const tweetsWithAds: ITweet[] = [];
        convertedThreads.forEach((tweet, index) => {
          tweetsWithAds.push(tweet);
          
          // After every 5 tweets, add an advertisement
          if ((index + 1) % 5 === 0) {
            tweetsWithAds.push({
              id: `ad-${Date.now()}-${index}`,
              threadId: `ad-${Date.now()}-${index}`,
              username: 'advertisement',
              displayName: 'Advertisement',
              content: 'Sponsored Content',
              timestamp: new Date().toISOString(),
              avatar: '/assets/ad-icon.png', // Use proper ad icon path
              likes: 0,
              replies: 0,
              reposts: 0,
              bookmarks: 0,
              views: '0',
              media: [],
              isLiked: false,
              isReposted: false,
              isBookmarked: false,
              replyTo: null,
              isAdvertisement: true,
              authorId: '',
              authorName: 'Advertisement',
              authorUsername: 'advertisement',
              authorAvatar: '/assets/ad-icon.png'
            });
          }
        });
        
        // If first page, replace tweets, otherwise append
        tweetsForYou = pageForYou === 1 ? tweetsWithAds : [...tweetsForYou, ...tweetsWithAds];
        
        // Check if there are more threads to load
        hasMoreForYou = convertedThreads.length === limit;
        pageForYou++;
        
        logger.debug('Updated tweets state', { 
          totalTweets: tweetsForYou.length, 
          hasMore: hasMoreForYou, 
          nextPage: pageForYou 
        });
      } else {
        logger.info('No threads received from API');
        hasMoreForYou = false;
      }
    } catch (err) {
      console.error('Error loading feed:', err);
      toastStore.showToast('Failed to load feed. Please try again.', 'error');
      errorForYou = err instanceof Error ? err.message : 'Failed to fetch tweets';
    } finally {
      isLoadingForYou = false;
    }
  }

  // Function to fetch tweets for the "Following" tab
  async function fetchTweetsFollowing(resetPage = false) {
    logger.info('Fetching tweets for the "Following" tab', { resetPage, page: pageFollowing });
    
    if (resetPage) {
      pageFollowing = 1;
      tweetsFollowing = [];
    }
    
    isLoadingFollowing = true;
    errorFollowing = null;
    
    try {
      if (!checkAuth()) return;
      
      logger.debug('Fetching following feed');
      // Here we'll call a specific API to get tweets from users the current user follows
      const response = await getFollowingThreads(pageFollowing, limit);
      
      if (response && response.threads) {
        logger.info(`Received ${response.threads.length} following threads from API`);
        
        // Convert threads to tweets
        let convertedThreads = response.threads.map(thread => threadToTweet(thread));
        
        // Insert advertisements
        const tweetsWithAds: ITweet[] = [];
        convertedThreads.forEach((tweet, index) => {
          tweetsWithAds.push(tweet);
          
          // After every 5 tweets, add an advertisement
          if ((index + 1) % 5 === 0) {
            tweetsWithAds.push({
              id: `ad-${Date.now()}-${index}`,
              threadId: `ad-${Date.now()}-${index}`,
              username: 'advertisement',
              displayName: 'Advertisement',
              content: 'Sponsored Content',
              timestamp: new Date().toISOString(),
              avatar: '/assets/ad-icon.png', // Use proper ad icon path
              likes: 0,
              replies: 0,
              reposts: 0,
              bookmarks: 0,
              views: '0',
              media: [],
              isLiked: false,
              isReposted: false,
              isBookmarked: false,
              replyTo: null,
              isAdvertisement: true,
              authorId: '',
              authorName: 'Advertisement',
              authorUsername: 'advertisement',
              authorAvatar: '/assets/ad-icon.png'
            });
          }
        });
        
        // If first page, replace tweets, otherwise append
        tweetsFollowing = pageFollowing === 1 ? tweetsWithAds : [...tweetsFollowing, ...tweetsWithAds];
        
        // Check if there are more threads to load
        hasMoreFollowing = convertedThreads.length === limit;
        pageFollowing++;
        
        logger.debug('Updated following tweets state', { 
          totalTweets: tweetsFollowing.length, 
          hasMore: hasMoreFollowing, 
          nextPage: pageFollowing 
        });
      } else {
        logger.info('No following threads received from API');
        hasMoreFollowing = false;
      }
    } catch (err) {
      console.error('Error loading following feed:', err);
      toastStore.showToast('Failed to load following feed. Please try again.', 'error');
      errorFollowing = err instanceof Error ? err.message : 'Failed to fetch following tweets';
    } finally {
      isLoadingFollowing = false;
    }
  }

  // Function to handle tab change
  function handleTabChange(tab: 'for-you' | 'following') {
    activeTab = tab;
    
    // Load data for the selected tab if it's empty
    if (tab === 'for-you' && tweetsForYou.length === 0) {
      fetchTweetsForYou(true);
    } else if (tab === 'following' && tweetsFollowing.length === 0) {
      fetchTweetsFollowing(true);
    }
  }

  async function fetchTrends() {
    logger.debug('Fetching trends');
    isTrendsLoading = true;
    
    try {
      const trendData = await getTrends(5);
      trends = trendData;
      logger.debug('Trends loaded', { trendsCount: trends.length });
    } catch (error) {
      console.error('Error loading trends:', error);
      toastStore.showToast('Failed to load trends. Please try again.', 'error');
      trends = [];
    } finally {
      isTrendsLoading = false;
    }
  }

  async function fetchSuggestedUsers() {
    logger.debug('Fetching suggested users');
    isSuggestedUsersLoading = true;
    
    try {
      const userData = await getSuggestedUsers(3);
      suggestedUsers = userData;
      logger.debug('Suggested users loaded', { count: suggestedUsers.length });
    } catch (error) {
      console.error('Error loading suggestions:', error);
      toastStore.showToast('Failed to load suggestions. Please try again.', 'error');
      suggestedUsers = [];
    } finally {
      isSuggestedUsersLoading = false;
    }
  }

  onMount(async () => {
    console.log('Feed page - Auth state:', authState.isAuthenticated, 'Current path:', window.location.pathname);
    
    // Let the Router handle redirects rather than doing it directly
    if (!authState.isAuthenticated) {
      logger.info('User not authenticated, letting Router handle the redirect');
      return;
    }
    
    // Load user profile first
    await fetchUserProfile();
    
    // Then fetch tweets, trends, etc.
    fetchTweetsForYou();
    fetchTweetsFollowing();
    
    // Fetch trends and suggestions in parallel
    Promise.all([
      fetchTrends(),
      fetchSuggestedUsers()
    ]).catch(error => {
      logger.error('Error fetching additional data:', error);
    });
  });

  // Note for the backend: these functions are placeholders that need to be implemented
  // We need to implement:
  // 1. getFollowingThreads in the API to get tweets from followed users
  // 2. Check if a user is part of a community
  // 3. Add advertisement functionality to the backend

  function toggleComposeModal() {
    logger.debug('Toggling compose modal', { currentState: showComposeModal });
    if (!authState.isAuthenticated) {
      toastStore.showToast('You need to log in to create posts', 'warning');
      return;
    }
    showComposeModal = !showComposeModal;
  }
  
  function openThreadModal(tweet: ITweet) {
    logger.debug('Opening thread modal', { tweetId: tweet.id });
    selectedTweet = tweet;
  }
  
  function closeThreadModal() {
    logger.debug('Closing thread modal');
    selectedTweet = null;
  }
  
  // When a new tweet is created, refresh the feed
  function handleNewTweet() {
    logger.info('New tweet created, refreshing feed');
    if (activeTab === 'for-you') {
      fetchTweetsForYou(true);
    } else {
      fetchTweetsFollowing(true);
    }
    toggleComposeModal();
  }
  
  // Handle tweet like
  async function handleTweetLike(event: CustomEvent) {
    const tweetId = event.detail;
    if (!authState.isAuthenticated) {
      toastStore.showToast('You need to log in to like posts', 'warning');
      return;
    }
    logger.info('Like tweet action', { tweetId });
    
    try {
      await likeThread(tweetId);
      toastStore.showToast('Tweet liked', 'success');
      
      // Update both tweet arrays
      tweetsForYou = tweetsForYou.map(tweet => {
        if (tweet.id === tweetId) {
          return { ...tweet, likes: (tweet.likes || 0) + 1, isLiked: true, is_liked: true };
        }
        return tweet;
      });
      
      tweetsFollowing = tweetsFollowing.map(tweet => {
        if (tweet.id === tweetId) {
          return { ...tweet, likes: (tweet.likes || 0) + 1, isLiked: true, is_liked: true };
        }
        return tweet;
      });
    } catch (error) {
      console.error('Error liking tweet:', error);
      toastStore.showToast('Failed to like tweet', 'error');
    }
  }
  
  // Handle tweet unlike
  async function handleTweetUnlike(event: CustomEvent) {
    const tweetId = event.detail;
    if (!authState.isAuthenticated) {
      toastStore.showToast('You need to log in to unlike posts', 'warning');
      return;
    }
    logger.info('Unlike tweet action', { tweetId });
    
    try {
      await unlikeThread(tweetId);
      toastStore.showToast('Tweet unliked', 'success');
      
      // Update both tweet arrays
      tweetsForYou = tweetsForYou.map(tweet => {
        if (tweet.id === tweetId) {
          return { ...tweet, likes: Math.max(0, (tweet.likes || 0) - 1), isLiked: false, is_liked: false };
        }
        return tweet;
      });
      
      tweetsFollowing = tweetsFollowing.map(tweet => {
        if (tweet.id === tweetId) {
          return { ...tweet, likes: Math.max(0, (tweet.likes || 0) - 1), isLiked: false, is_liked: false };
        }
        return tweet;
      });
    } catch (error) {
      console.error('Error unliking tweet:', error);
      toastStore.showToast('Failed to unlike tweet', 'error');
    }
  }
  
  // Handle tweet reply
  function handleTweetReply(event: CustomEvent) {
    const tweetId = event.detail;
    if (!authState.isAuthenticated) {
      toastStore.showToast('You need to log in to reply to posts', 'warning');
      return;
    }
    logger.info('Reply to tweet action', { tweetId });
    
    // Find the tweet in either array
    const tweetToReply = tweetsForYou.find(t => t.id === tweetId) || 
                         tweetsFollowing.find(t => t.id === tweetId);
    
    if (!tweetToReply) {
      toastStore.showToast('Cannot find the tweet to reply to', 'error');
      return;
    }
    
    // Store the tweet to reply to and open the compose modal
    selectedTweet = tweetToReply;
    showComposeModal = true;
  }
  
  // New function: Handle tweet repost
  async function handleTweetRepost(event: CustomEvent) {
    const tweetId = event.detail;
    if (!authState.isAuthenticated) {
      toastStore.showToast('You need to log in to repost', 'warning');
      return;
    }
    logger.info('Repost tweet action', { tweetId });
    
    try {
      await repostThread(tweetId);
      toastStore.showToast('Tweet reposted', 'success');
      
      // Update both tweet arrays
      tweetsForYou = tweetsForYou.map(tweet => {
        if (tweet.id === tweetId) {
          return { ...tweet, reposts: (tweet.reposts || 0) + 1, isReposted: true };
        }
        return tweet;
      });
      
      tweetsFollowing = tweetsFollowing.map(tweet => {
        if (tweet.id === tweetId) {
          return { ...tweet, reposts: (tweet.reposts || 0) + 1, isReposted: true };
        }
        return tweet;
      });
    } catch (error) {
      toastStore.showToast('Failed to repost tweet', 'error');
    }
  }
  
  // New function: Handle tweet unrepost
  async function handleTweetUnrepost(event: CustomEvent) {
    const tweetId = event.detail;
    if (!authState.isAuthenticated) {
      toastStore.showToast('You need to log in to remove a repost', 'warning');
      return;
    }
    logger.info('Unrepost tweet action', { tweetId });
    
    try {
      await removeRepost(tweetId);
      toastStore.showToast('Repost removed', 'success');
      
      // Update both tweet arrays
      tweetsForYou = tweetsForYou.map(tweet => {
        if (tweet.id === tweetId) {
          return { ...tweet, reposts: Math.max(0, (tweet.reposts || 0) - 1), isReposted: false };
        }
        return tweet;
      });
      
      tweetsFollowing = tweetsFollowing.map(tweet => {
        if (tweet.id === tweetId) {
          return { ...tweet, reposts: Math.max(0, (tweet.reposts || 0) - 1), isReposted: false };
        }
        return tweet;
      });
    } catch (error) {
      toastStore.showToast('Failed to remove repost', 'error');
    }
  }
  
  async function handleTweetBookmark(event: CustomEvent) {
    const tweetId = event.detail;
    if (!authState.isAuthenticated) {
      toastStore.showToast('You need to log in to bookmark posts', 'warning');
      return;
    }
    logger.info('Bookmark tweet action', { tweetId });
    
    try {
      // Attempt to bookmark the thread
      console.log(`Attempting to bookmark thread: ${tweetId}`);
      const response = await bookmarkThread(tweetId);
      console.log(`Bookmark response:`, response);
      toastStore.showToast('Tweet bookmarked', 'success');
      
      // Update UI state to reflect the bookmark
      tweetsForYou = tweetsForYou.map(tweet => {
        if (tweet.id === tweetId) {
          console.log(`Marking tweet ${tweetId} as bookmarked`);
          return { ...tweet, bookmarks: (tweet.bookmarks || 0) + 1, isBookmarked: true };
        }
        return tweet;
      });
      
      tweetsFollowing = tweetsFollowing.map(tweet => {
        if (tweet.id === tweetId) {
          return { ...tweet, bookmarks: (tweet.bookmarks || 0) + 1, isBookmarked: true };
        }
        return tweet;
      });
    } catch (error) {
      console.error('Error bookmarking tweet:', error);
      toastStore.showToast('Failed to bookmark tweet', 'error');
    }
  }
  
  async function handleTweetUnbookmark(event: CustomEvent) {
    const tweetId = event.detail;
    if (!authState.isAuthenticated) {
      toastStore.showToast('You need to log in to remove bookmarks', 'warning');
      return;
    }
    logger.info('Unbookmark tweet action', { tweetId });
    
    try {
      // Attempt to remove the bookmark
      console.log(`Attempting to remove bookmark from thread: ${tweetId}`);
      const response = await removeBookmark(tweetId);
      console.log(`Unbookmark response:`, response);
      toastStore.showToast('Bookmark removed', 'success');
      
      // Update UI state to reflect the unbookmark
      tweetsForYou = tweetsForYou.map(tweet => {
        if (tweet.id === tweetId) {
          console.log(`Marking tweet ${tweetId} as not bookmarked`);
          return { ...tweet, bookmarks: Math.max(0, (tweet.bookmarks || 0) - 1), isBookmarked: false };
        }
        return tweet;
      });
      
      tweetsFollowing = tweetsFollowing.map(tweet => {
        if (tweet.id === tweetId) {
          return { ...tweet, bookmarks: Math.max(0, (tweet.bookmarks || 0) - 1), isBookmarked: false };
        }
        return tweet;
      });
    } catch (error) {
      console.error('Error removing bookmark:', error);
      toastStore.showToast('Failed to remove bookmark', 'error');
    }
  }
  
  // Load replies for a specific thread
  async function handleLoadReplies(event: CustomEvent) {
    const threadId = event.detail;
    logger.debug(`Loading replies for thread: ${threadId}`);
    await fetchRepliesForThread(threadId);
  }

  // Function to fetch replies for a given thread
  async function fetchRepliesForThread(threadId: string) {
    logger.debug(`Fetching replies for thread: ${threadId}`);
    
    try {
      const response = await getThreadReplies(threadId);
      
      if (response && response.replies && response.replies.length > 0) {
        logger.info(`Received ${response.replies.length} replies for thread ${threadId}`);
        
        // Debug the raw reply data structure
        console.log('Sample reply structure for thread', threadId, ':', response.replies[0]);
        
        // Create a more detailed mapping for replies that properly extracts user data
        const convertedReplies = response.replies.map(reply => {
          // Extract core data
          const replyData = reply.reply || reply;
          
          // Handle user data which might be nested or at the top level
          const userData = reply.user || {};
          
          // Build a comprehensive reply object that ensures all fields are populated
          const enrichedReply = {
            id: replyData.id,
            thread_id: replyData.thread_id || threadId,
            content: replyData.content || '',
            created_at: replyData.created_at || new Date().toISOString(),
            author_id: userData.id || replyData.user_id,
            author_username: userData.username || reply.author_username,
            author_name: userData.name || reply.author_name,
            author_avatar: userData.profile_picture_url || reply.author_avatar,
            parent_id: replyData.parent_id,
            metrics: {
              likes: reply.likes_count || 0,
              replies: 0 // Replies to replies not tracked yet
            }
          };
          
          const convertedReply = threadToTweet(enrichedReply);
          
          // Ensure the parent references are set properly
          convertedReply.replyTo = threadId as any; // Use type assertion to avoid type error
          (convertedReply as any).parentReplyId = replyData.parent_id;
          
          return convertedReply;
        });
        
        // Store replies in the map for the thread
        repliesMap.set(threadId, convertedReplies);
        
        // Process nested replies (replies to replies)
        convertedReplies.forEach(reply => {
          const parentId = (reply as any).parentReplyId;
          if (parentId) {
            // If this reply has a parent that is not the main thread
            const parentReplies = nestedRepliesMap.get(parentId) || [];
            nestedRepliesMap.set(parentId, [...parentReplies, reply]);
          }
        });
        
        // Trigger reactivity update
        repliesMap = repliesMap;
        nestedRepliesMap = nestedRepliesMap;
        
        logger.debug(`Replies loaded for thread ${threadId}`, { count: convertedReplies.length });
      } else {
        logger.warn(`No replies returned for thread ${threadId}`);
        repliesMap.set(threadId, []);
        repliesMap = repliesMap;
      }
    } catch (error) {
      logger.error(`Error fetching replies for thread ${threadId}:`, error);
      toastStore.showToast('Failed to load replies. Please try again.', 'error');
      repliesMap.set(threadId, []);
      repliesMap = repliesMap;
    }
  }

  // Get the current tweets array based on active tab
  $: currentTweets = activeTab === 'for-you' ? tweetsForYou : tweetsFollowing;
  $: isLoading = activeTab === 'for-you' ? isLoadingForYou : isLoadingFollowing;
  $: error = activeTab === 'for-you' ? errorForYou : errorFollowing;
  $: hasMore = activeTab === 'for-you' ? hasMoreForYou : hasMoreFollowing;
  
  // Function to load more tweets based on active tab
  function loadMoreTweets() {
    if (activeTab === 'for-you') {
      fetchTweetsForYou();
    } else {
      fetchTweetsFollowing();
    }
  }
</script>

<MainLayout on:toggleComposeModal={toggleComposeModal} {username} {displayName} {avatar}>
  <div class="feed-container {isDarkMode ? 'dark-theme' : ''}">
    <div class="feed-header">
      <h1 class="feed-title">Home</h1>
    </div>

    <div class="feed-tabs">
      <button 
        class="feed-tab {activeTab === 'for-you' ? 'active' : ''}" 
        on:click={() => handleTabChange('for-you')}
        aria-selected={activeTab === 'for-you'}
        role="tab"
      >
        For You
      </button>
      <button 
        class="feed-tab {activeTab === 'following' ? 'active' : ''}" 
        on:click={() => handleTabChange('following')}
        aria-selected={activeTab === 'following'}
        role="tab"
      >
        Following
      </button>
    </div>

    <!-- Mobile compose tweet button for smaller screens -->
    {#if isMobile}
      <div class="feed-compose">
        <div class="compose-avatar">
          <img src={avatar} alt={username} />
        </div>
        <div class="compose-input-container">
          <button 
            class="compose-input" 
            on:click={toggleComposeModal}
            aria-label="Compose new post"
          >
            What's happening?
          </button>
          <div class="compose-tools">
            <button class="compose-tweet-tool">
              <ImageIcon size="20" />
            </button>
            <button class="compose-tweet-tool">
              <FileIcon size="20" />
            </button>
            <button class="compose-tweet-tool">
              <BarChartIcon size="20" />
            </button>
            <button class="compose-tweet-tool">
              <SmileIcon size="20" />
            </button>
          </div>
        </div>
      </div>
    {/if}
    
    <!-- Tweet List -->
    <div class="feed-items">
      <!-- Loading state when first loading tab -->
      {#if isLoading && currentTweets.length === 0}
        <div class="feed-loading">
          <div class="feed-loading-spinner"></div>
        </div>
      <!-- Error state -->
      {:else if error}
        <div class="empty-state">
          <div class="empty-state-title">Something went wrong</div>
          <div class="empty-state-message">{error}</div>
          <button 
            class="btn btn-primary" 
            on:click={() => activeTab === 'for-you' ? fetchTweetsForYou(true) : fetchTweetsFollowing(true)}
          >
            Try Again
          </button>
        </div>
      <!-- Empty state -->
      {:else if currentTweets.length === 0 && !isLoading}
        <div class="empty-state">
          <div class="empty-state-title">
            {#if activeTab === 'for-you'}
              No posts yet
            {:else}
              You're not following anyone yet
            {/if}
          </div>
          <div class="empty-state-message">
            {#if activeTab === 'for-you'}
              Start the conversation by creating your first post
            {:else}
              When you follow people, their posts will show up here
            {/if}
          </div>
          {#if activeTab === 'for-you'}
            <button 
              class="btn btn-primary" 
              on:click={toggleComposeModal}
            >
              Create First Post
            </button>
          {:else}
            <a 
              href="/explore" 
              class="btn btn-primary"
            >
              Find People to Follow
            </a>
          {/if}
        </div>
      <!-- Tweets list -->
      {:else}
        {#each currentTweets as tweet, index (tweet.id || `tweet-${index}`)}
          {#if tweet.isAdvertisement}
            <!-- Advertisement card using our CSS classes -->
            <div class="tweet-card">
              <div class="tweet-header">
                <div class="tweet-avatar">
                  <img src={tweet.avatar} alt="Advertisement" />
                </div>
                <div class="tweet-user-info">
                  <span class="tweet-user-name">{tweet.displayName}</span>
                  <span class="tweet-user-handle">@{tweet.username}</span>
                  <span class="tweet-ad-label">Advertisement</span>
                </div>
              </div>
              <div class="tweet-content">
                {tweet.content}
              </div>
              <div class="tweet-media">
                <div class="ad-content">
                  <p>Sponsored content goes here</p>
                </div>
              </div>
            </div>
          {:else}
            <TweetCard 
              tweet={tweet} 
              isDarkMode={isDarkMode} 
              isAuthenticated={authState.isAuthenticated}
              isLiked={tweet.isLiked || false}
              isReposted={tweet.isReposted || false}
              isBookmarked={tweet.isBookmarked || false}
              on:reply={handleTweetReply}
              on:repost={handleTweetRepost}
              on:unrepost={handleTweetUnrepost}
              on:like={handleTweetLike}
              on:unlike={handleTweetUnlike}
              on:bookmark={handleTweetBookmark}
              on:removeBookmark={handleTweetUnbookmark}
              on:loadReplies={handleLoadReplies}
              replies={repliesMap.get(tweet.id) || []}
              showReplies={false}
              nestedRepliesMap={nestedRepliesMap}
            />
          {/if}
        {/each}
        
        <!-- Loading more state -->
        {#if isLoading}
          <div class="feed-loading">
            <div class="feed-loading-spinner"></div>
          </div>
        <!-- Load more button -->
        {:else if hasMore}
          <div class="feed-pagination">
            <button 
              class="feed-load-more" 
              on:click={loadMoreTweets}
            >
              Load More
            </button>
          </div>
        {/if}
      {/if}
      
    </div>
  </div>
</MainLayout>

{#if showComposeModal}
  <ComposeTweet 
    {isDarkMode}
    on:close={toggleComposeModal}
    on:tweet={handleNewTweet}
    avatar={avatar}
    replyTo={selectedTweet}
  />
{/if}

<!-- Toast notifications -->
<Toast />

<!-- Debug panel -->
<DebugPanel />

<style>
  .tweet-ad-label {
    background-color: var(--color-primary-light);
    color: var(--color-primary);
    padding: var(--space-1) var(--space-2);
    border-radius: var(--radius-full);
    font-size: var(--font-size-xs);
    margin-left: var(--space-2);
  }

  .ad-content {
    background-color: var(--bg-secondary);
    border-radius: var(--radius-md);
    padding: var(--space-3);
    color: var(--text-secondary);
  }
  
</style>