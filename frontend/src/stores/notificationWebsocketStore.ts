import { writable } from "svelte/store";
import { getAuthToken } from "../utils/auth";
import appConfig from "../config/appConfig";
import { createLoggerWithPrefix } from "../utils/logger";
import { notificationStore } from "./notificationStore";
import type { Notification } from "./notificationStore";

const logger = createLoggerWithPrefix("NotificationWebSocketStore");

export interface NotificationWebSocketState {
  connected: boolean;
  reconnecting: boolean;
  lastError: string | null;
}

const initialState: NotificationWebSocketState = {
  connected: false,
  reconnecting: false,
  lastError: null
};

type MessageHandler = (message: any) => void;
const messageHandlers: MessageHandler[] = [];

function createNotificationWebSocketStore() {
  const { subscribe, update, set } = writable<NotificationWebSocketState>(initialState);

  let ws: WebSocket | null = null;
  let reconnectAttempts = 0;
  let reconnectTimeout: number | null = null;
  const maxReconnectAttempts = 3; 
  let reconnectDelay = 1000;

  const connect = () => {
    if (ws && (ws.readyState === WebSocket.CONNECTING || ws.readyState === WebSocket.OPEN)) {
      logger.info("WebSocket already connecting or connected");
      console.log("[NotificationWebSocket] Connection already active or in progress");
      return;
    }

    try {
      if (ws) {
        try {
          ws.close();
        } catch (e) {
          logger.debug("Error closing existing WebSocket connection:", e);
          console.error("[NotificationWebSocket] Error closing existing connection:", e);
        }
        ws = null;
      }

      const wsBaseUrl = "ws://localhost:8083/api/v1";
      logger.debug(`Using direct WebSocket base URL: ${wsBaseUrl}`);
      console.log("[NotificationWebSocket] Using base URL:", wsBaseUrl);

      const token = getAuthToken();
      if (!token) {
        logger.error("No auth token available, cannot connect to WebSocket");
        console.error("[NotificationWebSocket] Authentication error: No token available");
        update(s => ({
          ...s,
          lastError: "No authentication token available"
        }));
        return;
      }

      let userId = "";
      try {
        if (token) {
          const base64Url = token.split(".")[1];
          const base64 = base64Url.replace(/-/g, "+").replace(/_/g, "/");
          const jsonPayload = decodeURIComponent(atob(base64).split("").map(function(c) {
            return "%" + ("00" + c.charCodeAt(0).toString(16)).slice(-2);
          }).join(""));

          const tokenData = JSON.parse(jsonPayload);
          userId = tokenData.user_id || tokenData.sub || "";
          logger.debug("Extracted user ID from token:", userId);
          console.log("[NotificationWebSocket] Extracted user ID:", userId.substring(0, 8) + "...");
        }
      } catch (e) {
        logger.error("Error decoding token:", e);
        console.error("[NotificationWebSocket] Failed to decode token:", e);
      }

      let wsUrl = `${wsBaseUrl}/notifications/ws`;

      const params: string[] = [];

      if (token) {
        params.push(`token=${encodeURIComponent(token)}`);
      }

      if (userId) {
        params.push(`user_id=${encodeURIComponent(userId)}`);
      }

      if (params.length > 0) {
        wsUrl += `?${params.join("&")}`;
      }

      logger.info(`Attempting to connect to Notification WebSocket: ${wsUrl.substring(0, 50)}...`);
      console.log("[NotificationWebSocket] Connecting to:", wsUrl.substring(0, 50) + "...");

      ws = new WebSocket(wsUrl);

      const connectionTimeout = setTimeout(() => {
        if (ws && ws.readyState !== WebSocket.OPEN) {
          logger.error("WebSocket connection timeout");
          console.error("[NotificationWebSocket] Connection timeout after 5 seconds");
          if (ws.readyState !== WebSocket.CLOSED && ws.readyState !== WebSocket.CLOSING) {
            ws.close();
          }

          attemptReconnect();
        }
      }, 5000); 

      ws.onopen = () => {
        clearTimeout(connectionTimeout);
        logger.info("Notification WebSocket connection established");
        console.log("[NotificationWebSocket] Connection successfully established");
        update(s => ({
          ...s,
          connected: true,
          reconnecting: false,
          lastError: null
        }));

        reconnectAttempts = 0;
        reconnectDelay = 1000; 
      };

      ws.onmessage = (event) => {
        try {
          logger.debug("Notification WebSocket message received:", event.data);
          const message = JSON.parse(event.data);

          if (message.type === "notification" && message.data) {
            notificationStore.addNotification(message.data as Notification);
          }

          if (message.type === "notification_bundle" && Array.isArray(message.notifications)) {
            message.notifications.forEach((notification: Notification) => {
              notificationStore.addNotification(notification);
            });
          }

          messageHandlers.forEach(handler => handler(message));
        } catch (e) {
          logger.error("Error parsing notification WebSocket message:", e);
          console.error("[NotificationWebSocket] Failed to parse message:", e, "Raw data:", event.data);
        }
      };

      ws.onerror = (error) => {
        clearTimeout(connectionTimeout);
        logger.error("Notification WebSocket error:", error);
        console.error("[NotificationWebSocket] Connection error:", error);

        console.log("[NotificationWebSocket] Connection details:", {
          url: wsUrl.substring(0, 50) + "...",
          readyState: ws?.readyState,
          protocol: ws?.protocol,
          userAgent: navigator.userAgent,
          timestamp: new Date().toISOString()
        });

        update(s => ({
          ...s,
          lastError: "Connection error"
        }));
      };

      ws.onclose = (event) => {
        clearTimeout(connectionTimeout);
        logger.info(`Notification WebSocket closed: code=${event.code}, reason="${event.reason}", wasClean=${event.wasClean}`);
        console.log("[NotificationWebSocket] Connection closed:", {
          code: event.code,
          reason: event.reason || "No reason provided",
          wasClean: event.wasClean,
          timestamp: new Date().toISOString()
        });

        const closeCodeMessages: Record<number, string> = {
          1000: "Normal closure",
          1001: "Going away (page unload)",
          1002: "Protocol error",
          1003: "Unsupported data",
          1005: "No status received",
          1006: "Abnormal closure (connection lost)",
          1007: "Invalid frame payload data",
          1008: "Policy violation",
          1009: "Message too big",
          1010: "Missing extension",
          1011: "Internal server error",
          1012: "Service restart",
          1013: "Try again later",
          1015: "TLS handshake failure"
        };

        const codeExplanation = closeCodeMessages[event.code] || "Unknown close code";
        console.log(`[NotificationWebSocket] Close code explanation: ${codeExplanation}`);

        update(s => ({
          ...s,
          connected: false
        }));

        if (event.code !== 1000) { 

          const currentPath = window.location.pathname;
          const notificationEnabledPaths = ["/feed", "/notifications"];

          if (notificationEnabledPaths.includes(currentPath)) {
            console.log("[NotificationWebSocket] Will attempt reconnect (non-clean close)");
            attemptReconnect();
          } else {
            console.log("[NotificationWebSocket] Not reconnecting - not on a notification-enabled page");
          }
        }
      };
    } catch (error) {
      logger.error("Failed to establish notification WebSocket connection:", error);
      console.error("[NotificationWebSocket] Fatal connection error:", error);

      const currentPath = window.location.pathname;
      const notificationEnabledPaths = ["/feed", "/notifications"];

      if (notificationEnabledPaths.includes(currentPath)) {
        attemptReconnect();
      }

      update(s => ({
        ...s,
        connected: false,
        lastError: "Failed to connect"
      }));
    }
  };

  const disconnect = () => {
    logger.info("Disconnecting notification WebSocket");
    console.log("[NotificationWebSocket] Disconnecting by request");

    if (ws) {
      try {
        ws.close(1000, "Disconnect requested");
        ws = null;
      } catch (e) {
        logger.error("Error closing notification WebSocket:", e);
        console.error("[NotificationWebSocket] Error while disconnecting:", e);
      }
    }

    if (reconnectTimeout !== null) {
      clearTimeout(reconnectTimeout);
      reconnectTimeout = null;
    }

    update(s => ({
      ...s,
      connected: false,
      reconnecting: false
    }));
  };

  const resetError = () => {
    update(state => ({
      ...state,
      lastError: null
    }));
  };

  const attemptReconnect = () => {
    if (reconnectTimeout !== null) {
      clearTimeout(reconnectTimeout);
    }

    if (reconnectAttempts >= maxReconnectAttempts) {
      logger.warn(`Maximum reconnect attempts (${maxReconnectAttempts}) reached. Giving up.`);
      console.warn(`[NotificationWebSocket] Giving up after ${maxReconnectAttempts} reconnect attempts`);
      update(s => ({
        ...s,
        reconnecting: false,
        lastError: `Failed to reconnect after ${maxReconnectAttempts} attempts`
      }));
      return;
    }

    reconnectAttempts++;

    const delay = Math.min(reconnectDelay * Math.pow(2, reconnectAttempts - 1), 30000);

    update(s => ({
      ...s,
      reconnecting: true
    }));

    logger.info(`Attempting to reconnect in ${delay}ms (attempt ${reconnectAttempts}/${maxReconnectAttempts})...`);

    reconnectTimeout = window.setTimeout(() => {
      logger.info(`Reconnecting now (attempt ${reconnectAttempts}/${maxReconnectAttempts})...`);
      connect();
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

  const isConnected = () => {
    return ws !== null && ws.readyState === WebSocket.OPEN;
  };

  const testConnection = async () => {
    try {

      const token = getAuthToken();
      if (!token) {
        return { success: false, message: "No auth token available" };
      }

      try {
        const response = await fetch(`${appConfig.api.baseUrl}/notifications`, {
          method: "GET",
          headers: {
            "Authorization": `Bearer ${token}`,
            "Content-Type": "application/json"
          }
        });

        if (!response.ok) {
          return {
            success: false,
            message: `API test failed: ${response.status} ${response.statusText}`,
            details: await response.text()
          };
        }

        logger.info("API test successful, backend is reachable");
      } catch (error) {
        return {
          success: false,
          message: "API test failed, cannot connect to backend",
          error
        };
      }

      return new Promise((resolve) => {
        const testWs = new WebSocket(`ws:

        const timeout = setTimeout(() => {
          testWs.close();
          resolve({
            success: false,
            message: "WebSocket connection test timed out"
          });
        }, 5000);

        testWs.onopen = () => {
          clearTimeout(timeout);
          testWs.close();
          resolve({
            success: true,
            message: "WebSocket connection test successful"
          });
        };

        testWs.onerror = (error) => {
          clearTimeout(timeout);
          testWs.close();
          resolve({
            success: false,
            message: "WebSocket connection test failed",
            error
          });
        };
      });
    } catch (error) {
      return {
        success: false,
        message: "WebSocket test failed with exception",
        error
      };
    }
  };

  return {
    subscribe,
    connect,
    disconnect,
    resetError,
    registerMessageHandler,
    isConnected,
    testConnection
  };
}

export const notificationWebsocketStore = createNotificationWebSocketStore();