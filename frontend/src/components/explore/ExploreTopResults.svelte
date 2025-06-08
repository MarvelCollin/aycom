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
      <div class="mb-6">
        <div class="flex justify-between items-center mb-3">
          <h3 class="font-bold text-lg text-primary dark:text-primary-dark">People</h3>
          <button class="text-blue-500 text-sm" on:click={() => handleViewAll('people')}>
            View all
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
    grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
    gap: var(--space-2);
    width: 100%;
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
    transition: all 0.2s ease;
    overflow: hidden;
  }
  
  .profile-card-container:hover {
    transform: translateY(-1px);
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
  }
  
  .profile-card-container-dark {
    background-color: var(--dark-bg-secondary, #1f2937);
    border-color: var(--border-color-dark, #374151);
  }
  
  .text-primary {
    color: var(--text-primary);
  }
  
  .text-primary-dark {
    color: var(--dark-text-primary);
  }
  
  @media (max-width: 768px) {
    .profiles-grid {
      grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
    }
  }
  
  @media (max-width: 576px) {
    .profiles-grid {
      grid-template-columns: repeat(2, 1fr);
    }
  }
</style> 