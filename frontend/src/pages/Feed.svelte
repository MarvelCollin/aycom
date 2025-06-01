<script lang="ts">
  import MainLayout from '../components/layout/MainLayout.svelte';
  import ComposeTweet from '../components/social/ComposeTweet.svelte';
  import TweetCard from '../components/social/TweetCard.svelte';
  import Toast from '../components/common/Toast.svelte';
  import DebugPanel from '../components/common/DebugPanel.svelte';
  import { onMount, onDestroy } from 'svelte';
  import { useAuth } from '../hooks/useAuth';
  import { useTheme } from '../hooks/useTheme';
  import type { ITweet, ITrend, ISuggestedFollow, IMedia } from '../interfaces/ISocialMedia';
  import type { IAuthStore } from '../interfaces/IAuth';
  import { getThreadsByUser, likeThread, unlikeThread, repostThread, bookmarkThread, removeBookmark, getAllThreads, getThreadReplies, getFollowingThreads, removeRepost, directFetchThreads } from '../api/thread';
  import { getTrends } from '../api/trends';
  import { getSuggestedUsers } from '../api/suggestions';
  import { createLoggerWithPrefix } from '../utils/logger';
  import { toastStore } from '../stores/toastStore';
  import { tweetInteractionStore } from '../stores/tweetStore';
  import { getProfile } from '../api/user';
  import appConfig, { checkApiHealth } from '../config/appConfig';
  
  import ImageIcon from 'svelte-feather-icons/src/icons/ImageIcon.svelte';
  import FileIcon from 'svelte-feather-icons/src/icons/FileIcon.svelte';
  import BarChartIcon from 'svelte-feather-icons/src/icons/BarChartIcon.svelte';
  import SmileIcon from 'svelte-feather-icons/src/icons/SmileIcon.svelte';

  const logger = createLoggerWithPrefix('Feed');

  const { getAuthState } = useAuth();
  const { theme } = useTheme();

  $: authState = getAuthState ? (getAuthState() as IAuthStore) : { 
    user_id: null, 
    is_authenticated: false, 
    access_token: null, 
    refresh_token: null 
  };
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

  let loadingMoreTweets = false;
  
  // Add this to store the cleanup function
  let scrollListenerCleanup: (() => void) | null = null;

  // Function to check scroll position and load more tweets if needed
  function checkScrollAndLoadMore() {
    // If already loading or no more to load, don't do anything
    if (loadingMoreTweets || isLoading) return;
    
    const scrollPosition = window.scrollY;
    const windowHeight = window.innerHeight;
    const bodyHeight = document.body.scrollHeight;
    
    // When user scrolls to near bottom (300px from bottom), load more
    if (bodyHeight - (scrollPosition + windowHeight) < 300) {
      loadingMoreTweets = true;
      logger.info('Near bottom of page, loading more tweets');
      
      loadMoreTweets().finally(() => {
        loadingMoreTweets = false;
      });
    }
  }
  
  // Scroll event listener setup (more reliable than IntersectionObserver)
  function setupScrollListener() {
    // Remove existing listener if any
    window.removeEventListener('scroll', checkScrollAndLoadMore);
    
    // Add new scroll listener
    window.addEventListener('scroll', checkScrollAndLoadMore);
    logger.info('Scroll listener set up for infinite scrolling');
    
    // Initial check in case page is already at bottom
    setTimeout(checkScrollAndLoadMore, 500);
    
    return () => {
      window.removeEventListener('scroll', checkScrollAndLoadMore);
    };
  }

  // Function to load more tweets with reset capability 
  async function loadMoreTweets() {
    try {
      if (activeTab === 'for-you') {
        // If no more tweets to fetch from API, go back to first page for infinite loop
        if (!hasMoreForYou) {
          logger.info('Reached end of available tweets, starting from beginning');
          pageForYou = 1; // Reset to first page
          hasMoreForYou = true; // Enable fetching again
        }
        
        await fetchTweetsForYou();
      } else {
        // If no more tweets to fetch from API, go back to first page for infinite loop
        if (!hasMoreFollowing) {
          logger.info('Reached end of available following tweets, starting from beginning');
          pageFollowing = 1; // Reset to first page
          hasMoreFollowing = true; // Enable fetching again
        }
        
        await fetchTweetsFollowing();
      }
    } catch (error) {
      logger.error('Error loading more tweets:', error);
    }
  }
  
  // Update onMount to have more debugging for API URLs
  onMount(async () => {
    console.log('[Feed Debug] Feed page mounted - Auth state:', authState.is_authenticated, 'Current path:', window.location.pathname);
    
    // Log API base URL for debugging
    console.log('[Feed Debug] API Base URL:', appConfig.api.baseUrl);
    
    // Check API health with more detailed logging
    console.log('[Feed Debug] Starting API health check...');
    
    try {
      // Test the API connection using /trends endpoint which we know is working
      console.log(`[Feed Debug] Testing API connection to ${appConfig.api.baseUrl}/trends`);
      const testResponse = await fetch(`${appConfig.api.baseUrl}/trends`, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
          'Accept': 'application/json'
        },
        mode: 'cors' // Explicitly set CORS mode
      });
      
      console.log(`[Feed Debug] API trends check response status:`, testResponse.status);
      
      if (testResponse.ok) {
        try {
          const testData = await testResponse.json();
          console.log('[Feed Debug] API connection test successful:', testData);
          
          if (testData && Array.isArray(testData)) {
            console.log('[Feed Debug] API is working properly, retrieved', testData.length, 'trends');
          }
        } catch (parseError) {
          console.error('[Feed Debug] Error parsing API test response:', parseError);
          
          // Try to get the raw text
          try {
            const rawText = await testResponse.text();
            console.log('[Feed Debug] Raw API response text:', rawText);
          } catch (textError) {
            console.error('[Feed Debug] Error getting raw response text:', textError);
          }
        }
      } else {
        console.error('[Feed Debug] API test failed with status:', testResponse.status);
      }
    } catch (connectionError) {
      console.error('[Feed Debug] API connection test error:', connectionError);
    }
    
    // Check API health and switch to fallback if needed
    await checkApiHealth();
    
    // Let the Router handle redirects rather than doing it directly
    if (!authState.is_authenticated) {
      logger.info('[Feed Debug] User not authenticated, letting Router handle the redirect');
      return;
    }
    
    // Load user profile first
    fetchUserProfile();
    
    // Then fetch tweets, trends, etc.
    console.log('[Feed Debug] Starting to fetch tweets for feed...');
    fetchTweetsForYou();
    fetchTweetsFollowing();
    
    // Fetch trends and suggestions in parallel
    Promise.all([
      fetchTrends(),
      fetchSuggestedUsers()
    ]).catch(error => {
      logger.error('[Feed Debug] Error fetching additional data:', error);
    });

    // Setup infinite scroll through scroll listener
    scrollListenerCleanup = setupScrollListener();
  });

  // Function to handle tab change
  function handleTabChange(tab: 'for-you' | 'following') {
    activeTab = tab;
    
    // Load data for the selected tab if it's empty
    if (tab === 'for-you' && tweetsForYou.length === 0) {
      fetchTweetsForYou(true);
    } else if (tab === 'following' && tweetsFollowing.length === 0) {
      fetchTweetsFollowing(true);
    }
    
    // Check scroll position after tab change
    setTimeout(checkScrollAndLoadMore, 100);
  }

  // Helper function to safely parse numeric metrics
  function parseMetric(value: any): number {
    if (value === undefined || value === null) return 0;
    if (typeof value === 'number') return value;
    if (typeof value === 'string') {
      const parsed = parseInt(value, 10);
      return isNaN(parsed) ? 0 : parsed;
    }
    return 0;
  }

  // Helper function to format timestamps
  function formatTimestamp(timestamp: string | Date | undefined): string {
    if (!timestamp) return new Date().toISOString();
    
    try {
      const date = new Date(timestamp);
      // Check if date is valid before converting to ISO string
      if (!isNaN(date.getTime())) {
        return date.toISOString();
      }
    } catch (error) {
      console.warn("Invalid date format:", timestamp);
    }
    
    return new Date().toISOString();
  }

  // Helper function to format profile pictures
  function formatProfilePicture(url: string | undefined): string {
    if (!url) return 'https://secure.gravatar.com/avatar/0?d=mp';
    
    // If already a full URL, return as is
    if (url.startsWith('http')) return url;
    
    const supabaseUrl = import.meta.env.VITE_SUPABASE_URL || 'https://your-supabase-url.supabase.co';
    return `${supabaseUrl}/storage/v1/object/public/tpaweb/${url}`;
  }

  // Authentication check
  function checkAuth() {
    if (!authState.is_authenticated) {
      logger.info('User not authenticated, redirecting to login page');
      
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
    if (resetPage) {
      pageForYou = 1;
      tweetsForYou = [];
    }
    
    if (!hasMoreForYou) {
      logger.info('No more tweets to fetch in For You feed');
      return;
    }
    
    if (isLoadingForYou && !resetPage) {
      logger.info('Already loading For You feed');
      return;
    }
    
    try {
      isLoadingForYou = true;
      errorForYou = null;
      
      // Log API URL for debugging
      console.log(`[Feed Debug] API URL: ${appConfig.api.baseUrl}`);
      logger.info(`Fetching For You feed - page ${pageForYou}, limit ${limit} from ${appConfig.api.baseUrl}/threads`);
      
      // Capture start time for performance measurement
      const startTime = performance.now();
      
      // Try the standard method first
      let response;
      try {
        console.log('[Feed Debug] Attempting standard API call via getAllThreads');
        response = await getAllThreads(pageForYou, limit);
      } catch (standardError) {
        console.error('[Feed Debug] Standard API call failed:', standardError);
        
        // If standard method fails, try direct fetch
        console.log('[Feed Debug] Falling back to direct fetch method');
        try {
          response = await directFetchThreads(pageForYou, limit);
          console.log('[Feed Debug] Direct fetch succeeded:', response);
        } catch (directError) {
          console.error('[Feed Debug] Direct fetch also failed:', directError);
          throw directError;
        }
      }
      
      const endTime = performance.now();
      
      // Log performance metrics
      logger.info(`API response time: ${Math.round(endTime - startTime)}ms`);
      
      // Log the raw API response to help debug
      console.log('[Feed Debug] Raw API response:', response);
      
      if (!response || !response.success) {
        const errorMessage = response && typeof response === 'object' && 'error' in response 
            ? response.error as string 
            : 'Failed to load tweets from server';
        
        logger.error('[Feed Debug] Error from API:', errorMessage);
        errorForYou = errorMessage;
        toastStore.showToast(errorForYou, 'error');
        return;
      }
      
      // Just use response.threads directly since that seems to be the format from the backend
      const threads = response.threads;
      
      if (!threads || !Array.isArray(threads) || threads.length === 0) {
        logger.info('[Feed Debug] No threads found in For You feed');
        hasMoreForYou = false;
        isLoadingForYou = false;
        return;
      }
      
      // Convert all threads to tweets
      const convertedThreads = threads.map(thread => {
        // Log each thread for debugging
        if (import.meta.env.DEV) {
          console.log('[Feed Debug] Thread from API:', thread);
        }
        
        // Use thread data directly from API
        const tweet: ITweet = {
          id: thread.id,
          content: thread.content || '',
          created_at: thread.created_at || new Date().toISOString(),
          updated_at: thread.updated_at,
          user_id: thread.user_id,
          username: thread.username,
          name: thread.name,
          profile_picture_url: thread.profile_picture_url,
          likes_count: thread.likes_count || 0,
          replies_count: thread.replies_count || 0,
          reposts_count: thread.reposts_count || 0,
          bookmark_count: thread.bookmark_count || 0,
          views_count: thread.views_count || 0,
          is_liked: !!thread.is_liked,
          is_reposted: !!thread.is_reposted, 
          is_bookmarked: !!thread.is_bookmarked,
          is_pinned: !!thread.is_pinned,
          parent_id: thread.parent_id || null,
          media: thread.media || [],
          community_id: thread.community_id,
          community_name: thread.community_name,
          is_advertisement: !!thread.is_advertisement
        };
        
        // Add to interaction store
        tweetInteractionStore.initTweet(tweet);
        
        return tweet;
      });
      
      // Add new tweets to our list
      tweetsForYou = [...tweetsForYou, ...convertedThreads];
      
      // Check if we have more tweets to fetch based on whether we got as many tweets as requested
      hasMoreForYou = convertedThreads.length >= limit;
      
      // Increment page for next fetch
      pageForYou++;
      
      logger.info(`[Feed Debug] Loaded ${convertedThreads.length} tweets for For You feed`);
      
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Error loading tweets. Please try again.';
      logger.error('[Feed Debug] Error fetching For You feed:', error);
      errorForYou = errorMessage;
      toastStore.showToast(errorForYou, 'error');
    } finally {
      isLoadingForYou = false;
    }
  }

  // Function to fetch tweets for the "Following" tab
  async function fetchTweetsFollowing(resetPage = false) {
    if (resetPage) {
      pageFollowing = 1;
      tweetsFollowing = [];
    }
    
    if (!hasMoreFollowing) {
      logger.info('No more tweets to fetch in Following feed');
      return;
    }
    
    if (isLoadingFollowing && !resetPage) {
      logger.info('Already loading Following feed');
      return;
    }
    
    try {
      isLoadingFollowing = true;
      errorFollowing = null;
      
      logger.info(`Fetching Following feed - page ${pageFollowing}, limit ${limit}`);
      const response = await getFollowingThreads(pageFollowing, limit);
      
      // Log the raw API response to help debug
      console.log('Raw API response for Following feed:', response);
      
      if (!response || !response.success) {
        logger.error('Error from API:', response);
        errorFollowing = response && typeof response === 'object' && 'error' in response 
          ? response.error as string 
          : 'Failed to load tweets from server';
        toastStore.showToast(errorFollowing || 'Failed to load tweets', 'error');
        return;
      }
      
      // Just use response.threads directly
      const threads = response.threads;
      
      if (!threads || !Array.isArray(threads) || threads.length === 0) {
        logger.info('No threads found in Following feed');
        hasMoreFollowing = false;
        isLoadingFollowing = false;
        return;
      }
      
      // Convert all threads to tweets
      const convertedThreads = threads.map(thread => {
        // Use thread data directly from API
        const tweet: ITweet = {
          id: thread.id,
          content: thread.content || '',
          created_at: thread.created_at || new Date().toISOString(),
          updated_at: thread.updated_at,
          user_id: thread.user_id,
          username: thread.username,
          name: thread.name,
          profile_picture_url: thread.profile_picture_url,
          likes_count: thread.likes_count || 0,
          replies_count: thread.replies_count || 0,
          reposts_count: thread.reposts_count || 0,
          bookmark_count: thread.bookmark_count || 0,
          views_count: thread.views_count || 0,
          is_liked: !!thread.is_liked,
          is_reposted: !!thread.is_reposted, 
          is_bookmarked: !!thread.is_bookmarked,
          is_pinned: !!thread.is_pinned,
          parent_id: thread.parent_id || null,
          media: thread.media || [],
          community_id: thread.community_id,
          community_name: thread.community_name,
          is_advertisement: !!thread.is_advertisement
        };
        
        // Add to interaction store
        tweetInteractionStore.initTweet(tweet);
        
        return tweet;
      });
      
      // Add new tweets to our list
      tweetsFollowing = [...tweetsFollowing, ...convertedThreads];
      
      // Check if we have more tweets to fetch
      hasMoreFollowing = convertedThreads.length >= limit;
      
      // Increment page for next fetch
      pageFollowing++;
      
      logger.info(`Loaded ${convertedThreads.length} tweets for Following feed`);
      
    } catch (error) {
      logger.error('Error fetching Following feed:', error);
      errorFollowing = error instanceof Error ? error.message : 'Error loading tweets. Please try again.';
      toastStore.showToast(errorFollowing || 'Error loading tweets', 'error');
    } finally {
      isLoadingFollowing = false;
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

  // Get the current tweets array based on active tab
  $: currentTweets = activeTab === 'for-you' ? tweetsForYou : tweetsFollowing;
  $: isLoading = activeTab === 'for-you' ? isLoadingForYou : isLoadingFollowing;
  $: error = activeTab === 'for-you' ? errorForYou : errorFollowing;
  $: hasMore = activeTab === 'for-you' ? hasMoreForYou : hasMoreFollowing;
  
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
          
          // Convert reply directly to the ITweet format
          const convertedReply: ITweet = {
            id: enrichedReply.id,
            content: enrichedReply.content || '',
            created_at: enrichedReply.created_at || new Date().toISOString(),
            user_id: enrichedReply.author_id,
            username: enrichedReply.author_username || 'anonymous',
            name: enrichedReply.author_name || 'User',
            profile_picture_url: enrichedReply.author_avatar || 'https://secure.gravatar.com/avatar/0?d=mp',
            likes_count: enrichedReply.metrics?.likes || 0,
            replies_count: 0,
            reposts_count: 0,
            bookmark_count: 0,
            views_count: 0,
            is_liked: false,
            is_reposted: false,
            is_bookmarked: false,
            is_pinned: false,
            parent_id: threadId,
            media: []
          };
          
          // Ensure the parent references are set properly
          convertedReply.parent_id = threadId;
          
          return convertedReply;
        });
        
        // Store replies in the map for the thread
        repliesMap.set(threadId, convertedReplies);
        
        // Process nested replies (replies to replies)
        convertedReplies.forEach(reply => {
          const parentId = reply.parent_id;
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

  function toggleComposeModal() {
    logger.debug('Toggling compose modal', { currentState: showComposeModal });
    if (!authState.is_authenticated) {
      toastStore.showToast('You need to log in to create posts', 'warning');
      return;
    }
    
    // Clear selectedTweet when opening for a new post (not a reply)
    if (!showComposeModal) {
      selectedTweet = null;
    }
    
    showComposeModal = !showComposeModal;
    
    logger.debug('Compose modal new state', { showComposeModal });
  }
  
  function openThreadModal(tweet: ITweet) {
    logger.debug('Opening thread modal', { tweetId: tweet.id });
    selectedTweet = tweet;
  }
  
  function closeThreadModal() {
    logger.debug('Closing thread modal');
    selectedTweet = null;
  }
  
  // Add this type to handle legacy properties in tweets
  type ExtendedTweet = ITweet & {
    threadId?: string;
    replyTo?: ITweet | null;
    userId?: string;
    displayName?: string;
    timestamp?: string;
    avatar?: string;
    likes?: number;
    replies?: number;
    reposts?: number;
    bookmarks?: number;
    isLiked?: boolean;
    isReposted?: boolean;
    isBookmarked?: boolean;
    isPinned?: boolean;
    [key: string]: any;
  };

  // Function to handle reply posted event
  function handleReplyPosted(event) {
    // @ts-ignore - Using legacy property for backward compatibility
    const { threadId, newReply } = event.detail;
    logger.info('Reply posted', { threadId });
    
    // Find the tweet that was replied to
    // @ts-ignore - Using legacy property for backward compatibility
    const repliedTweet = tweetsForYou.find(t => String(t.id) === String(threadId)) || 
                         tweetsFollowing.find(t => String(t.id) === String(threadId));
                         
    if (repliedTweet) {
      // Increment the reply count
      repliedTweet.replies_count = (parseInt(String(repliedTweet.replies_count)) || 0) + 1;
      
      // Update the store
      // @ts-ignore - Using legacy property for backward compatibility
      tweetInteractionStore.updateTweetInteraction(String(threadId), {
        replies: repliedTweet.replies_count
      });
      
      // Add the reply to our replies map if it exists
      if (repliesMap.has(threadId)) {
        // @ts-ignore - Using legacy property for backward compatibility
        const currentReplies = repliesMap.get(threadId) || [];
        // Convert the new reply directly to ITweet format
        const processedNewReply: ITweet = {
          id: newReply.id || `new-reply-${Date.now()}`,
          content: newReply.content || '',
          created_at: newReply.created_at || new Date().toISOString(),
          user_id: newReply.user_id,
          username: newReply.username || 'anonymous',
          name: newReply.name || 'User',
          profile_picture_url: newReply.profile_picture_url || 'https://secure.gravatar.com/avatar/0?d=mp',
          likes_count: newReply.likes_count || 0,
          replies_count: 0,
          reposts_count: 0,
          bookmark_count: 0,
          views_count: 0,
          is_liked: false,
          is_reposted: false,
          is_bookmarked: false,
          is_pinned: false,
          parent_id: threadId,
          media: newReply.media || []
        };
        // Set the parent_id to link to the original thread
        // @ts-ignore - Using legacy property for backward compatibility
        processedNewReply.parent_id = threadId;
        // @ts-ignore - Using legacy property for backward compatibility
        repliesMap.set(threadId, [processedNewReply, ...currentReplies]);
        repliesMap = repliesMap; // Trigger reactivity
      }
    }
    
    // Close the compose modal
    showComposeModal = false;
    selectedTweet = null;
  }
  
  // When a new tweet is created, refresh the feed
  function handleNewPost() {
    logger.info('New tweet created, refreshing feed');
    if (activeTab === 'for-you') {
      fetchTweetsForYou(true);
    } else {
      fetchTweetsFollowing(true);
    }
    toggleComposeModal();
  }
  
  // Handle tweet like
  async function handleLikeClick(event: CustomEvent) {
    const tweetId = event.detail;
    if (!authState.is_authenticated) {
      toastStore.showToast('You need to log in to like posts', 'warning');
      return;
    }
    
    try {
      await likeThread(tweetId);
      tweetInteractionStore.updateTweetInteraction(tweetId, { is_liked: true });
      toastStore.showToast('Post liked', 'success');
    } catch (error) {
      toastStore.showToast('Failed to like post', 'error');
      // Revert the optimistic update
      tweetInteractionStore.updateTweetInteraction(tweetId, { is_liked: false });
    }
  }
  
  // Handle tweet unlike
  async function handleUnlikeClick(event: CustomEvent) {
    const tweetId = event.detail;
    if (!authState.is_authenticated) {
      toastStore.showToast('You need to log in to unlike posts', 'warning');
      return;
    }
    
    try {
      await unlikeThread(tweetId);
      tweetInteractionStore.updateTweetInteraction(tweetId, { is_liked: false });
      toastStore.showToast('Post unliked', 'success');
    } catch (error) {
      toastStore.showToast('Failed to unlike post', 'error');
      // Revert the optimistic update
      tweetInteractionStore.updateTweetInteraction(tweetId, { is_liked: true });
    }
  }
  
  // Handle tweet reply
  function handleReply(event) {
    const tweetId = event.detail;
    if (!authState.is_authenticated) {
      toastStore.showToast('You need to log in to reply', 'warning');
      return;
    }
    
    // Find the tweet in either array
    let tweetToReply = tweetsForYou.find(t => String(t.id) === String(tweetId)) || 
                       tweetsFollowing.find(t => String(t.id) === String(tweetId));
    
    // If not found in main tweets, check in replies
    if (!tweetToReply) {
      // Check in all reply collections
      for (const [threadId, replies] of repliesMap.entries()) {
        const foundReply = replies.find(r => String(r.id) === String(tweetId));
        if (foundReply) {
          tweetToReply = foundReply;
          break;
        }
      }
    }
    
    // If still not found, check in nested replies
    if (!tweetToReply) {
      for (const [parentReplyId, nestedReplies] of nestedRepliesMap.entries()) {
        const foundReply = nestedReplies.find(r => String(r.id) === String(tweetId));
        if (foundReply) {
          tweetToReply = foundReply;
          break;
        }
      }
    }
    
    if (!tweetToReply) {
      console.error(`Cannot find tweet with ID ${tweetId} to reply to`);
      toastStore.showToast('Cannot find the tweet to reply to', 'error');
      return;
    }
    
    console.log('Found tweet to reply to:', tweetToReply);
    
    // Update the reply count in the store
    const currentReplies = tweetInteractionStore.getInteractionStatus(String(tweetId))?.replies || 0;
    tweetInteractionStore.updateTweetInteraction(String(tweetId), {
      replies: currentReplies + 1
    });
    
    // Store the tweet to reply to and open the compose modal
    selectedTweet = tweetToReply;
    showComposeModal = true;
    
    // Log for debugging
    console.log('Opened compose modal for reply to:', {
      id: selectedTweet.id,
      content: selectedTweet.content?.substring(0, 30) + '...',
      showComposeModal
    });
  }
  
  // New function: Handle tweet repost
  async function handleTweetRepost(event: CustomEvent) {
    const tweetId = event.detail;
    if (!authState.is_authenticated) {
      toastStore.showToast('You need to log in to repost', 'warning');
      return;
    }
    logger.info('Repost tweet action', { tweetId });
    
    try {
      await repostThread(tweetId);
      tweetInteractionStore.updateTweetInteraction(tweetId, { is_reposted: true });
      toastStore.showToast('Tweet reposted', 'success');
    } catch (error) {
      toastStore.showToast('Failed to repost tweet', 'error');
      // Revert the optimistic update
      tweetInteractionStore.updateTweetInteraction(tweetId, { is_reposted: false });
    }
  }
  
  // New function: Handle tweet unrepost
  async function handleTweetUnrepost(event: CustomEvent) {
    const tweetId = event.detail;
    if (!authState.is_authenticated) {
      toastStore.showToast('You need to log in to remove a repost', 'warning');
      return;
    }
    logger.info('Unrepost tweet action', { tweetId });
    
    try {
      await removeRepost(tweetId);
      tweetInteractionStore.updateTweetInteraction(tweetId, { is_reposted: false });
      toastStore.showToast('Repost removed', 'success');
    } catch (error) {
      toastStore.showToast('Failed to remove repost', 'error');
      // Revert the optimistic update
      tweetInteractionStore.updateTweetInteraction(tweetId, { is_reposted: true });
    }
  }
  
  async function handleTweetBookmark(event: CustomEvent) {
    const tweetId = event.detail;
    if (!authState.is_authenticated) {
      toastStore.showToast('You need to log in to bookmark posts', 'warning');
      return;
    }
    logger.info('Bookmark tweet action', { tweetId });
    
    try {
      // Attempt to bookmark the thread
      console.log(`Attempting to bookmark thread: ${tweetId}`);
      const response = await bookmarkThread(tweetId);
      console.log(`Bookmark response:`, response);
      
      // Update bookmark state in the store
      tweetInteractionStore.updateTweetInteraction(tweetId, { is_bookmarked: true });
      toastStore.showToast('Tweet bookmarked', 'success');
    } catch (error) {
      console.error('Error bookmarking tweet:', error);
      toastStore.showToast('Failed to bookmark tweet', 'error');
      // Revert the optimistic update
      tweetInteractionStore.updateTweetInteraction(tweetId, { is_bookmarked: false });
    }
  }
  
  async function handleTweetUnbookmark(event: CustomEvent) {
    const tweetId = event.detail;
    if (!authState.is_authenticated) {
      toastStore.showToast('You need to log in to remove bookmarks', 'warning');
      return;
    }
    logger.info('Unbookmark tweet action', { tweetId });
    
    try {
      // Attempt to remove the bookmark
      console.log(`Attempting to remove bookmark from thread: ${tweetId}`);
      const response = await removeBookmark(tweetId);
      console.log(`Unbookmark response:`, response);
      
      // Update bookmark state in the store
      tweetInteractionStore.updateTweetInteraction(tweetId, { is_bookmarked: false });
      toastStore.showToast('Bookmark removed', 'success');
    } catch (error) {
      console.error('Error removing bookmark:', error);
      toastStore.showToast('Failed to remove bookmark', 'error');
      // Revert the optimistic update
      tweetInteractionStore.updateTweetInteraction(tweetId, { is_bookmarked: true });
    }
  }
  
  // Load replies for a specific thread
  async function handleLoadReplies(event: CustomEvent) {
    const threadId = event.detail;
    logger.debug(`Loading replies for thread: ${threadId}`);
    await fetchRepliesForThread(threadId);
  }

  // Fix the onDestroy reference to use scrollListenerCleanup
  onDestroy(() => {
    // Clean up scroll listener when component is destroyed
    if (scrollListenerCleanup) scrollListenerCleanup();
  });
</script>

<MainLayout {username} {displayName} {avatar} on:toggleComposeModal={toggleComposeModal} on:posted={handleNewPost}>
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

    <div class="feed-content">
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
                {#if error}
                  API Connection Issue
                {:else}
                  No posts yet
                {/if}
              {:else}
                {#if error}
                  API Connection Issue
                {:else}
                  You're not following anyone yet
                {/if}
              {/if}
            </div>
            <div class="empty-state-message">
              {#if activeTab === 'for-you'}
                {#if error}
                  {error}
                  <div class="api-diagnosis">
                    Current API URL: {appConfig.api.baseUrl}
                    <br>
                    <span class="api-hint">Check that your Docker containers are running with <code>docker-compose ps</code></span>
                  </div>
                {:else}
                  Start the conversation by creating your first post
                {/if}
              {:else}
                {#if error}
                  {error}
                  <div class="api-diagnosis">
                    Current API URL: {appConfig.api.baseUrl}
                    <br>
                    <span class="api-hint">Check that your Docker containers are running with <code>docker-compose ps</code></span>
                  </div>
                {:else}
                  When you follow people, their posts will show up here
                {/if}
              {/if}
            </div>
            {#if error}
              <button 
                class="btn btn-primary" 
                on:click={() => activeTab === 'for-you' ? fetchTweetsForYou(true) : fetchTweetsFollowing(true)}
              >
                Try Again
              </button>
            {:else}
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
            {/if}
          </div>
        <!-- Tweets list -->
        {:else}
          {#each currentTweets as tweet, index (tweet.id || `tweet-${index}`)}
            {#if tweet.is_advertisement}
              <!-- Advertisement card using our CSS classes -->
              <div class="tweet-card">
                <div class="tweet-header">
                  <div class="tweet-avatar">
                    <img src={tweet.profile_picture_url} alt="Advertisement" />
                  </div>
                  <div class="tweet-user-info">
                    <span class="tweet-user-name">{tweet.name}</span>
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
                isAuth={authState.is_authenticated}
                isLiked={tweet.is_liked || false}
                isReposted={tweet.is_reposted || false}
                isBookmarked={tweet.is_bookmarked || false}
                on:reply={handleReply}
                on:repost={handleTweetRepost}
                on:unrepost={handleTweetUnrepost}
                on:like={handleLikeClick}
                on:unlike={handleUnlikeClick}
                on:bookmark={handleTweetBookmark}
                on:removeBookmark={handleTweetUnbookmark}
                on:loadReplies={handleLoadReplies}
                replies={repliesMap.get(tweet.id) || []}
                showReplies={false}
                nestedRepliesMap={nestedRepliesMap}
              />
            {/if}
          {/each}
          
          <!-- Loading indicator for infinite scrolling -->
          {#if loadingMoreTweets}
            <div class="feed-pagination">
              <div class="feed-loading">
                <div class="feed-loading-spinner"></div>
              </div>
            </div>
          {/if}
        {/if}
      </div>
    </div>
  </div>
</MainLayout>

<!-- Toast notifications -->
<Toast />

<!-- Debug panel -->
<DebugPanel />

<!-- Add ComposeTweetModal with selectedTweet -->
{#if showComposeModal}
  <ComposeTweet 
    avatar={avatar}
    isDarkMode={isDarkMode}
    parent_tweet={selectedTweet as ExtendedTweet}
    on:close={() => { 
      showComposeModal = false;
      selectedTweet = null;
    }}
    on:posted={selectedTweet ? handleReplyPosted : handleNewPost}
  />
{/if}

<style lang="css">
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
  
  /* Add missing feed styles */
  .feed-container {
    width: 100%;
    border-right: 1px solid var(--border-color, #e5e7eb);
    min-height: 100vh;
  }
  
  .feed-container.dark-theme {
    border-right: 1px solid var(--border-color-dark, #1e293b);
  }
  
  .feed-header {
    padding: 16px;
    position: sticky;
    top: 0;
    background-color: var(--bg-primary, #ffffff);
    backdrop-filter: blur(10px);
    z-index: 10;
    border-bottom: 1px solid var(--border-color, #e5e7eb);
  }
  
  .dark-theme .feed-header {
    background-color: var(--bg-primary-dark, #0f172a);
    border-bottom: 1px solid var(--border-color-dark, #1e293b);
  }
  
  .feed-title {
    font-size: 20px;
    font-weight: 700;
    margin: 0;
  }
  
  .feed-tabs {
    display: flex;
    border-bottom: 1px solid var(--border-color, #e5e7eb);
  }
  
  .dark-theme .feed-tabs {
    border-bottom: 1px solid var(--border-color-dark, #1e293b);
  }
  
  .feed-tab {
    flex: 1;
    text-align: center;
    padding: 16px 0;
    font-weight: 600;
    background: transparent;
    border: none;
    cursor: pointer;
    position: relative;
    color: var(--text-primary, #1f2937);
  }
  
  .dark-theme .feed-tab {
    color: var(--text-primary-dark, #f8fafc);
  }
  
  .feed-tab.active::after {
    content: '';
    position: absolute;
    bottom: 0;
    left: 50%;
    transform: translateX(-50%);
    width: 70px;
    height: 4px;
    background-color: var(--color-primary, #1d9bf0);
    border-radius: 9999px;
  }
  
  .feed-compose {
    display: flex;
    padding: 16px;
    border-bottom: 1px solid var(--border-color, #e5e7eb);
  }
  
  .dark-theme .feed-compose {
    border-bottom: 1px solid var(--border-color-dark, #1e293b);
  }
  
  .compose-avatar {
    width: 48px;
    height: 48px;
    border-radius: 50%;
    overflow: hidden;
    margin-right: 12px;
  }
  
  .compose-avatar img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }
  
  .compose-input-container {
    flex: 1;
  }
  
  .compose-input {
    width: 100%;
    border: none;
    background: transparent;
    padding: 12px 0;
    font-size: 20px;
    color: var(--text-secondary, #4b5563);
    text-align: left;
    cursor: pointer;
  }
  
  .dark-theme .compose-input {
    color: var(--text-secondary-dark, #94a3b8);
  }
  
  .compose-tools {
    display: flex;
    gap: 16px;
    margin-top: 12px;
  }
  
  .compose-tweet-tool {
    background: transparent;
    border: none;
    color: var(--color-primary, #1d9bf0);
    cursor: pointer;
    padding: 0;
    display: flex;
    align-items: center;
  }
  
  .feed-items {
    padding-bottom: 16px;
  }
  
  .feed-loading {
    display: flex;
    justify-content: center;
    padding: 32px 0;
  }
  
  .feed-loading-spinner {
    width: 32px;
    height: 32px;
    border: 3px solid rgba(29, 155, 240, 0.2);
    border-top-color: var(--color-primary, #1d9bf0);
    border-radius: 50%;
    animation: spinner 1s ease-in-out infinite;
  }
  
  @keyframes spinner {
    0% { transform: rotate(0deg); }
    100% { transform: rotate(360deg); }
  }
  
  .empty-state {
    padding: 48px 16px;
    text-align: center;
  }
  
  .empty-state-title {
    font-size: 20px;
    font-weight: 700;
    margin-bottom: 8px;
    color: var(--text-primary, #1f2937);
  }
  
  .dark-theme .empty-state-title {
    color: var(--text-primary-dark, #f8fafc);
  }
  
  .empty-state-message {
    font-size: 15px;
    color: var(--text-secondary, #4b5563);
    margin-bottom: 16px;
  }
  
  .dark-theme .empty-state-message {
    color: var(--text-secondary-dark, #94a3b8);
  }
  
  .btn {
    padding: 8px 16px;
    font-weight: 600;
    border-radius: 9999px;
    border: none;
    cursor: pointer;
  }
  
  .btn-primary {
    background-color: var(--color-primary, #1d9bf0);
    color: white;
  }
  
  .feed-pagination {
    padding: 20px 0;
    display: flex;
    justify-content: center;
    margin-top: 10px;
    margin-bottom: 30px;
  }
  
  .feed-loading-spinner {
    width: 32px;
    height: 32px;
    border: 3px solid rgba(29, 155, 240, 0.2);
    border-top-color: var(--color-primary, #1d9bf0);
    border-radius: 50%;
    animation: spinner 1s ease-in-out infinite;
  }
  
  @keyframes spinner {
    0% { transform: rotate(0deg); }
    100% { transform: rotate(360deg); }
  }
  
  .api-diagnosis {
    background-color: var(--bg-secondary);
    border-radius: var(--radius-md);
    padding: var(--space-2);
    margin-top: var(--space-2);
    font-size: var(--font-size-sm);
    color: var(--text-secondary);
    font-family: monospace;
  }
  
  .api-hint {
    display: block;
    margin-top: var(--space-2);
    font-style: italic;
  }
  
  .api-hint code {
    background-color: var(--bg-tertiary);
    padding: 2px 4px;
    border-radius: var(--radius-sm);
  }
</style>