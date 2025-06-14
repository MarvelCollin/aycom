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
  
  // Use the imported chatApi methods
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
    content: string;
    timestamp: string;
    sender_id: string;
    sender_name?: string;
    sender_avatar?: string;
    is_read: boolean;
    is_deleted: boolean;
    attachments?: Attachment[];
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
    selectedChat = { ...chat, messages: [] };
    
    // On mobile, hide the chat list
    if (isMobile) {
      showMobileMenu = false;
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
          sender_avatar: msg.sender_avatar || null
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
      logger.error(`Error fetching messages for chat ${chat.id}:`, error);
      toastStore.showToast('Failed to load messages', 'error');
    } finally {
      isLoadingMessages = false;
    }
    
    // Connect to WebSocket for this chat
    try {
      websocketStore.connect(chat.id);
      logger.info(`Connected to WebSocket for chat ${chat.id}`);
    } catch (error) {
      logger.error(`Error connecting to WebSocket for chat ${chat.id}:`, error);
    }
  }
  
  async function startChatWithUser(user: StandardUser) {
    // Logic to start a new chat with a user
    logger.debug(`Starting new chat with user: ${user.username}`);
    
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
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString()
        };
        
        // Add to chats and select it
        chats = [newChat, ...chats];
        filteredChats = [newChat, ...filteredChats];
        selectChat(newChat);
        
        // Close the modal
        showNewChatModal = false;
      }
    } catch (error) {
      logger.error('Failed to create chat', error);
      toastStore.showToast('Failed to create chat', 'error');
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
    // Return a display name based on chat participants
    return chat.name || (chat.participants && chat.participants[0]?.display_name) || 'Chat';
  }
  
  function formatTimeAgo(timestamp: string | number) {
    // Simple time formatter
    if (!timestamp) return '';
    
    const date = typeof timestamp === 'string' ? new Date(timestamp) : new Date(timestamp);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffSecs = Math.floor(diffMs / 1000);
    const diffMins = Math.floor(diffSecs / 60);
    const diffHours = Math.floor(diffMins / 60);
    const diffDays = Math.floor(diffHours / 24);
    
    if (diffSecs < 60) {
      return 'just now';
    } else if (diffMins < 60) {
      return `${diffMins}m ago`;
    } else if (diffHours < 24) {
      return `${diffHours}h ago`;
    } else if (diffDays < 7) {
      return `${diffDays}d ago`;
    } else {
      return date.toLocaleDateString(undefined, { month: 'short', day: 'numeric' });
    }
  }
  
  // Message handling functions
  async function sendMessage() {
    // Logic to send a message
    if (!newMessage.trim() || !selectedChat) return;
    
    const content = newMessage.trim();
    // Clear input right away for better UX
    newMessage = '';
    
    // Add optimistic update - add message to UI immediately
    const tempMessageId = `temp-${Date.now()}`;
    const tempMessage: Message = {
      id: tempMessageId,
      content,
      timestamp: new Date().toISOString(),
      sender_id: $authStore.user_id || '',
      sender_name: displayName,
      sender_avatar: avatar,
      is_read: false,
      is_deleted: false,
      attachments: selectedAttachments.length > 0 ? [...selectedAttachments] : undefined
    };
    
    // Clear attachments
    selectedAttachments = [];
    
    // Update UI optimistically
    if (selectedChat?.messages) {
      selectedChat = {
        ...selectedChat,
        messages: [...selectedChat.messages, tempMessage]
      };
    }
    
    try {
      // Call API to send message
      const response = await apiSendMessage(selectedChat.id, {
        content: content,
        attachments: selectedAttachments.map(att => att.id)
      });
      
      if (response && response.id) {
        // Replace temp message with real one
        if (selectedChat?.messages) {
          selectedChat = {
            ...selectedChat,
            messages: selectedChat.messages.map(msg => 
              msg.id === tempMessageId ? { ...msg, id: response.id } : msg
            )
          };
        }
        
        // Update chat in the list with last message
        const updatedChats = chats.map(chat => {
          if (chat.id === selectedChat?.id) {
            return {
              ...chat,
              last_message: {
                content,
                timestamp: new Date().toISOString(),
                sender_id: $authStore.user_id || '',
                sender_name: displayName
              }
            };
          }
          return chat;
        });
        
        // Update chats and move the active chat to top
        chats = [
          updatedChats.find(c => c.id === selectedChat?.id) as Chat,
          ...updatedChats.filter(c => c.id !== selectedChat?.id)
        ];
        
        // Also update filtered chats
        filteredChats = [
          ...filteredChats.filter(c => c.id === selectedChat?.id),
          ...filteredChats.filter(c => c.id !== selectedChat?.id)
        ];
      }
    } catch (error) {
      logger.error('Failed to send message', error);
      toastStore.showToast('Failed to send message', 'error');
      
      // Remove the optimistic update on error
      if (selectedChat?.messages) {
        selectedChat = {
          ...selectedChat,
          messages: selectedChat.messages.filter(msg => msg.id !== tempMessageId)
        };
      }
    }
  }
  
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
                  timestamp: newLastMessage.timestamp,
                  sender_id: newLastMessage.sender_id,
                  sender_name: newLastMessage.sender_name
                }
              };
            }
            return chat;
          });
          
          // Also update filtered chats
          filteredChats = filteredChats.map(chat => {
            if (chat.id === selectedChat?.id) {
              return {
                ...chat,
                last_message: {
                  content: newLastMessage.content,
                  timestamp: newLastMessage.timestamp,
                  sender_id: newLastMessage.sender_id,
                  sender_name: newLastMessage.sender_name
                }
              };
            }
            return chat;
          });
        }
        }
    } catch (error) {
      logger.error('Failed to unsend message', error);
      toastStore.showToast('Failed to unsend message', 'error');
      
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
  
  // Handle WebSocket messages
  function handleWebSocketMessage(message: any) {
    logger.debug('Received WebSocket message:', message);
    
    // Only process messages for the currently selected chat
    if (!selectedChat || message.chat_id !== selectedChat.id) {
      logger.debug('Message not for current chat, ignoring');
      return;
    }
    
    // Handle text messages
    if (message.type === 'text' || !message.type) {
      // Create a standardized message object
      const newMessage = {
        id: message.id || message.message_id,
        chat_id: message.chat_id,
        content: message.content,
        sender_id: message.sender_id || message.user_id,
        sender_name: message.sender_name || 'User',
        sender_avatar: message.sender_avatar,
        timestamp: message.timestamp || new Date().toISOString(),
        is_read: message.is_read || false,
        is_edited: message.is_edited || false,
        is_deleted: message.is_deleted || false
      };
      
      // Check if this is updating a temporary message
      if (message.temp_id) {
        const tempIndex = selectedChat.messages.findIndex(m => m.id === message.temp_id);
        if (tempIndex >= 0) {
          // Replace the temporary message with the confirmed one
          selectedChat.messages[tempIndex] = newMessage;
          selectedChat = { ...selectedChat }; // Trigger reactivity
          logger.debug('Updated temporary message with confirmed message');
          return;
        }
      }
      
      // Otherwise add as a new message
      selectedChat.messages = [...selectedChat.messages, newMessage];
      selectedChat = { ...selectedChat }; // Trigger reactivity
      
      // Scroll to bottom
      setTimeout(() => {
        const messagesContainer = document.querySelector('.messages-container');
        if (messagesContainer) {
          messagesContainer.scrollTop = messagesContainer.scrollHeight;
        }
      }, 100);
    }
    
    // Handle delete/unsend messages
    if (message.type === 'delete' || message.type === 'unsend') {
      if (message.message_id) {
        selectedChat.messages = selectedChat.messages.map(msg => 
          msg.id === message.message_id ? { ...msg, is_deleted: true, content: 'Message deleted' } : msg
        );
        selectedChat = { ...selectedChat }; // Trigger reactivity
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
    
    // Run API connection test first
    testApiConfig();
    
    // Initialize user profile and chats
    if ($authStore && $authStore.is_authenticated) {
      fetchUserProfile();
      fetchChats().then(() => {
        // Initialize WebSocket connection if a chat is selected
        if (selectedChat && selectedChat.id) {
          // Connect to WebSocket for selected chat
          try {
            websocketStore.connect(selectedChat.id);
          } catch (error) {
            logger.error('Error connecting to WebSocket:', error);
          }
        }
      });
    }
    
    // Register WebSocket message handler
    setMessageHandler(handleWebSocketMessage);
    
    return () => {
      window.removeEventListener('resize', checkViewport);
      
      // Disconnect WebSocket connections when component unmounts
      if (selectedChat && selectedChat.id) {
        try {
          websocketStore.disconnect(selectedChat.id);
        } catch (error) {
          logger.error('Error disconnecting from WebSocket:', error);
        }
      }
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
  
  // Call this function to load the chat list
  async function fetchChats() {
    if (!$authStore.is_authenticated) return;
    
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
      
      logger.info(`Filtered ${mappedChats.length} chats down to ${uniqueChats.length} unique chats`);
      
      // Set the chats and filtered chats
      chats = uniqueChats;
      filteredChats = [...uniqueChats];
    } catch (error) {
      logger.error('Failed to load chats', error);
      toastStore.showToast('Failed to load chats', 'error');
      chats = [];
      filteredChats = [];
    } finally {
      isLoadingChats = false;
    }
  }
  
  // Helper function to map API chat data to client format
  function mapApiChatsToClientFormat(apiChats: any[]): Chat[] {
    // First convert all chats to our format
    return apiChats.map(chat => ({
      id: chat.id,
      type: chat.is_group_chat || chat.type === 'group' ? 'group' : 'individual',
      name: chat.name || '',
      avatar: chat.profile_picture_url || chat.avatar_url || chat.avatar,
      profile_picture_url: chat.profile_picture_url || chat.avatar_url || chat.avatar,
      participants: chat.participants?.map((p: any) => ({
        id: p.user_id || p.id,
        username: p.username || '',
        display_name: p.display_name || p.username || p.name || 'User',
        avatar: p.avatar_url || p.profile_picture_url || p.avatar || null,
        is_verified: p.is_verified || false
      })) || [],
      last_message: chat.last_message ? {
        content: chat.last_message.content || '',
        timestamp: chat.last_message.timestamp || Date.now(),
        sender_id: chat.last_message.sender_id || chat.last_message.user_id || '',
        sender_name: chat.last_message.sender_name || chat.last_message.username || ''
      } : undefined,
      messages: [],
      unread_count: chat.unread_count || 0,
      created_at: chat.created_at || new Date().toISOString(),
      updated_at: chat.updated_at || chat.created_at || new Date().toISOString()
    }));
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
</script>

<div class="custom-message-layout {isDarkMode ? 'dark-theme' : ''}">
  <!-- Mobile header -->
  {#if isMobile}
    <div class="mobile-header">
      <button class="mobile-menu-button" on:click={toggleMobileMenu} aria-label="Toggle mobile menu">>
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
    <div class="mobile-overlay" 
         on:click={toggleMobileMenu} 
         on:keydown={(e) => e.key === 'Enter' && toggleMobileMenu()}
         role="button"
         tabindex="0"
         aria-label="Close mobile menu"></div>
  {/if}

  <!-- Main content area -->
  <div class="custom-content-area">
    <!-- WebSocket Status Display -->
    <div class="websocket-status">
      <div class="status-indicator {$websocketStore.connected ? 'connected' : 'disconnected'}">
        {$websocketStore.connected ? '🟢' : '🔴'} WebSocket: {$websocketStore.connected ? 'Connected' : 'Disconnected'}
      </div>
      {#if $websocketStore.lastError}
        <div class="status-error">Error: {$websocketStore.lastError}</div>
      {/if}
      {#if Object.keys($websocketStore.connectionStatus).length > 0}
        <div class="chat-connections">
          {#each Object.entries($websocketStore.connectionStatus) as [chatId, status]}
            <span class="chat-status {status}">Chat {chatId.slice(-8)}: {status}</span>
          {/each}
        </div>
      {/if}
    </div>

    <div class="message-container {isDarkMode ? 'dark-theme' : ''}">
  <div class="middle-section">
        <!-- Chat header -->
    <div class="section-header">
      <h1>Messages</h1>
      <div class="button-group">        <button
          class="compose-button"
          on:click={() => showNewChatModal = true}
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
            <button class="clear-search" on:click={() => searchQuery = ''} aria-label="Clear search">
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
          <div class="empty-state">
            <p>No conversations yet</p>
            <button class="compose-button" on:click={() => showNewChatModal = true}>
              Start a new chat
            </button>
          </div>
        {:else}
          {#each filteredChats as chat}
            <div 
              class="chat-item {selectedChat?.id === chat.id ? 'selected' : ''}" 
              on:click={() => selectChat(chat)}
              on:keydown={(e) => e.key === 'Enter' && selectChat(chat)}
              role="button"
              tabindex="0"
              aria-label="Open chat with {getChatDisplayName(chat)}"
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
              <button class="back-button" on:click={() => selectedChat = null} aria-label="Back to chat list">
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
            {#if isLoadingMessages}
              <div class="loading-container">
                <div class="loading-spinner"></div>
                <p>Loading messages...</p>
              </div>
            {:else if selectedChat.messages && selectedChat.messages.length > 0}
        {#each selectedChat.messages as message}
                <div class="message-item {message.sender_id === $authStore.user_id ? 'own-message' : ''} {message.is_deleted ? 'deleted' : ''}">
              {#if message.sender_id !== $authStore.user_id}
                    <div class="message-avatar">
                      {#if message.sender_avatar}
                        <img src={message.sender_avatar} alt={message.sender_name} />
                      {:else}
                        <div class="avatar-placeholder" style="background-color: {getAvatarColor(message.sender_name || 'User')}">
                          {(message.sender_name || 'User').charAt(0).toUpperCase()}
                        </div>
              {/if}
              </div>
                  {/if}
                  
                  <div class="message-bubble">
                    {#if message.sender_id !== $authStore.user_id && selectedChat.type === 'group'}
                      <div class="sender-name">{message.sender_name || 'User'}</div>
                    {/if}
                    
                    {#if message.is_deleted}
                      <div class="deleted-message">Message deleted</div>
                    {:else}
                      <div class="content-text">{message.content}</div>
                      
                      {#if message.attachments && message.attachments.length > 0}
                        <div class="attachments-container">
                  {#each message.attachments as attachment}
                            {#if attachment.type === 'image'}
                              <img src={attachment.url} alt="Attachment" class="image-attachment" />
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
                        
                        {#if message.sender_id === $authStore.user_id}
                          <div class="message-actions">
                            <button class="action-button" on:click={() => unsendMessage(message.id)} aria-label="Delete message">
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
                <button class="attachment-button" on:click={() => handleAttachment('image')} aria-label="Add image">
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" width="20" height="20">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
            </svg>
        </button>
                <button class="attachment-button" on:click={() => handleAttachment('gif')} aria-label="Add GIF">
                  <span class="gif-button">GIF</span>
                </button>
              </div>
            </div>
            
        <button
              class="send-button {newMessage.trim() ? 'active' : ''}"
          disabled={!newMessage.trim()}
              on:click={sendMessage}
              aria-label="Send message"
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
            <p>Send private messages to a friend or group</p>            <button 
              class="compose-button"
              on:click={() => showNewChatModal = true}
            >
              New Message
            </button>
      </div>
    {/if}
      </div>
  </div>
</div>

<!-- Add the modals -->
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

{#if showNewChatModal}
  <NewChatModal
    onCancel={() => showNewChatModal = false}
    onUserSelect={startChatWithUser}
  />
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
  
  /* WebSocket Status Styles */
  .websocket-status {
    position: fixed;
    top: 10px;
    right: 10px;
    z-index: 1000;
    background: rgba(0, 0, 0, 0.8);
    color: white;
    padding: 8px 12px;
    border-radius: 6px;
    font-size: 12px;
    font-family: monospace;
    border: 1px solid rgba(255, 255, 255, 0.2);
  }
  
  .status-indicator {
    margin-bottom: 4px;
  }
  
  .status-indicator.connected {
    color: #4ade80;
  }
  
  .status-indicator.disconnected {
    color: #ef4444;
  }
  
  .status-error {
    color: #fbbf24;
    font-size: 11px;
    margin-bottom: 4px;
  }
  
  .chat-connections {
    display: flex;
    flex-wrap: wrap;
    gap: 4px;
  }
  
  .chat-status {
    padding: 2px 6px;
    border-radius: 3px;
    font-size: 10px;
    background: rgba(255, 255, 255, 0.1);
  }
  
  .chat-status.connected {
    background: rgba(74, 222, 128, 0.2);
    color: #4ade80;
  }
  
  .chat-status.disconnected {
    background: rgba(239, 68, 68, 0.2);
    color: #ef4444;
  }
  
  .chat-status.connecting {
    background: rgba(251, 191, 36, 0.2);
    color: #fbbf24;
  }
  
  .chat-status.error {
    background: rgba(239, 68, 68, 0.3);
    color: #ef4444;
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
  
  /* Dark theme fixes for message bubbles and icons */
  .message-container.dark-theme .message-item.own-message .message-bubble {
    background-color: #4b5563;
    color: white;
  }
  
  .message-container.dark-theme .message-item .action-button {
    color: #e5e7eb;
  }
  
  .message-container.dark-theme .message-actions svg {
    stroke: #e5e7eb;
  }
  
  .message-container.dark-theme .message-input-container {
    background-color: #1f2937;
  }
  
  .message-container.dark-theme .input-wrapper textarea {
    background-color: #374151;
    color: white;
    border-color: #4b5563;
  }
  
  .message-container.dark-theme .attachment-button svg {
    stroke: #e5e7eb;
  }
  
  .message-container.dark-theme .send-button {
    background-color: #3b82f6;
  }
  
  .message-container.dark-theme .gif-button {
    color: #e5e7eb;
  }
  
  /* Message bubble styling improvements */
  .message-bubble {
    padding: 12px 16px;
    border-radius: 18px;
    max-width: 80%;
    word-wrap: break-word;
    position: relative;
    box-shadow: 0 1px 2px rgba(0,0,0,0.1);
    margin: 4px 0;
    display: inline-block;
  }
  
  .own-message .message-bubble {
    background-color: #3b82f6;
    color: white;
    border-bottom-right-radius: 4px;
    margin-left: auto;
  }
  
  .message-container.dark-theme .own-message .message-bubble {
    background-color: #3b82f6;
  }
  
  .message-item:not(.own-message) .message-bubble {
    background-color: #f3f4f6;
    color: #1f2937;
    border-bottom-left-radius: 4px;
    margin-right: auto;
  }
  
  .message-container.dark-theme .message-item:not(.own-message) .message-bubble {
    background-color: #374151;
    color: #f3f4f6;
  }
  
  .message-item {
    display: flex;
    margin-bottom: 8px;
    position: relative;
    width: 100%;
    padding: 0 16px;
  }
  
  /* Fix for the just now message bubble */
  .message-container.dark-theme .message-item {
    color: white;
  }
  
  .message-container.dark-theme .content-text {
    color: white;
  }
  
  /* Fix for the delete icon in message bubbles */
  .message-actions {
    display: flex;
    align-items: center;
    margin-left: 8px;
  }
  
  .action-button {
    background: transparent;
    border: none;
    padding: 4px;
    border-radius: 50%;
    cursor: pointer;
    opacity: 0.7;
    transition: opacity 0.2s;
    display: flex;
    align-items: center;
    justify-content: center;
  }
  
  .action-button:hover {
    opacity: 1;
    background-color: rgba(0,0,0,0.1);
  }
  
  .message-container.dark-theme .action-button:hover {
    background-color: rgba(255,255,255,0.1);
  }
  
  .action-button svg {
    width: 16px;
    height: 16px;
  }
  
  /* Message timestamp styling */
  .timestamp {
    font-size: 0.7rem;
    opacity: 0.7;
    margin-right: 4px;
  }
  
  /* Message avatar styling */
  .message-avatar {
    width: 36px;
    height: 36px;
    border-radius: 50%;
    margin-right: 8px;
    flex-shrink: 0;
    overflow: hidden;
  }
  
  .message-avatar img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }
  
  .message-footer {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-top: 4px;
    font-size: 0.7rem;
  }
  
  /* Input area styling */
  .input-wrapper {
    display: flex;
    flex: 1;
    background-color: #f3f4f6;
    border-radius: 24px;
    padding: 8px 16px;
    align-items: center;
  }
  
  .message-container.dark-theme .input-wrapper {
    background-color: #374151;
  }
  
  .input-wrapper textarea {
    flex: 1;
    border: none;
    background: transparent;
    resize: none;
    padding: 8px 0;
    outline: none;
    color: inherit;
    font-family: inherit;
    font-size: 0.95rem;
  }
  
  .attachment-buttons {
    display: flex;
    align-items: center;
    margin-left: 8px;
  }
  
  .attachment-button {
    background: transparent;
    border: none;
    color: #6b7280;
    cursor: pointer;
    padding: 4px;
    margin-left: 4px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: all 0.2s;
  }
  
  .attachment-button:hover {
    background-color: rgba(0,0,0,0.05);
    color: #3b82f6;
  }
  
  .message-container.dark-theme .attachment-button:hover {
    background-color: rgba(255,255,255,0.1);
  }
  
  .gif-button {
    font-weight: bold;
    font-size: 0.8rem;
  }
  
  .loading-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 100%;
    color: var(--text-secondary);
    padding: 20px;
  }
  
  .loading-spinner {
    width: 30px;
    height: 30px;
    border: 3px solid rgba(255, 255, 255, 0.2);
    border-radius: 50%;
    border-top-color: var(--accent-color);
    animation: spin 1s linear infinite;
    margin-bottom: 10px;
  }
  
  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }
</style>


