<script lang="ts">
  import { useTheme } from '../../hooks/useTheme';

  // Props
  export let type: 'search-results' | 'trends' | 'user-profile' | 'media' = 'search-results';
  export let count: number = 5;

  const { theme } = useTheme();
  $: isDarkMode = $theme === 'dark';
</script>

{#if type === 'search-results'}
  <div class="space-y-4">
    {#each Array(count) as _, i}
      <div class="flex items-start p-4 border-b border-gray-200 dark:border-gray-800 animate-pulse">
        <div class="w-12 h-12 rounded-full bg-gray-200 dark:bg-gray-700 mr-3"></div>
        <div class="flex-1">
          <div class="h-4 bg-gray-200 dark:bg-gray-700 rounded w-1/4 mb-2"></div>
          <div class="h-3 bg-gray-200 dark:bg-gray-700 rounded w-1/3 mb-2"></div>
          <div class="h-20 bg-gray-200 dark:bg-gray-700 rounded w-full"></div>
        </div>
      </div>
    {/each}
  </div>
{:else if type === 'trends'}
  <div class="rounded-xl bg-gray-50 dark:bg-gray-900 p-4">
    <div class="flex items-center justify-between pb-3 border-b border-gray-200 dark:border-gray-800">
      <div class="h-5 bg-gray-200 dark:bg-gray-700 rounded w-1/3 animate-pulse"></div>
    </div>
    <div class="divide-y divide-gray-200 dark:divide-gray-800">
      {#each Array(10) as _, i}
        <div class="py-3 animate-pulse">
          <div class="h-4 bg-gray-200 dark:bg-gray-700 rounded w-1/2 mb-2"></div>
          <div class="h-3 bg-gray-200 dark:bg-gray-700 rounded w-1/4"></div>
        </div>
      {/each}
    </div>
  </div>
{:else if type === 'user-profile'}
  <div class="flex items-center p-4 animate-pulse">
    <div class="w-12 h-12 rounded-full bg-gray-200 dark:bg-gray-700 mr-3"></div>
    <div>
      <div class="h-4 bg-gray-200 dark:bg-gray-700 rounded w-24 mb-2"></div>
      <div class="h-3 bg-gray-200 dark:bg-gray-700 rounded w-20"></div>
    </div>
  </div>
{:else if type === 'media'}
  <div class="grid grid-cols-3 gap-1 p-4">
    {#each Array(9) as _, i}
      <div class="aspect-square bg-gray-200 dark:bg-gray-700 rounded animate-pulse"></div>
    {/each}
  </div>
{/if}

<style>
  @keyframes pulse {
    0%, 100% {
      opacity: 1;
    }
    50% {
      opacity: 0.7;
    }
  }
  
  .animate-pulse {
    animation: pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite;
  }
</style> 