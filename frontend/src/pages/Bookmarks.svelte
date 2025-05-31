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
  import SearchIcon from 'svelte-feather-icons/src/icons/SearchIcon.svelte';
  import XIcon from 'svelte-feather-icons/src/icons/XIcon.svelte';

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
  let searchInputElement: HTMLInputElement;
  
  // Filter bookmarks based on search query
  $: {
    if (searchQuery.trim() === '') {
      filteredBookmarks = bookmarkedTweets;
    } else {
      const query = searchQuery.toLowerCase();
      filteredBookmarks = bookmarkedTweets.filter(tweet => 
        tweet.content.toLowerCase().includes(query) || 
        (tweet.username && tweet.username.toLowerCase().includes(query)) || 
        (tweet.displayName && tweet.displayName.toLowerCase().includes(query))
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
    if (searchInputElement) {
      searchInputElement.focus();
    }
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
        
        logger.debug('User profile loaded', { username, displayName });
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
      
      if (response && response.threads && Array.isArray(response.threads)) {
        logger.info(`Received ${response.threads.length} bookmarks from API`);
        
        // Convert bookmarks to tweets format
        const convertedTweets = response.threads.map(bookmark => {
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

  onMount(() => {
    logger.debug('Bookmarks page mounted');
    if (checkAuth()) {
      // First load user profile
      fetchUserProfile()
        .then(() => {
          // Then load bookmarks
          fetchBookmarkedTweets();
        });
    }
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
