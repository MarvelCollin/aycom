<script lang="ts">
  export let tweet = {
    id: 0,
    username: '',
    displayName: '',
    avatar: '👤',
    content: '',
    timestamp: '',
    likes: 0,
    replies: 0,
    reposts: 0,
    views: '0'
  };
</script>

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