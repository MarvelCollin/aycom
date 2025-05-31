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
  import { getThreadsByUser, likeThread, unlikeThread, repostThread, bookmarkThread, removeBookmark, getAllThreads, getThreadReplies, getFollowingThreads, removeRepost } from '../api/thread';
  import { getTrends } from '../api/trends';
  import { getSuggestedUsers } from '../api/suggestions';
  import { createLoggerWithPrefix } from '../utils/logger';
  import { toastStore } from '../stores/toastStore';
  import { tweetInteractionStore } from '../stores/tweetStore';
  import { getProfile } from '../api/user';
  
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
  
  // Add this after the variable declarations
  let loadingTimeout: ReturnType<typeof setTimeout> | null = null;
  let fetchAttempts = {
    forYou: 0,
    following: 0
  };
  const maxFetchAttempts = 3;
  const loadingTimeoutDuration = 15000; // 15 seconds

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
  
  // Update onMount to use the cleanup function
  onMount(() => {
    console.log('Feed page - Auth state:', authState.is_authenticated, 'Current path:', window.location.pathname);
    
    // Let the Router handle redirects rather than doing it directly
    if (!authState.is_authenticated) {
      logger.info('User not authenticated, letting Router handle the redirect');
      return;
    }
    
    // Load user profile first
    fetchUserProfile();
    
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

  // Convert thread data to tweet format
  function threadToTweet(thread: any): ITweet {
    // Always log the raw thread data to diagnose issues
    console.log('Converting thread to tweet (raw data):', JSON.stringify(thread, null, 2));
    
    if (!thread || typeof thread !== 'object') {
      console.error('Invalid thread data received:', thread);
      // Return a placeholder tweet with error indicator
      return {
        id: `error-${Date.now()}`,
        content: 'Error loading tweet content',
        created_at: new Date().toISOString(),
        user_id: '',
        username: 'error',
        name: 'Error Loading Data',
        profile_picture_url: 'https://secure.gravatar.com/avatar/0?d=mp',
        likes_count: 0,
        replies_count: 0,
        reposts_count: 0,
        bookmark_count: 0,
        is_liked: false,
        is_reposted: false,
        is_bookmarked: false,
        is_pinned: false,
        parent_id: null,
        media: []
      };
    }

    // Extract user information with fallbacks for different backend formats
    const userId = thread.user_id || thread.UserID || thread.user?.id || thread.author_id || thread.user?.user_id || '';
    const username = thread.username || thread.Username || thread.user?.username || thread.author_username || 'anonymous';
    const displayName = thread.name || thread.DisplayName || thread.display_name || thread.user?.name || thread.user?.display_name || thread.authorName || 'User';
    const profilePic = formatProfilePicture(thread.profile_picture_url || thread.ProfilePicture || thread.user?.profile_picture_url || thread.author_avatar || thread.avatar);
    
    // Handle metrics with extensive fallbacks
    const likesCount = parseMetric(thread.likes_count || thread.LikesCount || thread.LikeCount || thread.like_count || thread.metrics?.likes);
    const repliesCount = parseMetric(thread.replies_count || thread.RepliesCount || thread.ReplyCount || thread.reply_count || thread.metrics?.replies);
    const repostsCount = parseMetric(thread.reposts_count || thread.RepostsCount || thread.RepostCount || thread.repost_count || thread.metrics?.reposts);
    const bookmarkCount = parseMetric(thread.bookmark_count || thread.BookmarkCount || thread.bookmarks_count || thread.view_count || thread.metrics?.bookmarks);
    
    // Handle interaction states with fallbacks
    const isLiked = Boolean(thread.is_liked || thread.IsLiked || thread.liked_by_user || thread.LikedByUser || false);
    const isReposted = Boolean(thread.is_reposted || thread.IsReposted || thread.reposted_by_user || thread.RepostedByUser || thread.is_repost || false);
    const isBookmarked = Boolean(thread.is_bookmarked || thread.IsBookmarked || thread.bookmarked_by_user || thread.BookmarkedByUser || false);
    const isPinned = Boolean(thread.is_pinned || thread.IsPinned || false);
    
    // Handle media array with type checking
    let mediaArray: IMedia[] = [];
    if (Array.isArray(thread.media)) {
      mediaArray = thread.media;
    } else if (Array.isArray(thread.Media)) {
      mediaArray = thread.Media;
    } else if (typeof thread.media === 'string') {
      try {
        // Sometimes the backend might send media as a JSON string
        const parsedMedia = JSON.parse(thread.media);
        if (Array.isArray(parsedMedia)) {
          mediaArray = parsedMedia;
        }
      } catch (e) {
        console.warn('Failed to parse media string:', thread.media);
      }
    }
    
    // Create the standardized tweet object
    const tweet: ITweet = {
      id: thread.id || '',
      content: thread.content || '',
      created_at: formatTimestamp(thread.created_at || thread.CreatedAt || new Date().toISOString()),
      updated_at: formatTimestamp(thread.updated_at || thread.UpdatedAt),
      
      user_id: userId,
      username: username,
      name: displayName,
      profile_picture_url: profilePic,
      
      likes_count: likesCount,
      replies_count: repliesCount,
      reposts_count: repostsCount,
      bookmark_count: bookmarkCount,
      views_count: parseMetric(thread.views_count || thread.view_count),
      
      media: mediaArray,
      
      is_liked: isLiked,
      is_reposted: isReposted,
      is_bookmarked: isBookmarked,
      is_pinned: isPinned,
      
      parent_id: thread.parent_id || thread.ParentID || null,
      
      community_id: thread.community_id || thread.CommunityID || null,
      community_name: thread.community_name || thread.CommunityName || null,
      
      is_advertisement: Boolean(thread.is_advertisement || thread.IsAdvertisement || false)
    };
    
    // Log the final converted tweet
    console.log('Converted tweet result:', tweet);
    return tweet;
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
      fetchAttempts.forYou++;
      isLoadingForYou = true;
      errorForYou = null;
      
      // Set a timeout to prevent infinite loading
      setLoadingTimeout('for-you');
      
      logger.info(`Fetching For You feed - page ${pageForYou}, limit ${limit}`);
      const response = await getAllThreads(pageForYou, limit);
      
      // Clear the timeout
      if (loadingTimeout) {
        clearTimeout(loadingTimeout);
        loadingTimeout = null;
      }
      
      // Handle explicit API errors
      if (response.error) {
        logger.error('Error fetching For You feed', { error: response.error });
        errorForYou = `Error loading tweets: ${response.error}`;
        
        // If we've reached max attempts, don't try again
        if (fetchAttempts.forYou >= maxFetchAttempts) {
          logger.warn(`Reached max fetch attempts (${maxFetchAttempts}) for For You feed`);
          isLoadingForYou = false;
          return;
        }
        
        // Try again after a delay if there was an error
        setTimeout(() => {
          fetchTweetsForYou(false);
        }, 3000);
        return;
      }
      
      // Reset attempt counter on success
      fetchAttempts.forYou = 0;
      
      // Validate threads array exists
      if (!response.threads) {
        logger.warn('Response missing threads array:', response);
        errorForYou = 'No threads received from server';
        isLoadingForYou = false;
        
        // Try again after a delay
        if (fetchAttempts.forYou < maxFetchAttempts) {
          setTimeout(() => {
            fetchTweetsForYou(false);
          }, 3000);
        }
        return;
      }
      
      // Ensure threads is an array
      const threads = Array.isArray(response.threads) ? response.threads : [];
      
      if (threads.length === 0) {
        logger.info('No threads found in For You feed');
        hasMoreForYou = false;
        isLoadingForYou = false;
        return;
      }
      
      // Process the response and update state
      const threadsMap = new Map();
      
      // First, convert all threads to tweets and create a map
      let convertedThreads = threads
        .filter(thread => thread && typeof thread === 'object') // Filter out invalid entries
        .map(thread => {
          console.log('Raw thread data from API:', thread);
          const tweet = threadToTweet(thread);
          console.log('Converted tweet:', tweet);
          // Use id instead of threadId since threadId doesn't exist on ITweet
          threadsMap.set(tweet.id, tweet);
          // Initialize the tweet in our global interaction store
          tweetInteractionStore.initTweet(tweet);
          return tweet;
        });
      
      // Next, add new tweets to our list
      tweetsForYou = [...tweetsForYou, ...convertedThreads];
      
      // Check if we have more tweets to fetch
      hasMoreForYou = convertedThreads.length >= limit;
      
      // Increment page for next fetch
      pageForYou++;
      
      logger.info(`Loaded ${convertedThreads.length} tweets for For You feed`);
      
    } catch (error) {
      logger.error('Error fetching For You feed', { error });
      errorForYou = 'Error loading tweets. Please try again.';
      
      // If we've reached max attempts, stop trying
      if (fetchAttempts.forYou >= maxFetchAttempts) {
        logger.warn(`Reached max fetch attempts (${maxFetchAttempts}) for For You feed`);
        isLoadingForYou = false;
        return;
      }
      
      // Try again after a delay
      setTimeout(() => {
        fetchTweetsForYou(false);
      }, 3000);
    } finally {
      isLoadingForYou = false;
      // Clear timeout if it exists
      if (loadingTimeout) {
        clearTimeout(loadingTimeout);
        loadingTimeout = null;
      }
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
      fetchAttempts.following++;
      isLoadingFollowing = true;
      errorFollowing = null;
      
      // Set a timeout to prevent infinite loading
      setLoadingTimeout('following');
      
      logger.info(`Fetching Following feed - page ${pageFollowing}, limit ${limit}`);
      const response = await getFollowingThreads(pageFollowing, limit);
      
      // Clear the timeout
      if (loadingTimeout) {
        clearTimeout(loadingTimeout);
        loadingTimeout = null;
      }
      
      if (response.error) {
        logger.error('Error fetching Following feed', { error: response.error });
        errorFollowing = `Error loading tweets: ${response.error}`;
        
        // If we've reached max attempts, don't try again
        if (fetchAttempts.following >= maxFetchAttempts) {
          logger.warn(`Reached max fetch attempts (${maxFetchAttempts}) for Following feed`);
          isLoadingFollowing = false;
          return;
        }
        
        // Try again after a delay if there was an error
        setTimeout(() => {
          fetchTweetsFollowing(false);
        }, 3000);
        return;
      }
      
      // Reset attempt counter on success
      fetchAttempts.following = 0;
      
      if (!response.threads || response.threads.length === 0) {
        logger.info('No threads found in Following feed');
        hasMoreFollowing = false;
        isLoadingFollowing = false;
        return;
      }
      
      // Process the response and update state
      const threadsMap = new Map();
      
      // First, convert all threads to tweets
      let convertedThreads = response.threads.map(thread => {
        const tweet = threadToTweet(thread);
        threadsMap.set(tweet.id, tweet);
        // Initialize the tweet in our global interaction store
        tweetInteractionStore.initTweet(tweet);
        return tweet;
      });
      
      // Next, add new tweets to our list
      tweetsFollowing = [...tweetsFollowing, ...convertedThreads];
      
      // Check if we have more tweets to fetch
      hasMoreFollowing = response.threads.length >= limit;
      
      // Increment page for next fetch
      pageFollowing++;
      
      logger.info(`Loaded ${convertedThreads.length} tweets for Following feed`);
      
    } catch (error) {
      logger.error('Error fetching Following feed', { error });
      errorFollowing = 'Error loading tweets. Please try again.';
    } finally {
      isLoadingFollowing = false;
      // Clear timeout if it exists
      if (loadingTimeout) {
        clearTimeout(loadingTimeout);
        loadingTimeout = null;
      }
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
          
          const convertedReply = threadToTweet(enrichedReply);
          
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
      // @ts-ignore - Using legacy property for backward compatibility
      if (repliesMap.has(threadId)) {
        // @ts-ignore - Using legacy property for backward compatibility
        const currentReplies = repliesMap.get(threadId) || [];
        const processedNewReply = threadToTweet(newReply);
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

  // Add this function to handle fetch timeout
  function setLoadingTimeout(type: 'for-you' | 'following') {
    // Clear any existing timeout
    if (loadingTimeout) {
      clearTimeout(loadingTimeout);
    }
    
    // Set a new timeout
    loadingTimeout = setTimeout(() => {
      if (type === 'for-you' && isLoadingForYou) {
        logger.warn('Loading timeout reached for "For You" feed');
        isLoadingForYou = false;
        errorForYou = 'Loading timed out. Please try again.';
      } else if (type === 'following' && isLoadingFollowing) {
        logger.warn('Loading timeout reached for "Following" feed');
        isLoadingFollowing = false;
        errorFollowing = 'Loading timed out. Please try again.';
      }
    }, loadingTimeoutDuration);
  }

  // Fix the onDestroy reference to use scrollListenerCleanup
  onDestroy(() => {
    // Clean up scroll listener when component is destroyed
    if (scrollListenerCleanup) scrollListenerCleanup();
    
    // Clean up loading timeout
    if (loadingTimeout) {
      clearTimeout(loadingTimeout);
    }
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
</style>