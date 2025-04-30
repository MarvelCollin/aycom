<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import type { ITweet } from '../../interfaces/ISocialMedia';

  // Props
  export let tweet: ITweet;
  export let isDarkMode: boolean = false;
  
  // Event dispatcher
  const dispatch = createEventDispatcher();
  
  // Format the timestamp to a relative time string
  function formatTimestamp(timestamp: string): string {
    const date = new Date(timestamp);
    const now = new Date();
    const seconds = Math.floor((now.getTime() - date.getTime()) / 1000);
    
    let interval = seconds / 31536000; // seconds in a year
    if (interval > 1) {
      return Math.floor(interval) + 'y';
    }
    interval = seconds / 2592000; // seconds in a month
    if (interval > 1) {
      return Math.floor(interval) + 'mo';
    }
    interval = seconds / 86400; // seconds in a day
    if (interval > 1) {
      return Math.floor(interval) + 'd';
    }
    interval = seconds / 3600; // seconds in an hour
    if (interval > 1) {
      return Math.floor(interval) + 'h';
    }
    interval = seconds / 60; // seconds in a minute
    if (interval > 1) {
      return Math.floor(interval) + 'm';
    }
    return Math.floor(seconds) + 's';
  }

  // Handle action clicks
  function handleReply() {
    dispatch('reply', tweet);
  }

  function handleRetweet() {
    dispatch('retweet', tweet);
  }

  function handleLike() {
    dispatch('like', tweet);
  }

  function handleShare() {
    dispatch('share', tweet);
  }
</script>

<div class="post {isDarkMode ? 'post-dark' : ''} border-b {isDarkMode ? 'border-gray-800' : 'border-gray-200'} hover:bg-opacity-50 {isDarkMode ? 'hover:bg-gray-900 bg-black text-white' : 'hover:bg-gray-50 bg-white text-black'} transition-colors">
  <div class="post-header">
    <div class="post-avatar-container w-12 h-12 rounded-full overflow-hidden {isDarkMode ? 'bg-gray-700' : 'bg-gray-200'} flex items-center justify-center">
      {#if typeof tweet.avatar === 'string' && tweet.avatar.startsWith('http')}
        <img src={tweet.avatar} alt={tweet.username} class="w-full h-full object-cover" />
      {:else}
        <div class="text-xl">{tweet.avatar}</div>
      {/if}
    </div>
    
    <div class="flex-1 min-w-0">
      <div class="flex items-center">
        <span class="font-bold {isDarkMode ? 'text-white' : 'text-black'} mr-1.5">{tweet.displayName}</span>
        <span class="text-gray-500 text-sm truncate">@{tweet.username}</span>
        <span class="text-gray-500 mx-1.5">·</span>
        <span class="text-gray-500 text-sm">{formatTimestamp(tweet.timestamp)}</span>
      </div>
    </div>
    
    <div class="text-gray-500 hover:text-blue-500 cursor-pointer">
      <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
        <path d="M6 10a2 2 0 11-4 0 2 2 0 014 0zM12 10a2 2 0 11-4 0 2 2 0 014 0zM16 12a2 2 0 100-4 2 2 0 000 4z" />
      </svg>
    </div>
  </div>
  
  <div class="post-content my-2 {isDarkMode ? 'text-white' : 'text-black'}">
    <p>{tweet.content}</p>
  </div>
  
  <div class="post-actions flex justify-between mt-3 {isDarkMode ? 'text-gray-400' : 'text-gray-500'}">
    <button class="post-action-btn flex items-center group" on:click={handleReply}>
      <div class="p-2 rounded-full {isDarkMode ? 'group-hover:bg-blue-900/30 group-hover:text-blue-400' : 'group-hover:bg-blue-50 group-hover:text-blue-500'} transition-colors">
        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
        </svg>
      </div>
      <span class="text-sm ml-1">{tweet.replies}</span>
    </button>
    
    <button class="post-action-btn flex items-center group" on:click={handleRetweet}>
      <div class="p-2 rounded-full {isDarkMode ? 'group-hover:bg-green-900/30 group-hover:text-green-400' : 'group-hover:bg-green-50 group-hover:text-green-500'} transition-colors">
        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4" />
        </svg>
      </div>
      <span class="text-sm ml-1">{tweet.reposts}</span>
    </button>
    
    <button class="post-action-btn flex items-center group" on:click={handleLike}>
      <div class="p-2 rounded-full {isDarkMode ? 'group-hover:bg-red-900/30 group-hover:text-red-400' : 'group-hover:bg-red-50 group-hover:text-red-500'} transition-colors">
        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z" />
        </svg>
      </div>
      <span class="text-sm ml-1">{tweet.likes}</span>
    </button>
    
    <button class="post-action-btn flex items-center group">
      <div class="p-2 rounded-full {isDarkMode ? 'group-hover:bg-blue-900/30 group-hover:text-blue-400' : 'group-hover:bg-blue-50 group-hover:text-blue-500'} transition-colors">
        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
        </svg>
      </div>
      <span class="text-sm ml-1">{tweet.views}</span>
    </button>
    
    <button class="post-action-btn flex items-center" on:click={handleShare}>
      <div class="p-2 rounded-full {isDarkMode ? 'hover:bg-blue-900/30 hover:text-blue-400' : 'hover:bg-blue-50 hover:text-blue-500'} transition-colors">
        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M8.684 13.342C8.886 12.938 9 12.482 9 12c0-.482-.114-.938-.316-1.342m0 2.684a3 3 0 110-2.684m0 2.684l6.632 3.316m-6.632-6l6.632-3.316m0 0a3 3 0 105.367-2.684 3 3 0 00-5.367 2.684zm0 9.316a3 3 0 105.368 2.684 3 3 0 00-5.368-2.684z" />
        </svg>
      </div>
    </button>
  </div>
</div>

<style>
  /* Additional custom styles that complement the components.css file */
  .post-avatar-container {
    flex-shrink: 0;
  }
  
  .post-action-btn {
    transition: color 0.2s;
  }
  
  /* Custom interaction styles for tweet actions */
  .post-actions button:hover {
    color: inherit;
  }
</style>