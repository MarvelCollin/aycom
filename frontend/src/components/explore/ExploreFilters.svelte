<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { useTheme } from '../../hooks/useTheme';
  import { createLoggerWithPrefix } from '../../utils/logger';
  
  const logger = createLoggerWithPrefix('ExploreFilters');
  const dispatch = createEventDispatcher();
  const { theme } = useTheme();
  
  // Reactive declarations
  $: isDarkMode = $theme === 'dark';
  
  // Props
  export let searchFilter: 'all' | 'following' | 'verified' = 'all';
  export let selectedCategory = 'all';
  export let threadCategories = [
    { id: 'all', name: 'All Categories' },
    { id: 'news', name: 'News' },
    { id: 'sports', name: 'Sports' },
    { id: 'entertainment', name: 'Entertainment' },
    { id: 'politics', name: 'Politics' },
    { id: 'tech', name: 'Technology' },
    { id: 'science', name: 'Science' },
    { id: 'health', name: 'Health' }
  ];
  
  // Handle filter change
  function handleFilterChange(filter: 'all' | 'following' | 'verified') {
    logger.debug('Search filter changed', { from: searchFilter, to: filter });
    dispatch('filterChange', filter);
  }
  
  // Handle category change
  function handleCategoryChange(category: string) {
    logger.debug('Category filter changed', { from: selectedCategory, to: category });
    dispatch('categoryChange', category);
  }
</script>

<div class="mt-3 flex flex-wrap gap-2">
  <!-- People filter -->
  <div class="relative inline-block text-left group">
    <button class="px-3 py-1 border border-gray-300 dark:border-gray-700 rounded-full text-sm font-medium flex items-center {isDarkMode ? 'text-white' : 'text-black'} hover:bg-gray-100 dark:hover:bg-gray-800">
      {searchFilter === 'all' ? 'Everyone' : searchFilter === 'following' ? 'People you follow' : 'Verified accounts'}
      <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 ml-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
      </svg>
    </button>
    <div class="absolute left-0 mt-1 w-48 rounded-md shadow-lg bg-white dark:bg-gray-900 ring-1 ring-black ring-opacity-5 hidden group-hover:block z-20">
      <div class="py-1">
        <button
          class="block px-4 py-2 text-sm w-full text-left {searchFilter === 'all' ? 'bg-gray-100 dark:bg-gray-800' : ''}"
          on:click={() => handleFilterChange('all')}
        >
          Everyone
        </button>
        <button
          class="block px-4 py-2 text-sm w-full text-left {searchFilter === 'following' ? 'bg-gray-100 dark:bg-gray-800' : ''}"
          on:click={() => handleFilterChange('following')}
        >
          People you follow
        </button>
        <button
          class="block px-4 py-2 text-sm w-full text-left {searchFilter === 'verified' ? 'bg-gray-100 dark:bg-gray-800' : ''}"
          on:click={() => handleFilterChange('verified')}
        >
          Verified accounts
        </button>
      </div>
    </div>
  </div>
  
  <!-- Category filter -->
  <div class="relative inline-block text-left group">
    <button class="px-3 py-1 border border-gray-300 dark:border-gray-700 rounded-full text-sm font-medium flex items-center {isDarkMode ? 'text-white' : 'text-black'} hover:bg-gray-100 dark:hover:bg-gray-800">
      {threadCategories.find(cat => cat.id === selectedCategory)?.name || 'All Categories'}
      <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 ml-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
      </svg>
    </button>
    <div class="absolute left-0 mt-1 w-48 rounded-md shadow-lg bg-white dark:bg-gray-900 ring-1 ring-black ring-opacity-5 hidden group-hover:block z-20">
      <div class="py-1">
        {#each threadCategories as category}
          <button
            class="block px-4 py-2 text-sm w-full text-left {selectedCategory === category.id ? 'bg-gray-100 dark:bg-gray-800' : ''}"
            on:click={() => handleCategoryChange(category.id)}
          >
            {category.name}
          </button>
        {/each}
      </div>
    </div>
  </div>
</div>

<style>
  /* Fix for dropdown display */
  .group:hover .hidden.group-hover\:block {
    display: block;
  }
</style> 