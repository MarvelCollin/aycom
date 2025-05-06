<script lang="ts">
  import { createEventDispatcher, onMount } from 'svelte';
  import { useTheme } from '../../hooks/useTheme';
  import LoadingSkeleton from '../common/LoadingSkeleton.svelte';
  
  const dispatch = createEventDispatcher();
  const { theme } = useTheme();
  
  // Props
  export let media: any[] = [];
  export let hasMore = false;
  export let isLoading = false;
  
  // Reactive declarations
  $: isDarkMode = $theme === 'dark';
  
  // Intersection observer for infinite scrolling
  let loadMoreTrigger;
  
  function setupInfiniteScroll() {
    if (typeof IntersectionObserver === 'undefined') {
      // Fallback for browsers that don't support IntersectionObserver
      return;
    }
    
    const observer = new IntersectionObserver((entries) => {
      const entry = entries[0];
      if (entry.isIntersecting && hasMore && !isLoading) {
        dispatch('loadMore');
      }
    }, { threshold: 0.1 });
    
    if (loadMoreTrigger) {
      observer.observe(loadMoreTrigger);
    }
    
    return () => {
      if (loadMoreTrigger) {
        observer.unobserve(loadMoreTrigger);
      }
    };
  }
  
  onMount(setupInfiniteScroll);
  
  $: if (media && !isLoading) {
    setupInfiniteScroll();
  }
</script>

<div class="p-4">
  <h2 class="font-bold text-xl mb-4">Media</h2>
  
  {#if media.length > 0}
    <div class="grid grid-cols-3 gap-1">
      {#each media as item}
        {#if item.media && item.media.length > 0}
          <a 
            href={`/thread/${item.id}`} 
            class="aspect-square block overflow-hidden bg-gray-200 dark:bg-gray-800 relative"
          >
            <!-- Display appropriate media based on type -->
            {#if item.media[0].type === 'image'}
              <img 
                src={item.media[0].url} 
                alt="Media content" 
                class="w-full h-full object-cover"
              />
            {:else if item.media[0].type === 'video'}
              <div class="w-full h-full flex items-center justify-center">
                <div class="absolute inset-0 bg-black opacity-50"></div>
                <svg xmlns="http://www.w3.org/2000/svg" class="h-12 w-12 text-white z-10" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z" />
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                <video 
                  src={item.media[0].url} 
                  class="w-full h-full object-cover absolute inset-0"
                ></video>
              </div>
            {:else if item.media[0].type === 'gif'}
              <img 
                src={item.media[0].url} 
                alt="GIF content" 
                class="w-full h-full object-cover"
              />
            {/if}
          </a>
        {/if}
      {/each}
    </div>
    
    <!-- Infinite scroll trigger -->
    {#if hasMore}
      <div 
        class="py-8 flex justify-center" 
        bind:this={loadMoreTrigger}
      >
        {#if isLoading}
          <div class="flex items-center space-x-2">
            <div class="w-4 h-4 rounded-full bg-blue-500 animate-pulse"></div>
            <div class="w-4 h-4 rounded-full bg-blue-500 animate-pulse" style="animation-delay: 0.2s"></div>
            <div class="w-4 h-4 rounded-full bg-blue-500 animate-pulse" style="animation-delay: 0.4s"></div>
          </div>
        {:else}
          <span class="text-gray-500 dark:text-gray-400 text-sm">Loading more media...</span>
        {/if}
      </div>
    {/if}
  {:else if isLoading}
    <LoadingSkeleton type="media" />
  {:else}
    <p class="text-center text-gray-500 dark:text-gray-400 py-8">
      No media content found for your search.
    </p>
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