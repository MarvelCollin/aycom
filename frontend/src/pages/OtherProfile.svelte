<script lang="ts">
  import { onMount } from 'svelte';
  import MainLayout from '../components/layout/MainLayout.svelte';
  import { useAuth } from '../hooks/useAuth';
  import { useTheme } from '../hooks/useTheme';
  import { isAuthenticated, getUserId } from '../utils/auth';
  import { getUserById, followUser, unfollowUser, reportUser, blockUser, unblockUser, checkFollowStatus, getUserFollowers, getUserFollowing } from '../api/user';
  import type { FollowUserResponse, UnfollowUserResponse } from '../api/user';
  import { getUserThreads, getUserReplies, getUserMedia } from '../api/thread';
  import { toastStore } from '../stores/toastStore';
  import TweetCard from '../components/social/TweetCard.svelte';
  import LoadingSkeleton from '../components/common/LoadingSkeleton.svelte';
  import type { ITweet } from '../interfaces/ISocialMedia';
  import { createLoggerWithPrefix } from '../utils/logger';
  
  const logger = createLoggerWithPrefix('OtherProfile');

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
  
  interface Thread {
    id: string;
    content: string;
    username: string;
    display_name: string;
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
    [key: string]: any;
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
    [key: string]: any;
  }
  
  interface ThreadMedia {
    id: string;
    url: string;
    type: 'image' | 'video' | 'gif';
    thread_id: string;
    created_at?: string;
    [key: string]: any; 
  }
  
  const { getAuthState } = useAuth();
  const { theme } = useTheme();
  
  export let userId: string;
  
  $: isDarkMode = $theme === 'dark';
  $: authState = getAuthState();
  $: currentUserId = getUserId();
  
  $: {
    logger.debug(`Profile rendering for userId: ${userId}, currentUserId: ${currentUserId}`);
  }
  
  // Define proper interface for profile data with snake_case
  interface ProfileData {
    id: string;
    username: string;
    name: string; // was displayName
    bio: string;
    profile_picture_url: string; // was profilePicture
    banner_url: string; // was backgroundBanner
    follower_count: number; // was followerCount
    following_count: number; // was followingCount
    created_at: string; // was joinedDate
    location: string;
    website: string;
    is_private: boolean; // was isPrivate
    is_following: boolean; // was isFollowing
    is_blocked: boolean; // was isBlocked
  }
  
  // Initial profile data using the new interface
  let profileData: ProfileData = {
    id: '',
    username: '',
    name: '',
    bio: '',
    profile_picture_url: '',
    banner_url: '',
    follower_count: 0,
    following_count: 0,
    created_at: '',
    location: '',
    website: '',
    is_private: false,
    is_following: false,
    is_blocked: false
  };
  
  let posts: Thread[] = [];
  let replies: Reply[] = [];
  let media: ThreadMedia[] = [];
  
  let activeTab = 'posts';
  let isLoading = true;
  let isFollowLoading = false;
  let isBlockLoading = false;
  let showReportModal = false;
  let reportReason = '';
  let showBlockConfirmModal = false;
  let showActionsDropdown = false;
  let showEditProfile = false;
  let errorMessage = '';
  let retryCount = 0;
  const MAX_RETRIES = 3;
  let isFollowRequestPending = false; 
  
  // Add state for followers/following modals
  let showFollowersModal = false;
  let showFollowingModal = false;
  let followersList = [];
  let followingList = [];
  let isLoadingFollowers = false;
  let isLoadingFollowing = false;
  let followersError = '';
  let followingError = '';
  
  function ensureTweetFormat(thread: any): ITweet {
    try {
      if (!thread || typeof thread !== 'object') {
        logger.warn('Invalid thread object provided to ensureTweetFormat');
        return {
          id: `invalid-${Math.random().toString(36).substring(2, 9)}`,
          thread_id: '',
          user_id: '',
          username: 'unknown',
          name: 'Unknown User',
          content: 'This content is unavailable',
          created_at: new Date().toISOString(),
          profile_picture_url: '',
          likes_count: 0,
          replies_count: 0,
          reposts_count: 0,
          bookmarks_count: 0,
          views_count: 0,
          media: [],
          is_liked: false,
          is_reposted: false,
          is_bookmarked: false,
          is_pinned: false,
          reply_to: null
        };
      }
    
      let username = thread.author_username || thread.authorUsername || thread.username || 'anonymous';
      
      let name = thread.author_name || thread.authorName || thread.display_name || 
                 thread.displayName || thread.name || username || 'User';
      
      let profile_picture_url = thread.author_avatar || thread.authorAvatar || 
                          thread.profile_picture_url || thread.profilePictureUrl || 
                          thread.avatar || 'https://secure.gravatar.com/avatar/0?d=mp';
      
      let created_at = thread.created_at || thread.createdAt || thread.timestamp || new Date().toISOString();
      if (typeof created_at === 'string' && !created_at.includes('T')) {
        created_at = new Date(created_at).toISOString();
      }
      
      const likes_count = Number(thread.likes_count || thread.like_count || thread.metrics?.likes || 0);
      const replies_count = Number(thread.replies_count || thread.reply_count || thread.metrics?.replies || 0);
      const reposts_count = Number(thread.reposts_count || thread.repost_count || thread.metrics?.reposts || 0);
      const bookmarks_count = Number(thread.bookmarks_count || thread.bookmark_count || thread.metrics?.bookmarks || 0);
      const views_count = Number(thread.views || thread.views_count || 0);
      
      const is_liked = Boolean(thread.is_liked || thread.isLiked || false);
      const is_reposted = Boolean(thread.is_repost || thread.isReposted || false);
      const is_bookmarked = Boolean(thread.is_bookmarked || thread.isBookmarked || false);
      const is_pinned = Boolean(
        thread.is_pinned === true || 
        thread.is_pinned === 'true' || 
        thread.is_pinned === 1 || 
        thread.is_pinned === '1' || 
        thread.is_pinned === 't' || 
        thread.IsPinned === true || 
        false
      );
      
      const media = Array.isArray(thread.media) ? thread.media : [];
        
      const id = thread.id || `thread-${Math.random().toString(36).substring(2, 9)}`;
      const user_id = thread.user_id || thread.userId || thread.author_id || thread.authorId || '';
        
      return {
        id,
        thread_id: thread.thread_id || id,
        user_id,
        username,
        name,
        content: thread.content || '',
        created_at: typeof created_at === 'string' ? created_at : new Date(created_at).toISOString(),
        profile_picture_url,
        likes_count,
        replies_count,
        reposts_count,
        bookmarks_count,
        views_count,
        media,
        is_liked,
        is_reposted,
        is_bookmarked,
        is_pinned,
        reply_to: thread.parent_id ? { id: thread.parent_id } as any : null
      };
    } catch (error: any) {
      logger.error('Error formatting tweet:', error);
      return {
        id: `error-${Math.random().toString(36).substring(2, 9)}`,
        thread_id: '',
        user_id: '',
        username: 'error',
        name: 'Error',
        content: 'Error loading tweet',
        created_at: new Date().toISOString(),
        profile_picture_url: '',
        likes_count: 0,
        replies_count: 0,
        reposts_count: 0,
        bookmarks_count: 0,
        views_count: 0,
        media: [],
        is_liked: false,
        is_reposted: false,
        is_bookmarked: false,
        is_pinned: false,
        reply_to: null
      };
    }
  }

  function handleReply(event) {
    const thread_id = event.detail;
    window.location.href = `/thread/${thread_id}`;
  }

  function setActiveTab(tab) {
    activeTab = tab;
    loadTabContent(tab);
  }
  
  function handleLoadError(error: any, context: string): string {
    logger.error(`Error in ${context}:`, error);
    
    const errorMessage = error?.message || String(error);
    
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
    
    return `Failed to load ${context}. Please try again later.`;
  }
  
  async function loadTabContent(tab: string) {
    if (isLoading) return;
    
    isLoading = true;
    errorMessage = '';
    
    try {
      if (profileData.is_private && !profileData.is_following && currentUserId !== userId) {
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
              const dateA = new Date(a.created_at);
              const dateB = new Date(b.created_at);
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
              const dateA = new Date(a.created_at);
              const dateB = new Date(b.created_at);
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
    retryCount = 0;
    
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
        logger.debug(`Raw API response for getUserById: ${JSON.stringify(response)}`);
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
        // Make sure we get the accurate follow status from the API
        const initialFollowState = response.user.is_following === true;
        
        logger.debug(`Initial follow state from API: ${initialFollowState}`);
        logger.debug(`Follow state data type: ${typeof response.user.is_following}`);
        logger.debug(`Raw follow state value: ${response.user.is_following}`);
        
        profileData = {
          id: response.user.id || '',
          username: response.user.username || '',
          name: response.user.display_name || response.user.name || '',
          bio: response.user.bio || '',
          profile_picture_url: response.user.profile_picture_url || response.user.avatar || '',
          banner_url: response.user.banner_url || response.user.background_banner_url || '',
          follower_count: typeof response.user.follower_count === 'number' ? response.user.follower_count : 0,
          following_count: typeof response.user.following_count === 'number' ? response.user.following_count : 0,
          created_at: response.user.created_at ? new Date(response.user.created_at).toLocaleDateString('en-US', { month: 'long', year: 'numeric' }) : '',
          location: response.user.location || '',
          website: response.user.website || '',
          is_private: response.user.is_private || false,
          is_following: initialFollowState, 
          is_blocked: response.user.is_blocked || false
        };
        
        logger.debug(`Profile data loaded: ${profileData.name} (@${profileData.username})`);
        logger.debug(`Following status in profileData: ${profileData.is_following}`);
        logger.debug(`Follower count: ${profileData.follower_count}, Following count: ${profileData.following_count}`);
        
        // Make an explicit check for follow status to ensure it's correct
        try {
          if (currentUserId && currentUserId !== userId) {
            logger.debug(`Making explicit follow status check for ${userId}`);
            try {
              const isActuallyFollowing = await checkFollowStatus(userId);
              logger.debug(`Explicit follow check result: ${isActuallyFollowing}`);
              
              // ALWAYS update the profile data with the latest follow status
              logger.debug(`Setting follow status to: ${isActuallyFollowing}`);
              profileData = {
                ...profileData,
                is_following: isActuallyFollowing
              };
              
              // Force UI update by logging the new state
              logger.debug(`Updated follow status in profileData: ${profileData.is_following}`);
            } catch (followCheckError) {
              // Don't throw or change the UI state - keep using the value from getUserById
              logger.error("Error during follow status check:", followCheckError);
              logger.debug("Continuing with initial follow status value from profile data");
            }
          }
        } catch (outerError) {
          // Catch any unexpected errors in the outer block but don't disrupt the UI
          logger.error("Unexpected error in follow status check block:", outerError);
        }
          
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
    // Handle follow/unfollow
  async function toggleFollow() {
    if (!currentUserId || currentUserId === userId) {
      return; // Can't follow yourself
    }
    
    if (isFollowRequestPending) {
      logger.debug('Follow request already in progress - ignoring duplicate request');
      return; // Prevent multiple simultaneous requests
    }
    
    isFollowRequestPending = true;
    // Store the current following state to use throughout the function
    const wasFollowing = profileData.is_following;
    
    try {
      // Log current state before changes
      logger.debug(`Toggle follow: current state is_following=${wasFollowing}, follower_count=${profileData.follower_count}`);
      logger.debug(`Target userId=${userId}, currentUserId=${currentUserId}`);
      
      // Optimistic update - change UI immediately (Twitter-like behavior)
      profileData = {
        ...profileData,
        is_following: !wasFollowing,
        follower_count: wasFollowing 
          ? Math.max(0, (profileData.follower_count || 0) - 1) 
          : (profileData.follower_count || 0) + 1
      };
      
      logger.debug(`${wasFollowing ? 'Unfollowing' : 'Following'} user ${userId}`);
      logger.debug(`Updated UI state: is_following=${profileData.is_following}`);
      
      // Make API call
      let apiResponse;
      try {
        apiResponse = wasFollowing 
          ? await unfollowUser(userId)
          : await followUser(userId);
          
        logger.debug(`API call response:`, apiResponse);
      } catch (apiError) {
        logger.error(`API call threw an exception:`, apiError);
        throw apiError;
      }
      
      if (!apiResponse || !apiResponse.success) {
        // Revert optimistic update on failure
        profileData = {
          ...profileData,
          is_following: wasFollowing,
          follower_count: wasFollowing 
            ? (profileData.follower_count || 0) + 1
            : Math.max(0, (profileData.follower_count || 0) - 1)
        };
        
        const errorMessage = apiResponse?.message || `Failed to ${wasFollowing ? 'unfollow' : 'follow'} user. Please try again.`;
        toastStore.showToast(errorMessage, 'error');
        logger.error(`Failed to ${wasFollowing ? 'unfollow' : 'follow'} user ${userId}: ${apiResponse?.message}`);
        logger.debug(`Reverted UI state: is_following=${profileData.is_following}`);
      } else {
        // Use enhanced response data to update UI more accurately based on actual server state
        let actualNewState: boolean;
        
        if (wasFollowing) {
          // When unfollowing, check is_now_following from the response
          actualNewState = apiResponse.is_now_following === false;
          
          if (apiResponse.was_following === false) {
            profileData = { ...profileData, is_following: false };
            // Don't show a toast when we confirm the unfollow (Twitter doesn't)
          } else {
            profileData = { ...profileData, is_following: false };
            // Don't show a toast when we confirm the unfollow (Twitter doesn't)
          }
        } else {
          // When following, check is_now_following from the response
          actualNewState = apiResponse.is_now_following === true;
          
          if (apiResponse.was_already_following === true) {
            profileData = { ...profileData, is_following: true };
            // Don't show a toast for follow action (Twitter doesn't)
          } else {
            profileData = { ...profileData, is_following: true };
            // Don't show a toast for follow action (Twitter doesn't)
          }
        }
        
        // Ensure profile data reflects the actual state from the server
        profileData = {
          ...profileData,
          is_following: actualNewState
        };
        
        logger.debug(`Successfully ${wasFollowing ? 'unfollowed' : 'followed'} user. Response:`, apiResponse);
        logger.debug(`Final UI state after API response: is_following=${profileData.is_following}`);
      }
    } catch (error: any) {
      // Revert optimistic update and show error
      profileData = {
        ...profileData,
        is_following: wasFollowing,
        follower_count: wasFollowing 
          ? (profileData.follower_count || 0) + 1
          : Math.max(0, (profileData.follower_count || 0) - 1)
      };
      
      let errorMessage = 'Failed to update follow status';
      if (error?.message) {
        errorMessage = `Error: ${error.message}`;
        
        // Add more specific messages for common errors
        if (error.message.includes('timeout')) {
          errorMessage = 'Request timed out. The server might be busy, please try again later.';
        } else if (error.message.includes('network')) {
          errorMessage = 'Network error. Please check your connection and try again.';
        }
      }
      
      toastStore.showToast(errorMessage, 'error');
      logger.error(`Error toggling follow status: ${error.message || 'Unknown error'}`);
      logger.debug(`Reverted UI state after error: is_following=${profileData.is_following}`);
    } finally {
      isFollowRequestPending = false;
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
    } catch (error: any) {
      logger.error('Error reporting user:', error);
      toastStore.showToast('Failed to report user. Please try again.', 'error');
    }
  }
  
  async function handleBlockUser() {
    if (isBlockLoading) return;
    
    isBlockLoading = true;
    try {
      let success;
      const action = profileData.is_blocked ? 'unblock' : 'block';
      
      if (profileData.is_blocked) {
        success = await unblockUser(profileData.id);
      } else {
        success = await blockUser(profileData.id);
      }
      
      if (success) {
        profileData = {
          ...profileData,
          is_blocked: !profileData.is_blocked
        };
        
        toastStore.showToast(
          profileData.is_blocked ? 'User blocked successfully' : 'User unblocked successfully',
          'success'
        );
        
        // Close the modal
        showBlockConfirmModal = false;
        showActionsDropdown = false;
        
        // If we just blocked the user and they were being followed, simulate unfollow
        if (profileData.is_blocked && profileData.is_following) {
          profileData = {
            ...profileData,
            is_following: false,
            follower_count: Math.max(0, profileData.follower_count - 1)
          };
        }
        
        // Clear content if we blocked the user
        if (profileData.is_blocked) {
          posts = [];
          replies = [];
          media = [];
        }
      } else {
        throw new Error(`Failed to ${action} user`);
      }
    } catch (error: any) {
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
    // Note: No dropdown functionality currently implemented
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
  
  // Replace the handleFollowLink function with this:
  async function handleFollowLink(type: 'following' | 'followers', userId: string) {
    if (!userId) {
      logger.error(`Cannot view ${type} - missing userId`);
      toastStore.showToast('Cannot view followers/following for this user', 'error');
      return;
    }
    
    logger.debug(`Loading ${type} for userId ${userId}`);
    
    if (type === 'followers') {
      showFollowersModal = true;
      await loadFollowers(userId);
    } else {
      showFollowingModal = true;
      await loadFollowing(userId);
    }
  }
  
  // Add functions to load followers and following
  async function loadFollowers(userId: string) {
    isLoadingFollowers = true;
    followersList = [];
    followersError = '';
    
    try {
      const response = await getUserFollowers(userId);
      console.log('Full followers response:', response);
      
      // Handle the nested structure correctly
      if (response && response.data && response.data.followers) {
        followersList = response.data.followers;
        logger.debug(`Loaded ${followersList.length} followers`);
      } else if (response && response.followers) {
        // Handle the non-nested structure as a fallback
        followersList = response.followers;
        logger.debug(`Loaded ${followersList.length} followers (direct structure)`);
      } else {
        throw new Error('Invalid response format');
      }
    } catch (error) {
      logger.error(`Error loading followers: ${error.message}`);
      followersError = `Failed to load followers: ${error.message}`;
      toastStore.showToast('Failed to load followers', 'error');
    } finally {
      isLoadingFollowers = false;
    }
  }
  
  async function loadFollowing(userId: string) {
    isLoadingFollowing = true;
    followingList = [];
    followingError = '';
    
    try {
      const response = await getUserFollowing(userId);
      console.log('Full following response:', response);
      
      // Handle the nested structure correctly
      if (response && response.data && response.data.following) {
        followingList = response.data.following;
        logger.debug(`Loaded ${followingList.length} following`);
      } else if (response && response.following) {
        // Handle the non-nested structure as a fallback
        followingList = response.following;
        logger.debug(`Loaded ${followingList.length} following (direct structure)`);
      } else {
        throw new Error('Invalid response format');
      }
    } catch (error) {
      logger.error(`Error loading following: ${error.message}`);
      followingError = `Failed to load following: ${error.message}`;
      toastStore.showToast('Failed to load following', 'error');
    } finally {
      isLoadingFollowing = false;
    }
  }
  
  // Function to navigate to a user profile
  function navigateToProfile(userId: string) {
    if (userId) {
      window.location.href = `/profile/${userId}`;
    }
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
        {#if profileData.banner_url}
          <img 
            src={profileData.banner_url} 
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
        {#if profileData.profile_picture_url}
          <img 
            src={profileData.profile_picture_url} 
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
      <!-- Profile header buttons -->
      <div class="profile-actions">
        {#if !errorMessage && profileData}
          {#if profileData.id !== currentUserId}
            <div class="profile-action-buttons">
              <!-- Add visible debug indicator -->
              <span class="debug-indicator" style="display:none;">Status: {profileData.is_following ? 'Following' : 'Not Following'}</span>
              <button 
                class={profileData.is_following ? 'profile-following-button' : 'profile-follow-button'}
                on:click={toggleFollow}
                disabled={isFollowRequestPending}
                data-following={profileData.is_following ? 'true' : 'false'}
                aria-label={profileData.is_following ? 'Unfollow' : 'Follow'}
              >
                {#if isFollowRequestPending}
                  <span class="loading-indicator"></span>
                {:else if profileData.is_following}
                  <span class="following-text">Following</span>
                  <span class="unfollow-text">Unfollow</span>
                {:else}
                  Follow
                {/if}
              </button>
            </div>
          {:else}
            <div class="profile-action-buttons">
              <button class="profile-edit-button" on:click={() => showEditProfile = true}>
                Edit profile
              </button>
            </div>
          {/if}
        {/if}
      </div>
      
      <div class="profile-name-container">
        <h1 class="profile-name">{profileData.name}</h1>
        <div class="profile-username">@{profileData.username}</div>
      </div>
        {#if profileData.is_blocked}
        <div class="profile-blocked-alert">
          <SlashIcon size="16" class="profile-alert-icon" />
          <span>You have blocked this user</span>
        </div>
      {:else if profileData.is_private && !profileData.is_following && currentUserId !== userId}
        <div class="profile-private-alert">
          <UserIcon size="16" class="profile-alert-icon" />
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
          <span>{formatJoinDate(profileData.created_at)}</span>
        </div>
      </div>
      
      <div class="profile-stats">
        <button 
          class="profile-stat" 
          on:click={() => handleFollowLink('following', profileData.id)}
        >
          <span class="profile-stat-count">{profileData.following_count}</span>
          <span>Following</span>
        </button>
        <button 
          class="profile-stat" 
          on:click={() => handleFollowLink('followers', profileData.id)}
        >
          <span class="profile-stat-count">{profileData.follower_count}</span>
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
      {#if profileData.is_blocked}
        <div class="profile-content-empty">
          <SlashIcon size="48" class="profile-content-empty-icon error" />
          <p class="profile-content-empty-title error">This user is blocked</p>
          <p class="profile-content-empty-text error">
            Unblock this user to see their content
          </p>
        </div>
      {:else if profileData.is_private && !profileData.is_following && currentUserId !== userId}
        <div class="profile-content-empty">
          <UserIcon size="48" class="profile-content-empty-icon" />
          <p class="profile-content-empty-title">This account is private</p>
          <p class="profile-content-empty-text">
            Follow this user to see their posts, replies, and media
          </p>
          <button 
            class="profile-follow-button"
            on:click={toggleFollow}
            disabled={isFollowRequestPending}
          >
            {isFollowRequestPending ? 'Processing...' : 'Follow'}
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
                  isLiked={post.is_liked}
                  isReposted={post.is_repost}
                  isBookmarked={post.is_bookmarked}
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
                  isLiked={reply.is_liked}
                  isReposted={reply.is_repost}
                  isBookmarked={reply.is_bookmarked}
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
          <h2>{profileData.is_blocked ? 'Unblock' : 'Block'} @{profileData.username}?</h2>
          <button class="modal-close-button" on:click={() => showBlockConfirmModal = false}>
            <XIcon size="20" />
          </button>
        </div>
        
        <div class="modal-content">
          <p class="modal-description">
            {#if profileData.is_blocked}
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
                {profileData.is_blocked ? 'Unblock' : 'Block'}
              {/if}
            </button>
          </div>
        </div>
      </div>
    </div>
  {/if}

  <!-- Followers Modal -->
  {#if showFollowersModal}
    <div class="modal-overlay" on:click={() => showFollowersModal = false}>
      <div class="modal-container" on:click|stopPropagation>
        <div class="modal-header">
          <h2>Followers</h2>
          <button class="modal-close-button" on:click={() => showFollowersModal = false}>
            <XIcon size="20" />
          </button>
        </div>
        
        <div class="modal-content">
          {#if isLoadingFollowers}
            <div class="modal-loading">
              <div class="loading-spinner"></div>
              <p>Loading followers...</p>
            </div>
          {:else if followersError}
            <div class="modal-error">
              <AlertCircleIcon size="24" />
              <p>{followersError}</p>
              <button class="modal-retry-button" on:click={() => loadFollowers(profileData.id)}>
                Retry
              </button>
            </div>
          {:else if followersList.length === 0}
            <div class="modal-empty">
              <p>@{profileData.username} doesn't have any followers yet.</p>
            </div>
          {:else}
            <div class="user-list">
              {#each followersList as follower}
                <div class="user-item" on:click={() => navigateToProfile(follower.id)}>
                  <div class="user-avatar">
                    <img 
                      src={follower.profile_picture_url || follower.avatar || '/images/default-avatar.png'} 
                      alt={follower.username}
                      on:error={(e) => {
                        const target = e.target as HTMLImageElement;
                        if (target) {
                          target.src = '/images/default-avatar.png';
                        }
                      }}
                    />
                  </div>
                  <div class="user-info">
                    <div class="user-name">{follower.display_name || follower.name}</div>
                    <div class="user-username">@{follower.username}</div>
                    {#if follower.bio}
                      <div class="user-bio">{follower.bio}</div>
                    {/if}
                  </div>
                  {#if follower.id !== currentUserId}
                    <button 
                      class={follower.is_following ? 'profile-following-button compact' : 'profile-follow-button compact'}
                      on:click|stopPropagation={async (e) => {
                        e.stopPropagation();
                        try {
                          if (follower.is_following) {
                            await unfollowUser(follower.id);
                            followersList = followersList.map(f => 
                              f.id === follower.id ? {...f, is_following: false} : f
                            );
                          } else {
                            await followUser(follower.id);
                            followersList = followersList.map(f => 
                              f.id === follower.id ? {...f, is_following: true} : f
                            );
                          }
                        } catch (error) {
                          logger.error(`Error toggling follow for user ${follower.id}:`, error);
                          toastStore.showToast('Failed to update follow status', 'error');
                        }
                      }}
                    >
                      {#if follower.is_following}
                        <span class="following-text">Following</span>
                        <span class="unfollow-text">Unfollow</span>
                      {:else}
                        Follow
                      {/if}
                    </button>
                  {/if}
                </div>
              {/each}
            </div>
          {/if}
        </div>
      </div>
    </div>
  {/if}

  <!-- Following Modal -->
  {#if showFollowingModal}
    <div class="modal-overlay" on:click={() => showFollowingModal = false}>
      <div class="modal-container" on:click|stopPropagation>
        <div class="modal-header">
          <h2>Following</h2>
          <button class="modal-close-button" on:click={() => showFollowingModal = false}>
            <XIcon size="20" />
          </button>
        </div>
        
        <div class="modal-content">
          {#if isLoadingFollowing}
            <div class="modal-loading">
              <div class="loading-spinner"></div>
              <p>Loading following...</p>
            </div>
          {:else if followingError}
            <div class="modal-error">
              <AlertCircleIcon size="24" />
              <p>{followingError}</p>
              <button class="modal-retry-button" on:click={() => loadFollowing(profileData.id)}>
                Retry
              </button>
            </div>
          {:else if followingList.length === 0}
            <div class="modal-empty">
              <p>@{profileData.username} isn't following anyone yet.</p>
            </div>
          {:else}
            <div class="user-list">
              {#each followingList as user}
                <div class="user-item" on:click={() => navigateToProfile(user.id)}>
                  <div class="user-avatar">
                    <img 
                      src={user.profile_picture_url || user.avatar || '/images/default-avatar.png'} 
                      alt={user.username}
                      on:error={(e) => {
                        const target = e.target as HTMLImageElement;
                        if (target) {
                          target.src = '/images/default-avatar.png';
                        }
                      }}
                    />
                  </div>
                  <div class="user-info">
                    <div class="user-name">{user.display_name || user.name}</div>
                    <div class="user-username">@{user.username}</div>
                    {#if user.bio}
                      <div class="user-bio">{user.bio}</div>
                    {/if}
                  </div>
                  {#if user.id !== currentUserId}
                    <button 
                      class={user.is_following ? 'profile-following-button compact' : 'profile-follow-button compact'}
                      on:click|stopPropagation={async (e) => {
                        e.stopPropagation();
                        try {
                          if (user.is_following) {
                            await unfollowUser(user.id);
                            followingList = followingList.map(f => 
                              f.id === user.id ? {...f, is_following: false} : f
                            );
                          } else {
                            await followUser(user.id);
                            followingList = followingList.map(f => 
                              f.id === user.id ? {...f, is_following: true} : f
                            );
                          }
                        } catch (error) {
                          logger.error(`Error toggling follow for user ${user.id}:`, error);
                          toastStore.showToast('Failed to update follow status', 'error');
                        }
                      }}
                    >
                      {#if user.is_following}
                        <span class="following-text">Following</span>
                        <span class="unfollow-text">Unfollow</span>
                      {:else}
                        Follow
                      {/if}
                    </button>
                  {/if}
                </div>
              {/each}
            </div>
          {/if}
        </div>
      </div>
    </div>
  {/if}
</MainLayout>

<style>
  /* Component variables */
  :root {
    --bg-color: #ffffff;
    --text-primary: #0f1419;
    --text-secondary: #536471;
    --border-color: #eff3f4;
    --bg-hover: rgba(0, 0, 0, 0.03);
    --bg-highlight: #f7f9fa;
    --accent-light: #f7f9fa;
  }

  :global(.dark-theme) {
    --bg-color: #000000;
    --bg-highlight: #080808;
    --text-primary: #e7e9ea;
    --text-secondary: #71767b;
    --border-color: #2f3336;
    --bg-hover: rgba(255, 255, 255, 0.03);
    --accent-light: #1e2328;
  }

  /* Profile container styling */
  .profile-container {
    width: 100%;
    max-width: 100%;
    margin: 0;
    position: relative;
    background-color: var(--bg-color);
    min-height: 100vh;
    border-left: 1px solid var(--border-color);
    border-right: 1px solid var(--border-color);
  }

  /* Profile header styling */
  .profile-header-container {
    position: relative;
    width: 100%;
    height: 200px;
    overflow: hidden;
  }

  .profile-header-back {
    position: absolute;
    top: 12px;
    left: 12px;
    z-index: 2;
    display: flex;
    align-items: center;
    justify-content: center;
    width: 36px;
    height: 36px;
    border-radius: 50%;
    background-color: rgba(0, 0, 0, 0.5);
    backdrop-filter: blur(8px);
    -webkit-backdrop-filter: blur(8px);
    color: white;
    cursor: pointer;
    transition: transform 0.2s, background-color 0.2s;
    border: 1px solid rgba(255, 255, 255, 0.1);
  }

  .profile-header-back:hover {
    transform: scale(1.05);
    background-color: rgba(0, 0, 0, 0.7);
  }

  .profile-banner-container {
    position: relative;
    width: 100%;
    height: 100%;
  }

  .profile-banner {
    width: 100%;
    height: 100%;
    object-fit: cover;
    background-color: #1da1f2;
  }

  .profile-banner-overlay {
    position: absolute;
    bottom: 0;
    left: 0;
    right: 0;
    height: 80px;
    background: linear-gradient(to top, rgba(0,0,0,0.5), transparent);
    pointer-events: none;
  }

  /* Avatar styling */
  .profile-avatar-container {
    position: relative;
    margin-top: -72px;
    margin-left: 16px;
    z-index: 1;
    margin-bottom: 12px;
  }

  .profile-avatar-wrapper {
    width: 132px;
    height: 132px;
    border-radius: 50%;
    border: 4px solid var(--bg-color);
    overflow: hidden;
    display: flex;
    align-items: center;
    justify-content: center;
    background-color: #222;
    cursor: pointer;
    padding: 0;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  }

  .profile-avatar {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  .profile-avatar-placeholder {
    display: flex;
    align-items: center;
    justify-content: center;
    background-color: var(--bg-secondary);
    color: var(--text-primary);
    font-size: 48px;
    font-weight: bold;
    width: 100%;
    height: 100%;
  }

  /* Profile details */
  .profile-details {
    padding: 4px 16px;
  }

  .profile-actions {
    display: flex;
    justify-content: flex-end;
    gap: 8px;
    margin-bottom: 12px;
  }

  .profile-action-buttons {
    display: flex;
    gap: 8px;
  }

  .profile-edit-button {
    padding: 6px 16px;
    border-radius: 20px;
    font-weight: 600;
    font-size: 14px;
    border: 1px solid #536471;
    background-color: transparent;
    color: var(--text-primary);
    cursor: pointer;
    transition: background-color 0.2s;
  }

  .profile-edit-button:hover {
    background-color: rgba(0, 0, 0, 0.05);
  }

  .profile-follow-button {
    padding: 6px 16px;
    border-radius: 20px;
    font-weight: bold;
    font-size: 14px;
    background-color: #000000;
    color: #ffffff;
    border: none;
    cursor: pointer;
    transition: background-color 0.15s ease;
    min-width: 102px; /* Match width with following button */
    text-align: center;
  }

  .profile-follow-button:hover {
    background-color: #272c30;
  }
  
  /* Following button - exactly match Twitter's styling */
  .profile-following-button {
    padding: 6px 16px;
    border-radius: 20px;
    font-weight: bold;
    font-size: 14px;
    background-color: #ffffff;
    color: #0f1419;
    border: 1px solid #cfd9de;
    cursor: pointer;
    transition: all 0.15s ease;
    position: relative;
    min-width: 102px; /* Prevent button width from changing between states */
    text-align: center;
  }

  .profile-following-button:hover {
    background-color: rgba(244, 33, 46, 0.1);
    color: #f4212e; /* Twitter's exact red color */
    border-color: rgba(244, 33, 46, 0.4);
  }
  
  /* Control text visibility for hover state */
  .profile-following-button .following-text {
    display: inline;
  }
  
  .profile-following-button .unfollow-text {
    display: none;
  }
  
  .profile-following-button:hover .following-text {
    display: none;
  }
  
  .profile-following-button:hover .unfollow-text {
    display: inline;
  }

  .profile-name-container {
    margin-bottom: 0;
  }

  .profile-name {
    font-size: 20px;
    font-weight: 700;
    line-height: 24px;
    margin: 0;
    color: var(--text-primary);
  }

  .profile-username {
    font-size: 15px;
    color: #536471;
    margin: 0;
  }

  .profile-bio {
    font-size: 15px;
    margin: 12px 0;
    white-space: pre-wrap;
    color: var(--text-primary);
  }

  .profile-meta {
    display: flex;
    gap: 16px;
    margin: 8px 0;
    color: #536471;
  }

  .profile-meta-item {
    display: flex;
    align-items: center;
    gap: 4px;
    color: #536471;
    font-size: 14px;
  }

  .profile-website {
    color: var(--color-primary);
    text-decoration: none;
  }

  .profile-website:hover {
    text-decoration: underline;
  }

  /* Profile stats */
  .profile-stats {
    display: flex;
    gap: 20px;
    margin: 8px 0 12px 0;
    color: #536471;
  }

  .profile-stat {
    display: flex;
    gap: 4px;
    color: #536471;
    font-size: 14px;
    cursor: pointer;
    background: none;
    border: none;
    padding: 0;
    font-family: inherit;
    text-align: left;
    transition: color 0.2s;
  }

  .profile-stat:hover {
    text-decoration: underline;
  }

  .profile-stat-count {
    font-weight: 700;
    color: var(--text-primary);
  }

  /* Tab Navigation */
  .profile-tabs {
    display: flex;
    border-bottom: 1px solid var(--border-color);
    position: sticky;
    top: 0;
    background-color: var(--bg-color);
    z-index: 10;
    backdrop-filter: blur(12px);
    -webkit-backdrop-filter: blur(12px);
  }

  .profile-tab {
    flex: 1;
    padding: 12px 8px;
    text-align: center;
    color: #536471;
    font-weight: 500;
    cursor: pointer;
    position: relative;
    transition: color 0.2s, background-color 0.2s;
    background-color: transparent;
    border: none;
    font-size: 15px;
  }

  .profile-tab:hover {
    background-color: var(--bg-hover);
    color: var(--text-primary);
  }

  .profile-tab.active {
    color: var(--text-primary);
    font-weight: 700;
  }

  .profile-tab.active::after {
    content: "";
    position: absolute;
    bottom: -1px;
    left: 50%;
    transform: translateX(-50%);
    width: 56px;
    height: 4px;
    border-radius: 9999px 9999px 0 0;
    background-color: var(--color-primary);
    animation: tabIndicatorAppear 0.3s ease;
  }

  @keyframes tabIndicatorAppear {
    from { width: 0; opacity: 0; }
    to { width: 56px; opacity: 1; }
  }

  /* Profile content */
  .profile-content {
    min-height: 300px;
    padding: 0;
    background-color: var(--bg-color);
  }

  .profile-content-empty {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 60px 32px;
    text-align: center;
    color: var(--text-secondary);
  }

  .profile-content-empty-icon {
    font-size: 48px;
    margin-bottom: 16px;
    opacity: 0.7;
  }

  .profile-content-empty-title {
    font-size: 31px;
    font-weight: 700;
    margin: 0 0 8px 0;
    color: var(--text-primary);
  }

  .profile-content-empty-text {
    font-size: 15px;
    color: #536471;
    max-width: 300px;
    margin: 0;
  }

  /* Tweet feed styling */
  .tweet-feed {
    display: flex;
    flex-direction: column;
  }

  .tweet-card-container {
    border-bottom: 1px solid var(--border-color);
    padding: 12px 0;
    transition: background-color 0.2s;
  }
  .tweet-card-container:hover {
    background-color: var(--bg-hover);
  }

  .reply-indicator {
    margin-bottom: 8px;
    font-size: 13px;
    color: #536471;
    padding: 0 16px;
  }

  .reply-indicator a {
    color: var(--color-primary);
    text-decoration: none;
    transition: text-decoration 0.2s;
  }

  .reply-indicator a:hover {
    text-decoration: underline;
  }

  /* Alerts */
  .profile-blocked-alert,
  .profile-private-alert {
    display: flex;
    align-items: center;
    padding: 12px 16px;
    margin: 12px 0;
    border-radius: 8px;
    background-color: var(--accent-light);
    color: var(--text-secondary);
    font-size: 14px;
  }
  
  .profile-blocked-alert :global(.profile-alert-icon),
  .profile-private-alert :global(.profile-alert-icon) {
    margin-right: 8px;
  }

  /* Loading indicators */
  .loading-indicator {
    display: inline-block;
    width: 16px;
    height: 16px;
    border: 2px solid transparent;
    border-top-color: currentColor;
    border-radius: 50%;
    animation: spin 1s linear infinite;
    margin-right: 4px;
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
  
  .error {
    color: #e0245e;
  }

  /* Dark mode adjustments */
  :global(.dark-theme) .profile-header-back {
    background-color: rgba(255, 255, 255, 0.1);
    border-color: rgba(255, 255, 255, 0.05);
  }

  :global(.dark-theme) .profile-header-back:hover {
    background-color: rgba(255, 255, 255, 0.15);
  }

  :global(.dark-theme) .profile-avatar-wrapper {
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
  }

  :global(.dark-theme) .profile-tabs {
    background-color: rgba(0, 0, 0, 0.8);
    border-bottom-color: var(--border-color);
  }
  
  :global(.dark-theme) .profile-follow-button {
    background-color: #ffffff;
    color: #000000;
  }
  
  :global(.dark-theme) .profile-follow-button:hover {
    background-color: #e6e6e6;
  }
  
  :global(.dark-theme) .profile-following-button {
    background-color: transparent;
    color: #e7e9ea;
    border: 1px solid #536471;
  }
  
  :global(.dark-theme) .profile-following-button:hover {
    background-color: rgba(244, 33, 46, 0.1);
    color: #f4212e;
    border-color: rgba(244, 33, 46, 0.4);
  }

  /* Media queries for responsive design */
  @media (max-width: 500px) {
    .profile-header-container {
      height: 160px;
    }
    
    .profile-avatar-container {
      margin-top: -50px;
      margin-left: 12px;
    }
    
    .profile-avatar-wrapper {
      width: 100px;
      height: 100px;
      border-width: 3px;
    }
    
    .profile-details {
      padding: 8px 12px;
    }
    
    .profile-name {
      font-size: 18px;
    }
    
    .profile-username {
      font-size: 14px;
    }
    
    .profile-meta,
    .profile-stats {
      gap: 12px;
    }
    
    .profile-tab {
      font-size: 14px;
      padding: 10px 6px;
    }
    
    .profile-tab.active::after {
      width: 40px;
    }
  }

  /* Modal styling for followers/following */
  .modal-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background-color: rgba(0, 0, 0, 0.6);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
    backdrop-filter: blur(4px);
  }
  
  .modal-container {
    background-color: var(--bg-color);
    border-radius: 16px;
    width: 90%;
    max-width: 600px;
    max-height: 80vh;
    overflow: hidden;
    display: flex;
    flex-direction: column;
    box-shadow: 0 8px 30px rgba(0, 0, 0, 0.3);
    animation: modalAppear 0.2s ease;
  }
  
  @keyframes modalAppear {
    from { opacity: 0; transform: scale(0.95); }
    to { opacity: 1; transform: scale(1); }
  }
  
  .modal-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 16px;
    border-bottom: 1px solid var(--border-color);
  }
  
  .modal-header h2 {
    font-size: 20px;
    font-weight: 700;
    margin: 0;
    color: var(--text-primary);
  }
  
  .modal-close-button {
    background: none;
    border: none;
    padding: 8px;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    border-radius: 50%;
    color: var(--text-secondary);
  }
  
  .modal-close-button:hover {
    background-color: var(--bg-hover);
    color: var(--text-primary);
  }
  
  .modal-content {
    padding: 0;
    overflow-y: auto;
    max-height: calc(80vh - 70px);
  }
  
  .modal-loading, .modal-error, .modal-empty {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 40px 16px;
    text-align: center;
    color: var(--text-secondary);
  }
  
  .modal-retry-button {
    margin-top: 16px;
    padding: 8px 16px;
    border-radius: 20px;
    background-color: var(--bg-color);
    border: 1px solid var(--border-color);
    color: var(--text-primary);
    cursor: pointer;
    font-weight: 500;
  }
  
  .modal-retry-button:hover {
    background-color: var(--bg-hover);
  }
  
  /* User list styling */
  .user-list {
    display: flex;
    flex-direction: column;
  }
  
  .user-item {
    display: flex;
    padding: 16px;
    border-bottom: 1px solid var(--border-color);
    cursor: pointer;
    transition: background-color 0.2s;
    align-items: center;
  }
  
  .user-item:hover {
    background-color: var(--bg-hover);
  }
  
  .user-avatar {
    width: 48px;
    height: 48px;
    border-radius: 50%;
    overflow: hidden;
    margin-right: 12px;
    flex-shrink: 0;
  }
  
  .user-avatar img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }
  
  .user-info {
    flex: 1;
    min-width: 0;
  }
  
  .user-name {
    font-weight: 700;
    color: var(--text-primary);
    font-size: 15px;
    margin-bottom: 2px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  
  .user-username {
    color: var(--text-secondary);
    font-size: 14px;
    margin-bottom: 4px;
  }
  
  .user-bio {
    color: var(--text-primary);
    font-size: 14px;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  
  .profile-follow-button.compact,
  .profile-following-button.compact {
    padding: 6px 12px;
    font-size: 13px;
    min-width: auto;
  }
  
  /* Dark mode adjustments for modals */
  :global(.dark-theme) .modal-container {
    background-color: #000;
    border: 1px solid #2f3336;
  }
  
  :global(.dark-theme) .modal-retry-button {
    background-color: #000;
    border-color: #2f3336;
  }
  
  :global(.dark-theme) .modal-retry-button:hover {
    background-color: rgba(255, 255, 255, 0.03);
  }
</style>
