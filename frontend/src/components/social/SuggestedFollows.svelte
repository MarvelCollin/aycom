<script lang="ts">
  import type { ISuggestedFollow } from '../../interfaces/ISocialMedia';
  
  export let isDarkMode = false;
  export let suggestedUsers: ISuggestedFollow[] = [
    { 
      displayName: 'Brainwalla', 
      username: 'brainwalla', 
      avatar: '🧠', 
      verified: true,
      followerCount: 12300000
    },
    { 
      displayName: 'Peach', 
      username: 'peach', 
      avatar: '🍑', 
      verified: true,
      followerCount: 8500000
    },
    { 
      displayName: 'YTuber', 
      username: 'ytuber', 
      avatar: '▶️', 
      verified: false,
      followerCount: 5700000
    }
  ];
  
  function formatFollowerCount(count: number): string {
    if (count >= 1000000) {
      return (count / 1000000).toFixed(1) + 'M';
    } else if (count >= 1000) {
      return (count / 1000).toFixed(1) + 'K';
    }
    return count.toString();
  }
  
  function toggleFollow(index: number) {
    suggestedUsers = suggestedUsers.map((user, i) => {
      if (i === index) {
        return { ...user, isFollowing: !user.isFollowing };
      }
      return user;
    });
  }
</script>

<div class="{isDarkMode ? 'bg-gray-800' : 'bg-gray-50'} rounded-2xl overflow-hidden">
  <h2 class="text-xl font-bold p-4">Who to follow</h2>
  
  {#each suggestedUsers as user, index}
    <div class="px-4 py-3 {isDarkMode ? 'hover:bg-gray-700/50' : 'hover:bg-gray-200/50'} cursor-pointer transition-colors">
      <div class="flex items-center justify-between">
        <div class="flex items-center">
          <div class="w-10 h-10 {isDarkMode ? 'bg-gray-700' : 'bg-gray-300'} rounded-full flex items-center justify-center overflow-hidden flex-shrink-0">
            <span>{user.avatar}</span>
          </div>
          <div class="ml-3">
            <div class="flex items-center">
              <p class="font-bold hover:underline">{user.displayName}</p>
              {#if user.verified}
                <span class="ml-1 text-blue-500">✓</span>
              {/if}
            </div>
            <p class="{isDarkMode ? 'text-gray-400' : 'text-gray-500'} text-sm">@{user.username}</p>
          </div>
        </div>
        <button 
          class="{user.isFollowing 
            ? 'bg-transparent border border-gray-500 text-gray-300 hover:border-red-500 hover:text-red-500 hover:bg-red-500/10' 
            : 'bg-black text-white hover:bg-gray-800'} 
            px-4 py-1.5 rounded-full font-bold transition-colors text-sm"
          on:click={() => toggleFollow(index)}
        >
          {user.isFollowing ? 'Following' : 'Follow'}
        </button>
      </div>
    </div>
  {/each}
  
  <a href="/connect" class="block p-4 text-blue-500 {isDarkMode ? 'hover:bg-gray-700/50' : 'hover:bg-gray-200/50'} transition-colors">
    Show more
  </a>
</div> 