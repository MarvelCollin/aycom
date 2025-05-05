<script lang="ts">
  import { onMount } from 'svelte';
  import MainLayout from '../components/layout/MainLayout.svelte';
  import { useAuth } from '../hooks/useAuth';
  import { useTheme } from '../hooks/useTheme';
  import type { IAuthStore } from '../interfaces/IAuth';
  import { createLoggerWithPrefix } from '../utils/logger';
  import { toastStore } from '../stores/toastStore';
  import { checkAuth as validateAuth, formatTimeAgo, handleApiError, getUserProfile } from '../utils/common';
  import { getNotifications, getMentions, markNotificationAsRead } from '../api/notifications';
  import { getProfile } from '../api/user';
  
  const logger = createLoggerWithPrefix('Notification');
  
  // Auth and theme
  const { getAuthState } = useAuth();
  const { theme } = useTheme();
  
  // Reactive declarations
  $: authState = getAuthState ? (getAuthState() as IAuthStore) : { userId: null, isAuthenticated: false, accessToken: null, refreshToken: null };
  $: isDarkMode = $theme === 'dark';
  
  // User profile data
  let username = '';
  let displayName = '';
  let avatar = 'https://secure.gravatar.com/avatar/0?d=mp'; // Default avatar with proper image URL
  let isLoadingProfile = true;
  
  // Notification state
  let isLoading = true;
  let activeTab: 'all' | 'mentions' = 'all';
  let notifications: Notification[] = [];
  let mentions: Notification[] = [];
  
  // Notification interface
  interface Notification {
    id: string;
    type: 'like' | 'repost' | 'follow' | 'mention';
    userId: string;
    username: string;
    displayName: string;
    avatar: string | null;
    timestamp: string;
    threadId?: string;
    threadContent?: string;
    isRead: boolean;
  }
  
  // Authentication check
  function validateUserAuth() {
    if (!authState.isAuthenticated) {
      toastStore.showToast('You need to log in to access notifications', 'warning');
      window.location.href = '/login';
      return false;
    }
    return true;
  }
  
  // Fetch user profile data
  async function fetchUserProfile() {
    isLoadingProfile = true;
    try {
      const profileData = await getUserProfile(authState);
      username = profileData.username;
      displayName = profileData.displayName;
      avatar = profileData.avatar;
      logger.debug('Profile loaded', { username });
    } catch (error) {
      const errorResponse = handleApiError(error);
      logger.error('Error fetching user profile:', errorResponse);
      username = `user_${authState.userId?.substring(0, 4)}`;
      displayName = `User ${authState.userId?.substring(0, 4)}`;
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
        userId: notification.user_id,
        username: notification.username,
        displayName: notification.display_name,
        avatar: notification.avatar,
        timestamp: notification.timestamp,
        threadId: notification.thread_id,
        threadContent: notification.thread_content,
        isRead: notification.is_read
      }));
      
      // Fetch mentions from API
      const mentionsData = await getMentions();
      mentions = mentionsData.map(mention => ({
        id: mention.id,
        type: 'mention',
        userId: mention.user_id,
        username: mention.username,
        displayName: mention.display_name,
        avatar: mention.avatar,
        timestamp: mention.timestamp,
        threadId: mention.thread_id,
        threadContent: mention.thread_content,
        isRead: mention.is_read
      }));
      
      logger.debug('Notifications loaded', { count: notifications.length + mentions.length });
    } catch (error) {
      const errorResponse = handleApiError(error);
      logger.error('Error fetching notifications:', errorResponse);
      toastStore.showToast('Failed to load notifications. Please try again.', 'error');
    } finally {
      isLoading = false;
    }
  }
  
  // Handle tab change
  function handleTabChange(tab: 'all' | 'mentions') {
    activeTab = tab;
    logger.debug('Tab changed', { tab });
  }
  
  // Navigate to thread or profile based on notification type
  async function handleNotificationClick(notification: Notification) {
    try {
      // Mark notification as read
      await markNotificationAsRead(notification.id);
      
      // Update local state
      if (notification.type === 'mention') {
        mentions = mentions.map(n => 
          n.id === notification.id ? { ...n, isRead: true } : n
        );
      } else {
        notifications = notifications.map(n => 
          n.id === notification.id ? { ...n, isRead: true } : n
        );
      }
      
      // Navigate based on notification type
      if (notification.type === 'mention') {
        // For mentions, navigate to the thread where mentioned
        window.location.href = `/thread/${notification.threadId}`;
      } else if (notification.threadId) {
        // For likes and reposts, navigate to the thread
        window.location.href = `/thread/${notification.threadId}`;
      } else if (notification.type === 'follow') {
        // For follows, navigate to the profile of the user who followed
        window.location.href = `/profile/${notification.username}`;
      }
    } catch (error) {
      const errorResponse = handleApiError(error);
      logger.error('Error marking notification as read:', errorResponse);
      toastStore.showToast('Failed to mark notification as read.', 'error');
      
      // Still navigate even if marking as read fails
      if (notification.threadId) {
        window.location.href = `/thread/${notification.threadId}`;
      } else if (notification.type === 'follow') {
        window.location.href = `/profile/${notification.username}`;
      }
    }
  }
  
  onMount(() => {
    logger.debug('Notification page mounted', { authState });
    if (validateUserAuth()) {
      // Fetch user profile first, then notifications
      fetchUserProfile().then(() => {
        fetchNotifications();
      });
    }
  });
</script>

<MainLayout
  username={username}
  displayName={displayName}
  avatar={avatar}
  on:toggleComposeModal={() => {}}
>
  <div class="min-h-screen border-x border-gray-200 dark:border-gray-800">
    <!-- Header -->
    <div class="sticky top-0 z-10 bg-white/80 dark:bg-black/80 backdrop-blur-md border-b border-gray-200 dark:border-gray-800 px-4 py-3">
      <h1 class="text-xl font-bold">Notifications</h1>
      
      <!-- Tabs -->
      <div class="flex mt-3 border-b border-gray-200 dark:border-gray-800">
        <button 
          class="px-4 py-3 font-medium {activeTab === 'all' ? 'text-blue-500 border-b-2 border-blue-500' : 'text-gray-500 dark:text-gray-400'}"
          on:click={() => handleTabChange('all')}
        >
          All
        </button>
        <button 
          class="px-4 py-3 font-medium {activeTab === 'mentions' ? 'text-blue-500 border-b-2 border-blue-500' : 'text-gray-500 dark:text-gray-400'}"
          on:click={() => handleTabChange('mentions')}
        >
          Mentions
        </button>
      </div>
    </div>
    
    <!-- Content Area -->
    <div class="overflow-y-auto">
      {#if isLoading}
        <div class="p-4">
          <div class="animate-pulse space-y-4">
            {#each Array(5) as _}
              <div class="flex space-x-4">
                <div class="rounded-full bg-gray-300 dark:bg-gray-700 h-12 w-12"></div>
                <div class="flex-1 space-y-2 py-1">
                  <div class="h-4 bg-gray-300 dark:bg-gray-700 rounded w-3/4"></div>
                  <div class="h-4 bg-gray-300 dark:bg-gray-700 rounded w-1/2"></div>
                </div>
              </div>
            {/each}
          </div>
        </div>
      {:else if (activeTab === 'all' && notifications.length === 0) || (activeTab === 'mentions' && mentions.length === 0)}
        <div class="flex flex-col items-center justify-center py-16 px-4 text-center">
          <h2 class="text-2xl font-bold mb-2">Nothing to see here—yet</h2>
          <p class="text-gray-500 dark:text-gray-400 mb-8 max-w-md">
            From likes to reposts and a whole lot more, this is where all the action happens.
          </p>
        </div>
      {:else if activeTab === 'all'}
        <div class="divide-y divide-gray-200 dark:divide-gray-800">
          {#each notifications as notification}
            <button 
              class="w-full text-left p-4 hover:bg-gray-50 dark:hover:bg-gray-900 transition {notification.isRead ? 'opacity-70' : ''}"
              on:click={() => handleNotificationClick(notification)}
            >
              <div class="flex">
                <div class="flex-shrink-0">
                  <div class="w-12 h-12 rounded-full bg-gray-200 dark:bg-gray-700 flex items-center justify-center overflow-hidden">
                    {#if notification.avatar}
                      <img src={notification.avatar} alt={notification.displayName} class="w-full h-full object-cover" />
                    {:else}
                      <span class="text-xl">👤</span>
                    {/if}
                  </div>
                </div>
                <div class="ml-3 flex-1">
                  <div class="flex items-center">
                    <span class="font-semibold">{notification.displayName}</span>
                    <span class="text-gray-500 dark:text-gray-400 ml-1">@{notification.username}</span>
                    <span class="text-gray-500 dark:text-gray-400 text-sm ml-auto">{formatTimeAgo(notification.timestamp)}</span>
                  </div>
                  
                  <p class="text-gray-500 dark:text-gray-400 mt-1">
                    {#if notification.type === 'like'}
                      liked your thread
                    {:else if notification.type === 'repost'}
                      reposted your thread
                    {:else if notification.type === 'follow'}
                      followed you
                    {/if}
                  </p>
                  
                  {#if notification.threadContent && (notification.type === 'like' || notification.type === 'repost')}
                    <p class="text-gray-800 dark:text-gray-200 mt-2 line-clamp-2">{notification.threadContent}</p>
                  {/if}
                </div>
              </div>
            </button>
          {/each}
        </div>
      {:else if activeTab === 'mentions'}
        <div class="divide-y divide-gray-200 dark:divide-gray-800">
          {#each mentions as mention}
            <button 
              class="w-full text-left p-4 hover:bg-gray-50 dark:hover:bg-gray-900 transition {mention.isRead ? 'opacity-70' : ''}"
              on:click={() => handleNotificationClick(mention)}
            >
              <div class="flex">
                <div class="flex-shrink-0">
                  <div class="w-12 h-12 rounded-full bg-gray-200 dark:bg-gray-700 flex items-center justify-center overflow-hidden">
                    {#if mention.avatar}
                      <img src={mention.avatar} alt={mention.displayName} class="w-full h-full object-cover" />
                    {:else}
                      <span class="text-xl">👤</span>
                    {/if}
                  </div>
                </div>
                <div class="ml-3 flex-1">
                  <div class="flex items-center">
                    <span class="font-semibold">{mention.displayName}</span>
                    <span class="text-gray-500 dark:text-gray-400 ml-1">@{mention.username}</span>
                    <span class="text-gray-500 dark:text-gray-400 text-sm ml-auto">{formatTimeAgo(mention.timestamp)}</span>
                  </div>
                  
                  <p class="text-gray-500 dark:text-gray-400 mt-1">mentioned you</p>
                  
                  {#if mention.threadContent}
                    <p class="text-gray-800 dark:text-gray-200 mt-2 line-clamp-2">{mention.threadContent}</p>
                  {/if}
                </div>
              </div>
            </button>
          {/each}
        </div>
      {/if}
    </div>
  </div>
</MainLayout>

<!-- 
  BACKEND REQUIREMENTS:
  1. Real-time notifications:
     - Implement WebSocket connection to deliver notifications in real-time
     - When events occur (likes, reposts, follows, mentions), push notifications to connected clients
     - Store notifications in database for persistence and retrieval when user connects

  2. Email notifications:
     - Set up email service integration (SendGrid, Amazon SES, etc.)
     - Create email templates for different notification types
     - Send emails for all notification events
     - Allow users to configure email notification preferences
     
  3. API Endpoints needed:
     - GET /api/notifications - Fetch all notifications for the authenticated user
     - GET /api/notifications/mentions - Fetch only mention notifications
     - PUT /api/notifications/:id/read - Mark a notification as read
     - PUT /api/notifications/read-all - Mark all notifications as read
     - DELETE /api/notifications/:id - Delete a notification
     
  4. Notification storage schema:
     - id: unique identifier
     - userId: ID of the user receiving the notification
     - actorId: ID of the user who performed the action
     - type: enum ('like', 'repost', 'follow', 'mention')
     - entityId: ID of the related entity (thread ID for likes/reposts/mentions)
     - content: Optional content snippet (for mentions/thread content)
     - isRead: boolean
     - createdAt: timestamp
     
  5. Real-time implementation considerations:
     - Use Socket.io or similar for WebSocket connections
     - Implement room-based subscription where each user joins their own notification room
     - When notification events occur, emit to the appropriate user's room
     - Handle reconnection gracefully to avoid missing notifications
-->

<style>
  /* Limit multiline text to 2 lines */
  .line-clamp-2 {
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }
  
  /* Skeleton loading animation */
  @keyframes pulse {
    0%, 100% { opacity: 0.5; }
    50% { opacity: 1; }
  }
  .animate-pulse {
    animation: pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite;
  }
  
  .notification-content {
    display: -webkit-box;
    -webkit-line-clamp: 2;
    line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  
  /* Active tab styling */
  .tab-active {
    color: rgb(29, 155, 240);
    font-weight: 600;
  }
  
  .tab-active::after {
    content: '';
    position: absolute;
    bottom: -1px;
    left: 0;
    right: 0;
    height: 4px;
    background-color: rgb(29, 155, 240);
    border-radius: 9999px;
  }
</style>
