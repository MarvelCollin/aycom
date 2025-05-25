<script lang="ts">
  import { onMount } from 'svelte';
  import { useTheme } from '../../hooks/useTheme';
  import MoreHorizontalIcon from 'svelte-feather-icons/src/icons/MoreHorizontalIcon.svelte';
  import SearchIcon from 'svelte-feather-icons/src/icons/SearchIcon.svelte';
  import XIcon from 'svelte-feather-icons/src/icons/XIcon.svelte';
  import type { ITrend, ISuggestedFollow } from '../../interfaces/ISocialMedia';
  import { getTrends } from '../../api/trends';
  import { getSuggestedUsers } from '../../api/suggestions';
  import { followUser, unfollowUser } from '../../api/user';
  import { toastStore } from '../../stores/toastStore';
  import { isAuthenticated } from '../../utils/auth';

  // Extend the ISuggestedFollow interface with our UI-specific properties
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
  let searchQuery = '';
  let showSearch = false;
  let followingStatus: Record<string, boolean> = {};
  let followLoading: Record<string, boolean> = {};

  onMount(() => {
    fetchTrends();
    fetchSuggestedUsers();
    
    // Check window width for responsive design
    const checkWidth = () => {
      windowWidth = window.innerWidth;
    };
    
    checkWidth();
    window.addEventListener('resize', checkWidth);
    
    return () => {
      window.removeEventListener('resize', checkWidth);
    };
  });

  async function fetchTrends() {
    isTrendsLoading = true;
    try {
      const fetchedTrends = await getTrends(5);
      trends = fetchedTrends || [];
    } catch (error) {
      console.error('Error loading trends:', error);
      toastStore.showToast('Failed to load trends', 'error');
      trends = [];
    } finally {
      isTrendsLoading = false;
    }
  }

  async function fetchSuggestedUsers() {
    isFollowSuggestionsLoading = true;
    try {
      const users = await getSuggestedUsers();
      suggestedFollows = users || [];
    } catch (error) {
      // Don't show toast error for auth issues, just log quietly
      console.debug('Note: Could not load suggested users - you may need to be logged in');
      suggestedFollows = [];
    } finally {
      isFollowSuggestionsLoading = false;
    }
  }

  async function handleToggleFollow(userId: string) {
    if (!isAuthenticated()) {
      window.location.href = '/login';
      return;
    }
    
    if (followLoading[userId]) return;
    
    try {
      followLoading[userId] = true;
      followingStatus = {...followingStatus};
      
      if (followingStatus[userId]) {
        await unfollowUser(userId);
        followingStatus[userId] = false;
        toastStore.showToast('User unfollowed', 'success');
      } else {
        await followUser(userId);
        followingStatus[userId] = true;
        toastStore.showToast('User followed', 'success');
      }
    } catch (error) {
      console.error('Error toggling follow status:', error);
      toastStore.showToast('Failed to update follow status', 'error');
    } finally {
      followLoading[userId] = false;
      followingStatus = {...followingStatus};
    }
  }

  function handleSearch(e) {
    if (e.key === 'Enter' && searchQuery.trim()) {
      window.location.href = `/explore?q=${encodeURIComponent(searchQuery.trim())}`;
    }
  }

  function toggleSearch() {
    showSearch = !showSearch;
    if (showSearch) {
      setTimeout(() => {
        document.getElementById('search-input')?.focus();
      }, 100);
    }
  }

  // Initialize following status for each suggested user
  $: {
    suggestedFollows.forEach(user => {
      if (followingStatus[user.username] === undefined) {
        followingStatus[user.username] = user.isFollowing || false;
      }
    });
  }
</script>

<div class="widgets-container {isTabletView ? 'widgets-container-tablet' : ''}">
  <div class="right-sidebar {isDarkMode ? 'right-sidebar-dark' : ''}">
    <!-- Search Widget (only for desktop and tablet) -->
    {#if !isTabletView}
      <div class="search-widget {isDarkMode ? 'search-widget-dark' : ''}">
        <div class="search-input-container">
          <div class="search-icon">
            <SearchIcon size="18" />
          </div>
          <input 
            type="text" 
            id="search-input"
            placeholder="Search" 
            class="search-input"
            bind:value={searchQuery}
            on:keydown={handleSearch}
          />
          {#if searchQuery}
            <button class="search-clear-btn" on:click={() => searchQuery = ''}>
              <XIcon size="16" />
            </button>
          {/if}
        </div>
      </div>
    {/if}
    
    <!-- Trends Widget -->
    <div class="sidebar-section {isDarkMode ? 'sidebar-section-dark' : ''}">
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
          {#each trends as trend, i}
            <div class="trend-item {isDarkMode ? 'trend-item-dark' : ''}">
              <div class="trend-header">
                <span class="trend-location">{trend.category || 'Trending'}</span>
                <button 
                  class="trend-more-options {isDarkMode ? 'trend-more-options-dark' : ''}" 
                  aria-label="More options"
                >
                  <MoreHorizontalIcon size="16" />
                </button>
              </div>
              <div class="trend-name">
                <a href={`/explore?q=${encodeURIComponent(trend.title)}`}>
                  {trend.title}
                </a>
              </div>
              <div class="trend-count">{trend.postCount || '0'} posts</div>
            </div>
          {/each}
        </div>
      {/if}
      
      <a href="/explore" class="trends-show-more {isDarkMode ? 'trends-show-more-dark' : ''}">
        Show more
      </a>
    </div>

    <!-- Who to Follow Widget -->
    <div class="sidebar-section {isDarkMode ? 'sidebar-section-dark' : ''}">
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
          {#each suggestedFollows.slice(0, isTabletView ? 3 : 5) as user, i}
            <div class="suggestion-item {isDarkMode ? 'suggestion-item-dark' : ''}">
              <div class="suggestion-avatar">
                <a href={`/user/${user.username}`}>
                  <img 
                    src={user.avatar || 'https://secure.gravatar.com/avatar/0?d=mp'} 
                    alt={user.displayName || user.username || 'User'} 
                  />
                </a>
              </div>
              <div class="suggestion-details">
                <a href={`/user/${user.username}`} class="suggestion-name">
                  {user.displayName || user.username || 'User'}
                </a>
                <a href={`/user/${user.username}`} class="suggestion-username">
                  @{user.username || 'user'}
                </a>
              </div>
              <div class="suggestion-action">
                <button 
                  class="follow-button {followingStatus[user.username] ? 'following' : ''} {isDarkMode ? 'follow-button-dark' : ''}"
                  on:click={() => handleToggleFollow(user.username)}
                  disabled={followLoading[user.username]}
                >
                  {#if followLoading[user.username]}
                    <span class="loading-dot"></span>
                  {:else}
                    {followingStatus[user.username] ? 'Following' : 'Follow'}
                  {/if}
                </button>
              </div>
            </div>
          {/each}
        </div>
      {/if}
      
      <a href="/discover-people" class="trends-show-more {isDarkMode ? 'trends-show-more-dark' : ''}">
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
  
  /* Search widget */
  .search-widget {
    margin-bottom: var(--space-4);
    position: sticky;
    top: 0;
    z-index: var(--z-sticky);
    background-color: var(--bg-primary);
    padding: var(--space-2) 0;
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
  
  /* Sidebar sections */
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
  
  /* Trends list styles */
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
  
  /* Suggestions list styles */
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
  
  .suggestion-avatar {
    width: 48px;
    height: 48px;
    border-radius: 50%;
    overflow: hidden;
    margin-right: var(--space-3);
    flex-shrink: 0;
  }
  
  .suggestion-avatar img {
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
  
  /* Footer styles */
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
  
  /* Responsive styles */
  @media (max-width: 1400px) {
    .sidebar-section {
      padding: var(--space-3);
    }
    
    .suggestion-avatar {
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
  }
</style>