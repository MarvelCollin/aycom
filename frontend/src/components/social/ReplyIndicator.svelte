<script lang="ts">
  export let count: number = 0;
  export let isDarkMode: boolean = false;
  export let isExpanded: boolean = false;
</script>

<div class="reply-indicator-container mt-2 mb-2 ml-14">
  <button 
    class="reply-indicator-button flex items-center py-1.5 px-3 rounded-full transition-colors {isDarkMode ? 'bg-gray-800 text-blue-400 hover:text-blue-300 hover:bg-gray-700' : 'bg-gray-50 text-blue-500 hover:text-blue-600 hover:bg-gray-100'}"
    on:click
  >
    {#if isExpanded}
      <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 15l7-7 7 7" />
      </svg>
    {:else}
      <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-2" viewBox="0 0 20 20" fill="currentColor">
        <path fill-rule="evenodd" d="M7.293 14.707a1 1 0 010-1.414L10.586 10 7.293 6.707a1 1 0 011.414-1.414l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0z" clip-rule="evenodd" />
      </svg>
    {/if}
    
    <span class="flex items-center">
      {#if count > 0}
        {isExpanded ? 'Hide' : 'View'} {count} {count === 1 ? 'reply' : 'replies'}
        {#if count > 3}
          <span class="ml-1.5 px-1.5 py-0.5 text-xs rounded-full bg-blue-500 text-white">
            {count}
          </span>
        {/if}
      {:else}
        {isExpanded ? 'Hide thread' : 'Reply thread'}
      {/if}
    </span>
  </button>
</div>

<style>
  .reply-indicator-button {
    border: 1px solid;
    border-color: transparent;
    box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
    font-size: 0.875rem;
  }
  
  .reply-indicator-container {
    position: relative;
    z-index: 1;
  }
</style> 