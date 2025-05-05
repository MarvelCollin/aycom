<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { createLoggerWithPrefix } from '../../utils/logger';
  
  const logger = createLoggerWithPrefix('ExploreTabs');
  const dispatch = createEventDispatcher();
  
  // Props
  export let activeTab: 'top' | 'latest' | 'people' | 'media' | 'communities' = 'top';
  
  // Handle tab change
  function handleTabChange(tab: 'top' | 'latest' | 'people' | 'media' | 'communities') {
    logger.debug('Tab changed', { from: activeTab, to: tab });
    dispatch('tabChange', tab);
  }
</script>

<div class="border-b border-gray-200 dark:border-gray-800">
  <div class="flex">
    <button 
      class="px-4 py-3 font-medium {activeTab === 'top' ? 'text-blue-500 border-b-2 border-blue-500' : 'text-gray-500 dark:text-gray-400'}"
      on:click={() => handleTabChange('top')}
    >
      Top
    </button>
    <button 
      class="px-4 py-3 font-medium {activeTab === 'latest' ? 'text-blue-500 border-b-2 border-blue-500' : 'text-gray-500 dark:text-gray-400'}"
      on:click={() => handleTabChange('latest')}
    >
      Latest
    </button>
    <button 
      class="px-4 py-3 font-medium {activeTab === 'people' ? 'text-blue-500 border-b-2 border-blue-500' : 'text-gray-500 dark:text-gray-400'}"
      on:click={() => handleTabChange('people')}
    >
      People
    </button>
    <button 
      class="px-4 py-3 font-medium {activeTab === 'media' ? 'text-blue-500 border-b-2 border-blue-500' : 'text-gray-500 dark:text-gray-400'}"
      on:click={() => handleTabChange('media')}
    >
      Media
    </button>
    <button 
      class="px-4 py-3 font-medium {activeTab === 'communities' ? 'text-blue-500 border-b-2 border-blue-500' : 'text-gray-500 dark:text-gray-400'}"
      on:click={() => handleTabChange('communities')}
    >
      Communities
    </button>
  </div>
</div> 