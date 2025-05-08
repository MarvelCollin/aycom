<script lang="ts">
  import { onMount } from 'svelte';
  import { addChatParticipant, removeChatParticipant, listChatParticipants } from '../../api/chat';
  import { createLoggerWithPrefix } from '../../utils/logger';
  import { searchUsers } from '../../api/user';
  import { getAuthToken } from '../../utils/auth';
  import appConfig from '../../config/appConfig';
  
  const API_BASE_URL = appConfig.api.baseUrl;
  const logger = createLoggerWithPrefix('ManageGroupMembers');
  
  // Define interfaces
  interface User {
    id: string;
    username: string;
    display_name?: string;
    avatar_url?: string;
    role?: string;
  }
  
  // Props
  export let chatId: string;
  export let onClose: () => void = () => {};
  export let onMembersUpdated: () => void = () => {};
  
  // State
  let currentParticipants: User[] = [];
  let availableUsers: User[] = [];
  let searchQuery = '';
  let isLoading = true;
  let isAddingMember = false;
  let isRemovingMember = false;
  let errorMessage = '';
  let successMessage = '';
  
  // Load participants and available users on mount
  onMount(async () => {
    try {
      await loadParticipants();
      await loadAvailableUsers();
      isLoading = false;
    } catch (error) {
      logger.error('Error initializing group management:', error);
      errorMessage = 'Failed to load group members. Please try again.';
      isLoading = false;
    }
  });
  
  // Load current participants
  async function loadParticipants(): Promise<void> {
    try {
      const response = await listChatParticipants(chatId);
      currentParticipants = response.participants || [];
    } catch (error) {
      logger.error('Error loading participants:', error);
      throw new Error('Failed to load current participants');
    }
  }
  
  // Load available users to add
  async function loadAvailableUsers(): Promise<void> {
    try {
      logger.debug('Attempting to load available users with searchUsers API');
      // Use common letter 'a' to get many results
      const response = await searchUsers('a', 1, 50);
      
      if (response && response.users && response.users.length > 0) {
        logger.debug('Users loaded successfully from searchUsers API', { count: response.users.length });
        
        // Filter out users who are already in the chat
        const participantIds = new Set(currentParticipants.map(p => p.id));
        
        // Map the user objects to match our format
        availableUsers = response.users
          .filter(user => !participantIds.has(user.id))
          .map(user => ({
            id: user.id,
            username: user.username,
            display_name: user.display_name || user.name || user.username,
            avatar_url: user.avatar_url || user.profile_picture_url || user.avatar || ''
          }));
          
        logger.debug('Filtered available users', { count: availableUsers.length });
      } else {
        // Try alternative queries if no results
        await tryAlternativeSearchQueries();
      }
    } catch (error) {
      logger.error('Error loading users with searchUsers API:', error);
      await tryAlternativeSearchQueries();
    }
  }
  
  // Try different common letters that should return users
  async function tryAlternativeSearchQueries(): Promise<void> {
    // Try a series of common letters that should return results
    const commonLetters = ['e', 'i', 'o', 's', 'm'];
    
    for (const letter of commonLetters) {
      try {
        logger.debug(`Trying alternative search with letter "${letter}"`);
        const response = await searchUsers(letter, 1, 50);
        
        if (response && response.users && response.users.length > 0) {
          logger.debug(`Users found with letter "${letter}"`, { count: response.users.length });
          
          // Filter out users who are already in the chat
          const participantIds = new Set(currentParticipants.map(p => p.id));
          
          availableUsers = response.users
            .filter(user => !participantIds.has(user.id))
            .map(user => ({
              id: user.id,
              username: user.username,
              display_name: user.display_name || user.name || user.username,
              avatar_url: user.avatar_url || user.profile_picture_url || user.avatar || ''
            }));
            
          logger.debug('Filtered available users from alternative search', { count: availableUsers.length });
          return; // Exit after finding users
        }
      } catch (err) {
        logger.warn(`Search with letter "${letter}" failed:`, err);
      }
    }
    
    logger.warn('All alternative searches failed to load users');
    errorMessage = 'Could not load available users. Try searching by username.';
  }
  
  // Filtered users based on search query
  $: filteredUsers = searchQuery 
    ? availableUsers.filter(user => 
        user.username.toLowerCase().includes(searchQuery.toLowerCase()) || 
        (user.display_name && user.display_name.toLowerCase().includes(searchQuery.toLowerCase()))
      )
    : availableUsers;
  
  // Real-time search for users when typing
  $: if (searchQuery && searchQuery.length > 1) {
    performUserSearch(searchQuery);
  }

  // Debounce search to avoid too many API calls
  let searchTimeout: ReturnType<typeof setTimeout> | null = null;
  
  function performUserSearch(query: string) {
    if (searchTimeout) {
      clearTimeout(searchTimeout);
    }
    
    // Only search if query is at least 2 characters
    if (query.length < 2) return;
    
    searchTimeout = setTimeout(async () => {
      logger.debug(`Performing user search for "${query}"`);
      
      try {
        const response = await searchUsers(query, 1, 20);
        
        if (response && response.users && response.users.length > 0) {
          logger.debug(`Found ${response.users.length} users matching "${query}"`);
          
          // Filter out users who are already in the chat
          const participantIds = new Set(currentParticipants.map(p => p.id));
          
          const newUsers = response.users
            .filter(user => !participantIds.has(user.id))
            .map(user => ({
              id: user.id,
              username: user.username,
              display_name: user.display_name || user.name || user.username,
              avatar_url: user.avatar_url || user.profile_picture_url || user.avatar || ''
            }));
          
          // Merge new users with existing ones, avoiding duplicates
          const existingIds = new Set(availableUsers.map(u => u.id));
          const uniqueNewUsers = newUsers.filter(u => !existingIds.has(u.id));
          
          if (uniqueNewUsers.length > 0) {
            availableUsers = [...availableUsers, ...uniqueNewUsers];
            logger.debug(`Added ${uniqueNewUsers.length} new users to available list`);
          }
        }
      } catch (error) {
        logger.warn(`User search for "${query}" failed:`, error);
      }
    }, 300); // 300ms debounce time
  }
  
  // Add a user to the chat
  async function handleAddMember(user: User): Promise<void> {
    errorMessage = '';
    successMessage = '';
    isAddingMember = true;
    
    try {
      await addChatParticipant(chatId, { user_id: user.id });
      
      // Update local state
      currentParticipants = [...currentParticipants, user];
      availableUsers = availableUsers.filter(u => u.id !== user.id);
      
      successMessage = `Added ${user.display_name || user.username} to the chat`;
      
      // Notify parent component
      onMembersUpdated();
    } catch (error) {
      logger.error('Error adding member:', error);
      errorMessage = 'Failed to add member. Please try again.';
    } finally {
      isAddingMember = false;
      
      // Clear success message after a delay
      if (successMessage) {
        setTimeout(() => {
          successMessage = '';
        }, 3000);
      }
    }
  }
  
  // Remove a user from the chat
  async function handleRemoveMember(user: User): Promise<void> {
    errorMessage = '';
    successMessage = '';
    isRemovingMember = true;
    
    try {
      await removeChatParticipant(chatId, user.id);
      
      // Update local state
      currentParticipants = currentParticipants.filter(p => p.id !== user.id);
      availableUsers = [...availableUsers, user];
      
      successMessage = `Removed ${user.display_name || user.username} from the chat`;
      
      // Notify parent component
      onMembersUpdated();
    } catch (error) {
      logger.error('Error removing member:', error);
      errorMessage = 'Failed to remove member. Please try again.';
    } finally {
      isRemovingMember = false;
      
      // Clear success message after a delay
      if (successMessage) {
        setTimeout(() => {
          successMessage = '';
        }, 3000);
      }
    }
  }
</script>

<div class="manage-members-modal">
  <div class="modal-header">
    <h2>Manage Group Members</h2>
    <button class="close-button" on:click={onClose}>✕</button>
  </div>
  
  <div class="modal-body">
    {#if isLoading}
      <div class="loading-state">Loading members...</div>
    {:else if errorMessage}
      <div class="error-message">{errorMessage}</div>
    {:else}
      {#if successMessage}
        <div class="success-message">{successMessage}</div>
      {/if}
      
      <div class="members-section">
        <h3>Current Members ({currentParticipants.length})</h3>
        
        {#if currentParticipants.length === 0}
          <p class="empty-state">No members in this chat</p>
        {:else}
          <div class="member-list">
            {#each currentParticipants as user (user.id)}
              <div class="member-item">
                <div class="user-avatar">
                  {#if user.avatar_url}
                    <img src={user.avatar_url} alt={user.username} />
                  {:else}
                    <div class="avatar-placeholder">
                      {(user.display_name || user.username)[0].toUpperCase()}
                    </div>
                  {/if}
                </div>
                
                <div class="user-info">
                  <div class="user-name">{user.display_name || user.username}</div>
                  <div class="user-username">@{user.username}</div>
                </div>
                
                {#if user.role === 'admin'}
                  <div class="role-badge admin">Admin</div>
                {:else}
                  <button 
                    class="remove-button" 
                    on:click={() => handleRemoveMember(user)}
                    disabled={isRemovingMember}
                  >
                    Remove
                  </button>
                {/if}
              </div>
            {/each}
          </div>
        {/if}
      </div>
      
      <div class="add-members-section">
        <h3>Add Members</h3>
        <div class="search-container">
          <div class="search-input-wrapper">
            <input 
              type="text" 
              bind:value={searchQuery} 
              placeholder="Search users..."
              class="search-input"
            />
            {#if searchQuery.trim() !== ''}
              <button class="clear-search-button" on:click={() => searchQuery = ''}>
                <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" width="16" height="16">
                  <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clip-rule="evenodd" />
                </svg>
              </button>
            {/if}
          </div>

          {#if searchQuery.trim() !== '' && filteredUsers.length > 0}
            <div class="search-dropdown">
              <div class="search-dropdown-section">
                <h4 class="search-dropdown-title">Users</h4>
                <ul class="search-dropdown-list">
                  {#each filteredUsers as user (user.id)}
                    <li>
                      <div class="dropdown-item" on:click={() => handleAddMember(user)}>
                        <div class="avatar-container">
                          {#if user.avatar_url}
                            <img src={user.avatar_url} alt={user.display_name || user.username} class="avatar-image" />
                          {:else}
                            <span class="avatar-placeholder">{(user.display_name || user.username)[0].toUpperCase()}</span>
                          {/if}
                        </div>
                        <div class="user-info">
                          <span class="user-name">{user.display_name || user.username}</span>
                          <span class="user-username">@{user.username}</span>
                        </div>
                        <button 
                          class="add-button" 
                          on:click|stopPropagation={() => handleAddMember(user)}
                          disabled={isAddingMember}
                        >
                          Add
                        </button>
                      </div>
                    </li>
                  {/each}
                </ul>
              </div>
            </div>
          {/if}
        </div>
        
        {#if searchQuery.trim() === ''}
          <p class="help-text">Type to search for users</p>
        {:else if filteredUsers.length === 0}
          <p class="empty-state">No users found</p>
        {/if}
      </div>
    {/if}
  </div>
  
  <div class="modal-footer">
    <button class="close-button-footer" on:click={onClose}>Close</button>
  </div>
</div>

<style>
  .manage-members-modal {
    display: flex;
    flex-direction: column;
    width: 100%;
    height: 100%;
    background-color: white;
    border-radius: 8px;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
    overflow: hidden;
  }
  
  .modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 16px;
    border-bottom: 1px solid #e5e5e5;
  }
  
  .modal-header h2 {
    margin: 0;
    font-size: 1.25rem;
  }
  
  .close-button {
    background: none;
    border: none;
    font-size: 1.25rem;
    cursor: pointer;
    color: #6c757d;
  }
  
  .modal-body {
    flex: 1;
    padding: 16px;
    overflow-y: auto;
  }
  
  .loading-state {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 100px;
    color: #6c757d;
  }
  
  .error-message {
    background-color: #f8d7da;
    color: #721c24;
    padding: 10px;
    border-radius: 4px;
    margin-bottom: 16px;
  }
  
  .success-message {
    background-color: #d4edda;
    color: #155724;
    padding: 10px;
    border-radius: 4px;
    margin-bottom: 16px;
  }
  
  h3 {
    margin-top: 0;
    margin-bottom: 12px;
    font-size: 1.1rem;
  }
  
  .members-section, .add-members-section {
    margin-bottom: 24px;
  }
  
  .search-container {
    position: relative;
    margin-bottom: 12px;
  }
  
  .search-input-wrapper {
    position: relative;
    display: flex;
    align-items: center;
  }
  
  .search-input {
    width: 100%;
    padding: 8px 12px;
    border: 1px solid #ced4da;
    border-radius: 4px;
    font-size: 1rem;
  }
  
  .clear-search-button {
    position: absolute;
    right: 8px;
    top: 50%;
    transform: translateY(-50%);
    background: none;
    border: none;
    padding: 4px;
    cursor: pointer;
    color: #6c757d;
    display: flex;
    align-items: center;
    justify-content: center;
  }
  
  .search-dropdown {
    position: absolute;
    top: calc(100% + 4px);
    left: 0;
    right: 0;
    max-height: 300px;
    overflow-y: auto;
    background-color: white;
    border: 1px solid #e5e5e5;
    border-radius: 4px;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
    z-index: 1000;
  }
  
  .search-dropdown-section {
    padding: 8px;
  }
  
  .search-dropdown-title {
    font-size: 0.75rem;
    color: #6c757d;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    margin: 4px 4px 12px;
    font-weight: 600;
  }
  
  .search-dropdown-list {
    list-style: none;
    padding: 0;
    margin: 0;
  }
  
  .dropdown-item {
    display: flex;
    align-items: center;
    padding: 8px;
    cursor: pointer;
    border: none;
    background: none;
    width: 100%;
    text-align: left;
  }
  
  .dropdown-item:hover {
    background-color: #f8f9fa;
  }
  
  .help-text {
    color: #6c757d;
    font-size: 0.9rem;
    text-align: center;
    margin: 16px 0;
  }
  
  .member-list {
    border: 1px solid #e5e5e5;
    border-radius: 4px;
    max-height: 300px;
    overflow-y: auto;
  }
  
  .member-item {
    display: flex;
    align-items: center;
    padding: 12px;
    border-bottom: 1px solid #e5e5e5;
  }
  
  .member-item:last-child {
    border-bottom: none;
  }
  
  .avatar-container, .user-avatar {
    width: 40px;
    height: 40px;
    border-radius: 50%;
    margin-right: 12px;
    overflow: hidden;
    flex-shrink: 0;
  }
  
  .avatar-image, .user-avatar img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }
  
  .avatar-placeholder {
    width: 100%;
    height: 100%;
    background-color: #6c757d;
    color: white;
    display: flex;
    align-items: center;
    justify-content: center;
    font-weight: bold;
  }
  
  .user-info {
    flex: 1;
  }
  
  .user-name {
    font-weight: 500;
    display: block;
  }
  
  .user-username {
    font-size: 0.875rem;
    color: #6c757d;
    display: block;
  }
  
  .role-badge {
    font-size: 0.75rem;
    padding: 2px 8px;
    border-radius: 12px;
    font-weight: 500;
  }
  
  .role-badge.admin {
    background-color: #e9ecef;
    color: #495057;
  }
  
  .remove-button, .add-button {
    padding: 4px 8px;
    border-radius: 4px;
    font-size: 0.875rem;
    cursor: pointer;
    white-space: nowrap;
  }
  
  .remove-button {
    background-color: #f8d7da;
    color: #721c24;
    border: 1px solid #f5c6cb;
  }
  
  .add-button {
    background-color: #d4edda;
    color: #155724;
    border: 1px solid #c3e6cb;
    margin-left: 8px;
  }
  
  .remove-button:disabled, .add-button:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
  
  .empty-state {
    text-align: center;
    color: #6c757d;
    padding: 16px;
  }
  
  .modal-footer {
    display: flex;
    justify-content: flex-end;
    padding: 16px;
    border-top: 1px solid #e5e5e5;
  }
  
  .close-button-footer {
    padding: 8px 16px;
    background-color: #6c757d;
    color: white;
    border: none;
    border-radius: 4px;
    cursor: pointer;
  }
</style> 