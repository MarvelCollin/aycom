<script lang="ts">
  // Example trending topics
  const trendingTopics = [
    { category: 'Technology', hashtag: '#AYCOM', posts: '15.2K' },
    { category: 'Business', hashtag: '#WebDevelopment', posts: '8.5K' },
    { category: 'Indonesia', hashtag: '#BINUS', posts: '5.2K' }
  ];
  
  // Example users to follow
  const suggestedUsers = [
    { name: 'Tesla', username: 'Tesla', avatar: '⚡' },
    { name: 'GitHub', username: 'github', avatar: '👨‍💻' }
  ];
</script>

<div class="hidden lg:block fixed top-0 bottom-0 right-0 w-80 border-l border-gray-800 p-4 overflow-y-auto">
  <!-- Search -->
  <div class="mb-6">
    <div class="relative">
      <span class="absolute left-3 top-3 text-gray-500">🔍</span>
      <input 
        type="text" 
        placeholder="Search" 
        class="w-full bg-gray-900 rounded-full py-2 pl-10 pr-4 text-white focus:outline-none focus:ring-1 focus:ring-blue-500"
      />
    </div>
  </div>
  
  <!-- Trends -->
  <div class="bg-gray-900 rounded-xl mb-6">
    <h2 class="text-xl font-bold p-4">Trends for you</h2>
    
    {#each trendingTopics as topic}
      <div class="p-4 hover:bg-gray-800 cursor-pointer transition-colors">
        <p class="text-xs text-gray-500">Trending in {topic.category}</p>
        <p class="font-bold">{topic.hashtag}</p>
        <p class="text-xs text-gray-500">{topic.posts} posts</p>
      </div>
    {/each}
    
    <a href="/trends" class="block p-4 text-blue-500 hover:bg-gray-800 rounded-b-xl">
      Show more
    </a>
  </div>
  
  <!-- Who to follow -->
  <div class="bg-gray-900 rounded-xl">
    <h2 class="text-xl font-bold p-4">Who to follow</h2>
    
    {#each suggestedUsers as user}
      <div class="p-4 hover:bg-gray-800 cursor-pointer transition-colors flex items-center">
        <div class="w-10 h-10 bg-gray-300 rounded-full flex items-center justify-center mr-3">
          <span>{user.avatar}</span>
        </div>
        <div class="flex-1">
          <p class="font-bold">{user.name}</p>
          <p class="text-gray-500 text-sm">@{user.username}</p>
        </div>
        <button class="bg-white text-black px-4 py-1 rounded-full font-bold">Follow</button>
      </div>
    {/each}
    
    <a href="/connect" class="block p-4 text-blue-500 hover:bg-gray-800 rounded-b-xl">
      Show more
    </a>
  </div>
</div> 