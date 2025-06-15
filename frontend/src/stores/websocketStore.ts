import { writable } from 'svelte/store';
import { getAuthToken } from '../utils/auth';
import appConfig from '../config/appConfig';
import { createLoggerWithPrefix } from '../utils/logger';

const logger = createLoggerWithPrefix('WebSocketStore');

export type MessageType = 'text' | 'typing' | 'read' | 'edit' | 'delete' | 'system';

export interface ChatMessage {
  type: MessageType;
  content?: string;
  user_id?: string;
  sender_id?: string;
  chat_id: string;
  timestamp?: Date | string;
  message_id?: string;
  is_edited?: boolean;
  is_deleted?: boolean;
  is_read?: boolean;
  is_system?: boolean;
  is_fallback?: boolean;
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

type MessageHandler = (message: ChatMessage) => void;
const messageHandlers: MessageHandler[] = [];

function createWebSocketStore() {
  const { subscribe, update, set } = writable<WebSocketState>(initialState);

  let reconnectAttempts = 0;
  let reconnectTimeouts: Record<string, number> = {};

  // Simplified URL building
  const buildWebSocketUrl = (chatId: string): string => {
    const token = getAuthToken();
    
    // Use consistent URL construction method
    // Get base URL from config, but fallback to current location
    let baseUrl = appConfig.api.wsUrl;
    if (!baseUrl) {
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
      const host = window.location.hostname;
      // Use location.port if available, otherwise default to 8083
      const port = window.location.port || '8083';
      baseUrl = `${protocol}//${host}:${port}/api/v1`;
    }
    
    // Ensure baseUrl doesn't end with a slash
    baseUrl = baseUrl.endsWith('/') ? baseUrl.slice(0, -1) : baseUrl;
    
    // Use hardcoded URL for development testing to ensure connection works
    let wsUrl = `ws://localhost:8083/api/v1/chats/${chatId}/ws`;
    
    // Extract user ID from token if available
    let userId = '';
    if (token) {
      try {
        const payload = JSON.parse(atob(token.split('.')[1]));
        userId = payload.user_id || payload.sub || '';
      } catch (e) {
        logger.warn('Could not extract user ID from token');
        console.error('[WebSocket] Failed to extract user ID from token:', e);
      }
    }
    
    // Add query parameters
    const params: string[] = [];
    if (token) params.push(`token=${encodeURIComponent(token)}`);
    if (userId) params.push(`user_id=${encodeURIComponent(userId)}`);
    
    if (params.length > 0) {
      wsUrl += `?${params.join('&')}`;
    }
    
    logger.info(`[WebSocket] Built connection URL: ${wsUrl}`);
    return wsUrl;
  };

  const connect = (chatId: string) => {
    logger.info(`Connecting to WebSocket for chat: ${chatId}`);
    console.log(`[WebSocket] Attempting to connect for chat ID: ${chatId}`);
    
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
          console.error(`[WebSocket] Error closing existing connection for chat ${chatId}:`, e);
        }
      }
      return state;
    });

    try {
      const wsUrl = buildWebSocketUrl(chatId);
      logger.info(`Attempting to connect to WebSocket: ${wsUrl}`);
      console.log(`[WebSocket] Attempting connection with URL: ${wsUrl}`);
      
      // Set a connection timeout
      const connectionTimeout = setTimeout(() => {
        logger.warn(`WebSocket connection timeout for chat ${chatId}`);
        console.warn(`[WebSocket] Connection timeout for chat ${chatId}`);
        
        update(s => ({ 
          ...s, 
          lastError: 'Connection timeout',
          connectionStatus: {
            ...s.connectionStatus,
            [chatId]: 'error'
          }
        }));
        
        // Use fallback mode
        useLocalFallback(chatId);
      }, 5000); // 5 second timeout
      
      const ws = new WebSocket(wsUrl);
      ws.onopen = () => {
        // Clear the timeout since we connected successfully
        clearTimeout(connectionTimeout);
        
        logger.info(`WebSocket connection established for chat ${chatId}`);
        console.log(`[WebSocket] Connection established successfully for chat ${chatId}`);
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
              console.error('[WebSocket] Failed to extract user ID for initial message:', e);
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
          console.error(`[WebSocket] Failed to send initial message for chat ${chatId}:`, e);
        }
        
        reconnectAttempts = 0;
      };
      
      ws.onmessage = (event) => {
        try {
          logger.debug(`WebSocket message received for chat ${chatId}:`, event.data);
          const message = JSON.parse(event.data);
          
          messageHandlers.forEach(handler => {
            // Ensure message has the required properties for ChatMessage
            const chatMessage: ChatMessage = {
              ...message,
              type: message.type || 'text',
              chat_id: message.chat_id || chatId
            };
            handler(chatMessage);
          });
        } catch (e) {
          logger.error(`Error parsing WebSocket message for chat ${chatId}:`, e);
          console.error(`[WebSocket] Failed to parse message for chat ${chatId}:`, e, 'Raw data:', event.data);
        }
      };
      
      ws.onerror = (error) => {
        // Clear the timeout
        clearTimeout(connectionTimeout);
        
        logger.error(`WebSocket error for chat ${chatId}:`, error);
        console.error(`[WebSocket] Error in connection for chat ${chatId}:`, error);
        console.log(`[WebSocket] Error details:`, {
          url: wsUrl,
          readyState: ws.readyState,
          protocol: ws.protocol,
          bufferedAmount: ws.bufferedAmount
        });
        
        update(s => ({ 
          ...s, 
          lastError: 'Connection error',
          connectionStatus: {
            ...s.connectionStatus,
            [chatId]: 'error'
          }
        }));
        
        // Use fallback mode
        useLocalFallback(chatId);
      };
      
      ws.onclose = (event) => {
        // Clear the timeout
        clearTimeout(connectionTimeout);
        
        logger.info(`WebSocket closed for chat ${chatId}: code=${event.code}, reason="${event.reason}", wasClean=${event.wasClean}`);
        console.log(`[WebSocket] Connection closed for chat ${chatId}:`, {
          code: event.code,
          reason: event.reason || 'No reason provided',
          wasClean: event.wasClean,
          timestamp: new Date().toISOString()
        });
        
        // Log explanations for common close codes
        const closeCodeMessages: Record<number, string> = {
          1000: 'Normal closure',
          1001: 'Going away (page unload)',
          1002: 'Protocol error',
          1003: 'Unsupported data',
          1005: 'No status received',
          1006: 'Abnormal closure (connection lost)',
          1007: 'Invalid frame payload data',
          1008: 'Policy violation',
          1009: 'Message too big',
          1010: 'Missing extension',
          1011: 'Internal server error',
          1012: 'Service restart',
          1013: 'Try again later',
          1015: 'TLS handshake failure'
        };
        
        const codeExplanation = closeCodeMessages[event.code] || 'Unknown close code';
        console.log(`[WebSocket] Close code explanation: ${codeExplanation}`);
        
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
          console.log(`[WebSocket] Will attempt reconnect for chat ${chatId} (non-clean close)`);
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
    } catch (error: unknown) {
      const errorMessage = error instanceof Error ? error.message : 'Connection creation failed';
      logger.error(`Failed to create WebSocket connection for chat ${chatId}:`, error);
      console.error(`[WebSocket] Connection creation error for chat ${chatId}:`, error);
      
      update(state => ({
        ...state,
        lastError: errorMessage,
        connectionStatus: {
          ...state.connectionStatus,
          [chatId]: 'error'
        }
      }));
      
      // Use fallback mode
      useLocalFallback(chatId);
    }
  };

  // Add a function to handle fallback mode when WebSocket is not available
  const useLocalFallback = (chatId: string) => {
    logger.info(`Using fallback mode for chat ${chatId}`);
    console.log(`[WebSocket] Using fallback mode for chat ${chatId}`);
    
    // Notify handlers that we're in fallback mode
    messageHandlers.forEach(handler => {
      try {
        const fallbackMessage: ChatMessage = {
          type: 'system',
          content: 'Using local mode due to connection issues. Messages may not be delivered to other users.',
          chat_id: chatId,
          timestamp: new Date().toISOString(),
          is_system: true,
          is_fallback: true
        };
        handler(fallbackMessage);
      } catch (e) {
        logger.error('Error notifying handler about fallback mode:', e);
      }
    });
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
    let result = false;
    
    update(state => {
      // Check if we have a connection for this chat
      const connection = state.chatConnections[chatId];
      if (connection && connection.readyState === WebSocket.OPEN) {
        result = true;
      }
      return state;
    });
    
    return result;
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