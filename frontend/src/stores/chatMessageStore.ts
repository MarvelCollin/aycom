import { derived, writable } from 'svelte/store';
import { websocketStore, type ChatMessage, type MessageType } from './websocketStore';
import { getAuthToken } from '../utils/auth';
import { createLoggerWithPrefix } from '../utils/logger';

const logger = createLoggerWithPrefix('ChatMessageStore');

export interface User {
  id: string;
  username: string;
  display_name: string;
  avatar_url?: string;
}

export interface MessageWithUser extends Omit<ChatMessage, 'user_id'> {
  user: User;
  user_id: string;
}

export interface ChatState {
  messages: Record<string, Record<string, MessageWithUser>>;
  typingUsers: Record<string, Record<string, Date>>;
  unreadCount: Record<string, number>;
  lastError: string | null;
}

const initialState: ChatState = {
  messages: {},
  typingUsers: {},
  unreadCount: {},
  lastError: null
};

// Standalone function to handle incoming WebSocket messages
// This is exported to be used by other modules
export function handleIncomingMessage(message: ChatMessage & { user?: User }) {
  chatMessageStore.addIncomingMessage(message);
}

function createChatMessageStore() {
  const { subscribe, update, set } = writable<ChatState>(initialState);

  // Subscribe to WebSocket store to receive messages
  const unsubscribeWsState = websocketStore.subscribe(wsState => {
    if (wsState.lastError) {
      update(state => ({ ...state, lastError: wsState.lastError }));
    }
  });
  
  // Register a message handler with the WebSocket store
  const unsubscribeMessageHandler = websocketStore.registerMessageHandler(handleIncomingMessage);

  // Initialize chat state for a given chat ID
  const initChat = (chatId: string) => {
    update(state => {
      if (!state.messages[chatId]) {
        return {
          ...state,
          messages: { ...state.messages, [chatId]: {} },
          typingUsers: { ...state.typingUsers, [chatId]: {} },
          unreadCount: { ...state.unreadCount, [chatId]: 0 }
        };
      }
      return state;
    });
  };

  // Connect to chat WebSocket and initialize state
  const connectToChat = (chatId: string) => {
    initChat(chatId);
    websocketStore.connect(chatId);
  };

  // Disconnect from chat WebSocket
  const disconnectFromChat = (chatId: string) => {
    websocketStore.disconnect(chatId);
  };

  // Send a message through WebSocket
  const sendMessage = (chatId: string, content: string, userId: string) => {
    const message: ChatMessage = {
      type: 'text',
      content,
      user_id: userId,
      chat_id: chatId,
      timestamp: new Date()
    };

    websocketStore.sendMessage(chatId, message);
    
    // Optimistically add the message to the local store
    // It will be updated with the correct ID when the server confirms receipt
    const tempId = `temp-${Date.now()}`;
    addMessage({
      ...message,
      message_id: tempId,
      user: { id: userId, username: '', display_name: '' } // Placeholder, will be updated
    });
  };

  // Send a typing indicator
  const sendTypingIndicator = (chatId: string, userId: string) => {
    const message: ChatMessage = {
      type: 'typing',
      user_id: userId,
      chat_id: chatId
    };

    websocketStore.sendMessage(chatId, message);
  };

  // Send a read receipt
  const sendReadReceipt = (chatId: string, messageId: string, userId: string) => {
    const message: ChatMessage = {
      type: 'read',
      user_id: userId,
      chat_id: chatId,
      message_id: messageId
    };

    websocketStore.sendMessage(chatId, message);
    
    // Update local state to mark this message as read
    update(state => {
      if (state.messages[chatId] && state.messages[chatId][messageId]) {
        const updatedMessages = { ...state.messages };
        updatedMessages[chatId] = { ...updatedMessages[chatId] };
        updatedMessages[chatId][messageId] = {
          ...updatedMessages[chatId][messageId],
          is_read: true
        };
        
        return {
          ...state,
          messages: updatedMessages,
          unreadCount: {
            ...state.unreadCount,
            [chatId]: Math.max(0, (state.unreadCount[chatId] || 0) - 1)
          }
        };
      }
      return state;
    });
  };

  // Add a message to the store
  const addMessage = (message: MessageWithUser) => {
    const { chat_id: chatId, message_id: messageId } = message;
    
    if (!chatId || !messageId) {
      logger.error('Cannot add message without chat_id and message_id', message);
      return;
    }
    
    update(state => {
      // Initialize chat data if not exists
      if (!state.messages[chatId]) {
        state = {
          ...state,
          messages: { ...state.messages, [chatId]: {} },
          typingUsers: { ...state.typingUsers, [chatId]: {} },
          unreadCount: { ...state.unreadCount, [chatId]: 0 }
        };
      }
      
      // Add the message
      const updatedMessages = { ...state.messages };
      updatedMessages[chatId] = { 
        ...updatedMessages[chatId], 
        [messageId]: message 
      };
      
      // Increment unread count if the message is from someone else
      const currentUserId = getCurrentUserId();
      const isFromOther = message.user_id !== currentUserId;
      
      return {
        ...state,
        messages: updatedMessages,
        unreadCount: {
          ...state.unreadCount,
          [chatId]: isFromOther 
            ? (state.unreadCount[chatId] || 0) + 1 
            : (state.unreadCount[chatId] || 0)
        }
      };
    });
  };

  // Process an incoming message from WebSocket
  const addIncomingMessage = (message: ChatMessage & { user?: User }) => {
    const { type, chat_id: chatId } = message;
    
    if (!chatId) {
      logger.error('Received message without chat_id', message);
      return;
    }
    
    // Initialize chat if needed
    initChat(chatId);
    
    switch (type) {
      case 'text':
        if (message.user) {
          addMessage(message as MessageWithUser);
        } else {
          logger.error('Received text message without user data', message);
        }
        break;
        
      case 'typing':
        updateTypingStatus(chatId, message.user_id);
        break;
        
      case 'read':
        if (message.message_id) {
          markMessageAsRead(chatId, message.message_id, message.user_id);
        }
        break;
        
      case 'edit':
        if (message.message_id) {
          updateMessage(message as MessageWithUser);
        }
        break;
        
      case 'delete':
        if (message.message_id) {
          deleteMessage(chatId, message.message_id);
        }
        break;
        
      default:
        logger.warn(`Unknown message type: ${type}`, message);
    }
  };

  // Update typing status for a user
  const updateTypingStatus = (chatId: string, userId: string) => {
    update(state => {
      if (!state.typingUsers[chatId]) {
        state = {
          ...state,
          typingUsers: { ...state.typingUsers, [chatId]: {} }
        };
      }
      
      const updatedTyping = { ...state.typingUsers };
      updatedTyping[chatId] = { 
        ...updatedTyping[chatId], 
        [userId]: new Date() 
      };
      
      return {
        ...state,
        typingUsers: updatedTyping
      };
    });
    
    // Automatically clear typing indicator after 3 seconds
    setTimeout(() => {
      clearTypingStatus(chatId, userId);
    }, 3000);
  };

  // Clear typing status for a user
  const clearTypingStatus = (chatId: string, userId: string) => {
    update(state => {
      if (!state.typingUsers[chatId]) return state;
      
      const updatedTyping = { ...state.typingUsers };
      updatedTyping[chatId] = { ...updatedTyping[chatId] };
      
      const userTypingTime = updatedTyping[chatId][userId];
      
      // Only clear if the typing indicator is older than 3 seconds
      if (userTypingTime && (new Date().getTime() - userTypingTime.getTime() >= 3000)) {
        delete updatedTyping[chatId][userId];
      }
      
      return {
        ...state,
        typingUsers: updatedTyping
      };
    });
  };

  // Mark a message as read
  const markMessageAsRead = (chatId: string, messageId: string, userId: string) => {
    update(state => {
      if (!state.messages[chatId] || !state.messages[chatId][messageId]) return state;
      
      const updatedMessages = { ...state.messages };
      updatedMessages[chatId] = { ...updatedMessages[chatId] };
      updatedMessages[chatId][messageId] = {
        ...updatedMessages[chatId][messageId],
        is_read: true
      };
      
      return {
        ...state,
        messages: updatedMessages
      };
    });
  };

  // Update a message (for edits)
  const updateMessage = (message: MessageWithUser) => {
    const { chat_id: chatId, message_id: messageId } = message;
    
    if (!chatId || !messageId) {
      logger.error('Cannot update message without chat_id and message_id', message);
      return;
    }
    
    update(state => {
      if (!state.messages[chatId] || !state.messages[chatId][messageId]) return state;
      
      const updatedMessages = { ...state.messages };
      updatedMessages[chatId] = { ...updatedMessages[chatId] };
      updatedMessages[chatId][messageId] = {
        ...updatedMessages[chatId][messageId],
        ...message,
        is_edited: true
      };
      
      return {
        ...state,
        messages: updatedMessages
      };
    });
  };

  // Delete a message
  const deleteMessage = (chatId: string, messageId: string) => {
    update(state => {
      if (!state.messages[chatId] || !state.messages[chatId][messageId]) return state;
      
      const updatedMessages = { ...state.messages };
      updatedMessages[chatId] = { ...updatedMessages[chatId] };
      updatedMessages[chatId][messageId] = {
        ...updatedMessages[chatId][messageId],
        content: '',
        is_deleted: true
      };
      
      return {
        ...state,
        messages: updatedMessages
      };
    });
  };

  // Reset error state
  const resetError = () => {
    update(state => ({ ...state, lastError: null }));
    websocketStore.resetError();
  };

  // Clean up on unmount
  const cleanup = () => {
    unsubscribeWsState();
    unsubscribeMessageHandler();
    websocketStore.disconnectAll();
  };

  // Helper to get current user ID from auth token
  const getCurrentUserId = (): string => {
    const token = getAuthToken();
    if (!token) return '';
    
    try {
      // JWT tokens are base64 encoded with 3 parts separated by dots
      const payload = token.split('.')[1];
      // Decode the base64 payload
      const decoded = JSON.parse(atob(payload));
      return decoded.sub || '';
    } catch (e) {
      logger.error('Failed to decode JWT token', e);
      return '';
    }
  };

  return {
    subscribe,
    connectToChat,
    disconnectFromChat,
    sendMessage,
    sendTypingIndicator,
    sendReadReceipt,
    addMessage,
    addIncomingMessage,
    updateTypingStatus,
    markMessageAsRead,
    updateMessage,
    deleteMessage,
    resetError,
    cleanup
  };
}

// Export the singleton store instance
export const chatMessageStore = createChatMessageStore();

// Derived store for getting messages for a specific chat
export function getMessagesForChat(chatId: string) {
  return derived(chatMessageStore, $chatStore => {
    const messages = $chatStore.messages[chatId] || {};
    return Object.values(messages).sort((a, b) => {
      const timeA = a.timestamp ? new Date(a.timestamp).getTime() : 0;
      const timeB = b.timestamp ? new Date(b.timestamp).getTime() : 0;
      return timeA - timeB;
    });
  });
}

// Derived store for getting typing users for a specific chat
export function getTypingUsersForChat(chatId: string) {
  return derived(chatMessageStore, $chatStore => {
    const typing = $chatStore.typingUsers[chatId] || {};
    return Object.keys(typing).filter(userId => {
      const typingTime = typing[userId];
      return typingTime && (new Date().getTime() - typingTime.getTime() < 3000);
    });
  });
}

// Derived store for getting unread count for a specific chat
export function getUnreadCountForChat(chatId: string) {
  return derived(chatMessageStore, $chatStore => {
    return $chatStore.unreadCount[chatId] || 0;
  });
} 