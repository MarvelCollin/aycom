<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { createLoggerWithPrefix } from '../../utils/logger';
  import { useTheme } from '../../hooks/useTheme';
  
  const logger = createLoggerWithPrefix('ExploreTabs');
  const dispatch = createEventDispatcher();
  const { theme } = useTheme();
  
  // Props
  export let activeTab: 'top' | 'latest' | 'people' | 'media' | 'communities' = 'top';
  
  // Reactive declarations
  $: isDarkMode = $theme === 'dark';
  
  // Handle tab change
  function handleTabChange(tab: 'top' | 'latest' | 'people' | 'media' | 'communities') {
    logger.debug('Tab changed', { from: activeTab, to: tab });
    dispatch('tabChange', tab);
  }
</script>

<div class="explore-tabs {isDarkMode ? 'explore-tabs-dark' : ''}">
  <div class="tabs-container">
    <button 
      class="tab-button {activeTab === 'top' ? 'active' : ''} {isDarkMode ? 'tab-button-dark' : ''}"
      on:click={() => handleTabChange('top')}
    >
      Top
    </button>
    <button 
      class="tab-button {activeTab === 'latest' ? 'active' : ''} {isDarkMode ? 'tab-button-dark' : ''}"
      on:click={() => handleTabChange('latest')}
    >
      Latest
    </button>
    <button 
      class="tab-button {activeTab === 'people' ? 'active' : ''} {isDarkMode ? 'tab-button-dark' : ''}"
      on:click={() => handleTabChange('people')}
    >
      People
    </button>
    <button 
      class="tab-button {activeTab === 'media' ? 'active' : ''} {isDarkMode ? 'tab-button-dark' : ''}"
      on:click={() => handleTabChange('media')}
    >
      Media
    </button>
    <button 
      class="tab-button {activeTab === 'communities' ? 'active' : ''} {isDarkMode ? 'tab-button-dark' : ''}"
      on:click={() => handleTabChange('communities')}
    >
      Communities
    </button>
  </div>
</div>

<style>
  .explore-tabs {
    border-bottom: 1px solid var(--border-color);
    background-color: var(--bg-primary);
  }
  
  .explore-tabs-dark {
    border-bottom: 1px solid var(--border-color-dark);
    background-color: var(--bg-primary-dark);
  }
  
  .tabs-container {
    display: flex;
    overflow-x: auto;
    scrollbar-width: none; /* Firefox */
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
    transition: color var(--transition-fast);
  }
  
  .tab-button-dark {
    color: var(--text-secondary-dark);
  }
  
  .tab-button:hover {
    color: var(--text-primary);
    background-color: var(--bg-hover);
  }
  
  .tab-button-dark:hover {
    color: var(--text-primary-dark);
    background-color: var(--bg-hover-dark);
  }
  
  .tab-button.active {
    color: var(--color-primary);
    font-weight: var(--font-weight-bold);
  }
  
  .tab-button.active::after {
    content: "";
    position: absolute;
    bottom: -1px;
    left: 0;
    right: 0;
    height: 4px;
    background-color: var(--color-primary);
    border-radius: var(--radius-sm) var(--radius-sm) 0 0;
  }
</style> 