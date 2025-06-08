<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { useTheme } from '../../hooks/useTheme';
  import { useAuth } from '../../hooks/useAuth';
  import { createLoggerWithPrefix } from '../../utils/logger';
  import Button from '../common/Button.svelte';
  import { formatStorageUrl } from '../../utils/common';
  import { getPublicUrl, SUPABASE_BUCKETS } from '../../utils/supabase';
  import type { IUser } from '../../interfaces/IUser';
  
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

  // Props - Accept either a full IUser object or the custom structure
  export let user: IUser | {
    id: string;
    name?: string;
    username?: string;
    avatar?: string;
    avatar_url?: string;
    profile_picture_url?: string;
    bio?: string;
    isVerified?: boolean;
    is_verified?: boolean;
    isFollowing?: boolean;
    is_following?: boolean;
    role?: string;
  };
  
  // Optional props
  export let showFollowButton = true;
  export let compact = false;
  
  // Computed values
  $: displayName = user.name || `User_${user.id.substring(0, 4)}`;
  $: username = user.username || `user_${user.id.substring(0, 4)}`;
  $: isFollowing = 'isFollowing' in user ? user.isFollowing : ('is_following' in user ? user.is_following : false);
  $: isVerified = 'isVerified' in user ? user.isVerified : ('is_verified' in user ? user.is_verified : false);
  $: avatarUrl = getProfilePictureUrl(user);
  
  // Function to get profile picture URL using Supabase
  function getProfilePictureUrl(user) {
    // Check all possible avatar field names
    const rawUrl = user.avatar_url || user.profile_picture_url || user.avatar || '';
    
    if (!rawUrl) return '';
    
    // If URL already contains supabase, it's already formatted
    if (rawUrl.includes('supabase')) {
      return rawUrl;
    }
    
    // If it's a path starting with /, use it directly
    if (rawUrl.startsWith('/')) {
      return getPublicUrl(SUPABASE_BUCKETS.MEDIA, `profiles${rawUrl}`);
    }
    
    // Extract filename and use it for Supabase path
    const parts = rawUrl.split('/');
    if (parts.length > 0) {
      const filename = parts[parts.length - 1];
      return getPublicUrl(SUPABASE_BUCKETS.MEDIA, `profiles/${filename}`);
    }
    
    // Fallback to original URL
    return rawUrl;
  }
  
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
        {#if isVerified}
          <span class="user-verified-badge">
            <CheckCircleIcon size="14" />
          </span>
        {/if}
      </h4>
      <p class="user-username">@{username}</p>
    </div>
    
    {#if ('bio' in user) && user.bio && !compact}
      <p class="user-bio">{user.bio}</p>
    {/if}
    
    {#if ('role' in user) && user.role}
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
    line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
    line-height: 1.4;
    /* Fallback for browsers that don't support line-clamp */
    max-height: calc(1.4em * 2);
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