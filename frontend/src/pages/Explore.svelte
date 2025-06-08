<script lang="ts">
  import { onMount } from 'svelte';
  import MainLayout from '../components/layout/MainLayout.svelte';
  import { useAuth } from '../hooks/useAuth';
  import { useTheme } from '../hooks/useTheme';
  import type { ITweet } from '../interfaces/ISocialMedia';
  import type { ITrend } from '../interfaces/ITrend';
  import type { IAuthStore } from '../interfaces/IAuth';
  import { createLoggerWithPrefix } from '../utils/logger';
  import { toastStore } from '../stores/toastStore';
  import { getTrends } from '../api/trends';
  import { searchUsers, getAllUsers } from '../api/user';
  import { searchThreads, searchThreadsWithMedia, getThreadsByHashtag } from '../api/thread';
  import { searchCommunities } from '../api/community';
  import { debounce } from '../utils/helpers';
  import { formatTimeAgo } from '../utils/common';
  
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
  import ProfileCard from '../components/explore/ProfileCard.svelte';
  import Toast from '../components/common/Toast.svelte';

  const logger = createLoggerWithPrefix('Explore');
  
  // Auth and theme
  const { getAuthState } = useAuth();
  const { theme } = useTheme();
  
  // Reactive declarations
  $: authState = getAuthState ? getAuthState() : { user_id: null, is_authenticated: false, access_token: null, refresh_token: null };
  $: isDarkMode = $theme === 'dark';
  $: sidebarUsername = authState?.user_id ? `User_${authState.user_id.substring(0, 4)}` : '';
  $: sidebarDisplayName = authState?.user_id ? `User ${authState.user_id.substring(0, 4)}` : '';
  $: sidebarAvatar = 'https://secure.gravatar.com/avatar/0?d=mp'; // Default avatar with proper image URL
  
  // Trends data
  let trends: ITrend[] = [];
  let isTrendsLoading = true;
  let suggestedFollows = [];
  
  // All Users
  let allUsers: Array<{
    id: string;
    username: string;
    displayName: string;
    avatar: string | null;
    bio?: string;
    isVerified: boolean;
    followerCount: number;
    isFollowing: boolean;
  }> = [];
  let isLoadingAllUsers = false;
  
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
        name: string;
        created_at: string;
        likes_count: number;
        replies_count: number;
        reposts_count: number;
        media?: Array<{
          type: string;
          url: string;
        }>;
        profile_picture_url?: string;
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
        displayName: string;
        avatar: string | null;
        bio?: string;
        isVerified: boolean;
        followerCount: number;
        isFollowing: boolean;
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
  
  // Store users fetched when the page loads
  let usersToDisplay: Array<{
    id: string;
    username: string;
    displayName: string;
    avatar: string | null;
    bio?: string;
    isVerified: boolean;
    followerCount: number;
    isFollowing: boolean;
  }> = [];
  let isLoadingUsers = false;
  
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
  
  // Authentication check - Updated to fix auth issue
  function checkAuth() {
    if (!authState || !authState.is_authenticated) {
      logger.error('User not authenticated, redirecting to login');
      toastStore.showToast('You need to log in to access explore', 'warning');
      setTimeout(() => {
        window.location.href = '/login';
      }, 1000);
      return false;
    }
    logger.info('Authentication check passed, user is authenticated');
    return true;
  }
  
  // Fetch all users when "Everyone" filter is selected
  async function fetchAllUsers() {
    try {
      isLoadingAllUsers = true;
      
      const { users, total } = await getAllUsers(1, 20, 'created_at', false);
      
      if (users && users.length > 0) {
        // Map backend response to the format expected by the frontend components
        allUsers = users.map(user => ({
          id: user.id,
          username: user.username,
          displayName: user.name || user.display_name || user.username,
          avatar: user.profile_picture_url || user.avatar || null,
          bio: user.bio || '',
          isVerified: user.is_verified || false,
          followerCount: user.follower_count || 0,
          isFollowing: user.is_following || false
        }));
        
        logger.debug('Fetched users:', { count: allUsers.length });
      } else {
        allUsers = [];
        logger.info('No users found');
      }
    } catch (error) {
      logger.error('Error fetching all users:', error);
      toastStore.showToast('Failed to load user recommendations', 'error');
      allUsers = [];
    } finally {
      isLoadingAllUsers = false;
    }
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
      const { users } = await searchUsers(query, 1, 3, {
        clientFuzzy: true, // Enable client-side fuzzy matching
        sort: 'follower_count' // Sort by follower count
      });
      searchResults.top.profiles = users.map(user => ({
        id: user.id,
        username: user.username,
        displayName: user.display_name || user.username,
        avatar: user.avatar,
        bio: user.bio,
        isVerified: user.is_verified || false,
        followerCount: user.follower_count || 0,
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
        searchUsers(searchQuery, 1, peoplePerPage, { 
          filter: filterOption,
          clientFuzzy: true,
          sort: 'follower_count'
        }),
        
        // Top threads
        searchThreads(searchQuery, 1, 10, {
          filter: filterOption,
          category: categoryOption,
          sort_by: 'popular'
        }),
        
        // Latest threads
        searchThreads(searchQuery, 1, 20, {
          filter: filterOption,
          category: categoryOption,
          sort_by: 'recent'
        }),
        
        // Media tab data
        searchThreadsWithMedia(searchQuery, 1, 12, {
          filter: filterOption,
          category: categoryOption
        }),
        
        // Communities tab data
        searchCommunities(searchQuery, 1, communitiesPerPage)
      ]);
      
      // Get top 3 profiles sorted by follower count for the Top tab
      const topProfiles = [...peopleData.users]
        .sort((a, b) => ((b.follower_count || 0) - (a.follower_count || 0)))
        .slice(0, 3);
      
      // Update search results
      searchResults = {
        top: {
          profiles: topProfiles.map(user => ({
            id: user.id,
            username: user.username,
            displayName: user.display_name || user.username,
            avatar: user.avatar || null,
            bio: user.bio,
            isVerified: user.is_verified || false,
            followerCount: user.follower_count || 0,
            isFollowing: user.is_following || false
          })),
          threads: topThreadsData.threads.map(thread => ({
            id: thread.id,
            content: thread.content,
            username: thread.author?.username || 'anonymous',
            name: thread.author?.display_name || 'User',
            created_at: thread.created_at || new Date().toISOString(),
            likes_count: thread.like_count || 0,
            replies_count: thread.reply_count || 0,
            reposts_count: thread.repost_count || 0,
            media: thread.media,
            profile_picture_url: thread.author?.avatar
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
            followerCount: user.follower_count || 0,
            isFollowing: user.is_following || false
          })),
          totalCount: peopleData.totalCount || 0,
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
    
    if (searchFilter === 'all' && !hasSearched) {
      // If changing to "Everyone" filter and not in search results view,
      // fetch all users to display
      fetchAllUsers();
    } else if (hasSearched) {
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
    peopleCurrentPage = page;
    searchResults.people.isLoading = true;
    
    try {
      const filterOption = searchFilter === 'following' ? 'following' : (searchFilter === 'verified' ? 'verified' : 'all');
      const { users, totalCount } = await searchUsers(
        searchQuery, 
        page, 
        peoplePerPage, 
        { 
          filter: filterOption,
          clientFuzzy: true
        }
      );
      
      searchResults.people = {
        users: users || [],
        totalCount: totalCount || 0,
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
    peoplePerPage = event.detail;
    searchResults.people.isLoading = true;
    peopleCurrentPage = 1;
    executeSearch();
  }
  
  // Pagination for communities tab
  async function handleCommunitiesPageChange(event) {
    const page = event.detail;
    communitiesCurrentPage = page;
    searchResults.communities.isLoading = true;
    
    try {
      const { communities, total_count } = await searchCommunities(searchQuery, page, communitiesPerPage);
      
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
    communitiesPerPage = event.detail;
    searchResults.communities.isLoading = true;
    communitiesCurrentPage = 1;
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
  
  onMount(async () => {
    logger.debug('Explore page mounted', { authState });
    
    // Check authentication state and initialize content if authenticated
    if (checkAuth()) {
      try {
        await Promise.all([
          loadRecentSearches(),
          fetchTrends(),
          fetchAllUsers() // Load all users on initial page load
        ]);
        
        // Set the default filter to "all" (Everyone)
        searchFilter = 'all';
        
        logger.info('Explore page initialized successfully');
      } catch (error) {
        logger.error('Error initializing explore page:', error);
        toastStore.showToast('Failed to load explore page content', 'error');
      }
    }
  });
</script>

<MainLayout
  username={sidebarUsername}
  displayName={sidebarDisplayName}
  avatar={sidebarAvatar}
  {trends}
  {suggestedFollows}
>
  <div class="explore-page {isDarkMode ? 'explore-page-dark' : ''}">
    <!-- Header -->
    <div class="explore-header {isDarkMode ? 'explore-header-dark' : ''}">
      <h1 class="explore-title">Explore</h1>
      
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
    <div class="explore-content">
      {#if hasSearched}
        <!-- Tabs for search results -->
        <ExploreTabs {activeTab} on:tabChange={handleTabChange} />
        
        <!-- Tab content -->
        {#if isSearching}
          <div class="explore-loading">
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
              peopleResults={searchResults.people.users}
              isLoading={searchResults.people.isLoading}
              peoplePerPage={peoplePerPage}
              on:pageChange={handlePeoplePageChange}
              on:peoplePerPageChange={handlePeoplePerPageChange}
              on:loadMore={() => handlePeoplePageChange({detail: peopleCurrentPage + 1})}
            />
          {:else if activeTab === 'media'}
            <ExploreMediaResults 
              media={searchResults.media.threads}
              isLoading={searchResults.media.isLoading}
              hasMore={searchResults.media.threads.length >= 12}
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
        <!-- Show trending content and all users when not searching -->
        <div class="explore-trending-section">
          <div class="explore-section">
            <h2 class="explore-section-title">What's happening</h2>
            <ExploreTrending 
              {trends}
              {isTrendsLoading}
              on:viewThreads={(event) => {
                // Handle viewing threads by hashtag
                const hashtag = event.detail;
                if (hashtag) {
                  searchQuery = hashtag;
                  executeSearch();
                }
              }}
            />
          </div>
          
          <div class="explore-section">
            <h2 class="explore-section-title">Suggested topics to follow</h2>
            <div class="explore-topic-list">
              {#each ['technology', 'programming', 'design', 'svelte', 'webdev'] as topic}
                <button 
                  class="explore-topic-chip {isDarkMode ? 'explore-topic-chip-dark' : ''}" 
                  on:click={() => {
                    searchQuery = topic;
                    executeSearch();
                  }}
                >
                  #{topic}
                </button>
              {/each}
            </div>
          </div>
          
          <!-- Added section to display all users -->
          <div class="explore-section people-to-follow-section">
            <h2 class="explore-section-title">People to follow</h2>
            {#if isLoadingAllUsers}
              <div class="explore-loading">
                <LoadingSkeleton type="profile" count={5} />
              </div>
            {:else if allUsers.length === 0}
              <div class="empty-state">
                <p>No users found</p>
              </div>
            {:else}
              <div class="users-grid">
                {#each allUsers as person}
                  <div class="user-card {isDarkMode ? 'user-card-dark' : ''}">
                    <ProfileCard profile={person} />
                  </div>
                {/each}
              </div>
            {/if}
          </div>
        </div>
      {/if}
    </div>
  </div>
</MainLayout>

<Toast />

<style>
  .explore-page {
    min-height: 100vh;
    background-color: var(--bg-primary);
    color: var(--text-primary);
  }
  
  .explore-page-dark {
    background-color: var(--bg-primary-dark);
    color: var(--text-primary-dark);
  }
  
  .explore-header {
    position: sticky;
    top: 0;
    z-index: var(--z-sticky);
    padding: var(--space-4) var(--space-4) var(--space-2);
    background-color: var(--bg-primary);
    backdrop-filter: blur(12px);
    -webkit-backdrop-filter: blur(12px);
    border-bottom: 1px solid var(--border-color);
    box-shadow: var(--shadow-sm);
  }
  
  .explore-header-dark {
    background-color: var(--bg-primary-dark);
    border-bottom: 1px solid var(--border-color-dark);
  }
  
  .explore-title {
    font-size: var(--font-size-xl);
    font-weight: var(--font-weight-bold);
    margin-bottom: var(--space-3);
    position: relative;
    display: inline-block;
  }
  
  .explore-title::after {
    content: "";
    position: absolute;
    left: 0;
    bottom: -5px;
    width: 30px;
    height: 3px;
    background-color: var(--color-primary);
    border-radius: var(--radius-full);
  }
  
  .explore-content {
    padding: var(--space-4);
  }
  
  .explore-loading {
    padding: var(--space-4);
  }
  
  .explore-trending-section {
    display: flex;
    flex-direction: column;
    gap: var(--space-6);
  }
  
  .explore-section {
    background-color: var(--bg-secondary);
    border-radius: var(--radius-lg);
    overflow: hidden;
    box-shadow: var(--shadow-sm);
    margin-bottom: var(--space-4);
  }
  
  .explore-section:last-child {
    margin-bottom: 0;
  }
  
  .explore-section-dark {
    background-color: var(--bg-secondary-dark);
    box-shadow: var(--shadow-sm-dark);
  }
  
  .explore-section-title {
    font-size: var(--font-size-lg);
    font-weight: var(--font-weight-bold);
    padding: var(--space-4);
    border-bottom: 1px solid var(--border-color);
    position: relative;
  }
  
  .explore-section-title::after {
    content: '';
    position: absolute;
    bottom: 0;
    left: var(--space-4);
    width: 40px;
    height: 3px;
    background: var(--color-primary);
    border-radius: var(--radius-full);
  }
  
  .empty-state {
    padding: var(--space-8);
    text-align: center;
    color: var(--text-secondary);
  }
  
  .explore-topic-list {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-2);
    padding: var(--space-4);
  }
  
  .explore-topic-chip {
    background-color: var(--bg-tertiary);
    color: var(--text-primary);
    padding: var(--space-2) var(--space-3);
    border-radius: var(--radius-full);
    border: 1px solid transparent;
    font-size: var(--font-size-sm);
    cursor: pointer;
    transition: all var(--transition-normal);
    box-shadow: var(--shadow-sm);
    display: flex;
    align-items: center;
    transform-origin: center;
  }
  
  .explore-topic-chip-dark {
    background-color: var(--bg-tertiary-dark);
    color: var(--text-primary-dark);
  }
  
  .explore-topic-chip:hover {
    background-color: var(--hover-bg);
    transform: translateY(-2px) scale(1.03);
    border-color: var(--color-primary);
    box-shadow: var(--shadow-md);
  }
  
  .explore-topic-chip:active {
    transform: translateY(0) scale(0.98);
  }
  
  @media (max-width: 600px) {
    .explore-trending-section {
      padding: var(--space-2);
    }
    
    .explore-section {
      padding: var(--space-3);
    }
  }
  
  /* People to follow section styling */
  .people-to-follow-section {
    margin-top: var(--space-6);
  }
  
  .users-grid {
    display: grid;
    grid-template-columns: 1fr;
    gap: 0.75rem;
    padding: 1rem;
  }
  
  @media (min-width: 640px) {
    .users-grid {
      grid-template-columns: repeat(1, 1fr);
    }
  }
  
  .user-card {
    background-color: var(--bg-secondary);
    border-radius: 1rem;
    overflow: hidden;
    transition: all 0.2s ease;
    border: 1px solid var(--border-color);
    box-shadow: 0 2px 5px rgba(0, 0, 0, 0.05);
    padding: 0.5rem 0;
  }
  
  .user-card:hover {
    transform: translateY(-2px);
    box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);
  }
  
  .user-card-dark {
    background-color: var(--bg-secondary-dark);
    border: 1px solid var(--border-color-dark);
    box-shadow: 0 2px 5px rgba(0, 0, 0, 0.2);
  }
  
  .user-card-dark:hover {
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
  }
</style>
