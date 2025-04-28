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
    
    // Dispatch an event to notify parent component
    dispatch('tweet', { content: newTweet });
    
    // Clear the input
    newTweet = '';
  }
  
  const dispatch = createEventDispatcher();
</script>

<div class="border-b {isDarkMode ? 'border-gray-800' : 'border-gray-200'} p-4">
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