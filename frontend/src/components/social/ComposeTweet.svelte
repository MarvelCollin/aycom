<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { 
    ImageIcon, 
    BarChart2Icon, 
    SmileIcon, 
    MapPinIcon
  } from 'svelte-feather-icons';
  
  export let avatar = "👤";
  export let isDarkMode = false;
  
  let newTweet = '';
  let files: File[] = [];
  let replyPermission = 'everyone';
  let category = '';
  const maxWords = 280;
  
  function postTweet() {
    if (newTweet.trim() === '') return;
    dispatch('tweet', { content: newTweet, files, replyPermission, category });
    newTweet = '';
    files = [];
    category = '';
    replyPermission = 'everyone';
    dispatch('close');
  }

  function handleFileChange(e: Event) {
    const input = e.target as HTMLInputElement;
    if (input.files) {
      files = Array.from(input.files);
    }
  }

  const dispatch = createEventDispatcher();
  function closeModal() { dispatch('close'); }
  $: wordCount = newTweet.trim().split(/\s+/).filter(Boolean).length;
  $: wordPercent = Math.min(100, Math.round((wordCount / maxWords) * 100));
</script>

<div class="modal-container">
  <div class="modal-overlay" on:click={closeModal}></div>
  <div class="modal-content {isDarkMode ? 'modal-content-dark' : ''}">
    <div class="modal-header {isDarkMode ? 'modal-header-dark' : ''}">
      <button 
        class="text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200"
        on:click={closeModal}
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
        <div class="w-12 h-12 {isDarkMode ? 'bg-gray-700' : 'bg-gray-300'} rounded-full flex items-center justify-center mr-4 overflow-hidden">
          <span>{avatar}</span>
        </div>
        <div class="flex-1">
          <textarea 
            bind:value={newTweet}
            placeholder="What is happening?!"
            class="w-full bg-transparent border-none outline-none {isDarkMode ? 'text-white placeholder-gray-500' : 'text-gray-900 placeholder-gray-400'} resize-none mb-2 text-xl"
            rows="3"
            maxlength={maxWords * 6}
          ></textarea>
          <div class="flex items-center gap-2 mb-2">
            <div class="relative w-8 h-8">
              <svg viewBox="0 0 36 36" class="absolute top-0 left-0 w-8 h-8">
                <path
                  d="M18 2a16 16 0 1 1 0 32 16 16 0 0 1 0-32"
                  fill="none"
                  stroke="#e5e7eb"
                  stroke-width="4"
                />
                <path
                  d="M18 2a16 16 0 1 1 0 32 16 16 0 0 1 0-32"
                  fill="none"
                  stroke="#2563eb"
                  stroke-width="4"
                  stroke-dasharray="100, 100"
                  stroke-dashoffset={100 - wordPercent}
                  style="transition: stroke-dashoffset 0.2s;"
                />
              </svg>
              <span class="absolute inset-0 flex items-center justify-center text-xs font-bold {isDarkMode ? 'text-white' : 'text-gray-900'}">{wordCount}</span>
            </div>
            <span class="text-xs {isDarkMode ? 'text-gray-400' : 'text-gray-500'}">/ {maxWords} words</span>
          </div>
          <div class="flex gap-2 mb-2">
            <input type="file" multiple accept="image/*,video/*,.gif" on:change={handleFileChange} class="hidden" id="file-upload" />
            <label for="file-upload" class="cursor-pointer p-2 rounded-full {isDarkMode ? 'hover:bg-blue-900 hover:bg-opacity-20' : 'hover:bg-blue-100'} text-blue-500">
              <ImageIcon size="18" />
            </label>
            {#if files.length > 0}
              <span class="text-xs {isDarkMode ? 'text-gray-400' : 'text-gray-500'}">{files.length} file(s) attached</span>
            {/if}
          </div>
          <div class="flex gap-2 mb-2">
            <input type="text" placeholder="Add categories (comma separated)" bind:value={category} class="w-1/2 px-2 py-1 rounded border {isDarkMode ? 'bg-gray-800 border-gray-700 text-white' : 'bg-gray-100 border-gray-300 text-gray-900'} text-xs" />
            <select bind:value={replyPermission} class="px-2 py-1 rounded border {isDarkMode ? 'bg-gray-800 border-gray-700 text-white' : 'bg-gray-100 border-gray-300 text-gray-900'} text-xs">
              <option value="everyone">Everyone can reply</option>
              <option value="following">Accounts you follow</option>
              <option value="verified">Verified accounts</option>
            </select>
          </div>
          <div class="flex justify-between items-center mt-2">
            <div class="flex text-blue-500 gap-2">
              <button class="p-2 rounded-full {isDarkMode ? 'hover:bg-blue-900 hover:bg-opacity-20' : 'hover:bg-blue-100'}">
                <BarChart2Icon size="18" />
              </button>
              <button class="p-2 rounded-full {isDarkMode ? 'hover:bg-blue-900 hover:bg-opacity-20' : 'hover:bg-blue-100'}">
                <SmileIcon size="18" />
              </button>
              <button class="p-2 rounded-full {isDarkMode ? 'hover:bg-blue-900 hover:bg-opacity-20' : 'hover:bg-blue-100'}">
                <MapPinIcon size="18" />
              </button>
            </div>
            <button 
              on:click={postTweet}
              class="px-4 py-2 bg-blue-500 text-white rounded-full font-bold hover:bg-blue-600 disabled:opacity-50"
              disabled={newTweet.trim() === '' || wordCount > maxWords}
            >
              Post
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</div>