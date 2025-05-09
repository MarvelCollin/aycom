<script lang="ts">
  import type { ITrend, ISuggestedFollow } from "../../interfaces/ISocialMedia";
  import { createEventDispatcher, onMount } from 'svelte';
  import { createLoggerWithPrefix } from '../../utils/logger';
  import { getAuthToken } from '../../utils/auth';
  import appConfig from '../../config/appConfig';

  export let isDarkMode: boolean = false;
  
  // Initialize with empty arrays
  let trends: ITrend[] = [];
  let suggestedFollows: ISuggestedFollow[] = [];
  let isTrendsLoading = true;
  let isSuggestedFollowsLoading = true;
  
  const API_BASE_URL = appConfig.api.baseUrl;
  const logger = createLoggerWithPrefix('RightSide');
  const dispatch = createEventDispatcher();

  // Function to navigate to a route
  function navigateTo(path) {
    console.log(`Navigating to: ${path}`);
    window.location.href = path;
  }
  
  // Helper function to format Supabase image URLs
  function formatSupabaseImageUrl(url: string): string {
    if (!url) return 'https://secure.gravatar.com/avatar/0?d=mp';
    
    // If already a full URL, return as is
    if (url.startsWith('http')) return url;
    
    // If it contains emoji or placeholder indicator, return default
    if (url.includes('👤')) return 'https://secure.gravatar.com/avatar/0?d=mp';
    
    // Otherwise, construct the Supabase URL
    const supabaseUrl = import.meta.env.VITE_SUPABASE_URL || 'https://your-supabase-url.supabase.co';
    return `${supabaseUrl}/storage/v1/object/public/tpaweb/${url}`;
  }

  // Dispatch follow event to parent component
  function handleFollow(username: string) {
    const user = suggestedFollows.find(user => user.username === username);
    if (user) {
      // Toggle following state
      user.isFollowing = !user.isFollowing;
      suggestedFollows = [...suggestedFollows]; // Force refresh
      
      // Dispatch event
      dispatch('follow', { username, isFollowing: user.isFollowing });
      
      // In a real app, make API call to follow/unfollow
      try {
        if (user.isFollowing) {
          logger.debug(`Following user: ${username}`);
          // API call would go here
        } else {
          logger.debug(`Unfollowing user: ${username}`);
          // API call would go here
        }
      } catch (error) {
        logger.error(`Error ${user.isFollowing ? 'following' : 'unfollowing'} user:`, error);
        // Revert state on error
        user.isFollowing = !user.isFollowing;
        suggestedFollows = [...suggestedFollows]; // Force refresh
      }
    }
  }

  // Handle trend click
  function handleTrendClick(trend: ITrend) {
    dispatch('trendClick', trend);
    // Navigate to a hashtag page
    logger.debug(`Clicked on trend: ${trend.title}`);
    // navigateTo(`/explore/hashtags/${trend.title.replace('#', '')}`);
  }

  // Directly fetch trending hashtags
  async function fetchTrendingHashtags() {
    isTrendsLoading = true;
    try {
      const token = getAuthToken();
      const response = await fetch(`${API_BASE_URL}/trends?limit=5`, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': token ? `Bearer ${token}` : ''
        },
        credentials: 'include'
      });
      
      if (!response.ok) {
        throw new Error(`Failed to fetch trends: ${response.status} ${response.statusText}`);
      }
      
      const data = await response.json();
      
      if (data && data.trends && Array.isArray(data.trends)) {
        trends = data.trends.map(trend => ({
          id: trend.id || `trend-${Math.random().toString(36).substring(2, 9)}`,
          category: trend.category || 'Trending',
          title: trend.title,
          postCount: trend.post_count || 0
        }));
        logger.debug(`Fetched ${trends.length} trends`);
      } else {
        // Use hardcoded trends if API fails
        useFallbackTrends();
      }
    } catch (error) {
      logger.error('Error fetching trends:', error);
      useFallbackTrends();
    } finally {
      isTrendsLoading = false;
    }
  }
  
  // Directly fetch suggested users to follow
  async function fetchSuggestedUsers() {
    isSuggestedFollowsLoading = true;
    try {
      const token = getAuthToken();
      const response = await fetch(`${API_BASE_URL}/users/suggestions?limit=3`, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': token ? `Bearer ${token}` : ''
        },
        credentials: 'include'
      });
      
      if (!response.ok) {
        throw new Error(`Failed to fetch suggested users: ${response.status} ${response.statusText}`);
      }
      
      const data = await response.json();
      
      if (data && data.users && Array.isArray(data.users)) {
        suggestedFollows = data.users.map(user => ({
          username: user.username,
          displayName: user.display_name || user.username,
          avatar: user.avatar_url || null,
          verified: user.verified || false,
          followerCount: user.follower_count || 0,
          isFollowing: user.is_following || false
        }));
        logger.debug(`Fetched ${suggestedFollows.length} suggested users`);
      } else {
        // Use hardcoded users if API fails
        useFallbackUsers();
      }
    } catch (error) {
      logger.error('Error fetching suggested users:', error);
      useFallbackUsers();
    } finally {
      isSuggestedFollowsLoading = false;
    }
  }

  // Fallback functions when API calls fail
  function useFallbackTrends() {
    trends = [
      { id: 'trend-1', category: 'Technology', title: '#AI', postCount: '125K' },
      { id: 'trend-2', category: 'Entertainment', title: '#Music', postCount: '98K' },
      { id: 'trend-3', category: 'News', title: '#BreakingNews', postCount: '87K' },
      { id: 'trend-4', category: 'Sports', title: '#Basketball', postCount: '76K' },
      { id: 'trend-5', category: 'Politics', title: '#Election', postCount: '65K' }
    ];
    logger.debug('Using fallback trends');
  }
  
  function useFallbackUsers() {
    suggestedFollows = [
      { 
        username: 'tech_insider', 
        displayName: 'Tech Insider', 
        avatar: 'https://i.pravatar.cc/150?u=tech_insider', 
        verified: true,
        followerCount: 1240000,
        isFollowing: false
      },
      { 
        username: 'travel_adventures', 
        displayName: 'Travel Adventures', 
        avatar: 'https://i.pravatar.cc/150?u=travel_adventures', 
        verified: true,
        followerCount: 890000,
        isFollowing: false
      },
      { 
        username: 'photo_daily', 
        displayName: 'Photography Daily', 
        avatar: 'https://i.pravatar.cc/150?u=photo_daily', 
        verified: false,
        followerCount: 625000,
        isFollowing: false
      }
    ];
    logger.debug('Using fallback users');
  }

  // Fetch data on component mount
  onMount(() => {
    fetchTrendingHashtags();
    fetchSuggestedUsers();
  });
</script>

<div class="search-container {isDarkMode ? 'search-dark' : ''} sticky top-0 z-10 pb-3 bg-inherit">
  <div class="relative">
    <input 
      type="text" 
      placeholder="Search" 
      class="search-input w-full rounded-full {isDarkMode ? 'bg-gray-800 border-gray-700 text-white' : 'bg-gray-100 border-gray-200'}"
    />
    <div class="search-icon">
      <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
      </svg>
    </div>
  </div>
</div>

<div class="sidebar {isDarkMode ? 'sidebar-dark' : ''} rounded-xl mb-4">
  <div class="sidebar-section">
    <h2 class="sidebar-title text-xl font-bold">What's happening</h2>
    
    {#if isTrendsLoading}
      <div class="py-4">
        <div class="animate-pulse">
          {#each Array(3) as _, i}
            <div class="py-3 {i !== 2 ? 'border-b' : ''} {isDarkMode ? 'border-gray-700' : 'border-gray-200'}">
              <div class="h-4 {isDarkMode ? 'bg-gray-700' : 'bg-gray-200'} rounded w-1/3 mb-2"></div>
              <div class="h-5 {isDarkMode ? 'bg-gray-700' : 'bg-gray-200'} rounded w-3/4 mb-2"></div>
              <div class="h-4 {isDarkMode ? 'bg-gray-700' : 'bg-gray-200'} rounded w-1/4"></div>
            </div>
          {/each}
        </div>
      </div>
    {:else if trends.length > 0}
      <ul>
        {#each trends as trend}
          <li 
            class="py-3 cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors {trends.indexOf(trend) !== trends.length - 1 ? 'border-b' : ''} {isDarkMode ? 'border-gray-700' : 'border-gray-200'}"
            on:click={() => handleTrendClick(trend)}
          >
            <p class="{isDarkMode ? 'text-gray-400' : 'text-gray-500'} text-sm">{trend.category}</p>
            <p class="font-semibold {isDarkMode ? 'text-white' : 'text-black'} my-0.5">{trend.title}</p>
            <p class="{isDarkMode ? 'text-gray-400' : 'text-gray-500'} text-sm">{trend.postCount} posts</p>
          </li>
        {/each}
      </ul>
      <div class="mt-3 pt-2">
        <a href="/explore" class="text-blue-500 hover:text-blue-600 text-sm font-medium">
          Show more
        </a>
      </div>
    {:else}
      <p class="{isDarkMode ? 'text-gray-400' : 'text-gray-500'} py-4">No trends available</p>
    {/if}
  </div>
</div>

<div class="sidebar {isDarkMode ? 'sidebar-dark' : ''} rounded-xl mb-4">
  <div class="sidebar-section">
    <h2 class="sidebar-title text-xl font-bold">Who to follow</h2>
    
    {#if isSuggestedFollowsLoading}
      <div class="py-4">
        <div class="animate-pulse">
          {#each Array(2) as _, i}
            <div class="flex items-center gap-3 py-3 {i !== 1 ? 'border-b' : ''} {isDarkMode ? 'border-gray-700' : 'border-gray-200'}">
              <div class="w-10 h-10 rounded-full {isDarkMode ? 'bg-gray-700' : 'bg-gray-200'}"></div>
              <div class="flex-1">
                <div class="h-4 {isDarkMode ? 'bg-gray-700' : 'bg-gray-200'} rounded w-3/4 mb-2"></div>
                <div class="h-3 {isDarkMode ? 'bg-gray-700' : 'bg-gray-200'} rounded w-1/2"></div>
              </div>
              <div class="h-8 w-20 {isDarkMode ? 'bg-gray-700' : 'bg-gray-200'} rounded-full"></div>
            </div>
          {/each}
        </div>
      </div>
    {:else if suggestedFollows.length > 0}
      <ul>
        {#each suggestedFollows as follow}
          <li class="flex items-center gap-3 py-3 {suggestedFollows.indexOf(follow) !== suggestedFollows.length - 1 ? 'border-b' : ''} {isDarkMode ? 'border-gray-700' : 'border-gray-200'}">
            <div class="w-10 h-10 rounded-full overflow-hidden {isDarkMode ? 'bg-gray-700' : 'bg-gray-200'} flex-shrink-0 flex items-center justify-center">
              <div class="flex items-center justify-center w-full h-full text-lg">
                <img src={formatSupabaseImageUrl(follow.avatar)} alt={follow.username} class="w-full h-full object-cover rounded-full" />
              </div>
            </div>
            <div class="flex-1 min-w-0">
              <div class="flex items-center">
                <p class="font-semibold {isDarkMode ? 'text-white' : 'text-black'} truncate">{follow.displayName}</p>
                {#if follow.verified}
                  <span class="ml-1 text-blue-500">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor">
                      <path fill-rule="evenodd" d="M6.267 3.455a3.066 3.066 0 001.745-.723 3.066 3.066 0 013.976 0 3.066 3.066 0 001.745.723 3.066 3.066 0 012.812 2.812c.051.643.304 1.254.723 1.745a3.066 3.066 0 010 3.976 3.066 3.066 0 00-.723 1.745 3.066 3.066 0 01-2.812 2.812 3.066 3.066 0 00-1.745.723 3.066 3.066 0 01-3.976 0 3.066 3.066 0 00-1.745-.723 3.066 3.066 0 01-2.812-2.812 3.066 3.066 0 00-.723-1.745 3.066 3.066 0 010-3.976 3.066 3.066 0 00.723-1.745 3.066 3.066 0 012.812-2.812zm7.44 5.252a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                    </svg>
                  </span>
                {/if}
              </div>
              <p class="{isDarkMode ? 'text-gray-400' : 'text-gray-500'} text-sm truncate">@{follow.username}</p>
            </div>
            <button 
              class="follow-button rounded-full text-sm font-bold px-4 py-1.5 transition-colors {follow.isFollowing 
                ? isDarkMode 
                  ? 'bg-transparent border border-gray-400 text-white hover:bg-gray-800' 
                  : 'bg-transparent border border-gray-300 text-black hover:bg-gray-100' 
                : isDarkMode 
                  ? 'bg-white text-black hover:bg-gray-200' 
                  : 'bg-black text-white hover:bg-gray-800'}"
              on:click={() => handleFollow(follow.username)}
            >
              {follow.isFollowing ? 'Following' : 'Follow'}
            </button>
          </li>
        {/each}
      </ul>
      <div class="mt-3 pt-2">
        <a href="/connect" class="text-blue-500 hover:text-blue-600 text-sm font-medium">
          Show more
        </a>
      </div>
    {:else}
      <p class="{isDarkMode ? 'text-gray-400' : 'text-gray-500'} py-4">No suggestions available</p>
    {/if}
  </div>
</div>

<div class="mt-4 text-center footer-links {isDarkMode ? 'footer-dark' : ''}">
  <div class="flex flex-wrap text-xs justify-between {isDarkMode ? 'text-gray-500' : 'text-gray-500'}">
    <button class="dark:bg-black hover:underline mb-2" on:click={() => navigateTo('/terms')}>Terms of Service</button>
    <button class="dark:bg-black hover:underline mb-2" on:click={() => navigateTo('/privacy')}>Privacy Policy</button>
    <button class="dark:bg-black hover:underline mb-2" on:click={() => navigateTo('/cookies')}>Cookie Policy</button>
    <button class="dark:bg-black hover:underline mb-2" on:click={() => navigateTo('/accessibility')}>Accessibility</button>
    <button class="dark:bg-black hover:underline mb-2" on:click={() => navigateTo('/ads')}>Ads Info</button>
    <button class="dark:bg-black hover:underline mb-2" on:click={() => navigateTo('/about')}>About</button>
  </div>
  <p class="text-xs mt-2 {isDarkMode ? 'text-gray-500' : 'text-gray-500'}">© 2023 AYCOM, Inc.</p>
</div>

<style>
  /* Dark mode styling */
  .sidebar {
    background-color: #f7f9fa;
    border: 1px solid #eff3f4;
    padding: 1rem;
  }
  
  .sidebar-dark {
    background-color: #16181c;
    border: 1px solid #2f3336;
  }
  
  .search-input {
    padding: 0.75rem 1rem 0.75rem 3rem;
    outline: none;
  }
  
  .search-icon {
    position: absolute;
    left: 1rem;
    top: 50%;
    transform: translateY(-50%);
    color: #536471;
  }
  
  .footer-links button {
    transition: color 0.2s;
  }
  
  .footer-dark button:hover {
    color: #e5e7eb;
  }
  
  .follow-button {
    transition: background-color 0.2s ease;
  }
</style>