<script lang="ts">
  import type { ITrend, ISuggestedFollow } from "../../interfaces/ISocialMedia";

  export let trends: ITrend[] = [];
  export let suggestedFollows: ISuggestedFollow[] = [];
</script>

<div class="hidden lg:block fixed right-0 w-[350px] h-screen overflow-y-auto p-4">
  <div class="bg-gray-100 dark:bg-gray-800 rounded-xl p-4 mb-4">
    <h2 class="text-xl font-bold mb-4">Trends for you</h2>
    {#if trends.length > 0}
      <ul>
        {#each trends as trend}
          <li class="mb-3">
            <p class="text-gray-500 text-sm">{trend.category}</p>
            <p class="font-semibold">{trend.title}</p>
            <p class="text-gray-500 text-sm">{trend.postCount} posts</p>
          </li>
        {/each}
      </ul>
    {:else}
      <p class="text-gray-500">No trends available</p>
    {/if}
  </div>

  <div class="bg-gray-100 dark:bg-gray-800 rounded-xl p-4">
    <h2 class="text-xl font-bold mb-4">Who to follow</h2>
    {#if suggestedFollows.length > 0}
      <ul>
        {#each suggestedFollows as follow}
          <li class="flex items-center gap-3 mb-3">
            <img src={follow.avatar} alt={follow.username} class="w-10 h-10 rounded-full" />
            <div class="flex-1">
              <p class="font-semibold">{follow.displayName}</p>
              <p class="text-gray-500 text-sm">@{follow.username}</p>
            </div>
            <button class="bg-blue-500 text-white rounded-full px-4 py-1 text-sm font-bold">
              Follow
            </button>
          </li>
        {/each}
      </ul>
    {:else}
      <p class="text-gray-500">No suggestions available</p>
    {/if}
  </div>
</div> 