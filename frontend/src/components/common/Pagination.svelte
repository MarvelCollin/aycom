<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { useTheme } from '../../hooks/useTheme';
  
  // Event dispatcher
  const dispatch = createEventDispatcher();
  
  // Props
  export let currentPage = 1;
  export let totalPages = 1;
  export let maxDisplayPages = 5;
  
  // Theme state
  const { theme } = useTheme();
  $: isDarkMode = $theme === 'dark';
  
  // Computed values
  $: pagesArray = generatePaginationArray(currentPage, totalPages, maxDisplayPages);
  $: hasPrevious = currentPage > 1;
  $: hasNext = currentPage < totalPages;
  
  // Generate the array of page numbers to display
  function generatePaginationArray(current: number, total: number, max: number): (number | string)[] {
    if (total <= max) {
      // Show all pages if total is less than max
      return Array.from({ length: total }, (_, i) => i + 1);
    }
    
    // Calculate half of max (rounded down)
    const half = Math.floor(max / 2);
    
    // Determine start and end
    let start = current - half;
    let end = current + half;
    
    // Adjust if out of bounds
    if (start < 1) {
      end = end + (1 - start);
      start = 1;
    }
    
    if (end > total) {
      start = Math.max(1, start - (end - total));
      end = total;
    }
    
    // Generate array
    const result: (number | string)[] = [];
    
    // Add first page and ellipsis if needed
    if (start > 1) {
      result.push(1);
      if (start > 2) result.push('...');
    }
    
    // Add page numbers
    for (let i = start; i <= end; i++) {
      result.push(i);
    }
    
    // Add ellipsis and last page if needed
    if (end < total) {
      if (end < total - 1) result.push('...');
      result.push(total);
    }
    
    return result;
  }
  
  // Handle page change
  function handlePageChange(page: number | string) {
    if (typeof page === 'number' && page !== currentPage) {
      currentPage = page;
      dispatch('pageChange', { page });
    }
  }
  
  // Go to previous page
  function goToPrevious() {
    if (hasPrevious) {
      handlePageChange(currentPage - 1);
    }
  }
  
  // Go to next page
  function goToNext() {
    if (hasNext) {
      handlePageChange(currentPage + 1);
    }
  }
</script>

<div class="pagination {isDarkMode ? 'dark' : ''}">
  <button 
    class="pagination-control {!hasPrevious ? 'disabled' : ''}"
    disabled={!hasPrevious}
    on:click={goToPrevious}
    aria-label="Previous page"
  >
    &lt;
  </button>
  
  {#each pagesArray as page}
    {#if page === "..."}
      <span class="pagination-ellipsis">...</span>
    {:else}
      <button 
        class="pagination-page {page === currentPage ? 'active' : ''}"
        on:click={() => handlePageChange(page)}
      >
        {page}
      </button>
    {/if}
  {/each}
  
  <button 
    class="pagination-control {!hasNext ? 'disabled' : ''}"
    disabled={!hasNext}
    on:click={goToNext}
    aria-label="Next page"
  >
    &gt;
  </button>
</div>

<style>
  .pagination {
    display: flex;
    align-items: center;
    gap: 0.25rem;
  }
  
  .pagination-control,
  .pagination-page {
    display: flex;
    align-items: center;
    justify-content: center;
    min-width: 2rem;
    height: 2rem;
    padding: 0 0.5rem;
    background-color: white;
    border: 1px solid #e2e8f0;
    border-radius: 0.25rem;
    font-size: 0.875rem;
    cursor: pointer;
    transition: all 0.2s ease;
  }
  
  .pagination.dark .pagination-control,
  .pagination.dark .pagination-page {
    background-color: #2d3748;
    border-color: #4a5568;
    color: white;
  }
  
  .pagination-control:hover,
  .pagination-page:hover {
    background-color: #edf2f7;
    border-color: #cbd5e0;
  }
  
  .pagination.dark .pagination-control:hover,
  .pagination.dark .pagination-page:hover {
    background-color: #4a5568;
    border-color: #718096;
  }
  
  .pagination-control.disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
  
  .pagination-control.disabled:hover {
    background-color: white;
    border-color: #e2e8f0;
  }
  
  .pagination.dark .pagination-control.disabled:hover {
    background-color: #2d3748;
    border-color: #4a5568;
  }
  
  .pagination-page.active {
    background-color: #3182ce;
    border-color: #3182ce;
    color: white;
  }
  
  .pagination.dark .pagination-page.active {
    background-color: #4299e1;
    border-color: #4299e1;
  }
  
  .pagination-ellipsis {
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 0 0.5rem;
    font-size: 0.875rem;
  }
</style> 