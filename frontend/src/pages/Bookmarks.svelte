<script lang="ts">
  import MainLayout from '../components/layout/MainLayout.svelte';
  import TweetCard from '../components/social/TweetCard.svelte';
  import Toast from '../components/common/Toast.svelte';
  import DebugPanel from '../components/common/DebugPanel.svelte';
  import { onMount } from 'svelte';
  import { useAuth } from '../hooks/useAuth';
  import { useTheme } from '../hooks/useTheme';
  import type { ITweet, ITrend, ISuggestedFollow } from '../interfaces/ISocialMedia';
  import type { IAuthStore } from '../interfaces/IAuth';
  import { likeThread, unlikeThread, repostThread, bookmarkThread, removeBookmark, getThreadReplies, removeRepost, getUserBookmarks } from '../api/thread';
  import { getTrends } from '../api/trends';
  import { getSuggestedUsers } from '../api/suggestions';
  import { createLoggerWithPrefix } from '../utils/logger';
  import { toastStore } from '../stores/toastStore';
  import { getProfile } from '../api/user';

  const logger = createLoggerWithPrefix('Bookmarks');

  // Get auth store methods
  const { getAuthState } = useAuth();
  // Get theme store
  const { theme } = useTheme();

  // Reactive declarations for auth and theme
  $: authState = getAuthState ? (getAuthState() as IAuthStore) : { userId: null, isAuthenticated: false, accessToken: null, refreshToken: null };
  $: isDarkMode = $theme === 'dark';
  
  // User profile data
  let username = '';
  let displayName = '';
  let avatar = 'https://secure.gravatar.com/avatar/0?d=mp'; // Default avatar image URL
  let isLoadingProfile = true;

  // State for bookmarked tweets
  let bookmarkedTweets: ITweet[] = [];
  let isLoading = true;
  let error: string | null = null;
  let selectedTweet: ITweet | null = null;
  
  // Pagination
  let page = 1;
  let limit = 10;
  let hasMore = true;
  
  // Trends data
  let trends: ITrend[] = [];
  let isTrendsLoading = true;
  
  // Suggested users to follow
  let suggestedUsers: ISuggestedFollow[] = [];
  let isSuggestedUsersLoading = true;

  // Add nestedRepliesMap to track replies at different levels
  let repliesMap = new Map();
  let nestedRepliesMap = new Map(); // For storing replies to replies

  // State for search functionality
  let searchQuery = '';
  let filteredBookmarks: ITweet[] = [];
  
  // Filter bookmarks based on search query
  $: {
    if (searchQuery.trim() === '') {
      filteredBookmarks = bookmarkedTweets;
    } else {
      const query = searchQuery.toLowerCase();
      filteredBookmarks = bookmarkedTweets.filter(tweet => 
        tweet.content.toLowerCase().includes(query) || 
        tweet.username.toLowerCase().includes(query) || 
        tweet.displayName.toLowerCase().includes(query)
      );
    }
  }
  
  // Function to handle search input change
  function handleSearchInput(event: Event) {
    searchQuery = (event.target as HTMLInputElement).value;
  }

  // Clear search query
  function clearSearch() {
    searchQuery = '';
  }

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
      views: '0', // We're temporarily using view_count for bookmarks, so display 0 for now
      media: thread.media || [],
      isLiked: thread.is_liked || false,
      isReposted: thread.is_repost || false,
      isBookmarked: true, // Always true for bookmarks page
      replyTo: null, // Will be populated later if this is a reply
      isAdvertisement: thread.is_advertisement || false,
      communityId: thread.community_id || null,
      communityName: thread.community_name || null,
      // Include additional fields for replies
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

  // Function to fetch bookmarked tweets
  async function fetchBookmarkedTweets(resetPage = false) {
    logger.info('Fetching bookmarked tweets', { resetPage, page });
    
    if (resetPage) {
      page = 1;
      bookmarkedTweets = [];
    }
    
    isLoading = true;
    error = null;
    
    try {
      if (!checkAuth()) return;
      
      logger.debug('Fetching bookmarks');
      const response = await getUserBookmarks(page, limit);
      
      console.log('Bookmarks API response:', response);
      
      if (response && response.bookmarks && Array.isArray(response.bookmarks)) {
        logger.info(`Received ${response.bookmarks.length} bookmarks from API`);
        
        // Convert bookmarks to tweets format - bookmarks are now directly the threads
        const convertedTweets = response.bookmarks.map(bookmark => {
          // Each bookmark is already the thread data
          const tweet = threadToTweet(bookmark);
          return tweet;
        });
        
        // If first page, replace tweets, otherwise append
        bookmarkedTweets = page === 1 ? convertedTweets : [...bookmarkedTweets, ...convertedTweets];
        
        // Use pagination info from API if available
        if (response.pagination) {
          hasMore = response.pagination.hasMore;
          // Set page to next page value
          page = response.pagination.page + 1;
        } else {
          // Fallback to old logic
          hasMore = convertedTweets.length === limit;
          page++;
        }
        
        logger.debug('Updated bookmarks state', { 
          totalBookmarks: bookmarkedTweets.length, 
          hasMore, 
          nextPage: page 
        });
      } else {
        logger.info('No bookmarks received from API or invalid format');
        hasMore = false;
      }
    } catch (err) {
      console.error('Error loading bookmarks:', err);
      toastStore.showToast('Failed to load bookmarks. Please try again.', 'error');
      error = err instanceof Error ? err.message : 'Failed to fetch bookmarks';
    } finally {
      isLoading = false;
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
    console.log('Bookmarks page - Auth state:', authState.isAuthenticated, 'Current path:', window.location.pathname);
    
    // Let the Router handle redirects rather than doing it directly
    if (!authState.isAuthenticated) {
      logger.info('User not authenticated, letting Router handle the redirect');
      return;
    }
    
    // Load user profile first
    await fetchUserProfile();
    
    // Then fetch bookmarked tweets
    fetchBookmarkedTweets();
    
    // Fetch trends and suggestions in parallel
    Promise.all([
      fetchTrends(),
      fetchSuggestedUsers()
    ]).catch(error => {
      logger.error('Error fetching additional data:', error);
    });
  });

  function openThreadModal(tweet: ITweet) {
    logger.debug('Opening thread modal', { tweetId: tweet.id });
    selectedTweet = tweet;
  }
  
  function closeThreadModal() {
    logger.debug('Closing thread modal');
    selectedTweet = null;
  }
  
  // Handle tweet actions - updating the bookmarked tweets array
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
      
      // Update bookmarked tweets array
      bookmarkedTweets = bookmarkedTweets.map(tweet => {
        if (tweet.id === tweetId) {
          return { ...tweet, likes: (tweet.likes || 0) + 1, isLiked: true };
        }
        return tweet;
      });
    } catch (error) {
      toastStore.showToast('Failed to like tweet', 'error');
    }
  }
  
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
      
      // Update bookmarked tweets array
      bookmarkedTweets = bookmarkedTweets.map(tweet => {
        if (tweet.id === tweetId) {
          return { ...tweet, likes: Math.max(0, (tweet.likes || 0) - 1), isLiked: false };
        }
        return tweet;
      });
    } catch (error) {
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
    
    // Find the tweet in the array
    const tweetToReply = bookmarkedTweets.find(t => t.id === tweetId);
    
    if (!tweetToReply) {
      toastStore.showToast('Cannot find the tweet to reply to', 'error');
      return;
    }
    
    // Store the tweet to reply to and navigate to the thread page
    window.location.href = `/thread/${tweetId}`;
  }
  
  // Handle tweet repost
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
      
      // Update bookmarked tweets array
      bookmarkedTweets = bookmarkedTweets.map(tweet => {
        if (tweet.id === tweetId) {
          return { ...tweet, reposts: (tweet.reposts || 0) + 1, isReposted: true };
        }
        return tweet;
      });
    } catch (error) {
      toastStore.showToast('Failed to repost tweet', 'error');
    }
  }
  
  // Handle tweet unrepost
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
      
      // Update bookmarked tweets array
      bookmarkedTweets = bookmarkedTweets.map(tweet => {
        if (tweet.id === tweetId) {
          return { ...tweet, reposts: Math.max(0, (tweet.reposts || 0) - 1), isReposted: false };
        }
        return tweet;
      });
    } catch (error) {
      toastStore.showToast('Failed to remove repost', 'error');
    }
  }
  
  // Handle tweet unbookmark - remove from the list
  async function handleTweetUnbookmark(event: CustomEvent) {
    const tweetId = event.detail;
    if (!authState.isAuthenticated) {
      toastStore.showToast('You need to log in to remove bookmarks', 'warning');
      return;
    }
    logger.info('Unbookmark tweet action', { tweetId });
    
    try {
      await removeBookmark(tweetId);
      toastStore.showToast('Bookmark removed', 'success');
      
      // Remove the tweet from the bookmarked tweets array
      bookmarkedTweets = bookmarkedTweets.filter(tweet => tweet.id !== tweetId);
    } catch (error) {
      toastStore.showToast('Failed to remove bookmark', 'error');
    }
  }
  
  // Handle tweet bookmark - add bookmark to the backend
  async function handleTweetBookmark(event: CustomEvent) {
    const tweetId = event.detail;
    if (!authState.isAuthenticated) {
      toastStore.showToast('You need to log in to bookmark posts', 'warning');
      return;
    }
    logger.info('Bookmark tweet action', { tweetId });
    
    try {
      await bookmarkThread(tweetId);
      toastStore.showToast('Tweet bookmarked', 'success');
      
      // Update bookmarks count in the local state
      bookmarkedTweets = bookmarkedTweets.map(tweet => {
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
        console.log('Sample reply structure from API:', response.replies[0]);
        
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
  
  // Function to load more bookmarked tweets
  function loadMoreBookmarks() {
    fetchBookmarkedTweets();
  }
</script>

<MainLayout
  username={username}
  displayName={displayName}
  avatar={avatar}
  trends={trends}
  suggestedFollows={suggestedUsers}
>
  <div class="min-h-screen border-x feed-container">
    <div class="sticky top-0 z-10 header-tabs border-b {isDarkMode ? 'bg-black border-gray-800' : 'bg-white border-gray-200'}">
      <div class="p-4">
        <h1 class="text-xl font-bold">Bookmarks</h1>
        <p class="text-sm text-gray-500 dark:text-gray-400">Your saved tweets</p>
        
        <div class="mt-3 relative">
          <input
            type="text"
            placeholder="Search bookmarks..."
            value={searchQuery}
            on:input={handleSearchInput}
            class="w-full px-4 py-2 pl-10 border rounded-full {isDarkMode ? 'bg-gray-800 border-gray-700 text-white placeholder-gray-500' : 'bg-gray-100 border-gray-200 text-gray-800 placeholder-gray-500'} focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
          <div class="absolute left-3 top-2.5 text-gray-500">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
              <path fill-rule="evenodd" d="M8 4a4 4 0 100 8 4 4 0 000-8zM2 8a6 6 0 1110.89 3.476l4.817 4.817a1 1 0 01-1.414 1.414l-4.816-4.816A6 6 0 012 8z" clip-rule="evenodd" />
            </svg>
          </div>
          {#if searchQuery}
            <button 
              class="absolute right-3 top-2.5 text-gray-500 hover:text-gray-700 dark:hover:text-gray-300"
              on:click={clearSearch}
            >
              <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clip-rule="evenodd" />
              </svg>
            </button>
          {/if}
        </div>
      </div>
    </div>
    
    <div class="tweet-list">
      {#if isLoading && bookmarkedTweets.length === 0}
        <div class="space-y-4 p-4">
          {#each Array(5) as _, i}
            <div class="animate-pulse flex space-x-4">
              <div class="rounded-full bg-gray-300 dark:bg-gray-700 h-10 w-10"></div>
              <div class="flex-1 space-y-3 py-1">
                <div class="h-2 bg-gray-300 dark:bg-gray-700 rounded"></div>
                <div class="space-y-2">
                  <div class="h-2 bg-gray-300 dark:bg-gray-700 rounded"></div>
                  <div class="h-2 bg-gray-300 dark:bg-gray-700 rounded w-5/6"></div>
                </div>
                <div class="h-24 bg-gray-300 dark:bg-gray-700 rounded"></div>
              </div>
            </div>
          {/each}
        </div>
      {:else if error}
        <div class="p-8 text-center">
          <p class="text-red-500 mb-4">{error}</p>
          <button 
            class="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600" 
            on:click={() => fetchBookmarkedTweets(true)}
          >
            Try Again
          </button>
        </div>
      {:else if bookmarkedTweets.length === 0 && !isLoading}
        <div class="p-8 text-center text-gray-500 dark:text-gray-400">
          <p class="mb-4">You haven't bookmarked any tweets yet</p>
          <a 
            href="/explore" 
            class="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600 inline-block"
          >
            Explore Content
          </a>
        </div>
      {:else if searchQuery && filteredBookmarks.length === 0}
        <div class="p-8 text-center text-gray-500 dark:text-gray-400">
          <p class="mb-4">No bookmarks match your search</p>
          <button 
            class="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600" 
            on:click={clearSearch}
          >
            Clear Search
          </button>
        </div>
      {:else}
        {#each filteredBookmarks as tweet (tweet.id)}
          <TweetCard 
            {tweet}
            {isDarkMode}
            isAuthenticated={authState.isAuthenticated}
            isLiked={tweet.isLiked || false}
            isReposted={tweet.isReposted || false}
            isBookmarked={true}
            inReplyToTweet={tweet.replyTo || null}
            replies={repliesMap.get(tweet.id) || []}
            nestedRepliesMap={nestedRepliesMap}
            nestingLevel={0}
            on:click={() => openThreadModal(tweet)}
            on:like={handleTweetLike}
            on:unlike={handleTweetUnlike}
            on:repost={(e) => tweet.isReposted ? handleTweetUnrepost(e) : handleTweetRepost(e)}
            on:reply={handleTweetReply}
            on:bookmark={handleTweetBookmark}
            on:removeBookmark={handleTweetUnbookmark}
            on:loadReplies={handleLoadReplies}
          />
        {/each}
        
        {#if isLoading}
          <div class="flex justify-center items-center p-4">
            <div class="h-8 w-8 border-t-2 border-b-2 border-blue-500 rounded-full animate-spin"></div>
          </div>
        {:else if hasMore && !searchQuery}
          <div class="p-4 text-center">
            <button 
              class="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600" 
              on:click={loadMoreBookmarks}
            >
              Load More
            </button>
          </div>
        {/if}
      {/if}
    </div>
  </div>
</MainLayout>

<Toast />

<DebugPanel />

<style>
  :global(.theme-light) {
    --bg-primary: #ffffff;
    --bg-secondary: #f5f8fa;
    --bg-tertiary: #ebeef0;
    --bg-overlay: rgba(255, 255, 255, 0.9);
    --border-color: #e1e8ed;
    --text-primary: #14171a;
    --text-secondary: #657786;
    --text-tertiary: #aab8c2;
    --accent-color: #1d9bf0;
    --accent-hover: #1a8cd8;
    --error-color: #e0245e;
    --success-color: #17bf63;
  }

  :global(.theme-dark) {
    --bg-primary: #000000;
    --bg-secondary: #15181c;
    --bg-tertiary: #212327;
    --bg-overlay: rgba(0, 0, 0, 0.9);
    --border-color: #2f3336;
    --text-primary: #ffffff;
    --text-secondary: #8899a6;
    --text-tertiary: #66757f;
    --accent-color: #1d9bf0;
    --accent-hover: #1a8cd8;
    --error-color: #e0245e;
    --success-color: #17bf63;
  }

  .feed-container {
    background-color: var(--bg-primary);
    color: var(--text-primary);
    border-color: var(--border-color);
  }

  .header-tabs {
    background-color: var(--bg-overlay);
    backdrop-filter: blur(12px);
    border-color: var(--border-color);
  }

  .tweet-list {
    background-color: var(--bg-primary);
  }

  @keyframes pulse {
    0%, 100% { opacity: 0.5; }
    50% { opacity: 1; }
  }
  .animate-pulse {
    animation: pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite;
  }
</style>
