<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import ProfileCard from './ProfileCard.svelte';
  import { useTheme } from '../../hooks/useTheme';
  
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
  
  // Handle follow user
  function handleFollow(event) {
    dispatch('follow', event.detail);
  }
  
  // Handle people per page change
  function handlePeoplePerPageChange(perPage) {
    dispatch('peoplePerPageChange', perPage);
  }
  
  // Handle load more
  function handleLoadMore() {
    dispatch('loadMore');
  }
</script>

<div class="p-4">
  <!-- Pagination options -->
  <div class="mb-4 flex justify-end">
    <div class="relative inline-block text-left group">
      <button class="px-3 py-1 border border-gray-300 dark:border-gray-700 rounded-full text-sm font-medium flex items-center {isDarkMode ? 'text-white' : 'text-black'} hover:bg-gray-100 dark:hover:bg-gray-800">
        Show {peoplePerPage} per page
        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 ml-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
        </svg>
      </button>
      <div class="absolute right-0 mt-1 w-36 rounded-md shadow-lg bg-white dark:bg-gray-900 ring-1 ring-black ring-opacity-5 hidden group-hover:block z-20">
        <div class="py-1">
          <button
            class="block px-4 py-2 text-sm w-full text-left {peoplePerPage === 25 ? 'bg-gray-100 dark:bg-gray-800' : ''}"
            on:click={() => handlePeoplePerPageChange(25)}
          >
            Show 25 per page
          </button>
          <button
            class="block px-4 py-2 text-sm w-full text-left {peoplePerPage === 30 ? 'bg-gray-100 dark:bg-gray-800' : ''}"
            on:click={() => handlePeoplePerPageChange(30)}
          >
            Show 30 per page
          </button>
          <button
            class="block px-4 py-2 text-sm w-full text-left {peoplePerPage === 35 ? 'bg-gray-100 dark:bg-gray-800' : ''}"
            on:click={() => handlePeoplePerPageChange(35)}
          >
            Show 35 per page
          </button>
        </div>
      </div>
    </div>
  </div>
  
  {#if isLoading}
    <div class="animate-pulse space-y-4">
      {#each Array(5) as _}
        <div class="flex space-x-4">
          <div class="rounded-full bg-gray-300 dark:bg-gray-700 h-12 w-12"></div>
          <div class="flex-1 space-y-2 py-1">
            <div class="h-4 bg-gray-300 dark:bg-gray-700 rounded w-1/4"></div>
            <div class="h-4 bg-gray-300 dark:bg-gray-700 rounded w-1/2"></div>
          </div>
          <div class="w-20 h-8 bg-gray-300 dark:bg-gray-700 rounded-full"></div>
        </div>
      {/each}
    </div>
  {:else if peopleResults.length > 0}
    <ul class="divide-y divide-gray-200 dark:divide-gray-800">
      {#each peopleResults as profile}
        <li>
          <ProfileCard {profile} on:follow={handleFollow} />
        </li>
      {/each}
    </ul>
    
    <!-- Load more button -->
    {#if !isLoading}
      <div class="mt-4 text-center">
        <button 
          class="px-4 py-2 bg-gray-200 dark:bg-gray-800 rounded-full text-sm font-medium hover:bg-gray-300 dark:hover:bg-gray-700"
          on:click={handleLoadMore}
        >
          Load more
        </button>
      </div>
    {/if}
  {:else}
    <div class="text-center py-10">
      <p class="text-gray-500 dark:text-gray-400">No users found</p>
    </div>
  {/if}
</div>

<style>
  /* Skeleton loading animation */
  @keyframes pulse {
    0%, 100% { opacity: 0.5; }
    50% { opacity: 1; }
  }
  .animate-pulse {
    animation: pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite;
  }
  
  /* Fix for dropdown display */
  .group:hover .hidden.group-hover\:block {
    display: block;
  }
</style> 