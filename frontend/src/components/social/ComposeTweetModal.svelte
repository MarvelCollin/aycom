<script lang="ts">
  import { createEventDispatcher, tick } from 'svelte';
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
  import { createThread, uploadThreadMedia, suggestThreadCategory } from '../../api/thread';
  import { toastStore } from '../../stores/toastStore';
  import { useTheme } from '../../hooks/useTheme';
  import { debounce } from '../../utils/helpers';
  
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
  
  // Category suggestion
  let suggestedCategory = '';
  let suggestedCategoryConfidence = 0;
  let isSuggestingCategory = false;
  let categoryTouched = false; // User has manually selected a category
  let selectedCategory = '';
  let allCategories: Record<string, number> = {};
  let categorySuggestionDebounceTimeout: ReturnType<typeof setTimeout> | null = null;
  
  const categoryOptions = [
    { value: 'technology', label: 'Technology', icon: 'laptop' },
    { value: 'entertainment', label: 'Entertainment', icon: 'film' },
    { value: 'health', label: 'Health', icon: 'heart' },
    { value: 'sports', label: 'Sports', icon: 'activity' },
    { value: 'business', label: 'Business', icon: 'briefcase' },
    { value: 'politics', label: 'Politics', icon: 'flag' },
    { value: 'education', label: 'Education', icon: 'book' },
    { value: 'gaming', label: 'Gaming', icon: 'controller' },
    { value: 'food', label: 'Food', icon: 'coffee' },
    { value: 'travel', label: 'Travel', icon: 'map' },
    { value: 'general', label: 'General', icon: 'hash' }
  ];
  
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
    suggestedCategory = '';
    suggestedCategoryConfidence = 0;
    categoryTouched = false;
    selectedCategory = '';
    allCategories = {};
  }
  
  // Debounced function to get category suggestions
  const getSuggestedCategory = debounce(async (content: string) => {
    // Don't suggest if user already manually selected
    if (categoryTouched) return;
    
    // Don't suggest if content is too short
    if (!content || content.trim().length < 10) {
      suggestedCategory = '';
      suggestedCategoryConfidence = 0;
      return;
    }
    
    try {
      isSuggestingCategory = true;
      
      const result = await suggestThreadCategory(content);
      suggestedCategory = result.category;
      suggestedCategoryConfidence = result.confidence;
      
      // Auto-select the suggested category if confidence is above 0.7
      if (suggestedCategory && suggestedCategoryConfidence > 0.7 && !categoryTouched) {
        selectedCategory = suggestedCategory;
      }
    } catch (error) {
      console.error("Error getting category suggestion:", error);
    } finally {
      isSuggestingCategory = false;
    }
  }, 500);
  
  // Watch newTweet for changes to trigger category suggestion
  $: if (newTweet) {
    getSuggestedCategory(newTweet);
  }
  
  function handleCategorySelect(category: string) {
    selectedCategory = category;
    categoryTouched = true;
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
        who_can_reply: 'everyone',
        category: selectedCategory || suggestedCategory || 'general'
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
            
            <!-- Category suggestion -->
            {#if newTweet && newTweet.length >= 10}
              <div class="category-suggestion-container">
                <div class="category-suggestion-header">
                  <span class="category-suggestion-title">
                    <span class="category-icon">#</span>
                    {isSuggestingCategory ? 'Analyzing content...' : 'Category'}
                  </span>
                  
                  {#if suggestedCategory && !categoryTouched}
                    <span class="category-suggestion-info">
                      AI suggested: <strong>{suggestedCategory}</strong>
                      {#if suggestedCategoryConfidence > 0}
                        ({Math.round(suggestedCategoryConfidence * 100)}% confidence)
                      {/if}
                    </span>
                  {/if}
                </div>
                
                <div class="category-options">
                  {#each categoryOptions as option}
                    <button 
                      class="category-option {selectedCategory === option.value ? 'selected' : ''} 
                             {!selectedCategory && suggestedCategory === option.value ? 'suggested' : ''}"
                      on:click={() => handleCategorySelect(option.value)}
                    >
                      {option.label}
                    </button>
                  {/each}
                </div>
              </div>
            {/if}
            
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
            <div class="compose-tweet-char-counter {isOverLimit ? 'over-limit' : ''} {isNearLimit ? 'near-limit' : ''}">
              {isOverLimit ? -charsRemaining : charsRemaining}
            </div>
            
            <button 
              class="compose-tweet-submit"
              on:click={handleSubmit}
              disabled={isPosting || isOverLimit || (newTweet.trim() === '' && files.length === 0)}
            >
              {isPosting ? 'Publishing...' : 'Post'}
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
  
  .compose-tweet-char-counter {
    margin-right: var(--space-3);
    color: var(--text-secondary);
    font-size: var(--font-size-sm);
  }
  
  .compose-tweet-char-counter.near-limit {
    color: var(--color-warning);
  }
  
  .compose-tweet-char-counter.over-limit {
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
  
  .category-suggestion-container {
    margin: 10px 0;
    padding: 10px;
    border-radius: 8px;
    background-color: rgba(0, 0, 0, 0.02);
    border: 1px solid rgba(0, 0, 0, 0.05);
  }
  
  .category-suggestion-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 10px;
  }
  
  .category-suggestion-title {
    display: flex;
    align-items: center;
    gap: 5px;
    font-weight: 600;
    font-size: 14px;
  }
  
  .category-icon {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 16px;
    height: 16px;
    font-weight: bold;
  }
  
  .category-suggestion-info {
    font-size: 12px;
    color: #555;
  }
  
  .category-options {
    display: flex;
    flex-wrap: wrap;
    gap: 5px;
  }
  
  .category-option {
    padding: 6px 12px;
    border-radius: 20px;
    font-size: 12px;
    background-color: #f1f1f1;
    border: 1px solid transparent;
    cursor: pointer;
    transition: all 0.2s ease;
  }
  
  .category-option:hover {
    background-color: #e0e0e0;
  }
  
  .category-option.selected {
    background-color: #1d9bf0;
    color: white;
  }
  
  .category-option.suggested {
    border: 1px dashed #1d9bf0;
    animation: pulse 2s infinite;
  }
  
  @keyframes pulse {
    0% {
      border-color: rgba(29, 155, 240, 0.5);
    }
    50% {
      border-color: rgba(29, 155, 240, 1);
    }
    100% {
      border-color: rgba(29, 155, 240, 0.5);
    }
  }

  /* Dark mode overrides */
  :global(.dark) .category-suggestion-container {
    background-color: rgba(255, 255, 255, 0.05);
    border-color: rgba(255, 255, 255, 0.1);
  }
  
  :global(.dark) .category-suggestion-info {
    color: #aaa;
  }
  
  :global(.dark) .category-option {
    background-color: #2f3336;
    color: #e0e0e0;
  }
  
  :global(.dark) .category-option:hover {
    background-color: #3f4246;
  }
</style> 