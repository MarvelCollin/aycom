import { derived, writable } from 'svelte/store';
import type { MessageType } from './websocketStore';
import { getAuthToken } from '../utils/auth';
import { createLoggerWithPrefix } from '../utils/logger';
import { registerChatMessageHandler } from '../api/chat';

const logger = createLoggerWithPrefix('ChatMessageStore');

export interface ChatMessage {
  type: MessageType;
  content?: string;
  user_id: string;
  chat_id: string;
  timestamp?: Date;
  message_id?: string;
  is_edited?: boolean;
  is_deleted?: boolean;
  is_read?: boolean;
}

let websocketConnect: (chatId: string) => void;
let websocketDisconnect: (chatId: string) => void;
let websocketSendMessage: (chatId: string, message: ChatMessage) => void;
let websocketSubscribe: (callback: (state: any) => void) => () => void;
let websocketRegisterMessageHandler: (handler: (message: any) => void) => () => void;
let websocketResetError: () => void;
let websocketDisconnectAll: () => void;

let unsubscribeWsState: (() => void) | null = null;
let unsubscribeMessageHandler: (() => void) | null = null;

export function setupWebsocketMethods(methods: {
  connect: (chatId: string) => void;
  disconnect: (chatId: string) => void;
  sendMessage: (chatId: string, message: ChatMessage) => void;
  subscribe: (callback: (state: any) => void) => () => void;
  registerMessageHandler: (handler: (message: any) => void) => () => void;
  resetError: () => void;
  disconnectAll: () => void;
}) {
  websocketConnect = methods.connect;
  websocketDisconnect = methods.disconnect;
  websocketSendMessage = methods.sendMessage;
  websocketSubscribe = methods.subscribe;
  websocketRegisterMessageHandler = methods.registerMessageHandler;
  websocketResetError = methods.resetError;
  websocketDisconnectAll = methods.disconnectAll;

  if (chatMessageStore) {
    unsubscribeWsState = websocketSubscribe(wsState => {
      if (wsState.lastError) {
        chatMessageStore.update(state => ({ ...state, lastError: wsState.lastError }));
      }
    });

    unsubscribeMessageHandler = websocketRegisterMessageHandler(handleIncomingMessage);
  }
}

export interface User {
  id: string;
  username: string;
  name: string;
  profile_picture_url?: string;
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

export function handleIncomingMessage(message: ChatMessage & { user?: User }) {
  chatMessageStore.addIncomingMessage(message);
}

function createChatMessageStore() {
  const { subscribe, update, set } = writable<ChatState>(initialState);

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

  const connectToChat = (chatId: string) => {
    initChat(chatId);
    websocketConnect(chatId);
  };

  const disconnectFromChat = (chatId: string) => {
    websocketDisconnect(chatId);
  };

  const sendMessage = (chatId: string, content: string, userId: string) => {

    const tempId = `temp-${Date.now()}`;
    const timestamp = new Date();

    const tempMessage: MessageWithUser = {
      type: 'text',
      content,
      user_id: userId,
      chat_id: chatId,
      message_id: tempId,
      timestamp,
      is_read: false,
      is_edited: false,
      is_deleted: false,
      user: { 
        id: userId, 
        username: '', 
        name: '' 
      }
    };

    addMessage(tempMessage);

    const message: ChatMessage = {
      type: 'text',
      content,
      user_id: userId,
      chat_id: chatId,
      timestamp: new Date(),
      message_id: tempId 
    };

    logger.debug('Sending message to WebSocket', { chatId, tempId });
    websocketSendMessage(chatId, message);
  };

  const sendTypingIndicator = (chatId: string, userId: string) => {
    const message: ChatMessage = {
      type: 'typing',
      user_id: userId,
      chat_id: chatId
    };

    websocketSendMessage(chatId, message);
  };

  const sendReadReceipt = (chatId: string, messageId: string, userId: string) => {
    const message: ChatMessage = {
      type: 'read',
      user_id: userId,
      chat_id: chatId,
      message_id: messageId
    };

    websocketSendMessage(chatId, message);

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

  const addMessage = (message: MessageWithUser) => {
    const { chat_id: chatId, message_id: messageId } = message;

    if (!chatId || !messageId) {
      logger.error('Cannot add message without chat_id and message_id', message);
      return;
    }

    update(state => {

      if (!state.messages[chatId]) {
        state = {
          ...state,
          messages: { ...state.messages, [chatId]: {} },
          typingUsers: { ...state.typingUsers, [chatId]: {} },
          unreadCount: { ...state.unreadCount, [chatId]: 0 }
        };
      }

      const updatedMessages = { ...state.messages };
      updatedMessages[chatId] = { 
        ...updatedMessages[chatId], 
        [messageId]: message 
      };

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

  const addIncomingMessage = (message: any) => {
    const { type, chat_id: chatId } = message;

    if (!chatId) {
      logger.error('Received message without chat_id', message);
      return;
    }

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

      case 'update':

        if (message.originalTempId && message.message_id) {
          updateMessageWithServerData(
            message.originalTempId,
            message.chat_id,
            message.message_id,
            message.timestamp
          );
        }
        break;
    }
  };

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

    setTimeout(() => {
      clearTypingStatus(chatId, userId);
    }, 3000);
  };

  const clearTypingStatus = (chatId: string, userId: string) => {
    update(state => {
      if (!state.typingUsers[chatId]) return state;

      const updatedTyping = { ...state.typingUsers };
      updatedTyping[chatId] = { ...updatedTyping[chatId] };

      const userTypingTime = updatedTyping[chatId][userId];

      if (userTypingTime && (new Date().getTime() - userTypingTime.getTime() >= 3000)) {
        delete updatedTyping[chatId][userId];
      }

      return {
        ...state,
        typingUsers: updatedTyping
      };
    });
  };

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

  const resetError = () => {
    update(state => ({ ...state, lastError: null }));
    websocketResetError();
  };

  const cleanup = () => {
    if (unsubscribeWsState) {
      unsubscribeWsState();
      unsubscribeWsState = null;
    }
    if (unsubscribeMessageHandler) {
      unsubscribeMessageHandler();
      unsubscribeMessageHandler = null;
    }
    if (websocketDisconnectAll) {
      websocketDisconnectAll();
    }
  };

  const getCurrentUserId = (): string => {
    const token = getAuthToken();
    if (!token) return '';

    try {

      const payload = token.split('.')[1];

      const decoded = JSON.parse(atob(payload));
      return decoded.sub || '';
    } catch (e) {
      logger.error('Failed to decode JWT token', e);
      return '';
    }
  };

  const updateMessageWithServerData = (tempMessageId: string, chatId: string, serverMessageId: string, serverTimestamp: Date) => {
    update(state => {

      if (!state.messages[chatId]) return state;

      const tempMessage = state.messages[chatId][tempMessageId];
      if (!tempMessage) {
        logger.warn('Could not find temporary message to update', { tempMessageId, chatId });
        return state;
      }

      const updatedMessages = { ...state.messages };
      updatedMessages[chatId] = { ...updatedMessages[chatId] };

      const updatedMessage = {
        ...tempMessage,
        message_id: serverMessageId,
        timestamp: serverTimestamp
      };

      updatedMessages[chatId][serverMessageId] = updatedMessage;
      delete updatedMessages[chatId][tempMessageId];

      logger.debug('Updated temporary message with server data', { 
        tempId: tempMessageId, 
        serverId: serverMessageId 
      });

      return {
        ...state,
        messages: updatedMessages
      };
    });
  };

  const clearChat = (chatId: string) => {
    update(state => {

      const updatedState = {
        ...state,
        messages: { ...state.messages }
      };

      delete updatedState.messages[chatId];

      updatedState.messages[chatId] = {};

      logger.debug(`Cleared messages for chat ${chatId}`);
      return updatedState;
    });
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
    clearTypingStatus,
    markMessageAsRead,
    update,
    updateMessage,
    deleteMessage,
    resetError,
    cleanup,
    updateMessageWithServerData,
    clearChat
  };
}

export const chatMessageStore = createChatMessageStore();

registerChatMessageHandler(handleIncomingMessage);

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

export function getTypingUsersForChat(chatId: string) {
  return derived(chatMessageStore, $chatStore => {
    const typing = $chatStore.typingUsers[chatId] || {};
    return Object.keys(typing).filter(userId => {
      const typingTime = typing[userId];
      return typingTime && (new Date().getTime() - typingTime.getTime() < 3000);
    });
  });
}

export function getUnreadCountForChat(chatId: string) {
  return derived(chatMessageStore, $chatStore => {
    return $chatStore.unreadCount[chatId] || 0;
  });
}