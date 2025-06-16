import { writable } from "svelte/store";
import { getAuthToken } from "../utils/auth";
import appConfig from "../config/appConfig";
import { createLoggerWithPrefix } from "../utils/logger";

const logger = createLoggerWithPrefix("WebSocketStore");

export type MessageType = "text" | "typing" | "read" | "edit" | "delete" | "system";

export interface ChatMessage {
  type: MessageType;
  content?: string;
  user_id?: string;
  sender_id?: string;
  sender_name?: string;
  sender_avatar?: string;
  chat_id: string;
  timestamp?: Date | string;
  message_id?: string;
  is_edited?: boolean;
  is_deleted?: boolean;
  is_read?: boolean;
  is_system?: boolean;
}

export interface WebSocketState {
  connected: boolean;
  reconnecting: boolean;
  lastError: string | null;
  chatConnections: Record<string, WebSocket>;
  connectionStatus: Record<string, "connecting" | "connected" | "disconnected" | "error">;
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

  const reconnectAttempts: Record<string, number> = {}; // Per-chat reconnection attempts
  let reconnectTimeouts: Record<string, number> = {};
  const lastConnectionAttempt: Record<string, number> = {}; // Throttling mechanism

  const buildWebSocketUrl = (chatId: string) => {
    try {
      const token = getAuthToken();
      if (!token) {
        throw new Error("No authentication token available");
      }

      // Derive WebSocket URL from the configured API base URL for consistency
      const apiUrl = new URL(appConfig.api.baseUrl);
      const protocol = apiUrl.protocol === "https:" ? "wss:" : "ws:";
      const hostname = apiUrl.hostname;
      const port = apiUrl.port || (protocol === "wss:" ? "443" : "80");

      // The path for the WebSocket connection
      const wsPath = `/api/v1/chats/${chatId}/ws`;

      const wsUrl = `${protocol}//${hostname}:${port}${wsPath}?token=${encodeURIComponent(token)}`;

      logger.info(`Built WebSocket URL: ${wsUrl}`);
      return wsUrl;
    } catch (e) {
      const errorMessage = e instanceof Error ? e.message : "Unknown error";
      logger.error(`Error building WebSocket URL: ${errorMessage}`);
      throw e;
    }
  };

  const connect = (chatId: string) => {
    logger.info(`Connecting to WebSocket for chat: ${chatId}`);

    // Throttle connection attempts - don't allow connections more frequent than every 2 seconds
    const now = Date.now();
    const lastAttempt = lastConnectionAttempt[chatId] || 0;
    if (now - lastAttempt < 2000) {
      logger.debug(`Connection throttled for chat ${chatId}, last attempt was ${now - lastAttempt}ms ago`);
      return;
    }
    lastConnectionAttempt[chatId] = now;

    // Validate chat ID format (UUID)
    if (!chatId || !/^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i.test(chatId)) {
      logger.error(`Invalid chat ID format: ${chatId}`);
      update(state => ({
        ...state,
        lastError: `Invalid chat ID format: ${chatId}`,
        connectionStatus: { ...state.connectionStatus, [chatId]: "error" }
      }));
      return;
    }

    update(state => ({
      ...state,
      connectionStatus: { ...state.connectionStatus, [chatId]: "connecting" },
      lastError: null
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
          logger.info(`Closed existing WebSocket connection for chat ${chatId}`);
        } catch (e) {
          logger.warn(`Error closing existing connection: ${e}`);
        }
        const connections = { ...state.chatConnections };
        delete connections[chatId];
        return { ...state, chatConnections: connections };
      }
      return state;
    });

    try {
      const wsUrl = buildWebSocketUrl(chatId);
      logger.info(`Attempting to connect to WebSocket: ${wsUrl}`);

      // Set a connection timeout
      const connectionTimeout = setTimeout(() => {
        logger.warn(`WebSocket connection timeout for chat ${chatId}`);

        update(s => ({
          ...s,
          lastError: `Connection timeout for chat ${chatId}`,
          connectionStatus: {
            ...s.connectionStatus,
            [chatId]: "error"
          }
        }));

        // Attempt to reconnect instead of falling back
        attemptReconnect(chatId);
      }, 15000); // Increased timeout to 15 seconds

      const ws = new WebSocket(wsUrl);

      ws.addEventListener("open", () => {
        clearTimeout(connectionTimeout);

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
            [chatId]: "connected"
          }
        }));

        // Reset reconnection attempts on successful connection
        reconnectAttempts[chatId] = 0;
        lastConnectionAttempt[chatId] = 0; // Clear throttling

        // Send an initial connection check message
        try {
          const token = getAuthToken();
          let userId = "";
          if (token) {
            try {
              const payload = JSON.parse(atob(token.split(".")[1]));
              userId = payload.user_id || payload.sub || "";
            } catch (e) {
              logger.warn("Could not extract user ID from token");
            }
          }

          const initialMessage = {
            type: "connection_check",
            user_id: userId,
            chat_id: chatId,
            timestamp: new Date().toISOString()
          };
          ws.send(JSON.stringify(initialMessage));
          logger.debug(`Sent initial connection check for chat ${chatId}`);
        } catch (e) {
          logger.error(`Error sending initial message for chat ${chatId}:`, e);
        }
      });
      ws.addEventListener("message", (event) => {
        handleWebSocketMessage(ws, chatId, event);
      });

      ws.addEventListener("error", (error) => {
        clearTimeout(connectionTimeout);

        logger.error(`WebSocket error for chat ${chatId}:`, error);

        update(s => ({
          ...s,
          lastError: `Connection error for chat ${chatId}`,
          connectionStatus: {
            ...s.connectionStatus,
            [chatId]: "error"
          }
        }));

        // Try to reconnect after a delay using per-chat counter
        const currentAttempts = reconnectAttempts[chatId] || 0;
        const reconnectDelay = Math.min(1000 * Math.pow(1.5, currentAttempts), 30000);
        reconnectAttempts[chatId] = currentAttempts + 1;

        logger.info(`Scheduling reconnection attempt ${currentAttempts + 1} in ${reconnectDelay}ms for chat ${chatId}`);
        reconnectTimeouts[chatId] = window.setTimeout(() => {
          attemptReconnect(chatId);
        }, reconnectDelay);
      });

      ws.addEventListener("close", (event) => {
        logger.info(`WebSocket connection closed for chat ${chatId}:`, event.code, event.reason);

        // Update connection status
        update(s => {
          const connections = { ...s.chatConnections };
          delete connections[chatId];

          return {
            ...s,
            connected: Object.keys(connections).length > 0,
            chatConnections: connections,
            connectionStatus: {
              ...s.connectionStatus,
              [chatId]: "disconnected"
            }
          };
        });

        // If this wasn't a normal closure, try to reconnect using per-chat counter
        if (event.code !== 1000 && event.code !== 1001) {
          const currentAttempts = reconnectAttempts[chatId] || 0;
          const reconnectDelay = Math.min(1000 * Math.pow(1.5, currentAttempts), 30000);
          reconnectAttempts[chatId] = currentAttempts + 1;

          logger.info(`Scheduling reconnection attempt ${currentAttempts + 1} in ${reconnectDelay}ms for chat ${chatId}`);
          reconnectTimeouts[chatId] = window.setTimeout(() => {
            attemptReconnect(chatId);
          }, reconnectDelay);
        }
      });

    } catch (error: unknown) {
      logger.error(`Error creating WebSocket connection for chat ${chatId}:`, error);

      update(s => ({
        ...s,
        lastError: `Connection error: ${error instanceof Error ? error.message : "Unknown error"}`,
        connectionStatus: {
          ...s.connectionStatus,
          [chatId]: "error"
        }
      }));

      // Attempt to reconnect instead of falling back
      attemptReconnect(chatId);
    }
  };

  const disconnect = (chatId: string) => {
    update(state => {
      if (state.chatConnections[chatId]) {
        state.chatConnections[chatId].close(1000, "Disconnect requested");
        const connections = { ...state.chatConnections };
        delete connections[chatId];

        const status = { ...state.connectionStatus };
        status[chatId] = "disconnected";

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
          state.chatConnections[chatId].close(1000, "Disconnect all requested");
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
        logger.warn(`No WebSocket connection for chat ${chatId}. Attempting to connect...`);
        connect(chatId);
        return {
          ...state,
          lastError: `Not connected to chat ${chatId}. Attempting to connect...`
        };
      }

      if (state.chatConnections[chatId].readyState !== WebSocket.OPEN) {
        logger.warn(`WebSocket for chat ${chatId} is not in OPEN state (current state: ${state.chatConnections[chatId].readyState}). Attempting to reconnect...`);

        // Close the existing connection
        try {
          state.chatConnections[chatId].close(1000, "Reconnecting due to non-OPEN state");
        } catch (e) {
          logger.error(`Error closing existing connection for chat ${chatId}:`, e);
        }

        // Remove from state
        const connections = { ...state.chatConnections };
        delete connections[chatId];

        // Schedule reconnection
        setTimeout(() => connect(chatId), 500);

        return {
          ...state,
          chatConnections: connections,
          lastError: "Connection not ready. Attempting to reconnect...",
          connectionStatus: {
            ...state.connectionStatus,
            [chatId]: "connecting"
          }
        };
      }

      try {
        const ws = state.chatConnections[chatId];
        // Ensure the message has the correct structure for WebSocket
        const wsMessage = {
          type: message.type || "text",
          content: message.content || "",
          user_id: message.user_id || message.sender_id || "",
          sender_id: message.sender_id || message.user_id || "",
          chat_id: chatId,
          timestamp: message.timestamp || new Date().toISOString(),
          message_id: message.message_id || `ws-${Date.now()}`,
          is_edited: message.is_edited || false,
          is_deleted: message.is_deleted || false,
          is_read: message.is_read || false
        };

        const serializedMessage = JSON.stringify(wsMessage);
        ws.send(serializedMessage);
        logger.debug(`WebSocket message sent to chat ${chatId}:`, wsMessage);

        return {
          ...state,
          lastError: null
        };
      } catch (e) {
        logger.error(`Error sending WebSocket message to chat ${chatId}:`, e);

        // Schedule reconnection on error
        setTimeout(() => connect(chatId), 1000);

        return {
          ...state,
          lastError: `Failed to send message: ${e instanceof Error ? e.message : "Unknown error"}`
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

  const attemptReconnect = (chatId: string) => {
    update(state => ({
      ...state,
      reconnecting: true,
      connectionStatus: {
        ...state.connectionStatus,
        [chatId]: "connecting"
      }
    }));

    const maxReconnectAttempts = 10; // Increased attempts for better reliability
    const currentAttempts = reconnectAttempts[chatId] || 0;

    if (currentAttempts >= maxReconnectAttempts) {
      logger.warn(`Maximum reconnect attempts (${maxReconnectAttempts}) reached for chat ${chatId}`);
      update(state => ({
        ...state,
        reconnecting: false,
        lastError: `Connection failed after ${maxReconnectAttempts} attempts. Please refresh the page.`,
        connectionStatus: {
          ...state.connectionStatus,
          [chatId]: "error"
        }
      }));
      return;
    }

    const baseDelay = 1000;
    const delay = Math.min(baseDelay * Math.pow(1.5, currentAttempts), 30000);
    reconnectAttempts[chatId] = currentAttempts + 1;

    if (reconnectTimeouts[chatId]) {
      clearTimeout(reconnectTimeouts[chatId]);
    }

    logger.info(`Attempting to reconnect to chat ${chatId} (attempt ${currentAttempts + 1}/${maxReconnectAttempts}) in ${delay}ms`);

    reconnectTimeouts[chatId] = window.setTimeout(() => {
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

// Function to handle incoming WebSocket messages
const handleWebSocketMessage = (ws: WebSocket, chatId: string, event: MessageEvent) => {
  try {
    // Log raw data for debugging
    logger.debug(`[WebSocket] Raw message received for chat ${chatId}`);

    // Try to parse the message as JSON
    let message: any;
    try {
      message = JSON.parse(event.data);
    } catch (parseError) {
      logger.debug("Failed to parse message as JSON");
      // Try to handle as plain text
      message = {
        type: "text",
        content: event.data,
        chat_id: chatId,
        timestamp: new Date().toISOString()
      };
    }

    // Ensure the message has the required properties
    const chatMessage: ChatMessage = {
      type: message.type || "text",
      content: message.content || "",
      chat_id: message.chat_id || chatId,
      user_id: message.user_id || message.sender_id || "",
      sender_id: message.sender_id || message.user_id || "",
      sender_name: message.sender_name || "User",
      sender_avatar: message.sender_avatar || null,
      timestamp: message.timestamp || new Date().toISOString(),
      message_id: message.message_id || message.id || `ws-${Date.now()}`,
      is_edited: message.is_edited || false,
      is_deleted: message.is_deleted || false,
      is_read: message.is_read || false,
      is_system: message.is_system || false
    };

    // Convert Unix timestamp to ISO string if needed
    if (typeof chatMessage.timestamp === "number") {
      chatMessage.timestamp = new Date(chatMessage.timestamp * 1000).toISOString();
    }

    // Log the processed message with more detail
    logger.debug(`[WebSocket] Processed message for chat ${chatId}:`, {
      type: chatMessage.type,
      content: chatMessage.content?.substring(0, 50),
      sender_id: chatMessage.sender_id,
      message_id: chatMessage.message_id
    });

    // Notify all registered handlers
    if (messageHandlers.length === 0) {
      logger.warn("[WebSocket] No message handlers registered to process message");
    }

    messageHandlers.forEach(handler => {
      try {
        logger.debug(`[WebSocket] Calling message handler for chat ${chatId}`, {
          handlerCount: messageHandlers.length,
          messageType: chatMessage.type
        });
        handler(chatMessage);
      } catch (handlerError) {
        logger.error(`[WebSocket] Error in message handler for chat ${chatId}:`, handlerError);
      }
    });
  } catch (e) {
    logger.error("Error handling WebSocket message", e);
  }
};