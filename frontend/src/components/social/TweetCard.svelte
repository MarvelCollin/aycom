<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import type { ITweet } from '../../interfaces/ISocialMedia';

  export let tweet: ITweet;
  export let isDarkMode: boolean = false;
  
  const dispatch = createEventDispatcher();
  
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
    dispatch('reply', tweet.id);
  }

  function handleRetweet() {
    dispatch('repost', tweet.id);
  }

  function handleLike() {
    dispatch('like', tweet.id);
  }

  function handleShare() {
    dispatch('share', tweet);
  }

  function handleClick() {
    dispatch('click', tweet);
  }
</script>

<div 
  class="tweet-card {isDarkMode ? 'tweet-card-dark' : ''} border-b {isDarkMode ? 'border-gray-800' : 'border-gray-200'} hover:bg-opacity-50 {isDarkMode ? 'hover:bg-gray-800 bg-gray-900 text-white' : 'hover:bg-gray-50 bg-white text-black'} transition-colors cursor-pointer"
  on:click={handleClick}
  on:keydown={(e) => e.key === 'Enter' && handleClick()}
  role="button"
  tabindex="0"
>
  <div class="tweet-header p-4">
    <div class="flex items-start">
      <div class="tweet-avatar-container w-12 h-12 rounded-full overflow-hidden {isDarkMode ? 'bg-gray-700' : 'bg-gray-200'} flex items-center justify-center mr-3">
        {#if typeof tweet.avatar === 'string' && tweet.avatar.startsWith('http')}
          <img src={tweet.avatar} alt={tweet.username} class="w-full h-full object-cover" />
        {:else}
          <div class="text-xl {isDarkMode ? 'text-gray-100' : ''}">{tweet.avatar}</div>
        {/if}
      </div>
      
      <div class="flex-1 min-w-0">
        <div class="flex items-center">
          <span class="font-bold {isDarkMode ? 'text-white' : 'text-black'} mr-1.5">{tweet.displayName}</span>
          <span class="{isDarkMode ? 'text-gray-400' : 'text-gray-500'} text-sm truncate">@{tweet.username}</span>
          <span class="{isDarkMode ? 'text-gray-400' : 'text-gray-500'} mx-1.5">·</span>
          <span class="{isDarkMode ? 'text-gray-400' : 'text-gray-500'} text-sm">{formatTimestamp(tweet.timestamp)}</span>
        </div>
        
        <div class="tweet-content my-2 {isDarkMode ? 'text-gray-100' : 'text-black'}">
          <p>{tweet.content}</p>
        </div>
        
        {#if tweet.media && tweet.media.length > 0}
          <div class="media-container mt-2 rounded-xl overflow-hidden {isDarkMode ? 'border border-gray-700' : ''}">
            {#if tweet.media.length === 1}
              <div class="single-media h-64 w-full">
                {#if tweet.media[0].type === 'Image'}
                  <img src={tweet.media[0].url} alt="Media" class="h-full w-full object-cover" />
                {:else if tweet.media[0].type === 'Video'}
                  <video src={tweet.media[0].url} controls class="h-full w-full object-contain">
                    <track kind="captions" src="/captions/en.vtt" srclang="en" label="English" />
                  </video>
                {:else}
                  <img src={tweet.media[0].url} alt="GIF" class="h-full w-full object-cover" />
                {/if}
              </div>
            {:else if tweet.media.length > 1}
              <div class="media-grid grid gap-1" style="grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));">
                {#each tweet.media.slice(0, 4) as media, index}
                  <div class="media-item h-40">
                    {#if media.type === 'Image'}
                      <img src={media.url} alt="Media" class="h-full w-full object-cover" />
                    {:else if media.type === 'Video'}
                      <video src={media.url} class="h-full w-full object-cover">
                        <track kind="captions" src="/captions/en.vtt" srclang="en" label="English" />
                      </video>
                    {:else}
                      <img src={media.url} alt="GIF" class="h-full w-full object-cover" />
                    {/if}
                  </div>
                {/each}
              </div>
            {/if}
          </div>
        {/if}
        
        <div class="flex justify-between mt-3 {isDarkMode ? 'text-gray-400' : 'text-gray-500'}">
          <div class="flex items-center">
            <button class="tweet-action-btn flex items-center {isDarkMode ? 'bg-black hover:bg-blue-500/10' : 'hover:bg-blue-100'} hover:text-blue-500 rounded-full p-2 transition-all" on:click|stopPropagation={handleReply}>
              <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
              </svg>
              <span>{tweet.replies}</span>
            </button>
          </div>
          <div class="flex items-center">
            <button class="tweet-action-btn flex items-center {isDarkMode ? 'bg-black hover:bg-green-500/10' : 'hover:bg-green-100'} hover:text-green-500 rounded-full p-2 transition-all" on:click|stopPropagation={handleRetweet}>
              <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
              </svg>
              <span>{tweet.reposts}</span>
            </button>
          </div>
          <div class="flex items-center">
            <button class="tweet-action-btn flex items-center {isDarkMode ? 'bg-black hover:bg-red-500/10' : 'hover:bg-red-100'} hover:text-red-500 rounded-full p-2 transition-all" on:click|stopPropagation={handleLike}>
              <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z" />
              </svg>
              <span>{tweet.likes}</span>
            </button>
          </div>
          <div class="flex items-center">
            <div class="flex items-center p-2 rounded-full {isDarkMode ? 'bg-black hover:bg-gray-700' : 'hover:bg-gray-100'} transition-all">
              <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
              </svg>
              <span>{tweet.views}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</div>

<style>
  .tweet-card {
    padding: 0.5rem 0;
  }
  
  .tweet-card-dark {
    background-color: #1a202c; /* Match with gray-900 */
  }

  .tweet-avatar-container {
    flex-shrink: 0;
  }
  
  .tweet-action-btn {
    transition: all 0.2s;
  }
  
  /* Media grid styling */
  .media-grid {
    max-height: 300px;
    overflow: hidden;
  }
</style>