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
  
  const logger = createLoggerWithPrefix('Explore');
  
  // Auth and theme
  const { getAuthState } = useAuth();
  const { theme } = useTheme();
  
  // Reactive declarations
  $: authState = getAuthState ? (getAuthState() as IAuthStore) : { userId: null, isAuthenticated: false, accessToken: null, refreshToken: null };
  $: isDarkMode = $theme === 'dark';
  $: sidebarUsername = authState?.userId ? `User_${authState.userId.substring(0, 4)}` : '';
  $: sidebarDisplayName = authState?.userId ? `User ${authState.userId.substring(0, 4)}` : '';
  $: sidebarAvatar = '👤'; // Placeholder avatar
  
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
  let topProfiles: Array<{
    id: string;
    username: string;
    displayName: string;
    avatar: string | null;
    bio?: string;
    isVerified: boolean;
    followerCount: number;
    isFollowing: boolean;
  }> = [];
  
  let topThreads: Array<{
    id: string;
    content: string;
    username: string;
    displayName: string;
    timestamp: string;
    likes: number;
    replies: number;
    reposts: number;
    media?: Array<{
      type: string;
      url: string;
    }>;
    avatar?: string;
  }> = [];
  
  let latestThreads: Array<{
    id: string;
    content: string;
    username: string;
    displayName: string;
    timestamp: string;
    likes: number;
    replies: number;
    reposts: number;
    media?: Array<{
      type: string;
      url: string;
    }>;
    avatar?: string;
  }> = [];
  
  let peopleResults: Array<{
    id: string;
    username: string;
    displayName: string;
    avatar: string | null;
    bio?: string;
    isVerified: boolean;
    followerCount: number;
    isFollowing: boolean;
  }> = [];
  
  let mediaResults: Array<{
    id: string;
    threadId: string;
    url: string;
    type: string;
    content: string;
    username: string;
    media_id?: string;
  }> = [];
  
  let communityResults: Array<{
    id: string;
    name: string;
    description: string;
    logo: string | null;
    memberCount: number;
    isJoined: boolean;
    isPending: boolean;
  }> = [];
  
  // Pagination
  let peoplePage = 1;
  let peoplePerPage = 25;
  let communitiesPage = 1;
  let communitiesPerPage = 25;
  let hasMoreMedia = true;
  let mediaPage = 1;
  
  // Loading states
  let isLoadingTop = false;
  let isLoadingLatest = false;
  let isLoadingPeople = false;
  let isLoadingMedia = false;
  let isLoadingCommunities = false;
  
  // Thread categories for filtering
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
  
  // Recommended profiles while typing
  let recommendedProfiles: Array<{
    id: string;
    username: string;
    displayName: string;
    avatar: string | null;
    bio?: string;
    isVerified: boolean;
    followerCount: number;
  }> = [];
  
  // Debounce setup for search
  let searchTimeout: ReturnType<typeof setTimeout> | null = null;
  
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
  
  // Save search to recent searches
  function saveSearch(query: string) {
    if (!query.trim()) return;
    
    // Update recentSearches array
    recentSearches = [query, ...recentSearches.filter(s => s !== query)].slice(0, 3);
    
    // Save to localStorage
    try {
      localStorage.setItem('recentSearches', JSON.stringify(recentSearches));
    } catch (error) {
      console.error('Error saving recent searches:', error);
    }
  }
  
  // Clear all recent searches
  function clearRecentSearches() {
    recentSearches = [];
    localStorage.removeItem('recentSearches');
    toastStore.showToast('Search history cleared', 'success');
  }
  
  // Handle search query change with debounce
  function handleSearchInput(event) {
    const query = event.target.value;
    searchQuery = query;
    
    // Clear previous timeout
    if (searchTimeout) {
      clearTimeout(searchTimeout);
    }
    
    // If query is empty, clear recommendations
    if (!query.trim()) {
      recommendedProfiles = [];
      return;
    }
    
    // Set new timeout for debounce
    searchTimeout = setTimeout(() => {
      fetchRecommendedProfiles(query);
    }, 300); // 300ms debounce
  }
  
  // Fetch recommended profiles while typing
  async function fetchRecommendedProfiles(query: string) {
    if (!query.trim()) return;
    
    isLoadingRecommendations = true;
    try {
      // Using the searchUsers endpoint with a limit of 5
      const response = await searchUsers(query, 1, 5);
      if (response && response.users) {
        recommendedProfiles = response.users.map(user => ({
          id: user.id,
          username: user.username,
          displayName: user.name || user.username,
          avatar: user.profile_picture_url || '👤',
          bio: user.bio || '',
          isVerified: user.is_verified || false,
          followerCount: user.follower_count || 0
        }));
      }
    } catch (error) {
      console.error('Error fetching recommended profiles:', error);
      recommendedProfiles = [];
    } finally {
      isLoadingRecommendations = false;
    }
  }
  
  // Execute search
  async function executeSearch() {
    if (!searchQuery.trim()) return;
    
    isSearching = true;
    showRecentSearches = false;
    saveSearch(searchQuery);
    
    // Reset tabs and pagination
    activeTab = 'top';
    peoplePage = 1;
    communitiesPage = 1;
    mediaPage = 1;
    
    // Execute searches for all tabs
    await Promise.all([
      fetchTopResults(),
      fetchLatestResults(),
      fetchPeopleResults(),
      fetchMediaResults(),
      fetchCommunityResults()
    ]);
    
    isSearching = false;
  }
  
  // Fetch top results
  async function fetchTopResults() {
    isLoadingTop = true;
    try {
      // Fetch top profiles
      const userResponse = await searchUsers(searchQuery, 1, 3, { 
        filter: searchFilter, 
        sortBy: 'followers' 
      });
      if (userResponse && userResponse.users) {
        topProfiles = userResponse.users.map(user => ({
          id: user.id,
          username: user.username,
          displayName: user.name || user.username,
          avatar: user.profile_picture_url || '👤',
          bio: user.bio || '',
          isVerified: user.is_verified || false,
          followerCount: user.follower_count || 0,
          isFollowing: user.is_following || false
        }));
      }
      
      // Fetch top threads
      const threadResponse = await searchThreads(searchQuery, 1, 10, {
        filter: searchFilter,
        category: selectedCategory !== 'all' ? selectedCategory : undefined,
        sortBy: 'likes'
      });
      if (threadResponse && threadResponse.threads) {
        topThreads = threadResponse.threads.map(thread => ({
          id: thread.id,
          content: thread.content,
          username: thread.username || 'user',
          displayName: thread.display_name || 'User',
          timestamp: thread.created_at,
          likes: thread.like_count || 0,
          replies: thread.reply_count || 0,
          reposts: thread.repost_count || 0,
          media: thread.media || []
        }));
      }
    } catch (error) {
      console.error('Error fetching top results:', error);
      toastStore.showToast('Failed to load top results', 'error');
      topProfiles = [];
      topThreads = [];
    } finally {
      isLoadingTop = false;
    }
  }
  
  // Fetch latest results
  async function fetchLatestResults() {
    isLoadingLatest = true;
    try {
      const response = await searchThreads(searchQuery, 1, 20, {
        filter: searchFilter,
        category: selectedCategory !== 'all' ? selectedCategory : undefined,
        sortBy: 'newest'
      });
      if (response && response.threads) {
        latestThreads = response.threads.map(thread => ({
          id: thread.id,
          content: thread.content,
          username: thread.username || 'user',
          displayName: thread.display_name || 'User',
          timestamp: thread.created_at,
          likes: thread.like_count || 0,
          replies: thread.reply_count || 0,
          reposts: thread.repost_count || 0,
          media: thread.media || []
        }));
      }
    } catch (error) {
      console.error('Error fetching latest results:', error);
      toastStore.showToast('Failed to load latest results', 'error');
      latestThreads = [];
    } finally {
      isLoadingLatest = false;
    }
  }
  
  // Fetch people results
  async function fetchPeopleResults(reset = true) {
    if (reset) {
      peoplePage = 1;
      peopleResults = [];
    }
    
    isLoadingPeople = true;
    try {
      const response = await searchUsers(searchQuery, peoplePage, peoplePerPage, {
        filter: searchFilter
      });
      if (response && response.users) {
        const formattedUsers = response.users.map(user => ({
          id: user.id,
          username: user.username,
          displayName: user.name || user.username,
          avatar: user.profile_picture_url || '👤',
          bio: user.bio || '',
          isVerified: user.is_verified || false,
          followerCount: user.follower_count || 0,
          isFollowing: user.is_following || false
        }));
        
        if (reset) {
          peopleResults = formattedUsers;
        } else {
          peopleResults = [...peopleResults, ...formattedUsers];
        }
        
        // Increment page for next load
        peoplePage++;
      }
    } catch (error) {
      console.error('Error fetching people results:', error);
      toastStore.showToast('Failed to load people results', 'error');
      if (reset) {
        peopleResults = [];
      }
    } finally {
      isLoadingPeople = false;
    }
  }
  
  // Fetch media results
  async function fetchMediaResults(reset = true) {
    if (reset) {
      mediaPage = 1;
      mediaResults = [];
      hasMoreMedia = true;
    }
    
    isLoadingMedia = true;
    try {
      const response = await searchThreadsWithMedia(searchQuery, mediaPage, 15, {
        filter: searchFilter,
        category: selectedCategory !== 'all' ? selectedCategory : undefined
      });
      
      if (response && response.threads) {
        const media: typeof mediaResults = [];
        
        // Extract media from threads
        response.threads.forEach(thread => {
          if (thread.media && thread.media.length > 0) {
            thread.media.forEach(item => {
              media.push({
                id: `${thread.id}-${item.media_id || Math.random().toString(36).substr(2, 9)}`,
                threadId: thread.id,
                url: item.url,
                type: item.type,
                content: thread.content,
                username: thread.username || 'user'
              });
            });
          }
        });
        
        if (reset) {
          mediaResults = media;
        } else {
          mediaResults = [...mediaResults, ...media];
        }
        
        // Check if there are more results
        hasMoreMedia = media.length > 0;
        mediaPage++;
      } else {
        hasMoreMedia = false;
      }
    } catch (error) {
      console.error('Error fetching media results:', error);
      toastStore.showToast('Failed to load media results', 'error');
      hasMoreMedia = false;
      if (reset) {
        mediaResults = [];
      }
    } finally {
      isLoadingMedia = false;
    }
  }
  
  // Fetch community results
  async function fetchCommunityResults(reset = true) {
    if (reset) {
      communitiesPage = 1;
      communityResults = [];
    }
    
    isLoadingCommunities = true;
    try {
      const response = await searchCommunities(searchQuery, communitiesPage, communitiesPerPage);
      if (response && response.communities) {
        const formattedCommunities = response.communities.map(community => ({
          id: community.id,
          name: community.name,
          description: community.description || '',
          logo: community.logo || '👥',
          memberCount: community.member_count || 0,
          isJoined: community.is_joined || false,
          isPending: community.is_pending || false
        }));
        
        if (reset) {
          communityResults = formattedCommunities;
        } else {
          communityResults = [...communityResults, ...formattedCommunities];
        }
        
        // Increment page for next load
        communitiesPage++;
      }
    } catch (error) {
      console.error('Error fetching community results:', error);
      toastStore.showToast('Failed to load community results', 'error');
      if (reset) {
        communityResults = [];
      }
    } finally {
      isLoadingCommunities = false;
    }
  }
  
  // Handle tab change
  function handleTabChange(event) {
    activeTab = event.detail;
  }
  
  // Handle community join request
  async function requestJoinCommunity(communityId: string) {
    // TODO: Implement API call to request joining community
    toastStore.showToast('Join request sent to community', 'success');
    
    // Update the local state to show pending
    communityResults = communityResults.map(community => {
      if (community.id === communityId) {
        return { ...community, isPending: true };
      }
      return community;
    });
  }
  
  // Handle follow user
  async function followUser(userId: string) {
    // TODO: Implement API call to follow user
    toastStore.showToast('Successfully followed user', 'success');
    
    // Update local state to show following
    const updateUserFollowing = (users) => {
      return users.map(user => {
        if (user.id === userId) {
          return { ...user, isFollowing: true };
        }
        return user;
      });
    };
    
    peopleResults = updateUserFollowing(peopleResults);
    topProfiles = updateUserFollowing(topProfiles);
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
  
  // Load threads by hashtag
  async function loadTrendingThreads(hashtag: string) {
    searchQuery = hashtag;
    activeTab = 'latest';
    isSearching = true;
    
    try {
      const response = await getThreadsByHashtag(hashtag, 1, 20);
      if (response && response.threads) {
        latestThreads = response.threads.map(thread => ({
          id: thread.id,
          content: thread.content,
          username: thread.username || 'user',
          displayName: thread.display_name || 'User',
          timestamp: thread.created_at,
          likes: thread.like_count || 0,
          replies: thread.reply_count || 0,
          reposts: thread.repost_count || 0,
          media: thread.media || []
        }));
      } else {
        latestThreads = [];
      }
    } catch (error) {
      console.error('Error loading hashtag threads:', error);
      toastStore.showToast('Failed to load threads for this hashtag', 'error');
      latestThreads = [];
    } finally {
      isSearching = false;
    }
  }
  
  // Handle filter change
  function handleFilterChange(event) {
    searchFilter = event.detail;
    if (searchQuery) {
      executeSearch();
    }
  }
  
  // Handle category change
  function handleCategoryChange(event) {
    selectedCategory = event.detail;
    if (searchQuery) {
      executeSearch();
    }
  }
  
  // Handle people per page change
  function handlePeoplePerPageChange(event) {
    peoplePerPage = event.detail;
    fetchPeopleResults();
  }
  
  // Handle communities per page change
  function handleCommunitiesPerPageChange(event) {
    communitiesPerPage = event.detail;
    fetchCommunityResults();
  }
  
  // Authentication check
  function checkAuth() {
    if (!authState.isAuthenticated) {
      toastStore.showToast('You need to log in to access explore', 'warning');
      window.location.href = '/login';
      return false;
    }
    return true;
  }
  
  // Handle scroll for media infinite loading
  function handleScroll(event) {
    if (activeTab !== 'media' || !hasMoreMedia || isLoadingMedia) return;
    
    const { scrollTop, scrollHeight, clientHeight } = event.target;
    
    // If scrolled near the bottom, load more media
    if (scrollHeight - scrollTop - clientHeight < 200) {
      fetchMediaResults(false);
    }
  }
  
  // Handle events from search component
  function handleSearchFocus() {
    showRecentSearches = true;
  }
  
  function handleRecentSearchSelect(event) {
    searchQuery = event.detail;
    executeSearch();
  }
  
  // Handle view all button from top results
  function handleViewAll(event) {
    if (event.detail === 'people') {
      activeTab = 'people';
    }
  }
  
  onMount(() => {
    if (checkAuth()) {
      loadRecentSearches();
      fetchTrends();
    }
  });
</script>