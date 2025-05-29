<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { useTheme } from '../../hooks/useTheme';
  
  const dispatch = createEventDispatcher();
  const { theme } = useTheme();
  
  // Reactive declarations
  $: isDarkMode = $theme === 'dark';
  
  // Props
  export let profile: {
    id: string;
    username: string;
    name: string;
    profile_picture_url: string | null;
    bio?: string;
    is_verified: boolean;
    follower_count: number;
    is_following: boolean;
  };
  export let showBio = true;
  export let showFollowerCount = true;
  export let compact = false;
  
  // Handle follow user
  function handleFollow() {
    dispatch('follow', profile.id);
  }
</script>

<div class="flex items-center justify-between {compact ? 'py-2' : 'py-4'}">
  <a href={`/profile/${profile.username}`} class="flex items-center flex-1 min-w-0">
    <div class="{compact ? 'w-10 h-10' : 'w-12 h-12'} rounded-full {isDarkMode ? 'bg-gray-700' : 'bg-gray-200'} flex items-center justify-center overflow-hidden mr-3">
      {#if typeof profile.profile_picture_url === 'string' && profile.profile_picture_url.startsWith('http')}
        <img src={profile.profile_picture_url} alt={profile.username} class="w-full h-full object-cover" />
      {:else}
        <span class="text-lg">👤</span>
      {/if}
    </div>
    <div class="flex-1 min-w-0">
      <div class="flex items-center">
        <p class="font-bold {isDarkMode ? 'text-white' : 'text-black'} truncate">{profile.name}</p>
        {#if profile.is_verified}
          <span class="ml-1 text-blue-500">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor">
              <path fill-rule="evenodd" d="M6.267 3.455a3.066 3.066 0 001.745-.723 3.066 3.066 0 013.976 0 3.066 3.066 0 001.745.723 3.066 3.066 0 012.812 2.812c.051.643.304 1.254.723 1.745a3.066 3.066 0 010 3.976 3.066 3.066 0 00-.723 1.745 3.066 3.066 0 01-2.812 2.812 3.066 3.066 0 00-1.745.723 3.066 3.066 0 01-3.976 0 3.066 3.066 0 00-1.745-.723 3.066 3.066 0 01-2.812-2.812 3.066 3.066 0 00-.723-1.745 3.066 3.066 0 010-3.976 3.066 3.066 0 00.723-1.745 3.066 3.066 0 012.812-2.812zm7.44 5.252a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
            </svg>
          </span>
        {/if}
      </div>
      <p class="{isDarkMode ? 'text-gray-400' : 'text-gray-500'} text-sm truncate">@{profile.username}</p>
      {#if showBio && profile.bio}
        <p class="text-sm text-gray-500 dark:text-gray-400 mt-1 truncate">{profile.bio}</p>
      {/if}
      {#if showFollowerCount && profile.follower_count > 0}
        <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
          <span class="font-semibold">{profile.follower_count}</span> {profile.follower_count === 1 ? 'follower' : 'followers'}
        </p>
      {/if}
    </div>
  </a>
  <button 
    class="ml-4 px-4 py-1.5 rounded-full {profile.is_following ? 'bg-transparent border border-gray-300 dark:border-gray-600' : 'bg-black dark:bg-white text-white dark:text-black'} font-bold text-sm"
    on:click={handleFollow}
  >
    {profile.is_following ? 'Following' : 'Follow'}
  </button>
</div> 