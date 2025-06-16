<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { writable, get } from 'svelte/store';
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
  import type { Toast as ToastType } from '../interfaces/IToast';
  import { formatRelativeTime } from '../utils/date'; 
  import { getAuthToken } from '../utils/auth';
  
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
  
  // Toast notifications - managed by toastStore
  
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
  
  // WebSocket connection status monitoring (no auto-reconnect to prevent loops)
  $: {
    if ($websocketStore) {
      const isWsConnected = $websocketStore.connected;
      logger.debug(`WebSocket connection status: ${isWsConnected ? 'connected' : 'disconnected'}`);
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
  
  // Function to handle WebSocket messages
  function handleWebSocketMessage(message: ChatMessage) {
    logDebug('Received WebSocket message');
    console.log('WebSocket message details:', message);
    
    if (!message || !message.chat_id) {
      logger.warn('Invalid message format received from WebSocket');
      return;
    }
    
    // Log detailed message info for debugging
    logger.debug(`[WebSocket] Message received for chat ${message.chat_id}:`, {
      type: message.type,
      content: message.content,
      sender: message.user_id || message.sender_id,
      timestamp: message.timestamp
    });
    
    // Process system messages differently
    if (message.type === 'system') {
      // Just log system messages for now
      logger.info(`System message: ${message.content}`);
      return;
    }
    
    // Check if this message is for the selected chat
    if (selectedChat && message.chat_id === selectedChat.id) {
      // Skip messages sent by the current user (already handled by optimistic updates)
      const currentUserId = $authStore.user_id;
      const messageSenderId = message.user_id || message.sender_id;
      
      if (messageSenderId === currentUserId) {
        logger.debug('Skipping own message from WebSocket (already displayed)');
        return;
      }
      
      // Update the messages in the selected chat
      if (message.type === 'text' && message.content) {
        // Check if message already exists (to avoid duplicates)
        const messageExists = selectedChat.messages.some(msg => 
          (msg.id === message.message_id) || 
          (message.message_id && msg.id === message.message_id)
        );

        if (messageExists) {
          logger.debug('Message already exists in chat, skipping');
          return;
        }

        // Create a properly formatted message object
        const newMessage: Message = {
          id: message.message_id || `ws-${Date.now()}`,
          chat_id: message.chat_id,
          sender_id: messageSenderId || '',
          content: message.content || '',
          timestamp: typeof message.timestamp === 'string' 
            ? message.timestamp 
            : message.timestamp instanceof Date
              ? message.timestamp.toISOString()
              : new Date().toISOString(),
          is_read: false,
          is_edited: false,
          is_deleted: false,
          sender_name: message.sender_name || 'User',
          sender_avatar: message.sender_avatar
        };
        
        logger.info(`Adding new message from WebSocket to chat ${message.chat_id}:`, newMessage);
        
        // Add the message to the selected chat
        selectedChat = {
          ...selectedChat,
          messages: [...(selectedChat.messages || []), newMessage]
        };
        
        // Scroll to bottom
        setTimeout(() => {
          const messagesContainer = document.querySelector('.messages-container');
          if (messagesContainer) {
            messagesContainer.scrollTop = messagesContainer.scrollHeight;
          }
        }, 100);
      }
    }
    
    // Update the last message in the chat list for all chats
    if (message.type === 'text' && message.content) {
      // Find the chat in our list
      const chatIndex = chats.findIndex(c => c.id === message.chat_id);
      if (chatIndex >= 0) {
        // Create a properly formatted last message
        const lastMessage: LastMessage = {
          content: message.content || '',
          timestamp: typeof message.timestamp === 'string' 
            ? message.timestamp 
            : message.timestamp instanceof Date
              ? message.timestamp.toISOString()
              : new Date().toISOString(),
          sender_id: message.user_id || message.sender_id || '',
          sender_name: message.sender_name || 'User'
        };
        
        // Update the chat with the new last message
        const updatedChat = {
          ...chats[chatIndex],
          last_message: lastMessage,
          // Increment unread count if this isn't the selected chat
          unread_count: selectedChat?.id === message.chat_id 
            ? chats[chatIndex].unread_count 
            : (chats[chatIndex].unread_count || 0) + 1
        };
        
        // Move this chat to the top of the list
        const updatedChats = [
          updatedChat,
          ...chats.filter(c => c.id !== message.chat_id)
        ];
        
        // Update the chat list
        chats = updatedChats;
        
        // Also update filtered chats
        const filteredIndex = filteredChats.findIndex(c => c.id === message.chat_id);
        if (filteredIndex >= 0) {
          const updatedFilteredChat = {
            ...filteredChats[filteredIndex],
            last_message: lastMessage,
            unread_count: selectedChat?.id === message.chat_id 
              ? filteredChats[filteredIndex].unread_count 
              : (filteredChats[filteredIndex].unread_count || 0) + 1
          };
          
          // Move this chat to the top of the filtered list
          filteredChats = [
            updatedFilteredChat,
            ...filteredChats.filter(c => c.id !== message.chat_id)
          ];
        }
        
        // Play notification sound if this is not the selected chat
        if (selectedChat?.id !== message.chat_id) {
          logger.debug('Would play notification sound for new message');
        }
      } else {
        logger.warn(`Chat with ID ${message.chat_id} not found in chat list`);
      }
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
        logWarn('No chats found in response');
        console.warn('API response details:', response);
        chats = [];
        filteredChats = [];
      }
    } catch (error) {
      logError('Failed to load chats', error);
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
      logError('Failed to search for users', error);
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
      logError('Failed to create chat', error);
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
  onMount(() => {
    // Check viewport size
    checkViewport();
    window.addEventListener('resize', checkViewport);
    
    // Function to initialize everything
    const initialize = async () => {
      // Fetch user profile
      try {
        const profileData = await getProfile();
        if (profileData) {
          username = profileData.username || '';
          displayName = profileData.display_name || profileData.username || 'User';
          avatar = profileData.profile_picture_url || 'https://secure.gravatar.com/avatar/0?d=mp';
        }
      } catch (error) {
        logError('Failed to load profile', error);
      } finally {
        isLoadingProfile = false;
      }
      
      // Load chats
      await fetchChats();
    };
    
    // Start initialization
    initialize();
    
    // Register WebSocket message handler
    const unregisterHandler = websocketStore.registerMessageHandler(handleWebSocketMessage);
    
    // Also set the handler in the chatApi for backward compatibility
    setMessageHandler(handleWebSocketMessage);
    
    logger.info('Message component mounted');
    
    // Return cleanup function
    return () => {
      if (unregisterHandler) unregisterHandler();
    };
  });
  
  // Clean up when component unmounts
  onDestroy(() => {
    window.removeEventListener('resize', checkViewport);
    logger.info('Disconnecting from all WebSocket connections');
    websocketStore.disconnectAll();
  });

  // Helper function to safely format timestamps
  function safeFormatRelativeTime(timestamp: string | Date | unknown): string {
    if (typeof timestamp === 'string') {
      return formatRelativeTime(timestamp);
    } else if (timestamp instanceof Date) {
      return formatRelativeTime(timestamp.toISOString());
    } else {
      // Default to current time if invalid
      return formatRelativeTime(new Date().toISOString());
    }
  }
  
  // Helper function to ensure timestamp is a string
  function ensureStringTimestamp(timestamp: string | Date | number | unknown): string {
    if (typeof timestamp === 'string') {
      return timestamp;
    } else if (timestamp instanceof Date) {
      return timestamp.toISOString();
    } else if (typeof timestamp === 'number') {
      return new Date(timestamp).toISOString();
    } else {
      return new Date().toISOString();
    }
  }
  
  // Chat interaction functions
  async function selectChat(chat: Chat | string) {
    let chatId: string;
    
    // Handle both string ID and Chat object
    if (typeof chat === 'string') {
      chatId = chat;
      // Find the chat in our list
      const chatObj = chats.find(c => c.id === chatId);
      if (!chatObj) {
        logger.error(`Chat with ID ${chatId} not found in chats list`);
        toastStore.showToast(`Chat not found. Please try again.`, 'error');
        return;
      }
      chat = chatObj;
    } else {
      chatId = chat.id;
    }
    
    logger.info(`Selecting chat: ${chatId}`);
    
    // Validate chat ID format
    if (!chatId || !/^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i.test(chatId)) {
      logger.error(`Invalid chat ID format: ${chatId}`);
      toastStore.showToast(`Invalid chat ID format: ${chatId}. Please try again or contact support.`, 'error');
      return;
    }
    
    selectedChat = { ...chat, messages: [] };
    
    // On mobile, hide the chat list
    if (isMobile) {
      showMobileMenu = false;
      handleMobileNavigation('showChat');
    }
    
    // Fetch messages for the selected chat
    isLoadingMessages = true;
    try {
      logger.debug(`Fetching messages for chat ${chatId}`);
      const response = await listMessages(chatId);
      
      if (response && response.messages) {
        // Sort messages by timestamp to ensure correct order (newest last)
        const sortedMessages = [...response.messages].sort((a, b) => {
          const timeA = new Date(a.timestamp).getTime();
          const timeB = new Date(b.timestamp).getTime();
          return timeA - timeB;
        });
        
        // Process messages to add any missing properties
        const processedMessages = sortedMessages.map(msg => ({
          ...msg,
          sender_name: msg.sender_name || 'User',
          sender_avatar: msg.sender_avatar || null,
          timestamp: ensureStringTimestamp(msg.timestamp)
        }));
        
        selectedChat = {
          ...selectedChat,
          messages: processedMessages
        };
        
        logger.info(`Loaded ${processedMessages.length} messages for chat ${chatId}`);
          
        // Scroll to bottom of messages
        setTimeout(() => {
          const messagesContainer = document.querySelector('.messages-container');
          if (messagesContainer) {
            messagesContainer.scrollTop = messagesContainer.scrollHeight;
          }
        }, 100);
      } else {
        logWarn(`No messages found for chat ${chatId}`);
        selectedChat = {
          ...selectedChat,
          messages: []
        };
      }
    } catch (error) {
      logError(`Error loading messages for chat ${chatId}`, error);
      toastStore.showToast({ message: 'Failed to load messages', type: 'error' });
      selectedChat = {
        ...selectedChat,
        messages: []
      };
    } finally {
      isLoadingMessages = false;
    }
    
    // Connect to WebSocket for this chat
    try {
      // Check if already connected to this chat
      const isConnected = (websocketStore as any).isConnected(chatId);
      if (!isConnected) {
        logger.info(`Connecting to WebSocket for chat ${chatId}`);
        (websocketStore as any).connect(chatId);
      } else {
        logger.debug(`Already connected to WebSocket for chat ${chatId}`);
      }
    } catch (error) {
      logError(`Error connecting to WebSocket for chat ${chatId}`, error);
      toastStore.showToast({ message: 'Could not establish real-time connection', type: 'warning' });
    }
    
    // Mark chat as read by resetting unread count
    chats = chats.map(c => {
      if (c.id === chatId) {
        return { ...c, unread_count: 0 };
      }
      return c;
    }) as Chat[];
    
    // Fix the filtered chats assignment
    filteredChats = [
      ...(filteredChats.filter(c => c.id === chatId)),
      ...(filteredChats.filter(c => c.id !== chatId))
    ].map(chat => ({
      ...chat,
      // Ensure that last_message.timestamp is always a string
      last_message: chat.last_message ? {
        ...chat.last_message,
        timestamp: typeof chat.last_message.timestamp === 'string'
          ? chat.last_message.timestamp
          : new Date(chat.last_message.timestamp).toISOString()
      } : undefined
    })) as Chat[];
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
    // Group chats should use their name
    if (chat.type === 'group' && chat.name && chat.name.trim() !== '') {
      return chat.name;
    }
    
    // For individual chats, show the other participant's name
    if (chat.participants && chat.participants.length > 0) {
      // Use only id for filtering as that's in the Participant type
      const otherParticipants = chat.participants.filter(p => p.id !== $authStore.user_id);
      
      if (otherParticipants.length > 0) {
        const participant = otherParticipants[0];
        return participant.display_name || participant.username || 'Unknown User';
      }
    }
    
    // Fallback to chat name or generic name
    return chat.name && chat.name.trim() !== '' ? chat.name : 'Chat';
  }
  
  async function handleSearch() {
    if (!searchQuery || searchQuery.trim() === '') {
      filteredChats = [...chats];
      return;
    }
    
    const query = searchQuery.toLowerCase().trim();
    filteredChats = chats.filter(chat => {
      // Search in chat name
      const chatName = getChatDisplayName(chat).toLowerCase();
      if (chatName.includes(query)) return true;
      
      // Search in last message
      if (chat.last_message && chat.last_message.content) {
        const messageContent = chat.last_message.content.toLowerCase();
        if (messageContent.includes(query)) return true;
      }
      
      // Search in participants
      if (chat.participants) {
        for (const participant of chat.participants) {
          const name = participant.display_name || participant.username || '';
          if (name.toLowerCase().includes(query)) return true;
        }
      }
      
      return false;
    });
  }
  
  // Function to send a message
  async function sendMessage(content: string) {
    if (!content || !content.trim() || !selectedChat) {
      return;
    }
    
    try {
      // Generate a unique temporary ID for this message
      const tempMessageId = `temp-${Date.now()}`;
      
      // Trim content and prevent empty messages
      content = content.trim();
      newMessage = '';
      
      // Create message object
      const message: Message = {
        id: tempMessageId,
        chat_id: selectedChat.id,
        sender_id: $authStore.user_id || '',
        content: content,
        timestamp: new Date().toISOString(),
        is_read: false,
        is_edited: false,
        is_deleted: false,
        sender_name: displayName || 'You',
        sender_avatar: avatar,
        is_local: true
      };
      
      // Optimistically add message to UI
      selectedChat = {
        ...selectedChat,
        messages: [...selectedChat.messages, message]
      };
      
      // Scroll to bottom after message is added
      setTimeout(() => {
        const messagesContainer = document.querySelector('.messages-container');
        if (messagesContainer) {
          messagesContainer.scrollTop = messagesContainer.scrollHeight;
        }
      }, 50);
      
      // Create a last message object for the chat list
      const newLastMessage: LastMessage = {
        content: content,
        timestamp: new Date().toISOString(),
        sender_id: $authStore.user_id || '',
        sender_name: displayName || 'You'
      };
      
      // Update chat list with the new message
      chats = chats.map(chat => {
        if (chat.id === selectedChat?.id) {
          return {
            ...chat,
            last_message: newLastMessage
          };
        }
        return chat;
      }) as Chat[];
        
      // Move the active chat to the top
      const activeChatId = selectedChat?.id;
      if (activeChatId) {
      const activeChat = chats.find(c => c.id === activeChatId);
      if (activeChat) {
          // Remove the active chat from the array
          const otherChats = chats.filter(c => c.id !== activeChatId);
          // Add it back at the beginning
          chats = [activeChat, ...otherChats];
          
          // Do the same for filtered chats
          const filteredActiveChat = filteredChats.find(c => c.id === activeChatId);
          if (filteredActiveChat) {
            const otherFilteredChats = filteredChats.filter(c => c.id !== activeChatId);
        filteredChats = [
              {
                ...filteredActiveChat,
                last_message: newLastMessage
              },
              ...otherFilteredChats
            ];
          }
        }
      }

      // Send message via API
      const messageData = {
        content: content,
        message_id: tempMessageId
      };
      
      // Log the API call attempt
      logInfo(`Sending message to chat ${selectedChat?.id} via API`);
      
      // First try to send via WebSocket for immediate real-time delivery
      const wsMessage: ChatMessage = {
        type: 'text',
        content: content,
        chat_id: selectedChat?.id || '',
        user_id: $authStore.user_id || '',
        sender_id: $authStore.user_id || '',
        sender_name: displayName || username || 'User',
        sender_avatar: avatar,
        message_id: tempMessageId,
        timestamp: new Date().toISOString()
      };
      
      // Send via WebSocket first for real-time delivery
      websocketStore.sendMessage(selectedChat?.id || '', wsMessage);
      logger.info(`Message sent via WebSocket to chat ${selectedChat?.id}`);
      
      // Then send via API for persistence
      try {
        const result = await apiSendMessage(selectedChat?.id || '', messageData);
        logInfo('Message sent successfully via API:', result);
        
        // Update the message to mark it as confirmed by the server
        if (selectedChat) {
          selectedChat = {
            ...selectedChat,
            messages: selectedChat.messages.map(msg => 
              msg.id === tempMessageId 
                ? { 
                    ...msg, 
                    is_local: false,
                    id: result?.message?.id || result?.message_id || msg.id
                  } 
                : msg
            )
          };
        }
      } catch (apiError) {
        logError('Failed to send message via API', apiError);
        toastStore.showToast({
          message: 'Network issue detected. Message may not be delivered.',
          type: 'error'
        });
        
        // Mark the message as potentially failed
        if (selectedChat) {
          selectedChat = {
            ...selectedChat,
            messages: selectedChat.messages.map(msg => 
              msg.id === tempMessageId 
                ? { ...msg, failed: true } 
                : msg
            )
          };
        }
      }
    } catch (error: unknown) {
      const errorMessage = error instanceof Error ? error.message : 'Unknown error';
      logError('Error sending message', errorMessage);
      toastStore.showToast({
        message: `Error sending message: ${errorMessage}`,
        type: 'error'
      });
    }
  }
  
  // Fix unsendMessage function to ensure timestamps are always strings
  async function unsendMessage(messageId: string) {
    // Logic to unsend/delete a message
    if (!selectedChat) return;
    
    // Find the message
    const message = selectedChat.messages.find(m => m.id === messageId);
    if (!message || message.sender_id !== $authStore.user_id) return;
    
    try {
      // Optimistically update UI
      selectedChat = {
        ...selectedChat,
        messages: selectedChat.messages.map(msg => 
          msg.id === messageId ? { ...msg, is_deleted: true, content: 'Message deleted' } : msg
        )
      };
      
      // Call API to unsend
      await apiUnsendMessage(selectedChat.id, messageId);
      
      // Update the chat if the deleted message was the last message
      const lastMessage = selectedChat.last_message;
      if (lastMessage && lastMessage.content === message.content) {
        // Find the previous message to use as the new last message
        const previousMessages = selectedChat.messages
          .filter(m => !m.is_deleted && m.id !== messageId)
          .sort((a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime());
        
        const newLastMessage = previousMessages[0];
          
        if (newLastMessage) {
          // Update the chat in the list
          chats = chats.map(chat => {
            if (chat.id === selectedChat?.id) {
              return {
                ...chat,
                last_message: {
                  content: newLastMessage.content,
                  timestamp: ensureStringTimestamp(newLastMessage.timestamp),
                  sender_id: newLastMessage.sender_id,
                  sender_name: newLastMessage.sender_name || ''
                }
              };
            }
            return chat;
          }) as Chat[];
          
          // Also update filtered chats
          filteredChats = filteredChats.map(chat => {
            if (chat.id === selectedChat?.id) {
              return {
                ...chat,
                last_message: {
                  content: newLastMessage.content,
                  timestamp: ensureStringTimestamp(newLastMessage.timestamp),
                  sender_id: newLastMessage.sender_id,
                  sender_name: newLastMessage.sender_name || ''
                }
              };
            }
            return chat;
          }) as Chat[];
        }
        }
    } catch (error: unknown) {
      const errorMessage = error instanceof Error ? error.message : 'Unknown error';
      logError('Failed to unsend message', errorMessage);
      toastStore.showToast({ message: `Failed to unsend message: ${errorMessage}`, type: 'error' });
      
      // Revert the optimistic update on error
      selectedChat = {
        ...selectedChat,
        messages: selectedChat.messages.map(msg => 
          msg.id === messageId ? { ...message } : msg
        )
      };
    }
  }
  
  /**
   * Handle creating a group chat
   */
  async function handleCreateGroupChat(data: any) {
    try {
      isLoadingChats = true;
      
      // Handle different data formats from different sources
      let chatData: { type: string; name: string; participants: string[] };
      
      if (data && data.chat) {
        // Handle format from onSuccess event: { chat: { ... } }
        chatData = {
          type: 'group',
          name: data.chat.name || 'New Group',
          participants: (data.chat.participants || []).map((p: any) => p.id || p)
        };
      } else if (data && data.name && data.participants) {
        // Handle direct format: { name: string; participants: string[] }
        chatData = {
          type: 'group',
          name: data.name,
          participants: data.participants
        };
    } else {
        throw new Error('Invalid group chat data format');
      }
      
      const response = await createChat(chatData);
      
      if (response && response.chat_id) {
        showCreateGroupModal = false;
        await fetchChats();
        selectChat(response.chat_id);
        toastStore.showToast(`Group chat "${chatData.name}" created successfully`, 'success');
      }
    } catch (error) {
      logError('Failed to create group chat', error);
    } finally {
      isLoadingChats = false;
    }
  }

  // WebSocket connection management
  const handleReconnect = () => {
    console.log('[WebSocket] Attempting to reconnect...');
    
    // Reconnect to the selected chat
    if (selectedChat) {
      console.log(`[WebSocket] Reconnecting to selected chat: ${selectedChat.id}`);
      websocketStore.connect(selectedChat.id);
      
      // Reconnect to recent chats
      const recentChats = chats.slice(0, 5); // Reconnect to 5 most recent chats
      recentChats.forEach(chat => {
        if (selectedChat && chat.id !== selectedChat.id) {
          console.log(`[WebSocket] Reconnecting to additional chat: ${chat.id}`);
          websocketStore.connect(chat.id);
        }
      });
    } else {
      // If no selected chat, just reconnect to recent chats
      const recentChats = chats.slice(0, 5);
      recentChats.forEach(chat => {
        console.log(`[WebSocket] Reconnecting to chat: ${chat.id}`);
        websocketStore.connect(chat.id);
      });
    }
  };





  // Function to test WebSocket connection
  const testWebSocketConnection = () => {
    if (selectedChat) {
      console.log(`[WebSocket] Testing connection for chat: ${selectedChat.id}`);
      
      // Get the WebSocket URL
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
      const hostname = window.location.hostname;
      const port = '8083';
      const token = getAuthToken();
      
      if (!token) {
        console.error('[WebSocket] No authentication token available');
        return;
      }
      
      // Test both URL formats to see which one works
      const url1 = `${protocol}//${hostname}:${port}/api/v1/chats/${selectedChat.id}/ws?token=${encodeURIComponent(token)}`;
      const url2 = `${protocol}//${hostname}:${port}/ws/chat/${selectedChat.id}?token=${encodeURIComponent(token)}`;
      
      console.log(`[WebSocket] Testing URL 1: ${url1}`);
      const ws1 = new WebSocket(url1);
      
      ws1.onopen = () => {
        console.log('[WebSocket] URL 1 connection successful!');
        ws1.send(JSON.stringify({type: 'test', content: 'Test message from URL 1'}));
      };
      
      ws1.onerror = (error) => {
        console.error('[WebSocket] URL 1 connection failed:', error);
        
        // Try the second URL format
        console.log(`[WebSocket] Testing URL 2: ${url2}`);
        const ws2 = new WebSocket(url2);
        
        ws2.onopen = () => {
          console.log('[WebSocket] URL 2 connection successful!');
          ws2.send(JSON.stringify({type: 'test', content: 'Test message from URL 2'}));
        };
        
        ws2.onerror = (error) => {
          console.error('[WebSocket] URL 2 connection failed:', error);
          console.error('[WebSocket] Both connection attempts failed');
        };
      };
    }
  };

  // Function to fetch messages for a chat
  async function fetchMessages(chatId: string) {
    if (!chatId) return;
    
    isLoadingMessages = true;
    
    try {
      const response = await chatApi.listMessages(chatId);
      
      if (response && response.messages) {
        // Update the selected chat with the messages
        if (selectedChat && selectedChat.id === chatId) {
          selectedChat = {
            ...selectedChat,
            messages: response.messages
          };
        }
      }
    } catch (error) {
      logError('Failed to load messages', error);
    } finally {
      isLoadingMessages = false;
    }
  }
  
  // Function to mark a chat as read
  async function markChatAsRead(chatId: string) {
    if (!chatId) return;
    
    try {
      // Find the chat in our list
      const chatIndex = chats.findIndex(c => c.id === chatId);
      if (chatIndex >= 0) {
        // Update the chat's unread count locally
        chats[chatIndex].unread_count = 0;
        
        // Also update the filtered chats
        const filteredIndex = filteredChats.findIndex(c => c.id === chatId);
        if (filteredIndex >= 0) {
          filteredChats[filteredIndex].unread_count = 0;
        }
        
        // Trigger a UI update
        chats = [...chats];
        filteredChats = [...filteredChats];
      }
      
      // TODO: Implement API call to mark chat as read when endpoint is available
      // For now, we're just updating the UI
    } catch (error) {
      logError('Failed to mark chat as read', error);
    }
  }

  // Fix the log function to handle the correct number of arguments
  function logDebug(message: string) {
    logger.debug(message);
  }
  
  function logInfo(message: string) {
    logger.info(message);
  }
  
  function logWarn(message: string) {
    logger.warn(message);
  }
  
  async function logError(message: string, error?: any) {
    if (error) {
      console.error(message, error);
      logger.error(message);
    } else {
      logger.error(message);
    }
  }
  
  // Helper function to ensure we have valid arrays for chats and filteredChats
  function ensureValidChatArrays() {
    if (!Array.isArray(chats)) {
      chats = [];
    }
    
    if (!Array.isArray(filteredChats)) {
      filteredChats = [];
    }
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

  /* Connection status styles moved to corresponding components or removed */
  
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

  /* Connection status styles */
  .connection-status-container {
    padding: 5px 10px;
    background-color: var(--bg-secondary);
    border-bottom: 1px solid var(--border-color);
    font-size: 0.8rem;
  }

  .connection-status {
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 5px 0;
  }

  .status-connected {
    color: #10b981;
    display: flex;
    align-items: center;
    gap: 5px;
  }

  .status-connecting {
    color: #f59e0b;
    display: flex;
    align-items: center;
    gap: 5px;
  }

  .status-disconnected {
    color: #ef4444;
    display: flex;
    align-items: center;
    gap: 5px;
  }

  .status-icon {
    font-size: 10px;
    line-height: 1;
  }

  .status-error {
    margin-top: 5px;
    padding: 5px 8px;
    background-color: rgba(239, 68, 68, 0.1);
    border-radius: 4px;
    color: #ef4444;
    display: flex;
    align-items: center;
    gap: 5px;
  }

  .error-icon {
    font-size: 12px;
  }

  .error-text {
    flex: 1;
    font-size: 0.8rem;
  }

  .error-dismiss {
    background: none;
    border: none;
    color: #ef4444;
    cursor: pointer;
    font-size: 16px;
    padding: 0;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .reconnect-button {
    background-color: transparent;
    border: 1px solid currentColor;
    color: inherit;
    border-radius: 4px;
    padding: 2px 8px;
    font-size: 0.7rem;
    cursor: pointer;
    margin-left: 8px;
  }

  .reconnect-button:hover {
    background-color: rgba(239, 68, 68, 0.1);
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
    <div class="connection-status-container">
      <div class="connection-status">
        {#if selectedChat}
          {#if $websocketStore.connectionStatus[selectedChat.id] === 'connected'}
            <div class="status-connected">
              <span class="status-icon">●</span>
              <span class="status-text">Connected</span>
            </div>
          {:else if $websocketStore.connectionStatus[selectedChat.id] === 'connecting'}
            <div class="status-connecting">
              <span class="status-icon">◌</span>
              <span class="status-text">Connecting...</span>
            </div>
          {:else if $websocketStore.connectionStatus[selectedChat.id] === 'disconnected' || $websocketStore.connectionStatus[selectedChat.id] === 'error'}
            <div class="status-disconnected">
              <span class="status-icon">○</span>
              <span class="status-text">Disconnected</span>
              <button class="reconnect-button" on:click={handleReconnect}>
                Reconnect Now
              </button>
              <button class="test-connection-button" on:click={testWebSocketConnection}>
                Test Connection
              </button>
            </div>
          {/if}
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
                      <div class="timestamp">{safeFormatRelativeTime(chat.last_message.timestamp)}</div>
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
                      <span class="timestamp">{safeFormatRelativeTime(message.timestamp)}</span>
                        
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

<Toast on:close={(e) => toastStore.removeToast(e.detail)} />


