import { writable } from 'svelte/store';
import { getAuthToken } from '../utils/auth';
import appConfig from '../config/appConfig';
import { createLoggerWithPrefix } from '../utils/logger';

const logger = createLoggerWithPrefix('WebSocketStore');

export type MessageType = 'text' | 'typing' | 'read' | 'edit' | 'delete';

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

export interface WebSocketState {
  connected: boolean;
  reconnecting: boolean;
  lastError: string | null;
  chatConnections: Record<string, WebSocket>;
  connectionStatus: Record<string, 'connecting' | 'connected' | 'disconnected' | 'error'>;
}

const initialState: WebSocketState = {
  connected: false,
  reconnecting: false,
  lastError: null,
  chatConnections: {},
  connectionStatus: {}
};

type MessageHandler = (message: any) => void;
const messageHandlers: MessageHandler[] = [];

function createWebSocketStore() {
  const { subscribe, update, set } = writable<WebSocketState>(initialState);

  let reconnectAttempts = 0;
  let reconnectTimeouts: Record<string, number> = {};

  // Simplified URL building
  const buildWebSocketUrl = (chatId: string): string => {
    const token = getAuthToken();
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const host = window.location.hostname;
    const port = '8083'; // Fixed port for API Gateway
    
    let wsUrl = `${protocol}//${host}:${port}/api/v1/chats/${chatId}/ws`;
    
    // Extract user ID from token if available
    let userId = '';
    if (token) {
      try {
        const payload = JSON.parse(atob(token.split('.')[1]));
        userId = payload.user_id || payload.sub || '';
      } catch (e) {
        logger.warn('Could not extract user ID from token');
      }
    }
    
    // Add query parameters
    const params = [];
    if (token) params.push(`token=${token}`);
    if (userId) params.push(`user_id=${userId}`);
    
    if (params.length > 0) {
      wsUrl += `?${params.join('&')}`;
    }
    
    return wsUrl;
  };

  const connect = (chatId: string) => {
    logger.info(`Connecting to WebSocket for chat: ${chatId}`);
    
    update(state => ({
      ...state,
      connectionStatus: { ...state.connectionStatus, [chatId]: 'connecting' }
    }));

    // Clear any existing timeout
    if (reconnectTimeouts[chatId]) {
      clearTimeout(reconnectTimeouts[chatId]);
      delete reconnectTimeouts[chatId];
    }

    // Close existing connection if any
    update(state => {
      if (state.chatConnections[chatId]) {
        try {
          state.chatConnections[chatId].close();
        } catch (e) {
          logger.warn(`Error closing existing connection: ${e}`);
        }
      }
      return state;
    });

    try {
      const wsUrl = buildWebSocketUrl(chatId);
      logger.info(`Attempting to connect to WebSocket: ${wsUrl}`);
      
      const ws = new WebSocket(wsUrl);
      ws.onopen = () => {
        logger.info(`WebSocket connection established for chat ${chatId}`);
        update(s => ({ 
          ...s, 
          connected: true, 
          reconnecting: false,
          lastError: null,
          chatConnections: {
            ...s.chatConnections,
            [chatId]: ws
          },
          connectionStatus: {
            ...s.connectionStatus,
            [chatId]: 'connected'
          }
        }));
        
        // Send an initial message to confirm connection
        try {
          const token = getAuthToken();
          let userId = '';
          if (token) {
            try {
              const payload = JSON.parse(atob(token.split('.')[1]));
              userId = payload.user_id || payload.sub || '';
            } catch (e) {
              logger.warn('Could not extract user ID from token');
            }
          }
          
          const initialMessage = {
            type: 'connection_check',
            user_id: userId,
            chat_id: chatId,
            timestamp: new Date()
          };
          ws.send(JSON.stringify(initialMessage));
          logger.debug(`Sent initial connection check for chat ${chatId}`);
        } catch (e) {
          logger.error(`Error sending initial message for chat ${chatId}:`, e);
        }
        
        reconnectAttempts = 0;
      };
      
      ws.onmessage = (event) => {
        try {
          logger.debug(`WebSocket message received for chat ${chatId}:`, event.data);
          const message = JSON.parse(event.data);
          
          messageHandlers.forEach(handler => handler(message));
        } catch (e) {
          logger.error(`Error parsing WebSocket message for chat ${chatId}:`, e);
        }
      };
      
      ws.onerror = (error) => {
        logger.error(`WebSocket error for chat ${chatId}:`, error);
        update(s => ({ 
          ...s, 
          lastError: 'Connection error',
          connectionStatus: {
            ...s.connectionStatus,
            [chatId]: 'error'
          }
        }));
      };
      
      ws.onclose = (event) => {
        logger.info(`WebSocket closed for chat ${chatId}: code=${event.code}, reason="${event.reason}", wasClean=${event.wasClean}`);
        
        update(s => {
          const connections = { ...s.chatConnections };
          delete connections[chatId];
          
          const status = { ...s.connectionStatus };
          status[chatId] = 'disconnected';
          
          return { 
            ...s, 
            connected: Object.keys(connections).length > 0,
            chatConnections: connections,
            connectionStatus: status
          };
        });
        
        if (event.code !== 1000) {
          logger.info(`Will attempt to reconnect to chat ${chatId} due to non-clean close`);
          attemptReconnect(chatId);
        }
      };
      
      update(state => ({ 
        ...state, 
        chatConnections: { 
          ...state.chatConnections, 
          [chatId]: ws 
        } 
      }));
      
    } catch (error) {
      logger.error('Failed to establish WebSocket connection:', error);
      
      update(state => ({
        ...state,
        lastError: 'Failed to connect',
        connectionStatus: {
          ...state.connectionStatus,
          [chatId]: 'error'
        }
      }));
      
      attemptReconnect(chatId);
    }
  };

  const disconnect = (chatId: string) => {
    update(state => {
      if (state.chatConnections[chatId]) {
        state.chatConnections[chatId].close(1000, 'Disconnect requested');
        const connections = { ...state.chatConnections };
        delete connections[chatId];
        
        const status = { ...state.connectionStatus };
        status[chatId] = 'disconnected';
        
        return {
          ...state,
          chatConnections: connections,
          connected: Object.keys(connections).length > 0,
          connectionStatus: status
        };
      }
      return state;
    });
  };

  const disconnectAll = () => {
    update(state => {
      Object.keys(state.chatConnections).forEach(chatId => {
        try {
          state.chatConnections[chatId].close(1000, 'Disconnect all requested');
        } catch (e) {
          logger.error(`Error closing WebSocket for chat ${chatId}:`, e);
        }
      });
      
      // Clear all timeouts
      Object.values(reconnectTimeouts).forEach(timeout => {
        clearTimeout(timeout);
      });
      reconnectTimeouts = {};
      
      return {
        ...state,
        chatConnections: {},
        connected: false,
        reconnecting: false,
        connectionStatus: {}
      };
    });
  };

  const sendMessage = (chatId: string, message: ChatMessage) => {
    update(state => {
      if (!state.chatConnections[chatId]) {
        logger.warn(`Cannot send message to chat ${chatId}: not connected`);
        connect(chatId);
        return state;
      }
      
      try {
        const ws = state.chatConnections[chatId];
        ws.send(JSON.stringify(message));
        logger.debug(`Message sent to chat ${chatId}:`, message);
      } catch (e) {
        logger.error(`Error sending message to chat ${chatId}:`, e);
        return {
          ...state,
          lastError: 'Failed to send message'
        };
      }
      
      return state;
    });
  };
  
  const resetError = () => {
    update(state => ({
      ...state,
      lastError: null
    }));
  };

  const attemptReconnect = (chatId: string) => {
    update(state => ({
      ...state,
      reconnecting: true,
      connectionStatus: {
        ...state.connectionStatus,
        [chatId]: 'connecting'
      }
    }));
    
    const maxReconnectAttempts = 5;
    
    if (reconnectAttempts >= maxReconnectAttempts) {
      logger.warn(`Maximum reconnect attempts (${maxReconnectAttempts}) reached for chat ${chatId}`);
      update(state => ({
        ...state,
        reconnecting: false,
        lastError: 'Maximum reconnect attempts reached',
        connectionStatus: {
          ...state.connectionStatus,
          [chatId]: 'error'
        }
      }));
      return;
    }
    
    const baseDelay = 1000;
    const delay = baseDelay * Math.pow(1.5, reconnectAttempts);
    reconnectAttempts++;
    
    if (reconnectTimeouts[chatId]) {
      clearTimeout(reconnectTimeouts[chatId]);
    }
    
    reconnectTimeouts[chatId] = window.setTimeout(() => {
      logger.info(`Attempting to reconnect to chat ${chatId} (attempt ${reconnectAttempts})`);
      delete reconnectTimeouts[chatId];
      connect(chatId);
    }, delay);
  };
  
  const registerMessageHandler = (handler: MessageHandler) => {
    messageHandlers.push(handler);
    
    return () => {
      const index = messageHandlers.indexOf(handler);
      if (index !== -1) {
        messageHandlers.splice(index, 1);
      }
    };
  };
  
  const isConnected = (chatId: string) => {
    let connected = false;
    
    update(state => {
      connected = !!state.chatConnections[chatId] && 
                  state.chatConnections[chatId].readyState === WebSocket.OPEN;
      return state;
    });
    
    return connected;
  };

  return {
    subscribe,
    connect,
    disconnect,
    disconnectAll,
    sendMessage,
    resetError,
    registerMessageHandler,
    isConnected
  };
}

export const websocketStore = createWebSocketStore();

let setupChatMessageStore: ((ws: any) => void) | null = null;

export function setWebSocketHandlers(setup: (ws: any) => void) {
  setupChatMessageStore = setup;
  
  if (setupChatMessageStore) {
    setupChatMessageStore(websocketStore);
  }
} 