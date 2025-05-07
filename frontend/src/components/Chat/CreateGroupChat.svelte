<script lang="ts">
  import { onMount } from 'svelte';
  import { createChat } from '../../api/chat';
  import { createLoggerWithPrefix } from '../../utils/logger';
  
  const logger = createLoggerWithPrefix('CreateGroupChat');
  
  // Define user interface
  interface User {
    id: string;
    username: string;
    display_name?: string;
    avatar_url?: string;
  }
  
  // Props
  export let onSuccess: (data: any) => void = () => {};
  export let onCancel: () => void = () => {};
  
  // State
  let chatName = '';
  let searchQuery = '';
  let availableUsers: User[] = [];
  let selectedUsers: User[] = [];
  let isLoading = false;
  let errorMessage = '';
  
  // Load available users on mount
  onMount(async () => {
    try {
      // Fetch available users that can be added to chat
      const response = await fetch('/api/v1/users/suggestions', {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        }
      });
      
      if (!response.ok) {
        throw new Error('Failed to load users');
      }
      
      const data = await response.json();
      availableUsers = data.users || [];
    } catch (error) {
      logger.error('Error loading users:', error);
      errorMessage = 'Failed to load users. Please try again.';
    }
  });
  
  // Filtered users based on search query
  $: filteredUsers = searchQuery 
    ? availableUsers.filter(user => 
        user.username.toLowerCase().includes(searchQuery.toLowerCase()) || 
        (user.display_name && user.display_name.toLowerCase().includes(searchQuery.toLowerCase()))
      )
    : availableUsers;
  
  // Check if a user is already selected
  function isUserSelected(userId: string): boolean {
    return selectedUsers.some(user => user.id === userId);
  }
  
  // Add user to selection
  function selectUser(user: User): void {
    if (!isUserSelected(user.id)) {
      selectedUsers = [...selectedUsers, user];
    }
  }
  
  // Remove user from selection
  function removeUser(userId: string): void {
    selectedUsers = selectedUsers.filter(user => user.id !== userId);
  }
  
  // Handle form submission
  async function handleSubmit(): Promise<void> {
    if (!chatName.trim()) {
      errorMessage = 'Please enter a chat name';
      return;
    }
    
    if (selectedUsers.length === 0) {
      errorMessage = 'Please select at least one user';
      return;
    }
    
    isLoading = true;
    errorMessage = '';
    
    try {
      // Create new group chat
      const response = await createChat({
        name: chatName,
        participants: selectedUsers.map(user => user.id),
        is_group: true
      });
      
      onSuccess(response);
    } catch (error) {
      logger.error('Error creating chat:', error);
      errorMessage = 'Failed to create chat. Please try again.';
    } finally {
      isLoading = false;
    }
  }
</script>

<div class="create-group-modal">
  <div class="modal-header">
    <h2>Create Group Chat</h2>
    <button class="close-button" on:click={onCancel}>✕</button>
  </div>
  
  <div class="modal-body">
    {#if errorMessage}
      <div class="error-message">{errorMessage}</div>
    {/if}
    
    <div class="form-group">
      <label for="chat-name">Group Name</label>
      <input 
        id="chat-name" 
        type="text" 
        bind:value={chatName} 
        placeholder="Enter group name"
      />
    </div>
    
    <div class="form-group">
      <label>Add Participants</label>
      <input 
        type="text" 
        bind:value={searchQuery} 
        placeholder="Search users..."
      />
    </div>
    
    <div class="selected-users">
      {#if selectedUsers.length > 0}
        <h4>Selected Users ({selectedUsers.length})</h4>
        <div class="user-chips">
          {#each selectedUsers as user (user.id)}
            <div class="user-chip">
              <span>{user.display_name || user.username}</span>
              <button on:click={() => removeUser(user.id)}>✕</button>
            </div>
          {/each}
        </div>
      {/if}
    </div>
    
    <div class="user-list">
      <h4>Available Users</h4>
      {#if filteredUsers.length === 0}
        <p class="empty-state">No users found</p>
      {:else}
        {#each filteredUsers as user (user.id)}
          <div 
            class="user-item {isUserSelected(user.id) ? 'selected' : ''}" 
            on:click={() => selectUser(user)}
          >
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
            {#if isUserSelected(user.id)}
              <div class="selected-indicator">✓</div>
            {/if}
          </div>
        {/each}
      {/if}
    </div>
  </div>
  
  <div class="modal-footer">
    <button class="cancel-button" on:click={onCancel} disabled={isLoading}>Cancel</button>
    <button 
      class="create-button" 
      on:click={handleSubmit} 
      disabled={isLoading || selectedUsers.length === 0 || !chatName.trim()}
    >
      {isLoading ? 'Creating...' : 'Create Group'}
    </button>
  </div>
</div>

<style>
  .create-group-modal {
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
  
  .form-group {
    margin-bottom: 16px;
  }
  
  .form-group label {
    display: block;
    margin-bottom: 8px;
    font-weight: 500;
  }
  
  input[type="text"] {
    width: 100%;
    padding: 8px 12px;
    border: 1px solid #ced4da;
    border-radius: 4px;
    font-size: 1rem;
  }
  
  .error-message {
    background-color: #f8d7da;
    color: #721c24;
    padding: 10px;
    border-radius: 4px;
    margin-bottom: 16px;
  }
  
  .selected-users {
    margin-bottom: 16px;
  }
  
  .user-chips {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }
  
  .user-chip {
    display: flex;
    align-items: center;
    background-color: #e9ecef;
    border-radius: 16px;
    padding: 4px 12px;
    font-size: 0.875rem;
  }
  
  .user-chip button {
    background: none;
    border: none;
    font-size: 0.75rem;
    margin-left: 8px;
    cursor: pointer;
    color: #6c757d;
  }
  
  .user-list {
    border: 1px solid #e5e5e5;
    border-radius: 4px;
    max-height: 300px;
    overflow-y: auto;
  }
  
  .user-list h4 {
    margin: 0;
    padding: 12px;
    background-color: #f8f9fa;
    border-bottom: 1px solid #e5e5e5;
  }
  
  .user-item {
    display: flex;
    align-items: center;
    padding: 12px;
    cursor: pointer;
    border-bottom: 1px solid #e5e5e5;
  }
  
  .user-item:last-child {
    border-bottom: none;
  }
  
  .user-item:hover {
    background-color: #f8f9fa;
  }
  
  .user-item.selected {
    background-color: #e9ecef;
  }
  
  .user-avatar {
    width: 40px;
    height: 40px;
    border-radius: 50%;
    margin-right: 12px;
    overflow: hidden;
    flex-shrink: 0;
  }
  
  .user-avatar img {
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
  }
  
  .user-username {
    font-size: 0.875rem;
    color: #6c757d;
  }
  
  .selected-indicator {
    color: #007bff;
    font-weight: bold;
  }
  
  .empty-state {
    padding: 16px;
    text-align: center;
    color: #6c757d;
  }
  
  .modal-footer {
    display: flex;
    justify-content: flex-end;
    padding: 16px;
    border-top: 1px solid #e5e5e5;
    gap: 12px;
  }
  
  .cancel-button {
    padding: 8px 16px;
    border: 1px solid #ced4da;
    background-color: white;
    border-radius: 4px;
    cursor: pointer;
  }
  
  .create-button {
    padding: 8px 16px;
    border: none;
    background-color: #007bff;
    color: white;
    border-radius: 4px;
    cursor: pointer;
  }
  
  .create-button:disabled {
    background-color: #6c757d;
    cursor: not-allowed;
  }
</style> 