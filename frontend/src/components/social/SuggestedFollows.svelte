<script lang="ts">
  import type { ISuggestedFollow } from '../../interfaces/ISocialMedia';
  
  export let suggestedUsers: ISuggestedFollow[] = [
    { displayName: 'Tesla', username: 'Tesla', avatar: '⚡' },
    { displayName: 'GitHub', username: 'github', avatar: '👨‍💻' }
  ];
</script>

<div class="bg-gray-900 rounded-xl">
  <h2 class="text-xl font-bold p-4">Who to follow</h2>
  
  {#each suggestedUsers as user}
    <div class="p-4 hover:bg-gray-800 cursor-pointer transition-colors flex items-center">
      <div class="w-10 h-10 bg-gray-300 rounded-full flex items-center justify-center mr-3">
        <span>{user.avatar}</span>
      </div>
      <div class="flex-1">
        <p class="font-bold">{user.displayName}</p>
        <p class="text-gray-500 text-sm">@{user.username}</p>
      </div>
      <button class="bg-white text-black px-4 py-1 rounded-full font-bold">Follow</button>
    </div>
  {/each}
  
  <a href="/connect" class="block p-4 text-blue-500 hover:bg-gray-800 rounded-b-xl">
    Show more
  </a>
</div> 