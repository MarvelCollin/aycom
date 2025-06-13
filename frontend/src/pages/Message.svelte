<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import LeftSide from '../components/layout/LeftSide.svelte';
  import { useAuth } from '../hooks/useAuth';
  import { useTheme } from '../hooks/useTheme';
  import type { IAuthStore } from '../interfaces/IAuth';
  import { createLoggerWithPrefix } from '../utils/logger';
  import { toastStore } from '../stores/toastStore';
  import { checkAuth, isWithinTime, handleApiError } from '../utils/common';
  import { listChats, listMessages, sendMessage as apiSendMessage, unsendMessage as apiUnsendMessage, searchMessages, createChat, getChatHistoryList } from '../api/chat';
  import { getProfile, searchUsers, getUserById, getAllUsers } from '../api/user';
  import { websocketStore } from '../stores/websocketStore';
  import type { ChatMessage, MessageType } from '../stores/websocketStore';
  import DebugPanel from '../components/common/DebugPanel.svelte';
  import CreateGroupChat from '../components/chat/CreateGroupChat.svelte';
  import ThemeToggle from '../components/common/ThemeToggle.svelte';
  import { transformApiUsers, type StandardUser } from '../utils/userTransform';
  import Toast from '../components/common/Toast.svelte';
  
  import '../styles/pages/messages.css'; // Import the CSS file
  
  // Interface definitions for type safety
  interface Attachment {
    id: string;
    type: 'image' | 'gif' | 'video';
    url: string;
    thumbnail?: string;
  }
  
  interface Message {
    id: string;
    content: string;
    timestamp: string;
    sender_id: string;
    sender_name: string;
    sender_avatar?: string;
    is_own: boolean;
    is_read: boolean;
    is_deleted: boolean;
    attachments: Attachment[];
  }
  
  interface Participant {
    id: string;
    username: string;
    display_name: string;
    avatar: string | null;
    is_verified: boolean;
  }
  
  interface LastMessage {
    content: string;
    timestamp: string | number;
    sender_id: string;
    sender_name: string;
  }

  interface Chat {
    id: string;
    type: 'individual' | 'group';
    name: string;
    avatar: string | null;
    participants: Participant[];
    last_message?: LastMessage;
    messages: Message[];
    unread_count: number;
    profile_picture_url: string | null;
  }
  
  const logger = createLoggerWithPrefix('Message');

  const { getAuthState } = useAuth();
  const { theme } = useTheme();
  
  // Reactive declarations
  $: authState = getAuthState ? (getAuthState() as IAuthStore) : { user_id: null, is_authenticated: false, access_token: null, refresh_token: null };
  $: isDarkMode = $theme === 'dark';
  
  // User profile data
  let username = '';
  let displayName = '';
  let avatar = 'https://secure.gravatar.com/avatar/0?d=mp'; // Default avatar with proper image URL
  let isLoadingProfile = true;

  // Chat state
  let selectedChat: Chat | null = null;
  let chats: Chat[] = [];
  let isLoadingChats = true;
  let newMessage = '';
  let searchQuery = '';
  let filteredChats: Chat[] = [];
  let selectedAttachments: Attachment[] = [];
  let isLoadingUsers = false;
  let userSearchResults: StandardUser[] = [];
  
  // Handle attachment selection - placeholder function for now
  function handleAttachment(type: 'image' | 'gif') {
    // Implementation can be added later
    logger.debug(`Attachment selected: ${type}`);
    // For now, just show a message
    toastStore.showToast(`${type} attachment feature coming soon`, 'info');
  }
  
  // Group chat modal state
  let showGroupChatModal = false;

  // Mobile detection
  let isMobile = false;
  let showMobileMenu = false;
  
  // Group chat handlers
  function handleGroupChatCreated(event: any) {
    if (event && event.detail && event.detail.chat) {
      logger.debug('Group chat created', { chatId: event.detail.chat.id });
      
      // Add the new group chat to the chat list
      const newChat = formatGroupChatForDisplay(event.detail.chat);
      chats = [newChat, ...chats];
      filteredChats = [newChat, ...filteredChats];
      
      // Select the new chat
      selectChat(newChat);
      
      // Close the modal
      showGroupChatModal = false;
      
      // Show success notification
      toastStore.showToast('Group chat created successfully', 'success');
    }
  }
  
  // Chat interaction functions
  function selectChat(chat: Chat) {
    selectedChat = chat;
    
    // Additional chat selection logic...
  }
  
  function startChatWithUser(user: StandardUser) {
    // Logic to start a new chat with a user
    logger.debug(`Starting new chat with user: ${user.username}`);
    
    // Implementation here...
  }
  
  function getAvatarColor(name: string) {
    // Simple hash function for consistent colors
    let hash = 0;
    for (let i = 0; i < name.length; i++) {
      hash = name.charCodeAt(i) + ((hash << 5) - hash);
    }
    
    // Convert to HSL with good saturation and lightness
    const h = Math.abs(hash) % 360;
    return `hsl(${h}, 70%, 60%)`;
  }
  
  function getChatDisplayName(chat: Chat) {
    // Return a display name based on chat participants
    return chat.name || (chat.participants && chat.participants[0]?.display_name) || 'Chat';
  }
  
  function formatTimeAgo(timestamp: string | number) {
    // Simple time formatter
    if (!timestamp) return '';
    
    // Implementation here...
    return 'just now';
  }
  
  // Message handling functions
  function sendMessage() {
    // Logic to send a message
    if (!newMessage.trim()) return;
    
    // Implementation will call apiSendMessage
    newMessage = ''; // Clear input after sending
  }
  
  function unsendMessage(messageId: string) {
    // Logic to unsend/delete a message
    // Implementation will call apiUnsendMessage
  }
  
  async function handleSearch() {
    // Logic to search for users or messages
    if (searchQuery === 'new') {
      isLoadingUsers = true;
      try {
        const users = await getAllUsers();
        userSearchResults = transformApiUsers(users);
                } catch (error) {
        logger.error('Failed to fetch users', error);
        toastStore.showToast('Failed to load users', 'error');
    } finally {
        isLoadingUsers = false;
      }
    }
  }
  
  onMount(() => {
    // Check viewport size
    const checkViewport = () => {
      isMobile = window.innerWidth < 768;
    };
    
    checkViewport();
    window.addEventListener('resize', checkViewport);
    
    // Initialize user profile and chats
    if (authState && authState.is_authenticated) {
      fetchUserProfile();
      fetchChats();
    }
    
    return () => {
      window.removeEventListener('resize', checkViewport);
          };
        });
        
  function toggleMobileMenu() {
    showMobileMenu = !showMobileMenu;
  }
  
  // Call this function to fetch profile data
  async function fetchUserProfile() {
    // Implementation here...
  }
  
  // Call this function to load the chat list
  async function fetchChats() {
    // Implementation here...
  }
  
  // Format chat data for display
  function formatGroupChatForDisplay(chatData: any): Chat {
                return {
      id: chatData.id,
      name: chatData.name || 'New Group Chat',
      type: 'group' as const,  // Use const assertion to fix the type
      last_message: undefined,
      avatar: null,
      participants: chatData.participants?.map((p: any) => ({
        id: p.id || p.user_id,
          username: p.username || '',
        display_name: p.display_name || p.username || `User`,
        avatar: p.avatar_url || p.avatar || null,
          is_verified: p.is_verified || false
      })) || [],
        messages: [],
        unread_count: 0,
        profile_picture_url: null 
      };
  }
</script>

<div class="custom-message-layout {isDarkMode ? 'dark-theme' : ''}">
  <!-- Mobile header -->
  {#if isMobile}
    <div class="mobile-header">
      <button class="mobile-menu-button" on:click={toggleMobileMenu}>
        <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <line x1="3" y1="12" x2="21" y2="12"></line>
          <line x1="3" y1="6" x2="21" y2="6"></line>
          <line x1="3" y1="18" x2="21" y2="18"></line>
        </svg>
      </button>
      <h1 class="mobile-title">Messages</h1>
      <ThemeToggle size="sm" />
  </div>
  {/if}

  <!-- Sidebar -->
  <div class="custom-sidebar {isMobile && !showMobileMenu ? 'hidden' : ''}">
    <LeftSide 
      {username}
      {displayName}
      {avatar}
      isCollapsed={false}
      isMobileMenu={isMobile && showMobileMenu}
    />
  </div>
  
  <!-- Mobile menu overlay -->
  {#if isMobile && showMobileMenu}
    <div class="mobile-overlay" on:click={toggleMobileMenu}></div>
  {/if}

  <!-- Main content area -->
  <div class="custom-content-area">
    <div class="message-container {isDarkMode ? 'dark-theme' : ''}">
  <div class="middle-section">
        <!-- Chat header -->
    <div class="section-header">
      <h1>Messages</h1>
      <div class="button-group">
            <button
              class="compose-button"
              on:click={() => searchQuery = 'new'}
              aria-label="New message"
            >
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" width="20" height="20">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
          </svg>
              New
        </button>
            <button
              class="compose-button"
              on:click={() => showGroupChatModal = true}
              aria-label="New group"
            >
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" width="20" height="20">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
          </svg>
              Group
        </button>
            {#if !isMobile}
        <ThemeToggle size="sm" />
            {/if}
      </div>
    </div>

        <!-- Search container -->
    <div class="search-container">
        <input
          type="text"
            placeholder="Search messages..." 
          bind:value={searchQuery}
            on:input={handleSearch}
          class="search-input"
        />
          {#if searchQuery}
            <button class="clear-search" on:click={() => searchQuery = ''}>
              <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" width="18" height="18">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        {/if}
      </div>
      
        <!-- Chat list -->
        <div class="chat-list">
          {#if searchQuery === 'new' && userSearchResults.length > 0}
            <!-- User search results -->
            <div class="search-results">
              <h3>Users</h3>
                {#each userSearchResults as user}
                <div 
                  class="chat-item" 
                  on:click={() => startChatWithUser(user)}
                  on:keydown={(e) => e.key === 'Enter' && startChatWithUser(user)}
                  tabindex="0"
                >
                  <div class="avatar">
                        {#if user.avatar}
                      <img src={user.avatar} alt={user.display_name} />
                        {:else}
                      <div class="avatar-placeholder" style="background-color: {getAvatarColor(user.username)}">
                        {(user.display_name || user.username).charAt(0).toUpperCase()}
                      </div>
                        {/if}
                      </div>
                  <div class="chat-details">
                    <div class="chat-name">
                      {user.display_name}
                      {#if user.is_verified}
                        <span class="verified-badge">✓</span>
                        {/if}
                      </div>
                    <div class="username">@{user.username}</div>
                        </div>
                        </div>
                {/each}
            </div>
          {:else if isLoadingChats}
            <div class="loading-container">
              <div class="loading-spinner"></div>
              <p>Loading chats...</p>
            </div>
          {:else if chats.length === 0}
            <div class="empty-state">
              <p>No conversations yet</p>
              <button class="compose-button" on:click={() => searchQuery = 'new'}>
                Start a new chat
                  </button>
          </div>
          {:else}
              {#each filteredChats as chat}
              <div 
                class="chat-item {selectedChat?.id === chat.id ? 'selected' : ''}" 
                    on:click={() => selectChat(chat)}
                on:keydown={(e) => e.key === 'Enter' && selectChat(chat)}
                tabindex="0"
                  >
                <div class="avatar">
                      {#if chat.avatar}
                    <img src={chat.avatar} alt={getChatDisplayName(chat)} />
                  {:else if chat.type === 'group'}
                    <div class="avatar-placeholder group-avatar">
                      <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
                      </svg>
                    </div>
                      {:else}
                    <div class="avatar-placeholder" style="background-color: {getAvatarColor(getChatDisplayName(chat))}">
                      {getChatDisplayName(chat).charAt(0).toUpperCase()}
                    </div>
                      {/if}
                    </div>
                <div class="chat-details">
                      <div class="chat-header">
                    <div class="chat-name">
                      {getChatDisplayName(chat)}
                      </div>
                    {#if chat.last_message?.timestamp}
                      <div class="timestamp">{formatTimeAgo(chat.last_message.timestamp)}</div>
                          {/if}
                  </div>
                  <div class="last-message">
                    {#if chat.last_message}
                      <span>{chat.last_message.content}</span>
                        {:else}
                          <span class="no-messages">No messages yet</span>
                        {/if}
                      </div>
                    </div>
                    {#if chat.unread_count > 0}
                  <div class="unread-badge">{chat.unread_count}</div>
                    {/if}
          </div>
            {/each}
      {/if}
    </div>
  </div>

  <!-- Right: Chat Content -->
  <div class="right-section">
    {#if selectedChat}
      <div class="chat-header">
            {#if isMobile}
              <button class="back-button" on:click={() => selectedChat = null}>
                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" width="24" height="24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
                </svg>
              </button>
            {/if}
            
            <div class="avatar">
          {#if selectedChat.avatar}
                <img src={selectedChat.avatar} alt={getChatDisplayName(selectedChat)} />
              {:else if selectedChat.type === 'group'}
                <div class="avatar-placeholder group-avatar">
                  <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
                  </svg>
                </div>
          {:else}
                <div class="avatar-placeholder" style="background-color: {getAvatarColor(getChatDisplayName(selectedChat))}">
                  {getChatDisplayName(selectedChat).charAt(0).toUpperCase()}
                </div>
          {/if}
        </div>
            
            <div class="user-info">
              <div class="display-name">{getChatDisplayName(selectedChat)}</div>
              {#if selectedChat.type === 'group'}
                <div class="participants-info">{selectedChat.participants.length} members</div>
              {/if}
        </div>
      </div>
      
      <div class="messages-container">
            {#if selectedChat.messages && selectedChat.messages.length > 0}
        {#each selectedChat.messages as message}
                <div class="message-item {message.is_own ? 'own-message' : ''} {message.is_deleted ? 'deleted' : ''}">
              {#if !message.is_own}
                    <div class="message-avatar">
                      {#if message.sender_avatar}
                        <img src={message.sender_avatar} alt={message.sender_name} />
                      {:else}
                        <div class="avatar-placeholder" style="background-color: {getAvatarColor(message.sender_name)}">
                          {message.sender_name.charAt(0).toUpperCase()}
                        </div>
              {/if}
              </div>
                  {/if}
                  
                  <div class="message-content">
                    {#if !message.is_own && selectedChat.type === 'group'}
                      <div class="sender-name">{message.sender_name}</div>
                    {/if}
                    
                    {#if message.is_deleted}
                      <div class="deleted-message">Message deleted</div>
                    {:else}
                      <div class="content-text">{message.content}</div>
                      
                      {#if message.attachments && message.attachments.length > 0}
                        <div class="attachments-container">
                  {#each message.attachments as attachment}
                            {#if attachment.type === 'image'}
                              <img src={attachment.url} alt="Image attachment" class="image-attachment" />
                            {:else if attachment.type === 'gif'}
                              <img src={attachment.url} alt="GIF attachment" class="gif-attachment" />
                    {:else if attachment.type === 'video'}
                              <video src={attachment.url} controls class="video-attachment">
                        Your browser does not support the video tag.
                      </video>
                    {/if}
                  {/each}
                </div>
              {/if}
                      
                      <div class="message-footer">
                        <span class="timestamp">{formatTimeAgo(message.timestamp)}</span>
                        
                        {#if message.is_own}
                          <div class="message-actions">
                            <button class="action-button" on:click={() => unsendMessage(message.id)}>
                              <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" width="16" height="16">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                              </svg>
                  </button>
                          </div>
                {/if}
              </div>
                    {/if}
            </div>
          </div>
        {/each}
            {:else}
              <div class="empty-messages">
                <p>No messages yet</p>
                <p class="start-chat-prompt">Start the conversation!</p>
              </div>
            {/if}
      </div>
      
      <div class="message-input-container">
            <div class="input-wrapper">
              <textarea 
                bind:value={newMessage}
                placeholder="Type a message..."
                rows="1"
                on:keydown={(e) => e.key === 'Enter' && !e.shiftKey && sendMessage()}
              ></textarea>
              
              <div class="attachment-buttons">
                <div class="attachment-button" on:click={() => handleAttachment('image')}>
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" width="20" height="20">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
            </svg>
        </div>
                <div class="attachment-button" on:click={() => handleAttachment('gif')}>
                  <span class="gif-button">GIF</span>
                </div>
              </div>
            </div>
            
        <button
              class="send-button {newMessage.trim() ? 'active' : ''}"
          disabled={!newMessage.trim()}
              on:click={sendMessage}
        >
              <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" width="20" height="20">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8" />
          </svg>
        </button>
      </div>
    {:else}
          <div class="no-chat-selected">
            <div class="empty-state-image">
              <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" width="64" height="64">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
        </svg>
            </div>
            <h2>Your Messages</h2>
            <p>Send private messages to a friend or group</p>
            <button 
              class="compose-button"
              on:click={() => searchQuery = 'new'}
            >
              New Message
        </button>
      </div>
    {/if}
      </div>
  </div>
</div>

<!-- Add the Group Chat modal -->
{#if showGroupChatModal}
  <div class="modal-overlay">
    <div class="modal-container">
      <CreateGroupChat 
        onSuccess={handleGroupChatCreated} 
        onCancel={() => showGroupChatModal = false} 
      />
    </div>
  </div>
{/if}

  <!-- Toast and DebugPanel -->
  <Toast />
  <DebugPanel />
</div>

<style>
  /* Custom layout styles */
  :global(body) {
    margin: 0;
    padding: 0;
    overflow: hidden;
  }
  
  .custom-message-layout {
    display: flex;
    width: 100%;
    height: 100vh;
    background-color: var(--bg-primary);
    color: white;
    overflow: hidden;
  }
  
  .custom-message-layout.dark-theme {
    background-color: var(--bg-primary-dark);
    color: white;
  }
  
  /* Make all text in the sidebar white */
  .custom-sidebar {
    width: 250px;
    min-width: 250px;
    border-right: 1px solid var(--border-color);
    height: 100vh;
    position: sticky;
    top: 0;
    z-index: 100;
    background-color: var(--bg-primary);
    color: white;
  }
  
  /* Override LeftSide component text colors */
  .custom-sidebar :global(.sidebar-nav-item),
  .custom-sidebar :global(.sidebar-nav-text),
  .custom-sidebar :global(.sidebar-profile-name),
  .custom-sidebar :global(.sidebar-profile-username),
  .custom-sidebar :global(.sidebar-logo-text) {
    color: white !important;
  }
  
  .custom-sidebar :global(.sidebar-nav-icon) {
    color: white !important;
  }
  
  .custom-sidebar.hidden {
    display: none;
  }
  
  .custom-content-area {
    flex: 1;
    display: flex;
    height: 100vh;
    overflow: hidden;
  }
  
  .message-container {
    display: flex;
    width: 100%;
    height: 100%;
    overflow: hidden;
  }
  
  /* Mobile styles */
  .mobile-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 12px 16px;
    border-bottom: 1px solid var(--border-color);
    background-color: var(--bg-primary);
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    z-index: 50;
    height: 56px;
  }
  
  .mobile-title {
    font-size: 18px;
    font-weight: 600;
    margin: 0;
    color: white;
  }
  
  .mobile-menu-button {
    background: none;
    border: none;
    color: white;
    cursor: pointer;
    padding: 4px;
    display: flex;
    align-items: center;
    justify-content: center;
  }
  
  .mobile-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background-color: rgba(0, 0, 0, 0.5);
    z-index: 90;
  }
  
  @media (max-width: 768px) {
    .custom-message-layout {
      flex-direction: column;
    }
    
    .custom-sidebar {
      position: fixed;
      top: 0;
      left: 0;
      bottom: 0;
      z-index: 100;
      width: 80%;
      max-width: 280px;
    }
    
    .custom-content-area {
      padding-top: 56px; /* Height of mobile header */
    }
    
    .message-container {
      height: calc(100vh - 56px);
    }
  }
</style>


