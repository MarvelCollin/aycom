<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { 
    ImageIcon, 
    BarChart2Icon, 
    SmileIcon, 
    CalendarIcon, 
    MapPinIcon
  } from 'svelte-feather-icons';
  
  export let avatar = "👤";
  export let isDarkMode = false;
  
  let newTweet = '';
  
  function postTweet() {
    if (newTweet.trim() === '') return;
    console.log('Posted:', newTweet);
    dispatch('tweet', { content: newTweet });
    newTweet = '';
    dispatch('close');
  }
  
  const dispatch = createEventDispatcher();
  
  function closeModal() {
    dispatch('close');
  }
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
      <div class="flex">
        <div class="w-12 h-12 {isDarkMode ? 'bg-gray-700' : 'bg-gray-300'} rounded-full flex items-center justify-center mr-4 overflow-hidden">
          <span>{avatar}</span>
        </div>
        <div class="flex-1">
          <textarea 
            bind:value={newTweet}
            placeholder="What is happening?!"
            class="w-full bg-transparent border-none outline-none {isDarkMode ? 'text-white placeholder-gray-500' : 'text-gray-900 placeholder-gray-400'} resize-none mb-4 text-xl"
            rows="3"
          ></textarea>
          <div class="flex justify-between items-center">
            <div class="flex text-blue-500">
              <button class="p-2 rounded-full {isDarkMode ? 'hover:bg-blue-900 hover:bg-opacity-20' : 'hover:bg-blue-100'}">
                <ImageIcon size="18" />
              </button>
              <button class="p-2 rounded-full {isDarkMode ? 'hover:bg-blue-900 hover:bg-opacity-20' : 'hover:bg-blue-100'}">
                <BarChart2Icon size="18" />
              </button>
              <button class="p-2 rounded-full {isDarkMode ? 'hover:bg-blue-900 hover:bg-opacity-20' : 'hover:bg-blue-100'}">
                <SmileIcon size="18" />
              </button>
              <button class="p-2 rounded-full {isDarkMode ? 'hover:bg-blue-900 hover:bg-opacity-20' : 'hover:bg-blue-100'}">
                <CalendarIcon size="18" />
              </button>
              <button class="p-2 rounded-full {isDarkMode ? 'hover:bg-blue-900 hover:bg-opacity-20' : 'hover:bg-blue-100'}">
                <MapPinIcon size="18" />
              </button>
            </div>
            <button 
              on:click={postTweet}
              class="px-4 py-2 bg-blue-500 text-white rounded-full font-bold hover:bg-blue-600 disabled:opacity-50"
              disabled={newTweet.trim() === ''}
            >
              Post
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</div>