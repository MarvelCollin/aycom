<script lang="ts">
  import { onMount } from 'svelte';
  import MainLayout from '../components/layout/MainLayout.svelte';
  import { useAuth } from '../hooks/useAuth';
  import { useTheme } from '../hooks/useTheme';
  import type { ITweet, ITrend } from '../interfaces/ISocialMedia';
  import type { IAuthStore } from '../interfaces/IAuth';
  import { createLoggerWithPrefix } from '../utils/logger';
  import { toastStore } from '../stores/toastStore';
  import { getTrends } from '../api/trends';
  import { searchUsers } from '../api/user';
  import { searchThreads, searchThreadsWithMedia, getThreadsByHashtag } from '../api/thread';
  import { searchCommunities } from '../api/community';
  import { debounce } from '../utils/helpers';
  
  // Import newly created components
  import ExploreSearch from '../components/explore/ExploreSearch.svelte';
  import ExploreFilters from '../components/explore/ExploreFilters.svelte';
  import ExploreTrending from '../components/explore/ExploreTrending.svelte';
  import ExploreTabs from '../components/explore/ExploreTabs.svelte';
  import ExploreTopResults from '../components/explore/ExploreTopResults.svelte';
  import ExploreLatestResults from '../components/explore/ExploreLatestResults.svelte';
  import ExplorePeopleResults from '../components/explore/ExplorePeopleResults.svelte';
  import ExploreMediaResults from '../components/explore/ExploreMediaResults.svelte';
  import ExploreCommunityResults from '../components/explore/ExploreCommunityResults.svelte';
  import LoadingSkeleton from '../components/common/LoadingSkeleton.svelte';
  import Toast from '../components/common/Toast.svelte';

  const logger = createLoggerWithPrefix('Explore');
  
  // Auth and theme
  const { getAuthState } = useAuth();
  const { theme } = useTheme();
  
  // Reactive declarations
  $: authState = getAuthState ? (getAuthState() as IAuthStore) : { userId: null, isAuthenticated: false, accessToken: null, refreshToken: null };
  $: isDarkMode = $theme === 'dark';
  $: sidebarUsername = authState?.userId ? `User_${authState.userId.substring(0, 4)}` : '';
  $: sidebarDisplayName = authState?.userId ? `User ${authState.userId.substring(0, 4)}` : '';
  $: sidebarAvatar = 'https://secure.gravatar.com/avatar/0?d=mp'; // Default avatar with proper image URL
  
  // Trends data
  let trends: ITrend[] = [];
  let isTrendsLoading = true;
  
  // Search state
  let searchQuery = '';
  let isSearching = false;
  let recentSearches: string[] = [];
  let searchFilter: 'all' | 'following' | 'verified' = 'all';
  let activeTab: 'top' | 'latest' | 'people' | 'media' | 'communities' = 'top';
  let showRecentSearches = false;
  let isLoadingRecommendations = false;
  
  // Results state
  let searchResults: {
    top: {
      profiles: Array<{
        id: string;
        username: string;
        displayName: string;
        avatar: string | null;
        bio?: string;
        isVerified: boolean;
        followerCount: number;
        isFollowing: boolean;
      }>;
      threads: Array<{
        id: string;
        content: string;
        username: string;
        displayName: string;
        timestamp: string;
        likes: number;
        replies: number;
        reposts: number;
        media?: Array<{
          url: string;
          type: string;
        }>;
        avatar?: string;
      }>;
      isLoading: boolean;
    };
    latest: {
      threads: Array<{
        id: string;
        content: string;
        username: string;
        displayName: string;
        timestamp: string;
        likes: number;
        replies: number;
        reposts: number;
        media?: Array<{
          url: string;
          type: string;
        }>;
        avatar?: string;
      }>;
      isLoading: boolean;
    };
    people: {
      users: Array<{
        id: string;
        username: string;
        display_name?: string;
        avatar?: string;
        bio?: string;
        is_verified?: boolean;
        is_following?: boolean;
      }>;
      totalCount: number;
      isLoading: boolean;
    };
    media: {
      threads: Array<{
        id: string;
        media?: Array<{
          url: string;
          type: string;
        }>;
      }>;
      isLoading: boolean;
    };
    communities: {
      communities: Array<{
        id: string;
        name: string;
        description?: string;
        logo?: string;
        member_count?: number;
        is_joined?: boolean;
        is_pending?: boolean;
      }>;
      totalCount: number;
      isLoading: boolean;
    };
  } = {
    top: {
      profiles: [],
      threads: [],
      isLoading: false
    },
    latest: {
      threads: [],
      isLoading: false
    },
    people: {
      users: [],
      totalCount: 0,
      isLoading: false
    },
    media: {
      threads: [],
      isLoading: false
    },
    communities: {
      communities: [],
      totalCount: 0,
      isLoading: false
    }
  };
  
  // Has user performed a search?
  let hasSearched = false;
  
  // Thread categories
  const threadCategories = [
    { id: 'all', name: 'All Categories' },
    { id: 'news', name: 'News' },
    { id: 'sports', name: 'Sports' },
    { id: 'entertainment', name: 'Entertainment' },
    { id: 'politics', name: 'Politics' },
    { id: 'tech', name: 'Technology' },
    { id: 'science', name: 'Science' },
    { id: 'health', name: 'Health' }
  ];
  let selectedCategory = 'all';
  
  // Pagination options for People and Communities tabs
  let peoplePerPage = 25; 
  let peopleCurrentPage = 1;
  let communitiesPerPage = 25;
  let communitiesCurrentPage = 1;
  let mediaPage = 1; // Media infinite scroll page
  
  // Authentication check
  function checkAuth() {
    if (!authState.isAuthenticated) {
      toastStore.showToast('You need to log in to access explore', 'warning');
      window.location.href = '/login';
      return false;
    }
    return true;
  }
  
  // Load recent searches from localStorage
  function loadRecentSearches() {
    try {
      const savedSearches = localStorage.getItem('recentSearches');
      if (savedSearches) {
        recentSearches = JSON.parse(savedSearches).slice(0, 3);
      }
    } catch (error) {
      console.error('Error loading recent searches:', error);
      recentSearches = [];
    }
  }
  
  // Save a search to recent searches
  function saveToRecentSearches(query: string) {
    if (!query.trim()) return;
    
    try {
      // Remove if it already exists (to avoid duplicates)
      const filteredSearches = recentSearches.filter(s => s !== query);
      
      // Add to beginning of array
      const updatedSearches = [query, ...filteredSearches].slice(0, 3);
      recentSearches = updatedSearches;
      
      // Save to localStorage
      localStorage.setItem('recentSearches', JSON.stringify(updatedSearches));
    } catch (error) {
      console.error('Error saving recent search:', error);
    }
  }
  
  // Load trending hashtags
  async function fetchTrends() {
    isTrendsLoading = true;
    try {
      const trendData = await getTrends(10);
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
  
  // Search for recommended profiles - with debounce from lodash-es
  const debouncedSearchProfiles = debounce(async (query: string) => {
    if (!query || query.length < 2) {
      searchResults.top.profiles = [];
      return;
    }
    
    isLoadingRecommendations = true;
    try {
      const { users } = await searchUsers(query, 1, 3);
      searchResults.top.profiles = users.map(user => ({
        id: user.id,
        username: user.username,
        displayName: user.display_name || user.username,
        avatar: user.avatar,
        bio: user.bio,
        isVerified: user.is_verified || false,
        followerCount: (user as any).follower_count || 0,
        isFollowing: user.is_following || false
      }));
    } catch (error) {
      console.error('Error searching profiles:', error);
      searchResults.top.profiles = [];
    } finally {
      isLoadingRecommendations = false;
    }
  }, 300);
  
  // Handle search input with debounce for recommendations
  function handleSearchInput(event) {
    searchQuery = event.detail;
    debouncedSearchProfiles(searchQuery);
    logger.debug('Search input updated', { searchQuery });
  }
  
  // Execute search
  async function executeSearch() {
    if (!searchQuery.trim()) return;
    
    // Save to recent searches
    saveToRecentSearches(searchQuery);
    
    // Hide recent searches dropdown
    showRecentSearches = false;
    
    // Mark as searching
    isSearching = true;
    hasSearched = true;
    
    try {
      // Set loading state for all tabs
      searchResults.top.isLoading = true;
      searchResults.latest.isLoading = true;
      searchResults.people.isLoading = true;
      searchResults.media.isLoading = true;
      searchResults.communities.isLoading = true;
      
      // Get filter options
      const filterOption = searchFilter === 'following' ? 'following' : (searchFilter === 'verified' ? 'verified' : 'all');
      const categoryOption = selectedCategory !== 'all' ? selectedCategory : undefined;
      
      // Fetch data for all tabs in parallel
      const [peopleData, topThreadsData, latestThreadsData, mediaData, communitiesData] = await Promise.all([
        // People tab data (also used for top profiles)
        searchUsers(searchQuery, 1, 25, { filter: filterOption }),
        
        // Top threads
        searchThreads(searchQuery, 1, 10, {
          filter: filterOption,
          category: categoryOption,
          sortBy: 'popular'
        }),
        
        // Latest threads
        searchThreads(searchQuery, 1, 20, {
          filter: filterOption,
          category: categoryOption,
          sortBy: 'recent'
        }),
        
        // Media tab data
        searchThreadsWithMedia(searchQuery, 1, 12, {
          filter: filterOption,
          category: categoryOption
        }),
        
        // Communities tab data
        searchCommunities(searchQuery, 1, 25)
      ]);
      
      // Update search results
      searchResults = {
        top: {
          profiles: peopleData.users.slice(0, 3).map(user => ({
            id: user.id,
            username: user.username,
            displayName: user.display_name || user.username,
            avatar: user.avatar || null,
            bio: user.bio,
            isVerified: user.is_verified || false,
            followerCount: (user as any).follower_count || 0,
            isFollowing: user.is_following || false
          })),
          threads: topThreadsData.threads.map(thread => ({
            id: thread.id,
            content: thread.content,
            username: thread.author?.username || 'anonymous',
            displayName: thread.author?.display_name || 'User',
            timestamp: thread.created_at || new Date().toISOString(),
            likes: thread.like_count || 0,
            replies: thread.reply_count || 0,
            reposts: thread.repost_count || 0,
            media: thread.media,
            avatar: thread.author?.avatar
          })) || [],
          isLoading: false
        },
        latest: {
          threads: latestThreadsData.threads.map(thread => ({
            id: thread.id,
            content: thread.content,
            username: thread.author?.username || 'anonymous',
            displayName: thread.author?.display_name || 'User',
            timestamp: thread.created_at || new Date().toISOString(),
            likes: thread.like_count || 0,
            replies: thread.reply_count || 0,
            reposts: thread.repost_count || 0,
            media: thread.media,
            avatar: thread.author?.avatar
          })) || [],
          isLoading: false
        },
        people: {
          users: peopleData.users.map(user => ({
            id: user.id,
            username: user.username,
            displayName: user.display_name || user.username,
            avatar: user.avatar || null,
            bio: user.bio,
            isVerified: user.is_verified || false,
            followerCount: (user as any).follower_count || 0,
            isFollowing: user.is_following || false
          })) || [],
          totalCount: peopleData.total_count || 0,
          isLoading: false
        },
        media: {
          threads: mediaData.threads || [],
          isLoading: false
        },
        communities: {
          communities: communitiesData.communities || [],
          totalCount: communitiesData.total_count || 0,
          isLoading: false
        }
      };
      
      logger.debug('Search completed', {
        query: searchQuery,
        peopleCount: peopleData.users.length,
        threadsCount: topThreadsData.threads.length,
        mediaCount: mediaData.threads.length
      });
      
    } catch (error) {
      console.error('Error executing search:', error);
      toastStore.showToast('Search failed. Please try again.', 'error');
    } finally {
      isSearching = false;
    }
  }
  
  // Handle search focus
  function handleSearchFocus() {
    showRecentSearches = true;
    logger.debug('Search focused');
  }
  
  // Handle selection of a recent search
  function handleRecentSearchSelect(event) {
    searchQuery = event.detail;
    logger.debug('Recent search selected', { searchQuery });
    executeSearch();
  }
  
  // Clear recent searches
  function clearRecentSearches() {
    recentSearches = [];
    localStorage.removeItem('recentSearches');
    logger.debug('Recent searches cleared');
  }
  
  // Handle filter change
  function handleFilterChange(event) {
    searchFilter = event.detail;
    logger.debug('Filter changed', { filter: searchFilter });
    if (hasSearched) {
      executeSearch();
    }
  }
  
  // Handle category change
  function handleCategoryChange(event) {
    const category = event.detail;
    selectedCategory = category;
    logger.debug('Category changed', { category });
    if (hasSearched) {
      executeSearch();
    }
  }
  
  // Handle tab change
  function handleTabChange(event) {
    activeTab = event.detail;
    logger.debug('Tab changed', { tab: activeTab });
  }
  
  // Handle hashtag click from trends
  function handleHashtagClick(event) {
    const hashtag = event.detail;
    searchQuery = hashtag;
    executeSearch();
  }
  
  // Pagination for people tab
  async function handlePeoplePageChange(event) {
    const page = event.detail;
    searchResults.people.isLoading = true;
    
    try {
      const filterOption = searchFilter === 'following' ? 'following' : (searchFilter === 'verified' ? 'verified' : 'all');
      const { users, total_count } = await searchUsers(searchQuery, page, 25, { filter: filterOption });
      
      searchResults.people = {
        users: users || [],
        totalCount: total_count || 0,
        isLoading: false
      };
    } catch (error) {
      console.error('Error loading people page:', error);
      toastStore.showToast('Failed to load more profiles', 'error');
      searchResults.people.isLoading = false;
    }
  }
  
  // Change people results per page
  function handlePeoplePerPageChange(event) {
    searchResults.people.isLoading = true;
    executeSearch();
  }
  
  // Pagination for communities tab
  async function handleCommunitiesPageChange(event) {
    const page = event.detail;
    searchResults.communities.isLoading = true;
    
    try {
      const { communities, total_count } = await searchCommunities(searchQuery, page, 25);
      
      searchResults.communities = {
        communities: communities || [],
        totalCount: total_count || 0,
        isLoading: false
      };
    } catch (error) {
      console.error('Error loading community page:', error);
      toastStore.showToast('Failed to load communities', 'error');
      searchResults.communities.isLoading = false;
    }
  }
  
  // Change communities results per page
  function handleCommunitiesPerPageChange(event) {
    searchResults.communities.isLoading = true;
    executeSearch();
  }
  
  // Load more media (for infinite scroll)
  async function loadMoreMedia() {
    if (searchResults.media.isLoading) return;
    
    searchResults.media.isLoading = true;
    mediaPage++;
    
    try {
      const filterOption = searchFilter === 'following' ? 'following' : 'all';
      const categoryOption = selectedCategory !== 'all' ? selectedCategory : undefined;
      
      const data = await searchThreadsWithMedia(searchQuery, mediaPage, 12, {
        filter: filterOption,
        category: categoryOption
      });
      
      // Append new media to existing results
      searchResults.media = {
        threads: [...searchResults.media.threads, ...(data.threads || [])],
        isLoading: false
      };
    } catch (error) {
      console.error('Error loading more media:', error);
      toastStore.showToast('Failed to load more media', 'error');
      searchResults.media.isLoading = false;
    }
  }
  
  // Get threads by hashtag (for trending section)
  async function getThreadsByHashtagName(hashtag: string) {
    try {
      // Remove # if it exists
      const cleanHashtag = hashtag.startsWith('#') ? hashtag.substring(1) : hashtag;
      
      // Set search query to hashtag
      searchQuery = `#${cleanHashtag}`;
      
      // Switch to latest tab
      activeTab = 'latest';
      
      // Execute search
      executeSearch();
    } catch (error) {
      console.error('Error getting threads by hashtag:', error);
      toastStore.showToast('Failed to load hashtag results', 'error');
    }
  }
  
  onMount(() => {
    logger.debug('Explore page mounted', { authState });
    if (checkAuth()) {
      loadRecentSearches();
      fetchTrends();
    }
  });
</script>

<MainLayout
  username={sidebarUsername}
  displayName={sidebarDisplayName}
  avatar={sidebarAvatar}
  trends={trends}
  on:toggleComposeModal={() => {}}
>
  <div class="min-h-screen border-x border-gray-200 dark:border-gray-800">
    <!-- Header -->
    <div class="sticky top-0 z-10 bg-white/80 dark:bg-black/80 backdrop-blur-md border-b border-gray-200 dark:border-gray-800 px-4 py-3">
      <!-- Search component -->
      <ExploreSearch 
        {searchQuery}
        {recentSearches}
        recommendedProfiles={searchResults.top.profiles}
        {showRecentSearches}
        {isSearching}
        {isLoadingRecommendations}
        on:input={handleSearchInput}
        on:search={executeSearch}
        on:focus={handleSearchFocus}
        on:selectRecentSearch={handleRecentSearchSelect}
        on:clearRecentSearches={clearRecentSearches}
      />
      
      <!-- Filters component -->
      <ExploreFilters 
        {searchFilter}
        {selectedCategory}
        {threadCategories}
        on:filterChange={handleFilterChange}
        on:categoryChange={handleCategoryChange}
      />
    </div>
    
    <!-- Content Area -->
    <div>
      {#if hasSearched}
        <!-- Tabs for search results -->
        <ExploreTabs {activeTab} on:tabChange={handleTabChange} />
        
        <!-- Tab content -->
        {#if isSearching}
          <div class="p-4">
            <LoadingSkeleton type="search-results" />
          </div>
        {:else}
          {#if activeTab === 'top'}
            <ExploreTopResults 
              topProfiles={searchResults.top.profiles}
              topThreads={searchResults.top.threads}
              isLoading={searchResults.top.isLoading}
              on:viewAll={(event) => {
                if (event.detail === 'people') {
                  activeTab = 'people';
                }
              }}
            />
          {:else if activeTab === 'latest'}
            <ExploreLatestResults 
              latestThreads={searchResults.latest.threads}
              isLoading={searchResults.latest.isLoading}
            />
          {:else if activeTab === 'people'}
            <ExplorePeopleResults 
              peopleResults={searchResults.people.users.map(user => ({
                id: user.id,
                username: user.username,
                displayName: user.display_name || user.username,
                avatar: user.avatar || null,
                bio: user.bio,
                isVerified: user.is_verified || false,
                followerCount: (user as any).follower_count || 0,
                isFollowing: user.is_following || false
              }))}
              isLoading={searchResults.people.isLoading}
              peoplePerPage={peoplePerPage}
              on:pageChange={handlePeoplePageChange}
              on:peoplePerPageChange={handlePeoplePerPageChange}
              on:loadMore={() => handlePeoplePageChange({detail: peopleCurrentPage + 1})}
            />
          {:else if activeTab === 'media'}
            <ExploreMediaResults 
              media={searchResults.media.threads.map(thread => ({
                id: thread.id,
                media: thread.media || []
              }))}
              hasMore={searchResults.media.threads.length >= 12}
              isLoading={searchResults.media.isLoading}
              on:loadMore={loadMoreMedia}
            />
          {:else if activeTab === 'communities'}
            <ExploreCommunityResults 
              communityResults={searchResults.communities.communities.map(community => ({
                id: community.id,
                name: community.name,
                description: community.description || '',
                logo: community.logo || null,
                memberCount: community.member_count || 0,
                isJoined: community.is_joined || false,
                isPending: community.is_pending || false
              }))}
              isLoading={searchResults.communities.isLoading}
              communitiesPerPage={communitiesPerPage}
              on:pageChange={handleCommunitiesPageChange}
              on:communitiesPerPageChange={handleCommunitiesPerPageChange}
              on:loadMore={() => handleCommunitiesPageChange({detail: communitiesCurrentPage + 1})}
            />
          {/if}
        {/if}
      {:else}
        <!-- Trending section when not searching -->
        <div class="p-4">
          {#if isTrendsLoading}
            <LoadingSkeleton type="trends" />
          {:else if trends.length > 0}
            <div class="bg-gray-50 dark:bg-gray-900 rounded-xl p-4">
              <div class="flex items-center justify-between pb-3 border-b border-gray-200 dark:border-gray-800">
                <h2 class="font-bold text-xl">Trending Now</h2>
              </div>
              
              <div class="divide-y divide-gray-200 dark:divide-gray-800">
                {#each trends as trend, i}
                  <div class="py-3">
                    <div class="flex items-center justify-between">
                      <div>
                        <button 
                          class="text-blue-500 font-medium hover:underline"
                          on:click={() => getThreadsByHashtagName(trend.title)}
                        >
                          #{trend.title}
                        </button>
                        <p class="text-sm text-gray-500 dark:text-gray-400">{trend.postCount || 0} posts</p>
                      </div>
                      <span class="text-gray-500 dark:text-gray-400 text-sm">#{i + 1}</span>
                    </div>
                  </div>
                {/each}
              </div>
            </div>
          {:else}
            <p class="text-center text-gray-500 dark:text-gray-400 py-8">No trending topics available.</p>
          {/if}
        </div>
      {/if}
    </div>
  </div>
</MainLayout>

<style>
  /* Tab styles */
  .tab-active {
    color: rgb(29, 155, 240);
    font-weight: 600;
  }
  
  .tab-active::after {
    content: '';
    position: absolute;
    bottom: -1px;
    left: 0;
    right: 0;
    height: 4px;
    background-color: rgb(29, 155, 240);
    border-radius: 9999px;
  }
</style>
