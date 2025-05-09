<script lang="ts">
  import { createEventDispatcher, onMount } from 'svelte';
  import { 
    ImageIcon, 
    BarChart2Icon, 
    SmileIcon, 
    MapPinIcon,
    XIcon,
    ZapIcon
  } from 'svelte-feather-icons';
  import { createThread, uploadThreadMedia, replyToThread } from '../../api/thread';
  import { getCategories } from '../../api/categories';
  import { predictThreadCategory } from '../../api/ai';
  import { createLoggerWithPrefix } from '../../utils/logger';
  import { toastStore } from '../../stores/toastStore';
  import { getAuthToken } from '../../utils/auth';
  import appConfig from '../../config/appConfig';
  import type { ITweet } from '../../interfaces/ISocialMedia';
  import { generateFilePreview, handleApiError } from '../../utils/common';
  
  const logger = createLoggerWithPrefix('ComposeTweet');
  
  export let avatar = "https://secure.gravatar.com/avatar/0?d=mp";
  export let isDarkMode = false;
  export let replyTo: ITweet | null = null;
  
  let newTweet = '';
  let files: File[] = [];
  let replyPermission = 'everyone';
  let categories: string[] = [];
  let categoryInput = '';
  let isPosting = false;
  let errorMessage = '';
  let availableCategories: Array<{id: string, name: string}> = [];
  let suggestedCategories: string[] = [];
  let isLoadingSuggestions = false;
  let showSuggestions = false;
  const maxWords = 280;
  
  $: isReplyMode = replyTo !== null;
  $: modalTitle = isReplyMode ? 'Reply to Tweet' : 'New Post';
  
  async function loadCategories() {
    try {
      const data = await getCategories();
      if (data.success) {
        availableCategories = data.categories;
      } else {
        availableCategories = data.categories;
      }
      
      logger.debug('Loaded categories', { count: availableCategories.length });
    } catch (error) {
      logger.error('Failed to load categories', { error });
    }
  }
  
  function addCategory(category: string) {
    if (!categories.includes(category) && category.trim()) {
      categories = [...categories, category.trim()];
      categoryInput = '';
      logger.debug('Added category', { category, totalCategories: categories.length });
    }
  }
  
  function removeCategory(category: string) {
    categories = categories.filter(c => c !== category);
    logger.debug('Removed category', { category, totalCategories: categories.length });
  }
  
  function handleCategoryKeydown(e: KeyboardEvent) {
    if (e.key === 'Enter' && categoryInput.trim()) {
      e.preventDefault();
      addCategory(categoryInput);
    }
  }
  
  async function handlePost() {
    if (newTweet.trim() === '' && files.length === 0) {
      errorMessage = 'Your tweet cannot be empty';
      return;
    }
    
    isPosting = true;
    errorMessage = '';
    
    try {
      let response;
      
      if (isReplyMode && replyTo) {
        const replyData = {
          content: newTweet,
          thread_id: replyTo.threadId || replyTo.id,
          parent_id: replyTo.id,
          mentioned_user_ids: categories,
        };
        
        logger.debug('Posting reply with data:', replyData);
        
        response = await replyToThread(replyTo.id.toString(), replyData);
        
        if (files.length > 0 && response.id) {
          try {
            logger.debug(`Uploading ${files.length} media files for reply ${response.id}`);
            await uploadThreadMedia(response.id, files);
          } catch (uploadError) {
            logger.error('Error uploading media for reply:', uploadError);
            toastStore.showToast('Your reply was created but media upload failed', 'warning');
          }
        }
        
        toastStore.showToast('Your reply was posted successfully', 'success');
      } else {
        const data = {
          content: newTweet,
          hashtags: categories,
        };
        
        logger.debug('Posting tweet with data:', data);
        
        response = await createThread(data);
        
        if (files.length > 0 && response.id) {
          try {
            logger.debug(`Uploading ${files.length} media files for thread ${response.id}`);
            await uploadThreadMedia(response.id, files);
          } catch (uploadError) {
            logger.error('Error uploading media:', uploadError);
            toastStore.showToast('Your post was created but media upload failed', 'warning');
          }
        }
        
        toastStore.showToast('Your post was created successfully', 'success');
      }
      
      newTweet = '';
      files = [];
      categories = [];
      
      dispatch('tweet', response);
    } catch (error) {
      logger.error('Error creating tweet/reply:', error);
      const errorResponse = handleApiError(error);
      errorMessage = errorResponse.message;
      toastStore.showToast(errorMessage, 'error');
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

  function getFilePreview(file: File) {
    return generateFilePreview(file);
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

  async function getSuggestedCategories() {
    if (newTweet.trim().length < 10) {
      return;
    }

    isLoadingSuggestions = true;
    showSuggestions = true;
    
    try {
      const result = await predictThreadCategory(newTweet);
      
      if (result.success) {
        const threshold = 0.05;
        suggestedCategories = Object.entries(result.all_categories)
          .filter(([_, confidence]) => (confidence as number) >= threshold)
          .map(([category]) => category)
          .filter(category => !categories.includes(category));
        
        logger.debug('Got AI suggested categories', { count: suggestedCategories.length });
      } else {
        suggestedCategories = [];
      }
    } catch (error) {
      logger.error('Failed to get AI suggested categories', { error });
      suggestedCategories = [];
    } finally {
      isLoadingSuggestions = false;
    }
  }

  function addSuggestedCategory(category: string) {
    if (!categories.includes(category)) {
      categories = [...categories, category];
      suggestedCategories = suggestedCategories.filter(c => c !== category);
      logger.debug('Added suggested category', { category });
    }
  }

  let debounceTimeout: ReturnType<typeof setTimeout>;
  $: {
    if (newTweet && !isReplyMode) {
      clearTimeout(debounceTimeout);
      debounceTimeout = setTimeout(getSuggestedCategories, 500);
    }
  }
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
        <XIcon size="24" />
      </button>
      <span class="text-gray-600 dark:text-gray-300">{modalTitle}</span>
      <div style="width:2.5rem"></div>
    </div>
    <div class="modal-body {isDarkMode ? 'bg-gray-900' : 'bg-white'}">
      {#if isReplyMode && replyTo}
        <div class="reply-to-container mb-3 p-3 {isDarkMode ? 'bg-gray-800 border-gray-700' : 'bg-gray-100 border-gray-200'} border rounded-lg">
          <div class="flex items-start">
            <div class="flex-shrink-0 mr-2">
              {#if typeof replyTo.avatar === 'string' && replyTo.avatar.startsWith('http')}
                <img src={replyTo.avatar} alt={replyTo.username} class="w-8 h-8 rounded-full" />
              {:else}
                <div class="w-8 h-8 rounded-full {isDarkMode ? 'bg-gray-700' : 'bg-gray-300'} flex items-center justify-center">
                  {replyTo.avatar || 'https://secure.gravatar.com/avatar/0?d=mp'}
                </div>
              {/if}
            </div>
            <div class="flex-1 min-w-0 text-sm">
              <div class="flex items-center">
                <span class="font-bold {isDarkMode ? 'text-white' : 'text-black'} mr-1">{replyTo.displayName || 'User'}</span>
                <span class="{isDarkMode ? 'text-gray-400' : 'text-gray-500'} truncate">@{replyTo.username || 'user'}</span>
              </div>
              <p class="text-sm truncate {isDarkMode ? 'text-gray-300' : 'text-gray-600'}">{replyTo.content}</p>
            </div>
          </div>
        </div>
      {/if}
      
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

          <div class="mb-3">
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
            
            {#if showSuggestions && suggestedCategories.length > 0}
              <div class="mt-2">
                <div class="flex items-center gap-1 mb-1">
                  <ZapIcon size="14" color="#FBBF24" />
                  <span class="text-xs {isDarkMode ? 'text-gray-300' : 'text-gray-600'}">AI-suggested categories:</span>
                </div>
                <div class="flex flex-wrap gap-2">
                  {#each suggestedCategories as category}
                    <button 
                      class="suggested-category-tag {isDarkMode ? 'suggested-category-tag-dark' : ''}"
                      on:click={() => addSuggestedCategory(category)}
                    >
                      {category}
                    </button>
                  {/each}
                </div>
              </div>
            {/if}
            
            {#if isLoadingSuggestions}
              <div class="mt-2 flex items-center gap-1">
                <div class="loading-spinner"></div>
                <span class="text-xs {isDarkMode ? 'text-gray-300' : 'text-gray-600'}">Analyzing content...</span>
              </div>
            {/if}
          </div>

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
              on:click={handlePost}
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

  .suggested-category-tag {
    display: inline-flex;
    align-items: center;
    padding: 0.25rem 0.5rem;
    background-color: rgba(59, 130, 246, 0.15);
    color: #1f2937;
    border: 1px dashed #3b82f6;
    border-radius: 9999px;
    font-size: 0.875rem;
    margin-top: 0.25rem;
    margin-right: 0.25rem;
    transition: all 0.2s ease;
  }
  
  .suggested-category-tag:hover {
    background-color: rgba(59, 130, 246, 0.25);
    border-style: solid;
  }
  
  .suggested-category-tag-dark {
    background-color: rgba(96, 165, 250, 0.15);
    color: #f3f4f6;
    border-color: #60a5fa;
  }
  
  .suggested-category-tag-dark:hover {
    background-color: rgba(96, 165, 250, 0.25);
  }
  
  .loading-spinner {
    width: 14px;
    height: 14px;
    border: 2px solid rgba(59, 130, 246, 0.3);
    border-radius: 50%;
    border-top-color: #3b82f6;
    animation: spin 1s linear infinite;
  }
  
  @keyframes spin {
    to { transform: rotate(360deg); }
  }
</style>