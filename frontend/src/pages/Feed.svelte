<script lang="ts">
  // Dummy data for tweets
  const tweets = [
    {
      id: 1,
      username: 'elonmusk',
      displayName: 'Elon Musk',
      avatar: '👨‍🚀',
      content: 'Just launched another rocket! 🚀 #SpaceX',
      timestamp: '2h',
      likes: 3240,
      replies: 421,
      reposts: 892,
      views: '1.2M'
    }
  ];
  
  let newTweet = '';
  
  function postTweet() {
    if (newTweet.trim() === '') return;
    console.log('Posted:', newTweet);
    newTweet = '';
  }
</script>

<div class="min-h-screen bg-black text-white">
  <!-- Navigation Sidebar -->
  <div class="fixed top-0 bottom-0 left-0 w-16 md:w-64 border-r border-gray-800 p-2 md:p-4">
    <div class="flex flex-col items-center md:items-start h-full">
      <!-- Logo -->
      <div class="p-2 mb-4 rounded-full hover:bg-gray-900">
        <span class="text-white text-2xl font-bold">AY</span>
      </div>
      
      <!-- Navigation Items -->
      <nav class="w-full">
        <a href="/home" class="flex items-center p-3 hover:bg-gray-900 rounded-full mb-1 font-medium">
          <span class="mr-4">🏠</span>
          <span class="hidden md:inline">Home</span>
        </a>
        <a href="/explore" class="flex items-center p-3 hover:bg-gray-900 rounded-full mb-1 font-medium">
          <span class="mr-4">🔍</span>
          <span class="hidden md:inline">Explore</span>
        </a>
        <a href="/notifications" class="flex items-center p-3 hover:bg-gray-900 rounded-full mb-1 font-medium">
          <span class="mr-4">🔔</span>
          <span class="hidden md:inline">Notifications</span>
        </a>
        <a href="/messages" class="flex items-center p-3 hover:bg-gray-900 rounded-full mb-1 font-medium">
          <span class="mr-4">✉️</span>
          <span class="hidden md:inline">Messages</span>
        </a>
        <a href="/profile" class="flex items-center p-3 hover:bg-gray-900 rounded-full mb-1 font-medium">
          <span class="mr-4">👤</span>
          <span class="hidden md:inline">Profile</span>
        </a>
      </nav>
      
      <!-- Post Button -->
      <button class="mt-4 w-12 h-12 md:w-full md:py-3 bg-blue-500 text-white rounded-full font-bold hover:bg-blue-600">
        <span class="md:hidden">+</span>
        <span class="hidden md:inline">Post</span>
      </button>
      
      <!-- User Profile -->
      <div class="mt-auto flex items-center p-3 hover:bg-gray-900 rounded-full cursor-pointer">
        <div class="w-8 h-8 bg-gray-300 rounded-full flex items-center justify-center mr-2">
          <span>👤</span>
        </div>
        <div class="hidden md:block">
          <p class="font-semibold">User Name</p>
          <p class="text-gray-500 text-sm">@username</p>
        </div>
      </div>
    </div>
  </div>
  
  <!-- Main Content -->
  <div class="ml-16 md:ml-64">
    <!-- Header -->
    <header class="sticky top-0 bg-black bg-opacity-80 backdrop-blur-sm z-10 p-4 border-b border-gray-800">
      <h1 class="text-xl font-bold">Home</h1>
    </header>
    
    <!-- Tweet Compose -->
    <div class="border-b border-gray-800 p-4">
      <div class="flex">
        <div class="w-12 h-12 bg-gray-300 rounded-full flex items-center justify-center mr-4">
          <span>👤</span>
        </div>
        <div class="flex-1">
          <textarea 
            bind:value={newTweet}
            placeholder="What is happening?!"
            class="w-full bg-transparent border-none outline-none text-white resize-none mb-4 text-xl"
            rows="3"
          ></textarea>
          
          <div class="flex justify-between items-center">
            <div class="flex text-blue-500">
              <button class="p-2 rounded-full hover:bg-blue-900 hover:bg-opacity-20">📷</button>
              <button class="p-2 rounded-full hover:bg-blue-900 hover:bg-opacity-20">📊</button>
              <button class="p-2 rounded-full hover:bg-blue-900 hover:bg-opacity-20">😀</button>
              <button class="p-2 rounded-full hover:bg-blue-900 hover:bg-opacity-20">📅</button>
              <button class="p-2 rounded-full hover:bg-blue-900 hover:bg-opacity-20">📍</button>
            </div>
            
            <button 
              on:click={postTweet}
              class="px-4 py-2 bg-blue-500 text-white rounded-full font-bold hover:bg-blue-600 disabled:opacity-50"
              disabled={newTweet.trim() === ''}
            >
              Post
            </button>
          </div>
        </div>
      </div>
    </div>
    
    <!-- Tweet Feed -->
    <div>
      {#each tweets as tweet (tweet.id)}
        <div class="p-4 border-b border-gray-800 hover:bg-gray-900 hover:bg-opacity-30 transition-colors">
          <div class="flex">
            <div class="w-12 h-12 bg-gray-300 rounded-full flex items-center justify-center mr-3">
              <span>{tweet.avatar}</span>
            </div>
            <div class="flex-1">
              <div class="flex items-center mb-1">
                <span class="font-bold">{tweet.displayName}</span>
                <span class="text-gray-500 ml-2">@{tweet.username}</span>
                <span class="text-gray-500 mx-1">·</span>
                <span class="text-gray-500">{tweet.timestamp}</span>
              </div>
              
              <p class="mb-3">{tweet.content}</p>
              
              <div class="flex justify-between text-gray-500 max-w-md">
                <button class="flex items-center hover:text-blue-500">
                  <span>💬</span>
                  <span class="ml-2 text-xs">{tweet.replies}</span>
                </button>
                <button class="flex items-center hover:text-green-500">
                  <span>🔄</span>
                  <span class="ml-2 text-xs">{tweet.reposts}</span>
                </button>
                <button class="flex items-center hover:text-pink-500">
                  <span>❤️</span>
                  <span class="ml-2 text-xs">{tweet.likes}</span>
                </button>
                <button class="flex items-center hover:text-blue-500">
                  <span>📊</span>
                  <span class="ml-2 text-xs">{tweet.views}</span>
                </button>
                <button class="flex items-center hover:text-blue-500">
                  <span>⬆️</span>
                </button>
              </div>
            </div>
          </div>
        </div>
      {/each}
    </div>
  </div>
  
  <!-- Right Sidebar -->
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
      
      <div class="p-4 hover:bg-gray-800 cursor-pointer transition-colors">
        <p class="text-xs text-gray-500">Trending in Technology</p>
        <p class="font-bold">#AYCOM</p>
        <p class="text-xs text-gray-500">15.2K posts</p>
      </div>
      
      <div class="p-4 hover:bg-gray-800 cursor-pointer transition-colors">
        <p class="text-xs text-gray-500">Trending in Business</p>
        <p class="font-bold">#WebDevelopment</p>
        <p class="text-xs text-gray-500">8.5K posts</p>
      </div>
      
      <div class="p-4 hover:bg-gray-800 cursor-pointer transition-colors">
        <p class="text-xs text-gray-500">Trending in Indonesia</p>
        <p class="font-bold">#BINUS</p>
        <p class="text-xs text-gray-500">5.2K posts</p>
      </div>
      
      <a href="/trends" class="block p-4 text-blue-500 hover:bg-gray-800 rounded-b-xl">
        Show more
      </a>
    </div>
    
    <!-- Who to follow -->
    <div class="bg-gray-900 rounded-xl">
      <h2 class="text-xl font-bold p-4">Who to follow</h2>
      
      <div class="p-4 hover:bg-gray-800 cursor-pointer transition-colors flex items-center">
        <div class="w-10 h-10 bg-gray-300 rounded-full flex items-center justify-center mr-3">
          <span>⚡</span>
        </div>
        <div class="flex-1">
          <p class="font-bold">Tesla</p>
          <p class="text-gray-500 text-sm">@Tesla</p>
        </div>
        <button class="bg-white text-black px-4 py-1 rounded-full font-bold">Follow</button>
      </div>
      
      <div class="p-4 hover:bg-gray-800 cursor-pointer transition-colors flex items-center">
        <div class="w-10 h-10 bg-gray-300 rounded-full flex items-center justify-center mr-3">
          <span>👨‍💻</span>
        </div>
        <div class="flex-1">
          <p class="font-bold">GitHub</p>
          <p class="text-gray-500 text-sm">@github</p>
        </div>
        <button class="bg-white text-black px-4 py-1 rounded-full font-bold">Follow</button>
      </div>
      
      <a href="/connect" class="block p-4 text-blue-500 hover:bg-gray-800 rounded-b-xl">
        Show more
      </a>
    </div>
  </div>
</div> 