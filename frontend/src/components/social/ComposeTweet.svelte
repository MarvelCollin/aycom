<script lang="ts">
  import { createEventDispatcher, onMount } from 'svelte';
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
  let categories: string[] = [];
  let categoryInput = '';
  let isPosting = false;
  let errorMessage = '';
  let availableCategories: Array<{id: string, name: string}> = [];
  const maxWords = 280;
  
  // Load available categories
  async function loadCategories() {
    try {
      // This would be implemented in a real application
      // For now, we'll use some mock data
      availableCategories = [
        { id: "1", name: "Technology" },
        { id: "2", name: "News" },
        { id: "3", name: "Sports" },
        { id: "4", name: "Entertainment" },
        { id: "5", name: "Politics" },
        { id: "6", name: "Science" },
        { id: "7", name: "Health" },
        { id: "8", name: "Business" }
      ];
      logger.debug('Loaded available categories', { count: availableCategories.length });
    } catch (error) {
      logger.error('Failed to load categories', { error });
    }
  }
  
  // Add a category to the selected list
  function addCategory(category: string) {
    if (!categories.includes(category) && category.trim()) {
      categories = [...categories, category.trim()];
      categoryInput = '';
      logger.debug('Added category', { category, totalCategories: categories.length });
    }
  }
  
  // Remove a category from the selected list
  function removeCategory(category: string) {
    categories = categories.filter(c => c !== category);
    logger.debug('Removed category', { category, totalCategories: categories.length });
  }
  
  // Handle category input keydown events
  function handleCategoryKeydown(e: KeyboardEvent) {
    if (e.key === 'Enter' && categoryInput.trim()) {
      e.preventDefault();
      addCategory(categoryInput);
    }
  }
  
  async function postTweet() {
    if (newTweet.trim() === '') return;
    
    logger.info('Posting new tweet', { 
      contentLength: newTweet.length, 
      filesCount: files.length,
      replyPermission,
      categories
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
      
      // Prepare thread data
      const threadData = {
        content: newTweet,
        who_can_reply: whoCanReply,
        categories: categories // Using the array of category names directly
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
      categories = [];
      categoryInput = '';
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
  
  // Initialize component
  onMount(() => {
    loadCategories();
  });
  
  $: wordCount = newTweet.trim().split(/\s+/).filter(Boolean).length;
  $: wordPercent = Math.min(100, Math.round((wordCount / maxWords) * 100));
  $: isNearLimit = wordCount > maxWords * 0.8;
  $: filteredCategories = categoryInput.trim() 
    ? availableCategories.filter(c => 
        c.name.toLowerCase().includes(categoryInput.toLowerCase()) && 
        !categories.includes(c.name)
      )
    : [];
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

          <!-- Categories Section -->
          <div class="mb-3">
            <!-- Selected Categories -->
            {#if categories.length > 0}
              <div class="flex flex-wrap gap-2 mb-2">
                {#each categories as category}
                  <div class="category-tag {isDarkMode ? 'category-tag-dark' : ''}">
                    {category}
                    <button class="ml-1 text-sm" on:click={() => removeCategory(category)}>×</button>
                  </div>
                {/each}
              </div>
            {/if}
            
            <!-- Category Input with Autocomplete -->
            <div class="relative">
              <input 
                type="text" 
                placeholder="Add categories" 
                bind:value={categoryInput} 
                on:keydown={handleCategoryKeydown}
                class="compose-category-input {isDarkMode ? 'compose-category-input-dark' : ''}"
                aria-label="Add categories"
              />
              
              {#if filteredCategories.length > 0 && categoryInput.trim()}
                <div class="category-dropdown {isDarkMode ? 'category-dropdown-dark' : ''}">
                  {#each filteredCategories as category}
                    <button 
                      class="category-option {isDarkMode ? 'category-option-dark' : ''}" 
                      on:click={() => addCategory(category.name)}
                    >
                      {category.name}
                    </button>
                  {/each}
                </div>
              {/if}
            </div>
          </div>

          <!-- Reply Permission Setting -->
          <div class="mb-3">
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

<style>
  .modal-container {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 50;
  }
  
  .modal-overlay {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background-color: rgba(0, 0, 0, 0.5);
  }
  
  .modal-content {
    position: relative;
    width: 100%;
    max-width: 600px;
    border-radius: 16px;
    overflow: hidden;
    background-color: white;
    box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
  }
  
  .modal-content-dark {
    background-color: #1a202c;
  }
  
  .modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0.75rem 1rem;
    border-bottom: 1px solid #e5e7eb;
  }
  
  .modal-header-dark {
    border-bottom: 1px solid #374151;
  }
  
  .modal-body {
    padding: 1rem;
  }
  
  .compose-textarea {
    width: 100%;
    padding: 0.5rem;
    margin-bottom: 0.5rem;
    border: none;
    resize: none;
    background-color: transparent;
    font-size: 1.25rem;
    outline: none;
    color: #1f2937;
  }
  
  .compose-textarea-dark {
    color: #f3f4f6;
  }
  
  .compose-controls {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding-top: 0.5rem;
    border-top: 1px solid #e5e7eb;
  }
  
  .compose-controls-dark {
    border-top: 1px solid #374151;
  }
  
  .compose-action-btn {
    padding: 0.5rem;
    border-radius: 9999px;
    display: flex;
    align-items: center;
    justify-content: center;
    color: #3b82f6;
    transition: background-color 0.2s;
  }
  
  .compose-action-btn:hover {
    background-color: rgba(59, 130, 246, 0.1);
  }
  
  .compose-action-btn-dark {
    color: #60a5fa;
  }
  
  .compose-action-btn-dark:hover {
    background-color: rgba(96, 165, 250, 0.1);
  }
  
  .compose-submit-btn {
    padding: 0.5rem 1rem;
    border-radius: 9999px;
    background-color: #3b82f6;
    color: white;
    font-weight: 700;
    transition: background-color 0.2s;
  }
  
  .compose-submit-btn:hover:not(:disabled) {
    background-color: #2563eb;
  }
  
  .compose-submit-btn:disabled {
    background-color: #93c5fd;
    cursor: not-allowed;
  }
  
  .compose-word-circle {
    position: relative;
    width: 2rem;
    height: 2rem;
  }
  
  .compose-word-circle-text {
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    font-size: 0.75rem;
    font-weight: 600;
  }
  
  .compose-word-limit {
    color: #ef4444;
  }
  
  .compose-file-preview {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
    margin-bottom: 1rem;
  }
  
  .compose-file-thumb {
    position: relative;
    width: 100px;
    height: 100px;
    border-radius: 0.5rem;
    overflow: hidden;
    background-color: #f3f4f6;
    border: 1px solid #e5e7eb;
  }
  
  .compose-file-thumb-dark {
    background-color: #374151;
    border: 1px solid #4b5563;
  }
  
  .compose-file-remove {
    position: absolute;
    top: 0.25rem;
    right: 0.25rem;
    width: 1.5rem;
    height: 1.5rem;
    border-radius: 9999px;
    background-color: rgba(0, 0, 0, 0.5);
    color: white;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 1rem;
    font-weight: 700;
  }
  
  .compose-category-input,
  .compose-reply-select {
    width: 100%;
    padding: 0.5rem;
    border-radius: 0.375rem;
    border: 1px solid #e5e7eb;
    background-color: white;
    color: #1f2937;
    font-size: 0.875rem;
  }
  
  .compose-category-input-dark,
  .compose-reply-select-dark {
    border: 1px solid #4b5563;
    background-color: #1a202c;
    color: #f3f4f6;
  }

  /* Category Tags */
  .category-tag {
    display: inline-flex;
    align-items: center;
    padding: 0.25rem 0.5rem;
    background-color: #e5e7eb;
    color: #1f2937;
    border-radius: 9999px;
    font-size: 0.875rem;
  }
  
  .category-tag-dark {
    background-color: #374151;
    color: #f3f4f6;
  }
  
  /* Category Dropdown */
  .category-dropdown {
    position: absolute;
    top: 100%;
    left: 0;
    right: 0;
    background-color: white;
    border: 1px solid #e5e7eb;
    border-radius: 0.375rem;
    margin-top: 0.25rem;
    max-height: 12rem;
    overflow-y: auto;
    z-index: 10;
    box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
  }
  
  .category-dropdown-dark {
    background-color: #1a202c;
    border: 1px solid #4b5563;
  }
  
  .category-option {
    display: block;
    width: 100%;
    text-align: left;
    padding: 0.5rem;
    color: #1f2937;
  }
  
  .category-option:hover {
    background-color: #f3f4f6;
  }
  
  .category-option-dark {
    color: #f3f4f6;
  }
  
  .category-option-dark:hover {
    background-color: #374151;
  }
</style>