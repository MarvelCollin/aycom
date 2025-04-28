<script lang="ts">
  import type { ITrend } from '../../interfaces/ISocialMedia';
  
  export let trends: ITrend[] = [
    { category: 'Trending in Technology', title: '#AYCOM', postCount: '15.2K' },
    { category: 'Trending in Business', title: '#WebDevelopment', postCount: '8.5K' },
    { category: 'Trending in Indonesia', title: '#BINUS', postCount: '5.2K' }
  ];
</script>

<div class="bg-gray-900 rounded-xl mb-6">
  <h2 class="text-xl font-bold p-4">Trends for you</h2>
  
  {#each trends as trend}
    <div class="p-4 hover:bg-gray-800 cursor-pointer transition-colors">
      <p class="text-xs text-gray-500">{trend.category}</p>
      <p class="font-bold">{trend.title}</p>
      <p class="text-xs text-gray-500">{trend.postCount} posts</p>
    </div>
  {/each}
  
  <a href="/trends" class="block p-4 text-blue-500 hover:bg-gray-800 rounded-b-xl">
    Show more
  </a>
</div> 