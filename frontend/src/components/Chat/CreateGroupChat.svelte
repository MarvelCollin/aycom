<script lang="ts">
  import { createEventDispatcher, onMount } from 'svelte';
  import { useAuth } from '../../hooks/useAuth';
  import * as api from '../../api';
  import { createLoggerWithPrefix } from '../../utils/logger';
  import { toastStore } from '../../stores/toastStore';
  import { handleApiError } from '../../utils/common';
  import { transformApiUsers, type StandardUser } from '../../utils/userTransform';
  import type { User, ApiUserResponse, CreateChatResponse } from '../../interfaces/IChat';

  const { searchUsers, getAllUsers, getUserById, createChat } = api;

  const logger = createLoggerWithPrefix('CreateGroupChat');
  const dispatch = createEventDispatcher();
  const { getAuthState } = useAuth();

  export let onSuccess: ((event: { detail: { chat: any } }) => void) | undefined = undefined;  
  export let onCancel: (() => void) | undefined = undefined;   

  let authState = getAuthState ? getAuthState() : { user_id: null };
  let groupName = '';
  let searchQuery = '';
  let searchResults: StandardUser[] = [];
  let selectedParticipants: StandardUser[] = [];
  let isLoading = false;
  let searchTimeout: ReturnType<typeof setTimeout> | null = null;
  let errorMessage = '';
  let groupNameInput: HTMLInputElement;
  let searchInput: HTMLInputElement;
  let selectedIndex = -1;

  onMount(() => {
    if (groupNameInput) {
      groupNameInput.focus();
    }
  });

  function getAvatarColor(name: string | undefined): string {
    const colors = [
      '#4F46E5', 
      '#0EA5E9', 
      '#10B981', 
      '#F59E0B', 
      '#EF4444', 
      '#8B5CF6', 
      '#EC4899', 
      '#06B6D4', 
    ];

    let hash = 0;
    if (!name) name = 'User';
    for (let i = 0; i < name.length; i++) {
      hash = name.charCodeAt(i) + ((hash << 5) - hash);
    }

    hash = Math.abs(hash);
    const index = hash % colors.length;

    return colors[index];
  }

  async function handleSearch(): Promise<void> {
    if (!searchQuery.trim()) {
      searchResults = [];
      return;
    }

    selectedIndex = -1;

    if (searchTimeout) clearTimeout(searchTimeout);

    searchTimeout = setTimeout(async () => {
      try {
        logger.debug('Searching for users:', { query: searchQuery });
        isLoading = true;
        errorMessage = '';
        const query = searchQuery.toLowerCase();

        const response = await searchUsers(searchQuery);

        logger.debug('Search users API response:', { 
          status: 'success',
          userCount: response.users?.length || 0
        });

        const users = response?.users || [];

        if (users && users.length > 0) {

          const transformedUsers = transformApiUsers(users);
          searchResults = transformedUsers.filter(user => 
            user.id !== authState.user_id && 
            !selectedParticipants.some(p => p.id === user.id)
          );
          logger.info('Retrieved users from API', { count: searchResults.length });        } else {
          logger.warn('No users found from API search');
          searchResults = [];
        }      } catch (error) {
        logger.error('Error searching users:', error);
        errorMessage = 'Failed to search for users. Please try again.';
        searchResults = [];
      } finally {
        isLoading = false;
      }
    }, 300); 
  }

  function handleSearchKeydown(event: KeyboardEvent): void {
    if (!searchResults.length) return;

    if (event.key === 'ArrowDown') {
      event.preventDefault();
      selectedIndex = Math.min(selectedIndex + 1, searchResults.length - 1);
    } else if (event.key === 'ArrowUp') {
      event.preventDefault();
      selectedIndex = Math.max(selectedIndex - 1, -1);
    } else if (event.key === 'Enter' && selectedIndex >= 0) {
      event.preventDefault();
      addParticipant(searchResults[selectedIndex]);
    }
  }

  function addParticipant(user: StandardUser): void {
    if (selectedParticipants.some(p => p.id === user.id)) {
      return;
    }

    selectedParticipants = [...selectedParticipants, user];
    searchResults = searchResults.filter(u => u.id !== user.id);
    logger.debug('Added participant', { userId: user.id, username: user.username });
    searchQuery = '';
    selectedIndex = -1;

    setTimeout(() => {
      if (searchInput) searchInput.focus();
    }, 0);
  }

  function removeParticipant(userId: string): void {
    selectedParticipants = selectedParticipants.filter(p => p.id !== userId);
    logger.debug('Removed participant:', { userId });
  }

  async function createGroupChat(): Promise<void> {
    if (!groupName.trim()) {
      errorMessage = 'Please enter a group name';
      groupNameInput.focus();
      return;
    }

    if (selectedParticipants.length === 0) {
      errorMessage = 'Please select at least one participant';
      searchInput.focus();
      return;
    }

    try {
      isLoading = true;
      errorMessage = '';

      const participantIds = selectedParticipants.map(p => p.id);
      const chatData = {
        name: groupName.trim(),
        type: 'group',
        participants: participantIds
      };

      logger.debug('Creating group chat:', { 
        name: groupName, 
        participantsCount: participantIds.length 
      });

      const response: CreateChatResponse = await createChat(chatData);

      if (response && response.chat) {
        logger.debug('Group chat created:', { chatId: response.chat.id });

        const fullParticipants = [...selectedParticipants];

        if (onSuccess) {
          onSuccess({ detail: { 
            chat: {
              ...response.chat,
              participants: fullParticipants
            }
          }});
        } else {
          dispatch('success', { 
            chat: {
              ...response.chat,
              participants: fullParticipants
            }
          });
        }

        toastStore.showToast('Group chat created successfully', 'success');
      } else {
        throw new Error('Invalid response from server');
      }
    } catch (error) {
      const errorDetail = handleApiError(error);
      errorMessage = 'Failed to create group chat. Please try again.';
      logger.error('Failed to create group chat:', errorDetail);
      toastStore.showToast('Failed to create group chat', 'error');
    } finally {
      isLoading = false;
    }
  }

  function handleInput(): void {

    if (errorMessage) errorMessage = '';

    handleSearch();
  }

  function cancel(): void {
    if (onCancel) {
      onCancel();
    } else {
      dispatch('cancel');
    }
  }
</script>

<div class="group-chat-modal">
  <div class="modal-header">
    <div class="header-content">
      <h2>Create a Group Chat</h2>
      {#if selectedParticipants.length > 0}
        <div class="member-preview">
          <div class="avatar-group">
            {#each selectedParticipants.slice(0, 3) as participant, i}
              <div 
                class="avatar-mini preview" 
                style="
                  background-color: {getAvatarColor(participant.displayName || participant.username)};
                  margin-right: -8px;
                  z-index: {3 - i};
                "
              >
                {#if participant.avatar}
                  <img src={participant.avatar} alt={participant.displayName || participant.username} />
                {:else}
                  <span>{(participant.displayName || participant.username || 'User').substring(0, 1).toUpperCase()}</span>
                {/if}
              </div>
            {/each}
            {#if selectedParticipants.length > 3}
              <div class="avatar-mini preview more-members">
                <span>+{selectedParticipants.length - 3}</span>
              </div>
            {/if}
          </div>
          <span class="member-count">{selectedParticipants.length} {selectedParticipants.length === 1 ? 'member' : 'members'}</span>
        </div>
      {/if}
    </div>
    <button class="close-button" on:click={cancel} aria-label="Close">
      <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" width="24" height="24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
      </svg>
    </button>
  </div>

  <div class="modal-body">
    {#if errorMessage}
      <div class="error-message">
        {errorMessage}
      </div>
    {/if}

    <div class="input-group">
      <label for="groupName">Group Name</label>
      <input 
        type="text" 
        id="groupName" 
        placeholder="Enter group name" 
        bind:value={groupName}
        bind:this={groupNameInput}
        on:input={() => errorMessage = ''}
        on:keydown={(e) => e.key === 'Enter' && searchInput.focus()}
      />
    </div>

    <div class="input-group">
      <label for="searchUsers">Add Participants</label>
      <input 
        type="text" 
        id="searchUsers" 
        placeholder="Search users by name or username" 
        bind:value={searchQuery}
        bind:this={searchInput}
        on:input={handleInput}
        on:keydown={handleSearchKeydown}
      />
    </div>

    {#if isLoading}
      <div class="loading">
        <span class="loading-spinner"></span>
        Searching users...
      </div>
    {:else if searchQuery && searchResults.length === 0}
      <div class="no-results">
        No users found matching "{searchQuery}"
      </div>
    {/if}

    {#if searchResults.length > 0}
      <div class="search-results">
        <h3>Search Results</h3>
        <ul>
          {#each searchResults as user, i}
            <li>
              <div 
                class="user-item {i === selectedIndex ? 'selected' : ''}" 
                on:click={() => addParticipant(user)}
                on:keydown={(e) => e.key === 'Enter' && addParticipant(user)}
                role="button"
                tabindex="0"
              >
                <div class="avatar" style="background-color: {getAvatarColor(user.displayName || user.username)}">
                  {#if user.avatar}
                    <img src={user.avatar} alt={user.displayName || user.username} />
                  {:else}
                    <span>{(user.displayName || user.username || 'User').substring(0, 1).toUpperCase()}</span>
                  {/if}
                </div>
                <div class="user-info">
                  <span class="display-name">{user.displayName || user.username}</span>
                  {#if user.username && user.username !== user.displayName}
                    <span class="username">@{user.username}</span>
                  {/if}
                </div>
                <div class="add-button" aria-label="Add user">
                  <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" width="20" height="20">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
                  </svg>
                </div>
              </div>
            </li>
          {/each}
        </ul>
      </div>
    {/if}

    <div class="selected-participants">
      <h3>Selected Participants ({selectedParticipants.length})</h3>
      {#if selectedParticipants.length === 0}
        <div class="no-selections">
          No participants selected yet. Search and add users above.
        </div>
      {:else}
        <div class="participant-chips">
          {#each selectedParticipants as participant}
            <div class="participant-chip">
              <div class="avatar-mini" style="background-color: {getAvatarColor(participant.displayName || participant.username)}">
                {#if participant.avatar}
                  <img src={participant.avatar} alt={participant.displayName || participant.username} />
                {:else}
                  <span>{(participant.displayName || participant.username || 'User').substring(0, 1).toUpperCase()}</span>
                {/if}
              </div>
              <span>{participant.displayName || participant.username}</span>
              <button 
                class="remove-button" 
                on:click={() => removeParticipant(participant.id)}
                aria-label="Remove participant"
              >
                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" width="16" height="16">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>
          {/each}
        </div>
      {/if}
    </div>
  </div>

  <div class="modal-footer">
    <button class="cancel-button" on:click={cancel} disabled={isLoading}>
      Cancel
    </button>
    <button 
      class="create-button" 
      on:click={createGroupChat} 
      disabled={isLoading || selectedParticipants.length === 0 || !groupName.trim()}
    >
      {isLoading ? 'Creating...' : 'Create Group Chat'}
    </button>
  </div>
</div>

<style>
  .group-chat-modal {
    background-color: var(--background-color, white);
    color: var(--text-color, black);
    border-radius: 12px;
    width: 100%;
    max-width: 600px;
    max-height: 90vh;
    display: flex;
    flex-direction: column;
  }

  :global(.dark) .group-chat-modal {
    --background-color: #1f2937;
    --text-color: #f3f4f6;
    --border-color: #374151;
    --input-bg: #111827;
    --button-bg: #3b82f6;
    --button-hover: #2563eb;
    --button-text: white;
    --error-bg: rgba(239, 68, 68, 0.2);
    --error-text: #ef4444;
  }

  .group-chat-modal {
    --border-color: #e5e7eb;
    --input-bg: #f9fafb;
    --button-bg: #3b82f6;
    --button-hover: #2563eb;
    --button-text: white;
    --error-bg: rgba(239, 68, 68, 0.1);
    --error-text: #ef4444;
  }

  .user-item.selected {
    background-color: rgba(59, 130, 246, 0.1);
  }

  :global(.dark) .user-item.selected {
    background-color: rgba(59, 130, 246, 0.2);
  }

  .modal-header {
    padding: 16px;
    border-bottom: 1px solid var(--border-color);
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
  }

  .header-content {
    display: flex;
    flex-direction: column;
  }

  .modal-header h2 {
    margin: 0;
    font-size: 1.25rem;
    font-weight: 600;
  }

  .member-preview {
    display: flex;
    align-items: center;
    margin-top: 8px;
  }

  .avatar-group {
    display: flex;
    align-items: center;
    margin-right: 8px;
  }

  .avatar-mini.preview {
    width: 24px;
    height: 24px;
    font-size: 0.7rem;
    border: 2px solid var(--background-color);
    box-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
    margin-right: -8px;
    color: white;
    font-weight: 600;
  }

  .more-members {
    background-color: var(--button-bg);
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 0.7rem;
  }

  .member-count {
    font-size: 0.85rem;
    color: var(--text-secondary, gray);
    margin-left: 12px;
  }

  .close-button {
    background: none;
    border: none;
    cursor: pointer;
    color: var(--text-color);
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 50%;
    width: 36px;
    height: 36px;
    transition: background-color 0.2s;
  }

  .close-button:hover {
    background-color: rgba(0, 0, 0, 0.05);
  }

  :global(.dark) .close-button:hover {
    background-color: rgba(255, 255, 255, 0.1);
  }

  .modal-body {
    padding: 16px;
    overflow-y: auto;
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .error-message {
    padding: 10px;
    background-color: var(--error-bg);
    color: var(--error-text);
    border-radius: 6px;
    font-size: 0.875rem;
  }

  .input-group {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .input-group label {
    font-size: 0.875rem;
    font-weight: 500;
  }

  .input-group input {
    padding: 10px 12px;
    border: 1px solid var(--border-color);
    border-radius: 6px;
    background-color: var(--input-bg);
    color: var(--text-color);
    font-size: 0.95rem;
  }

  .input-group input:focus {
    outline: none;
    border-color: var(--button-bg);
    box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.15);
  }

  .loading {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 0.875rem;
    color: var(--text-secondary, gray);
    padding: 8px 0;
  }

  .loading-spinner {
    width: 16px;
    height: 16px;
    border: 2px solid rgba(0, 0, 0, 0.1);
    border-top-color: var(--button-bg);
    border-radius: 50%;
    animation: spinner 0.8s linear infinite;
  }

  @keyframes spinner {
    to {
      transform: rotate(360deg);
    }
  }

  :global(.dark) .loading-spinner {
    border-color: rgba(255, 255, 255, 0.1);
    border-top-color: var(--button-bg);
  }

  .no-results {
    padding: 8px 0;
    font-size: 0.875rem;
    color: var(--text-secondary, gray);
  }

  .search-results {
    margin-top: 8px;
  }

  .search-results h3, .selected-participants h3 {
    font-size: 1rem;
    font-weight: 600;
    margin: 0 0 8px 0;
  }

  .search-results ul {
    list-style: none;
    padding: 0;
    margin: 0;
    max-height: 200px;
    overflow-y: auto;
    border: 1px solid var(--border-color);
    border-radius: 6px;
  }

  .user-item {
    display: flex;
    align-items: center;
    padding: 10px 12px;
    width: 100%;
    text-align: left;
    cursor: pointer;
    color: var(--text-color);
    transition: background-color 0.2s;
    border-bottom: 1px solid var(--border-color);
  }

  .user-item:last-child {
    border-bottom: none;
  }

  .user-item:hover {
    background-color: rgba(0, 0, 0, 0.05);
  }

  :global(.dark) .user-item:hover {
    background-color: rgba(255, 255, 255, 0.05);
  }

  .avatar, .avatar-mini {
    width: 40px;
    height: 40px;
    border-radius: 50%;
    overflow: hidden;
    display: flex;
    align-items: center;
    justify-content: center;
    margin-right: 12px;
    color: white;
    font-weight: 600;
  }

  .avatar-mini {
    width: 24px;
    height: 24px;
    font-size: 0.75rem;
    margin-right: 8px;
  }

  .avatar img, .avatar-mini img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  .user-info {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
  }

  .display-name {
    font-weight: 500;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .username {
    font-size: 0.75rem;
    color: var(--text-secondary, gray);
  }

  .add-button {
    color: var(--button-bg);
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 4px;
    border-radius: 50%;
    transition: background-color 0.2s;
  }

  .user-item:hover .add-button {
    background-color: rgba(59, 130, 246, 0.1);
  }

  .selected-participants {
    margin-top: 16px;
  }

  .no-selections {
    color: var(--text-secondary, gray);
    font-size: 0.875rem;
    font-style: italic;
    padding: 8px 0;
  }

  .participant-chips {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    padding: 8px 0;
  }

  .participant-chip {
    display: flex;
    align-items: center;
    background-color: rgba(59, 130, 246, 0.1);
    border-radius: 20px;
    padding: 4px 8px;
    font-size: 0.875rem;
  }

  .remove-button {
    background: none;
    border: none;
    color: var(--text-color);
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    margin-left: 8px;
    padding: 2px;
    border-radius: 50%;
    transition: background-color 0.2s;
  }

  .remove-button:hover {
    background-color: rgba(0, 0, 0, 0.1);
  }

  :global(.dark) .remove-button:hover {
    background-color: rgba(255, 255, 255, 0.1);
  }

  .modal-footer {
    padding: 16px;
    border-top: 1px solid var(--border-color);
    display: flex;
    justify-content: flex-end;
    gap: 12px;
  }

  .cancel-button {
    padding: 8px 16px;
    border: 1px solid var(--border-color);
    background-color: transparent;
    color: var(--text-color);
    border-radius: 6px;
    font-weight: 500;
    cursor: pointer;
    transition: background-color 0.2s;
  }

  .cancel-button:hover:not(:disabled) {
    background-color: rgba(0, 0, 0, 0.05);
  }

  :global(.dark) .cancel-button:hover:not(:disabled) {
    background-color: rgba(255, 255, 255, 0.05);
  }

  .create-button {
    padding: 8px 16px;
    border: none;
    background-color: var(--button-bg);
    color: var(--button-text);
    border-radius: 6px;
    font-weight: 500;
    cursor: pointer;
    transition: background-color 0.2s;
  }

  .create-button:hover:not(:disabled) {
    background-color: var(--button-hover);
  }

  .create-button:disabled, .cancel-button:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }
</style>