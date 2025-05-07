<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { chatMessageStore, getMessagesForChat, getTypingUsersForChat } from '../../stores/chatMessageStore';
  import { createLoggerWithPrefix } from '../../utils/logger';
  import { fade } from 'svelte/transition';
  
  const logger = createLoggerWithPrefix('ChatWindow');
  
  // Props
  export let chatId: string;
  export let userId: string;
  export let participants: any[] = [];
  
  // Local state
  let messageInput = '';
  let messageContainer: HTMLElement;
  let isTyping = false;
  let typingTimeout: number | null = null;
  
  // Reactive stores
  $: messages = getMessagesForChat(chatId);
  $: typingUsers = getTypingUsersForChat(chatId);
  
  // Convert timestamp to readable format
  function formatTimestamp(timestamp: Date | string | undefined): string {
    if (!timestamp) return '';
    
    const date = typeof timestamp === 'string' ? new Date(timestamp) : timestamp;
    return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  }
  
  // Get user display name from participants array
  function getUserDisplayName(userId: string): string {
    const participant = participants.find(p => p.id === userId);
    return participant ? participant.display_name || participant.username : 'Unknown User';
  }
  
  // Handle sending a message
  function handleSendMessage() {
    if (!messageInput.trim()) return;
    
    chatMessageStore.sendMessage(chatId, messageInput, userId);
    messageInput = '';
    
    // Clear typing indicator when sending a message
    clearTypingIndicator();
  }
  
  // Handle keydown in the input field
  function handleKeydown(event: KeyboardEvent) {
    if (event.key === 'Enter' && !event.shiftKey) {
      event.preventDefault();
      handleSendMessage();
    } else {
      handleTyping();
    }
  }
  
  // Handle typing indicator
  function handleTyping() {
    if (!isTyping) {
      isTyping = true;
      chatMessageStore.sendTypingIndicator(chatId, userId);
    }
    
    // Clear any existing timeout
    if (typingTimeout) {
      clearTimeout(typingTimeout);
    }
    
    // Set a new timeout to clear the typing state
    typingTimeout = window.setTimeout(() => {
      isTyping = false;
    }, 2000);
  }
  
  // Clear typing indicator
  function clearTypingIndicator() {
    isTyping = false;
    if (typingTimeout) {
      clearTimeout(typingTimeout);
      typingTimeout = null;
    }
  }
  
  // Get typing indicator text
  $: typingIndicatorText = $typingUsers
    .filter(id => id !== userId)
    .map(getUserDisplayName)
    .join(', ');
  
  // Scroll to bottom when new messages arrive
  $: if ($messages && messageContainer) {
    // Schedule the scroll for the next tick to ensure the DOM is updated
    setTimeout(() => {
      messageContainer.scrollTop = messageContainer.scrollHeight;
    }, 0);
  }
  
  // Connect to chat when component mounts
  onMount(() => {
    chatMessageStore.connectToChat(chatId);
    
    // Mark all messages as read when opening the chat
    $messages.forEach(message => {
      if (!message.is_read && message.user_id !== userId && message.message_id) {
        chatMessageStore.sendReadReceipt(chatId, message.message_id, userId);
      }
    });
  });
  
  // Disconnect and clean up when component is destroyed
  onDestroy(() => {
    chatMessageStore.disconnectFromChat(chatId);
    clearTypingIndicator();
  });
</script>

<div class="chat-window">
  <div class="messages-container" bind:this={messageContainer}>
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
    border-radius: 8px;
    background-color: #f8f9fa;
    overflow: hidden;
  }
  
  .messages-container {
    flex: 1;
    padding: 16px;
    overflow-y: auto;
    scroll-behavior: smooth;
  }
  
  .message {
    margin-bottom: 12px;
    max-width: 80%;
  }
  
  .incoming {
    align-self: flex-start;
  }
  
  .outgoing {
    align-self: flex-end;
    margin-left: auto;
  }
  
  .sender {
    font-size: 0.8rem;
    margin-bottom: 2px;
    color: #6c757d;
  }
  
  .bubble {
    padding: 8px 12px;
    border-radius: 16px;
    word-wrap: break-word;
  }
  
  .incoming .bubble {
    background-color: #e9ecef;
    border-top-left-radius: 4px;
  }
  
  .outgoing .bubble {
    background-color: #007bff;
    color: white;
    border-top-right-radius: 4px;
  }
  
  .metadata {
    display: flex;
    font-size: 0.7rem;
    color: #6c757d;
    margin-top: 2px;
    gap: 4px;
  }
  
  .outgoing .metadata {
    justify-content: flex-end;
  }
  
  .deleted-message {
    font-style: italic;
    color: #6c757d;
  }
  
  .edited-indicator, .read-indicator {
    font-size: 0.7rem;
  }
  
  .typing-indicator {
    font-size: 0.8rem;
    color: #6c757d;
    margin-bottom: 8px;
    font-style: italic;
  }
  
  .input-container {
    display: flex;
    padding: 8px;
    border-top: 1px solid #dee2e6;
    background-color: white;
  }
  
  textarea {
    flex: 1;
    resize: none;
    border: 1px solid #ced4da;
    border-radius: 20px;
    padding: 8px 12px;
    font-family: inherit;
    outline: none;
  }
  
  textarea:focus {
    border-color: #007bff;
  }
  
  .send-button {
    margin-left: 8px;
    border: none;
    background-color: #007bff;
    color: white;
    border-radius: 20px;
    padding: 8px 16px;
    cursor: pointer;
  }
  
  .send-button:disabled {
    background-color: #6c757d;
    cursor: not-allowed;
  }
</style> 