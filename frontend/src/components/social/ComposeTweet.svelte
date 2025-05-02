<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { 
    ImageIcon, 
    BarChart2Icon, 
    SmileIcon, 
    MapPinIcon
  } from 'svelte-feather-icons';
  import { createThread, uploadThreadMedia } from '../../api/thread';
  import { createLoggerWithPrefix } from '../../utils/logger';
  import { toastStore } from '../../stores/toastStore';
  
  // Create a logger for this component
  const logger = createLoggerWithPrefix('ComposeTweet');
  
  export let avatar = "👤";
  export let isDarkMode = false;
  
  let newTweet = '';
  let files: File[] = [];
  let replyPermission = 'everyone';
  let category = '';
  let isPosting = false;
  let errorMessage = '';
  const maxWords = 280;
  
  async function postTweet() {
    if (newTweet.trim() === '') return;
    
    logger.info('Posting new tweet', { 
      contentLength: newTweet.length, 
      filesCount: files.length,
      replyPermission,
      hasCategory: category.trim() !== ''
    });
    
    isPosting = true;
    errorMessage = '';
    
    try {
      // Check authentication first
      const authData = localStorage.getItem('auth');
      if (!authData) {
        logger.error('Authentication required', { error: 'No authentication data found' }, { showToast: true });
        errorMessage = 'You must be logged in to post. Please log in and try again.';
        isPosting = false;
        return;
      }
      
      try {
        const auth = JSON.parse(authData);
        if (!auth.accessToken) {
          logger.error('Authentication required', { error: 'No access token found' }, { showToast: true });
          errorMessage = 'Your session is invalid. Please log in again.';
          isPosting = false;
          return;
        }
        
        logger.debug('Auth token available for request', { tokenExists: !!auth.accessToken });
      } catch (parseError) {
        logger.error('Authentication error', { error: parseError }, { showToast: true });
        errorMessage = 'Authentication error. Please log in again.';
        isPosting = false;
        return;
      }
      
      // Convert replyPermission to match backend format
      const whoCanReply = replyPermission === 'everyone' 
        ? 'Everyone' 
        : replyPermission === 'following' 
          ? 'Accounts You Follow' 
          : 'Verified Accounts';
      
      // Prepare categories if any
      const categories = category.trim() 
        ? category.split(',').map(cat => cat.trim()).filter(Boolean) 
        : [];
      
      // Prepare thread data
      const threadData = {
        content: newTweet,
        who_can_reply: whoCanReply,
        categories
      };
      
      logger.debug('Creating thread', threadData);
      
      // Post the thread
      const createdThread = await createThread(threadData);
      logger.info('Thread created successfully', { threadId: createdThread.thread_id });
      
      // If there are files, upload them
      if (files.length > 0 && createdThread.thread_id) {
        logger.debug('Uploading media files', { count: files.length, threadId: createdThread.thread_id });
        await uploadThreadMedia(createdThread.thread_id, files);
        logger.info('Media uploaded successfully');
      }
      
      // Reset form
      newTweet = '';
      files = [];
      category = '';
      replyPermission = 'everyone';
      
      // Notify parent
      dispatch('tweet', threadData);
      dispatch('close');
      
      // Show success message
      logger.info('Tweet posted successfully', null, { showToast: true });
    } catch (error) {
      console.error('Error posting tweet:', error);
      toastStore.showToast('Failed to post tweet. Please try again.', 'error');
    } finally {
      isPosting = false;
    }
  }

  function handleFileChange(e: Event) {
    const input = e.target as HTMLInputElement;
    if (input.files) {
      const newFiles = Array.from(input.files);
      files = newFiles;
      logger.debug('Files selected', { count: files.length, types: files.map(f => f.type) });
    }
  }

  function removeFile(index: number) {
    logger.debug('Removing file', { index, fileName: files[index].name });
    files = files.filter((_, i) => i !== index);
  }

  const dispatch = createEventDispatcher();
  function closeModal() { 
    logger.debug('Closing compose modal');
    dispatch('close'); 
  }
  $: wordCount = newTweet.trim().split(/\s+/).filter(Boolean).length;
  $: wordPercent = Math.min(100, Math.round((wordCount / maxWords) * 100));
  $: isNearLimit = wordCount > maxWords * 0.8;
</script>

<div class="modal-container">
  <div class="modal-overlay" on:click={closeModal} on:keydown={(e) => e.key === 'Escape' && closeModal()} role="button" tabindex="0"></div>
  <div class="modal-content {isDarkMode ? 'modal-content-dark' : ''}">
    <div class="modal-header {isDarkMode ? 'modal-header-dark' : ''}">
      <button 
        class="text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200"
        on:click={closeModal}
        aria-label="Close dialog"
      >
        <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
        </svg>
      </button>
      <span class="text-gray-600 dark:text-gray-300">New Post</span>
      <div style="width:2.5rem"></div>
    </div>
    <div class="modal-body {isDarkMode ? 'bg-gray-900' : 'bg-white'}">
      <div class="flex mb-4">
        <div class="w-12 h-12 {isDarkMode ? 'bg-gray-700' : 'bg-gray-300'} rounded-full flex items-center justify-center mr-4 overflow-hidden flex-shrink-0">
          <span>{avatar}</span>
        </div>
        <div class="flex-1">
          <textarea 
            bind:value={newTweet}
            placeholder="What is happening?!"
            class="compose-textarea {isDarkMode ? 'compose-textarea-dark' : ''}"
            rows="3"
            maxlength={maxWords * 6}
          ></textarea>

          <div class="flex items-center gap-2 mb-3">
            <div class="compose-word-circle">
              <svg viewBox="0 0 36 36">
                <path
                  d="M18 2a16 16 0 1 1 0 32 16 16 0 0 1 0-32"
                  fill="none"
                  stroke={isDarkMode ? "#374151" : "#e5e7eb"}
                  stroke-width="4"
                />
                <path
                  d="M18 2a16 16 0 1 1 0 32 16 16 0 0 1 0-32"
                  fill="none"
                  stroke={isNearLimit ? "#ef4444" : "#3b82f6"}
                  stroke-width="4"
                  stroke-dasharray="100, 100"
                  stroke-dashoffset={100 - wordPercent}
                  style="transition: stroke-dashoffset 0.2s ease;"
                />
              </svg>
              <span class="compose-word-circle-text {isDarkMode ? 'text-white' : 'text-gray-900'} {isNearLimit ? 'compose-word-limit' : ''}">{wordCount}</span>
            </div>
            <span class="text-xs {isDarkMode ? 'text-gray-400' : 'text-gray-500'}">/ {maxWords} words</span>
          </div>

          {#if files.length > 0}
            <div class="compose-file-preview">
              {#each files as file, i}
                <div class="compose-file-thumb {isDarkMode ? 'compose-file-thumb-dark' : ''}">
                  {#if file.type.startsWith('image/')}
                    <img src={URL.createObjectURL(file)} alt="preview" class="w-full h-full object-cover" />
                  {:else if file.type.startsWith('video/')}
                    <video src={URL.createObjectURL(file)} class="w-full h-full object-cover">
                      <track kind="captions" src="" label="English" />
                    </video>
                  {:else}
                    <div class="flex items-center justify-center h-full p-2 text-center">
                      <span class="text-xs overflow-hidden">{file.name}</span>
                    </div>
                  {/if}
                  <button class="compose-file-remove" type="button" on:click={() => removeFile(i)}>&times;</button>
                </div>
              {/each}
            </div>
          {/if}

          {#if errorMessage}
            <div class="text-red-500 text-sm mb-3">{errorMessage}</div>
          {/if}

          <div class="flex flex-wrap gap-2 mb-3">
            <input 
              type="text" 
              placeholder="Add categories (comma separated)" 
              bind:value={category} 
              class="compose-category-input {isDarkMode ? 'compose-category-input-dark' : ''}" 
            />
            <select 
              bind:value={replyPermission} 
              class="compose-reply-select {isDarkMode ? 'compose-reply-select-dark' : ''}"
            >
              <option value="everyone">Everyone can reply</option>
              <option value="following">Accounts you follow</option>
              <option value="verified">Verified accounts</option>
            </select>
          </div>

          <div class="compose-controls {isDarkMode ? 'compose-controls-dark' : ''}">
            <div class="flex flex-1 gap-1">
              <input type="file" multiple accept="image/*,video/*,.gif" on:change={handleFileChange} class="hidden" id="file-upload" />
              <label for="file-upload" class="compose-action-btn {isDarkMode ? 'compose-action-btn-dark' : ''}" title="Add media">
                <ImageIcon size="18" />
              </label>
              <button class="compose-action-btn {isDarkMode ? 'compose-action-btn-dark' : ''}" title="Add poll">
                <BarChart2Icon size="18" />
              </button>
              <button class="compose-action-btn {isDarkMode ? 'compose-action-btn-dark' : ''}" title="Add emoji">
                <SmileIcon size="18" />
              </button>
              <button class="compose-action-btn {isDarkMode ? 'compose-action-btn-dark' : ''}" title="Add location">
                <MapPinIcon size="18" />
              </button>
              {#if files.length > 0}
                <span class="text-xs self-center ml-2 {isDarkMode ? 'text-gray-400' : 'text-gray-500'}">{files.length} {files.length === 1 ? 'file' : 'files'}</span>
              {/if}
            </div>
            <button 
              on:click={postTweet}
              class="compose-submit-btn"
              disabled={newTweet.trim() === '' || wordCount > maxWords || isPosting}
            >
              {isPosting ? 'Posting...' : 'Post'}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</div>