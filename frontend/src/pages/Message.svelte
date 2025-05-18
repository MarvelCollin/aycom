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
  import { transformApiUsers, type StandardUser } from '../utils/userTransform';
  
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
  
  // Group chat modal state
  let showGroupChatModal = false;

  // WebSocket state
  let wsUnsubscribe: () => void;

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
  
  // Handle successful group chat creation
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
  
  // Format group chat data for display
  function formatGroupChatForDisplay(chatData: any): Chat {
    return {
      id: chatData.id,
      name: chatData.name || 'New Group Chat',
      type: 'group',
      lastMessage: undefined,
      avatar: null,
      participants: chatData.participants?.map(p => ({
        id: p.id || p.user_id,
        username: p.username || '',
        displayName: p.display_name || p.username || `User`,
        avatar: p.avatar_url || p.avatar || null,
        isVerified: p.is_verified || false
      })) || [],
      messages: [],
      unreadCount: 0
    };
  }

  // Message interfaces
  interface Message {
    id: string;
    content: string;
    timestamp: string;
    senderId: string;
    senderName: string;
    senderAvatar?: string;
    isOwn: boolean;
    isRead: boolean;
    isDeleted: boolean;
    attachments: Attachment[];
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
    lastMessage?: LastMessage;
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
  
  interface LastMessage {
    content: string;
    timestamp: string | number;
    senderId: string;
    senderName: string;
  }
  
  // Add user search results state
  let userSearchResults: StandardUser[] = [];

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
            if (chatToSelect && chatToSelect.id) {
              logger.debug('Restoring selected chat from localStorage', { chatId: savedChatId });
              selectChat(chatToSelect);
            } else {
              logger.debug('Saved chat ID not found in current chat list', { savedChatId });
              // Clear the invalid saved chat ID
              localStorage.removeItem('selectedChatId');
            }
          }
        } catch (error) {
          logger.warn('Failed to restore chat from localStorage', error);
          // Clear potentially corrupted localStorage
          localStorage.removeItem('selectedChatId');
        }
      });
    });

    // Setup WebSocket message handler
    setupWebSocketHandler();
  });

  onDestroy(() => {
    // Cleanup WebSocket connections when component is destroyed
    if (wsUnsubscribe) {
      wsUnsubscribe();
    }
    websocketStore.disconnectAll();
  });

  // Setup WebSocket message handler
  function setupWebSocketHandler() {
    // Unsubscribe from any existing handlers
    if (wsUnsubscribe) {
      wsUnsubscribe();
    }

    // Register a handler for incoming WebSocket messages
    wsUnsubscribe = websocketStore.registerMessageHandler(handleWebSocketMessage);
    logger.info('WebSocket message handler registered');
    
    // Check all existing connections and reconnect any that are closed
    if (selectedChat && selectedChat.id) {
      if (!websocketStore.isConnected(selectedChat.id)) {
        logger.info(`Reconnecting to WebSocket for chat ${selectedChat.id}`);
        websocketStore.connect(selectedChat.id);
      }
    }
  }

  // Handle incoming WebSocket messages
  function handleWebSocketMessage(message: any) {
    logger.debug('WebSocket message received', message);
    
    // Only process if we have a selected chat and message is related to it
    if (!selectedChat || !message.chat_id || message.chat_id !== selectedChat.id) {
      return;
    }

    // Handle different message types
    switch (message.type) {
      case 'text':
        handleIncomingTextMessage(message);
        break;
      case 'typing':
        // Could implement typing indicator here
        break;
      case 'read':
        handleReadReceipt(message);
        break;
      case 'edit':
        handleEditMessage(message);
        break;
      case 'delete':
        handleDeleteMessage(message);
        break;
      default:
        logger.warn('Unknown WebSocket message type', { type: message.type });
    }
  }

  // Handle incoming text message
  function handleIncomingTextMessage(message: any) {
    if (!selectedChat) return;

    // Only process if it's not our own message (those are handled during send)
    if (message.user_id === authState.userId) return;

    logger.debug('Processing incoming text message', { messageId: message.message_id });

    // Find sender info from participants
    const sender = selectedChat.participants.find(p => p.id === message.user_id);
    
    // Create new message object
    const newIncomingMessage: Message = {
      id: message.message_id,
      content: message.content || '',
      timestamp: message.timestamp ? message.timestamp.toString() : (Date.now() / 1000).toString(),
      senderId: message.user_id,
      senderName: sender ? (sender.displayName || sender.username) : `User ${message.user_id.substring(0, 4)}`,
      senderAvatar: sender ? sender.avatar || undefined : undefined,
      isOwn: false,
      isRead: false,
      isDeleted: false,
      attachments: message.attachments || []
    };

    // Add message to chat
    selectedChat.messages = [...selectedChat.messages, newIncomingMessage];

    // Update lastMessage in chat
    selectedChat.lastMessage = {
      content: message.content,
      timestamp: message.timestamp || (Date.now() / 1000),
      senderId: message.user_id,
      senderName: newIncomingMessage.senderName
    };

    // Force scroll to bottom to show new message
    setTimeout(() => {
      const messagesContainer = document.querySelector('.messages-container');
      if (messagesContainer) {
        messagesContainer.scrollTop = messagesContainer.scrollHeight;
      }
    }, 100);
  }

  // Handle read receipt
  function handleReadReceipt(message: any) {
    if (!selectedChat) return;

    // Mark messages as read
    selectedChat.messages = selectedChat.messages.map(msg => {
      if (!msg.isRead && message.message_ids && message.message_ids.includes(msg.id)) {
        return { ...msg, isRead: true };
      }
      return msg;
    });
  }

  // Handle edit message
  function handleEditMessage(message: any) {
    if (!selectedChat) return;

    // Update edited message
    selectedChat.messages = selectedChat.messages.map(msg => {
      if (msg.id === message.message_id) {
        return { 
          ...msg, 
          content: message.content,
          isEdited: true
        };
      }
      return msg;
    });

    // Update last message if this was the last one
    if (selectedChat.lastMessage && selectedChat.lastMessage.content) {
      const lastMsg = selectedChat.messages[selectedChat.messages.length - 1];
      if (lastMsg && lastMsg.id === message.message_id) {
        selectedChat.lastMessage.content = message.content;
      }
    }
  }

  // Handle delete message
  function handleDeleteMessage(message: any) {
    if (!selectedChat) return;

    // Mark message as deleted
    selectedChat.messages = selectedChat.messages.map(msg => {
      if (msg.id === message.message_id) {
        return { ...msg, isDeleted: true };
      }
      return msg;
    });

    // Update last message if needed
    if (selectedChat && selectedChat.lastMessage) {
      const lastMessageContent = selectedChat.lastMessage.content;
      
      if (selectedChat.messages.some(m => 
        m.id === message.message_id && 
        m.content === lastMessageContent)) {
        // Find the new last non-deleted message
        const lastNonDeletedMessage = [...selectedChat.messages]
          .reverse()
          .find(m => !m.isDeleted);
        
        if (lastNonDeletedMessage) {
          selectedChat.lastMessage = {
            content: lastNonDeletedMessage.content,
            timestamp: lastNonDeletedMessage.timestamp,
            senderId: lastNonDeletedMessage.senderId,
            senderName: lastNonDeletedMessage.senderName
          };
        } else {
          selectedChat.lastMessage = {
            content: '',
            timestamp: Date.now() / 1000,
            senderId: '',
            senderName: ''
          };
        }
      }
    }
  }

  // Fetch chats
  async function fetchChats() {
    isLoadingChats = true;
    try {
      const response = await getChatHistoryList();
      
      // Add more detailed logging to debug the response structure
      logger.debug('Raw chat history response:', response);
      
      // Determine where the chats array is located in the response
      let chatArray: any[] = [];
      if (response && response.chats && Array.isArray(response.chats)) {
        // Standard format: { chats: [...] }
        chatArray = response.chats;
        logger.debug(`Found ${chatArray.length} chats in response.chats`);
        
        // Debug first chat object to see all properties
        if (chatArray.length > 0) {
          logger.debug('First chat object structure:', JSON.stringify(chatArray[0], null, 2));
        }
      } else if (response && Array.isArray(response)) {
        // Alternative format: direct array
        chatArray = response;
        logger.debug(`Found ${chatArray.length} chats in direct array response`);
        
        // Debug first chat object to see all properties
        if (chatArray.length > 0) {
          logger.debug('First chat object structure:', JSON.stringify(chatArray[0], null, 2));
        }
      } else if (response && typeof response === 'object') {
        // Try to find an array property that might contain chats
        const possibleChatArrays = Object.entries(response)
          .filter(([_, value]) => Array.isArray(value) && value.length > 0)
          .map(([key, value]) => ({ key, value: value as any[] }));
        
        if (possibleChatArrays.length > 0) {
          // Use the first array property found
          const firstArray = possibleChatArrays[0];
          chatArray = firstArray.value;
          logger.debug(`Found ${chatArray.length} chats in response.${firstArray.key}`);
          
          // Debug first chat object to see all properties
          if (chatArray.length > 0) {
            logger.debug('First chat object structure:', JSON.stringify(chatArray[0], null, 2));
          }
        } else {
          logger.warn('No chat arrays found in response object', response);
        }
      }
      
      // Debug: Log all users with chats and their last messages
      console.log('===== CHAT DEBUG INFO =====');
      console.log('All chats received from API:', chatArray);
      
      if (chatArray.length > 0) {
        console.log(`Found ${chatArray.length} chats`);
        
        // Track unique users across all chats
        const allUsers = new Map();
        
        chatArray.forEach((chat, index) => {
          const chatId = chat.id || chat.Id;
          
          // Skip chats without an ID
          if (!chatId) {
            logger.warn(`Skipping chat at index ${index} because it has no ID`);
            return;
          }
          
          console.log(`\nChat #${index + 1} (ID: ${chatId}):`)
          console.log(`- Is group: ${chat.is_group_chat || false}`);
          console.log(`- Name: ${chat.name || 'Unnamed'}`);
          
          // Log participants
          if (Array.isArray(chat.participants)) {
            console.log(`- Participants (${chat.participants.length}):`);
            chat.participants.forEach(p => {
              const userId = p.id || p.user_id || '';
              const username = p.username || '';
              const displayName = p.display_name || p.displayName || '';
              
              console.log(`  * User: ${displayName || username || 'Unknown'} (ID: ${userId})`);
              
              // Add to all users map
              if (userId && !allUsers.has(userId)) {
                allUsers.set(userId, { 
                  username: username, 
                  displayName: displayName 
                });
              }
            });
          } else {
            console.log('- No participants data available');
          }
          
          // Log last message
          if (chat.last_message) {
            console.log('- Last message:');
            if (typeof chat.last_message === 'string') {
              console.log(`  * Content: ${chat.last_message}`);
            } else {
              console.log(`  * Content: ${chat.last_message.content || ''}`);
              console.log(`  * Sender: ${chat.last_message.sender_id || ''}`);
              console.log(`  * Timestamp: ${chat.last_message.timestamp || ''}`);
            }
          } else if (chat.lastMessage) {
            console.log('- Last message:');
            if (typeof chat.lastMessage === 'string') {
              console.log(`  * Content: ${chat.lastMessage}`);
            } else {
              console.log(`  * Content: ${chat.lastMessage.content || chat.lastMessage.Content || ''}`);
              console.log(`  * Sender: ${chat.lastMessage.sender_id || chat.lastMessage.SenderId || ''}`);
              console.log(`  * Timestamp: ${chat.lastMessage.timestamp || chat.lastMessage.Timestamp || ''}`);
            }
          } else {
            console.log('- No last message available');
          }
        });
        
        // Log summary of all users
        console.log('\nAll Users Summary:');
        console.log(`Total unique users: ${allUsers.size}`);
        allUsers.forEach((userData, userId) => {
          console.log(`- ${userData.displayName || userData.username || 'Unknown'} (ID: ${userId})`);
        });
        console.log('===== END DEBUG INFO =====');
        
        // Filter out any chats without valid IDs before mapping
        const validChats = chatArray.filter(chat => {
          // Check all possible ID field names
          const chatId = chat.id || chat.Id || chat.chat_id;
          
          if (!chatId) {
            logger.warn('Ignoring chat without a valid ID', chat);
            return false;
          }
          
          // Make sure the id field is set in expected format for downstream code
          if (!chat.id && chatId) {
            chat.id = chatId;
          }
          
          return true;
        });
        
        chats = validChats.map((chat: any) => {
          // Process each chat as before
          // Get the chat ID - also check for chat_id
          const chatId = chat.id || chat.Id || chat.chat_id;
          logger.debug(`Processing chat ${chatId}`, chat);
          
          // Determine if this is a group chat
          const isGroup = chat.is_group_chat || chat.IsGroupChat || chat.is_group || chat.isGroup || false;
          
          // Process participants list
          let processedParticipants: Participant[] = [];
          if (Array.isArray(chat.participants)) { 
            // First pass: Create basic participants from available data
            processedParticipants = chat.participants.map((p: any) => {
              // Handle different property naming formats
              const userId = p.id || p.user_id || p.userId || '';
              const username = p.username || '';
              const displayName = p.display_name || p.displayName || p.name || p.username || '';
              const avatar = p.profile_picture_url || p.profilePictureUrl || p.avatar || null;
              const isVerified = p.is_verified || p.isVerified || false;
              
              return {
                id: userId,
                username: username,
                displayName: displayName,
                avatar: avatar,
                isVerified: isVerified
              };
            });
            
            // Second pass (async): Fetch missing user data and update
            processedParticipants.forEach(async (participant, index) => {
              // Only fetch data if username or displayName is missing
              if (!participant.username || !participant.displayName) {
                try {
                  logger.debug(`Fetching missing user data for participant ${participant.id}`);
                  const userData = await getUserById(participant.id);
                  
                  if (userData) {
                    // Update the participant with fetched data
                    processedParticipants[index] = {
                      ...participant,
                      username: userData.username || `user_${participant.id.substring(0, 4)}`,
                      displayName: userData.name || userData.display_name || `User ${participant.id.substring(0, 4)}`,
                      avatar: userData.profile_picture_url || participant.avatar,
                      isVerified: userData.is_verified || participant.isVerified
                    };
                    
                    logger.debug(`Updated participant data: ${JSON.stringify(processedParticipants[index])}`);
                    
                    // Force a refresh of the chat display
                    chats = [...chats];
                    filteredChats = [...chats];
                  }
                } catch (error) {
                  logger.error(`Failed to fetch user data for ${participant.id}`, error);
                }
              }
            });
          }
          
          // Try to determine chat name
          let chatName: string;
          let chatAvatar: string | null = null;
          
          // Get name directly if provided and valid
          if (chat.name && chat.name !== 'Chat' && chat.name !== 'null null' && chat.name !== 'New Chat') {
            chatName = chat.name;
          }
          // Check if there's a display_name property
          else if (chat.display_name && chat.display_name !== 'null null') {
            chatName = chat.display_name;
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
          let lastMessageData: LastMessage;
          if (chat.last_message) {
            if (typeof chat.last_message === 'string') {
              lastMessageData = {
                content: chat.last_message,
                timestamp: Date.now() / 1000,
                senderId: '',
                senderName: ''
              };
            } else {
              // Extract sender information
              const senderId = chat.last_message.sender_id || chat.last_message.user_id || '';
              let senderName = '';
              
              // Try to find sender in participant list first
              if (senderId) {
                const sender = processedParticipants.find(p => p.id === senderId);
                if (sender) {
                  senderName = sender.displayName || sender.username || '';
                }
                
                // If sender wasn't found or didn't have a name, try to fetch it
                if (!senderName && senderId !== authState.userId) {
                  // We'll try to fetch this asynchronously
                  getUserById(senderId).then(userData => {
                    if (userData) {
                      // Find this chat and update its last message sender
                      const chatIndex = chats.findIndex(c => c.id === chatId);
                      if (chatIndex >= 0 && chats[chatIndex].lastMessage) {
                        chats[chatIndex].lastMessage.senderName = userData.name || userData.display_name || userData.username || `User ${senderId.substring(0, 4)}`;
                        // Force UI update
                        chats = [...chats];
                        filteredChats = [...chats];
                      }
                    }
                  }).catch(err => {
                    logger.error(`Failed to fetch last message sender data for ${senderId}:`, err);
                  });
                }
                
                // For current user messages
                if (senderId === authState.userId) {
                  senderName = 'You';
                }
              }
              
              lastMessageData = {
                content: chat.last_message.content || '',
                timestamp: chat.last_message.timestamp || Date.now() / 1000,
                senderId: senderId,
                senderName: senderName
              };
            }
          } else if (chat.lastMessage) {
            if (typeof chat.lastMessage === 'string') {
              lastMessageData = {
                content: chat.lastMessage,
                timestamp: Date.now() / 1000,
                senderId: '',
                senderName: ''
              };
            } else {
              // Extract sender information
              const senderId = chat.lastMessage.sender_id || chat.lastMessage.SenderId || '';
              let senderName = '';
              
              // Try to find sender in participant list first
              if (senderId) {
                const sender = processedParticipants.find(p => p.id === senderId);
                if (sender) {
                  senderName = sender.displayName || sender.username || '';
                }
                
                // If sender wasn't found or didn't have a name, try to fetch it
                if (!senderName && senderId !== authState.userId) {
                  // We'll try to fetch this asynchronously
                  getUserById(senderId).then(userData => {
                    if (userData) {
                      // Find this chat and update its last message sender
                      const chatIndex = chats.findIndex(c => c.id === chatId);
                      if (chatIndex >= 0 && chats[chatIndex].lastMessage) {
                        chats[chatIndex].lastMessage.senderName = userData.name || userData.display_name || userData.username || `User ${senderId.substring(0, 4)}`;
                        // Force UI update
                        chats = [...chats];
                        filteredChats = [...chats];
                      }
                    }
                  }).catch(err => {
                    logger.error(`Failed to fetch last message sender data for ${senderId}:`, err);
                  });
                }
                
                // For current user messages
                if (senderId === authState.userId) {
                  senderName = 'You';
                }
              }
              
              lastMessageData = {
                content: chat.lastMessage.content || chat.lastMessage.Content || '',
                timestamp: chat.lastMessage.timestamp || chat.lastMessage.Timestamp || Date.now() / 1000,
                senderId: senderId,
                senderName: senderName
              };
            }
          } else {
            // Create a default last message when none exists
            lastMessageData = {
              content: '',
              timestamp: Date.now() / 1000,
              senderId: '',
              senderName: ''
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
    if (chat.name && chat.name !== 'Chat' && chat.name !== 'null null' && chat.name !== 'New Chat') {
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
        const displayName = otherParticipant.displayName || otherParticipant.username;
        if (displayName) {
          return displayName;
        } else {
          // If there's no name, try using just "User" with part of their ID
          return `User ${otherParticipant.id.substring(0, 4)}`;
        }
      }
      
      // If we couldn't find another participant, use the first participant
      const participant = chat.participants[0];
      const name = participant.displayName || participant.username;
      if (name) {
        return name;
      } else {
        return `User ${participant.id.substring(0, 4)}`;
      }
    }
    
    // If all else fails, use the chat ID to create a name
    if (chat.id) {
      return `Chat ${chat.id.substring(0, 6)}`;
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
            senderId: lastNonDeletedMessage.senderId,
            senderName: lastNonDeletedMessage.senderName
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
    // Disconnect from previous chat's WebSocket if any
    if (selectedChat && selectedChat.id) {
      websocketStore.disconnect(selectedChat.id);
    }

    // Validate that chat and chat.id exist and are valid
    if (!chat || !chat.id) {
      logger.error('Attempted to select a chat with invalid ID', { chat });
      toastStore.showToast('Error: Invalid chat selection', 'error');
      return;
    }

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
      
      console.log(`===== MESSAGES DEBUG INFO (Chat: ${chat.id}) =====`);
      console.log('Raw message response:', response);
      
      if (response && response.messages && Array.isArray(response.messages)) {
        console.log(`Received ${response.messages.length} messages for chat ${chat.id}`);
        
        // Log detailed information about each message
        response.messages.forEach((msg, index) => {
          console.log(`\nMessage #${index + 1}:`);
          console.log(`- ID: ${msg.id || msg.message_id || 'unknown'}`);
          console.log(`- Sender ID: ${msg.sender_id || msg.user_id || 'unknown'}`);
          console.log(`- Content: ${msg.content || 'empty'}`);
          console.log(`- Timestamp: ${msg.timestamp || 'none'}`);
          console.log(`- Is deleted: ${msg.is_deleted || false}`);
          
          // Log user data if available
          if (msg.user) {
            console.log(`- User data:`);
            console.log(`  * ID: ${msg.user.id || 'unknown'}`);
            console.log(`  * Username: ${msg.user.username || 'unknown'}`);
            console.log(`  * Display name: ${msg.user.display_name || 'unknown'}`);
          } else {
            console.log(`- No user data available`);
          }
        });
        
        console.log('===== END MESSAGES DEBUG INFO =====');
        
        // Create a map of user data we've fetched to avoid duplicate API calls
        const userDataCache = new Map();
        
        // Transform the messages
        const messagesPromises = response.messages.map(async (msg: any) => {
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
          
          // Extract user data if available
          let senderName = '';
          let senderAvatar = '';
          
          // Check if user data is provided directly in the message
          if (msg.user) {
            // Log the user data for debugging
            logger.debug(`Message ${id} has user data:`, {
              user_id: msg.user.id,
              username: msg.user.username || 'Not provided',
              display_name: msg.user.display_name || 'Not provided',
              profile_picture_url: msg.user.profile_picture_url || 'No profile picture'
            });
            
            // Use display_name or username from the message's user data
            senderName = msg.user.display_name || msg.user.username || `User ${senderId.substring(0, 4)}`;
            senderAvatar = msg.user.profile_picture_url || msg.user.avatar || '';
          } else {
            // Try to find the user in the participants list
            const senderParticipant = chat.participants.find(p => 
              p.id === senderId || (p.id === `${senderId}`)
            );
            
            if (senderParticipant) {
              senderName = senderParticipant.displayName || senderParticipant.username || `User ${senderId.substring(0, 4)}`;
              senderAvatar = senderParticipant.avatar || '';
            } else {
              // If user not in participants and not in message, fetch from API
              // Check if we already fetched this user
              if (!userDataCache.has(senderId)) {
                logger.debug(`Fetching user data for sender ${senderId}`);
                
                try {
                  // Fetch user data from API
                  const userData = await getUserById(senderId);
                  if (userData) {
                    userDataCache.set(senderId, userData);
                    logger.debug(`Retrieved user data for ${senderId}:`, userData);
                  }
                } catch (error) {
                  logger.error(`Failed to fetch user data for ${senderId}:`, error);
                }
              }
              
              // Use the cached user data if available
              const userData = userDataCache.get(senderId);
              if (userData) {
                senderName = userData.name || userData.display_name || userData.username || `User ${senderId.substring(0, 4)}`;
                senderAvatar = userData.profile_picture_url || '';
                
                logger.debug(`Using API data for user ${senderId}:`, {
                  name: senderName,
                  avatar: senderAvatar
                });
              } else {
                // Generate a name from the sender ID if no user data found
                senderName = senderId === authState.userId ? displayName : `User ${senderId.substring(0, 4)}`;
                senderAvatar = senderId === authState.userId ? (avatar || '') : '';
                
                logger.debug(`Using generated data for user ${senderId}`);
              }
            }
          }
          
          // Use the sender's name if it's the current user
          if (senderId === authState.userId) {
            senderName = displayName; // Use the logged-in user's display name
            senderAvatar = avatar || '';    // Use the logged-in user's avatar
          }
          
          logger.debug(`Message ${id} processed:`, { 
            sender: senderId,
            senderName: senderName,
            isCurrentUser: senderId === authState.userId
          });
          
          return {
            id: id,
            senderId: senderId,
            senderName: senderName,
            senderAvatar: senderAvatar,
            content: content,
            timestamp: timestamp.toString(),
            isDeleted: isDeleted,
            attachments: msg.attachments || [],
            isOwn: senderId === authState.userId
          };
        });
        
        // Wait for all user data fetching to complete
        selectedChat.messages = await Promise.all(messagesPromises);
        
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
      
      // Connect to WebSocket for this chat
      websocketStore.connect(chat.id);
      logger.info(`Connected to WebSocket for chat ${chat.id}`);
      
      // Check WebSocket connection status after a short delay
      setTimeout(() => {
        const isConnected = websocketStore.isConnected(chat.id);
        logger.info(`WebSocket connection status for chat ${chat.id}: ${isConnected ? 'Connected' : 'Not connected'}`);
        
        // Update UI to show connection status
        const wsStatusElement = document.querySelector('.ws-status');
        if (wsStatusElement) {
          wsStatusElement.setAttribute('data-connected', isConnected ? 'true' : 'false');
        }
      }, 500);
      
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
        isRead: true, // Own messages are already read
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
      
      // Make sure we have an active WebSocket connection
      let isWsConnected = websocketStore.isConnected(selectedChat.id);
      
      // If not connected, try to connect
      if (!isWsConnected) {
        logger.info(`WebSocket not connected for chat ${selectedChat.id}, attempting to connect`);
        websocketStore.connect(selectedChat.id);
        
        // Wait a moment to allow connection to establish
        await new Promise(resolve => setTimeout(resolve, 100));
        
        // Check connection status again
        isWsConnected = websocketStore.isConnected(selectedChat.id);
        logger.info(`WebSocket connection status after attempt: ${isWsConnected ? 'Connected' : 'Not connected'}`);
      }
      
      // Send via WebSocket if connection is active, otherwise fall back to API
      if (isWsConnected) {
        // Prepare message for WebSocket
        const wsMessage: ChatMessage = {
          type: 'text' as MessageType,
          content: sentContent,
          user_id: authState.userId as string,
          chat_id: selectedChat.id,
          timestamp: new Date(),
          message_id: tempId
        };
        
        // Send via WebSocket
        websocketStore.sendMessage(selectedChat.id, wsMessage);
        logger.info('Message sent via WebSocket', { chatId: selectedChat.id, tempId });
      } else {
        // Fall back to API if WebSocket is not connected
        logger.info('WebSocket not connected, sending via API', { chatId: selectedChat.id });
        
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
          }
        }
      }
      
      // Update the last message in the chat list
      selectedChat.lastMessage = {
        content: sentContent || 'Sent an attachment',
        timestamp: (Date.now() / 1000).toString(),
        senderId: authState.userId as string,
        senderName: 'You' // Current user is the sender
      };
      
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
    logger.info('Starting search with query:', { query });

    // First, filter local chats by name
    let results = chats.filter(chat => 
      chat.name.toLowerCase().includes(query) || 
      chat.participants.some(p => 
        (p.displayName && p.displayName.toLowerCase().includes(query)) || 
        (p.username && p.username.toLowerCase().includes(query))
      )
    );
    logger.info('Filtered local chats:', { count: results.length });
    
    // Search for users via API
    try {
      logger.info('Calling searchUsers API with query:', { query: searchQuery });
      const response = await searchUsers(searchQuery);
      
      logger.info('Search users API response:', { 
        status: 'success',
        userCount: response.users?.length || 0,
        totalResults: response.totalCount || 0
      });
      
      // Get users from the response
      const users = response?.users || [];
      
      if (users && users.length > 0) {
        // Transform API user results using our utility function
        userSearchResults = transformApiUsers(users);
        logger.info('Retrieved users from API', { 
          count: userSearchResults.length, 
          firstUser: userSearchResults[0]?.displayName
        });
      } else {
        logger.warn('No users found from API, trying to get all users', { query: searchQuery });
        
        // Try to get all users from the new all users endpoint
        try {
          // First try to use the all users endpoint
          const allUsersResponse = await getAllUsers(30, 1, 'username', true);
          
          if (allUsersResponse.users && allUsersResponse.users.length > 0) {
            // Filter users client-side based on the query
            const filteredUsers = allUsersResponse.users.filter(user => 
              (user.username && user.username.toLowerCase().includes(query)) ||
              (user.display_name && user.display_name.toLowerCase().includes(query))
            );
            
            if (filteredUsers.length > 0) {
              userSearchResults = transformApiUsers(filteredUsers);
              logger.info('Retrieved filtered users from all users list', { 
                count: userSearchResults.length 
              });
            } else {
              // If no users match the filter, use client-side fallback
              findUsersFromLocalChats(query);
            }
          } else {
            // If no users returned from all users API, try with basic search as fallback
            logger.warn('No users returned from all users endpoint, trying search endpoint with empty query');
            const fallbackResponse = await searchUsers(" ");
            
            if (fallbackResponse.users && fallbackResponse.users.length > 0) {
              // Filter users client-side based on the query
              const filteredUsers = fallbackResponse.users.filter(user => 
                (user.username && user.username.toLowerCase().includes(query)) ||
                (user.display_name && user.display_name.toLowerCase().includes(query))
              );
              
              if (filteredUsers.length > 0) {
                userSearchResults = transformApiUsers(filteredUsers);
                logger.info('Retrieved filtered users from search with empty query', { 
                  count: userSearchResults.length 
                });
              } else {
                findUsersFromLocalChats(query);
              }
            } else {
              findUsersFromLocalChats(query);
            }
          }
        } catch (fallbackError) {
          logger.error('Error getting all users:', fallbackError);
          
          // Try with basic search as fallback
          try {
            logger.warn('Falling back to search endpoint with empty query');
            const fallbackResponse = await searchUsers(" ");
            
            if (fallbackResponse.users && fallbackResponse.users.length > 0) {
              // Filter users client-side based on the query
              const filteredUsers = fallbackResponse.users.filter(user => 
                (user.username && user.username.toLowerCase().includes(query)) ||
                (user.display_name && user.display_name.toLowerCase().includes(query))
              );
              
              if (filteredUsers.length > 0) {
                userSearchResults = transformApiUsers(filteredUsers);
                logger.info('Retrieved filtered users from search with empty query', { 
                  count: userSearchResults.length 
                });
              } else {
                findUsersFromLocalChats(query);
              }
            } else {
              findUsersFromLocalChats(query);
            }
          } catch (secondFallbackError) {
            logger.error('Error with search fallback:', secondFallbackError);
            findUsersFromLocalChats(query);
          }
        }
      }
    } catch (error) {
      logger.error('Error searching users:', error);
      userSearchResults = [];
      
      // Fallback to client-side searching if API fails
      findUsersFromLocalChats(query);
    }
    
    // If we have a selected chat, also search messages
    if (selectedChat) {
      try {
        logger.debug('Searching messages in chat:', { chatId: selectedChat.id, query: searchQuery });
        const response = await searchMessages(selectedChat.id, searchQuery);
        
        logger.debug('Search messages API response:', { 
          status: 'success',
          messageCount: response.messages?.length || 0
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
    
    // Update the filtered chats to trigger UI update
    filteredChats = [...results];
    
    logger.info('Search completed', { 
      query: searchQuery, 
      chatResults: filteredChats.length,
      userResults: userSearchResults.length
    });
  }
  
  // Helper function to find users from local chats
  function findUsersFromLocalChats(query: string) {
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
    
    // Convert participants to StandardUser format if we found any matches
    if (uniqueUsers.size > 0) {
      userSearchResults = Array.from(uniqueUsers.values()).map(p => ({
        id: p.id,
        username: p.username,
        displayName: p.displayName,
        avatar: p.avatar,
        isVerified: p.isVerified
      }));
      logger.info('Using local chat participants for search results', { count: userSearchResults.length });
    } else {
      userSearchResults = [];
    }
  }

  // Start a new chat with a user
  async function startChatWithUser(user: StandardUser) {
    // Convert StandardUser to Participant for compatibility
    const participant: Participant = {
      id: user.id,
      username: user.username,
      displayName: user.displayName,
      avatar: user.avatar,
      isVerified: user.isVerified
    };
    
    // Add more logging
    logger.debug('Starting chat with user', { userId: user.id, username: user.username });
    
    // Check if we already have a chat with this user
    const existingChat = chats.find(chat => 
      chat.type === 'individual' && 
      chat.participants.some(p => p.id === user.id)
    );
    
    if (existingChat) {
      // Validate the existing chat has an ID
      if (!existingChat.id) {
        logger.error('Found existing chat but it has no ID', { existingChat });
        toastStore.showToast('Error finding chat. Please try again.', 'error');
        return;
      }
      
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
      <div class="button-group">
        <button class="compose-button" on:click={() => showGroupChatModal = true}>
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" width="20" height="20">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
          </svg>
          <span>Group</span>
        </button>
        <button class="compose-button" on:click={() => searchQuery = 'new'}>
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" width="20" height="20">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
          </svg>
          <span>New</span>
        </button>
      </div>
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
              <h4 class="search-dropdown-title">Users ({userSearchResults.length})</h4>
              <ul class="search-dropdown-list">
                {#each userSearchResults as user}
                  <li>
                    <button class="dropdown-item" on:click={() => startChatWithUser(user)}>
                      <div class="avatar-container" style="background-color: {getAvatarColor(user.displayName || user.username)}">
                        {#if user.avatar}
                          <img src={user.avatar} alt={user.displayName || user.username} class="avatar-image" />
                        {:else}
                          <span class="avatar-placeholder">{(user.displayName || user.username).substring(0, 1).toUpperCase()}</span>
                        {/if}
                      </div>
                      <div class="user-info">
                        <span class="user-name">{user.displayName || user.username}</span>
                        {#if user.username && user.username !== user.displayName}
                          <span class="user-username">@{user.username}</span>
                        {/if}
                      </div>
                      <div class="start-chat-btn">
                        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" width="16" height="16">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
                        </svg>
                        <span>Chat</span>
                      </div>
                    </button>
                  </li>
                {/each}
              </ul>
            </div>
          {/if}
          
          {#if filteredChats.length > 0}
            <div class="search-dropdown-section">
              <h4 class="search-dropdown-title">Conversations ({filteredChats.length})</h4>
              <ul class="search-dropdown-list">
                {#each filteredChats as chat}
                  <li>
                    <button class="dropdown-item" on:click={() => selectChat(chat)}>
                      <div class="avatar-container" style="background-color: {getAvatarColor(getChatDisplayName(chat))}">
                        {#if chat.avatar}
                          <img src={chat.avatar} alt={chat.name} class="avatar-image" />
                        {:else}
                          <span class="avatar-placeholder">{getChatDisplayName(chat).substring(0, 1).toUpperCase()}</span>
                        {/if}
                        {#if chat.type === 'group'}
                          <span class="group-indicator">
                            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" width="12" height="12">
                              <path d="M13 6a3 3 0 11-6 0 3 3 0 016 0zM18 8a2 2 0 11-4 0 2 2 0 014 0zM14 15a4 4 0 00-8 0v3h8v-3zM6 8a2 2 0 11-4 0 2 2 0 014 0zM16 18v-3a5.972 5.972 0 00-.75-2.906A3.005 3.005 0 0119 15v3h-3zM4.75 12.094A5.973 5.973 0 004 15v3H1v-3a3 3 0 013.75-2.906z" />
                            </svg>
                          </span>
                        {/if}
                      </div>
                      <div class="chat-info">
                        <div class="chat-header">
                          <span class="chat-name">{getChatDisplayName(chat)}</span>
                          <span class="chat-time">{chat.lastMessage ? formatTimeAgo(chat.lastMessage.timestamp) : ''}</span>
                        </div>
                        <div class="chat-preview">
                          {#if chat.lastMessage && chat.lastMessage.content}
                            {#if chat.lastMessage.senderName}
                              <span class="message-sender">{chat.lastMessage.senderName}:</span>
                            {/if}
                            <span class="message-content">{chat.lastMessage.content.substring(0, 30)}{chat.lastMessage.content.length > 30 ? '...' : ''}</span>
                          {:else}
                            <span class="no-messages">No messages yet</span>
                          {/if}
                        </div>
                      </div>
                    </button>
                  </li>
                {/each}
              </ul>
            </div>
          {/if}
          
          {#if userSearchResults.length === 0 && filteredChats.length === 0}
            <div class="search-dropdown-section">
              <p class="no-results-message">No results found for "{searchQuery}"</p>
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
          {#if chats.length === 0}
            <div class="empty-state">
              <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" width="40" height="40">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
              </svg>
              <p>No conversations yet</p>
              <button class="start-chat-button" on:click={() => searchQuery = 'new'}>Start a Chat</button>
            </div>
          {:else}
            <div class="empty-state">
              <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" width="40" height="40">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
              </svg>
              <p>No results found</p>
              <button class="clear-search-btn" on:click={clearSearch}>Clear Search</button>
            </div>
          {/if}
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
                        <span class="avatar-placeholder">👤</span>
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
                        <span class="avatar-placeholder">{getChatDisplayName(chat).substring(0, 1).toUpperCase()}</span>
                      {/if}
                      {#if chat.type === 'group'}
                        <span class="group-indicator">
                          <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" width="12" height="12">
                            <path d="M13 6a3 3 0 11-6 0 3 3 0 016 0zM18 8a2 2 0 11-4 0 2 2 0 014 0zM14 15a4 4 0 00-8 0v3h8v-3zM6 8a2 2 0 11-4 0 2 2 0 014 0zM16 18v-3a5.972 5.972 0 00-.75-2.906A3.005 3.005 0 0119 15v3h-3zM4.75 12.094A5.973 5.973 0 004 15v3H1v-3a3 3 0 013.75-2.906z" />
                          </svg>
                        </span>
                      {/if}
                    </div>
                    <div class="chat-info">
                      <div class="chat-header">
                        <span class="chat-name">{getChatDisplayName(chat)}</span>
                        <span class="chat-time">{chat.lastMessage ? formatTimeAgo(chat.lastMessage.timestamp) : ''}</span>
                      </div>
                      <div class="chat-preview">
                        {#if chat.lastMessage && chat.lastMessage.content}
                          {#if chat.lastMessage.senderName}
                            <span class="message-sender">{chat.lastMessage.senderName}:</span>
                          {/if}
                          <span class="message-content">{chat.lastMessage.content.substring(0, 30)}{chat.lastMessage.content.length > 30 ? '...' : ''}</span>
                        {:else}
                          <span class="no-messages">No messages yet</span>
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
            <span class="avatar-placeholder">👥</span>
          {/if}
        </div>
        <div class="chat-title">
          <h2>{getChatDisplayName(selectedChat)}</h2>
          <p class="group-info">
            {selectedChat.type === 'group' ? `${selectedChat.participants.length} members` : ''}
            <span class="ws-status" data-connected="false" title="WebSocket connection status"></span>
          </p>
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
    --background-color: #111827;
    --text-color: #f3f4f6;
    --border-color: #374151;
    --hover-bg: #1f2937;
    --active-bg: rgba(29, 78, 216, 0.15);
    --message-bg: #2d3748;
    --own-message-bg: #3b82f6;
    --input-bg: #1f2937;
    --text-primary: #f3f4f6;
    --text-secondary: #9ca3af;
    --text-tertiary: #6b7280;
    --avatar-fallback-bg: #374151;
    --chat-item-border: #1f2937;
    --chat-list-bg: #111827;
    --active-border: #3b82f6;
  }

  /* Light mode variables */
  .message-container {
    --border-color: #e2e8f0;
    --hover-bg: #f7fafc;
    --active-bg: #e5efff;
    --message-bg: #e2e8f0;
    --own-message-bg: #3b82f6;
    --input-bg: #f7fafc;
    --text-primary: #1f2937;
    --text-secondary: #6b7280;
    --text-tertiary: #9ca3af;
    --avatar-fallback-bg: #e2e8f0;
    --chat-item-border: #f3f4f6;
    --chat-list-bg: white;
    --active-border: #3b82f6;
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
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .section-header h1 {
    font-size: 1.25rem;
    font-weight: bold;
    margin: 0;
    color: var(--text-primary);
  }

  .compose-button {
    display: flex;
    align-items: center;
    background-color: var(--own-message-bg);
    color: white;
    border: none;
    border-radius: 9999px;
    padding: 6px 12px;
    font-size: 0.875rem;
    font-weight: 500;
    cursor: pointer;
    transition: background-color 0.2s;
  }

  .compose-button svg {
    margin-right: 4px;
  }

  .compose-button:hover {
    background-color: #2563eb;
  }

  .search-container {
    padding: 12px 16px;
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
    padding: 10px 16px 10px 40px;
    border-radius: 9999px;
    border: 1px solid var(--border-color);
    background-color: var(--input-bg);
    color: var(--text-color);
    font-size: 0.95rem;
    transition: all 0.2s ease;
  }

  .search-input:focus {
    outline: none;
    border-color: var(--own-message-bg);
    box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.15);
  }

  .search-input-wrapper::before {
    content: "";
    position: absolute;
    left: 14px;
    width: 16px;
    height: 16px;
    background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' fill='none' viewBox='0 0 24 24' stroke='%236B7280'%3E%3Cpath stroke-linecap='round' stroke-linejoin='round' stroke-width='2' d='M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z'%3E%3C/path%3E%3C/svg%3E");
    background-repeat: no-repeat;
    background-position: center;
    background-size: contain;
    pointer-events: none;
  }

  :global(.dark) .search-input-wrapper::before {
    background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' fill='none' viewBox='0 0 24 24' stroke='%239CA3AF'%3E%3Cpath stroke-linecap='round' stroke-linejoin='round' stroke-width='2' d='M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z'%3E%3C/path%3E%3C/svg%3E");
  }

  .clear-search-button {
    position: absolute;
    right: 12px;
    background: none;
    border: none;
    padding: 4px;
    color: var(--text-secondary);
    cursor: pointer;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: all 0.2s ease;
  }

  .clear-search-button:hover {
    color: var(--text-primary);
    background-color: var(--hover-bg);
  }

  .search-dropdown {
    position: absolute;
    top: calc(100% + 4px);
    left: 16px;
    right: 16px;
    max-height: 400px;
    overflow-y: auto;
    background-color: var(--background-color);
    border: 1px solid var(--border-color);
    border-radius: 12px;
    box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
    z-index: 1000;
  }

  .search-dropdown-section {
    padding: 12px;
  }

  .search-dropdown-section + .search-dropdown-section {
    border-top: 1px solid var(--border-color);
  }

  .search-dropdown-title {
    font-size: 0.75rem;
    color: var(--text-secondary);
    text-transform: uppercase;
    letter-spacing: 0.5px;
    margin: 4px 4px 12px;
    font-weight: 600;
  }

  .dropdown-item {
    padding: 10px 12px;
    border-radius: 8px;
    margin-bottom: 4px;
    transition: all 0.15s ease;
    display: flex;
    align-items: center;
    width: 100%;
    text-align: left;
    border: none;
    background: none;
    cursor: pointer;
    color: var(--text-primary);
    position: relative;
  }

  .dropdown-item:hover {
    background-color: var(--hover-bg);
  }
  
  .start-chat-btn {
    display: flex;
    align-items: center;
    padding: 4px 8px;
    border-radius: 4px;
    background-color: var(--own-message-bg);
    color: white;
    font-size: 0.75rem;
    font-weight: 500;
    margin-left: auto;
    border: none;
    cursor: pointer;
    transition: background-color 0.2s;
  }
  
  .start-chat-btn svg {
    margin-right: 4px;
  }
  
  .start-chat-btn:hover {
    background-color: #2563eb;
  }
  
  .no-results-message {
    color: var(--text-secondary);
    font-style: italic;
    text-align: center;
    padding: 20px 0;
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
    display: flex;
    align-items: center;
  }
  
  .ws-status {
    display: inline-block;
    width: 8px;
    height: 8px;
    border-radius: 50%;
    margin-left: 6px;
    background-color: #ef4444;
  }
  
  .ws-status[data-connected="true"] {
    background-color: #10b981;
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

  .message-sender {
    font-weight: 600;
    margin-right: 0.25rem;
    color: var(--color-primary);
    display: inline-block;
  }
  
  .message-content {
    opacity: 0.9;
    display: inline;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  
  .chat-preview {
    display: flex;
    align-items: center;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .group-indicator {
    display: flex;
    align-items: center;
    margin-left: 4px;
    color: var(--text-secondary, #6c757d);
  }

  .no-messages {
    color: var(--text-tertiary, #9ca3af);
    font-style: italic;
  }

  /* Chat List */
  .chat-list {
    overflow-y: auto;
    flex: 1;
    background-color: var(--chat-list-bg);
  }

  .chat-items {
    list-style: none;
    margin: 0;
    padding: 0;
  }

  .chat-item {
    display: flex;
    align-items: center;
    padding: 12px 16px;
    border-bottom: 1px solid var(--chat-item-border);
    background: none;
    border-left: 3px solid transparent;
    border-right: none;
    border-top: none;
    width: 100%;
    text-align: left;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .chat-item:hover {
    background-color: var(--hover-bg);
  }

  .chat-item.active {
    background-color: var(--active-bg);
    border-left: 3px solid var(--active-border);
  }

  .avatar-container {
    width: 48px;
    height: 48px;
    border-radius: 50%;
    overflow: hidden;
    background-color: var(--avatar-fallback-bg);
    display: flex;
    align-items: center;
    justify-content: center;
    margin-right: 12px;
    flex-shrink: 0;
    position: relative;
    box-shadow: 0 2px 5px rgba(0, 0, 0, 0.1);
  }

  .avatar-image {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  .avatar-placeholder {
    color: var(--text-primary);
    font-size: 1.2rem;
    font-weight: 600;
  }

  .group-indicator {
    position: absolute;
    bottom: 0;
    right: 0;
    background-color: var(--own-message-bg);
    color: white;
    width: 18px;
    height: 18px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 10px;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.2);
    border: 2px solid var(--chat-list-bg);
  }

  .chat-info {
    flex: 1;
    min-width: 0;
    margin-right: 8px;
  }

  .chat-header {
    display: flex;
    justify-content: space-between;
    align-items: baseline;
    margin-bottom: 4px;
  }

  .chat-name {
    font-weight: 600;
    font-size: 0.95rem;
    color: var(--text-primary);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .chat-time {
    font-size: 0.7rem;
    color: var(--text-secondary);
    flex-shrink: 0;
  }

  .chat-preview {
    display: flex;
    color: var(--text-secondary);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    font-size: 0.85rem;
    line-height: 1.3;
  }

  .message-sender {
    font-weight: 500;
    margin-right: 0.25rem;
    color: var(--text-primary);
    display: inline-block;
  }
  
  .message-content {
    opacity: 0.95;
    display: inline;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    color: var(--text-secondary);
  }

  .no-messages {
    color: var(--text-tertiary);
    font-style: italic;
    font-size: 0.85rem;
  }

  .unread-badge {
    background-color: var(--own-message-bg);
    color: white;
    font-size: 0.75rem;
    min-width: 20px;
    height: 20px;
    padding: 0 6px;
    border-radius: 10px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-weight: 600;
    flex-shrink: 0;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.2);
  }

  .search-section-title {
    font-size: 0.875rem;
    font-weight: 600;
    color: var(--text-primary);
    margin: 16px 16px 8px;
    padding-bottom: 8px;
    border-bottom: 1px solid var(--border-color);
  }

  /* Add these new styles for the button group and modal */
  .button-group {
    display: flex;
    gap: 8px;
  }

  .modal-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background-color: rgba(0, 0, 0, 0.5);
    display: flex;
    justify-content: center;
    align-items: center;
    z-index: 1000;
  }

  .modal-container {
    width: 90%;
    max-width: 600px;
    max-height: 90vh;
    border-radius: 12px;
    overflow: hidden;
    box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
  }
</style>

<!-- Add the Debug Panel component -->
<DebugPanel />
