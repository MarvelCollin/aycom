<script lang="ts">
  import type { ITrend, ISuggestedFollow } from "../../interfaces/ISocialMedia";
  import ComposeTweet from "../social/ComposeTweet.svelte";

  export let trends: ITrend[] = [];
  export let suggestedFollows: ISuggestedFollow[] = [];
  export let isDarkMode: boolean = false;

  function navigateTo(path) {
    console.log(`Navigating to: ${path}`);
    // Implement actual navigation logic here when paths are created
    // window.location.href = path;
  }
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
        {#each trends as trend}
          <li class="py-3 {trends.indexOf(trend) !== trends.length - 1 ? 'border-b' : ''} {isDarkMode ? 'border-gray-700' : 'border-gray-200'}">
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
        {#each suggestedFollows as follow}
          <li class="flex items-center gap-3 py-3 {suggestedFollows.indexOf(follow) !== suggestedFollows.length - 1 ? 'border-b' : ''} {isDarkMode ? 'border-gray-700' : 'border-gray-200'}">
            <div class="w-10 h-10 rounded-full overflow-hidden {isDarkMode ? 'bg-gray-700' : 'bg-gray-200'} flex-shrink-0 flex items-center justify-center">
              <div class="flex items-center justify-center w-full h-full text-lg">
                {#if follow.avatar && !follow.avatar.includes('👤')}
                  <img src={follow.avatar} alt={follow.username} class="w-full h-full object-cover rounded-full" />
                {:else}
                  <img src="https://secure.gravatar.com/avatar/0?d=mp" alt={follow.username} class="w-full h-full object-cover rounded-full" />
                {/if}
              </div>
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
            <button class="follow-button rounded-full text-sm font-bold px-4 py-1.5 transition-colors {isDarkMode ? 'bg-white text-black hover:bg-gray-200' : 'bg-black text-white hover:bg-gray-800'}">
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

<div class="mt-4 text-center footer-links {isDarkMode ? 'footer-dark' : ''}">
  <div class="flex flex-wrap text-xs justify-between {isDarkMode ? 'text-gray-500' : 'text-gray-500'}">
    <button class="hover:underline mb-2" on:click={() => navigateTo('/terms')}>Terms of Service</button>
    <button class="hover:underline mb-2" on:click={() => navigateTo('/privacy')}>Privacy Policy</button>
    <button class="hover:underline mb-2" on:click={() => navigateTo('/cookies')}>Cookie Policy</button>
    <button class="hover:underline mb-2" on:click={() => navigateTo('/accessibility')}>Accessibility</button>
    <button class="hover:underline mb-2" on:click={() => navigateTo('/ads')}>Ads Info</button>
    <button class="hover:underline mb-2" on:click={() => navigateTo('/about')}>About</button>
  </div>
  <p class="text-xs mt-2 {isDarkMode ? 'text-gray-500' : 'text-gray-500'}">© 2023 AYCOM, Inc.</p>
</div>

<style>
  /* Dark mode styling */
  .sidebar {
    background-color: #f7f9fa;
    border: 1px solid #eff3f4;
    padding: 1rem;
  }
  
  .sidebar-dark {
    background-color: #16181c;
    border: 1px solid #2f3336;
  }
  
  .search-input {
    padding: 0.75rem 1rem 0.75rem 3rem;
    outline: none;
  }
  
  .search-icon {
    position: absolute;
    left: 1rem;
    top: 50%;
    transform: translateY(-50%);
    color: #536471;
  }
  
  .footer-links button {
    transition: color 0.2s;
  }
  
  .footer-dark button:hover {
    color: #e5e7eb;
  }
  
  .follow-button {
    transition: background-color 0.2s ease;
  }
</style>