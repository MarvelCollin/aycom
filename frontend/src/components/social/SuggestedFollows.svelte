<script lang="ts">
  import type { ISuggestedFollow } from '../../interfaces/ISocialMedia';
  
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
  
  // Format follower count for display
  function formatFollowerCount(count: number): string {
    if (count >= 1000000) {
      return (count / 1000000).toFixed(1) + 'M';
    } else if (count >= 1000) {
      return (count / 1000).toFixed(1) + 'K';
    }
    return count.toString();
  }
  
  // Toggle follow state (in a real app, this would call an API)
  function toggleFollow(index: number) {
    suggestedUsers = suggestedUsers.map((user, i) => {
      if (i === index) {
        return { ...user, isFollowing: !user.isFollowing };
      }
      return user;
    });
  }
</script>

<div class="bg-gray-900 rounded-xl mb-6">
  <h2 class="text-xl font-bold p-4">Who to follow</h2>
  
  {#each suggestedUsers as user, index}
    <div class="p-4 hover:bg-gray-800 cursor-pointer transition-colors flex items-center">
      <div class="w-10 h-10 bg-gray-300 rounded-full flex items-center justify-center mr-3">
        <span>{user.avatar}</span>
      </div>
      <div class="flex-1">
        <div class="flex items-center">
          <p class="font-bold">{user.displayName}</p>
          {#if user.verified}
            <span class="ml-1 text-blue-500">✓</span>
          {/if}
        </div>
        <p class="text-gray-500 text-sm">@{user.username}</p>
      </div>
      <button 
        class="bg-white hover:bg-gray-200 text-black px-4 py-1 rounded-full font-bold transition-colors"
        on:click={() => toggleFollow(index)}
      >
        {user.isFollowing ? 'Following' : 'Follow'}
      </button>
    </div>
  {/each}
</div> 