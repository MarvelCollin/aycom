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
  let selectedCategory = '';
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
  
  function setCategory(category: string) {
    if (category.trim()) {
      selectedCategory = category.trim();
      categoryInput = '';
      logger.debug('Set category', { category });
    }
  }
  
  function clearCategory() {
    selectedCategory = '';
    logger.debug('Cleared category');
  }
  
  function handleCategoryKeydown(e: KeyboardEvent) {
    if (e.key === 'Enter' && categoryInput.trim()) {
      e.preventDefault();
      setCategory(categoryInput);
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
        const replyData: any = {
          content: newTweet,
          thread_id: replyTo.threadId || replyTo.thread_id || replyTo.id,
          mentioned_user_ids: selectedCategory ? [selectedCategory] : [],
        };
        
        // Check if we're replying to a reply or to a thread
        // @ts-ignore - parentReplyId might be from our custom enriched objects
        if (replyTo.parentReplyId || replyTo.parent_reply_id) {
          // This is a reply-to-reply
          replyData.parent_reply_id = replyTo.id;
        }
        
        logger.debug('Posting reply with data:', replyData);
        
        // Get the thread ID from the reply or use reply's ID if no thread ID
        const threadId = String(replyTo.threadId || replyTo.thread_id || replyTo.id);
        
        response = await replyToThread(threadId, replyData);
        
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
        
        // Auto-refresh to show the new reply - emit refresh event
        dispatch('refreshReplies', {
          threadId: threadId,
          parentReplyId: replyData.parent_reply_id,
          newReply: response
        });
      } else {
        const data = {
          content: newTweet,
          hashtags: selectedCategory ? [selectedCategory] : [],
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
        
        // Auto-refresh feed
        dispatch('refreshFeed');
      }
      
      newTweet = '';
      files = [];
      selectedCategory = '';
      
      dispatch('tweet', response);
      
      // Close the modal after posting
      closeModal();
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
        c.name !== selectedCategory
      )
    : [];

  async function getSuggestedCategories() {
    if (newTweet.trim().length < 10 || isReplyMode) {
      return;
    }

    isLoadingSuggestions = true;
    showSuggestions = true;
    
    try {
      // Add a timeout to prevent UI hanging if the AI service is slow
      const timeoutPromise = new Promise<any>((_, reject) => 
        setTimeout(() => reject(new Error('AI suggestion timeout')), 10000)
      );
      
      const resultPromise = predictThreadCategory(newTweet);
      
      // Race between the actual API call and timeout
      const result = await Promise.race([resultPromise, timeoutPromise]);
      
      if (result.success && result.all_categories && Object.keys(result.all_categories).length > 0) {
        // Always select the highest confidence category
        const sortedCategories = Object.entries(result.all_categories)
          .sort((a, b) => (b[1] as number) - (a[1] as number));
          
        // Set the top category
        const topCategory = sortedCategories[0][0];
        setCategory(topCategory);
        
        // Get other suggestions to show
        const threshold = 0.05;
        suggestedCategories = sortedCategories
          .slice(1) // Skip the first one as it's already selected
          .filter(([_, confidence]) => (confidence as number) >= threshold)
          .map(([category]) => category);
        
        logger.debug('Got AI suggested categories', { 
          selected: topCategory,
          confidence: result.all_categories[topCategory],
          suggestions: suggestedCategories.length
        });
      } else {
        logger.warn('AI prediction failed:', result.error || 'No categories returned');
        suggestedCategories = [];
        
        // Only show toast for specific errors, not for normal API failures
        if (result.error && !result.error.includes('too short')) {
          toastStore.showToast('Couldn\'t suggest categories', 'warning');
        }
      }
    } catch (error) {
      logger.error('Failed to get AI suggested categories', { error });
      suggestedCategories = [];
      
      // Don't display timeout errors to users
      if (!(error instanceof Error && error.message === 'AI suggestion timeout')) {
        toastStore.showToast('Couldn\'t suggest categories', 'warning');
      }
    } finally {
      isLoadingSuggestions = false;
    }
  }

  function selectSuggestedCategory(category: string) {
    setCategory(category);
    suggestedCategories = suggestedCategories.filter(c => c !== category);
    logger.debug('Selected suggested category', { category });
  }

  let debounceTimeout: ReturnType<typeof setTimeout>;
  $: {
    if (newTweet && !isReplyMode) {
      clearTimeout(debounceTimeout);
      if (newTweet.trim().length >= 10) {
        debounceTimeout = setTimeout(getSuggestedCategories, 1000);
      }
    }
  }
</script>

<div class="compose-tweet-container">
  <div class="compose-tweet-overlay" on:click={closeModal} on:keydown={(e) => e.key === 'Escape' && closeModal()} role="button" tabindex="0"></div>
  <div class="compose-tweet-modal {isDarkMode ? 'compose-tweet-modal-dark' : ''}">
    <div class="compose-tweet-header {isDarkMode ? 'compose-tweet-header-dark' : ''}">
      <button 
        class="compose-tweet-close-btn"
        on:click={closeModal}
        aria-label="Close dialog"
      >
        <XIcon size="24" />
      </button>
      <span class="compose-tweet-title">{modalTitle}</span>
      <div class="compose-tweet-spacer"></div>
    </div>
    <div class="compose-tweet-body {isDarkMode ? 'compose-tweet-body-dark' : ''}">
      {#if isReplyMode && replyTo}
        <div class="compose-tweet-reply-to {isDarkMode ? 'compose-tweet-reply-to-dark' : ''}">
          <div class="compose-tweet-reply-content">
            <div class="compose-tweet-reply-avatar-container">
              {#if typeof replyTo.avatar === 'string' && replyTo.avatar.startsWith('http')}
                <img src={replyTo.avatar} alt={replyTo.username} class="compose-tweet-reply-avatar" />
              {:else}
                <div class="compose-tweet-reply-avatar-placeholder">
                  {replyTo.avatar || 'https://secure.gravatar.com/avatar/0?d=mp'}
                </div>
              {/if}
            </div>
            <div class="compose-tweet-reply-info">
              <div class="compose-tweet-reply-author">
                <span class="compose-tweet-reply-name">{replyTo.displayName || 'User'}</span>
                <span class="compose-tweet-reply-username">@{replyTo.username || 'user'}</span>
              </div>
              <p class="compose-tweet-reply-text">{replyTo.content}</p>
            </div>
          </div>
        </div>
      {/if}
      
      <div class="compose-tweet-input-wrapper">
        <div class="compose-tweet-avatar-wrapper">
          <div class="compose-tweet-avatar">
            <span>{avatar}</span>
          </div>
        </div>
        <div class="compose-tweet-input-area">
          <textarea 
            bind:value={newTweet}
            placeholder="What is happening?!"
            class="compose-tweet-textarea"
            rows="3"
            maxlength={maxWords * 6}
          ></textarea>

          <div class="compose-tweet-word-count">
            <div class="compose-tweet-word-circle">
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
              <span class="compose-tweet-word-count-text {isNearLimit ? 'near-limit' : ''}">{wordCount}</span>
            </div>
            <span class="compose-tweet-word-limit">/ {maxWords} words</span>
          </div>

          {#if files.length > 0}
            <div class="compose-tweet-media-preview">
              <div class="compose-tweet-media-grid {files.length === 1 ? 'single' : ''}">
                {#each files as file, i}
                  <div class="compose-tweet-media-item">
                    {#if file.type.startsWith('image/')}
                      <img src={URL.createObjectURL(file)} alt="preview" class="compose-tweet-media-img" />
                    {:else if file.type.startsWith('video/')}
                      <video src={URL.createObjectURL(file)} class="compose-tweet-media-img">
                        <track kind="captions" src="" label="English" />
                      </video>
                    {:else}
                      <div class="compose-tweet-media-placeholder">
                        <span>{file.name}</span>
                      </div>
                    {/if}
                    <button class="compose-tweet-media-remove" type="button" on:click={() => removeFile(i)}>&times;</button>
                  </div>
                {/each}
              </div>
            </div>
          {/if}

          {#if errorMessage}
            <div class="compose-tweet-error">{errorMessage}</div>
          {/if}

          <div class="compose-tweet-category">
            <div class="compose-tweet-category-header">
              <span class="compose-tweet-category-label">Category</span>
            </div>
            
            {#if selectedCategory}
              <div class="compose-tweet-category-tags">
                <div class="compose-tweet-category-tag {isDarkMode ? 'compose-tweet-category-tag-dark' : ''}">
                  {selectedCategory}
                  <button class="compose-tweet-category-remove" on:click={clearCategory}>×</button>
                </div>
              </div>
            {/if}
            
            <div class="compose-tweet-category-input-wrapper">
              <input 
                type="text" 
                placeholder="Select a category" 
                bind:value={categoryInput} 
                on:keydown={handleCategoryKeydown}
                class="compose-tweet-category-input {isDarkMode ? 'compose-tweet-category-input-dark' : ''}"
                aria-label="Select a category"
              />
              
              {#if filteredCategories.length > 0 && categoryInput.trim()}
                <div class="compose-tweet-category-dropdown {isDarkMode ? 'compose-tweet-category-dropdown-dark' : ''}">
                  {#each filteredCategories as category}
                    <button 
                      class="compose-tweet-category-option {isDarkMode ? 'compose-tweet-category-option-dark' : ''}" 
                      on:click={() => setCategory(category.name)}
                    >
                      {category.name}
                    </button>
                  {/each}
                </div>
              {/if}
            </div>
            
            {#if showSuggestions && suggestedCategories.length > 0}
              <div class="compose-tweet-suggestions">
                <div class="compose-tweet-suggestions-header">
                  <ZapIcon size="14" color="#FBBF24" />
                  <span class="compose-tweet-suggestions-label">Other suggested categories:</span>
                </div>
                <div class="compose-tweet-suggestions-tags">
                  {#each suggestedCategories as category}
                    <button 
                      class="compose-tweet-suggestion-tag {isDarkMode ? 'compose-tweet-suggestion-tag-dark' : ''}"
                      on:click={() => selectSuggestedCategory(category)}
                    >
                      {category}
                    </button>
                  {/each}
                </div>
              </div>
            {/if}
            
            {#if isLoadingSuggestions}
              <div class="compose-tweet-loading">
                <div class="compose-tweet-loading-spinner"></div>
                <span class="compose-tweet-loading-text">Analyzing content...</span>
              </div>
            {/if}
          </div>

          <div class="compose-tweet-reply-settings">
            <select 
              bind:value={replyPermission} 
              class="compose-tweet-reply-select {isDarkMode ? 'compose-tweet-reply-select-dark' : ''}"
            >
              <option value="everyone">Everyone can reply</option>
              <option value="following">Accounts you follow</option>
              <option value="verified">Verified accounts</option>
            </select>
          </div>

          <div class="compose-tweet-actions">
            <div class="compose-tweet-tools">
              <input type="file" multiple accept="image/*,video/*,.gif" on:change={handleFileChange} class="compose-tweet-file-input" id="file-upload" />
              <label for="file-upload" class="compose-tweet-tool" title="Add media">
                <ImageIcon size="18" />
              </label>
              <button class="compose-tweet-tool" title="Add poll">
                <BarChart2Icon size="18" />
              </button>
              <button class="compose-tweet-tool" title="Add emoji">
                <SmileIcon size="18" />
              </button>
              <button class="compose-tweet-tool" title="Add location">
                <MapPinIcon size="18" />
              </button>
              {#if files.length > 0}
                <span class="compose-tweet-file-count">{files.length} {files.length === 1 ? 'file' : 'files'}</span>
              {/if}
            </div>
            <button 
              on:click={handlePost}
              class="compose-tweet-submit"
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
  /* All the existing style can be removed since we're now using the component CSS classes from compose-tweet.css */
</style>