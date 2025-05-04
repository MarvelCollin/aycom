<script lang="ts">
  import MainLayout from '../components/layout/MainLayout.svelte';
  import ComposeTweet from '../components/social/ComposeTweet.svelte';
  import TweetCard from '../components/social/TweetCard.svelte';
  import { onMount } from 'svelte';
  import { useAuth } from '../hooks/useAuth';
  import { useTheme } from '../hooks/useTheme';
  import type { ITweet, ITrend, ISuggestedFollow } from '../interfaces/ISocialMedia';
  import type { IAuthStore } from '../interfaces/IAuth';
  import { getThreadsByUser } from '../api/thread';
  import { getTrends } from '../api/trends';
  import { getSuggestedUsers } from '../api/suggestions';
  import { createLoggerWithPrefix } from '../utils/logger';
  import { toastStore } from '../stores/toastStore';

  // Create a logger for this component
  const logger = createLoggerWithPrefix('Feed');

  // Accept the route prop for conditionally rendering content
  export let route: string;

  // Get auth store methods - updated to include subscribe
  const { getAuthState, subscribe } = useAuth();
  // Get theme store
  const { theme } = useTheme();

  // Subscribe to auth store
  let authState: IAuthStore;
  subscribe(state => {
    authState = state;
  });

  // Reactive declarations for theme
  $: isDarkMode = $theme === 'dark';
  
  // User profile data
  let userDetails = {
    username: '',
    displayName: '',
    avatar: '👤',
    userId: '',
    email: '',
    isVerified: false,
    joinDate: ''
  };

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

  // Fetch user profile when authenticated
  async function fetchUserProfile() {
    if (!authState.isAuthenticated || !authState.accessToken) {
      return;
    }
    
    try {
      const apiUrl = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8081/api/v1';
      const token = authState.accessToken;
      
      logger.debug('Fetching user profile');
      const response = await fetch(`${apiUrl}/users/profile`, {
        headers: {
          'Authorization': `Bearer ${token}`
        }
      });
      
      if (response.ok) {
        const data = await response.json();
        if (data && data.user) {
          logger.info('User profile fetched successfully');
          userDetails = {
            username: data.user.username || `user_${authState.userId?.substring(0, 4)}`,
            displayName: data.user.displayName || `User ${authState.userId?.substring(0, 4)}`,
            avatar: data.user.profilePictureUrl || '👤',
            userId: data.user.id || authState.userId || '',
            email: data.user.email || '',
            isVerified: data.user.isVerified || false,
            joinDate: data.user.createdAt ? new Date(data.user.createdAt).toLocaleDateString() : ''
          };
        }
      } else {
        logger.warn('Failed to fetch user profile', { status: response.status });
        // Fallback to basic user info from auth state
        userDetails = {
          username: `user_${authState.userId?.substring(0, 4)}`,
          displayName: `User ${authState.userId?.substring(0, 4)}`,
          avatar: '👤',
          userId: authState.userId || '',
          email: '',
          isVerified: false,
          joinDate: ''
        };
      }
    } catch (err) {
      console.error('Error fetching user profile:', err);
      logger.error('Error fetching user profile', { error: err });
      toastStore.showToast('Failed to load user profile. Using default values.', 'warning');
      
      // Fallback to basic user info from auth state
      userDetails = {
        username: `user_${authState.userId?.substring(0, 4)}`,
        displayName: `User ${authState.userId?.substring(0, 4)}`,
        avatar: '👤',
        userId: authState.userId || '',
        email: '',
        isVerified: false,
        joinDate: ''
      };
    }
  }

  // Convert thread data to tweet format
  function threadToTweet(thread: any): ITweet {
    logger.debug('Converting thread to tweet', { threadId: thread.thread_id });
    return {
      id: thread.thread_id,
      username: thread.user?.username || `user_${thread.user_id.substring(0, 4)}`,
      displayName: thread.user?.display_name || `User ${thread.user_id.substring(0, 4)}`,
      avatar: thread.user?.avatar_url || '👤',
      content: thread.content,
      timestamp: thread.created_at,
      likes: thread.likes?.length || 0,
      replies: thread.replies?.length || 0,
      reposts: thread.reposts?.length || 0,
      views: thread.view_count?.toString() || '0',
      media: thread.media || []
    };
  }

  async function fetchTweets(resetPage = false) {
    logger.info('Fetching tweets', { resetPage, page });
    
    if (resetPage) {
      page = 1;
      tweets = [];
    }
    
    if (!authState.isAuthenticated || !authState.userId) {
      logger.warn('User not authenticated, aborting tweet fetch');
      isLoading = false;
      return;
    }
    
    isLoading = true;
    error = null;
    
    try {
      // Fetch threads from API
      logger.debug('Making API call to getThreadsByUser', { userId: 'me', page, limit });
      const response = await getThreadsByUser('me');
      
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
    if (authState.isAuthenticated) {
      await fetchUserProfile();
    }
    fetchTweets();
    fetchTrends();
    fetchSuggestedUsers();
  });

  function toggleComposeModal() {
    logger.debug('Toggling compose modal', { currentState: showComposeModal });
    showComposeModal = !showComposeModal;
    // Add a small delay before toggling to prevent render loops
    setTimeout(() => {
      showComposeModal = !showComposeModal;
    }, 0);
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
  
  function handleTweetLike(event: CustomEvent) {
    const tweetId = event.detail;
    logger.info('Like tweet action', { tweetId }, { showToast: true });
    // TODO: Implement like functionality with API
  }
  
  function handleTweetRepost(event: CustomEvent) {
    const tweetId = event.detail;
    logger.info('Repost tweet action', { tweetId }, { showToast: true });
    // TODO: Implement repost functionality with API
  }
  
  function handleTweetReply(event: CustomEvent) {
    const tweetId = event.detail;
    logger.info('Reply to tweet action', { tweetId }, { showToast: true });
    // TODO: Implement reply functionality with API
  }
</script>

<MainLayout
  username={userDetails.username}
  displayName={userDetails.displayName}
  avatar={userDetails.avatar}
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
      {#if authState.isAuthenticated && userDetails.displayName}
        <div class="p-4 bg-blue-50 dark:bg-blue-900/30 mb-2">
          <div class="flex items-center">
            <div class="w-10 h-10 {isDarkMode ? 'bg-gray-700' : 'bg-gray-300'} rounded-full flex items-center justify-center mr-3 overflow-hidden flex-shrink-0">
              {#if typeof userDetails.avatar === 'string' && userDetails.avatar.startsWith('http')}
                <img src={userDetails.avatar} alt={userDetails.username} class="w-full h-full object-cover" />
              {:else}
                <span>{userDetails.avatar}</span>
              {/if}
            </div>
            <div>
              <h2 class="font-bold text-lg dark:text-white">Welcome, {userDetails.displayName}!</h2>
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
            on:click={handleTweetClick}
            on:like={handleTweetLike}
            on:repost={handleTweetRepost}
            on:reply={handleTweetReply}
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
    on:close={() => { showComposeModal = false; }}
    on:tweet={handleNewTweet}
    avatar={userDetails.avatar} 
  />
{/if}

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