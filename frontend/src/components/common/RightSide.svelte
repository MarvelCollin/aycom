<script lang="ts">
  import SearchBar from '../social/SearchBar.svelte';
  import TrendsList from '../social/TrendsList.svelte';
  import SuggestedFollows from '../social/SuggestedFollows.svelte';
  import type { ITrend, ISuggestedFollow } from '../../interfaces/ISocialMedia';
  
  export let isDarkMode = false;
  export let trends: ITrend[] = [];
  export let suggestedUsers: ISuggestedFollow[] = [];
  export let showPremium = true;
</script>

<div class="hidden lg:block w-[350px] flex-shrink-0 pl-7">
  <div class="sticky top-0 w-[350px] h-screen overflow-y-auto py-1">
    <!-- Search -->
    <div class="mb-4 mt-1 px-1">
      <SearchBar {isDarkMode} />
    </div>
    
    <!-- Premium Subscription -->
    {#if showPremium}
      <div class="{isDarkMode ? 'bg-gray-800' : 'bg-gray-100'} rounded-2xl mb-4 p-4 transition-colors">
        <h2 class="text-xl font-bold mb-2">Subscribe to Premium</h2>
        <p class="text-sm mb-4 {isDarkMode ? 'text-gray-300' : 'text-gray-600'}">Subscribe to unlock new features and if eligible, receive a share of revenue.</p>
        <button class="bg-blue-500 text-white px-4 py-2 rounded-full font-bold w-full hover:bg-blue-600 transition-colors">
          Subscribe
        </button>
      </div>
    {/if}
    
    <!-- Trends -->
    <div class="mb-4">
      <TrendsList {isDarkMode} {trends} />
    </div>
    
    <!-- Who to follow -->
    <div class="mb-4">
      <SuggestedFollows {isDarkMode} suggestedUsers={suggestedUsers} />
    </div>
    
    <!-- Footer links -->
    <div class="text-xs {isDarkMode ? 'text-gray-500' : 'text-gray-400'} mt-4 mb-10 flex flex-wrap px-4">
      <a href="#" class="mr-2 hover:underline">Terms of Service</a>
      <a href="#" class="mr-2 hover:underline">Privacy Policy</a>
      <a href="#" class="mr-2 hover:underline">Cookie Policy</a>
      <a href="#" class="mr-2 hover:underline">Accessibility</a>
      <a href="#" class="mr-2 hover:underline">Ads info</a>
      <span class="mr-2">© 2023 AYCOM</span>
    </div>
  </div>
</div>
