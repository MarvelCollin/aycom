<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import type { ExtendedTweet } from "../../interfaces/ITweet.extended";
  import Linkify from "../common/Linkify.svelte";
  import { formatStorageUrl } from "../../utils/common";
  import { formatRelativeTime } from "../../utils/date";
  import RefreshCwIcon from "svelte-feather-icons/src/icons/RefreshCwIcon.svelte";

  export let originalTweet: ExtendedTweet;
  export let isDarkMode: boolean = false;

  const dispatch = createEventDispatcher();

  function handleOriginalTweetClick() {
    dispatch("clickOriginal", originalTweet);
  }

  function formatProfilePicture(url: string | undefined): string {
    if (!url) return "/images/default-avatar.png";
    return formatStorageUrl(url);
  }
</script>

<div class="reposted-content-container {isDarkMode ? "dark" : ""}" on:click|stopPropagation={handleOriginalTweetClick}>
  <div class="reposted-header">
    <div class="repost-icon">
      <RefreshCwIcon size="14" />
    </div>
    <span class="reposted-text">Original post</span>
  </div>

  <div class="original-tweet">
    <div class="tweet-author">
      <img
        src={formatProfilePicture(originalTweet.profile_picture_url)}
        alt={originalTweet.name || originalTweet.username || "User"}
        class="author-avatar"
      />
      <div class="author-info">
        <div class="author-name">
          {originalTweet.name || originalTweet.username || "User"}
          {#if originalTweet.is_verified || originalTweet?.user?.is_verified || originalTweet?.author?.is_verified}
            <span class="verified-badge">✓</span>
          {/if}
        </div>
        <div class="author-username">@{originalTweet.username || "anonymous"}</div>
      </div>
    </div>

    <div class="tweet-content">
      <Linkify text={originalTweet.content || ""} />
    </div>

    {#if originalTweet.media && originalTweet.media.length > 0}
      <div class="tweet-media">
        {#each originalTweet.media.slice(0, 2) as media}
          {#if media.type === "image"}
            <img src={formatStorageUrl(media.url)} alt="Tweet media" class="media-item" />
          {:else if media.type === "video"}
            <div class="video-container">
              <video src={formatStorageUrl(media.url)} controls class="media-item">
                Your browser does not support video playback.
              </video>
            </div>
          {/if}
        {/each}

        {#if originalTweet.media.length > 2}
          <div class="more-media">+{originalTweet.media.length - 2} more</div>
        {/if}
      </div>
    {/if}

    <div class="tweet-timestamp">
      {formatRelativeTime(originalTweet.created_at)}
    </div>
  </div>
</div>

<style>
  .reposted-content-container {
    border: 1px solid #e1e8ed;
    border-radius: 12px;
    padding: 12px;
    margin: 10px 0;
    background-color: #f8f9fa;
    cursor: pointer;
    transition: background-color 0.2s ease;
  }

  .reposted-content-container.dark {
    border: 1px solid #2f3336;
    background-color: #15181c;
    color: #e7e9ea;
  }

  .reposted-content-container:hover {
    background-color: #f0f2f5;
  }

  .reposted-content-container.dark:hover {
    background-color: #1d2023;
  }

  .reposted-header {
    display: flex;
    align-items: center;
    margin-bottom: 8px;
    color: #536471;
    font-size: 13px;
  }

  .dark .reposted-header {
    color: #71767b;
  }

  .repost-icon {
    margin-right: 4px;
    color: #536471;
  }

  .dark .repost-icon {
    color: #71767b;
  }

  .original-tweet {
    margin-left: 4px;
  }

  .tweet-author {
    display: flex;
    align-items: center;
    margin-bottom: 8px;
  }

  .author-avatar {
    width: 24px;
    height: 24px;
    border-radius: 50%;
    margin-right: 8px;
    object-fit: cover;
  }

  .author-info {
    display: flex;
    flex-direction: column;
  }

  .author-name {
    font-weight: bold;
    font-size: 14px;
    color: #0f1419;
    display: flex;
    align-items: center;
  }

  .dark .author-name {
    color: #e7e9ea;
  }

  .verified-badge {
    color: #1d9bf0;
    font-size: 12px;
    margin-left: 3px;
  }

  .author-username {
    font-size: 13px;
    color: #536471;
  }

  .dark .author-username {
    color: #71767b;
  }

  .tweet-content {
    font-size: 14px;
    line-height: 1.4;
    margin-bottom: 10px;
    word-break: break-word;
    color: #0f1419;
  }

  .dark .tweet-content {
    color: #e7e9ea;
  }

  .tweet-media {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(120px, 1fr));
    grid-gap: 4px;
    margin-bottom: 10px;
    border-radius: 8px;
    overflow: hidden;
  }

  .media-item {
    width: 100%;
    max-height: 150px;
    object-fit: cover;
    border-radius: 4px;
  }

  .video-container {
    position: relative;
    width: 100%;
    padding-bottom: 56.25%; /* 16:9 Aspect Ratio */
  }

  .video-container video {
    position: absolute;
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  .more-media {
    background-color: rgba(0, 0, 0, 0.5);
    color: white;
    display: flex;
    justify-content: center;
    align-items: center;
    font-size: 14px;
    height: 150px;
  }

  .tweet-timestamp {
    font-size: 12px;
    color: #536471;
  }

  .dark .tweet-timestamp {
    color: #71767b;
  }
</style>