<script lang="ts">
  import { onMount, onDestroy, tick } from "svelte";
  import { fade } from "svelte/transition";
  import { getThread, getThreadReplies, getReplyReplies, replyToThread } from "../api/thread";
  import TweetCard from "../components/social/TweetCard.svelte";
  import ComposeTweetModal from "../components/social/ComposeTweetModal.svelte";
  import { toastStore } from "../stores/toastStore";
  import { authStore } from "../stores/authStore";
  import { useTheme } from "../hooks/useTheme";
  import { formatStorageUrl } from "../utils/common";
  import ArrowLeftIcon from "svelte-feather-icons/src/icons/ArrowLeftIcon.svelte";
  import type { ITweet } from "../interfaces/ISocialMedia";
  import type { ExtendedTweet } from "../interfaces/ITweet.extended";
  import { ensureTweetFormat } from "../interfaces/ITweet.extended";
  import MediaOverlay from "../components/media/MediaOverlay.svelte";

  const { theme } = useTheme();

  export let threadId: string;
  export const passedThread: any = null;  

  let thread: ExtendedTweet | null = null;
  let replies: ExtendedTweet[] = [];
  let isLoading = true;
  let nestedRepliesMap = new Map<string, ExtendedTweet[]>();
  let showRepliesMap = new Map<string, boolean>();
  let repliesMap = new Map<string, ExtendedTweet[]>();
  
  let showReplyModal = false;
  let replyToTweet: ITweet | null = null;
  let replyText = '';
  let isSubmitting = false;
  
  let showMediaOverlay = false;
  let currentMediaIndex = 0;
  let currentMediaArray: any[] = [];
  let currentTweet: ExtendedTweet | null = null;
  
  let isDarkMode = false;

  $: isDarkMode = $theme === "dark";

  function normalizeReplyStructure(replies) {
    if (!Array.isArray(replies)) return [];
    
    return replies.map(replyItem => {
      if (replyItem.reply && typeof replyItem.reply === 'object') {
        console.log('DEBUG: Normalizing nested reply structure', {
          before: {
            id: replyItem.id || 'no direct id',
            content: replyItem.content || '(no direct content)',
            nested_id: replyItem.reply.id || 'no nested id',
            nested_content: replyItem.reply.content || '(no nested content)'
          }
        });
        
        const normalizedReply = {
          ...replyItem.reply,
          id: replyItem.reply.id || replyItem.id,
          user: replyItem.user || replyItem.reply.user,
          user_data: replyItem.user_data || replyItem.reply.user_data,
          author: replyItem.author || replyItem.reply.author
        };
        
        console.log('DEBUG: After normalization:', {
          id: normalizedReply.id,
          content: normalizedReply.content || '(still empty)'
        });
        
        return normalizedReply;
      }
      
      return replyItem;
    });
  }

  async function handleLoadReplies(event: CustomEvent<string>) {
    const threadId = event.detail;
    console.log(`Loading replies for thread ${threadId}`);
    
    if (!threadId) {
      console.error('No thread ID provided');
      return;
    }
    
    const isCurrentlyShowing = showRepliesMap.get(threadId) || false;
    showRepliesMap.set(threadId, !isCurrentlyShowing);
    
    if (isCurrentlyShowing) {
      showRepliesMap = new Map(showRepliesMap);
      return;
    }
    
    if (repliesMap.has(threadId)) {
      showRepliesMap = new Map(showRepliesMap);
      return;
    }
    
    try {
      const response = await getThreadReplies(threadId);
      console.log('DEBUG: API response for thread replies:', response);
      if (response && response.replies) {
        console.log(`DEBUG: Received ${response.replies.length} replies for thread ${threadId}`);
        
        if (response.replies.length > 0) {
          console.log('DEBUG: First reply structure:', {
            direct: response.replies[0],
            has_reply_property: typeof (response.replies[0] as any).reply !== 'undefined',
            reply_property: (response.replies[0] as any).reply ? {
              id: (response.replies[0] as any).reply.id,
              content: (response.replies[0] as any).reply.content || '(empty)',
              created_at: (response.replies[0] as any).reply.created_at
            } : 'no reply property',
            direct_content: (response.replies[0] as any).content || '(empty)',
            direct_id: (response.replies[0] as any).id || 'no id',
            user: (response.replies[0] as any).user ? {
              id: (response.replies[0] as any).user.id,
              username: (response.replies[0] as any).user.username
            } : 'no user property'
          });
        }
        
        const normalizedReplies = normalizeReplyStructure(response.replies);
        
        repliesMap.set(threadId, normalizedReplies);
        repliesMap = new Map(repliesMap);
      }
    } catch (err: unknown) {
      const errorMessage = err instanceof Error ? err.message : 'Unknown error';
      console.error(`Error loading replies for thread ${threadId}:`, err);
      toastStore.showToast(`Failed to load replies: ${errorMessage}`, 'error');
    } finally {
      showRepliesMap = new Map(showRepliesMap);
    }
  }

  async function handleReply(event) {
    const threadId = event.detail;
    console.log(`Handling reply to thread: ${threadId}`);
    
    const targetTweet = thread;
    if (!targetTweet) {
      console.error(`Tweet with ID ${threadId} not found`);
      toastStore.showToast('Error finding the tweet to reply to', 'error');
      return;
    }
    
    replyToTweet = targetTweet;
    showReplyModal = true;
  }
  
  async function submitReply() {
    if (!replyToTweet || !replyText.trim()) return;
    
    const typedReplyToTweet = replyToTweet as ExtendedTweet;
    if (!typedReplyToTweet.id) return;
    
    try {
      isSubmitting = true;
      toastStore.showToast('Posting reply...', 'info');
      
      console.log("Attempting to post reply to thread:", typedReplyToTweet.id);
      console.log("Reply content:", replyText);
      
      if (!authStore.isAuthenticated()) {
        throw new Error('Authentication required. Please log in.');
      }
      
      const response = await replyToThread(typedReplyToTweet.id, {
        content: replyText.trim()
      });
      
      console.log("Reply API response:", response);
      
      showReplyModal = false;
      toastStore.showToast('Reply posted successfully!', 'success');
      
      replyToTweet = null;
      replyText = '';
      isSubmitting = false;
      
      try {
        const replyId = typedReplyToTweet.id;
        if (replyId) {
          const updatedReplies = await getThreadReplies(replyId);
          
          if (updatedReplies && updatedReplies.replies) {
            repliesMap.set(replyId, updatedReplies.replies);
            showRepliesMap.set(replyId, true);
            repliesMap = new Map(repliesMap);
            showRepliesMap = new Map(showRepliesMap);
            
            if (thread) {
              thread.replies_count += 1;
              thread = { ...thread }; 
            }
          }
        }
      } catch (refreshErr) {
        console.warn("Error refreshing replies after posting:", refreshErr);
      }
    } catch (error: unknown) {
      const errorMessage = error instanceof Error ? error.message : 'Unknown error';
      console.error('Error posting reply:', error);
      toastStore.showToast(`Failed to post reply: ${errorMessage}`, 'error');
      isSubmitting = false;
    }
  }
  
  function handleReplyModalClose() {
    showReplyModal = false;
    replyToTweet = null;
    replyText = '';
    isSubmitting = false;
  }

  function formatThreadData(responseData): ExtendedTweet {
    try {
      console.log("Formatting thread data from API response:", responseData);
      
      if (!responseData) {
        console.error("Empty or null response data provided to formatThreadData");
        return createEmptyThreadData();
      }

      const formattedId = String(responseData.id || responseData.thread_id || threadId || "");
      
      if (!formattedId) {
        console.error("No valid ID found in thread data");
        return createEmptyThreadData();
      }

      return {
        id: formattedId,
        content: responseData.content || "",
        created_at: responseData.created_at || new Date().toISOString(),
        updated_at: responseData.updated_at || undefined,
        user_id: String(responseData.user_id || responseData.userId || responseData.author_id || ""),
        username: String(responseData.username || responseData.author_username || ""),
        name: String(responseData.name || responseData.display_name || "User"),
        profile_picture_url: responseData.profile_picture_url || "",
        likes_count: Number(responseData.likes_count || 0),
        replies_count: Number(responseData.replies_count || 0),
        reposts_count: Number(responseData.reposts_count || 0),
        bookmark_count: Number(responseData.bookmark_count || 0),
        views_count: Number(responseData.views_count || 0),
        is_liked: Boolean(responseData.is_liked || false),
        is_bookmarked: Boolean(responseData.is_bookmarked || false),
        is_reposted: Boolean(responseData.is_reposted || false),
        is_pinned: Boolean(responseData.is_pinned || false),
        is_verified: Boolean(responseData.is_verified || false),
        media: Array.isArray(responseData.media) 
          ? responseData.media.map(m => ({
              id: String(m.id || ""),
              url: String(m.url || ""),
              type: String(m.type || "image"),
              thumbnail_url: String(m.thumbnail_url || m.url || ""),
              alt_text: ""
            })) 
          : [],
        thread_id: formattedId,
        author_id: String(responseData.user_id || responseData.userId || responseData.author_id || ""),
        parent_id: responseData.parent_id || null,
        community_id: responseData.community_id || null,
        community_name: responseData.community_name || null
      };
    } catch (error) {
      console.error("Error formatting thread data:", error);
      return createEmptyThreadData();
    }
  }
  
  function createEmptyThreadData(): ExtendedTweet {
    return {
      id: String(threadId || ""),
      content: "",
      created_at: new Date().toISOString(),
      updated_at: undefined,
      user_id: "",
      username: "",
      name: "User",
      profile_picture_url: "",
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
      media: [],
      thread_id: String(threadId || ""),
      author_id: "",
      parent_id: null,
      community_id: null,
      community_name: null
    };
  }

  function formatReplyData(replyData): ExtendedTweet | null {
    try {
      if (!replyData) {
        console.error("Empty or null reply data provided to formatReplyData");
        return null;
      }
      
      const formattedId = String(replyData.id || "");
      
      if (!formattedId) {
        console.error("No valid ID found in reply data");
        return null;
      }
      
      return {
        id: formattedId,
        content: replyData.content || "",
        created_at: replyData.created_at || new Date().toISOString(),
        updated_at: replyData.updated_at || null,
        thread_id: String(replyData.thread_id || threadId || ""),
        user_id: String(replyData.user_id || replyData.userId || replyData.author_id || ""),
        username: String(replyData.username || replyData.author_username || ""),
        name: String(replyData.name || replyData.display_name || "User"),
        profile_picture_url: replyData.profile_picture_url || "",
        is_verified: Boolean(replyData.is_verified || false),
        likes_count: Number(replyData.likes_count || 0),
        replies_count: Number(replyData.replies_count || 0),
        reposts_count: Number(replyData.reposts_count || 0),
        bookmark_count: Number(replyData.bookmark_count || 0),
        views_count: Number(replyData.views_count || 0),
        is_liked: Boolean(replyData.is_liked || false),
        is_bookmarked: Boolean(replyData.is_bookmarked || false),
        is_reposted: Boolean(replyData.is_reposted || false),
        is_pinned: Boolean(replyData.is_pinned || false),
        parent_id: replyData.parent_id ? String(replyData.parent_id) : null,
        parent_content: replyData.parent_content || null,
        parent_user: replyData.parent_user ? {
          id: String(replyData.parent_user.id || ""),
          username: String(replyData.parent_user.username || ""),
          name: String(replyData.parent_user.name || ""),
          profile_picture_url: replyData.parent_user.profile_picture_url || ""
        } : null,
        media: Array.isArray(replyData.media) 
          ? replyData.media.map(m => ({
              id: String(m.id || ""),
              url: String(m.url || ""),
              type: String(m.type || "image"),
              thumbnail_url: String(m.thumbnail_url || m.url || ""),
              alt_text: ""
            })) 
          : [],
        author_id: String(replyData.user_id || replyData.userId || replyData.author_id || ""),
        community_id: replyData.community_id || null,
        community_name: replyData.community_name || null
      };
    } catch (error) {
      console.error("Error formatting reply data:", error);
      return null;
    }
  }

  async function loadThreadWithReplies() {
    isLoading = !thread; 
    let apiDataLoaded = false;
    
    try {
      if (!threadId) {
        console.error("No thread ID available");
        isLoading = false;
        return;
      }

      console.log("Loading thread with ID:", threadId);

      let sessionData = null;
      try {
        const storedThread = sessionStorage.getItem("lastViewedThread");
        if (storedThread) {
          const parsedThread = JSON.parse(storedThread);
          if (String(parsedThread.id) === String(threadId)) {
            sessionData = parsedThread;
          }
        }
      } catch (e) {
        console.error("Error accessing sessionStorage:", e);
      }

      const response = await getThread(threadId);
      console.log("Thread data from API:", response);

      if (!response) {
        throw new Error("Failed to load thread data");
      }

      const apiThread = formatThreadData(response);
      console.log("Processed API thread:", apiThread);
      apiDataLoaded = true;

      if (sessionData) {
        thread = {
          ...apiThread,
          likes_count: (sessionData as any).likes_count !== undefined ? (sessionData as any).likes_count : apiThread.likes_count,
          is_liked: (sessionData as any).is_liked !== undefined ? (sessionData as any).is_liked : apiThread.is_liked,
          is_bookmarked: (sessionData as any).is_bookmarked !== undefined ? (sessionData as any).is_bookmarked : apiThread.is_bookmarked,
          is_reposted: (sessionData as any).is_reposted !== undefined ? (sessionData as any).is_reposted : apiThread.is_reposted,
          bookmark_count: (sessionData as any).bookmark_count !== undefined ? (sessionData as any).bookmark_count : apiThread.bookmark_count,
          reposts_count: (sessionData as any).reposts_count !== undefined ? (sessionData as any).reposts_count : apiThread.reposts_count
        };
      } else {
        thread = apiThread;
      }

      if (!thread.username && (sessionData as any)?.username) {
        thread.username = (sessionData as any).username;
        thread.name = (sessionData as any).name || thread.name;
      }

      if ((!thread.content || thread.content.trim() === "") && (sessionData as any)?.content) {
        thread.content = (sessionData as any).content;
      }

      console.log("Final merged thread data:", thread);

      await loadReplies(threadId);

      if (apiDataLoaded) {
        sessionStorage.removeItem("lastViewedThread");
      }

    } catch (error) {
      console.error("Error loading thread:", error);
      toastStore.showToast("Failed to load thread details", "error");

      try {
        const storedThread = sessionStorage.getItem("lastViewedThread");
        if (storedThread && !apiDataLoaded) {
          const parsedThread = JSON.parse(storedThread);

          if (String(parsedThread.id) === String(threadId)) {
            console.log("Using stored thread data as fallback:", parsedThread);
            thread = formatThreadData(parsedThread);
          }
        }
      } catch (storageError) {
        console.error("Error accessing stored thread data for fallback:", storageError);
      }

    } finally {
      isLoading = false;
    }
  }

  async function loadReplies(threadId: string) {
    try {
      const response = await getThreadReplies(threadId);
      console.log("Replies data:", response);

      if (response && response.replies && Array.isArray(response.replies)) {
        console.log(`DEBUG: Received ${response.replies.length} replies for thread ${threadId}`);
        
        if (response.replies.length > 0) {
          console.log('DEBUG: First reply structure:', response.replies[0]);
        }
        
        const normalizedReplies = normalizeReplyStructure(response.replies);
        
        replies = normalizedReplies
          .map(formatReplyData)
          .filter(reply => reply !== null); 

        repliesMap.set(threadId, replies);
        showRepliesMap.set(threadId, true);
        
        if (replies.length > 0) {
          console.log("Processed reply structure:", replies[0]);
        }

        for (const reply of replies) {
          if (reply && reply.replies_count && reply.replies_count > 0) {
            await loadNestedReplies(reply.id);
          }
        }
      }
    } catch (error) {
      console.error("Error loading replies:", error);
    }
  }

  async function loadNestedReplies(replyId: string) {
    try {
      const response = await getReplyReplies(replyId);
      console.log(`Nested replies for ${replyId}:`, response);

      if (response && response.replies && Array.isArray(response.replies)) {    
        const processedReplies = response.replies
          .map(formatReplyData)
          .filter(reply => reply !== null); 

        nestedRepliesMap.set(replyId, processedReplies);
        nestedRepliesMap = new Map(nestedRepliesMap); 
      }
    } catch (error) {
      console.error(`Error loading nested replies for ${replyId}:`, error);
    }
  }

  function handleRefreshReplies(event: any) {
    const { threadId, parentReplyId, newReply } = event.detail;

    if (!newReply) return;

    const processedReply = formatReplyData(newReply);
    if (!processedReply) return;

    if (threadId === thread?.id && !parentReplyId) {
      replies = [processedReply, ...replies];

      if (thread) {
        thread.replies_count = (thread.replies_count || 0) + 1;
      }
      
      const currentReplies = repliesMap.get(threadId) || [];
      repliesMap.set(threadId, [processedReply, ...currentReplies]);
      repliesMap = new Map(repliesMap);
    }
    else if (parentReplyId) {
      const currentNestedReplies = nestedRepliesMap.get(parentReplyId) || [];
      nestedRepliesMap.set(parentReplyId, [processedReply, ...currentNestedReplies]);
      nestedRepliesMap = new Map(nestedRepliesMap);

      const parentReplyIndex = replies.findIndex(r => r.id === parentReplyId);
      if (parentReplyIndex >= 0) {
        replies[parentReplyIndex].replies_count = (replies[parentReplyIndex].replies_count || 0) + 1;
        replies = [...replies];
      }
    }

    showReplyModal = false;
  }

  function handleLike(event) { 
    console.log('Like thread:', event.detail); 
  }
  function handleUnlike(event) { 
    console.log('Unlike thread:', event.detail); 
  }
  function handleBookmark(event) { 
    console.log('Bookmark thread:', event.detail); 
  }
  function handleRemoveBookmark(event) { 
    console.log('Remove bookmark from thread:', event.detail); 
  }
  function handleRepost(event) { 
    console.log('Repost thread:', event.detail); 
  }

  // Handle media overlay functionality
  function openMediaOverlay(tweet: ExtendedTweet, mediaIndex: number = 0) {
    if (!tweet.media || tweet.media.length === 0) return;
    
    currentTweet = tweet;
    currentMediaArray = tweet.media;
    currentMediaIndex = mediaIndex;
    showMediaOverlay = true;
  }

  function closeMediaOverlay() {
    showMediaOverlay = false;
    currentTweet = null;
    currentMediaArray = [];
    currentMediaIndex = 0;
  }

  function handleMediaClick(event: CustomEvent) {
    const { tweet, mediaIndex } = event.detail;
    openMediaOverlay(tweet, mediaIndex || 0);
  }

  // Function to add mock images to the current thread for testing
  function triggerMockImages() {
    if (!thread) return;

    const mockImages = [
      {
        id: "mock-1",
        type: "image",
        url: "https://picsum.photos/800/600?random=1",
        thumbnail_url: "https://picsum.photos/400/300?random=1",
        alt_text: "Mock Image 1"
      },
      {
        id: "mock-2", 
        type: "image",
        url: "https://picsum.photos/800/700?random=2",
        thumbnail_url: "https://picsum.photos/400/350?random=2",
        alt_text: "Mock Image 2"
      },
      {
        id: "mock-3",
        type: "image", 
        url: "https://picsum.photos/900/600?random=3",
        thumbnail_url: "https://picsum.photos/450/300?random=3",
        alt_text: "Mock Image 3"
      }
    ];

    // Add mock images to the thread
    thread = {
      ...thread,
      media: mockImages
    };

    console.log("Mock images added to thread:", thread.media);
    toastStore.showToast("Mock images added! Click on any image to test the overlay.", "success");
  }

  // Initialize component
  onMount(() => {
    // Check if we have thread data in sessionStorage from TweetCard navigation
    try {
      const storedThread = sessionStorage.getItem("lastViewedThread");
      if (storedThread) {
        const parsedThread = JSON.parse(storedThread);

        // Verify this is the correct thread for the current page
        // Compare as strings to avoid type mismatch issues
        if (String(parsedThread.id) === String(threadId)) {
          // Use the stored thread data as an initial render
          thread = formatThreadData(parsedThread);
          console.log("Using thread data from sessionStorage for initial render:", thread);
          
          // If we have complete data from sessionStorage, we'll show it immediately
          // while waiting for the API response
          if (thread.content && thread.username) {
            isLoading = false;
          }
        } else {
          console.log("Stored thread ID does not match current threadId, ignoring stored data");
          console.log(`Stored: ${String(parsedThread.id)}, Current: ${String(threadId)}`);
        }

        // Keep session storage until we confirm API data loads successfully
      }
    } catch (error) {
      console.error("Error parsing stored thread data:", error);
    }

    // Always load fresh thread data from API to ensure it's up-to-date
    loadThreadWithReplies();
  });

  // Clean up subscription on component destruction
  onDestroy(() => {
    // No cleanup needed since we're not using store subscription
  });
</script>

<svelte:head>
  <title>{thread ? `${thread.name || "User"}'s Post` : "Thread"} | AYCOM</title>
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
        <ArrowLeftIcon size="20" class={isDarkMode ? "text-white" : "text-black"} />
      </button>
      <span class="ml-2 font-semibold">Thread</span>
    </div>

    {#if isLoading}
      <div class="p-4 text-center">
        <div class="loader"></div>
        <p class="mt-4 {isDarkMode ? "text-gray-300" : "text-gray-600"}">Loading thread...</p>
      </div>
    {:else if thread}
      <div transition:fade={{ duration: 200 }}>
        <!-- Main Thread -->
        <TweetCard
          tweet={thread}
          {isDarkMode}
          isAuth={authStore.isAuthenticated()}
          showReplies={true}
          {replies}
          {nestedRepliesMap}
          on:reply={handleReply}
          on:like={handleLike}
          on:unlike={handleUnlike}
          on:bookmark={handleBookmark}
          on:removeBookmark={handleRemoveBookmark}
          on:repost={handleRepost}
          on:loadReplies={handleLoadReplies}
          on:mediaClick={handleMediaClick}
        />

        <!-- Reply Modal (improved from Feed.svelte) -->
        <!-- Now handled by inline modal at the bottom of the component -->

        <!-- Reply to Thread Button -->
        <div class="p-4" style="display: flex; gap: 12px; flex-wrap: wrap;">
          <button
            class="reply-button"
            on:click={() => {
              if (!authStore.isAuthenticated()) {
                toastStore.showToast("Please log in to reply", "info");
                return;
              }
              if (thread) {
                replyToTweet = thread;
                showReplyModal = true;
              }
            }}
          >
            Reply to this thread
          </button>

          <!-- Mock Images Test Button -->
          <button
            class="trigger-mock-button"
            on:click={triggerMockImages}
          >
            🖼️ Trigger Mock Images
          </button>
        </div>

        <!-- Debug info to show thread and replies data -->
        {#if thread && import.meta.env?.DEV}
          <details class="p-4 bg-gray-800 text-white text-xs rounded-lg m-2">
            <summary>Debug thread data</summary>
            <pre>{JSON.stringify(thread, null, 2)}</pre>
          </details>
        {/if}

        <!-- Replies Section Display -->
        {#if repliesMap.has(thread.id) && showRepliesMap.get(thread.id)}
          <div class="replies-section">
            <div class="replies-header">
              <h3 class="text-lg font-semibold {isDarkMode ? 'text-white' : 'text-black'}">
                Replies ({thread.replies_count || 0})
              </h3>
            </div>
            
            {#each (repliesMap.get(thread.id) || []) as reply (reply.id)}
              <div class="reply-item">
                <TweetCard
                  tweet={reply}
                  {isDarkMode}
                  isAuth={authStore.isAuthenticated()}
                  nestingLevel={1}
                  showReplies={false}
                  replies={nestedRepliesMap.get(reply.id) || []}
                  {nestedRepliesMap}
                  on:reply={handleReply}
                  on:like={handleLike}
                  on:unlike={handleUnlike}
                  on:bookmark={handleBookmark}
                  on:removeBookmark={handleRemoveBookmark}
                  on:repost={handleRepost}
                  on:loadReplies={handleLoadReplies}
                  on:mediaClick={handleMediaClick}
                />
              </div>
            {/each}
          </div>
        {/if}

        <!-- Replies List with visual connector line (fallback) -->
        {#if replies && replies.length > 0 && !repliesMap.has(thread.id)}
          <div class="reply-section">
            <div class="reply-separator"></div>
            <!-- Replies are rendered by TweetCard component -->
          </div>
        {/if}
      </div>
    {:else}
      <div class="p-4 text-center {isDarkMode ? "text-gray-300" : "text-gray-600"}">
        <p>Thread not found</p>
      </div>
    {/if}
  </div>
</div>

{#if showReplyModal && replyToTweet}
  <div 
    class="aycom-reply-overlay" 
    on:click={handleReplyModalClose}
    role="dialog"
    aria-modal="true"
    aria-labelledby="reply-modal-title"
    on:keydown={(e) => e.key === 'Escape' && handleReplyModalClose()}
    tabindex="-1"
  >
    <div 
      class="aycom-reply-modal {isDarkMode ? 'aycom-dark-theme' : 'aycom-light-theme'}" 
      on:click|stopPropagation
      on:keydown={(e) => e.key === 'Enter' && e.stopPropagation()}
      role="dialog"
      tabindex="-1"
    >
      <div class="aycom-reply-header">
        <h3 id="reply-modal-title" class="aycom-reply-title">
          Reply to @{replyToTweet ? (replyToTweet as ExtendedTweet).username : ''}
        </h3>
        <button 
          class="aycom-reply-close-btn" 
          on:click={handleReplyModalClose}
          aria-label="Close reply dialog"
        >×</button>
      </div>
      
      <div class="aycom-reply-body">
        <div class="aycom-original-tweet">
          <div class="aycom-tweet-user">
            <img 
              src={(replyToTweet as ExtendedTweet).profile_picture_url || "https://secure.gravatar.com/avatar/0?d=mp"} 
              alt={(replyToTweet as ExtendedTweet).name || (replyToTweet as ExtendedTweet).username}
              class="aycom-profile-pic"
            />
            <div class="aycom-user-info">
              <div class="aycom-display-name">{(replyToTweet as ExtendedTweet).name || (replyToTweet as ExtendedTweet).username}</div>
              <div class="aycom-username">@{(replyToTweet as ExtendedTweet).username}</div>
            </div>
          </div>
          <div class="aycom-tweet-content">{(replyToTweet as ExtendedTweet).content}</div>
          
          <!-- Reply line connector -->
          <div class="aycom-reply-connector" aria-hidden="true"></div>
        </div>
        
        <div class="aycom-reply-form">
          <div class="aycom-form-user">
            <img 
              src="https://secure.gravatar.com/avatar/0?d=mp"
              alt="Your profile" 
              class="aycom-profile-pic"
            />
            <div class="aycom-input-container">
              <textarea
                bind:value={replyText}
                placeholder="Post your reply"
                class="aycom-reply-input"
                rows="4"
              ></textarea>
            </div>
          </div>
          
          <div class="aycom-reply-actions">
            <div class="aycom-reply-tools">
              <button class="aycom-tool-btn" title="Add media" aria-label="Add media">
                <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                  <rect x="3" y="3" width="18" height="18" rx="2" ry="2"/>
                  <circle cx="8.5" cy="8.5" r="1.5"/>
                  <polyline points="21 15 16 10 5 21"/>
                </svg>
              </button>
              <button class="aycom-tool-btn" title="Add emoji" aria-label="Add emoji">
                <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                  <circle cx="12" cy="12" r="10"/>
                  <path d="M8 14s1.5 2 4 2 4-2 4-2"/>
                  <line x1="9" y1="9" x2="9.01" y2="9"/>
                  <line x1="15" y1="9" x2="15.01" y2="9"/>
                </svg>
              </button>
            </div>
            <div class="aycom-submit-container">
              <span class="aycom-char-count">{replyText.length} / 280</span>
              <button 
                class="aycom-submit-btn" 
                disabled={!replyText.trim() || isSubmitting} 
                on:click={submitReply}
              >
                {isSubmitting ? 'Posting...' : 'Reply'}
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
{/if}

{#if showMediaOverlay && currentTweet}
  <MediaOverlay 
    isOpen={showMediaOverlay} 
    on:close={() => showMediaOverlay = false}
    currentIndex={currentMediaIndex}
    mediaItems={currentMediaArray}
    threadData={currentTweet}
    {isDarkMode}
  />
{/if}

<style>
  /* Base styles */
  .thread-detail-container {
    min-height: calc(100vh - 60px);
    max-height: 100vh;
    overflow-y: auto;
    padding: 0;
    background-color: var(--light-bg-primary);
    color: var(--light-text-primary);
  }

  .thread-content {
    height: 100%;
    overflow-y: auto;
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

  .replies-section {
    margin-top: 1rem;
    border-top: 1px solid var(--light-border-color);
  }

  :global([data-theme="dark"]) .replies-section {
    border-top: 1px solid var(--dark-border-color);
  }

  .replies-header {
    padding: 1rem;
    background-color: var(--light-bg-secondary);
  }

  :global([data-theme="dark"]) .replies-header {
    background-color: var(--dark-bg-secondary);
  }

  .reply-item {
    border-bottom: 1px solid var(--light-border-color);
  }

  :global([data-theme="dark"]) .reply-item {
    border-bottom: 1px solid var(--dark-border-color);
  }

  .reply-item:last-child {
    border-bottom: none;
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

  /* Reply Modal Styles */
  .aycom-reply-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background-color: rgba(0, 0, 0, 0.75);
    display: flex;
    justify-content: center;
    align-items: center;
    z-index: 1000;
    backdrop-filter: blur(4px);
    -webkit-backdrop-filter: blur(4px);
    animation: fadeIn 0.2s ease-in-out;
  }
  
  .aycom-reply-modal {
    width: 100%;
    max-width: 600px;
    max-height: 90vh;
    border-radius: 16px;
    overflow: hidden;
    box-shadow: 0 8px 30px rgba(0, 0, 0, 0.5);
    animation: slideUp 0.3s ease-out forwards;
    display: flex;
    flex-direction: column;
  }

  .aycom-reply-modal.aycom-dark-theme {
    background-color: #15202b;
    color: #e6e9ef;
  }

  .aycom-reply-modal.aycom-light-theme {
    background-color: #ffffff;
    color: #1c1c1c;
  }
  
  .aycom-reply-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 12px 16px;
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);
    min-height: 53px;
  }

  .aycom-light-theme .aycom-reply-header {
    border-bottom: 1px solid rgba(0, 0, 0, 0.1);
  }
  
  .aycom-reply-title {
    font-size: 20px;
    font-weight: 700;
    margin: 0;
  }
  
  .aycom-reply-close-btn {
    background: none;
    border: none;
    font-size: 24px;
    cursor: pointer;
    color: inherit;
    width: 32px;
    height: 32px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: background-color 0.2s ease;
  }
  
  .aycom-reply-close-btn:hover {
    background-color: rgba(255, 255, 255, 0.1);
  }

  .aycom-light-theme .aycom-reply-close-btn:hover {
    background-color: rgba(0, 0, 0, 0.1);
  }
  
  .aycom-reply-body {
    padding: 0;
    flex: 1;
    overflow-y: auto;
    max-height: calc(90vh - 53px);
  }
  
  .aycom-original-tweet {
    padding: 16px 16px 0;
    position: relative;
  }
  
  .aycom-tweet-user {
    display: flex;
    align-items: flex-start;
    gap: 12px;
    margin-bottom: 12px;
  }
  
  .aycom-profile-pic {
    width: 40px;
    height: 40px;
    border-radius: 50%;
    object-fit: cover;
    flex-shrink: 0;
  }
  
  .aycom-user-info {
    flex: 1;
  }
  
  .aycom-display-name {
    font-weight: 700;
    font-size: 15px;
    line-height: 1.3;
  }
  
  .aycom-username {
    color: #71767b;
    font-size: 15px;
    line-height: 1.3;
  }

  .aycom-light-theme .aycom-username {
    color: #536471;
  }
  
  .aycom-tweet-content {
    font-size: 16px;
    line-height: 1.5;
    margin-left: 52px;
    margin-bottom: 12px;
    white-space: pre-wrap;
    word-wrap: break-word;
  }
  
  .aycom-reply-connector {
    position: absolute;
    left: 36px;
    top: 68px;
    bottom: -12px;
    width: 2px;
    background-color: #2f3336;
  }

  .aycom-light-theme .aycom-reply-connector {
    background-color: #cfd9de;
  }
  
  .aycom-reply-form {
    padding: 4px 16px 16px;
    position: relative;
  }
  
  .aycom-form-user {
    display: flex;
    align-items: flex-start;
    gap: 12px;
  }
  
  .aycom-input-container {
    flex: 1;
  }
  
  .aycom-reply-input {
    width: 100%;
    background: none;
    border: none;
    color: inherit;
    font-size: 20px;
    line-height: 1.5;
    resize: none;
    outline: none;
    font-family: inherit;
    padding: 12px 0;
    min-height: 120px;
  }
  
  .aycom-reply-input::placeholder {
    color: #71767b;
  }

  .aycom-light-theme .aycom-reply-input::placeholder {
    color: #536471;
  }
  
  .aycom-reply-actions {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-top: 12px;
    margin-left: 52px;
  }
  
  .aycom-reply-tools {
    display: flex;
    gap: 16px;
  }
  
  .aycom-tool-btn {
    background: none;
    border: none;
    color: #1d9bf0;
    cursor: pointer;
    padding: 8px;
    border-radius: 50%;
    transition: background-color 0.2s ease;
    display: flex;
    align-items: center;
    justify-content: center;
  }
  
  .aycom-tool-btn:hover {
    background-color: rgba(29, 155, 240, 0.1);
  }
  
  .aycom-submit-container {
    display: flex;
    align-items: center;
    gap: 12px;
  }
  
  .aycom-char-count {
    font-size: 13px;
    color: #71767b;
  }

  .aycom-light-theme .aycom-char-count {
    color: #536471;
  }
  
  .aycom-submit-btn {
    background-color: #1d9bf0;
    color: white;
    border: none;
    border-radius: 9999px;
    font-weight: 700;
    font-size: 15px;
    padding: 8px 16px;
    cursor: pointer;
    transition: background-color 0.2s ease;
    min-width: 80px;
  }
  
  .aycom-submit-btn:hover:not(:disabled) {
    background-color: #1a8cd8;
  }
  
  .aycom-submit-btn:disabled {
    background-color: #1e3a5f;
    cursor: not-allowed;
    opacity: 0.5;
  }

  .aycom-light-theme .aycom-submit-btn:disabled {
    background-color: #cfd9de;
    color: #ffffff;
  }

  @keyframes fadeIn {
    from {
      opacity: 0;
    }
    to {
      opacity: 1;
    }
  }
  
  @keyframes slideUp {
    from {
      transform: translateY(50px);
      opacity: 0;
    }
    to {
      transform: translateY(0);
      opacity: 1;
    }
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

  /* Trigger Mock Images Button */
  .trigger-mock-button {
    transition: all 0.2s;
    display: flex;
    align-items: center;
    justify-content: center;
    background-color: #e1306c;
    color: white;
    border: none;
    border-radius: 9999px;
    padding: 0.75rem 1.5rem;
    font-weight: 600;
    margin-top: 1rem;
    font-size: 14px;
  }

  .trigger-mock-button:hover {
    transform: translateY(-1px);
    background-color: #c12d5a;
    box-shadow: 0 2px 5px rgba(0, 0, 0, 0.1);
  }

  :global([data-theme="dark"]) .trigger-mock-button {
    background-color: #e1306c;
    box-shadow: 0 2px 5px rgba(0, 0, 0, 0.3);
  }
</style>
