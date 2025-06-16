<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { useTheme } from '../../hooks/useTheme';
  import { createLoggerWithPrefix } from '../../utils/logger';
  
  const logger = createLoggerWithPrefix('ExploreSearch');
  const dispatch = createEventDispatcher();
  const { theme } = useTheme();
  
  // Reactive declarations
  $: isDarkMode = $theme === 'dark';
  
  // Props
  export let searchQuery = '';
  export let recentSearches: string[] = [];
  export let recommendedProfiles: Array<{
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
  export let showRecentSearches = false;
  export let isSearching = false;
  export let isLoadingRecommendations = false;
  
  // Handle search input
  function handleSearchInput(event) {
    const value = event.target.value;
    logger.debug('Search input changed', { value });
    
    // Dispatch the input event immediately
    dispatch('input', value);
    
    // If the query is empty, hide recent searches
    if (!value || value.length === 0) {
      showRecentSearches = false;
      logger.debug('Hiding recent searches due to empty query');
    }
    
    // Log search query changes
    console.log('Search query changed to:', value);
  }
  
  // Function to get color based on fuzzy match score
  function getFuzzyMatchColor(score: number | undefined): string {
    if (!score) return '';
    
    if (score >= 0.8) return 'fuzzy-match-high';
    if (score >= 0.6) return 'fuzzy-match-medium';
    if (score >= 0.3) return 'fuzzy-match-low';
    return '';
  }
  
  // Function to get fuzzy match label
  function getFuzzyMatchLabel(score: number | undefined): string {
    if (!score) return '';
    
    if (score >= 0.8) return 'Strong match';
    if (score >= 0.6) return 'Good match';
    if (score >= 0.3) return 'Possible match';
    return '';
  }
  
  // Handle search execution
  function executeSearch() {
    logger.debug('Search executed', { query: searchQuery });
    dispatch('search');
  }
  
  // Handle search field focus
  function handleFocus() {
    logger.debug('Search field focused');
    dispatch('focus');
  }
  
  // Handle search keystroke
  function handleKeydown(event) {
    // If Enter key is pressed, immediately trigger search
    if (event.key === 'Enter') {
      logger.debug('Search triggered via Enter key', { query: searchQuery || '(empty)' });
      
      // Hide dropdowns when search is executed
      showRecentSearches = false;
      
      // This will ensure the full search results are shown instead of the dropdown
      isSearching = true;
      
      // Dispatch the search event to trigger the full search
      dispatch('search');
      
      // Force the dropdown to close by dispatching an event
      dispatch('enterPressed');
    }
  }
  
  // Handle selection of a recent search
  function selectRecentSearch(search: string) {
    logger.debug('Recent search selected', { search });
    dispatch('selectRecentSearch', search);
  }
  
  // Clear recent searches
  function clearRecentSearches() {
    logger.debug('Recent searches cleared');
    dispatch('clearRecentSearches');
  }
  
  // Clear search input
  function clearSearch() {
    logger.debug('Search input cleared');
    dispatch('clearSearch');
  }
  
  // Log when recommendations are loaded
  $: {
    if (!isLoadingRecommendations && recommendedProfiles.length > 0) {
      logger.debug('Profile recommendations loaded', { count: recommendedProfiles.length });
    }
  }
</script>

<div class="search-container">
  <!-- Search bar -->
  <div class="search-input-wrapper {isDarkMode ? 'search-input-wrapper-dark' : ''}">
    <div class="search-icon-container">
      <svg xmlns="http://www.w3.org/2000/svg" class="search-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
      </svg>
    </div>
    
    <input 
      type="text" 
      placeholder="Search with Fuzzy Matching (typo-friendly)" 
      value={searchQuery}
      on:input={handleSearchInput}
      on:focus={handleFocus}
      on:keydown={handleKeydown}
      class="search-input {isDarkMode ? 'search-input-dark' : ''}"
      aria-label="Fuzzy search with Damerau-Levenshtein distance (0.3 threshold)"
      title="Search using Damerau-Levenshtein fuzzy matching - tolerates misspellings (now with 0.3 threshold)"
    />
    
    {#if searchQuery}
      <button 
        class="search-clear-button {isDarkMode ? 'search-clear-button-dark' : ''}"
        on:click={clearSearch}
        aria-label="Clear search"
      >
        <svg xmlns="http://www.w3.org/2000/svg" class="search-clear-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor">
          <circle cx="12" cy="12" r="10" fill={isDarkMode ? "#4e555d" : "#e7eaec"}></circle>
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 9l-6 6M9 9l6 6" />
        </svg>
      </button>
    {/if}
    
    <!-- Fuzzy search indicator -->
    <div class="fuzzy-search-indicator">
      <span class="fuzzy-search-badge">
        <svg xmlns="http://www.w3.org/2000/svg" class="fuzzy-search-icon" viewBox="0 0 20 20" fill="currentColor">
          <path fill-rule="evenodd" d="M8 4a4 4 0 100 8 4 4 0 000-8zM2 8a6 6 0 1110.89 3.476l4.817 4.817a1 1 0 01-1.414 1.414l-4.816-4.816A6 6 0 012 8z" clip-rule="evenodd" />
        </svg>
        Fuzzy Search Enabled
      </span>
    </div>
  </div>
  
  <!-- Recent searches dropdown -->
  {#if showRecentSearches && recentSearches.length > 0 && !searchQuery}
    <div class="search-dropdown {isDarkMode ? 'search-dropdown-dark' : ''}">
      <div class="search-dropdown-header">
        <h3 class="search-dropdown-title">Recent</h3>
        <button 
          class="search-dropdown-clear-button"
          on:click={clearRecentSearches}
        >
          Clear all
        </button>
      </div>
      <ul class="search-recent-list">
        {#each recentSearches as search}
          <li>
            <button 
              class="search-recent-item {isDarkMode ? 'search-recent-item-dark' : ''}"
              on:click={() => selectRecentSearch(search)}
            >
              <svg xmlns="http://www.w3.org/2000/svg" class="search-recent-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              {search}
            </button>
          </li>
        {/each}
      </ul>
    </div>
  {/if}
  
  <!-- Recommended profiles dropdown -->
  {#if searchQuery && recommendedProfiles.length > 0 && !isSearching}
    <div class="search-dropdown {isDarkMode ? 'search-dropdown-dark' : ''}">
      <div class="search-dropdown-header">
        <h3 class="search-dropdown-title">Suggested Profiles</h3>
      </div>
      <ul class="search-profiles-list">
        {#each recommendedProfiles as profile}
          <li>
            <a 
              href={`/user/${profile.username}`}
              class="search-profile-item {isDarkMode ? 'search-profile-item-dark' : ''}"
            >
              <div class="search-profile-content">
                <div class="search-profile-avatar {isDarkMode ? 'search-profile-avatar-dark' : ''}">
                  {#if typeof profile.avatar === 'string' && profile.avatar.startsWith('http')}
                    <img src={profile.avatar} alt={profile.username} class="search-profile-img" />
                  {:else}
                    <span class="search-profile-placeholder">{profile.displayName.charAt(0)}</span>
                  {/if}
                </div>
                <div class="search-profile-info">
                  <div class="search-profile-name-wrapper">
                    <p class="search-profile-name {isDarkMode ? 'search-profile-name-dark' : ''}">{profile.displayName}</p>
                    {#if profile.isVerified}
                      <span class="search-profile-verified">
                        <svg xmlns="http://www.w3.org/2000/svg" class="search-verified-icon" viewBox="0 0 20 20" fill="currentColor">
                          <path fill-rule="evenodd" d="M6.267 3.455a3.066 3.066 0 001.745-.723 3.066 3.066 0 013.976 0 3.066 3.066 0 001.745.723 3.066 3.066 0 012.812 2.812c.051.643.304 1.254.723 1.745a3.066 3.066 0 010 3.976 3.066 3.066 0 00-.723 1.745 3.066 3.066 0 01-2.812 2.812 3.066 3.066 0 00-1.745.723 3.066 3.066 0 01-3.976 0 3.066 3.066 0 00-1.745-.723 3.066 3.066 0 01-2.812-2.812 3.066 3.066 0 00-.723-1.745 3.066 3.066 0 010-3.976 3.066 3.066 0 00.723-1.745 3.066 3.066 0 012.812-2.812zm7.44 5.252a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                        </svg>
                      </span>
                    {/if}
                    
                    <!-- Fuzzy match indicator -->
                    {#if profile.fuzzyMatchScore && profile.fuzzyMatchScore > 0}
                      <span class="fuzzy-match-indicator {getFuzzyMatchColor(profile.fuzzyMatchScore)}" title="{getFuzzyMatchLabel(profile.fuzzyMatchScore)} ({Math.round(profile.fuzzyMatchScore * 100)}% similarity)">
                        <svg xmlns="http://www.w3.org/2000/svg" class="fuzzy-match-icon" viewBox="0 0 20 20" fill="currentColor">
                          <path fill-rule="evenodd" d="M8 4a4 4 0 100 8 4 4 0 000-8zM2 8a6 6 0 1110.89 3.476l4.817 4.817a1 1 0 01-1.414 1.414l-4.816-4.816A6 6 0 012 8z" clip-rule="evenodd" />
                        </svg>
                        {Math.round(profile.fuzzyMatchScore * 100)}%
                      </span>
                    {/if}
                  </div>
                  <p class="search-profile-username {isDarkMode ? 'search-profile-username-dark' : ''}">@{profile.username}</p>
                </div>
              </div>
            </a>
          </li>
        {/each}
        <li class="search-dropdown-footer">
          <button 
            class="search-query-button"
            on:click={executeSearch}
          >
            Search for "{searchQuery}"
          </button>
        </li>
      </ul>
    </div>
  {/if}
</div>

<style>
  .search-container {
    position: relative;
    width: 100%;
  }
  
  .search-input-wrapper {
    position: relative;
    background-color: var(--bg-tertiary);
    border-radius: var(--radius-full);
    transition: all 0.2s ease;
    border: 1px solid transparent;
    display: flex;
    align-items: center;
    height: 42px;
  }
  
  .search-input-wrapper-dark {
    background-color: var(--dark-bg-tertiary);
    border: 1px solid var(--dark-bg-tertiary);
  }
  
  .search-input-wrapper:focus-within {
    border-color: var(--color-primary);
    background-color: var(--bg-primary);
  }
  
  .search-input-wrapper-dark:focus-within {
    background-color: var(--dark-bg-primary);
  }
  
  .search-icon-container {
    display: flex;
    align-items: center;
    justify-content: center;
    padding-left: 12px;
  }
  
  .search-icon {
    width: 18px;
    height: 18px;
    color: var(--text-secondary);
  }
  
  .search-input {
    width: 100%;
    padding: 8px 8px 8px 8px;
    border-radius: var(--radius-full);
    border: none;
    background-color: transparent;
    color: var(--text-primary);
    font-size: var(--font-size-md);
    outline: none;
    caret-color: var(--color-primary);
  }
  
  .search-input-dark {
    color: var(--dark-text-primary);
  }
  
  .search-input::placeholder {
    color: var(--text-tertiary);
    opacity: 0.8;
  }
  
  .search-input-dark::placeholder {
    color: var(--dark-text-tertiary);
  }
  
  .search-clear-button {
    height: 22px;
    width: 22px;
    display: flex;
    align-items: center;
    justify-content: center;
    background: transparent;
    border: none;
    cursor: pointer;
    margin-right: 12px;
    padding: 0;
  }
  
  .search-clear-icon {
    width: 18px;
    height: 18px;
    color: var(--text-primary);
  }
  
  .search-dropdown {
    position: absolute;
    top: calc(100% + 4px);
    left: 0;
    right: 0;
    background-color: var(--bg-primary);
    border-radius: 14px;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1), 0 3px 10px rgba(0, 0, 0, 0.1);
    border: 1px solid var(--border-color);
    z-index: var(--z-dropdown);
    max-height: 500px;
    overflow-y: auto;
  }
  
  .search-dropdown-dark {
    background-color: var(--dark-bg-primary);
    border-color: var(--dark-border-color);
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.3), 0 3px 10px rgba(0, 0, 0, 0.3);
  }
  
  .search-dropdown-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 12px 16px;
    border-bottom: 1px solid var(--border-color);
  }
  
  .search-dropdown-dark .search-dropdown-header {
    border-color: var(--dark-border-color);
  }
  
  .search-dropdown-title {
    font-size: var(--font-size-md);
    font-weight: var(--font-weight-bold);
    margin: 0;
    color: var(--text-primary);
  }
  
  .search-dropdown-clear-button {
    background: none;
    border: none;
    color: var(--color-primary);
    font-size: var(--font-size-sm);
    font-weight: var(--font-weight-medium);
    cursor: pointer;
    padding: 4px 8px;
    border-radius: var(--radius-md);
    transition: background-color 0.2s;
  }
  
  .search-dropdown-clear-button:hover {
    background-color: var(--hover-primary);
  }
  
  .search-recent-list, 
  .search-profiles-list {
    list-style: none;
    padding: 0;
    margin: 0;
  }
  
  .search-recent-item {
    display: flex;
    align-items: center;
    width: 100%;
    padding: 12px 16px;
    text-align: left;
    background: none;
    border: none;
    cursor: pointer;
    font-size: var(--font-size-md);
    color: var(--text-primary);
    transition: background-color 0.2s;
  }
  
  .search-recent-item-dark {
    color: var(--dark-text-primary);
  }
  
  .search-recent-item:hover {
    background-color: var(--hover-bg);
  }
  
  .search-recent-item-dark:hover {
    background-color: var(--dark-hover-bg);
  }
  
  .search-recent-icon {
    width: 18px;
    height: 18px;
    margin-right: 12px;
    color: var(--text-secondary);
  }
  
  .search-profile-item {
    display: block;
    padding: 12px 16px;
    text-decoration: none;
    transition: background-color 0.2s;
  }
  
  .search-profile-item:hover {
    background-color: var(--hover-bg);
  }
  
  .search-profile-item-dark:hover {
    background-color: var(--dark-hover-bg);
  }
  
  .search-profile-content {
    display: flex;
    align-items: center;
  }
  
  .search-profile-avatar {
    width: 40px;
    height: 40px;
    border-radius: 50%;
    overflow: hidden;
    margin-right: 12px;
    background-color: var(--bg-tertiary);
    display: flex;
    align-items: center;
    justify-content: center;
  }
  
  .search-profile-avatar-dark {
    background-color: var(--dark-bg-tertiary);
  }
  
  .search-profile-img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }
  
  .search-profile-placeholder {
    color: var(--text-secondary);
    font-weight: var(--font-weight-bold);
    font-size: var(--font-size-lg);
  }
  
  .search-profile-info {
    flex: 1;
  }
  
  .search-profile-name-wrapper {
    display: flex;
    align-items: center;
  }
  
  .search-profile-name {
    margin: 0;
    font-weight: var(--font-weight-bold);
    color: var(--text-primary);
  }
  
  .search-profile-name-dark {
    color: var(--dark-text-primary);
  }
  
  .search-profile-verified {
    margin-left: 4px;
    display: flex;
  }
  
  .search-verified-icon {
    width: 16px;
    height: 16px;
    color: var(--color-primary);
  }
  
  .search-profile-username {
    margin: 2px 0 0;
    font-size: var(--font-size-sm);
    color: var(--text-secondary);
  }
  
  .search-profile-username-dark {
    color: var(--dark-text-secondary);
  }
  
  .search-dropdown-footer {
    padding: 12px 16px;
    border-top: 1px solid var(--border-color);
  }
  
  .search-dropdown-dark .search-dropdown-footer {
    border-color: var(--dark-border-color);
  }
  
  .search-query-button {
    width: 100%;
    padding: 8px 16px;
    background-color: var(--bg-primary);
    border: 1px solid var(--border-color);
    border-radius: var(--radius-full);
    color: var(--color-primary);
    font-weight: var(--font-weight-medium);
    cursor: pointer;
    transition: background-color 0.2s;
  }
  
  .search-query-button:hover {
    background-color: var(--hover-primary);
  }
  
  /* Fuzzy search styles */
  .fuzzy-search-indicator {
    position: absolute;
    right: 48px;
    top: 50%;
    transform: translateY(-50%);
    display: flex;
    align-items: center;
  }
  
  .fuzzy-search-badge {
    display: flex;
    align-items: center;
    font-size: var(--font-size-xs);
    color: var(--color-primary);
    background-color: rgba(var(--color-primary-rgb), 0.1);
    padding: 2px 6px;
    border-radius: var(--radius-full);
  }
  
  .fuzzy-search-icon {
    width: 12px;
    height: 12px;
    margin-right: 4px;
  }
  
  .fuzzy-match-indicator {
    display: flex;
    align-items: center;
    font-size: var(--font-size-xs);
    padding: 2px 6px;
    border-radius: var(--radius-full);
    margin-left: 6px;
    font-weight: 500;
  }
  
  .fuzzy-match-icon {
    width: 10px;
    height: 10px;
    margin-right: 2px;
  }
  
  .fuzzy-match-high {
    color: #16a34a;
    background-color: rgba(22, 163, 74, 0.1);
  }
  
  .fuzzy-match-medium {
    color: #ca8a04;
    background-color: rgba(202, 138, 4, 0.1);
  }
  
  .fuzzy-match-low {
    color: #dc2626;
    background-color: rgba(220, 38, 38, 0.1);
  }
</style> 