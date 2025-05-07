<script lang="ts">
  import { onMount } from 'svelte';
  import LeftSide from '../components/layout/LeftSide.svelte';
  import { useAuth } from '../hooks/useAuth';
  import { useTheme } from '../hooks/useTheme';
  import type { IAuthStore } from '../interfaces/IAuth';
  import { createLoggerWithPrefix } from '../utils/logger';
  import { toastStore } from '../stores/toastStore';
  import { checkAuth, isWithinTime, handleApiError } from '../utils/common';
  import { listChats, listMessages, sendMessage as apiSendMessage, unsendMessage as apiUnsendMessage, searchMessages, createChat } from '../api/chat';
  import { getProfile, searchUsers } from '../api/user';
  import '../styles/magniview.css'
  
  const logger = createLoggerWithPrefix('Message');

  const { getAuthState } = useAuth();
  const { theme } = useTheme();
  
  // Reactive declarations
  $: authState = getAuthState ? (getAuthState() as IAuthStore) : { userId: null, isAuthenticated: false, accessToken: null, refreshToken: null };
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

  // Dynamic class for mobile view
  let chatSelectedClass = '';
  $: {
    chatSelectedClass = selectedChat ? 'chat-selected' : '';
  }

  // Update class on the container
  $: if (typeof document !== 'undefined') {
    const container = document.querySelector('.message-container');
    if (container) {
      if (selectedChat && window.innerWidth <= 768) {
        container.classList.add('chat-selected');
      } else {
        container.classList.remove('chat-selected');
      }
    }
  }

  // Message interfaces
  interface Message {
    id: string;
    senderId: string;
    senderName: string;
    senderAvatar: string | null;
    content: string;
    timestamp: string;
    isDeleted: boolean;
    attachments: Attachment[];
    isOwn: boolean;
  }

  interface Attachment {
    id: string;
    type: 'image' | 'gif' | 'video';
    url: string;
    thumbnail?: string;
  }

  interface Chat {
    id: string;
    type: 'individual' | 'group';
    name: string;
    avatar: string | null;
    participants: Participant[];
    lastMessage?: {
      content: string;
      timestamp: string;
      senderId: string;
    };
    messages: Message[];
    unreadCount: number;
  }

  interface Participant {
    id: string;
    username: string;
    displayName: string;
    avatar: string | null;
    isVerified: boolean;
  }
  
  // Add user search results state
  let userSearchResults: Participant[] = [];

  // Fetch user profile data using the API directly
  async function fetchUserProfile() {
    isLoadingProfile = true;
    try {
      const response = await getProfile();
      if (response && response.user) {
        const userData = response.user;
        username = userData.username || `user_${authState.userId?.substring(0, 4)}`;
        displayName = userData.name || userData.display_name || `User ${authState.userId?.substring(0, 4)}`;
        avatar = userData.profile_picture_url || 'https://secure.gravatar.com/avatar/0?d=mp';
        logger.debug('Profile loaded', { username });
      } else {
        logger.warn('No user data received from API');
        username = `user_${authState.userId?.substring(0, 4)}`;
        displayName = `User ${authState.userId?.substring(0, 4)}`;
      }
    } catch (error) {
      const errorResponse = handleApiError(error);
      logger.error('Error fetching user profile:', errorResponse);
      username = `user_${authState.userId?.substring(0, 4)}`;
      displayName = `User ${authState.userId?.substring(0, 4)}`;
    } finally {
      isLoadingProfile = false;
    }
  }

  // Authentication check on component load
  onMount(() => {
    if (!checkAuth(authState, 'messages')) {
      return;
    }
    
    // Fetch user profile first, then chats
    fetchUserProfile().then(() => {
      fetchChats().then(() => {
        // Try to restore the last selected chat from localStorage
        try {
          const savedChatId = localStorage.getItem('selectedChatId');
          if (savedChatId && chats.length > 0) {
            const chatToSelect = chats.find(c => c.id === savedChatId);
            if (chatToSelect) {
              selectChat(chatToSelect);
              logger.debug('Restored selected chat from localStorage', { chatId: savedChatId });
            }
          }
        } catch (error) {
          logger.warn('Failed to restore chat from localStorage', error);
        }
      });
    });
  });

  // Fetch chats
  async function fetchChats() {
    isLoadingChats = true;
    try {
      const response = await listChats();
      logger.debug('Raw chat data from API:', { chats: response.chats });
      
      // Backend returns { chats: [...] }
      if (response && response.chats && Array.isArray(response.chats)) {
        chats = response.chats.map((chat: any) => {
          // Get the chat ID
          const chatId = chat.id || chat.Id;
          logger.debug(`Processing chat ${chatId}`, chat);
          
          // Determine if this is a group chat
          const isGroup = chat.is_group_chat || chat.IsGroupChat || chat.is_group || false;
          
          // Process participants list
          let processedParticipants: Participant[] = [];
          if (Array.isArray(chat.participants)) { 
            processedParticipants = chat.participants.map((p: any) => ({
              id: p.id || p.user_id || p.userId || '',
              username: p.username || '',
              displayName: p.display_name || p.displayName || p.username || 'User',
              avatar: p.profile_picture_url || p.avatar || null,
              isVerified: p.is_verified || false
            }));
          }
          
          // Try to determine chat name
          let chatName: string;
          let chatAvatar: string | null = null;
          
          // Get name directly if provided and valid
          if (chat.name && chat.name !== 'Chat' && chat.name !== 'null null' && chat.name !== 'New Chat') {
            chatName = chat.name;
          }
          // For individual chats, try to use the other participant's name
          else if (!isGroup && processedParticipants.length > 0) {
            // Find the other participant (not the current user)
            const otherParticipant = processedParticipants.find(p => 
              p.id !== authState.userId && 
              p.id !== `${authState.userId}`
            );
            
            if (otherParticipant) {
              chatName = otherParticipant.displayName || otherParticipant.username || 'Chat Partner';
              chatAvatar = otherParticipant.avatar;
            } else {
              // If we couldn't find another participant, use the first one
              chatName = processedParticipants[0].displayName || 
                        processedParticipants[0].username || 
                        'Chat';
              chatAvatar = processedParticipants[0].avatar;
            }
          }
          // For group chats without a name
          else if (isGroup) {
            chatName = `Group (${processedParticipants.length} members)`;
          } 
          // Check creator info if available
          else if (chat.created_by || chat.createdBy) {
            const creatorId = chat.created_by || chat.createdBy;
            // If current user is creator
            if (creatorId === authState.userId) {
              chatName = "My Chat";
            } else {
              chatName = "Chat";
            }
          }
          // Default fallback
          else {
            chatName = "Chat";
          }
          
          // Process last message 
          let lastMessageData;
          if (chat.last_message) {
            if (typeof chat.last_message === 'string') {
              lastMessageData = {
                content: chat.last_message,
                timestamp: Date.now() / 1000,
                senderId: ''
              };
            } else {
              lastMessageData = {
                content: chat.last_message.content || '',
                timestamp: chat.last_message.timestamp || Date.now() / 1000,
                senderId: chat.last_message.sender_id || ''
              };
            }
          } else if (chat.lastMessage) {
            if (typeof chat.lastMessage === 'string') {
              lastMessageData = {
                content: chat.lastMessage,
                timestamp: Date.now() / 1000,
                senderId: ''
              };
            } else {
              lastMessageData = {
                content: chat.lastMessage.content || chat.lastMessage.Content || '',
                timestamp: chat.lastMessage.timestamp || chat.lastMessage.Timestamp || Date.now() / 1000,
                senderId: chat.lastMessage.sender_id || chat.lastMessage.SenderId || ''
              };
            }
          } else {
            // Create a default last message when none exists
            lastMessageData = {
              content: '',
              timestamp: Date.now() / 1000,
              senderId: ''
            };
          }
          
          // Create the chat object with properly formatted data
          const formattedChat: Chat = {
            id: chatId,
            type: isGroup ? 'group' : 'individual',
            name: chatName,
            avatar: chatAvatar,
            participants: processedParticipants,
            lastMessage: lastMessageData,
            messages: [],
            unreadCount: chat.unread_count || 0
          };
          
          logger.debug('Processed chat:', { 
            id: formattedChat.id, 
            name: formattedChat.name,
            type: formattedChat.type,
            participantsCount: formattedChat.participants.length,
            lastMessage: formattedChat.lastMessage ? formattedChat.lastMessage.content : 'No last message'
          });
          
          return formattedChat;
        });
        
        filteredChats = [...chats];
        logger.debug('Chats loaded', { count: chats.length });
      } else {
        logger.warn('No chats found or invalid response format');
        chats = [];
        filteredChats = [];
      }
    } catch (error) {
      const errorResponse = handleApiError(error);
      logger.error('Error fetching chats:', errorResponse);
      toastStore.showToast('Failed to load chats. Please try again.', 'error');
      chats = [];
      filteredChats = [];
    } finally {
      isLoadingChats = false;
    }
  }
  
  // Helper function to get a chat name from participants
  function getParticipantName(chat: any): string {
    if (chat.participants && chat.participants.length > 0) {
      const participant = chat.participants[0];
      return participant.display_name || participant.username || 'Chat';
    }
    return 'Chat';
  }

  // Function to get the display name for a chat based on participants
  function getChatDisplayName(chat: Chat): string {
    // If chat has a name that's not the default placeholder, use it
    if (chat.name && chat.name !== 'Chat' && chat.name !== 'null null') {
      return chat.name;
    }
    
    // If it's a group chat with no name or a default name
    if (chat.type === 'group') {
      return `Group (${chat.participants.length} members)`;
    }
    
    // For individual chats, find the other participant (not the current user)
    if (chat.participants && chat.participants.length > 0) {
      // Find participant that isn't the current user
      const otherParticipant = chat.participants.find(p => 
        p.id !== authState.userId && 
        p.id !== `${authState.userId}`
      );
      
      if (otherParticipant) {
        return otherParticipant.displayName || otherParticipant.username || 'Chat Partner';
      }
      
      // If we couldn't find another participant, use the first participant
      const participant = chat.participants[0];
      return participant.displayName || participant.username || 'Chat';
    }
    
    // Ultimate fallback
    return 'Chat';
  }

  // Unsend a message using the API
  async function unsendMessage(messageId: string) {
    if (!selectedChat) return;

    try {
      await apiUnsendMessage(selectedChat.id, messageId);
      
      // Update local state after successful API call
      selectedChat.messages = selectedChat.messages.map(m => 
        m.id === messageId ? { ...m, isDeleted: true } : m
      );

      // If this was the last message, update the lastMessage property
      if (selectedChat.lastMessage && selectedChat.messages.find(m => m.id === messageId)) {
        const lastNonDeletedMessage = [...selectedChat.messages]
          .reverse()
          .find(m => !m.isDeleted);
        
        if (lastNonDeletedMessage) {
          selectedChat.lastMessage = {
            content: lastNonDeletedMessage.content,
            timestamp: lastNonDeletedMessage.timestamp,
            senderId: lastNonDeletedMessage.senderId
          };
        } else {
          selectedChat.lastMessage = undefined;
        }
      }
      
      logger.debug('Message unsent', { messageId });
      toastStore.showToast('Message unsent', 'success');
    } catch (error) {
      const errorResponse = handleApiError(error);
      logger.error('Failed to unsend message:', errorResponse);
      toastStore.showToast('Failed to unsend message. Please try again.', 'error');
    }
  }

  // Select a chat and load messages
  async function selectChat(chat: Chat) {
    selectedChat = chat;
    
    // Save selected chat ID to localStorage
    try {
      localStorage.setItem('selectedChatId', chat.id);
    } catch (error) {
      logger.warn('Failed to save selected chat to localStorage', error);
    }
    
    try {
      logger.debug(`Selecting chat ${chat.id} and loading messages`);
      const response = await listMessages(chat.id);
      
      if (response && response.messages && Array.isArray(response.messages)) {
        logger.debug(`Received ${response.messages.length} messages for chat ${chat.id}`);
        
        // Transform the messages to the expected format
        selectedChat.messages = response.messages.map((msg: any) => {
          // Handle inconsistent field names
          const id = msg.id || msg.message_id || msg.Id;
          const senderId = msg.sender_id || msg.user_id || msg.SenderId;
          const content = msg.content || msg.Content || '';
          let timestamp = msg.timestamp || msg.Timestamp || Date.now() / 1000;
          const isDeleted = msg.is_deleted || msg.IsDeleted || false;
          
          // Ensure timestamp is a number
          if (typeof timestamp === 'string') {
            timestamp = parseInt(timestamp);
          }
          
          logger.debug(`Processing message ${id} from sender ${senderId}`, { 
            messageType: typeof msg,
            timestamp: timestamp 
          });
          
          return {
            id: id,
            senderId: senderId,
            senderName: msg.sender_name || displayName,
            senderAvatar: msg.sender_avatar || avatar,
            content: content,
            timestamp: timestamp.toString(),
            isDeleted: isDeleted,
            attachments: msg.attachments || [],
            isOwn: senderId === authState.userId
          };
        });
        
        // Sort messages by timestamp (oldest first)
        selectedChat.messages.sort((a, b) => {
          const timestampA = parseInt(a.timestamp);
          const timestampB = parseInt(b.timestamp);
          return timestampA - timestampB;
        });
        
        logger.debug(`Processed ${selectedChat.messages.length} messages for display`);
      } else {
        logger.warn(`No messages or invalid response format for chat ${chat.id}`);
        selectedChat.messages = [];
      }
      
      // Reset unread count
      selectedChat.unreadCount = 0;
      
    } catch (error) {
      const errorResponse = handleApiError(error);
      logger.error('Error loading messages:', errorResponse);
      toastStore.showToast('Failed to load messages. Please try again.', 'error');
    }
  }

  // Send a message using the API
  async function sendMessage() {
    if (!selectedChat || (!newMessage.trim() && selectedAttachments.length === 0)) return;
    try {
      // First, add a temporary message to the UI for immediate feedback
      const tempId = `temp-${Date.now()}`;
      const tempMessage: Message = {
        id: tempId,
        senderId: authState.userId as string,
        senderName: displayName,
        senderAvatar: avatar,
        content: newMessage.trim(),
        timestamp: (Date.now() / 1000).toString(), // Unix timestamp as string
        isDeleted: false,
        attachments: selectedAttachments,
        isOwn: true
      };
      
      // Add to UI immediately
      selectedChat.messages = [...selectedChat.messages, tempMessage];
      
      // Prepare data to send to API
      const messageData = {
        content: newMessage.trim(),
        message_id: tempId, // Send the temp ID so the server can link the response
        attachments: selectedAttachments.map(attachment => ({
          type: attachment.type,
          url: attachment.url
        }))
      };
      
      logger.debug(`Sending message to chat ${selectedChat.id}`, { tempId });
      
      // Clear the input fields immediately for better UX
      const sentContent = newMessage; // Keep a copy for error cases
      newMessage = '';
      const sentAttachments = [...selectedAttachments];
      selectedAttachments = [];
      
      // Call the API
      const response = await apiSendMessage(selectedChat.id, messageData);
      
      logger.debug(`Message send response received`, { response });
      
      if (response && response.message) {
        // Extract server-assigned message ID and data
        const serverMsgId = response.message.id || response.message.message_id;
        
        if (serverMsgId) {
          logger.debug(`Message saved with server ID: ${serverMsgId}`);
          
          // Update the temporary message with server data
          selectedChat.messages = selectedChat.messages.map(m => {
            if (m.id === tempId) {
              // Update with server data
              return {
                ...m,
                id: serverMsgId,
                content: response.message.content || m.content,
                timestamp: (response.message.timestamp || m.timestamp).toString()
              };
            }
            return m;
          });
          
          // Update the last message in the chat list
          selectedChat.lastMessage = {
            content: response.message.content || sentContent || 'Sent an attachment',
            timestamp: response.message.timestamp?.toString() || (Date.now() / 1000).toString(),
            senderId: response.message.sender_id || authState.userId as string
          };
        }
      }
      
      // Force scroll to bottom to show new message
      setTimeout(() => {
        const messagesContainer = document.querySelector('.messages-container');
        if (messagesContainer) {
          messagesContainer.scrollTop = messagesContainer.scrollHeight;
        }
      }, 100);
      
    } catch (error) {
      const errorResponse = handleApiError(error);
      logger.error('Failed to send message:', errorResponse);
      toastStore.showToast('Failed to send message. Please try again.', 'error');
      
      // Remove the temporary message on error
      if (selectedChat) {
        selectedChat.messages = selectedChat.messages.filter(m => 
          !m.id.startsWith('temp-') || m.id === `temp-${Date.now()}`
        );
      }
    }
  }

  // Search chats and messages
  async function searchChats() {
    if (!searchQuery.trim()) {
      filteredChats = [...chats];
      userSearchResults = [];
      return;
    }

    const query = searchQuery.toLowerCase();
    logger.debug('Starting search with query:', { query });

    // First, filter local chats by name
    let results = chats.filter(chat => 
      chat.name.toLowerCase().includes(query)
    );
    logger.debug('Filtered local chats:', { count: results.length });
    
    // Search for users via API - this should search ALL users in the system, not just those we've chatted with
    try {
      logger.debug('Calling searchUsers API with query:', { query: searchQuery });
      const response = await searchUsers(searchQuery);
      
      logger.debug('Search users API response:', { 
        status: 'success',
        responseData: JSON.stringify(response)
      });
      
      // Handle different response formats - check both direct response.users and response.data.users
      const users = response?.users || (response?.data?.users || []);
      
      if (users && users.length > 0) {
        // Transform API user results to match our participant format
        userSearchResults = users.map(user => ({
          id: user.id,
          username: user.username,
          displayName: user.name,
          avatar: user.profile_picture_url,
          bio: user.bio,
          isVerified: user.is_verified,
          isFollowing: user.is_following,
        }));
        logger.debug('Retrieved users from API', { count: userSearchResults.length, users: userSearchResults.map(u => u.username) });
      } else {
        userSearchResults = [];
        logger.warn('No users found or invalid API response format', { response });
      }
    } catch (error) {
      logger.error('Error searching users:', error);
      userSearchResults = [];
      // Fallback to client-side searching if API fails
      const uniqueUsers = new Map<string, Participant>();
      
      chats.forEach(chat => {
        if (chat.participants && chat.participants.length > 0) {
          chat.participants.forEach(participant => {
            // Skip if it's the current user or already in the map
            if (participant.id === authState.userId || uniqueUsers.has(participant.id)) {
              return;
            }
            
            // Check if user matches search query
            const displayName = participant.displayName || participant.username || '';
            const username = participant.username || '';
            
            if (displayName.toLowerCase().includes(query) || 
                username.toLowerCase().includes(query)) {
              uniqueUsers.set(participant.id, participant);
            }
          });
        }
      });
      
      userSearchResults = Array.from(uniqueUsers.values());
      logger.debug('Using fallback search results', { count: userSearchResults.length });
    }
    
    // If we have a selected chat, also search messages
    if (selectedChat) {
      try {
        logger.debug('Searching messages in chat:', { chatId: selectedChat.id, query: searchQuery });
        const response = await searchMessages(selectedChat.id, searchQuery);
        
        logger.debug('Search messages API response:', { 
          status: 'success',
          responseData: JSON.stringify(response)
        });
        
        if (response && response.messages && response.messages.length > 0) {
          // If the current chat isn't in results but has matching messages, add it
          if (!results.some(chat => chat.id === selectedChat!.id)) {
            results.push(selectedChat);
            logger.debug('Added current chat to results due to matching messages', 
              { chatId: selectedChat.id, matchingMessages: response.messages.length });
          }
        }
      } catch (error) {
        logger.error('Error searching messages:', error);
      }
    }
    
    filteredChats = results;
    logger.debug('Search completed', { 
      query: searchQuery, 
      chatResults: filteredChats.length,
      userResults: userSearchResults.length,
      userDetails: userSearchResults.map(u => u.username).join(', ')
    });
  }

  // Start a new chat with a user
  async function startChatWithUser(user: Participant) {
    // Add more logging
    logger.debug('Starting chat with user', { userId: user.id, username: user.username });
    
    // Check if we already have a chat with this user
    const existingChat = chats.find(chat => 
      chat.type === 'individual' && 
      chat.participants.some(p => p.id === user.id)
    );
    
    if (existingChat) {
      // If chat exists, select it
      logger.debug('Found existing chat', { chatId: existingChat.id });
      selectChat(existingChat);
    } else {
      // Create a new chat with this user via API
      try {
        const chatData = {
          type: 'individual',
          participants: [user.id]
        };
        
        logger.debug('Creating new chat', { chatData });
        const response = await createChat(chatData);
        logger.debug('Create chat API response', { response });
        
        if (response && response.chat) {
          // Format the received chat to match our chat structure
          const newChat: Chat = {
            id: response.chat.id,
            name: user.displayName || user.username,
            type: 'individual',
            lastMessage: undefined, // Using undefined instead of null to satisfy TypeScript
            avatar: user.avatar,
            participants: [
              {
                id: user.id,
                username: user.username,
                displayName: user.displayName,
                avatar: user.avatar,
                isVerified: user.isVerified
              },
              // Add current user as participant for display name calculation
              {
                id: authState.userId as string,
                username: username,
                displayName: displayName,
                avatar: avatar,
                isVerified: false
              }
            ],
            messages: [],
            unreadCount: 0
          };
          
          logger.debug('Adding new chat to list', { newChatId: newChat.id });
          chats = [newChat, ...chats];
          filteredChats = [newChat, ...filteredChats];
          selectChat(newChat);
          
          toastStore.showToast(`Chat with ${user.displayName || user.username} created`, 'success');
        } else {
          logger.error('Invalid response format from create chat API', { response });
          toastStore.showToast(`Failed to create chat: Invalid response`, 'error');
        }
      } catch (error) {
        const errorDetail = handleApiError(error);
        logger.error('Failed to create chat:', errorDetail);
        toastStore.showToast(`Failed to create chat with ${user.displayName || user.username}`, 'error');
      }
      
      // Clear search
      clearSearch();
    }
  }

  // Clear search
  function clearSearch() {
    searchQuery = '';
    filteredChats = [...chats];
    userSearchResults = [];
    logger.debug('Search cleared');
  }

  // Handle file attachments
  function handleAttachment(type: 'image' | 'gif' | 'video') {
    // Create a file input element
    const fileInput = document.createElement('input');
    fileInput.type = 'file';
    
    if (type === 'image') {
      fileInput.accept = 'image/*';
    } else if (type === 'gif') {
      fileInput.accept = 'image/gif';
    } else if (type === 'video') {
      fileInput.accept = 'video/*';
    }
    
    // Set up the change handler
    fileInput.onchange = async (e) => {
      const target = e.target as HTMLInputElement;
      if (!target.files || target.files.length === 0) return;
      
      const file = target.files[0];
      
      try {
        // In a real implementation, you would upload this file to your server/storage
        // and get back a URL. For now, we'll use a local object URL.
        const localUrl = URL.createObjectURL(file);
        
        const attachment: Attachment = {
          id: `temp-${Date.now()}`,
          type,
          url: localUrl,
        };
        
        selectedAttachments = [...selectedAttachments, attachment];
        toastStore.showToast(`${type} attached`, 'success');
        
        logger.debug('Attachment added', { type, filename: file.name, size: file.size });
      } catch (error) {
        logger.error('Failed to process attachment', { error });
        toastStore.showToast('Failed to attach file. Please try again.', 'error');
      }
    };
    
    // Trigger the file dialog
    fileInput.click();
  }

  // Filter chats when search query changes
  $: if (searchQuery !== undefined) {
    searchChats();
  }

  // Define our own formatTimeAgo function to ensure timestamps display correctly
  function formatTimeAgo(timestamp: string | number): string {
    if (!timestamp) return '';
    
    let date: Date;
    
    // Convert the timestamp to a Date object based on type
    if (typeof timestamp === 'string') {
      // Try parsing as ISO string first
      date = new Date(timestamp);
      
      // If invalid date, try to parse as Unix timestamp in seconds
      if (isNaN(date.getTime())) {
        date = new Date(parseInt(timestamp) * 1000);
      }
    } else if (typeof timestamp === 'number') {
      if (timestamp < 31536000000) {
        date = new Date(timestamp * 1000);
      } else {
        date = new Date(timestamp);
      }
    } else {
      return '';
    }
    
    if (isNaN(date.getTime())) {
      return '';
    }
    
    const now = new Date();
    const diffInSeconds = Math.floor((now.getTime() - date.getTime()) / 1000);
    
    if (diffInSeconds < 60) {
      return 'Just now';
    }
    
    if (diffInSeconds < 3600) {
      const minutes = Math.floor(diffInSeconds / 60);
      return `${minutes}m`;
    }
    
    if (diffInSeconds < 86400) {
      const hours = Math.floor(diffInSeconds / 3600);
      return `${hours}h`;
    }
    
    // Less than a week
    if (diffInSeconds < 604800) {
      const days = Math.floor(diffInSeconds / 86400);
      return `${days}d`;
    }
    
    // Less than a month
    if (diffInSeconds < 2592000) {
      const weeks = Math.floor(diffInSeconds / 604800);
      return `${weeks}w`;
    }
    
    // Less than a year
    if (diffInSeconds < 31536000) {
      const months = Math.floor(diffInSeconds / 2592000);
      return `${months}mo`;
    }
    
    // More than a year
    const years = Math.floor(diffInSeconds / 31536000);
    return `${years}y`;
  }

  // Get a consistent color based on name
  function getAvatarColor(name: string): string {
    // Default colors
    const colors = [
      '#4F46E5', // indigo
      '#0EA5E9', // sky
      '#10B981', // emerald
      '#F59E0B', // amber
      '#EF4444', // red
      '#8B5CF6', // violet
      '#EC4899', // pink
      '#06B6D4', // cyan
    ];
    
    // Get a deterministic index based on the name
    let hash = 0;
    if (!name) name = 'Chat';
    for (let i = 0; i < name.length; i++) {
      hash = name.charCodeAt(i) + ((hash << 5) - hash);
    }
    
    // Convert to positive number and get index
    hash = Math.abs(hash);
    const index = hash % colors.length;
    
    return colors[index];
  }

  // Get the first letter or two initials from a name
  function getInitials(name: string): string {
    if (!name) return '?';
    
    // If the name contains spaces, get first letters of first and last words
    if (name.includes(' ')) {
      const parts = name.split(' ').filter(Boolean);
      if (parts.length >= 2) {
        return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase();
      }
    }
    
    // Otherwise just return the first letter
    return name[0].toUpperCase();
  }
</script>

<div class="message-container {chatSelectedClass}">
  <!-- Left navigation/profile -->
  <div class="left-sidebar">
    <LeftSide username={username} displayName={displayName} avatar={avatar} />
  </div>

  <!-- Middle: Chat List / Search -->
  <div class="middle-section">
    <div class="section-header">
      <h1>Messages</h1>
    </div>
    <div class="search-container">
      <div class="search-input-wrapper">
        <input
          type="text"
          bind:value={searchQuery}
          placeholder="Search Direct Messages"
          class="search-input"
        />
        {#if searchQuery.trim() !== ''}
          <button class="clear-search-button" on:click={clearSearch}>
            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" width="16" height="16">
              <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clip-rule="evenodd" />
            </svg>
          </button>
        {/if}
      </div>
      
      {#if searchQuery.trim() !== '' && (userSearchResults.length > 0 || filteredChats.length > 0)}
        <div class="search-dropdown">
          {#if userSearchResults.length > 0}
            <div class="search-dropdown-section">
              <h4 class="search-dropdown-title">Users</h4>
              <ul class="search-dropdown-list">
                {#each userSearchResults as user}
                  <li>
                    <button class="dropdown-item" on:click={() => startChatWithUser(user)}>
                      <div class="avatar-container" style="background-color: {getAvatarColor(user.displayName || user.username)}">
                        {#if user.avatar}
                          <img src={user.avatar} alt={user.displayName || user.username} class="avatar-image" />
                        {:else}
                          <span class="avatar-placeholder">{getInitials(user.displayName || user.username)}</span>
                        {/if}
                      </div>
                      <div class="user-info">
                        <span class="user-name">{user.displayName || user.username}</span>
                        {#if user.username && user.username !== user.displayName}
                          <span class="user-username">@{user.username}</span>
                        {/if}
                      </div>
                    </button>
                  </li>
                {/each}
              </ul>
            </div>
          {/if}
          
          {#if filteredChats.length > 0}
            <div class="search-dropdown-section">
              <h4 class="search-dropdown-title">Conversations</h4>
              <ul class="search-dropdown-list">
                {#each filteredChats as chat}
                  <li>
                    <button class="dropdown-item" on:click={() => selectChat(chat)}>
                      <div class="avatar-container" style="background-color: {getAvatarColor(getChatDisplayName(chat))}">
                        {#if chat.avatar}
                          <img src={chat.avatar} alt={chat.name} class="avatar-image" />
                        {:else}
                          <span class="avatar-placeholder">{getInitials(getChatDisplayName(chat))}</span>
                        {/if}
                      </div>
                      <div class="user-info">
                        <span class="user-name">{getChatDisplayName(chat)}</span>
                        <div class="chat-preview">
                          {#if chat.lastMessage && chat.lastMessage.content}
                            {chat.lastMessage.content.substring(0, 30)}{chat.lastMessage.content.length > 30 ? '...' : ''}
                          {:else}
                            No messages yet
                          {/if}
                        </div>
                      </div>
                    </button>
                  </li>
                {/each}
              </ul>
            </div>
          {/if}
        </div>
      {/if}
    </div>
    <div class="chat-list">
      {#if isLoadingChats}
        <div class="loading-message">Loading chats...</div>
      {:else if userSearchResults.length === 0 && filteredChats.length === 0}
        <div class="empty-message">
          {chats.length === 0 ? 'No conversations yet' : 'No results found'}
        </div>
      {:else}
        {#if userSearchResults.length > 0}
          <div class="search-results-section">
            <h3 class="search-section-title">Users</h3>
            <ul class="user-results">
              {#each userSearchResults as user}
                <li>
                  <button class="user-result-item" on:click={() => startChatWithUser(user)}>
                    <div class="avatar-container" style="background-color: {getAvatarColor(user.displayName || user.username)}">
                      {#if user.avatar}
                        <img src={user.avatar} alt={user.displayName || user.username} class="avatar-image" />
                      {:else}
                        <span class="avatar-placeholder">{getInitials(user.displayName || user.username)}</span>
                      {/if}
                    </div>
                    <div class="user-info">
                      <span class="user-name">{user.displayName || user.username}</span>
                      {#if user.username && user.username !== user.displayName}
                        <span class="user-username">@{user.username}</span>
                      {/if}
                    </div>
                  </button>
                </li>
              {/each}
            </ul>
          </div>
        {/if}
        
        {#if filteredChats.length > 0}
          <div class="search-results-section">
            {#if userSearchResults.length > 0}
              <h3 class="search-section-title">Chats</h3>
            {/if}
            <ul class="chat-items">
              {#each filteredChats as chat}
                <li>
                  <button
                    class="chat-item {selectedChat?.id === chat.id ? 'active' : ''}"
                    on:click={() => selectChat(chat)}
                  >
                    <div class="avatar-container" style="background-color: {getAvatarColor(getChatDisplayName(chat))}">
                      {#if chat.avatar}
                        <img src={chat.avatar} alt={chat.name} class="avatar-image" />
                      {:else}
                        <span class="avatar-placeholder">{getInitials(getChatDisplayName(chat))}</span>
                      {/if}
                    </div>
                    <div class="chat-info">
                      <div class="chat-header">
                        <span class="chat-name">{getChatDisplayName(chat)}</span>
                        <span class="chat-time">{chat.lastMessage ? formatTimeAgo(chat.lastMessage.timestamp) : ''}</span>
                      </div>
                      <div class="chat-preview">
                        {#if chat.lastMessage && chat.lastMessage.content}
                          {chat.lastMessage.content.substring(0, 30)}{chat.lastMessage.content.length > 30 ? '...' : ''}
                        {:else}
                          No messages yetasdasd
                        {/if}
                      </div>
                    </div>
                    {#if chat.unreadCount > 0}
                      <span class="unread-badge">{chat.unreadCount}</span>
                    {/if}
                  </button>
                </li>
              {/each}
            </ul>
          </div>
        {/if}
      {/if}
    </div>
  </div>

  <!-- Right: Chat Content -->
  <div class="right-section">
    {#if selectedChat}
      <!-- Chat header -->
      <div class="chat-header">
        <div class="chat-avatar" style="background-color: {getAvatarColor(getChatDisplayName(selectedChat))}">
          {#if selectedChat.avatar}
            <img src={selectedChat.avatar} alt={selectedChat.name} class="avatar-image" />
          {:else}
            <span class="avatar-placeholder">{getInitials(getChatDisplayName(selectedChat))}</span>
          {/if}
        </div>
        <div class="chat-title">
          <h2>{getChatDisplayName(selectedChat)}</h2>
          <p class="group-info">{selectedChat.type === 'group' ? `${selectedChat.participants.length} members` : ''}</p>
        </div>
      </div>
      
      <!-- Messages -->
      <div class="messages-container">
        {#each selectedChat.messages as message}
          <div class="message-wrapper {message.isOwn ? 'own-message' : 'other-message'}">
            <div class="message-bubble">
              {#if !message.isOwn}
                <div class="sender-name">{message.senderName}</div>
              {/if}
              <div class={message.isDeleted ? 'deleted-message' : ''}>
                {message.isDeleted ? 'This message was deleted' : message.content}
              </div>
              {#if message.attachments.length > 0}
                <div class="attachments">
                  {#each message.attachments as attachment}
                    {#if attachment.type === 'image' || attachment.type === 'gif'}
                      <img src={attachment.url} alt="Attachment" class="attachment-image" />
                    {:else if attachment.type === 'video'}
                      <video src={attachment.url} controls class="attachment-video">
                        <track kind="captions" src="path-to-captions.vtt" label="English" default />
                        Your browser does not support the video tag.
                      </video>
                    {/if}
                  {/each}
                </div>
              {/if}
              <div class="message-meta">
                {formatTimeAgo(message.timestamp)}
                {#if message.isOwn && !message.isDeleted && isWithinTime(message.timestamp)}
                  <button class="unsend-button" on:click={() => unsendMessage(message.id)}>
                    Unsend
                  </button>
                {/if}
              </div>
            </div>
          </div>
        {/each}
      </div>
      
      <!-- Message input -->
      <div class="message-input-container">
        <div class="message-actions">
          <button class="action-button" on:click={() => clearSearch()} aria-label="Clear search">
            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor">
              <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clip-rule="evenodd" />
            </svg>
          </button>
          <button class="action-button" on:click={() => handleAttachment('image')} aria-label="Attach image">
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
            </svg>
          </button>
          <button class="action-button" on:click={() => handleAttachment('gif')} aria-label="Attach GIF">
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z" />
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
          </button>
        </div>
        <input
          type="text"
          bind:value={newMessage}
          placeholder="Start a new message"
          class="message-input"
          on:keydown={(e) => e.key === 'Enter' && sendMessage()}
        />
        <button
          class="send-button"
          on:click={sendMessage}
          disabled={!newMessage.trim()}
          aria-label="Send message"
        >
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8" />
          </svg>
        </button>
      </div>
    {:else}
      <div class="empty-chat">
        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" class="empty-chat-icon">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M8 10h.01M12 10h.01M16 10h.01M9 16H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-5l-5 5v-5z" />
        </svg>
        <h2>Select a conversation</h2>
        <p>Choose from your existing chats or start a new conversation to begin messaging.</p>
        <button class="new-message-button" on:click={() => searchQuery = 'new'}>
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" width="20" height="20">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
          </svg>
          New message
        </button>
      </div>
    {/if}
  </div>
</div>

<style>
  /* Main Container */
  .message-container {
    display: grid;
    grid-template-columns: 288px 300px 1fr;
    height: 100vh;
    background-color: var(--background-color, white);
    color: var(--text-color, black);
  }

  /* Dark mode overrides */
  :global(.dark) .message-container {
    --background-color: black;
    --text-color: white;
    --border-color: #2d3748;
    --hover-bg: #1a202c;
    --active-bg: rgba(29, 78, 216, 0.2);
    --message-bg: #2d3748;
    --own-message-bg: #3b82f6;
    --input-bg: #1a202c;
  }

  /* Light mode variables */
  .message-container {
    --border-color: #e2e8f0;
    --hover-bg: #f7fafc;
    --active-bg: rgba(59, 130, 246, 0.1);
    --message-bg: #e2e8f0;
    --own-message-bg: #3b82f6;
    --input-bg: #f7fafc;
  }

  /* Left Sidebar */
  .left-sidebar {
    border-right: 1px solid var(--border-color);
    height: 100%;
    overflow-y: auto;
    min-width: 288px;
  }

  /* Middle Section */
  .middle-section {
    display: flex;
    flex-direction: column;
    border-right: 1px solid var(--border-color);
    height: 100%;
  }

  .section-header {
    padding: 16px;
    border-bottom: 1px solid var(--border-color);
  }

  .section-header h1 {
    font-size: 1.25rem;
    font-weight: bold;
  }

  .search-container {
    padding: 12px;
    border-bottom: 1px solid var(--border-color);
    position: relative;
  }

  .search-input-wrapper {
    position: relative;
    display: flex;
    align-items: center;
  }

  .search-input {
    width: 100%;
    padding: 8px 16px;
    border-radius: 9999px;
    border: 1px solid var(--border-color);
    background-color: var(--input-bg);
    color: var(--text-color);
  }

  .clear-search-button {
    position: absolute;
    right: 8px;
    background: none;
    border: none;
    padding: 4px;
    color: var(--text-secondary, #6c757d);
    cursor: pointer;
    border-radius: 50%;
  }

  .clear-search-button:hover {
    background-color: var(--hover-bg);
  }

  .search-dropdown {
    position: absolute;
    top: calc(100% + 4px);
    left: 12px;
    right: 12px;
    max-height: 350px;
    overflow-y: auto;
    background-color: var(--background-color);
    border: 1px solid var(--border-color);
    border-radius: 8px;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
    z-index: 1000;
  }

  .search-dropdown-section {
    padding: 8px;
  }

  .search-dropdown-section + .search-dropdown-section {
    border-top: 1px solid var(--border-color);
  }

  .search-dropdown-title {
    font-size: 12px;
    color: var(--text-secondary, #6c757d);
    text-transform: uppercase;
    letter-spacing: 0.5px;
    margin: 4px 8px 8px;
  }

  .dropdown-item {
    padding: 8px;
    border-radius: 6px;
    margin-bottom: 2px;
    transition: background-color 0.15s ease;
  }

  .dropdown-item:hover {
    background-color: var(--hover-bg);
  }

  .dropdown-item .avatar-container {
    width: 36px;
    height: 36px;
    margin-right: 10px;
  }

  .dropdown-item .user-name {
    font-weight: 600;
    display: block;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .dropdown-item .user-username,
  .dropdown-item .chat-preview {
    font-size: 12px;
    color: var(--text-secondary, #6c757d);
    display: block;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .chat-info {
    flex: 1;
    min-width: 0;
  }

  .chat-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .search-container {
    padding: 12px;
    border-bottom: 1px solid var(--border-color);
    position: relative;
  }

  .search-input-wrapper {
    position: relative;
    display: flex;
    align-items: center;
  }

  .chat-name {
    font-weight: 600;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .chat-time {
    font-size: 0.75rem;
    color: gray;
    margin-left: 8px;
    flex-shrink: 0;
  }

  .chat-preview {
    color: gray;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    font-size: 0.875rem;
  }

  .unread-badge {
    background-color: #3b82f6;
    color: white;
    font-size: 0.75rem;
    width: 20px;
    height: 20px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    margin-left: 8px;
    flex-shrink: 0;
  }

  /* Right Section */
  .right-section {
    display: flex;
    flex-direction: column;
    height: 100%;
  }

  .chat-header {
    display: flex;
    align-items: center;
    padding: 16px;
    border-bottom: 1px solid var(--border-color);
  }

  .chat-avatar {
    width: 40px;
    height: 40px;
    border-radius: 50%;
    overflow: hidden;
    background-color: #6b7280;
    display: flex;
    align-items: center;
    justify-content: center;
    margin-right: 12px;
  }

  .chat-title h2 {
    font-weight: bold;
    margin: 0;
  }

  .group-info {
    font-size: 0.875rem;
    color: gray;
    margin: 0;
  }

  .messages-container {
    flex: 1;
    overflow-y: auto;
    padding: 16px;
    background-color: var(--bg-light, #f9fafb);
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  :global(.dark) .messages-container {
    background-color: var(--bg-dark, #121212);
  }

  .message-wrapper {
    display: flex;
    animation: fade-in 0.3s ease;
  }

  @keyframes fade-in {
    from { opacity: 0; transform: translateY(10px); }
    to { opacity: 1; transform: translateY(0); }
  }

  .own-message {
    justify-content: flex-end;
  }

  .other-message {
    justify-content: flex-start;
  }

  .message-bubble {
    max-width: 75%;
    padding: 10px 14px;
    border-radius: 18px;
    position: relative;
    box-shadow: 0 1px 2px rgba(0,0,0,0.1);
  }

  .own-message .message-bubble {
    background-color: var(--own-message-bg, #3b82f6);
    color: white;
    border-bottom-right-radius: 6px;
  }

  .other-message .message-bubble {
    background-color: var(--message-bg, #f3f4f6);
    color: var(--text-color, black);
    border-bottom-left-radius: 6px;
  }

  :global(.dark) .other-message .message-bubble {
    background-color: var(--message-bg-dark, #2d3748);
    color: white;
  }

  .sender-name {
    font-weight: 600;
    margin-bottom: 3px;
    font-size: 0.85rem;
    color: var(--text-secondary, #6c757d);
  }

  .deleted-message {
    font-style: italic;
    color: var(--text-tertiary, #9ca3af);
    font-size: 0.9rem;
  }

  .attachments {
    margin-top: 8px;
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .attachment-image, .attachment-video {
    max-width: 100%;
    border-radius: 8px;
  }

  .message-meta {
    display: flex;
    align-items: center;
    font-size: 0.7rem;
    margin-top: 4px;
    color: rgba(255, 255, 255, 0.8);
    justify-content: flex-end;
  }

  .other-message .message-meta {
    color: var(--text-secondary, #6c757d);
  }

  .unsend-button {
    margin-left: 6px;
    background: none;
    border: none;
    color: inherit;
    text-decoration: underline;
    cursor: pointer;
    padding: 0;
    font-size: 0.7rem;
    opacity: 0.8;
    transition: opacity 0.2s;
  }

  .unsend-button:hover {
    opacity: 1;
  }

  .message-input-container {
    display: flex;
    align-items: center;
    padding: 12px 16px;
    border-top: 1px solid var(--border-color);
    background-color: var(--bg-light, white);
  }

  :global(.dark) .message-input-container {
    background-color: var(--bg-dark, #121212);
  }

  .message-actions {
    display: flex;
    margin-right: 8px;
  }

  .action-button {
    background: none;
    border: none;
    color: var(--text-secondary, #6c757d);
    width: 36px;
    height: 36px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    margin-right: 4px;
    transition: all 0.2s;
  }

  .action-button:hover {
    background-color: var(--hover-bg, #f7fafc);
    color: var(--own-message-bg, #3b82f6);
  }

  .message-input {
    flex: 1;
    padding: 10px 16px;
    border-radius: 24px;
    border: 1px solid var(--border-color, #e2e8f0);
    background-color: var(--input-bg, #f7fafc);
    color: var(--text-color, black);
    font-size: 0.95rem;
    transition: border-color 0.2s;
  }

  .message-input:focus {
    outline: none;
    border-color: var(--own-message-bg, #3b82f6);
  }

  .send-button {
    width: 40px;
    height: 40px;
    border-radius: 50%;
    background-color: var(--own-message-bg, #3b82f6);
    color: white;
    border: none;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    margin-left: 8px;
    transition: background-color 0.2s, transform 0.1s;
  }

  .send-button:hover {
    background-color: var(--own-message-bg-hover, #2563eb);
    transform: scale(1.05);
  }

  .send-button:active {
    transform: scale(0.95);
  }

  .send-button[disabled] {
    opacity: 0.5;
    cursor: not-allowed;
    transform: none;
  }

  .send-button svg {
    width: 20px;
    height: 20px;
  }

  .empty-chat {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 100%;
    text-align: center;
    padding: 0 16px;
    color: var(--text-secondary, #6c757d);
    background-color: var(--bg-light, #f9fafb);
  }

  :global(.dark) .empty-chat {
    background-color: var(--bg-dark, #121212);
    color: #e5e7eb;
  }

  .empty-chat h2 {
    font-size: 1.5rem;
    font-weight: bold;
    margin-bottom: 10px;
    color: var(--text-primary, #1f2937);
  }

  :global(.dark) .empty-chat h2 {
    color: white;
  }

  .empty-chat p {
    color: var(--text-secondary, #6c757d);
    margin-bottom: 24px;
    max-width: 400px;
    line-height: 1.5;
  }

  .new-message-button {
    background-color: var(--own-message-bg, #3b82f6);
    color: white;
    font-weight: 600;
    padding: 10px 24px;
    border-radius: 9999px;
    border: none;
    cursor: pointer;
    transition: background-color 0.2s;
  }

  .new-message-button:hover {
    background-color: var(--own-message-bg-hover, #2563eb);
  }

  .empty-message {
    text-align: center;
    padding: 2rem;
    color: var(--text-secondary, #6c757d);
    font-style: italic;
  }

  .empty-chat-icon {
    width: 80px;
    height: 80px;
    margin-bottom: 20px;
    color: var(--own-message-bg, #3b82f6);
    opacity: 0.7;
  }
</style>
