<script lang="ts">
  import { createEventDispatcher, onMount } from 'svelte';
  import { fade } from 'svelte/transition';
  import { useAuth } from '../../hooks/useAuth';
  import { useTheme } from '../../hooks/useTheme';
  import type { ICommunity } from '../../interfaces/ISocialMedia';
  
  // Props
  export let avatar = "👤";
  export let username = "";
  export let displayName = "";
  export let isAdmin = false;
  
  // Event dispatcher
  const dispatch = createEventDispatcher();
  
  // Theme
  const { theme } = useTheme();
  $: isDarkMode = $theme === 'dark';
  
  // Auth
  const { getAuthState } = useAuth();
  
  // State
  let content = "";
  let files: File[] = [];
  let previewUrls: string[] = [];
  let selectedCommunity: ICommunity | null = null;
  let isPoll = false;
  let pollOptions = ["", ""];
  let pollDuration = 1; // days
  let replyPrivacy = "everyone"; // everyone, following, verified
  let categories: string[] = [];
  let availableCategories = [
    "Technology", "Politics", "Entertainment", "Sports",
    "Science", "Health", "Business", "Education", "Art", "Travel"
  ];
  let isScheduled = false;
  let scheduledDate: string = "";
  let scheduledTime: string = "";
  let showCategoryInput = false;
  let newCategory = "";
  let isPosting = false;
  let error = "";
  let isPrivacyDropdownOpen = false;
  let isScheduleDropdownOpen = false;
  
  // Constants
  const MAX_CHARS = 280;
  const MAX_FILES = 4;
  const MAX_POLL_OPTIONS = 4;
  const MAX_CATEGORIES = 3;
  
  // Computed values
  $: charsLeft = MAX_CHARS - content.length;
  $: progressPercentage = (content.length / MAX_CHARS) * 100;
  $: isValid = content.trim().length > 0 && content.length <= MAX_CHARS;
  $: canAddPollOption = pollOptions.length < MAX_POLL_OPTIONS;
  $: canAddCategory = categories.length < MAX_CATEGORIES;
  $: progressColor = progressPercentage < 70 ? 'text-blue-500' : 
                     progressPercentage < 90 ? 'text-yellow-500' : 'text-red-500';
  $: scheduleValid = !isScheduled || (scheduledDate && scheduledTime);
  $: scheduledDateTime = isScheduled && scheduledDate && scheduledTime ? 
      new Date(`${scheduledDate}T${scheduledTime}`) : null;
  $: isPastDate = scheduledDateTime ? scheduledDateTime < new Date() : false;
  
  onMount(() => {
    // Set default scheduled date and time to tomorrow at noon
    const tomorrow = new Date();
    tomorrow.setDate(tomorrow.getDate() + 1);
    tomorrow.setHours(12, 0, 0, 0);
    
    scheduledDate = tomorrow.toISOString().split('T')[0];
    scheduledTime = '12:00';
  });
  
  // Functions
  function closeModal() {
    dispatch('close');
  }
  
  function handleFileInput(e: Event) {
    const input = e.target as HTMLInputElement;
    const newFiles = Array.from(input.files || []);
    
    if (files.length + newFiles.length > MAX_FILES) {
      error = `You can only attach up to ${MAX_FILES} files`;
      return;
    }
    
    error = "";
    
    newFiles.forEach(file => {
      if (files.length >= MAX_FILES) return;
      
      const fileType = file.type.split('/')[0];
      if (fileType !== 'image' && fileType !== 'video') {
        error = "Only images and videos are allowed";
        return;
      }
      
      files = [...files, file];
      
      // Create preview URL
      const url = URL.createObjectURL(file);
      previewUrls = [...previewUrls, url];
    });
    
    // Reset input field
    input.value = "";
  }
  
  function removeFile(index: number) {
    // Revoke object URL to prevent memory leak
    URL.revokeObjectURL(previewUrls[index]);
    
    files = files.filter((_, i) => i !== index);
    previewUrls = previewUrls.filter((_, i) => i !== index);
    error = "";
  }
  
  function togglePoll() {
    isPoll = !isPoll;
    if (isPoll) {
      // Reset poll options
      pollOptions = ["", ""];
      pollDuration = 1;
    }
  }
  
  function addPollOption() {
    if (canAddPollOption) {
      pollOptions = [...pollOptions, ""];
    }
  }
  
  function removePollOption(index: number) {
    if (pollOptions.length > 2) {
      pollOptions = pollOptions.filter((_, i) => i !== index);
    }
  }
  
  function toggleCategory(category: string) {
    if (categories.includes(category)) {
      categories = categories.filter(c => c !== category);
    } else if (canAddCategory) {
      categories = [...categories, category];
    }
  }
  
  function addCustomCategory() {
    if (newCategory.trim() && !availableCategories.includes(newCategory) && canAddCategory) {
      availableCategories = [...availableCategories, newCategory];
      categories = [...categories, newCategory];
      newCategory = "";
      showCategoryInput = false;
    }
  }
  
  function toggleSchedule() {
    isScheduled = !isScheduled;
    isScheduleDropdownOpen = false;
  }
  
  async function handleSubmit() {
    if (!isValid || isPosting || !scheduleValid || isPastDate) return;
    
    isPosting = true;
    
    try {
      // Create form data for files
      const formData = new FormData();
      formData.append("content", content);
      formData.append("replyPrivacy", replyPrivacy);
      
      if (categories.length > 0) {
        formData.append("categories", JSON.stringify(categories));
      }
      
      if (selectedCommunity) {
        formData.append("communityId", selectedCommunity.id);
      }
      
      if (isAdmin) {
        formData.append("isAdvertisement", "true");
      }
      
      if (isPoll) {
        const validOptions = pollOptions.filter(option => option.trim().length > 0);
        if (validOptions.length < 2) {
          error = "A poll needs at least 2 options";
          isPosting = false;
          return;
        }
        formData.append("isPoll", "true");
        formData.append("pollOptions", JSON.stringify(validOptions));
        formData.append("pollDuration", pollDuration.toString());
      }
      
      if (isScheduled && scheduledDateTime) {
        formData.append("isScheduled", "true");
        formData.append("scheduledAt", scheduledDateTime.toISOString());
      }
      
      files.forEach((file, index) => {
        formData.append(`file${index}`, file);
      });
      
      // Mock thread creation - replace with actual API call
      await new Promise(resolve => setTimeout(resolve, 1000));
      
      // Success!
      dispatch('created');
      closeModal();
    } catch (err) {
      error = "Failed to create thread. Please try again.";
      console.error(err);
    } finally {
      isPosting = false;
    }
  }
</script>

<div class="fixed inset-0 z-50 flex items-center justify-center overflow-x-hidden overflow-y-auto bg-black bg-opacity-50" transition:fade>
  <div class="relative w-full max-w-xl mx-4 md:mx-auto" on:click|stopPropagation>
    <div class="relative rounded-lg shadow-xl {isDarkMode ? 'bg-gray-900 text-white' : 'bg-white text-black'}">
      <!-- Header -->
      <div class="flex items-center justify-between px-4 py-3 border-b {isDarkMode ? 'border-gray-700' : 'border-gray-200'}">
        <button 
          class="p-1 rounded-full hover:bg-gray-200 hover:bg-opacity-20" 
          on:click={closeModal}
        >
          <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
        
        {#if isScheduled}
          <div class="flex items-center text-blue-500 px-3 py-1 rounded-full {isDarkMode ? 'bg-blue-900 bg-opacity-30' : 'bg-blue-50'}">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 mr-1.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            <span class="text-sm font-medium">Scheduled</span>
          </div>
        {:else}
          <div class="px-3"></div> <!-- Spacer -->
        {/if}
        
        <button 
          class="px-4 py-1.5 rounded-full font-bold text-sm {isValid && scheduleValid && !isPastDate ? 'bg-blue-500 hover:bg-blue-600 text-white' : 'bg-blue-300 cursor-not-allowed text-white'} {isDarkMode ? 'bg-opacity-70' : ''}"
          disabled={!isValid || !scheduleValid || isPastDate || isPosting}
          on:click={handleSubmit}
        >
          {isPosting ? 'Posting...' : isScheduled ? 'Schedule' : 'Post'}
        </button>
      </div>
      
      <!-- Main Content -->
      <div class="px-4 py-3">
        <div class="flex space-x-3">
          <!-- Avatar -->
          <div class="flex-shrink-0">
            <div class="w-10 h-10 rounded-full bg-gray-300 flex items-center justify-center overflow-hidden {isDarkMode ? 'bg-gray-700' : 'bg-gray-200'}">
              {#if typeof avatar === 'string' && avatar.startsWith('http')}
                <img src={avatar} alt={username} class="w-full h-full object-cover" />
              {:else}
                <div class="text-xl">{avatar}</div>
              {/if}
            </div>
          </div>
          
          <!-- Input Area -->
          <div class="flex-1 min-w-0">
            {#if selectedCommunity}
              <div class="mb-2 px-2 py-1 rounded-full inline-flex items-center {isDarkMode ? 'bg-gray-800' : 'bg-gray-100'}">
                <span class="text-sm text-blue-500 font-medium">Posting to {selectedCommunity.name}</span>
                <button class="ml-1 text-gray-500" on:click={() => selectedCommunity = null}>
                  <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor">
                    <path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd" />
                  </svg>
                </button>
              </div>
            {/if}
            
            <!-- Textarea -->
            <textarea
              class="w-full min-h-[120px] resize-none border-0 bg-transparent p-0 placeholder-gray-500 focus:ring-0 {isDarkMode ? 'text-white' : 'text-black'}"
              placeholder="What's happening?"
              bind:value={content}
              maxlength={MAX_CHARS}
            ></textarea>
            
            <!-- Previews -->
            {#if previewUrls.length > 0}
              <div class="mb-3 grid grid-cols-2 gap-2 rounded-lg overflow-hidden {previewUrls.length === 1 ? 'grid-cols-1' : ''}">
                {#each previewUrls as url, i}
                  <div class="relative aspect-square">
                    {#if files[i]?.type.startsWith('image/')}
                      <img src={url} alt="Preview" class="w-full h-full object-cover" />
                    {:else if files[i]?.type.startsWith('video/')}
                      <video src={url} class="w-full h-full object-cover" controls></video>
                    {/if}
                    <button 
                      class="absolute top-1 right-1 bg-black bg-opacity-70 rounded-full p-1"
                      on:click={() => removeFile(i)}
                    >
                      <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-white" viewBox="0 0 20 20" fill="currentColor">
                        <path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd" />
                      </svg>
                    </button>
                  </div>
                {/each}
              </div>
            {/if}
            
            <!-- Poll UI -->
            {#if isPoll}
              <div class="mt-3 border rounded-lg p-3 {isDarkMode ? 'border-gray-700' : 'border-gray-200'}">
                <h3 class="font-medium mb-2">Poll options</h3>
                {#each pollOptions as option, i}
                  <div class="flex items-center mb-2">
                    <input 
                      type="text"
                      placeholder={`Option ${i+1}`}
                      class="flex-1 p-2 rounded-lg {isDarkMode ? 'bg-gray-800 text-white border-gray-700' : 'bg-gray-100 text-black border-gray-200'} border"
                      bind:value={pollOptions[i]}
                    />
                    {#if i > 1}
                      <button 
                        class="ml-2 text-gray-500 hover:text-red-500"
                        on:click={() => removePollOption(i)}
                      >
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                          <path fill-rule="evenodd" d="M9 2a1 1 0 00-.894.553L7.382 4H4a1 1 0 000 2v10a2 2 0 002 2h8a2 2 0 002-2V6a1 1 0 100-2h-3.382l-.724-1.447A1 1 0 0011 2H9zM7 8a1 1 0 012 0v6a1 1 0 11-2 0V8zm5-1a1 1 0 00-1 1v6a1 1 0 102 0V8a1 1 0 00-1-1z" clip-rule="evenodd" />
                        </svg>
                      </button>
                    {/if}
                  </div>
                {/each}
                
                {#if canAddPollOption}
                  <button 
                    class="text-blue-500 font-medium text-sm flex items-center mt-2"
                    on:click={addPollOption}
                  >
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 mr-1" viewBox="0 0 20 20" fill="currentColor">
                      <path fill-rule="evenodd" d="M10 3a1 1 0 011 1v5h5a1 1 0 110 2h-5v5a1 1 0 11-2 0v-5H4a1 1 0 110-2h5V4a1 1 0 011-1z" clip-rule="evenodd" />
                    </svg>
                    Add option
                  </button>
                {/if}
                
                <div class="mt-3">
                  <label class="block text-sm font-medium mb-1">Poll length</label>
                  <select 
                    bind:value={pollDuration}
                    class="w-full p-2 rounded-lg {isDarkMode ? 'bg-gray-800 text-white border-gray-700' : 'bg-gray-100 text-black border-gray-200'} border"
                  >
                    <option value={1}>1 day</option>
                    <option value={3}>3 days</option>
                    <option value={7}>1 week</option>
                    <option value={30}>1 month</option>
                  </select>
                </div>
              </div>
            {/if}
            
            <!-- Categories -->
            {#if categories.length > 0}
              <div class="mt-3 flex flex-wrap gap-2">
                {#each categories as category}
                  <div class="bg-blue-100 text-blue-800 rounded-full px-3 py-1 text-sm flex items-center {isDarkMode ? 'bg-blue-900 bg-opacity-50 text-blue-300' : ''}">
                    <span>{category}</span>
                    <button class="ml-1.5" on:click={() => toggleCategory(category)}>
                      <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor">
                        <path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd" />
                      </svg>
                    </button>
                  </div>
                {/each}
              </div>
            {/if}
            
            <!-- Schedule Info -->
            {#if isScheduled && scheduledDateTime}
              <div class="mt-3 flex items-center bg-blue-100 text-blue-800 rounded-lg px-3 py-2 text-sm {isDarkMode ? 'bg-blue-900 bg-opacity-30 text-blue-300' : ''}">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                <span>
                  Will be posted on {scheduledDateTime.toLocaleDateString()} at {scheduledDateTime.toLocaleTimeString([], {hour: '2-digit', minute:'2-digit'})}
                </span>
                {#if isPastDate}
                  <div class="ml-auto text-red-500 font-medium">Cannot schedule in the past!</div>
                {/if}
              </div>
            {/if}
            
            <!-- Error Message -->
            {#if error}
              <div class="mt-3 text-red-500 text-sm">{error}</div>
            {/if}
          </div>
        </div>
      </div>
      
      <!-- Footer -->
      <div class="px-4 py-3 flex items-center justify-between border-t {isDarkMode ? 'border-gray-700' : 'border-gray-200'}">
        <!-- Action Icons -->
        <div class="flex items-center space-x-5">
          <!-- Image Button -->
          <div class="relative">
            <label class="cursor-pointer text-blue-500 hover:text-blue-600" title="Add photos or video">
              <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
              </svg>
              <input type="file" class="hidden" accept="image/*,video/*" multiple on:change={handleFileInput} />
            </label>
            {#if files.length > 0}
              <span class="absolute -top-2 -right-2 bg-blue-500 text-white text-xs rounded-full w-4 h-4 flex items-center justify-center">
                {files.length}
              </span>
            {/if}
          </div>
          
          <!-- Poll Button -->
          <button 
            class={`text-blue-500 hover:text-blue-600 ${isPoll ? 'text-blue-600' : ''}`} 
            title="Create poll"
            on:click={togglePoll}
          >
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
            </svg>
          </button>
          
          <!-- Categories Button -->
          <div class="relative">
            <button 
              class="text-blue-500 hover:text-blue-600" 
              title="Add categories"
              on:click={() => showCategoryInput = !showCategoryInput}
            >
              <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A1.994 1.994 0 013 12V7a4 4 0 014-4z" />
              </svg>
            </button>
            
            {#if categories.length > 0}
              <span class="absolute -top-2 -right-2 bg-blue-500 text-white text-xs rounded-full w-4 h-4 flex items-center justify-center">
                {categories.length}
              </span>
            {/if}
            
            {#if showCategoryInput}
              <div class="absolute bottom-full left-0 mb-2 p-3 rounded-lg shadow-lg {isDarkMode ? 'bg-gray-800' : 'bg-white'} border {isDarkMode ? 'border-gray-700' : 'border-gray-200'} z-10 w-64">
                <div class="mb-2">
                  <h4 class="font-medium text-sm mb-1">Select categories ({MAX_CATEGORIES} max)</h4>
                  <div class="flex flex-wrap gap-2 mb-3">
                    {#each availableCategories as category}
                      <button 
                        class="px-2 py-1 text-xs rounded-full {categories.includes(category) ? 'bg-blue-500 text-white' : isDarkMode ? 'bg-gray-700 text-gray-300' : 'bg-gray-200 text-gray-700'}"
                        disabled={!categories.includes(category) && !canAddCategory}
                        on:click={() => toggleCategory(category)}
                      >
                        {category}
                      </button>
                    {/each}
                  </div>
                  
                  <div class="flex">
                    <input 
                      type="text" 
                      placeholder="New category" 
                      class="flex-1 p-1 text-sm border rounded {isDarkMode ? 'bg-gray-700 border-gray-600 text-white' : 'bg-white border-gray-300 text-black'}"
                      bind:value={newCategory}
                    />
                    <button 
                      class="ml-2 px-2 py-1 rounded bg-blue-500 text-white text-sm"
                      on:click={addCustomCategory}
                      disabled={!newCategory.trim() || !canAddCategory}
                    >
                      Add
                    </button>
                  </div>
                </div>
              </div>
            {/if}
          </div>
          
          <!-- Schedule Button -->
          <div class="relative">
            <button 
              class={`text-blue-500 hover:text-blue-600 ${isScheduled ? 'text-blue-600' : ''}`}
              title="Schedule post"
              on:click={() => isScheduleDropdownOpen = !isScheduleDropdownOpen}
            >
              <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            </button>
            
            {#if isScheduled}
              <span class="absolute -top-2 -right-2 bg-blue-500 text-white text-xs rounded-full w-4 h-4 flex items-center justify-center">
                ✓
              </span>
            {/if}
            
            {#if isScheduleDropdownOpen}
              <div class="absolute bottom-full left-0 mb-2 p-3 rounded-lg shadow-lg {isDarkMode ? 'bg-gray-800' : 'bg-white'} border {isDarkMode ? 'border-gray-700' : 'border-gray-200'} z-10 w-64">
                <h4 class="font-medium text-sm mb-2">Schedule for later</h4>
                <div class="toggle-switch mb-2">
                  <label class="flex items-center cursor-pointer">
                    <div class="relative">
                      <input type="checkbox" class="sr-only" bind:checked={isScheduled}>
                      <div class="w-10 h-5 bg-gray-400 rounded-full shadow-inner"></div>
                      <div class="dot absolute w-4 h-4 bg-white rounded-full shadow -left-1 -top-1 transition {isScheduled ? 'transform translate-x-7 bg-blue-500' : ''}"></div>
                    </div>
                    <div class="ml-3 text-sm font-medium">
                      Schedule this thread
                    </div>
                  </label>
                </div>
                
                {#if isScheduled}
                  <div class="space-y-2">
                    <div>
                      <label class="block text-xs font-medium mb-1">Date</label>
                      <input 
                        type="date" 
                        bind:value={scheduledDate}
                        min={new Date().toISOString().split('T')[0]}
                        class="w-full p-1.5 border rounded {isDarkMode ? 'bg-gray-700 border-gray-600 text-white' : 'bg-white border-gray-300 text-black'}"
                      />
                    </div>
                    <div>
                      <label class="block text-xs font-medium mb-1">Time</label>
                      <input 
                        type="time" 
                        bind:value={scheduledTime}
                        class="w-full p-1.5 border rounded {isDarkMode ? 'bg-gray-700 border-gray-600 text-white' : 'bg-white border-gray-300 text-black'}"
                      />
                    </div>
                  </div>
                {/if}
                
                <div class="mt-3 flex justify-end">
                  <button 
                    class="px-4 py-1.5 rounded bg-blue-500 text-white text-sm"
                    on:click={toggleSchedule}
                  >
                    {isScheduled ? 'Confirm' : 'Set schedule'}
                  </button>
                </div>
              </div>
            {/if}
          </div>
          
          <!-- Privacy Settings -->
          <div class="relative">
            <button 
              class="text-blue-500 hover:text-blue-600" 
              title="Who can reply"
              on:click={() => isPrivacyDropdownOpen = !isPrivacyDropdownOpen}
            >
              <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 11V7a4 4 0 118 0m-4 8v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2z" />
              </svg>
            </button>
            
            {#if isPrivacyDropdownOpen}
              <div class="absolute bottom-full left-0 mb-2 rounded-lg shadow-lg {isDarkMode ? 'bg-gray-800' : 'bg-white'} border {isDarkMode ? 'border-gray-700' : 'border-gray-200'} z-10 w-48">
                <h4 class="font-medium text-sm pt-3 px-3">Who can reply</h4>
                <div class="py-1">
                  <button 
                    class="w-full text-left px-3 py-2 flex items-center {replyPrivacy === 'everyone' ? 'bg-blue-500 bg-opacity-10 text-blue-500' : ''}"
                    on:click={() => {replyPrivacy = 'everyone'; isPrivacyDropdownOpen = false;}}
                  >
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9" />
                    </svg>
                    Everyone
                  </button>
                  
                  <button 
                    class="w-full text-left px-3 py-2 flex items-center {replyPrivacy === 'following' ? 'bg-blue-500 bg-opacity-10 text-blue-500' : ''}"
                    on:click={() => {replyPrivacy = 'following'; isPrivacyDropdownOpen = false;}}
                  >
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
                    </svg>
                    Accounts you follow
                  </button>
                  
                  <button 
                    class="w-full text-left px-3 py-2 flex items-center {replyPrivacy === 'verified' ? 'bg-blue-500 bg-opacity-10 text-blue-500' : ''}"
                    on:click={() => {replyPrivacy = 'verified'; isPrivacyDropdownOpen = false;}}
                  >
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
                    </svg>
                    Verified accounts
                  </button>
                </div>
              </div>
            {/if}
          </div>
        </div>
        
        <!-- Character Counter -->
        <div class="flex items-center">
          <!-- Divider line -->
          <div class="h-6 w-px bg-gray-300 mx-3 {isDarkMode ? 'bg-gray-700' : ''}"></div>
          
          <!-- Progress Circle -->
          <div class="relative h-6 w-6">
            <svg class="w-6 h-6" viewBox="0 0 24 24">
              <!-- Background Circle -->
              <circle cx="12" cy="12" r="10" fill="none" stroke={isDarkMode ? '#374151' : '#f3f4f6'} stroke-width="2" />
              
              <!-- Progress Circle -->
              <circle 
                cx="12" 
                cy="12" 
                r="10" 
                fill="none" 
                stroke="currentColor" 
                stroke-width="2" 
                class={progressColor}
                stroke-dasharray={Math.PI * 20} 
                stroke-dashoffset={Math.PI * 20 * (1 - progressPercentage / 100)} 
                transform="rotate(-90 12 12)" 
              />
              
              <!-- Show character count when near limit -->
              {#if charsLeft <= 20}
                <text 
                  x="12" 
                  y="13" 
                  text-anchor="middle" 
                  class={charsLeft <= 0 ? 'text-red-500' : charsLeft <= 10 ? 'text-yellow-500' : 'text-blue-500'}
                  style="font-size: 9px; font-weight: bold;"
                >
                  {charsLeft}
                </text>
              {/if}
            </svg>
          </div>
        </div>
      </div>
    </div>
  </div>
</div>

<style>
  /* Ensure model stays visible even with overflow */
  :global(body.modal-open) {
    overflow: hidden;
  }
  
  /* Style for toggle button transition */
  .dot {
    transition: all 0.3s ease-in-out;
  }
  
  /* Progress circle animation */
  circle {
    transition: stroke-dashoffset 0.3s ease;
  }
  
  /* Styles for file upload elements */
  input[type="file"] {
    width: 0.1px;
    height: 0.1px;
    opacity: 0;
    overflow: hidden;
    position: absolute;
    z-index: -1;
  }
  
  /* Textarea autogrow */
  textarea {
    overflow: hidden;
    transition: height 0.1s ease;
  }
</style>