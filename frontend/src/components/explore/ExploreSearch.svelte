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
  }> = [];
  export let showRecentSearches = false;
  export let isSearching = false;
  export let isLoadingRecommendations = false;
  
  // Handle search input
  function handleSearchInput(event) {
    const value = event.target.value;
    logger.debug('Search input changed', { value });
    dispatch('input', value);
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
    if (event.key === 'Enter') {
      logger.debug('Search triggered via Enter key', { query: searchQuery || '(empty)' });
      dispatch('search');
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
    <input 
      type="text" 
      placeholder="Search AYCOM" 
      value={searchQuery}
      on:input={handleSearchInput}
      on:focus={handleFocus}
      on:keydown={handleKeydown}
      class="search-input {isDarkMode ? 'search-input-dark' : ''}"
    />
    <button 
      class="search-icon-button {isDarkMode ? 'search-icon-button-dark' : ''}"
      on:click={executeSearch}
      aria-label="Search"
    >
      <svg xmlns="http://www.w3.org/2000/svg" class="search-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
      </svg>
    </button>
    
    {#if searchQuery}
      <button 
        class="search-clear-button {isDarkMode ? 'search-clear-button-dark' : ''}"
        on:click={clearSearch}
        aria-label="Clear search"
      >
        <svg xmlns="http://www.w3.org/2000/svg" class="search-clear-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
        </svg>
      </button>
    {/if}
  </div>
  
  <!-- Recent searches dropdown -->
  {#if showRecentSearches && recentSearches.length > 0 && !searchQuery}
    <div class="search-dropdown {isDarkMode ? 'search-dropdown-dark' : ''}">
      <div class="search-dropdown-header">
        <h3 class="search-dropdown-title">Recent searches</h3>
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
    margin-bottom: var(--space-2);
    background-color: var(--bg-tertiary);
    border-radius: var(--radius-full);
    box-shadow: var(--shadow-sm);
    transition: all var(--transition-normal);
    border: 2px solid transparent;
    display: flex;
    align-items: center;
  }
  
  .search-input-wrapper-dark {
    background-color: var(--dark-bg-tertiary, rgba(255, 255, 255, 0.1));
    border-color: var(--border-color-dark);
  }
  
  .search-input-wrapper:focus-within {
    border-color: var(--color-primary);
    box-shadow: var(--shadow-md), 0 0 0 2px rgba(var(--color-primary-rgb), 0.2);
  }
  
  .search-input {
    width: 100%;
    padding: var(--space-3) var(--space-12) var(--space-3) var(--space-12);
    border-radius: var(--radius-full);
    border: none;
    background-color: transparent;
    color: var(--text-primary);
    font-size: var(--font-size-md);
    transition: all var(--transition-normal);
    outline: none;
  }
  
  .search-input-dark {
    color: var(--dark-text-primary, #fff);
  }
  
  .search-input::placeholder {
    color: var(--text-tertiary);
    opacity: 0.8;
  }
  
  .search-input-dark::placeholder {
    color: var(--dark-text-tertiary, rgba(255, 255, 255, 0.5));
  }
  
  .search-icon-button {
    position: absolute;
    left: var(--space-3);
    top: 50%;
    transform: translateY(-50%);
    padding: var(--space-1);
    background: none;
    border: none;
    color: var(--text-secondary);
    cursor: pointer;
    transition: color var(--transition-fast);
    display: flex;
    align-items: center;
    justify-content: center;
  }
  
  .search-icon-button-dark {
    color: var(--dark-text-secondary, rgba(255, 255, 255, 0.6));
  }
  
  .search-icon-button:hover,
  .search-input:focus + .search-icon-button {
    color: var(--color-primary);
  }
  
  .search-icon {
    width: 20px;
    height: 20px;
  }
  
  .search-dropdown {
    position: absolute;
    top: 100%;
    left: 0;
    right: 0;
    background-color: var(--bg-primary);
    border-radius: var(--radius-lg);
    box-shadow: var(--shadow-lg);
    border: 1px solid var(--border-color);
    margin-top: var(--space-2);
    z-index: var(--z-dropdown);
    max-height: 500px;
    overflow-y: auto;
    animation: fadeInDown 0.3s ease-out;
  }
  
  @keyframes fadeInDown {
    from {
      opacity: 0;
      transform: translateY(-10px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }
  
  .search-dropdown-dark {
    background-color: var(--dark-bg-primary);
    border-color: var(--border-color-dark);
    box-shadow: 0 8px 30px rgba(0, 0, 0, 0.5);
  }
  
  .search-dropdown-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: var(--space-3) var(--space-4);
    border-bottom: 1px solid var(--border-color);
  }
  
  .search-dropdown-title {
    font-weight: var(--font-weight-bold);
    color: var(--text-primary);
    font-size: var(--font-size-md);
    margin: 0;
  }
  
  .search-dropdown-clear-button {
    background: none;
    border: none;
    color: var(--color-primary);
    font-size: var(--font-size-sm);
    cursor: pointer;
    padding: var(--space-1) var(--space-2);
    border-radius: var(--radius-md);
    transition: all var(--transition-fast);
  }
  
  .search-dropdown-clear-button:hover {
    background-color: rgba(var(--color-primary-rgb), 0.1);
    text-decoration: underline;
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
    padding: var(--space-3) var(--space-4);
    width: 100%;
    text-align: left;
    background: none;
    border: none;
    color: var(--text-primary);
    font-size: var(--font-size-md);
    cursor: pointer;
    transition: background-color var(--transition-fast);
    border-left: 3px solid transparent;
  }
  
  .search-recent-item-dark {
    color: var(--text-primary-dark);
  }
  
  .search-recent-item:hover {
    background-color: var(--hover-bg);
    border-left-color: var(--color-primary);
  }
  
  .search-recent-icon {
    width: 18px;
    height: 18px;
    margin-right: var(--space-3);
    color: var(--text-tertiary);
    flex-shrink: 0;
  }
  
  .search-profile-item {
    display: flex;
    padding: var(--space-3) var(--space-4);
    text-decoration: none;
    transition: background-color var(--transition-fast);
    border-left: 3px solid transparent;
  }
  
  .search-profile-item:hover {
    background-color: var(--hover-bg);
    border-left-color: var(--color-primary);
  }
  
  .search-profile-item-dark:hover {
    background-color: var(--bg-hover-dark);
  }
  
  .search-profile-content {
    display: flex;
    align-items: center;
    width: 100%;
  }
  
  .search-profile-avatar {
    width: 40px;
    height: 40px;
    border-radius: 50%;
    margin-right: var(--space-3);
    overflow: hidden;
    background-color: var(--bg-tertiary);
    flex-shrink: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    box-shadow: var(--shadow-sm);
    border: 1px solid var(--border-color);
  }
  
  .search-profile-avatar-dark {
    background-color: var(--bg-tertiary-dark);
    border-color: var(--border-color-dark);
  }
  
  .search-profile-img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }
  
  .search-profile-placeholder {
    font-size: var(--font-size-lg);
    color: var(--text-tertiary);
    text-transform: uppercase;
  }
  
  .search-profile-info {
    flex: 1;
    min-width: 0;
  }
  
  .search-profile-name-wrapper {
    display: flex;
    align-items: center;
  }
  
  .search-profile-name {
    font-weight: var(--font-weight-bold);
    color: var(--text-primary);
    margin: 0;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  
  .search-profile-name-dark {
    color: var(--text-primary-dark);
  }
  
  .search-profile-verified {
    margin-left: var(--space-1);
    color: var(--color-primary);
    display: inline-flex;
  }
  
  .search-verified-icon {
    width: 16px;
    height: 16px;
  }
  
  .search-profile-username {
    color: var(--text-secondary);
    font-size: var(--font-size-sm);
    margin: var(--space-1) 0 0;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  
  .search-profile-username-dark {
    color: var(--text-secondary-dark);
  }
  
  .search-dropdown-footer {
    padding: var(--space-3);
    border-top: 1px solid var(--border-color);
    text-align: center;
  }
  
  .search-query-button {
    background-color: var(--color-primary);
    color: white;
    border: none;
    border-radius: var(--radius-full);
    padding: var(--space-2) var(--space-4);
    font-weight: var(--font-weight-medium);
    cursor: pointer;
    transition: background-color var(--transition-fast), transform var(--transition-fast);
    width: 100%;
  }
  
  .search-query-button:hover {
    background-color: var(--color-primary-hover);
    transform: translateY(-1px);
  }
  
  .search-query-button:active {
    transform: translateY(0);
  }
  
  @media (max-width: 500px) {
    .search-dropdown {
      position: fixed;
      top: 60px;
      left: 0;
      right: 0;
      border-radius: 0;
      max-height: calc(100vh - 60px);
      margin-top: 0;
      border-top: 1px solid var(--border-color);
      border-left: none;
      border-right: none;
      box-shadow: none;
    }
    
    .search-dropdown-dark {
      border-top: 1px solid var(--border-color-dark);
    }
    
    .search-input {
      font-size: var(--font-size-base);
      padding: var(--space-2) var(--space-10) var(--space-2) var(--space-10);
    }
    
    .search-icon {
      width: 18px;
      height: 18px;
    }
  }
  
  .search-clear-button {
    position: absolute;
    right: var(--space-3);
    top: 50%;
    transform: translateY(-50%);
    padding: var(--space-1);
    background: none;
    border: none;
    color: var(--text-secondary);
    cursor: pointer;
    transition: color var(--transition-fast);
    display: flex;
    align-items: center;
    justify-content: center;
    opacity: 0.6;
  }
  
  .search-clear-button:hover {
    opacity: 1;
    color: var(--color-error, #e53e3e);
  }
  
  .search-clear-button-dark {
    color: var(--dark-text-secondary, rgba(255, 255, 255, 0.6));
  }
  
  .search-clear-button-dark:hover {
    color: var(--dark-color-error, #fc8181);
  }
  
  .search-clear-icon {
    width: 16px;
    height: 16px;
  }
</style> 