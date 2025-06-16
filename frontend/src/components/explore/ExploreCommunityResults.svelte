<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { useTheme } from '../../hooks/useTheme';
  import CommunityCard from './CommunityCard.svelte';
  import { createLoggerWithPrefix } from '../../utils/logger';

  const logger = createLoggerWithPrefix('ExploreCommunityResults');
  const dispatch = createEventDispatcher();
  const { theme } = useTheme();

  $: isDarkMode = $theme === 'dark';

  type RawCommunity = {
    id: string;
    name: string;
    description?: string;
    logo?: string | null;
    logo_url?: string | null;
    avatar?: string | null;
    memberCount?: number;
    member_count?: number;
    isJoined?: boolean;
    is_joined?: boolean;
    isPending?: boolean;
    is_pending?: boolean;
  };

  export let communityResults: RawCommunity[] = [];
  export let isLoading = false;
  export let communitiesPerPage = 25;
  export let currentPage = 1;
  export let totalCount = 0;

  $: processedCommunities = communityResults
    .filter(community => community && typeof community === 'object')
    .map(community => {

      return {
        id: community.id || '',
        name: community.name || '',
        description: community.description || '',
        logo: community.logo || community.logo_url || community.avatar || null,
        memberCount: community.memberCount || community.member_count || 0,
        isJoined: community.isJoined || community.is_joined || false,
        isPending: community.isPending || community.is_pending || false
      };
    });

  $: totalPages = Math.ceil(totalCount / communitiesPerPage);

  function changePage(page: number) {
    logger.debug('Changing page', { page });
    dispatch('pageChange', page);
  }

  const perPageOptions = [25, 30, 35];

  function handlePerPageChange(e) {
    const newValue = parseInt(e.target.value);
    logger.debug('Changing results per page', { value: newValue });
    dispatch('communitiesPerPageChange', newValue);
  }

  function handleJoinRequest(event) {
    const { communityId } = event.detail;
    logger.debug('Join request for community', { communityId });
    dispatch('joinRequest', communityId);
  }

  function handleCommunitiesPerPageChange(perPage) {
    logger.debug('Changing communities per page', { from: communitiesPerPage, to: perPage });
    dispatch('communitiesPerPageChange', perPage);
  }

  function handleLoadMore() {
    logger.debug('Loading more community results');
    dispatch('loadMore');
  }

  $: {
    if (!isLoading) {
      if (processedCommunities.length > 0) {
        logger.debug('Community results loaded', { count: processedCommunities.length });
      } else {
        logger.debug('No community results found');
      }
    }
  }
</script>

<div class="p-4">
  <!-- Pagination options -->
  <div class="mb-4 flex justify-end">
    <div class="relative inline-block text-left group">
      <button class="px-3 py-1 border border-gray-300 dark:border-gray-700 rounded-full text-sm font-medium flex items-center {isDarkMode ? 'text-white' : 'text-black'} hover:bg-gray-100 dark:hover:bg-gray-800">
        Show {communitiesPerPage} per page
        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 ml-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
        </svg>
      </button>
      <div class="absolute right-0 mt-1 w-36 rounded-md shadow-lg bg-white dark:bg-gray-900 ring-1 ring-black ring-opacity-5 hidden group-hover:block z-20">
        <div class="py-1">
          <button
            class="block px-4 py-2 text-sm w-full text-left {communitiesPerPage === 25 ? 'bg-gray-100 dark:bg-gray-800' : ''}"
            on:click={() => handleCommunitiesPerPageChange(25)}
          >
            Show 25 per page
          </button>
          <button
            class="block px-4 py-2 text-sm w-full text-left {communitiesPerPage === 30 ? 'bg-gray-100 dark:bg-gray-800' : ''}"
            on:click={() => handleCommunitiesPerPageChange(30)}
          >
            Show 30 per page
          </button>
          <button
            class="block px-4 py-2 text-sm w-full text-left {communitiesPerPage === 35 ? 'bg-gray-100 dark:bg-gray-800' : ''}"
            on:click={() => handleCommunitiesPerPageChange(35)}
          >
            Show 35 per page
          </button>
        </div>
      </div>
    </div>
  </div>

  {#if isLoading}
    <div class="animate-pulse space-y-4">
      {#each Array(3) as _}
        <div class="bg-gray-200 dark:bg-gray-800 h-24 rounded-lg"></div>
      {/each}
    </div>
  {:else if processedCommunities.length === 0}
    <div class="text-center py-8">
      <p class="text-gray-500 dark:text-gray-400">No communities found matching your search.</p>
    </div>
  {:else}
    <div class="space-y-4">
      {#each processedCommunities as community}
          <CommunityCard {community} on:joinRequest={handleJoinRequest} />
      {/each}

      <!-- Pagination controls -->
      {#if totalCount > communitiesPerPage}
        <div class="mt-6 flex flex-wrap justify-between items-center gap-4 border-t border-gray-200 dark:border-gray-800 pt-4">
          <div class="flex items-center">
            <span class="text-sm text-gray-500 dark:text-gray-400 mr-2">Show:</span>
            <select 
              class="bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-700 rounded-md px-2 py-1 text-sm"
              bind:value={communitiesPerPage}
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

  @keyframes pulse {
    0%, 100% { opacity: 0.5; }
    50% { opacity: 1; }
  }
  .animate-pulse {
    animation: pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite;
  }

  .group:hover .hidden.group-hover\:block {
    display: block;
  }
</style>