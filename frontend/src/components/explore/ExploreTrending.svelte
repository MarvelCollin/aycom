<script lang="ts">
  import { createEventDispatcher, onMount } from 'svelte';
  import { useTheme } from '../../hooks/useTheme';
  import type { ITrend } from '../../interfaces/ITrend';
  import { createLoggerWithPrefix } from '../../utils/logger';

  const logger = createLoggerWithPrefix('ExploreTrending');
  const dispatch = createEventDispatcher();
  const { theme } = useTheme();

  $: isDarkMode = $theme === 'dark';

  export let trends: ITrend[] = [];
  export let isTrendsLoading = false;

  let sampleTrends = [
    { title: 'programming', category: 'Technology', post_count: 125 },
    { title: 'design', category: 'Creative', post_count: 98 },
    { title: 'webdev', category: 'Technology', post_count: 87 }
  ];

  let showSampleTrends = false;

  onMount(() => {

    if (trends.length === 0 && !isTrendsLoading) {
      setTimeout(() => {
        showSampleTrends = true;
      }, 500);
    }
  });

  function handleHashtagClick(hashtag: string) {
    logger.debug('Hashtag clicked', { hashtag });
    dispatch('hashtagClick', hashtag);
  }

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
</script>

<div class="twitter-trends-container {isDarkMode ? 'twitter-trends-container-dark' : ''}">
  <h2 class="trends-header">What's happening</h2>

  {#if isTrendsLoading}
    <div class="trends-loading">
      <div class="trend-skeleton"></div>
      <div class="trend-skeleton"></div>
      <div class="trend-skeleton"></div>
    </div>
  {:else if trends.length > 0}
    <div class="trends-list">
      {#each trends as trend, i}
        <button 
          class="trend-item" 
          on:click={() => handleHashtagClick(trend.title || trend.name || '')}
        >
          <div class="trend-content">
            <div class="trend-category">{trend.category || 'Trending'}</div>
            <div class="trend-tag">#{trend.title || trend.name || ''}</div>
            <div class="trend-metrics">{trend.post_count || trend.tweet_count || 0} posts</div>
          </div>
          <div class="trend-more">
            <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <polyline points="9 18 15 12 9 6"></polyline>
            </svg>
          </div>
        </button>
      {/each}
    </div>
  {:else}
    <div class="trends-empty">
      <div class="trends-empty-icon">
        <svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
          <line x1="12" y1="20" x2="12" y2="10"></line>
          <line x1="18" y1="20" x2="18" y2="4"></line>
          <line x1="6" y1="20" x2="6" y2="16"></line>
        </svg>
      </div>
      <h3 class="trends-empty-title">No trends available</h3>
      <p class="trends-empty-text">Check back soon for trending topics</p>

      {#if showSampleTrends}
        <div class="sample-trends">
          <h4 class="sample-trends-title">Try these topics</h4>
          <div class="sample-trends-grid">
            {#each sampleTrends as trend}
              <button 
                class="sample-trend-tag"
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
  .twitter-trends-container {
    background-color: var(--bg-primary);
    border-radius: 16px;
    overflow: hidden;
  }

  .twitter-trends-container-dark {
    background-color: var(--dark-bg-primary);
  }

  .trends-header {
    padding: 12px 16px;
    font-size: 20px;
    font-weight: 800;
    border-bottom: 1px solid var(--border-color);
    margin: 0;
    color: var(--text-primary);
  }

  .twitter-trends-container-dark .trends-header {
    color: var(--dark-text-primary);
    border-bottom-color: var(--dark-border-color);
  }

  .trends-list {
    display: flex;
    flex-direction: column;
  }

  .trend-item {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 12px 16px;
    border-bottom: 1px solid var(--border-color);
    background-color: transparent;
    border-left: none;
    border-right: none;
    border-top: none;
    text-align: left;
    cursor: pointer;
    transition: background-color 0.2s ease;
    width: 100%;
  }

  .trend-item:hover {
    background-color: var(--hover-bg);
  }

  .twitter-trends-container-dark .trend-item {
    border-bottom-color: var(--dark-border-color);
  }

  .twitter-trends-container-dark .trend-item:hover {
    background-color: var(--dark-hover-bg);
  }

  .trend-content {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .trend-category {
    font-size: 13px;
    color: var(--text-secondary);
  }

  .twitter-trends-container-dark .trend-category {
    color: var(--dark-text-secondary);
  }

  .trend-tag {
    font-size: 15px;
    font-weight: 700;
    color: var(--text-primary);
  }

  .twitter-trends-container-dark .trend-tag {
    color: var(--dark-text-primary);
  }

  .trend-metrics {
    font-size: 13px;
    color: var(--text-secondary);
  }

  .twitter-trends-container-dark .trend-metrics {
    color: var(--dark-text-secondary);
  }

  .trend-more {
    color: var(--text-secondary);
  }

  .twitter-trends-container-dark .trend-more {
    color: var(--dark-text-secondary);
  }

  .trends-loading {
    padding: 16px;
  }

  .trend-skeleton {
    height: 68px;
    background: linear-gradient(
      90deg,
      var(--bg-tertiary) 0%,
      var(--bg-secondary) 50%,
      var(--bg-tertiary) 100%
    );
    border-radius: 8px;
    margin-bottom: 12px;
    animation: pulse 1.5s ease-in-out infinite;
  }

  .twitter-trends-container-dark .trend-skeleton {
    background: linear-gradient(
      90deg,
      var(--dark-bg-tertiary) 0%,
      var(--dark-bg-secondary) 50%,
      var(--dark-bg-tertiary) 100%
    );
  }

  @keyframes pulse {
    0% {
      opacity: 0.6;
    }
    50% {
      opacity: 0.3;
    }
    100% {
      opacity: 0.6;
    }
  }

  .trends-empty {
    padding: 32px 16px;
    text-align: center;
  }

  .trends-empty-icon {
    color: var(--text-secondary);
    margin-bottom: 16px;
  }

  .twitter-trends-container-dark .trends-empty-icon {
    color: var(--dark-text-secondary);
  }

  .trends-empty-title {
    font-size: 18px;
    font-weight: 700;
    color: var(--text-primary);
    margin: 0 0 8px;
  }

  .twitter-trends-container-dark .trends-empty-title {
    color: var(--dark-text-primary);
  }

  .trends-empty-text {
    font-size: 15px;
    color: var(--text-secondary);
    margin: 0;
  }

  .twitter-trends-container-dark .trends-empty-text {
    color: var(--dark-text-secondary);
  }

  .sample-trends {
    margin-top: 24px;
  }

  .sample-trends-title {
    font-size: 16px;
    font-weight: 600;
    color: var(--text-primary);
    margin: 0 0 12px;
  }

  .twitter-trends-container-dark .sample-trends-title {
    color: var(--dark-text-primary);
  }

  .sample-trends-grid {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    justify-content: center;
  }

  .sample-trend-tag {
    background-color: var(--color-primary-bg);
    color: var(--color-primary);
    border: none;
    border-radius: 16px;
    padding: 8px 16px;
    font-size: 14px;
    font-weight: 500;
    cursor: pointer;
    transition: background-color 0.2s;
  }

  .sample-trend-tag:hover {
    background-color: var(--hover-primary);
  }
</style>