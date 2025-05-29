<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import MessageCircleIcon from 'svelte-feather-icons/src/icons/MessageCircleIcon.svelte';
  import HeartIcon from 'svelte-feather-icons/src/icons/HeartIcon.svelte';
  import RefreshCwIcon from 'svelte-feather-icons/src/icons/RefreshCwIcon.svelte';
  
  export let thread: any;
  export let isDarkMode = false;
  
  // Format timestamp function for tweet display
  function formatTime(timestamp: string): string {
    try {
      const date = new Date(timestamp);
      const now = new Date();
      const diffMs = now.getTime() - date.getTime();
      const diffSec = Math.floor(diffMs / 1000);
      const diffMin = Math.floor(diffSec / 60);
      const diffHours = Math.floor(diffMin / 60);
      const diffDays = Math.floor(diffHours / 24);
      
      if (diffSec < 60) return `${diffSec}s ago`;
      if (diffMin < 60) return `${diffMin}m ago`;
      if (diffHours < 24) return `${diffHours}h ago`;
      if (diffDays < 7) return `${diffDays}d ago`;
      
      return date.toLocaleDateString();
    } catch (error) {
      console.error('Error formatting time:', error);
      return 'some time ago';
    }
  }
  
  const dispatch = createEventDispatcher();
  
  function handleClick() {
    dispatch('click', thread);
  }
</script>

<div class="thread-card-container">
  <button 
    class="thread-card-button p-4 w-full text-left border-b dark:border-gray-800 hover:bg-gray-50 dark:hover:bg-gray-900 transition cursor-pointer"
    on:click={handleClick}
    aria-label="View thread by {thread.name || 'User'}"
  >
    <article class="thread-content">
      <div class="flex">
        <div class="flex-shrink-0 mr-3">
          <div class="w-10 h-10 rounded-full bg-gray-200 dark:bg-gray-700 overflow-hidden flex items-center justify-center">
            {#if thread.profile_picture_url}
              <img src={thread.profile_picture_url} alt={thread.name} class="w-full h-full object-cover" />
            {:else}
              <div class="text-lg font-bold text-gray-500">{(thread.name || 'User').charAt(0).toUpperCase()}</div>
            {/if}
          </div>
        </div>
        
        <div class="flex-1 min-w-0">
          <div class="flex items-center">
            <p class="font-bold {isDarkMode ? 'text-white' : 'text-black'} mr-1">{thread.name || 'User'}</p>
            <p class="text-gray-500 dark:text-gray-400 text-sm truncate">@{thread.username || 'user'}</p>
            <span class="mx-1 text-gray-500 dark:text-gray-400">·</span>
            <time datetime={new Date(thread.created_at).toISOString()} class="text-gray-500 dark:text-gray-400 text-sm">
              {formatTime(thread.created_at)}
            </time>
          </div>
        
          <p class="mt-1 mb-2 {isDarkMode ? 'text-white' : 'text-black'}">{thread.content}</p>
          
          {#if thread.media && thread.media.length > 0}
            <div class="mb-2 rounded-lg overflow-hidden border dark:border-gray-700">
              <img src={thread.media[0].url} alt="Media attached to thread" class="w-full h-48 object-cover" />
            </div>
          {/if}
          
          <div class="flex mt-2 text-gray-500 dark:text-gray-400 text-sm">
            <div class="flex items-center mr-4">
              <div class="mr-1"><MessageCircleIcon size="16" /></div>
              <span>{thread.replies_count || 0}</span>
            </div>
            <div class="flex items-center mr-4">
              <div class="mr-1"><RefreshCwIcon size="16" /></div>
              <span>{thread.reposts_count || 0}</span>
            </div>
            <div class="flex items-center">
              <div class="mr-1"><HeartIcon size="16" /></div>
              <span>{thread.likes_count || 0}</span>
            </div>
          </div>
        </div>
      </div>
    </article>
  </button>
</div>

<style>
  .thread-card-button {
    background: none;
    border: none;
    font: inherit;
    padding: 0;
    display: block;
    width: 100%;
  }
  
  .thread-card-button {
    padding: 1rem;
  }
  
  .thread-content {
    width: 100%;
  }
</style> 