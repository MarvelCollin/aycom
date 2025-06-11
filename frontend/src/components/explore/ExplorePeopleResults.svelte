<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { useTheme } from '../../hooks/useTheme';
  import { useAuth } from '../../hooks/useAuth';
  import { createLoggerWithPrefix } from '../../utils/logger';
  import Button from '../common/Button.svelte';
  import LoadingSkeleton from '../common/LoadingSkeleton.svelte';
  import Pagination from '../common/Pagination.svelte';
  import PerPageSelector from '../common/PerPageSelector.svelte';
  import { formatNumber } from '../../utils/common';
  
  const logger = createLoggerWithPrefix('ExplorePeopleResults');
  const dispatch = createEventDispatcher();
  const { theme } = useTheme();
  const { getAuthState } = useAuth();
  
  // Reactive declarations
  $: isDarkMode = $theme === 'dark';
  $: authState = getAuthState ? getAuthState() : { user_id: null, is_authenticated: false };
  
  // Props
  export let peopleResults: Array<{
    id: string;
    username: string;
    displayName: string;
    avatar: string | null;
    bio?: string;
    isVerified: boolean;
    followerCount: number;
    isFollowing: boolean;
  }> = [];
  export let isLoading = false;
  export let peoplePerPage = 20;
  export let currentPage = 1;
  export let totalCount = 0;
  
  // Calculate total pages
  $: totalPages = Math.max(1, Math.ceil(totalCount / peoplePerPage));
  
  // Handle page change
  function handlePageChange(event: CustomEvent<number>) {
    logger.debug('Changing page', { page: event.detail });
    currentPage = event.detail;
    dispatch('pageChange', event.detail);
  }
  
  // Page size options
  const perPageOptions = [25, 30, 35];
  
  // Handle per page change
  function handlePerPageChange(event: CustomEvent<number>) {
    const newValue = event.detail;
    logger.debug('Changing results per page', { value: newValue });
    peoplePerPage = newValue;
    dispatch('peoplePerPageChange', newValue);
  }
  
  // Handle follow user
  function handleFollow(userId: string) {
    logger.debug('Follow request initiated', { userId });
    
    // Find the user in results
    const userIndex = peopleResults.findIndex(user => user.id === userId);
    if (userIndex !== -1) {
      // Create a copy of the array with the updated user
      peopleResults = [
        ...peopleResults.slice(0, userIndex),
        {
          ...peopleResults[userIndex],
          isFollowing: !peopleResults[userIndex].isFollowing
        },
        ...peopleResults.slice(userIndex + 1)
      ];
    }
    
    // Dispatch event to parent
    dispatch('follow', userId);
  }
  
  // Handle profile click
  function handleProfileClick(userId: string) {
    logger.debug('Profile click', { userId });
    dispatch('profileClick', userId);
    
    // Navigate to user profile
    const user = peopleResults.find(u => u.id === userId);
    if (user) {
      window.location.href = `/user/${user.username}`;
    }
  }
  
  // Log when people results change
  $: {
    if (!isLoading) {
      if (peopleResults && peopleResults.length > 0) {
        logger.debug('People results loaded', { count: peopleResults.length });
        console.log('People results data:', peopleResults);
        // Check for any issues with the data structure
        for (let i = 0; i < peopleResults.length; i++) {
          const person = peopleResults[i];
          console.log(`Person ${i}:`, {
            id: person.id || 'MISSING ID',
            username: person.username || 'MISSING USERNAME',
            displayName: person.displayName || 'MISSING DISPLAY NAME',
            avatar: person.avatar,
            bio: person.bio,
            isVerified: person.isVerified,
            followerCount: person.followerCount,
            isFollowing: person.isFollowing
          });
          
          // Check for potential issues
          if (!person.id) console.error('Person missing ID!');
          if (!person.username) console.error('Person missing username!');
          if (!person.displayName) console.error('Person missing display name!');
        }
      } else {
        logger.debug('No people results found');
        console.log('Empty people results array:', peopleResults);
      }
    }
  }
</script>

<div class="people-results {isDarkMode ? 'people-results-dark' : ''}">
  {#if isLoading}
    <div class="people-loading">
      {#each Array(5) as _, i}
        <div class="twitter-profile-skeleton"></div>
      {/each}
    </div>
  {:else if peopleResults.length > 0}
    <div class="twitter-people-list">
      {#each peopleResults as user (user.id)}
        <div class="twitter-profile-card">
          <div class="twitter-profile-avatar" on:click={() => handleProfileClick(user.id)}>
            {#if user.avatar}
              <img src={user.avatar} alt={user.username} class="twitter-profile-img" />
            {:else}
              <div class="twitter-profile-placeholder">{user.displayName.charAt(0)}</div>
            {/if}
          </div>
          
          <div class="twitter-profile-content">
            <div class="twitter-profile-header">
              <div class="twitter-profile-info" on:click={() => handleProfileClick(user.id)}>
                <div class="twitter-profile-name-row">
                  <span class="twitter-profile-name">{user.displayName}</span>
                  {#if user.isVerified}
                    <span class="twitter-verified-badge">
                      <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" width="16" height="16" fill="var(--color-primary)">
                        <path d="M22.5 12.5c0-1.58-.875-2.95-2.148-3.6.154-.435.238-.905.238-1.4 0-2.21-1.71-3.998-3.818-3.998-.47 0-.92.084-1.336.25C14.818 2.415 13.51 1.5 12 1.5s-2.816.917-3.437 2.25c-.415-.165-.866-.25-1.336-.25-2.11 0-3.818 1.79-3.818 4 0 .494.083.964.237 1.4-1.272.65-2.147 2.018-2.147 3.6 0 1.495.782 2.798 1.942 3.486-.02.17-.032.34-.032.514 0 2.21 1.708 4 3.818 4 .47 0 .92-.086 1.335-.25.62 1.334 1.926 2.25 3.437 2.25s2.818-.916 3.437-2.25c.415.163.865.248 1.336.248 2.11 0 3.818-1.79 3.818-4 0-.174-.012-.344-.033-.513 1.158-.687 1.943-1.99 1.943-3.484zm-6.616-3.334l-4.334 6.5c-.145.217-.382.334-.625.334-.143 0-.288-.04-.416-.126l-.115-.094-2.415-2.415c-.293-.293-.293-.768 0-1.06s.768-.294 1.06 0l1.77 1.767 3.825-5.74c.23-.345.696-.436 1.04-.207.346.23.44.696.21 1.04z" />
                      </svg>
                    </span>
                  {/if}
                </div>
                <span class="twitter-profile-username">@{user.username}</span>
                {#if user.bio}
                  <p class="twitter-profile-bio">{user.bio}</p>
                {/if}
              </div>
              
              <div class="twitter-profile-follow">
                <button 
                  class="twitter-follow-button {user.isFollowing ? 'following' : ''}"
                  on:click={() => handleFollow(user.id)}
                >
                  {user.isFollowing ? 'Following' : 'Follow'}
                </button>
              </div>
            </div>
            
            {#if user.followerCount > 0}
              <div class="twitter-profile-stats">
                <span class="twitter-profile-followers">
                  <strong>{formatNumber(user.followerCount)}</strong> followers
                </span>
              </div>
            {/if}
          </div>
        </div>
      {/each}
    </div>
    
    <div class="twitter-people-footer">
      <div class="twitter-pagination-wrapper">
        <Pagination 
          totalItems={totalCount} 
          perPage={peoplePerPage} 
          currentPage={currentPage} 
          on:pageChange={handlePageChange}
        />
      </div>
      
      <div class="twitter-perpage-wrapper">
        <PerPageSelector 
          perPage={peoplePerPage} 
          options={[10, 20, 50]} 
          on:perPageChange={handlePerPageChange}
        />
      </div>
    </div>
  {:else}
    <div class="twitter-people-empty">
      <div class="twitter-people-empty-icon">
        <svg xmlns="http://www.w3.org/2000/svg" width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
          <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
          <circle cx="9" cy="7" r="4"></circle>
          <path d="M23 21v-2a4 4 0 0 0-3-3.87"></path>
          <path d="M16 3.13a4 4 0 0 1 0 7.75"></path>
        </svg>
      </div>
      <h3 class="twitter-people-empty-title">No users found</h3>
      <p class="twitter-people-empty-text">Try adjusting your search or filters</p>
    </div>
  {/if}
</div>

<style>
  .people-results {
    width: 100%;
  }
  
  .people-loading {
    display: flex;
    flex-direction: column;
    gap: 12px;
    padding: 0 0 16px 0;
  }
  
  .twitter-profile-skeleton {
    height: 80px;
    background: linear-gradient(
      90deg,
      var(--bg-tertiary) 0%,
      var(--bg-secondary) 50%,
      var(--bg-tertiary) 100%
    );
    border-radius: 16px;
    animation: pulse 1.5s ease-in-out infinite;
  }
  
  .people-results-dark .twitter-profile-skeleton {
    background: linear-gradient(
      90deg,
      var(--dark-bg-tertiary) 0%,
      var(--dark-bg-secondary) 50%,
      var(--dark-bg-tertiary) 100%
    );
  }
  
  @keyframes pulse {
    0% {
      opacity: 0.6;
    }
    50% {
      opacity: 0.3;
    }
    100% {
      opacity: 0.6;
    }
  }
  
  .twitter-people-list {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }
  
  .twitter-profile-card {
    display: flex;
    gap: 12px;
    padding: 12px;
    border-radius: 16px;
    transition: background-color 0.2s;
  }
  
  .twitter-profile-card:hover {
    background-color: var(--hover-bg);
  }
  
  .people-results-dark .twitter-profile-card:hover {
    background-color: var(--dark-hover-bg);
  }
  
  .twitter-profile-avatar {
    width: 48px;
    height: 48px;
    border-radius: 50%;
    overflow: hidden;
    flex-shrink: 0;
    cursor: pointer;
  }
  
  .twitter-profile-img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }
  
  .twitter-profile-placeholder {
    width: 100%;
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
    background-color: var(--bg-tertiary);
    color: var(--text-secondary);
    font-weight: bold;
    font-size: 20px;
  }
  
  .people-results-dark .twitter-profile-placeholder {
    background-color: var(--dark-bg-tertiary);
    color: var(--dark-text-secondary);
  }
  
  .twitter-profile-content {
    flex: 1;
    min-width: 0;
  }
  
  .twitter-profile-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 4px;
  }
  
  .twitter-profile-info {
    flex: 1;
    min-width: 0;
    cursor: pointer;
  }
  
  .twitter-profile-name-row {
    display: flex;
    align-items: center;
    gap: 4px;
    margin-bottom: 1px;
  }
  
  .twitter-profile-name {
    font-weight: bold;
    color: var(--text-primary);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  
  .people-results-dark .twitter-profile-name {
    color: var(--dark-text-primary);
  }
  
  .twitter-verified-badge {
    display: flex;
    align-items: center;
    flex-shrink: 0;
  }
  
  .twitter-profile-username {
    color: var(--text-secondary);
    font-size: 14px;
    display: block;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  
  .people-results-dark .twitter-profile-username {
    color: var(--dark-text-secondary);
  }
  
  .twitter-profile-bio {
    font-size: 14px;
    color: var(--text-primary);
    margin: 4px 0 0;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  
  .people-results-dark .twitter-profile-bio {
    color: var(--dark-text-primary);
  }
  
  .twitter-profile-follow {
    flex-shrink: 0;
    margin-left: 16px;
  }
  
  .twitter-follow-button {
    background-color: var(--text-primary);
    color: var(--bg-primary);
    border: none;
    border-radius: 9999px;
    padding: 6px 16px;
    font-size: 14px;
    font-weight: bold;
    cursor: pointer;
    transition: background-color 0.2s;
    white-space: nowrap;
  }
  
  .twitter-follow-button:hover {
    background-color: var(--text-primary);
    opacity: 0.9;
  }
  
  .twitter-follow-button.following {
    background-color: transparent;
    color: var(--text-primary);
    border: 1px solid var(--border-color);
  }
  
  .twitter-follow-button.following:hover {
    border-color: rgba(var(--color-danger-rgb), 0.4);
    color: var(--color-danger);
    background-color: rgba(var(--color-danger-rgb), 0.1);
  }
  
  .people-results-dark .twitter-follow-button {
    background-color: var(--dark-text-primary);
    color: var(--dark-bg-primary);
  }
  
  .people-results-dark .twitter-follow-button.following {
    background-color: transparent;
    color: var(--dark-text-primary);
    border: 1px solid var(--dark-border-color);
  }
  
  .twitter-profile-stats {
    font-size: 14px;
    color: var(--text-secondary);
    margin-top: 4px;
  }
  
  .people-results-dark .twitter-profile-stats {
    color: var(--dark-text-secondary);
  }
  
  .twitter-profile-followers strong {
    color: var(--text-primary);
    font-weight: bold;
  }
  
  .people-results-dark .twitter-profile-followers strong {
    color: var(--dark-text-primary);
  }
  
  .twitter-people-footer {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-top: 20px;
    flex-wrap: wrap;
    gap: 16px;
  }
  
  .twitter-pagination-wrapper {
    flex: 1;
  }
  
  .twitter-perpage-wrapper {
    flex-shrink: 0;
  }
  
  .twitter-people-empty {
    padding: 40px 16px;
    text-align: center;
  }
  
  .twitter-people-empty-icon {
    color: var(--text-secondary);
    margin-bottom: 16px;
  }
  
  .people-results-dark .twitter-people-empty-icon {
    color: var(--dark-text-secondary);
  }
  
  .twitter-people-empty-title {
    font-size: 20px;
    font-weight: bold;
    color: var(--text-primary);
    margin: 0 0 8px;
  }
  
  .people-results-dark .twitter-people-empty-title {
    color: var(--dark-text-primary);
  }
  
  .twitter-people-empty-text {
    font-size: 15px;
    color: var(--text-secondary);
    margin: 0;
  }
  
  .people-results-dark .twitter-people-empty-text {
    color: var(--dark-text-secondary);
  }
  
  @media (max-width: 640px) {
    .twitter-profile-bio {
      display: none;
    }
    
    .twitter-follow-button {
      padding: 4px 12px;
      font-size: 13px;
    }
    
    .twitter-people-footer {
      flex-direction: column;
      align-items: center;
      gap: 12px;
    }
    
    .twitter-pagination-wrapper {
      width: 100%;
      display: flex;
      justify-content: center;
    }
  }
</style> 