<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { useTheme } from '../../hooks/useTheme';
  
  const dispatch = createEventDispatcher();
  const { theme } = useTheme();
  
  // Reactive declarations
  $: isDarkMode = $theme === 'dark';
  
  // Updated props to match parent component
  export let id: string;
  export let username: string;
  export let displayName: string;
  export let avatar: string | null = null;
  export let bio: string = "";
  export let isVerified: boolean = false;
  export let followerCount: number = 0;
  export let isFollowing: boolean = false;
  export let showBio: boolean = false; // Default to false to make cards more compact
  export let showFollowerCount: boolean = true;
  export let compact: boolean = true; // Default to compact
  export let onToggleFollow: () => void = () => {};
  
  // Log props for debugging
  $: console.log('ProfileCard props:', { id, username, displayName, isVerified });
  
  // Handle card click to navigate to user profile
  function handleCardClick() {
    // Navigate to user profile
    window.location.href = `/user/${username}`;
    dispatch('profileClick', id);
  }

  // Get correct avatar URL
  $: avatarUrl = avatar || `https://secure.gravatar.com/avatar/${id}?d=identicon&s=200`;
</script>

<div 
  class="profile-card {isDarkMode ? 'profile-card-dark' : ''} {compact ? 'compact' : ''}"
  on:click={handleCardClick}
  on:keydown={(e) => e.key === 'Enter' && handleCardClick()}
  tabindex="0"
  role="button"
>
  <div class="profile-content">
    <div class="avatar-container">
      {#if avatarUrl && typeof avatarUrl === 'string' && avatarUrl.startsWith('http')}
        <div class="image-wrapper">
          <img src={avatarUrl} alt={username} class="avatar-image" loading="lazy" />
        </div>
      {:else}
        <div class="avatar-fallback">
          <span class="avatar-initial">{displayName.charAt(0).toUpperCase()}</span>
        </div>
      {/if}
    </div>
    
    <div class="profile-info">
      <div class="name-container">
        <p class="display-name">{displayName}</p>
        {#if isVerified}
          <span class="verified-badge">
            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor">
              <path fill-rule="evenodd" d="M6.267 3.455a3.066 3.066 0 001.745-.723 3.066 3.066 0 013.976 0 3.066 3.066 0 001.745.723 3.066 3.066 0 012.812 2.812c.051.643.304 1.254.723 1.745a3.066 3.066 0 010 3.976 3.066 3.066 0 00-.723 1.745 3.066 3.066 0 01-2.812 2.812 3.066 3.066 0 00-1.745.723 3.066 3.066 0 01-3.976 0 3.066 3.066 0 00-1.745-.723 3.066 3.066 0 01-2.812-2.812 3.066 3.066 0 00-.723-1.745 3.066 3.066 0 010-3.976 3.066 3.066 0 00.723-1.745 3.066 3.066 0 012.812-2.812zm7.44 5.252a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
            </svg>
          </span>
        {/if}
      </div>
      
      <p class="username">@{username}</p>
      
      {#if showBio && bio}
        <p class="bio-text">{bio}</p>
      {/if}
      
      {#if showFollowerCount}
        <p class="follower-count">
          <span class="follower-number">{followerCount}</span> {followerCount === 1 ? 'follower' : 'followers'}
        </p>
      {/if}
    </div>
  </div>
</div>

<style>
  .profile-card {
    display: flex;
    align-items: center;
    justify-content: flex-start;
    padding: var(--space-2) var(--space-3);
    border-radius: var(--radius-md);
    cursor: pointer;
    transition: all 0.2s ease;
    gap: var(--space-2);
    background-color: transparent;
  }
  
  .profile-card:hover {
    background-color: var(--bg-hover);
    transform: translateY(-1px);
  }
  
  .profile-card-dark:hover {
    background-color: var(--dark-hover-bg);
  }
  
  .profile-card.compact {
    padding: var(--space-2);
  }
  
  .profile-content {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    flex: 1;
    min-width: 0;
  }
  
  .avatar-container {
    flex-shrink: 0;
    border: 1px solid transparent;
    width: 40px !important; /* Smaller avatar */
    height: 40px !important;
    border-radius: 50%;
    position: relative;
    display: flex;
    align-items: center;
    justify-content: center;
    overflow: hidden;
    transition: all 0.2s ease;
  }
  
  .profile-card:hover .avatar-container {
    border-color: var(--color-primary);
  }
  
  .avatar-image {
    width: 100%;
    height: 100%;
    object-fit: cover;
    object-position: center;
  }
  
  .avatar-fallback {
    width: 100%;
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
    background-color: var(--color-primary-light);
    color: white;
  }
  
  .avatar-initial {
    font-size: 1.2rem;
    font-weight: bold;
  }
  
  .profile-info {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 0px;
  }
  
  .name-container {
    display: flex;
    align-items: center;
    gap: var(--space-1);
  }
  
  .display-name {
    font-weight: var(--font-weight-bold);
    color: var(--text-primary);
    margin: 0;
    font-size: var(--font-size-sm);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  
  .username {
    color: var(--text-secondary);
    font-size: var(--font-size-xs);
    margin: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  
  .bio-text {
    color: var(--text-secondary);
    font-size: var(--font-size-xs);
    margin: var(--space-1) 0 0 0;
    overflow: hidden;
    text-overflow: ellipsis;
    display: -webkit-box;
    -webkit-box-orient: vertical;
    -webkit-line-clamp: 1;
    line-height: 1.3;
  }
  
  .follower-count {
    color: var(--text-tertiary);
    font-size: var(--font-size-xs);
    margin: 0;
  }
  
  .follower-number {
    font-weight: var(--font-weight-semibold);
  }
  
  .verified-badge {
    display: inline-flex;
    color: var(--color-primary);
    width: 14px;
    height: 14px;
  }
</style> 