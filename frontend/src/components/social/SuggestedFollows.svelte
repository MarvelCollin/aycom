<script lang="ts">
  import type { ISuggestedFollow } from '../../interfaces/ISocialMedia';
  import { followUser, unfollowUser } from '../../api/user';
  import { toastStore } from '../../stores/toastStore';
  import { createLoggerWithPrefix } from '../../utils/logger';
  
  const logger = createLoggerWithPrefix('SuggestedFollows');
  
  export let isDarkMode = false;
  export let suggestedUsers: ISuggestedFollow[] = [
    { 
      name: 'Brainwalla', 
      username: 'brainwalla', 
      profile_picture_url: '🧠', 
      is_verified: true,
      follower_count: 12300000,
      user_id: '1',
      is_following: false
    },
    { 
      name: 'Peach', 
      username: 'peach', 
      profile_picture_url: '🍑', 
      is_verified: true,
      follower_count: 8500000,
      user_id: '2',
      is_following: false
    },
    { 
      name: 'YTuber', 
      username: 'ytuber', 
      profile_picture_url: '▶️', 
      is_verified: false,
      follower_count: 5700000,
      user_id: '3',
      is_following: false
    }
  ];
  
  // Track loading state per user
  let followingInProgress: Record<string, boolean> = {};
  
  function formatFollowerCount(count: number): string {
    if (count >= 1000000) {
      return (count / 1000000).toFixed(1) + 'M';
    } else if (count >= 1000) {
      return (count / 1000).toFixed(1) + 'K';
    }
    return count.toString();
  }
  
  async function toggleFollow(index: number) {
    const user = suggestedUsers[index];
    if (!user || followingInProgress[user.user_id]) return;
    
    // Set loading state
    followingInProgress[user.user_id] = true;
    
    try {
      // Optimistically update UI
      suggestedUsers = suggestedUsers.map((u, i) => {
        if (i === index) {
          return { ...u, is_following: !u.is_following };
        }
        return u;
      });
      
      // Make API call
      const userId = user.user_id;
      const wasFollowing = !user.is_following; // We already toggled it above
      
      logger.debug(`${wasFollowing ? 'Unfollowing' : 'Following'} user ${userId}`);
      
      // API call - note we use the opposite of current state since we already updated it
      const response = wasFollowing 
        ? await unfollowUser(userId)
        : await followUser(userId);
      
      logger.debug('Follow/unfollow response:', response);
      
      // Show success message
      if (response.success) {
        toastStore.showToast(`${wasFollowing ? 'Unfollowed' : 'Followed'} @${user.username}`, 'success');
      } else {
        // If request failed, revert UI
        suggestedUsers = suggestedUsers.map((u, i) => {
          if (i === index) {
            return { ...u, is_following: wasFollowing };
          }
          return u;
        });
        toastStore.showToast(response.message || 'Failed to update follow status', 'error');
      }
    } catch (error) {
      logger.error('Error toggling follow:', error);
      
      // Revert UI change
      suggestedUsers = suggestedUsers.map((u, i) => {
        if (i === index) {
          return { ...u, is_following: !u.is_following };
        }
        return u;
      });
      
      toastStore.showToast('Failed to update follow status', 'error');
    } finally {
      // Clear loading state
      followingInProgress[user.user_id] = false;
    }
  }
</script>

<div class="{isDarkMode ? 'bg-gray-800' : 'bg-gray-50'} rounded-2xl overflow-hidden">
  <h2 class="text-xl font-bold p-4">Who to follow</h2>
  
  {#each suggestedUsers as user, index}
    <div class="px-4 py-3 {isDarkMode ? 'hover:bg-gray-700/50' : 'hover:bg-gray-200/50'} cursor-pointer transition-colors">
      <div class="flex items-center justify-between">
        <div class="flex items-center">
          <div class="w-10 h-10 {isDarkMode ? 'bg-gray-700' : 'bg-gray-300'} rounded-full flex items-center justify-center overflow-hidden flex-shrink-0">
            <span>{user.profile_picture_url}</span>
          </div>
          <div class="ml-3">
            <div class="flex items-center">
              <p class="font-bold hover:underline">{user.name}</p>
              {#if user.is_verified}
                <span class="ml-1 text-blue-500">✓</span>
              {/if}
            </div>
            <p class="{isDarkMode ? 'text-gray-400' : 'text-gray-500'} text-sm">@{user.username}</p>
          </div>
        </div>
        <button 
          class="{user.is_following 
            ? 'bg-transparent border border-gray-500 text-gray-300 hover:border-red-500 hover:text-red-500 hover:bg-red-500/10' 
            : 'bg-black text-white hover:bg-gray-800'} 
            px-4 py-1.5 rounded-full font-bold transition-colors text-sm"
          on:click={(e) => {
            e.stopPropagation();
            toggleFollow(index);
          }}
          disabled={followingInProgress[user.user_id]}
        >
          {followingInProgress[user.user_id] ? 
            '...' : 
            user.is_following ? 'Following' : 'Follow'}
        </button>
      </div>
    </div>
  {/each}
  
  <a href="/connect" class="block p-4 text-blue-500 {isDarkMode ? 'hover:bg-gray-700/50' : 'hover:bg-gray-200/50'} transition-colors">
    Show more
  </a>
</div> 