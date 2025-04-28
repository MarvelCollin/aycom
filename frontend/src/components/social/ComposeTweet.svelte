<script lang="ts">
  export let avatar = "👤";
  
  let newTweet = '';
  
  function postTweet() {
    if (newTweet.trim() === '') return;
    console.log('Posted:', newTweet);
    
    // Dispatch an event to notify parent component
    dispatch('tweet', { content: newTweet });
    
    // Clear the input
    newTweet = '';
  }
  
  import { createEventDispatcher } from 'svelte';
  const dispatch = createEventDispatcher();
</script>

<div class="border-b border-gray-800 p-4">
  <div class="flex">
    <div class="w-12 h-12 bg-gray-300 rounded-full flex items-center justify-center mr-4">
      <span>{avatar}</span>
    </div>
    <div class="flex-1">
      <textarea 
        bind:value={newTweet}
        placeholder="What is happening?!"
        class="w-full bg-transparent border-none outline-none text-white resize-none mb-4 text-xl"
        rows="3"
      ></textarea>
      
      <div class="flex justify-between items-center">
        <div class="flex text-blue-500">
          <button class="p-2 rounded-full hover:bg-blue-900 hover:bg-opacity-20">📷</button>
          <button class="p-2 rounded-full hover:bg-blue-900 hover:bg-opacity-20">📊</button>
          <button class="p-2 rounded-full hover:bg-blue-900 hover:bg-opacity-20">😀</button>
          <button class="p-2 rounded-full hover:bg-blue-900 hover:bg-opacity-20">📅</button>
          <button class="p-2 rounded-full hover:bg-blue-900 hover:bg-opacity-20">📍</button>
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