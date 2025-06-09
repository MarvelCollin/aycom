<script lang="ts">
  import { onMount } from 'svelte';
  import { getUserBookmarks } from '../api/thread';
  import MainLayout from '../components/layout/MainLayout.svelte';

  let bookmarks: any[] = [];
  let loading = true;
  let error: string | null = null;

  onMount(async () => {
    try {
      loading = true;
      const response = await getUserBookmarks('me');
      console.log('Bookmarks API response:', response);
      
      // Handle response structure based on the backend format
      if (response && response.success) {
        bookmarks = response.bookmarks || [];
        console.log('Successfully loaded bookmarks:', bookmarks.length);
      } else {
        error = 'Failed to load bookmarks';
      }
    } catch (err: any) {
      console.error('Error fetching bookmarks:', err);
      error = err.message || 'Failed to load bookmarks';
    } finally {
      loading = false;
    }
  });
</script>

<MainLayout>
  <div class="bookmarks-page">
    <h1>Bookmarks</h1>
    
    {#if loading}
      <div class="loading">Loading bookmarks...</div>
    {:else if error}
      <div class="error-message">
        <p>{error}</p>
      </div>
    {:else if bookmarks.length === 0}
      <div class="empty-state">
        <h3>No bookmarks yet</h3>
        <p>Threads you bookmark will appear here.</p>
      </div>
    {:else}
      <div class="bookmarks-list">
        {#each bookmarks as bookmark}
          <div class="bookmark-item">
            <div class="bookmark-header">
              <div class="user-info">
                {#if bookmark.profile_picture_url}
                  <img src={bookmark.profile_picture_url} alt={bookmark.username || 'User'} class="avatar" />
                {/if}
                <div class="user-details">
                  <span class="display-name">{bookmark.name || 'Unknown'}</span>
                  <span class="username">@{bookmark.username || 'user'}</span>
                </div>
              </div>
              <div class="bookmark-date">
                {new Date(bookmark.created_at).toLocaleDateString()}
              </div>
            </div>
            
            <div class="thread-content">
              <p>{bookmark.content}</p>
            </div>
            
            {#if bookmark.media && bookmark.media.length > 0}
              <div class="media-container">
                {#each bookmark.media as media}
                  {#if media.type === 'image'}
                    <img src={media.url} alt="Thread media" class="thread-media" />
                  {:else if media.type === 'video'}
                    <video src={media.url} controls class="thread-media">
                      Your browser does not support video playback.
                    </video>
                  {/if}
                {/each}
              </div>
            {/if}
            
            <div class="thread-stats">
              <div class="stat-item">
                <span class="icon">❤️</span> {bookmark.likes_count || 0}
              </div>
              <div class="stat-item">
                <span class="icon">💬</span> {bookmark.replies_count || 0}
              </div>
              <div class="stat-item">
                <span class="icon">🔁</span> {bookmark.reposts_count || 0}
              </div>
            </div>
          </div>
        {/each}
      </div>
    {/if}
  </div>
</MainLayout>

<style>
  .bookmarks-page {
    max-width: 800px;
    margin: 0 auto;
    padding: 20px;
  }

  h1 {
    margin-bottom: 20px;
  }

  .loading {
    text-align: center;
    padding: 40px;
    color: var(--text-secondary);
  }

  .error-message {
    color: var(--color-error, #ff0000);
    padding: 20px;
    background-color: var(--error-bg, #fff1f0);
    border-radius: 8px;
    margin: 20px 0;
  }

  .empty-state {
    text-align: center;
    padding: 40px 20px;
    color: var(--text-secondary);
  }

  .bookmarks-list {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .bookmark-item {
    padding: 16px;
    border: 1px solid var(--border-color, #eaeaea);
    border-radius: 8px;
    background-color: var(--bg-secondary, white);
  }

  .bookmark-header {
    display: flex;
    justify-content: space-between;
    margin-bottom: 12px;
  }

  .user-info {
    display: flex;
    align-items: center;
  }

  .avatar {
    width: 40px;
    height: 40px;
    border-radius: 50%;
    margin-right: 10px;
  }

  .user-details {
    display: flex;
    flex-direction: column;
  }

  .display-name {
    font-weight: bold;
  }

  .username {
    color: var(--text-secondary);
    font-size: 0.9em;
  }

  .bookmark-date {
    color: var(--text-secondary);
    font-size: 0.9em;
  }

  .thread-content {
    margin-bottom: 16px;
  }

  .media-container {
    margin-bottom: 16px;
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }

  .thread-media {
    max-width: 100%;
    border-radius: 8px;
    max-height: 400px;
    object-fit: contain;
  }

  .thread-stats {
    display: flex;
    gap: 24px;
    color: var(--text-secondary);
  }

  .stat-item {
    display: flex;
    align-items: center;
    gap: 4px;
  }

  .icon {
    font-size: 1.2em;
  }
</style>
