<script lang="ts">
  import type { ITrend } from '../../interfaces/ISocialMedia';
  
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

<div class="bg-gray-900 rounded-xl mb-6">
  <h2 class="text-xl font-bold p-4">What's happening</h2>
  
  {#each displayTrends as trend}
    <div class="p-4 hover:bg-gray-800 cursor-pointer transition-colors">
      <p class="text-xs text-gray-500">{trend.category}</p>
      <p class="font-bold">{trend.title}</p>
      <p class="text-xs text-gray-500">{trend.postCount} posts</p>
    </div>
  {/each}
  
  <a href="/explore/trends" class="block p-4 text-blue-500 hover:bg-gray-800 rounded-b-xl">
    Show more
  </a>
</div> 