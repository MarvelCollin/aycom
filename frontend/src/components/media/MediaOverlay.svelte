<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import ChevronLeftIcon from "svelte-feather-icons/src/icons/ChevronLeftIcon.svelte";
  import ChevronRightIcon from "svelte-feather-icons/src/icons/ChevronRightIcon.svelte";
  import XIcon from "svelte-feather-icons/src/icons/XIcon.svelte";
  import HeartIcon from "svelte-feather-icons/src/icons/HeartIcon.svelte";
  import MessageCircleIcon from "svelte-feather-icons/src/icons/MessageCircleIcon.svelte";
  import RepeatIcon from "svelte-feather-icons/src/icons/RepeatIcon.svelte";
  import BookmarkIcon from "svelte-feather-icons/src/icons/BookmarkIcon.svelte";
  import ShareIcon from "svelte-feather-icons/src/icons/ShareIcon.svelte";
  import DownloadIcon from "svelte-feather-icons/src/icons/DownloadIcon.svelte";

  const dispatch = createEventDispatcher();
  export let isOpen = false;
  export let mediaItems: Array<{
    id: string;
    type: 'image' | 'video';
    url: string;
    thumbnail?: string;
    alt?: string;
  }> = [];
  export let currentIndex = 0;
  export let threadData: any = null;
  export let isDarkMode = false;

  $: currentMedia = mediaItems[currentIndex];
  $: hasMultipleItems = mediaItems.length > 1;

  function closeOverlay() {
    dispatch('close');
  }

  function previousMedia() {
    if (currentIndex > 0) {
      currentIndex = currentIndex - 1;
    } else {
      currentIndex = mediaItems.length - 1; 
    }
    dispatch('navigate', { index: currentIndex });
  }

  function nextMedia() {
    if (currentIndex < mediaItems.length - 1) {
      currentIndex = currentIndex + 1;
    } else {
      currentIndex = 0; 
    }
    dispatch('navigate', { index: currentIndex });
  }

  function handleKeydown(event: KeyboardEvent) {
    if (!isOpen) return;

    switch (event.key) {
      case 'Escape':
        closeOverlay();
        break;
      case 'ArrowLeft':
        if (hasMultipleItems) previousMedia();
        break;
      case 'ArrowRight':
        if (hasMultipleItems) nextMedia();
        break;
    }
  }

  function handleLike() {
    console.log('Mock: Like interaction in media overlay');
    dispatch('like', { threadId: threadData?.id });
  }

  function handleReply() {
    console.log('Mock: Reply interaction in media overlay');
    dispatch('reply', { threadId: threadData?.id });
  }

  function handleRepost() {
    console.log('Mock: Repost interaction in media overlay');
    dispatch('repost', { threadId: threadData?.id });
  }

  function handleBookmark() {
    console.log('Mock: Bookmark interaction in media overlay');
    dispatch('bookmark', { threadId: threadData?.id });
  }

  function handleShare() {
    console.log('Mock: Share interaction in media overlay');
    dispatch('share', { threadId: threadData?.id, mediaUrl: currentMedia?.url });
  }

  function handleDownload() {
    console.log('Mock: Download interaction in media overlay');
    dispatch('download', { mediaUrl: currentMedia?.url, mediaId: currentMedia?.id });
  }

  function handleBackdropClick(event: MouseEvent) {
    if (event.target === event.currentTarget) {
      closeOverlay();
    }
  }
</script>

<svelte:window on:keydown={handleKeydown} />

{#if isOpen && currentMedia}  <div 
    class="media-overlay" 
    on:click={handleBackdropClick}
    on:keydown={handleKeydown}
    role="dialog"
    aria-modal="true"
    aria-label="Media overlay"
    tabindex="-1"
  >
    <!-- Header Controls -->
    <div class="overlay-header">
      <div class="header-left">
        <button class="icon-btn" on:click={closeOverlay} aria-label="Close overlay">
          <XIcon size="24" />
        </button>
        {#if hasMultipleItems}
          <span class="media-counter">
            {currentIndex + 1} / {mediaItems.length}
          </span>
        {/if}
      </div>

      <div class="header-right">
        <button class="icon-btn" on:click={handleDownload} aria-label="Download media">
          <DownloadIcon size="20" />
        </button>
        <button class="icon-btn" on:click={handleShare} aria-label="Share media">
          <ShareIcon size="20" />
        </button>
      </div>
    </div>

    <!-- Media Content -->
    <div class="media-container">
      {#if hasMultipleItems}
        <button 
          class="nav-btn nav-left" 
          on:click={previousMedia}
          aria-label="Previous media"
        >
          <ChevronLeftIcon size="32" />
        </button>
      {/if}

      <div class="media-content">
        {#if currentMedia.type === 'image'}
          <img 
            src={currentMedia.url} 
            alt={currentMedia.alt || 'Media content'}
            class="media-image"
          />
        {:else if currentMedia.type === 'video'}
          <video 
            src={currentMedia.url} 
            controls
            class="media-video"
            poster={currentMedia.thumbnail}
          >
            <track kind="captions" />
            Your browser does not support the video tag.
          </video>
        {/if}
      </div>

      {#if hasMultipleItems}
        <button 
          class="nav-btn nav-right" 
          on:click={nextMedia}
          aria-label="Next media"
        >
          <ChevronRightIcon size="32" />
        </button>
      {/if}
    </div>

    <!-- Media Thumbnails (if multiple items) -->
    {#if hasMultipleItems}
      <div class="thumbnails-container">
        {#each mediaItems as media, index}
          <button 
            class="thumbnail-btn {index === currentIndex ? 'active' : ''}"
            on:click={() => currentIndex = index}
            aria-label="View media {index + 1}"
          >
            {#if media.type === 'image'}
              <img src={media.thumbnail || media.url} alt="Thumbnail {index + 1}" />
            {:else}
              <div class="video-thumbnail">
                <img src={media.thumbnail || media.url} alt="Video thumbnail {index + 1}" />
                <div class="play-icon">▶</div>
              </div>
            {/if}
          </button>
        {/each}
      </div>
    {/if}

    <!-- Interaction Controls -->
    <div class="interaction-controls">
      <div class="controls-content">
        <button class="control-btn like-btn" on:click={handleLike}>
          <HeartIcon size="20" />
          <span>Like</span>
        </button>

        <button class="control-btn reply-btn" on:click={handleReply}>
          <MessageCircleIcon size="20" />
          <span>Reply</span>
        </button>

        <button class="control-btn repost-btn" on:click={handleRepost}>
          <RepeatIcon size="20" />
          <span>Repost</span>
        </button>

        <button class="control-btn bookmark-btn" on:click={handleBookmark}>
          <BookmarkIcon size="20" />
          <span>Bookmark</span>
        </button>
      </div>
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
    background: rgba(0, 0, 0, 0.95);
    z-index: 1000;
    display: flex;
    flex-direction: column;
    backdrop-filter: blur(4px);
  }

  .overlay-header {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    z-index: 1001;
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1rem;
    background: linear-gradient(to bottom, rgba(0, 0, 0, 0.7), transparent);
  }

  .header-left {
    display: flex;
    align-items: center;
    gap: 1rem;
  }

  .header-right {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .media-counter {
    color: white;
    font-size: 0.9rem;
    font-weight: 500;
  }

  .icon-btn {
    background: rgba(255, 255, 255, 0.1);
    border: none;
    border-radius: 50%;
    width: 40px;
    height: 40px;
    display: flex;
    align-items: center;
    justify-content: center;
    color: white;
    cursor: pointer;
    transition: background-color 0.2s ease;
  }

  .icon-btn:hover {
    background: rgba(255, 255, 255, 0.2);
  }

  .media-container {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    position: relative;
    padding: 2rem;
  }

  .media-content {
    max-width: 90vw;
    max-height: 80vh;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .media-image, .media-video {
    max-width: 100%;
    max-height: 100%;
    object-fit: contain;
    border-radius: 8px;
    box-shadow: 0 4px 20px rgba(0, 0, 0, 0.5);
  }

  .nav-btn {
    position: absolute;
    top: 50%;
    transform: translateY(-50%);
    background: rgba(255, 255, 255, 0.1);
    border: none;
    border-radius: 50%;
    width: 50px;
    height: 50px;
    display: flex;
    align-items: center;
    justify-content: center;
    color: white;
    cursor: pointer;
    transition: all 0.2s ease;
    z-index: 1002;
  }

  .nav-btn:hover {
    background: rgba(255, 255, 255, 0.2);
    transform: translateY(-50%) scale(1.1);
  }

  .nav-left {
    left: 2rem;
  }

  .nav-right {
    right: 2rem;
  }

  .thumbnails-container {
    position: absolute;
    bottom: 100px;
    left: 50%;
    transform: translateX(-50%);
    display: flex;
    gap: 0.5rem;
    padding: 1rem;
    background: rgba(0, 0, 0, 0.7);
    border-radius: 12px;
    backdrop-filter: blur(10px);
  }

  .thumbnail-btn {
    width: 60px;
    height: 60px;
    border: 2px solid transparent;
    border-radius: 8px;
    overflow: hidden;
    cursor: pointer;
    transition: all 0.2s ease;
    position: relative;
  }

  .thumbnail-btn:hover {
    border-color: rgba(255, 255, 255, 0.5);
    transform: scale(1.05);
  }

  .thumbnail-btn.active {
    border-color: #1da1f2;
    transform: scale(1.1);
  }

  .thumbnail-btn img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  .video-thumbnail {
    position: relative;
    width: 100%;
    height: 100%;
  }

  .play-icon {
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    color: white;
    font-size: 12px;
    text-shadow: 0 2px 4px rgba(0, 0, 0, 0.8);
  }

  .interaction-controls {
    position: absolute;
    bottom: 0;
    left: 0;
    right: 0;
    background: linear-gradient(to top, rgba(0, 0, 0, 0.8), transparent);
    padding: 2rem 1rem 1rem;
  }

  .controls-content {
    display: flex;
    justify-content: center;
    gap: 2rem;
    max-width: 500px;
    margin: 0 auto;
  }

  .control-btn {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.25rem;
    background: rgba(255, 255, 255, 0.1);
    border: none;
    border-radius: 12px;
    padding: 0.75rem 1rem;
    color: white;
    cursor: pointer;
    transition: all 0.2s ease;
    min-width: 70px;
  }

  .control-btn:hover {
    background: rgba(255, 255, 255, 0.2);
    transform: translateY(-2px);
  }

  .control-btn span {
    font-size: 0.75rem;
    font-weight: 500;
  }

  .like-btn:hover {
    background: rgba(231, 76, 60, 0.2);
    color: #e74c3c;
  }

  .reply-btn:hover {
    background: rgba(29, 161, 242, 0.2);
    color: #1da1f2;
  }

  .repost-btn:hover {
    background: rgba(46, 204, 113, 0.2);
    color: #2ecc71;
  }

  .bookmark-btn:hover {
    background: rgba(241, 196, 15, 0.2);
    color: #f1c40f;
  }

  @media (max-width: 768px) {
    .overlay-header {
      padding: 0.75rem;
    }

    .media-container {
      padding: 1rem;
    }

    .nav-btn {
      width: 40px;
      height: 40px;
    }

    .nav-left {
      left: 1rem;
    }

    .nav-right {
      right: 1rem;
    }

    .thumbnails-container {
      bottom: 80px;
      padding: 0.5rem;
    }

    .thumbnail-btn {
      width: 50px;
      height: 50px;
    }

    .controls-content {
      gap: 1rem;
    }

    .control-btn {
      min-width: 60px;
      padding: 0.5rem 0.75rem;
    }

    .control-btn span {
      font-size: 0.7rem;
    }
  }
</style>