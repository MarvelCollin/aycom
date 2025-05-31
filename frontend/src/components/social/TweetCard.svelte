<script lang="ts">
  import { createEventDispatcher, onMount, onDestroy } from 'svelte';
  import { createLoggerWithPrefix } from '../../utils/logger';
  import { isAuthenticated as checkAuth } from '../../utils/auth';
  import { likeThread, unlikeThread, replyToThread, getReplyReplies, likeReply, unlikeReply, bookmarkThread, removeBookmark } from '../../api/thread';
  import { tweetInteractionStore } from '../../stores/tweetInteractionStore';
  import { notificationStore } from '../../stores/notificationStore';
  import { toastStore } from '../../stores/toastStore';
  import type { ITweet, IMedia } from '../../interfaces/ISocialMedia';
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

  // Extended tweet interface for additional properties that might be in the data
  interface ExtendedTweet extends ITweet {
    // Fields for retweets and bookmarks
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
    
    // Nested data
    bookmarked_thread?: any;
    bookmarked_reply?: any;
    parent_reply?: any;
    parent_thread?: any;
    isComment?: boolean;
    
    // User data
    user?: {
      id?: string;
      username?: string;
      name?: string;
      profile_picture_url?: string;
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
    };
  }
  
  // Initialize logger
  const logger = createLoggerWithPrefix('TweetCard');
  
  // Reactive tweet store setup
  export let tweet: ITweet | ExtendedTweet;
  export let isDarkMode: boolean = false;
  export let isAuth: boolean = false;
  
  export let isLiked: boolean = false;
  export let isReposted: boolean = false;
  export let isBookmarked: boolean = false;
  
  // Changed from export let to export const since it's only for external reference
  export const inReplyToTweet: ITweet | null = null;
  export let replies: ITweet[] = [];
  export let showReplies: boolean = false;
  export let nestingLevel: number = 0;
  const MAX_NESTING_LEVEL = 3;
  export let nestedRepliesMap: Map<string, ITweet[]> = new Map();
  
  const dispatch = createEventDispatcher();
  
  // Process the tweet and create a standardized version to work with
  $: processedTweet = processTweetContent(tweet);
  
  // Connect to the interaction store
  $: storeInteraction = $tweetInteractionStore?.get(processedTweet.id);
  
  // For interaction states, use the store if available, otherwise use the tweet's values
  $: effectiveIsLiked = storeInteraction?.is_liked ?? processedTweet.is_liked;
  $: effectiveIsReposted = storeInteraction?.is_reposted ?? processedTweet.is_reposted;
  $: effectiveIsBookmarked = storeInteraction?.is_bookmarked ?? processedTweet.is_bookmarked;
  
  // For count values, use the store if available, otherwise use the tweet's values
  $: effectiveLikes = storeInteraction?.likes ?? parseCount(processedTweet.likes_count);
  $: effectiveReplies = storeInteraction?.replies ?? parseCount(processedTweet.replies_count);
  $: effectiveReposts = storeInteraction?.reposts ?? parseCount(processedTweet.reposts_count);
  $: effectiveBookmarks = storeInteraction?.bookmarks ?? parseCount(processedTweet.bookmark_count);
  
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
  
  // Current request IDs to prevent race conditions
  let currentLikeRequestId = 0;
  let currentRepostRequestId = 0;
  let currentBookmarkRequestId = 0;
  
  // Update the conditional rendering logic to make the replies toggle more visible
  // First, modify how we determine if a tweet has replies
  $: hasReplies = effectiveReplies > 0 || parseCount(processedTweet.replies_count) > 0;
  
  $: processedReplies = replies.map(reply => processTweetContent(reply));
  
  // Subscribe to the tweet interaction store
  let storeInteraction: any = undefined;
  $: tweetId = typeof processedTweet.id === 'number' ? String(processedTweet.id) : processedTweet.id;
  
  // Initialize the tweet in the store on mount
  onMount(() => {
    if (processedTweet) {
      tweetInteractionStore.initTweet(processedTweet);
    }
  });

  // When processedTweet changes, make sure it's initialized in the store
  $: if (processedTweet && processedTweet.id) {
    tweetInteractionStore.initTweet(processedTweet);
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

  function processTweetContent(rawTweet: any): ExtendedTweet {
    if (!rawTweet) {
      console.error('Invalid tweet data provided to processTweetContent:', rawTweet);
      return createPlaceholderTweet();
    }
    
    try {
      // Make a deep copy to avoid modifying the original
      const processed: ExtendedTweet = {
        ...rawTweet,
        // Ensure all required ITweet fields exist with fallbacks
        id: rawTweet.id || rawTweet.thread_id || rawTweet.threadId || `unknown-${Date.now()}`,
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
        replies_count: safeParseNumber(rawTweet.replies_count || rawTweet.RepliesCount || rawTweet.ReplyCount || rawTweet.reply_count || rawTweet.metrics?.replies),
        reposts_count: safeParseNumber(rawTweet.reposts_count || rawTweet.RepostsCount || rawTweet.RepostCount || rawTweet.repost_count || rawTweet.metrics?.reposts),
        bookmark_count: safeParseNumber(rawTweet.bookmark_count || rawTweet.BookmarkCount || rawTweet.bookmarks_count || rawTweet.metrics?.bookmarks),
        
        // Interaction states with fallbacks
        is_liked: Boolean(rawTweet.is_liked || rawTweet.IsLiked || rawTweet.liked_by_user || rawTweet.LikedByUser || false),
        is_reposted: Boolean(rawTweet.is_reposted || rawTweet.IsReposted || rawTweet.reposted_by_user || rawTweet.RepostedByUser || rawTweet.is_repost || false),
        is_bookmarked: Boolean(rawTweet.is_bookmarked || rawTweet.IsBookmarked || rawTweet.bookmarked_by_user || rawTweet.BookmarkedByUser || false),
        is_pinned: Boolean(rawTweet.is_pinned || rawTweet.IsPinned || rawTweet.pinned || false),
        
        // Media with validation
        media: validateMedia(rawTweet.media || rawTweet.Media || []),
        
        // Other fields with fallbacks
        parent_id: rawTweet.parent_id || rawTweet.ParentID || rawTweet.parent_thread_id || null,
        community_id: rawTweet.community_id || rawTweet.CommunityID || null,
        community_name: rawTweet.community_name || rawTweet.CommunityName || null
      };
      
      return processed;
    } catch (error) {
      console.error('Error processing tweet content:', error, rawTweet);
      return createPlaceholderTweet();
    }
  }
  
  // Helper function to create a placeholder tweet when data is invalid
  function createPlaceholderTweet(): ExtendedTweet {
    return {
      id: `error-${Date.now()}`,
      content: 'Error loading content',
      created_at: new Date().toISOString(),
      user_id: '',
      username: 'error',
      name: 'Error Loading Data',
      profile_picture_url: 'https://secure.gravatar.com/avatar/0?d=mp',
      likes_count: 0,
      replies_count: 0,
      reposts_count: 0,
      bookmark_count: 0,
      is_liked: false,
      is_reposted: false,
      is_bookmarked: false,
      is_pinned: false,
      parent_id: null,
      media: []
    };
  }
  
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
  
  // Helper function to extract profile picture with fallbacks
  function extractProfilePicture(rawTweet: any): string {
    const picUrl = rawTweet.profile_picture_url || 
      rawTweet.ProfilePicture || 
      rawTweet.author_avatar || 
      rawTweet.avatar ||
      rawTweet.user?.profile_picture_url ||
      rawTweet.author?.profile_picture_url ||
      rawTweet.user_data?.profile_picture_url;
      
    if (!picUrl) return 'https://secure.gravatar.com/avatar/0?d=mp';
    
    // If already a full URL, return as is
    if (picUrl.startsWith('http')) return picUrl;
    
    // Try to construct a proper URL
    try {
      const supabaseUrl = import.meta.env.VITE_SUPABASE_URL || 'https://your-supabase-url.supabase.co';
      return `${supabaseUrl}/storage/v1/object/public/tpaweb/${picUrl}`;
    } catch (e) {
      return 'https://secure.gravatar.com/avatar/0?d=mp';
    }
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
    if (!Array.isArray(media)) {
      try {
        // Try to parse if it's a string
        if (typeof media === 'string') {
          const parsed = JSON.parse(media);
          if (Array.isArray(parsed)) return parsed;
        }
        return [];
      } catch (e) {
        return [];
      }
    }
    
    // Filter out invalid media items
    return media.filter(item => item && item.url).map(item => ({
      id: item.id || `media-${Date.now()}-${Math.random().toString(36).substring(2, 9)}`,
      url: item.url,
      type: item.type || 'image',
      thumbnail: item.thumbnail || item.url,
      alt: item.alt || 'Media attachment'
    }));
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

  function handleReply() {
    if (!checkAuth()) {
      toastStore.showToast('Please log in to reply to posts', 'info');
      return;
    }
    
    console.log(`Triggering reply for tweet: ${processedTweet.id}`);
    
    // Add visual feedback for the click
    const replyBtn = document.querySelector('.tweet-reply-btn');
    if (replyBtn) {
      replyBtn.classList.add('clicked');
      setTimeout(() => {
        replyBtn.classList.remove('clicked');
      }, 300);
    }
    
    dispatch('reply', safeToString(processedTweet.id));
  }

  async function handleRetweet() {
    if (!checkAuth()) {
      toastStore.showToast('Please log in to repost', 'info');
      return;
    }

    // Update the repost state through the store
    tweetInteractionStore.updateTweetInteraction(tweetId, {
      isReposted: !effectiveIsReposted,
      reposts: effectiveReposts + (!effectiveIsReposted ? 1 : -1),
      pendingRepost: true
    });
    dispatch('repost', tweetId);
  }

  // Handle like/unlike
  async function handleLikeClick() {
    if (!checkAuth()) {
      dispatch('like');
      return;
    }
    
    if (isLikeLoading) return;
    isLikeLoading = true;
    repliesErrorState = false;
    
    // Create a unique request ID to track this request
    const requestId = ++currentLikeRequestId;
    let isLikeRequestPending = true;
    
    try {
      // Optimistically update UI first
      const newStatus = !effectiveIsLiked;
      const newCount = newStatus ? effectiveLikes + 1 : effectiveLikes - 1;
      
      // Optimistically update the store
      tweetInteractionStore.updateTweetInteraction(tweetId, {
        is_liked: newStatus,
        likes: newCount,
        pending_like: true
      });
      
      // Make the API call
      const apiCall = newStatus ? likeThread : unlikeThread;
      await apiCall(tweetId);
      
      // Only update if this is still the current request
      if (requestId === currentLikeRequestId) {
        // Update the store with the final state
        tweetInteractionStore.updateTweetInteraction(tweetId, {
          is_liked: newStatus,
          likes: newCount,
          pending_like: false
        });
      }
    } catch (error) {
      console.error('Error toggling like:', error);
      
      if (requestId === currentLikeRequestId) {
        // Revert the optimistic update
          tweetInteractionStore.updateTweetInteraction(tweetId, {
          is_liked: !effectiveIsLiked,
          likes: effectiveIsLiked ? effectiveLikes + 1 : effectiveLikes - 1,
          pending_like: false
        });
        
          repliesErrorState = true;
      }
    } finally {
      if (requestId === currentLikeRequestId) {
        isLikeLoading = false;
        isLikeRequestPending = false;
      }
    }
  }

  // Handle bookmark toggle
  async function toggleBookmarkStatus(event: Event) {
    event.stopPropagation();
    
    if (!isAuthenticated) {
      toastStore.showToast('Please log in to bookmark tweets', 'error');
      return;
    }
    
    // Determine the current status
    const status = storeInteraction?.is_bookmarked || processedTweet.is_bookmarked || false;
    
    // Update the UI optimistically
    tweetInteractionStore.updateTweetInteraction(String(processedTweet.id), {
      is_bookmarked: !status,
      bookmarks: !status ? effectiveBookmarks + 1 : effectiveBookmarks - 1,
      pending_bookmark: true
    });
    
    try {
      if (!status) {
        await bookmarkThread(processedTweet.id);
        toastStore.showToast('Tweet bookmarked', 'success');
      } else {
        await removeBookmark(processedTweet.id);
        toastStore.showToast('Bookmark removed', 'success');
      }
      
      // Confirm the update was successful
      tweetInteractionStore.updateTweetInteraction(String(processedTweet.id), {
        pending_bookmark: false
      });
    } catch (error) {
      console.error('Error toggling bookmark:', error);
      toastStore.showToast('Failed to update bookmark', 'error');
          
      // Revert the optimistic update
      tweetInteractionStore.updateTweetInteraction(String(processedTweet.id), {
        is_bookmarked: status,
        bookmarks: status ? effectiveBookmarks + 1 : effectiveBookmarks - 1,
        pending_bookmark: false
      });
    }
  }

  function toggleReplies() {
    showReplies = !showReplies;
    
    if (showReplies && (!replies || replies.length === 0)) {
      console.log('Loading replies for tweet:', processedTweet.id);
      
      // Show loading state
      const replyContainer = document.getElementById(`replies-container-${tweetId}`);
      if (replyContainer) {
        replyContainer.classList.add('loading-replies');
      }
      
      // Dispatch event to load replies
      dispatch('loadReplies', safeToString(processedTweet.id));
      
      // Auto-load nested replies for all first-level replies when expanding
      if (replies && replies.length > 0 && nestingLevel === 0) {
        replies.forEach(async (reply) => {
          if (reply && reply.replies > 0) {
            try {
              // Ensure we have a string ID
              const replyId = safeToString(reply.id);
              const nestedRepliesData = await getReplyReplies(replyId);
              if (nestedRepliesData && nestedRepliesData.replies) {
                nestedRepliesMap.set(replyId, nestedRepliesData.replies.map(r => processTweetContent(r)));
                nestedRepliesMap = new Map(nestedRepliesMap);
              }
            } catch (error) {
              console.error(`Error pre-loading nested replies for ${reply.id}:`, error);
            }
          }
        });
      }
    } else {
      console.log('Hiding replies for tweet:', processedTweet.id);
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
    
    try {
      // Check if this is a reply ID - if so, we need to load replies to a reply
      if (nestedRepliesMap) {
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
          } catch (error) {
            console.error(`Error loading nested replies for reply ${replyId}:`, error);
            // Add a retry flag as an object to avoid type errors
            const retryKey = `retry_${replyId}`;
            const retryObject = { retry: true };
            nestedRepliesMap.set(retryKey, retryObject as any);
            // Set empty array as a fallback
            nestedRepliesMap.set(replyId, []);
            
            // Show user error message
            toastStore.showToast('Failed to load replies. Please try again.', 'error');
          } finally {
            // Remove loading state
            nestedRepliesMap.delete(loadingKey);
            // Force reactivity update
            nestedRepliesMap = new Map(nestedRepliesMap);
            
            // Remove loading class from reply container
            if (replyContainer) {
              replyContainer.classList.remove('loading-nested-replies');
            }
          }
        } else {
          console.log(`Already loading replies for ${replyId}, ignoring duplicate request`);
        }
      } else {
        // Just pass the event up to parent for thread replies
        dispatch('loadReplies', replyId);
      }
    } catch (error) {
      console.error(`Error in handleLoadNestedReplies for reply ${replyId}:`, error);
      toastStore.showToast('Failed to load replies. Please try again.', 'error');
      
      if (replyId && nestedRepliesMap) {
        // Remove loading state
        const loadingKey = `loading_${replyId}`;
        nestedRepliesMap.delete(loadingKey);
        
        // Set empty array for this reply's replies
        nestedRepliesMap.set(replyId, []);
        
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

      // Optimistic UI update
      reply.isLiked = true;
      if (typeof reply.likes === 'number') {
        reply.likes += 1;
      }
      
      // Call API
      try {
        await likeReply(String(replyId));
      } catch (error) {
        console.error('Error liking reply:', error);
        
        // Check for "already liked" error - don't revert UI
        const errorMsg = error instanceof Error ? error.message.toLowerCase() : '';
        const isAlreadyLiked = errorMsg.includes('already liked');
        
        if (!isAlreadyLiked) {
          // Revert optimistic update
          reply.isLiked = false;
          if (typeof reply.likes === 'number' && reply.likes > 0) {
            reply.likes -= 1;
          }
          toastStore.showToast('Failed to like reply. Please try again.', 'error');
        }
      } finally {
        // Clear loading state
        loadingState.like = false;
        replyActionsLoading.set(String(replyId), loadingState);
        replyActionsLoading = new Map(replyActionsLoading);
      }
    } catch (error) {
      console.error('Unhandled error in handleLikeReply:', error);
      toastStore.showToast('An unexpected error occurred. Please try again later.', 'error');
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
      replyActionsLoading = new Map(replyActionsLoading);

      // Optimistic UI update
      reply.isLiked = false;
      if (typeof reply.likes === 'number' && reply.likes > 0) {
        reply.likes -= 1;
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
          reply.isLiked = true;
          if (typeof reply.likes === 'number') {
            reply.likes += 1;
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
      reply.isBookmarked = true;
      if (typeof reply.bookmarks === 'number') {
        reply.bookmarks += 1;
      }
      
      // Call API
      try {
        await likeReply(String(replyId));
      } catch (error) {
        console.error('Error bookmarking reply:', error);
        
        // Check for "already bookmarked" error - don't revert UI
        const errorMsg = error instanceof Error ? error.message.toLowerCase() : '';
        const isAlreadyBookmarked = errorMsg.includes('already bookmarked');
        
        if (!isAlreadyBookmarked) {
          // Revert optimistic update
          reply.isBookmarked = false;
          if (typeof reply.bookmarks === 'number' && reply.bookmarks > 0) {
            reply.bookmarks -= 1;
          }
          toastStore.showToast('Failed to bookmark reply. Please try again.', 'error');
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
      reply.isBookmarked = false;
      if (typeof reply.bookmarks === 'number' && reply.bookmarks > 0) {
        reply.bookmarks -= 1;
      }
      
      // Call API
      try {
        await unlikeReply(String(replyId));
      } catch (error) {
        console.error('Error unbookmarking reply:', error);
        
        // Check for "not bookmarked" or "not found" error - don't revert UI
        const errorMsg = error instanceof Error ? error.message.toLowerCase() : '';
        const isNotBookmarked = errorMsg.includes('not bookmarked') || errorMsg.includes('not found');
        
        if (!isNotBookmarked) {
          // Revert optimistic update
          reply.isBookmarked = true;
          if (typeof reply.bookmarks === 'number') {
            reply.bookmarks += 1;
          }
          toastStore.showToast('Failed to remove bookmark from reply. Please try again.', 'error');
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
    
    // Get the userId from the processed tweet, ensuring it's a string
    // Try all possible ID field variants in order of preference
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
      window.location.href = `/user/${effectiveUserId}`;
      return;
    }
    
    // Otherwise fall back to username if it's valid
    if (username && username !== 'anonymous' && username !== 'user' && username !== 'unknown') {
      console.log(`✅ Falling back to username for navigation: ${username}`);
      window.location.href = `/user/${username}`;
    } else {
      console.error("❌ Navigation failed: No valid ID or username available", { 
        username, providedUserId: userId 
      });
    }
  }

  function debugUserData() {
    // This function can be called to inspect all tweets in the component
    console.group('🔍 TWEET DEBUGGING');
    console.log('Main tweet:', {
      id: tweet.id,
      thread_id: tweet.thread_id || tweet.threadId,
      username: {
        processed: processedTweet.username,
        original: tweet.username,
        authorUsername: tweet.authorUsername,
        author_username: tweet.author_username,
        fromThread: tweet.thread?.author?.username,
        fromUser: tweet.user?.username,
        fromAuthor: tweet.author?.username,
        fromUserData: tweet.user_data?.username,
        isValid: isValidUsername(processedTweet.username),
      },
      displayName: {
        processed: processedTweet.displayName,
        original: tweet.displayName,
        authorName: tweet.authorName, 
        name: tweet.name,
        display_name: tweet.display_name,
        fromThread: tweet.thread?.author?.name,
        fromUser: tweet.user?.name,
        fromAuthor: tweet.author?.name,
        fromUserData: tweet.user_data?.name,
        isValid: isValidDisplayName(processedTweet.displayName),
      },
      userId: {
        processed: processedTweet.userId || processedTweet.authorId || processedTweet.author_id || processedTweet.user_id,
        original: tweet.userId,
        authorId: tweet.authorId,
        author_id: tweet.author_id,
        user_id: tweet.user_id,
        fromThread: tweet.thread?.author?.id,
        fromUser: tweet.user?.id,
        fromAuthor: tweet.author?.id,
        fromUserData: tweet.user_data?.id,
      },
      // Cast to ExtendedTweet to access non-standard properties
      bookmarkedThread: (tweet as ExtendedTweet).bookmarked_thread || null,
      contenWithUsers: tweet.content && tweet.content.includes('@') ? 
        tweet.content.match(/@([a-zA-Z0-9_]+)/g) : null
    });
    
    if (replies.length > 0) {
      console.log('First reply:', {
        id: replies[0].id,
        thread_id: replies[0].thread_id || replies[0].threadId,
        username: {
          original: replies[0].username,
          authorUsername: replies[0].authorUsername,
          author_username: replies[0].author_username,
          fromThread: replies[0].thread?.author?.username,
          fromUser: replies[0].user?.username,
          fromAuthor: replies[0].author?.username,
          fromUserData: replies[0].user_data?.username
        },
        displayName: {
          original: replies[0].displayName,
          authorName: replies[0].authorName, 
          name: replies[0].name,
          display_name: replies[0].display_name,
          fromThread: replies[0].thread?.author?.name,
          fromUser: replies[0].user?.name,
          fromAuthor: replies[0].author?.name,
          fromUserData: replies[0].user_data?.name
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
        const replyCount = typeof oldParent.replies === 'number' ? 
          oldParent.replies + 1 : 
          Number(oldParent.replies || 0) + 1;
        
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

  // Create a copy of isAuthenticated from the parameter to avoid redeclaration
  $: isAuthenticated = isAuth;

  // Handle repost/unrepost
  async function handleRepostClick(event: Event) {
    event.stopPropagation();
    
    if (!checkAuth()) {
      toastStore.showToast('Please log in to repost', 'error');
      return;
    }
    
    const status = storeInteraction?.is_reposted || processedTweet.is_reposted || false;
    
    // Update UI optimistically
    tweetInteractionStore.updateTweetInteraction(String(processedTweet.id), {
      is_reposted: !status,
      reposts: !status ? effectiveReposts + 1 : effectiveReposts - 1,
      pending_repost: true
    });
    
    try {
      if (!status) {
        dispatch('repost');
      } else {
        dispatch('unrepost');
      }
      
      // Final update
      tweetInteractionStore.updateTweetInteraction(String(processedTweet.id), {
        pending_repost: false
      });
    } catch (error) {
      console.error('Error toggling repost:', error);
      toastStore.showToast('Failed to update repost', 'error');
      
      // Revert the optimistic update
      tweetInteractionStore.updateTweetInteraction(String(processedTweet.id), {
        is_reposted: status,
        reposts: status ? effectiveReposts + 1 : effectiveReposts - 1,
        pending_repost: false
      });
    }
  }
</script>

<div class="tweet-card {isDarkMode ? 'tweet-card-dark' : ''}">
  <div class="tweet-card-container">
    <div class="tweet-card-content">
      <div class="tweet-card-header">
        <a href={`/user/${processedTweet.userId || processedTweet.authorId || processedTweet.author_id || processedTweet.user_id || processedTweet.username}`}
          class="tweet-avatar"
          on:click|preventDefault={(e) => navigateToUserProfile(e, processedTweet.username, processedTweet.userId || processedTweet.authorId || processedTweet.author_id || processedTweet.user_id)}
          on:keydown={(e) => e.key === 'Enter' && navigateToUserProfile(e, processedTweet.username, processedTweet.userId || processedTweet.authorId || processedTweet.author_id || processedTweet.user_id)}>
          {#if typeof processedTweet.avatar === 'string' && processedTweet.avatar.startsWith('http')}
            <img src={processedTweet.avatar} alt={processedTweet.username} class="tweet-avatar-image" />
          {:else}
            <div class="tweet-avatar-placeholder">
              <div class="tweet-avatar-text">{processedTweet.username ? processedTweet.username[0].toUpperCase() : 'U'}</div>
            </div>
          {/if}
        </a>
        <div class="tweet-content-container">
          <div class="tweet-author-info">
            <a href={`/user/${processedTweet.userId || processedTweet.authorId || processedTweet.author_id || processedTweet.user_id || processedTweet.username}`}
              class="tweet-author-name {isDarkMode ? 'tweet-author-name-dark' : ''}"
              on:click|preventDefault={(e) => navigateToUserProfile(e, processedTweet.username, processedTweet.userId || processedTweet.authorId || processedTweet.author_id || processedTweet.user_id)}
              on:keydown={(e) => e.key === 'Enter' && navigateToUserProfile(e, processedTweet.username, processedTweet.userId || processedTweet.authorId || processedTweet.author_id || processedTweet.user_id)}>
              {#if processedTweet.displayName === 'User'}
                {console.warn('❌ MISSING DISPLAY NAME:', {id: processedTweet.id, tweetId: processedTweet.tweetId, username: processedTweet.username})}
                <span class="tweet-error-text">User</span>
              {:else}
                {processedTweet.displayName}
              {/if}
            </a>
            <a href={`/user/${processedTweet.userId || processedTweet.authorId || processedTweet.author_id || processedTweet.user_id || processedTweet.username}`}
              class="tweet-author-username {isDarkMode ? 'tweet-author-username-dark' : ''}"
              on:click|preventDefault={(e) => navigateToUserProfile(e, processedTweet.username, processedTweet.userId || processedTweet.authorId || processedTweet.author_id || processedTweet.user_id)}
              on:keydown={(e) => e.key === 'Enter' && navigateToUserProfile(e, processedTweet.username, processedTweet.userId || processedTweet.authorId || processedTweet.author_id || processedTweet.user_id)}>
              {#if processedTweet.username === 'user'}
                {console.warn('❌ MISSING USERNAME:', {id: processedTweet.id, tweetId: processedTweet.tweetId, displayName: processedTweet.displayName})}
                <span class="tweet-error-text">@user</span>
              {:else}
                @{processedTweet.username}
              {/if}
            </a>
            <span class="tweet-dot-separator {isDarkMode ? 'tweet-dot-separator-dark' : ''}">·</span>
            <span class="tweet-timestamp {isDarkMode ? 'tweet-timestamp-dark' : ''}">{formatTimeAgo(processedTweet.timestamp)}</span>
          </div>
          
          <div class="tweet-text {isDarkMode ? 'tweet-text-dark' : ''}">
            <p>{processedTweet.content || ''}</p>
          </div>
          
          {#if processedTweet.media && processedTweet.media.length > 0}
            <div class="tweet-media-container {isDarkMode ? 'tweet-media-container-dark' : ''}">
              {#if processedTweet.media.length === 1}
                <div class="tweet-media-single">
                  {#if processedTweet.media[0].type === 'Image'}
                    <img src={processedTweet.media[0].url} alt="Media" class="tweet-media-img" />
                  {:else if processedTweet.media[0].type === 'Video'}
                    <video src={processedTweet.media[0].url} controls class="tweet-media-video">
                      <track kind="captions" src="/captions/en.vtt" srclang="en" label="English" />
                    </video>
                  {:else}
                    <img src={processedTweet.media[0].url} alt="GIF" class="tweet-media-img" />
                  {/if}
                </div>
              {:else if processedTweet.media.length > 1}
                <div class="tweet-media-grid">
                  {#each processedTweet.media.slice(0, 4) as media, index (media.url || index)}
                    <div class="tweet-media-item">
                      {#if media.type === 'Image'}
                        <img src={media.url} alt="Media" class="tweet-media-img" />
                      {:else if media.type === 'Video'}
                        <video src={media.url} class="tweet-media-video">
                          <track kind="captions" src="/captions/en.vtt" srclang="en" label="English" />
                        </video>
                      {:else}
                        <img src={media.url} alt="GIF" class="tweet-media-img" />
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
                <span class="tweet-action-count {hasReplies ? 'tweet-reply-count-highlight' : ''}">{effectiveReplies}</span>
              </button>
              {#if !showReplies}
                <button 
                  class="view-replies-btn" 
                  on:click|stopPropagation={toggleReplies}
                  aria-label="View all replies"
                >
                  {#if hasReplies}
                    View {effectiveReplies} {effectiveReplies === 1 ? 'reply' : 'replies'}
                  {:else}
                    View replies
                  {/if}
                </button>
              {/if}
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
                class="tweet-action-btn tweet-like-btn {effectiveIsLiked ? 'active' : ''} {isLikeLoading ? 'loading' : ''} {isDarkMode ? 'tweet-action-btn-dark' : ''}" 
                on:click|stopPropagation={handleLikeClick}
                aria-label="{effectiveIsLiked ? 'Unlike' : 'Like'}"
                disabled={isLikeLoading}
              >
                {#if isLikeLoading}
                  <div class="tweet-action-loading"></div>
                  <HeartIcon size="20" fill={effectiveIsLiked ? "currentColor" : "none"} class="tweet-action-icon hidden" />
                {:else}
                  <HeartIcon size="20" fill={effectiveIsLiked ? "currentColor" : "none"} class="tweet-action-icon" />
                {/if}
                <span class="tweet-action-count">{effectiveLikes}</span>
                <span class="like-status-text">{effectiveIsLiked ? 'Liked' : 'Like'}</span>
              </button>
            </div>
            <div class="tweet-action-item">
              <button 
                class="tweet-action-btn tweet-bookmark-btn {effectiveIsBookmarked ? 'active' : ''} {isBookmarkLoading ? 'loading' : ''} {isDarkMode ? 'tweet-action-btn-dark' : ''}" 
                on:click|stopPropagation={toggleBookmarkStatus}
                aria-label="{effectiveIsBookmarked ? 'Remove bookmark' : 'Bookmark'}"
                disabled={isBookmarkLoading}
              >
                {#if isBookmarkLoading}
                  <div class="tweet-action-loading"></div>
                  <BookmarkIcon size="20" fill={effectiveIsBookmarked ? "currentColor" : "none"} class="tweet-action-icon hidden" />
                {:else}
                  <BookmarkIcon size="20" fill={effectiveIsBookmarked ? "currentColor" : "none"} class="tweet-action-icon" />
                {/if}
                <span class="tweet-action-count">{effectiveBookmarks}</span>
              </button>
            </div>
            <div class="tweet-action-item">
              <div class="tweet-views-count {isDarkMode ? 'tweet-views-count-dark' : ''}">
                <EyeIcon size="20" class="tweet-action-icon" />
                <span class="tweet-action-count">{processedTweet.views || '0'}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</div>

{#if nestingLevel === 0}
  <div class="tweet-replies-toggle-container">
    <button 
      class="tweet-replies-toggle {isDarkMode ? 'tweet-replies-toggle-dark' : ''}"
      on:click|stopPropagation={toggleReplies}
      aria-expanded={showReplies}
      aria-controls="replies-container-{tweetId}"
    >
      {#if showReplies}
        <ChevronUpIcon size="16" class="tweet-replies-toggle-icon" />
        Hide replies
      {:else}
        <ChevronDownIcon size="16" class="tweet-replies-toggle-icon" />
        {#if hasReplies}
          View replies ({effectiveReplies})
        {:else}
          View replies
        {/if}
      {/if}
    </button>
  </div>
{/if}

{#if showReplies}
  <div id="replies-container-{tweetId}" class="tweet-replies-container {isDarkMode ? 'tweet-replies-container-dark' : ''}">
    {#if replies.length === 0}
      <div class="tweet-replies-empty {isDarkMode ? 'tweet-replies-empty-dark' : ''}">
        <div class="tweet-replies-empty-icon">
          <MessageCircleIcon size="20" />
        </div>
        <div class="tweet-replies-empty-text">
          No replies yet. Be the first to reply!
        </div>
        <button 
          class="tweet-replies-empty-btn" 
          on:click|stopPropagation={handleReply}
        >
          Reply
        </button>
      </div>
    {:else}
      {#if replies.length > 0 && nestingLevel < MAX_NESTING_LEVEL}
        {#each processedReplies as reply (reply.id || reply.tweetId)}
          <div id="reply-{reply.id}-container" class="nested-reply-container">
            <svelte:self 
              tweet={reply}
              {isDarkMode}
              {isAuthenticated}
              isLiked={reply.isLiked || reply.is_liked || false}
              isReposted={reply.isReposted || false}
              isBookmarked={reply.isBookmarked || false}
              inReplyToTweet={null}
              replies={nestedRepliesMap.get(String(reply.id)) || []} 
              showReplies={false}
              nestingLevel={nestingLevel + 1}
              {nestedRepliesMap}
              on:reply={handleNestedReply}
              on:like={handleNestedLike}
              on:unlike={handleNestedLike}
              on:repost={handleNestedRepost}
              on:bookmark={handleNestedBookmark}
              on:removeBookmark={handleNestedBookmark}
              on:loadReplies={handleLoadNestedReplies}
            />
            
            {#if reply.replies > 0}
              {#if nestedRepliesMap.has(`retry_${reply.id}`)}
                <!-- Show retry button when loading failed -->
                <div class="nested-replies-retry-container">
                  <button 
                    class="nested-replies-retry-btn" 
                    on:click|stopPropagation={() => retryLoadNestedReplies(reply.id)}
                  >
                    <RefreshCwIcon size="14" />
                    Failed to load replies. Retry?
                  </button>
                </div>
              {:else if !nestedRepliesMap.has(String(reply.id)) && reply.replies > 0}
                <!-- Show view replies button when not loaded yet -->
                <div class="nested-replies-view-container">
                  <button 
                    class="nested-replies-view-btn" 
                    on:click|stopPropagation={() => handleLoadNestedReplies({ detail: String(reply.id) })}
                  >
                    <ChevronDownIcon size="14" />
                    View {reply.replies} {reply.replies === 1 ? 'reply' : 'replies'}
                  </button>
                </div>
              {/if}
            {/if}
          </div>
        {/each}
      {:else if replies.length > 0}
        {#each processedReplies as reply, index (reply.id || reply.tweetId || `reply-${reply.timestamp}-${reply.username}-${index}`)}
          <div class="tweet-reply-item {isDarkMode ? 'tweet-reply-item-dark' : ''}">
            <div class="tweet-reply-content">
              <a 
                href={`/user/${reply.userId || reply.authorId || reply.author_id || reply.user_id || reply.username}`}
                class="tweet-reply-avatar"
                on:click|preventDefault={(e) => navigateToUserProfile(e, reply.username, reply.userId || reply.authorId || reply.author_id || reply.user_id)}
                on:keydown={(e) => e.key === 'Enter' && navigateToUserProfile(e, reply.username, reply.userId || reply.authorId || reply.author_id || reply.user_id)}
              >
                {#if typeof reply.avatar === 'string' && reply.avatar.startsWith('http')}
                  <img src={reply.avatar} alt={reply.username} class="tweet-reply-avatar-img" />
                {:else}
                  <div class="tweet-reply-avatar-placeholder">{reply.username ? reply.username.charAt(0).toUpperCase() : 'U'}</div>
                {/if}
              </a>
              <div class="tweet-reply-body">
                <div class="tweet-reply-author">
                  <a 
                    href={`/user/${reply.userId || reply.authorId || reply.author_id || reply.user_id || reply.username}`}
                    class="tweet-reply-author-name {isDarkMode ? 'tweet-reply-author-name-dark' : ''}"
                    on:click|preventDefault={(e) => navigateToUserProfile(e, reply.username, reply.userId || reply.authorId || reply.author_id || reply.user_id)}
                    on:keydown={(e) => e.key === 'Enter' && navigateToUserProfile(e, reply.username, reply.userId || reply.authorId || reply.author_id || reply.user_id)}
                  >{reply.displayName || 'User'}</a>
                  <a 
                    href={`/user/${reply.userId || reply.authorId || reply.author_id || reply.user_id || reply.username}`}
                    class="tweet-reply-author-username {isDarkMode ? 'tweet-reply-author-username-dark' : ''}"
                    on:click|preventDefault={(e) => navigateToUserProfile(e, reply.username, reply.userId || reply.authorId || reply.author_id || reply.user_id)}
                    on:keydown={(e) => e.key === 'Enter' && navigateToUserProfile(e, reply.username, reply.userId || reply.authorId || reply.author_id || reply.user_id)}
                  >@{reply.username || 'user'}</a>
                  <span class="tweet-reply-dot-separator {isDarkMode ? 'tweet-reply-dot-separator-dark' : ''}">·</span>
                  <span class="tweet-reply-timestamp {isDarkMode ? 'tweet-reply-timestamp-dark' : ''}">{formatTimeAgo(reply.timestamp)}</span>
                </div>
                <div class="tweet-reply-text {isDarkMode ? 'tweet-reply-text-dark' : ''}">
                  <p>{reply.content || 'No content available'}</p>
                  
                  {#if reply.media && reply.media.length > 0}
                    <div class="tweet-reply-media">
                      <img src={reply.media[0].url} alt="Media" class="tweet-reply-media-img" />
                    </div>
                  {/if}
                </div>
                <div class="tweet-reply-actions {isDarkMode ? 'tweet-reply-actions-dark' : ''}">
                  <button class="tweet-reply-action-btn tweet-reply-reply-btn {isDarkMode ? 'tweet-reply-action-btn-dark' : ''}" on:click|stopPropagation={() => dispatch('reply', reply.id)}>
                    <MessageCircleIcon size="16" class="tweet-reply-action-icon" />
                    <span>Reply</span>
                  </button>
                  
                  <button class="tweet-reply-action-btn tweet-reply-like-btn {reply.isLiked ? 'active' : ''} {(replyActionsLoading.get(String(reply.id))?.like) ? 'loading' : ''} {isDarkMode ? 'tweet-reply-action-btn-dark' : ''}" 
                    on:click|stopPropagation={(e) => {
                      e.preventDefault();
                      reply.isLiked ? handleUnlikeReply(reply.id) : handleLikeReply(reply.id);
                    }}>
                    {#if replyActionsLoading.get(String(reply.id))?.like}
                      <div class="tweet-reply-action-loading"></div>
                    {:else if reply.isLiked}
                      <HeartIcon size="16" fill="currentColor" class="tweet-reply-action-icon" />
                    {:else}
                      <HeartIcon size="16" class="tweet-reply-action-icon" />
                    {/if}
                    <span>Like</span>
                  </button>
                  
                  <button class="tweet-reply-action-btn tweet-reply-bookmark-btn {reply.isBookmarked ? 'active' : ''} {(replyActionsLoading.get(String(reply.id))?.bookmark) ? 'loading' : ''} {isDarkMode ? 'tweet-reply-action-btn-dark' : ''}" 
                    on:click|stopPropagation={(e) => {
                      e.preventDefault();
                      reply.isBookmarked ? handleUnbookmarkReply(reply.id) : handleBookmarkReply(reply.id);
                    }}>
                    {#if replyActionsLoading.get(String(reply.id))?.bookmark}
                      <div class="tweet-reply-action-loading"></div>
                    {:else if reply.isBookmarked}
                      <BookmarkIcon size="16" fill="currentColor" class="tweet-reply-action-icon" />
                    {:else}
                      <BookmarkIcon size="16" class="tweet-reply-action-icon" />
                    {/if}
                    <span>Save</span>
                  </button>
                </div>
              </div>
            </div>
          </div>
        {/each}
      {/if}
    {/if}
  </div>
{/if}

<style>
  /* Note: Some CSS selectors may appear unused but are needed for dynamic class creation
     or are used by JavaScript functions like classList.add/remove */
  .tweet-card {
    padding: 0.5rem 0;
    border-bottom: 1px solid var(--border-color);
    background-color: var(--bg-primary);
    color: var(--text-primary);
    transition: background-color var(--transition-fast);
  }

  .tweet-card-dark {
    background-color: var(--bg-primary-dark);
    color: var(--text-primary-dark);
    border-bottom: 1px solid var(--border-color-dark);
  }

    /* Tweet avatar container is not used in the template */
  
  .tweet-actions {
    display: flex;
    justify-content: space-between;
    padding: 0.5rem 0;
    margin-top: 0.5rem;
  }
  
  .tweet-actions-dark {
    color: var(--text-secondary-dark);
  }

  .tweet-action-item {
    display: flex;
    align-items: center;
  }

  .tweet-action-btn {
    display: flex;
    align-items: center;
    padding: 0.5rem;
    border-radius: 9999px;
    transition: all var(--transition-fast);
    cursor: pointer;
    background-color: transparent;
    border: none;
    color: var(--text-secondary);
    position: relative;
    min-width: 65px;
  }

  .tweet-action-btn:hover {
    background-color: rgba(var(--color-primary-rgb), 0.1);
    color: var(--color-primary);
  }

  .tweet-action-btn.active {
    color: var(--color-primary);
  }

  .tweet-action-btn.loading {
    pointer-events: none;
  }

  .tweet-action-loading {
    position: absolute;
    left: 0.5rem;
    width: 20px;
    height: 20px;
    border: 2px solid rgba(var(--color-primary-rgb), 0.3);
    border-radius: 50%;
    border-top-color: var(--color-primary);
    animation: spin 0.8s linear infinite;
  }

  .tweet-action-icon {
    margin-right: 0.25rem;
    flex-shrink: 0;
  }
  
  .tweet-action-icon.hidden {
    opacity: 0;
  }

  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }

  .tweet-action-count {
    font-size: 0.875rem;
  }
  
  .like-status-text {
    font-size: 0;
    width: 0;
    height: 0;
    overflow: hidden;
    position: absolute;
    left: -9999px;
  }
  
  /* Only show the like status on mobile and tablets */
  @media (max-width: 768px) {
    .like-status-text {
      font-size: 0.75rem;
      width: auto;
      height: auto;
      position: static;
      margin-left: 0.25rem;
      overflow: visible;
      display: none;
    }
    
    .tweet-like-btn.active .like-status-text {
      display: inline;
      color: var(--color-primary);
      font-weight: 500;
    }
  }

  .view-replies-btn {
    background-color: var(--hover-primary);
    border: 1px solid var(--color-primary-light);
    color: var(--color-primary);
    font-size: 0.9rem;
    margin-left: 0.75rem;
    cursor: pointer;
    padding: 0.5rem 0.75rem;
    border-radius: var(--radius-full);
    transition: all 0.2s;
    display: flex;
    align-items: center;
    gap: 0.25rem;
    font-weight: 600;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  }

  .view-replies-btn:hover {
    background-color: var(--color-primary-light);
    transform: translateY(-2px);
    box-shadow: 0 3px 6px rgba(0, 0, 0, 0.15);
  }

  .view-replies-btn:active {
    transform: translateY(0);
    box-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
  }

  .tweet-replies-toggle {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 0.5rem;
    background-color: var(--hover-primary);
    border: none;
    color: var(--color-primary);
    font-weight: 600;
    padding: 0.75rem;
    cursor: pointer;
    width: 100%;
    border-radius: var(--radius-md);
    margin-top: 0.5rem;
    transition: all 0.2s;
    border: 1px solid var(--border-color);
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  }

  .tweet-replies-toggle:hover {
    background-color: var(--color-primary-light);
    transform: translateY(-1px);
    box-shadow: 0 2px 5px rgba(0, 0, 0, 0.15);
  }

  .tweet-replies-toggle-dark {
    background-color: var(--bg-secondary-dark);
    border-color: var(--border-color-dark);
  }

  .loading-replies {
    position: relative;
  }

  .loading-replies::after {
    content: "";
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 5px;
    background: linear-gradient(
      90deg,
      transparent,
      var(--color-primary),
      transparent
    );
    animation: loading-animation 1.5s infinite;
  }
  
  .loading-nested-replies {
    position: relative;
    opacity: 0.8;
  }
  
  .loading-nested-replies::before {
    content: "";
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background-color: rgba(255, 255, 255, 0.1);
    z-index: 1;
  }
  
  .loading-nested-replies::after {
    content: "";
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 3px;
    background: linear-gradient(
      90deg,
      transparent,
      var(--color-primary),
      transparent
    );
    animation: loading-animation 1.5s infinite;
    z-index: 2;
  }

  @keyframes loading-animation {
    0% { transform: translateX(-100%); }
    100% { transform: translateX(100%); }
  }

  .tweet-reply-count-highlight {
    color: var(--color-primary);
    font-weight: 600;
  }

  .tweet-action-btn.clicked {
    transform: scale(1.1);
    transition: transform 0.2s;
  }
  
  .tweet-replies-toggle-icon {
    transition: transform 0.2s;
  }
  
  /* Used in the template */
  .has-replies {
    font-weight: 500;
  }
  
  /* Additional styles for tweet reply action buttons */
  .tweet-reply-action-btn {
    background-color: var(--bg-secondary);
    color: var(--text-secondary);
    border: none;
    border-radius: var(--radius-full);
    padding: 0.375rem 0.625rem;
    display: flex;
    align-items: center;
    gap: 4px;
    cursor: pointer;
    transition: all 0.2s;
  }

  .tweet-reply-action-btn-dark {
    background-color: var(--bg-secondary-dark);
    color: var(--text-secondary-dark);
  }

  .tweet-reply-action-btn:hover {
    background-color: var(--hover-primary);
    color: var(--color-primary);
  }

  .tweet-reply-action-btn-dark:hover {
    background-color: var(--hover-primary);
    color: var(--color-primary);
  }

  .tweet-reply-action-btn.active {
    color: var(--color-primary);
  }

  .tweet-reply-action-btn-dark.active {
    color: var(--color-primary);
  }

  .tweet-replies-empty {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 2rem 1rem;
    text-align: center;
    background-color: var(--bg-secondary);
    border-radius: var(--radius-md);
    margin: 1rem 0;
  }
  
  .tweet-replies-empty-dark {
    background-color: var(--bg-secondary-dark);
  }
  
  .tweet-replies-empty-icon {
    margin-bottom: 0.5rem;
    color: var(--text-secondary);
  }
  
  .tweet-replies-empty-text {
    margin-bottom: 1rem;
    color: var(--text-secondary);
    font-size: 0.95rem;
  }
  
  .tweet-replies-empty-btn {
    background-color: var(--color-primary);
    color: white;
    border: none;
    padding: 0.5rem 1.25rem;
    border-radius: var(--radius-full);
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s;
  }
  
  .tweet-replies-empty-btn:hover {
    background-color: var(--color-primary-hover);
    transform: translateY(-1px);
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  }

  .nested-reply-container {
    margin-left: 1rem;
    border-left: 2px solid var(--border-color);
    padding-left: 1rem;
    position: relative;
  }
  
  .nested-replies-retry-container,
  .nested-replies-view-container {
    margin-top: 0.5rem;
    margin-bottom: 1rem;
    padding-left: 3rem;
  }
  
  .nested-replies-retry-btn {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    background-color: rgba(255, 193, 7, 0.1);
    color: #e6a700;
    border: 1px solid rgba(255, 193, 7, 0.2);
    padding: 0.5rem 1rem;
    border-radius: 9999px;
    font-size: 0.875rem;
    cursor: pointer;
    transition: all 0.2s;
  }
  
  .nested-replies-retry-btn:hover {
    background-color: rgba(255, 193, 7, 0.2);
    transform: translateY(-1px);
  }
  
  .nested-replies-view-btn {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    background-color: rgba(29, 155, 240, 0.1);
    color: #1d9bf0;
    border: 1px solid rgba(29, 155, 240, 0.2);
    padding: 0.5rem 1rem;
    border-radius: 9999px;
    font-size: 0.875rem;
    cursor: pointer;
    transition: all 0.2s;
  }
  
  .nested-replies-view-btn:hover {
    background-color: rgba(29, 155, 240, 0.2);
    transform: translateY(-1px);
  }

  /* Add to the existing styles */
  .tweet-reply-action-btn.loading {
    pointer-events: none;
    opacity: 0.8;
  }
  
  .tweet-reply-action-loading {
    width: 16px;
    height: 16px;
    border: 2px solid rgba(var(--color-primary-rgb), 0.3);
    border-radius: 50%;
    border-top-color: var(--color-primary);
    animation: spin 0.8s linear infinite;
    margin-right: 0.25rem;
  }

  .tweet-reply-avatar-placeholder {
    width: 32px;
    height: 32px;
    border-radius: 50%;
    background-color: var(--color-primary-light);
    color: var(--color-primary);
    display: flex;
    align-items: center;
    justify-content: center;
    font-weight: bold;
    font-size: 14px;
  }
</style>