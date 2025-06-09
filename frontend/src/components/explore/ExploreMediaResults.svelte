<script lang="ts">
  import { createEventDispatcher, onMount, onDestroy } from 'svelte';
  import { useTheme } from '../../hooks/useTheme';
  import { createLoggerWithPrefix } from '../../utils/logger';
  import LoadingSkeleton from '../common/LoadingSkeleton.svelte';
  
  const logger = createLoggerWithPrefix('ExploreMediaResults');
  const dispatch = createEventDispatcher();
  const { theme } = useTheme();
  
  // Reactive declarations
  $: isDarkMode = $theme === 'dark';
  
  // Define raw thread types to handle different API response formats
  type MediaItem = {
    url: string;
    type: string;
  };
  
  type ThreadWithMedia = {
    id: string;
    media?: MediaItem[];
    attachments?: MediaItem[];
    images?: string[];
    videos?: string[];
  };
  
  // Props
  export let media: ThreadWithMedia[] = [];
  export let isLoading = false;
  export let hasMore = true;
  
  // Infinite scroll handling
  let observer: IntersectionObserver;
  let loadMoreTrigger: HTMLDivElement;
  let isIntersecting = false;
  
  // Set up intersection observer for infinite scroll
  function setupIntersectionObserver() {
    if (typeof IntersectionObserver !== 'undefined') {
      observer = new IntersectionObserver(
        (entries) => {
          const [entry] = entries;
          isIntersecting = entry.isIntersecting;
          
          if (isIntersecting && hasMore && !isLoading) {
            logger.debug('Load more trigger intersected, loading more media');
            loadMore();
      }
        },
        {
          root: null,
          rootMargin: '100px',
          threshold: 0.1
        }
      );
    
    if (loadMoreTrigger) {
      observer.observe(loadMoreTrigger);
    }
    }
  }
  
  // Load more media
  function loadMore() {
    logger.debug('Loading more media items');
    dispatch('loadMore');
  }
  
  // Get appropriate media element based on type
  function getMediaElement(mediaItem: { url: string; type?: string }) {
    if (!mediaItem || !mediaItem.url) {
      return null;
    }
    
    // Extract file extension for type inference if not provided
    const url = mediaItem.url;
    const fileExt = url.split('.').pop()?.toLowerCase() || '';
    
    // Determine media type from explicit type or file extension
    const type = mediaItem.type || '';
    
    if (type.includes('video') || ['mp4', 'webm', 'mov'].includes(fileExt)) {
      return {
        type: 'video',
        url: mediaItem.url
      };
    } else if (type.includes('gif') || fileExt === 'gif') {
      return {
        type: 'gif',
        url: mediaItem.url
      };
    } else {
      return {
        type: 'image',
        url: mediaItem.url
    };
  }
  }
  
  // Navigate to thread detail
  function navigateToThread(threadId: string) {
    window.location.href = `/thread/${threadId}`;
  }
  
  // Process media with proper error handling and format normalization
  function processThreadMedia(threads: ThreadWithMedia[]) {
    if (!threads || !Array.isArray(threads)) {
      logger.warn('Invalid threads data:', threads);
      return [];
    }
    
    logger.debug(`Processing ${threads.length} threads for media`);
    
    return threads
      .filter(thread => thread && typeof thread === 'object')
      .flatMap(thread => {
        // Check for various media properties
        let threadMedia: MediaItem[] = [];
        
        // Handle standard media array
        if (thread.media && Array.isArray(thread.media)) {
          threadMedia = thread.media;
        }
        // Handle attachments array
        else if (thread.attachments && Array.isArray(thread.attachments)) {
          threadMedia = thread.attachments;
        }
        // Handle separate images/videos arrays
        else {
          // Handle images array
          if (thread.images && Array.isArray(thread.images)) {
            threadMedia = [
              ...threadMedia,
              ...thread.images.map(url => ({ url, type: 'image' }))
            ];
          }
          
          // Handle videos array
          if (thread.videos && Array.isArray(thread.videos)) {
            threadMedia = [
              ...threadMedia,
              ...thread.videos.map(url => ({ url, type: 'video' }))
            ];
          }
        }
        
        // Map each media to a standardized format with thread reference
        return threadMedia
          .map(mediaItem => ({
            threadId: thread.id,
            media: getMediaElement(mediaItem)
          }))
          .filter(item => item.media !== null);
      });
  }
  
  // Get all media items flattened
  $: flattenedMedia = processThreadMedia(media);
  
  onMount(() => {
    logger.debug('ExploreMediaResults component mounted');
    setupIntersectionObserver();
  });
  
  onDestroy(() => {
    if (observer && loadMoreTrigger) {
      observer.unobserve(loadMoreTrigger);
  }
  });
</script>

<div class="p-4">
  <h2 class="font-bold text-xl mb-4">Media</h2>
  
  {#if isLoading && flattenedMedia.length === 0}
    <div class="animate-pulse grid grid-cols-3 gap-2">
      {#each Array(9) as _}
        <div class="aspect-square bg-gray-200 dark:bg-gray-800 rounded-md"></div>
      {/each}
    </div>
  {:else if flattenedMedia.length === 0}
    <div class="text-center py-8">
      <p class="text-gray-500 dark:text-gray-400">No media found matching your search.</p>
    </div>
  {:else}
    <div class="grid grid-cols-3 gap-2">
      {#each flattenedMedia as item}
        <button 
          class="aspect-square bg-gray-100 dark:bg-gray-900 flex items-center justify-center overflow-hidden rounded-md hover:opacity-90 transition-opacity"
          on:click={() => navigateToThread(item.threadId)}
        >
          {#if item.media && item.media.type === 'video'}
            <video src={item.media.url} class="w-full h-full object-cover" />
          {:else if item.media && item.media.type === 'gif'}
            <img src={item.media.url} alt="GIF" class="w-full h-full object-cover" />
          {:else if item.media}
            <img src={item.media.url} alt="Image" class="w-full h-full object-cover" />
          {/if}
        </button>
      {/each}
    </div>
    
    <!-- Infinite scroll load trigger -->
    {#if hasMore}
      <div 
        bind:this={loadMoreTrigger}
        class="w-full h-10 flex items-center justify-center my-4"
      >
        {#if isLoading}
          <div class="loader"></div>
        {:else}
          <div class="w-full h-1"></div>
        {/if}
      </div>
    {/if}
  {/if}
</div>

<style>
  .animate-pulse {
    animation: pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite;
  }
  
  @keyframes pulse {
    0%, 100% { opacity: 0.5; }
    50% { opacity: 1; }
  }
  
  .loader {
    border: 2px solid rgba(0, 0, 0, 0.1);
    border-radius: 50%;
    border-top: 2px solid #3498db;
    width: 20px;
    height: 20px;
    animation: spin 1s linear infinite;
  }
  
  @keyframes spin {
    0% { transform: rotate(0deg); }
    100% { transform: rotate(360deg); }
  }
  
  :global(.dark) .loader {
    border: 2px solid rgba(255, 255, 255, 0.1);
    border-top: 2px solid #3498db;
  }
</style> 