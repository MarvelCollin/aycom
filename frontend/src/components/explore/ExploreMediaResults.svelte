<script lang="ts">
  import { createLoggerWithPrefix } from '../../utils/logger';
  
  const logger = createLoggerWithPrefix('ExploreMediaResults');
  
  // Props
  export let mediaResults: Array<{
    id: string;
    threadId: string;
    url: string;
    type: string;
    content: string;
    username: string;
    media_id?: string;
  }> = [];
  export let isLoading = false;
  
  // Log when media results change
  $: {
    if (!isLoading) {
      if (mediaResults.length > 0) {
        logger.debug('Media results loaded', { count: mediaResults.length });
      } else {
        logger.debug('No media results found');
      }
    }
  }
</script>

<div class="p-4">
  {#if isLoading && mediaResults.length === 0}
    <div class="grid grid-cols-3 gap-1">
      {#each Array(9) as _}
        <div class="aspect-square bg-gray-300 dark:bg-gray-700 animate-pulse rounded-md"></div>
      {/each}
    </div>
  {:else if mediaResults.length > 0}
    <div class="grid grid-cols-3 gap-1">
      {#each mediaResults as media}
        <a href={`/thread/${media.threadId}`} class="aspect-square rounded-md overflow-hidden relative block bg-gray-100 dark:bg-gray-800">
          {#if media.type === 'image'}
            <img src={media.url} alt={media.content} class="w-full h-full object-cover" />
          {:else if media.type === 'video'}
            <video src={media.url} class="w-full h-full object-cover"></video>
            <div class="absolute inset-0 flex items-center justify-center">
              <div class="bg-black bg-opacity-50 rounded-full p-2">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z" />
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
              </div>
            </div>
          {/if}
          <div class="absolute bottom-0 left-0 right-0 bg-gradient-to-t from-black to-transparent p-2">
            <div class="text-white text-xs truncate">
              @{media.username}
            </div>
          </div>
        </a>
      {/each}
    </div>
    
    <!-- Loading indicator for infinite scroll -->
    {#if isLoading}
      <div class="flex justify-center mt-6 mb-2">
        <div class="h-8 w-8 border-t-2 border-b-2 border-blue-500 rounded-full animate-spin"></div>
      </div>
    {/if}
  {:else}
    <div class="text-center py-10">
      <p class="text-gray-500 dark:text-gray-400">No media found</p>
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
  
  /* Spinner animation */
  @keyframes spin {
    0% { transform: rotate(0deg); }
    100% { transform: rotate(360deg); }
  }
  .animate-spin {
    animation: spin 1s linear infinite;
  }
</style> 