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
}

const initialState: WebSocketState = {
  connected: false,
  reconnecting: false,
  lastError: null,
  chatConnections: {},
};

type MessageHandler = (message: any) => void;
const messageHandlers: MessageHandler[] = [];

function createWebSocketStore() {
  const { subscribe, update, set } = writable<WebSocketState>(initialState);

  let reconnectAttempts = 0;
  let reconnectTimeout: number | null = null;

  const connect = (chatId: string) => {
    update(state => {
      if (state.chatConnections[chatId]) {
        logger.info(`Already connected to chat ${chatId}`);
        return state;
      }

      try {
        const apiUrl = appConfig.api.baseUrl;
        let wsProtocol = 'ws:';
        if (apiUrl.startsWith('https:') || window.location.protocol === 'https:') {
          wsProtocol = 'wss:';
        }
        
        const domain = apiUrl.replace(/^https?:\/\//, '').split('/')[0];
        
        const apiPath = apiUrl.replace(/^https?:\/\/[^/]+/, '');
        
        const token = getAuthToken();
        
        let userId = '';
        try {
          if (token) {
            const base64Url = token.split('.')[1];
            const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
            const jsonPayload = decodeURIComponent(atob(base64).split('').map(function(c) {
              return '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2);
            }).join(''));
            
            const tokenData = JSON.parse(jsonPayload);
            userId = tokenData.user_id || tokenData.sub || '';
            logger.debug('Extracted user ID from token:', userId);
          }
        } catch (e) {
          logger.error('Error decoding token:', e);
        }
        
        let wsUrl = `${wsProtocol}//${domain}${apiPath}/chats/${chatId}/ws`;
        
        const params: string[] = [];
        
        if (token) {
          params.push(`token=${token}`);
        }
        
        if (userId) {
          params.push(`user_id=${userId}`);
        }
        
        if (params.length > 0) {
          wsUrl += `?${params.join('&')}`;
        }
        
        logger.info(`Attempting to connect to WebSocket: ${wsUrl}`);
        
        const ws = new WebSocket(wsUrl);
        
        ws.onopen = () => {
          logger.info(`WebSocket connection established for chat ${chatId}`);
          update(s => ({ 
            ...s, 
            connected: true, 
            reconnecting: false,
            lastError: null 
          }));
          
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
          logger.error(`WebSocket URL was: ${wsUrl}`);
          logger.error(`WebSocket ready state: ${ws.readyState}`);
          update(s => ({ 
            ...s, 
            lastError: 'Connection error' 
          }));
        };
        
        ws.onclose = (event) => {
          logger.info(`WebSocket closed for chat ${chatId}: code=${event.code}, reason="${event.reason}", wasClean=${event.wasClean}`);
          
          update(s => {
            const connections = { ...s.chatConnections };
            delete connections[chatId];
            
            return { 
              ...s, 
              connected: Object.keys(connections).length > 0,
              chatConnections: connections
            };
          });
          
          if (event.code !== 1000) {
            logger.info(`Will attempt to reconnect to chat ${chatId} due to non-clean close`);
            attemptReconnect(chatId);
          }
        };
        
        return { 
          ...state, 
          chatConnections: { 
            ...state.chatConnections, 
            [chatId]: ws 
          } 
        };
      } catch (error) {
        logger.error('Failed to establish WebSocket connection:', error);
        
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
      if (state.chatConnections[chatId]) {
        state.chatConnections[chatId].close(1000, 'Disconnect requested');
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
      Object.keys(state.chatConnections).forEach(chatId => {
        try {
          state.chatConnections[chatId].close(1000, 'Disconnect all requested');
        } catch (e) {
          logger.error(`Error closing WebSocket for chat ${chatId}:`, e);
        }
      });
      
      if (reconnectTimeout !== null) {
        clearTimeout(reconnectTimeout);
        reconnectTimeout = null;
      }
      
      return {
        ...state,
        chatConnections: {},
        connected: false,
        reconnecting: false
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
      reconnecting: true
    }));
    
    const maxReconnectAttempts = 10;
    
    if (reconnectAttempts >= maxReconnectAttempts) {
      logger.warn(`Maximum reconnect attempts (${maxReconnectAttempts}) reached for chat ${chatId}`);
      update(state => ({
        ...state,
        reconnecting: false,
        lastError: 'Maximum reconnect attempts reached'
      }));
      return;
    }
    
    const baseDelay = 1000;
    const delay = baseDelay * Math.pow(1.5, reconnectAttempts);
    reconnectAttempts++;
    
    if (reconnectTimeout !== null) {
      clearTimeout(reconnectTimeout);
    }
    
    reconnectTimeout = window.setTimeout(() => {
      logger.info(`Attempting to reconnect to chat ${chatId} (attempt ${reconnectAttempts})`);
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