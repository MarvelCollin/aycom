<script lang="ts">
  import { onMount } from "svelte";
  import { useTheme } from "../../hooks/useTheme";
  import MoreHorizontalIcon from "svelte-feather-icons/src/icons/MoreHorizontalIcon.svelte";
  import SearchIcon from "svelte-feather-icons/src/icons/SearchIcon.svelte";
  import XIcon from "svelte-feather-icons/src/icons/XIcon.svelte";
  import type { ISuggestedFollow } from "../../interfaces/ISocialMedia";
import type { ITrend } from "../../interfaces/ITrend";  import { getTrends } from "../../api/trends";
  import { getSuggestedUsers } from "../../api/suggestions";
  import { followUser, unfollowUser, searchUsers } from "../../api/user";
  import { toastStore } from "../../stores/toastStore";
  import { isAuthenticated } from "../../utils/auth";
  import { debounce } from "../../utils/helpers";
  import { transformApiUsers } from "../../utils/userTransform";
  import type { StandardUser } from "../../utils/userTransform";

  interface ExtendedSuggestedFollow extends ISuggestedFollow {
    isFollowingLoading?: boolean;
  }

  export let isDarkMode = false;
  export let trends: ITrend[] = [];
  export let suggestedFollows: ExtendedSuggestedFollow[] = [];
  export let isTabletView = false;
  let isTrendsLoading = true;
  let isFollowSuggestionsLoading = true;
  let windowWidth = 0;
  let searchQuery = "";
  let showSearch = false;
  const followingStatus: Record<string, boolean> = {};
  const followLoading: Record<string, boolean> = {};

  let searchResults: StandardUser[] = [];
  let isSearching = false;
  let showSearchResults = false;
  let searchContainer: HTMLElement;
  onMount(() => {
    fetchTrends();
    fetchSuggestedUsers();

    const checkWidth = () => {
      windowWidth = window.innerWidth;
    };

    checkWidth();
    window.addEventListener("resize", checkWidth);

    const handleClickOutside = (event) => {
      if (searchContainer && !searchContainer.contains(event.target)) {
        showSearchResults = false;
      }
    };

    document.addEventListener("click", handleClickOutside);

    return () => {
      window.removeEventListener("resize", checkWidth);
      document.removeEventListener("click", handleClickOutside);
    };
  });

  async function fetchTrends() {
    isTrendsLoading = true;
    try {

      const fetchedTrends = await getTrends(10);
      trends = fetchedTrends || [];
      console.log("Fetched trends:", trends);
    } catch (error) {
      console.error("Error loading trends:", error);
      toastStore.showToast("Failed to load trends", "error");
      trends = [];
    } finally {
      isTrendsLoading = false;
    }
  }

  async function fetchSuggestedUsers() {
    isFollowSuggestionsLoading = true;
    try {
      const users = await getSuggestedUsers(isTabletView ? 3 : 3, "follower_count");
      suggestedFollows = users || [];
    } catch (error) {

      console.debug("Note: Could not load suggested users - you may need to be logged in");

      suggestedFollows = [];

      if (suggestedFollows.length === 0) {
        console.debug("Using default suggested users as fallback");
      }
    } finally {
      isFollowSuggestionsLoading = false;
    }
  }

  async function handleToggleFollow(userId: string) {
    if (!userId) {
      console.error("Invalid user ID for follow action");
      return;
    }

    if (!isAuthenticated()) {
      window.location.href = "/login";
      return;
    }

    if (followLoading[userId]) return;

    try {
      followLoading[userId] = true;

      const userIndex = suggestedFollows.findIndex(user =>
        (user.user_id === userId) || (user.id === userId)
      );
      if (userIndex === -1) {
        console.error("User not found in suggested follows list");
        return;
      }

      const user = suggestedFollows[userIndex];
      const wasFollowing = user.is_following;

      suggestedFollows = suggestedFollows.map(user => {
        if (user.user_id === userId) {
          return { ...user, is_following: !user.is_following };
        }
        return user;
      });

      let response;
      if (wasFollowing) {
        response = await unfollowUser(userId);
        if (response.success) {
          toastStore.showToast("User unfollowed", "success");
        }
      } else {
        response = await followUser(userId);
        if (response.success) {
          toastStore.showToast("User followed", "success");
        }
      }

      if (!response || !response.success) {
        suggestedFollows = suggestedFollows.map(user => {
          if (user.user_id === userId) {
            return { ...user, is_following: wasFollowing };
          }
          return user;
        });
        toastStore.showToast(response?.message || "Failed to update follow status", "error");
      }
    } catch (error) {
      console.error("Error toggling follow status:", error);

      suggestedFollows = [...suggestedFollows]; 
      toastStore.showToast("Failed to update follow status", "error");
    } finally {
      followLoading[userId] = false;
    }
  }  

  const debouncedSearch = debounce(async (query: string) => {
    if (!query.trim()) {
      searchResults = [];
      showSearchResults = false;
      isSearching = false;
      return;
    }

    try {
      isSearching = true;
      console.log("Searching for users with query:", query.trim());

      const response = await searchUsers(query.trim(), 1, 5);
      console.log("Search API response:", response);

      const users = response.users || [];
      searchResults = transformApiUsers(users);

      console.log("Transformed search results:", searchResults);
      showSearchResults = true;

      if (searchResults.length > 0) {
        console.log(`Found ${searchResults.length} users matching "${query.trim()}"`);
      }

    } catch (error) {
      console.error("Error searching users:", error);
      toastStore.showToast("Failed to search users. Please try again.", "error");
      searchResults = [];
      showSearchResults = false;
    } finally {
      isSearching = false;
    }
  }, 300);

  function handleSearchInput() {
    debouncedSearch(searchQuery);
  }

  function handleSearch(e) {
    if (e.key === "Enter" && searchQuery.trim()) {

      if (searchResults.length > 0) {
        navigateToProfile(searchResults[0].username);
      } else {

        window.location.href = `/explore?q=${encodeURIComponent(searchQuery.trim())}`;
      }
    } else if (e.key === "Escape") {
      clearSearch();
    }
  }

  function navigateToProfile(username: string) {
    window.location.href = `/user/${username}`;
    clearSearch();
  }

  function clearSearch() {
    searchQuery = "";
    searchResults = [];
    showSearchResults = false;
    isSearching = false;
  }

  function handleSearchFocus() {
    if (searchQuery.trim() && searchResults.length > 0) {
      showSearchResults = true;
    }
  }

  function handleSearchBlur() {

    setTimeout(() => {
      showSearchResults = false;
    }, 200);
  }

  function toggleSearch() {
    showSearch = !showSearch;
    if (showSearch) {
      setTimeout(() => {
        document.getElementById("search-input")?.focus();
      }, 100);
    }
  }

  $: {
    suggestedFollows.forEach(user => {
      if (followingStatus[user.username] === undefined) {
        followingStatus[user.username] = user.is_following || false;
      }
    });
  }
</script>

<div class="widgets-container {isTabletView ? "widgets-container-tablet" : ""}">
  <div class="right-sidebar {isDarkMode ? "right-sidebar-dark" : ""}">    <!-- Search Widget (only for desktop and tablet) -->
    {#if !isTabletView}
      <div class="search-widget {isDarkMode ? "search-widget-dark" : ""}" bind:this={searchContainer}>
        <div class="search-input-container">
          <div class="search-icon">
            <SearchIcon size="18" />
          </div>
          <input
            type="text"
            id="search-input"
            placeholder="Search users"
            class="search-input"
            bind:value={searchQuery}
            on:input={handleSearchInput}
            on:keydown={handleSearch}
            on:focus={handleSearchFocus}
            on:blur={handleSearchBlur}
          />
          {#if searchQuery}
            <button class="search-clear-btn" on:click={clearSearch}>
              <XIcon size="16" />
            </button>
          {/if}
        </div>

        <!-- Search Results Dropdown -->
        {#if showSearchResults}
          <div class="search-results {isDarkMode ? "search-results-dark" : ""}">
            {#if isSearching}
              <div class="search-loading">
                <div class="search-loading-spinner"></div>
                <span>Searching...</span>
              </div>
            {:else if searchResults.length > 0}
              <div class="search-results-list">
                {#each searchResults as user}
                  <button
                    class="search-result-item {isDarkMode ? "search-result-item-dark" : ""}"
                    on:click={() => navigateToProfile(user.username)}
                  >
                    <div class="search-result-avatar">
                      <img
                        src={user.profile_picture_url}
                        alt={user.name || user.username}
                      />
                    </div>
                    <div class="search-result-details">
                      <div class="search-result-name">
                        {user.name || user.username}
                        {#if user.is_verified}
                          <span class="verified-badge">✓</span>
                        {/if}
                      </div>
                      <div class="search-result-username">@{user.username}</div>
                      {#if user.bio}
                        <div class="search-result-bio">{user.bio.slice(0, 60)}{user.bio.length > 60 ? "..." : ""}</div>
                      {/if}
                    </div>
                  </button>
                {/each}
              </div>
              {#if searchResults.length === 5}
                <div class="search-show-more">
                  <button
                    class="search-show-more-btn {isDarkMode ? "search-show-more-btn-dark" : ""}"
                    on:click={() => window.location.href = `/explore?q=${encodeURIComponent(searchQuery.trim())}`}
                  >
                    View all results for "{searchQuery}"
                  </button>
                </div>
              {/if}
            {:else}
              <div class="search-no-results">
                <p>No users found for "{searchQuery}"</p>
                <button
                  class="search-show-more-btn {isDarkMode ? "search-show-more-btn-dark" : ""}"
                  on:click={() => window.location.href = `/explore?q=${encodeURIComponent(searchQuery.trim())}`}
                >
                  Search in all content
                </button>
              </div>
            {/if}
          </div>
        {/if}
      </div>
    {/if}

    <!-- Premium Advertisement Widget -->
    <div class="premium-ad-widget {isDarkMode ? "premium-ad-widget-dark" : ""}">
      <a href="/premium" class="premium-ad-link">
        <div class="premium-ad-content">
          <div class="premium-header">
            <div class="premium-checkmark">
              <svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                <path d="M9 12L11 14L15 10M21 12C21 16.9706 16.9706 21 12 21C7.02944 21 3 16.9706 3 12C3 7.02944 7.02944 3 12 3C16.9706 3 21 7.02944 21 12Z" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
              </svg>
            </div>
            <h3 class="premium-title">Get Verified</h3>
          </div>
          <p class="premium-description">
            Stand out with the blue checkmark and increase your credibility on AYCOM
          </p>
          <div class="premium-benefits">
            <div class="premium-benefit">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                <path d="M20 6L9 17L4 12" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
              </svg>
              <span>Blue verification badge</span>
            </div>
            <div class="premium-benefit">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                <path d="M20 6L9 17L4 12" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
              </svg>
              <span>Increased visibility</span>
            </div>
            <div class="premium-benefit">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                <path d="M20 6L9 17L4 12" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
              </svg>
              <span>Enhanced credibility</span>
            </div>
          </div>
          <div class="premium-cta">
            <span class="premium-cta-text">Apply Now</span>
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path d="M5 12H19M19 12L12 5M19 12L12 19" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
          </div>
        </div>
      </a>
    </div>

    <!-- Trends Widget -->
    <div class="sidebar-section {isDarkMode ? "sidebar-section-dark" : ""}">
      <h3 class="sidebar-title">Trends for you</h3>

      {#if isTrendsLoading}
        <div class="trends-loading">
          <div class="trends-loading-spinner"></div>
        </div>
      {:else if trends.length === 0}
        <div class="trends-empty">
          <p>No trends available</p>
        </div>
      {:else}
        <div class="trends-list">
          <!-- Only display the top 3 trending hashtags -->
          {#each trends.slice(0, 3) as trend, i}
            <div class="trend-item {isDarkMode ? "trend-item-dark" : ""}">
              <div class="trend-header">
                <span class="trend-location">{trend.category || "Trending"}</span>
                <button
                  class="trend-more-options {isDarkMode ? "trend-more-options-dark" : ""}"
                  aria-label="More options"
                >
                  <MoreHorizontalIcon size="16" />
                </button>
              </div>
              <div class="trend-name">
                <a href={`/explore?q=${encodeURIComponent((trend.title || trend.name || "").replace(/^#/, ""))}`}>
                  {trend.title || trend.name || ""}
                </a>
              </div>
              <div class="trend-count">{trend.post_count || trend.tweet_count || "0"} posts</div>
            </div>
          {/each}
        </div>
      {/if}

      <a href="/explore" class="trends-show-more {isDarkMode ? "trends-show-more-dark" : ""}">
        Show more
      </a>
    </div>

    <!-- Who to Follow Widget -->
    <div class="sidebar-section {isDarkMode ? "sidebar-section-dark" : ""}">
      <h3 class="sidebar-title">Who to follow</h3>

      {#if isFollowSuggestionsLoading}
        <div class="suggestions-loading">
          <div class="suggestions-loading-spinner"></div>
        </div>
      {:else if suggestedFollows.length === 0}
        <div class="suggestions-empty">
          <p>No suggestions available</p>
        </div>
      {:else}
        <div class="suggestions-list">
          {#each suggestedFollows.slice(0, 3) as user, i}
            <div class="suggestion-item {isDarkMode ? "suggestion-item-dark" : ""}">
              <div class="suggested-user-avatar">
                <img
                  src={user.profile_picture_url}
                  alt={user.name || user.username || "User"}
                  class="suggested-user-img"
                />
              </div>
              <div class="suggestion-details">
                <a href={`/user/${user.username}`} class="suggestion-name">
                  {user.name || user.username || "User"}
                </a>
                <a href={`/user/${user.username}`} class="suggestion-username">
                  @{user.username || "user"}
                </a>
              </div>
              <div class="suggestion-action">
                <button
                  class="follow-button {followingStatus[user.username] ? "following" : ""} {isDarkMode ? "follow-button-dark" : ""}"
                  on:click={() => handleToggleFollow(user.user_id || user.id || "")}
                  disabled={followLoading[user.user_id || user.id || ""]}
                >
                  {#if followLoading[user.user_id || user.id || ""]}
                    <span class="loading-dot"></span>
                  {:else}
                    {followingStatus[user.username] ? "Following" : "Follow"}
                  {/if}
                </button>
              </div>
            </div>
          {/each}
        </div>
      {/if}

      <a href="/discover-people" class="trends-show-more {isDarkMode ? "trends-show-more-dark" : ""}">
        Show more
      </a>
    </div>

    <!-- Footer links - only show on desktop -->
    {#if !isTabletView}
      <div class="footer-links">
        <div class="footer-links-list">
          <a href="/about" class="footer-link">About</a>
          <a href="/help" class="footer-link">Help Center</a>
          <a href="/terms" class="footer-link">Terms of Service</a>
          <a href="/privacy" class="footer-link">Privacy Policy</a>
          <a href="/cookies" class="footer-link">Cookie Policy</a>
          <a href="/accessibility" class="footer-link">Accessibility</a>
          <a href="/ads-info" class="footer-link">Ads Info</a>
        </div>
        <div class="footer-copyright">© {new Date().getFullYear()} AYCOM, Inc.</div>
      </div>
    {/if}
  </div>
</div>

<style>
  .widgets-container {
    height: 100vh;
    position: sticky;
    top: 0;
    overflow-y: auto;
    scrollbar-width: thin;
    scrollbar-color: var(--text-tertiary) transparent;
    transition: all var(--transition-normal);
    padding-right: var(--space-4);
  }

  .widgets-container-tablet {
    height: auto;
    position: relative;
    padding: 0;
    margin-top: var(--space-4);
    border-top: 1px solid var(--border-color);
  }

  .widgets-container::-webkit-scrollbar {
    width: 6px;
  }

  .widgets-container::-webkit-scrollbar-track {
    background: transparent;
  }

  .widgets-container::-webkit-scrollbar-thumb {
    background-color: var(--text-tertiary);
    border-radius: var(--radius-full);
  }

  .right-sidebar {
    padding: var(--space-4) 0;
    height: 100%;
  }

  .widgets-container-tablet .right-sidebar {
    padding: var(--space-4) var(--space-4);
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-4);
  }

  .search-widget {
    margin-bottom: var(--space-4);
    position: sticky;
    top: 0;
    z-index: var(--z-sticky);
    background-color: var(--bg-primary);
    padding: var(--space-2) 0;
    position: relative;
  }

  .search-widget-dark {
    background-color: var(--dark-bg-primary);
  }

  .search-input-container {
    position: relative;
    display: flex;
    align-items: center;
    background-color: var(--bg-secondary);
    border-radius: var(--radius-full);
    padding: 0 var(--space-4);
    transition: background-color var(--transition-fast), box-shadow var(--transition-fast);
  }

  .search-input-container:focus-within {
    background-color: var(--bg-primary);
    box-shadow: 0 0 0 1px var(--color-primary), 0 0 0 4px rgba(var(--color-primary-rgb), 0.2);
  }

  .search-icon {
    color: var(--text-secondary);
    display: flex;
    align-items: center;
  }

  .search-input {
    flex: 1;
    border: none;
    background: transparent;
    padding: var(--space-3);
    font-size: var(--font-size-base);
    color: var(--text-primary);
    outline: none;
    width: 100%;
  }

  .search-input::placeholder {
    color: var(--text-secondary);
  }

  .search-clear-btn {
    background: transparent;
    border: none;
    color: var(--text-secondary);
    display: flex;
    align-items: center;
    justify-content: center;
    padding: var(--space-1);
    border-radius: 50%;
    cursor: pointer;
  }

  .search-clear-btn:hover {
    background-color: var(--bg-hover);
    color: var(--color-primary);
  }

  .search-results {
    position: absolute;
    top: 100%;
    left: 0;
    right: 0;
    background-color: var(--bg-primary);
    border-radius: var(--radius-lg);
    box-shadow: 0 8px 32px rgba(0, 0, 0, 0.12);
    border: 1px solid var(--border-color);
    max-height: 400px;
    overflow-y: auto;
    z-index: 1000;
    margin-top: var(--space-2);
  }

  .search-results-dark {
    background-color: var(--dark-bg-primary);
    border-color: var(--border-color-dark);
    box-shadow: 0 8px 32px rgba(0, 0, 0, 0.24);
  }

  .search-loading {
    display: flex;
    align-items: center;
    justify-content: center;
    padding: var(--space-4);
    gap: var(--space-2);
    color: var(--text-secondary);
  }

  .search-loading-spinner {
    width: 16px;
    height: 16px;
    border: 2px solid rgba(var(--color-primary-rgb), 0.2);
    border-top-color: var(--color-primary);
    border-radius: 50%;
    animation: spin 1s linear infinite;
  }

  .search-results-list {
    padding: var(--space-2) 0;
  }

  .search-result-item {
    width: 100%;
    display: flex;
    align-items: flex-start;
    padding: var(--space-3) var(--space-4);
    border: none;
    background: transparent;
    text-align: left;
    cursor: pointer;
    transition: background-color var(--transition-fast);
  }

  .search-result-item:hover {
    background-color: var(--bg-hover);
  }

  .search-result-item-dark:hover {
    background-color: var(--dark-hover-bg);
  }

  .search-result-avatar {
    width: 40px;
    height: 40px;
    border-radius: 50%;
    overflow: hidden;
    margin-right: var(--space-3);
    flex-shrink: 0;
  }

  .search-result-avatar img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  .search-result-details {
    flex: 1;
    min-width: 0;
  }

  .search-result-name {
    font-weight: var(--font-weight-bold);
    color: var(--text-primary);
    display: flex;
    align-items: center;
    gap: var(--space-1);
    margin-bottom: var(--space-1);
  }

  .verified-badge {
    color: var(--color-primary);
    font-size: var(--font-size-sm);
  }

  .search-result-username {
    font-size: var(--font-size-sm);
    color: var(--text-secondary);
    margin-bottom: var(--space-1);
  }

  .search-result-bio {
    font-size: var(--font-size-sm);
    color: var(--text-secondary);
    line-height: 1.4;
  }

  .search-no-results {
    padding: var(--space-4);
    text-align: center;
    color: var(--text-secondary);
  }

  .search-show-more {
    padding: var(--space-2) var(--space-4);
    border-top: 1px solid var(--border-color);
  }

  .search-show-more-btn {
    width: 100%;
    padding: var(--space-2);
    background: transparent;
    border: none;
    color: var(--color-primary);
    text-align: center;
    cursor: pointer;
    border-radius: var(--radius-md);
    transition: background-color var(--transition-fast);
    font-size: var(--font-size-sm);
  }

  .search-show-more-btn:hover {
    background-color: var(--bg-hover);
  }

  .search-show-more-btn-dark:hover {
    background-color: var(--dark-hover-bg);
  }

  .premium-ad-widget {
    margin-bottom: var(--space-4);
    border-radius: var(--radius-lg);
    overflow: hidden;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
    background: linear-gradient(135deg, #1DA1F2 0%, #14A085 100%);
    position: relative;
  }

  .premium-ad-widget::before {
    content: '';
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: linear-gradient(135deg, rgba(29, 161, 242, 0.1) 0%, rgba(20, 160, 133, 0.1) 100%);
    pointer-events: none;
  }

  .premium-ad-widget-dark {
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
  }

  .premium-ad-link {
    display: block;
    text-decoration: none;
    color: white;
    transition: transform var(--transition-fast), box-shadow var(--transition-fast);
  }

  .premium-ad-link:hover {
    transform: translateY(-2px);
    box-shadow: 0 4px 16px rgba(0, 0, 0, 0.15);
  }

  .premium-ad-content {
    padding: var(--space-4);
    position: relative;
    z-index: 1;
  }

  .premium-header {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    margin-bottom: var(--space-3);
  }

  .premium-checkmark {
    color: white;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .premium-title {
    font-size: var(--font-size-lg);
    font-weight: var(--font-weight-bold);
    color: white;
    margin: 0;
  }

  .premium-description {
    color: rgba(255, 255, 255, 0.9);
    font-size: var(--font-size-sm);
    margin-bottom: var(--space-3);
    line-height: 1.4;
  }

  .premium-benefits {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
    margin-bottom: var(--space-4);
  }

  .premium-benefit {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    color: rgba(255, 255, 255, 0.95);
    font-size: var(--font-size-sm);
  }

  .premium-benefit svg {
    flex-shrink: 0;
    color: rgba(255, 255, 255, 0.9);
  }

  .premium-cta {
    display: flex;
    align-items: center;
    justify-content: space-between;
    background-color: rgba(255, 255, 255, 0.15);
    padding: var(--space-2) var(--space-3);
    border-radius: var(--radius-full);
    backdrop-filter: blur(10px);
    border: 1px solid rgba(255, 255, 255, 0.2);
    transition: all var(--transition-fast);
  }

  .premium-ad-link:hover .premium-cta {
    background-color: rgba(255, 255, 255, 0.25);
    transform: translateX(4px);
  }

  .premium-cta-text {
    font-weight: var(--font-weight-bold);
    color: white;
    font-size: var(--font-size-sm);
  }

  .premium-cta svg {
    color: white;
    transition: transform var(--transition-fast);
  }

  .premium-ad-link:hover .premium-cta svg {
    transform: translateX(2px);
  }

  .widgets-container-tablet .premium-ad-widget {
    flex: 1;
    min-width: 280px;
    margin-bottom: 0;
  }

  @media (max-width: 992px) {
    .widgets-container-tablet .right-sidebar {
      padding: var(--space-4) 0;
    }

    .widgets-container-tablet .sidebar-section {
      min-width: 240px;
    }

    .widgets-container-tablet .premium-ad-widget {
      min-width: 240px;
    }
  }

  @media (max-width: 768px) {
    .premium-ad-content {
      padding: var(--space-3);
    }

    .premium-title {
      font-size: var(--font-size-base);
    }

    .premium-description {
      font-size: var(--font-size-xs);
    }

    .premium-benefit {
      font-size: var(--font-size-xs);
    }

    .premium-cta-text {
      font-size: var(--font-size-xs);
    }
  }

  .sidebar-section {
    background-color: var(--bg-secondary);
    border-radius: var(--radius-lg);
    padding: var(--space-4);
    margin-bottom: var(--space-4);
  }

  .widgets-container-tablet .sidebar-section {
    flex: 1;
    min-width: 280px;
    margin-bottom: 0;
  }

  .sidebar-section-dark {
    background-color: var(--dark-bg-secondary);
  }

  .sidebar-title {
    font-size: var(--font-size-lg);
    font-weight: 700;
    margin-bottom: var(--space-4);
  }

  .trends-loading,
  .suggestions-loading {
    display: flex;
    justify-content: center;
    padding: var(--space-4) 0;
  }

  .trends-loading-spinner,
  .suggestions-loading-spinner {
    width: 24px;
    height: 24px;
    border: 3px solid rgba(var(--color-primary-rgb), 0.2);
    border-top-color: var(--color-primary);
    border-radius: 50%;
    animation: spin 1s linear infinite;
  }

  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }

  .trends-empty,
  .suggestions-empty {
    text-align: center;
    padding: var(--space-4) 0;
    color: var(--text-secondary);
  }

  .trends-list {
    margin-bottom: var(--space-2);
  }

  .trend-item {
    padding: var(--space-3) 0;
    border-bottom: 1px solid var(--border-color);
    transition: background-color var(--transition-fast);
    cursor: pointer;
  }

  .trend-item:last-child {
    border-bottom: none;
  }

  .trend-item:hover {
    background-color: var(--bg-hover);
  }

  .trend-item-dark:hover {
    background-color: var(--dark-hover-bg);
  }

  .trend-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: var(--space-1);
  }

  .trend-location {
    font-size: var(--font-size-xs);
    color: var(--text-secondary);
  }

  .trend-more-options {
    width: 24px;
    height: 24px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    background: transparent;
    border: none;
    color: var(--text-secondary);
    cursor: pointer;
    transition: background-color var(--transition-fast), color var(--transition-fast);
  }

  .trend-more-options:hover {
    background-color: var(--bg-hover);
    color: var(--color-primary);
  }

  .trend-more-options-dark:hover {
    background-color: var(--dark-hover-bg);
  }

  .trend-name {
    font-weight: var(--font-weight-bold);
    margin-bottom: var(--space-1);
    word-break: break-word;
  }

  .trend-name a {
    color: var(--text-primary);
    text-decoration: none;
  }

  .trend-count {
    font-size: var(--font-size-xs);
    color: var(--text-secondary);
  }

  .trends-show-more {
    display: block;
    padding: var(--space-3) 0;
    color: var(--color-primary);
    text-decoration: none;
    font-size: var(--font-size-sm);
    transition: color var(--transition-fast);
  }

  .trends-show-more:hover {
    text-decoration: underline;
  }

  .trends-show-more-dark {
    color: var(--color-primary-light);
  }

  .suggestions-list {
    margin-bottom: var(--space-2);
  }

  .suggestion-item {
    display: flex;
    align-items: center;
    padding: var(--space-3) 0;
    border-bottom: 1px solid var(--border-color);
    transition: background-color var(--transition-fast);
  }

  .suggestion-item:last-child {
    border-bottom: none;
  }

  .suggestion-item:hover {
    background-color: var(--bg-hover);
  }

  .suggestion-item-dark:hover {
    background-color: var(--dark-hover-bg);
  }

  .suggested-user-avatar {
    width: 48px;
    height: 48px;
    border-radius: 50%;
    overflow: hidden;
    margin-right: var(--space-3);
    flex-shrink: 0;
  }

  .suggested-user-img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  .suggestion-details {
    flex: 1;
    min-width: 0;
    margin-right: var(--space-2);
  }

  .suggestion-name {
    font-weight: var(--font-weight-bold);
    color: var(--text-primary);
    text-decoration: none;
    display: block;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .suggestion-username {
    font-size: var(--font-size-sm);
    color: var(--text-secondary);
    text-decoration: none;
    display: block;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .suggestion-action {
    flex-shrink: 0;
  }

  .follow-button {
    background-color: var(--color-primary);
    color: white !important;
    border: none;
    border-radius: var(--radius-full);
    padding: var(--space-1) var(--space-3);
    font-size: var(--font-size-sm);
    font-weight: var(--font-weight-bold);
    cursor: pointer;
    transition: background-color var(--transition-fast);
    min-width: 80px;
    text-align: center;
  }

  .follow-button:hover {
    background-color: var(--color-primary-dark);
    color: white !important;
  }

  .follow-button.following {
    background-color: transparent;
    color: var(--text-primary);
    border: 1px solid var(--border-color);
  }

  .follow-button.following:hover {
    background-color: var(--error-bg);
    color: var(--error);
    border-color: var(--error-bg);
  }

  .follow-button-dark {
    background-color: var(--color-primary);
    color: white !important;
  }

  .follow-button-dark:hover {
    background-color: var(--color-primary-dark);
    color: white !important;
  }

  .follow-button-dark.following {
    background-color: transparent;
    color: var(--text-primary-dark);
    border: 1px solid var(--border-color-dark);
  }

  .loading-dot {
    display: inline-block;
    width: 8px;
    height: 8px;
    background-color: currentColor;
    border-radius: 50%;
    animation: pulse 1.5s infinite ease-in-out;
  }

  @keyframes pulse {
    0% {
      transform: scale(0.5);
      opacity: 0.3;
    }
    50% {
      transform: scale(1);
      opacity: 1;
    }
    100% {
      transform: scale(0.5);
      opacity: 0.3;
    }
  }

  .footer-links {
    padding: var(--space-2) 0;
    font-size: var(--font-size-xs);
    color: var(--text-tertiary);
  }

  .footer-links-list {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-2);
    margin-bottom: var(--space-2);
  }

  .footer-link {
    color: var(--text-tertiary);
    text-decoration: none;
    transition: color var(--transition-fast);
  }

  .footer-link:hover {
    color: var(--color-primary);
    text-decoration: underline;
  }

  .footer-copyright {
    margin-top: var(--space-2);
  }

  @media (max-width: 1400px) {
    .sidebar-section {
      padding: var(--space-3);
    }

    .suggested-user-avatar {
      width: 40px;
      height: 40px;
    }
  }

  @media (max-width: 1200px) {
    .footer-links-list {
      flex-direction: column;
      gap: var(--space-1);
    }

    .follow-button {
      min-width: 70px;
      padding: var(--space-1) var(--space-2);
    }
  }

  @media (max-width: 992px) {
    .widgets-container-tablet .right-sidebar {
      padding: var(--space-4) 0;
    }

    .widgets-container-tablet .sidebar-section {
      min-width: 240px;
    }

    .widgets-container-tablet .premium-ad-widget {
      min-width: 240px;
    }
  }

  @media (max-width: 768px) {
    .premium-ad-content {
      padding: var(--space-3);
    }

    .premium-title {
      font-size: var(--font-size-base);
    }

    .premium-description {
      font-size: var(--font-size-xs);
    }

    .premium-benefit {
      font-size: var(--font-size-xs);
    }

    .premium-cta-text {
      font-size: var(--font-size-xs);
    }
  }
</style>