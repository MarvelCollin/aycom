<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { useTheme } from '../../hooks/useTheme';
  import type { ITrend } from '../../interfaces/ISocialMedia';
  import { createLoggerWithPrefix } from '../../utils/logger';
  
  const logger = createLoggerWithPrefix('ExploreTrending');
  const dispatch = createEventDispatcher();
  const { theme } = useTheme();
  
  // Reactive declarations
  $: isDarkMode = $theme === 'dark';
  
  // Props
  export let trends: ITrend[] = [];
  export let isTrendsLoading = false;
  
  // Load trending hashtags
  function loadTrendingThreads(hashtag: string) {
    logger.debug('Loading trending hashtag', { hashtag });
    dispatch('loadTrending', hashtag);
  }
  
  // Log when trends are loaded
  $: {
    if (!isTrendsLoading) {
      if (trends.length > 0) {
        logger.debug('Trends loaded', { count: trends.length });
      } else {
        logger.debug('No trends available');
      }
    }
  }
  
  // Handle hashtag click
  function handleHashtagClick(hashtag: string) {
    dispatch('hashtagClick', hashtag);
  }
  
  // Handle view thread by hashtag
  function viewThreadsByHashtag(hashtag: string) {
    dispatch('viewThreads', hashtag);
  }
</script>

<div class="bg-gray-50 dark:bg-gray-900 rounded-xl p-4">
  <div class="flex items-center justify-between pb-3 border-b border-gray-200 dark:border-gray-800">
    <h2 class="font-bold text-xl">Trending Now</h2>
  </div>
  
  {#if trends.length > 0}
    <div class="divide-y divide-gray-200 dark:divide-gray-800">
      {#each trends as trend, i}
        <div class="py-3">
          <div class="flex items-center justify-between">
            <div>
              <button 
                class="text-blue-500 font-medium hover:underline"
                on:click={() => viewThreadsByHashtag(trend.title)}
              >
                #{trend.title}
              </button>
              <p class="text-sm text-gray-500 dark:text-gray-400">{trend.postCount || 0} posts</p>
              {#if trend.category && trend.category !== 'Trending'}
                <span class="mt-1 inline-block px-2 py-0.5 bg-gray-200 dark:bg-gray-800 rounded-full text-xs">
                  {trend.category}
                </span>
              {/if}
            </div>
            <span class="text-gray-500 dark:text-gray-400 text-sm">#{i + 1}</span>
          </div>
        </div>
      {/each}
    </div>
  {:else}
    <p class="text-center text-gray-500 dark:text-gray-400 py-4">No trending topics available.</p>
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