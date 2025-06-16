<script lang="ts">
  import { onMount } from "svelte";

  export let isDarkMode = false;
  let searchQuery = "";
  let isFocused = false;
  let recentSearches: string[] = [];

  onMount(() => {
    const savedSearches = localStorage.getItem("recentSearches");
    if (savedSearches) {
      recentSearches = JSON.parse(savedSearches).slice(0, 3);
    }
  });

  function saveSearch(query: string) {
    if (!query.trim()) return;

    const newSearches = [
      query,
      ...recentSearches.filter(s => s !== query)
    ].slice(0, 3);

    recentSearches = newSearches;
    localStorage.setItem("recentSearches", JSON.stringify(newSearches));
  }

  function clearAllSearches() {
    recentSearches = [];
    localStorage.removeItem("recentSearches");
  }

  function handleSearch() {
    if (searchQuery.trim()) {
      saveSearch(searchQuery);
      window.location.href = `/explore?q=${encodeURIComponent(searchQuery)}`;
    }
  }

  function handleKeyPress(event: KeyboardEvent) {
    if (event.key === "Enter") {
      handleSearch();
    }
  }

  function selectRecentSearch(search: string) {
    searchQuery = search;
    handleSearch();
  }

  function toggleFocus(value: boolean) {
    setTimeout(() => {
      isFocused = value;
    }, 100);
  }
</script>

<div class="relative mb-3">
  <div class="relative">
    <button
      class="absolute left-3 top-[14px] {isDarkMode ? "text-gray-400" : "text-gray-500"}"
      on:click={handleSearch}
    >
      🔍
    </button>
    <input
      type="text"
      bind:value={searchQuery}
      placeholder="Search"
      class="w-full {isDarkMode ? "bg-gray-800 text-white" : "bg-gray-100 text-gray-900"} rounded-full py-3 pl-10 pr-4 focus:outline-none focus:ring-1 focus:ring-blue-500"
      on:focus={() => toggleFocus(true)}
      on:blur={() => toggleFocus(false)}
      on:keypress={handleKeyPress}
    />
  </div>

  {#if isFocused && recentSearches.length > 0}
    <div class="absolute left-0 right-0 mt-1 {isDarkMode ? "bg-gray-900 border-gray-800" : "bg-white border-gray-200"} border rounded-xl shadow-lg z-10 overflow-hidden">
      <div class="flex justify-between items-center p-3 border-b {isDarkMode ? "border-gray-700 border-opacity-50" : "border-gray-200"}">
        <h3 class="font-bold">Recent searches</h3>
        <button
          class="text-blue-500 text-sm hover:underline"
          on:click={clearAllSearches}
        >
          Clear all
        </button>
      </div>
      <ul>
        {#each recentSearches as search}
          <li>
            <button
              class="w-full text-left p-3 {isDarkMode ? "hover:bg-gray-800" : "hover:bg-gray-100"} flex items-center"
              on:click={() => selectRecentSearch(search)}
            >
              <span class="{isDarkMode ? "text-gray-400" : "text-gray-500"} mr-3">🔍</span>
              {search}
            </button>
          </li>
        {/each}
      </ul>
    </div>
  {/if}
</div>