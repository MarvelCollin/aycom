<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import LeftSide from '../components/layout/LeftSide.svelte';
  import { useTheme } from '../hooks/useTheme';
  import type { IAuthStore } from '../interfaces/IAuth';
  import { createLoggerWithPrefix } from '../utils/logger';
  import { toastStore } from '../stores/toastStore';
  import { authStore } from '../stores/authStore';
  import { checkAuth, isWithinTime, handleApiError } from '../utils/common';
  import * as chatApi from '../api/chat';
  import { getProfile, searchUsers, getUserById, getAllUsers } from '../api/user';
  import { websocketStore } from '../stores/websocketStore';
  import type { ChatMessage, MessageType } from '../stores/websocketStore';
  import DebugPanel from '../components/common/DebugPanel.svelte';
  import CreateGroupChat from '../components/chat/CreateGroupChat.svelte';
  import NewChatModal from '../components/chat/NewChatModal.svelte';
  import ThemeToggle from '../components/common/ThemeToggle.svelte';
  import { transformApiUsers, type StandardUser } from '../utils/userTransform';
  import Toast from '../components/common/Toast.svelte';
  import { formatRelativeTime } from '../utils/date'; 
  
  const { 
    listChats, 
    listMessages, 
    sendMessage: apiSendMessage, 
    unsendMessage: apiUnsendMessage, 
    searchMessages, 
    createChat, 
    getChatHistoryList, 
    testApiConnection, 
    logAuthTokenInfo, 
    setMessageHandler,
    listChatParticipants
  } = chatApi;
  
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
    chat_id: string;
    sender_id: string;
    sender_name?: string;
    sender_avatar?: string;
    content: string;
    timestamp: string | number | Date;
    is_read: boolean;
    is_edited: boolean;
    is_deleted: boolean;
    failed?: boolean;
    is_local?: boolean;
    attachments?: Array<{
      id: string;
      type: string;
      url: string;
      name?: string;
    }>;
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
    timestamp: string;  // Only allow string type
    sender_id: string;
    sender_name?: string;
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
    created_at: string;
    updated_at: string;
  }
  
  const logger = createLoggerWithPrefix('Message');

  const { theme } = useTheme();
  
  // Reactive declarations
  $: authState = $authStore;
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
  let isLoadingMessages = false;
  let newMessage = '';
  let searchQuery = '';
  let filteredChats: Chat[] = [];
  let selectedAttachments: Attachment[] = [];
  let isLoadingUsers = false;
  let userSearchResults: StandardUser[] = [];
  let userSearchQuery = '';
  
  // Mobile view state
  let isMobile = false;
  let showMobileMenu = false;
  
  // Modal visibility flags
  let showNewChatModal = false;
  let showCreateGroupModal = false;
  let showDebug = false;
  
  // Toast notifications
  let toasts = [];
  
  // Group chat modal state
  let showGroupChatModal = false;
  
  // Handle attachment selection - placeholder function for now
  function handleAttachment(type: 'image' | 'gif') {
    // Implementation can be added later
    logger.debug(`Attachment selected: ${type}`);
    // For now, just show a message
    toastStore.showToast(`${type} attachment feature coming soon`, 'info');
  }
  
  // Function to check viewport size and set mobile state
  function checkViewport() {
    isMobile = window.innerWidth < 768;
  }
  
  // Initialize WebSocket monitoring
  $: {
    // Monitor connection status changes and reconnect if needed
    if ($websocketStore) {
      const isWsConnected = $websocketStore.connected;
      
      if (!isWsConnected && selectedChat) {
        logger.debug(`WebSocket detected as disconnected with chat ${selectedChat.id} selected. Will attempt reconnect.`);
        setTimeout(() => initializeWebSocketConnections(), 1000);
      }
    }
  }
  
  // Improved function to initialize WebSocket connections for active chats
  function initializeWebSocketConnections() {
    try {
      // First priority: connect to selected chat
      if (selectedChat) {
        logger.info(`Connecting to WebSocket for selected chat: ${selectedChat.id}`);
        try {
          // Use type assertion to avoid TypeScript error
          (websocketStore as any).connect(selectedChat.id);
        } catch (err) {
          logger.error(`Error connecting to WebSocket for selected chat: ${err}`);
        }
      }
      
      // Second priority: connect to most recent chats (up to 3)
      if (chats && chats.length > 0) {
        const recentChats = chats.slice(0, 3); // Limit to 3 recent chats
        
        for (const chat of recentChats) {
          // Skip if it's the selected chat (already connected)
          if (selectedChat && chat.id === selectedChat.id) continue;
          
          logger.debug(`Connecting to WebSocket for recent chat: ${chat.id}`);
          try {
            // Use type assertion to avoid TypeScript error
            (websocketStore as any).connect(chat.id);
          } catch (err) {
            // Log but don't show UI error for background connections
            logger.error(`Error connecting to WebSocket for chat ${chat.id}: ${err}`);
          }
        }
      }
    } catch (error) {
      logger.error('Error initializing WebSocket connections:', error);
    }
  }
  
  // WebSocket handling
  function handleWebSocketMessage(message: any) {
    // Processing websocket message...
    const senderId = message.sender_id;
    
    if (senderId === $authStore.user_id) {
      logger.debug('Skipping own message from WebSocket:', message);
      return;
    }
    
    logger.info(`Received new message in chat ${message.chat_id} from ${senderId}`);

    // Format the message with all required properties
    const newMessage: Message = {
      id: message.message_id || `ws-${Date.now()}`,
      chat_id: message.chat_id || '',
      content: message.content || '',
      timestamp: ensureStringTimestamp(message.timestamp || new Date()),
      sender_id: senderId,
      sender_name: 'User', // Default name
      sender_avatar: undefined,
      is_read: false,
      is_edited: false,
      is_deleted: message.is_deleted || false
    };
      
    // Update the selected chat if this message belongs to it
    if (selectedChat && selectedChat.id === message.chat_id) {
      selectedChat = {
        ...selectedChat,
        messages: [...selectedChat.messages, newMessage]
      };
      
      // Scroll to bottom
      setTimeout(() => {
        const messagesContainer = document.querySelector('.messages-container');
        if (messagesContainer) {
          messagesContainer.scrollTop = messagesContainer.scrollHeight;
        }
      }, 100);
    }
    
    // Create a properly formatted last message
    const lastMessage: LastMessage = {
      content: message.content || '',
      timestamp: ensureStringTimestamp(typeof message.timestamp === 'string' ? message.timestamp : new Date().toISOString()),
      sender_id: senderId,
      sender_name: 'User'
    };

    // Update the chat in the list with the new last message
    const updatedChats = chats.map(chat => {
      if (chat.id === message.chat_id) {
        return {
          ...chat,
          last_message: lastMessage,
          unread_count: selectedChat?.id === chat.id ? 0 : (chat.unread_count || 0) + 1
        };
      }
      return chat;
    }) as Chat[]; // Type assertion
    
    // Update chats state
    chats = updatedChats;
    
    // Also update filtered chats
    filteredChats = filteredChats.map(chat => {
      if (chat.id === message.chat_id) {
        return {
          ...chat,
          last_message: lastMessage,
          unread_count: selectedChat?.id === chat.id ? 0 : (chat.unread_count || 0) + 1
        };
      }
      return chat;
    }) as Chat[];
    
    // Play notification sound if not viewing this chat
    if (selectedChat?.id !== message.chat_id) {
      // Add sound notification here if needed
    }
  }
  
  // Helper function to send messages via WebSocket
  function sendWebSocketMessage(chatId: string, content: string) {
    const chatMessage: ChatMessage = {
      type: 'text',
      content: content,
      chat_id: chatId,
      user_id: $authStore.user_id || ''
    };
    
    // Use any to bypass TypeScript type checking
    (websocketStore as any).sendMessage(chatId, chatMessage);
  }
  
  // Mobile navigation handling
  function handleMobileNavigation(view: string): void {
    if (view === 'back' || view === 'showChats') {
      selectedChat = null;
    } else if (view === 'showChat' && selectedChat) {
      // Already handled
    }
  }
  
  // Mobile menu toggle
  function toggleMobileMenu() {
    showMobileMenu = !showMobileMenu;
  }
  
  // Helper functions
  function formatGroupChatForDisplay(apiChat: any): Chat {
    let avatar = null;
    
    // Format to match our Chat type
    return {
      id: apiChat.id,
      type: apiChat.is_group_chat ? 'group' : 'individual',
      name: apiChat.name || 'Group Chat',
      avatar: avatar,
      participants: (apiChat.participants || []).map(p => ({
        id: p.id || p.user_id,
        username: p.username || 'User',
        display_name: p.display_name || p.username || 'User',
        avatar: p.profile_picture_url || p.avatar || null,
        is_verified: p.is_verified || false
      })),
      messages: [],
      unread_count: 0,
      profile_picture_url: null,
      created_at: apiChat.created_at || new Date().toISOString(),
      updated_at: apiChat.updated_at || new Date().toISOString()
    };
  }
  
  function formatTimeForChat(timestamp: string | number | Date): string {
    let date: Date;
    
    if (timestamp instanceof Date) {
      date = timestamp;
    } else if (typeof timestamp === 'number') {
      date = new Date(timestamp);
    } else {
      date = new Date(timestamp);
    }
    
    return date.toLocaleString('en-US', {
      hour: 'numeric',
      minute: 'numeric',
      hour12: true
    });
  }
  
  async function fetchChats() {
    isLoadingChats = true;
    
    try {
      const response = await getChatHistoryList();
      
      if (response && response.chats) {
        // Process chats
        const processedChats = response.chats.map(chat => {
          // Format the chat to match our Chat interface
          const chatObj: Chat = {
            id: chat.id,
            type: chat.is_group_chat ? 'group' : 'individual',
            name: chat.name || '',
            avatar: null,
            participants: (chat.participants || []).map(p => {
              // Ensure we have proper participant data
              return {
                id: p.id || p.user_id || '',
                username: p.username || '',
                display_name: p.display_name || p.name || p.username || '',
                avatar: p.profile_picture_url || p.avatar || null,
                is_verified: p.is_verified || false
              };
            }),
            messages: [],
            unread_count: chat.unread_count || 0,
            profile_picture_url: null,
            created_at: chat.created_at || new Date().toISOString(),
            updated_at: chat.updated_at || new Date().toISOString()
          };
          
          // Add last message if available
          if (chat.last_message) {
            chatObj.last_message = {
              content: chat.last_message.content || '',
              timestamp: ensureStringTimestamp(chat.last_message.timestamp || chat.last_message.sent_at || new Date()),
              sender_id: chat.last_message.sender_id || '',
              sender_name: chat.last_message.sender_name || 'User'
            };
          }
          
          return chatObj;
        });
        
        // Sort chats by update time, newest first
        processedChats.sort((a, b) => {
          const timeA = new Date(a.updated_at).getTime();
          const timeB = new Date(b.updated_at).getTime();
          return timeB - timeA;
        });
        
        chats = processedChats;
        filteredChats = [...processedChats];
        
        // If no chat is selected yet and we have chats, select the first one
        if (!selectedChat && processedChats.length > 0) {
          selectChat(processedChats[0]);
        }
      } else {
        logger.warn('No chats found in response', response);
        chats = [];
        filteredChats = [];
      }
    } catch (error) {
      handleApiError(error, 'Failed to load chats');
      chats = [];
      filteredChats = [];
    } finally {
      isLoadingChats = false;
    }
  }
  
  function formatMessageTime(timestamp: string | number | Date): string {
    let stringTimestamp: string;
    
    if (timestamp instanceof Date) {
      stringTimestamp = timestamp.toISOString();
    } else if (typeof timestamp === 'number') {
      stringTimestamp = new Date(timestamp).toISOString();
    } else {
      stringTimestamp = timestamp;
    }
    
    return formatRelativeTime(stringTimestamp);
  }

  /**
   * Search for users to add to chats
   */
  async function searchForUsers(query: string) {
    if (!query || query.length < 2) {
      userSearchResults = [];
      return;
    }
    
    userSearchQuery = query;
    isLoadingUsers = true;
    
    try {
      const results = await searchUsers(query);
      userSearchResults = transformApiUsers(results);
    } catch (error) {
      handleApiError(error, 'Failed to search for users');
      userSearchResults = [];
    } finally {
      isLoadingUsers = false;
    }
  }
  
  /**
   * Initialize a new chat with a user
   */
  async function initiateNewChat(data: any) {
    try {
      isLoadingChats = true;
      
      // Check if we received an object with chat data or just a user ID
      let chatData;
      if (typeof data === 'string') {
        // Handle legacy format (just user ID)
        chatData = {
          type: 'individual',
          participants: [data]
        };
      } else if (typeof data === 'object') {
        // Use the object data directly, but ensure it uses the correct field name
        chatData = data;
        
        // Convert participant_ids to participants if needed
        if (data.participant_ids && !data.participants) {
          chatData = {
            ...data,
            participants: data.participant_ids
          };
          delete chatData.participant_ids;
        }
      } else {
        throw new Error('Invalid chat data format');
      }
      
      const response = await createChat(chatData);
      
      if (response && response.chat_id) {
        showNewChatModal = false;
        await fetchChats(); // Use fetchChats instead of loadChats
        selectChat(response.chat_id);
      }
    } catch (error) {
      handleApiError(error, 'Failed to create chat');
    } finally {
      isLoadingChats = false;
    }
  }
  
  function getUserDisplayName(userId: string): string {
    // Check if this is the current user
    if (userId === $authStore.user_id) {
      return displayName || 'You';
    }
    
    // Check if the user is in the participants list of any chat
    for (const chat of chats) {
      const participant = chat.participants.find(p => p.id === userId);
      if (participant) {
        return participant.display_name || participant.username || 'Unknown User';
      }
    }
    
    // If we can't find the user, return a generic name with the ID
    const shortId = userId.substring(0, 4);
    return `User ${shortId}`;
  }
  
  function getOtherParticipant(chat: Chat): Participant | undefined {
    if (chat.type !== 'individual') return undefined;
    
    return chat.participants.find(p => p.id !== $authStore.user_id);
  }

  // Initialize connections when component mounts
  onMount(async () => {
    // Check viewport size
    checkViewport();
    window.addEventListener('resize', checkViewport);
    
    // Fetch user profile
    try {
      const profileData = await getProfile();
      if (profileData) {
        username = profileData.username || '';
        displayName = profileData.display_name || profileData.username || 'User';
        avatar = profileData.profile_picture_url || 'https://secure.gravatar.com/avatar/0?d=mp';
      }
    } catch (error) {
      logger.error('Failed to load profile:', error);
    } finally {
      isLoadingProfile = false;
    }
    
    // Load chats
    await fetchChats();
    
    // Register WebSocket message handler
    setMessageHandler(handleWebSocketMessage);
    
    // Initialize WebSockets with 500ms delay to ensure chats are loaded
    setTimeout(initializeWebSocketConnections, 500);
    
    logger.info('Message component mounted');
  });
  
  // Clean up when component unmounts
  onDestroy(() => {
    window.removeEventListener('resize', checkViewport);
    logger.info('Disconnecting from all WebSocket connections');
    websocketStore.disconnectAll();
  });

  // Add a helper function to ensure timestamp is always a string
  function ensureStringTimestamp(timestamp: string | number | Date): string {
    if (timestamp instanceof Date) {
      return timestamp.toISOString();
    } else if (typeof timestamp === 'number') {
      return new Date(timestamp).toISOString();
    }
    return timestamp;
  }

  // Mark as active in URL
  if (typeof window !== 'undefined') {
    const url = new URL(window.location.href);
    url.searchParams.set('chat', targetChat.id);
    window.history.replaceState({}, '', url.toString());
  }
</script>

<style>
  /* Fix for component-level layout issues */
  :global(body, html, #app) {
    height: 100%;
    margin: 0;
    padding: 0;
    overflow: hidden;
    background-color: var(--bg-primary, #121417);
  }
  
  :global(.app-container) {
    height: 100vh;
    width: 100%;
    display: flex;
    overflow: hidden;
  }

  .websocket-status {
    position: absolute;
    top: 10px;
    right: 10px;
    z-index: 10;
    display: flex;
    flex-direction: column;
    align-items: flex-end;
    gap: 5px;
    transition: transform 0.3s ease, opacity 0.3s ease;
  }
  
  .websocket-status.attention {
    animation: pulse 2s infinite;
  }
  
  @keyframes pulse {
    0% { transform: scale(1); }
    50% { transform: scale(1.05); }
    100% { transform: scale(1); }
  }
  
  .status-indicator {
    display: flex;
    align-items: center;
    gap: 5px;
    padding: 5px 10px;
    border-radius: 20px;
    font-size: 12px;
    background-color: rgba(0, 0, 0, 0.1);
    backdrop-filter: blur(5px);
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
    transition: all 0.3s ease;
  }
  
  .status-indicator.connected {
    color: #4caf50;
    background-color: rgba(76, 175, 80, 0.1);
  }
  
  .status-indicator.disconnected {
    color: #f44336;
    background-color: rgba(244, 67, 54, 0.2);
    border: 1px solid rgba(244, 67, 54, 0.4);
  }
  
  .status-icon {
    font-size: 14px;
    animation: blink 2s infinite;
  }
  
  @keyframes blink {
    0% { opacity: 0.5; }
    50% { opacity: 1; }
    100% { opacity: 0.5; }
  }
  
  .status-error {
    display: flex;
    align-items: center;
    gap: 5px;
    padding: 5px 10px;
    border-radius: 20px;
    font-size: 12px;
    background-color: rgba(244, 67, 54, 0.2);
    color: #f44336;
    backdrop-filter: blur(5px);
    box-shadow: 0 2px 8px rgba(244, 67, 54, 0.25);
    border: 1px solid rgba(244, 67, 54, 0.4);
    max-width: 300px;
  }
  
  .reconnect-button {
    background-color: #f44336;
    color: white;
    border: none;
    border-radius: 4px;
    padding: 4px 10px;
    margin-left: 8px;
    font-size: 12px;
    font-weight: bold;
    cursor: pointer;
    display: flex;
    align-items: center;
    gap: 4px;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
    transition: all 0.2s ease;
  }
  
  .reconnect-button:hover {
    background-color: #d32f2f;
    transform: translateY(-1px);
    box-shadow: 0 3px 6px rgba(0, 0, 0, 0.3);
  }
  
  .error-dismiss {
    background: none;
    border: none;
    color: #f44336;
    cursor: pointer;
    font-size: 16px;
    padding: 0 4px;
    margin-left: 5px;
    font-weight: bold;
  }
  
  .error-dismiss:hover {
    color: #d32f2f;
  }
</style>

<div class="custom-message-layout {isDarkMode ? 'dark-theme' : ''}">
  <!-- Sidebar -->
  <div class="custom-sidebar {isMobile && !showMobileMenu ? 'hidden' : ''}">
    <LeftSide 
      {username}
      {displayName}
      {avatar}
      isCollapsed={false}
      isMobileMenu={isMobile && showMobileMenu}
      on:closeMobileMenu={() => showMobileMenu = false}
    />
  </div>
  
  <!-- Mobile menu overlay -->
  {#if isMobile && showMobileMenu}
    <div class="mobile-overlay" 
         on:click={toggleMobileMenu} 
         on:keydown={(e) => e.key === 'Enter' && toggleMobileMenu()}
         role="button"
         tabindex="0"
         aria-label="Close mobile menu"></div>
  {/if}

  <!-- Main content area -->
  <div class="custom-content-area">
    <!-- Mobile header -->
    {#if isMobile}
      <div class="mobile-header">
        <button class="mobile-menu-button" on:click={toggleMobileMenu} aria-label="Toggle mobile menu">
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

    <!-- WebSocket Status Display -->
    <div class="websocket-status {!$websocketStore.connected ? 'attention' : ''}">
      <div class="status-indicator {$websocketStore.connected ? 'connected' : 'disconnected'}">
        <span class="status-icon">{$websocketStore.connected ? '●' : '○'}</span>
        <span class="status-text">Real-time {$websocketStore.connected ? 'Connected' : 'Disconnected'}</span>
        {#if !$websocketStore.connected}
          <button class="reconnect-button" on:click={() => {
            toastStore.showToast('Reconnecting to chat servers...', 'info');
            initializeWebSocketConnections();
          }}>
            <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M21.5 2v6h-6M2.5 22v-6h6M2 11.5a10 10 0 0 1 18.8-4.3M22 12.5a10 10 0 0 1-18.8 4.3"/>
            </svg>
            Reconnect Now
          </button>
        {/if}
      </div>
      {#if $websocketStore.lastError}
        <div class="status-error">
          <span class="error-icon">⚠️</span>
          <span class="error-text">{$websocketStore.lastError}</span>
          <button class="error-dismiss" on:click={() => (websocketStore as any).resetError()}>×</button>
        </div>
      {/if}
    </div>

    <div class="message-container {isDarkMode ? 'dark-theme' : ''}">
      <!-- Middle section - Chat list -->
      <div class="middle-section {selectedChat && isMobile ? 'hidden' : ''}">
        <!-- Chat header -->
        <div class="chat-list-header">
          <h2 class="page-title">Messages</h2>
          <div class="header-actions">
            <button class="msg-new-message-button" on:click={() => showNewChatModal = true}>
              <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <line x1="12" y1="5" x2="12" y2="19"></line>
                <line x1="5" y1="12" x2="19" y2="12"></line>
          </svg>
              <span>New</span>
        </button>
      </div>
    </div>

        <!-- Search container -->
        <div class="msg-search-container">
        <input
          type="text"
            placeholder="Search messages..." 
          bind:value={searchQuery}
            on:input={handleSearch}
            class="msg-search-input"
        />
          {#if searchQuery}
            <button class="msg-clear-search" on:click={() => { searchQuery = ''; handleSearch(); }} aria-label="Clear search">
              <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" width="18" height="18">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        {/if}
      </div>

         <!-- Chat list -->
      <div class="chat-list">
        {#if isLoadingChats}
          <div class="loading-container">
            <div class="loading-spinner"></div>
            <p>Loading chats...</p>
          </div>
        {:else if chats.length === 0}
            <div class="msg-empty-state">
              <div class="msg-empty-state-icon">
                <svg xmlns="http://www.w3.org/2000/svg" width="70" height="70" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1" stroke-linecap="round" stroke-linejoin="round">
                  <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"></path>
                </svg>
              </div>
              <h2>No conversations yet</h2>
              <p>Start a new conversation with friends</p>
              <button class="msg-new-message-button" on:click={() => showNewChatModal = true}>
              Start a new chat
            </button>
          </div>
        {:else}
            {#each filteredChats as chat (chat.id)}
            <div 
                class="msg-chat-item {selectedChat?.id === chat.id ? 'selected' : ''}"
              on:click={() => selectChat(chat)}
              on:keydown={(e) => e.key === 'Enter' && selectChat(chat)}
              role="button"
              tabindex="0"
              >
                {#if chat.type === 'individual'}
                  <div class="msg-avatar">
                    {#if getOtherParticipant(chat)?.avatar}
                      <img src={getOtherParticipant(chat)?.avatar || ''} alt={getOtherParticipant(chat)?.display_name || ''} />
                      {:else}
                    <div class="avatar-placeholder" style="background-color: {getAvatarColor(getChatDisplayName(chat))}">
                      {getChatDisplayName(chat).charAt(0).toUpperCase()}
                    </div>
                      {/if}
                    </div>
                {:else}
                  <div class="msg-avatar">
                    {#if chat.avatar}
                      <img src={chat.avatar} alt={chat.name} />
                    {:else}
                      <div class="avatar-placeholder" style="background-color: {getAvatarColor(chat.name)}">
                        {chat.name.charAt(0).toUpperCase()}
                      </div>
                    {/if}
                  </div>
                {/if}
                
                <div class="chat-details">
                      <div class="chat-header">
                    <div class="chat-name">
                      {getChatDisplayName(chat)}
                      </div>
                    {#if chat.last_message?.timestamp}
                      <div class="timestamp">{formatMessageTime(chat.last_message.timestamp)}</div>
                          {/if}
                  </div>
                  <div class="msg-last-message">
                    {#if chat.last_message}
                      <span>{chat.last_message.content}</span>
                        {:else}
                      <span class="msg-no-messages">No messages yet</span>
                        {/if}
                      </div>
                    </div>
                    {#if chat.unread_count > 0}
                  <div class="msg-unread-badge">{chat.unread_count}</div>
                    {/if}
          </div>
            {/each}
      {/if}
    </div>
  </div>

      <!-- Right section - Chat content -->
      <div class="right-section {selectedChat && isMobile ? 'full-width' : ''}">
    {#if selectedChat}
          <!-- Chat header -->
          <div class="msg-chat-header">
            {#if selectedChat.type === 'individual'}
              <div class="msg-avatar">
                {#if getOtherParticipant(selectedChat)?.avatar}
                  <img src={getOtherParticipant(selectedChat)?.avatar || ''} alt={getOtherParticipant(selectedChat)?.display_name || ''} />
          {:else}
                <div class="avatar-placeholder" style="background-color: {getAvatarColor(getChatDisplayName(selectedChat))}">
                  {getChatDisplayName(selectedChat).charAt(0).toUpperCase()}
                </div>
          {/if}
        </div>
            {:else}
              <div class="msg-avatar">
                {#if selectedChat.avatar}
                  <img src={selectedChat.avatar} alt={selectedChat.name} />
                {:else}
                  <div class="avatar-placeholder" style="background-color: {getAvatarColor(selectedChat.name)}">
                    {selectedChat.name.charAt(0).toUpperCase()}
                  </div>
                {/if}
              </div>
            {/if}
            
            <div class="msg-chat-header-info">
              <div class="msg-chat-header-name">{getChatDisplayName(selectedChat)}</div>
              {#if selectedChat.type === 'individual' && getOtherParticipant(selectedChat)}
                <div class="msg-chat-header-status">
                  {getOtherParticipant(selectedChat)?.is_verified ? '✓ Verified' : 'Online'}
                </div>
              {:else}
                <div class="msg-chat-header-status">
                  {selectedChat.participants.length} members
                </div>
              {/if}
        </div>
            
            <div class="msg-chat-header-actions">
              <button class="msg-action-icon" on:click={() => {/* Implement video call */}} aria-label="Video call">
                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" width="20" height="20">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" />
                </svg>
              </button>
              <button class="msg-action-icon" on:click={() => {/* Implement call */}} aria-label="Voice call">
                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" width="20" height="20">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 5a2 2 0 012-2h3.28a1 1 0 01.948.684l1.498 4.493a1 1 0 01-.502 1.21l-2.257 1.13a11.042 11.042 0 005.516 5.516l1.13-2.257a1 1 0 011.21-.502l4.493 1.498a1 1 0 01.684.949V19a2 2 0 01-2 2h-1C9.716 21 3 14.284 3 6V5z" />
                </svg>
              </button>
              <button class="msg-action-icon" on:click={() => {/* Implement options */}} aria-label="More options">
                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" width="20" height="20">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 5v.01M12 12v.01M12 19v.01M12 6a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2z" />
                </svg>
              </button>
            </div>
      </div>
      
      <div class="messages-container">
            {#if isLoadingMessages}
              <div class="loading-container">
                <div class="loading-spinner"></div>
                <p>Loading messages...</p>
              </div>
            {:else if selectedChat.messages && selectedChat.messages.length > 0}
        {#each selectedChat.messages as message}
                <div class="msg-conversation-item {message.sender_id === $authStore.user_id ? 'own-message' : ''} {message.is_deleted ? 'deleted' : ''} {message.failed ? 'failed' : ''}">
              {#if message.sender_id !== $authStore.user_id}
                    <div class="msg-avatar">
                      {#if message.sender_avatar}
                        <img src={message.sender_avatar} alt={message.sender_name} />
                      {:else}
                        <div class="avatar-placeholder" style="background-color: {getAvatarColor(message.sender_name || 'User')}">
                          {(message.sender_name || 'User').charAt(0).toUpperCase()}
                        </div>
              {/if}
              </div>
                  {/if}
                  
                  <div class="message-bubble {message.sender_id === $authStore.user_id ? 'sent' : 'received'}" class:failed={message.failed} class:is-local={message.is_local}>
                    <!-- Message content -->
                    <div class="message-content">
                    {#if message.is_deleted}
                        <span class="deleted-message">Message deleted</span>
                    {:else}
                        <p>{message.content}</p>
                        
                        <!-- Show retry option for local/failed messages -->
                        {#if message.failed || message.is_local}
                          <div class="message-error">
                            <span class="error-text">Not sent to server</span>
                            <button class="retry-btn" on:click={() => {
                              // Copy message content back to input field
                              newMessage = message.content;
                              // Remove the failed message
                              if (selectedChat?.messages) {
                                selectedChat = {
                                  ...selectedChat,
                                  messages: selectedChat.messages.filter(msg => msg.id !== message.id)
                                };
                              }
                            }}>Retry</button>
                          </div>
                        {/if}
                        
                        <!-- Message attachments -->
                      {#if message.attachments && message.attachments.length > 0}
                          <div class="message-attachments">
                  {#each message.attachments as attachment}
                              <div class="attachment">
                            {#if attachment.type === 'image'}
                                  <img src={attachment.url} alt="" />
                                {:else if attachment.type === 'file'}
                                  <div class="file-attachment">
                                    <span class="file-name">{attachment.name}</span>
                                    <a href={attachment.url} download>Download</a>
                                  </div>
                    {/if}
                              </div>
                  {/each}
                </div>
              {/if}
                      {/if}
                    </div>
                      
                    <!-- Message footer with timestamp -->
                      <div class="message-footer">
                      <span class="timestamp">{formatMessageTime(message.timestamp)}</span>
                        
                      <!-- Message actions for sent messages -->
                      {#if !message.is_deleted && message.sender_id === $authStore.user_id && !message.is_local}
                          <div class="message-actions">
                          <button class="action-btn" on:click={() => unsendMessage(message.id)}>
                            <span class="material-icons">delete</span>
                  </button>
                          </div>
                {/if}
              </div>
            </div>
          </div>
        {/each}
            {:else}
              <div class="msg-empty-state">
                <div class="msg-empty-state-icon">
                  <svg xmlns="http://www.w3.org/2000/svg" width="70" height="70" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1" stroke-linecap="round" stroke-linejoin="round">
                    <circle cx="12" cy="12" r="10"></circle>
                    <line x1="8" y1="12" x2="16" y2="12"></line>
                  </svg>
                </div>
                <h2>No messages yet</h2>
                <p>Start the conversation!</p>
              </div>
            {/if}
      </div>
      
          <div class="msg-message-input-container">
            <div class="msg-input-wrapper">
              <textarea 
                bind:value={newMessage}
                placeholder="Type a message..."
                rows="1"
                on:keydown={(e) => e.key === 'Enter' && !e.shiftKey && sendMessage(newMessage)}
              ></textarea>
              
              <div class="msg-attachment-buttons">
                <button class="msg-attachment-button" on:click={() => handleAttachment('image')} aria-label="Add image">
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" width="20" height="20">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
            </svg>
        </button>
                <button class="msg-attachment-button" on:click={() => handleAttachment('gif')} aria-label="Add GIF">
                  <span class="msg-gif-button">GIF</span>
                </button>
              </div>
            </div>
            
        <button
              class="msg-send-button {newMessage.trim() ? 'active' : ''}"
          disabled={!newMessage.trim()}
              on:click={() => sendMessage(newMessage)}
              aria-label="Send message"
        >
              <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" width="20" height="20">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8" />
          </svg>
        </button>
      </div>
    {:else}
          <div class="msg-empty-state">
            <div class="msg-empty-state-icon">
              <svg xmlns="http://www.w3.org/2000/svg" width="70" height="70" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1" stroke-linecap="round" stroke-linejoin="round">
                <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"></path>
        </svg>
            </div>
            <h2>Your Messages</h2>
            <p>Send private messages to a friend or group</p>
            <button class="msg-new-message-button" on:click={() => showNewChatModal = true}>
              New Message
            </button>
      </div>
    {/if}
      </div>
  </div>
</div>
    </div>

<!-- Modals -->
{#if showNewChatModal}
  <NewChatModal
    {isLoadingUsers}
    {userSearchResults}
    searchKeyword={userSearchQuery}
    onCancel={() => showNewChatModal = false}
    on:close={() => showNewChatModal = false}
    on:search={(e) => searchForUsers(e.detail)}
    on:createChat={(e) => initiateNewChat(e.detail)}
  />
{/if}

{#if showCreateGroupModal}
  <CreateGroupChat 
    onCancel={() => showCreateGroupModal = false}
    onSuccess={(e) => handleCreateGroupChat(e.detail)}
    on:close={() => showCreateGroupModal = false}
    on:createGroup={(e) => handleCreateGroupChat(e.detail)}
  />
{/if}

{#if showDebug}
  <DebugPanel 
    on:close={() => showDebug = false}
    on:testConnection={() => testApiConnection()}
    on:checkAuth={() => logAuthTokenInfo()}
  />
{/if}

<Toast {toasts} on:close={(e) => toastStore.removeToast(e.detail)} />


