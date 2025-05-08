import { writable } from 'svelte/store';
import { getAuthToken } from '../utils/auth';
import appConfig from '../config/appConfig';
import { createLoggerWithPrefix } from '../utils/logger';
import { processWebSocketMessage } from '../api/chat';

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
}

const initialState: WebSocketState = {
  connected: false,
  reconnecting: false,
  lastError: null,
  chatConnections: {},
};

// Create a registry of message handlers 
// This allows decoupling the WebSocket store from specific handlers
type MessageHandler = (message: any) => void;
const messageHandlers: MessageHandler[] = [];

function createWebSocketStore() {
  const { subscribe, update, set } = writable<WebSocketState>(initialState);

  // Reconnect logic with exponential backoff
  let reconnectAttempts = 0;
  let reconnectTimeout: number | null = null;

  const connect = (chatId: string) => {
    update(state => {
      // Check if already connected
      if (state.chatConnections[chatId]) {
        logger.info(`Already connected to chat ${chatId}`);
        return state;
      }

      try {
        // Construct WebSocket URL based on API URL
        const apiUrl = appConfig.api.baseUrl;
        let wsProtocol = 'ws:';
        if (apiUrl.startsWith('https:') || window.location.protocol === 'https:') {
          wsProtocol = 'wss:';
        }
        
        // Get the domain part of the API URL without protocol
        const domain = apiUrl.replace(/^https?:\/\//, '').split('/')[0];
        
        // Get the API path without domain
        const apiPath = apiUrl.replace(/^https?:\/\/[^/]+/, '');
        
        // Construct WebSocket URL with the complete path - no token needed
        const wsUrl = `${wsProtocol}//${domain}${apiPath}/chats/${chatId}/ws`;
        
        logger.info(`Connecting to WebSocket: ${wsUrl}`);
        
        // Connect without a token
        const ws = new WebSocket(wsUrl);
        
        ws.onopen = () => {
          logger.info('WebSocket connection established');
          update(s => ({ 
            ...s, 
            connected: true, 
            reconnecting: false,
            lastError: null 
          }));
          
          // Reset reconnect attempts on successful connection
          reconnectAttempts = 0;
        };
        
        ws.onmessage = (event) => {
          try {
            const message = JSON.parse(event.data);
            
            // Process the message using the chat API handler
            processWebSocketMessage(message);
            
            // Also pass message to any registered handlers
            messageHandlers.forEach(handler => handler(message));
          } catch (e) {
            logger.error('Error parsing WebSocket message:', e);
          }
        };
        
        ws.onerror = (error) => {
          logger.error('WebSocket error:', error);
          update(s => ({ 
            ...s, 
            lastError: 'Connection error' 
          }));
        };
        
        ws.onclose = (event) => {
          logger.info(`WebSocket closed: ${event.code} ${event.reason}`);
          
          // Remove the closed connection
          update(s => {
            const connections = { ...s.chatConnections };
            delete connections[chatId];
            
            return { 
              ...s, 
              connected: Object.keys(connections).length > 0,
              chatConnections: connections
            };
          });
          
          // Attempt to reconnect unless this was a clean close
          if (event.code !== 1000) {
            attemptReconnect(chatId);
          }
        };
        
        // Store the connection
        return { 
          ...state, 
          chatConnections: { 
            ...state.chatConnections, 
            [chatId]: ws 
          } 
        };
      } catch (error) {
        logger.error('Failed to establish WebSocket connection:', error);
        
        // Attempt to reconnect
        attemptReconnect(chatId);
        
        return { 
          ...state, 
          lastError: 'Failed to connect' 
        };
      }
    });
  };

  const disconnect = (chatId: string) => {
    update(state => {
      const ws = state.chatConnections[chatId];
      if (ws) {
        ws.close(1000, 'Client disconnecting');
        
        // Remove this connection
        const connections = { ...state.chatConnections };
        delete connections[chatId];
        
        return { 
          ...state, 
          chatConnections: connections,
          connected: Object.keys(connections).length > 0
        };
      }
      return state;
    });
  };

  const disconnectAll = () => {
    update(state => {
      // Close all connections
      Object.values(state.chatConnections).forEach(ws => {
        ws.close(1000, 'Client disconnecting');
      });
      
      return { 
        ...initialState 
      };
    });
    
    // Cancel any pending reconnect
    if (reconnectTimeout) {
      clearTimeout(reconnectTimeout);
      reconnectTimeout = null;
    }
  };

  const sendMessage = (chatId: string, message: ChatMessage) => {
    update(state => {
      const ws = state.chatConnections[chatId];
      if (!ws || ws.readyState !== WebSocket.OPEN) {
        logger.error(`Cannot send message - No active connection for chat ${chatId}`);
        return { 
          ...state, 
          lastError: 'No active connection' 
        };
      }
      
      try {
        ws.send(JSON.stringify(message));
        return state;
      } catch (error) {
        logger.error('Error sending message:', error);
        return { 
          ...state, 
          lastError: 'Failed to send message' 
        };
      }
    });
  };

  const resetError = () => {
    update(state => ({ 
      ...state, 
      lastError: null 
    }));
  };

  // Private function to handle reconnection
  const attemptReconnect = (chatId: string) => {
    update(s => ({ 
      ...s, 
      reconnecting: true 
    }));
    
    // Use exponential backoff for reconnect attempts
    const delay = Math.min(30000, 1000 * Math.pow(2, reconnectAttempts));
    
    logger.info(`Attempting to reconnect in ${delay}ms (attempt ${reconnectAttempts + 1})`);
    
    // Clear any existing timeout
    if (reconnectTimeout) {
      clearTimeout(reconnectTimeout);
    }
    
    // Set up the next reconnect attempt
    reconnectTimeout = window.setTimeout(() => {
      reconnectAttempts++;
      connect(chatId);
    }, delay);
  };

  // Register a message handler
  const registerMessageHandler = (handler: MessageHandler) => {
    messageHandlers.push(handler);
    return () => {
      // Return function to unregister handler
      const index = messageHandlers.indexOf(handler);
      if (index !== -1) {
        messageHandlers.splice(index, 1);
      }
    };
  };

  // Check if a specific chat is connected
  const isConnected = (chatId: string) => {
    let connected = false;
    
    update(state => {
      const ws = state.chatConnections[chatId];
      connected = ws && ws.readyState === WebSocket.OPEN;
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

// Export the singleton store instance
export const websocketStore = createWebSocketStore(); 