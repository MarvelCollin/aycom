<script lang="ts">
  import { onMount } from 'svelte';
  import MainLayout from '../components/layout/MainLayout.svelte';
  import { useAuth } from '../hooks/useAuth';
  import { useTheme } from '../hooks/useTheme';
  import type { IAuthStore } from '../interfaces/IAuth';
  import { createLoggerWithPrefix } from '../utils/logger';
  import { toastStore } from '../stores/toastStore';
  import { checkAuth, isWithinTime, formatTimeAgo, handleApiError } from '../utils/common';
  import { listChats, listMessages, sendMessage as apiSendMessage, unsendMessage as apiUnsendMessage, searchMessages } from '../api/chat';
  import { getProfile } from '../api/user';
  
  const logger = createLoggerWithPrefix('Message');
  
  // Auth and theme
  const { getAuthState } = useAuth();
  const { theme } = useTheme();
  
  // Reactive declarations
  $: authState = getAuthState ? (getAuthState() as IAuthStore) : { userId: null, isAuthenticated: false, accessToken: null, refreshToken: null };
  $: isDarkMode = $theme === 'dark';
  
  // User profile data
  let username = '';
  let displayName = '';
  let avatar = 'https://secure.gravatar.com/avatar/0?d=mp'; // Default avatar with proper image URL
  let isLoadingProfile = true;

  // Chat state
  let selectedChat: Chat | null = null;
  let chats: Chat[] = [];
  let isLoadingChats = true;
  let newMessage = '';
  let searchQuery = '';
  let filteredChats: Chat[] = [];
  let selectedAttachments: Attachment[] = [];

  // Message interfaces
  interface Message {
    id: string;
    senderId: string;
    senderName: string;
    senderAvatar: string | null;
    content: string;
    timestamp: string;
    isDeleted: boolean;
    attachments: Attachment[];
    isOwn: boolean;
  }

  interface Attachment {
    id: string;
    type: 'image' | 'gif' | 'video';
    url: string;
    thumbnail?: string;
  }

  interface Chat {
    id: string;
    type: 'individual' | 'group';
    name: string;
    avatar: string | null;
    participants: Participant[];
    lastMessage?: {
      content: string;
      timestamp: string;
      senderId: string;
    };
    messages: Message[];
    unreadCount: number;
  }

  interface Participant {
    id: string;
    username: string;
    displayName: string;
    avatar: string | null;
  }
  
  // Fetch user profile data using the API directly
  async function fetchUserProfile() {
    isLoadingProfile = true;
    try {
      const response = await getProfile();
      if (response && response.user) {
        const userData = response.user;
        username = userData.username || `user_${authState.userId?.substring(0, 4)}`;
        displayName = userData.name || userData.display_name || `User ${authState.userId?.substring(0, 4)}`;
        avatar = userData.profile_picture_url || 'https://secure.gravatar.com/avatar/0?d=mp';
        logger.debug('Profile loaded', { username });
      } else {
        logger.warn('No user data received from API');
        username = `user_${authState.userId?.substring(0, 4)}`;
        displayName = `User ${authState.userId?.substring(0, 4)}`;
      }
    } catch (error) {
      const errorResponse = handleApiError(error);
      logger.error('Error fetching user profile:', errorResponse);
      username = `user_${authState.userId?.substring(0, 4)}`;
      displayName = `User ${authState.userId?.substring(0, 4)}`;
    } finally {
      isLoadingProfile = false;
    }
  }

  // Authentication check on component load
  onMount(() => {
    if (!checkAuth(authState, 'messages')) {
      return;
    }
    
    // Fetch user profile first, then chats
    fetchUserProfile().then(() => {
      fetchChats();
    });
  });

  // Fetch chats
  async function fetchChats() {
    isLoadingChats = true;
    
    try {
      const response = await listChats();
      
      if (response && response.chats) {
        chats = response.chats.map((chat: any) => ({
          id: chat.id,
          type: chat.type || 'individual',
          name: chat.name || getParticipantName(chat),
          avatar: chat.avatar || null,
          participants: chat.participants || [],
          lastMessage: chat.last_message ? {
            content: chat.last_message.content,
            timestamp: chat.last_message.timestamp,
            senderId: chat.last_message.sender_id
          } : undefined,
          messages: [],
          unreadCount: chat.unread_count || 0
        }));
        
        filteredChats = [...chats];
        logger.debug('Chats loaded', { count: chats.length });
      } else {
        logger.warn('No chats found or invalid response format');
        chats = [];
        filteredChats = [];
      }
    } catch (error) {
      const errorResponse = handleApiError(error);
      logger.error('Error fetching chats:', errorResponse);
      toastStore.showToast('Failed to load chats. Please try again.', 'error');
      chats = [];
      filteredChats = [];
    } finally {
      isLoadingChats = false;
    }
  }
  
  // Helper function to get a chat name from participants
  function getParticipantName(chat: any): string {
    if (chat.participants && chat.participants.length > 0) {
      const participant = chat.participants[0];
      return participant.display_name || participant.username || 'Chat';
    }
    return 'Chat';
  }

  // Unsend a message using the API
  async function unsendMessage(messageId: string) {
    if (!selectedChat) return;

    try {
      await apiUnsendMessage(selectedChat.id, messageId);
      
      // Update local state after successful API call
      selectedChat.messages = selectedChat.messages.map(m => 
        m.id === messageId ? { ...m, isDeleted: true } : m
      );

      // If this was the last message, update the lastMessage property
      if (selectedChat.lastMessage && selectedChat.messages.find(m => m.id === messageId)) {
        const lastNonDeletedMessage = [...selectedChat.messages]
          .reverse()
          .find(m => !m.isDeleted);
        
        if (lastNonDeletedMessage) {
          selectedChat.lastMessage = {
            content: lastNonDeletedMessage.content,
            timestamp: lastNonDeletedMessage.timestamp,
            senderId: lastNonDeletedMessage.senderId
          };
        } else {
          selectedChat.lastMessage = undefined;
        }
      }
      
      logger.debug('Message unsent', { messageId });
      toastStore.showToast('Message unsent', 'success');
    } catch (error) {
      const errorResponse = handleApiError(error);
      logger.error('Failed to unsend message:', errorResponse);
      toastStore.showToast('Failed to unsend message. Please try again.', 'error');
    }
  }

  // Select a chat and load messages
  async function selectChat(chat: Chat) {
    selectedChat = chat;
    
    try {
      const response = await listMessages(chat.id);
      
      if (response && response.messages) {
        // Update chat with fetched messages
        selectedChat.messages = response.messages.map((msg: any) => ({
          id: msg.id,
          senderId: msg.sender_id,
          senderName: msg.sender_name,
          senderAvatar: msg.sender_avatar,
          content: msg.content,
          timestamp: msg.timestamp,
          isDeleted: msg.is_deleted || false,
          attachments: msg.attachments || [],
          isOwn: msg.sender_id === authState.userId
        }));
      }
      
      // Mark messages as read
      selectedChat.unreadCount = 0;
      logger.debug('Chat selected and messages loaded', { chatId: chat.id, messageCount: selectedChat.messages.length });
    } catch (error) {
      const errorResponse = handleApiError(error);
      logger.error('Error loading messages:', errorResponse);
      toastStore.showToast('Failed to load messages. Please try again.', 'error');
    }
  }

  // Send a message using the API
  async function sendMessage() {
    if (!selectedChat || (!newMessage.trim() && selectedAttachments.length === 0)) return;
    
    try {
      const messageData = {
        content: newMessage.trim(),
        attachments: selectedAttachments.map(attachment => ({
          type: attachment.type,
          url: attachment.url
        }))
      };
      
      const response = await apiSendMessage(selectedChat.id, messageData);
      
      if (response && response.message) {
        // Add the new message to the chat
        const newMsg: Message = {
          id: response.message.id,
          senderId: response.message.sender_id,
          senderName: response.message.sender_name || displayName,
          senderAvatar: response.message.sender_avatar || avatar,
          content: response.message.content,
          timestamp: response.message.timestamp,
          isDeleted: false,
          attachments: response.message.attachments || [],
          isOwn: true
        };
        
        selectedChat.messages = [...selectedChat.messages, newMsg];
        selectedChat.lastMessage = {
          content: newMsg.content || 'Sent an attachment',
          timestamp: newMsg.timestamp,
          senderId: newMsg.senderId
        };
        
        newMessage = '';
        selectedAttachments = [];
        logger.debug('Message sent', { messageId: newMsg.id });
      }
    } catch (error) {
      const errorResponse = handleApiError(error);
      logger.error('Failed to send message:', errorResponse);
      toastStore.showToast('Failed to send message. Please try again.', 'error');
    }
  }

  // Search chats and messages
  async function searchChats() {
    if (!searchQuery.trim()) {
      filteredChats = [...chats];
      return;
    }

    // First, filter local chats by name
    let results = chats.filter(chat => 
      chat.name.toLowerCase().includes(searchQuery.toLowerCase())
    );
    
    // If we have a selected chat, also search messages
    if (selectedChat) {
      try {
        const response = await searchMessages(selectedChat.id, searchQuery);
        
        if (response && response.messages && response.messages.length > 0) {
          // If the current chat isn't in results but has matching messages, add it
          if (!results.some(chat => chat.id === selectedChat!.id)) {
            results.push(selectedChat);
          }
        }
      } catch (error) {
        logger.error('Error searching messages:', error);
      }
    }
    
    filteredChats = results;
    logger.debug('Chats searched', { query: searchQuery, resultCount: filteredChats.length });
  }

  // Clear search
  function clearSearch() {
    searchQuery = '';
    filteredChats = [...chats];
    logger.debug('Search cleared');
  }

  // Handle file attachments
  function handleAttachment(type: 'image' | 'gif' | 'video') {
    // For now, we'll keep this simplified without file picker integration
    // In a real implementation, you would use a file input element
    
    const dummyUrl = type === 'image' 
      ? 'https://via.placeholder.com/300' 
      : type === 'gif' 
        ? 'https://media.giphy.com/media/3o7TKSjRrfIPjeiVyg/giphy.gif'
        : 'https://sample-videos.com/video123/mp4/240/big_buck_bunny_240p_1mb.mp4';
        
    const attachment: Attachment = {
      id: `temp-${Date.now()}`,
      type,
      url: dummyUrl,
    };
    
    selectedAttachments = [...selectedAttachments, attachment];
    toastStore.showToast(`${type} attached`, 'success');
    
    logger.debug('Attachment added', { type, attachmentCount: selectedAttachments.length });
  }

  // Filter chats when search query changes
  $: if (searchQuery !== undefined) {
    searchChats();
  }
</script>

<MainLayout
  username={username}
  displayName={displayName}
  avatar={avatar}
  on:toggleComposeModal={() => {}}
>
  <div class="flex h-screen border-x border-gray-200 dark:border-gray-800">
    <!-- Left sidebar - Chat list -->
    <div class="w-full md:w-2/5 lg:w-1/3 border-r border-gray-200 dark:border-gray-800">
      <div class="sticky top-0 z-10 bg-white/80 dark:bg-black/80 backdrop-blur-md border-b border-gray-200 dark:border-gray-800 px-4 py-3">
        <h1 class="text-xl font-bold">Messages</h1>
        
        <!-- Search -->
        <div class="relative mt-3">
          <input 
            type="text" 
            bind:value={searchQuery}
            placeholder="Search Direct Messages" 
            class="w-full rounded-full pl-10 pr-4 py-2 {isDarkMode ? 'bg-gray-800 border-gray-700 text-white' : 'bg-gray-100 border-gray-200'}"
          />
          <div class="absolute left-3 top-2.5 text-gray-500">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
            </svg>
          </div>
        </div>
      </div>
      
      <!-- Chat list -->
      <div class="overflow-y-auto">
        {#if isLoadingChats}
          <div class="p-4">
            <div class="animate-pulse space-y-4">
              {#each Array(5) as _}
                <div class="flex space-x-4">
                  <div class="rounded-full bg-gray-300 dark:bg-gray-700 h-12 w-12"></div>
                  <div class="flex-1 space-y-2 py-1">
                    <div class="h-4 bg-gray-300 dark:bg-gray-700 rounded w-3/4"></div>
                    <div class="h-4 bg-gray-300 dark:bg-gray-700 rounded w-1/2"></div>
                  </div>
                </div>
              {/each}
            </div>
          </div>
        {:else if filteredChats.length === 0}
          <div class="flex flex-col items-center justify-center py-16 px-4 text-center">
            <h2 class="text-2xl font-bold mb-2">No messages found</h2>
            <p class="text-gray-500 dark:text-gray-400 mb-8 max-w-md">
              {searchQuery ? 'Try a different search term.' : 'Start a conversation with someone.'}
            </p>
          </div>
        {:else}
          <div class="divide-y divide-gray-200 dark:divide-gray-800">
            {#each filteredChats as chat}
              <button 
                class="w-full text-left p-4 hover:bg-gray-50 dark:hover:bg-gray-900 transition {selectedChat?.id === chat.id ? 'bg-blue-50 dark:bg-blue-900/20' : ''}"
                on:click={() => selectChat(chat)}
              >
                <div class="flex">
                  <div class="flex-shrink-0">
                    <div class="w-12 h-12 rounded-full bg-gray-200 dark:bg-gray-700 flex items-center justify-center overflow-hidden">
                      {#if chat.avatar}
                        <img src={chat.avatar} alt={chat.name} class="w-full h-full object-cover" />
                      {:else}
                        <span class="text-xl">👤</span>
                      {/if}
                    </div>
                  </div>
                  <div class="ml-3 flex-1">
                    <div class="flex items-center justify-between">
                      <span class="font-semibold">{chat.name}</span>
                      <span class="text-gray-500 dark:text-gray-400 text-sm">
                        {chat.lastMessage ? formatTimeAgo(chat.lastMessage.timestamp) : ''}
                      </span>
                    </div>
                    <div class="flex items-center justify-between mt-1">
                      <p class="text-gray-500 dark:text-gray-400 truncate">
                        {chat.lastMessage ? chat.lastMessage.content : 'No messages yet'}
                      </p>
                      {#if chat.unreadCount > 0}
                        <span class="bg-blue-500 text-white text-xs rounded-full w-5 h-5 flex items-center justify-center">
                          {chat.unreadCount}
                        </span>
                      {/if}
                    </div>
                  </div>
                </div>
              </button>
            {/each}
          </div>
        {/if}
      </div>
    </div>
    
    <!-- Right side - Chat content or placeholder -->
    <div class="hidden md:flex md:w-3/5 lg:w-2/3 flex-col">
      {#if selectedChat}
        <!-- Chat header -->
        <div class="sticky top-0 z-10 bg-white/80 dark:bg-black/80 backdrop-blur-md border-b border-gray-200 dark:border-gray-800 px-4 py-3 flex items-center">
          <div class="flex-shrink-0">
            <div class="w-10 h-10 rounded-full bg-gray-200 dark:bg-gray-700 flex items-center justify-center overflow-hidden">
              {#if selectedChat.avatar}
                <img src={selectedChat.avatar} alt={selectedChat.name} class="w-full h-full object-cover" />
              {:else}
                <span class="text-lg">👤</span>
              {/if}
            </div>
          </div>
          <div class="ml-3">
            <h2 class="font-bold">{selectedChat.name}</h2>
            <p class="text-gray-500 dark:text-gray-400 text-sm">
              {selectedChat.type === 'group' ? `${selectedChat.participants.length} members` : ''}
            </p>
          </div>
          <div class="ml-auto flex space-x-2">
            <button class="text-gray-500 hover:text-gray-700 dark:hover:text-gray-300">
              <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            </button>
          </div>
        </div>
        
        <!-- Chat messages -->
        <div class="flex-1 overflow-y-auto p-4 space-y-4">
          {#each selectedChat.messages as message}
            <div class="flex {message.isOwn ? 'justify-end' : 'justify-start'}">
              <div class="max-w-[70%] {message.isOwn ? 'bg-blue-500 text-white rounded-l-lg rounded-tr-lg' : 'bg-gray-200 dark:bg-gray-700 rounded-r-lg rounded-tl-lg'} p-3">
                {#if !message.isOwn}
                  <div class="font-semibold">{message.senderName}</div>
                {/if}
                <div class="{message.isDeleted ? 'italic text-gray-400 dark:text-gray-500' : ''}">
                  {message.isDeleted ? 'This message was deleted' : message.content}
                </div>
                {#if message.attachments.length > 0}
                  <div class="mt-2 space-y-2">
                    {#each message.attachments as attachment}
                      {#if attachment.type === 'image' || attachment.type === 'gif'}
                        <img src={attachment.url} alt="Attachment" class="rounded-lg max-w-full" />
                      {:else if attachment.type === 'video'}
                        <video src={attachment.url} controls class="rounded-lg max-w-full">
                          <track kind="captions" src="path-to-captions.vtt" label="English" default />
                          Your browser does not support the video tag.
                        </video>
                      {/if}
                    {/each}
                  </div>
                {/if}
                <div class="text-xs mt-1 {message.isOwn ? 'text-blue-100' : 'text-gray-500 dark:text-gray-400'}">
                  {formatTimeAgo(message.timestamp)}
                  {#if message.isOwn && !message.isDeleted && isWithinTime(message.timestamp)}
                    <button class="ml-2 hover:underline" on:click={() => unsendMessage(message.id)}>
                      Unsend
                    </button>
                  {/if}
                </div>
              </div>
            </div>
          {/each}
        </div>
        
        <!-- Message input -->
        <div class="border-t border-gray-200 dark:border-gray-800 p-4">
          <div class="flex">
            <div class="flex space-x-2 mr-2">
              <button 
                class="text-gray-600 dark:text-gray-400 hover:text-gray-800 dark:hover:text-gray-200"
                on:click={() => clearSearch()}
                aria-label="Clear search"
              >
                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                  <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clip-rule="evenodd" />
                </svg>
              </button>
              <button
                class="text-gray-500 hover:text-blue-500 p-2 rounded-full hover:bg-gray-100 dark:hover:bg-gray-800"
                on:click={() => handleAttachment('image')}
                aria-label="Attach image"
              >
                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
                </svg>
              </button>
              <button
                class="text-gray-500 hover:text-blue-500 p-2 rounded-full hover:bg-gray-100 dark:hover:bg-gray-800"
                on:click={() => handleAttachment('gif')}
                aria-label="Attach GIF"
              >
                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z" />
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
              </button>
            </div>
            <input 
              type="text" 
              bind:value={newMessage}
              placeholder="Start a new message" 
              class="flex-1 rounded-full px-4 py-2 {isDarkMode ? 'bg-gray-800 border-gray-700 text-white' : 'bg-gray-100 border-gray-200'}"
              on:keydown={(e) => e.key === 'Enter' && sendMessage()}
            />
            <button
              class="p-2 rounded-full bg-blue-500 hover:bg-blue-600 text-white"
              on:click={sendMessage}
              disabled={!newMessage.trim()}
              aria-label="Send message"
            >
              <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8" />
              </svg>
            </button>
          </div>
        </div>
      {:else}
        <!-- Placeholder when no chat is selected -->
        <div class="flex flex-col items-center justify-center h-full px-4 text-center">
          <h2 class="text-2xl font-bold mb-2">Select a message</h2>
          <p class="text-gray-500 dark:text-gray-400 mb-8 max-w-md">
            Choose a conversation from the list or start a new one.
          </p>
        </div>
      {/if}
    </div>
  </div>
</MainLayout>

<style>
  /* Limit multiline text to 1 line */
  .truncate {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  
  /* Skeleton loading animation */
  @keyframes pulse {
    0%, 100% { opacity: 0.5; }
    50% { opacity: 1; }
  }
  .animate-pulse {
    animation: pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite;
  }
</style>
