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
  import { searchUsers, getAllUsers, getFollowing } from '../api/user';
  import { searchThreads, searchThreadsWithMedia, getThreadsByHashtag } from '../api/thread';
  import { searchCommunities, getCommunities } from '../api/community';
  import { debounce, stringSimilarity } from '../utils/helpers';
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
  import CommunityCard from '../components/explore/CommunityCard.svelte';
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
      pagination: {
        total_count: number;
        current_page: number;
        total_pages: number;
      };
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
      totalCount: number;
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
      pagination: {
        total_count: 0,
        current_page: 1,
        total_pages: 0
      },
      isLoading: false
    },
    media: {
      threads: [],
      totalCount: 0,
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
  
  // State variables for communities display
  let communitiesToDisplay: Array<{
    id: string;
    name: string;
    description?: string;
    logo?: string | null;
    member_count?: number;
    is_joined?: boolean;
    is_pending?: boolean;
  }> = [];
  let isLoadingCommunities = false;
  let defaultActiveTab: 'people' | 'communities' = 'people';
  
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
    isLoadingUsers = true;
    try {
      const response = await getAllUsers(1, 20, 'created_at', false);
      
      console.log('fetchAllUsers response:', response);
      const users = response.users || [];
      
      if (users && users.length > 0) {
        // Map backend response to the format expected by the frontend components
        usersToDisplay = users.map(user => ({
          id: user.id,
          username: user.username,
          displayName: user.name || user.display_name || user.username,
          avatar: user.profile_picture_url || user.avatar || null,
          bio: user.bio || '',
          isVerified: user.is_verified || false,
          followerCount: user.follower_count || 0,
          isFollowing: user.is_following || false
        }));
        
        console.log('Mapped users for display:', usersToDisplay);
        logger.debug('Fetched users:', { count: usersToDisplay.length });
      } else {
        usersToDisplay = [];
        logger.info('No users found');
      }
    } catch (error) {
      logger.error('Error fetching all users:', error);
      toastStore.showToast('Failed to load users', 'error');
      usersToDisplay = [];
    } finally {
      isLoadingUsers = false;
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
      const trendData = await getTrends(10); // Explicitly get top 10 trending hashtags
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
      const { users } = await searchUsers(query.trim(), 1, 10, {
        clientFuzzy: true, // Enable client-side fuzzy matching
        sort: 'follower_count' // Sort by follower count
      });
      
      // Apply Damerau-Levenshtein distance for fuzzy matching
      // Filter and sort users based on relevance
      const relevantUsers = users
        .map(user => {
          // Calculate similarity based on username and display name
          const usernameSimilarity = stringSimilarity(
            query.toLowerCase(),
            user.username.toLowerCase()
          );
          
          const displayNameSimilarity = stringSimilarity(
            query.toLowerCase(),
            (user.display_name || '').toLowerCase()
          );
          
          // Use the better match of the two
          const similarity = Math.max(usernameSimilarity, displayNameSimilarity);
          
          return {
            user,
            similarity
          };
        })
        .filter(item => item.similarity > 0.3) // Only include relevant matches
        .sort((a, b) => b.similarity - a.similarity) // Sort by similarity (highest first)
        .slice(0, 3); // Take top 3 matches
      
      searchResults.top.profiles = relevantUsers.map(item => ({
        id: item.user.id,
        username: item.user.username,
        displayName: item.user.display_name || item.user.username,
        avatar: item.user.avatar,
        bio: item.user.bio,
        isVerified: item.user.is_verified || false,
        followerCount: item.user.follower_count || 0,
        isFollowing: item.user.is_following || false
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
  
  // Handle search from search component
  function handleSearch(event) {
    // If we're on the communities tab in non-search mode, start with the communities tab
    if (!hasSearched && defaultActiveTab === 'communities') {
      activeTab = 'communities';
    }
    executeSearch();
  }
  
  // Handle clear search input
  function handleClearSearch() {
    searchQuery = '';
    hasSearched = false;
    showRecentSearches = false;
  }
  
  // Handle toggle follow for a user
  function handleToggleFollow(userId) {
    // Find the user in usersToDisplay
    const userIndex = usersToDisplay.findIndex(user => user.id === userId);
    if (userIndex !== -1) {
      // Toggle the following state
      const user = usersToDisplay[userIndex];
      const updatedUser = {...user, isFollowing: !user.isFollowing};
      
      // Update the array
      usersToDisplay = [
        ...usersToDisplay.slice(0, userIndex),
        updatedUser,
        ...usersToDisplay.slice(userIndex + 1)
      ];
      
      // Call the API in the background
      handleFollowUser({ detail: userId });
    }
  }
  
  // Main search function
  async function executeSearch() {
    // Handle empty search query - still perform search to show all items
    if (!searchQuery || searchQuery.trim() === '') {
      logger.debug('Empty search query, showing all items');
    } else {
      // Add to recent searches if not already there
      if (!recentSearches.includes(searchQuery)) {
        recentSearches = [searchQuery, ...recentSearches.slice(0, 2)];
        localStorage.setItem('recentSearches', JSON.stringify(recentSearches));
      }
    }
    
    isSearching = true;
    hasSearched = true;
    showRecentSearches = false;
    
    try {
      const filterOption = searchFilter === 'following' ? 'following' : (searchFilter === 'verified' ? 'verified' : 'all');
      const categoryOption = selectedCategory !== 'all' ? selectedCategory : undefined;
      
      logger.debug('Starting search', {
        query: searchQuery || '(empty - showing all)',
        filter: filterOption,
        category: categoryOption,
        sortBy: 'popular' 
      });
      
      // Safely execute each API call with error handling
      const safeApiCall = async (apiFunction, ...args) => {
        try {
          const result = await apiFunction(...args);
          
          // Check if the result is null or undefined
          if (result === null || result === undefined) {
            logger.warn(`API call ${apiFunction.name} returned null or undefined`);
            return getDefaultResultForFunction(apiFunction.name);
          }
          
          return result;
        } catch (error) {
          logger.error(`Error in API call ${apiFunction.name}:`, error);
          // Return a safe default value based on expected structure
          return getDefaultResultForFunction(apiFunction.name);
        }
      };
      
      // Helper function to get default results for different API functions
      function getDefaultResultForFunction(functionName) {
        switch(functionName) {
          case 'searchUsers':
            return { users: [], totalCount: 0 };
          case 'searchThreads':
          case 'searchThreadsWithMedia':  
            return { threads: [] };
          case 'searchCommunities':
            return { 
              communities: [], 
              total_count: 0,
              pagination: {
                total_count: 0,
                current_page: 1,
                per_page: communitiesPerPage,
                total_pages: 0
              }
            };
          default:
            return {};
        }
      }
      
      // Fetch data for all tabs in parallel with error handling
      const [peopleData, topThreadsData, latestThreadsData, mediaData, communitiesData] = await Promise.all([
        // People tab data (also used for top profiles)
        safeApiCall(searchUsers, searchQuery, 1, peoplePerPage, { 
          filter: filterOption,
          clientFuzzy: true,
          sort: 'follower_count'
        }),
        
        // Top threads
        safeApiCall(searchThreads, searchQuery, 1, 10, {
          filter: filterOption,
          category: categoryOption,
          sort_by: 'popular'
        }),
        
        // Latest threads
        safeApiCall(searchThreads, searchQuery, 1, 20, {
          filter: filterOption,
          category: categoryOption,
          sort_by: 'recent'
        }),
        
        // Media tab data
        safeApiCall(searchThreadsWithMedia, searchQuery, 1, 12, {
          filter: filterOption,
          category: categoryOption
        }),
        
        // Communities tab data
        safeApiCall(searchCommunities, searchQuery, 1, communitiesPerPage)
          .catch(error => {
            logger.error('Failed to search communities:', error);
            return {
              communities: [],
              total_count: 0,
              pagination: {
                total_count: 0,
                current_page: 1,
                per_page: communitiesPerPage,
                total_pages: 0
              }
            };
          })
      ]);
      
      // Process people results
      console.log("Raw people data:", peopleData);
      
      // Add safety check to ensure peopleData has the expected structure
      let processedPeopleData = peopleData;
      if (!processedPeopleData || typeof processedPeopleData !== 'object') {
        console.error("Invalid people data received:", processedPeopleData);
        processedPeopleData = { users: [], totalCount: 0 };
      }
      
      const peopleUsers = (processedPeopleData.users || []).map(user => {
        // Log each user for debugging
        console.log("Processing user:", user);
        
        if (!user || typeof user !== 'object') {
          console.error("Invalid user object:", user);
          return null;
        }
        
        return {
          id: user.id || "",
          username: user.username || "",
          displayName: user.name || user.display_name || user.username || "",
          avatar: user.profile_picture_url || user.avatar || null,
          bio: user.bio || '',
          isVerified: user.is_verified || false,
          followerCount: user.follower_count || 0,
          isFollowing: user.is_following || false
        };
      }).filter(user => user !== null); // Filter out any null entries from invalid data
      
      console.log("Mapped people users:", peopleUsers);
      
      // Process top threads results
      const topThreads = (topThreadsData.threads || []).map(thread => ({
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
      }));
      
      // Get top 3 profiles sorted by follower count for the Top tab
      const topProfiles = [...peopleUsers]
        .sort((a, b) => ((b.follower_count || 0) - (a.follower_count || 0)))
        .slice(0, 3);
      
      // Process communities data
      console.log("Raw communities data:", communitiesData);
      
      // Add safety check to ensure communitiesData has the expected structure
      let processedCommunitiesData = communitiesData;
      if (!processedCommunitiesData || typeof processedCommunitiesData !== 'object') {
        console.error("Invalid communities data received:", processedCommunitiesData);
        processedCommunitiesData = { communities: [], total_count: 0 };
      }
      
      // Process community items to ensure proper format
      const normalizedCommunities = (processedCommunitiesData.communities || []).map(community => {
        if (!community || typeof community !== 'object') {
          console.error("Invalid community object:", community);
          return null;
        }
        
        return {
          id: community.id || '',
          name: community.name || '',
          description: community.description || '',
          logo: community.logo || community.logo_url || community.avatar || null,
          member_count: community.member_count || community.memberCount || 0,
          is_joined: community.is_joined || community.isJoined || false,
          is_pending: community.is_pending || community.isPending || false
        };
      }).filter(community => community !== null);
      
      // Update search results
      searchResults = {
        top: {
          profiles: topProfiles,
          threads: topThreads,
          isLoading: false
        },
        latest: {
          threads: (latestThreadsData?.threads || []).map(thread => ({
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
          })),
          isLoading: false
        },
        people: {
          users: peopleUsers,
          totalCount: processedPeopleData?.totalCount || processedPeopleData?.total || 0,
          pagination: processedPeopleData?.pagination || { 
            total_count: processedPeopleData?.totalCount || processedPeopleData?.total || 0,
            current_page: processedPeopleData?.currentPage || 1,
            total_pages: Math.ceil((processedPeopleData?.totalCount || processedPeopleData?.total || 0) / peoplePerPage)
          },
          isLoading: false
        },
        media: {
          threads: [...(mediaData?.threads || [])],
          totalCount: mediaData?.total_count || mediaData?.total || 0,
          isLoading: false
        },
        communities: {
          communities: normalizedCommunities,
          totalCount: processedCommunitiesData?.total_count || processedCommunitiesData?.total || 0,
          isLoading: false
        }
      };
      
      // Update the usersToDisplay for the People tab
      usersToDisplay = peopleUsers;
      
      logger.debug('Search completed', {
        query: searchQuery,
        filter: filterOption,
        peopleCount: processedPeopleData?.users?.length || 0,
        threadsCount: topThreadsData?.threads?.length || 0,
        mediaCount: mediaData?.threads?.length || 0,
        totalPeopleCount: processedPeopleData?.totalCount || processedPeopleData?.total || 0
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
    const newFilter = event.detail;
    logger.debug('Filter changed', { from: searchFilter, to: newFilter });
    console.log('Filter changed from', searchFilter, 'to', newFilter);
    
    // Update filter state
    searchFilter = newFilter;
    
    // Handle based on current state
    if (hasSearched && searchQuery.trim() !== '') {
      // If user has already searched, immediately re-execute with new filter
      console.log('In search mode, re-executing search with new filter');
      executeSearch();
    } else {
      // If not in search results view, fetch data based on selected filter and active tab
      console.log('Not in search mode, fetching data for filter:', newFilter, 'and tab:', defaultActiveTab);
      if (defaultActiveTab === 'people') {
        isLoadingUsers = true;
        if (searchFilter === 'all') {
          // Fetch all users if filter is set to "Everyone"
          console.log('Fetching all users');
          fetchAllUsers();
        } else if (searchFilter === 'following') {
          // Fetch followed users if filter is set to "Following"
          console.log('Fetching followed users');
          fetchFollowedUsers();
        } else if (searchFilter === 'verified') {
          // Fetch verified users if filter is set to "Verified"
          console.log('Fetching verified users');
          fetchVerifiedUsers();
        }
      } else if (defaultActiveTab === 'communities') {
        isLoadingCommunities = true;
        // For communities, we currently only support fetching all communities
        // In a production app, you would implement filter options for communities too
        fetchAllCommunities();
      }
    }
  }
  
  // Fetch followed users for the "following" filter when no search has been done
  async function fetchFollowedUsers() {
    if (!authState.is_authenticated) {
      toastStore.showToast('You need to be logged in to view followed users', 'warning');
      return;
    }
    
    isLoadingUsers = true;
    try {
      // Get the current user's ID
      const userId = authState.user_id;
      console.log("Fetching following users for:", userId);
      
      // Use getFollowing API instead of searchUsers for the following filter
      const response = await getFollowing(userId || '', 1, 20);
      console.log("Following API response:", response);
      
      // Define type for user objects
      interface UserObject {
        id: string;
        username: string;
        name?: string;
        display_name?: string;
        profile_picture_url?: string;
        avatar?: string;
        bio?: string;
        is_verified?: boolean;
        follower_count?: number;
        [key: string]: any;
      }
      
      let followingUsers: UserObject[] = [];
      
      // Handle different response structures
      if (response.data && response.data.following) {
        followingUsers = response.data.following as UserObject[];
      } else if (response.following) {
        followingUsers = response.following as UserObject[];
      }
      
      console.log("Found following users:", followingUsers);
      
      usersToDisplay = followingUsers.map(user => ({
        id: user.id,
        username: user.username,
        displayName: user.name || user.display_name || user.username,
        avatar: user.profile_picture_url || user.avatar || null,
        bio: user.bio || '',
        isVerified: user.is_verified || false,
        followerCount: user.follower_count || 0,
        isFollowing: true
      }));
    } catch (error) {
      console.error('Error fetching followed users:', error);
      toastStore.showToast('Failed to load followed users', 'error');
      usersToDisplay = [];
    } finally {
      isLoadingUsers = false;
    }
  }
  
  // Fetch verified users for the "verified" filter when no search has been done
  async function fetchVerifiedUsers() {
    isLoadingUsers = true;
    try {
      console.log('Fetching verified users...');
      
      // Instead of using searchUsers, use getAllUsers and then filter for verified users
      const response = await getAllUsers(1, 50, 'created_at', false);
      console.log('All users response for verified filtering:', response);
      
      // Extract users and filter for only verified ones
      const allUsers = response.users || [];
      const verifiedUsers = allUsers.filter(user => user.is_verified === true);
      
      console.log('Filtered verified users:', verifiedUsers);
      
      if (verifiedUsers.length > 0) {
        // Map backend response to the format expected by the frontend components
        usersToDisplay = verifiedUsers.map(user => ({
          id: user.id,
          username: user.username,
          displayName: user.name || user.display_name || user.username,
          avatar: user.profile_picture_url || user.avatar || null,
          bio: user.bio || '',
          isVerified: true, // We know these users are verified
          followerCount: user.follower_count || 0,
          isFollowing: user.is_following || false
        }));
        
        console.log('Mapped verified users for display:', usersToDisplay);
        logger.debug('Fetched verified users:', { count: usersToDisplay.length });
      } else {
        console.log('No verified users found');
        usersToDisplay = [];
        
        // If no verified users found through filtering, try the searchUsers approach as backup
        try {
          const searchResponse = await searchUsers('', 1, 20, { filter: 'verified' });
          console.log('Search API verified users response:', searchResponse);
          
          if (searchResponse && searchResponse.users && Array.isArray(searchResponse.users) && searchResponse.users.length > 0) {
            usersToDisplay = searchResponse.users.map(user => ({
              id: user.id || '',
              username: user.username || '',
              displayName: user.name || user.display_name || user.username || '',
              avatar: user.profile_picture_url || user.avatar || null,
              bio: user.bio || '',
              isVerified: true,
              followerCount: user.follower_count || 0,
              isFollowing: user.is_following || false
            }));
            console.log('Found verified users from search API:', usersToDisplay);
          }
        } catch (searchError) {
          console.error('Backup search method for verified users failed:', searchError);
        }
      }
    } catch (error) {
      logger.error('Error fetching verified users:', error);
      toastStore.showToast('Failed to load verified users', 'error');
      usersToDisplay = [];
    } finally {
      isLoadingUsers = false;
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
    
    // If this is a communities or people tab, also update the default tab
    // so when exiting search mode, we're on the right tab
    if (activeTab === 'communities' || activeTab === 'people') {
      defaultActiveTab = activeTab;
    }
  }
  
  // Handle clicking on a trending hashtag
  async function handleHashtagClick(event) {
    const hashtag = event.detail;
    logger.debug('Hashtag clicked', { hashtag });
    
    try {
      // Show loading state
      hasSearched = true;
      isSearching = true;
      searchQuery = hashtag;
      
      // Get threads for this hashtag sorted by likes
      const hashtagThreadsData = await getThreadsByHashtag(hashtag, 1, 20);
      
      // Update the latest tab with these results
      searchResults.latest.threads = (hashtagThreadsData?.threads || []).map(thread => ({
        id: thread.id,
        content: thread.content,
        username: thread.username || thread.author?.username || 'anonymous',
        displayName: thread.name || thread.author?.display_name || 'User',
        timestamp: thread.created_at || new Date().toISOString(),
        likes: thread.likes_count || thread.like_count || 0,
        replies: thread.replies_count || thread.reply_count || 0,
        reposts: thread.reposts_count || thread.repost_count || 0,
        media: thread.media,
        avatar: thread.profile_picture_url || thread.author?.avatar
      }));
      
      // Also update the top results with properly formatted data for that component
      searchResults.top.threads = (hashtagThreadsData?.threads || []).map(thread => ({
        id: thread.id,
        content: thread.content,
        username: thread.username || thread.author?.username || 'anonymous',
        name: thread.name || thread.author?.display_name || 'User',
        created_at: thread.created_at || new Date().toISOString(),
        likes_count: thread.likes_count || thread.like_count || 0,
        replies_count: thread.replies_count || thread.reply_count || 0,
        reposts_count: thread.reposts_count || thread.repost_count || 0,
        media: thread.media,
        profile_picture_url: thread.profile_picture_url || thread.author?.avatar
      })).slice(0, 5);
      
      // Set active tab to latest to show the results
      activeTab = 'latest';
      
    } catch (error) {
      console.error('Error fetching hashtag threads:', error);
      toastStore.showToast('Failed to load hashtag threads', 'error');
    } finally {
      isSearching = false;
    }
  }
  
  // Pagination for people tab
  async function handlePeoplePageChange(event) {
    const page = event.detail;
    peopleCurrentPage = page;
    searchResults.people.isLoading = true;
    
    try {
      const filterOption = searchFilter === 'following' ? 'following' : (searchFilter === 'verified' ? 'verified' : 'all');
      
      console.log(`Fetching people page ${page} with filter ${filterOption}, query: '${searchQuery}'`);
      
      const response = await searchUsers(
        searchQuery, 
        page, 
        peoplePerPage, 
        { 
          filter: filterOption,
          clientFuzzy: true
        }
      );
      
      console.log('People pagination response:', response);
      
      // Safety check for users array
      if (!response || !response.users) {
        console.error("Invalid response format from searchUsers:", response);
        searchResults.people.isLoading = false;
        return;
      }
      
      // Map the API response to the format expected by the component
      const mappedUsers = response.users.map(user => {
        console.log('Mapping user:', user);
        
        if (!user || typeof user !== 'object') {
          console.error("Invalid user object:", user);
          return null;
        }
        
        return {
          id: user.id || '',
          username: user.username || '',
          displayName: user.name || user.display_name || user.username || '',
          avatar: user.profile_picture_url || user.avatar || null,
          bio: user.bio || '',
          isVerified: user.is_verified || false,
          followerCount: user.follower_count || 0,
          isFollowing: user.is_following || false
        };
      }).filter(user => user !== null); // Filter out any null entries from invalid data
      
      console.log('Mapped users:', mappedUsers);
      
      // Update the displayed users
      searchResults.people = {
        users: mappedUsers,
        totalCount: response.totalCount || response.total || 0,
        pagination: response.pagination || {
          total_count: response.totalCount || response.total || 0,
          current_page: page,
          total_pages: Math.ceil((response.totalCount || response.total || 0) / peoplePerPage)
        },
        isLoading: false
      };
      
      // Also update the main users display array
      usersToDisplay = mappedUsers;
      
      console.log(`Updated people results: ${mappedUsers.length} users, total: ${response.totalCount || response.total || 0}`);
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
        totalCount: data.total_count || data.total || searchResults.media.totalCount || 0,
        isLoading: false
      };
    } catch (error) {
      console.error('Error loading more media:', error);
      toastStore.showToast('Failed to load more media', 'error');
      searchResults.media.isLoading = false;
    }
  }
  
  // Handle follow user
  function handleFollowUser(event) {
    const userId = event.detail;
    logger.debug('Follow user requested', { userId });
    // Implement follow user logic here if needed
    // For now, just log the action
    toastStore.showToast('Follow feature will be implemented soon', 'info');
  }
  
  // Handle profile click
  function handleProfileClick(event) {
    const userId = event.detail;
    logger.debug('Profile clicked', { userId });
    // Navigate to user profile
    window.location.href = `/user/${userId}`;
  }
  
  // Handle join community
  function handleJoinCommunity(event) {
    const { communityId } = event.detail;
    logger.debug('Join community requested', { communityId });
    // Implement join community logic here
    toastStore.showToast('Join community feature will be implemented soon', 'info');
  }
  
  // Fetch all communities when "Everyone" filter is selected
  async function fetchAllCommunities() {
    isLoadingCommunities = true;
    try {
      const params = {
        page: 1,
        limit: communitiesPerPage,
        is_approved: true
      };
      
      const response = await getCommunities(params);
      
      console.log('fetchAllCommunities response:', response);
      const communities = response.communities || [];
      
      if (communities && communities.length > 0) {
        // Map backend response to the format expected by the frontend components
        communitiesToDisplay = communities.map(community => ({
          id: community.id || '',
          name: community.name || '',
          description: community.description || '',
          logo: community.logo_url || null,
          member_count: community.member_count || 0,
          // These properties might not exist in the getCommunities response
          is_joined: false, // Default to false since we can't know from getCommunities
          is_pending: false // Default to false since we can't know from getCommunities
        }));
        
        console.log('Mapped communities for display:', communitiesToDisplay);
        logger.debug('Fetched communities:', { count: communitiesToDisplay.length });
      } else {
        communitiesToDisplay = [];
        logger.info('No communities found');
      }
    } catch (error) {
      logger.error('Error fetching all communities:', error);
      toastStore.showToast('Failed to load communities', 'error');
      communitiesToDisplay = [];
    } finally {
      isLoadingCommunities = false;
    }
  }
  
  // Handle tab change for default view
  function handleDefaultTabChange(newTab) {
    defaultActiveTab = newTab;
    logger.debug('Default tab changed', { tab: defaultActiveTab });
    
    // Load data for the selected tab if needed
    if (defaultActiveTab === 'communities' && communitiesToDisplay.length === 0 && !isLoadingCommunities) {
      fetchAllCommunities();
    }
  }
  
  onMount(async () => {
    logger.debug('Explore page mounted', { authState });
    console.log('Explore page mounted, auth state:', authState);
    
    // Check authentication state and initialize content if authenticated
    if (checkAuth()) {
      try {
        // Set the default filter to "all" (Everyone)
        searchFilter = 'all';
        console.log('Setting default filter to:', searchFilter);
        
        // Load initial user list using the "all" filter
        console.log('Loading initial users with filter:', searchFilter);
        await fetchAllUsers();
        
        // Load initial communities list
        await fetchAllCommunities();
        
        logger.info('Explore page initialized successfully');
      } catch (error) {
        logger.error('Error initializing explore page:', error);
        toastStore.showToast('Failed to load explore page content', 'error');
      }
    }
  });
</script>

<!-- Enhanced page layout using full width -->
<MainLayout 
  username={sidebarUsername}
  displayName={sidebarDisplayName}
  avatar={sidebarAvatar}
  trends={trends}
  suggestedFollows={suggestedFollows}
  pageTitle="Explore"
>
  <div class="explore-page-content {isDarkMode ? 'explore-page-content-dark' : ''}">
    <!-- Page header with search -->
    <div class="page-header {isDarkMode ? 'page-header-dark' : ''}">
      <h1 class="page-title {isDarkMode ? 'page-title-dark' : ''}">Explore</h1>
      <p class="page-subtitle {isDarkMode ? 'page-subtitle-dark' : ''}">Discover people, communities, and conversations</p>
      
      <!-- Search bar -->
      <div class="search-container">
        <ExploreSearch 
          bind:searchQuery={searchQuery}
          bind:showRecentSearches={showRecentSearches}
          recentSearches={recentSearches}
          on:search={handleSearch}
          on:input={handleSearchInput}
          on:focus={handleSearchFocus}
          on:selectRecentSearch={handleRecentSearchSelect}
          on:clearRecentSearches={clearRecentSearches}
          on:clearSearch={handleClearSearch}
        />
      </div>
    </div>
    
    <!-- Filters in a dedicated container -->
    <div class="filter-container {isDarkMode ? 'filter-container-dark' : ''}">
      <ExploreFilters 
        bind:searchFilter={searchFilter}
        bind:selectedCategory={selectedCategory}
        threadCategories={threadCategories}
        on:filterChange={handleFilterChange}
        on:categoryChange={handleCategoryChange}
      />
    </div>
    
    <!-- Tab navigation for search results -->
    {#if hasSearched}
      <div class="tabs-container {isDarkMode ? 'tabs-container-dark' : ''}">
        <ExploreTabs 
          activeTab={activeTab}
          on:tabChange={handleTabChange}
        />
      </div>
      
      <!-- Search results based on active tab -->
      <div class="search-results-container {isDarkMode ? 'search-results-container-dark' : ''}">
        {#if activeTab === 'top'}
          <ExploreTopResults 
            topProfiles={searchResults.top.profiles}
            topThreads={searchResults.top.threads}
            isLoading={searchResults.top.isLoading}
            on:profileClick={handleProfileClick}
            on:follow={handleFollowUser}
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
            totalCount={searchResults.people.totalCount}
            peoplePerPage={peoplePerPage}
            currentPage={peopleCurrentPage}
            on:pageChange={handlePeoplePageChange}
            on:peoplePerPageChange={handlePeoplePerPageChange}
            on:follow={handleFollowUser}
            on:profileClick={handleProfileClick}
          />
        {:else if activeTab === 'media'}
          <!-- Media results component -->
          <ExploreMediaResults 
            media={searchResults.media.threads}
            isLoading={searchResults.media.isLoading}
            hasMore={mediaPage * 12 < searchResults.media.totalCount}
            on:loadMore={loadMoreMedia}
          />
        {:else if activeTab === 'communities'}
          <!-- Communities results component -->
          <ExploreCommunityResults 
            communityResults={searchResults.communities.communities}
            isLoading={searchResults.communities.isLoading}
            totalCount={searchResults.communities.totalCount}
            communitiesPerPage={communitiesPerPage}
            currentPage={communitiesCurrentPage}
            on:pageChange={handleCommunitiesPageChange}
            on:communitiesPerPageChange={handleCommunitiesPerPageChange}
            on:joinRequest={handleJoinCommunity}
          />
        {/if}
      </div>
    {:else}
      <!-- When not searching, show tabs to select between People and Communities -->
      <div class="tabs-container {isDarkMode ? 'tabs-container-dark' : ''}">
        <div class="flex border-b border-gray-200 dark:border-gray-700">
          <button 
            class="py-3 px-6 font-medium {defaultActiveTab === 'people' ? 'text-blue-500 border-b-2 border-blue-500' : 'text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300'}"
            on:click={() => handleDefaultTabChange('people')}
          >
            People
          </button>
          <button 
            class="py-3 px-6 font-medium {defaultActiveTab === 'communities' ? 'text-blue-500 border-b-2 border-blue-500' : 'text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300'}"
            on:click={() => handleDefaultTabChange('communities')}
          >
            Communities
          </button>
        </div>
      </div>
      
      <!-- Content based on selected default tab -->
      <div class="section-container {isDarkMode ? 'section-container-dark' : ''}">
        {#if defaultActiveTab === 'people'}
          <h2 class="section-title {isDarkMode ? 'section-title-dark' : ''}">People</h2>
          
          {#if isLoadingUsers}
            <div class="loading-container">
              <LoadingSkeleton type="profile" count={6} />
            </div>
          {:else if usersToDisplay.length > 0}
            <div class="grid-container">
              {#each usersToDisplay as user (user.id)}
                <div class="card {isDarkMode ? 'card-dark' : ''}">
                  <ProfileCard
                    id={user.id}
                    username={user.username}
                    displayName={user.displayName}
                    avatar={user.avatar}
                    bio={user.bio}
                    isVerified={user.isVerified}
                    followerCount={user.followerCount}
                    isFollowing={user.isFollowing}
                    onToggleFollow={() => handleToggleFollow(user.id)}
                  />
                </div>
              {/each}
            </div>
          {:else}
            <div class="empty-state {isDarkMode ? 'empty-state-dark' : ''}">
              <p class="empty-state-message">No users found. Try a different filter.</p>
            </div>
          {/if}
        {:else if defaultActiveTab === 'communities'}
          <h2 class="section-title {isDarkMode ? 'section-title-dark' : ''}">Communities</h2>
          
          {#if isLoadingCommunities}
            <div class="loading-container">
              <LoadingSkeleton type="community" count={6} />
            </div>
          {:else if communitiesToDisplay.length > 0}
            <div class="space-y-4">
              {#each communitiesToDisplay as community (community.id)}
                <CommunityCard {community} on:joinRequest={handleJoinCommunity} />
              {/each}
            </div>
          {:else}
            <div class="empty-state {isDarkMode ? 'empty-state-dark' : ''}">
              <p class="empty-state-message">No communities found. Try a different filter.</p>
            </div>
          {/if}
        {/if}
      </div>
    {/if}
  </div>
</MainLayout>

<Toast />

<style>
  /* Explore page styles */
  .explore-page-content {
    width: 100%;
    background-color: var(--bg-primary);
    color: var(--text-primary);
  }
  
  .explore-page-content-dark {
    background-color: var(--dark-bg-primary);
    color: var(--dark-text-primary);
  }
  
  /* Page header styles */
  .page-header {
    padding: var(--space-5) var(--space-4);
    background-color: var(--bg-primary);
    border-bottom: 1px solid var(--border-color);
    position: relative;
  }
  
  .page-header-dark {
    background-color: var(--dark-bg-primary);
    border-bottom: 1px solid var(--border-color-dark);
  }
  
  .page-title {
    font-size: var(--font-size-xxl);
    font-weight: var(--font-weight-bold);
    margin: 0 0 var(--space-2);
    color: var(--text-primary);
  }
  
  .page-title-dark {
    color: var(--dark-text-primary);
  }
  
  .page-subtitle {
    font-size: var(--font-size-md);
    margin: 0 0 var(--space-4);
    color: var(--text-secondary);
    max-width: 600px;
  }
  
  .page-subtitle-dark {
    color: var(--dark-text-secondary);
  }
  
  .search-container {
    width: 100%;
    max-width: 600px;
  }

  /* Filter container */
  .filter-container {
    background-color: var(--bg-primary);
  }
  
  .filter-container-dark {
    background-color: var(--dark-bg-primary);
  }
  
  /* Tabs container */
  .tabs-container {
    background-color: var(--bg-primary);
    border-bottom: 1px solid var(--border-color);
  }
  
  .tabs-container-dark {
    background-color: var(--dark-bg-primary);
    border-bottom: 1px solid var(--border-color-dark);
  }
  
  /* Section layout */
  .section-container {
    padding: var(--space-4);
    background-color: var(--bg-primary);
  }
  
  .section-container-dark {
    background-color: var(--dark-bg-primary);
  }
  
  .section-title {
    font-size: var(--font-size-xl);
    font-weight: var(--font-weight-bold);
    margin-bottom: var(--space-4);
    color: var(--text-primary);
  }
  
  .section-title-dark {
    color: var(--dark-text-primary);
  }
  
  /* Grid layout */
  .grid-container {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
    gap: var(--space-3);
    width: 100%;
  }
  
  /* Loading state */
  .loading-container {
    padding: var(--space-4);
  }
  
  /* Card styling */
  .card {
    background-color: var(--bg-secondary);
    border-radius: var(--radius-md);
    overflow: hidden;
    transition: transform 0.2s ease, box-shadow 0.2s ease;
    border: 1px solid var(--border-color);
    padding: var(--space-1);
  }
  
  .card:hover {
    transform: translateY(-1px);
    box-shadow: var(--shadow-sm);
  }
  
  .card-dark {
    background-color: var(--dark-bg-secondary);
    border: 1px solid var(--border-color-dark);
  }
  
  /* Empty state */
  .empty-state {
    padding: var(--space-8) var(--space-4);
    text-align: center;
    background-color: var(--bg-secondary);
    border-radius: var(--radius-lg);
    border: 1px solid var(--border-color);
    margin-bottom: var(--space-4);
  }
  
  .empty-state-dark {
    background-color: var(--dark-bg-secondary);
    border-color: var(--border-color-dark);
  }
  
  .empty-state-message {
    color: var(--text-secondary);
    font-size: var(--font-size-lg);
    margin: 0;
  }
  
  /* Search results container */
  .search-results-container {
    padding-bottom: var(--space-8);
    background-color: var(--bg-primary);
  }
  
  .search-results-container-dark {
    background-color: var(--dark-bg-primary);
  }
  
  /* Responsive adjustments */
  @media (max-width: 768px) {
    .page-title {
      font-size: var(--font-size-xl);
    }
    
    .page-header {
      padding: var(--space-4) var(--space-3);
    }
    
    .section-container {
      padding: var(--space-3);
    }
    
    .grid-container {
      grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
      gap: var(--space-2);
    }
  }
  
  @media (max-width: 576px) {
    .grid-container {
      grid-template-columns: repeat(2, 1fr);
      gap: var(--space-2);
    }
    
    .page-subtitle {
      font-size: var(--font-size-sm);
    }
  }
</style>
