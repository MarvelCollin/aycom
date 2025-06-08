<script lang="ts">
  import { createEventDispatcher, onMount } from 'svelte';
  import { useTheme } from '../../hooks/useTheme';
  import type { ITrend } from '../../interfaces/ITrend';
  import { createLoggerWithPrefix } from '../../utils/logger';
  
  const logger = createLoggerWithPrefix('ExploreTrending');
  const dispatch = createEventDispatcher();
  const { theme } = useTheme();
  
  // Reactive declarations
  $: isDarkMode = $theme === 'dark';
  
  // Props
  export let trends: ITrend[] = [];
  export let isTrendsLoading = false;
  
  // Sample trends for empty state
  let sampleTrends = [
    { title: 'programming', category: 'Technology', post_count: 125 },
    { title: 'design', category: 'Creative', post_count: 98 },
    { title: 'webdev', category: 'Technology', post_count: 87 }
  ];
  
  // Animation timing
  let showSampleTrends = false;
  
  onMount(() => {
    // Show sample trends after a delay if no real trends are available
    if (trends.length === 0 && !isTrendsLoading) {
      setTimeout(() => {
        showSampleTrends = true;
      }, 500);
    }
  });
  
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
        showSampleTrends = false;
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
      <div class="trending-item-skeleton animate-pulse" style="animation-delay: 0.2s"></div>
      <div class="trending-item-skeleton animate-pulse" style="animation-delay: 0.4s"></div>
    </div>
  {:else if trends.length > 0}
    <div class="trending-list">
      {#each trends as trend, i}
        <div class="trending-item">
          <div class="trending-item-content">
            <div>
              <button 
                class="trending-hashtag"
                on:click={() => handleHashtagClick(trend.title || trend.name || '')}
              >
                #{trend.title || trend.name || ''}
              </button>
              <p class="trending-post-count">{trend.post_count || trend.tweet_count || 0} posts</p>
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
    <div class="trending-empty">
      <div class="trending-empty-icon pulse-animation">
        <svg xmlns="http://www.w3.org/2000/svg" width="50" height="50" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
          <line x1="12" y1="20" x2="12" y2="10"></line>
          <line x1="18" y1="20" x2="18" y2="4"></line>
          <line x1="6" y1="20" x2="6" y2="16"></line>
        </svg>
      </div>
      <p class="trending-empty-text">No trending topics yet</p>
      <p class="trending-empty-subtext">Be the first to start a trending conversation!</p>
      
      {#if showSampleTrends}
        <div class="sample-trends fade-in">
          <p class="sample-trends-title">Try exploring these topics:</p>
          <div class="sample-trends-list">
            {#each sampleTrends as trend}
              <button 
                class="sample-trend-chip"
                on:click={() => handleHashtagClick(trend.title)}
              >
                #{trend.title}
              </button>
            {/each}
          </div>
        </div>
      {/if}
    </div>
  {/if}
</div>

<style>
  .trending-container {
    background-color: var(--bg-secondary);
    border-radius: var(--radius-lg);
    padding: var(--space-4);
    margin-bottom: var(--space-4);
    box-shadow: var(--shadow-sm);
    border: 1px solid var(--border-color);
    transition: transform var(--transition-normal), box-shadow var(--transition-normal);
  }
  
  .trending-container:hover {
    transform: translateY(-2px);
    box-shadow: var(--shadow-md);
  }
  
  .trending-container-dark {
    background-color: var(--bg-secondary-dark);
    border-color: var(--border-color-dark);
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
    position: relative;
    display: inline-block;
  }
  
  .trending-title::after {
    content: "";
    position: absolute;
    left: 0;
    bottom: -5px;
    width: 30px;
    height: 3px;
    background-color: var(--color-primary);
    border-radius: var(--radius-full);
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
    transition: background-color var(--transition-fast);
  }
  
  .trending-item:hover {
    background-color: var(--hover-bg);
    border-radius: var(--radius-md);
    padding-left: var(--space-2);
    padding-right: var(--space-2);
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
    transition: all var(--transition-fast);
    text-align: left;
    font-size: var(--font-size-md);
  }
  
  .trending-hashtag:hover {
    color: var(--color-primary-hover);
    text-decoration: underline;
    transform: scale(1.02);
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
    padding: var(--space-1) var(--space-2);
    background-color: var(--bg-tertiary);
    border-radius: var(--radius-full);
    font-size: var(--font-size-xs);
    color: var(--text-secondary);
    font-weight: var(--font-weight-medium);
  }
  
  .trending-container-dark .trending-category {
    background-color: var(--bg-tertiary-dark);
    color: var(--text-secondary-dark);
  }
  
  .trending-rank {
    font-size: var(--font-size-sm);
    color: var(--text-secondary);
    font-weight: var(--font-weight-bold);
    background-color: var(--bg-tertiary);
    padding: var(--space-1) var(--space-2);
    border-radius: var(--radius-full);
  }
  
  .trending-container-dark .trending-rank {
    color: var(--text-secondary-dark);
    background-color: var(--bg-tertiary-dark);
  }
  
  .trending-empty {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: var(--space-6) var(--space-4);
    text-align: center;
    background: linear-gradient(135deg, 
      rgba(var(--color-primary-rgb), 0.05) 0%, 
      rgba(var(--color-primary-rgb), 0.02) 100%);
    border-radius: var(--radius-lg);
    min-height: 200px;
  }
  
  .trending-container-dark .trending-empty {
    background: linear-gradient(135deg, 
      rgba(var(--color-primary-rgb), 0.1) 0%, 
      rgba(var(--color-primary-rgb), 0.05) 100%);
  }
  
  .trending-empty-icon {
    color: var(--color-primary);
    margin-bottom: var(--space-3);
    opacity: 0.8;
  }
  
  .pulse-animation {
    animation: pulse-fade 2s infinite ease-in-out;
  }
  
  @keyframes pulse-fade {
    0% { opacity: 0.5; transform: scale(0.95); }
    50% { opacity: 1; transform: scale(1.05); }
    100% { opacity: 0.5; transform: scale(0.95); }
  }
  
  .trending-empty-text {
    color: var(--text-primary);
    font-weight: var(--font-weight-bold);
    font-size: var(--font-size-lg);
    margin-bottom: var(--space-2);
  }
  
  .trending-container-dark .trending-empty-text {
    color: var(--text-primary-dark);
  }
  
  .trending-empty-subtext {
    color: var(--text-secondary);
    font-size: var(--font-size-sm);
    margin-bottom: var(--space-4);
  }
  
  .trending-container-dark .trending-empty-subtext {
    color: var(--text-secondary-dark);
  }
  
  .sample-trends {
    margin-top: var(--space-4);
    width: 100%;
  }
  
  .fade-in {
    animation: fadeIn 0.5s ease-in-out;
  }
  
  @keyframes fadeIn {
    from { opacity: 0; transform: translateY(10px); }
    to { opacity: 1; transform: translateY(0); }
  }
  
  .sample-trends-title {
    font-size: var(--font-size-sm);
    color: var(--text-secondary);
    margin-bottom: var(--space-2);
  }
  
  .sample-trends-list {
    display: flex;
    flex-wrap: wrap;
    justify-content: center;
    gap: var(--space-2);
  }
  
  .sample-trend-chip {
    background-color: var(--bg-primary);
    color: var(--color-primary);
    padding: var(--space-1) var(--space-3);
    border-radius: var(--radius-full);
    font-size: var(--font-size-sm);
    border: 1px solid var(--color-primary);
    cursor: pointer;
    transition: all var(--transition-fast);
  }
  
  .sample-trend-chip:hover {
    background-color: var(--color-primary);
    color: white;
    transform: translateY(-2px);
    box-shadow: var(--shadow-sm);
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