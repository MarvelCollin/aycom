<script lang="ts">
  import type { ITrend } from '../../interfaces/ISocialMedia';
  
  export let isDarkMode = false;
  export let trends: ITrend[] = [
    { category: 'Trending', title: '#Trump', postCount: '1.95M' },
    { category: 'Trending', title: '#nct2d', postCount: '145K' },
    { category: 'Trending in Indonesia', title: 'Pagi', postCount: '6,974' },
    { category: 'Trending in Indonesia', title: '#reddysoekinspirasi', postCount: '5.4K' },
    { category: 'Trending in Indonesia', title: '#AYCOMLaunch', postCount: '12.8K' }
  ];
  
  export let showAllTrends = false;
  
  // Display only 5 trends by default
  $: displayTrends = showAllTrends ? trends : trends.slice(0, 5);
</script>

<div class="{isDarkMode ? 'bg-gray-800' : 'bg-gray-50'} rounded-2xl overflow-hidden">
  <h2 class="text-xl font-bold p-4">What's happening</h2>
  
  {#each displayTrends as trend, i}
    <div class="px-4 py-3 {isDarkMode ? 'hover:bg-gray-700/50' : 'hover:bg-gray-200/50'} cursor-pointer transition-colors">
      <p class="text-xs {isDarkMode ? 'text-gray-400' : 'text-gray-500'} mb-1">{trend.category}</p>
      <p class="font-bold text-base">{trend.title}</p>
      <p class="text-xs {isDarkMode ? 'text-gray-400' : 'text-gray-500'} mt-1">{trend.postCount} posts</p>
    </div>
  {/each}
  
  <a href="/explore/trends" class="block p-4 text-blue-500 {isDarkMode ? 'hover:bg-gray-700/50' : 'hover:bg-gray-200/50'} transition-colors">
    Show more
  </a>
</div> 