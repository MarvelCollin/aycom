<script lang="ts">
  import MainLayout from '../components/layout/MainLayout.svelte';
  import TweetCard from '../components/social/TweetCard.svelte';
  import ComposeTweet from '../components/social/ComposeTweet.svelte';
  import { onMount, tick } from 'svelte';
  import { getAllThreads, getThreadReplies, replyToThread } from '../api/thread';
  import { useTheme } from '../hooks/useTheme';
  import { formatStorageUrl } from '../utils/common';
  import type { ITweet } from '../interfaces/ISocialMedia';
  import type { ExtendedTweet } from '../interfaces/ITweet.extended';
  import { ensureTweetFormat } from '../interfaces/ITweet.extended';
  import { toastStore } from '../stores/toastStore';
  import { authStore } from '../stores/authStore';
  import { fade } from 'svelte/transition';
  import appConfig from '../config/appConfig';

  // Extended interface for needs
  interface Thread {
    id: string;
    content: string;
    created_at: string;
    updated_at?: string;
    username: string;
    name?: string;
    user_id: string;
    profile_picture_url: string;
    replies_count: number;
    likes_count: number;
    is_liked: boolean;
    is_bookmarked: boolean;
    media?: Array<{url: string, type: string}>;
  }

  // Get theme status
  const { theme } = useTheme();
  let isDarkMode: boolean;
  
  // Subscribe to theme changes
  theme.subscribe(val => {
    isDarkMode = val === 'dark';
  });

  let threads: ExtendedTweet[] = [];
  let loading = true;
  let error = null;
  
  // Map to store replies for each thread
  let repliesMap = new Map<string, ExtendedTweet[]>();
  // Map to track which threads have their replies showing
  let showRepliesMap = new Map<string, boolean>();

  // Reply modal state
  let showReplyModal = false;
  let replyToTweet: ITweet | null = null;
  let replyText = '';
  let isSubmitting = false;

  // Function to normalize reply structure - some APIs return nested 'reply' object
  function normalizeReplyStructure(replies) {
    if (!Array.isArray(replies)) return [];
    
    return replies.map(replyItem => {
      // If the item has a 'reply' property containing the actual reply data
      if (replyItem.reply && typeof replyItem.reply === 'object') {
        console.log('DEBUG: Normalizing nested reply structure', {
          before: {
            id: replyItem.id || 'no direct id',
            content: replyItem.content || '(no direct content)',
            nested_id: replyItem.reply.id || 'no nested id',
            nested_content: replyItem.reply.content || '(no nested content)'
          }
        });
        
        // Create a normalized reply object merging the data
        const normalizedReply = {
          ...replyItem.reply,
          id: replyItem.reply.id || replyItem.id,
          // Preserve any user data that's at the root level
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
      
      // If the structure is already flat, return as is
      return replyItem;
    });
  }

  // Function to handle loading replies for a thread
  async function handleLoadReplies(event: CustomEvent<string>) {
    const threadId = event.detail;
    console.log(`Loading replies for thread ${threadId}`);
    
    if (!threadId) {
      console.error('No thread ID provided');
      return;
    }
    
    // Toggle showing replies
    const isCurrentlyShowing = showRepliesMap.get(threadId) || false;
    showRepliesMap.set(threadId, !isCurrentlyShowing);
    
    // If we're hiding replies, just update and return
    if (isCurrentlyShowing) {
      showRepliesMap = new Map(showRepliesMap);
      return;
    }
    
    // If we already have replies, just show them
    if (repliesMap.has(threadId)) {
      showRepliesMap = new Map(showRepliesMap);
      return;
    }
    
    try {
      const response = await getThreadReplies(threadId);
      console.log('DEBUG: API response for thread replies:', response);
      if (response && response.replies) {
        console.log(`DEBUG: Received ${response.replies.length} replies for thread ${threadId}`);
        
        // Inspect structure of the first reply if available
        if (response.replies.length > 0) {
          console.log('DEBUG: First reply structure:', {
            direct: response.replies[0],
            has_reply_property: typeof response.replies[0].reply !== 'undefined',
            reply_property: response.replies[0].reply ? {
              id: response.replies[0].reply.id,
              content: response.replies[0].reply.content || '(empty)',
              created_at: response.replies[0].reply.created_at
            } : 'no reply property',
            direct_content: response.replies[0].content || '(empty)',
            direct_id: response.replies[0].id || 'no id',
            user: response.replies[0].user ? {
              id: response.replies[0].user.id,
              username: response.replies[0].user.username
            } : 'no user property'
          });
        }
        
        // Normalize the reply structure if needed
        const normalizedReplies = normalizeReplyStructure(response.replies);
        
        // Store the normalized replies in our map
        repliesMap.set(threadId, normalizedReplies);
        // Force reactivity
        repliesMap = new Map(repliesMap);
      }
    } catch (err: unknown) {
      const errorMessage = err instanceof Error ? err.message : 'Unknown error';
      console.error(`Error loading replies for thread ${threadId}:`, err);
      toastStore.showToast(`Failed to load replies: ${errorMessage}`, 'error');
    } finally {
      // Force reactivity update
      showRepliesMap = new Map(showRepliesMap);
    }
  }
  
  // Handle like action
  function handleLike(threadId: string) {
    console.log(`Like thread: ${threadId}`);
    // Implement like functionality here
  }

  // Handle unlike action
  function handleUnlike(threadId: string) {
    console.log(`Unlike thread: ${threadId}`);
    // Implement unlike functionality here
  }

  // Handle repost action
  function handleRepost(threadId: string) {
    console.log(`Repost thread: ${threadId}`);
    // Implement repost functionality here
  }

  // Handle bookmark action
  function handleBookmark(threadId: string) {
    console.log(`Bookmark thread: ${threadId}`);
    // Implement bookmark functionality here
  }

  // Handle remove bookmark action
  function handleRemoveBookmark(threadId: string) {
    console.log(`Remove bookmark from thread: ${threadId}`);
    // Implement remove bookmark functionality here
  }

  // Handle thread click - navigate to thread detail page
  function handleThreadClick(event) {
    const tweet = event.detail as ExtendedTweet;
    if (!tweet || !tweet.id) {
      console.error('Invalid tweet data for navigation', tweet);
      return;
    }
    
    window.location.href = `/thread/${tweet.id}`;
  }

  // Handle reply to thread
  async function handleReply(event) {
    const threadId = event.detail;
    console.log(`Handling reply to thread: ${threadId}`);
    
    // Find the tweet to reply to
    const targetTweet = threads.find(t => t.id === threadId);
    if (!targetTweet) {
      console.error(`Tweet with ID ${threadId} not found`);
      toastStore.showToast('Error finding the tweet to reply to', 'error');
      return;
    }
    
    // Set the reply target and show the modal
    replyToTweet = targetTweet;
    showReplyModal = true;
  }
  
  // Handle reply submission from the modal
  async function submitReply() {
    // Proper null check before accessing properties
    if (!replyToTweet || !replyText.trim()) return;
    
    // Type assertion for replyToTweet after the null check
    const typedReplyToTweet = replyToTweet as ExtendedTweet;
    if (!typedReplyToTweet.id) return;
    
    try {
      // Set submitting state
      isSubmitting = true;
      toastStore.showToast('Posting reply...', 'info');
      
      // Debug info
      console.log("Attempting to post reply to thread:", typedReplyToTweet.id);
      console.log("Reply content:", replyText);
      
      if (!authStore.isAuthenticated()) {
        throw new Error('Authentication required. Please log in.');
      }
      
      // Use the imported replyToThread function instead of direct fetch
      const response = await replyToThread(typedReplyToTweet.id, {
        content: replyText.trim()
      });
      
      console.log("Reply API response:", response);
      
      // Close the modal immediately to improve perceived performance
      showReplyModal = false;
      toastStore.showToast('Reply posted successfully!', 'success');
      
      // Reset state
      replyToTweet = null;
      replyText = '';
      isSubmitting = false;
      
      // Refresh data
      try {
        // Store the ID before nulling out replyToTweet
        const replyId = typedReplyToTweet.id;
        if (replyId) {
          const updatedReplies = await getThreadReplies(replyId);
          
          if (updatedReplies && updatedReplies.replies) {
            // Update the replies in our state
            repliesMap.set(replyId, updatedReplies.replies);
            showRepliesMap.set(replyId, true);
            repliesMap = new Map(repliesMap);
            showRepliesMap = new Map(showRepliesMap);
            
            // Update the thread's reply count in the UI
            const targetThread = threads.find(t => t.id === replyId);
            if (targetThread) {
              targetThread.replies_count += 1;
              threads = [...threads]; // Trigger reactivity
            }
          }
        }
      } catch (refreshErr) {
        console.warn("Error refreshing replies after posting:", refreshErr);
        // Don't fail the whole operation if just the refresh failed
      }
    } catch (error: unknown) {
      const errorMessage = error instanceof Error ? error.message : 'Unknown error';
      console.error('Error posting reply:', error);
      toastStore.showToast(`Failed to post reply: ${errorMessage}`, 'error');
      isSubmitting = false;
    }
  }
  
  // Handle modal close
  function handleReplyModalClose() {
    showReplyModal = false;
    replyToTweet = null;
    replyText = '';
    isSubmitting = false;
  }

  // Load threads function
  async function loadThreads() {
    console.log('Loading threads...');
    loading = true;
    error = null;

    try {
      const response = await getAllThreads(1, 20);
      console.log('Thread API response:', response);
      
      if (response && response.success && Array.isArray(response.threads)) {
        threads = response.threads as ExtendedTweet[];
        console.log('Loaded threads:', threads.length);
      } else if (response && Array.isArray(response)) {
        // Handle case where API returns threads directly as array
        threads = response as ExtendedTweet[];
        console.log('Loaded threads directly:', threads.length);
      } else {
        console.error('Invalid API response format:', response);
        threads = [];
        error = 'No threads available right now. Try again later.' as any;
      }
    } catch (err) {
      console.error('Error loading threads:', err);
      if (err instanceof Error && err.message.includes('401')) {
        // If it's an auth error, don't show it to the user, just show empty state
        threads = [];
        error = 'No threads available right now. Try again later.' as any;
      } else {
        // For other errors, show a helpful message
        error = 'Unable to load threads. Please check your connection and try again.' as any;
      }
    } finally {
      loading = false;
    }
  }

  // Convert Thread to ITweet for compatibility with TweetCard
  function threadToTweet(thread: Thread): ExtendedTweet {
    // Map media items to ensure type is one of the allowed values
    // Also format media URLs through formatStorageUrl
    const mappedMedia = (thread.media || []).map(item => ({
      url: formatStorageUrl(item.url),
      type: mapMediaType(item.type),
      alt_text: ''  // Add required alt_text property
    }));
    
    // Process the profile picture URL through formatStorageUrl
    const formattedProfilePicture = formatStorageUrl(thread.profile_picture_url);
    
    return {
      id: thread.id,
      thread_id: thread.id,  // Add thread_id for better compatibility
      content: thread.content,
      created_at: thread.created_at,
      updated_at: thread.updated_at,
      username: thread.username,
      name: thread.name || thread.username,
      user_id: thread.user_id,
      author_id: thread.user_id,  // Add author_id for compatibility
      profile_picture_url: formattedProfilePicture,
      likes_count: thread.likes_count,
      replies_count: thread.replies_count,
      reposts_count: 0,
      bookmark_count: 0,
      views_count: 0,
      is_liked: thread.is_liked,
      is_reposted: false,
      is_bookmarked: thread.is_bookmarked,
      is_pinned: false,
      parent_id: null,
      media: mappedMedia
    };
  }
  
  // Helper function to map media types to allowed values
  function mapMediaType(type: string): 'image' | 'video' | 'gif' {
    if (type === 'video') return 'video';
    if (type === 'gif') return 'gif';
    return 'image'; // Default to image for any other type
  }

  // Load on mount
  onMount(() => {
    loadThreads();
  });
</script>

<MainLayout>
  <div class="feed-container {isDarkMode ? 'feed-container-dark' : ''}">
    <h1 class="feed-title {isDarkMode ? 'feed-title-dark' : ''}">Feed</h1>
    
    {#if loading}
      <div class="loading {isDarkMode ? 'loading-dark' : ''}">
        <div class="loading-spinner"></div>
        <span>Loading threads...</span>
      </div>
    {:else if error}
      <div class="error {isDarkMode ? 'error-dark' : ''}">
        <p>{error}</p>
        <button class="retry-button {isDarkMode ? 'retry-button-dark' : ''}" on:click={loadThreads}>Retry</button>
      </div>
    {:else if threads.length === 0}
      <div class="empty {isDarkMode ? 'empty-dark' : ''}">No threads found</div>
    {:else}
      <div class="threads-list">
        {#each threads as thread (thread.id)}
          <TweetCard 
            tweet={thread}
            isAuth={authStore.isAuthenticated()}
            replies={repliesMap.get(thread.id) || []}
            showReplies={showRepliesMap.get(thread.id) || false}
            on:loadReplies={handleLoadReplies}
            on:like={() => handleLike(thread.id)}
            on:unlike={() => handleUnlike(thread.id)}
            on:repost={() => handleRepost(thread.id)}
            on:bookmark={() => handleBookmark(thread.id)}
            on:removeBookmark={() => handleRemoveBookmark(thread.id)}
            on:reply={handleReply}
            on:click={handleThreadClick}
          />
        {/each}
      </div>
    {/if}
  </div>
</MainLayout>

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
      class="aycom-reply-modal aycom-dark-theme" 
      on:click|stopPropagation
      role="document"
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
              src={(replyToTweet as ExtendedTweet).profile_picture_url || "https://secure.gravatar.com/avatar/0?d=mp"} 
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

<style>
  .feed-container {
    max-width: 800px;
    margin: 0 auto;
    padding: 20px;
    background-color: var(--bg-primary);
    color: var(--text-primary);
  }
  
  .feed-container-dark {
    background-color: var(--bg-primary-dark);
    color: var(--text-primary-dark);
  }

  .feed-title {
    margin-bottom: 1rem;
    font-size: 1.5rem;
    font-weight: 700;
    color: var(--text-primary);
  }
  
  .feed-title-dark {
    color: var(--text-primary-dark);
  }

  .loading, .error, .empty {
    text-align: center;
    padding: 40px;
    font-size: 1.1rem;
    background-color: var(--bg-secondary);
    border-radius: 10px;
    margin: 20px 0;
  }
  
  .loading-dark, .error-dark, .empty-dark {
    background-color: var(--bg-secondary-dark);
    color: var(--text-primary-dark);
  }
  
  .loading {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 1rem;
  }
  
  .loading-spinner {
    width: 40px;
    height: 40px;
    border: 4px solid rgba(0, 0, 0, 0.1);
    border-left-color: var(--color-primary);
    border-radius: 50%;
    animation: spin 1s linear infinite;
  }
  
  .loading-dark .loading-spinner {
    border-color: rgba(255, 255, 255, 0.1);
    border-left-color: var(--color-primary);
  }
  
  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }

  .error button {
    margin-top: 10px;
    padding: 10px 20px;
    background: var(--color-primary);
    color: white;
    border: none;
    border-radius: 5px;
    cursor: pointer;
    font-weight: 600;
    transition: all 0.2s ease;
  }
  
  .error button:hover {
    background: var(--color-primary-hover);
    transform: translateY(-2px);
  }

  .threads-list {
    display: flex;
    flex-direction: column;
    gap: 1px;
    border-radius: 10px;
    overflow: hidden;
    border: 1px solid var(--border-color);
  }
  
  :global(.dark-theme) .threads-list {
    border-color: var(--border-color-dark);
  }
  
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
    overflow: auto;
    box-shadow: 0 8px 30px rgba(0, 0, 0, 0.5);
    background-color: #15202b;
    color: #e6e9ef;
    animation: slideUp 0.3s ease-out forwards;
    display: flex;
    flex-direction: column;
  }
  
  .aycom-reply-body {
    padding: 0;
    flex: 1;
    overflow-y: auto;
    max-height: calc(90vh - 53px); /* Subtract header height */
  }
  
  .aycom-reply-connector {
    position: absolute;
    left: 36px;
    bottom: -16px;
    width: 2px;
    height: 16px;
    background-color: #38444d;
    z-index: 1;
  }
  
  .aycom-tweet-user {
    display: flex;
    align-items: center;
    margin-bottom: 8px;
  }
  
  .aycom-profile-pic {
    width: 40px;
    height: 40px;
    border-radius: 50%;
    object-fit: cover;
    margin-right: 12px;
  }
  
  .aycom-user-info {
    display: flex;
    flex-direction: column;
  }
  
  .aycom-display-name {
    font-weight: 700;
    font-size: 15px;
    color: #e6e9ef;
  }
  
  .aycom-tweet-content {
    font-size: 15px;
    line-height: 1.4;
    white-space: pre-wrap;
    overflow-wrap: break-word;
    color: #e6e9ef;
  }
  
  .aycom-reply-form {
    padding: 16px;
    padding-top: 20px;
  }
  
  .aycom-form-user {
    display: flex;
    align-items: center;
    margin-bottom: 16px;
  }
  
  .aycom-input-container {
    flex: 1;
  }
  
  .aycom-reply-actions {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  
  .aycom-reply-tools {
    display: flex;
    gap: 8px;
  }
  
  .aycom-submit-container {
    display: flex;
    gap: 8px;
  }
  
  .aycom-reply-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 12px 16px;
    border-bottom: 1px solid #38444d;
    position: sticky;
    top: 0;
    background-color: #15202b;
    z-index: 5;
  }
  
  .aycom-reply-title {
    font-size: 18px;
    font-weight: 700;
    margin: 0;
    color: #e6e9ef;
  }
  
  .aycom-reply-close-btn {
    background: transparent;
    border: none;
    color: #8899a6;
    font-size: 24px;
    line-height: 1;
    cursor: pointer;
    padding: 0;
    width: 34px;
    height: 34px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: background-color 0.2s;
  }
  
  .aycom-reply-close-btn:hover {
    background-color: rgba(255, 255, 255, 0.1);
  }
  
  .aycom-reply-input {
    width: 100%;
    padding: 12px;
    border: 1px solid #38444d;
    border-radius: 8px;
    font-size: 16px;
    line-height: 1.5;
    resize: vertical;
    min-height: 100px;
    color: #e6e9ef;
    background-color: #1e2732;
    transition: border-color 0.2s;
  }
  
  .aycom-reply-input:focus {
    outline: none;
    border-color: #1d9bf0;
  }
  
  .aycom-tool-btn {
    background: transparent;
    border: none;
    color: #8899a6;
    font-size: 20px;
    cursor: pointer;
    padding: 0;
    width: 34px;
    height: 34px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: background-color 0.2s;
  }
  
  .aycom-tool-btn:hover {
    background-color: rgba(255, 255, 255, 0.1);
  }
  
  .aycom-char-count {
    color: #8899a6;
    font-size: 14px;
    display: flex;
    align-items: center;
  }
  
  .aycom-submit-btn {
    padding: 10px 20px;
    background: #1d9bf0;
    color: white;
    border: none;
    border-radius: 5px;
    cursor: pointer;
    font-weight: 600;
    transition: all 0.2s ease;
  }
  
  .aycom-submit-btn:hover {
    background: #1a8cd8;
    transform: translateY(-2px);
  }
  
  .aycom-submit-btn:disabled {
    background: #65676b;
    cursor: not-allowed;
    opacity: 0.7;
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
      transform: translateY(30px);
      opacity: 0;
    }
    to {
      transform: translateY(0);
      opacity: 1;
    }
  }

  .aycom-original-tweet {
    padding: 16px;
    border-bottom: 1px solid #38444d;
    position: relative;
  }
  
  .aycom-username {
    color: #8899a6;
    font-size: 14px;
  }
</style>
