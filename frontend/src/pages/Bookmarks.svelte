<script lang="ts">
  import { onMount } from 'svelte';
  import MainLayout from '../components/layout/MainLayout.svelte';
  import { useAuth } from '../hooks/useAuth';
  import { useTheme } from '../hooks/useTheme';
  import type { ITweet } from '../interfaces/ISocialMedia';
  import type { IAuthStore } from '../interfaces/IAuth';
  import { createLoggerWithPrefix } from '../utils/logger';
  import { toastStore } from '../stores/toastStore';
  import TweetCard from '../components/social/TweetCard.svelte';
  
  const logger = createLoggerWithPrefix('Bookmarks');
  
  // Auth and theme
  const { getAuthState } = useAuth();
  const { theme } = useTheme();
  
  // Reactive declarations
  $: authState = getAuthState ? (getAuthState() as IAuthStore) : { userId: null, isAuthenticated: false, accessToken: null, refreshToken: null };
  $: isDarkMode = $theme === 'dark';
  $: sidebarUsername = authState?.userId ? `User_${authState.userId.substring(0, 4)}` : '';
  $: sidebarDisplayName = authState?.userId ? `User ${authState.userId.substring(0, 4)}` : '';
  $: sidebarAvatar = 'https://secure.gravatar.com/avatar/0?d=mp'; // Default avatar with proper image URL
  
  // Bookmarks state
  let isLoading = true;
  let bookmarks: ITweet[] = [];
  let filteredBookmarks: ITweet[] = [];
  let searchQuery = '';
  
  // Authentication check
  function checkAuth() {
    if (!authState.isAuthenticated) {
      toastStore.showToast('You need to log in to access bookmarks', 'warning');
      window.location.href = '/login';
      return false;
    }
    return true;
  }
  
  // Fetch bookmarks
  async function fetchBookmarks() {
    isLoading = true;
    
    try {
      // In a real implementation, this would be an API call to fetch bookmarks
      // For example: const response = await fetch('/api/bookmarks');
      
      // Simulate API response with mock data
      setTimeout(() => {
        // Mock bookmarks data
        bookmarks = [
          {
            id: 'tweet1',
            username: '@Brawlhalla',
            displayName: 'Brawlhalla',
            avatar: null,
            content: 'The Brawlhalla: Guardians Crossover Event has begun! Choose Titan, Hunter, Warlock, or all three and Brawl in the Grand Tournament! Your Ghost Companions and Ghaisthuns, topped with with Heavy Ammo, are waiting for you now!',
            timestamp: '2023-05-27T10:00:00Z',
            likes: 621,
            replies: 27,
            reposts: 71,
            bookmarks: 0,
            views: '135K',
            isLiked: false,
            isReposted: false,
            isBookmarked: true,
            // Additional properties used in UI
            authorId: 'brawlhalla',
            authorName: 'Brawlhalla',
            authorUsername: '@Brawlhalla',
            authorAvatar: null,
            verified: true,
            likesCount: 621,
            commentsCount: 27,
            repostsCount: 71,
            viewsCount: '135K',
            createdAt: '2023-05-27T10:00:00Z',
            media: [
              {
                type: 'image',
                url: 'https://example.com/brawlhalla-event.jpg',
                alt: 'Brawlhalla event promotional image'
              }
            ],
            poll: null,
            quoteTweet: null,
            replyingTo: null
          },
          {
            id: 'tweet2',
            username: '@AstralisCS',
            displayName: 'Astralis Counter-Strike',
            avatar: null,
            content: '@clue vs on train vs ENCE\n16-8 W/D\n19/24 ADR\nrating: 1.25',
            timestamp: '2023-08-12T15:30:00Z',
            likes: 320,
            replies: 15,
            reposts: 42,
            bookmarks: 0,
            views: '98K',
            isLiked: false,
            isReposted: false,
            isBookmarked: true,
            // Additional properties used in UI
            authorId: 'astraliscs',
            authorName: 'Astralis Counter-Strike',
            authorUsername: '@AstralisCS',
            authorAvatar: null,
            verified: true,
            likesCount: 320,
            commentsCount: 15,
            repostsCount: 42,
            viewsCount: '98K',
            createdAt: '2023-08-12T15:30:00Z',
            media: [],
            poll: null,
            quoteTweet: null,
            replyingTo: null
          },
          {
            id: 'tweet3',
            username: '@AstralisCS',
            displayName: 'Astralis Counter-Strike',
            avatar: null,
            content: 'advice',
            timestamp: '2023-04-15T09:15:00Z',
            likes: 510,
            replies: 30,
            reposts: 85,
            bookmarks: 0,
            views: '120K',
            isLiked: true,
            isReposted: false,
            isBookmarked: true,
            // Additional properties used in UI
            authorId: 'astraliscs',
            authorName: 'Astralis Counter-Strike',
            authorUsername: '@AstralisCS',
            authorAvatar: null,
            verified: true,
            likesCount: 510,
            commentsCount: 30,
            repostsCount: 85,
            viewsCount: '120K',
            createdAt: '2023-04-15T09:15:00Z',
            media: [
              {
                type: 'image',
                url: 'https://example.com/astralis-advice.jpg',
                alt: 'Astralis advice image'
              }
            ],
            poll: null,
            quoteTweet: null,
            replyingTo: null
          }
        ];
        
        filteredBookmarks = [...bookmarks];
        isLoading = false;
        logger.debug('Bookmarks loaded', { count: bookmarks.length });
      }, 1000);
      
    } catch (error) {
      console.error('Error fetching bookmarks:', error);
      toastStore.showToast('Failed to load bookmarks. Please try again.', 'error');
      isLoading = false;
    }
  }
  
  // Filter bookmarks based on search query
  function filterBookmarks() {
    if (!searchQuery.trim()) {
      filteredBookmarks = [...bookmarks];
      return;
    }
    
    const query = searchQuery.toLowerCase();
    filteredBookmarks = bookmarks.filter(bookmark => 
      bookmark.content.toLowerCase().includes(query) || 
      bookmark.authorName?.toLowerCase().includes(query) || 
      bookmark.authorUsername?.toLowerCase().includes(query) ||
      bookmark.displayName.toLowerCase().includes(query) || 
      bookmark.username.toLowerCase().includes(query)
    );
    
    logger.debug('Bookmarks filtered', { query, resultsCount: filteredBookmarks.length });
  }
  
  // Handle bookmark removal
  function handleRemoveBookmark(event) {
    const tweetId = event.detail;
    
    // In a real implementation, this would be an API call to remove bookmark
    // For example: await fetch(`/api/bookmarks/${tweetId}`, { method: 'DELETE' });
    
    // Update local state
    bookmarks = bookmarks.filter(bookmark => bookmark.id !== tweetId);
    filteredBookmarks = filteredBookmarks.filter(bookmark => bookmark.id !== tweetId);
    
    logger.debug('Bookmark removed', { tweetId });
    toastStore.showToast('Bookmark removed', 'success');
  }
  
  // Watch for search query changes
  $: if (searchQuery !== undefined) {
    filterBookmarks();
  }
  
  onMount(() => {
    logger.debug('Bookmarks page mounted', { authState });
    if (checkAuth()) {
      fetchBookmarks();
    }
  });
</script>

<MainLayout
  username={sidebarUsername}
  displayName={sidebarDisplayName}
  avatar={sidebarAvatar}
  on:toggleComposeModal={() => {}}
>
  <div class="min-h-screen border-x border-gray-200 dark:border-gray-800">
    <!-- Header -->
    <div class="sticky top-0 z-10 bg-white/80 dark:bg-black/80 backdrop-blur-md border-b border-gray-200 dark:border-gray-800 px-4 py-3">
      <h1 class="text-xl font-bold">Bookmarks</h1>
      
      <!-- Search -->
      <div class="relative mt-3">
        <input 
          type="text" 
          bind:value={searchQuery}
          placeholder="Search Bookmarks" 
          class="w-full rounded-full pl-10 pr-4 py-2 {isDarkMode ? 'bg-gray-800 border-gray-700 text-white' : 'bg-gray-100 border-gray-200'}"
        />
        <div class="absolute left-3 top-2.5 text-gray-500">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
          </svg>
        </div>
      </div>
    </div>
    
    <!-- Content -->
    <div class="divide-y divide-gray-200 dark:divide-gray-800">
      {#if isLoading}
        <div class="p-4">
          <div class="animate-pulse space-y-4">
            {#each Array(3) as _}
              <div class="flex space-x-4">
                <div class="rounded-full bg-gray-300 dark:bg-gray-700 h-12 w-12"></div>
                <div class="flex-1 space-y-2 py-1">
                  <div class="h-4 bg-gray-300 dark:bg-gray-700 rounded w-3/4"></div>
                  <div class="h-4 bg-gray-300 dark:bg-gray-700 rounded w-1/2"></div>
                  <div class="h-24 bg-gray-300 dark:bg-gray-700 rounded w-full mt-2"></div>
                  <div class="h-4 bg-gray-300 dark:bg-gray-700 rounded w-full mt-2"></div>
                </div>
              </div>
            {/each}
          </div>
        </div>
      {:else if filteredBookmarks.length === 0}
        <div class="flex flex-col items-center justify-center py-16 px-4 text-center">
          {#if bookmarks.length === 0}
            <h2 class="text-2xl font-bold mb-2">You haven't added any Bookmarks yet</h2>
            <p class="text-gray-500 dark:text-gray-400 mb-8 max-w-md">
              When you do, they'll show up here.
            </p>
          {:else}
            <h2 class="text-2xl font-bold mb-2">No results found</h2>
            <p class="text-gray-500 dark:text-gray-400 mb-8 max-w-md">
              Try a different search term.
            </p>
          {/if}
        </div>
      {:else}
        {#each filteredBookmarks as bookmark}
          <div class="border-b border-gray-200 dark:border-gray-800">
            <TweetCard 
              tweet={bookmark} 
              on:removeBookmark={handleRemoveBookmark} 
            />
          </div>
        {/each}
      {/if}
    </div>
  </div>
</MainLayout>

<style>
  /* Skeleton loading animation */
  @keyframes pulse {
    0%, 100% { opacity: 0.5; }
    50% { opacity: 1; }
  }
  .animate-pulse {
    animation: pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite;
  }
</style>
