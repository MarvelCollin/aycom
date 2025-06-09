<script lang="ts">
  import MainLayout from '../components/layout/MainLayout.svelte';
  import TweetCard from '../components/social/TweetCard.svelte';
  import Toast from '../components/common/Toast.svelte';
  import DebugPanel from '../components/common/DebugPanel.svelte';
  import { onMount } from 'svelte';
  import { useAuth } from '../hooks/useAuth';
  import { useTheme } from '../hooks/useTheme';
  import type { ITweet, ISuggestedFollow } from '../interfaces/ISocialMedia';
  import type { ITrend } from '../interfaces/ITrend';
  import type { IMedia } from '../interfaces/IMedia';
  import type { IAuthStore } from '../interfaces/IAuth';
  import { likeThread, unlikeThread, repostThread, bookmarkThread, removeBookmark, getThreadReplies, removeRepost, getUserBookmarks, searchBookmarks } from '../api/thread';
  import { getTrends } from '../api/trends';
  import { getSuggestedUsers } from '../api/suggestions';
  import { createLoggerWithPrefix } from '../utils/logger';
  import { toastStore } from '../stores/toastStore';
  import { getProfile } from '../api/user';
  import { getUserId } from '../utils/auth';
  import SearchIcon from 'svelte-feather-icons/src/icons/SearchIcon.svelte';
  import XIcon from 'svelte-feather-icons/src/icons/XIcon.svelte';

  const logger = createLoggerWithPrefix('Bookmarks');

  // Get auth store methods
  const { getAuthState } = useAuth();
  // Get theme store
  const { theme } = useTheme();

  // Reactive declarations for auth and theme
  $: authState = getAuthState ? (getAuthState() as IAuthStore) : { user_id: null, is_authenticated: false, access_token: null, refresh_token: null };
  $: isDarkMode = $theme === 'dark';
  
  // User profile data
  let username = '';
  let name = '';
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
  let searchInputElement: HTMLInputElement;
  let isSearching = false;
  let searchDebounceTimer: number;
  
  // Filter bookmarks based on search query
  $: {
    if (searchQuery.trim() === '') {
      filteredBookmarks = bookmarkedTweets;
    } else {
      // Client-side filtering is only used when not actively searching
      // The actual search is performed by the API when the user finishes typing
      if (!isSearching) {
        const query = searchQuery.toLowerCase();
        filteredBookmarks = bookmarkedTweets.filter(tweet => 
          tweet.content.toLowerCase().includes(query) || 
          (tweet.username && tweet.username.toLowerCase().includes(query)) || 
          (tweet.name && tweet.name.toLowerCase().includes(query))
        );
      }
    }
  }
  
  // Function to handle search input change with debounce
  function handleSearchInput(event: Event) {
    const value = (event.target as HTMLInputElement).value;
    searchQuery = value;
    
    // Clear any existing timer
    if (searchDebounceTimer) {
      clearTimeout(searchDebounceTimer);
    }
    
    // Set a new timer for 500ms
    if (value.trim() !== '') {
      searchDebounceTimer = setTimeout(() => {
        performSearch(value);
      }, 500) as unknown as number;
    } else {
      // If search is cleared, reset to regular bookmarks
      fetchBookmarkedTweets(true);
    }
  }
  
  // Function to perform the actual search
  async function performSearch(query: string) {
    if (!query.trim()) return;
    
    isSearching = true;
    isLoading = true;
    error = null;
    
    try {
      if (!checkAuth()) return;
      
      logger.debug(`Searching bookmarks for: "${query}"`);
      const response = await searchBookmarks(query, 1, limit);
      
      if (response && response.threads && Array.isArray(response.threads)) {
        logger.info(`Received ${response.threads.length} search results from API`);
        
        // Format the search results as tweets
        const searchResults = response.threads.map(thread => {
          const tweet: ITweet = {
            id: thread.id,
            content: thread.content || '',
            created_at: thread.created_at || new Date().toISOString(),
            updated_at: thread.updated_at,
            user_id: thread.user_id,
            username: thread.username || 'anonymous',
            name: thread.name || 'User',
            profile_picture_url: thread.profile_picture_url || 'https://secure.gravatar.com/avatar/0?d=mp',
            likes_count: thread.likes_count || 0,
            replies_count: thread.replies_count || 0,
            reposts_count: thread.reposts_count || 0,
            bookmark_count: thread.bookmark_count || 0,
            is_liked: !!thread.is_liked,
            is_reposted: !!thread.is_reposted,
            is_bookmarked: true, // Always true for bookmarks page
            is_pinned: !!thread.is_pinned,
            parent_id: thread.parent_id || null,
            media: thread.media || []
          };
          return tweet;
        });
        
        // Update filtered bookmarks with search results
        filteredBookmarks = searchResults;
        
        // Update pagination info
        if (response.pagination) {
          hasMore = response.pagination.has_more;
        } else {
          hasMore = searchResults.length >= limit;
        }
      } else {
        logger.info('No search results received from API or invalid format');
        filteredBookmarks = [];
        hasMore = false;
      }
    } catch (err) {
      console.error('Error searching bookmarks:', err);
      toastStore.showToast('Failed to search bookmarks. Please try again.', 'error');
      error = err instanceof Error ? err.message : 'Failed to search bookmarks';
      filteredBookmarks = [];
    } finally {
      isLoading = false;
      isSearching = false;
    }
  }

  // Clear search query
  function clearSearch() {
    searchQuery = '';
    if (searchInputElement) {
      searchInputElement.focus();
    }
    
    // Reset to regular bookmarks
    fetchBookmarkedTweets(true);
  }

  // Authentication check
  function checkAuth() {
    if (!authState.is_authenticated) {
      logger.info('User not authenticated, redirecting to login page');
      
      // Only redirect if we're not already on the login page
      const currentPath = window.location.pathname;
      if (currentPath !== '/login') {
        logger.debug('Redirecting to login page');
        
        // Store the current path to redirect back after login
        try {
          localStorage.setItem('redirectAfterLogin', currentPath);
          logger.debug('Saved redirect path:', currentPath);
        } catch (err) {
          logger.error('Failed to save redirect path:', err);
        }
        
        window.location.href = '/login';
      }
      return false;
    }
    
    const userId = getUserId();
    if (!userId) {
      logger.warn('User authenticated but no user ID available, refreshing auth state');
      // Here you might want to refresh the auth state or redirect to login
      window.location.href = '/login';
      return false;
    }
    
    logger.debug('User authenticated with ID:', userId);
    return true;
  }
  
  // Fetch user profile data using the API directly
  async function fetchUserProfile() {
    isLoadingProfile = true;
    try {
      const response = await getProfile();
      if (response && response.user) {
        username = response.user.username || '';
        name = response.user.name || response.user.display_name || username;
        
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
        
        logger.debug('User profile loaded', { username, name });
      }
    } catch (err) {
      logger.error('Error loading user profile', err);
      toastStore.showToast('Failed to load your profile. Some features may be limited.', 'error');
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
      const response = await getUserBookmarks('me', page, limit);
      
      console.log('Bookmarks API response:', response);
      
      // Check if response has the expected structure
      if (!response) {
        logger.warn('Empty response from bookmarks API');
        hasMore = false;
        isLoading = false;
        return;
      }
      
      // Ensure bookmarks property exists, default to empty array if not
      const bookmarks = response.bookmarks || [];
      
      if (!Array.isArray(bookmarks)) {
        logger.warn('Bookmarks is not an array:', bookmarks);
        hasMore = false;
        isLoading = false;
        return;
      }
      
      logger.info(`Received ${bookmarks.length} bookmarks from API`);
      
      if (bookmarks.length === 0) {
        logger.info('Empty bookmarks array returned from API');
        hasMore = false;
        isLoading = false;
        return;
      }
      
      // Log the first bookmark for debugging
      if (bookmarks[0]) {
        logger.debug('First bookmark data structure:', JSON.stringify(bookmarks[0], null, 2));
      }
      
      // Use threads directly from API
      const receivedThreads = bookmarks.map(thread => {
        // Ensure all required ITweet fields are present
        const tweet: ITweet = {
          id: thread.id,
          content: thread.content || '',
          created_at: thread.created_at || new Date().toISOString(),
          updated_at: thread.updated_at,
          user_id: thread.user_id,
          username: thread.username || 'anonymous',
          name: thread.name || 'User',
          profile_picture_url: thread.profile_picture_url || 'https://secure.gravatar.com/avatar/0?d=mp',
          likes_count: thread.likes_count || 0,
          replies_count: thread.replies_count || 0,
          reposts_count: thread.reposts_count || 0,
          bookmark_count: thread.bookmark_count || 0,
          is_liked: !!thread.is_liked,
          is_reposted: !!thread.is_reposted,
          is_bookmarked: true, // Always true for bookmarks page
          is_pinned: !!thread.is_pinned,
          parent_id: thread.parent_id || null,
          media: thread.media || []
        };
        return tweet;
      });
      
      // If first page, replace tweets, otherwise append
      bookmarkedTweets = page === 1 ? receivedThreads : [...bookmarkedTweets, ...receivedThreads];
      filteredBookmarks = bookmarkedTweets;
      
      // Use pagination info from API if available
      if (response.pagination) {
        hasMore = response.pagination.has_more;
        // Set page to next page value
        page = response.pagination.current_page + 1;
        logger.debug('Pagination info from API', response.pagination);
      } else {
        // Fallback to old logic
        hasMore = receivedThreads.length >= limit;
        page++;
        logger.debug('Using fallback pagination logic', { hasMore, nextPage: page });
      }
      
      logger.debug('Updated bookmarks state', { 
        totalBookmarks: bookmarkedTweets.length, 
        hasMore, 
        nextPage: page 
      });
    } catch (err) {
      console.error('Error loading bookmarks:', err);
      toastStore.showToast('Failed to load bookmarks. Please try again.', 'error');
      error = err instanceof Error ? err.message : 'Failed to fetch bookmarks';
      logger.error('Bookmarks error details:', { error, userId: authState?.user_id });
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

  onMount(() => {
    logger.debug('Bookmarks page mounted');
    
    if (!checkAuth()) {
      logger.warn('Authentication check failed in onMount');
      return;
    }
    
    logger.debug('Auth check passed, loading profile and bookmarks');
    
    // First load user profile
    fetchUserProfile()
      .then(() => {
        // Then load bookmarks
        logger.debug('Profile loaded, fetching bookmarks');
        return fetchBookmarkedTweets(true);
      })
      .catch(error => {
        logger.error('Error in component initialization:', error);
        toastStore.showToast('Error loading your data. Please refresh the page.', 'error');
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
    if (!authState.is_authenticated) {
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
          return { ...tweet, likes_count: (tweet.likes_count || 0) + 1, is_liked: true };
        }
        return tweet;
      });
    } catch (error) {
      toastStore.showToast('Failed to like tweet', 'error');
    }
  }
  
  async function handleTweetUnlike(event: CustomEvent) {
    const tweetId = event.detail;
    if (!authState.is_authenticated) {
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
          return { ...tweet, likes_count: Math.max(0, (tweet.likes_count || 0) - 1), is_liked: false };
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
    if (!authState.is_authenticated) {
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
    if (!authState.is_authenticated) {
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
          return { ...tweet, reposts_count: (tweet.reposts_count || 0) + 1, is_reposted: true };
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
    if (!authState.is_authenticated) {
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
          return { ...tweet, reposts_count: Math.max(0, (tweet.reposts_count || 0) - 1), is_reposted: false };
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
    if (!authState.is_authenticated) {
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
    if (!authState.is_authenticated) {
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
          return { ...tweet, bookmark_count: (tweet.bookmark_count || 0) + 1, is_bookmarked: true };
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
          
          // Map directly to ITweet format
          const convertedReply: ITweet = {
            id: replyData.id,
            content: replyData.content || '',
            created_at: replyData.created_at || new Date().toISOString(),
            user_id: userData.id || replyData.user_id,
            username: userData.username || reply.author_username || 'anonymous',
            name: userData.name || reply.author_name || 'User',
            profile_picture_url: userData.profile_picture_url || reply.author_avatar || 'https://secure.gravatar.com/avatar/0?d=mp',
            likes_count: reply.likes_count || 0,
            replies_count: 0, // Replies to replies not tracked yet
            reposts_count: 0,
            bookmark_count: 0,
            is_liked: false,
            is_reposted: false,
            is_bookmarked: false,
            is_pinned: false,
            parent_id: threadId,
            media: []
          };
          
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

  // Function to handle reply posted event
  function handleReplyPosted(event) {
    // @ts-ignore - Using legacy property for backward compatibility
    const { threadId, newReply } = event.detail;
    logger.info('Reply posted', { threadId });
    
    // Find the tweet that was replied to
    // @ts-ignore - Using legacy property for backward compatibility
    const repliedTweet = bookmarkedTweets.find(t => String(t.id) === String(threadId));
                         
    if (repliedTweet) {
      // Increment the reply count
      repliedTweet.replies_count = (parseInt(String(repliedTweet.replies_count)) || 0) + 1;
      
      // Add the reply to our replies map if it exists
      // @ts-ignore - Using legacy property for backward compatibility
      if (repliesMap.has(threadId)) {
        // @ts-ignore - Using legacy property for backward compatibility
        const currentReplies = repliesMap.get(threadId) || [];
        
        // Map the reply directly to ITweet format
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
          is_liked: false,
          is_reposted: false,
          is_bookmarked: false,
          is_pinned: false,
          parent_id: threadId,
          media: newReply.media || []
        };
        
        // @ts-ignore - Using legacy property for backward compatibility
        repliesMap.set(threadId, [processedNewReply, ...currentReplies]);
        repliesMap = repliesMap; // Trigger reactivity
      }
    }
  }
</script>

<MainLayout
  username={username}
  displayName={name}
  avatar={avatar}
>
  <div class="bookmarks-container">
    <!-- Header -->
    <div class="bookmarks-header">
      <h1 class="page-title">Bookmarks</h1>
      <p class="page-subtitle">@{username}</p>
      
      <!-- Search box -->
      <div class="search-container">
        <div class="search-input-wrapper">
          <div class="search-icon">
            <SearchIcon size="18" />
          </div>
          <input
            bind:this={searchInputElement}
            type="text"
            placeholder="Search your bookmarks"
            class="search-input"
            value={searchQuery}
            on:input={handleSearchInput}
          />
          {#if searchQuery}
            <button class="search-clear-button" on:click={clearSearch} aria-label="Clear search">
              <XIcon size="18" />
            </button>
          {/if}
        </div>
      </div>
    </div>
    
    <!-- Content Area -->
    <div class="bookmarks-content">
      {#if isLoading}
        <div class="bookmarks-loading">
          <div class="bookmarks-skeleton">
            {#each Array(3) as _}
              <div class="tweet-skeleton">
                <div class="tweet-skeleton-avatar"></div>
                <div class="tweet-skeleton-content">
                  <div class="tweet-skeleton-header"></div>
                  <div class="tweet-skeleton-body"></div>
                  <div class="tweet-skeleton-actions"></div>
                </div>
              </div>
            {/each}
          </div>
        </div>
      {:else if error}
        <div class="bookmarks-error">
          <p class="error-message">{error}</p>
          <button class="retry-button" on:click={() => fetchBookmarkedTweets(true)}>Try Again</button>
        </div>
      {:else if filteredBookmarks.length === 0}
        <div class="bookmarks-empty">
          {#if searchQuery}
            <h2 class="empty-title">No results found</h2>
            <p class="empty-message">We couldn't find any results for "{searchQuery}"</p>
            <button class="clear-search-button" on:click={clearSearch}>Clear search</button>
          {:else}
            <h2 class="empty-title">You haven't saved any posts yet</h2>
            <p class="empty-message">When you do, they'll show up here.</p>
          {/if}
        </div>
      {:else}
        <div class="bookmarks-list">
          {#each filteredBookmarks as tweet}
            <div class="bookmark-item {isDarkMode ? 'bookmark-item-dark' : ''}">
              <TweetCard 
                {tweet}
                isDarkMode={isDarkMode}
                on:like={async (e) => handleTweetLike(e)}
                on:unlike={async (e) => handleTweetUnlike(e)}
                on:repost={async (e) => handleTweetRepost(e)}
                on:unrepost={async (e) => handleTweetUnrepost(e)}
                on:bookmark={async (e) => handleTweetBookmark(e)}
                on:unbookmark={async (e) => handleTweetUnbookmark(e)}
                on:reply={async (e) => handleTweetReply(e)}
                on:viewReplies={async (e) => handleLoadReplies(e)}
                on:showProfile={(e) => openThreadModal(e.detail)}
              />
            </div>
          {/each}
          
          {#if hasMore}
            <div class="load-more-container">
              <button 
                class="load-more-button {isDarkMode ? 'load-more-button-dark' : ''}"
                on:click={loadMoreBookmarks}
                disabled={isLoading}
              >
                {isLoading ? 'Loading...' : 'Load More'}
              </button>
            </div>
          {/if}
        </div>
      {/if}
    </div>
  </div>
</MainLayout>

<Toast />

<DebugPanel />

<style>
  .bookmarks-container {
    min-height: 100vh;
    background-color: var(--bg-primary);
    color: var(--text-primary);
    border-left: 1px solid var(--border-color);
    border-right: 1px solid var(--border-color);
  }
  
  .bookmarks-container-dark {
    background-color: var(--bg-primary-dark);
    color: var(--text-primary-dark);
    border-left: 1px solid var(--border-color-dark);
    border-right: 1px solid var(--border-color-dark);
  }
  
  .bookmarks-header {
    position: sticky;
    top: 0;
    z-index: var(--z-sticky);
    padding: var(--space-3) var(--space-4);
    background-color: var(--bg-primary);
    backdrop-filter: blur(12px);
    -webkit-backdrop-filter: blur(12px);
    border-bottom: 1px solid var(--border-color);
  }
  
  .bookmarks-header-dark {
    background-color: var(--bg-primary-dark);
    border-bottom: 1px solid var(--border-color-dark);
  }
  
  .page-title {
    font-size: var(--font-size-xl);
    font-weight: var(--font-weight-bold);
    margin-bottom: var(--space-3);
  }
  
  .page-subtitle {
    font-size: var(--font-size-md);
    color: var(--text-secondary);
  }
  
  .search-container {
    margin-top: var(--space-3);
    margin-bottom: var(--space-1);
  }
  
  .search-input-wrapper {
    position: relative;
  }
  
  .search-input {
    width: 100%;
    padding: var(--space-2) var(--space-4) var(--space-2) var(--space-10);
    border-radius: var(--radius-full);
    border: 1px solid var(--border-color);
    background-color: var(--bg-tertiary);
    color: var(--text-primary);
    font-size: var(--font-size-md);
    transition: all var(--transition-fast);
  }
  
  .search-input-dark {
    border-color: var(--border-color-dark);
    background-color: var(--bg-tertiary-dark);
    color: var(--text-primary-dark);
  }
  
  .search-input:focus {
    outline: none;
    background-color: var(--bg-primary);
    border-color: var(--color-primary);
  }
  
  .search-icon {
    position: absolute;
    left: var(--space-3);
    top: 50%;
    transform: translateY(-50%);
    color: var(--text-tertiary);
    background: none;
    border: none;
    display: flex;
    align-items: center;
    justify-content: center;
  }
  
  .search-icon-dark {
    color: var(--text-tertiary-dark);
  }
  
  .search-icon {
    width: 18px;
    height: 18px;
  }
  
  .search-clear-button {
    position: absolute;
    right: var(--space-3);
    top: 50%;
    transform: translateY(-50%);
    color: var(--text-tertiary);
    background: none;
    border: none;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
  }
  
  .search-clear-button-dark {
    color: var(--text-tertiary-dark);
  }
  
  .clear-icon {
    width: 18px;
    height: 18px;
  }
  
  .bookmarks-content {
    padding-bottom: var(--space-6);
  }
  
  .bookmarks-loading {
    padding: var(--space-4);
  }
  
  .bookmarks-skeleton {
    display: flex;
    flex-direction: column;
    gap: var(--space-4);
  }
  
  .tweet-skeleton {
    display: flex;
    gap: var(--space-3);
    padding: var(--space-4);
    border-bottom: 1px solid var(--border-color);
  }
  
  .tweet-skeleton-avatar {
    width: 48px;
    height: 48px;
    border-radius: 50%;
    background-color: var(--bg-tertiary);
    flex-shrink: 0;
  }
  
  .tweet-skeleton-content {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
  }
  
  .tweet-skeleton-header {
    height: 20px;
    background-color: var(--bg-tertiary);
    border-radius: var(--radius-sm);
    width: 60%;
  }
  
  .tweet-skeleton-body {
    height: 60px;
    background-color: var(--bg-tertiary);
    border-radius: var(--radius-sm);
    width: 100%;
  }
  
  .tweet-skeleton-actions {
    height: 20px;
    background-color: var(--bg-tertiary);
    border-radius: var(--radius-sm);
    width: 80%;
  }
  
  .bookmarks-error {
    padding: var(--space-8);
    text-align: center;
  }
  
  .error-message {
    color: var(--color-error);
    margin-bottom: var(--space-4);
  }
  
  .retry-button {
    padding: var(--space-2) var(--space-4);
    background-color: var(--color-primary);
    color: white;
    border: none;
    border-radius: var(--radius-full);
    cursor: pointer;
    font-weight: var(--font-weight-medium);
  }
  
  .bookmarks-empty {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: var(--space-16) var(--space-4);
    text-align: center;
  }
  
  .empty-title {
    font-size: var(--font-size-2xl);
    font-weight: var(--font-weight-bold);
    margin-bottom: var(--space-2);
  }
  
  .empty-message {
    color: var(--text-secondary);
    margin-bottom: var(--space-4);
    max-width: 400px;
  }
  
  .clear-search-button {
    padding: var(--space-2) var(--space-4);
    background-color: var(--color-primary);
    color: white;
    border: none;
    border-radius: var(--radius-full);
    cursor: pointer;
    font-weight: var(--font-weight-medium);
  }
  
  .bookmarks-list {
    display: flex;
    flex-direction: column;
  }
  
  .bookmark-item {
    border-bottom: 1px solid var(--border-color);
  }
  
  .bookmark-item-dark {
    border-bottom: 1px solid var(--border-color-dark);
  }
  
  .load-more-container {
    padding: var(--space-4);
    display: flex;
    justify-content: center;
  }
  
  .load-more-button {
    padding: var(--space-2) var(--space-6);
    background-color: var(--color-primary);
    color: white;
    border: none;
    border-radius: var(--radius-full);
    cursor: pointer;
    font-weight: var(--font-weight-medium);
  }
  
  .load-more-button:disabled {
    opacity: 0.7;
    cursor: not-allowed;
  }
  
  .load-more-button-dark {
    background-color: var(--color-primary-dark);
  }
  
  @keyframes pulse {
    0%, 100% { opacity: 0.5; }
    50% { opacity: 0.8; }
  }
  
  .animate-pulse {
    animation: pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite;
  }
</style>
