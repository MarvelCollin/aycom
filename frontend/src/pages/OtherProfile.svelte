<script lang="ts">
  import { onMount } from 'svelte';
  import MainLayout from '../components/layout/MainLayout.svelte';
  import { useAuth } from '../hooks/useAuth';
  import { useTheme } from '../hooks/useTheme';
  import { isAuthenticated, getUserId } from '../utils/auth';
  import { getUserById, followUser, unfollowUser, reportUser, blockUser, unblockUser } from '../api/user';
  import { getUserThreads, getUserReplies, getUserMedia } from '../api/thread';
  import { toastStore } from '../stores/toastStore';
  import ThreadCard from '../components/explore/ThreadCard.svelte';
  import TweetCard from '../components/social/TweetCard.svelte';
  import LoadingSkeleton from '../components/common/LoadingSkeleton.svelte';
  import type { ITweet } from '../interfaces/ISocialMedia';
  import { page } from '../stores/routeStore';
  import { createLoggerWithPrefix } from '../utils/logger';
  
  const logger = createLoggerWithPrefix('OtherProfile');

  // Import Feather icons
  import CalendarIcon from 'svelte-feather-icons/src/icons/CalendarIcon.svelte';
  import XIcon from 'svelte-feather-icons/src/icons/XIcon.svelte';
  import UserIcon from 'svelte-feather-icons/src/icons/UserIcon.svelte';
  import FlagIcon from 'svelte-feather-icons/src/icons/FlagIcon.svelte';
  import ShieldIcon from 'svelte-feather-icons/src/icons/ShieldIcon.svelte';
  import SlashIcon from 'svelte-feather-icons/src/icons/SlashIcon.svelte';
  import AlertCircleIcon from 'svelte-feather-icons/src/icons/AlertCircleIcon.svelte';
  import MoreHorizontalIcon from 'svelte-feather-icons/src/icons/MoreHorizontalIcon.svelte';
  import ArrowLeftIcon from 'svelte-feather-icons/src/icons/ArrowLeftIcon.svelte';
  import LinkIcon from 'svelte-feather-icons/src/icons/LinkIcon.svelte';
  import MapPinIcon from 'svelte-feather-icons/src/icons/MapPinIcon.svelte';
  
  // Define interfaces for our data structures
  interface Thread {
    id: string;
    content: string;
    username: string;
    displayName: string;
    timestamp: string;
    likes: number;
    replies: number;
    reposts: number;
    created_at: string;
    author_id?: string;
    author_username?: string;
    author_name?: string;
    author_avatar?: string;
    likes_count?: number;
    replies_count?: number;
    is_liked?: boolean;
    is_reposted?: boolean;
    is_bookmarked?: boolean;
    media?: Array<{
      type: string;
      url: string;
    }>;
    avatar?: string;
    [key: string]: any; // For any additional properties
  }
  
  interface Reply {
    id: string;
    content: string;
    created_at: string;
    thread_id: string;
    thread_author: string;
    author_id?: string;
    author_username?: string;
    author_name?: string;
    author_avatar?: string;
    likes_count?: number;
    is_liked?: boolean;
    [key: string]: any; // For any additional properties
  }
  
  interface ThreadMedia {
    id: string;
    url: string;
    type: 'image' | 'video' | 'gif';
    thread_id: string;
    created_at?: string;
    [key: string]: any; // For any additional properties
  }
  
  // Auth and theme
  const { getAuthState } = useAuth();
  const { theme } = useTheme();
  
  // Get userId from URL parameter
  export let userId: string;
  
  // Reactive declarations
  $: isDarkMode = $theme === 'dark';
  $: authState = getAuthState();
  $: currentUserId = getUserId();
  
  $: {
    // Log whenever userId changes
    logger.debug(`Profile rendering for userId: ${userId}, currentUserId: ${currentUserId}`);
  }
  
  // Profile data
  let profileData = {
    id: '',
    username: '',
    displayName: '',
    bio: '',
    profilePicture: '',
    backgroundBanner: '',
    followerCount: 0,
    followingCount: 0,
    joinedDate: '',
    location: '',
    website: '',
    isPrivate: false,
    isFollowing: false,
    isBlocked: false
  };
  
  // Content data with types
  let posts: Thread[] = [];
  let replies: Reply[] = [];
  let media: ThreadMedia[] = [];
  
  // UI state
  let activeTab = 'posts';
  let isLoading = true;
  let isFollowLoading = false;
  let isBlockLoading = false;
  let showReportModal = false;
  let reportReason = '';
  let showBlockConfirmModal = false;
  let showActionsDropdown = false;
  let errorMessage = '';
  let retryCount = 0;
  const MAX_RETRIES = 3;
  
  // Helper function to ensure an object has all ITweet properties
  function ensureTweetFormat(thread: any): ITweet {
    try {
      if (!thread || typeof thread !== 'object') {
        logger.warn('Invalid thread object provided to ensureTweetFormat');
        // Return a minimal valid object to prevent rendering errors
        return {
          id: `invalid-${Math.random().toString(36).substring(2, 9)}`,
          threadId: '',
          userId: '',
          username: 'unknown',
          displayName: 'Unknown User',
          content: 'This content is unavailable',
          timestamp: new Date().toISOString(),
          avatar: '',
          likes: 0,
          replies: 0,
          reposts: 0,
          bookmarks: 0,
          views: 0,
          media: [],
          isLiked: false,
          isReposted: false,
          isBookmarked: false,
          isPinned: false,
          replyTo: null
        };
      }
    
      // Get username from all possible sources
      let username = thread.author_username || thread.authorUsername || thread.username || 'anonymous';
      
      // Get display name from all possible sources
      let displayName = thread.author_name || thread.authorName || thread.display_name || 
                        thread.displayName || username || 'User';
      
      // Get profile picture from all possible sources
      let profilePicture = thread.author_avatar || thread.authorAvatar || 
                          thread.profile_picture_url || thread.profilePictureUrl || 
                          thread.avatar || 'https://secure.gravatar.com/avatar/0?d=mp';
      
      // Use the created_at timestamp if available, fall back to UTC now
      let timestamp = thread.created_at || thread.createdAt || thread.timestamp || new Date().toISOString();
      if (typeof timestamp === 'string' && !timestamp.includes('T')) {
        // Convert to ISO format if it's not already
        timestamp = new Date(timestamp).toISOString();
      }
      
      // Normalize metrics
      const likes = Number(thread.likes_count || thread.like_count || thread.metrics?.likes || 0);
      const replies = Number(thread.replies_count || thread.reply_count || thread.metrics?.replies || 0);
      const reposts = Number(thread.reposts_count || thread.repost_count || thread.metrics?.reposts || 0);
      const bookmarks = Number(thread.bookmarks_count || thread.bookmark_count || thread.metrics?.bookmarks || 0);
      const views = Number(thread.views || thread.views_count || 0);
      
      // Normalize interaction states
      const isLiked = Boolean(thread.is_liked || thread.isLiked || false);
      const isReposted = Boolean(thread.is_repost || thread.isReposted || false);
      const isBookmarked = Boolean(thread.is_bookmarked || thread.isBookmarked || false);
      const isPinned = Boolean(
        thread.is_pinned === true || 
        thread.is_pinned === 'true' || 
        thread.is_pinned === 1 || 
        thread.is_pinned === '1' || 
        thread.is_pinned === 't' || 
        thread.IsPinned === true || 
        false
      );
      
      // Ensure media array is valid
      const media = Array.isArray(thread.media) ? thread.media : [];
        
      // Ensure thread ID and thread author ID exist
      const id = thread.id || `thread-${Math.random().toString(36).substring(2, 9)}`;
      const userId = thread.user_id || thread.userId || thread.author_id || thread.authorId || '';
        
      return {
        id,
        threadId: thread.thread_id || id,
        userId,
        username,
        displayName,
        content: thread.content || '',
        timestamp: typeof timestamp === 'string' ? timestamp : new Date(timestamp).toISOString(),
        avatar: profilePicture,
        likes,
        replies,
        reposts,
        bookmarks,
        views,
        media,
        isLiked,
        isReposted,
        isBookmarked,
        isPinned,
        replyTo: thread.parent_id ? { id: thread.parent_id } as any : null
      };
    } catch (error) {
      logger.error('Error formatting tweet:', error);
      // Return a minimal valid object to prevent rendering errors
      return {
        id: `error-${Math.random().toString(36).substring(2, 9)}`,
        threadId: '',
        userId: '',
        username: 'error',
        displayName: 'Error',
        content: 'Error loading tweet',
        timestamp: new Date().toISOString(),
        avatar: '',
        likes: 0,
        replies: 0,
        reposts: 0,
        bookmarks: 0,
        views: 0,
        media: [],
        isLiked: false,
        isReposted: false,
        isBookmarked: false,
        isPinned: false,
        replyTo: null
      };
    }
  }

  function handleReply(event) {
    const threadId = event.detail;
    window.location.href = `/thread/${threadId}`;
  }

  function setActiveTab(tab) {
    activeTab = tab;
    loadTabContent(tab);
  }
  
  // Helper function to handle loading errors with specific messages
  function handleLoadError(error: any, context: string): string {
    logger.error(`Error in ${context}:`, error);
    
    const errorMessage = error?.message || String(error);
    
    // Check for specific error patterns
    if (errorMessage.includes('invalid UUID format')) {
      return `Invalid user ID format. Please use a valid username or ID.`;
    } else if (errorMessage.includes('not found') || errorMessage.includes('404')) {
      return `User not found. The account may have been deleted or username changed.`;
    } else if (errorMessage.includes('403') || errorMessage.includes('forbidden')) {
      return `You don't have permission to view this user's ${context}.`;
    } else if (errorMessage.includes('429') || errorMessage.includes('too many')) {
      return `Too many requests. Please wait a moment and try again.`;
    } else if (errorMessage.includes('timeout') || errorMessage.includes('timed out')) {
      return `Request timed out. The server might be busy, please try again later.`;
    } else if (errorMessage.includes('500')) {
      return `Server error while loading ${context}. Please try again later.`;
    }
    
    // Default message
    return `Failed to load ${context}. Please try again later.`;
  }
  
  async function loadTabContent(tab: string) {
    if (isLoading) return; // Don't load if already loading
    
    isLoading = true;
    errorMessage = '';
    
    try {
      // Check if user is private and not following
      if (profileData.isPrivate && !profileData.isFollowing && currentUserId !== userId) {
        logger.debug('User is private and not following, skipping content load');
        isLoading = false;
        return;
      }
      
      if (tab === 'posts') {
        // Load user's threads
        logger.debug(`Loading posts for user ${userId}`);
        try {
          const postsData = await getUserThreads(userId);
          
          // Safety check for valid data structure
          if (!postsData || (!postsData.threads && !postsData.data)) {
            logger.warn('Received invalid posts data structure:', postsData);
            posts = [];
            if (!posts.length) {
              logger.debug('No posts found for user');
            }
          } else {
            // Convert threads and ensure proper format
            const threadsArray = postsData.threads || postsData.data || [];
            posts = threadsArray.map(thread => ensureTweetFormat(thread));
            logger.debug(`Loaded ${posts.length} posts`);
            
            // Sort by creation date (newest first)
            posts.sort((a, b) => {
              const dateA = new Date(a.timestamp);
              const dateB = new Date(b.timestamp);
              return dateB.getTime() - dateA.getTime();
            });
          }
        } catch (error: any) {
          logger.error('Error loading posts:', error);
          posts = [];
          
          if (error.message?.includes('not found') || error.message?.includes('404')) {
            errorMessage = `User "${userId}" not found or has deleted their account.`;
          } else if (error.message?.includes('invalid UUID format')) {
            errorMessage = `Invalid user ID format. Please check the profile URL.`;
          } else if (error.message?.includes('timed out')) {
            errorMessage = `Request timed out. The server might be busy, please try again later.`;
          } else {
            errorMessage = `Error loading posts: ${error.message || 'Unknown error'}`;
          }
          
          throw error;
        }
      } else if (tab === 'replies') {
        // Load user's replies
        logger.debug(`Loading replies for user ${userId}`);
        try {
          const repliesData = await getUserReplies(userId);
          
          // Safety check for valid data structure
          if (!repliesData || (!repliesData.replies && !repliesData.data)) {
            logger.warn('Received invalid replies data structure:', repliesData);
            replies = [];
            if (!replies.length) {
              logger.debug('No replies found for user');
            }
          } else {
            const repliesArray = repliesData.replies || repliesData.data || [];
            replies = repliesArray.map(reply => ensureTweetFormat(reply));
            logger.debug(`Loaded ${replies.length} replies`);
            
            // Sort replies by date (newest first)
            replies.sort((a, b) => {
              const dateA = new Date(a.timestamp);
              const dateB = new Date(b.timestamp);
              return dateB.getTime() - dateA.getTime();
            });
          }
        } catch (error: any) {
          logger.error('Error loading replies:', error);
          replies = [];
          
          if (error.message?.includes('not found') || error.message?.includes('404')) {
            errorMessage = `User "${userId}" not found or has deleted their account.`;
          } else if (error.message?.includes('invalid UUID format')) {
            errorMessage = `Invalid user ID format. Please check the profile URL.`;
          } else if (error.message?.includes('timed out')) {
            errorMessage = `Request timed out. The server might be busy, please try again later.`;
          } else {
            errorMessage = `Error loading replies: ${error.message || 'Unknown error'}`;
          }
          
          throw error;
        }
      } else if (tab === 'media') {
        // Load user's media posts
        logger.debug(`Loading media for user ${userId}`);
        try {
          const mediaData = await getUserMedia(userId);
          
          // Safety check for valid data structure
          if (!mediaData || (!mediaData.media && !mediaData.data)) {
            logger.warn('Received invalid media data structure:', mediaData);
            media = [];
            if (!media.length) {
              logger.debug('No media found for user');
            }
          } else {
            // Ensure media items have all required fields
            const mediaArray = mediaData.media || mediaData.data || [];
            media = mediaArray.map(item => ({
              id: item.id || `media-${Math.random().toString(36).substr(2, 9)}`,
              url: item.url || '',
              type: item.type || 'image',
              thread_id: item.thread_id || '',
              created_at: item.created_at || new Date().toISOString()
            }));
            logger.debug(`Loaded ${media.length} media items`);

            // Sort by date (newest first)
            media.sort((a, b) => {
              const dateA = new Date(a.created_at || '');
              const dateB = new Date(b.created_at || '');
              return dateB.getTime() - dateA.getTime();
            });
          }
        } catch (error: any) {
          logger.error('Error loading media:', error);
          media = [];
          
          if (error.message?.includes('not found') || error.message?.includes('404')) {
            errorMessage = `User "${userId}" not found or has deleted their account.`;
          } else if (error.message?.includes('invalid UUID format')) {
            errorMessage = `Invalid user ID format. Please check the profile URL.`;
          } else if (error.message?.includes('timed out')) {
            errorMessage = `Request timed out. The server might be busy, please try again later.`;
          } else {
            errorMessage = `Error loading media: ${error.message || 'Unknown error'}`;
          }
          
          throw error;
        }
      }
      
      retryCount = 0; // Reset retry count on success
      activeTab = tab; // Update the active tab state to match loaded content
      
      if ((tab === 'posts' && !posts.length) || 
          (tab === 'replies' && !replies.length) || 
          (tab === 'media' && !media.length)) {
        logger.debug(`No content found for ${tab} tab`);
      }
      
    } catch (error: any) {
      logger.error(`Error loading ${tab}:`, error);
      logger.error(`Error in ${tab}:`, error);
      
      // Use handleLoadError to get a user-friendly message if we don't already have one
      if (!errorMessage) {
        errorMessage = handleLoadError(error, tab);
      }
      
      toastStore.showToast(errorMessage, 'error');
    } finally {
      isLoading = false;
    }
  }
  
  async function loadProfileData() {
    isLoading = true;
    errorMessage = '';
    
    try {
      logger.debug(`Loading profile data for userId: ${userId}`);
      
      // Validate userId
      if (!userId || userId === 'undefined') {
        throw new Error('Invalid user ID');
      }
      
      // Use a try/catch here to give a more specific error if this call fails
      let response;
      try {
        response = await getUserById(userId);
      } catch (error: any) {
        logger.error(`Error getting user by ID: ${error.message}`);
        if (error.message?.includes('404') || error.message?.includes('not found')) {
          throw new Error(`User "${userId}" not found. The account may have been deleted.`);
        } else if (error.message?.includes('invalid UUID format')) {
          throw new Error(`Invalid user ID format: ${userId}`);
        } else {
          throw error; // Re-throw the original error
        }
      }
      
      if (response && response.user) {
        // Store initial data for comparison
        const initialFollowState = response.user.is_following || false;
        
        profileData = {
          id: response.user.id || '',
          username: response.user.username || '',
          displayName: response.user.display_name || response.user.name || '',
          bio: response.user.bio || '',
          profilePicture: response.user.profile_picture_url || response.user.avatar || '',
          backgroundBanner: response.user.banner_url || response.user.background_banner_url || '',
          followerCount: typeof response.user.follower_count === 'number' ? response.user.follower_count : 0,
          followingCount: typeof response.user.following_count === 'number' ? response.user.following_count : 0,
          joinedDate: response.user.created_at ? new Date(response.user.created_at).toLocaleDateString('en-US', { month: 'long', year: 'numeric' }) : '',
          location: response.user.location || '',
          website: response.user.website || '',
          isPrivate: response.user.is_private || false,
          isFollowing: initialFollowState, 
          isBlocked: response.user.is_blocked || false
        };
        
        logger.debug(`Profile data loaded: ${profileData.displayName} (@${profileData.username})`);
        logger.debug(`Following status: ${initialFollowState}`);
        logger.debug(`Follower count: ${profileData.followerCount}, Following count: ${profileData.followingCount}`);
          
        // Load initial tab content after profile loads
        await loadTabContent('posts');
      } else {
        logger.error('User profile not found in response:', response);
        errorMessage = 'User not found';
        toastStore.showToast('User not found', 'error');
      }
    } catch (error: any) {
      errorMessage = handleLoadError(error, 'profile');
      toastStore.showToast(errorMessage, 'error');
    } finally {
      isLoading = false;
    }
  }
  
  async function handleFollow() {
    if (isFollowLoading) return;
    
    isFollowLoading = true;
    try {
      let success;
      const action = profileData.isFollowing ? 'unfollow' : 'follow';
      const originalFollowState = profileData.isFollowing;
      const originalFollowerCount = profileData.followerCount;
      
      // Optimistic UI update
      profileData = {
        ...profileData,
        isFollowing: !originalFollowState,
        followerCount: originalFollowState 
          ? Math.max(0, originalFollowerCount - 1) 
          : originalFollowerCount + 1
      };
      
      logger.debug(`Attempting to ${action} user ${profileData.id}, previous state: ${originalFollowState}`);
      
      if (originalFollowState) {
        success = await unfollowUser(profileData.id);
        logger.debug(`Unfollow attempt result: ${success}`);
      } else {
        success = await followUser(profileData.id);
        logger.debug(`Follow attempt result: ${success}`);
      }
      
      if (success) {
        toastStore.showToast(
          profileData.isFollowing ? 'You are now following this user' : 'You have unfollowed this user',
          'success'
        );
        
        // If we just followed a private user, reload the content
        if (profileData.isFollowing && profileData.isPrivate) {
          await loadTabContent(activeTab);
        }
      } else {
        // Revert UI state if API call failed
        profileData = {
          ...profileData,
          isFollowing: originalFollowState,
          followerCount: originalFollowerCount
        };
        
        throw new Error(`Failed to ${action} user - API returned failure`);
      }
    } catch (error) {
      logger.error('Error updating follow status:', error);
      toastStore.showToast('Failed to update follow status. Please try again.', 'error');
    } finally {
      isFollowLoading = false;
    }
  }
  
  async function submitReport() {
    if (!reportReason.trim()) {
      toastStore.showToast('Please provide a reason for the report', 'error');
      return;
    }
    
    try {
      const result = await reportUser(profileData.id, reportReason);
      
      if (result) {
        toastStore.showToast('User reported successfully. Our team will review this report.', 'success');
        showReportModal = false;
        reportReason = '';
      } else {
        throw new Error('Failed to report user');
      }
    } catch (error) {
      logger.error('Error reporting user:', error);
      toastStore.showToast('Failed to report user. Please try again.', 'error');
    }
  }
  
  async function handleBlockUser() {
    if (isBlockLoading) return;
    
    isBlockLoading = true;
    try {
      let success;
      const action = profileData.isBlocked ? 'unblock' : 'block';
      
      if (profileData.isBlocked) {
        success = await unblockUser(profileData.id);
      } else {
        success = await blockUser(profileData.id);
      }
      
      if (success) {
        profileData = {
          ...profileData,
          isBlocked: !profileData.isBlocked
        };
        
        toastStore.showToast(
          profileData.isBlocked ? 'User blocked successfully' : 'User unblocked successfully',
          'success'
        );
        
        // Close the modal
        showBlockConfirmModal = false;
        showActionsDropdown = false;
        
        // If we just blocked the user and they were being followed, simulate unfollow
        if (profileData.isBlocked && profileData.isFollowing) {
          profileData = {
            ...profileData,
            isFollowing: false,
            followerCount: Math.max(0, profileData.followerCount - 1)
          };
        }
        
        // Clear content if we blocked the user
        if (profileData.isBlocked) {
          posts = [];
          replies = [];
          media = [];
        }
      } else {
        throw new Error(`Failed to ${action} user`);
      }
    } catch (error) {
      logger.error('Error updating block status:', error);
      toastStore.showToast('Failed to update block status. Please try again.', 'error');
    } finally {
      isBlockLoading = false;
    }
  }
  
  function formatJoinDate(dateString) {
    if (!dateString) return '';
    
    const date = new Date(dateString);
    return `Joined ${date.toLocaleString('default', { month: 'long' })} ${date.getFullYear()}`;
  }
  
  function retryLoad() {
    if (retryCount < MAX_RETRIES) {
      retryCount++;
      if (profileData.id) {
        loadTabContent(activeTab);
      } else {
        loadProfileData();
      }
    } else {
      toastStore.showToast('Maximum retries reached. Please refresh the page.', 'error');
    }
  }
  
  // Close dropdowns when clicking outside
  function handleClickOutside(event) {
    if (showActionsDropdown && !event.target.closest('.dropdown-container')) {
      showActionsDropdown = false;
    }
  }
    onMount(() => {
    // Check if user is authenticated
    if (!isAuthenticated()) {
      logger.warn('User not authenticated, redirecting to login');
      window.location.href = '/login';
      return;
    }
    
    logger.debug('OtherProfile component mounted with userId:', userId);
    
    // Only load profile if we have a userId
    if (userId && userId !== 'undefined') {
      loadProfileData();
    } else {
      logger.error('No userId provided or invalid userId');
      errorMessage = 'Invalid user ID';
      isLoading = false;
      toastStore.showToast('Invalid user profile ID', 'error');
      // Redirect to home after a short delay if ID is invalid
      setTimeout(() => {
        window.location.href = '/';
      }, 2000);
    }
    
    // Add click listener for closing dropdowns
    document.addEventListener('click', handleClickOutside);
    
    // Return cleanup function
    return () => {
      document.removeEventListener('click', handleClickOutside);
      if (retryCount > 0) {
        logger.debug(`Component unmounted with ${retryCount} retry attempts`);
      }
    };
  });
  
  // Helper function for following/follower links
  function handleFollowLink(type: 'following' | 'followers', userId: string) {
    if (!userId) {
      logger.error(`Cannot navigate to ${type} - missing userId`);
      toastStore.showToast('Cannot view followers/following for this user', 'error');
      return;
    }
    
    logger.debug(`Navigating to ${type} page for userId ${userId}`);
    window.location.href = `/${type}/${userId}`;
  }
</script>

<MainLayout>
  <div class="profile-container">
    <!-- Header with back button -->
    <div class="profile-header-container">
      <button class="profile-header-back" on:click={() => window.history.back()}>
        <ArrowLeftIcon size="20" />
      </button>
      
      <div class="profile-banner-container">
        {#if profileData.backgroundBanner}
          <img 
            src={profileData.backgroundBanner} 
            alt="Banner" 
            class="profile-banner"
            on:error={(e) => {
              const target = e.target as HTMLImageElement;
              if (target) {
                target.src = '/images/default-banner.png';
              }
            }}
          />
        {/if}
        <div class="profile-banner-overlay"></div>
      </div>
    </div>
    
    <!-- Profile info section -->
    <div class="profile-avatar-container">
      <div class="profile-avatar-wrapper">
        {#if profileData.profilePicture}
          <img 
            src={profileData.profilePicture} 
            alt="Profile"
            class="profile-avatar"
            on:error={(e) => {
              const target = e.target as HTMLImageElement;
              if (target) {
                target.src = '/images/default-avatar.png';
              }
            }}
          />
        {:else}
          <div class="profile-avatar profile-avatar-placeholder">
            <UserIcon size="48" />
          </div>
        {/if}
      </div>
    </div>
    
    <div class="profile-details">
      {#if !isLoading && currentUserId !== userId}
        <div class="profile-actions">
          <div class="dropdown-container">
            <button 
              class="profile-edit-button"
              on:click={() => showActionsDropdown = !showActionsDropdown}
              title="More actions"
            >
              <MoreHorizontalIcon size="20" />
            </button>
            
            {#if showActionsDropdown}
              <div class="dropdown-menu">
                <button 
                  class="dropdown-item"
                  on:click={() => {
                    showBlockConfirmModal = true;
                    showActionsDropdown = false;
                  }}
                >
                  <ShieldIcon size="16" class="dropdown-icon" />
                  {profileData.isBlocked ? 'Unblock @' + profileData.username : 'Block @' + profileData.username}
                </button>
                
                <button 
                  class="dropdown-item"
                  on:click={() => {
                    showReportModal = true;
                    showActionsDropdown = false;
                  }}
                >
                  <FlagIcon size="16" class="dropdown-icon" />
                  Report @{profileData.username}
                </button>
              </div>
            {/if}
          </div>
          
          <button 
            class={profileData.isFollowing ? 'profile-following-button' : 'profile-follow-button'}
            on:click={handleFollow}
            disabled={isFollowLoading}
          >
            {#if isFollowLoading}
              <span class="loading-spinner"></span>
            {:else}
              {profileData.isFollowing ? 'Following' : 'Follow'}
            {/if}
          </button>
        </div>
      {/if}
      
      <div class="profile-name-container">
        <h1 class="profile-name">{profileData.displayName}</h1>
        <div class="profile-username">@{profileData.username}</div>
      </div>
      
      {#if profileData.isBlocked}
        <div class="profile-blocked-alert">
          <SlashIcon size="16" class="alert-icon" />
          <span>You have blocked this user</span>
        </div>
      {:else if profileData.isPrivate && !profileData.isFollowing && currentUserId !== userId}
        <div class="profile-private-alert">
          <UserIcon size="16" class="alert-icon" />
          <span>This account is private. Follow to see their posts.</span>
        </div>
      {:else if profileData.bio}
        <p class="profile-bio">{profileData.bio}</p>
      {/if}
      
      <div class="profile-meta">
        {#if profileData.location}
          <div class="profile-meta-item">
            <MapPinIcon size="14" />
            <span>{profileData.location}</span>
          </div>
        {/if}
        
        {#if profileData.website}
          <div class="profile-meta-item">
            <LinkIcon size="14" />
            <a 
              href={profileData.website.startsWith('http') ? profileData.website : `https://${profileData.website}`}
              target="_blank" 
              rel="noopener noreferrer"
              class="profile-website"
            >
              {profileData.website.replace(/^https?:\/\/(www\.)?/, '')}
            </a>
          </div>
        {/if}
        
        <div class="profile-meta-item">
          <CalendarIcon size="14" />
          <span>{formatJoinDate(profileData.joinedDate)}</span>
        </div>
      </div>
      
      <div class="profile-stats">
        <button 
          class="profile-stat" 
          on:click={() => handleFollowLink('following', profileData.id)}
        >
          <span class="profile-stat-count">{profileData.followingCount}</span>
          <span>Following</span>
        </button>
        <button 
          class="profile-stat" 
          on:click={() => handleFollowLink('followers', profileData.id)}
        >
          <span class="profile-stat-count">{profileData.followerCount}</span>
          <span>Followers</span>
        </button>
      </div>
    </div>
    
    <!-- Tab Navigation -->
    <div class="profile-tabs">
      <button 
        class="profile-tab {activeTab === 'posts' ? 'active' : ''}"
        on:click={() => setActiveTab('posts')}
      >
        Posts
      </button>
      <button 
        class="profile-tab {activeTab === 'replies' ? 'active' : ''}"
        on:click={() => setActiveTab('replies')}
      >
        Replies
      </button>
      <button 
        class="profile-tab {activeTab === 'media' ? 'active' : ''}"
        on:click={() => setActiveTab('media')}
      >
        Media
      </button>
    </div>
    
    <!-- Tab Content -->
    <div class="profile-content">
      {#if profileData.isBlocked}
        <div class="profile-content-empty">
          <SlashIcon size="48" class="profile-content-empty-icon error" />
          <p class="profile-content-empty-title error">This user is blocked</p>
          <p class="profile-content-empty-text error">
            Unblock this user to see their content
          </p>
        </div>
      {:else if profileData.isPrivate && !profileData.isFollowing && currentUserId !== userId}
        <div class="profile-content-empty">
          <UserIcon size="48" class="profile-content-empty-icon" />
          <p class="profile-content-empty-title">This account is private</p>
          <p class="profile-content-empty-text">
            Follow this user to see their posts, replies, and media
          </p>
          <button 
            class="profile-follow-button"
            on:click={handleFollow}
            disabled={isFollowLoading}
          >
            {isFollowLoading ? 'Processing...' : 'Follow'}
          </button>
        </div>
      {:else if isLoading}
        <LoadingSkeleton type="threads" count={3} />
      {:else if errorMessage}
        <div class="profile-content-empty">
          <AlertCircleIcon size="48" class="profile-content-empty-icon error" />
          <p class="profile-content-empty-text">{errorMessage}</p>
          <button 
            class="profile-follow-button"
            on:click={retryLoad}
          >
            Retry
          </button>
        </div>
      {:else if activeTab === 'posts'}
        {#if posts.length === 0}
          <div class="profile-content-empty">
            <p class="profile-content-empty-title">No posts yet</p>
            <p class="profile-content-empty-text">
              {currentUserId === userId ? "Share your first thought!" : "@" + profileData.username + " hasn't posted yet"}
            </p>
          </div>
        {:else}
          <div class="tweet-feed">
            {#each posts as post (post.id)}
              <div class="tweet-card-container">
                <TweetCard 
                  tweet={ensureTweetFormat(post)} 
                  isDarkMode={isDarkMode} 
                  isAuthenticated={true}
                  isLiked={post.isLiked || post.is_liked}
                  isReposted={post.isReposted || post.is_repost}
                  isBookmarked={post.isBookmarked || post.is_bookmarked}
                  on:reply={handleReply}
                />
              </div>
            {/each}
          </div>
        {/if}
      {:else if activeTab === 'replies'}
        {#if replies.length === 0}
          <div class="profile-content-empty">
            <p class="profile-content-empty-title">No replies yet</p>
            <p class="profile-content-empty-text">
              {currentUserId === userId ? "Join the conversation!" : "This user hasn't replied to any posts yet."}
            </p>
          </div>
        {:else}
          <div class="tweet-feed">
            {#each replies as reply (reply.id)}
              <div class="tweet-card-container">
                <div class="reply-indicator">
                  <span>Replying to</span>
                  <a href={`/thread/${reply.thread_id}`}>thread</a>
                </div>
                <TweetCard 
                  tweet={ensureTweetFormat(reply)} 
                  isDarkMode={isDarkMode} 
                  isAuthenticated={true}
                  isLiked={reply.isLiked || reply.is_liked}
                  isReposted={reply.isReposted || reply.is_repost}
                  isBookmarked={reply.isBookmarked || reply.is_bookmarked}
                  on:reply={handleReply}
                />
              </div>
            {/each}
          </div>
        {/if}
      {:else if activeTab === 'media'}
        {#if media.length === 0}
          <div class="profile-content-empty">
            <p class="profile-content-empty-title">No media yet</p>
            <p class="profile-content-empty-text">
              {currentUserId === userId ? "Share photos, videos, or GIFs!" : "This user hasn't shared any media yet."}
            </p>
          </div>
        {:else}
          <div class="media-grid">
            {#each media as item (item.id)}
              <a 
                href={`/thread/${item.thread_id}`} 
                class="media-grid-item"
              >
                {#if item.type === 'image'}
                  <img 
                    src={item.url} 
                    alt="Media content" 
                    class="media-image" 
                    loading="lazy"
                    on:error={(e) => {
                      const target = e.target as HTMLImageElement;
                      if (target) {
                        target.src = '/images/placeholder.png';
                      }
                    }}
                  />
                {:else if item.type === 'video'}
                  <div class="media-video-container">
                    <video src={item.url} class="media-video">
                      <track kind="captions" label="English" src="" default />
                    </video>
                    <div class="media-video-play-button">
                      <div class="media-video-play-icon"></div>
                    </div>
                  </div>
                {:else if item.type === 'gif'}
                  <div class="media-gif-container">
                    <img 
                      src={item.url} 
                      alt="GIF content" 
                      class="media-image" 
                      loading="lazy"
                      on:error={(e) => {
                        const target = e.target as HTMLImageElement;
                        if (target) {
                          target.src = '/images/placeholder.png';
                        }
                      }}
                    />
                    <div class="media-gif-indicator">GIF</div>
                  </div>
                {/if}
              </a>
            {/each}
          </div>
        {/if}
      {/if}
    </div>
  </div>
  
  <!-- Report Modal -->
  {#if showReportModal}
    <div class="modal-overlay" on:click={() => showReportModal = false}>
      <div class="modal-container" on:click|stopPropagation>
        <div class="modal-header">
          <h2>Report @{profileData.username}</h2>
          <button class="modal-close-button" on:click={() => showReportModal = false}>
            <XIcon size="20" />
          </button>
        </div>
        
        <div class="modal-content">
          <p class="modal-description">Please tell us why you're reporting this account.</p>
          
          <textarea
            bind:value={reportReason}
            class="modal-textarea"
            rows="4"
            placeholder="Describe the issue..."
            maxlength="500"
          ></textarea>
          
          <div class="modal-actions">
            <button class="modal-cancel-button" on:click={() => showReportModal = false}>
              Cancel
            </button>
            <button 
              class="modal-action-button danger"
              on:click={submitReport}
              disabled={!reportReason.trim()}
            >
              Submit Report
            </button>
          </div>
        </div>
      </div>
    </div>
  {/if}
  
  <!-- Block Confirmation Modal -->
  {#if showBlockConfirmModal}
    <div class="modal-overlay" on:click={() => showBlockConfirmModal = false}>
      <div class="modal-container" on:click|stopPropagation>
        <div class="modal-header">
          <h2>{profileData.isBlocked ? 'Unblock' : 'Block'} @{profileData.username}?</h2>
          <button class="modal-close-button" on:click={() => showBlockConfirmModal = false}>
            <XIcon size="20" />
          </button>
        </div>
        
        <div class="modal-content">
          <p class="modal-description">
            {#if profileData.isBlocked}
              You will be able to follow this user and see their posts again.
            {:else}
              They will not be able to follow you or view your posts, and you will not see their posts or notifications.
            {/if}
          </p>
          
          <div class="modal-actions">
            <button class="modal-cancel-button" on:click={() => showBlockConfirmModal = false}>
              Cancel
            </button>
            <button 
              class="modal-action-button danger"
              on:click={handleBlockUser}
              disabled={isBlockLoading}
            >
              {#if isBlockLoading}
                <span class="loading-spinner"></span>
                <span>Processing...</span>
              {:else}
                {profileData.isBlocked ? 'Unblock' : 'Block'}
              {/if}
            </button>
          </div>
        </div>
      </div>
    </div>
  {/if}
</MainLayout>

<style>
  /* Add any component-specific styles here */
  .dropdown-container {
    position: relative;
  }
  
  .dropdown-menu {
    position: absolute;
    right: 0;
    top: 100%;
    margin-top: 4px;
    background: var(--bg-color);
    border: 1px solid var(--border-color);
    border-radius: var(--radius-md);
    box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
    z-index: 10;
    overflow: hidden;
  }
  
  .dropdown-item {
    display: flex;
    align-items: center;
    width: 100%;
    padding: 12px 16px;
    text-align: left;
    background: none;
    border: none;
    color: var(--text-primary);
    font-size: var(--font-size-sm);
    cursor: pointer;
  }
  
  .dropdown-item:hover {
    background-color: var(--bg-hover);
  }
  
  .dropdown-icon {
    margin-right: 12px;
  }
  
  .profile-blocked-alert,
  .profile-private-alert {
    display: flex;
    align-items: center;
    padding: 12px;
    margin: 12px 0;
    border-radius: var(--radius-md);
  }
  
  .profile-blocked-alert {
    background-color: rgba(244, 33, 46, 0.1);
    color: #e0245e;
  }
  
  .profile-private-alert {
    background-color: var(--bg-secondary);
    color: var(--text-primary);
  }
  
  .alert-icon {
    margin-right: 8px;
  }
  
  .loading-spinner {
    display: inline-block;
    width: 16px;
    height: 16px;
    border: 2px solid transparent;
    border-top-color: currentColor;
    border-radius: 50%;
    animation: spin 1s linear infinite;
    margin-right: 4px;
  }
  
  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }
  
  .profile-website {
    color: var(--color-primary);
    text-decoration: none;
  }
  
  .profile-website:hover {
    text-decoration: underline;
  }
  
  /* Error styling */
  .error {
    color: #e0245e;
  }
  
  /* Make the profile-stat buttons look like links */
  button.profile-stat {
    background: none;
    border: none;
    padding: 0;
    cursor: pointer;
    font-family: inherit;
    font-size: inherit;
    color: inherit;
    text-align: left;
  }
  
  button.profile-stat:hover {
    text-decoration: underline;
  }
</style>
