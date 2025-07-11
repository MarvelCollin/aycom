<script lang="ts">
  import { onMount } from "svelte";
  import MainLayout from "../components/layout/MainLayout.svelte";
  import { useAuth } from "../hooks/useAuth";
  import { useTheme } from "../hooks/useTheme";
  import type { ITweet } from "../interfaces/ISocialMedia";
  import type { ITrend } from "../interfaces/ITrend";
  import type { IAuthStore } from "../interfaces/IAuth";
  import { createLoggerWithPrefix } from "../utils/logger";
  import { toastStore } from "../stores/toastStore";
  import { getTrends } from "../api/trends";
  import { searchUsers, getAllUsers, getFollowing } from "../api/user";
  import { searchThreads, searchThreadsWithMedia, getThreadsByHashtag } from "../api/thread";
  import { searchCommunities, getCommunities } from "../api/community";
  import { debounce } from "../utils/helpers";
  import { formatTimeAgo } from "../utils/common";
  import { improvedSearchUsers } from "../api/userApi";

  import ExploreSearch from "../components/explore/ExploreSearch.svelte";
  import ExploreFilters from "../components/explore/ExploreFilters.svelte";
  import ExploreTrending from "../components/explore/ExploreTrending.svelte";
  import ExploreTabs from "../components/explore/ExploreTabs.svelte";
  import ExploreTopResults from "../components/explore/ExploreTopResults.svelte";
  import ExploreLatestResults from "../components/explore/ExploreLatestResults.svelte";
  import ExplorePeopleResults from "../components/explore/ExplorePeopleResults.svelte";
  import ExploreMediaResults from "../components/explore/ExploreMediaResults.svelte";
  import ExploreCommunityResults from "../components/explore/ExploreCommunityResults.svelte";
  import LoadingSkeleton from "../components/common/LoadingSkeleton.svelte";
  import ProfileCard from "../components/explore/ProfileCard.svelte";
  import CommunityCard from "../components/explore/CommunityCard.svelte";
  import Toast from "../components/common/Toast.svelte";

  const logger = createLoggerWithPrefix("Explore");

  const { getAuthState } = useAuth();
  const { theme } = useTheme();

  $: authState = getAuthState ? getAuthState() : { user_id: null, is_authenticated: false, access_token: null, refresh_token: null };
  $: isDarkMode = $theme === "dark";
  $: sidebarUsername = authState?.user_id ? `User_${authState.user_id.substring(0, 4)}` : "";
  $: sidebarDisplayName = authState?.user_id ? `User ${authState.user_id.substring(0, 4)}` : "";
  $: sidebarAvatar = "https://secure.gravatar.com/avatar/0?d=mp"; 

  let trends: ITrend[] = [];
  let isTrendsLoading = true;
  const suggestedFollows = [];

  const allUsers: Array<{
    id: string;
    username: string;
    displayName: string;
    avatar: string | null;
    bio?: string;
    isVerified: boolean;
    followerCount: number;
    isFollowing: boolean;
  }> = [];
  const isLoadingAllUsers = false;

  let searchQuery = "";
  let isSearching = false;
  let recentSearches: string[] = [];
  let searchFilter: "all" | "following" | "verified" = "all";
  let activeTab: "trending" | "media" | "people" | "communities" | "latest" = "trending";
  let showRecentSearches = false;
  let isLoadingRecommendations = false;

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
        per_page: number;
        has_more?: boolean;
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
        total_pages: 0,
        per_page: 25
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

  let hasSearched = false;

  let usersToDisplay: Array<{
    id: string;
    username: string;
    displayName: string;
    avatar: string | null;
    bio?: string;
    isVerified: boolean;
    followerCount: number;
    isFollowing: boolean;
    fuzzyMatchScore?: number;
  }> = [];
  let isLoadingUsers = false;

  type ThreadCategory = {
    id: string;
    label: string;
    value: string;
  };

  const threadCategories: ThreadCategory[] = [
    { id: "tech", label: "Technology", value: "technology" },
    { id: "entertainment", label: "Entertainment", value: "entertainment" },
    { id: "sports", label: "Sports", value: "sports" },
    { id: "politics", label: "Politics", value: "politics" },
    { id: "science", label: "Science", value: "science" },
    { id: "health", label: "Health", value: "health" },
    { id: "business", label: "Business", value: "business" },
    { id: "lifestyle", label: "Lifestyle", value: "lifestyle" }
  ];

  let selectedCategory = "all";

  let peoplePerPage = 25;
  let peopleCurrentPage = 1;
  let communitiesPerPage = 25;
  let communitiesCurrentPage = 1;
  let mediaPage = 1; 

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
  let defaultActiveTab: "trending" | "media" | "people" | "communities" | "latest" = "trending";

  function checkAuth() {
    if (!authState || !authState.is_authenticated) {
      logger.error("User not authenticated, redirecting to login");
      toastStore.showToast("You need to log in to access explore", "warning");
      setTimeout(() => {
        window.location.href = "/login";
      }, 1000);
      return false;
    }
    logger.info("Authentication check passed, user is authenticated");
    return true;
  }

  async function fetchAllUsers() {
    isLoadingUsers = true;
    try {
      const response = await getAllUsers(1, 20, "created_at", false);

      console.log("fetchAllUsers response:", response);
      const users = response.users || [];

      if (users && users.length > 0) {

        usersToDisplay = users.map(user => ({
          id: user.id,
          username: user.username,
          displayName: user.name || user.display_name || user.username,
          avatar: user.profile_picture_url || user.avatar || null,
          bio: user.bio || "",
          isVerified: user.is_verified || false,
          followerCount: user.follower_count || 0,
          isFollowing: user.is_following || false
        }));

        console.log("Mapped users for display:", usersToDisplay);
        logger.debug("Fetched users:", { count: usersToDisplay.length });
      } else {
        usersToDisplay = [];
        logger.info("No users found");
      }
    } catch (error) {
      logger.error("Error fetching all users:", error);
      toastStore.showToast("Failed to load users", "error");
      usersToDisplay = [];
    } finally {
      isLoadingUsers = false;
    }
  }

  function loadRecentSearches() {
    try {
      const savedSearches = localStorage.getItem("recentSearches");
      if (savedSearches) {
        recentSearches = JSON.parse(savedSearches).slice(0, 3);
      }
    } catch (error) {
      console.error("Error loading recent searches:", error);
      recentSearches = [];
    }
  }

  function saveToRecentSearches(query: string) {
    if (!query.trim()) return;

    try {

      const filteredSearches = recentSearches.filter(s => s !== query);

      const updatedSearches = [query, ...filteredSearches].slice(0, 3);
      recentSearches = updatedSearches;

      localStorage.setItem("recentSearches", JSON.stringify(updatedSearches));
    } catch (error) {
      console.error("Error saving recent search:", error);
    }
  }

  async function fetchTrends() {
    isTrendsLoading = true;
    try {
      const trendData = await getTrends(10); 
      trends = trendData;
      logger.debug("Trends loaded", { trendsCount: trends.length });
    } catch (error) {
      console.error("Error loading trends:", error);
      toastStore.showToast("Failed to load trends. Please try again.", "error");
      trends = [];
    } finally {
      isTrendsLoading = false;
    }
  }

  const debouncedSearchProfiles = debounce(async (query: string) => {
    console.log("Running debouncedSearchProfiles with query:", query);

    if (!query || query.length < 2) {
      searchResults.top.profiles = [];
      searchResults.top.isLoading = false;
      console.log("Query too short, cleared profiles");
      return;
    }

    searchResults.top.isLoading = true;
    isLoadingRecommendations = true;

    try {
      console.log("Fetching user profiles for:", query);

      const { users } = await searchUsers(query.trim(), 1, 15, {
        sort: "follower_count" 
      });

      console.log("Recommended profiles API response:", users);

      if (!users || users.length === 0) {
        console.log("No profiles found for query:", query);
        searchResults.top.profiles = [];
      } else {

        searchResults.top.profiles = users.slice(0, 6).map(user => ({
        id: user.id,
        username: user.username,
        displayName: user.display_name || user.username,
        avatar: user.avatar,
          bio: user.bio || "",
        isVerified: user.is_verified || false,
        followerCount: user.follower_count || 0,
        isFollowing: user.is_following || false
      }));

        console.log("Displaying recommended profiles:", searchResults.top.profiles);
      }

      isSearching = true;
      hasSearched = true;

      searchResults = { ...searchResults };

    } catch (error) {
      logger.error("Error searching profiles:", error);
      console.error("Error in profile search:", error);
      searchResults.top.profiles = [];
    } finally {
      searchResults.top.isLoading = false;
      isLoadingRecommendations = false;
    }
  }, 300);

  function handleSearchInput(event) {
    const query = event.detail;
    searchQuery = query;
    console.log("Search input changed:", query);

    if (query && query.length > 1) {
      isSearching = true;
      hasSearched = true;

      activeTab = "trending";

      searchResults.top.isLoading = true;

      debouncedSearchProfiles(query);
    } else {

      if (!query || query.length === 0) {
        isSearching = false;
        hasSearched = false;
        activeTab = "trending";
      }
    }

    logger.debug("Search input updated", { searchQuery });
  }

  function handleSearch(event) {

    if (!hasSearched && defaultActiveTab === "communities") {
      activeTab = "communities";
    } else {

      activeTab = "trending";
    }
    executeSearch();
  }

  function handleClearSearch() {
    searchQuery = "";
    hasSearched = false;
    isSearching = false;
    showRecentSearches = false; 
    activeTab = "trending"; 
  }

  function handleToggleFollow(userId) {

    const userIndex = usersToDisplay.findIndex(user => user.id === userId);
    if (userIndex !== -1) {

      const user = usersToDisplay[userIndex];
      const updatedUser = {...user, isFollowing: !user.isFollowing};

      usersToDisplay = [
        ...usersToDisplay.slice(0, userIndex),
        updatedUser,
        ...usersToDisplay.slice(userIndex + 1)
      ];

      handleFollowUser({ detail: userId });
    }
  }

  async function executeSearch() {
    if (!searchQuery || searchQuery.length < 2) {
      hasSearched = false;
      isSearching = false;
      return;
    }

    saveToRecentSearches(searchQuery);

    isSearching = true;
    hasSearched = true;
    showRecentSearches = false; 
    activeTab = "trending"; 

    searchResults.top.isLoading = true;
    searchResults.latest.isLoading = true;
    searchResults.people.isLoading = true;
    searchResults.media.isLoading = true;
    searchResults.communities.isLoading = true;

    try {
      const filterOption = searchFilter === "following" ? "following" : (searchFilter === "verified" ? "verified" : "all");
      const categoryOption = selectedCategory !== "all" ? selectedCategory : undefined;

      logger.debug("Starting search with Damerau-Levenshtein fuzzy matching", {
        query: searchQuery || "(empty - showing all)",
        filter: filterOption,
        category: categoryOption,
        sortBy: "popular",
        fuzzyThreshold: "0.3 (modified for better matching)"
      });

      const safeApiCall = async (apiFunction, ...args) => {
        try {
          const result = await apiFunction(...args);

          if (result === null || result === undefined) {
            logger.warn(`API call ${apiFunction.name} returned null or undefined`);
            return getDefaultResultForFunction(apiFunction.name);
          }

          return result;
        } catch (error) {
          logger.error(`Error in API call ${apiFunction.name}:`, error);

          return getDefaultResultForFunction(apiFunction.name);
        }
      };

      function getDefaultResultForFunction(functionName) {
        switch(functionName) {
          case "searchUsers":
            return { users: [], totalCount: 0 };
          case "searchThreads":
          case "searchThreadsWithMedia":
            return { threads: [] };
          case "searchCommunities":
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

      const [peopleData, topThreadsData, latestThreadsData, mediaData, communitiesData] = await Promise.all([

        safeApiCall(improvedSearchUsers, searchQuery, filterOption, 1, peoplePerPage),

        safeApiCall(searchThreads, searchQuery, 1, 10, {
          filter: filterOption,
          category: categoryOption,
          sort_by: "popular"
        }),

        safeApiCall(searchThreads, searchQuery, 1, 20, {
          filter: filterOption,
          category: categoryOption,
          sort_by: "recent"
        }),

        safeApiCall(searchThreadsWithMedia, searchQuery, 1, 12, {
          filter: filterOption,
          category: categoryOption
        }),

        safeApiCall(searchCommunities, searchQuery, 1, communitiesPerPage)
      ]);

      console.log("Raw people data:", peopleData);

      let processedPeopleData = peopleData;
      if (!processedPeopleData || typeof processedPeopleData !== "object") {
        console.error("Invalid people data received:", processedPeopleData);
        processedPeopleData = { data: { users: [], pagination: { total_count: 0 } } };
      }

      const peopleDataUsers = processedPeopleData.data?.users || processedPeopleData.users || [];
      const totalPeopleCount = processedPeopleData.data?.pagination?.total_count || processedPeopleData.totalCount || processedPeopleData.total || 0;

      const peopleUsers = peopleDataUsers.map(user => {

        console.log("Processing user:", user);

        if (!user || typeof user !== "object") {
          console.error("Invalid user object:", user);
          return null;
        }

        return {
          id: user.id || "",
          username: user.username || "",
          displayName: user.name || user.username || "",
          avatar: user.profile_picture_url || null,
          bio: user.bio || "",
          isVerified: user.is_verified || false,
          followerCount: user.follower_count || 0,
          isFollowing: false 
        };
      }).filter(user => user !== null); 

      console.log("Mapped people users:", peopleUsers);

      function calculateSimpleStringSimilarity(a: string, b: string): number {

        if (a.includes(b) || b.includes(a)) {
          return 0.8; 
        }

        let commonPrefix = 0;
        const minLength = Math.min(a.length, b.length);
        for (let i = 0; i < minLength; i++) {
          if (a[i] === b[i]) {
            commonPrefix++;
          } else {
            break;
          }
        }

        return commonPrefix > 0 ? commonPrefix / Math.max(a.length, b.length) : 0.3;
      }

      const topThreads = (topThreadsData.threads || []).map(thread => ({
        id: thread.id,
        content: thread.content,
        username: thread.author?.username || "anonymous",
        name: thread.author?.display_name || "User",
        created_at: thread.created_at || new Date().toISOString(),
        likes_count: thread.like_count || 0,
        replies_count: thread.reply_count || 0,
        reposts_count: thread.repost_count || 0,
        media: thread.media,
        profile_picture_url: thread.author?.avatar
      }));

      const topProfiles = [...peopleUsers]
        .sort((a, b) => ((b.follower_count || 0) - (a.follower_count || 0)))
        .slice(0, 3);

      console.log("Raw communities data:", communitiesData);

      let processedCommunitiesData = communitiesData;
      if (!processedCommunitiesData || typeof processedCommunitiesData !== "object") {
        console.error("Invalid communities data received:", processedCommunitiesData);
        processedCommunitiesData = { communities: [], total_count: 0 };
      }

      const normalizedCommunities = (processedCommunitiesData.communities || []).map(community => {
        if (!community || typeof community !== "object") {
          console.error("Invalid community object:", community);
          return null;
        }

        return {
          id: community.id || "",
          name: community.name || "",
          description: community.description || "",
          logo: community.logo || community.logo_url || community.avatar || null,
          member_count: community.member_count || community.memberCount || 0,
          is_joined: community.is_joined || community.isJoined || false,
          is_pending: community.is_pending || community.isPending || false
        };
      }).filter(community => community !== null);

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
            username: thread.author?.username || "anonymous",
            displayName: thread.author?.display_name || "User",
            timestamp: thread.created_at || new Date().toISOString(),
            likes: thread.likes_count || thread.like_count || 0,
            replies: thread.replies_count || thread.reply_count || 0,
            reposts: thread.reposts_count || thread.repost_count || 0,
            media: thread.media || [],
            avatar: thread.profile_picture_url || thread.author?.profile_picture_url || null
          })),
          isLoading: false
        },
        people: {
          users: peopleUsers,
          totalCount: totalPeopleCount,
          pagination: processedPeopleData?.data?.pagination || processedPeopleData?.pagination || {
            current_page: 1,
            total_pages: Math.ceil(totalPeopleCount / peoplePerPage),
            per_page: peoplePerPage,
            total_count: totalPeopleCount
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

      usersToDisplay = peopleUsers;

      logger.debug("Search completed", {
        query: searchQuery,
        filter: filterOption,
        peopleCount: peopleDataUsers?.length || 0,
        threadsCount: topThreadsData?.threads?.length || 0,
        mediaCount: mediaData?.threads?.length || 0,
        totalPeopleCount: totalPeopleCount
      });

    } catch (error) {
      console.error("Error executing search:", error);
      toastStore.showToast("Search failed. Please try again.", "error");
    } finally {
      isSearching = false;
    }
  }

  function handleSearchFocus() {
    showRecentSearches = true;
    logger.debug("Search focused");
  }

  function handleRecentSearchSelect(event) {
    searchQuery = event.detail;
    logger.debug("Recent search selected", { searchQuery });
    executeSearch();
  }

  function clearRecentSearches() {
    recentSearches = [];
    localStorage.removeItem("recentSearches");
    logger.debug("Recent searches cleared");
  }

  function handleFilterChange(event) {
    const newFilter = event.detail;
    logger.debug("Filter changed", { from: searchFilter, to: newFilter });
    console.log("Filter changed from", searchFilter, "to", newFilter);

    searchFilter = newFilter;

    if (hasSearched && searchQuery.trim() !== "") {

      console.log("In search mode, re-executing search with new filter");
      executeSearch();
    } else {

      console.log("Not in search mode, fetching data for filter:", newFilter, "and tab:", defaultActiveTab);
      if (defaultActiveTab === "people") {
        isLoadingUsers = true;
        if (searchFilter === "all") {

          console.log("Fetching all users");
          fetchAllUsers();
        } else if (searchFilter === "following") {

          console.log("Fetching followed users");
          fetchFollowedUsers();
        } else if (searchFilter === "verified") {

          console.log("Fetching verified users");
          fetchVerifiedUsers();
        }
      } else if (defaultActiveTab === "communities") {
        isLoadingCommunities = true;

        fetchAllCommunities();
      }
    }
  }

  async function fetchFollowedUsers() {
    if (!authState.is_authenticated) {
      toastStore.showToast("You need to be logged in to view followed users", "warning");
      return;
    }

    isLoadingUsers = true;
    try {

      const userId = authState.user_id;
      console.log("Fetching following users for:", userId);

      const response = await getFollowing(userId || "", 1, 20);
      console.log("Following API response:", response);

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
        bio: user.bio || "",
        isVerified: user.is_verified || false,
        followerCount: user.follower_count || 0,
        isFollowing: true
      }));
    } catch (error) {
      console.error("Error fetching followed users:", error);
      toastStore.showToast("Failed to load followed users", "error");
      usersToDisplay = [];
    } finally {
      isLoadingUsers = false;
    }
  }

  async function fetchVerifiedUsers() {
    isLoadingUsers = true;
    try {
      console.log("Fetching verified users...");

      const response = await getAllUsers(1, 50, "created_at", false);
      console.log("All users response for verified filtering:", response);

      const allUsers = response.users || [];
      const verifiedUsers = allUsers.filter(user => user.is_verified === true);

      console.log("Filtered verified users:", verifiedUsers);

      if (verifiedUsers.length > 0) {

        usersToDisplay = verifiedUsers.map(user => ({
          id: user.id,
          username: user.username,
          displayName: user.name || user.display_name || user.username,
          avatar: user.profile_picture_url || user.avatar || null,
          bio: user.bio || "",
          isVerified: true, 
          followerCount: user.follower_count || 0,
          isFollowing: user.is_following || false
        }));

        console.log("Mapped verified users for display:", usersToDisplay);
        logger.debug("Fetched verified users:", { count: usersToDisplay.length });
      } else {
        console.log("No verified users found");
        usersToDisplay = [];

        try {
          const searchResponse = await searchUsers("", 1, 20, { filter: "verified" });
          console.log("Search API verified users response:", searchResponse);

          if (searchResponse && searchResponse.users && Array.isArray(searchResponse.users) && searchResponse.users.length > 0) {
            usersToDisplay = searchResponse.users.map(user => ({
              id: user.id || "",
              username: user.username || "",
              displayName: user.name || user.display_name || user.username || "",
              avatar: user.profile_picture_url || user.avatar || null,
              bio: user.bio || "",
              isVerified: true,
              followerCount: user.follower_count || 0,
              isFollowing: user.is_following || false
            }));
            console.log("Found verified users from search API:", usersToDisplay);
          }
        } catch (searchError) {
          console.error("Backup search method for verified users failed:", searchError);
        }
      }
    } catch (error) {
      logger.error("Error fetching verified users:", error);
      toastStore.showToast("Failed to load verified users", "error");
      usersToDisplay = [];
    } finally {
      isLoadingUsers = false;
    }
  }

  function handleCategoryChange(event) {
    const category = event.detail;
    selectedCategory = category;
    logger.debug("Category changed", { category });
    if (hasSearched) {
      executeSearch();
    }
  }

  function handleTabChange(event) {
    activeTab = event.detail;
    logger.debug("Tab changed", { tab: activeTab });

    if (activeTab === "communities" || activeTab === "people") {
      defaultActiveTab = activeTab;
    }
  }

  async function handleHashtagClick(event) {
    const hashtag = event.detail;
    logger.debug("Hashtag clicked", { hashtag });

    isSearching = true;
    hasSearched = true;
    activeTab = "latest";  
    searchQuery = hashtag;

    const hashtagThreadsData = await getThreadsByHashtag(hashtag, 1, 20);

    if (hashtagThreadsData?.threads) {

      searchResults.latest.threads = (hashtagThreadsData?.threads || []).map(thread => {

        let timestamp;
        try {
          const date = new Date(thread.created_at || thread.timestamp || new Date());
          timestamp = !isNaN(date.getTime()) ? thread.created_at : new Date().toISOString();
        } catch (e) {
          console.error("Invalid date in thread:", thread.created_at);
          timestamp = new Date().toISOString();
        }

        return {
        id: thread.id || "",
        content: thread.content || "",
        username: thread.username || thread.author?.username || "",
        displayName: thread.display_name || thread.author?.display_name || thread.username || "",
          timestamp: timestamp,
        likes: thread.likes_count || thread.like_count || 0,
        replies: thread.replies_count || thread.reply_count || 0,
        reposts: thread.reposts_count || thread.repost_count || 0,
        media: thread.media || [],
        avatar: thread.profile_picture_url || thread.author?.profile_picture_url || null
        };
      });

      searchResults.top.threads = (hashtagThreadsData?.threads || []).map(thread => {

        let created_at;
        try {
          const date = new Date(thread.created_at || thread.timestamp || new Date());
          created_at = !isNaN(date.getTime()) ? thread.created_at : new Date().toISOString();
        } catch (e) {
          console.error("Invalid date in thread for top tab:", thread.created_at);
          created_at = new Date().toISOString();
        }

        return {
        id: thread.id || "",
        content: thread.content || "",
        username: thread.username || thread.author?.username || "",
        name: thread.display_name || thread.author?.display_name || thread.username || "",
          created_at: created_at,
        likes_count: thread.likes_count || thread.like_count || 0,
        replies_count: thread.replies_count || thread.reply_count || 0,
        reposts_count: thread.reposts_count || thread.repost_count || 0,
        media: thread.media || [],
        profile_picture_url: thread.profile_picture_url || thread.author?.profile_picture_url || null
        };
      });
    } else {
      console.error("Error fetching hashtag threads");
      toastStore.showToast("Failed to load hashtag threads", "error");
    }
  }

  async function handlePeoplePageChange(event) {
    const page = event.detail;
    peopleCurrentPage = page;
    searchResults.people.isLoading = true;

    try {
      const filterOption = searchFilter === "following" ? "following" : (searchFilter === "verified" ? "verified" : "all");

      console.log(`Fetching people page ${page} with filter ${filterOption}, query: '${searchQuery}'`);

      const response = await improvedSearchUsers(searchQuery, filterOption, page, peoplePerPage);

      console.log("People pagination response:", response);

      if (response && response.success) {

        const peopleResults = response.data?.users || [];
        const peopleCount = peopleResults.length;
        const totalPeopleCount = response.data?.pagination?.total_count || 0;

        searchResults.people.users = peopleResults.map(user => ({
          id: user.id || "",
          username: user.username || "",
          displayName: user.name || "",
          avatar: user.profile_picture_url || null,
          bio: user.bio || "",
          isVerified: user.is_verified || false,
          followerCount: user.follower_count || 0,
          isFollowing: false 
        }));

        searchResults.people.pagination = {
          current_page: page,
          total_pages: response.data?.pagination?.total_pages || 1,
          total_count: response.data?.pagination?.total_count || 0,
          per_page: peoplePerPage
        };

        searchResults.people.isLoading = false;

        logger.debug(`Found ${peopleCount} people, total: ${totalPeopleCount}`);

        usersToDisplay = searchResults.people.users;

        console.log(`Updated people results: ${peopleCount} users, total: ${totalPeopleCount}`);
      }
    } catch (error) {
      console.error("Error loading people page:", error);
      toastStore.showToast("Failed to load more profiles", "error");
      searchResults.people.isLoading = false;
    }
  }

  function handlePeoplePerPageChange(event) {
    peoplePerPage = event.detail;
    searchResults.people.isLoading = true;
    peopleCurrentPage = 1;
    executeSearch();
  }

  async function handleCommunitiesPageChange(event) {
    const page = event.detail;
    communitiesCurrentPage = page;
    searchResults.communities.isLoading = true;

    try {
      const response = await searchCommunities(searchQuery, page, communitiesPerPage);
      const communities = response.communities || [];

      let totalCount = 0;

      if (response.total_count !== undefined) {

        totalCount = response.total_count;

      } else if (response.total !== undefined) {

        totalCount = response.total;

      } else if (response.pagination && response.pagination.total_count !== undefined) {

        totalCount = response.pagination.total_count;
      }

      searchResults.communities = {
        communities,
        totalCount,
        isLoading: false
      };
    } catch (error) {
      console.error("Error loading community page:", error);
      toastStore.showToast("Failed to load communities", "error");
      searchResults.communities.isLoading = false;
    }
  }

  function handleCommunitiesPerPageChange(event) {
    communitiesPerPage = event.detail;
    searchResults.communities.isLoading = true;
    communitiesCurrentPage = 1;
    executeSearch();
  }

  async function loadMoreMedia() {
    if (searchResults.media.isLoading) return;

    searchResults.media.isLoading = true;
    mediaPage++;

    try {
      const filterOption = searchFilter === "following" ? "following" : "all";
      const categoryOption = selectedCategory !== "all" ? selectedCategory : undefined;

      const data = await searchThreadsWithMedia(searchQuery, mediaPage, 12, {
        filter: filterOption,
        category: categoryOption
      });

      searchResults.media = {
        threads: [...searchResults.media.threads, ...(data.threads || [])],
        totalCount: data.total_count || data.total || searchResults.media.totalCount || 0,
        isLoading: false
      };
    } catch (error) {
      console.error("Error loading more media:", error);
      toastStore.showToast("Failed to load more media", "error");
      searchResults.media.isLoading = false;
    }
  }

  function handleFollowUser(event) {
    const userId = event.detail;
    logger.debug("Follow user requested", { userId });

    toastStore.showToast("Follow feature will be implemented soon", "info");
  }

  function handleProfileClick(event) {
    const userId = event.detail;
    logger.debug("Profile clicked", { userId });

    window.location.href = `/user/${userId}`;
  }

  function handleJoinCommunity(event) {
    const { communityId } = event.detail;
    logger.debug("Join community requested", { communityId });

    toastStore.showToast("Join community feature will be implemented soon", "info");
  }

  async function fetchAllCommunities() {
    isLoadingCommunities = true;
    try {
      const params = {
        page: 1,
        limit: communitiesPerPage,
        is_approved: true
      };

      const response = await getCommunities(params);

      console.log("fetchAllCommunities response:", response);
      const communities = response.communities || [];

      if (communities && communities.length > 0) {

        communitiesToDisplay = communities.map(community => ({
          id: community.id || "",
          name: community.name || "",
          description: community.description || "",
          logo: community.logo_url || null,
          member_count: community.member_count || 0,

          is_joined: false, 
          is_pending: false 
        }));

        console.log("Mapped communities for display:", communitiesToDisplay);
        logger.debug("Fetched communities:", { count: communitiesToDisplay.length });
      } else {
        communitiesToDisplay = [];
        logger.info("No communities found");
      }
    } catch (error) {
      logger.error("Error fetching all communities:", error);
      toastStore.showToast("Failed to load communities", "error");
      communitiesToDisplay = [];
    } finally {
      isLoadingCommunities = false;
    }
  }

  function handleDefaultTabChange(newTab: "trending" | "media" | "people" | "communities" | "latest") {
    defaultActiveTab = newTab;
    logger.debug("Default tab changed", { tab: defaultActiveTab });

    if (defaultActiveTab === "trending" && trends.length === 0 && !isTrendsLoading) {

      fetchTrends();
    } else if (defaultActiveTab === "communities" && communitiesToDisplay.length === 0 && !isLoadingCommunities) {
      fetchAllCommunities();
    } else if (defaultActiveTab === "people" && usersToDisplay.length === 0 && !isLoadingUsers) {
      fetchAllUsers();
    } else if (defaultActiveTab === "media" && searchResults.media.threads.length === 0 && !searchResults.media.isLoading) {

      console.log("Media tab selected, could load media content here");
    } else if (defaultActiveTab === "latest" && searchResults.latest.threads.length === 0 && !searchResults.latest.isLoading) {

      console.log("Latest tab selected, could load latest content here");
    }
  }

  onMount(async () => {
    logger.debug("Explore component mounted");

    loadRecentSearches();

    await fetchTrends();

    if (checkAuth()) {
      await fetchAllUsers();
    }

    if (!hasSearched) {
      activeTab = "trending";
    }
  });
</script>

<!-- Enhanced page layout using full width -->
<MainLayout
  username={sidebarUsername}
  displayName={sidebarDisplayName}
  avatar={sidebarAvatar}
  {trends}
  {suggestedFollows}
  pageTitle="Explore"
>
  <div class="explore-page-content {isDarkMode ? "explore-page-content-dark" : ""}">
    <!-- Modern header with search -->
    <div class="page-header {isDarkMode ? "page-header-dark" : ""}">
      <div class="search-container">
        <ExploreSearch
          bind:searchQuery
          bind:showRecentSearches
          {recentSearches}
          recommendedProfiles={searchResults.top.profiles}
          {isSearching}
          {isLoadingRecommendations}
          on:search={handleSearch}
          on:input={handleSearchInput}
          on:focus={handleSearchFocus}
          on:selectRecentSearch={handleRecentSearchSelect}
          on:clearRecentSearches={clearRecentSearches}
          on:clearSearch={handleClearSearch}
          on:enterPressed={() => {

            isSearching = true;
            hasSearched = true;
          }}
        />
      </div>
    </div>

    <!-- Modern Filters with pill design -->
    <div class="filter-container {isDarkMode ? "filter-container-dark" : ""}">
      <div class="filter-pills">
        <button
          class="filter-pill {searchFilter === "all" ? "active" : ""}"
          on:click={() => { searchFilter = "all"; handleFilterChange({detail: "all"}); }}
        >
          For you
        </button>
        <button
          class="filter-pill {searchFilter === "following" ? "active" : ""}"
          on:click={() => { searchFilter = "following"; handleFilterChange({detail: "following"}); }}
        >
          Following
        </button>
        <button
          class="filter-pill {searchFilter === "verified" ? "active" : ""}"
          on:click={() => { searchFilter = "verified"; handleFilterChange({detail: "verified"}); }}
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
          class="category-dropdown {isDarkMode ? "category-dropdown-dark" : ""}"
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
      <div class="tabs-container {isDarkMode ? "tabs-container-dark" : ""}">
        <div class="modern-tabs">
          <button
            class="modern-tab {activeTab === "trending" ? "active" : ""}"
            on:click={() => handleTabChange({detail: "trending"})}
          >
            Trending
          </button>
          <button
            class="modern-tab {activeTab === "media" ? "active" : ""}"
            on:click={() => handleTabChange({detail: "media"})}
          >
            Media
          </button>
          <button
            class="modern-tab {activeTab === "people" ? "active" : ""}"
            on:click={() => handleTabChange({detail: "people"})}
          >
            People
          </button>
          <button
            class="modern-tab {activeTab === "communities" ? "active" : ""}"
            on:click={() => handleTabChange({detail: "communities"})}
          >
            Communities
          </button>
          <button
            class="modern-tab {activeTab === "latest" ? "active" : ""}"
            on:click={() => handleTabChange({detail: "latest"})}
          >
            Latest
          </button>
        </div>
      </div>

      <!-- Search results based on active tab -->
      <div class="search-results-container {isDarkMode ? "search-results-container-dark" : ""}">
        {#if activeTab === "trending"}
          <ExploreTrending
            {trends}
            {isTrendsLoading}
            on:hashtagClick={handleHashtagClick}
          />
        {:else if activeTab === "latest"}
          <ExploreLatestResults
            latestThreads={searchResults.latest.threads}
            isLoading={searchResults.latest.isLoading}
          />
        {:else if activeTab === "people"}
          <ExplorePeopleResults
            peopleResults={searchResults.people.users}
            isLoading={searchResults.people.isLoading}
            totalCount={searchResults.people.totalCount}
            {peoplePerPage}
            currentPage={peopleCurrentPage}
            on:pageChange={handlePeoplePageChange}
            on:peoplePerPageChange={handlePeoplePerPageChange}
            on:follow={handleFollowUser}
            on:profileClick={handleProfileClick}
          />
        {:else if activeTab === "media"}
          <!-- Media results component -->
          <ExploreMediaResults
            media={searchResults.media.threads}
            isLoading={searchResults.media.isLoading || isLoadingUsers}
            hasMore={mediaPage * 12 < searchResults.media.totalCount}
            on:loadMore={loadMoreMedia}
          />
        {:else if activeTab === "communities"}
          <!-- Communities results component -->
          <ExploreCommunityResults
            communityResults={searchResults.communities.communities}
            isLoading={searchResults.communities.isLoading}
            totalCount={searchResults.communities.totalCount}
            {communitiesPerPage}
            currentPage={communitiesCurrentPage}
            on:pageChange={handleCommunitiesPageChange}
            on:communitiesPerPageChange={handleCommunitiesPerPageChange}
            on:joinRequest={handleJoinCommunity}
          />
        {/if}
      </div>
    {:else}
      <!-- When not searching, show tabs to select between Trending, People and Communities -->
      <div class="tabs-container {isDarkMode ? "tabs-container-dark" : ""}">
        <div class="modern-tabs">
          <button
            class="modern-tab {defaultActiveTab === "trending" ? "active" : ""}"
            on:click={() => handleDefaultTabChange("trending")}
          >
            <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="tab-icon"><polyline points="23 6 13.5 15.5 8.5 10.5 1 18"></polyline><polyline points="17 6 23 6 23 12"></polyline></svg>
            <span>Trending</span>
          </button>
          <button
            class="modern-tab {defaultActiveTab === "media" ? "active" : ""}"
            on:click={() => handleDefaultTabChange("media")}
          >
            <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="tab-icon"><rect x="3" y="3" width="18" height="18" rx="2" ry="2"></rect><circle cx="8.5" cy="8.5" r="1.5"></circle><polyline points="21 15 16 10 5 21"></polyline></svg>
            <span>Media</span>
          </button>
          <button
            class="modern-tab {defaultActiveTab === "people" ? "active" : ""}"
            on:click={() => handleDefaultTabChange("people")}
          >
            <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="tab-icon"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"></path><circle cx="12" cy="7" r="4"></circle></svg>
            <span>People</span>
          </button>
          <button
            class="modern-tab {defaultActiveTab === "communities" ? "active" : ""}"
            on:click={() => handleDefaultTabChange("communities")}
          >
            <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="tab-icon"><path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path><circle cx="9" cy="7" r="4"></circle><path d="M23 21v-2a4 4 0 0 0-3-3.87"></path><path d="M16 3.13a4 4 0 0 1 0 7.75"></path></svg>
            <span>Communities</span>
          </button>
          <button
            class="modern-tab {defaultActiveTab === "latest" ? "active" : ""}"
            on:click={() => handleDefaultTabChange("latest")}
          >
            <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="tab-icon"><circle cx="12" cy="12" r="10"></circle><polyline points="12 6 12 12 16 14"></polyline></svg>
            <span>Latest</span>
          </button>
        </div>
      </div>

      <!-- Content based on selected default tab -->
      <div class="content-container {isDarkMode ? "content-container-dark" : ""}">
        {#if defaultActiveTab === "trending"}
          <div class="section-header">
            <h2 class="section-title {isDarkMode ? "section-title-dark" : ""}">
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
        {:else if defaultActiveTab === "media"}
          <div class="section-header">
            <h2 class="section-title {isDarkMode ? "section-title-dark" : ""}">
              <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="section-icon"><rect x="3" y="3" width="18" height="18" rx="2" ry="2"></rect><circle cx="8.5" cy="8.5" r="1.5"></circle><polyline points="21 15 16 10 5 21"></polyline></svg>
              Media
            </h2>
            <p class="section-description">Posts with photos and videos</p>
          </div>

          <ExploreMediaResults
            media={searchResults.media.threads}
            isLoading={searchResults.media.isLoading || isLoadingUsers}
            hasMore={mediaPage * 12 < searchResults.media.totalCount}
            on:loadMore={loadMoreMedia}
          />
        {:else if defaultActiveTab === "people"}
          <div class="section-header">
            <h2 class="section-title {isDarkMode ? "section-title-dark" : ""}">
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
                <div class="card {isDarkMode ? "card-dark" : ""}">
                  <ProfileCard
                    id={user.id}
                    username={user.username}
                    displayName={user.displayName}
                    avatar={user.avatar}
                    bio={user.bio}
                    isVerified={user.isVerified}
                    followerCount={user.followerCount}
                    isFollowing={user.isFollowing}
                    fuzzyMatchScore={user.fuzzyMatchScore}
                    onToggleFollow={() => handleToggleFollow(user.id)}
                  />
                </div>
              {/each}
            </div>
          {:else}
            <div class="empty-state {isDarkMode ? "empty-state-dark" : ""}">
              <p class="empty-state-message">No users found. Try a different filter.</p>
            </div>
          {/if}
        {:else if defaultActiveTab === "communities"}
          <div class="section-header">
            <h2 class="section-title {isDarkMode ? "section-title-dark" : ""}">
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
            <div class="empty-state {isDarkMode ? "empty-state-dark" : ""}">
              <p class="empty-state-message">No communities found. Try a different filter.</p>
            </div>
          {/if}
        {:else if defaultActiveTab === "latest"}
          <div class="section-header">
            <h2 class="section-title {isDarkMode ? "section-title-dark" : ""}">
              <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="section-icon"><circle cx="12" cy="12" r="10"></circle><polyline points="12 6 12 12 16 14"></polyline></svg>
              Latest
            </h2>
            <p class="section-description">Recent posts from everyone</p>
          </div>

          <ExploreLatestResults
            latestThreads={searchResults.latest.threads}
            isLoading={searchResults.latest.isLoading || isLoadingUsers}
          />
        {/if}
      </div>
    {/if}
  </div>
</MainLayout>

<Toast />

<style>

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