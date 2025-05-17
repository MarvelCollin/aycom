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

<div class="filter-container">
  <!-- People filter buttons -->
  <div class="filter-button-group {isDarkMode ? 'filter-button-group-dark' : ''}">
    <button 
      class="filter-button {searchFilter === 'all' ? 'active' : ''} {isDarkMode ? 'filter-button-dark' : ''}"
      on:click={() => handleFilterChange('all')}
    >
      Everyone
    </button>
    <button 
      class="filter-button {searchFilter === 'following' ? 'active' : ''} {isDarkMode ? 'filter-button-dark' : ''}"
      on:click={() => handleFilterChange('following')}
    >
      People you follow
    </button>
    <button 
      class="filter-button {searchFilter === 'verified' ? 'active' : ''} {isDarkMode ? 'filter-button-dark' : ''}"
      on:click={() => handleFilterChange('verified')}
    >
      Verified only
    </button>
  </div>
  
  <!-- Category dropdown -->
  <div class="category-container">
    <select 
      class="category-select {isDarkMode ? 'category-select-dark' : ''}"
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
  .filter-container {
    padding-top: var(--space-3);
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: var(--space-2);
    margin-bottom: var(--space-3);
  }
  
  .filter-button-group {
    display: flex;
    background-color: var(--bg-tertiary);
    border-radius: var(--radius-full);
    padding: var(--space-1);
    margin-right: var(--space-2);
  }
  
  .filter-button-group-dark {
    background-color: var(--bg-tertiary-dark);
  }
  
  .filter-button {
    padding: var(--space-1) var(--space-3);
    font-size: var(--font-size-sm);
    border-radius: var(--radius-full);
    border: none;
    background: none;
    color: var(--text-secondary);
    cursor: pointer;
    transition: all var(--transition-fast);
  }
  
  .filter-button-dark {
    color: var(--text-secondary-dark);
  }
  
  .filter-button:hover {
    color: var(--text-primary);
  }
  
  .filter-button-dark:hover {
    color: var(--text-primary-dark);
  }
  
  .filter-button.active {
    background-color: var(--bg-primary);
    color: var(--text-primary);
    box-shadow: var(--shadow-sm);
  }
  
  .filter-button-dark.active {
    background-color: var(--bg-primary-dark);
    color: var(--text-primary-dark);
    box-shadow: var(--shadow-sm-dark);
  }
  
  .category-container {
    flex: 1;
  }
  
  .category-select {
    width: 100%;
    background-color: var(--bg-tertiary);
    border: none;
    border-radius: var(--radius-full);
    font-size: var(--font-size-sm);
    padding: var(--space-2) var(--space-3);
    color: var(--text-primary);
    appearance: none;
    background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='16' height='16' viewBox='0 0 24 24' fill='none' stroke='currentColor' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpolyline points='6 9 12 15 18 9'%3E%3C/polyline%3E%3C/svg%3E");
    background-repeat: no-repeat;
    background-position: right var(--space-2) center;
    background-size: 16px;
    cursor: pointer;
  }
  
  .category-select-dark {
    background-color: var(--bg-tertiary-dark);
    color: var(--text-primary-dark);
  }
  
  .category-select:focus {
    outline: none;
    border-color: var(--color-primary);
  }
  
  @media (min-width: 768px) {
    .category-select {
      width: auto;
    }
  }
</style> 