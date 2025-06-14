<script lang="ts">
  // @ts-nocheck
  import { createEventDispatcher, onMount, onDestroy } from 'svelte';
  import { createLoggerWithPrefix } from '../../utils/logger';
  import { isAuthenticated as checkAuth, getUserId } from '../../utils/auth';
  import { useAuth } from '../../hooks/useAuth';
  import * as api from '../../api';
  import { tweetInteractionStore } from '../../stores/tweetInteractionStore';
  import { notificationStore } from '../../stores/notificationStore';
  import { toastStore } from '../../stores/toastStore';
  import { formatStorageUrl } from '../../utils/common';
  import type { ITweet } from '../../interfaces/ISocialMedia';
  import type { IMedia } from '../../interfaces/IMedia';
  import Linkify from '../common/Linkify.svelte';
  import MessageCircleIcon from 'svelte-feather-icons/src/icons/MessageCircleIcon.svelte';
  import RefreshCwIcon from 'svelte-feather-icons/src/icons/RefreshCwIcon.svelte';
  import HeartIcon from 'svelte-feather-icons/src/icons/HeartIcon.svelte';
  import BookmarkIcon from 'svelte-feather-icons/src/icons/BookmarkIcon.svelte';
  import MoreHorizontalIcon from 'svelte-feather-icons/src/icons/MoreHorizontalIcon.svelte';
  import ArrowUpIcon from 'svelte-feather-icons/src/icons/ArrowUpIcon.svelte';
  import AlertTriangleIcon from 'svelte-feather-icons/src/icons/AlertTriangleIcon.svelte';
  import TrashIcon from 'svelte-feather-icons/src/icons/Trash2Icon.svelte';
  import UserIcon from 'svelte-feather-icons/src/icons/UserIcon.svelte';
  import CornerUpRightIcon from 'svelte-feather-icons/src/icons/CornerUpRightIcon.svelte';
  import EyeIcon from 'svelte-feather-icons/src/icons/EyeIcon.svelte';
  import ChevronUpIcon from 'svelte-feather-icons/src/icons/ChevronUpIcon.svelte';
  import ChevronDownIcon from 'svelte-feather-icons/src/icons/ChevronDownIcon.svelte';
  import CheckCircleIcon from 'svelte-feather-icons/src/icons/CheckCircleIcon.svelte';
  import { authStore } from '../../stores/authStore';
  import { useTheme } from '../../hooks/useTheme';
  import XIcon from 'svelte-feather-icons/src/icons/XIcon.svelte';
  import UsersIcon from 'svelte-feather-icons/src/icons/UsersIcon.svelte';
  import { navigate } from '../../utils/navigation';
  
  // Extract the API methods we need
  const { 
    likeThread, 
    unlikeThread, 
    replyToThread, 
    getReplyReplies, 
    likeReply, 
    unlikeReply, 
    bookmarkThread, 
    removeBookmark, 
    deleteThread, 
    repostThread, 
    removeRepost 
  } = api;
  
  interface ExtendedTweet extends ITweet {
    retweet_id?: string;
    threadId?: string;
    thread_id?: string;
    tweetId?: string;
    userId?: string;
    authorId?: string;
    author_id?: string;
    author_username?: string;
    authorUsername?: string;
    authorName?: string;
    avatar?: string;
    displayName?: string;
    display_name?: string;
    timestamp?: string;
    
    // Legacy metric fields
    likes?: number;
    replies?: number;
    reposts?: number;
    bookmarks?: number;
    views?: number;
    
    // Legacy interaction state fields
    isLiked?: boolean;
    isReposted?: boolean;
    isBookmarked?: boolean;
    isPinned?: boolean;
    is_verified?: boolean; // Add is_verified property to match ITweet
    is_admin?: boolean; // Add is_admin property to match ITweet
    
    // Nested data
    bookmarked_thread?: any;
    bookmarked_reply?: any;
    parent_reply?: any;
    parent_thread?: any;
    isComment?: boolean;
    
    // Poll data
    poll?: any;
    
    // User data
    user?: {
      id?: string;
      username?: string;
      name?: string;
      profile_picture_url?: string;
      is_verified?: boolean;
      verified?: boolean;
      is_admin?: boolean;
    };
    thread?: {
      author?: {
        id?: string;
        username?: string;
        name?: string;
      };
    };
    user_data?: {
      id?: string;
      username?: string;
      name?: string;
    };
    author?: {
      id?: string;
      username?: string;
      name?: string;
      profile_picture_url?: string;
      is_verified?: boolean;
      is_admin?: boolean;
    };
    
    // Original data reference for debugging
    _originalData?: any;
  }
  
  // Initialize logger
  const logger = createLoggerWithPrefix('TweetCard');
  
  // Initialize auth hook with proper destructuring
  const { getAuthState, getAuthToken, refreshToken, checkAndRefreshTokenIfNeeded } = useAuth();
  
  // Create a default empty tweet to use when processedTweet is null
  const defaultEmptyTweet: ExtendedTweet = {
    id: 'empty',
    content: '',
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
    user_id: '',
    username: '',
    name: '',
    profile_picture_url: '',
    likes_count: 0,
    replies_count: 0,
    reposts_count: 0,
    bookmark_count: 0,
    is_liked: false,
    is_reposted: false,
    is_bookmarked: false,
    is_pinned: false,
    is_verified: false,
    is_admin: false,
    media: [],
    parent_id: null
  };
  
  // This will be injected from props when component is used
  export let isAuth: boolean = false;
  
  // Set up local authentication state - make this writable
  let authState: any = getAuthState();
  let isAuthenticated = Boolean(authState?.is_authenticated || isAuth);
  
  // Store auth update function for later use
  const updateAuthState = () => {
    authState = getAuthState();
    isAuthenticated = Boolean(authState?.is_authenticated || isAuth);
  };
  
  $: {
    // Keep isAuthenticated up to date with props and auth state
    isAuthenticated = Boolean(authState?.is_authenticated || isAuth);
  }
  
  // Reactive tweet store setup
  export let tweet: ITweet | ExtendedTweet;
  
  // Update isDarkMode to use themeStore from useTheme
let isDarkMode: boolean = false;
const { theme: themeStore } = useTheme();
themeStore.subscribe(theme => {
  isDarkMode = theme === 'dark';
});
  
  // Changed from export let to export const since they're only for external reference
  export const isLiked: boolean = false;
  export const isReposted: boolean = false;
  export const isBookmarked: boolean = false;
  
  // Changed from export let to export const since it's only for external reference
  export const inReplyToTweet: ITweet | null = null;
  export let replies: (ITweet | ExtendedTweet)[] = [];
  export let showReplies: boolean = false;
  export let nestingLevel: number = 0;
  const MAX_NESTING_LEVEL = 3;
  export let nestedRepliesMap: Map<string, (ITweet | ExtendedTweet)[]> = new Map();
  
  const dispatch = createEventDispatcher();
    // Process the tweet and create a standardized version to work with
  $: processedTweet = processTweetContent(tweet);
  
  // Create a safe version of processedTweet that's never null
  $: safeTweet = processedTweet || defaultEmptyTweet;
  
  // Add a non-null assertion for TypeScript
  let nonNullTweet: ExtendedTweet;
  $: {
    if (processedTweet === null) {
      nonNullTweet = defaultEmptyTweet;
    } else {
      nonNullTweet = processedTweet;
    }
  }
  
  // Define missing variables
  let className = "";
  let isReply = false;
  let isRepost = false;

  // Add a function to process content with entities
  function processContentWithEntities(content: string): string {
    // Simple implementation to handle mentions and hashtags
    if (!content) return '';
    
    // Replace mentions with links
    content = content.replace(/@(\w+)/g, '<a href="/user/$1">@$1</a>');
    
    // Replace hashtags with links
    content = content.replace(/#(\w+)/g, '<a href="/hashtag/$1">#$1</a>');
    
    return content;
  }
  
  // Subscribe to the tweet interaction store - declare first
  let storeInteraction: any = undefined;
  
  // Connect to the interaction store - use null check
  $: storeInteraction = $tweetInteractionStore?.get(safeTweet.id);
  
  // For interaction states, use the store if available, otherwise use the tweet's values with null check
  $: effectiveIsLiked = storeInteraction?.is_liked ?? safeTweet.is_liked;
  $: effectiveIsReposted = storeInteraction?.is_reposted ?? safeTweet.is_reposted;
  $: effectiveIsBookmarked = storeInteraction?.is_bookmarked ?? safeTweet.is_bookmarked;
  
  // For count values, use the store if available, otherwise use the tweet's values with null check
  $: effectiveLikes = storeInteraction?.likes ?? parseCount(safeTweet.likes_count);
  $: effectiveReplies = storeInteraction?.replies ?? parseCount(safeTweet.replies_count);
  $: effectiveReposts = storeInteraction?.reposts ?? parseCount(safeTweet.reposts_count);
  $: effectiveBookmarks = storeInteraction?.bookmarks ?? parseCount(safeTweet.bookmark_count);
  
  // Track loading states for interaction buttons
  let isLikeLoading = false;
  let isRepostLoading = false;
  let isBookmarkLoading = false;
  let isShowingReplies = showReplies;
  
  // For tracking reply loading states
  let isLoadingReplies = false;
  let repliesErrorState = false;
    // For loading nested replies
  let isLoadingNestedReplies = new Map<string, boolean>();
  let nestedRepliesErrorState = new Map<string, boolean>();
  
  // For tracking reply loading states
  let replyActionsLoading = new Map<string, any>();
  
  // Current request IDs to prevent race conditions
  let currentLikeRequestId = 0;
  let currentRepostRequestId = 0;
  let currentBookmarkRequestId = 0;
  
  // Update the conditional rendering logic to make the replies toggle more visible
  // First, modify how we determine if a tweet has replies with null check
  $: hasReplies = effectiveReplies > 0 || parseCount(safeTweet.replies_count) > 0;
  
  // Filter out null values from processed replies and cast to the correct type
  $: processedReplies = replies
    .map(reply => processTweetContent(reply))
    .filter((reply): reply is ExtendedTweet => reply !== null);
  
  // Set tweetId with null check
  $: tweetId = safeTweet.id;
  
  // Initialize the tweet in the store on mount
  onMount(() => {
    tweetInteractionStore.initTweet(safeTweet);
  });

  // When processedTweet changes, make sure it's initialized in the store
  $: if (safeTweet) {
    tweetInteractionStore.initTweet(safeTweet);
  }

  // Subscribe to the store for this specific tweet
  $: {
    if (tweetId) {
      tweetInteractionStore.subscribe(store => {
        storeInteraction = store.get(tweetId);
      });
    }
  }

  function isValidUsername(username: string | undefined | null): boolean {
    return !!username && 
           username !== 'anonymous' && 
           username !== 'user' && 
           username !== 'unknown' &&
           username !== 'undefined';
  }

  function isValidDisplayName(name: string | undefined | null): boolean {
    return !!name && 
           name !== 'User' && 
           name !== 'Anonymous User' &&
           name !== 'Anonymous' &&
           name !== 'undefined';
  }

  function processTweetContent(rawTweet: any): ExtendedTweet | null {
    if (!rawTweet) {
      console.error('Invalid tweet data provided to processTweetContent:', rawTweet);
      return null;
    }
    
    try {
      // Log verification fields from the API response
      console.debug('TWEET VERIFICATION CHECK:', {
        id: rawTweet.id || 'unknown',
        username: rawTweet.username,
        isVerifiedDirect: rawTweet.is_verified,
        userIsVerified: rawTweet.user?.is_verified,
        authorIsVerified: rawTweet.author?.is_verified
      });
      
      // Log admin status fields from API response
      console.debug('TWEET ADMIN CHECK:', {
        id: rawTweet.id || 'unknown',
        username: rawTweet.username,
        isAdminDirect: rawTweet.is_admin,
        userIsAdmin: rawTweet.user?.is_admin,
        authorIsAdmin: rawTweet.author?.is_admin
      });
      
      // Make a deep copy to avoid modifying the original
      const processed: ExtendedTweet = {
        ...rawTweet,
        // Ensure all required ITweet fields exist with fallbacks
        id: rawTweet.id || rawTweet.thread_id || rawTweet.threadId || `unknown-${Date.now()}-${Math.random().toString(36).substring(2, 9)}`,
        content: rawTweet.content || rawTweet.Content || '',
        created_at: rawTweet.created_at || rawTweet.CreatedAt || new Date().toISOString(),
        updated_at: rawTweet.updated_at || rawTweet.UpdatedAt,
        
        // User information with extensive fallbacks
        user_id: extractUserId(rawTweet),
        username: extractUsername(rawTweet),
        name: extractDisplayName(rawTweet),
        profile_picture_url: extractProfilePicture(rawTweet),
        
        // Interaction metrics with fallbacks
        likes_count: safeParseNumber(rawTweet.likes_count || rawTweet.LikesCount || rawTweet.LikeCount || rawTweet.like_count || rawTweet.metrics?.likes),
        replies_count: safeParseNumber(rawTweet.replies_count || rawTweet.RepliesCount || rawTweet.ReplyCount || rawTweet.reply_count || rawTweet.repliesCount || rawTweet.replyCount || rawTweet.replies || rawTweet.metrics?.replies || 0),
        reposts_count: safeParseNumber(rawTweet.reposts_count || rawTweet.RepostsCount || rawTweet.RepostCount || rawTweet.repost_count || rawTweet.metrics?.reposts),
        bookmark_count: safeParseNumber(rawTweet.bookmark_count || rawTweet.BookmarkCount || rawTweet.bookmarks_count || rawTweet.metrics?.bookmarks),
        
        // Interaction states with fallbacks
        is_liked: Boolean(rawTweet.is_liked || rawTweet.IsLiked || rawTweet.liked_by_user || rawTweet.LikedByUser || false),
        is_reposted: Boolean(rawTweet.is_reposted || rawTweet.IsReposted || rawTweet.reposted_by_user || rawTweet.RepostedByUser || rawTweet.is_repost || false),
        is_bookmarked: Boolean(rawTweet.is_bookmarked || rawTweet.IsBookmarked || rawTweet.bookmarked_by_user || rawTweet.BookmarkedByUser || false),
        is_pinned: Boolean(rawTweet.is_pinned || rawTweet.IsPinned || rawTweet.pinned || false),
        is_verified: isVerified(rawTweet),
        is_admin: Boolean(rawTweet.is_admin || rawTweet.IsAdmin || rawTweet.user?.is_admin || rawTweet.author?.is_admin || false),
        
        // Media with validation
        media: validateMedia(rawTweet.media || rawTweet.Media || []),
        
        // Poll data
        poll: rawTweet.poll || null,
        
        // Store original data for debugging
        _originalData: rawTweet
      };
      
      console.log(`Tweet processed - verified status: ${processed.is_verified}, admin status: ${processed.is_admin} (${processed.name || processed.displayName})`);
      
      // Before the return statement, make sure is_verified is set
      processed.is_verified = isVerified(rawTweet);
      
      return processed;
    } catch (error) {
      console.error('Error processing tweet content:', error, rawTweet);
      return null;
    }
  }
  
  // Placeholder function removed
  
  // Helper function to extract user ID with fallbacks
  function extractUserId(rawTweet: any): string {
    return rawTweet.user_id || 
      rawTweet.UserID || 
      rawTweet.userId ||
      rawTweet.author_id || 
      rawTweet.authorId ||
      rawTweet.user?.id || 
      rawTweet.author?.id ||
      rawTweet.thread?.author?.id ||
      rawTweet.user_data?.id || 
      '';
  }
  
  // Helper function to extract username with fallbacks
  function extractUsername(rawTweet: any): string {
    return rawTweet.username || 
      rawTweet.Username || 
      rawTweet.author_username || 
      rawTweet.authorUsername ||
      rawTweet.user?.username || 
      rawTweet.author?.username ||
      rawTweet.thread?.author?.username ||
      rawTweet.user_data?.username || 
      'anonymous';
  }
  
  // Helper function to extract display name with fallbacks
  function extractDisplayName(rawTweet: any): string {
    return rawTweet.name || 
      rawTweet.DisplayName || 
      rawTweet.display_name || 
      rawTweet.authorName ||
      rawTweet.author_name ||
      rawTweet.user?.name || 
      rawTweet.user?.display_name ||
      rawTweet.author?.name ||
      rawTweet.thread?.author?.name ||
      rawTweet.user_data?.name || 
      'User';
  }
  
  // Helper function to extract profile picture
  function extractProfilePicture(rawTweet: any): string {
    const picUrl = rawTweet.profile_picture_url || 
      rawTweet.ProfilePicture || 
      rawTweet.author_avatar || 
      rawTweet.avatar ||
      rawTweet.user?.profile_picture_url ||
      rawTweet.author?.profile_picture_url ||
      rawTweet.user_data?.profile_picture_url;
      
    if (!picUrl) return '';
    
    // Use the formatStorageUrl utility function to handle all URL formatting
    return formatStorageUrl(picUrl);
  }
  
  // Helper function to safely parse numbers
  function safeParseNumber(value: any): number {
    if (value === undefined || value === null) return 0;
    if (typeof value === 'number') return value;
    if (typeof value === 'string') {
      const parsed = parseInt(value, 10);
      return isNaN(parsed) ? 0 : parsed;
    }
    return 0;
  }
  
  // Helper function to validate media
  function validateMedia(media: any): IMedia[] {
    console.log("Validating media:", media);
    
    if (!media) {
      console.warn("Media is null or undefined");
      return [];
    }
    
    if (!Array.isArray(media)) {
      console.warn("Media is not an array, attempting to parse:", typeof media);
      try {
        // Try to parse if it's a string
        if (typeof media === 'string') {
          const parsed = JSON.parse(media);
          if (Array.isArray(parsed)) {
            console.log("Successfully parsed media string into array:", parsed);
            return validateMedia(parsed); // Recursively validate the parsed array
        }
        }
        console.warn("Could not convert media to array");
        return [];
      } catch (e) {
        console.error("Error parsing media string:", e);
        return [];
      }
    }
    
    // Filter out invalid media items and format URLs
    const validatedMedia = media
      .filter(item => {
        if (!item) {
          console.warn("Filtered out null/undefined media item");
          return false;
        }
        if (!item.url) {
          console.warn("Filtered out media item with no URL:", item);
          return false;
        }
        return true;
      })
              .map(item => {
        try {
          const formattedUrl = formatStorageUrl(item.url);
          
          return {
      id: item.id || `media-${Date.now()}-${Math.random().toString(36).substring(2, 15)}`,
            url: formattedUrl,
      type: item.type || 'image',
            thumbnail: item.thumbnail ? formatStorageUrl(item.thumbnail) : formattedUrl,
      alt_text: item.alt_text || item.alt || 'Media attachment'
          };
        } catch (error) {
          console.error("Error formatting URL for media item:", error, item);
          return null;
        }
      })
      .filter(item => item !== null);
    
    console.log("Validated media result:", validatedMedia);
    return validatedMedia;
  }
  
  // Helper function to format timestamp (custom implementation instead of timeago.js)
  function formatTimeAgo(timestamp: string | undefined): string {
    if (!timestamp) return '';
    
    try {
      const date = new Date(timestamp);
      const now = new Date();
      const seconds = Math.floor((now.getTime() - date.getTime()) / 1000);
      
      // Check if the date is valid
      if (isNaN(date.getTime())) {
        return '';
      }
      
      // Less than a minute
      if (seconds < 60) {
        return 'just now';
      }
      
      // Less than an hour
      if (seconds < 3600) {
        const minutes = Math.floor(seconds / 60);
        return `${minutes}m`;
      }
      
      // Less than a day
      if (seconds < 86400) {
        const hours = Math.floor(seconds / 3600);
        return `${hours}h`;
      }
      
      // Less than a week
      if (seconds < 604800) {
        const days = Math.floor(seconds / 86400);
        return `${days}d`;
      }
      
      // Format as date
      const options: Intl.DateTimeFormatOptions = { month: 'short', day: 'numeric' };
      // Add year if it's not the current year
      if (date.getFullYear() !== now.getFullYear()) {
        options.year = 'numeric';
      }
      
      return date.toLocaleDateString(undefined, options);
    } catch (e) {
      console.error('Error formatting date:', e);
      return '';
    }
  }

  // Helper function to safely parse interaction counts
  function parseCount(value: any): number {
    if (value === undefined || value === null) return 0;
    if (typeof value === 'number') return value;
    if (typeof value === 'string') {
      const parsed = parseInt(value, 10);
      return isNaN(parsed) ? 0 : parsed;
    }
    return 0;
  }
  
  // Safely convert a value to string for event dispatching
  function safeToString(value: any): string {
    if (value === undefined || value === null) return '';
    return String(value);
  }

  // Function to handle replies
  async function handleReply() {
    if (!isAuthenticated) {
      navigate('/login', {
        replace: true,
        showToast: true,
        toastMessage: 'Please log in to reply',
        toastType: 'info'
      });
      return;
    }
    
    dispatch('reply', safeTweet);
  }
  
  // Handle repost click
  async function handleRepostClick(event: Event) {
    event.preventDefault();
    event.stopPropagation();
    
    if (!isAuthenticated) {
      // Redirect to login if not authenticated
      navigate('/login', { 
        replace: true,
        showToast: true,
        toastMessage: 'Please log in to repost',
        toastType: 'info'
      });
      return;
    }
    
    if (isRepostLoading) return;
    
    // Generate a unique request ID to prevent race conditions
    const requestId = ++currentRepostRequestId;
    isRepostLoading = true;
    
    try {
      // Optimistically update the UI
      const newRepostState = !effectiveIsReposted;
      const newRepostCount = effectiveReposts + (newRepostState ? 1 : -1);
      
      tweetInteractionStore.updateTweetInteraction(safeTweet.id, {
        is_reposted: newRepostState,
        reposts: newRepostCount
      });
      
      // Make the API call
      if (newRepostState) {
        await repostThread(safeTweet.id);
      } else {
        await removeRepost(safeTweet.id);
      }
      
      // If this isn't the latest request, ignore the result
      if (requestId !== currentRepostRequestId) return;
      
      // Show success toast
      toastStore.showToast(`Post ${newRepostState ? 'reposted' : 'unreposted'} successfully`, 'success');
      
    } catch (error) {
      console.error('Error toggling repost status:', error);
      
      // If this isn't the latest request, ignore the error
      if (requestId !== currentRepostRequestId) return;
      
      // Revert the optimistic update on error
      tweetInteractionStore.updateTweetInteraction(safeTweet.id, {
        is_reposted: effectiveIsReposted,
        reposts: effectiveReposts
      });
      
      toastStore.showToast(`Failed to ${effectiveIsReposted ? 'unrepost' : 'repost'} the post. Please try again.`, 'error');
    } finally {
      // If this is the latest request, reset loading state
      if (requestId === currentRepostRequestId) {
        isRepostLoading = false;
      }
    }
  }
  
  // Handle like click
  async function handleLikeClick(event: Event) {
    event.preventDefault();
    event.stopPropagation();
    
    if (!isAuthenticated) {
      // Redirect to login if not authenticated
      navigate('/login', { 
        replace: true,
        showToast: true,
        toastMessage: 'Please log in to like posts',
        toastType: 'info'
      });
      return;
    }
    
    if (isLikeLoading) return;
    isLikeLoading = true;
    
    try {
      // Optimistically update the UI
      const newLikeState = !effectiveIsLiked;
      const newLikeCount = effectiveLikes + (newLikeState ? 1 : -1);
      
      tweetInteractionStore.updateTweetInteraction(safeTweet.id, {
        is_liked: newLikeState,
        likes: newLikeCount
      });
      
      // Trigger heart animation if liking
      if (newLikeState) {
        heartAnimating = true;
        setTimeout(() => {
          heartAnimating = false;
        }, 800);
      }
      
      // Make the API call
      if (newLikeState) {
        await likeThread(safeTweet.id);
      } else {
        await unlikeThread(safeTweet.id);
      }
      
      // Show success toast
      toastStore.showToast(`Post ${newLikeState ? 'liked' : 'unliked'} successfully`, 'success');
      
    } catch (error) {
      console.error('Error toggling like status:', error);
      
      // Revert the optimistic update on error
      tweetInteractionStore.updateTweetInteraction(safeTweet.id, {
        is_liked: effectiveIsLiked,
        likes: effectiveLikes
      });
      
      toastStore.showToast(`Failed to ${effectiveIsLiked ? 'unlike' : 'like'} the post. Please try again.`, 'error');
    } finally {
      isLikeLoading = false;
    }
  }
  
  // Function to toggle replies visibility
  function toggleReplies() {
    isShowingReplies = !isShowingReplies;
    
    if (isShowingReplies && !replies.length) {
      loadReplies();
    }
  }
  
  // Function to load replies
  async function loadReplies() {
    if (isLoadingReplies) return;
    
    isLoadingReplies = true;
    repliesErrorState = false;
    
    try {
      const response = await getReplyReplies(safeTweet.id);
      replies = response.data || [];
    } catch (error) {
      console.error('Error loading replies:', error);
      repliesErrorState = true;
      toastStore.showToast('Failed to load replies. Please try again.', 'error');
    } finally {
      isLoadingReplies = false;
    }
  }

  function handleShare() {
    dispatch('share', processedTweet);
  }

  function handleClick() {
    dispatch('click', processedTweet);
  }

  function handleNestedReply(event) {
    // Ensure we're passing a string ID
    const replyId = typeof event.detail === 'string' ? event.detail : String(event.detail);
    dispatch('reply', replyId);
  }

  async function handleLoadNestedReplies(event) {
    // Ensure replyId is a string
    const replyId = typeof event.detail === 'string' ? event.detail : String(event.detail);
    
    if (!replyId) {
      console.error("Missing reply ID in handleLoadNestedReplies");
      return;
    }
    
    console.log(`Loading nested replies for reply: ${replyId}`);
    
    // Add loading state to the specific reply
    const replyContainer = document.querySelector(`#reply-${replyId}-container`);
    if (replyContainer) {
      replyContainer.classList.add('loading-nested-replies');
    }

    // Keep track of loading state
    const loadingKey = `loading_${replyId}`;
    // Fix the comparison by checking for the object's loading property instead of direct comparison
    const isLoading = nestedRepliesMap.get(loadingKey) && 
                     (nestedRepliesMap.get(loadingKey) as any)?.loading === true;
    
    // Only attempt to load if not already loading
    if (!isLoading) {
      // Use a temporary object for loading state to avoid type errors
      const loadingState = { loading: true };
      nestedRepliesMap.set(loadingKey, loadingState as any);
      nestedRepliesMap = new Map(nestedRepliesMap);
    
      try {
        // Fetch replies to the reply with a page limit
        const response = await getReplyReplies(replyId, 1, 20);
        
        if (response && response.replies && response.replies.length > 0) {
          console.log(`Received ${response.replies.length} nested replies for reply ${replyId}`);
          
          // Process replies for display
          const processedReplies = response.replies.map(reply => {
            return processTweetContent(reply);
          });
          
          // Update the nested replies map
          nestedRepliesMap.set(replyId, processedReplies);
          
          // Update total count for pagination
          const countKey = `total_count_${replyId}`;
          const countObject = { count: response.total_count || processedReplies.length };
          nestedRepliesMap.set(countKey, countObject as any);
        } else {
          console.warn(`No nested replies returned for reply ${replyId}`);
          nestedRepliesMap.set(replyId, []);
        }
        
        // Remove loading state
        nestedRepliesMap.delete(loadingKey);
        // Force reactivity update
        nestedRepliesMap = new Map(nestedRepliesMap);
        
        // Remove loading class from reply container
        if (replyContainer) {
          replyContainer.classList.remove('loading-nested-replies');
        }
      } catch (error) {
        console.error(`Error loading nested replies for reply ${replyId}:`, error);
        // Add a retry flag as an object to avoid type errors
        const retryKey = `retry_${replyId}`;
        const retryObject = { retry: true };
        nestedRepliesMap.set(retryKey, retryObject as any);
        
        // Force reactivity update
        nestedRepliesMap = new Map(nestedRepliesMap);
      }
    }
  }
  
  // Function to retry loading nested replies
  async function retryLoadNestedReplies(replyId) {
    // Clear any existing retry flag
    const retryKey = `retry_${replyId}`;
    nestedRepliesMap.delete(retryKey);
    
    // Trigger load through normal channel
    handleLoadNestedReplies({ detail: replyId });
  }

  function handleNestedLike(event) {
    if (event.type === 'unlike') {
      dispatch('unlike', event.detail);
    } else {
      dispatch('like', event.detail);
    }
  }

  function handleNestedBookmark(event) {
    if (event.type === 'removeBookmark') {
      dispatch('removeBookmark', event.detail);
    } else {
      dispatch('bookmark', event.detail);
    }
  }

  function handleNestedRepost(event) {
    dispatch('repost', event.detail);
  }

  function showLoginModal() {
    toastStore.showToast('You need to be logged in to perform this action', 'error');
  }

  async function handleLikeReply(replyId: any) {
    try {
      if (!checkAuth()) {
        showLoginModal();
        return;
      }

      const reply = replies.find(r => r.id === replyId);
      if (!reply) return;
      
      // Check if already in loading state
      const loadingState = replyActionsLoading.get(String(replyId)) || {};
      if (loadingState.like) return;
      
      // Set loading state
      loadingState.like = true;
      replyActionsLoading.set(String(replyId), loadingState);
      replyActionsLoading = new Map(replyActionsLoading);
      
      // Add heart animation
      replyHeartAnimations.set(String(replyId), true);
      setTimeout(() => {
        replyHeartAnimations.delete(String(replyId));
      }, 800);
      
      // Optimistic UI update
      reply.is_liked = true;
      if (typeof reply.likes_count === 'number') {
        reply.likes_count += 1;
      }
      
      // Provide haptic feedback on mobile devices
      if (window.navigator && window.navigator.vibrate) {
        try {
          window.navigator.vibrate(30); // lighter vibration for reply likes
        } catch (e) {
          // Ignore vibration API errors
        }
      }
      
      // Call API only if online
      if (navigator && navigator.onLine) {
        try {
          await likeReply(String(replyId));
        } catch (error) {
          console.error('Error liking reply:', error);
          
          // Check for "already liked" error - don't revert UI
          const errorMsg = error instanceof Error ? error.message.toLowerCase() : '';
          const isAlreadyLiked = errorMsg.includes('already liked');
          if (!isAlreadyLiked) {
            // Revert optimistic update
            reply.is_liked = false;
            if (typeof reply.likes_count === 'number' && reply.likes_count > 0) {
              reply.likes_count -= 1;
            }
            toastStore.showToast('Failed to like reply. Please try again.', 'error');
          }
        } 
      } else {
        // Store offline likes in localStorage for later syncing
        try {
          const offlineReplyLikes = JSON.parse(localStorage.getItem('offlineReplyLikes') || '{}');
          offlineReplyLikes[replyId] = { action: 'like', timestamp: Date.now() };
          localStorage.setItem('offlineReplyLikes', JSON.stringify(offlineReplyLikes));
          toastStore.showToast('Liked! Will be synced when you\'re back online.', 'info', 2000);
        } catch (e) {
          console.error('Failed to save offline reply like', e);
        }
      }
    } catch (error) {
      console.error('Unhandled error in handleLikeReply:', error);
      toastStore.showToast('An unexpected error occurred. Please try again later.', 'error');
    } finally {
      // Clear loading state
      const loadingState = replyActionsLoading.get(String(replyId)) || {};
      loadingState.like = false;
      replyActionsLoading.set(String(replyId), loadingState);
      replyActionsLoading = new Map(replyActionsLoading);
    }
  }

  async function handleUnlikeReply(replyId: any) {
    try {
      if (!checkAuth()) {
        showLoginModal();
        return;
      }

      const reply = replies.find(r => r.id === replyId);
      if (!reply) return;
      
      // Check if already in loading state
      const loadingState = replyActionsLoading.get(String(replyId)) || {};
      if (loadingState.like) return;
      
      // Set loading state
      loadingState.like = true;
      replyActionsLoading.set(String(replyId), loadingState);
      replyActionsLoading = new Map(replyActionsLoading);      // Optimistic UI update
      reply.is_liked = false;
      if (typeof reply.likes_count === 'number' && reply.likes_count > 0) {
        reply.likes_count -= 1;
      }
      
      // Call API
      try {
        await unlikeReply(String(replyId));
      } catch (error) {
        console.error('Error unliking reply:', error);
        
        // Check for "not liked" or "not found" error - don't revert UI
        const errorMsg = error instanceof Error ? error.message.toLowerCase() : '';
        const isNotLiked = errorMsg.includes('not liked') || errorMsg.includes('not found');
          if (!isNotLiked) {
          // Revert optimistic update
          reply.is_liked = true;
          if (typeof reply.likes_count === 'number') {
            reply.likes_count += 1;
          }
          toastStore.showToast('Failed to unlike reply. Please try again.', 'error');
        }
      } finally {
        // Clear loading state
        loadingState.like = false;
        replyActionsLoading.set(String(replyId), loadingState);
        replyActionsLoading = new Map(replyActionsLoading);
      }
    } catch (error) {
      console.error('Unhandled error in handleUnlikeReply:', error);
      toastStore.showToast('An unexpected error occurred. Please try again later.', 'error');
    }
  }

  async function handleBookmarkReply(replyId: any) {
    try {
      if (!checkAuth()) {
        showLoginModal();
        return;
      }

      const reply = replies.find(r => r.id === replyId);
      if (!reply) return;
      
      // Check if already in loading state
      const loadingState = replyActionsLoading.get(String(replyId)) || {};
      if (loadingState.bookmark) return;
      
      // Set loading state
      loadingState.bookmark = true;
      replyActionsLoading.set(String(replyId), loadingState);
      replyActionsLoading = new Map(replyActionsLoading);
      
      // Optimistic UI update
      reply.is_bookmarked = true;
      if (typeof reply.bookmark_count === 'number') {
        reply.bookmark_count += 1;
      }
      
      // Call API
      try {
        // Using the correct API call for bookmarking
        await bookmarkThread(String(replyId));
        toastStore.showToast('Reply bookmarked', 'success');
      } catch (error) {
        console.error('Error bookmarking reply:', error);
        
        // Check for "already bookmarked" error - don't revert UI
        const errorMsg = error instanceof Error ? error.message.toLowerCase() : '';
        const isAlreadyBookmarked = errorMsg.includes('already bookmarked') || errorMsg.includes('already exists');
        
        if (!isAlreadyBookmarked) {
          // Revert optimistic update
          reply.is_bookmarked = false;
          if (typeof reply.bookmark_count === 'number' && reply.bookmark_count > 0) {
            reply.bookmark_count -= 1;
          }
          toastStore.showToast('Failed to bookmark reply. Please try again.', 'error');
        } else {
          // If it's already bookmarked, just confirm to the user
          toastStore.showToast('Reply is already bookmarked', 'info');
        }
      } finally {
        // Clear loading state
        loadingState.bookmark = false;
        replyActionsLoading.set(String(replyId), loadingState);
        replyActionsLoading = new Map(replyActionsLoading);
      }
    } catch (error) {
      console.error('Unhandled error in handleBookmarkReply:', error);
      toastStore.showToast('An unexpected error occurred. Please try again later.', 'error');
    }
  }

  async function handleUnbookmarkReply(replyId: any) {
    try {
      if (!checkAuth()) {
        showLoginModal();
        return;
      }

      const reply = replies.find(r => r.id === replyId);
      if (!reply) return;
      
      // Check if already in loading state
      const loadingState = replyActionsLoading.get(String(replyId)) || {};
      if (loadingState.bookmark) return;
      
      // Set loading state
      loadingState.bookmark = true;
      replyActionsLoading.set(String(replyId), loadingState);
      replyActionsLoading = new Map(replyActionsLoading);
      
      // Optimistic UI update
      reply.is_bookmarked = false;
      if (typeof reply.bookmark_count === 'number' && reply.bookmark_count > 0) {
        reply.bookmark_count -= 1;
      }
      
      // Call API
      try {
        // Using the correct API call for unbookmarking
        await removeBookmark(String(replyId));
        toastStore.showToast('Bookmark removed', 'success');
      } catch (error) {
        console.error('Error unbookmarking reply:', error);
        
        // Check for "not bookmarked" or "not found" error - don't revert UI
        const errorMsg = error instanceof Error ? error.message.toLowerCase() : '';
        const isNotBookmarked = errorMsg.includes('not bookmarked') || errorMsg.includes('not found');
        
        if (!isNotBookmarked) {
          // Revert optimistic update
          reply.is_bookmarked = true;
          if (typeof reply.bookmark_count === 'number') {
            reply.bookmark_count += 1;
          }
          toastStore.showToast('Failed to remove bookmark from reply. Please try again.', 'error');
        } else {
          // If it's already not bookmarked, just confirm to the user
          toastStore.showToast('Reply was not bookmarked', 'info');
        }
      } finally {
        // Clear loading state
        loadingState.bookmark = false;
        replyActionsLoading.set(String(replyId), loadingState);
        replyActionsLoading = new Map(replyActionsLoading);
      }
    } catch (error) {
      console.error('Unhandled error in handleUnbookmarkReply:', error);
      toastStore.showToast('An unexpected error occurred. Please try again later.', 'error');
    }
  }

  function navigateToUserProfile(event: MouseEvent | KeyboardEvent, username: string, userId?: string | number | null) {
    event.stopPropagation(); // Prevent the main tweet click
    
    const effectiveUserId = userId ? String(userId) : 
      processedTweet.userId ? String(processedTweet.userId) : 
      processedTweet.authorId ? String(processedTweet.authorId) : 
      processedTweet.author_id ? String(processedTweet.author_id) : 
      processedTweet.user_id ? String(processedTweet.user_id) : null;
    
    // For debugging
    console.log("🔍 User Navigation Debug:", { 
      username, 
      providedUserId: userId,
      effectiveUserId,
      availableFields: {
        userId: processedTweet.userId,
        authorId: processedTweet.authorId,
        author_id: processedTweet.author_id,
        user_id: processedTweet.user_id,
        displayName: processedTweet.displayName,
      },
      tweetInfo: {
        id: processedTweet.id,
        content: processedTweet.content?.substring(0, 30) + '...',
      }
    });
    
    // If we have a user ID, use that for navigation (most reliable)
    if (effectiveUserId) {
      console.log(`✅ Using userId for navigation: ${effectiveUserId}`);
      navigate(`/user/${effectiveUserId}`);
      return;
    }
    
    // Otherwise fall back to username if it's valid
    if (username && username !== 'anonymous' && username !== 'user' && username !== 'unknown') {
      console.log(`✅ Falling back to username for navigation: ${username}`);
      navigate(`/user/${username}`);
    } else {
      console.error("❌ Navigation failed: No valid ID or username available", { 
        username, providedUserId: userId 
      });
    }
  }

  function debugUserData() {
    // This function can be called to inspect all tweets in the component    console.group('🔍 TWEET DEBUGGING');
    const extTweet = tweet as ExtendedTweet;
    console.log('Main tweet:', {
      id: tweet.id,
      thread_id: extTweet.thread_id || extTweet.threadId,
      username: {
        processed: processedTweet.username,
        original: tweet.username,
        authorUsername: extTweet.authorUsername,
        author_username: extTweet.author_username,
        fromThread: extTweet.thread?.author?.username,
        fromUser: extTweet.user?.username,
        fromAuthor: extTweet.author?.username,
        fromUserData: extTweet.user_data?.username,
        isValid: isValidUsername(processedTweet.username),
      },
      displayName: {
        processed: processedTweet.displayName,
        original: extTweet.displayName,
        authorName: extTweet.authorName, 
        name: tweet.name,
        display_name: extTweet.display_name,
        fromThread: extTweet.thread?.author?.name,
        fromUser: extTweet.user?.name,
        fromAuthor: extTweet.author?.name,
        fromUserData: extTweet.user_data?.name,
        isValid: isValidDisplayName(processedTweet.displayName),
      },
      userId: {
        processed: processedTweet.userId || processedTweet.authorId || processedTweet.author_id || processedTweet.user_id,
        original: extTweet.userId,
        authorId: extTweet.authorId,
        author_id: extTweet.author_id,
        user_id: tweet.user_id,
        fromThread: extTweet.thread?.author?.id,
        fromUser: extTweet.user?.id,
        fromAuthor: extTweet.author?.id,
        fromUserData: extTweet.user_data?.id,
      },
      // Cast to ExtendedTweet to access non-standard properties
      bookmarkedThread: extTweet.bookmarked_thread || null,
      contenWithUsers: tweet.content && tweet.content.includes('@') ? 
        tweet.content.match(/@([a-zA-Z0-9_]+)/g) : null
    });
      if (replies.length > 0) {
      const extReply = replies[0] as ExtendedTweet;
      console.log('First reply:', {
        id: replies[0].id,
        thread_id: extReply.thread_id || extReply.threadId,
        username: {
          original: replies[0].username,
          authorUsername: extReply.authorUsername,
          author_username: extReply.author_username,
          fromThread: extReply.thread?.author?.username,
          fromUser: extReply.user?.username,
          fromAuthor: extReply.author?.username,
          fromUserData: extReply.user_data?.username        },
        displayName: {
          original: extReply.displayName,
          authorName: extReply.authorName, 
          name: replies[0].name,
          display_name: extReply.display_name,
          fromThread: extReply.thread?.author?.name,
          fromUser: extReply.user?.name,
          fromAuthor: extReply.author?.name,
          fromUserData: extReply.user_data?.name
        }
      });
    }
    console.groupEnd();
  }

  // Call this when the component mounts
  $: {
    if (tweet) {
      setTimeout(debugUserData, 500); // Delay to allow processing to complete
    }
  }

  // Handle refresh events from ComposeTweet component
  export function handleRefreshReplies(event) {
    const { threadId, parentReplyId, newReply } = event.detail;
    
    // Process the new reply through our content processor
    const processedNewReply = processTweetContent(newReply);
    
    // If refreshing replies for a thread
    if (threadId === processedTweet.id && !parentReplyId) {
      // Add the new reply to our replies array
      replies = [processedNewReply, ...replies];
      // Make sure replies are visible
      showReplies = true;
    }
    // If refreshing replies for a specific reply (nested reply)
    else if (parentReplyId) {
      // Get the current nested replies for this parent reply
      let currentNestedReplies = nestedRepliesMap.get(parentReplyId) || [];
      // Add the new reply to the nested replies
      nestedRepliesMap.set(parentReplyId, [processedNewReply, ...currentNestedReplies]);
      // Update the map to trigger reactivity
      nestedRepliesMap = new Map(nestedRepliesMap);
      
      // Find the parent reply
      const parentIndex = replies.findIndex(r => r.id === parentReplyId);
      if (parentIndex >= 0) {
        // Create a new array with all the existing replies
        const newReplies = [...replies];
          // Create a new object for the parent reply with the incremented count
        const oldParent = replies[parentIndex];
        const replyCount = typeof (oldParent as ExtendedTweet).replies === 'number' ? 
          (oldParent as ExtendedTweet).replies! + 1 : 
          Number((oldParent as ExtendedTweet).replies || 0) + 1;
        
        // Create a new parent reply object with updated count
        const newParent = { ...oldParent };
        // Use TypeScript's any type for the assignment to bypass type checking
        (newParent as any).replies = replyCount;
        
        // Replace the parent in the new array
        newReplies[parentIndex] = newParent;
        
        // Update the replies array
        replies = newReplies;
      }
    }
  }

  // Inside the <script> section, add this:
  let heartAnimating = false;
  const replyHeartAnimations = new Map<string, boolean>();
  
  // Add an onMount to initialize from store
  onMount(() => {
    // Initialize tweet in interaction store
    if (tweet) {
      tweetInteractionStore.initTweet(tweet);
      
      // Check if we need to sync with server
      syncInteractionWithServer();
    }
    
    // Check authentication status
    const checkAuthStatus = () => {
      updateAuthState();
      console.log('Authentication status:', isAuthenticated);
    };
    
    // Run immediately and set up refresh interval
    checkAuthStatus();
    const authCheckInterval = setInterval(checkAuthStatus, 60000); // Check every minute
    
    // Set up auth change listener
    window.addEventListener('auth:changed', checkAuthStatus);
    
    // Add visibility change listener to sync when returning to page
    document.addEventListener('visibilitychange', handleVisibilityChange);
    
    return () => {
      clearInterval(authCheckInterval);
      window.removeEventListener('auth:changed', checkAuthStatus);
      document.removeEventListener('visibilitychange', handleVisibilityChange);
    };
  });
  
  onDestroy(() => {
    document.removeEventListener('visibilitychange', handleVisibilityChange);
    window.removeEventListener('auth:changed', () => {});
  });
  
  // Function to sync interactions when returning to the tab
  function handleVisibilityChange() {
    if (document.visibilityState === 'visible') {
      syncInteractionWithServer();
      syncOfflineLikes();
    }
  }
  
  // Function to sync with server
  function syncInteractionWithServer() {
    if (navigator && navigator.onLine) {
      tweetInteractionStore.syncWithServer();
    }
  }
  
  // Function to sync offline likes
  function syncOfflineLikes() {
    if (navigator && navigator.onLine) {
      try {
        // Check for offline likes for threads
        const offlineLikes = JSON.parse(localStorage.getItem('offlineLikes') || '{}');
        if (Object.keys(offlineLikes).length > 0) {
          console.log(`Found ${Object.keys(offlineLikes).length} offline likes to sync`);
          
          // Process each offline like
          Object.entries(offlineLikes).forEach(async ([id, data]: [string, any]) => {
            try {
              const action = data.action;
              const apiCall = action === 'like' ? likeThread : unlikeThread;
              await apiCall(id);
              
              // Update the store
              tweetInteractionStore.updateTweetInteraction(id, {
                is_liked: action === 'like',
                pending_like: false
              });
              
              // Remove from localStorage after successful sync
              delete offlineLikes[id];
            } catch (error) {
              console.error(`Failed to sync offline ${data.action} for tweet ${id}:`, error);
            }
          });
          
          // Save the updated offline likes
          localStorage.setItem('offlineLikes', JSON.stringify(offlineLikes));
        }
        
        // Check for offline likes for replies
        const offlineReplyLikes = JSON.parse(localStorage.getItem('offlineReplyLikes') || '{}');
        if (Object.keys(offlineReplyLikes).length > 0) {
          console.log(`Found ${Object.keys(offlineReplyLikes).length} offline reply likes to sync`);
          
          // Process each offline reply like
          Object.entries(offlineReplyLikes).forEach(async ([id, data]: [string, any]) => {
            try {
              const action = data.action;
              const apiCall = action === 'like' ? likeReply : unlikeReply;
              await apiCall(id);
              
              // Remove from localStorage after successful sync
              delete offlineReplyLikes[id];
            } catch (error) {
              console.error(`Failed to sync offline ${data.action} for reply ${id}:`, error);
            }
          });
          
          // Save the updated offline reply likes
          localStorage.setItem('offlineReplyLikes', JSON.stringify(offlineReplyLikes));
        }
      } catch (e) {
        console.error('Error syncing offline likes:', e);
      }
    }
  }

  // Helper function to extract verified status with robust fallbacks
  function isVerified(rawTweet: any): boolean {
    if (!rawTweet) return false;
    
    // Check all possible verification flags
    return Boolean(
      rawTweet.is_verified || 
      rawTweet.IsVerified || 
      rawTweet.verified ||
      rawTweet.user?.is_verified || 
      rawTweet.user?.verified ||
      rawTweet.author?.is_verified ||
      rawTweet.author?.verified
    );
  }

  // In the script section, add isCurrentUserAuthor and dropdown toggle functionality
  // Add this after the other reactive declarations
  $: isCurrentUserAuthor = authState?.user_id === processedTweet?.user_id;
  let isSettingsDropdownOpen = false;

  function toggleSettingsDropdown(e) {
    e.stopPropagation();
    isSettingsDropdownOpen = !isSettingsDropdownOpen;
  }

  // Close dropdown if clicked outside
  function handleClickOutside(e) {
    if (isSettingsDropdownOpen) {
      isSettingsDropdownOpen = false;
    }
  }

  // Add a state variable for the delete confirmation modal
  let showDeleteConfirmationModal = false;
  let tweetToDelete: string | null = null;

  // In the script section, update the handleDeleteTweet function
  async function handleDeleteTweet(e) {
    e.stopPropagation();
    
    if (!isCurrentUserAuthor) {
      toastStore.showToast('You can only delete your own posts', 'error');
      return;
    }
    
    // Show the modal instead of using browser confirm
    tweetToDelete = String(processedTweet.id);
    showDeleteConfirmationModal = true;
    isSettingsDropdownOpen = false; // Close the dropdown
  }

  // Add a new function to perform the actual deletion
  async function confirmDeleteTweet() {
    if (tweetToDelete === null) {
      console.error('No tweet ID to delete');
      return;
    }
    
    try {
      await deleteThread(tweetToDelete);
      toastStore.showToast('Post deleted successfully', 'success');
      
      // Notify parent component that tweet was deleted
      dispatch('deleted', { id: tweetToDelete });
      
      // Remove the tweet from the DOM
      const tweetElement = document.getElementById(`tweet-${tweetToDelete}`);
      if (tweetElement) {
        tweetElement.style.height = `${tweetElement.offsetHeight}px`;
        tweetElement.style.overflow = 'hidden';
        
        // Add animation
        setTimeout(() => {
          tweetElement.style.height = '0';
          tweetElement.style.opacity = '0';
          tweetElement.style.margin = '0';
          tweetElement.style.padding = '0';
          tweetElement.style.transition = 'all 0.3s ease';
        }, 10);
        
        // Remove after animation
        setTimeout(() => {
          tweetElement.remove();
        }, 300);
      }
      
    } catch (error) {
      console.error('Error deleting tweet:', error);
      toastStore.showToast('Failed to delete post. Please try again.', 'error');
    } finally {
      // Close the modal
      showDeleteConfirmationModal = false;
      tweetToDelete = null;
    }
  }

  // Add a function to cancel deletion
  function cancelDeleteTweet() {
    showDeleteConfirmationModal = false;
    tweetToDelete = null;
  }

  // Listen for clicks to close the dropdown
  onMount(() => {
    document.addEventListener('click', handleClickOutside);
  });

  onDestroy(() => {
    document.removeEventListener('click', handleClickOutside);
  });

  // Helper function to check if a tweet is from an admin user
  function isAdminTweet(tweet: ExtendedTweet | null): boolean {
    if (!tweet) return false;
    return Boolean(tweet.is_admin || tweet.user?.is_admin || tweet.author?.is_admin);
  }

  // Process the tweet data into a standardized format
  function processTweet(rawTweet: any): ExtendedTweet | null {
    if (!rawTweet) {
      console.error('Received null or undefined tweet data');
      return null;
    }
    
    try {
      // Extract unique ID with fallbacks
      const id = rawTweet.id || 
        rawTweet.thread_id || 
        rawTweet.threadId || 
        rawTweet._id || 
        rawTweet.ID;
      
      if (!id) {
        console.error('Tweet has no ID:', rawTweet);
        return null;
      }
      
      // Extract content with fallbacks
      const content = rawTweet.content || 
        rawTweet.text || 
        rawTweet.body || 
        rawTweet.message || 
        '';
      
      // Extract timestamps with fallbacks
      const createdAt = rawTweet.created_at || 
        rawTweet.createdAt || 
        rawTweet.timestamp || 
        new Date().toISOString();
        
      const updatedAt = rawTweet.updated_at || 
        rawTweet.updatedAt || 
        createdAt;
      
      // Extract user data with fallbacks
      const username = extractUsername(rawTweet);
      const displayName = extractDisplayName(rawTweet);
      const userId = extractUserId(rawTweet);
      const profilePicture = extractProfilePicture(rawTweet);
      
      // Extract metrics with fallbacks
      const likesCount = safeParseNumber(rawTweet.likes_count || rawTweet.likesCount || rawTweet.likes);
      const repliesCount = safeParseNumber(rawTweet.replies_count || rawTweet.repliesCount || rawTweet.replies);
      const repostsCount = safeParseNumber(rawTweet.reposts_count || rawTweet.repostsCount || rawTweet.reposts);
      const bookmarkCount = safeParseNumber(rawTweet.bookmark_count || rawTweet.bookmarkCount || rawTweet.bookmarks);
      
      // Extract interaction states with fallbacks
      const isLiked = rawTweet.is_liked || 
        rawTweet.isLiked || 
        rawTweet.liked_by_user || 
        rawTweet.likedByUser || 
        false;
        
      const isReposted = rawTweet.is_reposted || 
        rawTweet.isReposted || 
        rawTweet.reposted_by_user || 
        rawTweet.repostedByUser || 
        false;
        
      const isBookmarked = rawTweet.is_bookmarked || 
        rawTweet.isBookmarked || 
        rawTweet.bookmarked_by_user || 
        rawTweet.bookmarkedByUser || 
        false;
      
      const isPinned = rawTweet.is_pinned || 
        rawTweet.isPinned || 
        false;
      
      // Create the standardized tweet object
      return {
        id,
        content,
        created_at: createdAt,
        updated_at: updatedAt,
        username,
        display_name: displayName,
        profile_picture_url: profilePicture,
        likes_count: likesCount,
        replies_count: repliesCount,
        reposts_count: repostsCount,
        bookmark_count: bookmarkCount,
        is_liked: isLiked,
        is_reposted: isReposted,
        is_bookmarked: isBookmarked,
        is_pinned: isPinned,
        user_id: userId,
        media: validateMedia(rawTweet.media),
        parent_id: rawTweet.parent_id || rawTweet.parentId || rawTweet.thread_id || '',
        poll: rawTweet.poll || null,
        _originalData: rawTweet
      };
    } catch (error) {
      console.error('Error processing tweet data:', error, rawTweet);
      return null;
    }
  }

  // Process the tweet content for display
  function processContent(rawTweet: any): string {
    if (!rawTweet) {
      console.error('Received null or undefined tweet data');
      return '';
    }
    
    try {
      // Extract content with fallbacks
      const content = rawTweet.content || 
        rawTweet.text || 
        rawTweet.body || 
        rawTweet.message || 
        '';
      
      // Process mentions, hashtags, and URLs
      return processContentWithEntities(content);
    } catch (error) {
      console.error('Error processing tweet content:', error, rawTweet);
      return '';
    }
  }
</script>

{#if tweet}
  <div class="tweet-card {className} {isDarkMode ? 'tweet-card-dark' : ''}" class:is-reply={isReply} class:is-repost={isRepost}>
    <!-- Tweet content -->
    <div class="tweet-container">
      <!-- Tweet header with user info -->
      <div class="tweet-header">
        <a href={`/user/${safeTweet.username}`}
          class="tweet-avatar"
          on:click|preventDefault={(e) => navigateToUserProfile(e, safeTweet.username, safeTweet.user_id)}
          on:keydown={(e) => e.key === 'Enter' && navigateToUserProfile(e, safeTweet.username, safeTweet.user_id)}>
          {#if safeTweet.profile_picture_url}
            <img src={safeTweet.profile_picture_url} alt={safeTweet.username} class="tweet-avatar-image" />
          {:else}
            <div class="tweet-avatar-placeholder">
              <div class="tweet-avatar-text">{safeTweet.username ? safeTweet.username[0].toUpperCase() : 'U'}</div>
            </div>
          {/if}
        </a>
        <div class="tweet-content-container">
          <div class="tweet-author-info">
            <a href={`/user/${safeTweet.username}`}
              class="tweet-author-name {isDarkMode ? 'tweet-author-name-dark' : ''}"
              on:click|preventDefault={(e) => navigateToUserProfile(e, safeTweet.username, safeTweet.user_id)}
              on:keydown={(e) => e.key === 'Enter' && navigateToUserProfile(e, safeTweet.username, safeTweet.user_id)}>
              <span class="display-name-text">{safeTweet.name || safeTweet.displayName || safeTweet.username}</span>
              {#if safeTweet.is_verified}
                <span class="user-verified-badge">
                  <CheckCircleIcon size="14" />
                </span>
              {/if}
            </a>
            <a href={`/user/${safeTweet.username}`}
              class="tweet-author-username {isDarkMode ? 'tweet-author-username-dark' : ''}"
              on:click|preventDefault={(e) => navigateToUserProfile(e, safeTweet.username, safeTweet.user_id)}
              on:keydown={(e) => e.key === 'Enter' && navigateToUserProfile(e, safeTweet.username, safeTweet.user_id)}>
              @{safeTweet.username}
            </a>
            <span class="tweet-dot-separator {isDarkMode ? 'tweet-dot-separator-dark' : ''}">·</span>
            <span class="tweet-timestamp {isDarkMode ? 'tweet-timestamp-dark' : ''}">{formatTimeAgo(safeTweet.created_at)}</span>
          </div>
          
          <div class="tweet-text {isDarkMode ? 'tweet-text-dark' : ''}">
            <p>{safeTweet.content}</p>
          </div>
          
          {#if safeTweet.media && safeTweet.media.length > 0}
            <div class="tweet-media-container {isDarkMode ? 'tweet-media-container-dark' : ''}">
              {#if safeTweet.media.length === 1}
                <!-- Single Media Display -->
                <div class="tweet-media-single">
                  {#if safeTweet.media[0].type === 'image'}
                    <img 
                      src={safeTweet.media[0].url} 
                      alt={safeTweet.media[0].alt_text || "Media"} 
                      class="tweet-media-img"
                    />
                  {:else if safeTweet.media[0].type === 'video'}
                    <video 
                      src={safeTweet.media[0].url} 
                      controls 
                      class="tweet-media-video"
                    >
                      <track kind="captions" src="/captions/en.vtt" srclang="en" label="English" />
                    </video>
                  {:else}
                    <img 
                      src={safeTweet.media[0].url} 
                      alt="GIF" 
                      class="tweet-media-img"
                    />
                  {/if}
                </div>
              {:else if safeTweet.media.length > 1}
                <!-- Multiple Media Grid -->
                <div class="tweet-media-grid">
                  {#each safeTweet.media.slice(0, 4) as media, index (media.url || index)}
                    <div class="tweet-media-item">
                      {#if media.type === 'image'}
                        <img 
                          src={media.url} 
                          alt={media.alt_text || "Media"} 
                          class="tweet-media-img"
                        />
                      {:else if media.type === 'video'}
                        <video 
                          src={media.url} 
                          controls 
                          class="tweet-media-video"
                        >
                          <track kind="captions" src="/captions/en.vtt" srclang="en" label="English" />
                        </video>
                      {:else}
                        <img 
                          src={media.url} 
                          alt="GIF" 
                          class="tweet-media-img"
                        />
                      {/if}
                    </div>
                  {/each}
                </div>
              {/if}
            </div>
          {/if}
          
          <div class="tweet-actions {isDarkMode ? 'tweet-actions-dark' : ''}">
            <div class="tweet-action-item">
              <button 
                class="tweet-action-btn tweet-reply-btn {hasReplies ? 'has-replies' : ''} {isDarkMode ? 'tweet-action-btn-dark' : ''}" 
                on:click|stopPropagation={handleReply} 
                aria-label="Reply to tweet"
              >
                <MessageCircleIcon size="20" class="tweet-action-icon" />
                <span class="tweet-action-count">{effectiveReplies}</span>
              </button>
            </div>
            <div class="tweet-action-item">
              <button 
                class="tweet-action-btn tweet-repost-btn {effectiveIsReposted ? 'active' : ''} {isDarkMode ? 'tweet-action-btn-dark' : ''}" 
                on:click|stopPropagation={handleRepostClick}
                aria-label="{effectiveIsReposted ? 'Undo repost' : 'Repost'}"
              >
                <RefreshCwIcon size="20" class="tweet-action-icon" />
                <span class="tweet-action-count">{effectiveReposts}</span>
              </button>
            </div>
            <div class="tweet-action-item">
              <button 
                class="tweet-action-btn tweet-like-btn {effectiveIsLiked ? 'active' : ''} {isLikeLoading ? 'loading' : ''} {heartAnimating ? 'animating' : ''} {isDarkMode ? 'tweet-action-btn-dark' : ''}" 
                on:click|stopPropagation={handleLikeClick}
                aria-label="{effectiveIsLiked ? 'Unlike this post' : 'Like this post'}"
              >
                <HeartIcon size="20" fill={effectiveIsLiked ? "currentColor" : "none"} class="tweet-action-icon" />
                <span class="tweet-action-count">{effectiveLikes}</span>
              </button>
            </div>
            <div class="tweet-action-item">
              <button 
                class="tweet-action-btn tweet-bookmark-btn {effectiveIsBookmarked ? 'active' : ''} {isDarkMode ? 'tweet-action-btn-dark' : ''}" 
                on:click|stopPropagation={() => {}}
                aria-label="{effectiveIsBookmarked ? 'Remove bookmark' : 'Bookmark'}"
              >
                <BookmarkIcon size="20" fill={effectiveIsBookmarked ? "currentColor" : "none"} class="tweet-action-icon" />
                <span class="tweet-action-count">{effectiveBookmarks}</span>
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
{:else}
  <!-- Fallback for null tweets -->
  <div class="tweet-card-error {isDarkMode ? 'tweet-card-error-dark' : ''}">
    <p>Unable to display this tweet</p>
  </div>
{/if}

<style>
  .tweet-card {
    border: 1px solid #e1e8ed;
    border-radius: 12px;
    padding: 16px;
    margin-bottom: 16px;
    background-color: white;
    transition: background-color 0.2s, border-color 0.2s;
  }
  
  .tweet-card-dark {
    background-color: #15202b;
    border-color: #38444d;
    color: #fff;
  }
  
  .tweet-container {
    display: flex;
    flex-direction: column;
  }
  
  .tweet-header {
    display: flex;
    margin-bottom: 8px;
  }
  
  .tweet-avatar {
    margin-right: 12px;
    flex-shrink: 0;
  }
  
  .tweet-avatar-image {
    width: 48px;
    height: 48px;
    border-radius: 50%;
    object-fit: cover;
  }
  
  .tweet-avatar-placeholder {
    width: 48px;
    height: 48px;
    border-radius: 50%;
    background-color: #1da1f2;
    display: flex;
    align-items: center;
    justify-content: center;
    color: white;
    font-weight: bold;
  }
  
  .tweet-content-container {
    flex: 1;
    min-width: 0; /* Allows proper text truncation */
  }
  
  .tweet-author-info {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    margin-bottom: 2px;
  }
  
  .tweet-author-name {
    font-weight: bold;
    color: #000;
    text-decoration: none;
    display: flex;
    align-items: center;
    margin-right: 4px;
  }
  
  .tweet-author-name-dark {
    color: #fff;
  }
  
  .tweet-author-username {
    color: #657786;
    text-decoration: none;
    margin-right: 4px;
  }
  
  .tweet-author-username-dark {
    color: #8899a6;
  }
  
  .tweet-dot-separator {
    margin: 0 4px;
    color: #657786;
  }
  
  .tweet-dot-separator-dark {
    color: #8899a6;
  }
  
  .tweet-timestamp {
    color: #657786;
    font-size: 14px;
  }
  
  .tweet-timestamp-dark {
    color: #8899a6;
  }
  
  .display-name-text {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    max-width: 200px;
  }
  
  .user-verified-badge {
    color: #1da1f2;
    margin-left: 4px;
    display: flex;
    align-items: center;
  }
  
  .tweet-text {
    margin: 8px 0;
    word-wrap: break-word;
    color: #000;
    line-height: 1.4;
  }
  
  .tweet-text-dark {
    color: #fff;
  }
  
  .tweet-media-container {
    margin-top: 12px;
    border-radius: 16px;
    overflow: hidden;
    max-width: 100%;
    border: 1px solid #e1e8ed;
  }
  
  .tweet-media-container-dark {
    border-color: #38444d;
  }
  
  .tweet-media-single {
    width: 100%;
    max-height: 400px;
    overflow: hidden;
    border-radius: 16px;
  }
  
  .tweet-media-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    grid-gap: 2px;
    border-radius: 16px;
    overflow: hidden;
    max-height: 400px;
  }
  
  .tweet-media-item {
    position: relative;
    padding-bottom: 100%;
    overflow: hidden;
  }
  
  .tweet-media-img, .tweet-media-video {
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    object-fit: cover;
  }
  
  .tweet-actions {
    display: flex;
    justify-content: space-between;
    margin-top: 12px;
    max-width: 425px;
  }
  
  .tweet-actions-dark {
    color: #8899a6;
  }
  
  .tweet-action-item {
    display: flex;
    align-items: center;
  }
  
  .tweet-action-btn {
    background: none;
    border: none;
    padding: 8px;
    border-radius: 50%;
    cursor: pointer;
    display: flex;
    align-items: center;
    color: #657786;
    transition: all 0.2s;
  }
  
  .tweet-action-btn-dark {
    color: #8899a6;
  }
  
  .tweet-action-btn:hover {
    background-color: rgba(29, 161, 242, 0.1);
    color: #1da1f2;
  }
  
  .tweet-action-btn.active {
    color: #1da1f2;
  }
  
  .tweet-like-btn.active {
    color: #e0245e;
  }
  
  .tweet-action-icon {
    margin-right: 4px;
  }
  
  .tweet-action-count {
    font-size: 14px;
  }
  
  .tweet-card-error {
    border: 1px solid #e1e8ed;
    border-radius: 12px;
    padding: 16px;
    margin-bottom: 16px;
    background-color: #f8f9fa;
    color: #657786;
    text-align: center;
  }
  
  .tweet-card-error-dark {
    background-color: #1c2938;
    border-color: #38444d;
    color: #8899a6;
  }
  
  .is-reply {
    margin-left: 24px;
    border-left: 2px solid #e1e8ed;
  }
  
  .is-repost {
    border-color: #17bf63;
  }
  
  .has-replies {
    color: #1da1f2;
  }
  
  .tweet-reply-count-highlight {
    color: #1da1f2;
    font-weight: bold;
  }
  
  /* Dark mode hover effects */
  .tweet-card-dark .tweet-action-btn:hover {
    background-color: rgba(29, 161, 242, 0.1);
  }
  
  .tweet-card-dark .tweet-like-btn:hover {
    background-color: rgba(224, 36, 94, 0.1);
  }
  
  .heart-animation {
    animation: heart-burst 0.8s steps(28) forwards;
  }
  
  @keyframes heart-burst {
    from { transform: scale(1); }
    to { transform: scale(1.2); }
  }
  
  .loading::before {
    content: "";
    display: inline-block;
    width: 12px;
    height: 12px;
    border-radius: 50%;
    border: 2px solid currentColor;
    border-top-color: transparent;
    animation: spin 1s linear infinite;
    margin-right: 4px;
  }
  
  @keyframes spin {
    to { transform: rotate(360deg); }
  }
</style>