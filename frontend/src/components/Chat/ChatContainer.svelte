<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import ChatWindow from './ChatWindow.svelte';
  import { listChatParticipants, listMessages, joinChat } from '../../api';
  import { getAuthToken } from '../../utils/auth';
  import { createLoggerWithPrefix } from '../../utils/logger';
  import { chatMessageStore } from '../../stores/chatMessageStore';
  import type { MessageType } from '../../stores/websocketStore';

  const logger = createLoggerWithPrefix('ChatContainer');

  export let chatId: string;

  type BasicUser = {
    id?: string;
    user_id?: string;
    username?: string;
    display_name?: string;
    avatar_url?: string;
    name?: string;
    profile_picture_url?: string;
  };

  type ChatMessage = {
    id?: string;
    message_id?: string;
    sender_id?: string;
    user_id?: string;
    chat_id?: string;
    content: string;
    timestamp: number | string;
    is_read?: boolean;
    is_deleted?: boolean;
    is_edited?: boolean;
    user?: BasicUser;
  };

  let participants: BasicUser[] = [];
  let userId = '';
  let isLoading = true;
  let error: string | null = null;
  let messages: ChatMessage[] = [];
  let isParticipantError = false;
  let isJoiningChat = false;

  async function handleJoinChat() {
    try {
      isJoiningChat = true;
      error = null;

      await joinChat(chatId);

      window.location.reload();
    } catch (err: any) {
      logger.error('Failed to join chat:', err);
      error = err.message || 'Failed to join chat. Please try again.';
      isJoiningChat = false;
    }
  }

  onMount(async () => {

    const token = getAuthToken();
    if (token) {
      try {

        const base64Url = token.split('.')[1];
        const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
        const jsonPayload = decodeURIComponent(atob(base64).split('').map(c => {
          return '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2);
        }).join(''));

        const tokenData = JSON.parse(jsonPayload);
        userId = tokenData.user_id || tokenData.sub || '';
        console.log('[ChatContainer] Extracted user ID from token:', userId);
      } catch (e) {
        logger.error('Failed to extract userID from token:', e);
        console.error('[ChatContainer] Auth token parsing error:', e);
      }
    } else {
      console.warn('[ChatContainer] No authentication token found');
    }

    try {
      isLoading = true;
      console.log('[ChatContainer] Loading chat data for chat ID:', chatId);

      try {
        const participantsResponse = await listChatParticipants(chatId);
        console.log('[ChatContainer] Participants response:', participantsResponse);

        if (participantsResponse && participantsResponse.participants) {
          participants = participantsResponse.participants.map((p: any) => ({
            id: p.user_id || p.id,
            username: p.username || '',
            display_name: p.display_name || p.name || p.username || 'Unknown User',
            avatar_url: p.avatar_url || p.profile_picture_url || null
          }));
          console.log('[ChatContainer] Processed participants:', participants);
        } else if (participantsResponse && Array.isArray(participantsResponse)) {

          participants = participantsResponse.map((p: any) => ({
            id: p.user_id || p.id,
            username: p.username || '',
            display_name: p.display_name || p.name || p.username || 'Unknown User',
            avatar_url: p.avatar_url || p.profile_picture_url || null
          }));
        }
      } catch (participantsError: any) {
        console.error('[ChatContainer] Failed to load participants:', participantsError);

        if (participantsError.message && participantsError.message.includes('not a participant in this chat')) {
          isParticipantError = true;
          error = 'You are not a participant in this chat. Please join the chat first.';
        }

      }

      if (!isParticipantError) {

        try {
          const messagesResponse = await listMessages(chatId);
          console.log('[ChatContainer] Messages response:', messagesResponse);

          if (messagesResponse && messagesResponse.messages) {
            messages = messagesResponse.messages;
          } else if (messagesResponse && Array.isArray(messagesResponse)) {

            messages = messagesResponse;
          } else {
            console.warn('[ChatContainer] Invalid or empty messages response');
            messages = []; 
          }
        } catch (messagesError: any) {
          console.error('[ChatContainer] Failed to load messages:', messagesError);

          if (messagesError.message && messagesError.message.includes('not a participant in this chat')) {
            isParticipantError = true;
            error = 'You are not a participant in this chat. Please join the chat first.';
          } else {
            error = messagesError instanceof Error ? messagesError.message : 'Failed to load messages';
          }

          messages = []; 
        }
      }

      if (!isParticipantError && messages.length > 0) {
        messages.forEach(message => {

          let sender = message.user;
          const senderId = message.user_id || message.sender_id || '';

          if (!sender && participants.length > 0) {
            sender = participants.find(p => p.id === senderId || p.user_id === senderId);
          }

          const messageId = message.message_id || message.id || '';

          let messageTimestamp;
          if (typeof message.timestamp === 'number') {
            messageTimestamp = new Date(message.timestamp * 1000);
          } else if (typeof message.timestamp === 'string') {
            messageTimestamp = new Date(message.timestamp);
          } else {
            messageTimestamp = new Date();
          }

          const storeMessage = {
            type: 'text' as MessageType,
            content: message.content,
            user_id: senderId,
            chat_id: chatId,
            message_id: messageId,
            timestamp: messageTimestamp,
            is_read: message.is_read || false,
            is_deleted: message.is_deleted || false,
            is_edited: message.is_edited || false,
            user: {
              id: senderId || 'unknown',
              username: sender?.username || 'Unknown',
              name: sender?.display_name || sender?.name || 'Unknown User',
              profile_picture_url: sender?.avatar_url || sender?.profile_picture_url
            }
          };

          chatMessageStore.addMessage(storeMessage);
        });
      }

      if (!isParticipantError) {
        chatMessageStore.connectToChat(chatId);
      }

      isLoading = false;
    } catch (err: unknown) {
      logger.error('Error loading chat data:', err);
      console.error('[ChatContainer] Error details:', err instanceof Error ? err.message : err);

      if (err instanceof Error && err.message.includes('not a participant in this chat')) {
        isParticipantError = true;
        error = 'You are not a participant in this chat. Please join the chat first.';
      } else {
        error = err instanceof Error ? err.message : 'Failed to load chat';
      }

      isLoading = false;

      if (err instanceof Error && (
          err.message.includes('authentication') || 
          err.message.includes('UNAUTHORIZED')
      )) {
        console.error('[ChatContainer] Authentication error detected');
        error = 'Authentication error. Please try logging in again.';
      }
    }
  });

  onDestroy(() => {
    if (chatId && !isParticipantError) {
      chatMessageStore.disconnectFromChat(chatId);
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
      {#if isParticipantError}
        <div class="action-buttons">
          <button on:click={handleJoinChat} disabled={isJoiningChat}>
            {isJoiningChat ? 'Joining...' : 'Join Chat'}
          </button>
          <button on:click={() => window.history.back()}>Go Back</button>
        </div>
      {:else}
        <button on:click={() => window.location.reload()}>Retry</button>
      {/if}
    </div>
  {:else if !userId}
    <div class="error-state">
      <p>You need to be logged in to view this chat</p>
    </div>
  {:else}
    <div class="chat-header">
      {#if participants.length > 0}
        <div class="participant-info">
          {#if chatId.startsWith('group_') || participants.length > 2}
            <h3>Group Chat</h3>
          {:else}
            <h3>Chat with: {participants.filter(p => p.user_id !== userId && p.id !== userId)
              .map(p => p.display_name || p.username || 'Unknown User')
              .join(', ') || 'Unknown User'}</h3>
          {/if}
        </div>
      {/if}
    </div>
    <ChatWindow 
      {chatId}
      {userId}
      {participants}
      initialMessages={messages}
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

  .chat-header {
    padding: 12px 16px;
    border-bottom: 1px solid #dee2e6;
    background-color: #f8f9fa;
  }

  .participant-info h3 {
    margin: 0;
    font-size: 1.1rem;
    font-weight: 500;
    color: #333;
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

  .error-state button:disabled {
    background-color: #6c757d;
    cursor: not-allowed;
  }

  .action-buttons {
    display: flex;
    gap: 10px;
    margin-top: 12px;
  }
</style>