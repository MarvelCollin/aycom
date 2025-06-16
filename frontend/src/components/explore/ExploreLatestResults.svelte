<script lang="ts">
  import ThreadCard from "./ThreadCard.svelte";
  import { createLoggerWithPrefix } from "../../utils/logger";

  const logger = createLoggerWithPrefix("ExploreLatestResults");

  export let latestThreads: Array<{
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

  $: {
    if (!isLoading) {
      if (latestThreads.length > 0) {
        logger.debug("Latest threads loaded", { count: latestThreads.length });
      } else {
        logger.debug("No latest threads found");
      }
    }
  }
</script>

<div class="divide-y divide-gray-200 dark:divide-gray-800">
  {#if isLoading}
    <div class="animate-pulse space-y-4 p-4">
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
  {:else if latestThreads.length > 0}
    {#each latestThreads as thread}
      <div class="p-4">
        <ThreadCard {thread} />
      </div>
    {/each}
  {:else}
    <div class="text-center py-10">
      <p class="text-gray-500 dark:text-gray-400">No results found</p>
    </div>
  {/if}
</div>

<style>

  @keyframes pulse {
    0%, 100% { opacity: 0.5; }
    50% { opacity: 1; }
  }
  .animate-pulse {
    animation: pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite;
  }
</style>