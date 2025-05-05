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
  import { getThreadsByUser, likeThread, unlikeThread, repostThread, bookmarkThread, removeBookmark, getAllThreads, getThreadReplies } from '../api/thread';
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

  // Add nestedRepliesMap to track replies at different levels
  let repliesMap = new Map();
  let nestedRepliesMap = new Map(); // For storing replies to replies

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
      likes: thread.like_count || thread.metrics?.likes || 0,
      replies: thread.reply_count || thread.metrics?.replies || 0,
      reposts: thread.repost_count || thread.metrics?.reposts || 0,
      bookmarks: thread.bookmark_count || (thread.view_count > 0 ? thread.view_count : 0) || thread.metrics?.bookmarks || 0,
      views: '0', // We're temporarily using view_count for bookmarks, so display 0 for now
      media: thread.media || [],
      isLiked: thread.is_liked || false,
      isReposted: thread.is_repost || false,
      isBookmarked: thread.is_bookmarked || false,
      replyTo: null // Will be populated later if this is a reply
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
        
        // Process threads to identify replies and link them to parent threads
        const threadsMap = new Map();
        
        // First, convert all threads to tweets and create a map
        const convertedThreads = response.threads.map(thread => {
          const tweet = threadToTweet(thread);
          threadsMap.set(tweet.threadId, tweet);
          return tweet;
        });
        
        // Then, process replies by linking them to parent threads
        for (const thread of response.threads) {
          if (thread.parent_thread_id && threadsMap.has(thread.parent_thread_id)) {
            const replyTweet = threadsMap.get(thread.thread_id || thread.id);
            const parentTweet = threadsMap.get(thread.parent_thread_id);
            
            if (replyTweet && parentTweet) {
              replyTweet.replyTo = parentTweet;
            }
          }
        }
        
        // If first page, replace tweets, otherwise append
        tweets = page === 1 ? convertedThreads : [...tweets, ...convertedThreads];
        
        // Check if there are more threads to load
        hasMore = convertedThreads.length === limit;
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
  
  // Update tweet reply count in the UI
  function updateTweetReplyCount(tweetId: string, count: number) {
    tweets = tweets.map(tweet => {
      if (String(tweet.id) === String(tweetId)) {
        return {
          ...tweet,
          replies: count
        };
      }
      return tweet;
    });
  }
  
  // Handle tweet actions
  function handleTweetClick(event: CustomEvent) {
    openThreadModal(event.detail);
  }
  
  // Find a tweet by ID in any of our data structures
  function findTweetById(tweetId: string | number): ITweet | undefined {
    // Convert to string for consistent comparison
    const id = String(tweetId);
    
    // First check main tweets array
    let foundTweet = tweets.find(t => String(t.id) === id);
    if (foundTweet) return foundTweet;
    
    // Check in replies
    for (const [parentId, replies] of repliesMap.entries()) {
      foundTweet = replies.find(r => String(r.id) === id);
      if (foundTweet) return foundTweet;
    }
    
    // Check in nested replies
    for (const [parentId, replies] of nestedRepliesMap.entries()) {
      foundTweet = replies.find(r => String(r.id) === id);
      if (foundTweet) return foundTweet;
    }
    
    return undefined;
  }
  
  async function handleTweetReply(event: CustomEvent) {
    const tweetId = event.detail;
    logger.info('Reply to tweet action', { tweetId });
    
    // Find the tweet to reply to
    const tweetToReply = findTweetById(tweetId);
    
    if (!tweetToReply) {
      logger.error('Tweet not found for reply', { tweetId });
      toastStore.showToast('Cannot find the tweet to reply to. Please try again.', 'error');
      
      // Try to load replies for this tweet ID in case it exists but we don't have it loaded
      if (typeof tweetId === 'string' || typeof tweetId === 'number') {
        try {
          await getThreadReplies(String(tweetId));
          // Retry finding the tweet after loading replies
          const retryFindTweet = findTweetById(tweetId);
          if (retryFindTweet) {
            selectedTweet = retryFindTweet;
            showComposeModal = true;
            return;
          }
        } catch (error) {
          logger.error('Failed to load replies for tweet', { tweetId, error });
        }
      }
      return;
    }
    
    // Store the tweet to reply to and open the compose modal
    selectedTweet = tweetToReply;
    showComposeModal = true;
  }
  
  async function handleTweetLike(event: CustomEvent) {
    const tweetId = event.detail;
    logger.info('Like tweet action', { tweetId });
    
    try {
      // Find the current tweet to check if it's already liked
      const currentTweet = tweets.find(t => t.id === tweetId);
      if (!currentTweet) {
        logger.error('Tweet not found for like action', { tweetId });
        return;
      }
      
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
  
  async function handleTweetUnlike(event: CustomEvent) {
    const tweetId = event.detail;
    logger.info('Unlike tweet action', { tweetId });
    
    try {
      // Find the current tweet
      const currentTweet = tweets.find(t => t.id === tweetId);
      if (!currentTweet) {
        logger.error('Tweet not found for unlike action', { tweetId });
        return;
      }
      
      // Call the unlike API
      await unlikeThread(tweetId);
      toastStore.showToast('Tweet unliked', 'success');
      
      // Update the likes count in the UI
      tweets = tweets.map(tweet => {
        if (tweet.id === tweetId) {
          return {
            ...tweet,
            likes: Math.max(0, (tweet.likes || 0) - 1), // Ensure likes don't go below 0
            isLiked: false
          };
        }
        return tweet;
      });
    } catch (error) {
      logger.error('Failed to unlike tweet', { error });
      toastStore.showToast('Failed to unlike tweet', 'error');
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
  
  async function handleTweetBookmark(event: CustomEvent) {
    const tweetId = event.detail;
    logger.info('Bookmark tweet action', { tweetId });
    
    try {
      // Find the current tweet
      const currentTweet = tweets.find(t => t.id === tweetId);
      if (!currentTweet) {
        logger.error('Tweet not found for bookmark action', { tweetId });
        return;
      }
      
      // Call the bookmark API
      await bookmarkThread(tweetId);
      toastStore.showToast('Tweet bookmarked', 'success');
      
      // Update the UI to show bookmarked state and update bookmark count
      tweets = tweets.map(tweet => {
        if (tweet.id === tweetId) {
          return {
            ...tweet,
            bookmarks: (tweet.bookmarks || 0) + 1,
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

  async function handleRemoveBookmark(event: CustomEvent) {
    const tweetId = event.detail;
    logger.info('Remove bookmark action', { tweetId });
    
    try {
      // Find the current tweet
      const currentTweet = tweets.find(t => t.id === tweetId);
      if (!currentTweet) {
        logger.error('Tweet not found for remove bookmark action', { tweetId });
        return;
      }
      
      // Call the remove bookmark API
      await removeBookmark(tweetId);
      toastStore.showToast('Bookmark removed', 'success');
      
      // Update the UI to show unbookmarked state and update bookmark count
      tweets = tweets.map(tweet => {
        if (tweet.id === tweetId) {
          return {
            ...tweet,
            bookmarks: Math.max(0, (tweet.bookmarks || 0) - 1), // Ensure bookmarks don't go below 0
            isBookmarked: false
          };
        }
        return tweet;
      });
    } catch (error) {
      logger.error('Failed to remove bookmark', { error });
      toastStore.showToast('Failed to remove bookmark', 'error');
    }
  }

  // Function to load nested replies
  async function loadNestedReplies(replyId: string) {
    try {
      // Check if we already have data for this reply (avoid duplicate requests)
      if (nestedRepliesMap.has(String(replyId)) && nestedRepliesMap.get(String(replyId))?.length > 0) {
        logger.debug(`Using cached nested replies for reply ${replyId}`);
        return; // Use cached data
      }

      // Set an empty array first to show loading state
      nestedRepliesMap.set(String(replyId), []);
      
      // Force a refresh to show loading state
      tweets = [...tweets];
      
      logger.info(`Loading nested replies for reply: ${replyId}`);
      
      // Call the API to get replies for this reply
      const response = await getThreadReplies(replyId);
      
      if (response && response.replies && response.replies.length > 0) {
        logger.info(`Received ${response.replies.length} nested replies from API for reply ${replyId}`);
        
        // Convert replies to tweet format
        const convertedReplies = response.replies.map(reply => threadToTweet(reply));
        
        // Store in the nested replies map (using string keys)
        nestedRepliesMap.set(String(replyId), convertedReplies);
        
        // Force a refresh to trigger UI update
        tweets = [...tweets];
      } else {
        logger.info(`No nested replies received for reply ${replyId}`);
        nestedRepliesMap.set(String(replyId), []);
        
        // Force a refresh to trigger UI update
        tweets = [...tweets];
      }
    } catch (error) {
      logger.error('Failed to load nested replies', { error, replyId });
      nestedRepliesMap.set(String(replyId), []);
      
      // Show error toast
      toastStore.showToast('Failed to load nested replies. Please try again.', 'error');
      
      // Force a refresh to update UI
      tweets = [...tweets];
    }
  }

  // Function to check if a thread ID exists within our nested replies structure
  function findThreadInNestedReplies(threadId: string): boolean {
    // First check in the main repliesMap
    for (const [parentId, replies] of repliesMap.entries()) {
      if (replies.some(reply => String(reply.id) === String(threadId))) {
        return true;
      }
    }
    
    // Then check in the nestedRepliesMap
    for (const [parentId, replies] of nestedRepliesMap.entries()) {
      if (replies.some(reply => String(reply.id) === String(threadId))) {
        return true;
      }
    }
    
    return false;
  }

  // Modify the handleLoadReplies function to check if it's a nested reply request
  async function handleLoadReplies(event: CustomEvent) {
    const threadId = event.detail;
    logger.info('Loading replies for thread', { threadId });
    
    // Check if this is a request for a reply that's already in our reply maps
    const isNestedReply = findThreadInNestedReplies(String(threadId));
    
    if (isNestedReply) {
      // Handle as a nested reply
      await loadNestedReplies(String(threadId));
      return;
    }
    
    try {
      // Set an empty array first to show loading state if we don't already have replies
      if (!repliesMap.has(String(threadId))) {
        repliesMap.set(String(threadId), []);
        // Force a refresh to show loading state
        tweets = [...tweets];
      }
      
      logger.info(`Fetching replies for thread ID: ${threadId}`);
      
      // Call the API to get replies for this thread
      const response = await getThreadReplies(threadId);
      
      if (response && response.replies && response.replies.length > 0) {
        logger.info(`Received ${response.replies.length} replies from API`);
        
        // Convert replies to tweet format
        const convertedReplies = response.replies.map(reply => threadToTweet(reply));
        
        // Store in the replies map using string keys
        repliesMap.set(String(threadId), convertedReplies);
        
        // Update the reply count in the UI
        updateTweetReplyCount(String(threadId), convertedReplies.length);
      } else {
        // Handle the case when there are no replies
        logger.info('No replies received from API');
        repliesMap.set(String(threadId), []);
        
        // Ensure the UI shows no replies
        updateTweetReplyCount(String(threadId), 0);
      }
      
      // Force a refresh of the tweets array to trigger UI update
      tweets = [...tweets];
    } catch (error) {
      logger.error('Failed to load replies', { error });
      toastStore.showToast('Failed to load replies. Please try again.', 'error');
      // Set empty array to avoid continuous loading attempts
      repliesMap.set(String(threadId), []);
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
            inReplyToTweet={tweet.replyTo || null}
            replies={repliesMap.get(tweet.id) || []}
            nestedRepliesMap={nestedRepliesMap}
            nestingLevel={0}
            on:click={handleTweetClick}
            on:like={handleTweetLike}
            on:unlike={handleTweetUnlike}
            on:repost={handleTweetRepost}
            on:reply={handleTweetReply}
            on:bookmark={handleTweetBookmark}
            on:removeBookmark={handleRemoveBookmark}
            on:loadReplies={handleLoadReplies}
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