<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { useTheme } from '../../hooks/useTheme';
  import { useAuth } from '../../hooks/useAuth';
  import { createLoggerWithPrefix } from '../../utils/logger';
  import Button from '../common/Button.svelte';
  
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
  export let peoplePerPage = 25;
  export let currentPage = 1;
  export let totalCount = 0;
  
  // Calculate total pages
  $: totalPages = Math.max(1, Math.ceil(totalCount / peoplePerPage));
  
  // Handle page change
  function changePage(page: number) {
    logger.debug('Changing page', { page });
    currentPage = page;
    dispatch('pageChange', page);
  }
  
  // Page size options
  const perPageOptions = [25, 30, 35];
  
  // Handle per page change
  function handlePerPageChange(e) {
    const newValue = parseInt(e.target.value);
    logger.debug('Changing results per page', { value: newValue });
    peoplePerPage = newValue;
    dispatch('peoplePerPageChange', newValue);
  }
  
  // Handle follow user
  function handleFollowUser(userId: string) {
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

<div class="people-results-container">
  <div class="people-header">
    <h2 class="people-title">People</h2>
    
    {#if totalCount > 0}
      <div class="results-count">
        <span>{totalCount} {totalCount === 1 ? 'result' : 'results'}</span>
      </div>
    {/if}
  </div>
  
  {#if isLoading}
    <div class="people-grid animate-pulse">
      {#each Array(peoplePerPage > 9 ? 9 : peoplePerPage) as _, i}
        <div class="profile-card-skeleton">
          <div class="flex items-center space-x-3 w-full">
            <div class="skeleton-avatar"></div>
            <div class="skeleton-content">
              <div class="skeleton-name"></div>
              <div class="skeleton-username"></div>
              <div class="skeleton-bio"></div>
            </div>
            <div class="skeleton-button"></div>
          </div>
        </div>
      {/each}
    </div>
  {:else if peopleResults.length === 0}
    <div class="empty-state">
      <div class="empty-icon">👤</div>
      <p class="empty-message">No users found matching your search criteria</p>
      <p class="empty-tip">Try adjusting your search or filters</p>
    </div>
  {:else}
    <div class="people-grid">
      {#each peopleResults as person (person.id)}
        <div class="profile-card {isDarkMode ? 'profile-card-dark' : ''}">
          <div class="profile-content">
            <div class="profile-avatar" on:click={() => handleProfileClick(person.id)}>
              {#if person.avatar}
                <img src={person.avatar} alt={person.displayName} class="avatar-image" />
              {:else}
                <div class="avatar-fallback">
                  <span>{person.displayName.charAt(0).toUpperCase()}</span>
                </div>
              {/if}
            </div>
            
            <div class="profile-info" on:click={() => handleProfileClick(person.id)}>
              <div class="profile-name-container">
                <h3 class="profile-name">
                  {person.displayName}
                  {#if person.isVerified}
                    <span class="verified-badge">
                      <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" class="verified-icon">
                        <path d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
                      </svg>
                    </span>
                  {/if}
                </h3>
                <p class="profile-username">@{person.username}</p>
              </div>
              
              {#if person.bio}
                <p class="profile-bio">{person.bio}</p>
              {/if}
              
              <div class="profile-stats">
                <span class="follower-count">
                  <strong>{person.followerCount}</strong> {person.followerCount === 1 ? 'follower' : 'followers'}
                </span>
              </div>
            </div>
            
            {#if authState.is_authenticated && person.id !== authState.user_id}
              <div class="profile-action">
                <button 
                  class="follow-button {person.isFollowing ? 'following' : ''}"
                  on:click|stopPropagation={() => handleFollowUser(person.id)}
                >
                  {#if person.isFollowing}
                    <span class="follow-icon">✓</span> Following
                  {:else}
                    <span class="follow-icon">+</span> Follow
                  {/if}
                </button>
              </div>
            {/if}
          </div>
        </div>
      {/each}
    </div>
    
    {#if totalPages > 1}
      <div class="pagination-container">
        <div class="pagination-select">
          <label for="perPage">Show:</label>
          <select 
            id="perPage"
            class="per-page-select {isDarkMode ? 'per-page-select-dark' : ''}"
            bind:value={peoplePerPage}
            on:change={handlePerPageChange}
          >
            {#each perPageOptions as option}
              <option value={option}>{option} per page</option>
            {/each}
          </select>
        </div>
        
        <div class="pagination-controls">
          <button 
            class="pagination-button {currentPage === 1 ? 'disabled' : ''}"
            disabled={currentPage === 1}
            on:click={() => changePage(1)}
          >
            &laquo;
          </button>
          
          <button 
            class="pagination-button {currentPage === 1 ? 'disabled' : ''}"
            disabled={currentPage === 1}
            on:click={() => changePage(currentPage - 1)}
          >
            &lt;
          </button>
          
          <div class="pagination-pages">
            {#if totalPages <= 5}
              {#each Array(totalPages) as _, i}
                <button 
                  class="pagination-page {i + 1 === currentPage ? 'active' : ''}"
                  on:click={() => changePage(i + 1)}
                >
                  {i + 1}
                </button>
              {/each}
            {:else}
              {#if currentPage > 3}
                <button class="pagination-page" on:click={() => changePage(1)}>1</button>
                {#if currentPage > 4}
                  <span class="pagination-ellipsis">...</span>
                {/if}
              {/if}
              
              {#each Array(Math.min(5, totalPages)) as _, i}
                {@const pageNum = currentPage <= 3 ? i + 1 : 
                            currentPage >= totalPages - 2 ? totalPages - 4 + i : 
                            currentPage - 2 + i}
                {#if pageNum > 0 && pageNum <= totalPages}
                  <button 
                    class="pagination-page {pageNum === currentPage ? 'active' : ''}"
                    on:click={() => changePage(pageNum)}
                  >
                    {pageNum}
                  </button>
                {/if}
              {/each}
              
              {#if currentPage < totalPages - 2}
                {#if currentPage < totalPages - 3}
                  <span class="pagination-ellipsis">...</span>
                {/if}
                <button class="pagination-page" on:click={() => changePage(totalPages)}>{totalPages}</button>
              {/if}
            {/if}
          </div>
          
          <button 
            class="pagination-button {currentPage === totalPages ? 'disabled' : ''}"
            disabled={currentPage === totalPages}
            on:click={() => changePage(currentPage + 1)}
          >
            &gt;
          </button>
          
          <button 
            class="pagination-button {currentPage === totalPages ? 'disabled' : ''}"
            disabled={currentPage === totalPages}
            on:click={() => changePage(totalPages)}
          >
            &raquo;
          </button>
        </div>
      </div>
    {/if}
  {/if}
</div>

<style>
  .people-results-container {
    width: 100%;
    margin: 0 auto;
    padding: 1rem 0;
  }
  
  .people-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1.5rem;
  }
  
  .people-title {
    font-size: 1.5rem;
    font-weight: 700;
    margin: 0;
  }
  
  .results-count {
    font-size: 0.9rem;
    color: #666;
    padding: 0.25rem 0.5rem;
    background-color: #f5f5f5;
    border-radius: 1rem;
  }
  
  .people-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
    gap: 1.5rem;
    margin-bottom: 2rem;
  }
  
  .profile-card {
    border-radius: 0.75rem;
    padding: 1.25rem;
    background-color: #fff;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
    transition: transform 0.2s ease, box-shadow 0.2s ease;
    border: 1px solid #eaeaea;
  }
  
  .profile-card:hover {
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
  }
  
  .profile-card-dark {
    background-color: #1a1a1a;
    border-color: #333;
  }
  
  .profile-content {
    display: flex;
    align-items: flex-start;
    width: 100%;
  }
  
  .profile-avatar {
    flex: 0 0 48px;
    width: 48px;
    height: 48px;
    border-radius: 50%;
    overflow: hidden;
    margin-right: 1rem;
    cursor: pointer;
  }
  
  .avatar-image {
    width: 100%;
    height: 100%;
    object-fit: cover;
    border-radius: 50%;
  }
  
  .avatar-fallback {
    width: 100%;
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
    background-color: #3b82f6;
    color: white;
    font-weight: 600;
  }
  
  .profile-info {
    flex: 1;
    min-width: 0;
    cursor: pointer;
  }
  
  .profile-name-container {
    margin-bottom: 0.25rem;
  }
  
  .profile-name {
    font-size: 1.1rem;
    font-weight: 600;
    margin: 0 0 0.125rem 0;
    display: flex;
    align-items: center;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  
  .verified-badge {
    display: inline-flex;
    margin-left: 0.25rem;
    color: #3b82f6;
  }
  
  .verified-icon {
    width: 1rem;
    height: 1rem;
  }
  
  .profile-username {
    font-size: 0.9rem;
    color: #666;
    margin: 0;
  }
  
  .profile-bio {
    font-size: 0.9rem;
    margin: 0.5rem 0;
    line-height: 1.4;
    overflow: hidden;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    color: #333;
  }
  
  :global(.dark) .profile-bio {
    color: #ddd;
  }
  
  :global(.dark) .profile-username {
    color: #aaa;
  }
  
  .profile-stats {
    font-size: 0.85rem;
    color: #666;
    margin-top: 0.25rem;
  }
  
  :global(.dark) .profile-stats {
    color: #aaa;
  }
  
  .profile-action {
    margin-left: 0.75rem;
    flex-shrink: 0;
  }
  
  .follow-button {
    font-size: 0.875rem;
    font-weight: 500;
    padding: 0.375rem 0.75rem;
    border-radius: 1.5rem;
    background-color: #3b82f6;
    color: white;
    border: none;
    cursor: pointer;
    display: flex;
    align-items: center;
    transition: background-color 0.2s ease;
  }
  
  .follow-button:hover {
    background-color: #2563eb;
  }
  
  .follow-button.following {
    background-color: transparent;
    color: #3b82f6;
    border: 1px solid #3b82f6;
  }
  
  .follow-button.following:hover {
    background-color: rgba(59, 130, 246, 0.1);
  }
  
  .follow-icon {
    margin-right: 0.25rem;
    font-size: 0.75rem;
  }
  
  .empty-state {
    text-align: center;
    padding: 3rem 1rem;
  }
  
  .empty-icon {
    font-size: 3rem;
    margin-bottom: 1rem;
    opacity: 0.5;
  }
  
  .empty-message {
    font-size: 1.1rem;
    font-weight: 500;
    margin-bottom: 0.5rem;
    color: #555;
  }
  
  .empty-tip {
    font-size: 0.9rem;
    color: #777;
  }
  
  .profile-card-skeleton {
    padding: 1.25rem;
    border-radius: 0.75rem;
    background-color: #fff;
    border: 1px solid #eaeaea;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
    width: 100%;
  }
  
  .flex {
    display: flex;
  }
  
  .items-center {
    align-items: center;
  }
  
  .space-x-3 > * + * {
    margin-left: 0.75rem;
  }
  
  .w-full {
    width: 100%;
  }
  
  .skeleton-avatar {
    width: 48px;
    height: 48px;
    border-radius: 50%;
    background-color: #e5e7eb;
    flex-shrink: 0;
  }
  
  .skeleton-content {
    flex: 1;
    min-width: 0;
  }
  
  .skeleton-name {
    height: 1.25rem;
    background-color: #e5e7eb;
    border-radius: 0.25rem;
    width: 60%;
    margin-bottom: 0.5rem;
  }
  
  .skeleton-username {
    height: 1rem;
    background-color: #e5e7eb;
    border-radius: 0.25rem;
    width: 40%;
    margin-bottom: 0.75rem;
  }
  
  .skeleton-bio {
    height: 1rem;
    background-color: #e5e7eb;
    border-radius: 0.25rem;
    width: 80%;
  }
  
  .skeleton-button {
    height: 2rem;
    width: 5rem;
    background-color: #e5e7eb;
    border-radius: 1.5rem;
    flex-shrink: 0;
  }
  
  .animate-pulse {
    animation: pulse 1.5s cubic-bezier(0.4, 0, 0.6, 1) infinite;
  }
  
  @keyframes pulse {
    0%, 100% {
      opacity: 1;
    }
    50% {
      opacity: 0.5;
    }
  }
  
  .pagination-container {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-top: 2rem;
    padding: 1rem 0;
    border-top: 1px solid #eaeaea;
  }
  
  :global(.dark) .pagination-container {
    border-top-color: #333;
  }
  
  .pagination-select {
    display: flex;
    align-items: center;
  }
  
  .pagination-select label {
    margin-right: 0.5rem;
    font-size: 0.875rem;
    color: #555;
  }
  
  :global(.dark) .pagination-select label {
    color: #aaa;
  }
  
  .per-page-select {
    padding: 0.375rem 0.75rem;
    border-radius: 0.375rem;
    border: 1px solid #d1d5db;
    background-color: white;
    font-size: 0.875rem;
    color: #4b5563;
    cursor: pointer;
  }
  
  .per-page-select:focus {
    outline: none;
    border-color: #3b82f6;
    box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.3);
  }
  
  .per-page-select-dark {
    border-color: #4b5563;
    background-color: #1f2937;
    color: #e5e7eb;
  }
  
  .pagination-controls {
    display: flex;
    align-items: center;
  }
  
  .pagination-button {
    padding: 0.5rem 0.75rem;
    border: 1px solid #d1d5db;
    background-color: white;
    color: #4b5563;
    cursor: pointer;
    margin: 0 0.125rem;
    transition: all 0.2s ease;
  }
  
  .pagination-button:hover:not(.disabled) {
    background-color: #f3f4f6;
    border-color: #9ca3af;
    color: #111827;
  }
  
  .pagination-button.disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
  
  :global(.dark) .pagination-button {
    border-color: #4b5563;
    background-color: #1f2937;
    color: #e5e7eb;
  }
  
  :global(.dark) .pagination-button:hover:not(.disabled) {
    background-color: #374151;
    border-color: #6b7280;
  }
  
  .pagination-pages {
    display: flex;
    align-items: center;
    margin: 0 0.25rem;
  }
  
  .pagination-page {
    padding: 0.5rem 0.75rem;
    margin: 0 0.125rem;
    border: 1px solid #d1d5db;
    background-color: white;
    color: #4b5563;
    cursor: pointer;
    transition: all 0.2s ease;
    min-width: 2.5rem;
    text-align: center;
  }
  
  .pagination-page:hover:not(.active) {
    background-color: #f3f4f6;
    border-color: #9ca3af;
    color: #111827;
  }
  
  .pagination-page.active {
    background-color: #3b82f6;
    border-color: #3b82f6;
    color: white;
    font-weight: 500;
  }
  
  :global(.dark) .pagination-page {
    border-color: #4b5563;
    background-color: #1f2937;
    color: #e5e7eb;
  }
  
  :global(.dark) .pagination-page:hover:not(.active) {
    background-color: #374151;
    border-color: #6b7280;
  }
  
  :global(.dark) .pagination-page.active {
    background-color: #3b82f6;
    border-color: #3b82f6;
    color: white;
  }
  
  .pagination-ellipsis {
    display: inline-block;
    padding: 0.5rem 0.375rem;
    color: #4b5563;
    font-weight: 500;
  }
  
  :global(.dark) .pagination-ellipsis {
    color: #e5e7eb;
  }
</style> 