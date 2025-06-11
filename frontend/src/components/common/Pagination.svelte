<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { useTheme } from '../../hooks/useTheme';
  
  const dispatch = createEventDispatcher();
  const { theme } = useTheme();
  
  // Reactive declarations
  $: isDarkMode = $theme === 'dark';
  
  // Props
  export let totalItems: number = 0;
  export let perPage: number = 20;
  export let currentPage: number = 1;
  export let maxDisplayPages: number = 5;
  
  // Calculate total pages
  $: totalPages = Math.max(1, Math.ceil(totalItems / perPage));
  
  // Visible page range
  $: {
    let start = Math.max(1, currentPage - Math.floor(maxDisplayPages / 2));
    let end = start + maxDisplayPages - 1;
    
    if (end > totalPages) {
      end = totalPages;
      start = Math.max(1, end - maxDisplayPages + 1);
    }
    
    visiblePages = Array.from({ length: end - start + 1 }, (_, i) => start + i);
  }
  
  let visiblePages: number[] = [];
  
  function goToPage(page: number) {
    if (page < 1 || page > totalPages || page === currentPage) return;
    dispatch('pageChange', page);
  }
  
  function showEllipsisBefore() {
    return visiblePages.length > 0 && visiblePages[0] > 1;
  }
  
  function showEllipsisAfter() {
    return visiblePages.length > 0 && visiblePages[visiblePages.length - 1] < totalPages;
  }
</script>

{#if totalPages > 1}
  <nav class="pagination {isDarkMode ? 'pagination-dark' : ''}">
    <button 
      class="pagination-action" 
      on:click={() => goToPage(currentPage - 1)} 
      disabled={currentPage === 1}
      aria-label="Previous page"
    >
      <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <polyline points="15 18 9 12 15 6"></polyline>
      </svg>
    </button>
    
    {#if showEllipsisBefore()}
      <button class="pagination-button" on:click={() => goToPage(1)}>1</button>
      {#if visiblePages[0] > 2}
        <span class="pagination-ellipsis">...</span>
      {/if}
    {/if}
    
    {#each visiblePages as page}
      <button 
        class="pagination-button {page === currentPage ? 'active' : ''}" 
        on:click={() => goToPage(page)}
      >
        {page}
      </button>
    {/each}
    
    {#if showEllipsisAfter()}
      {#if visiblePages[visiblePages.length - 1] < totalPages - 1}
        <span class="pagination-ellipsis">...</span>
      {/if}
      <button class="pagination-button" on:click={() => goToPage(totalPages)}>{totalPages}</button>
    {/if}
    
    <button 
      class="pagination-action" 
      on:click={() => goToPage(currentPage + 1)} 
      disabled={currentPage === totalPages}
      aria-label="Next page"
    >
      <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <polyline points="9 18 15 12 9 6"></polyline>
      </svg>
    </button>
  </nav>
{/if}

<style>
  .pagination {
    display: flex;
    align-items: center;
    gap: 4px;
  }
  
  .pagination-dark {
    color: var(--dark-text-secondary);
  }
  
  .pagination-button,
  .pagination-action {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    min-width: 32px;
    height: 32px;
    border-radius: 9999px;
    padding: 0 8px;
    background: none;
    border: 1px solid var(--border-color);
    color: var(--text-primary);
    font-size: 14px;
    cursor: pointer;
    transition: background-color 0.2s;
  }
  
  .pagination-button:hover,
  .pagination-action:hover:not([disabled]) {
    background-color: var(--hover-bg);
  }
  
  .pagination-button.active {
    background-color: var(--color-primary);
    color: white;
    border-color: var(--color-primary);
    font-weight: bold;
  }
  
  .pagination-action[disabled] {
    cursor: not-allowed;
    opacity: 0.5;
  }
  
  .pagination-ellipsis {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    min-width: 32px;
    color: var(--text-secondary);
  }
  
  .pagination-dark .pagination-button,
  .pagination-dark .pagination-action {
    border-color: var(--dark-border-color);
    color: var(--dark-text-primary);
  }
  
  .pagination-dark .pagination-button:hover,
  .pagination-dark .pagination-action:hover:not([disabled]) {
    background-color: var(--dark-hover-bg);
  }
  
  .pagination-dark .pagination-ellipsis {
    color: var(--dark-text-secondary);
  }
  
  @media (max-width: 640px) {
    .pagination-button,
    .pagination-action {
      min-width: 28px;
      height: 28px;
      padding: 0 6px;
      font-size: 13px;
    }
    
    .pagination {
      gap: 2px;
    }
  }
</style> 