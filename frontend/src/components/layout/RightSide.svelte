<script lang="ts">
  import type { ITrend, ISuggestedFollow } from "../../interfaces/ISocialMedia";

  export let trends: ITrend[] = [];
  export let suggestedFollows: ISuggestedFollow[] = [];
  export let isDarkMode: boolean = false;
</script>

<div class="search-container {isDarkMode ? 'search-dark' : ''} sticky top-0 z-10 pb-3 bg-inherit">
  <div class="relative">
    <input 
      type="text" 
      placeholder="Search" 
      class="search-input w-full rounded-full {isDarkMode ? 'bg-gray-800 border-gray-700 text-white' : 'bg-gray-100 border-gray-200'}"
    />
    <div class="search-icon">
      <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
      </svg>
    </div>
  </div>
</div>

<div class="sidebar {isDarkMode ? 'sidebar-dark' : ''} rounded-xl mb-4">
  <div class="sidebar-section">
    <h2 class="sidebar-title text-xl font-bold">What's happening</h2>
    {#if trends.length > 0}
      <ul>
        {#each trends as trend, i}
          <li class="py-3 {i < trends.length - 1 ? 'border-b' : ''} {isDarkMode ? 'border-gray-700' : 'border-gray-200'}">
            <p class="{isDarkMode ? 'text-gray-400' : 'text-gray-500'} text-sm">{trend.category}</p>
            <p class="font-semibold {isDarkMode ? 'text-white' : 'text-black'} my-0.5">{trend.title}</p>
            <p class="{isDarkMode ? 'text-gray-400' : 'text-gray-500'} text-sm">{trend.postCount} posts</p>
          </li>
        {/each}
      </ul>
      <div class="mt-3 pt-2">
        <a href="/trends" class="text-blue-500 hover:text-blue-600 text-sm font-medium">
          Show more
        </a>
      </div>
    {:else}
      <p class="{isDarkMode ? 'text-gray-400' : 'text-gray-500'}">No trends available</p>
    {/if}
  </div>
</div>

<div class="sidebar {isDarkMode ? 'sidebar-dark' : ''} rounded-xl mb-4">
  <div class="sidebar-section">
    <h2 class="sidebar-title text-xl font-bold">Who to follow</h2>
    {#if suggestedFollows.length > 0}
      <ul>
        {#each suggestedFollows as follow, i}
          <li class="flex items-center gap-3 py-3 {i < suggestedFollows.length - 1 ? 'border-b' : ''} {isDarkMode ? 'border-gray-700' : 'border-gray-200'}">
            <div class="w-10 h-10 rounded-full overflow-hidden {isDarkMode ? 'bg-gray-700' : 'bg-gray-200'} flex-shrink-0 flex items-center justify-center">
              {#if typeof follow.avatar === 'string' && follow.avatar.startsWith('http')}
                <img src={follow.avatar} alt={follow.username} class="w-full h-full object-cover" />
              {:else}
                <div class="flex items-center justify-center w-full h-full text-lg">{follow.avatar || '👤'}</div>
              {/if}
            </div>
            <div class="flex-1 min-w-0">
              <div class="flex items-center">
                <p class="font-semibold {isDarkMode ? 'text-white' : 'text-black'} truncate">{follow.displayName}</p>
                {#if follow.verified}
                  <span class="ml-1 text-blue-500">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor">
                      <path fill-rule="evenodd" d="M6.267 3.455a3.066 3.066 0 001.745-.723 3.066 3.066 0 013.976 0 3.066 3.066 0 001.745.723 3.066 3.066 0 012.812 2.812c.051.643.304 1.254.723 1.745a3.066 3.066 0 010 3.976 3.066 3.066 0 00-.723 1.745 3.066 3.066 0 01-2.812 2.812 3.066 3.066 0 00-1.745.723 3.066 3.066 0 01-3.976 0 3.066 3.066 0 00-1.745-.723 3.066 3.066 0 01-2.812-2.812 3.066 3.066 0 00-.723-1.745 3.066 3.066 0 010-3.976 3.066 3.066 0 00.723-1.745 3.066 3.066 0 012.812-2.812zm7.44 5.252a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                    </svg>
                  </span>
                {/if}
              </div>
              <p class="{isDarkMode ? 'text-gray-400' : 'text-gray-500'} text-sm truncate">@{follow.username}</p>
            </div>
            <button class="btn-follow {isDarkMode ? 'btn-follow-dark' : ''} rounded-full text-sm font-bold bg-black text-white dark:bg-white dark:text-black px-4 py-1.5 hover:bg-opacity-90 transition-colors">
              Follow
            </button>
          </li>
        {/each}
      </ul>
      <div class="mt-3 pt-2">
        <a href="/connect" class="text-blue-500 hover:text-blue-600 text-sm font-medium">
          Show more
        </a>
      </div>
    {:else}
      <p class="{isDarkMode ? 'text-gray-400' : 'text-gray-500'}">No suggestions available</p>
    {/if}
  </div>
</div>

<div class="{isDarkMode ? 'text-gray-400' : 'text-gray-500'} text-xs">
  <div class="flex flex-wrap gap-2">
    <a href="#" class="hover:underline">Terms of Service</a>
    <a href="#" class="hover:underline">Privacy Policy</a>
    <a href="#" class="hover:underline">Cookie Policy</a>
    <a href="#" class="hover:underline">Accessibility</a>
    <a href="#" class="hover:underline">Ads info</a>
    <a href="#" class="hover:underline">More</a>
  </div>
  <p class="mt-2">© {new Date().getFullYear()} AYCOM</p>
</div>