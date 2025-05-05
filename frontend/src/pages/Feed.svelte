<!-- This will act like Home Page - jantan lupa bro  -->

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
  import { getThreadsByUser, likeThread, unlikeThread, repostThread, bookmarkThread, removeBookmark, getAllThreads } from '../api/thread';
  import { getTrends } from '../api/trends';
  import { getSuggestedUsers } from '../api/suggestions';
  import { createLoggerWithPrefix } from '../utils/logger';
  import { toastStore } from '../stores/toastStore';

  const logger = createLoggerWithPrefix('Feed');

  export let route: string;

  // Get auth store methods
  const { getAuthState } = useAuth();
  // Get theme store
  const { theme } = useTheme();

  // Reactive declarations for auth and theme
  $: authState = getAuthState ? (getAuthState() as IAuthStore) : { userId: null, isAuthenticated: false, accessToken: null, refreshToken: null };
  $: isDarkMode = $theme === 'dark';
  // User information for sidebar
  $: sidebarUsername = authState?.userId ? `User_${authState.userId.substring(0, 4)}` : '';
  $: sidebarDisplayName = authState?.userId ? `User ${authState.userId.substring(0, 4)}` : '';
  $: sidebarAvatar = '👤'; // Placeholder avatar

  // State for tweets and compose modal
  let tweets: ITweet[] = [];
  let isLoading = true;
  let error: string | null = null;
  let showComposeModal: boolean = false;
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

  // Convert thread data to tweet format
  function threadToTweet(thread: any): ITweet {
    // Process content field if it contains user metadata
    // Format: [USER:username@displayName]content
    let username = 'anonymous';
    let displayName = 'User';
    let content = thread.content || '';
    let profilePicture = '';
    
    if (typeof content === 'string') {
      // Look for enhanced user metadata that includes profile picture
      // Format: [USER:username@displayName@profileUrl]content
      const enhancedMetadataRegex = /^\[USER:([^@\]]+)@([^@\]]+)@([^\]]+)\](.*)/;
      const match = enhancedMetadataRegex.exec(content);
      
      if (match) {
        username = match[1] || username;
        displayName = match[2] || displayName;
        profilePicture = match[3] || '';
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
      }
    } catch (error) {
      console.warn("Invalid date format in thread:", thread.created_at);
    }
    
    return {
      id: thread.id,
      threadId: thread.thread_id || thread.id,
      username: username,
      displayName: displayName,
      content: content,
      timestamp: timestamp,
      avatar: profilePicture || '👤', // Use real profile picture or fallback to emoji
      likes: thread.metrics?.likes || 0,
      replies: thread.metrics?.replies || 0,
      reposts: thread.metrics?.reposts || 0,
      views: thread.view_count?.toString() || '0',
      media: thread.media || [],
      isLiked: false,
      isReposted: false,
      isBookmarked: false
    };
  }

  async function fetchTweets(resetPage = false) {
    logger.info('Fetching tweets', { resetPage, page });
    
    if (resetPage) {
      page = 1;
      tweets = [];
    }
    
    // Remove the authentication check to allow viewing without login
    isLoading = true;
    error = null;
    
    try {
      let response;
      
      // If we're on the personal profile page and logged in, get only user's threads
      // Otherwise, get the global feed with all threads
      if (route === '/profile' && authState.isAuthenticated && authState.userId) {
        if (!authState.userId) {
          throw new Error('User ID is required for profile page');
        }
        logger.debug('Fetching user\'s own threads for profile page');
        response = await getThreadsByUser(authState.userId);
      } else if (route === '/profile' && !authState.isAuthenticated) {
        // Handle case where user tries to view profile but isn't logged in
        isLoading = false;
        error = 'Please log in to view your profile';
        return;
      } else {
        // For home/feed page, get all threads
        logger.debug('Fetching global feed for home page');
        response = await getAllThreads(page, limit);
      }
      
      if (response && response.threads) {
        logger.info(`Received ${response.threads.length} threads from API`);
        const newTweets = response.threads.map(threadToTweet);
        
        // If first page, replace tweets, otherwise append
        tweets = page === 1 ? newTweets : [...tweets, ...newTweets];
        
        // Check if there are more threads to load
        hasMore = newTweets.length === limit;
        page++;
        
        logger.debug('Updated tweets state', { 
          totalTweets: tweets.length, 
          hasMore, 
          nextPage: page 
        });
      } else {
        logger.info('No threads received from API');
        hasMore = false;
      }
    } catch (err) {
      console.error('Error loading feed:', err);
      toastStore.showToast('Failed to load feed. Please try again.', 'error');
      error = err instanceof Error ? err.message : 'Failed to fetch tweets';
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
    logger.info('Feed component mounted');
    fetchTweets();
    fetchTrends();
    fetchSuggestedUsers();
  });

  function toggleComposeModal() {
    logger.debug('Toggling compose modal', { currentState: showComposeModal });
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
    fetchTweets(true);
    toggleComposeModal();
  }
  
  // Handle tweet actions
  function handleTweetClick(event: CustomEvent) {
    openThreadModal(event.detail);
  }
  
  async function handleTweetLike(event: CustomEvent) {
    const tweetId = event.detail;
    logger.info('Like tweet action', { tweetId });
    
    try {
      // Call the like API
      await likeThread(tweetId);
      toastStore.showToast('Tweet liked', 'success');
      
      // Update the likes count in the UI
      tweets = tweets.map(tweet => {
        if (tweet.id === tweetId) {
          return {
            ...tweet,
            likes: (tweet.likes || 0) + 1,
            isLiked: true
          };
        }
        return tweet;
      });
    } catch (error) {
      logger.error('Failed to like tweet', { error });
      toastStore.showToast('Failed to like tweet', 'error');
    }
  }
  
  async function handleTweetRepost(event: CustomEvent) {
    const tweetId = event.detail;
    logger.info('Repost tweet action', { tweetId });
    
    try {
      // Call the repost API
      await repostThread(tweetId);
      toastStore.showToast('Tweet reposted', 'success');
      
      // Update the reposts count in the UI
      tweets = tweets.map(tweet => {
        if (tweet.id === tweetId) {
          return {
            ...tweet,
            reposts: (tweet.reposts || 0) + 1,
            isReposted: true
          };
        }
        return tweet;
      });
      
      // Refresh the feed to show the repost
      fetchTweets(true);
    } catch (error) {
      logger.error('Failed to repost tweet', { error });
      toastStore.showToast('Failed to repost tweet', 'error');
    }
  }
  
  async function handleTweetReply(event: CustomEvent) {
    const tweetId = event.detail;
    logger.info('Reply to tweet action', { tweetId });
    
    // Find the tweet to reply to
    const tweetToReply = tweets.find(t => t.id === tweetId);
    if (!tweetToReply) {
      logger.error('Tweet not found', { tweetId });
      return;
    }
    
    // Store the tweet to reply to and open the compose modal
    selectedTweet = tweetToReply;
    showComposeModal = true;
  }
  
  async function handleTweetBookmark(event: CustomEvent) {
    const tweetId = event.detail;
    logger.info('Bookmark tweet action', { tweetId });
    
    try {
      // Call the bookmark API
      await bookmarkThread(tweetId);
      toastStore.showToast('Tweet bookmarked', 'success');
      
      // Update the UI to show bookmarked state
      tweets = tweets.map(tweet => {
        if (tweet.id === tweetId) {
          return {
            ...tweet,
            isBookmarked: true
          };
        }
        return tweet;
      });
    } catch (error) {
      logger.error('Failed to bookmark tweet', { error });
      toastStore.showToast('Failed to bookmark tweet', 'error');
    }
  }
</script>

<MainLayout
  username={sidebarUsername}
  displayName={sidebarDisplayName}
  avatar={sidebarAvatar}
  trends={trends}
  suggestedFollows={suggestedUsers}
  on:toggleComposeModal={toggleComposeModal}
>
  <!-- Dynamic Content Area -->
  <div class="min-h-screen border-x border-gray-200 dark:border-gray-800">
    <!-- Dynamic Header based on route -->
    <div class="sticky top-0 z-10 bg-white/80 dark:bg-black/80 backdrop-blur-md border-b border-gray-200 dark:border-gray-800 px-4 py-3">
      <div class="flex justify-between items-center">
        {#if route === '/home' || route === '/feed'}
          <h1 class="text-xl font-bold">Home</h1>
        {:else if route === '/explore'}
          <h1 class="text-xl font-bold">Explore</h1>
        {:else if route === '/notifications'}
          <h1 class="text-xl font-bold">Notifications</h1>
        {:else if route === '/messages'}
          <h1 class="text-xl font-bold">Messages</h1>
        {:else if route === '/profile'}
          <h1 class="text-xl font-bold">Profile</h1>
        {/if}
      </div>
    </div>
    
    {#if route === '/home' || route === '/feed'}
      {#if !authState.isAuthenticated}
        <div class="p-4 mb-2 bg-blue-50 dark:bg-blue-900 rounded-md">
          <div class="flex items-center justify-between">
            <div>
              <h2 class="font-bold text-lg dark:text-white">Welcome to AYCOM!</h2>
              <p class="text-sm text-gray-600 dark:text-gray-300">Sign in to post and interact with threads.</p>
            </div>
            <div class="flex space-x-2">
              <a href="/login" class="px-4 py-2 bg-transparent border border-blue-500 text-blue-500 hover:bg-blue-500 hover:text-white rounded transition-colors">
                Log in
              </a>
              <a href="/register" class="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600 transition-colors">
                Sign up
              </a>
            </div>
          </div>
        </div>
      {:else if authState.isAuthenticated && sidebarDisplayName}
        <div class="p-4 mb-2">
          <div class="flex items-center">
            <div class="w-10 h-10 {isDarkMode ? 'bg-gray-700' : 'bg-gray-300'} rounded-full flex items-center justify-center mr-3 overflow-hidden flex-shrink-0">
              <span>{sidebarAvatar}</span>
            </div>
            <div>
              <h2 class="font-bold text-lg dark:text-white">Welcome, {sidebarDisplayName}!</h2>
              <p class="text-sm text-gray-600 dark:text-gray-300">We're glad to see you today. Here's your feed.</p>
            </div>
          </div>
        </div>
      {/if}
      
      {#if isLoading && tweets.length === 0}
        <div class="flex justify-center items-center p-8">
          <div class="spinner h-8 w-8 border-t-2 border-b-2 border-blue-500 rounded-full animate-spin"></div>
        </div>
      {:else if error}
        <div class="p-4 text-center text-red-500">
          <p>{error}</p>
          <button 
            class="mt-2 px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600" 
            on:click={() => fetchTweets(true)}
          >
            Try Again
          </button>
        </div>
      {:else if tweets.length === 0}
        <div class="p-8 text-center text-gray-500 dark:text-gray-400">
          <p class="mb-4">No posts yet</p>
          <button 
            class="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600" 
            on:click={toggleComposeModal}
          >
            Create First Post
          </button>
        </div>
      {:else}
        {#each tweets as tweet (tweet.id)}
          <TweetCard 
            {tweet}
            {isDarkMode}
            isAuthenticated={authState.isAuthenticated}
            isLiked={tweet.isLiked || false}
            isReposted={tweet.isReposted || false}
            isBookmarked={tweet.isBookmarked || false}
            on:click={handleTweetClick}
            on:like={handleTweetLike}
            on:repost={handleTweetRepost}
            on:reply={handleTweetReply}
            on:bookmark={handleTweetBookmark}
          />
        {/each}
        
        {#if isLoading}
          <div class="flex justify-center items-center p-4">
            <div class="spinner h-8 w-8 border-t-2 border-b-2 border-blue-500 rounded-full animate-spin"></div>
          </div>
        {:else if hasMore}
          <div class="p-4 text-center">
            <button 
              class="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600" 
              on:click={() => fetchTweets()}
            >
              Load More
            </button>
          </div>
        {/if}
      {/if}
    {:else if route === '/explore'}
      <!-- Explore Content -->
      <div class="p-4">
        <div class="bg-gray-100 dark:bg-gray-800 rounded-lg p-4">
          <h2 class="text-lg font-semibold mb-2">Trending Topics</h2>
          <p>Explore page content coming soon...</p>
        </div>
      </div>
    {:else if route === '/notifications'}
      <!-- Notifications Content -->
      <div class="p-4">
        <div class="bg-gray-100 dark:bg-gray-800 rounded-lg p-4">
          <h2 class="text-lg font-semibold mb-2">All Notifications</h2>
          <p>No notifications to display.</p>
        </div>
      </div>
    {:else if route === '/messages'}
      <!-- Messages Content -->
      <div class="p-4">
        <div class="bg-gray-100 dark:bg-gray-800 rounded-lg p-4">
          <h2 class="text-lg font-semibold mb-2">Messages</h2>
          <p>Your message inbox is empty.</p>
        </div>
      </div>
    {:else if route === '/profile'}
      <!-- Profile Content -->
      <div class="p-4">
        <div class="bg-gray-100 dark:bg-gray-800 rounded-lg p-4">
          <h2 class="text-lg font-semibold mb-2">Your Profile</h2>
          <p>Profile information will appear here.</p>
        </div>
      </div>
    {/if}
  </div>
</MainLayout>

{#if showComposeModal}
  <ComposeTweet 
    {isDarkMode}
    on:close={toggleComposeModal}
    on:tweet={handleNewTweet}
    avatar={sidebarAvatar}
    replyTo={selectedTweet}
  />
{/if}

<!-- Toast notifications -->
<Toast />

<!-- Debug panel -->
<DebugPanel />

<style>
  /* Spinner animation */
  @keyframes spin {
    0% { transform: rotate(0deg); }
    100% { transform: rotate(360deg); }
  }
  .animate-spin {
    animation: spin 1s linear infinite;
  }
</style>