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
  export let selectedCategory: string = 'all';
  export let threadCategories: Array<{ id: string; name: string }> = [];
  
  // Handle filter change
  function handleFilterChange(filter: 'all' | 'following' | 'verified') {
    logger.debug('Search filter changed', { from: searchFilter, to: filter });
    dispatch('filterChange', filter);
  }
  
  // Handle category change
  function handleCategoryChange(event) {
    logger.debug('Category filter changed', { from: selectedCategory, to: event.target.value });
    dispatch('categoryChange', event.target.value);
  }
</script>

<div class="pt-3 flex flex-wrap items-center gap-2">
  <!-- People filter buttons -->
  <div class="flex bg-gray-100 dark:bg-gray-800 rounded-full p-1 mr-2">
    <button 
      class={`px-3 py-1 text-sm rounded-full ${searchFilter === 'all' ? 'bg-white dark:bg-gray-900 shadow' : 'text-gray-600 dark:text-gray-300'}`}
      on:click={() => handleFilterChange('all')}
    >
      Everyone
    </button>
    <button 
      class={`px-3 py-1 text-sm rounded-full ${searchFilter === 'following' ? 'bg-white dark:bg-gray-900 shadow' : 'text-gray-600 dark:text-gray-300'}`}
      on:click={() => handleFilterChange('following')}
    >
      People you follow
    </button>
    <button 
      class={`px-3 py-1 text-sm rounded-full ${searchFilter === 'verified' ? 'bg-white dark:bg-gray-900 shadow' : 'text-gray-600 dark:text-gray-300'}`}
      on:click={() => handleFilterChange('verified')}
    >
      Verified only
    </button>
  </div>
  
  <!-- Category dropdown -->
  <div class="flex-1">
    <select 
      class="w-full md:w-auto bg-gray-100 dark:bg-gray-800 border-0 rounded-full text-sm px-3 py-2 focus:ring-blue-500"
      value={selectedCategory}
      on:change={handleCategoryChange}
    >
      {#each threadCategories as category}
        <option value={category.id}>{category.name}</option>
      {/each}
    </select>
  </div>
</div>

<style>
  /* Fix for dropdown display */
  .group:hover .hidden.group-hover\:block {
    display: block;
  }
</style> 