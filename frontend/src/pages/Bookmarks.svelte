<script lang="ts">
  import { onMount } from "svelte";
  import { getUserBookmarks, searchBookmarks, likeThread, unlikeThread, removeBookmark } from "../api/thread";
  import MainLayout from "../components/layout/MainLayout.svelte";
  import TweetCard from "../components/social/TweetCard.svelte";
  import { useTheme } from "../hooks/useTheme";
  import SearchIcon from "svelte-feather-icons/src/icons/SearchIcon.svelte";
  import XIcon from "svelte-feather-icons/src/icons/XIcon.svelte";
  import { toastStore } from "../stores/toastStore";

  let bookmarks: any[] = [];
  let loading = true;
  let error: string | null = null;
  let searchQuery: string = "";
  let isSearching = false;
  let searchResults: any[] = [];

  const { theme } = useTheme();
  $: isDarkMode = $theme === "dark";

  async function fetchBookmarks() {
    try {
      loading = true;
      const response = await getUserBookmarks("me");
      console.log("Bookmarks API response:", response);

      if (response && response.success) {
        bookmarks = response.bookmarks || [];
        console.log("Successfully loaded bookmarks:", bookmarks.length);
      } else {
        error = "Failed to load bookmarks";
      }
    } catch (err: any) {
      console.error("Error fetching bookmarks:", err);
      error = err.message || "Failed to load bookmarks";
    } finally {
      loading = false;
    }
  }

  // Function to handle search
  async function handleSearch() {
    if (!searchQuery.trim()) {
      isSearching = false;
      return;
    }

    try {
      isSearching = true;
      loading = true;

      // Try to use the server-side search API
      try {
        const response = await searchBookmarks(searchQuery);

        if (response && response.success) {
          searchResults = response.bookmarks || [];
          console.log("Search results from API:", searchResults.length);
        } else {
          // Fall back to client-side search if server search fails
          performClientSideSearch();
        }
      } catch (err) {
        console.error("API search failed, falling back to client-side search:", err);
        performClientSideSearch();
      }
    } catch (err: any) {
      console.error("Error searching bookmarks:", err);
      error = err.message || "Failed to search bookmarks";
      searchResults = [];
    } finally {
      loading = false;
    }
  }

  // Perform client-side search as fallback
  function performClientSideSearch() {
    const query = searchQuery.toLowerCase().trim();
    searchResults = bookmarks.filter(bookmark => {
      const content = (bookmark.content || "").toLowerCase();
      const username = (bookmark.username || "").toLowerCase();
      const name = (bookmark.name || bookmark.displayName || "").toLowerCase();

      return content.includes(query) ||
             username.includes(query) ||
             name.includes(query);
    });
    console.log("Client-side search results:", searchResults.length);
  }

  // Function to clear search
  function clearSearch() {
    searchQuery = "";
    isSearching = false;
    searchResults = [];
  }

  // Handle search form submission
  function handleSubmit(event: Event) {
    event.preventDefault();
    handleSearch();
  }

  // Handle like event
  async function handleLike(event: CustomEvent) {
    const threadId = event.detail;
    try {
      await likeThread(threadId);

      // Update the bookmark or search result
      const updatedBookmarks = isSearching ? [...searchResults] : [...bookmarks];
      const index = updatedBookmarks.findIndex(b => b.id === threadId);

      if (index !== -1) {
        updatedBookmarks[index].is_liked = true;
        updatedBookmarks[index].likes_count = (updatedBookmarks[index].likes_count || 0) + 1;

        if (isSearching) {
          searchResults = updatedBookmarks;
        } else {
          bookmarks = updatedBookmarks;
        }
      }

      toastStore.showToast("Post liked", "success");
    } catch (error: any) {
      console.error("Error liking thread:", error);
      toastStore.showToast("Failed to like post. Please try again.", "error");
    }
  }

  // Handle unlike event
  async function handleUnlike(event: CustomEvent) {
    const threadId = event.detail;
    try {
      await unlikeThread(threadId);

      // Update the bookmark or search result
      const updatedBookmarks = isSearching ? [...searchResults] : [...bookmarks];
      const index = updatedBookmarks.findIndex(b => b.id === threadId);

      if (index !== -1) {
        updatedBookmarks[index].is_liked = false;
        updatedBookmarks[index].likes_count = Math.max(0, (updatedBookmarks[index].likes_count || 0) - 1);

        if (isSearching) {
          searchResults = updatedBookmarks;
        } else {
          bookmarks = updatedBookmarks;
        }
      }

      toastStore.showToast("Post unliked", "success");
    } catch (error: any) {
      console.error("Error unliking thread:", error);
      toastStore.showToast("Failed to unlike post. Please try again.", "error");
    }
  }

  // Handle remove bookmark event
  async function handleRemoveBookmark(event: CustomEvent) {
    const threadId = event.detail;
    try {
      await removeBookmark(threadId);

      // Remove the item from bookmarks list
      if (isSearching) {
        searchResults = searchResults.filter(bookmark => bookmark.id !== threadId);
        // Also remove from main bookmarks list
        bookmarks = bookmarks.filter(bookmark => bookmark.id !== threadId);
      } else {
        bookmarks = bookmarks.filter(bookmark => bookmark.id !== threadId);
      }

      toastStore.showToast("Post removed from bookmarks", "success");
    } catch (error: any) {
      console.error("Error removing bookmark:", error);
      toastStore.showToast("Failed to remove bookmark. Please try again.", "error");
    }
  }

  onMount(fetchBookmarks);
</script>

<MainLayout>
  <div class="bookmarks-page">
    <div class="bookmarks-header">
      <h1>Bookmarks</h1>
      <form class="search-form" on:submit={handleSubmit}>
        <div class="search-container">
          <SearchIcon size="16" class="search-icon" />
          <input
            type="text"
            placeholder="Search bookmarks"
            bind:value={searchQuery}
            class="search-input"
          />
          {#if searchQuery}
            <button type="button" class="clear-button" on:click={clearSearch} aria-label="Clear search">
              <XIcon size="16" />
            </button>
          {/if}
        </div>
        <button type="submit" class="search-button">Search</button>
      </form>
    </div>

    {#if loading}
      <div class="loading">Loading bookmarks...</div>
    {:else if error}
      <div class="error-message">
        <p>{error}</p>
      </div>
    {:else if isSearching && searchResults.length === 0}
      <div class="empty-state">
        <h3>No matching bookmarks found</h3>
        <p>Try a different search term or <button class="reset-button" on:click={clearSearch}>view all bookmarks</button>.</p>
      </div>
    {:else if !isSearching && bookmarks.length === 0}
      <div class="empty-state">
        <h3>No bookmarks yet</h3>
        <p>Threads you bookmark will appear here.</p>
      </div>
    {:else}
      <div class="bookmarks-list">
        {#each isSearching ? searchResults : bookmarks as bookmark (bookmark.id)}
          <TweetCard
            tweet={bookmark}
            {isDarkMode}
            on:like={handleLike}
            on:unlike={handleUnlike}
            on:removeBookmark={handleRemoveBookmark}
          />
        {/each}
      </div>
    {/if}
  </div>
</MainLayout>

<style>
  .bookmarks-page {
    max-width: 800px;
    margin: 0 auto;
    padding: 20px;
  }

  .bookmarks-header {
    display: flex;
    flex-direction: column;
    gap: 16px;
    margin-bottom: 24px;
  }

  h1 {
    margin-bottom: 0;
  }

  .search-form {
    display: flex;
    gap: 8px;
  }

  .search-container {
    position: relative;
    flex-grow: 1;
    display: flex;
    align-items: center;
  }

  .search-icon {
    position: absolute;
    left: 12px;
    color: var(--text-secondary);
  }

  .search-input {
    width: 100%;
    padding: 10px 36px;
    border: 1px solid var(--border-color, #eaeaea);
    border-radius: 20px;
    font-size: 14px;
    background-color: var(--bg-secondary);
    color: var(--text-primary);
  }

  .clear-button {
    position: absolute;
    right: 12px;
    background: none;
    border: none;
    cursor: pointer;
    color: var(--text-secondary);
    padding: 0;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .search-button {
    padding: 10px 16px;
    background-color: var(--color-primary);
    color: white;
    border: none;
    border-radius: 20px;
    font-weight: bold;
    cursor: pointer;
    transition: background-color 0.2s;
  }

  .search-button:hover {
    background-color: var(--color-primary-hover);
  }

  .loading {
    text-align: center;
    padding: 40px;
    color: var(--text-secondary);
  }

  .error-message {
    color: var(--color-error, #ff0000);
    padding: 20px;
    background-color: var(--error-bg, #fff1f0);
    border-radius: 8px;
    margin: 20px 0;
  }

  .empty-state {
    text-align: center;
    padding: 40px 20px;
    color: var(--text-secondary);
  }

  .bookmarks-list {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .reset-button {
    background: none;
    border: none;
    color: var(--color-primary);
    cursor: pointer;
    font-weight: bold;
    padding: 0;
    text-decoration: underline;
  }
</style>