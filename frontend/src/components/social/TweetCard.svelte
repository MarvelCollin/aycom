<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import type { ITweet } from '../../interfaces/ISocialMedia.d.ts';
  import { toastStore } from '../../stores/toastStore';
  import { formatTimeAgo, processUserMetadata } from '../../utils/common';
  import { likeThread, unlikeThread, bookmarkThread, removeBookmark, likeReply, unlikeReply, bookmarkReply, removeReplyBookmark } from '../../api/thread';
  
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

  export let tweet: ITweet;
  export let isDarkMode: boolean = false;
  export let isAuthenticated: boolean = false;
  
  export let isLiked: boolean = false;
  export let isReposted: boolean = false;
  export let isBookmarked: boolean = false;
  
  export let inReplyToTweet: ITweet | null = null;
  export let replies: ITweet[] = [];
  export let showReplies: boolean = false;
  export let nestingLevel: number = 0;
  const MAX_NESTING_LEVEL = 3;
  export let nestedRepliesMap: Map<string, ITweet[]> = new Map();
  
  const dispatch = createEventDispatcher();
  
  $: processedTweet = processTweetContent(tweet);
  
  $: processedReplies = replies.map(reply => processTweetContent(reply));
  
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
        processedTweet.username = processedTweet.author.username;
      }
      if (isValidDisplayName(processedTweet.author.name) && !isValidDisplayName(processedTweet.displayName)) {
        console.log(`[${processedTweet.id}] Using author.name:`, processedTweet.author.name);
        processedTweet.displayName = processedTweet.author.name;
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
          processedTweet.username = processedTweet.user_data.username;
        }
        if (isValidDisplayName(processedTweet.user_data.name)) {
          console.log(`[${processedTweet.id}] Using user_data.name:`, processedTweet.user_data.name);
          processedTweet.displayName = processedTweet.user_data.name;
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
          ? processedTweet.content.match(/\/profile\/([a-zA-Z0-9_]+)/)[1] 
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

  function handleReply() {
    if (!isAuthenticated) {
      toastStore.showToast('Please log in to reply to posts', 'info');
      return;
    }
    dispatch('reply', processedTweet.id);
  }

  function handleRetweet() {
    if (!isAuthenticated) {
      toastStore.showToast('Please log in to repost', 'info');
      return;
    }
    dispatch('repost', processedTweet.id);
    isReposted = !isReposted;
  }

  function handleLike() {
    if (!isAuthenticated) {
      toastStore.showToast('Please log in to like posts', 'info');
      return;
    }
    
    if (isLiked) {
      dispatch('unlike', processedTweet.id);
    } else {
      dispatch('like', processedTweet.id);
    }
  }

  function handleBookmark() {
    if (!isAuthenticated) {
      toastStore.showToast('Please log in to bookmark posts', 'info');
      return;
    }
    
    console.log(`Bookmark action - current status: ${isBookmarked ? 'bookmarked' : 'not bookmarked'}`);
    
    if (isBookmarked) {
      dispatch('removeBookmark', processedTweet.id);
    } else {
      dispatch('bookmark', processedTweet.id);
    }
  }

  function toggleReplies() {
    showReplies = !showReplies;
    if (showReplies) {
      console.log('Loading replies for tweet:', processedTweet.id);
      dispatch('loadReplies', processedTweet.id);
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
    dispatch('reply', event.detail);
  }

  function handleLoadNestedReplies(event) {
    dispatch('loadReplies', event.detail);
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

      reply.isLiked = true;
      
      await likeReply(String(replyId));
    } catch (error) {
      console.error('Error liking reply:', error);
      const reply = replies.find(r => r.id === replyId);
      if (reply) reply.isLiked = false;
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

      reply.isLiked = false;
      
      await unlikeReply(String(replyId));
    } catch (error) {
      console.error('Error unliking reply:', error);
      const reply = replies.find(r => r.id === replyId);
      if (reply) reply.isLiked = true;
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

      reply.isBookmarked = true;
      
      await bookmarkReply(String(replyId));
    } catch (error) {
      console.error('Error bookmarking reply:', error);
      const reply = replies.find(r => r.id === replyId);
      if (reply) reply.isBookmarked = false;
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

      reply.isBookmarked = false;
      
      await removeReplyBookmark(String(replyId));
    } catch (error) {
      console.error('Error unbookmarking reply:', error);
      const reply = replies.find(r => r.id === replyId);
      if (reply) reply.isBookmarked = true;
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
      bookmarkedThread: tweet.bookmarked_thread || null,
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
</script>

<div 
  class="tweet-card {isDarkMode ? 'tweet-card-dark' : ''} {nestingLevel > 0 ? 'nested-tweet' : 'main-tweet'} border-b {isDarkMode ? 'border-gray-800' : 'border-gray-200'} hover:bg-opacity-50 {isDarkMode ? 'hover:bg-gray-800 bg-gray-900 text-white' : 'hover:bg-gray-50 bg-white text-black'} transition-colors cursor-pointer"
  style="margin-left: {nestingLevel * 12}px;"
  on:click={handleClick}
  on:keydown={(e) => e.key === 'Enter' && handleClick()}
  role="article"
  tabindex="0"
  aria-label="Tweet by {processedTweet.displayName}"
>
  {#if inReplyToTweet && nestingLevel === 0}
    <div class="reply-context px-4 pt-2 pb-0">
      <div class="flex items-center text-sm {isDarkMode ? 'text-gray-400' : 'text-gray-500'}">
        <CornerUpRightIcon size="16" class="mr-2" />
        <span>Replying to <a 
          href={`/user/${inReplyToTweet.userId || inReplyToTweet.authorId || inReplyToTweet.author_id || inReplyToTweet.user_id || inReplyToTweet.username}`}
          class="text-blue-500 hover:underline"
          on:click|preventDefault={(e) => navigateToUserProfile(e, inReplyToTweet.username, inReplyToTweet.userId || inReplyToTweet.authorId || inReplyToTweet.author_id || inReplyToTweet.user_id)}
        >@{inReplyToTweet.username}</a></span>
      </div>
      <div class="ml-5 pl-4 border-l {isDarkMode ? 'border-gray-700' : 'border-gray-200'} my-1">
        <div class="text-sm {isDarkMode ? 'text-gray-300' : 'text-gray-600'} line-clamp-1">
          {inReplyToTweet.content}
        </div>
      </div>
    </div>
  {/if}

  {#if nestingLevel > 0}
    <div class="nested-reply-indicator {isDarkMode ? 'border-gray-700' : 'border-gray-300'}"></div>
  {/if}

  <div class="tweet-header p-4 relative">
    <div class="flex items-start">
      <a 
        href={`/user/${processedTweet.userId || processedTweet.authorId || processedTweet.author_id || processedTweet.user_id || processedTweet.username}`}
        class="tweet-avatar-container w-12 h-12 rounded-full overflow-hidden {isDarkMode ? 'bg-gray-700' : 'bg-gray-200'} flex items-center justify-center mr-3"
        on:click|preventDefault={(e) => navigateToUserProfile(e, processedTweet.username, processedTweet.userId || processedTweet.authorId || processedTweet.author_id || processedTweet.user_id)}
        on:keydown={(e) => e.key === 'Enter' && navigateToUserProfile(e, processedTweet.username, processedTweet.userId || processedTweet.authorId || processedTweet.author_id || processedTweet.user_id)}
      >
        {#if typeof processedTweet.avatar === 'string' && processedTweet.avatar.startsWith('http')}
          <img src={processedTweet.avatar} alt={processedTweet.username} class="w-full h-full object-cover" />
        {:else}
          <div class="text-xl {isDarkMode ? 'text-gray-100' : ''}">{processedTweet.avatar}</div>
        {/if}
      </a>
      
      <div class="flex-1 min-w-0">
        <div class="flex items-center">
          <a 
            href={`/user/${processedTweet.userId || processedTweet.authorId || processedTweet.author_id || processedTweet.user_id || processedTweet.username}`}
            class="font-bold {isDarkMode ? 'text-white' : 'text-black'} mr-1.5 hover:underline"
            on:click|preventDefault={(e) => navigateToUserProfile(e, processedTweet.username, processedTweet.userId || processedTweet.authorId || processedTweet.author_id || processedTweet.user_id)}
            on:keydown={(e) => e.key === 'Enter' && navigateToUserProfile(e, processedTweet.username, processedTweet.userId || processedTweet.authorId || processedTweet.author_id || processedTweet.user_id)}
          >
            {#if processedTweet.displayName === 'User'}
              {console.warn('❌ MISSING DISPLAY NAME:', {id: processedTweet.id, tweetId: processedTweet.tweetId, username: processedTweet.username})}
              <span class="text-red-500">User</span>
            {:else}
              {processedTweet.displayName}
            {/if}
          </a>
          <a 
            href={`/user/${processedTweet.userId || processedTweet.authorId || processedTweet.author_id || processedTweet.user_id || processedTweet.username}`}
            class="{isDarkMode ? 'text-gray-400' : 'text-gray-500'} text-sm truncate hover:underline"
            on:click|preventDefault={(e) => navigateToUserProfile(e, processedTweet.username, processedTweet.userId || processedTweet.authorId || processedTweet.author_id || processedTweet.user_id)}
            on:keydown={(e) => e.key === 'Enter' && navigateToUserProfile(e, processedTweet.username, processedTweet.userId || processedTweet.authorId || processedTweet.author_id || processedTweet.user_id)}
          >
            {#if processedTweet.username === 'user'}
              {console.warn('❌ MISSING USERNAME:', {id: processedTweet.id, tweetId: processedTweet.tweetId, displayName: processedTweet.displayName})}
              <span class="text-red-500">@user</span>
            {:else}
              @{processedTweet.username}
            {/if}
          </a>
          <span class="{isDarkMode ? 'text-gray-400' : 'text-gray-500'} mx-1.5">·</span>
          <span class="{isDarkMode ? 'text-gray-400' : 'text-gray-500'} text-sm">{formatTimeAgo(processedTweet.timestamp)}</span>
        </div>
        
        <div class="tweet-content my-2 {isDarkMode ? 'text-gray-100' : 'text-black'}">
          <p>{processedTweet.content || ''}</p>
        </div>
        
        {#if processedTweet.media && processedTweet.media.length > 0}
          <div class="media-container mt-2 rounded-xl overflow-hidden {isDarkMode ? 'border border-gray-700' : ''}">
            {#if processedTweet.media.length === 1}
              <div class="single-media h-64 w-full">
                {#if processedTweet.media[0].type === 'Image'}
                  <img src={processedTweet.media[0].url} alt="Media" class="h-full w-full object-cover" />
                {:else if processedTweet.media[0].type === 'Video'}
                  <video src={processedTweet.media[0].url} controls class="h-full w-full object-contain">
                    <track kind="captions" src="/captions/en.vtt" srclang="en" label="English" />
                  </video>
                {:else}
                  <img src={processedTweet.media[0].url} alt="GIF" class="h-full w-full object-cover" />
                {/if}
              </div>
            {:else if processedTweet.media.length > 1}
              <div class="media-grid grid gap-1" style="grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));">
                {#each processedTweet.media.slice(0, 4) as media, index (media.url || index)}
                  <div class="media-item h-40">
                    {#if media.type === 'Image'}
                      <img src={media.url} alt="Media" class="h-full w-full object-cover" />
                    {:else if media.type === 'Video'}
                      <video src={media.url} class="h-full w-full object-cover">
                        <track kind="captions" src="/captions/en.vtt" srclang="en" label="English" />
                      </video>
                    {:else}
                      <img src={media.url} alt="GIF" class="h-full w-full object-cover" />
                    {/if}
                  </div>
                {/each}
              </div>
            {/if}
          </div>
        {/if}
        
        <div class="flex justify-between mt-3 {isDarkMode ? 'text-gray-400' : 'text-gray-500'}">
          <div class="flex items-center">
            <button class="tweet-action-btn flex items-center rounded-full p-2 transition-colors {isDarkMode ? 'dark-btn hover:bg-blue-900/30' : 'light-btn hover:bg-blue-100'} hover:text-blue-500 {processedTweet.replies > 0 ? 'has-replies text-blue-500' : ''}" on:click|stopPropagation={handleReply} aria-label="Reply to tweet">
              <MessageCircleIcon size="20" class="mr-1.5" />
              <span class="reply-count {processedTweet.replies > 0 ? 'bg-blue-100 dark:bg-blue-900/40 px-1.5 py-0.5 rounded-full text-blue-600 dark:text-blue-400 font-medium' : ''}">{isNaN(processedTweet.replies) ? 0 : processedTweet.replies}</span>
              {#if processedTweet.replies > 0}
                <ChevronRightIcon size="14" class="ml-0.5 text-blue-500" />
              {/if}
            </button>
          </div>
          <div class="flex items-center">
            <button 
              class="tweet-action-btn flex items-center rounded-full p-2 transition-colors {isDarkMode ? 'dark-btn hover:bg-green-900/30' : 'light-btn hover:bg-green-100'} {isReposted ? 'text-green-500' : ''} hover:text-green-500" 
              on:click|stopPropagation={handleRetweet}
              aria-label="{isReposted ? 'Undo repost' : 'Repost'}"
            >
              <RefreshCwIcon size="20" class="mr-1.5" />
              <span class="count-badge">{isNaN(processedTweet.reposts) ? 0 : processedTweet.reposts}</span>
            </button>
          </div>
          <div class="flex items-center">
            <button 
              class="tweet-action-btn flex items-center rounded-full p-2 transition-colors {isDarkMode ? 'dark-btn hover:bg-red-900/30' : 'light-btn hover:bg-red-100'} {isLiked ? 'text-red-500' : ''} hover:text-red-500" 
              on:click|stopPropagation={handleLike}
              aria-label="{isLiked ? 'Unlike' : 'Like'}"
            >
              <HeartIcon size="20" fill={isLiked ? "currentColor" : "none"} class="mr-1.5" />
              <span class="count-badge">{isNaN(processedTweet.likes) ? 0 : processedTweet.likes}</span>
            </button>
          </div>
          <div class="flex items-center">
            <button 
              class="tweet-action-btn flex items-center rounded-full p-2 transition-colors {isDarkMode ? 'dark-btn hover:bg-blue-900/30' : 'light-btn hover:bg-blue-100'} {isBookmarked ? 'text-blue-500' : ''} hover:text-blue-500" 
              on:click|stopPropagation={handleBookmark}
              aria-label="{isBookmarked ? 'Remove bookmark' : 'Bookmark'}"
            >
              <BookmarkIcon size="20" fill={isBookmarked ? "currentColor" : "none"} class="mr-1.5" />
              <span class="count-badge">{isNaN(processedTweet.bookmarks) ? 0 : processedTweet.bookmarks}</span>
            </button>
          </div>
          <div class="flex items-center">
            <div class="flex items-center p-2 rounded-full transition-colors {isDarkMode ? 'dark-btn hover:bg-gray-700' : 'light-btn hover:bg-gray-100'}">
              <EyeIcon size="20" class="mr-1.5" />
              <span class="count-badge">{typeof processedTweet.views === 'string' ? processedTweet.views : '0'}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</div>

{#if nestingLevel === 0 && (replies.length > 0 || processedTweet.replies > 0)}
  <div class="ml-12 mt-1 mb-2">
    <button 
      class="text-sm flex items-center p-1.5 rounded-full {isDarkMode ? 'dark-btn text-gray-400 hover:text-blue-400 hover:bg-blue-900/20' : 'light-btn text-gray-500 hover:text-blue-500 hover:bg-blue-100'}"
      on:click|stopPropagation={toggleReplies}
      aria-expanded={showReplies}
      aria-controls="replies-container"
    >
      {#if showReplies}
        <ChevronUpIcon size="16" class="mr-1.5" />
      {:else}
        <ChevronDownIcon size="16" class="mr-1.5" />
      {/if}
      {#if replies.length > 0}
        {showReplies ? 'Hide' : 'Show'} {replies.length} {replies.length === 1 ? 'reply' : 'replies'}
      {:else}
        {showReplies ? 'Hide replies' : 'Show replies'}
      {/if}
    </button>
  </div>
{/if}

{#if showReplies}
  <div id="replies-container" class="replies-container {isDarkMode ? 'bg-gray-900' : 'bg-white'} ml-12 border-l {isDarkMode ? 'border-gray-700' : 'border-gray-200'} pl-4 pb-2">
    {#if replies.length === 0}
      <div class="py-4 text-center {isDarkMode ? 'text-gray-400' : 'text-gray-500'}">
        <div class="animate-pulse">Loading replies...</div>
      </div>
    {:else}
      {#if replies.length > 0 && nestingLevel < MAX_NESTING_LEVEL}
        {#each processedReplies as reply (reply.id || reply.tweetId)}
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
        {/each}
      {:else if replies.length > 0}
        {#each processedReplies as reply, index (reply.id || reply.tweetId || `reply-${reply.timestamp}-${reply.username}-${index}`)}
          <div class="reply-item py-3 {isDarkMode ? 'border-b border-gray-800' : 'border-b border-gray-200'}">
            <div class="flex">
              <a 
                href={`/user/${reply.userId || reply.authorId || reply.author_id || reply.user_id || reply.username}`}
                class="w-10 h-10 rounded-full overflow-hidden {isDarkMode ? 'bg-gray-700' : 'bg-gray-200'} flex items-center justify-center mr-3 flex-shrink-0"
                on:click|preventDefault={(e) => navigateToUserProfile(e, reply.username, reply.userId || reply.authorId || reply.author_id || reply.user_id)}
                on:keydown={(e) => e.key === 'Enter' && navigateToUserProfile(e, reply.username, reply.userId || reply.authorId || reply.author_id || reply.user_id)}
              >
                {#if typeof reply.avatar === 'string' && reply.avatar.startsWith('http')}
                  <img src={reply.avatar} alt={reply.username} class="w-full h-full object-cover" />
                {:else}
                  <div class="text-lg {isDarkMode ? 'text-gray-100' : ''}">{reply.avatar}</div>
                {/if}
              </a>
              <div class="flex-1">
                <div class="flex items-center">
                  <a 
                    href={`/user/${reply.userId || reply.authorId || reply.author_id || reply.user_id || reply.username}`}
                    class="font-bold {isDarkMode ? 'text-white' : 'text-black'} mr-1.5 hover:underline"
                    on:click|preventDefault={(e) => navigateToUserProfile(e, reply.username, reply.userId || reply.authorId || reply.author_id || reply.user_id)}
                    on:keydown={(e) => e.key === 'Enter' && navigateToUserProfile(e, reply.username, reply.userId || reply.authorId || reply.author_id || reply.user_id)}
                  >{reply.displayName || 'User'}</a>
                  <a 
                    href={`/user/${reply.userId || reply.authorId || reply.author_id || reply.user_id || reply.username}`}
                    class="{isDarkMode ? 'text-gray-400' : 'text-gray-500'} text-sm truncate hover:underline"
                    on:click|preventDefault={(e) => navigateToUserProfile(e, reply.username, reply.userId || reply.authorId || reply.author_id || reply.user_id)}
                    on:keydown={(e) => e.key === 'Enter' && navigateToUserProfile(e, reply.username, reply.userId || reply.authorId || reply.author_id || reply.user_id)}
                  >@{reply.username || 'user'}</a>
                  <span class="{isDarkMode ? 'text-gray-400' : 'text-gray-500'} mx-1.5">·</span>
                  <span class="{isDarkMode ? 'text-gray-400' : 'text-gray-500'} text-sm">{formatTimeAgo(reply.timestamp)}</span>
                </div>
                <div class="my-2 {isDarkMode ? 'text-gray-100' : 'text-black'}">
                  <p>{reply.content || 'No content available'}</p>
                  
                  {#if reply.media && reply.media.length > 0}
                    <div class="mt-2 rounded-lg overflow-hidden">
                      <img src={reply.media[0].url} alt="Media" class="h-40 w-full object-cover" />
                    </div>
                  {/if}
                </div>
                <div class="flex text-sm {isDarkMode ? 'text-gray-400' : 'text-gray-500'} mt-2">
                  <button class="flex items-center mr-4 hover:text-blue-500 p-1 rounded-full {isDarkMode ? 'dark-btn hover:bg-blue-900/30' : 'light-btn hover:bg-blue-100'}" on:click|stopPropagation={() => dispatch('reply', reply.id)}>
                    <MessageCircleIcon size="16" class="mr-1" />
                    <span>Reply</span>
                  </button>
                  
                  <button class="flex items-center mr-4 p-1 rounded-full {reply.isLiked ? 'text-red-500' : ''} {isDarkMode ? 'dark-btn hover:bg-red-900/30 hover:text-red-500' : 'light-btn hover:bg-red-100 hover:text-red-500'}" 
                    on:click|stopPropagation={(e) => {
                      e.preventDefault();
                      reply.isLiked ? handleUnlikeReply(reply.id) : handleLikeReply(reply.id);
                    }}>
                    {#if reply.isLiked}
                      <HeartIcon size="16" fill="currentColor" class="mr-1" />
                    {:else}
                      <HeartIcon size="16" class="mr-1" />
                    {/if}
                    <span>Like</span>
                  </button>
                  
                  <button class="flex items-center mr-4 p-1 rounded-full {reply.isBookmarked ? 'text-blue-500' : ''} {isDarkMode ? 'dark-btn hover:bg-blue-900/30 hover:text-blue-500' : 'light-btn hover:bg-blue-100 hover:text-blue-500'}" 
                    on:click|stopPropagation={(e) => {
                      e.preventDefault();
                      reply.isBookmarked ? handleUnbookmarkReply(reply.id) : handleBookmarkReply(reply.id);
                    }}>
                    {#if reply.isBookmarked}
                      <BookmarkIcon size="16" fill="currentColor" class="mr-1" />
                    {:else}
                      <BookmarkIcon size="16" class="mr-1" />
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

{#if nestingLevel === 0}
  <div class="mt-1 mb-2 ml-14">
    <button 
      class="flex items-center text-sm py-1.5 px-3 rounded-full border {isDarkMode ? 'bg-gray-800 border-gray-700 text-blue-400 hover:bg-gray-700' : 'bg-gray-50 border-gray-200 text-blue-500 hover:bg-gray-100'} transition-colors"
      on:click|stopPropagation={toggleReplies}
      aria-expanded={showReplies}
    >
      {#if !showReplies}
        <ChevronRightIcon size="16" class="mr-1.5" />
      {:else}
        <ChevronUpIcon size="16" class="mr-1.5" />
      {/if}
      
      {#if processedTweet.replies > 0}
        {showReplies ? 'Hide' : 'View'} {processedTweet.replies} {processedTweet.replies === 1 ? 'reply' : 'replies'}
      {:else}
        {showReplies ? 'Hide thread' : 'Reply to thread'}
      {/if}
    </button>
  </div>
{/if}

{#if nestingLevel > 0 && !showReplies}
  <div class="mt-1 mb-2 ml-12">
    <button 
      class="flex items-center text-xs py-1 px-2 rounded-full border {isDarkMode ? 'bg-gray-800 border-gray-700 text-blue-400 hover:bg-gray-700' : 'bg-gray-50 border-gray-200 text-blue-500 hover:bg-gray-100'} transition-colors"
      on:click|stopPropagation={toggleReplies}
    >
      <ChevronRightIcon size="14" class="mr-1.5" />
      
      {#if processedTweet.replies > 0}
        {processedTweet.replies} {processedTweet.replies === 1 ? 'reply' : 'replies'}
      {:else}
        Continue thread
      {/if}
    </button>
  </div>
{/if}

<style>
  .tweet-card {
    padding: 0.5rem 0;
  }
  
  .tweet-card-dark {
    background-color: #1a202c;
  }

  .nested-tweet {
    padding-left: 0.5rem;
    border-radius: 0.5rem;
    margin-top: 0.5rem;
    margin-bottom: 0.5rem;
    position: relative;
  }

  .nested-reply-indicator {
    position: absolute;
    top: 0;
    left: -1px;
    bottom: 0;
    border-left-width: 2px;
    width: 2px;
    opacity: 0.7;
  }

  .tweet-avatar-container {
    flex-shrink: 0;
  }
  
  .tweet-action-btn {
    transition: all 0.2s;
  }
  
  .dark-btn {
    background-color: transparent;
  }
  
  .dark-btn:hover {
    background-color: rgba(59, 130, 246, 0.2);
  }
  
  .light-btn {
    background-color: transparent;
  }
  
  .light-btn:hover {
    background-color: rgba(29, 155, 240, 0.1);
  }
  
  .media-grid {
    max-height: 300px;
    overflow: hidden;
  }
  
  .line-clamp-1 {
    display: -webkit-box;
    -webkit-line-clamp: 1;
    line-clamp: 1;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }
  
  @keyframes pulse {
    0%, 100% { opacity: 1; }
    50% { opacity: 0.5; }
  }
  .animate-pulse {
    animation: pulse 1.5s cubic-bezier(0.4, 0, 0.6, 1) infinite;
  }

  .reply-count {
    min-width: 1rem;
    text-align: center;
  }

  .replies-container {
    position: relative;
  }

  .replies-container:before {
    content: '';
    position: absolute;
    left: 0;
    top: 0;
    bottom: 0;
    width: 2px;
    background-color: currentColor;
    opacity: 0.2;
  }

  .has-replies {
    font-weight: 500;
  }

  .count-badge {
    min-width: 1.5rem;
    text-align: center;
    font-weight: 500;
    display: inline-block;
  }
  
  .tweet-action-btn {
    padding: 6px 10px;
    border-radius: 9999px;
  }
</style>