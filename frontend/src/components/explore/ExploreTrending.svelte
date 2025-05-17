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

<div class="trending-container {isDarkMode ? 'trending-container-dark' : ''}">
  <div class="trending-header">
    <h2 class="trending-title">Trending Now</h2>
  </div>
  
  {#if isTrendsLoading}
    <div class="trending-loading">
      <div class="trending-item-skeleton animate-pulse"></div>
      <div class="trending-item-skeleton animate-pulse"></div>
      <div class="trending-item-skeleton animate-pulse"></div>
    </div>
  {:else if trends.length > 0}
    <div class="trending-list">
      {#each trends as trend, i}
        <div class="trending-item">
          <div class="trending-item-content">
            <div>
              <button 
                class="trending-hashtag"
                on:click={() => viewThreadsByHashtag(trend.title)}
              >
                #{trend.title}
              </button>
              <p class="trending-post-count">{trend.postCount || 0} posts</p>
              {#if trend.category && trend.category !== 'Trending'}
                <span class="trending-category">
                  {trend.category}
                </span>
              {/if}
            </div>
            <span class="trending-rank">#{i + 1}</span>
          </div>
        </div>
      {/each}
    </div>
  {:else}
    <p class="trending-empty">No trending topics available.</p>
  {/if}
</div>

<style>
  .trending-container {
    background-color: var(--bg-secondary);
    border-radius: var(--radius-lg);
    padding: var(--space-4);
    margin-bottom: var(--space-4);
  }
  
  .trending-container-dark {
    background-color: var(--bg-secondary-dark);
  }
  
  .trending-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding-bottom: var(--space-3);
    border-bottom: 1px solid var(--border-color);
    margin-bottom: var(--space-3);
  }
  
  .trending-container-dark .trending-header {
    border-bottom-color: var(--border-color-dark);
  }
  
  .trending-title {
    font-weight: var(--font-weight-bold);
    font-size: var(--font-size-xl);
    color: var(--text-primary);
  }
  
  .trending-container-dark .trending-title {
    color: var(--text-primary-dark);
  }
  
  .trending-list {
    display: flex;
    flex-direction: column;
  }
  
  .trending-item {
    padding: var(--space-3) 0;
    border-bottom: 1px solid var(--border-color);
  }
  
  .trending-container-dark .trending-item {
    border-bottom-color: var(--border-color-dark);
  }
  
  .trending-item:last-child {
    border-bottom: none;
  }
  
  .trending-item-content {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
  }
  
  .trending-hashtag {
    color: var(--color-primary);
    font-weight: var(--font-weight-medium);
    background: none;
    border: none;
    padding: 0;
    cursor: pointer;
    transition: text-decoration var(--transition-fast);
    text-align: left;
  }
  
  .trending-hashtag:hover {
    text-decoration: underline;
  }
  
  .trending-post-count {
    font-size: var(--font-size-sm);
    color: var(--text-secondary);
    margin: var(--space-1) 0;
  }
  
  .trending-container-dark .trending-post-count {
    color: var(--text-secondary-dark);
  }
  
  .trending-category {
    display: inline-block;
    margin-top: var(--space-1);
    padding: 0 var(--space-2);
    background-color: var(--bg-tertiary);
    border-radius: var(--radius-full);
    font-size: var(--font-size-xs);
    color: var(--text-secondary);
  }
  
  .trending-container-dark .trending-category {
    background-color: var(--bg-tertiary-dark);
    color: var(--text-secondary-dark);
  }
  
  .trending-rank {
    font-size: var(--font-size-sm);
    color: var(--text-secondary);
  }
  
  .trending-container-dark .trending-rank {
    color: var(--text-secondary-dark);
  }
  
  .trending-empty {
    color: var(--text-secondary);
    text-align: center;
    padding: var(--space-4) 0;
  }
  
  .trending-container-dark .trending-empty {
    color: var(--text-secondary-dark);
  }
  
  .trending-loading {
    padding: var(--space-2) 0;
  }
  
  .trending-item-skeleton {
    height: 80px;
    margin-bottom: var(--space-3);
    background-color: var(--bg-tertiary);
    border-radius: var(--radius-md);
  }
  
  .trending-container-dark .trending-item-skeleton {
    background-color: var(--bg-tertiary-dark);
  }
  
  /* Animation */
  @keyframes pulse {
    0%, 100% { opacity: 0.5; }
    50% { opacity: 0.8; }
  }
  
  .animate-pulse {
    animation: pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite;
  }
</style> 