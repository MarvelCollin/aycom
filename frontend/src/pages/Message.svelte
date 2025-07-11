<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import { writable, get } from "svelte/store";
  import LeftSide from "../components/layout/LeftSide.svelte";
  import { useTheme } from "../hooks/useTheme";
  import type { IAuthStore } from "../interfaces/IAuth";
  import { createLoggerWithPrefix } from "../utils/logger";
  import { toastStore } from "../stores/toastStore";
  import { authStore } from "../stores/authStore";
  import { checkAuth, isWithinTime, handleApiError } from "../utils/common";
  import * as chatApi from "../api/chat";
  import { getProfile, searchUsers, getUserById, getAllUsers } from "../api/user";
  import { websocketStore } from "../stores/websocketStore";
  import type { ChatMessage, MessageType } from "../stores/websocketStore";
  import DebugPanel from "../components/common/DebugPanel.svelte";
  import CreateGroupChat from "../components/chat/CreateGroupChat.svelte";
  import NewChatModal from "../components/chat/NewChatModal.svelte";
  import ManageGroupMembers from "../components/chat/ManageGroupMembers.svelte";
  import ThemeToggle from "../components/common/ThemeToggle.svelte";
  import { transformApiUsers, type StandardUser } from "../utils/userTransform";
  import Toast from "../components/common/Toast.svelte";
  import type { Toast as ToastType, ToastType as ToastTypeEnum } from "../interfaces/IToast";
  import { formatRelativeTime } from "../utils/date";
  import { getAuthToken } from "../utils/auth";
  import { uploadMedia } from "../api/media";
  import type { Attachment, Chat, LastMessage, Message, Participant } from "../interfaces/IChat";

  function logError(message: string, error: any) {
    logger.error(message, error);
    console.error(message, error);
  }

  function logInfo(message: string, data?: any) {
    logger.info(message, data);
    console.log(message, data);
  }

  function logDebug(message: string, data?: any) {
    logger.debug(message, data);
    console.log(message, data);
  }

  function logWarn(message: string, data?: any) {
    logger.warn(message, data);
    console.warn(message, data);
  }

  function ensureStringTimestamp(timestamp: string | number | Date): string {
    if (typeof timestamp === "string") {
      return timestamp;
    } else if (timestamp instanceof Date) {
      return timestamp.toISOString();
    } else if (typeof timestamp === "number") {
      return new Date(timestamp).toISOString();
    }
    return new Date().toISOString();
  }

  async function sendMessageToApi(chatId: string, messageData: any) {

    const formattedData = {
      content: messageData.content || "" 
    };

    if (messageData.attachment) {
      formattedData["attachment"] = messageData.attachment;
    }

    logger.debug(`Sending message to chat ${chatId} with data:`, JSON.stringify(formattedData));

    return await chatApi.sendMessage(chatId, formattedData);
  }

  const {
    listChats,
    listMessages,
    sendMessage: apiSendMessage,
    unsendMessage: apiUnsendMessage,
    searchMessages,
    createChat,
    getChatHistoryList,
    testApiConnection,
    logAuthTokenInfo,
    setMessageHandler,
    listChatParticipants,
    deleteChat
  } = chatApi;

  import "../styles/pages/messages.css"; 

  interface Attachment {
    id: string;
    type: "image" | "gif" | "video";
    url: string;
    thumbnail?: string;
  }

  interface Message {
    id: string;
    chat_id: string;
    sender_id: string;
    sender_name?: string;
    sender_avatar?: string;
    content: string;
    timestamp: string | number | Date;
    is_read: boolean;
    is_edited: boolean;
    is_deleted: boolean;
    failed?: boolean;
    is_local?: boolean;
    attachments?: Array<{
      id: string;
      type: string;
      url: string;
      name?: string;
    }>;
  }

  interface Participant {
    id: string;
    username: string;
    display_name: string;
    avatar: string | null;
    is_verified: boolean;
  }

  interface LastMessage {
    content: string;
    timestamp: string;  
    sender_id: string;
    sender_name?: string;
  }

  interface Chat {
    id: string;
    type: "individual" | "group";
    name: string;
    avatar: string | null;
    participants: Participant[];
    last_message?: LastMessage;
    messages: Message[];
    unread_count: number;
    profile_picture_url: string | null;
    created_at: string;
    updated_at: string;
  }

  const logger = createLoggerWithPrefix("Message");

  const { theme } = useTheme();

  $: authState = $authStore;
  $: isDarkMode = $theme === "dark";

  let username = "";
  let displayName = "";
  let avatar = "https://secure.gravatar.com/avatar/0?d=mp"; 
  let isLoadingProfile = true;

  let selectedChat: Chat | null = null;
  let chats: Chat[] = [];
  let isLoadingChats = true;
  let isLoadingMessages = false;
  let newMessage = "";
  let searchQuery = "";
  let filteredChats: Chat[] = [];
  let selectedAttachments: Attachment[] = [];
  let isLoadingUsers = false;
  let userSearchResults: StandardUser[] = [];
  let userSearchQuery = "";

  let isMobile = false;
  let showMobileMenu = false;

  let showNewChatModal = false;
  let showCreateGroupModal = false;
  let showManageGroupModal = false;
  let showDebug = false;

  const showGroupChatModal = false;

  let isUploading = false;

  let showDeleteConfirm = false;
  let chatToDelete: Chat | null = null;

  async function handleAttachment(type: "image" | "gif") {
    logger.debug(`Attachment selection requested: ${type}`);

    const fileInput = document.createElement("input");
    fileInput.type = "file";
    fileInput.multiple = false;

    if (type === "image") {
      fileInput.accept = "image/jpeg,image/png,image/jpg";
    } else if (type === "gif") {
      fileInput.accept = "image/gif";
    }

    fileInput.onchange = async (event) => {
      const target = event.target as HTMLInputElement;
      const files = target.files;

      if (!files || files.length === 0) {
        return;
      }

      const file = files[0];

      const MAX_FILE_SIZE = 10 * 1024 * 1024; 
      if (file.size > MAX_FILE_SIZE) {
        toastStore.showToast({
          message: "File is too large. Maximum size is 10MB.",
          type: "error"
        });
        return;
      }

      try {
        isUploading = true;

        const result = await uploadMedia(file, "chat");

        if (!result || !result.url) {
          throw new Error("Failed to upload file");
        }

        const attachment: Attachment = {
          id: `temp-${Date.now()}`,
          type: result.mediaType === "video" ? "video" :
                (type === "gif" ? "gif" : "image"),
          url: result.url
        };

        selectedAttachments = [...selectedAttachments, attachment];

        logger.info(`Attachment uploaded successfully: ${attachment.type}`);

        if (selectedChat) {

          await sendMessageWithAttachment(attachment);
        }
      } catch (error) {
        logger.error("Failed to upload attachment:");
        toastStore.showToast({
          message: "Failed to upload attachment. Please try again.",
          type: "error"
        });
      } finally {
        isUploading = false;
      }
    };

    fileInput.click();
  }

  async function sendMessageWithAttachment(attachment: Attachment) {
    if (!selectedChat) return;

    try {

      const tempMessageId = `temp-${Date.now()}`;

      const content = JSON.stringify({
        text: "",
        attachment: {
          type: attachment.type,
          url: attachment.url
        }
      });

      const message: Message = {
        id: tempMessageId,
        chat_id: selectedChat.id,
        sender_id: $authStore.user_id || "",
        content: content,
        timestamp: new Date().toISOString(),
        is_read: false,
        is_edited: false,
        is_deleted: false,
        sender_name: displayName || "You",
        sender_avatar: avatar,
        is_local: true,
        attachments: [attachment]
      };

      selectedChat = {
        ...selectedChat,
        messages: [...selectedChat.messages, message]
      };

      setTimeout(() => {
        const messagesContainer = document.querySelector(".messages-container");
        if (messagesContainer) {
          messagesContainer.scrollTop = messagesContainer.scrollHeight;
        }
      }, 50);

      const newLastMessage: LastMessage = {
        content: attachment.type === "image" ? "📷 Image" :
                 attachment.type === "gif" ? "🎞️ GIF" :
                 attachment.type === "video" ? "🎥 Video" : "Attachment",
        timestamp: new Date().toISOString(),
        sender_id: $authStore.user_id || "",
        sender_name: displayName || "You"
      };

      chats = chats.map(chat => {
        if (chat.id === selectedChat?.id) {
          return {
            ...chat,
            last_message: newLastMessage
          };
        }
        return chat;
      }) as Chat[];

      const activeChatId = selectedChat?.id;
      if (activeChatId) {
        const activeChat = chats.find(c => c.id === activeChatId);
        if (activeChat) {

          const otherChats = chats.filter(c => c.id !== activeChatId);

          chats = [activeChat, ...otherChats];

          const filteredActiveChat = filteredChats.find(c => c.id === activeChatId);
          if (filteredActiveChat) {
            const otherFilteredChats = filteredChats.filter(c => c.id !== activeChatId);
            filteredChats = [
              {
                ...filteredActiveChat,
                last_message: newLastMessage
              },
              ...otherFilteredChats
            ];
          }
        }
      }

      const messageData = {
        content: content,
        message_id: tempMessageId,
        attachments: [attachment]
      };

      logger.debug(`Sending message with attachment to chat ${selectedChat?.id} via API`);

      const wsMessage = {
        type: "text" as MessageType,
        content: content,
        chat_id: selectedChat?.id || "",
        user_id: $authStore.user_id || "",
        sender_id: $authStore.user_id || "",
        sender_name: displayName || username || "User",
        sender_avatar: avatar,
        message_id: tempMessageId,
        timestamp: new Date().toISOString()
      };

      websocketStore.sendMessage(selectedChat?.id || "", wsMessage);
      logger.debug(`Message with attachment sent via WebSocket to chat ${selectedChat?.id}`);

      try {
        const result = await sendMessageToApi(selectedChat?.id || "", {
          content: messageData.content,
          message_id: messageData.message_id
        });
        logger.debug("Message with attachment sent successfully via API");

        if (selectedChat) {
          selectedChat = {
            ...selectedChat,
            messages: selectedChat.messages.map(msg =>
              msg.id === tempMessageId
                ? {
                    ...msg,
                    is_local: false,
                    id: result?.message?.id || result?.message_id || msg.id
                  }
                : msg
            )
          };
        }
      } catch (apiError) {
        logger.error("Failed to send message with attachment via API");
        toastStore.showToast({
          message: "Network issue detected. Message may not be delivered.",
          type: "error"
        });

        if (selectedChat) {
          selectedChat = {
            ...selectedChat,
            messages: selectedChat.messages.map(msg =>
              msg.id === tempMessageId
                ? { ...msg, failed: true }
                : msg
            )
          };
        }
      }
    } catch (error: unknown) {
      const errorMessage = error instanceof Error ? error.message : "Unknown error";
      logger.error("Error sending message with attachment");
      toastStore.showToast({
        message: `Error sending message with attachment: ${errorMessage}`,
        type: "error"
      });
    } finally {

      selectedAttachments = [];
    }
  }

  function checkViewport() {
    isMobile = window.innerWidth < 768;
  }

  $: {
    if ($websocketStore) {
      const isWsConnected = $websocketStore.connected;
      logger.debug(`WebSocket connection status: ${isWsConnected ? "connected" : "disconnected"}`);
    }
  }

  function initializeWebSocketConnections() {
    try {

      if (selectedChat) {
        logger.info(`Connecting to WebSocket for selected chat: ${selectedChat.id}`);
        try {

          (websocketStore as any).connect(selectedChat.id);
        } catch (err) {
          logger.error(`Error connecting to WebSocket for selected chat: ${err}`);
        }
      }

      if (chats && chats.length > 0) {
        const recentChats = chats.slice(0, 3); 

        for (const chat of recentChats) {

          if (selectedChat && chat.id === selectedChat.id) continue;

          logger.debug(`Connecting to WebSocket for recent chat: ${chat.id}`);
          try {

            (websocketStore as any).connect(chat.id);
          } catch (err) {

            logger.error(`Error connecting to WebSocket for chat ${chat.id}: ${err}`);
          }
        }
      }
    } catch (error) {
      logger.error("Error initializing WebSocket connections:", error);
    }
  }

  function handleWebSocketMessage(message: ChatMessage) {
    logDebug("Received WebSocket message");
    console.log("WebSocket message details:", message);

    if (!message || !message.chat_id) {
      logger.warn("Invalid message format received from WebSocket");
      return;
    }

    logger.debug(`[WebSocket] Message received for chat ${message.chat_id}:`, {
      type: message.type,
      content: message.content,
      sender: message.user_id || message.sender_id,
      timestamp: message.timestamp
    });

    if (message.type === "system") {

      logger.info(`System message: ${message.content}`);
      return;
    }      
      if (selectedChat && message.chat_id === selectedChat.id) {

        const currentUserId = $authStore.user_id;
        const messageSenderId = message.user_id || message.sender_id;

        if (messageSenderId === currentUserId) {
          logger.debug("Skipping own message from WebSocket (already displayed optimistically)", {
            messageId: message.message_id,
            content: message.content?.substring(0, 50)
          });
          return;
        }

        if (message.type === "text" && message.content) {

          const messageExists = selectedChat.messages.some(msg => {

            if (msg.id === message.message_id || (message.message_id && msg.id === message.message_id)) {
              return true;
            }

            if (msg.content === message.content &&
                msg.sender_id === messageSenderId &&
                msg.timestamp && message.timestamp) {
              const timeDiff = Math.abs(new Date(msg.timestamp).getTime() - new Date(message.timestamp).getTime());
              if (timeDiff < 5000) { 
                return true;
              }
            }

            if (msg.id?.startsWith("temp-") && message.content === msg.content && messageSenderId === msg.sender_id) {
              return true;
            }

            return false;
          });

          if (messageExists) {
            logger.debug("Message already exists in chat, skipping duplicate", {
              messageId: message.message_id,
              content: message.content?.substring(0, 50)
            });
            return;
          }

        const newMessage: Message = {
          id: message.message_id || `ws-${Date.now()}`,
          chat_id: message.chat_id,
          sender_id: messageSenderId || "",
          content: message.content || "",
          timestamp: typeof message.timestamp === "string"
            ? message.timestamp
            : message.timestamp instanceof Date
              ? message.timestamp.toISOString()
              : new Date().toISOString(),
          is_read: false,
          is_edited: false,
          is_deleted: false,
          sender_name: message.sender_name || "User",
          sender_avatar: message.sender_avatar
        };

        logger.info(`Adding new message from WebSocket to chat ${message.chat_id}:`, newMessage);

        selectedChat = {
          ...selectedChat,
          messages: [...(selectedChat.messages || []), newMessage]
        };

        setTimeout(() => {
          const messagesContainer = document.querySelector(".messages-container");
          if (messagesContainer) {
            messagesContainer.scrollTop = messagesContainer.scrollHeight;
          }
        }, 100);
      }
    }

    if (message.type === "text" && message.content) {

      const chatIndex = chats.findIndex(c => c.id === message.chat_id);
      if (chatIndex >= 0) {

        const lastMessage: LastMessage = {
          content: message.content || "",
          timestamp: typeof message.timestamp === "string"
            ? message.timestamp
            : message.timestamp instanceof Date
              ? message.timestamp.toISOString()
              : new Date().toISOString(),
          sender_id: message.user_id || message.sender_id || "",
          sender_name: message.sender_name || "User"
        };

        const updatedChat = {
          ...chats[chatIndex],
          last_message: lastMessage,

          unread_count: selectedChat?.id === message.chat_id
            ? chats[chatIndex].unread_count
            : (chats[chatIndex].unread_count || 0) + 1
        };

        const updatedChats = [
          updatedChat,
          ...chats.filter(c => c.id !== message.chat_id)
        ];

        chats = deduplicateChats(updatedChats);

        const filteredIndex = filteredChats.findIndex(c => c.id === message.chat_id);
        if (filteredIndex >= 0) {
          const updatedFilteredChat = {
            ...filteredChats[filteredIndex],
            last_message: lastMessage,
            unread_count: selectedChat?.id === message.chat_id
              ? filteredChats[filteredIndex].unread_count
              : (filteredChats[filteredIndex].unread_count || 0) + 1
          };

          const tempFilteredChats = [
            updatedFilteredChat,
            ...filteredChats.filter(c => c.id !== message.chat_id)
          ];
          filteredChats = deduplicateChats(tempFilteredChats);
        }

        if (selectedChat?.id !== message.chat_id) {
          logger.debug("Would play notification sound for new message");
        }
      } else {
        logger.warn(`Chat with ID ${message.chat_id} not found in chat list`);
      }
    }
  }

  function sendWebSocketMessage(chatId: string, content: string) {
    const chatMessage: ChatMessage = {
      type: "text",
      content: content,
      chat_id: chatId,
      user_id: $authStore.user_id || ""
    };

    (websocketStore as any).sendMessage(chatId, chatMessage);
  }

  function handleMobileNavigation(view: string): void {
    if (view === "back" || view === "showChats") {
      selectedChat = null;
    } else if (view === "showChat" && selectedChat) {

    }
  }

  function toggleMobileMenu() {
    showMobileMenu = !showMobileMenu;
  }

  function formatGroupChatForDisplay(apiChat: any): Chat {
    const avatar = null;

    return {
      id: apiChat.id,
      type: apiChat.is_group_chat ? "group" : "individual",
      name: apiChat.name || "Group Chat",
      avatar: avatar,
      participants: (apiChat.participants || []).map(p => ({
        id: p.id || p.user_id,
        username: p.username || "User",
        display_name: p.display_name || p.username || "User",
        avatar: p.profile_picture_url || p.avatar || null,
        is_verified: p.is_verified || false
      })),
      messages: [],
      unread_count: 0,
      profile_picture_url: null,
      created_at: apiChat.created_at || new Date().toISOString(),
      updated_at: apiChat.updated_at || new Date().toISOString()
    };
  }

  function formatTimeForChat(timestamp: string | number | Date): string {
    let date: Date;

    if (timestamp instanceof Date) {
      date = timestamp;
    } else if (typeof timestamp === "number") {
      date = new Date(timestamp);
    } else {
      date = new Date(timestamp);
    }

    return date.toLocaleString("en-US", {
      hour: "numeric",
      minute: "numeric",
      hour12: true
    });
  }

  function deduplicateChats(chatList: Chat[]): Chat[] {
    const chatMap = new Map<string, Chat>();

    chatList.forEach(chat => {
      if (chat && chat.id && !chatMap.has(chat.id)) {
        chatMap.set(chat.id, chat);
      }
    });

    const result = Array.from(chatMap.values());
    logger.debug(`Deduplicated ${chatList.length} chats to ${result.length} unique chats`);
    return result;
  }

  async function fetchChats() {
    isLoadingChats = true;

    try {
      const response = await getChatHistoryList();

      if (response && response.chats) {

        const processedChats = response.chats.map(chat => {

          const chatObj: Chat = {
            id: chat.id,
            type: chat.is_group_chat ? "group" : "individual",
            name: chat.name || "",
            avatar: null,
            participants: (chat.participants || []).map(p => {

              return {
                id: p.id || p.user_id || "",
                username: p.username || "",
                display_name: p.display_name || p.name || p.username || "",
                avatar: p.profile_picture_url || p.avatar || null,
                is_verified: p.is_verified || false
              };
            }),
            messages: [],
            unread_count: chat.unread_count || 0,
            profile_picture_url: null,
            created_at: chat.created_at || new Date().toISOString(),
            updated_at: chat.updated_at || new Date().toISOString()
          };

          if (chat.last_message) {
            chatObj.last_message = {
              content: chat.last_message.content || "",
              timestamp: ensureStringTimestamp(chat.last_message.timestamp || chat.last_message.sent_at || new Date()),
              sender_id: chat.last_message.sender_id || "",
              sender_name: chat.last_message.sender_name || "User"
            };
          }

          return chatObj;        });

        const uniqueChats = deduplicateChats(processedChats);

        uniqueChats.sort((a, b) => {
          const timeA = new Date(a.updated_at).getTime();
          const timeB = new Date(b.updated_at).getTime();
          return timeB - timeA;
        });

        chats = uniqueChats;
        filteredChats = [...uniqueChats];

        if (!selectedChat && processedChats.length > 0) {
          selectChat(processedChats[0]);
        }
      } else {
        logWarn("No chats found in response");
        console.warn("API response details:", response);
        chats = [];
        filteredChats = [];
      }
    } catch (error) {
      logError("Failed to load chats", error);
      chats = [];
      filteredChats = [];
    } finally {
      isLoadingChats = false;
    }
  }

  function formatMessageTime(timestamp: string | number | Date): string {
    let stringTimestamp: string;

    if (timestamp instanceof Date) {
      stringTimestamp = timestamp.toISOString();
    } else if (typeof timestamp === "number") {
      stringTimestamp = new Date(timestamp).toISOString();
    } else {
      stringTimestamp = timestamp;
    }

    return formatRelativeTime(stringTimestamp);
  }

  async function searchForUsers(query: string) {
    if (!query || query.length < 2) {
      userSearchResults = [];
      return;
    }

    userSearchQuery = query;
    isLoadingUsers = true;

    try {
      const results = await searchUsers(query);
      userSearchResults = transformApiUsers(results);
    } catch (error) {
      logError("Failed to search for users", error);
      userSearchResults = [];
    } finally {
      isLoadingUsers = false;
    }
  }

  async function initiateNewChat(data: any) {
    try {
      isLoadingChats = true;

      let chatData;
      if (typeof data === "string") {

        chatData = {
          type: "individual",
          participants: [data]
        };
      } else if (typeof data === "object") {

        chatData = data;

        if (data.participant_ids && !data.participants) {
          chatData = {
            ...data,
            participants: data.participant_ids
          };
          delete chatData.participant_ids;
        }
      } else {
        throw new Error("Invalid chat data format");
      }

      const response = await createChat(chatData);

      if (response && response.chat_id) {
        showNewChatModal = false;
        await fetchChats(); 
        selectChat(response.chat_id);
      }
    } catch (error) {
      logError("Failed to create chat", error);
    } finally {
      isLoadingChats = false;
    }
  }

  function getUserDisplayName(userId: string): string {

    if (userId === $authStore.user_id) {
      return displayName || "You";
    }

    for (const chat of chats) {
      const participant = chat.participants.find(p => p.id === userId);
      if (participant) {
        return participant.display_name || participant.username || "Unknown User";
      }
    }

    const shortId = userId.substring(0, 4);
    return `User ${shortId}`;
  }

  function getOtherParticipant(chat: Chat): Participant | undefined {
    if (chat.type !== "individual") return undefined;

    return chat.participants.find(p => p.id !== $authStore.user_id);
  }

  onMount(() => {

    checkViewport();
    window.addEventListener("resize", checkViewport);

    const initialize = async () => {

    try {
      const profileData = await getProfile();
      if (profileData) {
        username = profileData.username || "";
        displayName = profileData.display_name || profileData.username || "User";
        avatar = profileData.profile_picture_url || "https://secure.gravatar.com/avatar/0?d=mp";
      }
    } catch (error) {
      logError("Failed to load profile", error);
    } finally {
      isLoadingProfile = false;
    }

    await fetchChats();
    };

    initialize();

    const unregisterHandler = websocketStore.registerMessageHandler(handleWebSocketMessage);

    setMessageHandler(handleWebSocketMessage);

    logger.info("Message component mounted");

    return () => {
      if (unregisterHandler) unregisterHandler();
    };
  });

  onDestroy(() => {
    window.removeEventListener("resize", checkViewport);
    logger.info("Disconnecting from all WebSocket connections");
    websocketStore.disconnectAll();
  });

  function safeFormatRelativeTime(timestamp: string | Date | unknown): string {
    if (typeof timestamp === "string") {
      return formatRelativeTime(timestamp);
    } else if (timestamp instanceof Date) {
      return formatRelativeTime(timestamp.toISOString());
    } else {

      return formatRelativeTime(new Date().toISOString());
    }
  }

  async function selectChat(chat: Chat | string) {
    let chatId: string;

    if (typeof chat === "string") {
      chatId = chat;

      const chatObj = chats.find(c => c.id === chatId);
      if (!chatObj) {
        logger.error(`Chat with ID ${chatId} not found in chats list`);
        toastStore.showToast({
          message: "Chat not found. Please try again.",
          type: "error"
        });
        return;
      }
      chat = chatObj;
    } else {
      chatId = chat.id;
    }

    logger.info(`Selecting chat: ${chatId}`);

    if (!chatId || !/^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i.test(chatId)) {
      logger.error(`Invalid chat ID format: ${chatId}`);
      toastStore.showToast({
        message: `Invalid chat ID format: ${chatId}. Please try again or contact support.`,
        type: "error"
      });
      return;
    }

    selectedChat = { ...chat, messages: [] };

    if (isMobile) {
      showMobileMenu = false;
      handleMobileNavigation("showChat");
    }

    isLoadingMessages = true;
    try {
      logger.debug(`Fetching messages for chat ${chatId}`);
      const response = await listMessages(chatId);

      if (response && response.messages) {

        const sortedMessages = [...response.messages].sort((a, b) => {
          const timeA = new Date(a.timestamp).getTime();
          const timeB = new Date(b.timestamp).getTime();
          return timeA - timeB;
        });

        const processedMessages = sortedMessages.map(msg => ({
          ...msg,
          sender_name: msg.sender_name || "User",
          sender_avatar: msg.sender_avatar || null,
          timestamp: ensureStringTimestamp(msg.timestamp)
        }));

        selectedChat = {
          ...selectedChat,
          messages: processedMessages
        };

        logger.info(`Loaded ${processedMessages.length} messages for chat ${chatId}`);

        setTimeout(() => {
          const messagesContainer = document.querySelector(".messages-container");
          if (messagesContainer) {
            messagesContainer.scrollTop = messagesContainer.scrollHeight;
          }
        }, 100);
      } else {
        logWarn(`No messages found for chat ${chatId}`);
        selectedChat = {
          ...selectedChat,
          messages: []
        };
      }
    } catch (error) {
      logError(`Error loading messages for chat ${chatId}`, error);
      toastStore.showToast({ message: "Failed to load messages", type: "error" });
      selectedChat = {
        ...selectedChat,
        messages: []
      };
    } finally {
      isLoadingMessages = false;
    }

    try {

      const isConnected = (websocketStore as any).isConnected(chatId);
      if (!isConnected) {
        logger.info(`Connecting to WebSocket for chat ${chatId}`);
        (websocketStore as any).connect(chatId);
      } else {
        logger.debug(`Already connected to WebSocket for chat ${chatId}`);
      }
    } catch (error) {
      logError(`Error connecting to WebSocket for chat ${chatId}`, error);
      toastStore.showToast({ message: "Could not establish real-time connection", type: "warning" });
    }

    chats = chats.map(c => {
      if (c.id === chatId) {
        return { ...c, unread_count: 0 };
      }
      return c;
    }) as Chat[];

    filteredChats = [
      ...(filteredChats.filter(c => c.id === chatId)),
      ...(filteredChats.filter(c => c.id !== chatId))
    ].map(chat => ({
      ...chat,

      last_message: chat.last_message ? {
        ...chat.last_message,
        timestamp: typeof chat.last_message.timestamp === "string"
          ? chat.last_message.timestamp
          : new Date(chat.last_message.timestamp).toISOString()
      } : undefined
    })) as Chat[];
  }

  function getAvatarColor(name: string) {

    let hash = 0;
    for (let i = 0; i < name.length; i++) {
      hash = name.charCodeAt(i) + ((hash << 5) - hash);
    }

    const h = Math.abs(hash) % 360;
    return `hsl(${h}, 70%, 60%)`;
  }

  function getChatDisplayName(chat: Chat) {

    if (chat.type === "group" && chat.name && chat.name.trim() !== "") {
      return chat.name;
    }

    if (chat.participants && chat.participants.length > 0) {

      const otherParticipants = chat.participants.filter(p => p.id !== $authStore.user_id);

      if (otherParticipants.length > 0) {
        const participant = otherParticipants[0];
        return participant.display_name || participant.username || "Unknown User";
      }
    }

    return chat.name && chat.name.trim() !== "" ? chat.name : "Chat";
  }

  async function handleSearch() {
    if (!searchQuery || searchQuery.trim() === "") {
      filteredChats = [...chats];
      return;
    }

    const query = searchQuery.toLowerCase().trim();
    filteredChats = chats.filter(chat => {

      const chatName = getChatDisplayName(chat).toLowerCase();
      if (chatName.includes(query)) return true;

      if (chat.last_message && chat.last_message.content) {
        const messageContent = chat.last_message.content.toLowerCase();
        if (messageContent.includes(query)) return true;
      }

      if (chat.participants) {
        for (const participant of chat.participants) {
          const name = participant.display_name || participant.username || "";
          if (name.toLowerCase().includes(query)) return true;
        }
      }

      return false;
    });
  }

  async function sendMessage(content: string) {
    if (!content || !content.trim() || !selectedChat) {
      return;
    }

    try {

      const tempMessageId = `temp-${Date.now()}`;

      content = content.trim();
      newMessage = "";

      const message: Message = {
      id: tempMessageId,
        chat_id: selectedChat.id,
      sender_id: $authStore.user_id || "",
        content: content,
        timestamp: new Date().toISOString(),
      is_read: false,
      is_edited: false,
        is_deleted: false,
        sender_name: displayName || "You",
        sender_avatar: avatar,
        is_local: true
      };

      selectedChat = {
        ...selectedChat,
        messages: [...selectedChat.messages, message]
      };

        setTimeout(() => {
          const messagesContainer = document.querySelector(".messages-container");
          if (messagesContainer) {
            messagesContainer.scrollTop = messagesContainer.scrollHeight;
          }
      }, 50);

      const newLastMessage: LastMessage = {
        content: content,
        timestamp: new Date().toISOString(),
                sender_id: $authStore.user_id || "",
        sender_name: displayName || "You"
      };

      chats = deduplicateChats(chats.map(chat => {
        if (chat.id === selectedChat?.id) {
          return {
            ...chat,
            last_message: newLastMessage
            };
          }
          return chat;
      }) as Chat[]);

      const activeChatId = selectedChat?.id;
      if (activeChatId) {
      const activeChat = chats.find(c => c.id === activeChatId);
      if (activeChat) {

          const otherChats = chats.filter(c => c.id !== activeChatId);

          chats = [activeChat, ...otherChats];

          const filteredActiveChat = filteredChats.find(c => c.id === activeChatId);
          if (filteredActiveChat) {
            const otherFilteredChats = filteredChats.filter(c => c.id !== activeChatId);
        filteredChats = deduplicateChats([
              {
                ...filteredActiveChat,
                last_message: newLastMessage
              },
              ...otherFilteredChats
            ]);
          }
        }
      }

      const messageData: Record<string, any> = {
        content: content, 
        sender_id: $authStore.user_id,
        attachment: selectedAttachments.length > 0 ? selectedAttachments[0] : null
      };

      logger.debug(`Sending message via API to chat ${selectedChat?.id}`);

      const wsMessage = {
        type: "text" as MessageType,
        content: content,
        chat_id: selectedChat?.id || "",
        user_id: $authStore.user_id || "",
        sender_id: $authStore.user_id || "",
        sender_name: displayName || username || "You",
        sender_avatar: avatar,
        message_id: tempMessageId,
        timestamp: new Date().toISOString()
      };

      try {

        websocketStore.sendMessage(selectedChat?.id || "", wsMessage);
        logger.debug(`Message sent via WebSocket to chat ${selectedChat?.id}`);
      } catch (wsError) {
        logWarn("Failed to send message via WebSocket, continuing with API", wsError);
      }

      try {
        const result = await sendMessageToApi(selectedChat?.id || "", messageData);
        logInfo("Message sent successfully via API");

        if (selectedChat) {
          selectedChat = {
            ...selectedChat,
            messages: selectedChat.messages.map(msg =>
              msg.id === tempMessageId
                ? {
                    ...msg,
                    is_local: false,
                    id: result?.message?.id || result?.message_id || msg.id
                  }
                : msg
            )
          };
        }
      } catch (apiError) {
        logError("Failed to send message via API", apiError);
        toastStore.showToast({
          message: "Failed to send message. Please try again.",
          type: "error"
        });

        if (selectedChat) {
          selectedChat = {
            ...selectedChat,
            messages: selectedChat.messages.filter(msg => msg.id !== tempMessageId)
          };
        }
      }
    } catch (error: unknown) {
      const errorMessage = error instanceof Error ? error.message : "Unknown error";
      logError("Error sending message", errorMessage);
      toastStore.showToast({
        message: `Error sending message: ${errorMessage}`,
        type: "error"
      });
    }
  }

  async function unsendMessage(messageId: string) {

    if (!selectedChat) return;

    const message = selectedChat.messages.find(m => m.id === messageId);
    if (!message || message.sender_id !== $authStore.user_id) return;

    try {

      selectedChat = {
        ...selectedChat,
        messages: selectedChat.messages.map(msg =>
          msg.id === messageId ? { ...msg, is_deleted: true, content: "Message deleted" } : msg
        )
      };

      await apiUnsendMessage(selectedChat.id, messageId);

      const lastMessage = selectedChat.last_message;
      if (lastMessage && lastMessage.content === message.content) {

        const previousMessages = selectedChat.messages
          .filter(m => !m.is_deleted && m.id !== messageId)
          .sort((a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime());

        const newLastMessage = previousMessages[0];

        if (newLastMessage) {

          chats = chats.map(chat => {
            if (chat.id === selectedChat?.id) {
              return {
                ...chat,
                last_message: {
                  content: newLastMessage.content,
                  timestamp: ensureStringTimestamp(newLastMessage.timestamp),
                  sender_id: newLastMessage.sender_id,
                  sender_name: newLastMessage.sender_name || ""
                }
              };
            }
            return chat;
          }) as Chat[];

          filteredChats = filteredChats.map(chat => {
            if (chat.id === selectedChat?.id) {
              return {
                ...chat,
                last_message: {
                  content: newLastMessage.content,
                  timestamp: ensureStringTimestamp(newLastMessage.timestamp),
                  sender_id: newLastMessage.sender_id,
                  sender_name: newLastMessage.sender_name || ""
                }
              };
            }
            return chat;
          }) as Chat[];
        }
      }
    } catch (error: unknown) {
      const errorMessage = error instanceof Error ? error.message : "Unknown error";
      logError("Failed to unsend message", errorMessage);
      toastStore.showToast({ message: `Failed to unsend message: ${errorMessage}`, type: "error" });

      if (selectedChat) {

        const originalMessage = selectedChat.messages.find(m => m.id === messageId);
        if (originalMessage) {
          selectedChat = {
            ...selectedChat,
            messages: selectedChat.messages.map(msg =>
              msg.id === messageId ? { ...originalMessage, is_deleted: false, content: originalMessage.content } : msg
            )
          };
        }
      }
    }
  }

  async function handleCreateGroupChat(data: any) {
    try {
      isLoadingChats = true;

      let chatData: { type: string; name: string; participants: string[] };

      if (data && data.chat) {

        chatData = {
          type: "group",
          name: data.chat.name || "New Group",
          participants: (data.chat.participants || []).map((p: any) => p.id || p)
        };
      } else if (data && data.name && data.participants) {

        chatData = {
          type: "group",
          name: data.name,
          participants: data.participants
        };
      } else {
        throw new Error("Invalid group chat data format");
      }

      logger.debug("Creating group chat with data:", chatData);
      const response = await createChat(chatData);

      if (response && (response.chat_id || response.chat)) {
        const newChatId = response.chat_id || response.chat?.id;
        showCreateGroupModal = false;

        logger.debug("Group chat created successfully, refreshing chat list");
        await fetchChats();

        if (newChatId) {
          const newChat = chats.find(c => c.id === newChatId);
          if (newChat) {
            selectChat(newChat);
          }
        }

        toastStore.showToast({
          message: `Group chat "${chatData.name}" created successfully`,
          type: "success"
        });
      }
    } catch (error) {
      logError("Failed to create group chat", error);
      toastStore.showToast({
        message: "Failed to create group chat. Please try again.",
        type: "error"
      });
    } finally {
      isLoadingChats = false;
    }
  }

  const handleReconnect = () => {
    console.log("[WebSocket] Attempting to reconnect...");

    if (selectedChat) {
      console.log(`[WebSocket] Reconnecting to selected chat: ${selectedChat.id}`);
      websocketStore.connect(selectedChat.id);

      const recentChats = chats.slice(0, 5); 
      recentChats.forEach(chat => {
        if (selectedChat && chat.id !== selectedChat.id) {
          console.log(`[WebSocket] Reconnecting to additional chat: ${chat.id}`);
          websocketStore.connect(chat.id);
        }
      });
    } else {

      const recentChats = chats.slice(0, 5);
      recentChats.forEach(chat => {
        console.log(`[WebSocket] Reconnecting to chat: ${chat.id}`);
        websocketStore.connect(chat.id);
      });
    }
  };

  const testWebSocketConnection = () => {
    if (selectedChat) {
      console.log(`[WebSocket] Testing connection for chat: ${selectedChat.id}`);

      const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
      const hostname = window.location.hostname;
      const port = "8083";
      const token = getAuthToken();

      if (!token) {
        console.error("[WebSocket] No authentication token available");
        return;
      }

      const url1 = `${protocol}
      const url2 = `${protocol}

      console.log(`[WebSocket] Testing URL 1: ${url1}`);
      const ws1 = new WebSocket(url1);

      ws1.onopen = () => {
        console.log("[WebSocket] URL 1 connection successful!");
        ws1.send(JSON.stringify({type: "test", content: "Test message from URL 1"}));
      };

      ws1.onerror = (error) => {
        console.error("[WebSocket] URL 1 connection failed:", error);

        console.log(`[WebSocket] Testing URL 2: ${url2}`);
        const ws2 = new WebSocket(url2);

        ws2.onopen = () => {
          console.log("[WebSocket] URL 2 connection successful!");
          ws2.send(JSON.stringify({type: "test", content: "Test message from URL 2"}));
        };

        ws2.onerror = (error) => {
          console.error("[WebSocket] URL 2 connection failed:", error);
          console.error("[WebSocket] Both connection attempts failed");
        };
      };
    }
  };

  async function fetchMessages(chatId: string) {
    if (!chatId) return;

    isLoadingMessages = true;

    try {
      const response = await chatApi.listMessages(chatId);

      if (response && response.messages) {

        if (selectedChat && selectedChat.id === chatId) {
          selectedChat = {
            ...selectedChat,
            messages: response.messages
          };
        }
      }
    } catch (error) {
      logError("Failed to load messages", error);
    } finally {
      isLoadingMessages = false;
    }
  }

  async function markChatAsRead(chatId: string) {
    if (!chatId) return;

    try {

      const chatIndex = chats.findIndex(c => c.id === chatId);
      if (chatIndex >= 0) {

        chats[chatIndex].unread_count = 0;

        const filteredIndex = filteredChats.findIndex(c => c.id === chatId);
        if (filteredIndex >= 0) {
          filteredChats[filteredIndex].unread_count = 0;
        }

        chats = [...chats];
        filteredChats = [...filteredChats];
      }

    } catch (error) {
      logError("Failed to mark chat as read", error);
    }
  }

  function ensureValidChatArrays() {
    if (!Array.isArray(chats)) {
      chats = [];
    }

    if (!Array.isArray(filteredChats)) {
      filteredChats = [];
    }
  }

  async function confirmDeleteChat() {
    if (!chatToDelete) return;

    try {
      await deleteChat(chatToDelete.id);

      chats = chats.filter(c => c.id !== chatToDelete?.id);
      filteredChats = filteredChats.filter(c => c.id !== chatToDelete?.id);

      if (selectedChat && selectedChat.id === chatToDelete.id) {
        selectedChat = null;
      }

      if (!selectedChat && chats.length > 0) {
        selectChat(chats[0]);
      }

      toastStore.showToast({
        message: "Conversation deleted successfully",
        type: "success"
      });
    } catch (error) {
      logger.error("Failed to delete conversation", error);
      toastStore.showToast({
        message: "Failed to delete conversation. Please try again.",
        type: "error"
      });
    } finally {

      chatToDelete = null;
      showDeleteConfirm = false;
    }
  }

  function cancelDeleteChat() {
    chatToDelete = null;
    showDeleteConfirm = false;
  }

  function requestDeleteChat(chatToRemove: Chat) {
    chatToDelete = chatToRemove;
    showDeleteConfirm = true;
    logger.debug(`Preparing to delete chat: ${chatToRemove.id}`);
  }

  function showToast(options: { message: string; type: ToastTypeEnum }) {
    toastStore.showToast(options);
  }
</script>

<style>
  .message-container {
    display: flex;
    height: 100vh;
    width: 100%;
    background-color: white;
  }

  .middle-section {
    width: 35%;
    min-width: 320px;
    border-right: 1px solid #e0e0e0;
    overflow: hidden;
    display: flex;
    flex-direction: column;
  }

  .message-container.dark-theme {
    background-color: #1a1a1a;
    color: #f0f0f0;
  }

  .dark-theme .middle-section,
  .dark-theme .right-section {
    background-color: #1a1a1a;
    border-color: #333;
  }

  .dark-theme input,
  .dark-theme textarea {
    background-color: #333;
    color: #fff;
    border-color: #444;
  }

  .dark-theme .msg-search-input {
    background-color: #333;
    color: #fff;
  }

  .dark-theme .msg-chat-item {
    border-color: #333;
  }

  .dark-theme .msg-chat-item:hover {
    background-color: #2a2a2a;
  }

  .dark-theme .msg-chat-item.selected {
    background-color: #2a2a2a;
  }

  .chat-list-header {
    padding: 16px;
    display: flex;
    justify-content: space-between;
    align-items: center;
    border-bottom: 1px solid #e0e0e0;
  }

  .msg-search-container {
    padding: 12px 16px;
    position: relative;
  }

  .msg-search-input {
    width: 100%;
    padding: 8px 32px 8px 12px;
    border: 1px solid #e0e0e0;
    border-radius: 20px;
    font-size: 14px;
  }

  .msg-clear-search {
    position: absolute;
    right: 24px;
    top: 50%;
    transform: translateY(-50%);
    background: none;
    border: none;
    cursor: pointer;
    color: #888;
  }

  .chat-list {
    flex: 1;
    overflow-y: auto;
    padding: 0;
  }

  .msg-chat-item {
    display: flex;
    align-items: center;
    padding: 12px 16px;
    border-bottom: 1px solid #f0f0f0;
    cursor: pointer;
    position: relative;
  }

  .msg-chat-item:hover {
    background-color: #f9f9f9;
  }

  .msg-chat-item.selected {
    background-color: #f0f0f0;
  }

  .msg-avatar {
    width: 48px;
    height: 48px;
    border-radius: 50%;
    margin-right: 12px;
    overflow: hidden;
    flex-shrink: 0;
  }

  .msg-avatar img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  .avatar-placeholder {
    width: 100%;
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
    color: white;
    font-weight: bold;
    font-size: 20px;
  }

  .chat-details {
    flex: 1;
    min-width: 0;
  }

  .chat-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 4px;
  }

  .chat-name {
    font-weight: 600;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .timestamp {
    color: #888;
    font-size: 12px;
    white-space: nowrap;
  }

  .msg-last-message {
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    color: #666;
    font-size: 14px;
  }

  .msg-no-messages {
    font-style: italic;
    color: #888;
  }

  .msg-unread-badge {
    background-color: #0066ff;
    color: white;
    border-radius: 50%;
    min-width: 20px;
    height: 20px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 12px;
    font-weight: bold;
    padding: 0 4px;
  }

  .right-section {
    flex: 1;
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  .msg-chat-header {
    padding: 16px;
    display: flex;
    align-items: center;
    border-bottom: 1px solid #e0e0e0;
    background-color: white;
  }

  .dark-theme .msg-chat-header {
    background-color: #1a1a1a;
    border-color: #333;
  }

  .msg-chat-header-info {
    flex: 1;
    margin-left: 12px;
  }

  .msg-chat-header-name {
    font-weight: 600;
    font-size: 16px;
  }

  .msg-chat-header-status {
    color: #888;
    font-size: 13px;
  }

  .msg-chat-header-actions {
    display: flex;
    gap: 8px;
  }

  .msg-action-icon {
    background: none;
    border: none;
    cursor: pointer;
    color: #555;
    padding: 8px;
    border-radius: 50%;
    transition: background-color 0.2s;
  }

  .msg-action-icon:hover {
    background-color: #f0f0f0;
  }

  .dark-theme .msg-action-icon:hover {
    background-color: #333;
  }

  .messages-container {
    flex: 1;
    overflow-y: auto;
    padding: 16px;
    background-color: #f7f7f7;
    display: flex;
    flex-direction: column;
  }

  .dark-theme .messages-container {
    background-color: #252525;
  }

  .loading-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 100%;
  }

  .loading-spinner {
    border: 4px solid rgba(0, 0, 0, 0.1);
    border-left-color: #0066ff;
    border-radius: 50%;
    width: 30px;
    height: 30px;
    animation: spin 1s linear infinite;
    margin-bottom: 16px;
  }

  .dark-theme .loading-spinner {
    border-color: rgba(255, 255, 255, 0.1);
    border-left-color: #0066ff;
  }

  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }

  .msg-conversation-item {
    display: flex;
    margin-bottom: 16px;
    max-width: 80%;
  }

  .msg-conversation-item.own-message {
    align-self: flex-end;
    flex-direction: row-reverse;
  }

  .message-bubble {
    position: relative;
    padding: 12px 16px;
    border-radius: 18px;
    box-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
    margin: 0 8px;
    max-width: 100%;
    word-break: break-word;
  }

  .message-bubble.sent {
    background-color: #3b82f6;
    color: white;
    border-bottom-right-radius: 4px;
  }

  .message-bubble.received {
    background-color: #e9e9e9;
    color: #333;
    border-bottom-left-radius: 4px;
  }

  .dark-theme .message-bubble.sent {
    background-color: #3b82f6;
    color: white;
  }

  .dark-theme .message-bubble.received {
    background-color: #2a2a2a;
    color: #f0f0f0;
  }

  .chat-actions {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .delete-chat-btn {
    display: none;
    background: none;
    border: none;
    cursor: pointer;
    padding: 4px;
    color: #888;
    border-radius: 50%;
    transition: all 0.2s ease;
  }

  .delete-chat-btn:hover {
    color: #ff4d4d;
    background-color: rgba(255, 77, 77, 0.1);
  }

  .msg-chat-item:hover .delete-chat-btn {
    display: flex;
  }

  .confirmation-dialog {
    position: fixed;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background-color: rgba(0, 0, 0, 0.5);
    display: flex;
    justify-content: center;
    align-items: center;
    z-index: 1000;
  }

  .dialog-content {
    background-color: white;
    border-radius: 8px;
    padding: 24px;
    max-width: 400px;
    width: 90%;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  }

  .dialog-content p {
    margin: 0 0 20px;
    font-size: 16px;
    text-align: center;
  }

  .dialog-actions {
    display: flex;
    justify-content: center;
    gap: 12px;
  }

  .confirm-button, .cancel-button {
    padding: 8px 16px;
    border-radius: 4px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .confirm-button {
    background-color: #ff4d4d;
    color: white;
    border: 1px solid #ff4d4d;
  }

  .confirm-button:hover {
    background-color: #ff3333;
  }

  .cancel-button {
    background-color: white;
    color: #333;
    border: 1px solid #ddd;
  }

  .cancel-button:hover {
    background-color: #f5f5f5;
  }

  .dark-theme .msg-clear-search {
    color: #bbb;
  }

  .image-button {
    padding: 0;
    margin: 0;
    border: none;
    background: transparent;
    cursor: pointer;
    display: block;
    width: 100%;
    border-radius: 8px;
    overflow: hidden;
  }

  .image-button:focus {
    outline: 2px solid #0066ff;
  }

  .image-button img {
    width: 100%;
    display: block;
  }

  .modal-backdrop {
    position: fixed;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background-color: rgba(0, 0, 0, 0.5);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
    padding: 20px;
    box-sizing: border-box;
  }

  .modal-container {
    background: white;
    border-radius: 12px;
    max-width: 600px;
    width: 100%;
    max-height: 90vh;
    overflow: hidden;
    box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04);
  }
</style>

<div class="custom-message-layout {isDarkMode ? "dark-theme" : ""}">
  <!-- Sidebar -->
  <div class="custom-sidebar {isMobile && !showMobileMenu ? "hidden" : ""}">
    <LeftSide
      {username}
      {displayName}
      {avatar}
      isCollapsed={false}
      isMobileMenu={isMobile && showMobileMenu}
      on:closeMobileMenu={() => showMobileMenu = false}
    />
  </div>

  <!-- Mobile menu overlay -->
  {#if isMobile && showMobileMenu}
    <div class="mobile-overlay"
         on:click={toggleMobileMenu}
         on:keydown={(e) => e.key === "Enter" && toggleMobileMenu()}
         role="button"
         tabindex="0"
         aria-label="Close mobile menu"></div>
  {/if}

  <!-- Main content area -->
  <div class="custom-content-area">
    <!-- Mobile header -->
    {#if isMobile}
      <div class="mobile-header">
        <button class="mobile-menu-button" on:click={toggleMobileMenu} aria-label="Toggle mobile menu">
          <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <line x1="3" y1="12" x2="21" y2="12"></line>
            <line x1="3" y1="6" x2="21" y2="6"></line>
            <line x1="3" y1="18" x2="21" y2="18"></line>
          </svg>
        </button>
        <h1 class="mobile-title">Messages</h1>
        <ThemeToggle size="sm" />
      </div>
    {/if}

    <!-- WebSocket Status Display -->
    <div class="connection-status-container">
      <div class="connection-status">
        {#if selectedChat}
          {#if $websocketStore.connectionStatus[selectedChat.id] === "connected"}
            <div class="status-connected">
              <span class="status-icon">●</span>
              <span class="status-text">Connected</span>
            </div>
          {:else if $websocketStore.connectionStatus[selectedChat.id] === "connecting"}
            <div class="status-connecting">
              <span class="status-icon">◌</span>
              <span class="status-text">Connecting...</span>
            </div>
          {:else if $websocketStore.connectionStatus[selectedChat.id] === "disconnected" || $websocketStore.connectionStatus[selectedChat.id] === "error"}
            <div class="status-disconnected">
              <span class="status-icon">○</span>
              <span class="status-text">Disconnected</span>
              <button class="reconnect-button" on:click={handleReconnect}>
                Reconnect Now
              </button>
              <button class="test-connection-button" on:click={testWebSocketConnection}>
                Test Connection
              </button>
            </div>
          {/if}
        {/if}
      </div>
      {#if $websocketStore.lastError}
        <div class="status-error">
          <span class="error-icon">⚠️</span>
          <span class="error-text">{$websocketStore.lastError}</span>
          <button class="error-dismiss" on:click={() => (websocketStore as any).resetError()}>×</button>
        </div>
      {/if}
    </div>

    <div class="message-container {isDarkMode ? "dark-theme" : ""}">
      <!-- Middle section - Chat list -->
      <div class="middle-section {selectedChat && isMobile ? "hidden" : ""}">
        <!-- Chat header -->
        <div class="chat-list-header">
          <h2 class="page-title">Messages</h2>
          <div class="header-actions">
            <button class="msg-new-message-button" on:click={() => showCreateGroupModal = true}>
              <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
                <circle cx="9" cy="7" r="4"></circle>
                <path d="M23 21v-2a4 4 0 0 0-3-3.87"></path>
                <path d="M16 3.13a4 4 0 0 1 0 7.75"></path>
              </svg>
              <span>New Group</span>
            </button>
            <button class="msg-new-message-button" on:click={() => showNewChatModal = true}>
              <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <line x1="12" y1="5" x2="12" y2="19"></line>
                <line x1="5" y1="12" x2="19" y2="12"></line>
          </svg>
              <span>New</span>
        </button>
      </div>
    </div>

        <!-- Search container -->
        <div class="msg-search-container">
        <input
          type="text"
            placeholder="Search messages..."
          bind:value={searchQuery}
            on:input={handleSearch}
            class="msg-search-input"
        />
          {#if searchQuery}
            <button class="msg-clear-search" on:click={() => { searchQuery = ""; handleSearch(); }} aria-label="Clear search">
              <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" width="18" height="18">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        {/if}
      </div>

         <!-- Chat list -->
      <div class="chat-list">
        {#if isLoadingChats}
          <div class="loading-container">
            <div class="loading-spinner"></div>
            <p>Loading chats...</p>
          </div>
        {:else if chats.length === 0}
            <div class="msg-empty-state">
              <div class="msg-empty-state-icon">
                <svg xmlns="http://www.w3.org/2000/svg" width="70" height="70" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1" stroke-linecap="round" stroke-linejoin="round">
                  <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"></path>
                </svg>
              </div>
              <h2>No conversations yet</h2>
              <p>Start a new conversation with friends</p>
              <button class="msg-new-message-button" on:click={() => showNewChatModal = true}>
              Start a new chat
            </button>
          </div>
        {:else}
            {#each filteredChats as chat (chat.id)}
            <div
                class="msg-chat-item {selectedChat?.id === chat.id ? "selected" : ""}"
              on:click={() => selectChat(chat)}
              on:keydown={(e) => e.key === "Enter" && selectChat(chat)}
              role="button"
              tabindex="0"
              >
                {#if chat.type === "individual"}
                  <div class="msg-avatar">
                    {#if getOtherParticipant(chat)?.avatar}
                      <img src={getOtherParticipant(chat)?.avatar || ""} alt={getOtherParticipant(chat)?.display_name || ""} />
                      {:else}
                    <div class="avatar-placeholder" style="background-color: {getAvatarColor(getChatDisplayName(chat))}">
                      {getChatDisplayName(chat).charAt(0).toUpperCase()}
                    </div>
                      {/if}
                    </div>
                {:else}
                  <div class="msg-avatar">
                    {#if chat.avatar}
                      <img src={chat.avatar} alt={chat.name} />
                    {:else}
                      <div class="avatar-placeholder" style="background-color: {getAvatarColor(chat.name)}">
                        {chat.name.charAt(0).toUpperCase()}
                      </div>
                    {/if}
                  </div>
                {/if}

                <div class="chat-details">
                      <div class="chat-header">
                    <div class="chat-name">
                      {getChatDisplayName(chat)}
                      </div>
                    {#if chat.last_message?.timestamp}
                      <div class="timestamp">{safeFormatRelativeTime(chat.last_message.timestamp)}</div>
                          {/if}
                  </div>
                  <div class="msg-last-message">
                    {#if chat.last_message}
                      <span>{chat.last_message.content}</span>
                        {:else}
                      <span class="msg-no-messages">No messages yet</span>
                        {/if}
                      </div>
                    </div>

                    <div class="chat-actions">
                    {#if chat.unread_count > 0}
                  <div class="msg-unread-badge">{chat.unread_count}</div>
                    {/if}
                      <!-- Delete button -->
                      <button
                        class="delete-chat-btn"
                        on:click|stopPropagation={() => requestDeleteChat(chat)}
                        aria-label="Delete conversation"
                      >
                        <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                          <path d="M3 6h18"></path>
                          <path d="M19 6v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6"></path>
                          <path d="M8 6V4c0-1 1-2 2-2h4c1 0 2 1 2 2v2"></path>
                        </svg>
                      </button>
                    </div>
          </div>
            {/each}
      {/if}
    </div>
  </div>

      <!-- Right section - Chat content -->
      <div class="right-section {selectedChat && isMobile ? "full-width" : ""}">
    {#if selectedChat}
          <!-- Chat header -->
          <div class="msg-chat-header">
            {#if selectedChat.type === "individual"}
              <div class="msg-avatar">
                {#if getOtherParticipant(selectedChat)?.avatar}
                  <img src={getOtherParticipant(selectedChat)?.avatar || ""} alt={getOtherParticipant(selectedChat)?.display_name || ""} />
          {:else}
                <div class="avatar-placeholder" style="background-color: {getAvatarColor(getChatDisplayName(selectedChat))}">
                  {getChatDisplayName(selectedChat).charAt(0).toUpperCase()}
                </div>
          {/if}
        </div>
            {:else}
              <div class="msg-avatar">
                {#if selectedChat.avatar}
                  <img src={selectedChat.avatar} alt={selectedChat.name} />
                {:else}
                  <div class="avatar-placeholder" style="background-color: {getAvatarColor(selectedChat.name)}">
                    {selectedChat.name.charAt(0).toUpperCase()}
                  </div>
                {/if}
              </div>
            {/if}

            <div class="msg-chat-header-info">
              <div class="msg-chat-header-name">{getChatDisplayName(selectedChat)}</div>
              {#if selectedChat.type === "individual" && getOtherParticipant(selectedChat)}
                <div class="msg-chat-header-status">
                  {getOtherParticipant(selectedChat)?.is_verified ? "✓ Verified" : "Online"}
                </div>
              {:else}
                <div class="msg-chat-header-status">
                  {selectedChat.participants.length} members
                </div>
              {/if}
        </div>

            <div class="msg-chat-header-actions">
              {#if selectedChat.type === "group"}
                <button
                  class="msg-action-icon"
                  on:click={() => {
                    logger.debug("Opening group management modal for chat:", selectedChat);
                    showManageGroupModal = true;
                  }}
                  aria-label="Manage group members"
                  title="Manage Members"
                >
                  <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" width="20" height="20">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197m13.5-9a2.5 2.5 0 11-5 0 2.5 2.5 0 015 0z" />
                  </svg>
                </button>
              {/if}
              <button class="msg-action-icon" on:click={() => {}} aria-label="More options">
                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" width="20" height="20">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 5v.01M12 12v.01M12 19v.01M12 6a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2z" />
                </svg>
              </button>
            </div>
      </div>

      <div class="messages-container">
            {#if isLoadingMessages}
              <div class="loading-container">
                <div class="loading-spinner"></div>
                <p>Loading messages...</p>
              </div>
            {:else if selectedChat.messages && selectedChat.messages.length > 0}
        {#each selectedChat.messages as message}
                <div class="msg-conversation-item {message.sender_id === $authStore.user_id ? "own-message" : ""} {message.is_deleted ? "deleted" : ""} {message.failed ? "failed" : ""}">
              {#if message.sender_id !== $authStore.user_id}
                    <div class="msg-avatar">
                      {#if message.sender_avatar}
                        <img src={message.sender_avatar} alt={message.sender_name} />
                      {:else}
                        <div class="avatar-placeholder" style="background-color: {getAvatarColor(message.sender_name || "User")}">
                          {(message.sender_name || "User").charAt(0).toUpperCase()}
                        </div>
              {/if}
              </div>
                  {/if}

                  <div class="message-bubble {message.sender_id === $authStore.user_id ? "sent" : "received"}" class:failed={message.failed} class:is-local={message.is_local}>
                    <!-- Message content -->
                    <div class="message-content">
                    {#if message.is_deleted}
                        <span class="deleted-message">Message deleted</span>
                    {:else}
                        <!-- Display the content text -->
                        {#if message.content && (!message.content.startsWith("{") || !message.content.includes("attachment"))}
                        <p>{message.content}</p>
                        {/if}

                        <!-- Display attachments if any -->
                        {#if message.attachments && message.attachments.length > 0}
                          <div class="attachments-container">
                            {#each message.attachments as attachment}
                              {#if attachment.type === "image"}
                                <div class="image-attachment">
                                  <button class="image-button" on:click={() => window.open(attachment.url, "_blank")} aria-label="View image">
                                    <img src={attachment.url} alt="Attachment" />
                                  </button>
                                </div>
                              {:else if attachment.type === "gif"}
                                <div class="gif-attachment">
                                  <button class="image-button" on:click={() => window.open(attachment.url, "_blank")} aria-label="View GIF">
                                    <img src={attachment.url} alt="GIF" />
                                  </button>
                                </div>
                              {:else if attachment.type === "video"}
                                <div class="video-attachment">
                                  <video controls>
                                    <source src={attachment.url} type="video/mp4">
                                    <track kind="captions" src="" label="English" default>
                                    Your browser does not support the video tag.
                                  </video>
                                </div>
                              {/if}
                            {/each}
                          </div>
                        {/if}

                        <!-- Try to parse message content for attachment info -->
                        {#if !message.attachments?.length && message.content && message.content.startsWith("{")}
                          <!-- Use a helper function approach to parse JSON safely -->
                          {@const parsedContent = (() => {
                            try {
                              return JSON.parse(message.content);
                            } catch (e) {
                              return null;
                            }
                          })()}

                          {#if parsedContent && parsedContent.attachment}
                            <div class="attachments-container">
                              {#if parsedContent.attachment.type === "image"}
                                <div class="image-attachment">
                                  <button class="image-button" on:click={() => window.open(parsedContent.attachment.url, "_blank")} aria-label="View image">
                                    <img src={parsedContent.attachment.url} alt="Attachment" />
                                  </button>
                                </div>
                              {:else if parsedContent.attachment.type === "gif"}
                                <div class="gif-attachment">
                                  <button class="image-button" on:click={() => window.open(parsedContent.attachment.url, "_blank")} aria-label="View GIF">
                                    <img src={parsedContent.attachment.url} alt="GIF" />
                                  </button>
                                </div>
                              {:else if parsedContent.attachment.type === "video"}
                                <div class="video-attachment">
                                  <video controls>
                                    <source src={parsedContent.attachment.url} type="video/mp4">
                                    <track kind="captions" src="" label="English" default>
                                    Your browser does not support the video tag.
                                  </video>
                                </div>
                              {/if}
                            </div>
                            {#if parsedContent.text}
                              <p>{parsedContent.text}</p>
                            {/if}
                          {/if}
                        {/if}

                        <!-- Show retry option for local/failed messages -->
                        {#if message.failed || message.is_local}
                          <div class="message-error">
                            <span class="error-text">Not sent to server</span>
                            <button class="retry-btn" on:click={() => {

                              newMessage = message.content;

                              if (selectedChat?.messages) {
                                selectedChat = {
                                  ...selectedChat,
                                  messages: selectedChat.messages.filter(msg => msg.id !== message.id)
                                };
                              }
                            }}>
                              <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                                <path d="M21 12a9 9 0 0 0-9-9 9.75 9.75 0 0 0-6.74 2.74L3 8"></path>
                                <path d="M3 3v5h5"></path>
                                <path d="M3 12a9 9 0 0 0 9 9 9.75 9.75 0 0 0 6.74-2.74L21 16"></path>
                                <path d="M16 21h5v-5"></path>
                              </svg>
                              Retry
                            </button>
                </div>
              {/if}
                      {/if}
                    </div>

                    <!-- Message footer with timestamp -->
                      <div class="message-footer">
                      <span class="timestamp" data-timestamp={message.timestamp}>{safeFormatRelativeTime(message.timestamp)}</span>

                      <!-- Message actions for sent messages -->
                      {#if !message.is_deleted && message.sender_id === $authStore.user_id && !message.is_local}
                          <div class="message-actions">
                          <button class="action-btn" on:click={() => unsendMessage(message.id)}>
                            <span class="material-icons">delete</span>
                  </button>
                          </div>
                {/if}
              </div>
            </div>
          </div>
        {/each}
            {:else}
              <div class="msg-empty-state">
                <div class="msg-empty-state-icon">
                  <svg xmlns="http://www.w3.org/2000/svg" width="70" height="70" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1" stroke-linecap="round" stroke-linejoin="round">
                    <circle cx="12" cy="12" r="10"></circle>
                    <line x1="8" y1="12" x2="16" y2="12"></line>
                  </svg>
                </div>
                <h2>No messages yet</h2>
                <p>Start the conversation!</p>
              </div>
            {/if}
      </div>

          <div class="msg-message-input-container">
            <div class="msg-input-wrapper">
              <textarea
                bind:value={newMessage}
                placeholder="Type a message..."
                rows="1"
                on:keydown={(e) => e.key === "Enter" && !e.shiftKey && sendMessage(newMessage)}
              ></textarea>

              <div class="msg-attachment-buttons">
                <button class="msg-attachment-button" on:click={() => handleAttachment("image")} aria-label="Add image">
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" width="20" height="20">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
            </svg>
        </button>
                <button class="msg-attachment-button" on:click={() => handleAttachment("gif")} aria-label="Add GIF">
                  <span class="msg-gif-button">GIF</span>
                </button>
              </div>
            </div>

        <button
              class="msg-send-button {newMessage.trim() ? "active" : ""}"
          disabled={!newMessage.trim()}
              on:click={() => sendMessage(newMessage)}
              aria-label="Send message"
        >
              <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" width="20" height="20">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8" />
          </svg>
        </button>
      </div>
    {:else}
          <div class="msg-empty-state">
            <div class="msg-empty-state-icon">
              <svg xmlns="http://www.w3.org/2000/svg" width="70" height="70" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1" stroke-linecap="round" stroke-linejoin="round">
                <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"></path>
        </svg>
            </div>
            <h2>Your Messages</h2>
            <p>Send private messages to a friend or group</p>
            <button class="msg-new-message-button" on:click={() => showNewChatModal = true}>
              New Message
            </button>
      </div>
    {/if}
      </div>
  </div>
</div>
    </div>

<!-- Modals -->
{#if showNewChatModal}
  <NewChatModal
    {isLoadingUsers}
    {userSearchResults}
    searchKeyword={userSearchQuery}
    onCancel={() => showNewChatModal = false}
    on:close={() => showNewChatModal = false}
    on:search={(e) => searchForUsers(e.detail)}
    on:createChat={(e) => initiateNewChat(e.detail)}
  />
{/if}

{#if showCreateGroupModal}
  <!-- Modal backdrop -->
  <div
    class="modal-backdrop"
    role="dialog"
    aria-modal="true"
    tabindex="-1"
    on:click={() => showCreateGroupModal = false}
    on:keydown={(e) => e.key === "Escape" && (showCreateGroupModal = false)}
  >
    <div
      class="modal-container"
      role="document"
    >
      <CreateGroupChat
        onCancel={() => showCreateGroupModal = false}
        onSuccess={(e) => handleCreateGroupChat(e.detail)}
        on:close={() => showCreateGroupModal = false}
        on:createGroup={(e) => handleCreateGroupChat(e.detail)}
      />
    </div>
  </div>
{/if}

{#if showManageGroupModal && selectedChat && selectedChat.type === "group"}
  <!-- Modal backdrop -->
  <div
    class="modal-backdrop"
    role="dialog"
    aria-modal="true"
    tabindex="-1"
    on:click={() => showManageGroupModal = false}
    on:keydown={(e) => e.key === "Escape" && (showManageGroupModal = false)}
  >
    <div
      class="modal-container"
      role="document"
    >
      <ManageGroupMembers
        chatId={selectedChat.id}
        currentChatParticipants={selectedChat.participants?.map(p => ({
          id: p.id,
          username: p.username || "",
          name: p.display_name || p.username || "",
          display_name: p.display_name || p.username || "",
          profile_picture_url: p.avatar || null,
          is_verified: p.is_verified || false,
          avatar: p.avatar || null
        })) || []}
        onClose={() => showManageGroupModal = false}
        onMembersUpdated={() => {
          logger.debug("Members updated, refreshing chats");
          fetchChats();
        }}
      />
    </div>
  </div>
{/if}

{#if showDebug}
  <DebugPanel
    on:close={() => showDebug = false}
    on:testConnection={() => testApiConnection()}
    on:checkAuth={() => logAuthTokenInfo()}
  />
{/if}

<Toast on:close={(e) => toastStore.removeToast(e.detail)} />

{#if showDeleteConfirm}
  <div class="confirmation-dialog">
    <div class="dialog-content">
      <p>Are you sure you want to delete this conversation?</p>
      <div class="dialog-actions">
        <button class="confirm-button" on:click={confirmDeleteChat}>Yes</button>
        <button class="cancel-button" on:click={cancelDeleteChat}>No</button>
      </div>
    </div>
  </div>
{/if}