<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import ProfileCard from './ProfileCard.svelte';
  import ThreadCard from './ThreadCard.svelte';
  import type { IMedia } from '../../interfaces/ISocialMedia';
  
  const dispatch = createEventDispatcher();
  
  // Props with proper type definitions
  export let topProfiles: Array<{
    id: string;
    username: string;
    displayName: string;
    avatar: string | null;
    bio?: string;
    isVerified: boolean;
    followerCount: number;
    isFollowing: boolean;
  }> = [];
  
  export let topThreads: Array<{
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
  }> = [];
  
  export let isLoading = false;
  
  // Handle follow user
  function handleFollow(event) {
    dispatch('follow', event.detail);
  }
</script>

<div class="p-4">
  {#if isLoading}
    <div class="animate-pulse space-y-4">
      {#each Array(5) as _}
        <div class="flex space-x-4">
          <div class="rounded-full bg-gray-300 dark:bg-gray-700 h-10 w-10"></div>
          <div class="flex-1 space-y-2 py-1">
            <div class="h-4 bg-gray-300 dark:bg-gray-700 rounded w-3/4"></div>
            <div class="space-y-2">
              <div class="h-4 bg-gray-300 dark:bg-gray-700 rounded"></div>
              <div class="h-4 bg-gray-300 dark:bg-gray-700 rounded w-5/6"></div>
            </div>
          </div>
        </div>
      {/each}
    </div>
  {:else}
    <!-- Top Profiles Section -->
    {#if topProfiles.length > 0}
      <div class="mb-6">
        <div class="flex justify-between items-center mb-3">
          <h3 class="font-bold text-lg">People</h3>
          <button class="text-blue-500 text-sm" on:click={() => dispatch('viewAll', 'people')}>
            View all
          </button>
        </div>
        
        <div class="space-y-4">
          {#each topProfiles as profile}
            <ProfileCard 
              {profile} 
              on:follow={handleFollow}
              compact={true}
            />
          {/each}
        </div>
      </div>
    {/if}
    
    <!-- Top Threads Section -->
    {#if topThreads.length > 0}
      <div>
        <h3 class="font-bold text-lg mb-3">Threads</h3>
        <div class="divide-y divide-gray-200 dark:divide-gray-800">
          {#each topThreads as thread}
            <ThreadCard {thread} />
          {/each}
        </div>
      </div>
    {:else if topProfiles.length === 0}
      <div class="text-center py-10">
        <p class="text-gray-500 dark:text-gray-400">No results found</p>
      </div>
    {/if}
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
</style> 