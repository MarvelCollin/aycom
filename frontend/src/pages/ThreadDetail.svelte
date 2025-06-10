<script lang="ts">
  import { onMount } from 'svelte';
  import { fade } from 'svelte/transition';
  import { getThread, getThreadReplies, getReplyReplies } from '../api/thread';
  import TweetCard from '../components/social/TweetCard.svelte';
  import ComposeTweet from '../components/social/ComposeTweet.svelte';
  import { toastStore } from '../stores/toastStore';
  import { authStore } from '../stores/authStore';
  import { useTheme } from '../hooks/useTheme';
  import { createLoggerWithPrefix } from '../utils/logger';
  import ArrowLeftIcon from 'svelte-feather-icons/src/icons/ArrowLeftIcon.svelte';
  import type { ITweet } from '../interfaces/ISocialMedia';
  import { type ExtendedTweet, ensureTweetFormat } from '../interfaces/ITweet.extended';

  const logger = createLoggerWithPrefix('ThreadDetail');
  const { theme } = useTheme();
  
  export let threadId: string;
  let thread: ExtendedTweet | null = null;
  let replies: ExtendedTweet[] = [];
  let isLoading = true;
  let nestedRepliesMap = new Map<string, ExtendedTweet[]>();
  let showReplyForm = false;
  let replyTo: ExtendedTweet | null = null;
  let isDarkMode = false;
  
  // Update isDarkMode based on the document's data-theme attribute
  $: {
    if (typeof document !== 'undefined') {
      isDarkMode = document.documentElement.getAttribute('data-theme') === 'dark';
    } else {
      isDarkMode = $theme === 'dark';
    }
  }

  // Function to load the thread and its replies
  async function loadThreadWithReplies() {
    isLoading = true;
    try {
      // Load the thread
      const threadData = await getThread(threadId);
      console.log('Thread data received from API:', threadData);
      
      // Process the thread data with our utility
      thread = ensureTweetFormat(threadData);
      console.log('Processed thread data:', thread);
      
      // Load replies to the thread
      await loadReplies(threadId);
      
    } catch (error) {
      console.error('Error loading thread:', error);
      toastStore.showToast('Failed to load thread details', 'error');
    } finally {
      isLoading = false;
    }
  }
  
  // Load replies for a thread
  async function loadReplies(threadId: string) {
    try {
      const response = await getThreadReplies(threadId);
      
      if (response && response.replies) {
        replies = (response.replies as any[]).map(reply => ensureTweetFormat(reply));
        
        // Load nested replies for each reply that has them
        for (const reply of replies) {
          if (reply.replies_count && reply.replies_count > 0) {
            await loadNestedReplies(reply.id);
          }
        }
      }
    } catch (error) {
      console.error('Error loading replies:', error);
    }
  }
  
  // Load nested replies
  async function loadNestedReplies(replyId: string) {
    try {
      const response = await getReplyReplies(replyId);
      
      if (response && response.replies) {
        const nestedReplies = (response.replies as any[]).map(reply => ensureTweetFormat(reply));
        nestedRepliesMap.set(replyId, nestedReplies);
        nestedRepliesMap = new Map(nestedRepliesMap); // Force reactivity update
      }
    } catch (error) {
      console.error(`Error loading nested replies for ${replyId}:`, error);
    }
  }
  
  // Handle refresh replies from ComposeTweet component
  function handleRefreshReplies(event: any) {
    const { threadId, parentReplyId, newReply } = event.detail;
    
    if (!newReply) return;
    
    const processedReply = ensureTweetFormat(newReply);
    
    // If refreshing replies for a thread
    if (threadId === thread?.id && !parentReplyId) {
      // Add the new reply to our replies array
      replies = [processedReply, ...replies];
      
      // Update reply count on thread
      if (thread) {
        thread.replies_count = (thread.replies_count || 0) + 1;
      }
    }
    // If refreshing replies for a specific reply (nested reply)
    else if (parentReplyId) {
      // Get current nested replies or empty array if none exist
      const currentNestedReplies = nestedRepliesMap.get(parentReplyId) || [];
      // Add the new reply to the nested replies
      nestedRepliesMap.set(parentReplyId, [processedReply, ...currentNestedReplies]);
      // Update the map to trigger reactivity
      nestedRepliesMap = new Map(nestedRepliesMap);
      
      // Find the parent reply and update its reply count
      const parentReplyIndex = replies.findIndex(r => r.id === parentReplyId);
      if (parentReplyIndex >= 0) {
        replies[parentReplyIndex].replies_count = (replies[parentReplyIndex].replies_count || 0) + 1;
        replies = [...replies]; // Force reactivity update
      }
    }
    
    // Close the reply form after a successful reply
    showReplyForm = false;
  }
  
  // Handle reply button click
  function handleReply(event: CustomEvent<string>) {
    if (!authStore.isAuthenticated()) {
      toastStore.showToast('Please log in to reply', 'info');
      return;
    }
    
    const targetId = event.detail;
    
    // If replying to the main thread
    if (thread && targetId === thread.id) {
      replyTo = ensureTweetFormat({
        id: thread.id,
        content: thread.content,
        username: thread.username || thread.author?.username || '',
        name: thread.name,
        display_name: thread.display_name,
        parent_id: null,
        is_pinned: false,
        avatar: thread.avatar || thread.profile_picture_url,
        user_id: thread.user_id,
        created_at: thread.created_at
      });
    } 
    // If replying to a reply
    else {
      // Find the reply either in the main replies array or in nested replies
      const targetReply = replies.find(r => r.id === targetId);
      
      if (targetReply) {
        replyTo = ensureTweetFormat({
          id: targetReply.id,
          thread_id: thread?.id,
          content: targetReply.content,
          username: targetReply.username,
          name: targetReply.name,
          parent_id: null,
          is_pinned: false,
          avatar: targetReply.avatar || targetReply.profile_picture_url,
          parentReplyId: targetReply.parentReplyId || targetReply.parent_reply_id,
          user_id: targetReply.user_id,
          created_at: targetReply.created_at
        });
      } else {
        // Check in nested replies
        for (const [parentId, nestedReplies] of nestedRepliesMap.entries()) {
          const nestedReply = nestedReplies.find(r => r.id === targetId);
          if (nestedReply) {
            replyTo = ensureTweetFormat({
              id: nestedReply.id,
              thread_id: thread?.id,
              content: nestedReply.content,
              username: nestedReply.username,
              name: nestedReply.name,
              parent_id: null,
              is_pinned: false,
              avatar: nestedReply.avatar || nestedReply.profile_picture_url,
              parentReplyId: parentId,
              user_id: nestedReply.user_id,
              created_at: nestedReply.created_at
            });
            break;
          }
        }
      }
    }
    
    showReplyForm = true;
  }
  
  // Like, unlike, bookmark, unbookmark handlers
  function handleLike(event: CustomEvent<string>) {
    const tweetId = event.detail;
    if (thread && tweetId === thread.id) {
      thread.is_liked = true;
      thread.likes_count = (thread.likes_count || 0) + 1;
      thread = { ...thread }; // Force reactivity
    } else {
      updateNestedInteraction(tweetId, 'like', true);
    }
  }
  
  function handleUnlike(event: CustomEvent<string>) {
    const tweetId = event.detail;
    if (thread && tweetId === thread.id) {
      thread.is_liked = false;
      thread.likes_count = Math.max(0, (thread.likes_count || 0) - 1);
      thread = { ...thread }; // Force reactivity
    } else {
      updateNestedInteraction(tweetId, 'like', false);
    }
  }
  
  function handleBookmark(event: CustomEvent<string>) {
    const tweetId = event.detail;
    if (thread && tweetId === thread.id) {
      thread.is_bookmarked = true;
      thread.bookmark_count = (thread.bookmark_count || 0) + 1;
      thread = { ...thread }; // Force reactivity
    } else {
      updateNestedInteraction(tweetId, 'bookmark', true);
    }
  }
  
  function handleRemoveBookmark(event: CustomEvent<string>) {
    const tweetId = event.detail;
    if (thread && tweetId === thread.id) {
      thread.is_bookmarked = false;
      thread.bookmark_count = Math.max(0, (thread.bookmark_count || 0) - 1);
      thread = { ...thread }; // Force reactivity
    } else {
      updateNestedInteraction(tweetId, 'bookmark', false);
    }
  }
  
  // Update like or bookmark status in nested replies
  function updateNestedInteraction(tweetId: string, type: 'like' | 'bookmark', isActive: boolean) {
    // Check in first level replies
    const replyIndex = replies.findIndex(r => r.id === tweetId);
    
    if (replyIndex >= 0) {
      if (type === 'like') {
        replies[replyIndex].is_liked = isActive;
        replies[replyIndex].likes_count = isActive 
          ? (replies[replyIndex].likes_count || 0) + 1
          : Math.max(0, (replies[replyIndex].likes_count || 0) - 1);
      } else {
        replies[replyIndex].is_bookmarked = isActive;
        replies[replyIndex].bookmark_count = isActive 
          ? (replies[replyIndex].bookmark_count || 0) + 1
          : Math.max(0, (replies[replyIndex].bookmark_count || 0) - 1);
      }
      
      // Force reactivity update
      replies = [...replies];
      return;
    }
    
    // Check in nested replies
    for (const [parentId, nestedReplies] of nestedRepliesMap.entries()) {
      const nestedReplyIndex = nestedReplies.findIndex(r => r.id === tweetId);
      if (nestedReplyIndex >= 0) {
        const updatedReplies = [...nestedReplies];
        
        if (type === 'like') {
          updatedReplies[nestedReplyIndex].is_liked = isActive;
          updatedReplies[nestedReplyIndex].likes_count = isActive 
            ? (updatedReplies[nestedReplyIndex].likes_count || 0) + 1
            : Math.max(0, (updatedReplies[nestedReplyIndex].likes_count || 0) - 1);
        } else {
          updatedReplies[nestedReplyIndex].is_bookmarked = isActive;
          updatedReplies[nestedReplyIndex].bookmark_count = isActive 
            ? (updatedReplies[nestedReplyIndex].bookmark_count || 0) + 1
            : Math.max(0, (updatedReplies[nestedReplyIndex].bookmark_count || 0) - 1);
        }
        
        // Update the map with the modified array
        nestedRepliesMap.set(parentId, updatedReplies);
        // Force reactivity update
        nestedRepliesMap = new Map(nestedRepliesMap);
        return;
      }
    }
  }
  
  // Initialize component
  onMount(() => {
    if (threadId) {
      loadThreadWithReplies();
    }
    
    // Add event listener to check for theme changes
    const observer = new MutationObserver((mutations) => {
      mutations.forEach((mutation) => {
        if (mutation.attributeName === 'data-theme') {
          isDarkMode = document.documentElement.getAttribute('data-theme') === 'dark';
        }
      });
    });
    
    observer.observe(document.documentElement, { attributes: true, attributeFilter: ['data-theme'] });
    
    // Initial check
    isDarkMode = document.documentElement.getAttribute('data-theme') === 'dark';
    
    return () => {
      observer.disconnect();
    };
  });
</script>

<svelte:head>
  <title>{thread ? `${thread.name || thread.display_name || 'User'}'s Post` : 'Thread'} | AYCOM</title>
</svelte:head>

<div class="thread-detail-container">
  <div class="thread-content">
    <!-- Back Button -->
    <div class="back-button-container">
      <button 
        class="back-button"
        on:click={() => window.history.back()}
        aria-label="Go back"
      >
        <ArrowLeftIcon size="20" class="{isDarkMode ? 'text-white' : 'text-black'}" />
      </button>
      <span class="ml-2 font-semibold">Thread</span>
    </div>

    {#if isLoading}
      <div class="p-4 text-center">
        <div class="loader"></div>
        <p class="mt-4 {isDarkMode ? 'text-gray-300' : 'text-gray-600'}">Loading thread...</p>
      </div>
    {:else if thread}
      <div transition:fade={{ duration: 200 }}>
        <!-- Main Thread -->
        <TweetCard 
          tweet={thread} 
          {isDarkMode}
          isAuth={authStore.isAuthenticated()}
          isLiked={thread.is_liked || false}
          isBookmarked={thread.is_bookmarked || false}
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
        {#if showReplyForm && replyTo}
          <div class="reply-form-container">
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
              class="reply-button"
              on:click={() => {
                if (thread) {
                  replyTo = ensureTweetFormat({
                    id: thread.id,
                    content: thread.content,
                    username: thread.username || '',
                    name: thread.name || '',
                    parent_id: null,
                    is_pinned: false,
                    avatar: thread.avatar || thread.profile_picture_url,
                    user_id: thread.user_id,
                    created_at: thread.created_at
                  });
                  showReplyForm = true;
                }
              }}
            >
              Reply to this thread
            </button>
          </div>
        {/if}
        
        <!-- Replies List with visual connection line -->
        {#if replies && replies.length > 0}
          <div class="reply-section">
            <div class="reply-separator"></div>
            <!-- Replies are already rendered by TweetCard component -->
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
  /* Base styles */
  .thread-detail-container {
    min-height: calc(100vh - 60px);
    padding: 0;
    background-color: var(--light-bg-primary);
    color: var(--light-text-primary);
  }
  
  /* Dark theme overrides - these will apply when dark theme is active */
  :global([data-theme="dark"]) .thread-detail-container {
    background-color: var(--dark-bg-primary);
    color: var(--dark-text-primary);
  }
  
  .thread-content {
    max-width: 800px;
    margin: 0 auto;
    background-color: var(--light-bg-primary);
  }
  
  :global([data-theme="dark"]) .thread-content {
    background-color: var(--dark-bg-primary);
  }
  
  .back-button-container {
    display: flex;
    align-items: center;
    padding: 1rem;
    border-bottom: 1px solid var(--light-border-color);
  }
  
  :global([data-theme="dark"]) .back-button-container {
    border-bottom: 1px solid var(--dark-border-color);
  }
  
  .back-button {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 36px;
    height: 36px;
    border-radius: 50%;
    transition: background-color 0.2s ease;
  }
  
  .back-button:hover {
    background-color: var(--light-hover-bg);
  }
  
  :global([data-theme="dark"]) .back-button:hover {
    background-color: var(--dark-hover-bg);
  }
  
  .loader {
    border: 4px solid rgba(0, 0, 0, 0.1);
    border-radius: 50%;
    border-top-color: var(--color-primary);
    width: 40px;
    height: 40px;
    animation: spin 1s linear infinite;
    margin: 0 auto;
  }
  
  :global([data-theme="dark"]) .loader {
    border: 4px solid rgba(255, 255, 255, 0.1);
    border-top-color: var(--color-primary);
  }
  
  .reply-section {
    position: relative;
    padding: 0 1rem;
  }
  
  .reply-separator {
    position: absolute;
    left: 2.5rem;
    top: 0;
    bottom: 0;
    width: 2px;
    background-color: var(--light-border-color);
    z-index: 0;
  }
  
  :global([data-theme="dark"]) .reply-separator {
    background-color: var(--dark-border-color);
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
    background-color: var(--color-primary);
    color: white;
    border: none;
    border-radius: 9999px;
    padding: 0.75rem 1.5rem;
    font-weight: 600;
    margin-top: 1rem;
  }
  
  .reply-button:hover {
    transform: translateY(-1px);
    background-color: var(--color-primary-hover);
    box-shadow: 0 2px 5px rgba(0, 0, 0, 0.1);
  }
  
  :global([data-theme="dark"]) .reply-button {
    background-color: var(--color-primary);
    box-shadow: 0 2px 5px rgba(0, 0, 0, 0.3);
  }
  
  .reply-form-container {
    margin-top: 1rem;
    border-radius: 16px;
    border: 1px solid var(--light-border-color);
  }
  
  :global([data-theme="dark"]) .reply-form-container {
    border-color: var(--dark-border-color);
  }
</style>
