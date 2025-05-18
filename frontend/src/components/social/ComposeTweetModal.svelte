<script lang="ts">
  import { createEventDispatcher } from 'svelte';
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
  import { createThread, uploadThreadMedia } from '../../api/thread';
  import { toastStore } from '../../stores/toastStore';
  import { useTheme } from '../../hooks/useTheme';
  
  export let isOpen = false;
  export let avatar = "https://secure.gravatar.com/avatar/0?d=mp";
  
  const { theme } = useTheme();
  $: isDarkMode = $theme === 'dark';
  
  let newTweet = '';
  let files: File[] = [];
  let fileInputRef: HTMLInputElement;
  let isPosting = false;
  let errorMessage = '';
  let previewImages: string[] = [];
  const maxChars = 280;
  
  $: charsRemaining = maxChars - newTweet.length;
  $: isOverLimit = charsRemaining < 0;
  $: isNearLimit = charsRemaining <= 20 && charsRemaining > 0;
  
  const dispatch = createEventDispatcher();
  
  function handleClose() {
    resetForm();
    dispatch('close');
  }
  
  function resetForm() {
    newTweet = '';
    files = [];
    previewImages = [];
    isPosting = false;
    errorMessage = '';
  }
  
  function handleFileSelect(e: Event) {
    const input = e.target as HTMLInputElement;
    if (input.files) {
      // Convert FileList to array
      const fileArray = Array.from(input.files);
      
      // Limit to 4 images
      const newFiles = fileArray.slice(0, 4 - files.length);
      
      if (files.length + newFiles.length > 4) {
        toastStore.showToast('Maximum 4 images allowed', 'warning');
      }
      
      files = [...files, ...newFiles];
      
      // Create preview URLs
      newFiles.forEach(file => {
        const reader = new FileReader();
        reader.onload = (e) => {
          const result = e.target?.result as string;
          previewImages = [...previewImages, result];
        };
        reader.readAsDataURL(file);
      });
    }
  }
  
  function removeImage(index: number) {
    files = files.filter((_, i) => i !== index);
    previewImages = previewImages.filter((_, i) => i !== index);
  }
  
  async function handleSubmit() {
    if (newTweet.trim() === '' && files.length === 0) {
      errorMessage = 'Your post cannot be empty';
      return;
    }
    
    if (isOverLimit) {
      errorMessage = `Your post exceeds the maximum character limit by ${-charsRemaining} characters`;
      return;
    }
    
    isPosting = true;
    errorMessage = '';
    
    try {
      const threadData = {
        content: newTweet,
        hashtags: [],
        who_can_reply: 'everyone'
      };
      
      const response = await createThread(threadData);
      
      if (files.length > 0) {
        await uploadThreadMedia(response.id, files);
      }
      
      toastStore.showToast('Your post was published successfully', 'success');
      resetForm();
      dispatch('posted', response);
      dispatch('close');
    } catch (error) {
      console.error('Error posting thread:', error);
      toastStore.showToast('Failed to publish your post. Please try again.', 'error');
      errorMessage = 'Failed to publish your post. Please try again.';
    } finally {
      isPosting = false;
    }
  }
</script>

{#if isOpen}
  <div class="modal-overlay" on:click={handleClose}>
    <div 
      class="modal-container {isDarkMode ? 'modal-container-dark' : ''}"
      on:click|stopPropagation={() => {}}
    >
      <div class="modal-header {isDarkMode ? 'modal-header-dark' : ''}">
        <button 
          class="modal-close-button {isDarkMode ? 'modal-close-button-dark' : ''}"
          on:click={handleClose}
          aria-label="Close"
        >
          <XIcon size="20" />
        </button>
        <span class="modal-title">Create a post</span>
      </div>
      
      <div class="compose-tweet-container">
        <div class="compose-tweet-header">
          <img src={avatar} alt="Your avatar" class="compose-tweet-avatar" />
          
          <div class="compose-tweet-input-area">
            <textarea 
              class="compose-tweet-textarea"
              placeholder="What's happening?"
              bind:value={newTweet}
              autofocus
            ></textarea>
            
            <!-- Media preview -->
            {#if previewImages.length > 0}
              <div class="compose-tweet-media-preview">
                <div class="compose-tweet-media-grid {previewImages.length === 1 ? 'single' : ''}">
                  {#each previewImages as preview, i}
                    <div class="compose-tweet-media-item">
                      <img src={preview} alt="Preview" class="compose-tweet-media-img" />
                      <button 
                        class="compose-tweet-media-remove"
                        on:click={() => removeImage(i)}
                        aria-label="Remove image"
                      >
                        <XIcon size="16" />
                      </button>
                    </div>
                  {/each}
                </div>
              </div>
            {/if}
            
            {#if errorMessage}
              <div class="compose-tweet-error">
                {errorMessage}
              </div>
            {/if}
          </div>
        </div>
        
        <div class="compose-tweet-actions">
          <div class="compose-tweet-tools">
            <label class="compose-tweet-tool" aria-label="Add images">
              <input 
                type="file" 
                class="compose-tweet-file-input" 
                accept="image/*" 
                multiple 
                bind:this={fileInputRef}
                on:change={handleFileSelect}
                disabled={files.length >= 4 || isPosting}
              />
              <ImageIcon size="20" />
            </label>
            
            <button class="compose-tweet-tool" aria-label="Add poll" disabled={isPosting}>
              <BarChart2Icon size="20" />
            </button>
            
            <button class="compose-tweet-tool" aria-label="Add emoji" disabled={isPosting}>
              <SmileIcon size="20" />
            </button>
            
            <button class="compose-tweet-tool" aria-label="Add location" disabled={isPosting}>
              <MapPinIcon size="20" />
            </button>
          </div>
          
          <div class="compose-tweet-submit-area">
            {#if newTweet.length > 0}
              <div class="compose-tweet-char-count {isNearLimit ? 'near-limit' : ''} {isOverLimit ? 'over-limit' : ''}">
                {charsRemaining}
              </div>
            {/if}
            
            <button 
              class="compose-tweet-submit"
              on:click={handleSubmit}
              disabled={isOverLimit || isPosting || (newTweet.trim() === '' && files.length === 0)}
            >
              {#if isPosting}
                <div class="loading-spinner-small"></div>
              {:else}
                Post
              {/if}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
{/if}

<style>
  .compose-tweet-container {
    padding: var(--space-3) var(--space-4);
  }
  
  .compose-tweet-header {
    display: flex;
    align-items: flex-start;
  }
  
  .compose-tweet-avatar {
    width: 48px;
    height: 48px;
    border-radius: 50%;
    margin-right: var(--space-3);
    object-fit: cover;
  }
  
  .compose-tweet-input-area {
    flex: 1;
    min-width: 0;
  }
  
  .compose-tweet-textarea {
    width: 100%;
    min-height: 120px;
    padding: var(--space-3) 0;
    border: none;
    background-color: transparent;
    color: var(--text-primary);
    font-size: var(--font-size-lg);
    resize: none;
    overflow-y: auto;
  }
  
  .compose-tweet-textarea:focus {
    outline: none;
  }
  
  .compose-tweet-textarea::placeholder {
    color: var(--text-tertiary);
  }
  
  .compose-tweet-media-preview {
    margin-top: var(--space-3);
    border-radius: var(--radius-md);
    overflow: hidden;
  }
  
  .compose-tweet-media-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 8px;
    border-radius: var(--radius-md);
  }
  
  .compose-tweet-media-grid.single {
    grid-template-columns: 1fr;
  }
  
  .compose-tweet-media-item {
    position: relative;
    aspect-ratio: 16/9;
    background-color: var(--bg-tertiary);
    border-radius: var(--radius-md);
    overflow: hidden;
  }
  
  .compose-tweet-media-img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }
  
  .compose-tweet-media-remove {
    position: absolute;
    top: 8px;
    right: 8px;
    width: 24px;
    height: 24px;
    border-radius: 50%;
    background-color: rgba(0, 0, 0, 0.5);
    color: white;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    border: none;
  }
  
  .compose-tweet-actions {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-top: var(--space-3);
    padding-top: var(--space-3);
    border-top: 1px solid var(--border-color);
  }
  
  .compose-tweet-tools {
    display: flex;
    gap: var(--space-3);
  }
  
  .compose-tweet-tool {
    color: var(--color-primary);
    background: transparent;
    border: none;
    cursor: pointer;
    padding: var(--space-1);
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: background-color var(--transition-fast);
  }
  
  .compose-tweet-tool:hover {
    background-color: var(--hover-primary);
  }
  
  .compose-tweet-tool:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
  
  .compose-tweet-file-input {
    display: none;
  }
  
  .compose-tweet-submit-area {
    display: flex;
    align-items: center;
  }
  
  .compose-tweet-char-count {
    margin-right: var(--space-3);
    color: var(--text-secondary);
    font-size: var(--font-size-sm);
  }
  
  .compose-tweet-char-count.near-limit {
    color: var(--color-warning);
  }
  
  .compose-tweet-char-count.over-limit {
    color: var(--color-danger);
  }
  
  .compose-tweet-submit {
    background-color: var(--color-primary);
    color: white;
    border: none;
    border-radius: var(--radius-full);
    padding: var(--space-2) var(--space-4);
    font-weight: var(--font-weight-bold);
    font-size: var(--font-size-md);
    cursor: pointer;
    transition: background-color var(--transition-fast), transform var(--transition-fast);
    display: flex;
    align-items: center;
    justify-content: center;
    min-width: 80px;
  }
  
  .compose-tweet-submit:hover {
    background-color: var(--color-primary-hover);
    transform: translateY(-1px);
  }
  
  .compose-tweet-submit:disabled {
    opacity: 0.5;
    cursor: not-allowed;
    transform: none;
  }
  
  .compose-tweet-error {
    margin-top: var(--space-2);
    color: var(--color-danger);
    font-size: var(--font-size-sm);
  }
  
  .loading-spinner-small {
    width: 16px;
    height: 16px;
    border: 2px solid rgba(255, 255, 255, 0.3);
    border-top-color: white;
    border-radius: 50%;
    animation: spin 1s linear infinite;
  }
  
  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }
</style> 