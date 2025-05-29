<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { useTheme } from '../../hooks/useTheme';
  
  const dispatch = createEventDispatcher();
  const { theme } = useTheme();
  
  // Reactive declarations
  $: isDarkMode = $theme === 'dark';
  
  // Props
  export let community: {
    id: string;
    name: string;
    description: string;
    logo: string | null;
    memberCount: number;
    isJoined: boolean;
    isPending: boolean;
  };
  
  // Handle join request
  function handleJoinRequest() {
    dispatch('joinRequest', community.id);
  }
</script>

<div class="py-4">
  <div class="flex items-start">
    <a href={`/community/${community.id}`} class="flex-shrink-0">
      <div class="w-16 h-16 rounded-md {isDarkMode ? 'bg-gray-700' : 'bg-gray-200'} flex items-center justify-center overflow-hidden mr-3">
        {#if typeof community.logo === 'string' && community.logo.startsWith('http')}
          <img src={community.logo} alt={community.name} class="w-full h-full object-cover" />
        {:else}
          <span class="text-2xl">👥</span>
        {/if}
      </div>
    </a>
    <div class="flex-1 min-w-0">
      <a href={`/community/${community.id}`} class="block">
        <h3 class="font-bold {isDarkMode ? 'text-white' : 'text-black'}">{community.name}</h3>
        <p class="text-sm text-gray-500 dark:text-gray-400 mt-1 line-clamp-2">{community.description}</p>
        <p class="text-xs text-gray-500 dark:text-gray-400 mt-2">
          <span class="font-semibold">{community.memberCount}</span> {community.memberCount === 1 ? 'member' : 'members'}
        </p>
      </a>
    </div>
    <div class="ml-4 flex-shrink-0">
      {#if community.isJoined}
        <span class="px-4 py-1.5 rounded-full bg-transparent border border-gray-300 dark:border-gray-600 font-bold text-sm flex items-center">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-1 text-green-500" viewBox="0 0 20 20" fill="currentColor">
            <path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd" />
          </svg>
          Joined
        </span>
      {:else if community.isPending}
        <span class="px-4 py-1.5 rounded-full bg-transparent border border-gray-300 dark:border-gray-600 font-bold text-sm flex items-center">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-1 text-yellow-500" viewBox="0 0 20 20" fill="currentColor">
            <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm1-12a1 1 0 10-2 0v4a1 1 0 00.293.707l2.828 2.829a1 1 0 101.415-1.415L11 9.586V6z" clip-rule="evenodd" />
          </svg>
          Pending
        </span>
      {:else}
        <button 
          class="px-4 py-1.5 rounded-full bg-blue-500 text-white font-bold text-sm hover:bg-blue-600"
          on:click={handleJoinRequest}
        >
          Request to Join
        </button>
      {/if}
    </div>
  </div>
</div>

<style>  /* Line clamp for truncating text */
  .line-clamp-2 {
    display: -webkit-box;
    -webkit-line-clamp: 2;
    line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
    /* Fallback for browsers that don't support line-clamp */
    line-height: 1.4;
    max-height: calc(1.4em * 2);
  }
</style> 