<script lang="ts">
  import type { ITweet } from '../../interfaces/ISocialMedia';
  import { 
    MessageSquareIcon, 
    RefreshCwIcon,
    HeartIcon,
    EyeIcon,
    ShareIcon,
    BookmarkIcon
  } from 'svelte-feather-icons';
  import { createEventDispatcher } from 'svelte';
  
  export let tweet: ITweet;
  export let isDarkMode = false;
  
  const dispatch = createEventDispatcher();
  
  function toggleBookmark() {
    dispatch('toggleBookmark', { tweetId: tweet.id });
  }

  function likeTweet() {
    dispatch('likeTweet', { tweetId: tweet.id });
  }

  function repostTweet() {
    dispatch('repostTweet', { tweetId: tweet.id });
  }
</script>

<article class="p-4 border-b {isDarkMode ? 'border-gray-800 border-opacity-80' : 'border-gray-200'} hover:{isDarkMode ? 'bg-gray-800/30' : 'bg-gray-100/30'} transition-colors cursor-pointer">
  <div class="flex">
    <div class="flex-shrink-0">
      <div class="w-12 h-12 rounded-full overflow-hidden flex items-center justify-center {isDarkMode ? 'bg-gray-700' : 'bg-gray-300'}">
        <span>{tweet.avatar}</span>
      </div>
    </div>
    <div class="ml-3 flex-1 overflow-hidden">
      <div class="flex items-center gap-1">
        <span class="font-bold hover:underline">{tweet.displayName}</span>
        <span class="{isDarkMode ? 'text-gray-400' : 'text-gray-500'} text-sm">@{tweet.username}</span>
        <span class="{isDarkMode ? 'text-gray-400' : 'text-gray-500'} text-sm">·</span>
        <span class="{isDarkMode ? 'text-gray-400' : 'text-gray-500'} text-sm hover:underline">{tweet.timestamp}</span>
      </div>
      
      <p class="mt-1 mb-2 whitespace-pre-wrap break-words leading-normal">{tweet.content}</p>
      
      <div class="flex justify-between mt-3 {isDarkMode ? 'text-gray-400' : 'text-gray-500'} max-w-md pr-2">
        <button class="group flex items-center hover:text-blue-500 transition-colors">
          <div class="p-2 rounded-full group-hover:bg-blue-500/10 transition-colors">
            <MessageSquareIcon size="16" />
          </div>
          <span class="ml-1 text-xs group-hover:text-blue-500">{tweet.replies}</span>
        </button>
        <button class="group flex items-center hover:text-green-500 transition-colors" on:click={repostTweet}>
          <div class="p-2 rounded-full group-hover:bg-green-500/10 transition-colors">
            <RefreshCwIcon size="16" color={tweet.reposted ? "rgb(34, 197, 94)" : "currentColor"} />
          </div>
          <span class="ml-1 text-xs {tweet.reposted ? 'text-green-500' : 'group-hover:text-green-500'}">{tweet.reposts}</span>
        </button>
        <button class="group flex items-center hover:text-pink-500 transition-colors" on:click={likeTweet}>
          <div class="p-2 rounded-full group-hover:bg-pink-500/10 transition-colors">
            <HeartIcon size="16" color={tweet.liked ? "rgb(244, 114, 182)" : "currentColor"} />
          </div>
          <span class="ml-1 text-xs {tweet.liked ? 'text-pink-500' : 'group-hover:text-pink-500'}">{tweet.likes}</span>
        </button>
        <button class="group flex items-center hover:text-blue-500 transition-colors">
          <div class="p-2 rounded-full group-hover:bg-blue-500/10 transition-colors">
            <EyeIcon size="16" />
          </div>
          <span class="ml-1 text-xs group-hover:text-blue-500">{tweet.views}</span>
        </button>
        <button class="group flex items-center hover:text-blue-500 transition-colors" on:click={toggleBookmark}>
          <div class="p-2 rounded-full group-hover:bg-blue-500/10 transition-colors">
            <BookmarkIcon size="16" color={tweet.bookmarked ? "rgb(59, 130, 246)" : "currentColor"} />
          </div>
        </button>
        <button class="group flex items-center hover:text-blue-500 transition-colors">
          <div class="p-2 rounded-full group-hover:bg-blue-500/10 transition-colors">
            <ShareIcon size="16" />
          </div>
        </button>
      </div>
    </div>
  </div>
</article> 