<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { fade, scale } from 'svelte/transition';
  import ChevronLeftIcon from 'svelte-feather-icons/src/icons/ChevronLeftIcon.svelte';
  import ChevronRightIcon from 'svelte-feather-icons/src/icons/ChevronRightIcon.svelte';
  import XIcon from 'svelte-feather-icons/src/icons/XIcon.svelte';
  import HeartIcon from 'svelte-feather-icons/src/icons/HeartIcon.svelte';
  import MessageCircleIcon from 'svelte-feather-icons/src/icons/MessageCircleIcon.svelte';
  import BookmarkIcon from 'svelte-feather-icons/src/icons/BookmarkIcon.svelte';
  import ShareIcon from 'svelte-feather-icons/src/icons/ShareIcon.svelte';
  import DownloadIcon from 'svelte-feather-icons/src/icons/DownloadIcon.svelte';

  interface MediaItem {
    id: string;
    url: string;
    type: 'image' | 'video';
    alt_text?: string;
    thumbnail_url?: string;
  }

  interface TweetData {
    id: string;
    username: string;
    name: string;
    profile_picture_url: string;
    content: string;
    likes_count: number;
    replies_count: number;
    reposts_count: number;
    bookmark_count: number;
    is_liked: boolean;
    is_bookmarked: boolean;
    is_reposted: boolean;
  }

  export let mediaItems: MediaItem[] = [];
  export let currentIndex: number = 0;
  export let tweetData: TweetData | null = null;
  export let isDarkMode: boolean = false;
  export let isVisible: boolean = false;

  const dispatch = createEventDispatcher();

  $: currentMedia = mediaItems[currentIndex] || null;
  $: hasPrevious = currentIndex > 0;
  $: hasNext = currentIndex < mediaItems.length - 1;

  function goToPrevious() {
    if (hasPrevious) {
      currentIndex = currentIndex - 1;
    }
  }

  function goToNext() {
    if (hasNext) {
      currentIndex = currentIndex + 1;
    }
  }

  function closeOverlay() {
    dispatch('close');
  }

  function handleKeydown(event: KeyboardEvent) {
    switch (event.key) {
      case 'Escape':
        closeOverlay();
        break;
      case 'ArrowLeft':
        goToPrevious();
        break;
      case 'ArrowRight':
        goToNext();
        break;
    }
  }

  function handleLike() {
    dispatch('like', { tweetId: tweetData?.id });
  }

  function handleReply() {
    dispatch('reply', { tweetId: tweetData?.id });
  }

  function handleBookmark() {
    dispatch('bookmark', { tweetId: tweetData?.id });
  }

  function handleShare() {
    dispatch('share', { tweetId: tweetData?.id, mediaUrl: currentMedia?.url });
  }

  function handleDownload() {
    if (currentMedia?.url) {
      const link = document.createElement('a');
      link.href = currentMedia.url;
      link.download = `media-${currentMedia.id}`;
      link.click();
    }
  }

  function formatCount(count: number): string {
    if (count >= 1000000) {
      return (count / 1000000).toFixed(1) + 'M';
    } else if (count >= 1000) {
      return (count / 1000).toFixed(1) + 'K';
    }
    return count.toString();
  }
</script>

<!-- svelte-ignore a11y-no-noninteractive-tabindex -->
{#if isVisible && currentMedia}
  <div 
    class="media-overlay {isDarkMode ? 'dark' : 'light'}"
    transition:fade={{ duration: 200 }}
    on:keydown={handleKeydown}
    tabindex="0"
    role="dialog"
    aria-modal="true"
    aria-label="Media viewer"
  >
    <!-- Background -->
    <div class="overlay-background" on:click={closeOverlay}></div>

    <!-- Content -->
    <div class="overlay-content" on:click|stopPropagation>
      <!-- Header -->
      <div class="overlay-header">
        <div class="header-left">
          <button class="nav-button close-button" on:click={closeOverlay} title="Close">
            <XIcon size="24" />
          </button>
          <div class="media-counter">
            {currentIndex + 1} of {mediaItems.length}
          </div>
        </div>

        <div class="header-right">
          <button class="action-button" on:click={handleDownload} title="Download">
            <DownloadIcon size="20" />
          </button>
          <button class="action-button" on:click={handleShare} title="Share">
            <ShareIcon size="20" />
          </button>
        </div>
      </div>

      <!-- Main Media Display -->
      <div class="media-container">
        <!-- Navigation buttons -->
        {#if hasPrevious}
          <button 
            class="nav-button prev-button" 
            on:click={goToPrevious}
            title="Previous image"
            transition:scale={{ duration: 150 }}
          >
            <ChevronLeftIcon size="32" />
          </button>
        {/if}

        {#if hasNext}
          <button 
            class="nav-button next-button" 
            on:click={goToNext}
            title="Next image"
            transition:scale={{ duration: 150 }}
          >
            <ChevronRightIcon size="32" />
          </button>
        {/if}

        <!-- Media content -->
        <div class="media-content">
          {#if currentMedia.type === 'image'}
            <img 
              src={currentMedia.url}
              alt={currentMedia.alt_text || `Image ${currentIndex + 1}`}
              class="media-image"
              loading="lazy"
            />
          {:else if currentMedia.type === 'video'}
            <video 
              src={currentMedia.url}
              class="media-video"
              controls
              autoplay
              muted
            >
              Your browser does not support the video tag.
            </video>
          {/if}
        </div>
      </div>

      <!-- Tweet Information and Interactions -->
      {#if tweetData}
        <div class="tweet-info">
          <div class="tweet-header">
            <img 
              src={tweetData.profile_picture_url || "https://secure.gravatar.com/avatar/0?d=mp"}
              alt={tweetData.name}
              class="profile-pic"
            />
            <div class="user-info">
              <div class="display-name">{tweetData.name}</div>
              <div class="username">@{tweetData.username}</div>
            </div>
          </div>

          {#if tweetData.content}
            <div class="tweet-content">
              {tweetData.content}
            </div>
          {/if}

          <!-- Interaction buttons -->
          <div class="interaction-bar">
            <button 
              class="interaction-button {tweetData.is_liked ? 'liked' : ''}"
              on:click={handleLike}
              title="Like"
            >
              <HeartIcon size="20" />
              <span class="count">{formatCount(tweetData.likes_count)}</span>
            </button>

            <button 
              class="interaction-button"
              on:click={handleReply}
              title="Reply"
            >
              <MessageCircleIcon size="20" />
              <span class="count">{formatCount(tweetData.replies_count)}</span>
            </button>

            <button 
              class="interaction-button {tweetData.is_bookmarked ? 'bookmarked' : ''}"
              on:click={handleBookmark}
              title="Bookmark"
            >
              <BookmarkIcon size="20" />
              <span class="count">{formatCount(tweetData.bookmark_count)}</span>
            </button>

            <button 
              class="interaction-button"
              on:click={handleShare}
              title="Share"
            >
              <ShareIcon size="20" />
            </button>
          </div>
        </div>
      {/if}

      <!-- Media thumbnails navigation -->
      {#if mediaItems.length > 1}
        <div class="thumbnail-nav">
          {#each mediaItems as media, index}
            <button 
              class="thumbnail-button {index === currentIndex ? 'active' : ''}"
              on:click={() => currentIndex = index}
            >
              <img 
                src={media.thumbnail_url || media.url}
                alt={`Media ${index + 1}`}
                class="thumbnail-image"
              />
            </button>
          {/each}
        </div>
      {/if}
    </div>
  </div>
{/if}

<style>
  .media-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    z-index: 9999;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .overlay-background {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.9);
  }

  .overlay-content {
    position: relative;
    width: 100%;
    height: 100%;
    max-width: 90vw;
    max-height: 90vh;
    display: flex;
    flex-direction: column;
    background: var(--overlay-bg);
    border-radius: 8px;
    overflow: hidden;
  }

  .overlay-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1rem;
    background: var(--header-bg);
    border-bottom: 1px solid var(--border-color);
  }

  .header-left {
    display: flex;
    align-items: center;
    gap: 1rem;
  }

  .header-right {
    display: flex;
    gap: 0.5rem;
  }

  .media-counter {
    color: var(--text-secondary);
    font-size: 0.9rem;
  }

  .media-container {
    position: relative;
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    background: var(--media-bg);
    overflow: hidden;
  }

  .media-content {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 100%;
    height: 100%;
  }

  .media-image {
    max-width: 100%;
    max-height: 100%;
    object-fit: contain;
  }

  .media-video {
    max-width: 100%;
    max-height: 100%;
  }

  .nav-button {
    position: absolute;
    background: rgba(0, 0, 0, 0.5);
    color: white;
    border: none;
    border-radius: 50%;
    width: 48px;
    height: 48px;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    transition: all 0.2s ease;
    z-index: 10;
  }

  .nav-button:hover {
    background: rgba(0, 0, 0, 0.7);
    transform: scale(1.1);
  }

  .close-button {
    position: static;
    background: transparent;
    color: var(--text-primary);
  }

  .prev-button {
    left: 1rem;
    top: 50%;
    transform: translateY(-50%);
  }

  .next-button {
    right: 1rem;
    top: 50%;
    transform: translateY(-50%);
  }

  .action-button {
    background: transparent;
    color: var(--text-primary);
    border: none;
    border-radius: 50%;
    width: 40px;
    height: 40px;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    transition: background-color 0.2s ease;
  }

  .action-button:hover {
    background: var(--hover-bg);
  }

  .tweet-info {
    padding: 1rem;
    background: var(--tweet-bg);
    border-top: 1px solid var(--border-color);
  }

  .tweet-header {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    margin-bottom: 0.5rem;
  }

  .profile-pic {
    width: 40px;
    height: 40px;
    border-radius: 50%;
    object-fit: cover;
  }

  .user-info {
    flex: 1;
  }

  .display-name {
    font-weight: 600;
    color: var(--text-primary);
  }

  .username {
    color: var(--text-secondary);
    font-size: 0.9rem;
  }

  .tweet-content {
    color: var(--text-primary);
    margin-bottom: 1rem;
    line-height: 1.5;
  }

  .interaction-bar {
    display: flex;
    gap: 2rem;
    align-items: center;
  }

  .interaction-button {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    background: transparent;
    border: none;
    color: var(--text-secondary);
    cursor: pointer;
    transition: color 0.2s ease;
    padding: 0.5rem;
    border-radius: 8px;
  }

  .interaction-button:hover {
    background: var(--hover-bg);
  }

  .interaction-button.liked {
    color: #e91e63;
  }

  .interaction-button.bookmarked {
    color: #1976d2;
  }

  .count {
    font-size: 0.9rem;
    min-width: 20px;
  }

  .thumbnail-nav {
    display: flex;
    gap: 0.5rem;
    padding: 1rem;
    background: var(--header-bg);
    overflow-x: auto;
    border-top: 1px solid var(--border-color);
  }

  .thumbnail-button {
    background: transparent;
    border: 2px solid transparent;
    border-radius: 4px;
    cursor: pointer;
    transition: border-color 0.2s ease;
    padding: 2px;
  }

  .thumbnail-button.active {
    border-color: var(--primary-color, #1976d2);
  }

  .thumbnail-image {
    width: 60px;
    height: 60px;
    object-fit: cover;
    border-radius: 4px;
  }

  .light {
    --overlay-bg: #ffffff;
    --header-bg: #f8f9fa;
    --media-bg: #000000;
    --tweet-bg: #ffffff;
    --text-primary: #1a1a1a;
    --text-secondary: #6b7280;
    --border-color: #e5e7eb;
    --hover-bg: #f3f4f6;
  }

  .dark {
    --overlay-bg: #1f2937;
    --header-bg: #111827;
    --media-bg: #000000;
    --tweet-bg: #1f2937;
    --text-primary: #f9fafb;
    --text-secondary: #9ca3af;
    --border-color: #374151;
    --hover-bg: #374151;
  }

  @media (max-width: 768px) {
    .overlay-content {
      max-width: 100vw;
      max-height: 100vh;
      border-radius: 0;
    }

    .interaction-bar {
      gap: 1rem;
    }

    .nav-button {
      width: 40px;
      height: 40px;
    }

    .prev-button {
      left: 0.5rem;
    }

    .next-button {
      right: 0.5rem;
    }
  }
</style>