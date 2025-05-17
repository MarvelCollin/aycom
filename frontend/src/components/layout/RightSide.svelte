<script lang="ts">
  import { onMount } from 'svelte';
  import { useTheme } from '../../hooks/useTheme';
  import MoreHorizontalIcon from 'svelte-feather-icons/src/icons/MoreHorizontalIcon.svelte';
  import type { ITrend, ISuggestedFollow } from '../../interfaces/ISocialMedia';
  import { getTrends } from '../../api/trends';
  import { getSuggestedUsers } from '../../api/suggestions';
  import { followUser, unfollowUser } from '../../api/user';
  import { toastStore } from '../../stores/toastStore';

  // Extend the ISuggestedFollow interface with our UI-specific properties
  interface ExtendedSuggestedFollow extends ISuggestedFollow {
    isFollowingLoading?: boolean;
  }

  export let isDarkMode = false;
  export let trends: ITrend[] = [];
  export let suggestedFollows: ExtendedSuggestedFollow[] = [];

  let isTrendsLoading = true;
  let isFollowSuggestionsLoading = true;

  onMount(async () => {
    fetchTrends();
    fetchSuggestedUsers();
  });

  async function fetchTrends() {
    isTrendsLoading = true;
    try {
      const fetchedTrends = await getTrends(5);
      trends = fetchedTrends || [];
    } catch (error) {
      console.error('Error loading trends:', error);
      toastStore.showToast('Failed to load trends', 'error');
      trends = [];
    } finally {
      isTrendsLoading = false;
    }
  }

  async function fetchSuggestedUsers() {
    isFollowSuggestionsLoading = true;
    try {
      const users = await getSuggestedUsers();
      suggestedFollows = users || [];
    } catch (error) {
      console.error('Error loading suggested users:', error);
      toastStore.showToast('Failed to load suggested users', 'error');
      suggestedFollows = [];
    } finally {
      isFollowSuggestionsLoading = false;
    }
  }

  async function handleFollowToggle(user: ExtendedSuggestedFollow) {
    if (!user) return;
    
    // Optimistic UI update
    const index = suggestedFollows.findIndex(u => u.username === user.username);
    if (index === -1) return;
    
    const isCurrentlyFollowing = suggestedFollows[index].isFollowing;
    
    // Create a copy of the array with the updated user
    const updatedSuggestedFollows = [...suggestedFollows];
    updatedSuggestedFollows[index] = {
      ...updatedSuggestedFollows[index],
      isFollowing: !isCurrentlyFollowing,
      isFollowingLoading: true
    };
    suggestedFollows = updatedSuggestedFollows;
    
    try {
      if (isCurrentlyFollowing) {
        await unfollowUser(user.username);
        toastStore.showToast(`Unfollowed @${user.username}`, 'success');
      } else {
        await followUser(user.username);
        toastStore.showToast(`Followed @${user.username}`, 'success');
      }
    } catch (error) {
      console.error('Error toggling follow:', error);
      toastStore.showToast(`Failed to ${isCurrentlyFollowing ? 'unfollow' : 'follow'} user`, 'error');
      
      // Revert the optimistic update
      const revertedSuggestedFollows = [...suggestedFollows];
      revertedSuggestedFollows[index] = {
        ...revertedSuggestedFollows[index],
        isFollowing: isCurrentlyFollowing,
        isFollowingLoading: false
      };
      suggestedFollows = revertedSuggestedFollows;
    } finally {
      // Update isFollowingLoading state
      const finalSuggestedFollows = [...suggestedFollows];
      finalSuggestedFollows[index] = {
        ...finalSuggestedFollows[index],
        isFollowingLoading: false
      };
      suggestedFollows = finalSuggestedFollows;
    }
  }
</script>

<div class="widgets-container">
  <!-- Trends Widget -->
  <div class="sidebar {isDarkMode ? 'sidebar-dark' : ''}">
    <div class="sidebar-section">
      <h3 class="sidebar-title">Trends for you</h3>
      
      {#if isTrendsLoading}
        <div class="trends-loading">
          <div class="trends-loading-spinner"></div>
        </div>
      {:else if trends.length === 0}
        <div class="trends-empty">
          <p>No trends available</p>
        </div>
      {:else}
        <div class="trends-list">
          {#each trends as trend, i}
            <div class="trend-item {isDarkMode ? 'trend-item-dark' : ''}">
              <div class="trend-header">
                <span class="trend-location">{trend.category || 'Trending'}</span>
                <button 
                  class="trend-more-options {isDarkMode ? 'trend-more-options-dark' : ''}" 
                  aria-label="More options"
                >
                  <MoreHorizontalIcon size="16" />
                </button>
              </div>
              <div class="trend-name">
                <a href={`/explore?q=${encodeURIComponent(trend.title)}`}>
                  {trend.title}
                </a>
              </div>
              <div class="trend-count">{trend.postCount || '0'} posts</div>
            </div>
          {/each}
        </div>
      {/if}
      
      <a href="/explore" class="trends-show-more {isDarkMode ? 'trends-show-more-dark' : ''}">
        Show more
      </a>
    </div>
  </div>

  <!-- Who to Follow Widget -->
  <div class="sidebar {isDarkMode ? 'sidebar-dark' : ''}">
    <div class="sidebar-section">
      <h3 class="sidebar-title">Who to follow</h3>
      
      {#if isFollowSuggestionsLoading}
        <div class="suggestions-loading">
          <div class="suggestions-loading-spinner"></div>
        </div>
      {:else if suggestedFollows.length === 0}
        <div class="suggestions-empty">
          <p>No suggestions available</p>
        </div>
      {:else}
        <div class="suggestions-list">
          {#each suggestedFollows as user, i}
            <div class="suggestion-item {isDarkMode ? 'suggestion-item-dark' : ''}">
              <div class="suggestion-avatar">
                <a href={`/user/${user.username}`}>
                  <img 
                    src={user.avatar || 'https://secure.gravatar.com/avatar/0?d=mp'} 
                    alt={user.displayName || user.username || 'User'} 
                  />
                </a>
              </div>
              <div class="suggestion-details">
                <a href={`/user/${user.username}`} class="suggestion-name">
                  {user.displayName || user.username || 'User'}
                </a>
                <a href={`/user/${user.username}`} class="suggestion-username">
                  @{user.username || 'user'}
                </a>
              </div>
              <div class="suggestion-action">
                <button 
                  class="follow-button {user.isFollowing ? 'following' : ''} {isDarkMode ? 'follow-button-dark' : ''}"
                  on:click={() => handleFollowToggle(user)}
                  disabled={user.isFollowingLoading}
                >
                  {#if user.isFollowingLoading}
                    <span class="loading-dot"></span>
                  {:else if user.isFollowing}
                    Following
                  {:else}
                    Follow
                  {/if}
                </button>
              </div>
            </div>
          {/each}
        </div>
      {/if}
      
      <a href="/connect" class="suggestions-show-more {isDarkMode ? 'suggestions-show-more-dark' : ''}">
        Show more
      </a>
    </div>
  </div>
</div>

<style>
  /* Loading indicators */
  .trends-loading, .suggestions-loading {
    display: flex;
    justify-content: center;
    align-items: center;
    padding: var(--space-4);
  }
  
  .trends-loading-spinner, .suggestions-loading-spinner {
    width: 20px;
    height: 20px;
    border: 2px solid var(--border-color);
    border-top: 2px solid var(--color-primary);
    border-radius: 50%;
    animation: spin 1s linear infinite;
  }
  
  @keyframes spin {
    0% { transform: rotate(0deg); }
    100% { transform: rotate(360deg); }
  }
  
  /* Empty states */
  .trends-empty, .suggestions-empty {
    padding: var(--space-4);
    color: var(--text-secondary);
    text-align: center;
    font-size: 14px;
  }
  
  /* Trend items */
  .trend-item {
    padding: var(--space-3);
    cursor: pointer;
    transition: background-color 0.2s;
    border-bottom: 1px solid var(--border-color);
  }
  
  .trend-item-dark {
    border-bottom: 1px solid var(--border-color-dark);
  }
  
  .trend-item:hover {
    background-color: var(--bg-hover);
  }
  
  .trend-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: var(--space-1);
  }
  
  .trend-location {
    font-size: 13px;
    color: var(--text-secondary);
  }
  
  .trend-more-options {
    color: var(--text-secondary);
    background: transparent;
    border: none;
    cursor: pointer;
    padding: var(--space-1);
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
  }
  
  .trend-more-options:hover {
    background-color: rgba(var(--color-primary-rgb), 0.1);
    color: var(--color-primary);
  }
  
  .trend-name {
    font-weight: 700;
    margin-bottom: var(--space-1);
  }
  
  .trend-name a {
    color: var(--text-primary);
    text-decoration: none;
  }
  
  .trend-count {
    font-size: 13px;
    color: var(--text-secondary);
  }
  
  /* Suggestion items */
  .suggestion-item {
    display: flex;
    align-items: center;
    padding: var(--space-3);
    border-bottom: 1px solid var(--border-color);
  }
  
  .suggestion-item-dark {
    border-bottom: 1px solid var(--border-color-dark);
  }
  
  .suggestion-avatar {
    width: 48px;
    height: 48px;
    margin-right: var(--space-3);
    border-radius: 50%;
    overflow: hidden;
  }
  
  .suggestion-avatar img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }
  
  .suggestion-details {
    flex: 1;
    min-width: 0;
    margin-right: var(--space-2);
  }
  
  .suggestion-name {
    font-weight: 700;
    color: var(--text-primary);
    text-decoration: none;
    display: block;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  
  .suggestion-username {
    font-size: 14px;
    color: var(--text-secondary);
    text-decoration: none;
    display: block;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  
  .suggestion-action {
    flex-shrink: 0;
  }
  
  .follow-button {
    background-color: var(--text-primary);
    color: var(--bg-primary);
    border: none;
    border-radius: 9999px;
    padding: 6px 16px;
    font-weight: 700;
    font-size: 14px;
    cursor: pointer;
    transition: background-color 0.2s;
  }
  
  .follow-button-dark {
    background-color: var(--text-primary-dark);
    color: var(--bg-primary-dark);
  }
  
  .follow-button:hover {
    opacity: 0.9;
  }
  
  .follow-button.following {
    background-color: transparent;
    color: var(--text-primary);
    border: 1px solid var(--border-color);
  }
  
  .follow-button-dark.following {
    color: var(--text-primary-dark);
    border: 1px solid var(--border-color-dark);
  }
  
  .loading-dot {
    display: inline-block;
    width: 14px;
    height: 14px;
    border: 2px solid currentColor;
    border-radius: 50%;
    border-top-color: transparent;
    animation: spin 0.8s linear infinite;
  }
  
  /* Show more links */
  .trends-show-more, .suggestions-show-more {
    display: block;
    padding: var(--space-3);
    color: var(--color-primary);
    text-decoration: none;
    font-size: 14px;
  }
  
  .trends-show-more:hover, .suggestions-show-more:hover {
    background-color: var(--bg-hover);
  }
</style>