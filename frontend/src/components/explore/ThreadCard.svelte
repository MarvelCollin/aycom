<script lang="ts">
  import { useTheme } from '../../hooks/useTheme';
  
  const { theme } = useTheme();
  
  // Reactive declarations
  $: isDarkMode = $theme === 'dark';
  
  // Props
  export let thread: {
    id: string;
    content: string;
    username: string;
    displayName: string;
    timestamp: string;
    likes: number;
    replies: number;
    reposts: number;
    media?: Array<{
      type: string;
      url: string;
    }>;
    avatar?: string;
  };
  
  // Format timestamp
  function formatTime(timestamp: string): string {
    try {
      const date = new Date(timestamp);
      const now = new Date();
      const diffMs = now.getTime() - date.getTime();
      const diffSec = Math.floor(diffMs / 1000);
      const diffMin = Math.floor(diffSec / 60);
      const diffHours = Math.floor(diffMin / 60);
      const diffDays = Math.floor(diffHours / 24);
      
      if (diffSec < 60) return `${diffSec}s`;
      if (diffMin < 60) return `${diffMin}m`;
      if (diffHours < 24) return `${diffHours}h`;
      if (diffDays < 7) return `${diffDays}d`;
      
      return date.toLocaleDateString();
    } catch (e) {
      return 'recently';
    }
  }
</script>

<div class="py-3">
  <div class="flex">
    <div class="w-10 h-10 rounded-full {isDarkMode ? 'bg-gray-700' : 'bg-gray-200'} flex items-center justify-center overflow-hidden mr-3">
      {#if thread.avatar && thread.avatar.startsWith('http')}
        <img src={thread.avatar} alt={thread.username} class="w-full h-full object-cover" />
      {:else}
        <span>👤</span>
      {/if}
    </div>
    <div>
      <div class="flex items-center">
        <span class="font-bold {isDarkMode ? 'text-white' : 'text-black'}">{thread.displayName}</span>
        <span class="text-gray-500 dark:text-gray-400 text-sm ml-2">@{thread.username}</span>
        <span class="text-gray-500 dark:text-gray-400 text-sm ml-2">· {formatTime(thread.timestamp)}</span>
      </div>
      <p class="mt-1">{thread.content}</p>
      
      {#if thread.media && thread.media.length > 0}
        <div class="mt-2 rounded-lg overflow-hidden">
          {#if thread.media[0].type === 'image'}
            <img src={thread.media[0].url} alt="Media content" class="w-full h-auto max-h-80 object-cover" />
          {:else if thread.media[0].type === 'video'}
            <video src={thread.media[0].url} controls class="w-full h-auto max-h-80 object-cover"></video>
          {/if}
        </div>
      {/if}
      
      <div class="flex mt-2 text-gray-500 dark:text-gray-400 text-sm">
        <div class="mr-4 flex items-center">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
          </svg>
          {thread.replies}
        </div>
        <div class="mr-4 flex items-center">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
          </svg>
          {thread.reposts}
        </div>
        <div class="flex items-center">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z" />
          </svg>
          {thread.likes}
        </div>
      </div>
    </div>
  </div>
</div> 