<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { fade } from 'svelte/transition';
  import { getThread, getThreadReplies, getReplyReplies } from '../api/thread';
  import TweetCard from '../components/social/TweetCard.svelte';
  import ComposeTweetModal from '../components/social/ComposeTweetModal.svelte';
  import { toastStore } from '../stores/toastStore';
  import { authStore } from '../stores/authStore';
  import { useTheme } from '../hooks/useTheme';
  import ArrowLeftIcon from 'svelte-feather-icons/src/icons/ArrowLeftIcon.svelte';
  import type { ITweet } from '../interfaces/ISocialMedia';
  
  const { theme } = useTheme();
  
  export let threadId: string;
  export let passedThread: any = null; // Accept thread data passed from TweetCard
  
  let thread: any = null;
  let replies: any[] = [];
  let isLoading = true;
  let nestedRepliesMap = new Map();
  let showReplyForm = false;
  let replyTo: any = null;
  let isDarkMode = false;
  
  // Update isDarkMode based on theme
  $: isDarkMode = $theme === 'dark';

  // Function to format thread data from the API response to a consistent structure
  function formatThreadData(responseData) {
    try {
      console.log('Formatting thread data from API response:', responseData);
      
      // Create standardized thread object using the API fields directly
      return {
        id: responseData.id || threadId,
        content: responseData.content || '',
        created_at: responseData.created_at || new Date().toISOString(),
        updated_at: responseData.updated_at,
        user_id: responseData.user_id || '',
        username: responseData.username || 'anonymous',
        name: responseData.name || 'User',
        profile_picture_url: responseData.profile_picture_url || '',
        likes_count: responseData.likes_count || 0,
        replies_count: responseData.replies_count || 0,
        reposts_count: responseData.reposts_count || 0,
        bookmark_count: responseData.bookmark_count || 0,
        views_count: responseData.views_count || 0,
        is_liked: responseData.is_liked || false,
        is_bookmarked: responseData.is_bookmarked || false,
        is_reposted: responseData.is_reposted || false,
        is_pinned: responseData.is_pinned || false,
        is_verified: responseData.is_verified || false,
        media: Array.isArray(responseData.media) ? responseData.media.map(m => ({
          id: m.id,
          url: m.url,
          type: m.type,
          thumbnail_url: m.thumbnail_url || m.url
        })) : []
      };
    } catch (error) {
      console.error('Error formatting thread data:', error);
      // Return a minimal safe object
      return {
        id: threadId,
        content: '',
        created_at: new Date().toISOString(),
        user_id: '',
        username: 'user',
        name: 'User',
        profile_picture_url: '',
        likes_count: 0,
        replies_count: 0,
        reposts_count: 0,
        bookmark_count: 0,
        views_count: 0,
        is_liked: false,
        is_bookmarked: false,
        is_reposted: false,
        is_pinned: false,
        is_verified: false,
        media: []
      };
    }
  }
  
  // Format reply data to a consistent structure
  function formatReplyData(replyData) {
    try {
      return {
        id: replyData.id,
        content: replyData.content || '',
        created_at: replyData.created_at || new Date().toISOString(),
        updated_at: replyData.updated_at,
        thread_id: replyData.thread_id || threadId,
        user_id: replyData.user_id || '',
        username: replyData.username || 'anonymous',
        name: replyData.name || 'User',
        profile_picture_url: replyData.profile_picture_url || '',
        is_verified: replyData.is_verified || false,
        likes_count: replyData.likes_count || 0,
        replies_count: replyData.replies_count || 0,
        reposts_count: replyData.reposts_count || 0,
        bookmark_count: replyData.bookmark_count || 0,
        is_liked: replyData.is_liked || false,
        is_bookmarked: replyData.is_bookmarked || false,
        is_reposted: replyData.is_reposted || false,
        parent_id: replyData.parent_id,
        media: Array.isArray(replyData.media) ? replyData.media.map(m => ({
          id: m.id,
          url: m.url,
          type: m.type,
          thumbnail_url: m.thumbnail_url || m.url
        })) : []
      };
    } catch (error) {
      console.error('Error formatting reply data:', error);
      // Return a basic object with required fields
      return {
        id: replyData.id || '',
        content: replyData.content || '',
        created_at: replyData.created_at || new Date().toISOString(),
        thread_id: replyData.thread_id || threadId,
        user_id: replyData.user_id || '',
        username: replyData.username || 'anonymous',
        name: replyData.name || 'User',
        profile_picture_url: replyData.profile_picture_url || '',
        likes_count: 0,
        replies_count: 0,
        media: []
      };
    }
  }

  // Function to load the thread and its replies
  async function loadThreadWithReplies() {
    isLoading = true;
    try {
      // Check if we have threadId
      if (!threadId) {
        console.error('No thread ID available');
        isLoading = false;
        return;
      }
      
      // Always load fresh thread data from API
      const response = await getThread(threadId);
      console.log('Thread data from API:', response);
      
      if (!response) {
        throw new Error('Failed to load thread data');
      }
      
      // Process the response using our formatter
      thread = formatThreadData(response);
      console.log('Processed thread:', thread);
      
      // Load replies
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
      // Get the replies from the API
      const response = await getThreadReplies(threadId);
      console.log('Replies data:', response);
      
      if (response && response.replies && Array.isArray(response.replies)) {
        // Process replies to make them compatible with TweetCard
        replies = response.replies
          .map(formatReplyData)
          .filter(reply => reply && reply.id); // Filter out any null or invalid replies
        
        // Check structure of first reply to validate
        if (replies.length > 0) {
          console.log('Processed reply structure:', replies[0]);
        }
        
        // Pre-load nested replies for each reply that has them
        for (const reply of replies) {
          if (reply && reply.replies_count && reply.replies_count > 0) {
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
      console.log(`Nested replies for ${replyId}:`, response);
      
      if (response && response.replies && Array.isArray(response.replies)) {
        // Process nested replies using the same formatter
        const processedReplies = response.replies
          .map(formatReplyData)
          .filter(reply => reply && reply.id); // Filter out any null entries
        
        nestedRepliesMap.set(replyId, processedReplies);
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
    
    // Process the new reply using the formatter
    const processedReply = formatReplyData(newReply);
    
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
      replyTo = {
        id: thread.id,
        content: thread.content || '',
        username: thread.username || '',
        name: thread.name || '',
        user_id: thread.user_id || '',
        created_at: thread.created_at || new Date().toISOString()
      };
    } 
    // If replying to a reply
    else {
      // Find the reply either in the main replies array or in nested replies
      const targetReply = replies.find(r => r.id === targetId);
      
      if (targetReply) {
        replyTo = {
          id: targetReply.id,
          thread_id: thread?.id,
          content: targetReply.content || '',
          username: targetReply.username || '',
          name: targetReply.name || '',
          user_id: targetReply.user_id || ''
        };
      } else {
        // Check in nested replies
        for (const [parentId, nestedReplies] of nestedRepliesMap.entries()) {
          if (!Array.isArray(nestedReplies)) continue;
          
          const nestedReply = nestedReplies.find(r => r.id === targetId);
          if (nestedReply) {
            replyTo = {
              id: nestedReply.id,
              thread_id: thread?.id,
              parent_reply_id: parentId,
              content: nestedReply.content || '',
              username: nestedReply.username || '',
              name: nestedReply.name || '',
              user_id: nestedReply.user_id || ''
            };
            break;
          }
        }
      }
    }
    
    showReplyForm = true;
  }
  
  // Handle like, unlike, bookmark, unbookmark events by forwarding them
  function handleLike(event) { dispatch('like', event.detail); }
  function handleUnlike(event) { dispatch('unlike', event.detail); }
  function handleBookmark(event) { dispatch('bookmark', event.detail); }
  function handleRemoveBookmark(event) { dispatch('removeBookmark', event.detail); }
  function handleRepost(event) { dispatch('repost', event.detail); }
  
  // Initialize component
  onMount(() => {
    // Check if we have thread data in sessionStorage from TweetCard navigation
    try {
      const storedThread = sessionStorage.getItem('lastViewedThread');
      if (storedThread) {
        const parsedThread = JSON.parse(storedThread);
        // Use the stored thread data as a quick initial render
        thread = formatThreadData(parsedThread);
        console.log('Using thread data from sessionStorage:', thread);
        // Remove from sessionStorage to avoid stale data on page refresh
        sessionStorage.removeItem('lastViewedThread');
      }
    } catch (error) {
      console.error('Error parsing stored thread data:', error);
    }
    
    // Always load fresh thread data from API to ensure it's up-to-date
    loadThreadWithReplies();
  });
  
  // Clean up subscription on component destruction
  onDestroy(() => {
    // No cleanup needed since we're not using store subscription
  });

  // Svelte's createEventDispatcher workaround
  function dispatch(event, data) {
    const customEvent = new CustomEvent(event, {
      detail: data,
      bubbles: true
    });
    document.dispatchEvent(customEvent);
  }
</script>

<svelte:head>
  <title>{thread ? `${thread.name || 'User'}'s Post` : 'Thread'} | AYCOM</title>
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
          showReplies={true}
          replies={replies}
          {nestedRepliesMap}
          on:reply={handleReply}
          on:like={handleLike}
          on:unlike={handleUnlike}
          on:bookmark={handleBookmark}
          on:removeBookmark={handleRemoveBookmark}
          on:repost={handleRepost}
          on:loadReplies={() => {}}
        />
        
        <!-- Reply Form -->
        {#if showReplyForm && replyTo}
          <div class="reply-form-container">
            <ComposeTweetModal 
              isOpen={showReplyForm}
              avatar={replyTo.profile_picture_url || "https://secure.gravatar.com/avatar/0?d=mp"}
              on:close={() => showReplyForm = false}
              on:posted={(event) => {
                showReplyForm = false;
                // Refresh the thread to show the new reply
                loadThreadWithReplies();
              }}
            />
          </div>
        {:else}
          <div class="p-4">
            <button 
              class="reply-button"
              on:click={() => {
                if (thread) {
                  replyTo = {
                    id: thread.id,
                    content: thread.content || '',
                    username: thread.username || '',
                    name: thread.name || '',
                    user_id: thread.user_id || '',
                    created_at: thread.created_at || new Date().toISOString()
                  };
                showReplyForm = true;
                }
              }}
            >
              Reply to this thread
            </button>
          </div>
        {/if}
        
        <!-- Debug info to show thread and replies data -->
        {#if thread && import.meta.env?.DEV}
          <details class="p-4 bg-gray-800 text-white text-xs rounded-lg m-2">
            <summary>Debug thread data</summary>
            <pre>{JSON.stringify(thread, null, 2)}</pre>
          </details>
        {/if}
        
        <!-- Replies List with visual connector line -->
        {#if replies && replies.length > 0}
          <div class="reply-section">
            <div class="reply-separator"></div>
            <!-- Replies are rendered by TweetCard component -->
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
