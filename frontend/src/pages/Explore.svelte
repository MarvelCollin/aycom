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
  let activeTab: 'top' | 'latest' | 'people' | 'media' | 'communities' | 'trending' = 'top';
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
  type ThreadCategory = {
    id: string;
    label: string;
    value: string;
  };
  
  let threadCategories: ThreadCategory[] = [
    { id: 'tech', label: 'Technology', value: 'technology' },
    { id: 'entertainment', label: 'Entertainment', value: 'entertainment' },
    { id: 'sports', label: 'Sports', value: 'sports' },
    { id: 'politics', label: 'Politics', value: 'politics' },
    { id: 'science', label: 'Science', value: 'science' },
    { id: 'health', label: 'Health', value: 'health' },
    { id: 'business', label: 'Business', value: 'business' },
    { id: 'lifestyle', label: 'Lifestyle', value: 'lifestyle' }
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
  let defaultActiveTab: 'trending' | 'people' | 'communities' = 'trending';
  
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
        .filter(item => item.similarity > 0.15) // Only include relevant matches (lower threshold for better fuzzy matching)
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
    // If the search query is empty, reset to showing trending hashtags
    if (!searchQuery || searchQuery.trim() === '') {
      logger.debug('Empty search query, showing trending tab');
      hasSearched = false;
      activeTab = 'trending';
      return;
    }

    // Add to recent searches if not already there
    if (!recentSearches.includes(searchQuery)) {
      recentSearches = [searchQuery, ...recentSearches.slice(0, 2)];
      localStorage.setItem('recentSearches', JSON.stringify(recentSearches));
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
          sort_by: 'popular',
          clientFuzzy: true
        }),
        
        // Latest threads
        safeApiCall(searchThreads, searchQuery, 1, 20, {
          filter: filterOption,
          category: categoryOption,
          sort_by: 'recent',
          clientFuzzy: true
        }),
        
        // Media tab data
        safeApiCall(searchThreadsWithMedia, searchQuery, 1, 12, {
          filter: filterOption,
          category: categoryOption,
          clientFuzzy: true
        }),
        
        // Communities tab data
        safeApiCall(searchCommunities, searchQuery, 1, communitiesPerPage, {
          clientFuzzy: true
        })
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
    
    // Change to search mode
    isSearching = true;
    hasSearched = true;
    activeTab = 'latest';  // Default to showing latest threads for this hashtag
    searchQuery = hashtag;
    
    // Get threads for this hashtag sorted by likes
    const hashtagThreadsData = await getThreadsByHashtag(hashtag, 1, 20);
    
    if (hashtagThreadsData?.threads) {
      // Update latest tab with these threads
      searchResults.latest.threads = (hashtagThreadsData?.threads || []).map(thread => ({
        id: thread.id || '',
        content: thread.content || '',
        username: thread.username || thread.author?.username || '',
        displayName: thread.display_name || thread.author?.display_name || thread.username || '',
        timestamp: thread.created_at || new Date().toISOString(),
        likes: thread.likes_count || thread.like_count || 0,
        replies: thread.replies_count || thread.reply_count || 0,
        reposts: thread.reposts_count || thread.repost_count || 0,
        media: thread.media || [],
        avatar: thread.profile_picture_url || thread.author?.profile_picture_url || null
      }));
      
      // Also update top tab
      searchResults.top.threads = (hashtagThreadsData?.threads || []).map(thread => ({
        id: thread.id || '',
        content: thread.content || '',
        username: thread.username || thread.author?.username || '',
        name: thread.display_name || thread.author?.display_name || thread.username || '',
        created_at: thread.created_at || new Date().toISOString(),
        likes_count: thread.likes_count || thread.like_count || 0,
        replies_count: thread.replies_count || thread.reply_count || 0,
        reposts_count: thread.reposts_count || thread.repost_count || 0,
        media: thread.media || [],
        profile_picture_url: thread.profile_picture_url || thread.author?.profile_picture_url || null
      }));
    } else {
      console.error('Error fetching hashtag threads');
      toastStore.showToast('Failed to load hashtag threads', 'error');
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
    if (defaultActiveTab === 'trending' && trends.length === 0 && !isTrendsLoading) {
      // If there are no trends, fetch them
      fetchTrends();
    } else if (defaultActiveTab === 'communities' && communitiesToDisplay.length === 0 && !isLoadingCommunities) {
      fetchAllCommunities();
    } else if (defaultActiveTab === 'people' && usersToDisplay.length === 0 && !isLoadingUsers) {
      fetchAllUsers();
    }
  }
  
  onMount(async () => {
    logger.debug('Explore component mounted');

    // Load recent searches from localStorage
    loadRecentSearches();
    
    // Load trending hashtags
    await fetchTrends();
    
    // If authenticated, fetch all users
    if (checkAuth()) {
      await fetchAllUsers();
    }
    
    // Show trending tab by default when no search is active
    if (!hasSearched) {
      activeTab = 'trending';
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
    <!-- Modern header with search -->
    <div class="page-header {isDarkMode ? 'page-header-dark' : ''}">
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
    
    <!-- Modern Filters with pill design -->
    <div class="filter-container {isDarkMode ? 'filter-container-dark' : ''}">
      <div class="filter-pills">
        <button 
          class="filter-pill {searchFilter === 'all' ? 'active' : ''}"
          on:click={() => { searchFilter = 'all'; handleFilterChange({detail: 'all'}); }}
        >
          For you
        </button>
        <button 
          class="filter-pill {searchFilter === 'following' ? 'active' : ''}"
          on:click={() => { searchFilter = 'following'; handleFilterChange({detail: 'following'}); }}
        >
          Following
        </button>
        <button 
          class="filter-pill {searchFilter === 'verified' ? 'active' : ''}"
          on:click={() => { searchFilter = 'verified'; handleFilterChange({detail: 'verified'}); }}
        >
          <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="pill-icon">
            <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"></path>
            <polyline points="22 4 12 14.01 9 11.01"></polyline>
          </svg>
          Verified
        </button>
      </div>
      
      <div class="category-select">
        <select 
          class="category-dropdown {isDarkMode ? 'category-dropdown-dark' : ''}"
          bind:value={selectedCategory}
          on:change={(e) => {
            const value = (e.target as HTMLSelectElement).value;
            handleCategoryChange({detail: value});
          }}
        >
          <option value="all">All Categories</option>
          {#each threadCategories as category}
            <option value={category.value}>{category.label}</option>
          {/each}
        </select>
      </div>
    </div>
    
    <!-- Tab navigation for search results -->
    {#if hasSearched}
      <div class="tabs-container {isDarkMode ? 'tabs-container-dark' : ''}">
        <div class="modern-tabs">
          <button 
            class="modern-tab {activeTab === 'top' ? 'active' : ''}"
            on:click={() => handleTabChange({detail: 'top'})}
          >
            Top
          </button>
          <button 
            class="modern-tab {activeTab === 'latest' ? 'active' : ''}"
            on:click={() => handleTabChange({detail: 'latest'})}
          >
            Latest
          </button>
          <button 
            class="modern-tab {activeTab === 'people' ? 'active' : ''}"
            on:click={() => handleTabChange({detail: 'people'})}
          >
            People
          </button>
          <button 
            class="modern-tab {activeTab === 'media' ? 'active' : ''}"
            on:click={() => handleTabChange({detail: 'media'})}
          >
            Media
          </button>
          <button 
            class="modern-tab {activeTab === 'communities' ? 'active' : ''}"
            on:click={() => handleTabChange({detail: 'communities'})}
          >
            Communities
          </button>
        </div>
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
      <!-- When not searching, show tabs to select between Trending, People and Communities -->
      <div class="tabs-container {isDarkMode ? 'tabs-container-dark' : ''}">
        <div class="modern-tabs">
          <button 
            class="modern-tab {defaultActiveTab === 'trending' ? 'active' : ''}"
            on:click={() => handleDefaultTabChange('trending')}
          >
            <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="tab-icon"><polyline points="23 6 13.5 15.5 8.5 10.5 1 18"></polyline><polyline points="17 6 23 6 23 12"></polyline></svg>
            <span>Trending</span>
          </button>
          <button 
            class="modern-tab {defaultActiveTab === 'people' ? 'active' : ''}"
            on:click={() => handleDefaultTabChange('people')}
          >
            <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="tab-icon"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"></path><circle cx="12" cy="7" r="4"></circle></svg>
            <span>People</span>
          </button>
          <button 
            class="modern-tab {defaultActiveTab === 'communities' ? 'active' : ''}"
            on:click={() => handleDefaultTabChange('communities')}
          >
            <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="tab-icon"><path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path><circle cx="9" cy="7" r="4"></circle><path d="M23 21v-2a4 4 0 0 0-3-3.87"></path><path d="M16 3.13a4 4 0 0 1 0 7.75"></path></svg>
            <span>Communities</span>
          </button>
        </div>
      </div>
      
      <!-- Content based on selected default tab -->
      <div class="content-container {isDarkMode ? 'content-container-dark' : ''}">
        {#if defaultActiveTab === 'trending'}
          <div class="section-header">
            <h2 class="section-title {isDarkMode ? 'section-title-dark' : ''}">
              <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="section-icon"><polyline points="23 6 13.5 15.5 8.5 10.5 1 18"></polyline><polyline points="17 6 23 6 23 12"></polyline></svg>
              Trending Hashtags
            </h2>
            <p class="section-description">Popular conversations happening right now</p>
          </div>
          
          <ExploreTrending 
            {trends}
            {isTrendsLoading}
            on:hashtagClick={handleHashtagClick}
          />
          
        {:else if defaultActiveTab === 'people'}
          <div class="section-header">
            <h2 class="section-title {isDarkMode ? 'section-title-dark' : ''}">
              <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="section-icon"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"></path><circle cx="12" cy="7" r="4"></circle></svg>
              People to Follow
            </h2>
            <p class="section-description">Connect with interesting profiles</p>
          </div>
          
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
          <div class="section-header">
            <h2 class="section-title {isDarkMode ? 'section-title-dark' : ''}">
              <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="section-icon"><path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path><circle cx="9" cy="7" r="4"></circle><path d="M23 21v-2a4 4 0 0 0-3-3.87"></path><path d="M16 3.13a4 4 0 0 1 0 7.75"></path></svg>
              Communities
            </h2>
            <p class="section-description">Join groups with shared interests</p>
          </div>
          
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
  /* Modern Explore page styles */
  .explore-page-content {
    width: 100%;
    background-color: var(--bg-primary);
    color: var(--text-primary);
    min-height: 100vh;
  }
  
  .explore-page-content-dark {
    background-color: var(--dark-bg-primary);
    color: var(--dark-text-primary);
  }
  
  /* Modern header styles */
  .page-header {
    padding: 8px 16px;
    position: sticky;
    top: 0;
    z-index: 10;
    background-color: var(--bg-primary);
    backdrop-filter: blur(12px);
    -webkit-backdrop-filter: blur(12px);
    border-bottom: 1px solid transparent;
  }
  
  .page-header-dark {
    background-color: rgba(var(--dark-bg-primary-rgb), 0.8);
    border-bottom-color: var(--dark-border-color);
  }
  
  .search-container {
    width: 100%;
    max-width: 600px;
    margin: 0 auto;
  }

  /* Modern filter container with pills */
  .filter-container {
    padding: 12px 16px;
    display: flex;
    justify-content: space-between;
    align-items: center;
    border-bottom: 1px solid var(--border-color);
  }
  
  .filter-container-dark {
    border-bottom-color: var(--dark-border-color);
  }
  
  .filter-pills {
    display: flex;
    gap: 8px;
    overflow-x: auto;
    scrollbar-width: none;
    -ms-overflow-style: none;
    padding-bottom: 4px;
  }
  
  .filter-pills::-webkit-scrollbar {
    display: none;
  }
  
  .filter-pill {
    background: none;
    border: 1px solid var(--border-color);
    border-radius: 9999px;
    padding: 6px 16px;
    font-size: 14px;
    font-weight: 500;
    color: var(--text-primary);
    cursor: pointer;
    transition: background-color 0.2s;
    display: flex;
    align-items: center;
    gap: 6px;
    white-space: nowrap;
  }
  
  .filter-pill:hover {
    background-color: var(--hover-bg);
  }
  
  .filter-pill.active {
    background-color: var(--text-primary);
    border-color: var(--text-primary);
    color: var(--bg-primary);
  }
  
  .filter-container-dark .filter-pill {
    color: var(--dark-text-primary);
    border-color: var(--dark-border-color);
  }
  
  .filter-container-dark .filter-pill:hover {
    background-color: var(--dark-hover-bg);
  }
  
  .filter-container-dark .filter-pill.active {
    background-color: var(--dark-text-primary);
    border-color: var(--dark-text-primary);
    color: var(--dark-bg-primary);
  }
  
  .pill-icon {
    width: 14px;
    height: 14px;
  }
  
  .category-select {
    position: relative;
  }
  
  .category-dropdown {
    appearance: none;
    background-color: transparent;
    border: 1px solid var(--border-color);
    border-radius: 9999px;
    padding: 6px 30px 6px 16px;
    font-size: 14px;
    color: var(--text-primary);
    cursor: pointer;
    background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='16' height='16' viewBox='0 0 24 24' fill='none' stroke='%23536471' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpolyline points='6 9 12 15 18 9'%3E%3C/polyline%3E%3C/svg%3E");
    background-repeat: no-repeat;
    background-position: right 10px center;
    background-size: 16px;
  }
  
  .category-dropdown-dark {
    color: var(--dark-text-primary);
    border-color: var(--dark-border-color);
    background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='16' height='16' viewBox='0 0 24 24' fill='none' stroke='%238b98a5' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpolyline points='6 9 12 15 18 9'%3E%3C/polyline%3E%3C/svg%3E");
  }
  
  /* Modern tabs container */
  .tabs-container {
    position: sticky;
    top: 60px;
    z-index: 5;
    background-color: var(--bg-primary);
    border-bottom: 1px solid var(--border-color);
  }
  
  .tabs-container-dark {
    background-color: var(--dark-bg-primary);
    border-bottom-color: var(--dark-border-color);
  }
  
  .modern-tabs {
    display: flex;
    overflow-x: auto;
    scrollbar-width: none;
    -ms-overflow-style: none;
  }
  
  .modern-tabs::-webkit-scrollbar {
    display: none;
  }
  
  .modern-tab {
    flex: 1;
    min-width: fit-content;
    padding: 16px;
    background: none;
    border: none;
    color: var(--text-secondary);
    font-size: 15px;
    font-weight: 500;
    position: relative;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    transition: color 0.2s;
  }
  
  .modern-tab:hover {
    color: var(--text-primary);
    background-color: var(--hover-bg);
  }
  
  .modern-tab.active {
    color: var(--text-primary);
    font-weight: 700;
  }
  
  .modern-tab.active::after {
    content: "";
    position: absolute;
    bottom: 0;
    left: 50%;
    transform: translateX(-50%);
    width: 56px;
    height: 4px;
    border-radius: 4px 4px 0 0;
    background-color: var(--color-primary);
  }
  
  .tabs-container-dark .modern-tab {
    color: var(--dark-text-secondary);
  }
  
  .tabs-container-dark .modern-tab:hover {
    color: var(--dark-text-primary);
    background-color: var(--dark-hover-bg);
  }
  
  .tabs-container-dark .modern-tab.active {
    color: var(--dark-text-primary);
  }
  
  .content-container {
    padding: 16px;
  }
  
  .content-container-dark {
    border-color: var(--dark-border-color);
  }
  
  .search-results-container {
    padding: 16px;
  }
  
  .search-results-container-dark {
    border-color: var(--dark-border-color);
  }
  
  /* Responsive adjustments */
  @media (max-width: 768px) {
    .filter-container {
      flex-direction: column;
      align-items: flex-start;
      gap: 12px;
    }
    
    .filter-pills {
      width: 100%;
    }
    
    .category-select {
      width: 100%;
    }
    
    .category-dropdown {
      width: 100%;
    }
    
    .modern-tab {
      padding: 12px 8px;
      font-size: 14px;
    }
    
    .modern-tab.active::after {
      width: 40px;
    }
    
    .page-header {
      position: sticky;
      top: 0;
      z-index: 10;
    }
  }
  
  @media (max-width: 576px) {
    .modern-tab span {
      display: none;
    }
    
    .modern-tab {
      padding: 12px;
    }
    
    .content-container,
    .search-results-container {
      padding: 12px 8px;
    }
  }
</style>
