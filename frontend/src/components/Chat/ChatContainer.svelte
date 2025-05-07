<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import ChatWindow from './ChatWindow.svelte';
  import { listChatParticipants } from '../../api/chat';
  import { getAuthToken } from '../../utils/auth';
  import { createLoggerWithPrefix } from '../../utils/logger';
  
  const logger = createLoggerWithPrefix('ChatContainer');
  
  // Props
  export let chatId: string;
  
  // State
  let participants = [];
  let userId = '';
  let isLoading = true;
  let error: string | null = null;
  
  // Load chat participants when component mounts
  onMount(async () => {
    try {
      // Get current user ID from auth token
      const token = getAuthToken();
      if (!token) {
        throw new Error('Not authenticated');
      }
      
      // Parse JWT to get user ID
      try {
        const payload = token.split('.')[1];
        const decoded = JSON.parse(atob(payload));
        userId = decoded.sub || '';
      } catch (e) {
        logger.error('Failed to decode JWT token', e);
        throw new Error('Failed to get user ID');
      }
      
      // Fetch participant data
      const response = await listChatParticipants(chatId);
      participants = response.participants || [];
      
      isLoading = false;
    } catch (err: unknown) {
      logger.error('Error loading chat participants:', err);
      error = err instanceof Error ? err.message : 'Failed to load chat';
      isLoading = false;
    }
  });
</script>

<div class="chat-container">
  {#if isLoading}
    <div class="loading-state">
      <p>Loading chat...</p>
    </div>
  {:else if error}
    <div class="error-state">
      <p>Error: {error}</p>
      <button on:click={() => window.location.reload()}>Retry</button>
    </div>
  {:else if !userId}
    <div class="error-state">
      <p>You need to be logged in to view this chat</p>
    </div>
  {:else}
    <ChatWindow 
      {chatId}
      {userId}
      {participants}
    />
  {/if}
</div>

<style>
  .chat-container {
    width: 100%;
    height: 100%;
    display: flex;
    flex-direction: column;
  }
  
  .loading-state,
  .error-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 100%;
    text-align: center;
    padding: 20px;
  }
  
  .error-state {
    color: #dc3545;
  }
  
  .error-state button {
    margin-top: 12px;
    padding: 8px 16px;
    background-color: #007bff;
    color: white;
    border: none;
    border-radius: 4px;
    cursor: pointer;
  }
</style> 