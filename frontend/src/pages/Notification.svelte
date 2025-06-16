<script lang="ts">
  import { onMount } from "svelte";
  import MainLayout from "../components/layout/MainLayout.svelte";
  import { useAuth } from "../hooks/useAuth";
  import { useTheme } from "../hooks/useTheme";
  import type { IAuthStore } from "../interfaces/IAuth";
  import { createLoggerWithPrefix } from "../utils/logger";
  import { toastStore } from "../stores/toastStore";
  import { checkAuth as validateAuth, formatTimeAgo, handleApiError } from "../utils/common";
  import { getNotifications, getMentions, markNotificationAsRead, getUserInteractionNotifications } from "../api/notifications";
  import { getProfile } from "../api/user";

  const logger = createLoggerWithPrefix("Notification");

  // Auth and theme
  const { getAuthState } = useAuth();
  const { theme } = useTheme();

  // Reactive declarations
  $: authState = getAuthState ? (getAuthState() as IAuthStore) : { user_id: null, is_authenticated: false, access_token: null, refresh_token: null };
  $: isDarkMode = $theme === "dark";

  // User profile data
  let username = "";
  let displayName = "";
  let avatar = "https://secure.gravatar.com/avatar/0?d=mp"; // Default avatar with proper image URL
  let isLoadingProfile = true;
    // Notification state
  let isLoading = true;
  let activeTab: "all" | "mentions" | "likes" | "bookmarks" | "replies" = "all";
  let notifications: Notification[] = [];
  let mentions: Notification[] = [];
  let userInteractions: any = null;    // Notification interface
  interface Notification {
    id: string;
    type: "like" | "repost" | "follow" | "mention" | "bookmark" | "reply";
    user_id: string;
    username: string;
    display_name: string;
    avatar: string | null;
    timestamp: string;
    thread_id?: string;
    thread_content?: string;
    is_read: boolean;
  }
    // Authentication check
  function validateUserAuth() {
    if (!authState.is_authenticated) {
      toastStore.showToast("You need to log in to access notifications", "warning");
      window.location.href = "/login";
      return false;
    }
    return true;
  }

  // Fetch user profile data
  async function fetchUserProfile() {
    isLoadingProfile = true;
    try {
      const response = await getProfile();
      if (response && response.user) {        const userData = response.user;
        username = userData.username || `user_${authState.user_id?.substring(0, 4)}`;
        displayName = userData.name || userData.display_name || `User ${authState.user_id?.substring(0, 4)}`;
        avatar = userData.profile_picture_url || "https://secure.gravatar.com/avatar/0?d=mp";
        logger.debug("Profile loaded", { username });      } else {
        logger.warn("No user data received from API");
        username = `user_${authState.user_id?.substring(0, 4)}`;
        displayName = `User ${authState.user_id?.substring(0, 4)}`;
      }    } catch (error) {
      const errorResponse = handleApiError(error);
      logger.error("Error fetching user profile:", errorResponse);
      username = `user_${authState.user_id?.substring(0, 4)}`;
      displayName = `User ${authState.user_id?.substring(0, 4)}`;
    } finally {
      isLoadingProfile = false;
    }
  }
    // Fetch notifications
  async function fetchNotifications() {
    isLoading = true;

    try {
      // Fetch notifications from API
      const notificationsData = await getNotifications();
      notifications = notificationsData.map(notification => ({
        id: notification.id,
        type: notification.type,
        user_id: notification.user_id,
        username: notification.username,
        display_name: notification.display_name,
        avatar: notification.avatar,
        timestamp: notification.timestamp,
        thread_id: notification.thread_id,
        thread_content: notification.thread_content,
        is_read: notification.is_read
      }));

      // Fetch mentions from API
      const mentionsData = await getMentions();
      mentions = mentionsData.map(mention => ({
        id: mention.id,
        type: "mention",
        user_id: mention.user_id,
        username: mention.username,
        display_name: mention.display_name,
        avatar: mention.avatar,
        timestamp: mention.timestamp,
        thread_id: mention.thread_id,
        thread_content: mention.thread_content,
        is_read: mention.is_read
      }));

      // Fetch user interactions (likes, bookmarks, replies)
      try {
        userInteractions = await getUserInteractionNotifications();
      } catch (error) {
        logger.warn("Failed to fetch user interactions:", error);
        userInteractions = { likes: [], bookmarks: [], replies: [], follows: [] };
      }

      logger.debug("Notifications loaded", {
        notifications: notifications.length,
        mentions: mentions.length,
        likes: userInteractions?.likes?.length || 0,
        bookmarks: userInteractions?.bookmarks?.length || 0,
        replies: userInteractions?.replies?.length || 0
      });
    } catch (error) {
      const errorResponse = handleApiError(error);
      logger.error("Error fetching notifications:", errorResponse);
      toastStore.showToast("Failed to load notifications. Please try again.", "error");
    } finally {
      isLoading = false;
    }
  }
  // Handle tab change
  function handleTabChange(tab: "all" | "mentions" | "likes" | "bookmarks" | "replies") {
    activeTab = tab;
    logger.debug("Tab changed", { tab });
  }

  // Get data for current tab
  function getCurrentTabData(): Notification[] {
    switch (activeTab) {
      case "all":
        return notifications;
      case "mentions":
        return mentions;
      case "likes":
        return userInteractions?.likes || [];
      case "bookmarks":
        return userInteractions?.bookmarks || [];
      case "replies":
        return userInteractions?.replies || [];
      default:
        return [];
    }
  }

  // Get count for tab badge
  function getTabCount(tab: string): number {
    switch (tab) {
      case "all":
        return notifications.length;
      case "mentions":
        return mentions.length;
      case "likes":
        return userInteractions?.likes?.length || 0;
      case "bookmarks":
        return userInteractions?.bookmarks?.length || 0;
      case "replies":
        return userInteractions?.replies?.length || 0;
      default:
        return 0;
    }
  }
    // Navigate to thread or profile based on notification type
  async function handleNotificationClick(notification: Notification) {
    try {
      // Mark notification as read
      await markNotificationAsRead(notification.id);
        // Update local state
      if (notification.type === "mention") {
        mentions = mentions.map(n =>
          n.id === notification.id ? { ...n, is_read: true } : n
        );
      } else if (activeTab === "likes" && userInteractions?.likes) {
        userInteractions.likes = userInteractions.likes.map(n =>
          n.id === notification.id ? { ...n, is_read: true } : n
        );
      } else if (activeTab === "bookmarks" && userInteractions?.bookmarks) {
        userInteractions.bookmarks = userInteractions.bookmarks.map(n =>
          n.id === notification.id ? { ...n, is_read: true } : n
        );
      } else if (activeTab === "replies" && userInteractions?.replies) {
        userInteractions.replies = userInteractions.replies.map(n =>
          n.id === notification.id ? { ...n, is_read: true } : n
        );
      } else {
        notifications = notifications.map(n =>
          n.id === notification.id ? { ...n, is_read: true } : n
        );
      }
        // Navigate based on notification type
      if (notification.type === "mention" || notification.type === "reply") {
        // For mentions and replies, navigate to the thread where mentioned/replied
        window.location.href = `/thread/${notification.thread_id}`;
      } else if (notification.thread_id && (notification.type === "like" || notification.type === "bookmark" || notification.type === "repost")) {
        // For likes, bookmarks, and reposts, navigate to the thread
        window.location.href = `/thread/${notification.thread_id}`;
      } else if (notification.type === "follow") {
        // For follows, navigate to the profile of the user who followed
        window.location.href = `/profile/${notification.username}`;
      }
    } catch (error) {
      const errorResponse = handleApiError(error);
      logger.error("Error marking notification as read:", errorResponse);
      toastStore.showToast("Failed to mark notification as read.", "error");
        // Still navigate even if marking as read fails
      if (notification.thread_id) {
        window.location.href = `/thread/${notification.thread_id}`;
      } else if (notification.type === "follow") {
        window.location.href = `/profile/${notification.username}`;
      }
    }
  }

  onMount(() => {
    logger.debug("Notification page mounted", { authState });
    if (validateUserAuth()) {
      // Fetch user profile first, then notifications
      fetchUserProfile().then(() => {
        fetchNotifications();
      });
    }
  });
</script>

<MainLayout
  {username}
  {displayName}
  {avatar}
>
  <div class="notification-container">
    <!-- Header -->
    <div class="notification-header">
      <h1 class="page-title">Notifications</h1>
        <!-- Tabs -->
      <div class="notification-tabs">
        <button
          class="tab-button {activeTab === "all" ? "active" : ""}"
          on:click={() => handleTabChange("all")}
        >
          All
          {#if getTabCount("all") > 0}
            <span class="tab-badge">{getTabCount("all")}</span>
          {/if}
        </button>
        <button
          class="tab-button {activeTab === "mentions" ? "active" : ""}"
          on:click={() => handleTabChange("mentions")}
        >
          Mentions
          {#if getTabCount("mentions") > 0}
            <span class="tab-badge">{getTabCount("mentions")}</span>
          {/if}
        </button>
        <button
          class="tab-button {activeTab === "likes" ? "active" : ""}"
          on:click={() => handleTabChange("likes")}
        >
          Likes
          {#if getTabCount("likes") > 0}
            <span class="tab-badge">{getTabCount("likes")}</span>
          {/if}
        </button>
        <button
          class="tab-button {activeTab === "bookmarks" ? "active" : ""}"
          on:click={() => handleTabChange("bookmarks")}
        >
          Bookmarks
          {#if getTabCount("bookmarks") > 0}
            <span class="tab-badge">{getTabCount("bookmarks")}</span>
          {/if}
        </button>
        <button
          class="tab-button {activeTab === "replies" ? "active" : ""}"
          on:click={() => handleTabChange("replies")}
        >
          Replies
          {#if getTabCount("replies") > 0}
            <span class="tab-badge">{getTabCount("replies")}</span>
          {/if}
        </button>
      </div>
    </div>
      <!-- Content Area -->
    <div class="notification-content">
      {#if isLoading}
        <div class="notification-loading">
          {#each Array(5) as _, i}
            <div class="notification-skeleton">
              <div class="notification-skeleton-icon"></div>
              <div class="notification-skeleton-avatar"></div>
              <div class="notification-skeleton-content">
                <div class="notification-skeleton-line"></div>
                <div class="notification-skeleton-line short"></div>
              </div>
              <div class="notification-skeleton-time"></div>
            </div>
          {/each}
        </div>
      {:else if getCurrentTabData().length === 0}
        <div class="notification-empty">
          <div class="notification-empty-icon">
            {#if activeTab === "likes"}
              <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M20.884 13.19c-1.351 2.48-4.001 5.12-8.379 7.67l-.503.3-.504-.3c-4.379-2.55-7.029-5.19-8.382-7.67-1.36-2.5-1.41-4.86-.514-6.67.887-1.79 2.647-2.91 4.601-3.01 1.651-.09 3.368.56 4.798 2.01 1.429-1.45 3.146-2.1 4.796-2.01 1.954.1 3.714 1.22 4.601 3.01.896 1.81.846 4.17-.514 6.67z"></path>
              </svg>
            {:else if activeTab === "bookmarks"}
              <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M19 21l-7-5-7 5V5a2 2 0 0 1 2-2h10a2 2 0 0 1 2 2z"></path>
              </svg>
            {:else if activeTab === "replies"}
              <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M21 11.5a8.38 8.38 0 0 1-.9 3.8 8.5 8.5 0 0 1-7.6 4.7 8.38 8.38 0 0 1-3.8-.9L3 21l1.9-5.7a8.38 8.38 0 0 1-.9-3.8 8.5 8.5 0 0 1 4.7-7.6 8.38 8.38 0 0 1 3.8-.9h.5a8.48 8.48 0 0 1 8 8v.5z"></path>
              </svg>
            {:else}
              <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9"></path>
                <path d="M13.73 21a2 2 0 0 1-3.46 0"></path>
              </svg>
            {/if}
          </div>
          <h2 class="notification-empty-title">
            {#if activeTab === "likes"}
              No likes yet
            {:else if activeTab === "bookmarks"}
              No bookmarks yet
            {:else if activeTab === "replies"}
              No replies yet
            {:else if activeTab === "mentions"}
              No mentions yet
            {:else}
              Nothing to see here—yet
            {/if}
          </h2>
          <p class="notification-empty-message">
            {#if activeTab === "likes"}
              When someone likes your posts, you'll see them here.
            {:else if activeTab === "bookmarks"}
              Posts you've bookmarked will appear here.
            {:else if activeTab === "replies"}
              Replies to your posts will show up here.
            {:else if activeTab === "mentions"}
              When someone mentions you, you'll see it here.
            {:else}
              From likes to reposts and a whole lot more, this is where all the action happens.
            {/if}
          </p>
        </div>
      {:else}
        <div class="notification-list">
          {#each getCurrentTabData() as notification}
            <button
              class="notification-item {notification.is_read ? "read" : ""} {activeTab}"
              on:click={() => handleNotificationClick(notification)}
            >
              <div class="notification-icon {notification.type}">
                {#if notification.type === "like"}
                  <svg viewBox="0 0 24 24" aria-hidden="true" fill="currentColor">
                    <g><path d="M20.884 13.19c-1.351 2.48-4.001 5.12-8.379 7.67l-.503.3-.504-.3c-4.379-2.55-7.029-5.19-8.382-7.67-1.36-2.5-1.41-4.86-.514-6.67.887-1.79 2.647-2.91 4.601-3.01 1.651-.09 3.368.56 4.798 2.01 1.429-1.45 3.146-2.1 4.796-2.01 1.954.1 3.714 1.22 4.601 3.01.896 1.81.846 4.17-.514 6.67z"></path></g>
                  </svg>
                {:else if notification.type === "bookmark"}
                  <svg viewBox="0 0 24 24" aria-hidden="true" fill="currentColor">
                    <g><path d="M19 21l-7-5-7 5V5a2 2 0 0 1 2-2h10a2 2 0 0 1 2 2z"></path></g>
                  </svg>
                {:else if notification.type === "reply"}
                  <svg viewBox="0 0 24 24" aria-hidden="true" fill="currentColor">
                    <g><path d="M21 11.5a8.38 8.38 0 0 1-.9 3.8 8.5 8.5 0 0 1-7.6 4.7 8.38 8.38 0 0 1-3.8-.9L3 21l1.9-5.7a8.38 8.38 0 0 1-.9-3.8 8.5 8.5 0 0 1 4.7-7.6 8.38 8.38 0 0 1 3.8-.9h.5a8.48 8.48 0 0 1 8 8v.5z"></path></g>
                  </svg>
                {:else if notification.type === "repost"}
                  <svg viewBox="0 0 24 24" aria-hidden="true" fill="currentColor">
                    <g><path d="M4.5 3.88l4.432 4.14-1.364 1.46L5.5 7.55V16c0 1.1.896 2 2 2H13v2H7.5c-2.209 0-4-1.79-4-4V7.55L1.432 9.48.068 8.02 4.5 3.88zM16.5 6H11V4h5.5c2.209 0 4 1.79 4 4v8.45l2.068-1.93 1.364 1.46-4.432 4.14-4.432-4.14 1.364-1.46 2.068 1.93V8c0-1.1-.896-2-2-2z"></path></g>
                  </svg>
                {:else if notification.type === "follow"}
                  <svg viewBox="0 0 24 24" aria-hidden="true" fill="currentColor">
                    <g><path d="M12 11.816c1.355 0 2.872-.15 3.84-1.256.814-.93 1.078-2.368.806-4.392-.38-2.825-2.117-4.512-4.646-4.512S7.734 3.343 7.354 6.17c-.272 2.022-.008 3.46.806 4.39.968 1.107 2.485 1.256 3.84 1.256zM8.84 6.368c.162-1.2.787-3.212 3.16-3.212s2.998 2.013 3.16 3.212c.207 1.55.057 2.627-.45 3.205-.455.52-1.266.743-2.71.743s-2.255-.223-2.71-.743c-.507-.578-.657-1.656-.45-3.205zm11.44 12.868c-.877-3.526-4.282-5.99-8.28-5.99s-7.403 2.464-8.28 5.99c-.172.692-.028 1.4.395 1.94.408.52 1.04.82 1.733.82h12.304c.693 0 1.325-.3 1.733-.82.424-.54.567-1.247.394-1.94zm-1.576 1.016c-.126.16-.316.246-.552.246H5.848c-.235 0-.426-.085-.552-.246-.137-.174-.18-.412-.12-.654.71-2.855 3.517-4.85 6.824-4.85s6.114 1.994 6.824 4.85c.06.242.017.48-.12.654z"></path></g>
                  </svg>
                {:else if notification.type === "mention"}
                  <svg viewBox="0 0 24 24" aria-hidden="true" fill="currentColor">
                    <g><path d="M12 11.816c1.355 0 2.872-.15 3.84-1.256.814-.93 1.078-2.368.806-4.392-.38-2.825-2.117-4.512-4.646-4.512S7.734 3.343 7.354 6.17c-.272 2.022-.008 3.46.806 4.39.968 1.107 2.485 1.256 3.84 1.256zM8.84 6.368c.162-1.2.787-3.212 3.16-3.212s2.998 2.013 3.16 3.212c.207 1.55.057 2.627-.45 3.205-.455.52-1.266.743-2.71.743s-2.255-.223-2.71-.743c-.507-.578-.657-1.656-.45-3.205zm11.44 12.868c-.877-3.526-4.282-5.99-8.28-5.99s-7.403 2.464-8.28 5.99c-.172.692-.028 1.4.395 1.94.408.52 1.04.82 1.733.82h12.304c.693 0 1.325-.3 1.733-.82.424-.54.567-1.247.394-1.94zm-1.576 1.016c-.126.16-.316.246-.552.246H5.848c-.235 0-.426-.085-.552-.246-.137-.174-.18-.412-.12-.654.71-2.855 3.517-4.85 6.824-4.85s6.114 1.994 6.824 4.85c.06.242.017.48-.12.654z"></path></g>
                  </svg>
                {/if}
              </div>
                <div class="notification-avatar">
                {#if notification.avatar}
                  <img src={notification.avatar} alt={notification.display_name} />
                {:else}
                  <div class="notification-avatar-placeholder">
                    {notification.display_name.charAt(0).toUpperCase()}
                  </div>
                {/if}
              </div>

              <div class="notification-content">
                <div class="notification-text">
                  <span class="notification-name">{notification.display_name}</span>
                  {#if notification.type === "like"}
                    liked your post
                  {:else if notification.type === "bookmark"}
                    bookmarked your post
                  {:else if notification.type === "reply"}
                    replied to your post
                  {:else if notification.type === "repost"}
                    reposted your post
                  {:else if notification.type === "follow"}
                    followed you
                  {:else if notification.type === "mention"}
                    mentioned you
                  {/if}
                </div>
                  {#if notification.thread_content && notification.type !== "follow"}
                  <div class="notification-thread">
                    {notification.thread_content}
                  </div>
                {/if}
              </div>

              <div class="notification-time">
                {formatTimeAgo(notification.timestamp)}
              </div>
            </button>
          {/each}
        </div>
      {/if}
    </div>
  </div>
</MainLayout>

<style>
  .notification-container {
    width: 100%;
    max-width: var(--container-width);
    margin: 0 auto;
    background-color: var(--bg-primary);
    color: var(--text-primary);
    border: 1px solid var(--border-color);
    border-top: none;
    border-bottom: none;
    min-height: 100vh;
  }

  /* Header styles */
  .notification-header {
    padding: var(--space-4);
    border-bottom: 1px solid var(--border-color);
    position: sticky;
    top: 0;
    background-color: var(--bg-primary);
    z-index: 10;
    backdrop-filter: blur(8px);
  }

  .page-title {
    font-size: var(--font-xl);
    font-weight: 700;
    margin-bottom: var(--space-3);
  }
    /* Tab styles */
  .notification-tabs {
    display: flex;
    border-bottom: 1px solid var(--border-color);
    overflow-x: auto;
    scrollbar-width: none;
    -ms-overflow-style: none;
  }

  .notification-tabs::-webkit-scrollbar {
    display: none;
  }

  .tab-button {
    flex: 1;
    min-width: fit-content;
    padding: var(--space-3) var(--space-2);
    background: transparent;
    color: var(--text-secondary);
    border: none;
    font-weight: 600;
    font-size: var(--font-sm);
    position: relative;
    cursor: pointer;
    transition: color 0.2s;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: var(--space-1);
    white-space: nowrap;
  }

  .tab-button:hover {
    color: var(--text-primary);
    background-color: var(--bg-hover);
  }

  .tab-button.active {
    color: var(--text-primary);
  }

  .tab-button.active::after {
    content: '';
    position: absolute;
    bottom: -1px;
    left: 0;
    right: 0;
    margin: 0 auto;
    width: 50%;
    height: 4px;
    background-color: var(--color-primary);
    border-radius: 2px;
  }

  .tab-badge {
    background-color: var(--color-primary);
    color: white;
    font-size: var(--font-xs);
    font-weight: 600;
    padding: 2px 6px;
    border-radius: 10px;
    min-width: 18px;
    height: 18px;
    display: flex;
    align-items: center;
    justify-content: center;
    line-height: 1;
  }

  /* Loading skeleton styles */
  .notification-loading {
    padding: var(--space-2) 0;
  }

  .notification-skeleton {
    display: flex;
    padding: var(--space-3) var(--space-4);
    align-items: flex-start;
    gap: var(--space-3);
    border-bottom: 1px solid var(--border-color);
    animation: pulse 1.5s infinite ease-in-out;
  }

  .notification-skeleton-icon {
    width: 16px;
    height: 16px;
    border-radius: 50%;
    background-color: var(--skeleton-color);
    flex-shrink: 0;
  }

  .notification-skeleton-avatar {
    width: 36px;
    height: 36px;
    border-radius: 50%;
    background-color: var(--skeleton-color);
    flex-shrink: 0;
  }

  .notification-skeleton-content {
    flex: 1;
  }

  .notification-skeleton-line {
    height: 14px;
    width: 80%;
    background-color: var(--skeleton-color);
    border-radius: 4px;
    margin-bottom: var(--space-2);
  }

  .notification-skeleton-line.short {
    width: 40%;
  }

  .notification-skeleton-time {
    width: 60px;
    height: 12px;
    background-color: var(--skeleton-color);
    border-radius: 4px;
    flex-shrink: 0;
  }

  @keyframes pulse {
    0% { opacity: 0.6; }
    50% { opacity: 0.8; }
    100% { opacity: 0.6; }
  }

  /* Empty state styles */
  .notification-empty {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    text-align: center;
    padding: var(--space-6);
    margin: var(--space-10) 0;
  }

  .notification-empty-icon {
    width: 60px;
    height: 60px;
    display: flex;
    align-items: center;
    justify-content: center;
    margin-bottom: var(--space-4);
    color: var(--text-secondary);
  }

  .notification-empty-title {
    font-size: var(--font-lg);
    font-weight: 700;
    margin-bottom: var(--space-2);
    color: var(--text-primary);
  }

  .notification-empty-message {
    font-size: var(--font-md);
    color: var(--text-secondary);
    max-width: 400px;
    line-height: 1.4;
  }

  /* Notification list styles */
  .notification-list {
    display: flex;
    flex-direction: column;
  }

  .notification-item {
    display: flex;
    padding: var(--space-3) var(--space-4);
    gap: var(--space-3);
    text-align: left;
    border: none;
    border-bottom: 1px solid var(--border-color);
    background-color: var(--bg-primary);
    cursor: pointer;
    transition: background-color 0.2s;
    align-items: flex-start;
    position: relative;
  }

  .notification-item:not(.read) {
    background-color: var(--bg-highlight);
  }

  .notification-item:hover {
    background-color: var(--bg-hover);
  }

  .notification-icon {
    width: 16px;
    height: 16px;
    display: flex;
    align-items: center;
    justify-content: center;
    margin-top: var(--space-1);
    flex-shrink: 0;
  }
    .notification-icon.like {
    color: var(--color-like);
  }

  .notification-icon.bookmark {
    color: var(--color-bookmark, #1DA1F2);
  }

  .notification-icon.reply {
    color: var(--color-reply, #1DA1F2);
  }

  .notification-icon.repost {
    color: var(--color-repost);
  }

  .notification-icon.follow {
    color: var(--color-primary);
  }

  .notification-icon.mention {
    color: var(--color-primary);
  }

  .notification-icon svg {
    width: 16px;
    height: 16px;
  }

  .notification-avatar {
    width: 36px;
    height: 36px;
    border-radius: 50%;
    overflow: hidden;
    flex-shrink: 0;
  }

  .notification-avatar img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  .notification-avatar-placeholder {
    width: 100%;
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
    background-color: var(--color-primary);
    color: white;
    font-weight: 600;
  }

  .notification-content {
    flex: 1;
    min-width: 0;
  }

  .notification-text {
    line-height: 1.4;
    font-size: var(--font-md);
    color: var(--text-primary);
    margin-bottom: var(--space-1);
  }

  .notification-name {
    font-weight: 600;
    color: var(--text-primary);
    margin-right: var(--space-1);
  }
    .notification-thread {
    font-size: var(--font-md);
    color: var(--text-secondary);
    margin-top: var(--space-1);
    line-height: 1.4;    overflow: hidden;
    text-overflow: ellipsis;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    line-clamp: 2;
    -webkit-box-orient: vertical;
    /* Fallback for browsers that don't support line-clamp */
    max-height: calc(1.4em * 2);
  }

  .notification-time {
    font-size: var(--font-sm);
    color: var(--text-tertiary);
    flex-shrink: 0;
    margin-left: var(--space-2);
  }

  /* Mention notifications */
  .notification-item.mention {
    border-left: 3px solid var(--color-primary);
  }

  /* Read state */
  .notification-item.read {
    opacity: 0.9;
  }
</style>
