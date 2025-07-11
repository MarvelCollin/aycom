<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import { useTheme } from "../../hooks/useTheme";

  const dispatch = createEventDispatcher();
  const { theme } = useTheme();

  $: isDarkMode = $theme === "dark";

  export let id: string;
  export let username: string;
  export let displayName: string;
  export let avatar: string | null = null;
  export let bio: string = "";
  export let isVerified: boolean = false;
  export let followerCount: number = 0;
  export let isFollowing: boolean = false;
  export let showBio: boolean = false; 
  export let showFollowerCount: boolean = true;
  export let compact: boolean = true; 
  export let onToggleFollow: () => void = () => {};
  export let fuzzyMatchScore: number | undefined = undefined; 

  $: console.log("ProfileCard props:", { id, username, displayName, isVerified, fuzzyMatchScore });

  function handleCardClick() {

    window.location.href = `/user/${username}`;
    dispatch("profileClick", id);
  }

  $: avatarUrl = avatar || `https:

  function getFuzzyMatchColor(score: number): string {
    if (score >= 0.8) return "fuzzy-match-high";
    if (score >= 0.6) return "fuzzy-match-medium";
    if (score >= 0.3) return "fuzzy-match-low";
    return "";
  }

  function getFuzzyMatchLabel(score: number): string {
    if (score >= 0.8) return "Strong match";
    if (score >= 0.6) return "Good match";
    if (score >= 0.3) return "Possible match";
    return "";
  }
</script>

<div
  class="profile-card {isDarkMode ? "profile-card-dark" : ""} {compact ? "compact" : ""}"
  on:click={handleCardClick}
  on:keydown={(e) => e.key === "Enter" && handleCardClick()}
  tabindex="0"
  role="button"
>
  <div class="profile-content">
    <div class="avatar-container">
      {#if avatarUrl && typeof avatarUrl === "string" && avatarUrl.startsWith("http")}
        <div class="image-wrapper">
          <img src={avatarUrl} alt={username} class="avatar-image" loading="lazy" />
        </div>
      {:else}
        <div class="avatar-fallback">
          <span class="avatar-initial">{displayName.charAt(0).toUpperCase()}</span>
        </div>
      {/if}
    </div>

    <div class="profile-info">
      <div class="name-container">
        <p class="display-name">{displayName}</p>
        {#if isVerified}
          <span class="verified-badge">
            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor">
              <path fill-rule="evenodd" d="M6.267 3.455a3.066 3.066 0 001.745-.723 3.066 3.066 0 013.976 0 3.066 3.066 0 001.745.723 3.066 3.066 0 012.812 2.812c.051.643.304 1.254.723 1.745a3.066 3.066 0 010 3.976 3.066 3.066 0 00-.723 1.745 3.066 3.066 0 01-2.812 2.812 3.066 3.066 0 00-1.745.723 3.066 3.066 0 01-3.976 0 3.066 3.066 0 00-1.745-.723 3.066 3.066 0 01-2.812-2.812 3.066 3.066 0 00-.723-1.745 3.066 3.066 0 010-3.976 3.066 3.066 0 00.723-1.745 3.066 3.066 0 012.812-2.812zm7.44 5.252a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
            </svg>
          </span>
        {/if}

        <!-- Fuzzy match indicator -->
        {#if fuzzyMatchScore !== undefined && fuzzyMatchScore > 0}
          <span class="fuzzy-match-badge {getFuzzyMatchColor(fuzzyMatchScore)}" title="{getFuzzyMatchLabel(fuzzyMatchScore)} ({Math.round(fuzzyMatchScore * 100)}% similarity)">
            <svg xmlns="http://www.w3.org/2000/svg" class="fuzzy-match-icon" viewBox="0 0 20 20" fill="currentColor">
              <path fill-rule="evenodd" d="M8 4a4 4 0 100 8 4 4 0 000-8zM2 8a6 6 0 1110.89 3.476l4.817 4.817a1 1 0 01-1.414 1.414l-4.816-4.816A6 6 0 012 8z" clip-rule="evenodd" />
            </svg>
            {Math.round(fuzzyMatchScore * 100)}%
          </span>
        {/if}
      </div>

      <p class="username">@{username}</p>

      {#if showBio && bio}
        <p class="bio-text">{bio}</p>
      {/if}

      {#if showFollowerCount}
        <p class="follower-count">
          <span class="follower-number">{followerCount}</span> {followerCount === 1 ? "follower" : "followers"}
        </p>
      {/if}
    </div>
  </div>
</div>

<style>
  .profile-card {
    display: flex;
    align-items: center;
    justify-content: flex-start;
    padding: var(--space-3) var(--space-4);
    border-radius: var(--radius-md);
    cursor: pointer;
    transition: all 0.3s ease;
    gap: var(--space-3);
    background-color: var(--bg-secondary);
    border: 1px solid var(--border-color);
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
  }

  .profile-card:hover {
    background-color: var(--bg-hover);
    transform: translateY(-2px);
    box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);
    border-color: var(--color-primary);
  }

  .profile-card-dark {
    background-color: var(--dark-bg-secondary);
    border-color: var(--dark-border-color);
  }

  .profile-card-dark:hover {
    background-color: var(--dark-hover-bg);
    box-shadow: 0 4px 8px rgba(0, 0, 0, 0.2);
    border-color: var(--color-primary);
  }

  .profile-card.compact {
    padding: var(--space-3);
  }

  .profile-content {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    flex: 1;
    min-width: 0;
  }

  .avatar-container {
    flex-shrink: 0;
    border: 2px solid var(--border-color);
    width: 48px !important;
    height: 48px !important;
    border-radius: 50%;
    position: relative;
    display: flex;
    align-items: center;
    justify-content: center;
    overflow: hidden;
    transition: all 0.2s ease;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  }

  .profile-card:hover .avatar-container {
    border-color: var(--color-primary);
    transform: scale(1.05);
  }

  .avatar-image {
    width: 100%;
    height: 100%;
    object-fit: cover;
    object-position: center;
  }

  .avatar-fallback {
    width: 100%;
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
    background-color: var(--color-primary-light);
    color: white;
  }

  .avatar-initial {
    font-size: 1.2rem;
    font-weight: bold;
  }

  .profile-info {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 0px;
  }

  .name-container {
    display: flex;
    align-items: center;
    gap: var(--space-1);
  }

  .display-name {
    font-weight: var(--font-weight-bold);
    color: var(--text-primary);
    margin: 0;
    font-size: var(--font-size-base);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .username {
    color: var(--text-secondary);
    font-size: var(--font-size-sm);
    margin: 2px 0 0 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .bio-text {
    color: var(--text-secondary);
    font-size: var(--font-size-sm);
    margin: var(--space-2) 0 0 0;
    overflow: hidden;
    text-overflow: ellipsis;
    display: -webkit-box;
    -webkit-box-orient: vertical;
    -webkit-line-clamp: 2;
    line-height: 1.4;
  }

  .follower-count {
    color: var(--text-tertiary);
    font-size: var(--font-size-sm);
    margin: var(--space-2) 0 0 0;
    display: flex;
    align-items: center;
    gap: 4px;
  }

  .follower-count::before {
    content: '';
    display: inline-block;
    width: 14px;
    height: 14px;
    background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' fill='none' viewBox='0 0 24 24' stroke='currentColor'%3E%3Cpath stroke-linecap='round' stroke-linejoin='round' stroke-width='2' d='M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z'%3E%3C/path%3E%3C/svg%3E");
    background-repeat: no-repeat;
    background-position: center;
    background-size: contain;
    opacity: 0.7;
  }

  .follower-number {
    font-weight: var(--font-weight-bold);
  }

  .verified-badge {
    display: inline-flex;
    color: var(--color-primary);
    width: 16px;
    height: 16px;
    margin-left: 4px;
  }

  .verified-badge svg {
    width: 16px;
    height: 16px;
    color: var(--color-primary);
  }

  .fuzzy-match-badge {
    display: flex;
    align-items: center;
    font-size: var(--font-size-xs);
    padding: 1px 4px;
    border-radius: var(--radius-full);
    margin-left: 4px;
    font-weight: 500;
    font-size: 0.65rem;
  }

  .fuzzy-match-icon {
    width: 10px;
    height: 10px;
    margin-right: 2px;
  }

  .fuzzy-match-high {
    color: #16a34a;
    background-color: rgba(22, 163, 74, 0.1);
  }

  .fuzzy-match-medium {
    color: #ca8a04;
    background-color: rgba(202, 138, 4, 0.1);
  }

  .fuzzy-match-low {
    color: #dc2626;
    background-color: rgba(220, 38, 38, 0.1);
  }
</style>