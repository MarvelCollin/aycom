<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { useTheme } from '../../hooks/useTheme';
  
  const dispatch = createEventDispatcher();
  const { theme } = useTheme();
  
  // Reactive declarations
  $: isDarkMode = $theme === 'dark';
  
  // Props with more flexible type definition
  export let profile: {
    id: string;
    username: string;
    name?: string;
    displayName?: string;
    profile_picture_url?: string | null;
    avatar?: string | null;
    bio?: string;
    is_verified?: boolean;
    isVerified?: boolean;
    follower_count?: number;
    followerCount?: number;
    is_following?: boolean;
    isFollowing?: boolean;
  };
  export let showBio = true;
  export let showFollowerCount = true;
  export let compact = false;
  
  // Handle follow user
  function handleFollow(event) {
    event.stopPropagation(); // Prevent navigation when follow button is clicked
    dispatch('follow', profile.id);
  }

  // Handle card click to navigate to user profile
  function handleCardClick() {
    // Navigate to user profile
    window.location.href = `/user/${profile.id}`;
    dispatch('profileClick', profile.id);
  }

  // Get the correct name property based on what's available
  $: displayName = profile.name || profile.displayName || profile.username;
  // Get the correct avatar property
  $: avatarUrl = profile.profile_picture_url || profile.avatar;
  // Get the correct verified status
  $: isVerified = profile.is_verified || profile.isVerified || false;
  // Get the correct follower count
  $: followerCount = profile.follower_count || profile.followerCount || 0;
  // Get the correct following status
  $: isFollowing = profile.is_following || profile.isFollowing || false;
</script>

<div 
  class="profile-card {isDarkMode ? 'profile-card-dark' : ''} {compact ? 'py-3' : 'py-4'}"
  on:click={handleCardClick}
  on:keydown={(e) => e.key === 'Enter' && handleCardClick()}
  tabindex="0"
  role="button"
>
  <div class="flex items-center flex-1 min-w-0">
    <div class="avatar-container {compact ? 'w-10 h-10' : 'w-14 h-14'} rounded-full overflow-hidden mr-4">
      {#if avatarUrl && typeof avatarUrl === 'string' && avatarUrl.startsWith('http')}
        <div class="image-wrapper">
          <img src={avatarUrl} alt={profile.username} class="w-full h-full object-cover" loading="lazy" />
        </div>
      {:else}
        <div class="w-full h-full flex items-center justify-center {isDarkMode ? 'bg-blue-900' : 'bg-blue-100'}">
          <span class="text-lg {isDarkMode ? 'text-blue-300' : 'text-blue-800'}">{displayName.charAt(0).toUpperCase()}</span>
        </div>
      {/if}
    </div>
    <div class="flex-1 min-w-0">
      <div class="flex items-center">
        <p class="font-bold {isDarkMode ? 'text-white' : 'text-gray-900'} truncate">{displayName}</p>
        {#if isVerified}
          <span class="verified-badge ml-1 text-blue-500">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor">
              <path fill-rule="evenodd" d="M6.267 3.455a3.066 3.066 0 001.745-.723 3.066 3.066 0 013.976 0 3.066 3.066 0 001.745.723 3.066 3.066 0 012.812 2.812c.051.643.304 1.254.723 1.745a3.066 3.066 0 010 3.976 3.066 3.066 0 00-.723 1.745 3.066 3.066 0 01-2.812 2.812 3.066 3.066 0 00-1.745.723 3.066 3.066 0 01-3.976 0 3.066 3.066 0 00-1.745-.723 3.066 3.066 0 01-2.812-2.812 3.066 3.066 0 00-.723-1.745 3.066 3.066 0 010-3.976 3.066 3.066 0 00.723-1.745 3.066 3.066 0 012.812-2.812zm7.44 5.252a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
            </svg>
          </span>
        {/if}
      </div>
      <p class="username {isDarkMode ? 'text-gray-400' : 'text-gray-600'} text-sm truncate">@{profile.username}</p>
      {#if showBio && profile.bio}
        <p class="bio-text text-sm {isDarkMode ? 'text-gray-400' : 'text-gray-600'} mt-1 truncate">{profile.bio}</p>
      {/if}
      {#if showFollowerCount}
        <p class="follower-count text-xs {isDarkMode ? 'text-gray-500' : 'text-gray-500'} mt-1">
          <span class="font-semibold">{followerCount}</span> {followerCount === 1 ? 'follower' : 'followers'}
        </p>
      {/if}
    </div>
  </div>
  <button 
    class="follow-button ml-3 px-4 py-1.5 rounded-full font-medium text-sm transition-all duration-200 
    {isFollowing 
      ? `border ${isDarkMode ? 'border-gray-600 text-white hover:bg-gray-800' : 'border-gray-300 text-gray-900 hover:bg-gray-100'}`
      : `${isDarkMode ? 'bg-blue-600 hover:bg-blue-700 text-white' : 'bg-blue-500 hover:bg-blue-600 text-white'}`}"
    on:click={handleFollow}
  >
    {isFollowing ? 'Following' : 'Follow'}
  </button>
</div>

<style>
  .profile-card {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding-left: 1rem;
    padding-right: 1rem;
    cursor: pointer;
    transition: background-color 0.2s ease;
  }
  
  .profile-card:hover {
    background-color: var(--hover-bg, rgba(0, 0, 0, 0.05));
  }
  
  .profile-card-dark:hover {
    background-color: var(--hover-bg-dark, rgba(255, 255, 255, 0.05));
  }
  
  .avatar-container {
    flex-shrink: 0;
    border: 2px solid transparent;
    width: 3.5rem !important;
    height: 3.5rem !important;
    position: relative;
    display: flex;
    align-items: center;
    justify-content: center;
    overflow: hidden;
  }
  
  .profile-card:hover .avatar-container {
    border-color: var(--color-primary);
  }
  
  .image-wrapper {
    width: 100%;
    height: 100%;
    display: block;
    position: relative;
  }
  
  .image-wrapper img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    object-position: center;
  }
  
  .verified-badge {
    display: inline-flex;
    animation: pulse 2s infinite;
  }
  
  @keyframes pulse {
    0% {
      opacity: 0.8;
    }
    50% {
      opacity: 1;
    }
    100% {
      opacity: 0.8;
    }
  }
  
  .follow-button {
    white-space: nowrap;
    transition: all 0.2s ease;
  }
  
  .username, .bio-text, .follower-count {
    line-height: 1.4;
  }
</style> 