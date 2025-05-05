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
  }> = [];
  export let showRecentSearches = false;
  export let isSearching = false;
  export let isLoadingRecommendations = false;
  
  // Handle search input
  function handleSearchInput(event) {
    const value = event.target.value;
    logger.debug('Search input changed', { value });
    dispatch('input', event);
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
      logger.debug('Search triggered via Enter key', { query: searchQuery });
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
  
  // Log when recommendations are loaded
  $: {
    if (!isLoadingRecommendations && recommendedProfiles.length > 0) {
      logger.debug('Profile recommendations loaded', { count: recommendedProfiles.length });
    }
  }
</script>

<div class="relative">
  <!-- Search bar -->
  <div class="relative">
    <input 
      type="text" 
      placeholder="Search AYCOM" 
      value={searchQuery}
      on:input={handleSearchInput}
      on:focus={handleFocus}
      on:keydown={handleKeydown}
      class="w-full py-2 pl-12 pr-4 rounded-full {isDarkMode ? 'bg-gray-800 border-gray-700 text-white' : 'bg-gray-100 border-gray-200 text-black'} border focus:outline-none focus:ring-2 focus:ring-blue-500"
    />
    <button 
      class="absolute left-4 top-1/2 transform -translate-y-1/2 text-gray-500 dark:text-gray-400"
      on:click={executeSearch}
    >
      <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
      </svg>
    </button>
  </div>
  
  <!-- Recent searches dropdown -->
  {#if showRecentSearches && recentSearches.length > 0 && !searchQuery}
    <div class="absolute top-full left-0 right-0 mt-1 bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-700 rounded-lg shadow-lg z-20">
      <div class="flex justify-between items-center px-4 py-2 border-b border-gray-200 dark:border-gray-700">
        <h3 class="font-medium">Recent searches</h3>
        <button 
          class="text-blue-500 text-sm hover:text-blue-600"
          on:click={clearRecentSearches}
        >
          Clear all
        </button>
      </div>
      <ul>
        {#each recentSearches as search}
          <li>
            <button 
              class="w-full px-4 py-3 text-left hover:bg-gray-100 dark:hover:bg-gray-800 flex items-center"
              on:click={() => selectRecentSearch(search)}
            >
              <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-3 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
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
    <div class="absolute top-full left-0 right-0 mt-1 bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-700 rounded-lg shadow-lg z-20">
      <ul>
        {#each recommendedProfiles as profile}
          <li>
            <a 
              href={`/profile/${profile.username}`}
              class="block px-4 py-3 hover:bg-gray-100 dark:hover:bg-gray-800"
            >
              <div class="flex items-center">
                <div class="w-10 h-10 rounded-full {isDarkMode ? 'bg-gray-700' : 'bg-gray-200'} flex items-center justify-center overflow-hidden mr-3">
                  {#if typeof profile.avatar === 'string' && profile.avatar.startsWith('http')}
                    <img src={profile.avatar} alt={profile.username} class="w-full h-full object-cover" />
                  {:else}
                    <span class="text-lg">{profile.avatar}</span>
                  {/if}
                </div>
                <div>
                  <div class="flex items-center">
                    <p class="font-bold {isDarkMode ? 'text-white' : 'text-black'}">{profile.displayName}</p>
                    {#if profile.isVerified}
                      <span class="ml-1 text-blue-500">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor">
                          <path fill-rule="evenodd" d="M6.267 3.455a3.066 3.066 0 001.745-.723 3.066 3.066 0 013.976 0 3.066 3.066 0 001.745.723 3.066 3.066 0 012.812 2.812c.051.643.304 1.254.723 1.745a3.066 3.066 0 010 3.976 3.066 3.066 0 00-.723 1.745 3.066 3.066 0 01-2.812 2.812 3.066 3.066 0 00-1.745.723 3.066 3.066 0 01-3.976 0 3.066 3.066 0 00-1.745-.723 3.066 3.066 0 01-2.812-2.812 3.066 3.066 0 00-.723-1.745 3.066 3.066 0 010-3.976 3.066 3.066 0 00.723-1.745 3.066 3.066 0 012.812-2.812zm7.44 5.252a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                        </svg>
                      </span>
                    {/if}
                  </div>
                  <p class="{isDarkMode ? 'text-gray-400' : 'text-gray-500'} text-sm">@{profile.username}</p>
                </div>
              </div>
            </a>
          </li>
        {/each}
        <li class="border-t border-gray-200 dark:border-gray-700">
          <button 
            class="w-full px-4 py-3 text-blue-500 text-center hover:bg-gray-100 dark:hover:bg-gray-800"
            on:click={executeSearch}
          >
            Search for "{searchQuery}"
          </button>
        </li>
      </ul>
    </div>
  {/if}
</div> 