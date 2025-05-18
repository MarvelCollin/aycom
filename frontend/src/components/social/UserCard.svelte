<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { useTheme } from '../../hooks/useTheme';
  import { useAuth } from '../../hooks/useAuth';
  import { createLoggerWithPrefix } from '../../utils/logger';
  import Button from '../common/Button.svelte';
  import { formatStorageUrl } from '../../utils/common';
  
  // Icons
  import UserIcon from 'svelte-feather-icons/src/icons/UserIcon.svelte';
  import CheckIcon from 'svelte-feather-icons/src/icons/CheckIcon.svelte';
  import PlusIcon from 'svelte-feather-icons/src/icons/PlusIcon.svelte';
  import CheckCircleIcon from 'svelte-feather-icons/src/icons/CheckCircleIcon.svelte';
  
  const logger = createLoggerWithPrefix('UserCard');
  const dispatch = createEventDispatcher();
  
  // Theme
  const { theme } = useTheme();
  $: isDarkMode = $theme === 'dark';

  // Props
  export let user: {
    id: string;
    name?: string;
    username?: string;
    avatar?: string;
    bio?: string;
    isVerified?: boolean;
    isFollowing?: boolean;
    role?: string;
  };
  
  // Optional props
  export let showFollowButton = true;
  export let compact = false;
  
  // Computed values
  $: displayName = user.name || `User_${user.id.substring(0, 4)}`;
  $: username = user.username || `user_${user.id.substring(0, 4)}`;
  $: isFollowing = user.isFollowing || false;
  $: avatarUrl = user.avatar ? formatStorageUrl(user.avatar) : '';
  
  // Handle click on user card
  function handleCardClick() {
    dispatch('click', { userId: user.id });
    window.location.href = `/user/${user.id}`;
  }
  
  // Handle follow/unfollow action
  async function handleFollowAction(event) {
    event.stopPropagation();
    
    try {
      isFollowing = !isFollowing;
      
      // In a real app, make API call here
      logger.debug(`${isFollowing ? 'Following' : 'Unfollowing'} user ${user.id}`);
      
      dispatch('followChange', { 
        userId: user.id, 
        isFollowing 
      });
    } catch (error) {
      // Revert state if error
      isFollowing = !isFollowing;
      logger.error(`Failed to ${isFollowing ? 'follow' : 'unfollow'} user:`, error);
    }
  }
</script>

<div 
  class="user-card {compact ? 'user-card-compact' : ''}"
  on:click={handleCardClick}
  on:keydown={(e) => e.key === 'Enter' && handleCardClick()}
  tabindex="0"
  role="button"
>
  <div class="user-avatar">
    {#if avatarUrl}
      <img src={avatarUrl} alt={displayName} />
    {:else}
      <div class="user-avatar-placeholder">
        <UserIcon size="24" />
      </div>
    {/if}
  </div>
  
  <div class="user-info">
    <div class="user-name-container">
      <h4 class="user-display-name">
        {displayName}
        {#if user.isVerified}
          <span class="user-verified-badge">
            <CheckCircleIcon size="14" />
          </span>
        {/if}
      </h4>
      <p class="user-username">@{username}</p>
    </div>
    
    {#if user.bio && !compact}
      <p class="user-bio">{user.bio}</p>
    {/if}
    
    {#if user.role}
      <div class="user-role-badge">
        {user.role}
      </div>
    {/if}
  </div>
  
  {#if showFollowButton}
    <div class="user-action" on:click|stopPropagation>
      <Button 
        variant={isFollowing ? "outlined" : "primary"} 
        size="small" 
        on:click={handleFollowAction}
        icon={isFollowing ? CheckIcon : PlusIcon}
      >
        {isFollowing ? 'Following' : 'Follow'}
      </Button>
    </div>
  {/if}
</div>

<style>
  .user-card {
    display: flex;
    align-items: center;
    padding: var(--space-3);
    border-radius: var(--radius-md);
    background-color: var(--bg-secondary);
    transition: background-color var(--transition-fast);
    cursor: pointer;
    text-decoration: none;
    color: var(--text-primary);
    gap: var(--space-3);
    border: 1px solid var(--border-color);
  }
  
  .user-card:hover {
    background-color: var(--bg-hover);
  }
  
  .user-card-compact {
    padding: var(--space-2);
  }
  
  .user-avatar {
    flex-shrink: 0;
    width: 48px;
    height: 48px;
    border-radius: 50%;
    overflow: hidden;
    border: 1px solid var(--border-color);
  }
  
  .user-card-compact .user-avatar {
    width: 40px;
    height: 40px;
  }
  
  .user-avatar img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }
  
  .user-avatar-placeholder {
    width: 100%;
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
    background-color: #333;
    color: var(--text-secondary);
  }
  
  .user-info {
    flex: 1;
    min-width: 0;
  }
  
  .user-name-container {
    display: flex;
    flex-direction: column;
  }
  
  .user-display-name {
    display: flex;
    align-items: center;
    gap: var(--space-1);
    font-weight: var(--font-weight-bold);
    margin: 0;
    font-size: var(--font-size-sm);
    text-overflow: ellipsis;
    white-space: nowrap;
    overflow: hidden;
  }
  
  .user-verified-badge {
    color: var(--color-primary);
    display: flex;
    align-items: center;
  }
  
  .user-username {
    color: var(--text-secondary);
    margin: 0;
    font-size: var(--font-size-xs);
    text-overflow: ellipsis;
    white-space: nowrap;
    overflow: hidden;
  }
  
  .user-bio {
    margin: var(--space-1) 0 0 0;
    font-size: var(--font-size-xs);
    color: var(--text-secondary);
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
    line-height: 1.4;
  }
  
  .user-role-badge {
    display: inline-block;
    background-color: var(--color-primary-light);
    color: var(--color-primary);
    padding: 0 var(--space-2);
    border-radius: var(--radius-full);
    font-size: var(--font-size-xs);
    margin-top: var(--space-1);
  }
  
  .user-action {
    flex-shrink: 0;
  }
</style> 