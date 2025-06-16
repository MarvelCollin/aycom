<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { chatMessageStore, getMessagesForChat, getTypingUsersForChat } from '../../stores/chatMessageStore';
  import { createLoggerWithPrefix } from '../../utils/logger';
  import { fade } from 'svelte/transition';
  import { listMessages, sendMessage as sendMessageApi } from '../../api';
  import type { MessageType } from '../../stores/websocketStore';

  const logger = createLoggerWithPrefix('ChatWindow');

  export let chatId: string;
  export let userId: string;
  export let participants: any[] = [];
  export let initialMessages: any[] = []; 

  let messageInput = '';
  let messageContainer: HTMLElement;
  let isTyping = false;
  let typingTimeout: number | null = null;
  let initialMessagesProcessed = false;
  let isLoadingMessages = false;
  let errorMessage = '';

  $: messages = getMessagesForChat(chatId);
  $: typingUsers = getTypingUsersForChat(chatId);

  async function loadMessagesHistory() {
    if (isLoadingMessages) return;

    try {
      isLoadingMessages = true;
      logger.debug('Loading message history for chat', { chatId });

      console.log(`[ChatWindow] Fetching messages for chat ID: ${chatId}`);

      const response = await listMessages(chatId);

      console.log('[ChatWindow] API response:', response);

      if (response && response.messages && Array.isArray(response.messages)) {
        logger.debug('Loaded messages from API', { count: response.messages.length });

        response.messages.forEach(msg => {

          const messageId = msg.message_id || msg.id || '';

          const messageUserId = msg.user_id || msg.sender_id || '';

          const processedMsg = {
            message_id: messageId,
            chat_id: msg.chat_id || chatId,
            content: msg.content,
            timestamp: msg.timestamp ? 
              (typeof msg.timestamp === 'number' ? 
                new Date(msg.timestamp * 1000) : 
                new Date(msg.timestamp)) : 
              new Date(),
            user_id: messageUserId,
            type: 'text' as MessageType,
            is_edited: msg.is_edited || false,
            is_deleted: msg.is_deleted || false,
            is_read: msg.is_read || false,
            user: msg.user || {
              id: messageUserId,
              username: '',
              name: getUserDisplayName(messageUserId)
            }
          };

          chatMessageStore.addMessage(processedMsg);
        });
      } else if (response && Array.isArray(response)) {

        logger.debug('Loaded messages from API (direct array)', { count: response.length });

        response.forEach(msg => {
          const messageId = msg.message_id || msg.id || '';
          const messageUserId = msg.user_id || msg.sender_id || '';

          const processedMsg = {
            message_id: messageId,
            chat_id: msg.chat_id || chatId,
            content: msg.content,
            timestamp: msg.timestamp ? 
              (typeof msg.timestamp === 'number' ? 
                new Date(msg.timestamp * 1000) : 
                new Date(msg.timestamp)) : 
              new Date(),
            user_id: messageUserId,
            type: 'text' as MessageType,
            is_edited: msg.is_edited || false,
            is_deleted: msg.is_deleted || false,
            is_read: msg.is_read || false,
            user: msg.user || {
              id: messageUserId,
              username: '',
              name: getUserDisplayName(messageUserId)
            }
          };

          chatMessageStore.addMessage(processedMsg);
        });
      } else {
        logger.warn('No messages returned from API or invalid response format', { response });
        console.warn('[ChatWindow] Invalid or empty messages response:', response);

        if (response) {
          console.log('[ChatWindow] Response keys:', Object.keys(response));
          console.log('[ChatWindow] Response type:', typeof response);

          if (response.messages) {
            console.log('[ChatWindow] Messages type:', typeof response.messages);
            console.log('[ChatWindow] Is array?', Array.isArray(response.messages));
          }
        }
      }
    } catch (error: any) {
      logger.error('Failed to load message history', error);

      if (error.message && error.message.includes('not a participant in this chat')) {
        errorMessage = 'You are not a participant in this chat. Please join the chat first.';
      } else {
      errorMessage = 'Failed to load messages. Please try refreshing.';
      }

      console.error('[ChatWindow] Error loading messages:', error?.message || error);

      if (error?.stack) {
        console.debug('[ChatWindow] Error stack:', error.stack);
      }
    } finally {
      isLoadingMessages = false;
    }
  }

  function formatTimestamp(timestamp: Date | string | undefined): string {
    if (!timestamp) return 'Just now';

    try {

      const date = typeof timestamp === 'string' ? new Date(timestamp) : timestamp;

      if (isNaN(date.getTime())) {
        logger.debug('Invalid timestamp detected', { timestamp });
        return 'Just now';
      }

      return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
    } catch (e) {
      logger.warn('Error formatting timestamp', e);
      return 'Just now';
    }
  }

  function getUserDisplayName(userId: string): string {

    const userMessage = $messages.find(m => m.user_id === userId);
    if (userMessage && userMessage.user && userMessage.user.name) {
      return userMessage.user.name;
    }

    const participant = participants.find(p => p.id === userId || p.user_id === userId);
    if (participant) {

      const displayName = participant.display_name || participant.name || participant.username;
      if (displayName) {
        return displayName;
      }
    }

    const shortId = userId.substring(0, 4);
    return `User ${shortId}`;
  }

  async function handleSendMessage() {
    if (!messageInput.trim()) return;

    const content = messageInput.trim();
    messageInput = ''; 

    clearTypingIndicator();

    try {
      logger.debug('Sending message to chat', { chatId });

      const tempId = `temp-${Date.now()}`;

      const messageData = {
        content: content,
        message_id: tempId 
      };

      try {
        const response = await sendMessageApi(chatId, messageData);
        logger.debug('Message sent successfully via API', { response });

        chatMessageStore.sendMessage(chatId, content, userId);

      } catch (apiError) {
        logger.error('Failed to send message via API, falling back to WebSocket only', apiError);

        chatMessageStore.sendMessage(chatId, content, userId);

        errorMessage = 'Network issue detected. Message may not be delivered.';
        setTimeout(() => {
          errorMessage = '';
        }, 3000);
      }
    } catch (error) {
      logger.error('Failed to send message', error);
      errorMessage = 'Failed to send message. Please try again.';
      setTimeout(() => {
        errorMessage = '';
      }, 3000);
    }
  }

  function handleKeydown(event: KeyboardEvent) {
    if (event.key === 'Enter' && !event.shiftKey) {
      event.preventDefault();
      handleSendMessage();
    } else {
      handleTyping();
    }
  }

  function handleTyping() {
    if (!isTyping) {
      isTyping = true;
      chatMessageStore.sendTypingIndicator(chatId, userId);
    }

    if (typingTimeout) {
      clearTimeout(typingTimeout);
    }

    typingTimeout = window.setTimeout(() => {
      isTyping = false;
    }, 2000);
  }

  function clearTypingIndicator() {
    isTyping = false;
    if (typingTimeout) {
      clearTimeout(typingTimeout);
      typingTimeout = null;
    }
  }

  $: typingIndicatorText = $typingUsers
    .filter(id => id !== userId)
    .map(getUserDisplayName)
    .join(', ');

  $: if ($messages && messageContainer) {

    setTimeout(() => {
      messageContainer.scrollTop = messageContainer.scrollHeight;
    }, 0);
  }

  onMount(() => {

    chatMessageStore.connectToChat(chatId);

    loadMessagesHistory();

    if (!initialMessagesProcessed && initialMessages.length > 0) {
      initialMessagesProcessed = true;

      initialMessages.forEach(message => {
        if (!message.is_read && message.sender_id !== userId && (message.id || message.message_id)) {
          chatMessageStore.sendReadReceipt(chatId, message.id || message.message_id, userId);
        }
      });
    } else {

      $messages.forEach(message => {
        if (!message.is_read && message.user_id !== userId && message.message_id) {
          chatMessageStore.sendReadReceipt(chatId, message.message_id, userId);
        }
      });
    }
  });

  onDestroy(() => {

    chatMessageStore.disconnectFromChat(chatId);
    clearTypingIndicator();
  });
</script>

<div class="chat-window">
  <div class="messages-container" bind:this={messageContainer}>
    {#if isLoadingMessages && $messages.length === 0}
      <div class="loading-messages">Loading messages...</div>
    {/if}

    {#each $messages as message (message.message_id)}
      <div 
        class="message {message.user_id === userId ? 'outgoing' : 'incoming'}"
        transition:fade={{ duration: 150 }}
      >
        {#if message.user_id !== userId}
          <div class="sender">{getUserDisplayName(message.user_id)}</div>
        {/if}
        <div class="bubble">
          {#if message.is_deleted}
            <span class="deleted-message">This message was deleted</span>
          {:else}
            {message.content}
          {/if}
        </div>
        <div class="metadata">
          <span class="timestamp">{formatTimestamp(message.timestamp)}</span>
          {#if message.is_edited}
            <span class="edited-indicator">(edited)</span>
          {/if}
          {#if message.is_read && message.user_id === userId}
            <span class="read-indicator">Read</span>
          {/if}
        </div>
      </div>
    {/each}

    {#if typingIndicatorText}
      <div class="typing-indicator" transition:fade={{ duration: 100 }}>
        {typingIndicatorText} {typingIndicatorText.includes(',') ? 'are' : 'is'} typing...
      </div>
    {/if}
  </div>

  <div class="input-container">
    <textarea 
      bind:value={messageInput} 
      on:keydown={handleKeydown}
      placeholder="Type a message..."
      rows="1"
    ></textarea>
    <button 
      class="send-button" 
      on:click={handleSendMessage}
      disabled={!messageInput.trim()}
    >
      Send
    </button>
  </div>
</div>

<style>
  .chat-window {
    display: flex;
    flex-direction: column;
    height: 100%;
  }

  .messages-container {
    flex: 1;
    overflow-y: auto;
    padding: 16px;
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .loading-messages {
    text-align: center;
    color: #777;
    padding: 20px;
    font-style: italic;
  }

  .message {
    max-width: 70%;
    margin-bottom: 8px;
  }

  .incoming {
    align-self: flex-start;
  }

  .outgoing {
    align-self: flex-end;
  }

  .sender {
    font-size: 0.8rem;
    margin-bottom: 2px;
    font-weight: 500;
    color: #555;
  }

  .bubble {
    padding: 10px 12px;
    border-radius: 12px;
    word-break: break-word;
  }

  .incoming .bubble {
    background-color: #f0f0f0;
    border-top-left-radius: 4px;
  }

  .outgoing .bubble {
    background-color: #0084ff;
    color: white;
    border-top-right-radius: 4px;
  }

  .metadata {
    font-size: 0.7rem;
    margin-top: 2px;
    display: flex;
    gap: 4px;
  }

  .incoming .metadata {
    justify-content: flex-start;
    color: #777;
  }

  .outgoing .metadata {
    justify-content: flex-end;
    color: #777;
  }

  .timestamp {
    color: #777;
  }

  .edited-indicator, .read-indicator {
    opacity: 0.8;
  }

  .deleted-message {
    font-style: italic;
    opacity: 0.7;
  }

  .typing-indicator {
    align-self: flex-start;
    font-size: 0.8rem;
    color: #666;
    font-style: italic;
    margin-top: 4px;
  }

  .input-container {
    display: flex;
    padding: 12px;
    border-top: 1px solid #e0e0e0;
    background: white;
  }

  textarea {
    flex: 1;
    padding: 10px 12px;
    border: 1px solid #ccc;
    border-radius: 18px;
    resize: none;
    font-family: inherit;
    font-size: inherit;
    outline: none;
  }

  .send-button {
    margin-left: 8px;
    border: none;
    background: #0084ff;
    color: white;
    border-radius: 50%;
    width: 36px;
    height: 36px;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: background-color 0.2s;
  }

  .send-button:disabled {
    background-color: #cccccc;
    cursor: not-allowed;
  }

  .send-button:hover:not(:disabled) {
    background-color: #0076e4;
  }
</style>