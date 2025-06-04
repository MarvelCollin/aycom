<script lang="ts">
  import MainLayout from '../components/layout/MainLayout.svelte';
  import TweetCard from '../components/social/TweetCard.svelte';
  import { onMount } from 'svelte';
  import { getAllThreads } from '../api/thread';
  import { useTheme } from '../hooks/useTheme';
  import { formatStorageUrl } from '../utils/common';
  import type { ITweet } from '../interfaces/ISocialMedia';

  interface Thread {
    id: string;
    content: string;
    created_at: string;
    updated_at?: string;
    username: string;
    name?: string;
    user_id: string;
    profile_picture_url: string;
    likes_count: number;
    replies_count: number;
    reposts_count: number;
    is_liked: boolean;
    is_reposted: boolean;
    is_bookmarked: boolean;
    is_pinned: boolean;
    media?: Array<{
      id: string;
      url: string;
      type: string;
    }>;
  }

  // State variables
  let threads: Thread[] = [];
  let isLoading = true;
  let error: string | null = null;
  let page = 1;
  let limit = 20;
  let totalCount = 0;
  
  // Get theme store
  const { theme } = useTheme();
  
  // Reactive declaration for dark mode
  $: isDarkMode = $theme === 'dark';

  // Load threads function
  async function loadThreads() {
    console.log('Loading threads...');
    isLoading = true;
    error = null;

    try {
      const response = await getAllThreads(page, limit);
      console.log('Thread API response:', response);
      
      if (response && response.success && Array.isArray(response.threads)) {
        threads = response.threads;
        totalCount = response.total_count || 0;
        console.log('Loaded threads:', threads.length, 'of total:', totalCount);
      } else if (response && Array.isArray(response)) {
        // Handle case where API returns threads directly as array
        threads = response;
        totalCount = response.length;
        console.log('Loaded threads directly:', threads.length);
      } else {
        console.error('Invalid API response format:', response);
        threads = [];
        error = 'No threads available right now. Try again later.';
      }
    } catch (err) {
      console.error('Error loading threads:', err);
      if (err instanceof Error && err.message.includes('401')) {
        // If it's an auth error, don't show it to the user, just show empty state
        threads = [];
        error = 'No threads available right now. Try again later.';
      } else {
        // For other errors, show a helpful message
        error = 'Unable to load threads. Please check your connection and try again.';
      }
    } finally {
      isLoading = false;
    }
  }

  // Convert Thread to ITweet for compatibility with TweetCard
  function threadToTweet(thread: Thread): ITweet {
    // Map media items to ensure type is one of the allowed values
    // Also format media URLs through formatStorageUrl
    const mappedMedia = (thread.media || []).map(item => ({
      url: formatStorageUrl(item.url),
      type: mapMediaType(item.type),
      alt_text: ''  // Add required alt_text property
    }));
    
    // Process the profile picture URL through formatStorageUrl
    const formattedProfilePicture = formatStorageUrl(thread.profile_picture_url);
    
    return {
      ...thread,
      name: thread.name || thread.username,
      profile_picture_url: formattedProfilePicture,
      bookmark_count: 0,
      parent_id: null,
      media: mappedMedia,
      views_count: 0
    };
  }
  
  // Helper function to map media types to allowed values
  function mapMediaType(type: string): 'image' | 'video' | 'gif' {
    if (type === 'video') return 'video';
    if (type === 'gif') return 'gif';
    return 'image'; // Default to image for any other type
  }

  // Authentication status - we'll assume the user is authenticated for now
  // This should be replaced with actual auth state in production
  const isAuthenticated = true;

  // Load on mount
  onMount(() => {
    loadThreads();
  });
</script>

<MainLayout>
  <div class="feed-container {isDarkMode ? 'feed-container-dark' : ''}">
    <h1 class="feed-title {isDarkMode ? 'feed-title-dark' : ''}">Feed</h1>
    
    {#if isLoading}
      <div class="loading {isDarkMode ? 'loading-dark' : ''}">
        <div class="loading-spinner"></div>
        <span>Loading threads...</span>
      </div>
    {:else if error}
      <div class="error {isDarkMode ? 'error-dark' : ''}">
        <p>{error}</p>
        <button class="retry-button {isDarkMode ? 'retry-button-dark' : ''}" on:click={loadThreads}>Retry</button>
      </div>
    {:else if threads.length === 0}
      <div class="empty {isDarkMode ? 'empty-dark' : ''}">No threads found</div>
    {:else}
      <div class="threads-list">
        {#each threads as thread (thread.id)}
          <TweetCard 
            tweet={threadToTweet(thread)} 
            {isDarkMode} 
            isAuth={isAuthenticated}
          />
        {/each}
      </div>
    {/if}
  </div>
</MainLayout>

<style>
  .feed-container {
    max-width: 800px;
    margin: 0 auto;
    padding: 20px;
    background-color: var(--bg-primary);
    color: var(--text-primary);
  }
  
  .feed-container-dark {
    background-color: var(--bg-primary-dark);
    color: var(--text-primary-dark);
  }

  .feed-title {
    margin-bottom: 1rem;
    font-size: 1.5rem;
    font-weight: 700;
    color: var(--text-primary);
  }
  
  .feed-title-dark {
    color: var(--text-primary-dark);
  }

  .loading, .error, .empty {
    text-align: center;
    padding: 40px;
    font-size: 1.1rem;
    background-color: var(--bg-secondary);
    border-radius: 10px;
    margin: 20px 0;
  }
  
  .loading-dark, .error-dark, .empty-dark {
    background-color: var(--bg-secondary-dark);
    color: var(--text-primary-dark);
  }
  
  .loading {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 1rem;
  }
  
  .loading-spinner {
    width: 40px;
    height: 40px;
    border: 4px solid rgba(0, 0, 0, 0.1);
    border-left-color: var(--color-primary);
    border-radius: 50%;
    animation: spin 1s linear infinite;
  }
  
  .loading-dark .loading-spinner {
    border-color: rgba(255, 255, 255, 0.1);
    border-left-color: var(--color-primary);
  }
  
  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }

  .error button {
    margin-top: 10px;
    padding: 10px 20px;
    background: var(--color-primary);
    color: white;
    border: none;
    border-radius: 5px;
    cursor: pointer;
    font-weight: 600;
    transition: all 0.2s ease;
  }
  
  .error button:hover {
    background: var(--color-primary-hover);
    transform: translateY(-2px);
  }

  .threads-list {
    display: flex;
    flex-direction: column;
    gap: 1px;
    border-radius: 10px;
    overflow: hidden;
    border: 1px solid var(--border-color);
  }
  
  :global(.dark-theme) .threads-list {
    border-color: var(--border-color-dark);
  }
</style>
