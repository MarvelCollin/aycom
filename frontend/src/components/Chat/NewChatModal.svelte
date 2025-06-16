<script lang="ts">
  import { onMount } from "svelte";
  import { createEventDispatcher } from "svelte";
  import { getAllUsers, searchUsers } from "../../api/user";
  import { transformApiUsers, type StandardUser } from "../../utils/userTransform";
  import { createLoggerWithPrefix } from "../../utils/logger";
  import { toastStore } from "../../stores/toastStore";

  const logger = createLoggerWithPrefix("NewChatModal");
  const dispatch = createEventDispatcher();

  export let onCancel: () => void;

  export let isLoadingUsers = false;
  export let userSearchResults: StandardUser[] = [];
  export let searchKeyword = "";

  let searchQuery = "";
  let users: StandardUser[] = [];
  let filteredUsers: StandardUser[] = [];
  let isLoading = false;
  let searchTimeout: ReturnType<typeof setTimeout>;

  onMount(async () => {
    await loadUsers();
  });

  async function loadUsers() {
    isLoading = true;
    try {
      console.log("Loading users...");

      const testResponse = await fetch("http://localhost:8083/api/v1/users?limit=50", {
        method: "GET",
        headers: {
          "Content-Type": "application/json"
        }
      });

      console.log("Direct API test - Status:", testResponse.status);

      if (testResponse.ok) {
        const testData = await testResponse.json();
        console.log("Direct API test - Data:", testData);

        if (testData && testData.users && Array.isArray(testData.users)) {
          users = transformApiUsers(testData.users);
          filteredUsers = users;
          console.log("Direct API success - Users loaded:", users.length);
          return;
        }
      }

      console.log("Trying getAllUsers API call...");
      const response = await getAllUsers(1, 50);
      console.log("getAllUsers response:", response);

      if (response && response.users && Array.isArray(response.users)) {
        users = transformApiUsers(response.users);
        console.log("Transformed users count:", users.length);
      } else if (response && Array.isArray(response)) {

        users = transformApiUsers(response);
        console.log("Transformed users count (direct array):", users.length);
      } else {
        console.error("Unexpected response format:", response);
        toastStore.showToast("Unexpected response format from server", "error");
        users = [];
      }

      filteredUsers = users;
      console.log("Final filtered users count:", filteredUsers.length);

      if (users.length === 0) {
        toastStore.showToast("No users found", "warning");
      }
    } catch (error) {
      console.error("Error in loadUsers:", error);
      logger.error("Failed to load users", error);
      toastStore.showToast(`Failed to load users: ${error instanceof Error ? error.message : String(error)}`, "error");
      users = [];
      filteredUsers = [];
    } finally {
      isLoading = false;
    }
  }

  function handleSearch() {
    if (searchTimeout) {
      clearTimeout(searchTimeout);
    }

    searchTimeout = setTimeout(() => {
      if (!searchQuery.trim()) {
        filteredUsers = users;
        return;
      }

      const query = searchQuery.toLowerCase();
      filteredUsers = users.filter(user =>
        user.username.toLowerCase().includes(query) ||
        (user.display_name && user.display_name.toLowerCase().includes(query))
      );
    }, 300);

    dispatch("search", searchQuery.trim());
  }

  function handleUserClick(user: StandardUser) {

    dispatch("createChat", {
      type: "individual",
      participants: [user.id]
    });

    if (typeof onCancel === "function") {
      onCancel();
    }
  }

  function getAvatarColor(name: string) {
    let hash = 0;
    for (let i = 0; i < name.length; i++) {
      hash = name.charCodeAt(i) + ((hash << 5) - hash);
    }
    const h = Math.abs(hash) % 360;
    return `hsl(${h}, 70%, 60%)`;
  }
</script>

<div class="modal-overlay" role="button" tabindex="0" on:click={onCancel} on:keydown={(e) => e.key === "Escape" && onCancel()}>
  <div class="modal-content" role="dialog" tabindex="0" on:click|stopPropagation on:keydown|stopPropagation>
    <div class="modal-header">
      <h2>New Message</h2>
      <button class="close-button" on:click={onCancel} aria-label="Close modal">
        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" width="24" height="24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
        </svg>
      </button>
    </div>

    <div class="search-container">
      <input
        type="text"
        placeholder="Search for users..."
        bind:value={searchQuery}
        on:input={handleSearch}
        class="search-input"
      />
    </div>

    <div class="users-list">
      {#if isLoading}
        <div class="loading-container">
          <div class="loading-spinner"></div>
          <p>Loading users...</p>
        </div>
      {:else if filteredUsers.length === 0}
        <div class="empty-state">
          <p>No users found</p>
        </div>
      {:else}        {#each filteredUsers as user}
          <button
            class="user-item"
            on:click={() => handleUserClick(user)}
            type="button"
          >
            <div class="avatar">
              {#if user.avatar}
                <img src={user.avatar} alt={user.display_name || user.username} />
              {:else}
                <div class="avatar-placeholder" style="background-color: {getAvatarColor(user.username)}">
                  {(user.display_name || user.username).charAt(0).toUpperCase()}
                </div>
              {/if}
            </div>
            <div class="user-details">
              <div class="user-name">
                {user.display_name || user.username}
                {#if user.is_verified}
                  <span class="verified-badge">✓</span>
                {/if}
              </div>
              <div class="username">@{user.username}</div>
            </div>
          </button>
        {/each}
      {/if}
    </div>
  </div>
</div>

<style>
  .modal-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background-color: rgba(0, 0, 0, 0.7);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
  }

  .modal-content {
    background-color: var(--bg-secondary, #ffffff);
    border-radius: 12px;
    width: 90%;
    max-width: 500px;
    max-height: 80vh;
    display: flex;
    flex-direction: column;
    border: 1px solid var(--border-color, #e5e7eb);
    color: var(--text-color, #333333);
  }

  .modal-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 20px 24px;
    border-bottom: 1px solid var(--border-color, #e5e7eb);
  }

  .modal-header h2 {
    margin: 0;
    color: var(--text-color, #333333);
    font-size: 18px;
    font-weight: 600;
  }

  .close-button {
    background: none;
    border: none;
    color: var(--text-secondary, #6b7280);
    cursor: pointer;
    padding: 4px;
    border-radius: 4px;
    transition: background-color 0.2s;
  }

  .close-button:hover {
    background-color: var(--hover-bg, rgba(0, 0, 0, 0.05));
    color: var(--text-color, #333333);
  }

  .search-container {
    padding: 20px 24px;
    border-bottom: 1px solid var(--border-color, #e5e7eb);
  }

  .search-input {
    width: 100%;
    padding: 12px 16px;
    background-color: var(--input-bg, #ffffff);
    border: 1px solid var(--border-color, #e5e7eb);
    border-radius: 8px;
    color: var(--text-color, #333333);
    font-size: 16px;
    outline: none;
  }

  .search-input:focus {
    border-color: var(--accent-color, #1d9bf0);
  }

  .search-input::placeholder {
    color: var(--text-secondary, #6b7280);
  }

  .users-list {
    flex: 1;
    overflow-y: auto;
    max-height: 400px;
  }

  .loading-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 40px;
    color: var(--text-secondary, #6b7280);
  }

  .loading-spinner {
    width: 32px;
    height: 32px;
    border: 3px solid var(--border-color, #e5e7eb);
    border-top: 3px solid var(--accent-color, #1d9bf0);
    border-radius: 50%;
    animation: spin 1s linear infinite;
    margin-bottom: 16px;
  }

  @keyframes spin {
    0% { transform: rotate(0deg); }
    100% { transform: rotate(360deg); }
  }

  .empty-state {
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 40px;
    color: var(--text-secondary, #6b7280);
  }
  .user-item {
    display: flex;
    align-items: center;
    padding: 16px 24px;
    cursor: pointer;
    transition: background-color 0.2s;
    border-bottom: 1px solid var(--border-color, #e5e7eb);
    background: none;
    border: none;
    border-bottom: 1px solid var(--border-color, #e5e7eb);
    width: 100%;
    text-align: left;
    color: inherit;
  }

  .user-item:hover {
    background-color: var(--hover-bg, rgba(0, 0, 0, 0.05));
  }

  .user-item:last-child {
    border-bottom: none;
  }

  .avatar {
    width: 48px;
    height: 48px;
    border-radius: 50%;
    margin-right: 16px;
    flex-shrink: 0;
  }

  .avatar img {
    width: 100%;
    height: 100%;
    border-radius: 50%;
    object-fit: cover;
  }

  .avatar-placeholder {
    width: 100%;
    height: 100%;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    color: white;
    font-weight: 600;
    font-size: 18px;
  }

  .user-details {
    flex: 1;
    min-width: 0;
  }

  .user-name {
    display: flex;
    align-items: center;
    gap: 6px;
    color: var(--text-color, #333333);
    font-weight: 600;
    font-size: 16px;
    margin-bottom: 4px;
  }

  .verified-badge {
    color: var(--accent-color, #1d9bf0);
    font-size: 14px;
  }

  .username {
    color: var(--text-secondary, #6b7280);
    font-size: 14px;
  }

  @media (max-width: 768px) {
    .modal-content {
      width: 95%;
      max-height: 90vh;
    }

    .modal-header,
    .search-container {
      padding: 16px 20px;
    }

    .user-item {
      padding: 12px 20px;
    }

    .avatar {
      width: 40px;
      height: 40px;
      margin-right: 12px;
    }

    .user-name {
      font-size: 15px;
    }

    .username {
      font-size: 13px;
    }
  }
</style>