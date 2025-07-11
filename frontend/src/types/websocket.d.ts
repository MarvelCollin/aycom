import type { ChatMessage } from "../stores/websocketStore";

declare module "../stores/websocketStore" {
  interface WebSocketStore {
    connect: (chatId: string) => void;
    disconnect: (chatId: string) => void;
    disconnectAll: () => void;
    sendMessage: (chatId: string, message: ChatMessage) => void;
    resetError: () => void;
    isConnected: (chatId: string) => boolean;
    registerMessageHandler: (handler: (message: any) => void) => () => void;
  }
}