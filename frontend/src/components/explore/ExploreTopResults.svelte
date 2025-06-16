<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import ProfileCard from './ProfileCard.svelte';
  import ThreadCard from './ThreadCard.svelte';
  import { createLoggerWithPrefix } from '../../utils/logger';
  import { useTheme } from '../../hooks/useTheme';
  
  const logger = createLoggerWithPrefix('ExploreTopResults');
  const dispatch = createEventDispatcher();
  const { theme } = useTheme();
  
  // Reactive declarations
  $: isDarkMode = $theme === 'dark';
  
  // Props with more flexible type definitions to handle different naming conventions
  export let topProfiles: Array<{
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
  }> = [];
  
  export let topThreads: Array<{
    id: string;
    content: string;
    username: string;
    name: string;
    created_at: string;
    likes_count: number;
    replies_count: number;
    reposts_count: number;
    media?: Array<{
      type: string;
      url: string;
    }>;
    profile_picture_url?: string;
  }> = [];
  
  export let isLoading = false;
  
  // Handle profile click
  function handleProfileClick(event) {
    const userId = event.detail;
    logger.debug('Profile click', { userId });
    dispatch('profileClick', userId);
  }
  
  // Handle view all
  function handleViewAll(section: string) {
    logger.debug('View all clicked', { section });
    dispatch('viewAll', section);
  }
</script>

<div class="p-4">
  {#if isLoading}
    <div class="animate-pulse">
      <!-- People section skeleton -->
      <div class="mb-6">
        <div class="flex justify-between items-center mb-3">
          <div class="h-6 bg-gray-300 dark:bg-gray-700 rounded w-24"></div>
          <div class="h-4 bg-gray-300 dark:bg-gray-700 rounded w-16"></div>
        </div>
        
        <div class="profiles-grid">
          {#each Array(6) as _}
            <div class="skeleton-card">
              <div class="flex items-center gap-2">
                <div class="rounded-full bg-gray-300 dark:bg-gray-700 h-8 w-8"></div>
                <div class="flex-1 space-y-1">
                  <div class="h-3 bg-gray-300 dark:bg-gray-700 rounded w-3/4"></div>
                  <div class="h-2 bg-gray-300 dark:bg-gray-700 rounded w-1/2"></div>
                </div>
              </div>
            </div>
          {/each}
        </div>
      </div>
    </div>
  {:else}
    <!-- Top Profiles Section -->
    {#if topProfiles.length > 0}
      <div class="mb-8">
        <div class="flex justify-between items-center mb-4">
          <h3 class="font-bold text-xl text-primary dark:text-primary-dark">
            <span class="flex items-center">
              <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="mr-2">
                <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"></path>
                <circle cx="12" cy="7" r="4"></circle>
              </svg>
              Recommended People
            </span>
          </h3>
          <button class="view-all-button" on:click={() => handleViewAll('people')}>
            <span>View all</span>
            <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M5 12h14"></path>
              <path d="M12 5l7 7-7 7"></path>
            </svg>
          </button>
        </div>
        
        <div class="profiles-grid">
          {#each topProfiles as profile}
            <div class="profile-card-container {isDarkMode ? 'profile-card-container-dark' : ''}">
              <ProfileCard 
                id={profile.id}
                username={profile.username}
                displayName={profile.displayName || profile.name || profile.username}
                avatar={profile.avatar || profile.profile_picture_url}
                bio={profile.bio || ""}
                isVerified={profile.isVerified || profile.is_verified || false}
                followerCount={profile.followerCount || profile.follower_count || 0}
                isFollowing={profile.isFollowing || profile.is_following || false}
                showBio={true}
                showFollowerCount={true}
                compact={false}
              />
            </div>
          {/each}
        </div>
      </div>
    {/if}
    
    <!-- Top Threads Section -->
    {#if topThreads.length > 0}
      <div>
        <h3 class="font-bold text-lg mb-3 text-primary dark:text-primary-dark">Threads</h3>
        <div class="divide-y divide-gray-200 dark:divide-gray-800">
          {#each topThreads as thread}
            <ThreadCard {thread} />
          {/each}
        </div>
      </div>
    {:else if topProfiles.length === 0}
      <div class="text-center py-10">
        <p class="text-gray-500 dark:text-gray-400">No results found</p>
      </div>
    {/if}
  {/if}
</div>

<style>
  /* Skeleton loading animation */
  @keyframes pulse {
    0%, 100% { opacity: 0.5; }
    50% { opacity: 1; }
  }
  .animate-pulse {
    animation: pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite;
  }
  
  .profiles-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
    gap: var(--space-4);
    width: 100%;
    margin-bottom: 24px;
  }
  
  .skeleton-card {
    background-color: var(--bg-secondary, #f8f9fa);
    border-radius: var(--radius-md);
    padding: var(--space-2);
    border: 1px solid var(--border-color, #e5e7eb);
  }
  
  .profile-card-container {
    background-color: var(--bg-secondary, #f8f9fa);
    border-radius: var(--radius-md);
    border: 1px solid var(--border-color, #e5e7eb);
    transition: all 0.3s ease;
    overflow: hidden;
    padding: var(--space-2);
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
    display: flex;
    flex-direction: column;
  }
  
  .profile-card-container:hover {
    transform: translateY(-2px);
    box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);
    border-color: var(--color-primary, #3b82f6);
  }
  
  .profile-card-container-dark {
    background-color: var(--dark-bg-secondary, #1f2937);
    border-color: var(--border-color-dark, #374151);
  }
  
  .profile-card-container-dark:hover {
    border-color: var(--color-primary, #3b82f6);
    box-shadow: 0 4px 12px rgba(59, 130, 246, 0.2);
  }
  
  .text-primary {
    color: var(--text-primary);
  }
  
  .text-primary-dark {
    color: var(--dark-text-primary);
  }
  
  @media (max-width: 768px) {
    .profiles-grid {
      grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
    }
  }
  
  @media (max-width: 576px) {
    .profiles-grid {
      grid-template-columns: repeat(2, 1fr);
      gap: var(--space-2);
    }
  }
  
  .view-all-button {
    display: flex;
    align-items: center;
    gap: 4px;
    background-color: var(--color-primary, #3b82f6);
    color: white;
    font-size: var(--font-size-sm);
    font-weight: var(--font-weight-medium);
    padding: 6px 12px;
    border-radius: var(--radius-md);
    transition: all 0.2s ease;
    border: none;
    cursor: pointer;
  }
  
  .view-all-button:hover {
    background-color: var(--color-primary-dark, #2563eb);
    transform: translateY(-1px);
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  }
  
  .view-all-button svg {
    transition: transform 0.2s ease;
  }
  
  .view-all-button:hover svg {
    transform: translateX(2px);
  }
</style> 