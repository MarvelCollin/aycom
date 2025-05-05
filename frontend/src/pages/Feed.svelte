<!-- This will act like Home Page -->

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
  import { getThreadsByUser, likeThread, unlikeThread, repostThread, bookmarkThread, removeBookmark, getAllThreads, getThreadReplies, getFollowingThreads } from '../api/thread';
  import { getTrends } from '../api/trends';
  import { getSuggestedUsers } from '../api/suggestions';
  import { createLoggerWithPrefix } from '../utils/logger';
  import { toastStore } from '../stores/toastStore';

  const logger = createLoggerWithPrefix('Feed');

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
  let tweetsForYou: ITweet[] = [];
  let tweetsFollowing: ITweet[] = [];
  let isLoadingForYou = true;
  let isLoadingFollowing = true;
  let errorForYou: string | null = null;
  let errorFollowing: string | null = null;
  let showComposeModal: boolean = false;
  let selectedTweet: ITweet | null = null;
  
  // Tab state
  let activeTab: 'for-you' | 'following' = 'for-you';
  
  // Pagination for both tabs
  let pageForYou = 1;
  let pageFollowing = 1;
  let limit = 10;
  let hasMoreForYou = true;
  let hasMoreFollowing = true;
  
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
      replyTo: null, // Will be populated later if this is a reply
      isAdvertisement: thread.is_advertisement || false,
      communityId: thread.community_id || null,
      communityName: thread.community_name || null
    };
  }

  // Authentication check
  function checkAuth() {
    if (!authState.isAuthenticated) {
      toastStore.showToast('You need to log in to access the feed', 'warning');
      return false;
    }
    return true;
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
        const tweetsWithAds = [];
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
              avatar: '📢',
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
              isAdvertisement: true
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
        const tweetsWithAds = [];
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
              avatar: '📢',
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
              isAdvertisement: true
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
    logger.info('Feed component mounted');
    if (authState.isAuthenticated) {
      fetchTweetsForYou();
      fetchTrends();
      fetchSuggestedUsers();
    }
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
  
  // Handle tweet actions - simplified versions that update both feed arrays
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
          return { ...tweet, likes: (tweet.likes || 0) + 1, isLiked: true };
        }
        return tweet;
      });
      
      tweetsFollowing = tweetsFollowing.map(tweet => {
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
      
      // Update both tweet arrays
      tweetsForYou = tweetsForYou.map(tweet => {
        if (tweet.id === tweetId) {
          return { ...tweet, likes: Math.max(0, (tweet.likes || 0) - 1), isLiked: false };
        }
        return tweet;
      });
      
      tweetsFollowing = tweetsFollowing.map(tweet => {
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
  
  // Other tweet action handlers would be similar - updating both arrays
  // For brevity, I'm not including all of them

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

<MainLayout
  username={sidebarUsername}
  displayName={sidebarDisplayName}
  avatar={sidebarAvatar}
  trends={trends}
  suggestedFollows={suggestedUsers}
  on:toggleComposeModal={toggleComposeModal}
>
  <!-- Content Area -->
  <div class="min-h-screen border-x border-gray-200 dark:border-gray-800">
    <!-- Header with Tabs -->
    <div class="sticky top-0 z-10 bg-white/80 dark:bg-black/80 backdrop-blur-md border-b border-gray-200 dark:border-gray-800">
      <!-- Tabs -->
      <div class="flex justify-between">
        <button 
          class="flex-1 py-4 text-center font-medium {activeTab === 'for-you' ? 'text-primary border-b-2 border-primary' : 'text-gray-500 dark:text-gray-400'}"
          on:click={() => handleTabChange('for-you')}
        >
          For you
        </button>
        <button 
          class="flex-1 py-4 text-center font-medium {activeTab === 'following' ? 'text-primary border-b-2 border-primary' : 'text-gray-500 dark:text-gray-400'}"
          on:click={() => handleTabChange('following')}
        >
          Following
        </button>
      </div>
    </div>
    
    <!-- Authentication Check Banner -->
    {#if !authState.isAuthenticated}
      <div class="p-4 mb-2 bg-blue-50 dark:bg-blue-900/30 rounded-md mx-4 mt-4">
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
    {/if}
    
    <!-- Compose Tweet Form - Only visible for authenticated users -->
    {#if authState.isAuthenticated}
      <div class="p-4 border-b border-gray-200 dark:border-gray-800">
        <div class="flex">
          <div class="w-10 h-10 {isDarkMode ? 'bg-gray-700' : 'bg-gray-200'} rounded-full flex items-center justify-center mr-3 overflow-hidden flex-shrink-0">
            <span>{sidebarAvatar}</span>
          </div>
          <div class="flex-1">
            <div 
              class="w-full min-h-[40px] px-4 py-2 rounded-3xl border border-gray-300 dark:border-gray-700 text-gray-500 dark:text-gray-400 cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-800"
              on:click={toggleComposeModal}
            >
              What's happening?
            </div>
            <div class="flex mt-2 -ml-2">
              <button class="p-2 text-primary rounded-full hover:bg-primary/10">
                <span class="sr-only">Add image</span>
                <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="3" width="18" height="18" rx="2" ry="2"></rect><circle cx="8.5" cy="8.5" r="1.5"></circle><polyline points="21 15 16 10 5 21"></polyline></svg>
              </button>
              <button class="p-2 text-primary rounded-full hover:bg-primary/10">
                <span class="sr-only">Add gif</span>
                <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path><polyline points="14 2 14 8 20 8"></polyline><line x1="16" y1="13" x2="8" y2="13"></line><line x1="16" y1="17" x2="8" y2="17"></line><polyline points="10 9 9 9 8 9"></polyline></svg>
              </button>
              <button class="p-2 text-primary rounded-full hover:bg-primary/10">
                <span class="sr-only">Add poll</span>
                <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="4 17 10 11 4 5"></polyline><line x1="12" y1="19" x2="20" y2="19"></line></svg>
              </button>
              <button class="p-2 text-primary rounded-full hover:bg-primary/10">
                <span class="sr-only">Add emoji</span>
                <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"></circle><path d="M8 14s1.5 2 4 2 4-2 4-2"></path><line x1="9" y1="9" x2="9.01" y2="9"></line><line x1="15" y1="9" x2="15.01" y2="9"></line></svg>
              </button>
            </div>
          </div>
        </div>
      </div>
    {/if}
    
    <!-- Tweet List -->
    <div>
      <!-- Loading state when first loading tab -->
      {#if isLoading && currentTweets.length === 0}
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
      <!-- Error state -->
      {:else if error}
        <div class="p-8 text-center">
          <p class="text-red-500 mb-4">{error}</p>
          <button 
            class="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600" 
            on:click={() => activeTab === 'for-you' ? fetchTweetsForYou(true) : fetchTweetsFollowing(true)}
          >
            Try Again
          </button>
        </div>
      <!-- Authentication required message -->
      {:else if !authState.isAuthenticated}
        <div class="p-8 text-center text-gray-500 dark:text-gray-400">
          <p class="mb-4">Please login to see your personalized feed</p>
        </div>
      <!-- Empty state -->
      {:else if currentTweets.length === 0 && !isLoading}
        <div class="p-8 text-center text-gray-500 dark:text-gray-400">
          <p class="mb-4">
            {#if activeTab === 'for-you'}
              No posts yet
            {:else}
              You're not following anyone yet, or they haven't posted
            {/if}
          </p>
          {#if activeTab === 'for-you'}
            <button 
              class="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600" 
              on:click={toggleComposeModal}
            >
              Create First Post
            </button>
          {:else}
            <a 
              href="/explore" 
              class="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600 inline-block"
            >
              Find People to Follow
            </a>
          {/if}
        </div>
      <!-- Tweets list -->
      {:else}
        {#each currentTweets as tweet (tweet.id)}
          {#if tweet.isAdvertisement}
            <div class="p-4 border-b border-gray-200 dark:border-gray-800 bg-gray-50 dark:bg-gray-900">
              <div class="flex items-center text-xs text-gray-500 mb-2">
                <span class="bg-blue-100 dark:bg-blue-900 text-blue-600 dark:text-blue-400 px-2 py-1 rounded-full">Advertisement</span>
              </div>
              <div class="flex space-x-3">
                <div class="flex-shrink-0">
                  <div class="w-10 h-10 bg-blue-100 dark:bg-blue-900 rounded-full flex items-center justify-center">
                    {tweet.avatar}
                  </div>
                </div>
                <div class="flex-1">
                  <div class="font-bold text-blue-600 dark:text-blue-500">
                    {tweet.displayName}
                  </div>
                  <div class="text-gray-700 dark:text-gray-300 mt-2">
                    {tweet.content}
                  </div>
                  <div class="bg-white dark:bg-gray-850 rounded-xl border border-gray-200 dark:border-gray-700 p-3 mt-3">
                    <p class="text-gray-600 dark:text-gray-400">Sponsored content goes here</p>
                  </div>
                </div>
              </div>
            </div>
          {:else}
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
              on:click={() => openThreadModal(tweet)}
              on:like={handleTweetLike}
              on:unlike={handleTweetUnlike}
              on:repost={() => {}}
              on:reply={handleTweetReply}
              on:bookmark={() => {}}
              on:removeBookmark={() => {}}
              on:loadReplies={() => {}}
            />
          {/if}
        {/each}
        
        <!-- Loading more state -->
        {#if isLoading}
          <div class="flex justify-center items-center p-4">
            <div class="h-8 w-8 border-t-2 border-b-2 border-blue-500 rounded-full animate-spin"></div>
          </div>
        <!-- Load more button -->
        {:else if hasMore}
          <div class="p-4 text-center">
            <button 
              class="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600" 
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
    avatar={sidebarAvatar}
    replyTo={selectedTweet}
  />
{/if}

<!-- Toast notifications -->
<Toast />

<!-- Debug panel -->
<DebugPanel />

<style>
  /* Skeleton loading animation */
  @keyframes pulse {
    0%, 100% { opacity: 0.5; }
    50% { opacity: 1; }
  }
  .animate-pulse {
    animation: pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite;
  }
  
  /* Spinner animation */
  @keyframes spin {
    0% { transform: rotate(0deg); }
    100% { transform: rotate(360deg); }
  }
  .animate-spin {
    animation: spin 1s linear infinite;
  }
</style>