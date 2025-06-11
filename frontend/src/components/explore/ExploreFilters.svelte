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
  <div class="filter-section">
    <label class="filter-label">Show:</label>
    <div class="filter-button-group {isDarkMode ? 'filter-button-group-dark' : ''}">
      <button 
        class="filter-button {searchFilter === 'all' ? 'active' : ''} {isDarkMode ? 'filter-button-dark' : ''}"
        on:click={() => handleFilterChange('all')}
      >
        <span class="filter-icon">👥</span>
        Everyone
      </button>
      <button 
        class="filter-button {searchFilter === 'following' ? 'active' : ''} {isDarkMode ? 'filter-button-dark' : ''}"
        on:click={() => handleFilterChange('following')}
      >
        <span class="filter-icon">👤</span>
        Following
      </button>
      <button 
        class="filter-button {searchFilter === 'verified' ? 'active' : ''} {isDarkMode ? 'filter-button-dark' : ''}"
        on:click={() => handleFilterChange('verified')}
      >
        <span class="filter-icon">✓</span>
        Verified
      </button>
    </div>
  </div>
  
  <!-- Category dropdown -->
  <div class="filter-section">
    <label class="filter-label">Category:</label>
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
      <span class="dropdown-icon">
        <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <polyline points="6 9 12 15 18 9"></polyline>
        </svg>
      </span>
    </div>
  </div>
</div>

<style>
  .filter-container {
    padding: var(--space-3) 0;
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: var(--space-4);
    margin-bottom: var(--space-3);
    border-bottom: 1px solid var(--border-color);
  }
  
  .filter-section {
    display: flex;
    align-items: center;
    gap: var(--space-2);
  }
  
  .filter-label {
    font-size: var(--font-size-sm);
    color: var(--text-secondary);
    font-weight: var(--font-weight-medium);
  }
  
  .filter-button-group {
    display: flex;
    background-color: var(--bg-tertiary);
    border-radius: var(--radius-full);
    padding: var(--space-1);
    box-shadow: var(--shadow-sm);
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
    display: flex;
    align-items: center;
    gap: var(--space-1);
  }
  
  .filter-icon {
    font-size: var(--font-size-sm);
    display: inline-flex;
    align-items: center;
    justify-content: center;
  }
  
  .filter-button-dark {
    color: var(--text-secondary-dark);
  }
  
  .filter-button:hover {
    color: var(--text-primary);
    transform: translateY(-1px);
  }
  
  .filter-button-dark:hover {
    color: var(--text-primary-dark);
  }
  
  .filter-button.active {
    background-color: var(--bg-primary);
    color: var(--color-primary);
    box-shadow: var(--shadow-sm);
    font-weight: var(--font-weight-medium);
  }
  
  .filter-button-dark.active {
    background-color: var(--bg-primary-dark);
    color: var(--color-primary);
    box-shadow: var(--shadow-sm-dark);
  }
  
  .category-container {
    position: relative;
    flex: 1;
  }
  
  .dropdown-icon {
    position: absolute;
    right: var(--space-3);
    top: 50%;
    transform: translateY(-50%);
    pointer-events: none;
    color: var(--text-secondary);
  }
  
  .category-select {
    width: 100%;
    background-color: var(--bg-tertiary);
    border: 1px solid transparent;
    border-radius: var(--radius-full);
    font-size: var(--font-size-sm);
    padding: var(--space-2) var(--space-3);
    padding-right: var(--space-8);
    color: var(--text-primary);
    appearance: none;
    cursor: pointer;
    transition: all var(--transition-fast);
    box-shadow: var(--shadow-sm);
  }
  
  .category-select-dark {
    background-color: var(--bg-tertiary-dark);
    color: var(--text-primary-dark);
  }
  
  /* Dropdown styling for light and dark mode */
  .category-select option {
    background-color: var(--bg-primary);
    color: var(--text-primary);
    padding: var(--space-2);
  }
  
  .category-select-dark option {
    background-color: var(--bg-primary-dark, #1a1a1a);
    color: var(--text-primary-dark, #ffffff);
  }
  
  /* Override browser defaults */
  select::-ms-expand {
    display: none;
  }
  
  /* For Firefox */
  select {
    -moz-appearance: none;
  }
  
  /* For Chrome and Safari */
  select::-webkit-dropdown-button {
    display: none;
  }
  
  /* For a consistent dropdown appearance */
  @supports (-moz-appearance:none) {
    .category-select option {
      padding: var(--space-3);
    }
  }
  
  .category-select:focus {
    outline: none;
    border-color: var(--color-primary);
    transform: translateY(-1px);
    box-shadow: var(--shadow-md);
  }
  
  .category-select:hover {
    border-color: var(--color-primary);
    transform: translateY(-1px);
  }
  
  @media (max-width: 600px) {
    .filter-container {
      flex-direction: column;
      align-items: flex-start;
      gap: var(--space-3);
    }
    
    .filter-section {
      width: 100%;
    }
    
    .category-select {
      width: 100%;
    }
  }
  
  @media (min-width: 768px) {
    .category-select {
      min-width: 150px;
    }
  }
</style> 