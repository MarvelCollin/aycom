<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { useTheme } from '../../hooks/useTheme';
  
  const dispatch = createEventDispatcher();
  const { theme } = useTheme();
  
  // Reactive declarations
  $: isDarkMode = $theme === 'dark';
  
  // Define a more flexible community type to handle different API response formats
  type CommunityData = {
    id: string;
    name: string;
    description?: string;
    logo?: string | null;
    logo_url?: string | null;
    avatar?: string | null;
    memberCount?: number;
    member_count?: number;
    isJoined?: boolean;
    is_joined?: boolean;
    isPending?: boolean;
    is_pending?: boolean;
  };
  
  // Props
  export let community: CommunityData;
  
  // Computed properties to handle different naming conventions
  $: logo = community.logo || community.logo_url || community.avatar || null;
  $: memberCount = community.memberCount || community.member_count || 0;
  $: isJoined = community.isJoined || community.is_joined || false;
  $: isPending = community.isPending || community.is_pending || false;
  
  // Handle join request
  function handleJoinRequest() {
    dispatch('joinRequest', {communityId: community.id});
  }
</script>

<div class="community-card {isDarkMode ? 'community-card-dark' : ''}">
  <div class="community-card-content">
    <a href={`/community/${community.id}`} class="community-logo-container">
      <div class="community-logo {isDarkMode ? 'community-logo-dark' : ''}">
        {#if typeof logo === 'string' && logo.startsWith('http')}
          <img src={logo} alt={community.name} class="community-logo-img" />
        {:else}
          <div class="community-logo-placeholder">
            <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
              <circle cx="9" cy="7" r="4"></circle>
              <path d="M23 21v-2a4 4 0 0 0-3-3.87"></path>
              <path d="M16 3.13a4 4 0 0 1 0 7.75"></path>
            </svg>
          </div>
        {/if}
      </div>
    </a>
    <div class="community-info">
      <a href={`/community/${community.id}`} class="community-link">
        <h3 class="community-name {isDarkMode ? 'community-name-dark' : ''}">{community.name}</h3>
        <p class="community-description {isDarkMode ? 'community-description-dark' : ''}">{community.description || 'No description provided'}</p>
        <div class="community-stats">
          <span class="community-member-count">
            <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
              <circle cx="9" cy="7" r="4"></circle>
              <path d="M23 21v-2a4 4 0 0 0-3-3.87"></path>
              <path d="M16 3.13a4 4 0 0 1 0 7.75"></path>
            </svg>
            {memberCount} {memberCount === 1 ? 'member' : 'members'}
          </span>
        </div>
      </a>
    </div>
    <div class="community-action">
      {#if isJoined}
        <span class="joined-badge {isDarkMode ? 'joined-badge-dark' : ''}">
          <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M20 6L9 17l-5-5"></path>
          </svg>
          Joined
        </span>
      {:else if isPending}
        <span class="pending-badge {isDarkMode ? 'pending-badge-dark' : ''}">
          <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="12" cy="12" r="10"></circle>
            <polyline points="12 6 12 12 16 14"></polyline>
          </svg>
          Pending
        </span>
      {:else}
        <button 
          class="join-button {isDarkMode ? 'join-button-dark' : ''}"
          on:click={handleJoinRequest}
        >
          Request to Join
        </button>
      {/if}
    </div>
  </div>
</div>

<style>
  .community-card {
    background-color: var(--bg-secondary);
    border-radius: var(--radius-lg);
    border: 1px solid var(--border-color);
    margin-bottom: var(--space-3);
    transition: transform var(--transition-normal), box-shadow var(--transition-normal);
    overflow: hidden;
  }
  
  .community-card:hover {
    transform: translateY(-2px);
    box-shadow: var(--shadow-md);
  }
  
  .community-card-dark {
    background-color: var(--dark-bg-secondary);
    border-color: var(--border-color-dark);
  }
  
  .community-card-content {
    padding: var(--space-4);
    display: flex;
    align-items: center;
  }
  
  .community-logo-container {
    flex-shrink: 0;
    margin-right: var(--space-3);
    text-decoration: none;
  }
  
  .community-logo {
    width: 60px;
    height: 60px;
    border-radius: var(--radius-lg);
    background-color: var(--bg-tertiary);
    display: flex;
    align-items: center;
    justify-content: center;
    overflow: hidden;
    border: 1px solid var(--border-color);
  }
  
  .community-logo-dark {
    background-color: var(--dark-bg-tertiary);
    border-color: var(--border-color-dark);
  }
  
  .community-logo-img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }
  
  .community-logo-placeholder {
    color: var(--text-tertiary);
  }
  
  .community-info {
    flex: 1;
    min-width: 0;
  }
  
  .community-link {
    text-decoration: none;
    display: block;
  }
  
  .community-name {
    font-weight: var(--font-weight-bold);
    font-size: var(--font-size-lg);
    margin: 0 0 var(--space-1);
    color: var(--text-primary);
  }
  
  .community-name-dark {
    color: var(--dark-text-primary);
  }
  
  .community-description {
    color: var(--text-secondary);
    font-size: var(--font-size-sm);
    margin: 0 0 var(--space-2);
    display: -webkit-box;
    -webkit-line-clamp: 2;
    line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
    line-height: 1.4;
    max-height: calc(1.4em * 2);
  }
  
  .community-description-dark {
    color: var(--dark-text-secondary);
  }
  
  .community-stats {
    display: flex;
    align-items: center;
  }
  
  .community-member-count {
    font-size: var(--font-size-xs);
    color: var(--text-tertiary);
    display: flex;
    align-items: center;
    gap: var(--space-1);
  }
  
  .community-action {
    margin-left: var(--space-3);
    flex-shrink: 0;
  }
  
  .join-button {
    padding: var(--space-2) var(--space-4);
    border-radius: var(--radius-full);
    background-color: var(--color-primary);
    color: white;
    font-weight: var(--font-weight-medium);
    font-size: var(--font-size-sm);
    border: none;
    cursor: pointer;
    transition: background-color var(--transition-fast), transform var(--transition-fast);
  }
  
  .join-button:hover {
    background-color: var(--color-primary-hover);
    transform: translateY(-1px);
  }
  
  .join-button-dark {
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.25);
  }
  
  .joined-badge, .pending-badge {
    display: flex;
    align-items: center;
    gap: var(--space-1);
    padding: var(--space-1) var(--space-3);
    border-radius: var(--radius-full);
    font-size: var(--font-size-xs);
    font-weight: var(--font-weight-medium);
    border: 1px solid var(--border-color);
    background-color: var(--bg-primary);
  }
  
  .joined-badge-dark, .pending-badge-dark {
    border-color: var(--border-color-dark);
    background-color: var(--dark-bg-primary);
  }
  
  .joined-badge {
    color: var(--color-success);
  }
  
  .pending-badge {
    color: var(--color-warning);
  }
  
  @media (max-width: 576px) {
    .community-logo {
      width: 48px;
      height: 48px;
    }
    
    .community-name {
      font-size: var(--font-size-md);
    }
    
    .community-description {
      font-size: var(--font-size-xs);
    }
  }
</style> 