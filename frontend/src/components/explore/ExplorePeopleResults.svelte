<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import ProfileCard from './ProfileCard.svelte';
  import { useTheme } from '../../hooks/useTheme';
  import { createLoggerWithPrefix } from '../../utils/logger';
  
  const logger = createLoggerWithPrefix('ExplorePeopleResults');
  const dispatch = createEventDispatcher();
  const { theme } = useTheme();
  
  // Reactive declarations
  $: isDarkMode = $theme === 'dark';
  
  // Props
  export let peopleResults: Array<{
    id: string;
    username: string;
    displayName: string;
    avatar: string | null;
    bio?: string;
    isVerified: boolean;
    followerCount: number;
    isFollowing: boolean;
  }> = [];
  export let isLoading = false;
  export let peoplePerPage = 25;
  export let currentPage = 1;
  export let totalCount = 0;
  
  // Calculate total pages
  $: totalPages = Math.ceil(totalCount / peoplePerPage);
  
  // Handle page change
  function changePage(page: number) {
    logger.debug('Changing page', { page });
    dispatch('pageChange', page);
  }
  
  // Page size options
  const perPageOptions = [25, 30, 35];
  
  // Handle per page change
  function handlePerPageChange(e) {
    const newValue = parseInt(e.target.value);
    logger.debug('Changing results per page', { value: newValue });
    dispatch('peoplePerPageChange', newValue);
  }
  
  // Handle load more
  function loadMore() {
    if (currentPage < totalPages) {
      logger.debug('Loading more people', { newPage: currentPage + 1 });
      dispatch('loadMore');
    }
  }
  
  // Handle follow user
  function handleFollow(event) {
    const userId = event.detail;
    logger.debug('Follow request initiated', { userId });
    dispatch('follow', userId);
  }
  
  // Log when people results change
  $: {
    if (!isLoading) {
      if (peopleResults.length > 0) {
        logger.debug('People results loaded', { count: peopleResults.length });
      } else {
        logger.debug('No people results found');
      }
    }
  }
</script>

<div class="p-4">
  {#if isLoading}
    <div class="animate-pulse space-y-4">
      {#each Array(3) as _}
        <div class="bg-gray-200 dark:bg-gray-800 h-24 rounded-lg"></div>
      {/each}
    </div>
  {:else if peopleResults.length === 0}
    <div class="text-center py-8">
      <p class="text-gray-500 dark:text-gray-400">No people found matching your search.</p>
    </div>
  {:else}
    <div class="space-y-4">
      {#each peopleResults as person}
        <div class="profile-result-card">
          <ProfileCard profile={person} on:follow={handleFollow} />
        </div>
      {/each}
      
      <!-- Pagination controls -->
      {#if totalCount > peoplePerPage}
        <div class="mt-6 flex flex-wrap justify-between items-center gap-4 border-t border-gray-200 dark:border-gray-800 pt-4">
          <div class="flex items-center">
            <span class="text-sm text-gray-500 dark:text-gray-400 mr-2">Show:</span>
            <select 
              class="bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-700 rounded-md px-2 py-1 text-sm"
              bind:value={peoplePerPage}
              on:change={handlePerPageChange}
            >
              {#each perPageOptions as option}
                <option value={option}>{option} per page</option>
              {/each}
            </select>
          </div>
          
          <div class="flex items-center space-x-1">
            <button 
              class="px-3 py-1 rounded {currentPage === 1 ? 'text-gray-400 cursor-not-allowed' : 'text-blue-500 hover:bg-blue-50 dark:hover:bg-gray-800'}"
              disabled={currentPage === 1}
              on:click={() => changePage(currentPage - 1)}
            >
              Previous
            </button>
            
            {#each Array(Math.min(5, totalPages)) as _, i}
              {#if totalPages <= 5 || (i < 3 && currentPage <= 3) || (i >= totalPages - 3 && currentPage >= totalPages - 2) || (i >= currentPage - 2 && i <= currentPage)}
                <button 
                  class="w-8 h-8 rounded-full flex items-center justify-center {i + 1 === currentPage ? 'bg-blue-500 text-white' : 'hover:bg-gray-100 dark:hover:bg-gray-800 text-gray-700 dark:text-gray-300'}"
                  on:click={() => changePage(i + 1)}
                >
                  {i + 1}
                </button>
              {:else if (i === 3 && currentPage > 3) || (i === totalPages - 4 && currentPage < totalPages - 2)}
                <span class="px-1">...</span>
              {/if}
            {/each}
            
        <button 
              class="px-3 py-1 rounded {currentPage === totalPages ? 'text-gray-400 cursor-not-allowed' : 'text-blue-500 hover:bg-blue-50 dark:hover:bg-gray-800'}"
              disabled={currentPage === totalPages}
              on:click={() => changePage(currentPage + 1)}
        >
              Next
        </button>
          </div>
      </div>
    {/if}
    </div>
  {/if}
</div>

<style>
  .animate-pulse {
    animation: pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite;
  }
  
  @keyframes pulse {
    0%, 100% { opacity: 0.5; }
    50% { opacity: 1; }
  }

  .profile-result-card {
    background-color: var(--bg-secondary, #f8f9fa);
    border-radius: 0.75rem;
    overflow: hidden;
    transition: all 0.2s ease;
    border: 1px solid var(--border-color, #e5e7eb);
    padding: 0.25rem 0;
  }
  
  .profile-result-card:hover {
    transform: translateY(-2px);
    box-shadow: 0 4px 6px rgba(0, 0, 0, 0.05);
  }
  
  :global(.dark) .profile-result-card {
    background-color: var(--bg-secondary-dark, #1f2937);
    border-color: var(--border-color-dark, #374151);
  }
  
  :global(.dark) .profile-result-card:hover {
    box-shadow: 0 4px 6px rgba(0, 0, 0, 0.2);
  }
</style> 