<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { useTheme } from '../../hooks/useTheme';
  import type { ITrend } from '../../interfaces/ISocialMedia';
  
  const dispatch = createEventDispatcher();
  const { theme } = useTheme();
  
  // Reactive declarations
  $: isDarkMode = $theme === 'dark';
  
  // Props
  export let trends: ITrend[] = [];
  export let isTrendsLoading = false;
  
  // Load trending hashtags
  function loadTrendingThreads(hashtag: string) {
    dispatch('loadTrending', hashtag);
  }
</script>

<div class="p-4">
  <h2 class="text-xl font-bold mb-4">Trending</h2>
  
  {#if isTrendsLoading}
    <div class="animate-pulse space-y-4">
      {#each Array(10) as _}
        <div class="flex space-x-4">
          <div class="flex-1 space-y-2 py-1">
            <div class="h-4 bg-gray-300 dark:bg-gray-700 rounded w-1/4"></div>
            <div class="h-4 bg-gray-300 dark:bg-gray-700 rounded w-3/4"></div>
          </div>
        </div>
      {/each}
    </div>
  {:else if trends.length > 0}
    <ul class="divide-y divide-gray-200 dark:divide-gray-800">
      {#each trends as trend, i}
        <li>
          <button 
            class="py-3 w-full text-left hover:bg-gray-50 dark:hover:bg-gray-900"
            on:click={() => loadTrendingThreads(trend.title)}
          >
            <div class="text-sm text-gray-500 dark:text-gray-400">#{i + 1} Trending</div>
            <div class="font-bold {isDarkMode ? 'text-white' : 'text-black'} mt-1">{trend.title}</div>
            <div class="text-sm text-gray-500 dark:text-gray-400">{trend.postCount} posts</div>
          </button>
        </li>
      {/each}
    </ul>
  {:else}
    <div class="text-center py-10">
      <p class="text-gray-500 dark:text-gray-400">No trending topics available</p>
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
</style> 