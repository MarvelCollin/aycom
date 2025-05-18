<script lang="ts">
  import { createEventDispatcher, onMount } from 'svelte';
  import { 
    ImageIcon, 
    BarChart2Icon, 
    SmileIcon, 
    MapPinIcon,
    XIcon,
    ZapIcon,
    AlertCircleIcon,
    UsersIcon
  } from 'svelte-feather-icons';
  import { createThread, uploadThreadMedia, replyToThread } from '../../api/thread';
  import { getCategories } from '../../api/categories';
  import { getCommunities } from '../../api/community';
  import { predictThreadCategory } from '../../api/ai';
  import { createLoggerWithPrefix } from '../../utils/logger';
  import { toastStore } from '../../stores/toastStore';
  import { getAuthToken, getUserRole } from '../../utils/auth';
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
  let availableCommunities: Array<{id: string, name: string}> = [];
  let suggestedCategories: string[] = [];
  let isLoadingSuggestions = false;
  let showSuggestions = false;
  let showScheduleOptions = false;
  let showCommunityOptions = false;
  let showPollOptions = false;
  let scheduledDate = '';
  let scheduledTime = '';
  let selectedCommunityId = '';
  let isAdmin = false;
  let isAdvertisement = false;
  const maxWords = 280;
  
  // Poll options
  let pollQuestion = '';
  let pollOptions = ['', ''];
  let pollExpiryHours = 24;
  let pollWhoCanVote = 'everyone';
  
  $: isReplyMode = replyTo !== null;
  $: modalTitle = isReplyMode ? 'Reply to Tweet' : 'New Post';
  $: canSchedule = !isReplyMode; // Only allow scheduling for new posts, not replies
  $: canPostForCommunity = !isReplyMode; // Only allow community posts for new posts, not replies
  
  // Use AlertCircleIcon for scheduling since it exists in the library
  const CalendarIcon = AlertCircleIcon;
  
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

  async function loadCommunities() {
    try {
      const data = await getCommunities();
      if (data.success) {
        availableCommunities = data.communities;
      } else {
        availableCommunities = [];
      }
      
      logger.debug('Loaded communities', { count: availableCommunities.length });
    } catch (error) {
      logger.error('Failed to load communities', { error });
    }
  }

  async function checkUserRole() {
    try {
      const role = await getUserRole();
      isAdmin = role === 'admin';
      logger.debug('Checked user role', { isAdmin });
    } catch (error) {
      logger.error('Failed to check user role', { error });
      isAdmin = false;
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

  function toggleScheduleOptions() {
    showScheduleOptions = !showScheduleOptions;
    if (showScheduleOptions) {
      // Set default time to current time + 1 hour
      const now = new Date();
      const tomorrow = new Date(now.getTime() + 24 * 60 * 60 * 1000);
      scheduledDate = tomorrow.toISOString().split('T')[0];
      
      const hours = now.getHours().toString().padStart(2, '0');
      const minutes = now.getMinutes().toString().padStart(2, '0');
      scheduledTime = `${hours}:${minutes}`;
    }
  }

  function toggleCommunityOptions() {
    showCommunityOptions = !showCommunityOptions;
  }

  function togglePollOptions() {
    showPollOptions = !showPollOptions;
    if (!showPollOptions) {
      // Reset poll state when closing
      pollQuestion = '';
      pollOptions = ['', ''];
      pollExpiryHours = 24;
      pollWhoCanVote = 'everyone';
    }
  }

  function addPollOption() {
    if (pollOptions.length < 4) {
      pollOptions = [...pollOptions, ''];
    }
  }

  function removePollOption(index: number) {
    if (pollOptions.length > 2) {
      pollOptions = pollOptions.filter((_, i) => i !== index);
    }
  }
  
  async function handlePost() {
    if (newTweet.trim() === '' && files.length === 0 && !showPollOptions) {
      errorMessage = 'Your post cannot be empty';
      return;
    }
    
    if (showPollOptions && (pollQuestion.trim() === '' || pollOptions.filter(opt => opt.trim() !== '').length < 2)) {
      errorMessage = 'Poll question and at least 2 options are required';
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
        const data: any = {
          content: newTweet,
          hashtags: selectedCategory ? [selectedCategory] : [],
          who_can_reply: replyPermission,
          is_advertisement: isAdmin && isAdvertisement,
        };
        
        // Add scheduled date if set
        if (showScheduleOptions && scheduledDate && scheduledTime) {
          const scheduledDateTime = new Date(`${scheduledDate}T${scheduledTime}`);
          if (!isNaN(scheduledDateTime.getTime())) {
            data.scheduled_at = scheduledDateTime.toISOString();
          }
        }
        
        // Add community ID if selected
        if (showCommunityOptions && selectedCommunityId) {
          data.community_id = selectedCommunityId;
        }
        
        // Add poll if created
        if (showPollOptions && pollQuestion.trim() && pollOptions.filter(opt => opt.trim() !== '').length >= 2) {
          const validOptions = pollOptions.filter(opt => opt.trim() !== '');
          data.poll = {
            question: pollQuestion,
            options: validOptions,
            closes_at: new Date(Date.now() + pollExpiryHours * 60 * 60 * 1000).toISOString(),
            who_can_vote: pollWhoCanVote
          };
        }
        
        logger.debug('Posting thread with data:', data);
        
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
      showScheduleOptions = false;
      showCommunityOptions = false;
      showPollOptions = false;
      scheduledDate = '';
      scheduledTime = '';
      selectedCommunityId = '';
      pollQuestion = '';
      pollOptions = ['', ''];
      pollExpiryHours = 24;
      pollWhoCanVote = 'everyone';
      isAdvertisement = false;
      
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
    loadCommunities();
    checkUserRole();
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
      
      {#if canPostForCommunity && showCommunityOptions}
        <div class="compose-tweet-community-selection {isDarkMode ? 'compose-tweet-community-selection-dark' : ''}">
          <h4>Post to Community</h4>
          <select 
            bind:value={selectedCommunityId} 
            class="compose-tweet-community-select {isDarkMode ? 'compose-tweet-community-select-dark' : ''}"
          >
            <option value="">Select a community</option>
            {#each availableCommunities as community}
              <option value={community.id}>{community.name}</option>
            {/each}
          </select>
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
          
          {#if showPollOptions}
            <div class="compose-tweet-poll-options {isDarkMode ? 'compose-tweet-poll-options-dark' : ''}">
              <div class="compose-tweet-poll-question">
                <input 
                  type="text" 
                  placeholder="Ask a question..." 
                  bind:value={pollQuestion}
                  class="compose-tweet-poll-question-input {isDarkMode ? 'compose-tweet-poll-question-input-dark' : ''}"
                />
              </div>
              <div class="compose-tweet-poll-choices">
                {#each pollOptions as option, index}
                  <div class="compose-tweet-poll-choice">
                    <input 
                      type="text" 
                      placeholder={`Option ${index + 1}`}
                      bind:value={pollOptions[index]}
                      class="compose-tweet-poll-choice-input {isDarkMode ? 'compose-tweet-poll-choice-input-dark' : ''}"
                    />
                    {#if index > 1 && pollOptions.length > 2}
                      <button 
                        class="compose-tweet-poll-choice-remove"
                        on:click={() => removePollOption(index)}
                      >×</button>
                    {/if}
                  </div>
                {/each}
              </div>
              {#if pollOptions.length < 4}
                <button 
                  class="compose-tweet-poll-add-option {isDarkMode ? 'compose-tweet-poll-add-option-dark' : ''}"
                  on:click={addPollOption}
                >
                  + Add Option
                </button>
              {/if}
              <div class="compose-tweet-poll-settings">
                <div class="compose-tweet-poll-duration">
                  <label for="poll-duration">Poll duration:</label>
                  <select 
                    id="poll-duration"
                    bind:value={pollExpiryHours}
                    class="compose-tweet-poll-select {isDarkMode ? 'compose-tweet-poll-select-dark' : ''}"
                  >
                    <option value={1}>1 hour</option>
                    <option value={6}>6 hours</option>
                    <option value={12}>12 hours</option>
                    <option value={24}>24 hours</option>
                    <option value={72}>3 days</option>
                    <option value={168}>7 days</option>
                  </select>
                </div>
                <div class="compose-tweet-poll-who-can-vote">
                  <label for="poll-who-can-vote">Who can vote:</label>
                  <select 
                    id="poll-who-can-vote"
                    bind:value={pollWhoCanVote}
                    class="compose-tweet-poll-select {isDarkMode ? 'compose-tweet-poll-select-dark' : ''}"
                  >
                    <option value="everyone">Everyone</option>
                    <option value="following">Accounts you follow</option>
                    <option value="verified">Verified accounts</option>
                  </select>
                </div>
              </div>
            </div>
          {/if}

          {#if canSchedule && showScheduleOptions}
            <div class="compose-tweet-schedule {isDarkMode ? 'compose-tweet-schedule-dark' : ''}">
              <h4>Schedule post</h4>
              <div class="compose-tweet-schedule-inputs">
                <input 
                  type="date" 
                  bind:value={scheduledDate}
                  min={new Date().toISOString().split('T')[0]}
                  class="compose-tweet-schedule-date {isDarkMode ? 'compose-tweet-schedule-date-dark' : ''}"
                />
                <input 
                  type="time" 
                  bind:value={scheduledTime}
                  class="compose-tweet-schedule-time {isDarkMode ? 'compose-tweet-schedule-time-dark' : ''}"
                />
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
          
          {#if isAdmin && !isReplyMode}
            <div class="compose-tweet-admin-options">
              <label class="compose-tweet-admin-option">
                <input type="checkbox" bind:checked={isAdvertisement} />
                <span>Mark as advertisement</span>
              </label>
            </div>
          {/if}

          <div class="compose-tweet-actions">
            <div class="compose-tweet-tools">
              <input type="file" multiple accept="image/*,video/*,.gif" on:change={handleFileChange} class="compose-tweet-file-input" id="file-upload" />
              <label for="file-upload" class="compose-tweet-tool" title="Add media">
                <ImageIcon size="18" />
              </label>
              <button class="compose-tweet-tool" title="Add poll" on:click={togglePollOptions}>
                <div class={showPollOptions ? 'active-tool' : ''}>
                <BarChart2Icon size="18" />
                </div>
              </button>
              <button class="compose-tweet-tool" title="Add emoji">
                <SmileIcon size="18" />
              </button>
              {#if canSchedule}
                <button class="compose-tweet-tool" title="Schedule post" on:click={toggleScheduleOptions}>
                  <div class={showScheduleOptions ? 'active-tool' : ''}>
                    <CalendarIcon size="18" />
                  </div>
              </button>
              {/if}
              {#if canPostForCommunity}
                <button class="compose-tweet-tool" title="Post to community" on:click={toggleCommunityOptions}>
                  <div class={showCommunityOptions ? 'active-tool' : ''}>
                    <UsersIcon size="18" />
                  </div>
                </button>
              {/if}
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
  /* Main container & overlay */
  .compose-tweet-container {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
    pointer-events: none; /* Allow clicks to pass through container */
  }
  
  .compose-tweet-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background-color: rgba(0, 0, 0, 0.5); /* Semi-transparent overlay */
    backdrop-filter: blur(2px); /* Slight blur effect */
    z-index: 1000;
    pointer-events: auto; /* Capture clicks on overlay */
  }
  
  /* Modal */
  .compose-tweet-modal {
    position: relative;
    width: 600px;
    max-width: 95vw;
    max-height: 90vh;
    background-color: white;
    border-radius: 16px;
    box-shadow: 0 8px 30px rgba(0, 0, 0, 0.12);
    overflow-y: auto;
    z-index: 1001;
    pointer-events: auto; /* Ensure modal captures clicks */
  }
  
  .compose-tweet-modal-dark {
    background-color: #1e293b;
    color: #f1f5f9;
  }
  
  /* Header */
  .compose-tweet-header {
    display: flex;
    align-items: center;
    padding: 12px 16px;
    border-bottom: 1px solid #e5e7eb;
    position: sticky;
    top: 0;
    background-color: white;
    z-index: 10;
  }
  
  .compose-tweet-header-dark {
    background-color: #1e293b;
    border-bottom: 1px solid #384152;
  }
  
  .compose-tweet-close-btn {
    background: none;
    border: none;
    cursor: pointer;
    color: #374151;
    padding: 8px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
  }
  
  .compose-tweet-header-dark .compose-tweet-close-btn {
    color: #e5e7eb;
  }
  
  .compose-tweet-close-btn:hover {
    background-color: rgba(0, 0, 0, 0.05);
  }
  
  .compose-tweet-header-dark .compose-tweet-close-btn:hover {
    background-color: rgba(255, 255, 255, 0.1);
  }
  
  .compose-tweet-title {
    margin-left: 16px;
    font-size: 18px;
    font-weight: 700;
  }
  
  .compose-tweet-spacer {
    flex: 1;
  }
  
  /* Body */
  .compose-tweet-body {
    padding: 16px;
  }
  
  .compose-tweet-body-dark {
    color: #f1f5f9;
  }
  
  /* Reply section */
  .compose-tweet-reply-to {
    padding: 12px;
    border-bottom: 1px solid #e5e7eb;
    margin-bottom: 16px;
  }
  
  .compose-tweet-reply-to-dark {
    border-bottom: 1px solid #384152;
  }
  
  .compose-tweet-reply-content {
    display: flex;
  }
  
  .compose-tweet-reply-avatar-container {
    margin-right: 12px;
  }
  
  .compose-tweet-reply-avatar {
    width: 40px;
    height: 40px;
    border-radius: 50%;
    object-fit: cover;
  }
  
  .compose-tweet-reply-avatar-placeholder {
    width: 40px;
    height: 40px;
    border-radius: 50%;
    background-color: #e5e7eb;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 12px;
  }
  
  .compose-tweet-reply-info {
    flex: 1;
  }
  
  .compose-tweet-reply-author {
    margin-bottom: 4px;
  }
  
  .compose-tweet-reply-name {
    font-weight: 700;
    margin-right: 4px;
  }
  
  .compose-tweet-reply-username {
    color: #6b7280;
  }
  
  .compose-tweet-reply-text {
    color: #374151;
    font-size: 15px;
  }
  
  .compose-tweet-body-dark .compose-tweet-reply-username {
    color: #9ca3af;
  }
  
  .compose-tweet-body-dark .compose-tweet-reply-text {
    color: #e5e7eb;
  }
  
  /* Input area */
  .compose-tweet-input-wrapper {
    display: flex;
  }
  
  .compose-tweet-avatar-wrapper {
    margin-right: 12px;
    flex-shrink: 0;
  }
  
  .compose-tweet-avatar {
    width: 48px;
    height: 48px;
    border-radius: 50%;
    background-color: #e5e7eb;
    display: flex;
    align-items: center;
    justify-content: center;
    overflow: hidden;
  }
  
  .compose-tweet-input-area {
    flex: 1;
  }
  
  .compose-tweet-textarea {
    width: 100%;
    min-height: 120px;
    padding: 8px 0;
    border: none;
    resize: none;
    font-size: 20px;
    color: #374151;
    background: transparent;
    margin-bottom: 16px;
  }
  
  .compose-tweet-body-dark .compose-tweet-textarea {
    color: #f1f5f9;
  }
  
  .compose-tweet-textarea:focus {
    outline: none;
  }
  
  .compose-tweet-textarea::placeholder {
    color: #9ca3af;
  }
  
  /* Word counter */
  .compose-tweet-word-count {
    display: flex;
    align-items: center;
    margin-bottom: 16px;
  }
  
  .compose-tweet-word-circle {
    position: relative;
    width: 32px;
    height: 32px;
    margin-right: 8px;
  }
  
  .compose-tweet-word-circle svg {
    transform: rotate(-90deg);
  }
  
  .compose-tweet-word-count-text {
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    font-size: 12px;
  }
  
  .compose-tweet-word-count-text.near-limit {
    color: #ef4444;
  }
  
  .compose-tweet-word-limit {
    font-size: 14px;
    color: #6b7280;
  }
  
  /* Media preview */
  .compose-tweet-media-preview {
    margin-bottom: 16px;
    border: 1px solid #e5e7eb;
    border-radius: 16px;
    overflow: hidden;
  }
  
  .compose-tweet-body-dark .compose-tweet-media-preview {
    border-color: #384152;
  }
  
  .compose-tweet-media-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    grid-gap: 2px;
    background-color: #e5e7eb;
  }
  
  .compose-tweet-body-dark .compose-tweet-media-grid {
    background-color: #384152;
  }
  
  .compose-tweet-media-grid.single {
    grid-template-columns: 1fr;
  }
  
  .compose-tweet-media-item {
    position: relative;
    aspect-ratio: 16/9;
    background-color: #f3f4f6;
    overflow: hidden;
  }
  
  .compose-tweet-body-dark .compose-tweet-media-item {
    background-color: #475569;
  }
  
  .compose-tweet-media-img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }
  
  .compose-tweet-media-placeholder {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 100%;
    height: 100%;
    padding: 8px;
    color: #6b7280;
    font-size: 14px;
    text-align: center;
  }
  
  .compose-tweet-media-remove {
    position: absolute;
    top: 8px;
    right: 8px;
    width: 28px;
    height: 28px;
    border-radius: 50%;
    background-color: rgba(0, 0, 0, 0.5);
    color: white;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 18px;
    border: none;
    cursor: pointer;
  }
  
  /* Category & tags */
  .compose-tweet-category {
    margin-bottom: 16px;
    padding-top: 16px;
    border-top: 1px solid #e5e7eb;
  }
  
  .compose-tweet-body-dark .compose-tweet-category {
    border-top: 1px solid #384152;
  }
  
  .compose-tweet-category-header {
    margin-bottom: 8px;
  }
  
  .compose-tweet-category-label {
    font-weight: 600;
    font-size: 14px;
  }
  
  .compose-tweet-category-tags {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    margin-bottom: 12px;
  }
  
  .compose-tweet-category-tag {
    display: flex;
    align-items: center;
    padding: 6px 12px;
    background-color: #e5e7eb;
    border-radius: 16px;
    font-size: 14px;
  }
  
  .compose-tweet-category-tag-dark {
    background-color: #384152;
  }
  
  .compose-tweet-category-remove {
    background: none;
    border: none;
    margin-left: 6px;
    cursor: pointer;
    font-size: 16px;
    line-height: 1;
  }
  
  .compose-tweet-category-input-wrapper {
    position: relative;
    margin-bottom: 12px;
  }
  
  .compose-tweet-category-input {
    width: 100%;
    padding: 10px 12px;
    border: 1px solid #e5e7eb;
    border-radius: 4px;
    font-size: 14px;
  }
  
  .compose-tweet-category-input-dark {
    background-color: #1e293b;
    border-color: #384152;
    color: #f1f5f9;
  }
  
  .compose-tweet-category-dropdown {
    position: absolute;
    top: 100%;
    left: 0;
    right: 0;
    background-color: white;
    border: 1px solid #e5e7eb;
    border-radius: 4px;
    box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
    z-index: 10;
    max-height: 200px;
    overflow-y: auto;
  }
  
  .compose-tweet-category-dropdown-dark {
    background-color: #1e293b;
    border-color: #384152;
  }
  
  .compose-tweet-category-option {
    width: 100%;
    padding: 8px 12px;
    text-align: left;
    background: none;
    border: none;
    cursor: pointer;
  }
  
  .compose-tweet-category-option:hover {
    background-color: #f3f4f6;
  }
  
  .compose-tweet-category-option-dark {
    color: #f1f5f9;
  }
  
  .compose-tweet-category-option-dark:hover {
    background-color: #334155;
  }
  
  /* Suggested categories */
  .compose-tweet-suggestions {
    margin-bottom: 12px;
  }
  
  .compose-tweet-suggestions-header {
    display: flex;
    align-items: center;
    margin-bottom: 8px;
  }
  
  .compose-tweet-suggestions-label {
    margin-left: 6px;
    font-size: 14px;
    color: #6b7280;
  }
  
  .compose-tweet-suggestions-tags {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }
  
  .compose-tweet-suggestion-tag {
    padding: 4px 10px;
    background-color: #f3f4f6;
    border: 1px solid #e5e7eb;
    border-radius: 16px;
    font-size: 14px;
    cursor: pointer;
  }
  
  .compose-tweet-suggestion-tag-dark {
    background-color: #334155;
    border-color: #475569;
    color: #f1f5f9;
  }
  
  .compose-tweet-suggestion-tag:hover {
    background-color: #e5e7eb;
  }
  
  .compose-tweet-suggestion-tag-dark:hover {
    background-color: #475569;
  }
  
  /* Loading state */
  .compose-tweet-loading {
    display: flex;
    align-items: center;
    margin-top: 8px;
  }
  
  .compose-tweet-loading-spinner {
    width: 16px;
    height: 16px;
    border: 2px solid #e5e7eb;
    border-top-color: #3b82f6;
    border-radius: 50%;
    animation: spinner 0.6s linear infinite;
    margin-right: 8px;
  }
  
  @keyframes spinner {
    to { transform: rotate(360deg); }
  }
  
  .compose-tweet-loading-text {
    font-size: 14px;
    color: #6b7280;
  }
  
  /* Reply settings */
  .compose-tweet-reply-settings {
    margin-bottom: 16px;
  }
  
  .compose-tweet-reply-select {
    padding: 8px 12px;
    font-size: 14px;
    border: 1px solid #e5e7eb;
    border-radius: 4px;
    background-color: white;
    color: #374151;
  }
  
  .compose-tweet-reply-select-dark {
    background-color: #1e293b;
    border-color: #384152;
    color: #f1f5f9;
  }
  
  /* Action buttons */
  .compose-tweet-actions {
    display: flex;
    justify-content: space-between;
    align-items: center;
    border-top: 1px solid #e5e7eb;
    padding-top: 16px;
  }
  
  .compose-tweet-body-dark .compose-tweet-actions {
    border-top: 1px solid #384152;
  }
  
  .compose-tweet-tools {
    display: flex;
    align-items: center;
  }
  
  .compose-tweet-tool {
    background: none;
    border: none;
    color: #3b82f6;
    width: 40px;
    height: 40px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    margin-right: 8px;
  }
  
  .compose-tweet-tool:hover {
    background-color: rgba(59, 130, 246, 0.1);
  }
  
  .compose-tweet-file-input {
    display: none;
  }
  
  .compose-tweet-file-count {
    font-size: 14px;
    color: #6b7280;
  }
  
  .compose-tweet-submit {
    background-color: #3b82f6;
    color: white;
    font-weight: 700;
    padding: 8px 16px;
    border-radius: 9999px;
    border: none;
    cursor: pointer;
  }
  
  .compose-tweet-submit:hover {
    background-color: #2563eb;
  }
  
  .compose-tweet-submit:disabled {
    background-color: #93c5fd;
    cursor: not-allowed;
  }
  
  .compose-tweet-error {
    color: #ef4444;
    margin-bottom: 16px;
    font-size: 14px;
  }
  
  .active-tool {
    color: #2563eb;
    font-weight: bold;
  }

  /* Styles for the new components */
  .compose-tweet-community-selection,
  .compose-tweet-schedule,
  .compose-tweet-poll-options {
    margin-bottom: 16px;
    padding: 12px;
    border-radius: 8px;
    background: var(--bg-secondary, #f8fafc);
  }
  
  .compose-tweet-community-selection-dark,
  .compose-tweet-schedule-dark,
  .compose-tweet-poll-options-dark {
    background: var(--bg-secondary-dark, #1e293b);
  }
  
  .compose-tweet-community-select,
  .compose-tweet-schedule-date,
  .compose-tweet-schedule-time,
  .compose-tweet-poll-select {
    width: 100%;
    padding: 8px;
    border-radius: 4px;
    border: 1px solid #e2e8f0;
    background: white;
    margin-top: 8px;
  }
  
  .compose-tweet-community-select-dark,
  .compose-tweet-schedule-date-dark,
  .compose-tweet-schedule-time-dark,
  .compose-tweet-poll-select-dark {
    border-color: #4b5563;
    background: #374151;
    color: #e5e7eb;
  }
  
  .compose-tweet-schedule-inputs {
    display: flex;
    gap: 8px;
  }
  
  .compose-tweet-poll-question,
  .compose-tweet-poll-choice {
    margin-bottom: 8px;
    position: relative;
  }
  
  .compose-tweet-poll-question-input,
  .compose-tweet-poll-choice-input {
    width: 100%;
    padding: 8px;
    border-radius: 4px;
    border: 1px solid #e2e8f0;
    background: white;
  }
  
  .compose-tweet-poll-question-input-dark,
  .compose-tweet-poll-choice-input-dark {
    border-color: #4b5563;
    background: #374151;
    color: #e5e7eb;
  }
  
  .compose-tweet-poll-choice-remove {
    position: absolute;
    right: 8px;
    top: 50%;
    transform: translateY(-50%);
    background: none;
    border: none;
    font-size: 18px;
    cursor: pointer;
    color: #ef4444;
  }
  
  .compose-tweet-poll-add-option {
    padding: 8px;
    background: none;
    border: 1px dashed #e2e8f0;
    border-radius: 4px;
    width: 100%;
    cursor: pointer;
    color: #3b82f6;
    margin-bottom: 12px;
  }
  
  .compose-tweet-poll-add-option-dark {
    border-color: #4b5563;
    color: #60a5fa;
  }
  
  .compose-tweet-poll-settings {
    display: flex;
    flex-wrap: wrap;
    gap: 16px;
  }
  
  .compose-tweet-poll-duration,
  .compose-tweet-poll-who-can-vote {
    flex: 1;
    min-width: 150px;
  }
  
  .compose-tweet-admin-options {
    margin-top: 12px;
    padding: 8px 0;
  }
  
  .compose-tweet-admin-option {
    display: flex;
    align-items: center;
    gap: 8px;
    cursor: pointer;
    color: var(--text-primary, #1f2937);
  }
</style>