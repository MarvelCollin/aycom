<script lang="ts">
  import { onMount } from 'svelte';
  import { listChats } from '../api/chat';
  import ChatContainer from '../components/Chat/ChatContainer.svelte';
  import { createLoggerWithPrefix } from '../utils/logger';
  
  const logger = createLoggerWithPrefix('ChatPage');
  
  // Define chat interface
  interface Chat {
    id: string;
    name?: string;
    is_group_chat?: boolean;
    created_by?: string;
    last_message?: {
      content?: string;
    };
    participants?: Array<{
      id?: string;
      user_id?: string;
      username?: string;
      display_name?: string;
    }>;
    created_at?: string;
    updated_at?: string;
  }
  
  // State
  let chats: Chat[] = [];
  let selectedChatId: string | null = null;
  let isLoading = true;
  let error: string | null = null;
  
  // Helper function to get chat name or participants display
  function getChatDisplayName(chat: Chat): string {
    if (chat.name && chat.name.trim() !== '') {
      return chat.name;
    }
    
    // Get names of participants if available
    if (chat.participants && chat.participants.length > 0) {
      const currentUserId = getUserIdFromToken();
      return chat.participants
        .filter(p => (p.id !== currentUserId && p.user_id !== currentUserId))
        .map(p => p.display_name || p.username || 'Unknown User')
        .join(', ');
    }
    
    return 'Unnamed Chat';
  }
  
  // Get user ID from JWT token
  function getUserIdFromToken(): string {
    const token = localStorage.getItem('token');
    if (!token) return '';
    
    try {
      const payload = token.split('.')[1];
      const decoded = JSON.parse(atob(payload));
      return decoded.sub || '';
    } catch (e) {
      logger.error('Failed to decode JWT token', e);
      return '';
    }
  }
  
  // Load user chats when component mounts
  onMount(async () => {
    try {
      isLoading = true;
      logger.debug('Loading chat list');
      
      const response = await listChats();
      chats = response.chats || [];
      logger.debug('Loaded chats', { count: chats.length, chats });
      
      // Try to restore previously selected chat from localStorage
      const savedChatId = localStorage.getItem('selectedChatId');
      
      if (savedChatId && chats.some(chat => chat.id === savedChatId)) {
        // If the saved chat exists in the loaded chats, select it
        selectedChatId = savedChatId;
        logger.debug('Restored previously selected chat', { chatId: savedChatId });
      } else if (chats.length > 0) {
        // Otherwise select the first chat
        selectedChatId = chats[0].id;
        // Save the selection
        localStorage.setItem('selectedChatId', selectedChatId);
        logger.debug('Selected first chat', { chatId: selectedChatId });
      }
      
      isLoading = false;
    } catch (err: unknown) {
      logger.error('Error loading chats:', err);
      error = err instanceof Error ? err.message : 'Failed to load chats';
      isLoading = false;
    }
  });
  
  // Handle chat selection
  function selectChat(chatId: string) {
    selectedChatId = chatId;
    // Save the selection for persistence between refreshes
    localStorage.setItem('selectedChatId', chatId);
    logger.debug('Selected chat', { chatId });
  }
</script>

<div class="chat-page">
  <div class="sidebar">
    <h2>Chats</h2>
    
    {#if isLoading}
      <p class="loading">Loading chats...</p>
    {:else if error}
      <p class="error">{error}</p>
    {:else if chats.length === 0}
      <p class="empty-state">No chats found</p>
    {:else}
      <ul class="chat-list">
        {#each chats as chat (chat.id)}
          <li 
            class="chat-item {selectedChatId === chat.id ? 'selected' : ''}"
            on:click={() => selectChat(chat.id)}
          >
            <div class="chat-name">{getChatDisplayName(chat)}</div>
            <div class="chat-preview">
              {chat.last_message?.content || 'No messages yet'}
            </div>
          </li>
        {/each}
      </ul>
    {/if} 
  </div>
  
  <div class="chat-area">
    {#if selectedChatId}
      <ChatContainer chatId={selectedChatId} />
    {:else}
      <div class="empty-chat-state">
        <p>Select a chat to start messaging</p>
      </div>
    {/if}
  </div>
</div>

<style>
  .chat-page {
    display: flex;
    height: 100%;
    width: 100%;
  }
  
  .sidebar {
    width: 300px;
    border-right: 1px solid #dee2e6;
    overflow-y: auto;
    background-color: #f8f9fa;
    padding: 16px;
    display: flex;
    flex-direction: column;
  }
  
  .sidebar h2 {
    margin-top: 0;
    margin-bottom: 16px;
    font-size: 1.25rem;
  }
  
  .chat-list {
    list-style: none;
    padding: 0;
    margin: 0;
  }
  
  .chat-item {
    padding: 12px;
    border-radius: 8px;
    margin-bottom: 8px;
    cursor: pointer;
    transition: background-color 0.2s;
  }
  
  .chat-item:hover {
    background-color: #e9ecef;
  }
  
  .chat-item.selected {
    background-color: #e9ecef;
  }
  
  .chat-name {
    font-weight: 500;
    margin-bottom: 4px;
  }
  
  .chat-preview {
    font-size: 0.85rem;
    color: #6c757d;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  
  .chat-area {
    flex: 1;
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }
  
  .empty-chat-state {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 100%;
    color: #6c757d;
    font-size: 1.1rem;
  }
  
  .loading, .error, .empty-state {
    padding: 12px;
    text-align: center;
  }
  
  .error {
    color: #dc3545;
  }
</style> 