<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { createLoggerWithPrefix } from '../../utils/logger';
  import { useTheme } from '../../hooks/useTheme';
  import { onMount } from 'svelte';
  
  const logger = createLoggerWithPrefix('ExploreTabs');
  const dispatch = createEventDispatcher();
  const { theme } = useTheme();
  
  // Props
  export let activeTab: 'top' | 'latest' | 'people' | 'media' | 'communities' = 'top';
  
  // Reactive declarations
  $: isDarkMode = $theme === 'dark';
  
  // References
  let tabsContainer: HTMLElement;
  let activeTabElement: HTMLElement;
  
  // Handle tab change
  function handleTabChange(tab: 'top' | 'latest' | 'people' | 'media' | 'communities') {
    logger.debug('Tab changed', { from: activeTab, to: tab });
    activeTab = tab;
    dispatch('tabChange', tab);
    
    // After update, scroll the active tab into view
    setTimeout(() => {
      const newActiveTab = document.querySelector(`.tab-button.active`) as HTMLElement;
      if (newActiveTab) {
        scrollTabIntoView(newActiveTab);
      }
    }, 10);
  }
  
  // Scroll the active tab into view
  function scrollTabIntoView(tabElement: HTMLElement) {
    if (tabsContainer && tabElement) {
      const containerRect = tabsContainer.getBoundingClientRect();
      const tabRect = tabElement.getBoundingClientRect();
      
      // Check if tab is not fully visible
      if (tabRect.left < containerRect.left || tabRect.right > containerRect.right) {
        // Calculate the scroll position to center the tab
        const scrollPosition = tabElement.offsetLeft - tabsContainer.offsetWidth / 2 + tabElement.offsetWidth / 2;
        tabsContainer.scrollTo({
          left: scrollPosition,
          behavior: 'smooth'
        });
      }
    }
  }
  
  // Initialize after component mounts
  onMount(() => {
    // Get active tab element and scroll it into view
    activeTabElement = document.querySelector('.tab-button.active');
    if (activeTabElement) {
      scrollTabIntoView(activeTabElement);
    }
  });
</script>

<div class="explore-tabs {isDarkMode ? 'explore-tabs-dark' : ''}">
  <div class="tabs-container" bind:this={tabsContainer}>
    <button 
      class="tab-button {activeTab === 'top' ? 'active' : ''} {isDarkMode ? 'tab-button-dark' : ''}"
      on:click={() => handleTabChange('top')}
      aria-selected={activeTab === 'top'}
      role="tab"
    >
      Top
    </button>
    <button 
      class="tab-button {activeTab === 'latest' ? 'active' : ''} {isDarkMode ? 'tab-button-dark' : ''}"
      on:click={() => handleTabChange('latest')}
      aria-selected={activeTab === 'latest'}
      role="tab"
    >
      Latest
    </button>
    <button 
      class="tab-button {activeTab === 'people' ? 'active' : ''} {isDarkMode ? 'tab-button-dark' : ''}"
      on:click={() => handleTabChange('people')}
      aria-selected={activeTab === 'people'}
      role="tab"
    >
      People
    </button>
    <button 
      class="tab-button {activeTab === 'media' ? 'active' : ''} {isDarkMode ? 'tab-button-dark' : ''}"
      on:click={() => handleTabChange('media')}
      aria-selected={activeTab === 'media'}
      role="tab"
    >
      Media
    </button>
    <button 
      class="tab-button {activeTab === 'communities' ? 'active' : ''} {isDarkMode ? 'tab-button-dark' : ''}"
      on:click={() => handleTabChange('communities')}
      aria-selected={activeTab === 'communities'}
      role="tab"
    >
      Communities
    </button>
  </div>
  <div class="tab-indicator-container">
    <div class="tab-indicator {activeTab} {isDarkMode ? 'tab-indicator-dark' : ''}"></div>
  </div>
</div>

<style>
  .explore-tabs {
    border-bottom: 1px solid var(--border-color);
    background-color: var(--bg-primary);
    position: relative;
  }
  
  .explore-tabs-dark {
    border-bottom: 1px solid var(--border-color-dark);
    background-color: var(--dark-bg-primary);
  }
  
  .tabs-container {
    display: flex;
    overflow-x: auto;
    scrollbar-width: none; /* Firefox */
    -ms-overflow-style: none; /* IE/Edge */
    position: relative;
    z-index: 2;
  }
  
  .tabs-container::-webkit-scrollbar {
    display: none; /* Chrome, Safari, Edge */
  }
  
  .tab-button {
    padding: var(--space-3) var(--space-4);
    font-weight: var(--font-weight-medium);
    background: none;
    border: none;
    cursor: pointer;
    color: var(--text-secondary);
    position: relative;
    white-space: nowrap;
    transition: color var(--transition-fast), background-color var(--transition-fast);
    flex: 1 1 auto;
    min-width: max-content;
  }
  
  .tab-button-dark {
    color: var(--dark-text-secondary, rgba(255, 255, 255, 0.7));
  }
  
  .tab-button:hover {
    color: var(--text-primary);
    background-color: var(--bg-hover);
  }
  
  .tab-button-dark:hover {
    color: var(--dark-text-primary, #fff);
    background-color: rgba(255, 255, 255, 0.1);
  }
  
  .tab-button.active {
    color: var(--color-primary);
    font-weight: var(--font-weight-bold);
  }
  
  .tab-indicator-container {
    position: absolute;
    bottom: 0;
    left: 0;
    width: 100%;
    height: 4px;
    overflow: hidden;
    z-index: 1;
  }
  
  .tab-indicator {
    position: absolute;
    height: 4px;
    background-color: var(--color-primary);
    border-radius: var(--radius-sm) var(--radius-sm) 0 0;
    transition: transform 0.3s ease;
  }
  
  .tab-indicator.top {
    width: 20%;
    transform: translateX(0%);
  }
  
  .tab-indicator.latest {
    width: 20%;
    transform: translateX(100%);
  }
  
  .tab-indicator.people {
    width: 20%;
    transform: translateX(200%);
  }
  
  .tab-indicator.media {
    width: 20%;
    transform: translateX(300%);
  }
  
  .tab-indicator.communities {
    width: 20%;
    transform: translateX(400%);
  }
  
  /* Responsive adjustments */
  @media (max-width: 576px) {
    .tab-button {
      padding: var(--space-2) var(--space-3);
      font-size: var(--font-size-sm);
    }
    
    /* Force equal width tabs in mobile */
    .tab-indicator {
      width: 20%;
    }
  }
</style> 