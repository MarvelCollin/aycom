<script lang="ts">
  import { onMount } from 'svelte';

  export let isDarkMode = false;
  let searchQuery = '';
  let isFocused = false;
  let recentSearches: string[] = [];
  
  // Load recent searches from local storage on mount
  onMount(() => {
    const savedSearches = localStorage.getItem('recentSearches');
    if (savedSearches) {
      recentSearches = JSON.parse(savedSearches).slice(0, 3);
    }
  });
  
  // Save a search to recent searches
  function saveSearch(query: string) {
    if (!query.trim()) return;
    
    // Remove duplicates and add new search at the beginning
    const newSearches = [
      query, 
      ...recentSearches.filter(s => s !== query)
    ].slice(0, 3);
    
    recentSearches = newSearches;
    localStorage.setItem('recentSearches', JSON.stringify(newSearches));
  }
  
  // Clear all recent searches
  function clearAllSearches() {
    recentSearches = [];
    localStorage.removeItem('recentSearches');
  }
  
  // Handle search submission
  function handleSearch() {
    if (searchQuery.trim()) {
      saveSearch(searchQuery);
      window.location.href = `/explore?q=${encodeURIComponent(searchQuery)}`;
    }
  }
  
  // Handle key press (Enter)
  function handleKeyPress(event: KeyboardEvent) {
    if (event.key === 'Enter') {
      handleSearch();
    }
  }
  
  // Handle clicking on a recent search
  function selectRecentSearch(search: string) {
    searchQuery = search;
    handleSearch();
  }
  
  // Toggle focus state
  function toggleFocus(value: boolean) {
    setTimeout(() => {
      isFocused = value;
    }, 100); // Small delay to allow for click events
  }
</script>

<div class="mb-6 relative">
  <div class="relative">
    <button 
      class="absolute left-3 top-3 {isDarkMode ? 'text-gray-400' : 'text-gray-500'}"
      on:click={handleSearch}
    >
      🔍
    </button>
    <input 
      type="text" 
      bind:value={searchQuery}
      placeholder="Search" 
      class="w-full {isDarkMode ? 'bg-gray-800 text-white' : 'bg-gray-100 text-gray-900'} rounded-full py-2 pl-10 pr-4 focus:outline-none focus:ring-1 focus:ring-blue-500"
      on:focus={() => toggleFocus(true)}
      on:blur={() => toggleFocus(false)}
      on:keypress={handleKeyPress}
    />
  </div>
  
  <!-- Recent searches dropdown -->
  {#if isFocused && recentSearches.length > 0}
    <div class="absolute left-0 right-0 mt-2 {isDarkMode ? 'bg-gray-900 border-gray-800' : 'bg-white border-gray-200'} border rounded-lg shadow-lg z-10">
      <div class="flex justify-between items-center p-3 border-b {isDarkMode ? 'border-gray-800' : 'border-gray-200'}">
        <h3 class="font-bold">Recent searches</h3>
        <button 
          class="text-blue-500 text-sm"
          on:click={clearAllSearches}
        >
          Clear all
        </button>
      </div>
      <ul>
        {#each recentSearches as search}
          <li>
            <button 
              class="w-full text-left p-3 {isDarkMode ? 'hover:bg-gray-800' : 'hover:bg-gray-100'} flex items-center"
              on:click={() => selectRecentSearch(search)}
            >
              <span class="{isDarkMode ? 'text-gray-400' : 'text-gray-500'} mr-3">🔍</span>
              {search}
            </button>
          </li>
        {/each}
      </ul>
    </div>
  {/if}
</div> 