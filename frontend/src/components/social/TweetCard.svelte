<script lang="ts">
  import { createEventDispatcher, onMount } from 'svelte';
  import type { ITweet } from '../../interfaces/ISocialMedia.d.ts';
  import { toastStore } from '../../stores/toastStore';
  import { tweetInteractionStore } from '../../stores/tweetStore';
  import { formatTimeAgo, processUserMetadata } from '../../utils/common';
  import { likeThread, unlikeThread, bookmarkThread, removeBookmark, likeReply, unlikeReply, bookmarkReply, removeReplyBookmark, getReplyReplies } from '../../api/thread';
  
  import MessageCircleIcon from 'svelte-feather-icons/src/icons/MessageCircleIcon.svelte';
  import RefreshCwIcon from 'svelte-feather-icons/src/icons/RefreshCwIcon.svelte';
  import HeartIcon from 'svelte-feather-icons/src/icons/HeartIcon.svelte';
  import BookmarkIcon from 'svelte-feather-icons/src/icons/BookmarkIcon.svelte';
  import EyeIcon from 'svelte-feather-icons/src/icons/EyeIcon.svelte';
  import ChevronUpIcon from 'svelte-feather-icons/src/icons/ChevronUpIcon.svelte';
  import ChevronDownIcon from 'svelte-feather-icons/src/icons/ChevronDownIcon.svelte';
  import ChevronRightIcon from 'svelte-feather-icons/src/icons/ChevronRightIcon.svelte';
  import ArrowRightIcon from 'svelte-feather-icons/src/icons/ArrowRightIcon.svelte';
  import XIcon from 'svelte-feather-icons/src/icons/XIcon.svelte';
  import ChevronLeftIcon from 'svelte-feather-icons/src/icons/ChevronLeftIcon.svelte';
  import CornerUpRightIcon from 'svelte-feather-icons/src/icons/CornerUpRightIcon.svelte';

  // Define a TypeScript type for a tweet with extended properties
  type ExtendedTweet = ITweet & { 
    bookmarked_thread?: any; 
    [key: string]: any; 
  }

  export let tweet: ITweet;
  export let isDarkMode: boolean = false;
  export let isAuthenticated: boolean = false;
  
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
  
  $: processedTweet = processTweetContent(tweet);
  
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
        storeInteraction = store.interactions.get(tweetId);
      });
    }
  }

  // Use store values for interaction counts and status if available
  $: effectiveIsLiked = storeInteraction?.isLiked ?? isLiked;
  $: effectiveIsReposted = storeInteraction?.isReposted ?? isReposted;
  $: effectiveIsBookmarked = storeInteraction?.isBookmarked ?? isBookmarked;
  
  // For count values, use the store if available, otherwise use the tweet's values
  $: effectiveLikes = storeInteraction?.likes ?? parseCount(processedTweet.likes);
  $: effectiveReplies = storeInteraction?.replies ?? parseCount(processedTweet.replies);
  $: effectiveReposts = storeInteraction?.reposts ?? parseCount(processedTweet.reposts);
  $: effectiveBookmarks = storeInteraction?.bookmarks ?? parseCount(processedTweet.bookmarks);
  
  // Track loading states for interaction buttons
  let isLikeLoading = false;
  let isBookmarkLoading = false;
  let isRepostLoading = false;
  
  // Track error states for retry functionality
  let likeErrorState = false;
  let bookmarkErrorState = false;
  
  // Request tracking to prevent race conditions
  let currentLikeRequestId = 0;
  let currentBookmarkRequestId = 0;
  let isLikeRequestPending = false;
  
  // Update the conditional rendering logic to make the replies toggle more visible
  // First, modify how we determine if a tweet has replies
  $: hasReplies = effectiveReplies > 0 || parseCount(processedTweet.replies) > 0;
  
  // Track loading and error states for reply actions
  let replyActionsLoading = new Map<string, { like?: boolean, bookmark?: boolean }>();
  
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

  function processTweetContent(originalTweet: ITweet): ITweet {
    console.log('Processing tweet content for tweet:', originalTweet.id, {
      rawTweet: originalTweet,
      hasUsername: !!originalTweet.username,
      hasDisplayName: !!originalTweet.displayName,
      hasAuthorId: !!originalTweet.authorId,
      hasUserId: !!originalTweet.userId,
      threadInfo: originalTweet.thread || originalTweet.threadInfo,
      author_id: originalTweet.author_id,
      user_id: originalTweet.user_id
    });
    
    const processedTweet = { ...originalTweet };
    
    // Clean up any "no reply" text that might be in the username field
    if (processedTweet.username === 'no reply') {
      processedTweet.username = '';
    }
    
    // Store the original ID as tweetId if not already set
    if (processedTweet.id && !processedTweet.tweetId) {
      processedTweet.tweetId = String(processedTweet.id);
    }
    
    // When tweet comes from a bookmarked thread, extract the actual author information
    if (processedTweet.thread_id && !isValidUsername(processedTweet.username)) {
      console.log(`[${processedTweet.id}] This appears to be a bookmarked tweet with thread_id`, processedTweet.thread_id);
    }
    
    // Extract author information from all possible locations
    
    // Check for embedded author object
    if (processedTweet.author && typeof processedTweet.author === 'object') {
      console.log(`[${processedTweet.id}] Found author object`, processedTweet.author);
      if (isValidUsername(processedTweet.author.username) && !isValidUsername(processedTweet.username)) {
        console.log(`[${processedTweet.id}] Using author.username:`, processedTweet.author.username);
        processedTweet.username = String(processedTweet.author.username || '');
      }
      if (isValidDisplayName(processedTweet.author.name) && !isValidDisplayName(processedTweet.displayName)) {
        console.log(`[${processedTweet.id}] Using author.name:`, processedTweet.author.name);
        processedTweet.displayName = String(processedTweet.author.name || '');
      }
    }
    
    // Check if thread or reply object contains author information
    if (processedTweet.thread && typeof processedTweet.thread === 'object') {
      console.log(`[${processedTweet.id}] Found thread object`, processedTweet.thread);
      if (processedTweet.thread.author) {
        const threadAuthor = processedTweet.thread.author;
        // Extract from thread author if available
        if (threadAuthor.username && !processedTweet.username) {
          console.log(`[${processedTweet.id}] Using thread.author.username:`, threadAuthor.username);
          processedTweet.username = threadAuthor.username;
        }
        if (threadAuthor.name && !processedTweet.displayName) {
          console.log(`[${processedTweet.id}] Using thread.author.name:`, threadAuthor.name);
          processedTweet.displayName = threadAuthor.name;
        }
        if (threadAuthor.id && !processedTweet.userId) {
          console.log(`[${processedTweet.id}] Using thread.author.id:`, threadAuthor.id);
          processedTweet.userId = threadAuthor.id;
        }
        if (threadAuthor.profile_picture_url && !processedTweet.avatar) {
          console.log(`[${processedTweet.id}] Using thread.author.profile_picture_url`);
          processedTweet.avatar = threadAuthor.profile_picture_url;
        }
      }
    }
    
    // Check if user or author object contains information
    if (processedTweet.user && typeof processedTweet.user === 'object') {
      console.log(`[${processedTweet.id}] Found user object`, processedTweet.user);
      const user = processedTweet.user;
      // Extract from user if available
      if (user.username && !processedTweet.username) {
        console.log(`[${processedTweet.id}] Using user.username:`, user.username);
        processedTweet.username = user.username;
      }
      if (user.name && !processedTweet.displayName) {
        console.log(`[${processedTweet.id}] Using user.name:`, user.name);
        processedTweet.displayName = user.name;
      }
      if (user.id && !processedTweet.userId) {
        console.log(`[${processedTweet.id}] Using user.id:`, user.id);
        processedTweet.userId = user.id;
      }
      if (user.profile_picture_url && !processedTweet.avatar) {
        console.log(`[${processedTweet.id}] Using user.profile_picture_url`);
        processedTweet.avatar = user.profile_picture_url;
      }
    }
    
    // Handle user ID fields - use snake_case variants too
    if (processedTweet.authorId && !processedTweet.userId) {
      console.log(`[${processedTweet.id}] Using authorId as userId:`, processedTweet.authorId);
      processedTweet.userId = processedTweet.authorId;
    } else if (processedTweet.author_id && !processedTweet.userId) {
      console.log(`[${processedTweet.id}] Using author_id as userId:`, processedTweet.author_id);
      processedTweet.userId = processedTweet.author_id;
    } else if (processedTweet.user_id && !processedTweet.userId) {
      console.log(`[${processedTweet.id}] Using user_id as userId:`, processedTweet.user_id);
      processedTweet.userId = processedTweet.user_id;
    }
    
    // Handle username fields - try camelCase and snake_case
    if (processedTweet.authorUsername && !processedTweet.username) {
      console.log(`[${processedTweet.id}] Using authorUsername as username:`, processedTweet.authorUsername);
      processedTweet.username = processedTweet.authorUsername;
    } else if (processedTweet.author_username && !processedTweet.username) {
      console.log(`[${processedTweet.id}] Using author_username as username:`, processedTweet.author_username);
      processedTweet.username = processedTweet.author_username;
    }
    
    // Handle display name fields - try all variants with different cases
    if (processedTweet.authorName && !processedTweet.displayName) {
      console.log(`[${processedTweet.id}] Using authorName as displayName:`, processedTweet.authorName);
      processedTweet.displayName = processedTweet.authorName;
    } else if (processedTweet.author_name && !processedTweet.displayName) {
      console.log(`[${processedTweet.id}] Using author_name as displayName:`, processedTweet.author_name);
      processedTweet.displayName = processedTweet.author_name;
    } else if (processedTweet.name && !processedTweet.displayName) {
      console.log(`[${processedTweet.id}] Using name as displayName:`, processedTweet.name);
      processedTweet.displayName = processedTweet.name;
    } else if (processedTweet.display_name && !processedTweet.displayName) {
      console.log(`[${processedTweet.id}] Using display_name as displayName:`, processedTweet.display_name);
      processedTweet.displayName = processedTweet.display_name;
    }
    
    // Handle avatar/profile picture - try all variants
    if (processedTweet.authorAvatar && !processedTweet.avatar) {
      console.log(`[${processedTweet.id}] Using authorAvatar as avatar`);
      processedTweet.avatar = processedTweet.authorAvatar;
    } else if (processedTweet.author_avatar && !processedTweet.avatar) {
      console.log(`[${processedTweet.id}] Using author_avatar as avatar`);
      processedTweet.avatar = processedTweet.author_avatar;
    } else if (processedTweet.profile_picture_url && !processedTweet.avatar) {
      console.log(`[${processedTweet.id}] Using profile_picture_url as avatar:`, processedTweet.profile_picture_url);
      processedTweet.avatar = processedTweet.profile_picture_url;
    }
    
    // If timestamp is in created_at but not in timestamp field
    if (processedTweet.created_at && !processedTweet.timestamp) {
      console.log(`[${processedTweet.id}] Using created_at as timestamp`);
      processedTweet.timestamp = processedTweet.created_at;
    }
    
    // Process content for embedded metadata if still missing key fields
    if ((typeof processedTweet.content === 'string') && 
        (!processedTweet.username || processedTweet.username === 'anonymous' || 
         processedTweet.username === 'user' || processedTweet.username === 'unknown')) {
      console.log(`[${processedTweet.id}] Extracting username/displayName from content`);
      const processed = processUserMetadata(processedTweet.content);
      
      if (processed.username && !processedTweet.username) {
        console.log(`[${processedTweet.id}] Extracted username from content:`, processed.username);
        processedTweet.username = processed.username;
      }
      
      if (processed.displayName && !processedTweet.displayName) {
        console.log(`[${processedTweet.id}] Extracted displayName from content:`, processed.displayName);
        processedTweet.displayName = processed.displayName;
      }
      
      processedTweet.content = processed.content;
    }
    
    // Handle special case - if username is 'anonymous' try to find a better username
    if (processedTweet.username === 'anonymous') {
      console.log(`[${processedTweet.id}] Found 'anonymous' username, searching for better alternatives`);
      
      // Look in user_data if it exists
      if (processedTweet.user_data && typeof processedTweet.user_data === 'object') {
        if (isValidUsername(processedTweet.user_data.username)) {
          console.log(`[${processedTweet.id}] Using user_data.username instead of anonymous:`, processedTweet.user_data.username);
          processedTweet.username = String(processedTweet.user_data.username || '');
        }
        if (isValidDisplayName(processedTweet.user_data.name)) {
          console.log(`[${processedTweet.id}] Using user_data.name:`, processedTweet.user_data.name);
          processedTweet.displayName = String(processedTweet.user_data.name || '');
        }
      }
      
      // Try to extract username from the URL if it exists in content
      if (processedTweet.content && processedTweet.content.includes('/profile/') && !isValidUsername(processedTweet.username)) {
        const urlMatch = processedTweet.content.match(/\/profile\/([a-zA-Z0-9_]+)/);
        if (urlMatch && urlMatch[1]) {
          console.log(`[${processedTweet.id}] Extracted username from profile URL:`, urlMatch[1]);
          processedTweet.username = urlMatch[1];
        }
      }
    }
    
    // Modify the final username/displayName checks to use the helper methods
    if (!isValidUsername(processedTweet.username)) {
      console.warn(`⚠️ [${processedTweet.id}] USING FALLBACK for username! No valid username found in:`, {
        username: originalTweet.username,
        authorUsername: originalTweet.authorUsername,
        author_username: originalTweet.author_username,
        extractedFromURL: processedTweet.content && processedTweet.content.match(/\/profile\/([a-zA-Z0-9_]+)/) 
          ? processedTweet.content.match(/\/profile\/([a-zA-Z0-9_]+)/)?.[1] || null 
          : null
      });
      
      // If this is a bookmarked tweet, try to extract author info from content or metadata
      if (processedTweet.content && processedTweet.content.includes('wrote:')) {
        const contentMatch = processedTweet.content.match(/([a-zA-Z0-9_]+)\s+wrote:/);
        if (contentMatch && contentMatch[1]) {
          console.log(`[${processedTweet.id}] Extracted username from content:`, contentMatch[1]);
          processedTweet.username = contentMatch[1];
          // Also try to extract display name if we can
          if (!isValidDisplayName(processedTweet.displayName) && processedTweet.content.includes('(')) {
            const nameMatch = processedTweet.content.match(/\(([^)]+)\)/);
            if (nameMatch && nameMatch[1]) {
              console.log(`[${processedTweet.id}] Extracted display name from content:`, nameMatch[1]);
              processedTweet.displayName = nameMatch[1];
            }
          }
        } else {
          processedTweet.username = 'user';
        }
      } else {
        processedTweet.username = 'user';
      }
    }
    
    if (!isValidDisplayName(processedTweet.displayName)) {
      console.warn(`⚠️ [${processedTweet.id}] USING FALLBACK for displayName! No valid display name found in:`, {
        displayName: originalTweet.displayName,
        authorName: originalTweet.authorName,
        author_name: originalTweet.author_name,
        name: originalTweet.name,
        display_name: originalTweet.display_name
      });
      processedTweet.displayName = 'User';
    }
    
    if (!processedTweet.avatar) {
      console.log(`[${processedTweet.id}] No avatar found, using null`);
      processedTweet.avatar = null;
    }
    
    console.log('Tweet processing result:', {
      id: processedTweet.id,
      username: processedTweet.username,
      displayName: processedTweet.displayName,
      hasAvatar: !!processedTweet.avatar
    });
    
    return processedTweet;
  }
  
  function formatTimestamp(timestamp: string): string {
    try {
      const date = new Date(timestamp);
      
      if (isNaN(date.getTime())) {
        return 'now';
      }
      
      const now = new Date();
      const seconds = Math.floor((now.getTime() - date.getTime()) / 1000);
      
      if (seconds < 0) return 'now';
      
      let interval = seconds / 31536000;
      if (interval > 1) {
        return Math.floor(interval) + 'y';
      }
      interval = seconds / 2592000;
      if (interval > 1) {
        return Math.floor(interval) + 'mo';
      }
      interval = seconds / 86400;
      if (interval > 1) {
        return Math.floor(interval) + 'd';
      }
      interval = seconds / 3600;
      if (interval > 1) {
        return Math.floor(interval) + 'h';
      }
      interval = seconds / 60;
      if (interval > 1) {
        return Math.floor(interval) + 'm';
      }
      return Math.floor(seconds) + 's';
    } catch (error) {
      console.error('Error formatting timestamp:', error);
      return 'now';
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
    if (!isAuthenticated) {
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
    if (!isAuthenticated) {
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

  // Handle like/unlike action
  async function handleLikeClick() {
    if (!isAuthenticated) {
      toastStore.showToast('Please log in to like posts', 'info');
      return;
    }

    if (isLikeLoading || isLikeRequestPending) {
      console.log('Like request already in progress, ignoring click');
      return;
    }

    const requestId = ++currentLikeRequestId;
    isLikeRequestPending = true;
    isLikeLoading = true;
    likeErrorState = false;

    // Determine if we're liking or unliking
    const isLiking = !effectiveIsLiked;
    const action = isLiking ? 'like' : 'unlike';

    try {
      // Optimistically update UI
      tweetInteractionStore.updateTweetInteraction(tweetId, {
        isLiked: isLiking,
        likes: effectiveLikes + (isLiking ? 1 : -1),
        pendingLike: true
      });

      console.log(`${action} action started for tweet ${tweetId}, isLiking=${isLiking}`);

      // Make API call
      if (isLiking) {
        await likeThread(tweetId);
      } else {
        await unlikeThread(tweetId);
      }

      // If this is still the most recent request
      if (requestId === currentLikeRequestId) {
        console.log(`${action} action completed successfully for tweet ${tweetId}`);
        // Update the store to remove pending state
        tweetInteractionStore.updateTweetInteraction(tweetId, {
          pendingLike: false
        });
      }
    } catch (error) {
      console.error(`Error ${action}ing tweet:`, error);
      
      // Only revert UI if this is the most recent request
      if (requestId === currentLikeRequestId) {
        // Check if this was an "already in state" error - don't revert the UI in that case
        const errorMsg = error instanceof Error ? error.message.toLowerCase() : '';
        const isAlreadyInState = (isLiking && errorMsg.includes('already liked')) || 
                              (!isLiking && (errorMsg.includes('not liked') || errorMsg.includes('not found')));
        
        if (!isAlreadyInState) {
          // Revert optimistic update
          tweetInteractionStore.updateTweetInteraction(tweetId, {
            isLiked: !isLiking,
            likes: effectiveLikes,
            pendingLike: false
          });
          likeErrorState = true;
          toastStore.showToast(`Failed to ${action} the tweet. Please try again.`, 'error', 5000);
        } else {
          // If it was already in the desired state, just update the UI to match reality
          tweetInteractionStore.updateTweetInteraction(tweetId, {
            pendingLike: false
          });
        }
      }
    } finally {
      // Only reset loading state if this is the most recent request
      if (requestId === currentLikeRequestId) {
        isLikeLoading = false;
        isLikeRequestPending = false;
      }
    }
  }

  async function handleBookmark() {
    if (!isAuthenticated) {
      toastStore.showToast('Please log in to bookmark posts', 'info');
      return;
    }
    
    // Prevent concurrent bookmark operations
    if (isBookmarkLoading) {
      return;
    }
    
    isBookmarkLoading = true;
    bookmarkErrorState = false;
    const requestId = ++currentBookmarkRequestId;
    
    try {
      // Ensure we have a valid ID
      if (!tweetId) {
        console.error('Cannot bookmark/unbookmark: Missing tweet ID');
        toastStore.showToast('Cannot process action: Invalid tweet ID', 'error');
        return;
      }
      
      // Set optimistic UI update based on current state
      const newBookmarkState = !effectiveIsBookmarked;
      
      // Optimistic UI update through the store
      tweetInteractionStore.updateTweetInteraction(tweetId, {
        isBookmarked: newBookmarkState,
        bookmarks: effectiveBookmarks + (newBookmarkState ? 1 : -1),
        pendingBookmark: true
      });
      
      // Trigger proper event using a string ID
      if (newBookmarkState) {
        dispatch('bookmark', String(tweetId));
      } else {
        dispatch('removeBookmark', String(tweetId));
      }
      
      // Call the appropriate API
      try {
        if (newBookmarkState) {
          await bookmarkThread(String(tweetId));
        } else {
          await removeBookmark(String(tweetId));
        }
        
        // Only update state if this is still the most recent request
        if (requestId !== currentBookmarkRequestId) {
          console.log('Ignoring stale bookmark response', requestId, currentBookmarkRequestId);
          return;
        }
        
        // Update the store to remove pending state
        tweetInteractionStore.updateTweetInteraction(tweetId, {
          pendingBookmark: false
        });
        
        // Clear any error state on success
        bookmarkErrorState = false;
      } catch (error) {
        console.error(`Error ${newBookmarkState ? 'bookmarking' : 'unbookmarking'} thread:`, error);
        
        // Only revert UI if this is still the most recent request
        if (requestId !== currentBookmarkRequestId) {
          console.log('Ignoring stale bookmark error', requestId, currentBookmarkRequestId);
          return;
        }
        
        // Check for already bookmarked/unbookmarked scenarios - don't revert the UI
        const errorMsg = error instanceof Error ? error.message.toLowerCase() : '';
        const isAlreadyInState = (newBookmarkState && errorMsg.includes('already bookmarked')) || 
                                (!newBookmarkState && (errorMsg.includes('not bookmarked') || errorMsg.includes('not found')));
        
        if (!isAlreadyInState) {
          // Set error state for retry functionality
          bookmarkErrorState = true;
          
          // Revert the optimistic update
          tweetInteractionStore.updateTweetInteraction(tweetId, {
            isBookmarked: !newBookmarkState,
            bookmarks: effectiveBookmarks,
            pendingBookmark: false
          });
          
          // Notify user
          const errorMessage = error instanceof Error ? error.message : `Failed to ${newBookmarkState ? 'bookmark' : 'remove bookmark'}. Please try again.`;
          toastStore.showToast(errorMessage, 'error', 5000);
        } else {
          // If it was already in the desired state, just update the UI to match reality
          tweetInteractionStore.updateTweetInteraction(tweetId, {
            pendingBookmark: false
          });
        }
      }
    } catch (error) {
      console.error('Unhandled error in handleBookmark:', error);
      toastStore.showToast('An unexpected error occurred. Please try again later.', 'error', 5000);
      bookmarkErrorState = true;
    } finally {
      // Only reset loading state if this is still the most recent request
      if (requestId === currentBookmarkRequestId) {
        isBookmarkLoading = false;
      }
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
      if (!isAuthenticated) {
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
      if (!isAuthenticated) {
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
      if (!isAuthenticated) {
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
        await bookmarkReply(String(replyId));
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
      if (!isAuthenticated) {
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
        await removeReplyBookmark(String(replyId));
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
                on:click|stopPropagation={handleRetweet}
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
                on:click|stopPropagation={handleBookmark}
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