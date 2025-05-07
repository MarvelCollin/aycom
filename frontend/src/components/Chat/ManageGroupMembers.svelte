<script lang="ts">
  import { onMount } from 'svelte';
  import { addChatParticipant, removeChatParticipant, listChatParticipants } from '../../api/chat';
  import { createLoggerWithPrefix } from '../../utils/logger';
  
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
      const response = await fetch('/api/v1/users/suggestions', {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        }
      });
      
      if (!response.ok) {
        throw new Error('Failed to load available users');
      }
      
      const data = await response.json();
      // Filter out users who are already in the chat
      const participants = currentParticipants.map(p => p.id);
      availableUsers = (data.users || []).filter((user: User) => 
        !participants.includes(user.id)
      );
    } catch (error) {
      logger.error('Error loading available users:', error);
      throw new Error('Failed to load available users');
    }
  }
  
  // Filtered users based on search query
  $: filteredUsers = searchQuery 
    ? availableUsers.filter(user => 
        user.username.toLowerCase().includes(searchQuery.toLowerCase()) || 
        (user.display_name && user.display_name.toLowerCase().includes(searchQuery.toLowerCase()))
      )
    : availableUsers;
  
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
        <input 
          type="text" 
          bind:value={searchQuery} 
          placeholder="Search users..."
          class="search-input"
        />
        
        {#if filteredUsers.length === 0}
          <p class="empty-state">No users found</p>
        {:else}
          <div class="user-list">
            {#each filteredUsers as user (user.id)}
              <div class="user-item">
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
                
                <button 
                  class="add-button" 
                  on:click={() => handleAddMember(user)}
                  disabled={isAddingMember}
                >
                  Add
                </button>
              </div>
            {/each}
          </div>
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
  
  .search-input {
    width: 100%;
    padding: 8px 12px;
    border: 1px solid #ced4da;
    border-radius: 4px;
    font-size: 1rem;
    margin-bottom: 12px;
  }
  
  .member-list, .user-list {
    border: 1px solid #e5e5e5;
    border-radius: 4px;
    max-height: 300px;
    overflow-y: auto;
  }
  
  .member-item, .user-item {
    display: flex;
    align-items: center;
    padding: 12px;
    border-bottom: 1px solid #e5e5e5;
  }
  
  .member-item:last-child, .user-item:last-child {
    border-bottom: none;
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