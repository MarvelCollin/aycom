<script lang="ts">
  import { onMount, onDestroy, createEventDispatcher } from 'svelte';
  import { fade, slide } from 'svelte/transition';
  import { getThread, getThreadReplies, getReplyReplies, replyToThread } from '../api/thread';
  import { uploadMedia } from '../api/media';
  import TweetCard from '../components/social/TweetCard.svelte';
  import ComposeTweetModal from '../components/social/ComposeTweetModal.svelte';
  import ComposeTweet from '../components/social/ComposeTweet.svelte';
  import MediaOverlay from '../components/media/MediaOverlay.svelte';
  import MainLayout from '../components/layout/MainLayout.svelte';
  import Spinner from '../components/common/Spinner.svelte';
  import { toastStore } from '../stores/toastStore';
  import { tweetInteractionStore } from '../stores/tweetInteractionStore';
  import { authStore } from '../stores/authStore';
  import { useTheme } from '../hooks/useTheme';
  import { generateFilePreview } from '../utils/common';
  import { createLoggerWithPrefix } from '../utils/logger';
  import type { ITweet } from '../interfaces/ISocialMedia';
  import type { IReply } from '../interfaces/ISocialMedia';
  import type { IMedia } from '../interfaces/IMedia';
  import ArrowLeftIcon from 'svelte-feather-icons/src/icons/ArrowLeftIcon.svelte';
  import HeartIcon from 'svelte-feather-icons/src/icons/HeartIcon.svelte';
  import BookmarkIcon from 'svelte-feather-icons/src/icons/BookmarkIcon.svelte';
  import RefreshCwIcon from 'svelte-feather-icons/src/icons/RefreshCwIcon.svelte';
  import MessageCircleIcon from 'svelte-feather-icons/src/icons/MessageCircleIcon.svelte';
  import MoreHorizontalIcon from 'svelte-feather-icons/src/icons/MoreHorizontalIcon.svelte';
  import ImageIcon from 'svelte-feather-icons/src/icons/ImageIcon.svelte';
  import VideoIcon from 'svelte-feather-icons/src/icons/VideoIcon.svelte';
  import XIcon from 'svelte-feather-icons/src/icons/XIcon.svelte';
  import SmileIcon from 'svelte-feather-icons/src/icons/SmileIcon.svelte';

  const logger = createLoggerWithPrefix('ThreadDetail');
  const dispatch = createEventDispatcher();
  const { theme } = useTheme();
  
  export let threadId;
  let thread = null;
  let replies = [];
  let isLoading = true;
  let nestedRepliesMap = new Map();
  let showReplyForm = false;
  let replyTo = null;
  let isLoadingReplies = false;
  let isDarkMode = false;
  
  $: isDarkMode = $theme === 'dark';

  // Function to load all replies and nested replies
  async function loadThreadWithReplies() {
    isLoading = true;
    try {
      // Load the thread
      const threadData = await getThread(threadId);
      thread = threadData;
      
      // Load replies to the thread
      await loadReplies(threadId);
      
      // Auto-load all nested replies for first-level replies
      await Promise.all(replies.map(async (reply) => {
        if (reply.replies > 0) {
          await loadNestedReplies(reply.id);
        }
      }));
    } catch (error) {
      console.error('Error loading thread:', error);
      toastStore.showToast('Failed to load thread details', 'error');
    } finally {
      isLoading = false;
    }
  }
  
  // Load replies for a thread
  async function loadReplies(threadId) {
    isLoadingReplies = true;
    try {
      const response = await getThreadReplies(threadId);
      if (response && response.replies) {
        replies = response.replies;
        
        // If the API provides a nestedRepliesMap, use it
        if (response.nestedRepliesMap) {
          console.log('API provided nested replies map:', response.nestedRepliesMap);
          
          // Convert the object back to a Map for our component
          const newNestedMap = new Map();
          Object.entries(response.nestedRepliesMap).forEach(([parentId, childReplies]) => {
            newNestedMap.set(parentId, childReplies);
          });
          
          nestedRepliesMap = newNestedMap;
        } else {
          // Fall back to the old way - load nested replies individually
          console.log('No nested replies map from API, loading individually');
        }
      }
    } catch (error) {
      console.error('Error loading replies:', error);
      toastStore.showToast('Failed to load replies', 'error');
    } finally {
      isLoadingReplies = false;
    }
  }
  
  // Load nested replies for a reply
  async function loadNestedReplies(replyId) {
    try {
      const replyIdStr = String(replyId);
      const response = await getReplyReplies(replyIdStr);
      if (response && response.replies && response.replies.length > 0) {
        nestedRepliesMap.set(replyIdStr, response.replies);
        // Force reactivity update
        nestedRepliesMap = new Map(nestedRepliesMap);
      }
    } catch (error) {
      console.error(`Error loading nested replies for reply ${replyId}:`, error);
    }
  }
  
  // Handle the reply action
  function handleReply(event) {
    if (!authStore.isAuthenticated()) {
      toastStore.showToast('Please log in to reply', 'info');
      return;
    }
    
    const targetId = event.detail;
    
    // If replying to the main thread
    if (targetId === thread?.id) {
      replyTo = {
        id: thread.id,
        content: thread.content,
        username: thread.username || thread.author?.username || thread.user?.username,
        displayName: thread.displayName || thread.author?.name || thread.user?.name,
        avatar: thread.avatar || thread.author?.profile_picture_url || thread.user?.profile_picture_url
      };
    } 
    // If replying to a reply
    else {
      // Find the reply either in the main replies array or in nested replies
      let targetReply = replies.find(r => r.id === targetId);
      
      if (!targetReply) {
        // Check in nested replies
        for (const [parentId, nestedReplies] of nestedRepliesMap.entries()) {
          targetReply = nestedReplies.find(r => r.id === targetId);
          if (targetReply) {
            // Set the parent_reply_id to enable reply-to-reply
            targetReply.parentReplyId = parentId;
            break;
          }
        }
      }
      
      if (targetReply) {
        replyTo = {
          id: targetReply.id,
          thread_id: thread.id,
          content: targetReply.content,
          username: targetReply.username || targetReply.author?.username || targetReply.user?.username,
          displayName: targetReply.displayName || targetReply.author?.name || targetReply.user?.name,
          avatar: targetReply.avatar || targetReply.author?.profile_picture_url || targetReply.user?.profile_picture_url,
          parentReplyId: targetReply.parentReplyId || targetReply.parent_reply_id
        };
      }
    }
    
    showReplyForm = true;
  }
  
  // Handle reply submission - auto-refresh the replies
  function handleRefreshReplies(event) {
    const { threadId, parentReplyId, newReply } = event.detail;
    
    // If it's a reply to the main thread
    if (!parentReplyId) {
      // Add the new reply to the beginning of the replies array
      replies = [newReply, ...replies];
    } 
    // If it's a reply to another reply (nested reply)
    else {
      // Get the current nested replies for this parent reply
      let currentNestedReplies = nestedRepliesMap.get(parentReplyId) || [];
      
      console.log('Adding nested reply to parent:', parentReplyId);
      console.log('Current nested replies:', currentNestedReplies);
      
      // Make sure the new reply has the parent_reply_id set
      const enhancedReply = {
        ...newReply,
        parent_reply_id: parentReplyId
      };
      
      // Add the new reply to the beginning of the nested replies
      nestedRepliesMap.set(parentReplyId, [enhancedReply, ...currentNestedReplies]);
      
      // Force reactivity update
      nestedRepliesMap = new Map(nestedRepliesMap);
      
      // Find the parent reply in the main replies array
      const parentReply = replies.find(r => r.id === parentReplyId);
      if (parentReply) {
        // Make a copy of the parent with the incremented reply count
        const updatedParentReply = {
          ...parentReply,
          replies: (parseInt(parentReply.replies || '0') + 1).toString()
        };
        
        // Update the parent reply in the replies array
        replies = replies.map(reply => 
          reply.id === parentReplyId ? updatedParentReply : reply
        );
        
        console.log('Updated parent reply with new reply count:', updatedParentReply.replies);
      } else {
        console.warn(`Parent reply ${parentReplyId} not found in main replies array`);
        
        // Check if it's a reply to a nested reply (multi-level nesting)
        for (const [topReplyId, nestedReplies] of nestedRepliesMap.entries()) {
          const nestedParent = nestedReplies.find(r => r.id === parentReplyId);
          if (nestedParent) {
            console.log(`Found parent reply ${parentReplyId} as a nested reply under ${topReplyId}`);
            
            // Increment the nested parent's reply count
            const updatedNestedParent = {
              ...nestedParent,
              replies: (parseInt(nestedParent.replies || '0') + 1).toString()
            };
            
            // Update the nested parent in the nested replies array
            const updatedNestedReplies = nestedReplies.map(reply => 
              reply.id === parentReplyId ? updatedNestedParent : reply
            );
            
            // Update the map
            nestedRepliesMap.set(topReplyId, updatedNestedReplies);
            
            // Also create a new entry in the map for this nested reply
            nestedRepliesMap.set(parentReplyId, [enhancedReply]);
            
            // Force reactivity update
            nestedRepliesMap = new Map(nestedRepliesMap);
            break;
          }
        }
      }
    }
    
    // Close the reply form
    showReplyForm = false;
  }
  
  // Like, bookmark handlers
  function handleLike(event) {
    const tweetId = event.detail;
    if (tweetId === thread.id) {
      thread.isLiked = true;
      thread.likes = (parseInt(thread.likes || '0') + 1).toString();
    } else {
      updateNestedInteraction(tweetId, 'like', true);
    }
  }
  
  function handleUnlike(event) {
    const tweetId = event.detail;
    if (tweetId === thread.id) {
      thread.isLiked = false;
      thread.likes = Math.max(0, parseInt(thread.likes || '0') - 1).toString();
    } else {
      updateNestedInteraction(tweetId, 'like', false);
    }
  }
  
  function handleBookmark(event) {
    const tweetId = event.detail;
    if (tweetId === thread.id) {
      thread.isBookmarked = true;
      thread.bookmarks = (parseInt(thread.bookmarks || '0') + 1).toString();
    } else {
      updateNestedInteraction(tweetId, 'bookmark', true);
    }
  }
  
  function handleRemoveBookmark(event) {
    const tweetId = event.detail;
    if (tweetId === thread.id) {
      thread.isBookmarked = false;
      thread.bookmarks = Math.max(0, parseInt(thread.bookmarks || '0') - 1).toString();
    } else {
      updateNestedInteraction(tweetId, 'bookmark', false);
    }
  }
  
  // Update like or bookmark status in nested replies
  function updateNestedInteraction(tweetId, type, isActive) {
    // Check in first level replies
    let reply = replies.find(r => r.id === tweetId);
    if (reply) {
      if (type === 'like') {
        reply.isLiked = isActive;
        reply.likes = isActive 
          ? (parseInt(reply.likes || '0') + 1).toString()
          : Math.max(0, (parseInt(reply.likes || '0') - 1)).toString();
      } else {
        reply.isBookmarked = isActive;
        reply.bookmarks = isActive 
          ? (parseInt(reply.bookmarks || '0') + 1).toString()
          : Math.max(0, (parseInt(reply.bookmarks || '0') - 1)).toString();
      }
      return;
    }
    
    // Check in nested replies
    for (const [parentId, nestedReplies] of nestedRepliesMap.entries()) {
      reply = nestedReplies.find(r => r.id === tweetId);
      if (reply) {
        if (type === 'like') {
          reply.isLiked = isActive;
          reply.likes = isActive 
            ? (parseInt(reply.likes || '0') + 1).toString()
            : Math.max(0, (parseInt(reply.likes || '0') - 1)).toString();
        } else {
          reply.isBookmarked = isActive;
          reply.bookmarks = isActive 
            ? (parseInt(reply.bookmarks || '0') + 1).toString()
            : Math.max(0, (parseInt(reply.bookmarks || '0') - 1)).toString();
        }
        
        // Update the map to trigger reactivity
        const updatedReplies = [...nestedRepliesMap.get(parentId)];
        nestedRepliesMap.set(parentId, updatedReplies);
        nestedRepliesMap = new Map(nestedRepliesMap);
        return;
      }
    }
  }
  
  onMount(() => {
    if (threadId) {
      loadThreadWithReplies();
    }
  });
</script>

<svelte:head>
  <title>{thread ? `${thread.displayName || 'User'}'s Post` : 'Thread'} | AYCOM</title>
</svelte:head>

<div class="thread-detail-container {isDarkMode ? 'bg-gray-900 text-white' : 'bg-white text-black'}">
  <div class="thread-content max-w-3xl mx-auto">
    <!-- Back Button -->
    <div class="back-button-container p-2">
      <button 
        class="back-button rounded-full p-2 {isDarkMode ? 'hover:bg-gray-800' : 'hover:bg-gray-100'} transition-colors"
        on:click={() => window.history.back()}
        aria-label="Go back"
      >
        <ArrowLeftIcon size="20" class="{isDarkMode ? 'text-white' : 'text-black'}" />
      </button>
      <span class="ml-2 font-semibold {isDarkMode ? 'text-white' : 'text-black'}">Thread</span>
    </div>

    {#if isLoading}
      <div class="p-4 text-center">
        <div class="loader {isDarkMode ? 'border-blue-500' : 'border-blue-600'}"></div>
        <p class="mt-4 {isDarkMode ? 'text-gray-300' : 'text-gray-600'}">Loading thread...</p>
      </div>
    {:else if thread}
      <div transition:fade={{ duration: 200 }}>
        <!-- Main Thread -->
        <TweetCard 
          tweet={thread} 
          {isDarkMode}
          isAuth={authStore.isAuthenticated()}
          isLiked={thread.isLiked || thread.is_liked || false}
          isBookmarked={thread.isBookmarked || thread.is_bookmarked || false}
          showReplies={true}
          replies={replies}
          {nestedRepliesMap}
          on:reply={handleReply}
          on:like={handleLike}
          on:unlike={handleUnlike}
          on:bookmark={handleBookmark}
          on:removeBookmark={handleRemoveBookmark}
          on:loadReplies={() => {}}
        />
        
        <!-- Reply Form -->
        {#if showReplyForm}
          <div class="reply-form-container p-4 {isDarkMode ? 'bg-gray-800' : 'bg-gray-50'} rounded-lg mt-4">
            <ComposeTweet 
              {isDarkMode} 
              {replyTo}
              on:close={() => showReplyForm = false}
              on:tweet={() => showReplyForm = false}
              on:refreshReplies={handleRefreshReplies}
            />
          </div>
        {:else}
          <div class="p-4">
            <button 
              class="reply-button py-2 px-4 rounded-full border {isDarkMode ? 'border-gray-700 bg-gray-800 text-blue-400 hover:bg-gray-700' : 'border-gray-200 bg-gray-50 text-blue-500 hover:bg-gray-100'} transition-colors w-full"
              on:click={() => {
                replyTo = thread;
                showReplyForm = true;
              }}
            >
              Reply to this thread
            </button>
          </div>
        {/if}
      </div>
    {:else}
      <div class="p-4 text-center {isDarkMode ? 'text-gray-300' : 'text-gray-600'}">
        <p>Thread not found</p>
      </div>
    {/if}
  </div>
</div>

<style>
  .thread-detail-container {
    min-height: calc(100vh - 60px);
    padding: 1rem 0;
  }
  
  .back-button-container {
    display: flex;
    align-items: center;
    padding: 0.75rem 1rem;
    border-bottom: 1px solid rgba(0, 0, 0, 0.1);
  }
  
  .back-button {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 36px;
    height: 36px;
    transition: background-color 0.2s ease;
  }
  
  .loader {
    border: 4px solid rgba(0, 0, 0, 0.1);
    border-radius: 50%;
    border-top-color: #3498db;
    width: 40px;
    height: 40px;
    animation: spin 1s linear infinite;
    margin: 0 auto;
  }
  
  @keyframes spin {
    0% { transform: rotate(0deg); }
    100% { transform: rotate(360deg); }
  }
  
  .reply-button {
    transition: all 0.2s;
    display: flex;
    align-items: center;
    justify-content: center;
  }
  
  .reply-button:hover {
    transform: translateY(-1px);
  }
</style>
