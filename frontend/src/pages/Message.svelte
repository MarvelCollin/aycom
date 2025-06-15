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
    setMessageHandler 
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
  
  // Handle attachment selection - placeholder function for now
  function handleAttachment(type: 'image' | 'gif') {
    // Implementation can be added later
    logger.debug(`Attachment selected: ${type}`);
    // For now, just show a message
    toastStore.showToast(`${type} attachment feature coming soon`, 'info');
  }
  
  // Group chat modal state
  let showGroupChatModal = false;
  let showNewChatModal = false;

  // Mobile detection
  let isMobile = false;
  let showMobileMenu = false;
  
  // Function to check viewport size and set mobile state
  function checkViewport() {
    isMobile = window.innerWidth < 768;
  }
  
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
  async function selectChat(chat: Chat) {
    logger.info(`Selecting chat: ${chat.id}`);
    
    // Validate chat ID format
    if (!chat.id || !/^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i.test(chat.id)) {
      logger.error(`Invalid chat ID format: ${chat.id}`);
      toastStore.showToast(`Invalid chat ID format: ${chat.id}. Please try again or contact support.`, 'error');
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
      logger.debug(`Fetching messages for chat ${chat.id}`);
      const response = await listMessages(chat.id);
      
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
        
        logger.info(`Loaded ${processedMessages.length} messages for chat ${chat.id}`);
          
        // Scroll to bottom of messages
        setTimeout(() => {
          const messagesContainer = document.querySelector('.messages-container');
          if (messagesContainer) {
            messagesContainer.scrollTop = messagesContainer.scrollHeight;
          }
        }, 100);
      } else {
        logger.warn(`No messages found for chat ${chat.id}`);
        selectedChat = {
          ...selectedChat,
          messages: []
        };
      }
    } catch (error) {
      logger.error(`Error loading messages for chat ${chat.id}:`, error);
      toastStore.showToast('Failed to load messages', 'error');
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
      const isConnected = websocketStore.isConnected(chat.id);
      if (!isConnected) {
        logger.info(`Connecting to WebSocket for chat ${chat.id}`);
        websocketStore.connect(chat.id);
      } else {
        logger.debug(`Already connected to WebSocket for chat ${chat.id}`);
      }
    } catch (error) {
      logger.error(`Error connecting to WebSocket for chat ${chat.id}:`, error);
      toastStore.showToast('Could not establish real-time connection', 'warning');
    }
    
    // Mark chat as read by resetting unread count
    chats = chats.map(c => {
      if (c.id === chat.id) {
        return { ...c, unread_count: 0 };
      }
      return c;
    });
    
    // Fix the filtered chats assignment
    filteredChats = [
      ...(filteredChats.filter(c => c.id === chat.id)),
      ...(filteredChats.filter(c => c.id !== chat.id))
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
  
  async function startChatWithUser(user: StandardUser) {
    // Logic to start a new chat with a user
    logger.debug(`Starting new chat with user: ${user.username}`);
    
    if (!user || !user.id) {
      logger.error('Cannot start chat: Invalid user data');
      toastStore.showToast('Invalid user data. Please try again.', 'error');
      return;
    }
    
    // Check if chat already exists
    const existingChat = chats.find(chat => 
      chat.type === 'individual' && 
      chat.participants.some(p => p.id === user.id)
    );
    
    if (existingChat) {
      // Chat already exists, just select it
      selectChat(existingChat);
      showNewChatModal = false; // Close the modal
      return;
    }
    
    try {
      // Create new chat
      const response = await createChat({
        type: 'individual',
        participants: [user.id],
        name: user.display_name || user.username
      });
      
      // The backend returns {chat: {...}}, so we need to extract the chat object
      const chatData = response.chat || response;
      
      if (chatData && chatData.id) {
        // Create a new chat object
        const newChat: Chat = {
          id: chatData.id,
          type: 'individual',
          name: user.display_name || user.username,
          avatar: user.avatar || null,
          participants: [{
            id: user.id,
            username: user.username,
            display_name: user.display_name || user.username,
            avatar: user.avatar || null,
            is_verified: user.is_verified
          }],
          messages: [],
          unread_count: 0,
          profile_picture_url: user.avatar || null,
          created_at: chatData.created_at || new Date().toISOString(),
          updated_at: chatData.updated_at || new Date().toISOString()
        };
        
        // Add to chats and select it
        chats = [newChat, ...chats];
        filteredChats = [newChat, ...filteredChats];
        selectChat(newChat);
        
        // Close the modal
        showNewChatModal = false;
      } else {
        logger.error('Failed to create chat: Invalid response', response);
        toastStore.showToast('Failed to create chat. Please try again.', 'error');
      }
    } catch (error) {
      logger.error('Failed to create chat', error);
      toastStore.showToast('Failed to create chat', 'error');
    }
  }
  
  async function createGroupChat(name: string, participants: StandardUser[]) {
    if (!name || !participants || participants.length === 0) {
      logger.error('Cannot create group chat: Missing required data');
      toastStore.showToast('Group name and participants are required', 'error');
      return;
    }
    
    logger.debug(`Creating group chat: ${name} with ${participants.length} participants`);
    
    try {
      // Create new group chat
      const response = await createChat({
        type: 'group',
        name: name,
        participants: participants.map(p => p.id)
      });
      
      // The backend returns {chat: {...}}, so we need to extract the chat object
      const chatData = response.chat || response;
      
      if (chatData && chatData.id) {
        // Create a new chat object
        const newChat: Chat = {
          id: chatData.id,
          type: 'group',
          name: name,
          avatar: null,
          participants: participants.map(p => ({
            id: p.id,
            username: p.username,
            display_name: p.display_name || p.username,
            avatar: p.avatar || null,
            is_verified: p.is_verified
          })),
          messages: [],
          unread_count: 0,
          profile_picture_url: null,
          created_at: chatData.created_at || new Date().toISOString(),
          updated_at: chatData.updated_at || new Date().toISOString()
        };
        
        // Add to chats and select it
        chats = [newChat, ...chats];
        filteredChats = [newChat, ...filteredChats];
        selectChat(newChat);
        
        // Close the modal
        showGroupChatModal = false;
      } else {
        logger.error('Failed to create group chat: Invalid response', response);
        toastStore.showToast('Failed to create group chat. Please try again.', 'error');
      }
    } catch (error: unknown) {
      const errorMessage = error instanceof Error ? error.message : 'Unknown error';
      logger.error('Failed to create group chat', error);
      toastStore.showToast('Failed to create group chat: ' + errorMessage, 'error');
    }
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
  
  // Message handling functions
  async function sendMessage(content: string) {
    if (!content.trim() || !selectedChat) return;

    // Clear input after sending
    newMessage = '';
    selectedAttachments = [];

    // Generate a temporary message ID
    const tempMessageId = `temp-${Date.now()}`;

    // Create a temporary message object
    const tempMessage: Message = {
      id: tempMessageId,
      chat_id: selectedChat?.id || '',
      content,
      timestamp: new Date().toISOString(),
      sender_id: $authStore.user_id || '',
      sender_name: displayName,
      sender_avatar: avatar,
      is_read: false,
      is_deleted: false,
      is_edited: false,
      attachments: selectedAttachments.length > 0 ? [...selectedAttachments] : undefined
    };

    try {
      // Add message to UI immediately (optimistic update)
      if (selectedChat) {
        selectedChat = {
          ...selectedChat,
          messages: [...selectedChat.messages, tempMessage]
        };

        // Scroll to bottom
        setTimeout(() => {
          const messagesContainer = document.querySelector('.messages-container');
          if (messagesContainer) {
            messagesContainer.scrollTop = messagesContainer.scrollHeight;
          }
        }, 100);
      }

      // Update chat in the list with last message
      const newLastMessage: LastMessage = {
        content,
        timestamp: ensureStringTimestamp(new Date().toISOString()),
        sender_id: $authStore.user_id || '',
        sender_name: displayName
      };

      // Fix the issue in chat.map where we're updating the last_message
      chats = chats.map(chat => {
        if (chat.id === selectedChat?.id) {
          return {
            ...chat,
            last_message: newLastMessage
          };
        }
        return chat;
      }) as Chat[];

      // Update chats and move the active chat to top
      const activeChatId = selectedChat?.id;
      const activeChat = chats.find(c => c.id === activeChatId);
      if (activeChat) {
        chats = [
          activeChat,
          ...chats.filter(c => c.id !== activeChatId)
        ] as Chat[];  // Type assertion
        
        // Fix the filtered chats assignment
        filteredChats = [
          ...(filteredChats.filter(c => c.id === activeChatId)),
          ...(filteredChats.filter(c => c.id !== activeChatId))
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

      // Send message to API
      // ...rest of function remains the same
    } catch (error: unknown) {
      const errorMessage = error instanceof Error ? error.message : 'Unknown error';
      logger.error('Error sending message:', errorMessage);
      toastStore.showToast(`Error sending message: ${errorMessage}`, 'error');
    }
  }
  
  // Add a helper function to ensure timestamp is always a string
  function ensureStringTimestamp(timestamp: string | number | Date): string {
    if (timestamp instanceof Date) {
      return timestamp.toISOString();
    } else if (typeof timestamp === 'number') {
      return new Date(timestamp).toISOString();
    }
    return timestamp;
  }
  
  // Update the map operation in unsendMessage function
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
      logger.error('Failed to unsend message', error);
      toastStore.showToast(`Failed to unsend message: ${errorMessage}`, 'error');
      
      // Revert the optimistic update on error
      selectedChat = {
        ...selectedChat,
        messages: selectedChat.messages.map(msg => 
          msg.id === messageId ? { ...message } : msg
        )
      };
    }
  }
  
  async function handleSearch() {
    // Filter existing chats
    if (searchQuery && searchQuery !== 'new') {
      filteredChats = chats.filter(chat => {
        const chatName = getChatDisplayName(chat).toLowerCase();
        const participants = chat.participants.map(p => 
          (p.username + p.display_name).toLowerCase()
        ).join(' ');
        
        return (
          chatName.includes(searchQuery.toLowerCase()) ||
          participants.includes(searchQuery.toLowerCase())
        );
      });
    } else {
      filteredChats = [...chats];
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

    // Move the updated chat to the top of the list
    const updatedChat = chats.find(chat => chat.id === message.chat_id);
    if (updatedChat) {
      chats = [
        updatedChat,
        ...chats.filter(chat => chat.id !== message.chat_id)
      ] as Chat[];  // Type assertion
      
      // Also update filtered chats
      filteredChats = [
        ...filteredChats.filter(chat => chat.id === message.chat_id),
        ...filteredChats.filter(chat => chat.id !== message.chat_id)
      ] as Chat[];  // Type assertion
    }

    // Show notification logic remains the same
  }
  
  // Initialize WebSocket connections for all chats
  async function initializeWebSocketConnections() {
    if (!$authStore.is_authenticated) return;
    
    try {
      logger.info('Initializing WebSocket connections for all chats');
      
      // Register the WebSocket message handler
      setMessageHandler(handleWebSocketMessage);
      
      // Get all chats
      const fetchedChats = await fetchChats();
      
      // Connect to WebSockets for all chats
      if (fetchedChats && fetchedChats.length > 0) {
        logger.info(`Connecting to WebSockets for ${fetchedChats.length} chats`);
        
        // Connect to each chat's WebSocket
        for (const chat of fetchedChats) {
          try {
            if (!websocketStore.isConnected(chat.id)) {
              websocketStore.connect(chat.id);
              logger.debug(`Connected to WebSocket for chat ${chat.id}`);
            }
          } catch (err) {
            const errorMessage = err instanceof Error ? err.message : String(err);
            logger.error(`Error connecting to WebSocket for chat ${chat.id}: ${errorMessage}`);
          }
        }
      }
      
      logger.info('WebSocket connections initialized');
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : String(err);
      logger.error(`Error initializing WebSocket connections: ${errorMessage}`);
    }
  }

  // Update onMount to use the new initialization function
  onMount(() => {
    // Check if we're on mobile
    checkViewport();
    window.addEventListener('resize', checkViewport);
    
    // Run API connection test first
    testApiConfig();
    
    // Initialize user profile and WebSocket connections
    if ($authStore && $authStore.is_authenticated) {
      fetchUserProfile();
      initializeWebSocketConnections();
    }
    
    return () => {
      window.removeEventListener('resize', checkViewport);
      
      // Disconnect WebSocket connections when component unmounts
      websocketStore.disconnectAll();
    };
  });
  
  onDestroy(() => {
    // Clean up WebSocket connections when component is destroyed
    websocketStore.disconnectAll();
  });
  
  // Test API configuration and connectivity
  async function testApiConfig() {
    try {
      logger.info('Testing API connection...');
      
      // Check authentication token first
      console.log('[Message] Checking authentication token...');
      const tokenInfo = logAuthTokenInfo();
      console.log('[Message] Auth token check result:', tokenInfo);
      
      if (!tokenInfo.success || tokenInfo.isExpired) {
        logger.error('Authentication token issue detected:', tokenInfo);
        toastStore.showToast(`Authentication error: ${tokenInfo.error || 'Token expired or invalid'}. Please log in again.`, 'error');
      }
      
      // Test API connection
      const result = await testApiConnection();
      logger.info('API connection test result:', result);
      
      if (!result.success) {
        logger.error('API connection test failed:', result.error);
        toastStore.showToast(`API connection error: ${result.error}. This may affect chat functionality.`, 'error');
      }
    } catch (error) {
      logger.error('Error running API connection test:', error);
    }
  }
  
  function toggleMobileMenu() {
    showMobileMenu = !showMobileMenu;
  }
  
  // Add a function to handle mobile navigation between chat list and chat content
  function handleMobileNavigation(action: 'showList' | 'showChat') {
    if (!isMobile) return;
    
    if (action === 'showList') {
      // Hide the selected chat view and show the chat list
      selectedChat = null;
    } else if (action === 'showChat' && selectedChat) {
      // Hide the chat list and show the selected chat
      // This happens automatically due to our CSS classes
    }
  }
  
  // Call this function to fetch profile data
  async function fetchUserProfile() {
    if (!$authStore.is_authenticated) return;
    
    isLoadingProfile = true;
    try {
      const response = await getProfile();
      if (response) {
        username = response.username || '';
        displayName = response.display_name || username;
        avatar = response.avatar_url || 'https://secure.gravatar.com/avatar/0?d=mp';
      }
    } catch (error) {
      logger.error('Failed to fetch profile', error);
      toastStore.showToast('Failed to load profile', 'error');
    } finally {
      isLoadingProfile = false;
    }
  }
  
  // Update fetchChats to return the fetched chats
  async function fetchChats() {
    if (!$authStore.is_authenticated) return [];
    
    isLoadingChats = true;
    try {
      logger.info('Fetching chats from API');
      const response = await getChatHistoryList();
      
      logger.info('API response received:', { 
        hasData: !!response,
        hasChats: response && 'chats' in response || (response && response.data && 'chats' in response.data),
        isChatsArray: response && 'chats' in response && Array.isArray(response.chats) || 
                     (response && response.data && 'chats' in response.data && Array.isArray(response.data.chats)),
        chatsLength: (response?.chats?.length || (response?.data?.chats?.length)) || 0,
        responseKeys: response ? Object.keys(response) : []
      });
      
      let rawChats: any[] = [];
      
      if (!response || (typeof response === 'object' && Object.keys(response).length === 0)) {
        // Empty response
        logger.info('API returned empty response, no chats available');
        rawChats = [];
      }
      else if (response && 'chats' in response && Array.isArray(response.chats)) {
        // Standard API response with chats property
        logger.info(`Processing ${response.chats.length} chats from API`);
        rawChats = response.chats;
      } 
      else if (response && response.data && 'chats' in response.data && Array.isArray(response.data.chats)) {
        // Response with data.chats structure
        logger.info(`Processing ${response.data.chats.length} chats from API data.chats`);
        rawChats = response.data.chats;
      }
      else if (Array.isArray(response)) {
        // API returns direct array
        logger.info(`Processing ${response.length} chats from direct array`);
        rawChats = response;
      } else if (response && typeof response === 'object') {
        // Try to find any array that might contain chats
        logger.warn('API returned unknown response structure:', response);
        const possibleChatArrays = Object.entries(response)
          .filter(([_, value]) => Array.isArray(value) && value.length > 0)
          .map(([key, value]) => ({ key, value: value as any[] }));
        
        if (possibleChatArrays.length > 0) {
          // Use the first array found
          const firstArray = possibleChatArrays[0].value;
          logger.info(`Using ${possibleChatArrays[0].key} as chat array with ${firstArray.length} items`);
          rawChats = firstArray;
        } else {
          logger.warn('No usable arrays found in response');
          rawChats = [];
        }
      } else {
        logger.warn('Unrecognized API response format, no chats available');
        rawChats = [];
      }
      
      // Map raw chats to client format
      const mappedChats = mapApiChatsToClientFormat(rawChats);
      
      // Filter out duplicate non-group chats with the same name
      const uniqueChats: Chat[] = [];
      const nameMap: Record<string, Chat> = {};
      
      // First pass: collect the most recent chat for each name
      mappedChats.forEach(chat => {
        // For group chats, always keep them
        if (chat.type === 'group') {
          uniqueChats.push(chat);
          return;
        }
        
        // For non-group chats with a name, keep only the most recent one
        if (chat.name && chat.name.trim() !== '') {
          const existingChat = nameMap[chat.name];
          if (!existingChat || 
              (chat.last_message && existingChat.last_message && 
               new Date(chat.last_message.timestamp) > new Date(existingChat.last_message.timestamp)) || 
              (!existingChat.last_message && chat.last_message) ||
              (!existingChat.last_message && !chat.last_message && new Date(chat.updated_at || chat.created_at) > new Date(existingChat.updated_at || existingChat.created_at))) {
            nameMap[chat.name] = chat;
          }
        } else {
          // Chats without names are always kept
          uniqueChats.push(chat);
        }
      });
      
      // Add the most recent chat for each name
      Object.values(nameMap).forEach(chat => {
        uniqueChats.push(chat);
      });
      
      // Sort chats by last message timestamp (newest first)
      uniqueChats.sort((a, b) => {
        const timeA = a.last_message?.timestamp 
          ? new Date(a.last_message.timestamp).getTime() 
          : new Date(a.updated_at || a.created_at).getTime();
        const timeB = b.last_message?.timestamp 
          ? new Date(b.last_message.timestamp).getTime() 
          : new Date(b.updated_at || b.created_at).getTime();
        return timeB - timeA;
      });
      
      // Update state
      chats = uniqueChats;
      filteredChats = uniqueChats;
      
      logger.info(`Loaded ${chats.length} chats`);
      return uniqueChats;
    } catch (error) {
      logger.error('Failed to fetch chats', error);
      toastStore.showToast('Failed to load chats', 'error');
      return [];
    } finally {
      isLoadingChats = false;
    }
  }
  
  // Helper function to map API chat data to client format
  function mapApiChatsToClientFormat(apiChats: any[]): Chat[] {
    logger.debug('Mapping API chats to client format', { count: apiChats.length });
    
    // First convert all chats to our format
    return apiChats.map(chat => {
      // Log each chat's structure for debugging
      logger.debug('Processing chat', { 
        id: chat.id,
        type: chat.is_group_chat || chat.type === 'group' ? 'group' : 'individual',
        name: chat.name || '',
        hasParticipants: !!chat.participants,
        participantCount: chat.participants?.length || 0
      });
      
      // Ensure participants is always an array
      const participants = Array.isArray(chat.participants) ? chat.participants : [];
      
      // Format the last_message with string timestamp
      const lastMessage = chat.last_message ? {
        content: chat.last_message.content || '',
        timestamp: ensureStringTimestamp(chat.last_message.timestamp || Date.now()),
        sender_id: chat.last_message.sender_id || chat.last_message.user_id || '',
        sender_name: chat.last_message.sender_name || chat.last_message.username || ''
      } : undefined;
      
      return {
        id: chat.id,
        type: chat.is_group_chat || chat.type === 'group' ? 'group' : 'individual',
        name: chat.name || '',
        avatar: chat.profile_picture_url || chat.avatar_url || chat.avatar,
        profile_picture_url: chat.profile_picture_url || chat.avatar_url || chat.avatar,
        participants: participants.map((p: any) => ({
          id: p.user_id || p.id,
          username: p.username || '',
          display_name: p.display_name || p.name || p.username || 'User',
          avatar: p.avatar_url || p.profile_picture_url || p.avatar || null,
          is_verified: p.is_verified || false
        })) || [],
        last_message: lastMessage,
        messages: [],
        unread_count: chat.unread_count || 0,
        created_at: chat.created_at || new Date().toISOString(),
        updated_at: chat.updated_at || chat.created_at || new Date().toISOString()
      };
    });
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
        profile_picture_url: null,
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString()
      };
  }

  // Fix the formatMessageTime function to ensure the argument to formatRelativeTime is always a string
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
   * Get the other participant in an individual chat
   */
  function getOtherParticipant(chat: Chat): Participant | undefined {
    if (chat.type !== 'individual') return undefined;
    
    return chat.participants.find(p => p.id !== $authStore.user_id);
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
    <div class="websocket-status">
      <div class="status-indicator {$websocketStore.connected ? 'connected' : 'disconnected'}">
        <span class="status-icon">{$websocketStore.connected ? '●' : '○'}</span>
        <span class="status-text">Real-time {$websocketStore.connected ? 'Connected' : 'Disconnected'}</span>
      </div>
      {#if $websocketStore.lastError}
        <div class="status-error">
          <span class="error-icon">⚠️</span>
          <span class="error-text">{$websocketStore.lastError}</span>
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
    on:close={() => showNewChatModal = false}
    on:search={(e) => searchForUsers(e.detail)}
    on:createChat={(e) => initiateNewChat(e.detail)}
  />
{/if}

{#if showCreateGroupModal}
  <CreateGroupChat 
    on:close={() => showCreateGroupModal = false}
    on:createGroup={(e) => createGroupChat(e.detail)}
    {userSearchResults}
    {isLoadingUsers}
    searchKeyword={userSearchQuery}
    on:search={(e) => searchForUsers(e.detail)}
  />
{/if}

{#if showDebug}
  <DebugPanel 
    on:close={() => showDebug = false}
    {chats}
    {selectedChat}
    authToken={$authStore.token}
    userId={$authStore.user_id}
    wsConnected={$websocketStore.connected}
    wsError={$websocketStore.lastError}
    on:testConnection={() => testApiConnection()}
    on:checkAuth={() => logAuthTokenInfo()}
  />
{/if}

<Toast {toasts} on:close={(e) => toastStore.removeToast(e.detail)} />


