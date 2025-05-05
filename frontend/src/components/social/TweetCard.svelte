<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import type { ITweet } from '../../interfaces/ISocialMedia.d.ts';
  import { toastStore } from '../../stores/toastStore';

  export let tweet: ITweet;
  export let isDarkMode: boolean = false;
  export let isAuthenticated: boolean = false;
  
  // Track interaction states
  export let isLiked: boolean = false;
  export let isReposted: boolean = false;
  export let isBookmarked: boolean = false;
  
  // For display of replied-to tweet
  export let inReplyToTweet: ITweet | null = null;
  // For display of replies to this tweet
  export let replies: ITweet[] = [];
  // Whether to show replies (can be toggled)
  export let showReplies: boolean = false;
  // Level of nesting for replies (0 for main tweets, increases for nested replies)
  export let nestingLevel: number = 0;
  // Maximum allowed nesting level
  const MAX_NESTING_LEVEL = 3;
  
  const dispatch = createEventDispatcher();
  
  // Process user metadata from content field
  // Format: [USER:username@displayName]content
  $: processedTweet = processTweetContent(tweet);
  
  function processTweetContent(originalTweet: ITweet): ITweet {
    // Create a copy of the tweet to avoid mutating the original
    const processedTweet = { ...originalTweet };
    
    // Check if the tweet already has a valid username (not 'anonymous' or 'unknown')
    // If so, we can assume it's already been processed
    if (processedTweet.username && 
        processedTweet.username !== 'anonymous' && 
        processedTweet.username !== 'user' &&
        processedTweet.username !== 'unknown') {
      return processedTweet;
    }
    
    if (typeof processedTweet.content === 'string') {
      const userMetadataRegex = /^\[USER:([^@\]]+)(?:@([^\]]+))?\](.*)/;
      const match = processedTweet.content.match(userMetadataRegex);
      
      if (match) {
        // Extract username and displayName from the content
        processedTweet.username = match[1] || processedTweet.username;
        processedTweet.displayName = match[2] || processedTweet.displayName;
        
        // Set the actual content without the metadata
        processedTweet.content = match[3] || '';
      }
    }
    
    return processedTweet;
  }
  
  function formatTimestamp(timestamp: string): string {
    try {
      const date = new Date(timestamp);
      
      // Check if date is valid
      if (isNaN(date.getTime())) {
        return 'now';
      }
      
      const now = new Date();
      const seconds = Math.floor((now.getTime() - date.getTime()) / 1000);
      
      if (seconds < 0) return 'now'; // Future dates or clock skew
      
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
    } catch (error) {
      console.error('Error formatting timestamp:', error);
      return 'now';
    }
  }

  // Handle action clicks with authentication check
  function handleReply() {
    if (!isAuthenticated) {
      toastStore.showToast('Please log in to reply to posts', 'info');
      return;
    }
    dispatch('reply', processedTweet.id);
  }

  function handleRetweet() {
    if (!isAuthenticated) {
      toastStore.showToast('Please log in to repost', 'info');
      return;
    }
    dispatch('repost', processedTweet.id);
    isReposted = !isReposted; // Toggle state locally
  }

  function handleLike() {
    if (!isAuthenticated) {
      toastStore.showToast('Please log in to like posts', 'info');
      return;
    }
    
    // Dispatch specific events based on current state (like or unlike)
    if (isLiked) {
      dispatch('unlike', processedTweet.id);
    } else {
      dispatch('like', processedTweet.id);
    }
    
    // Don't toggle state locally - let the parent component update based on API response
    // isLiked = !isLiked;
  }

  function handleBookmark() {
    if (!isAuthenticated) {
      toastStore.showToast('Please log in to bookmark posts', 'info');
      return;
    }
    
    // Log current state for debugging
    console.log(`Bookmark action on tweet ${processedTweet.id}. Current bookmark state: ${isBookmarked}`);
    
    // Dispatch different events based on current bookmark state
    if (isBookmarked) {
      console.log(`Removing bookmark for tweet ${processedTweet.id}`);
      dispatch('removeBookmark', processedTweet.id);
    } else {
      console.log(`Adding bookmark for tweet ${processedTweet.id}`);
      dispatch('bookmark', processedTweet.id);
    }
    
    // Don't toggle state locally - let the parent component update state based on API response
    // isBookmarked = !isBookmarked;
  }

  function toggleReplies() {
    showReplies = !showReplies;
    if (showReplies && replies.length === 0) {
      dispatch('loadReplies', processedTweet.id);
    }
  }

  function handleShare() {
    dispatch('share', processedTweet);
  }

  function handleClick() {
    dispatch('click', processedTweet);
  }

  // Handle reply to a reply
  function handleNestedReply(event) {
    // Forward the event to parent with the reply ID
    dispatch('reply', event.detail);
  }

  // Handle loading replies for nested tweets
  function handleLoadNestedReplies(event) {
    // Forward the loadReplies event to parent
    dispatch('loadReplies', event.detail);
  }

  // Handle nested like/unlike
  function handleNestedLike(event) {
    // Forward the like/unlike event to parent
    if (event.type === 'unlike') {
      dispatch('unlike', event.detail);
    } else {
      dispatch('like', event.detail);
    }
  }

  // Handle nested bookmark events
  function handleNestedBookmark(event) {
    // Forward events to parent
    if (event.type === 'removeBookmark') {
      dispatch('removeBookmark', event.detail);
    } else {
      dispatch('bookmark', event.detail);
    }
  }

  function handleNestedRepost(event) {
    dispatch('repost', event.detail);
  }
</script>

<div 
  class="tweet-card {isDarkMode ? 'tweet-card-dark' : ''} {nestingLevel > 0 ? 'nested-tweet' : 'main-tweet'} border-b {isDarkMode ? 'border-gray-800' : 'border-gray-200'} hover:bg-opacity-50 {isDarkMode ? 'hover:bg-gray-800 bg-gray-900 text-white' : 'hover:bg-gray-50 bg-white text-black'} transition-colors cursor-pointer"
  style="margin-left: {nestingLevel * 12}px;"
  on:click={handleClick}
  on:keydown={(e) => e.key === 'Enter' && handleClick()}
  role="button"
  tabindex="0"
>
  <!-- If this is a reply, show the replied-to tweet -->
  {#if inReplyToTweet && nestingLevel === 0}
    <div class="reply-context px-4 pt-2 pb-0">
      <div class="flex items-center text-sm {isDarkMode ? 'text-gray-400' : 'text-gray-500'}">
        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 10h10a8 8 0 018 8v2M3 10l6 6m-6-6l6-6" />
        </svg>
        <span>Replying to <span class="text-blue-500">@{inReplyToTweet.username}</span></span>
      </div>
      <div class="ml-5 pl-4 border-l {isDarkMode ? 'border-gray-700' : 'border-gray-200'} my-1">
        <div class="text-sm {isDarkMode ? 'text-gray-300' : 'text-gray-600'} line-clamp-1">
          {inReplyToTweet.content}
        </div>
      </div>
    </div>
  {/if}

  <div class="tweet-header p-4">
    <div class="flex items-start">
      <div class="tweet-avatar-container w-12 h-12 rounded-full overflow-hidden {isDarkMode ? 'bg-gray-700' : 'bg-gray-200'} flex items-center justify-center mr-3">
        {#if typeof processedTweet.avatar === 'string' && processedTweet.avatar.startsWith('http')}
          <img src={processedTweet.avatar} alt={processedTweet.username} class="w-full h-full object-cover" />
        {:else}
          <div class="text-xl {isDarkMode ? 'text-gray-100' : ''}">{processedTweet.avatar}</div>
        {/if}
      </div>
      
      <div class="flex-1 min-w-0">
        <div class="flex items-center">
          <span class="font-bold {isDarkMode ? 'text-white' : 'text-black'} mr-1.5">{processedTweet.displayName || 'User'}</span>
          <span class="{isDarkMode ? 'text-gray-400' : 'text-gray-500'} text-sm truncate">@{processedTweet.username || 'user'}</span>
          <span class="{isDarkMode ? 'text-gray-400' : 'text-gray-500'} mx-1.5">·</span>
          <span class="{isDarkMode ? 'text-gray-400' : 'text-gray-500'} text-sm">{formatTimestamp(processedTweet.timestamp)}</span>
        </div>
        
        <div class="tweet-content my-2 {isDarkMode ? 'text-gray-100' : 'text-black'}">
          <p>{processedTweet.content || ''}</p>
        </div>
        
        {#if processedTweet.media && processedTweet.media.length > 0}
          <div class="media-container mt-2 rounded-xl overflow-hidden {isDarkMode ? 'border border-gray-700' : ''}">
            {#if processedTweet.media.length === 1}
              <div class="single-media h-64 w-full">
                {#if processedTweet.media[0].type === 'Image'}
                  <img src={processedTweet.media[0].url} alt="Media" class="h-full w-full object-cover" />
                {:else if processedTweet.media[0].type === 'Video'}
                  <video src={processedTweet.media[0].url} controls class="h-full w-full object-contain">
                    <track kind="captions" src="/captions/en.vtt" srclang="en" label="English" />
                  </video>
                {:else}
                  <img src={processedTweet.media[0].url} alt="GIF" class="h-full w-full object-cover" />
                {/if}
              </div>
            {:else if processedTweet.media.length > 1}
              <div class="media-grid grid gap-1" style="grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));">
                {#each processedTweet.media.slice(0, 4) as media, index}
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
        
        <!-- Action buttons -->
        <div class="flex justify-between mt-3 {isDarkMode ? 'text-gray-400' : 'text-gray-500'}">
          <div class="flex items-center">
            <button class="tweet-action-btn flex items-center rounded-full p-2 transition-colors {isDarkMode ? 'dark-btn hover:bg-blue-900/30' : 'light-btn hover:bg-blue-100'} hover:text-blue-500" on:click|stopPropagation={handleReply}>
              <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
              </svg>
              <span>{isNaN(processedTweet.replies) ? 0 : processedTweet.replies}</span>
            </button>
          </div>
          <div class="flex items-center">
            <button 
              class="tweet-action-btn flex items-center rounded-full p-2 transition-colors {isDarkMode ? 'dark-btn hover:bg-green-900/30' : 'light-btn hover:bg-green-100'} {isReposted ? 'text-green-500' : ''} hover:text-green-500" 
              on:click|stopPropagation={handleRetweet}
            >
              <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 mr-1" fill={isReposted ? "currentColor" : "none"} viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
              </svg>
              <span>{isNaN(processedTweet.reposts) ? 0 : processedTweet.reposts}</span>
            </button>
          </div>
          <div class="flex items-center">
            <button 
              class="tweet-action-btn flex items-center rounded-full p-2 transition-colors {isDarkMode ? 'dark-btn hover:bg-red-900/30' : 'light-btn hover:bg-red-100'} {isLiked ? 'text-red-500' : ''} hover:text-red-500" 
              on:click|stopPropagation={handleLike}
            >
              <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 mr-1" fill={isLiked ? "currentColor" : "none"} viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z" />
              </svg>
              <span>{isNaN(processedTweet.likes) ? 0 : processedTweet.likes}</span>
            </button>
          </div>
          <div class="flex items-center">
            <button 
              class="tweet-action-btn flex items-center rounded-full p-2 transition-colors {isDarkMode ? 'dark-btn hover:bg-blue-900/30' : 'light-btn hover:bg-blue-100'} {isBookmarked ? 'text-blue-500' : ''} hover:text-blue-500" 
              on:click|stopPropagation={handleBookmark}
            >
              <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 mr-1" fill={isBookmarked ? "currentColor" : "none"} viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 5a2 2 0 012-2h10a2 2 0 012 2v16l-7-3.5L5 21V5z" />
              </svg>
              <!-- Show bookmark count -->
              <span>{isNaN(processedTweet.bookmarks) ? 0 : processedTweet.bookmarks}</span>
            </button>
          </div>
          <div class="flex items-center">
            <div class="flex items-center p-2 rounded-full transition-colors {isDarkMode ? 'dark-btn hover:bg-gray-700' : 'light-btn hover:bg-gray-100'}">
              <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
              </svg>
              <span>{typeof processedTweet.views === 'string' ? processedTweet.views : '0'}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</div>

<!-- Show replies button (only for main tweets and if replies exist or can be loaded) -->
{#if nestingLevel === 0 && (replies.length > 0 || processedTweet.replies > 0)}
  <div class="ml-12 mt-1 mb-2">
    <button 
      class="text-sm flex items-center p-1.5 rounded-full {isDarkMode ? 'dark-btn text-gray-400 hover:text-blue-400 hover:bg-blue-900/20' : 'light-btn text-gray-500 hover:text-blue-500 hover:bg-blue-100'}"
      on:click|stopPropagation={toggleReplies}
    >
      <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        {#if showReplies}
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 15l7-7 7 7" />
        {:else}
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
        {/if}
      </svg>
      {#if replies.length > 0}
        {showReplies ? 'Hide' : 'Show'} {replies.length} {replies.length === 1 ? 'reply' : 'replies'}
      {:else}
        {showReplies ? 'Hide replies' : 'Show replies'}
      {/if}
    </button>
  </div>
{/if}

<!-- Replies section -->
{#if showReplies}
  <div class="replies-container {isDarkMode ? 'bg-gray-900' : 'bg-white'} ml-12 border-l {isDarkMode ? 'border-gray-700' : 'border-gray-200'} pl-4">
    {#if replies.length === 0}
      <!-- Loading state -->
      <div class="py-4 text-center {isDarkMode ? 'text-gray-400' : 'text-gray-500'}">
        <div class="animate-pulse">Loading replies...</div>
      </div>
    {:else}
      <!-- Display replies -->
      {#each replies as reply (reply.id)}
        <!-- Render each reply as a nested TweetCard if not exceeding max nesting level -->
        {#if nestingLevel < MAX_NESTING_LEVEL}
          <svelte:self 
            tweet={reply}
            {isDarkMode}
            {isAuthenticated}
            isLiked={reply.isLiked || false}
            isReposted={reply.isReposted || false}
            isBookmarked={reply.isBookmarked || false}
            inReplyToTweet={null}
            replies={[]} 
            showReplies={false}
            nestingLevel={nestingLevel + 1}
            on:reply={handleNestedReply}
            on:like={handleNestedLike}
            on:unlike={handleNestedLike}
            on:repost={handleNestedRepost}
            on:bookmark={handleNestedBookmark}
            on:removeBookmark={handleNestedBookmark}
            on:loadReplies={handleLoadNestedReplies}
          />
        {:else}
          <!-- Simple reply rendering for max nesting level -->
          <div class="reply-item py-3 {isDarkMode ? 'border-b border-gray-800' : 'border-b border-gray-200'}">
            <div class="flex">
              <div class="w-10 h-10 rounded-full overflow-hidden {isDarkMode ? 'bg-gray-700' : 'bg-gray-200'} flex items-center justify-center mr-3 flex-shrink-0">
                {#if typeof reply.avatar === 'string' && reply.avatar.startsWith('http')}
                  <img src={reply.avatar} alt={reply.username} class="w-full h-full object-cover" />
                {:else}
                  <div class="text-lg {isDarkMode ? 'text-gray-100' : ''}">{reply.avatar}</div>
                {/if}
              </div>
              <div class="flex-1">
                <div class="flex items-center">
                  <span class="font-bold {isDarkMode ? 'text-white' : 'text-black'} mr-1.5">{reply.displayName || 'User'}</span>
                  <span class="{isDarkMode ? 'text-gray-400' : 'text-gray-500'} text-sm truncate">@{reply.username || 'user'}</span>
                  <span class="{isDarkMode ? 'text-gray-400' : 'text-gray-500'} mx-1.5">·</span>
                  <span class="{isDarkMode ? 'text-gray-400' : 'text-gray-500'} text-sm">{formatTimestamp(reply.timestamp)}</span>
                </div>
                <div class="my-2 {isDarkMode ? 'text-gray-100' : 'text-black'}">
                  <p>{reply.content || ''}</p>
                </div>
                <div class="flex text-sm {isDarkMode ? 'text-gray-400' : 'text-gray-500'} mt-2">
                  <button class="flex items-center mr-4 hover:text-blue-500 p-1 rounded-full {isDarkMode ? 'dark-btn hover:bg-blue-900/30' : 'light-btn hover:bg-blue-100'}" on:click|stopPropagation={() => dispatch('reply', reply.id)}>
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
                    </svg>
                    {isNaN(reply.replies) ? 0 : reply.replies}
                  </button>
                  <button class="flex items-center mr-4 hover:text-red-500 p-1 rounded-full {isDarkMode ? 'dark-btn hover:bg-red-900/30' : 'light-btn hover:bg-red-100'}" on:click|stopPropagation={() => dispatch('like', reply.id)}>
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-1" fill={reply.isLiked ? "currentColor" : "none"} viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z" />
                    </svg>
                    {isNaN(reply.likes) ? 0 : reply.likes}
                  </button>
                  <button class="flex items-center mr-4 hover:text-blue-500 p-1 rounded-full {isDarkMode ? 'dark-btn hover:bg-blue-900/30' : 'light-btn hover:bg-blue-100'}" on:click|stopPropagation={() => dispatch('bookmark', reply.id)}>
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-1" fill={reply.isBookmarked ? "currentColor" : "none"} viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 5a2 2 0 012-2h10a2 2 0 012 2v16l-7-3.5L5 21V5z" />
                    </svg>
                    {isNaN(reply.bookmarks) ? 0 : reply.bookmarks}
                  </button>
                </div>
              </div>
            </div>
          </div>
        {/if}
      {/each}
      {#if replies.length > 0 && nestingLevel === 0}
        <div class="text-center py-2">
          <button class="text-blue-500 text-sm hover:underline px-3 py-1.5 rounded-full {isDarkMode ? 'dark-btn hover:bg-blue-900/20' : 'light-btn hover:bg-blue-50'}">
            Show more replies
          </button>
        </div>
      {/if}
    {/if}
  </div>
{/if}

<!-- Add "View replies" message when replies exist but aren't expanded -->
{#if nestingLevel === 0 && processedTweet.replies > 0 && !showReplies}
  <div class="mt-2 mb-0 ml-14">
    <button 
      class="text-sm {isDarkMode ? 'text-blue-400 hover:text-blue-300' : 'text-blue-500 hover:text-blue-600'}"
      on:click|stopPropagation={toggleReplies}
    >
      View {processedTweet.replies} {processedTweet.replies === 1 ? 'reply' : 'replies'}
    </button>
  </div>
{/if}

<style>
  .tweet-card {
    padding: 0.5rem 0;
  }
  
  .tweet-card-dark {
    background-color: #1a202c; /* Match with gray-900 */
  }

  .nested-tweet {
    padding-left: 0.5rem;
    border-radius: 0.5rem;
    margin-top: 0.5rem;
    margin-bottom: 0.5rem;
  }

  .tweet-avatar-container {
    flex-shrink: 0;
  }
  
  .tweet-action-btn {
    transition: all 0.2s;
  }
  
  /* Button styles for dark and light mode */
  .dark-btn {
    background-color: transparent;
  }
  
  .dark-btn:hover {
    background-color: rgba(59, 130, 246, 0.2);
  }
  
  .light-btn {
    background-color: transparent;
  }
  
  /* Media grid styling */
  .media-grid {
    max-height: 300px;
    overflow: hidden;
  }
  
  /* For truncating long text in reply context */
  .line-clamp-1 {
    display: -webkit-box;
    -webkit-line-clamp: 1;
    line-clamp: 1;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }
  
  /* Animation for loading indicator */
  @keyframes pulse {
    0%, 100% { opacity: 1; }
    50% { opacity: 0.5; }
  }
  .animate-pulse {
    animation: pulse 1.5s cubic-bezier(0.4, 0, 0.6, 1) infinite;
  }
</style>