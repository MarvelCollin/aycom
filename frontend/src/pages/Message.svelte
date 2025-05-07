<script lang="ts">
  import { onMount } from 'svelte';
  import LeftSide from '../components/layout/LeftSide.svelte';
  import { useAuth } from '../hooks/useAuth';
  import { useTheme } from '../hooks/useTheme';
  import type { IAuthStore } from '../interfaces/IAuth';
  import { createLoggerWithPrefix } from '../utils/logger';
  import { toastStore } from '../stores/toastStore';
  import { checkAuth, isWithinTime, formatTimeAgo, handleApiError } from '../utils/common';
  import { listChats, listMessages, sendMessage as apiSendMessage, unsendMessage as apiUnsendMessage, searchMessages, createChat } from '../api/chat';
  import { getProfile, searchUsers } from '../api/user';
  import '../styles/magniview.css'
  
  const logger = createLoggerWithPrefix('Message');

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

  // Dynamic class for mobile view
  let chatSelectedClass = '';
  $: {
    chatSelectedClass = selectedChat ? 'chat-selected' : '';
  }

  // Update class on the container
  $: if (typeof document !== 'undefined') {
    const container = document.querySelector('.message-container');
    if (container) {
      if (selectedChat && window.innerWidth <= 768) {
        container.classList.add('chat-selected');
      } else {
        container.classList.remove('chat-selected');
      }
    }
  }

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
    isVerified: boolean;
  }
  
  // Add user search results state
  let userSearchResults: Participant[] = [];

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
      // Backend returns { chats: [...] }
      if (response && response.chats && Array.isArray(response.chats)) {
        chats = response.chats.map((chat: any) => ({
          id: chat.id || chat.Id,
          type: chat.is_group_chat || chat.IsGroupChat ? 'group' : 'individual',
          name: chat.name || chat.Name || getParticipantName(chat),
          avatar: chat.avatar || null, // You may want to add avatar logic for group/individual
          participants: chat.participants || [], // If not present, leave as empty array
          lastMessage: chat.last_message || chat.lastMessage || undefined,
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
      if (response && response.messages && Array.isArray(response.messages)) {
        selectedChat.messages = response.messages.map((msg: any) => ({
          id: msg.id || msg.Id,
          senderId: msg.sender_id || msg.SenderId,
          senderName: msg.sender_name || '',
          senderAvatar: msg.sender_avatar || '',
          content: msg.content || msg.Content,
          timestamp: msg.timestamp || msg.Timestamp,
          isDeleted: msg.is_deleted || msg.IsDeleted || false,
          attachments: msg.attachments || [],
          isOwn: (msg.sender_id || msg.SenderId) === authState.userId
        }));
      } else {
        selectedChat.messages = [];
      }
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
        const newMsg: Message = {
          id: response.message.id || response.message.Id,
          senderId: response.message.sender_id || response.message.SenderId,
          senderName: response.message.sender_name || displayName,
          senderAvatar: response.message.sender_avatar || avatar,
          content: response.message.content || response.message.Content,
          timestamp: response.message.timestamp || response.message.Timestamp,
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
      userSearchResults = [];
      return;
    }

    const query = searchQuery.toLowerCase();
    logger.debug('Starting search with query:', { query });

    // First, filter local chats by name
    let results = chats.filter(chat => 
      chat.name.toLowerCase().includes(query)
    );
    logger.debug('Filtered local chats:', { count: results.length });
    
    // Search for users via API - this should search ALL users in the system, not just those we've chatted with
    try {
      logger.debug('Calling searchUsers API with query:', { query: searchQuery });
      const response = await searchUsers(searchQuery);
      
      logger.debug('Search users API response:', { 
        status: 'success',
        responseData: JSON.stringify(response)
      });
      
      // Handle different response formats - check both direct response.users and response.data.users
      const users = response?.users || (response?.data?.users || []);
      
      if (users && users.length > 0) {
        // Transform API user results to match our participant format
        userSearchResults = users.map(user => ({
          id: user.id,
          username: user.username,
          displayName: user.name,
          avatar: user.profile_picture_url,
          bio: user.bio,
          isVerified: user.is_verified,
          isFollowing: user.is_following,
        }));
        logger.debug('Retrieved users from API', { count: userSearchResults.length, users: userSearchResults.map(u => u.username) });
      } else {
        userSearchResults = [];
        logger.warn('No users found or invalid API response format', { response });
      }
    } catch (error) {
      logger.error('Error searching users:', error);
      userSearchResults = [];
      // Fallback to client-side searching if API fails
      const uniqueUsers = new Map<string, Participant>();
      
      chats.forEach(chat => {
        if (chat.participants && chat.participants.length > 0) {
          chat.participants.forEach(participant => {
            // Skip if it's the current user or already in the map
            if (participant.id === authState.userId || uniqueUsers.has(participant.id)) {
              return;
            }
            
            // Check if user matches search query
            const displayName = participant.displayName || participant.username || '';
            const username = participant.username || '';
            
            if (displayName.toLowerCase().includes(query) || 
                username.toLowerCase().includes(query)) {
              uniqueUsers.set(participant.id, participant);
            }
          });
        }
      });
      
      userSearchResults = Array.from(uniqueUsers.values());
      logger.debug('Using fallback search results', { count: userSearchResults.length });
    }
    
    // If we have a selected chat, also search messages
    if (selectedChat) {
      try {
        logger.debug('Searching messages in chat:', { chatId: selectedChat.id, query: searchQuery });
        const response = await searchMessages(selectedChat.id, searchQuery);
        
        logger.debug('Search messages API response:', { 
          status: 'success',
          responseData: JSON.stringify(response)
        });
        
        if (response && response.messages && response.messages.length > 0) {
          // If the current chat isn't in results but has matching messages, add it
          if (!results.some(chat => chat.id === selectedChat!.id)) {
            results.push(selectedChat);
            logger.debug('Added current chat to results due to matching messages', 
              { chatId: selectedChat.id, matchingMessages: response.messages.length });
          }
        }
      } catch (error) {
        logger.error('Error searching messages:', error);
      }
    }
    
    filteredChats = results;
    logger.debug('Search completed', { 
      query: searchQuery, 
      chatResults: filteredChats.length,
      userResults: userSearchResults.length,
      userDetails: userSearchResults.map(u => u.username).join(', ')
    });
  }

  // Start a new chat with a user
  async function startChatWithUser(user: Participant) {
    // Add more logging
    logger.debug('Starting chat with user', { userId: user.id, username: user.username });
    
    // Check if we already have a chat with this user
    const existingChat = chats.find(chat => 
      chat.type === 'individual' && 
      chat.participants.some(p => p.id === user.id)
    );
    
    if (existingChat) {
      // If chat exists, select it
      logger.debug('Found existing chat', { chatId: existingChat.id });
      selectChat(existingChat);
    } else {
      // Create a new chat with this user via API
      try {
        const chatData = {
          type: 'individual',
          participants: [user.id]
        };
        
        logger.debug('Creating new chat', { chatData });
        const response = await createChat(chatData);
        logger.debug('Create chat API response', { response });
        
        if (response && response.chat) {
          // Format the received chat to match our chat structure
          const newChat: Chat = {
            id: response.chat.id,
            name: user.displayName || user.username,
            type: 'individual',
            lastMessage: undefined, // Using undefined instead of null to satisfy TypeScript
            avatar: user.avatar,
            participants: [
              {
                id: user.id,
                username: user.username,
                displayName: user.displayName,
                avatar: user.avatar,
                isVerified: user.isVerified
              },
              // Current user is also a participant, but we don't need to add it here
              // as the backend should already include it
            ],
            messages: [],
            unreadCount: 0
          };
          
          // Add the new chat to the list and select it
          logger.debug('Adding new chat to list', { newChatId: newChat.id });
          chats = [newChat, ...chats];
          filteredChats = [newChat, ...filteredChats];
          selectChat(newChat);
          
          toastStore.showToast(`Chat with ${user.displayName || user.username} created`, 'success');
        } else {
          logger.error('Invalid response format from create chat API', { response });
          toastStore.showToast(`Failed to create chat: Invalid response`, 'error');
        }
      } catch (error) {
        const errorDetail = handleApiError(error);
        logger.error('Failed to create chat:', errorDetail);
        toastStore.showToast(`Failed to create chat with ${user.displayName || user.username}`, 'error');
      }
      
      // Clear search
      clearSearch();
    }
  }

  // Clear search
  function clearSearch() {
    searchQuery = '';
    filteredChats = [...chats];
    userSearchResults = [];
    logger.debug('Search cleared');
  }

  // Handle file attachments
  function handleAttachment(type: 'image' | 'gif' | 'video') {
    // Create a file input element
    const fileInput = document.createElement('input');
    fileInput.type = 'file';
    
    if (type === 'image') {
      fileInput.accept = 'image/*';
    } else if (type === 'gif') {
      fileInput.accept = 'image/gif';
    } else if (type === 'video') {
      fileInput.accept = 'video/*';
    }
    
    // Set up the change handler
    fileInput.onchange = async (e) => {
      const target = e.target as HTMLInputElement;
      if (!target.files || target.files.length === 0) return;
      
      const file = target.files[0];
      
      try {
        // In a real implementation, you would upload this file to your server/storage
        // and get back a URL. For now, we'll use a local object URL.
        const localUrl = URL.createObjectURL(file);
        
        const attachment: Attachment = {
          id: `temp-${Date.now()}`,
          type,
          url: localUrl,
        };
        
        selectedAttachments = [...selectedAttachments, attachment];
        toastStore.showToast(`${type} attached`, 'success');
        
        logger.debug('Attachment added', { type, filename: file.name, size: file.size });
      } catch (error) {
        logger.error('Failed to process attachment', { error });
        toastStore.showToast('Failed to attach file. Please try again.', 'error');
      }
    };
    
    // Trigger the file dialog
    fileInput.click();
  }

  // Filter chats when search query changes
  $: if (searchQuery !== undefined) {
    searchChats();
  }
</script>

<div class="message-container {chatSelectedClass}">
  <!-- Left navigation/profile -->
  <div class="left-sidebar">
    <LeftSide username={username} displayName={displayName} avatar={avatar} />
  </div>

  <!-- Middle: Chat List / Search -->
  <div class="middle-section">
    <div class="section-header">
      <h1>Messages</h1>
    </div>
    <div class="search-container">
      <div class="search-input-wrapper">
        <input
          type="text"
          bind:value={searchQuery}
          placeholder="Search Direct Messages"
          class="search-input"
        />
        {#if searchQuery.trim() !== ''}
          <button class="clear-search-button" on:click={clearSearch}>
            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" width="16" height="16">
              <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clip-rule="evenodd" />
            </svg>
          </button>
        {/if}
      </div>
      
      {#if searchQuery.trim() !== '' && (userSearchResults.length > 0 || filteredChats.length > 0)}
        <div class="search-dropdown">
          {#if userSearchResults.length > 0}
            <div class="search-dropdown-section">
              <h4 class="search-dropdown-title">Users</h4>
              <ul class="search-dropdown-list">
                {#each userSearchResults as user}
                  <li>
                    <button class="dropdown-item" on:click={() => startChatWithUser(user)}>
                      <div class="avatar-container">
                        {#if user.avatar}
                          <img src={user.avatar} alt={user.displayName || user.username} class="avatar-image" />
                        {:else}
                          <span class="avatar-placeholder">👤</span>
                        {/if}
                      </div>
                      <div class="user-info">
                        <span class="user-name">{user.displayName || user.username}</span>
                        {#if user.username && user.username !== user.displayName}
                          <span class="user-username">@{user.username}</span>
                        {/if}
                      </div>
                    </button>
                  </li>
                {/each}
              </ul>
            </div>
          {/if}
          
          {#if filteredChats.length > 0}
            <div class="search-dropdown-section">
              <h4 class="search-dropdown-title">Conversations</h4>
              <ul class="search-dropdown-list">
                {#each filteredChats as chat}
                  <li>
                    <button class="dropdown-item" on:click={() => selectChat(chat)}>
                      <div class="avatar-container">
                        {#if chat.avatar}
                          <img src={chat.avatar} alt={chat.name} class="avatar-image" />
                        {:else}
                          <span class="avatar-placeholder">👤</span>
                        {/if}
                      </div>
                      <div class="user-info">
                        <span class="user-name">{chat.name}</span>
                        {#if chat.lastMessage}
                          <span class="chat-preview">{chat.lastMessage.content.substring(0, 30)}{chat.lastMessage.content.length > 30 ? '...' : ''}</span>
                        {/if}
                      </div>
                    </button>
                  </li>
                {/each}
              </ul>
            </div>
          {/if}
        </div>
      {/if}
    </div>
    <div class="chat-list">
      {#if isLoadingChats}
        <div class="loading-message">Loading chats...</div>
      {:else if userSearchResults.length === 0 && filteredChats.length === 0}
        <div class="empty-message">
          {chats.length === 0 ? 'No conversations yet' : 'No results found'}
        </div>
      {:else}
        {#if userSearchResults.length > 0}
          <div class="search-results-section">
            <h3 class="search-section-title">Users</h3>
            <ul class="user-results">
              {#each userSearchResults as user}
                <li>
                  <button class="user-result-item" on:click={() => startChatWithUser(user)}>
                    <div class="avatar-container">
                      {#if user.avatar}
                        <img src={user.avatar} alt={user.displayName || user.username} class="avatar-image" />
                      {:else}
                        <span class="avatar-placeholder">👤</span>
                      {/if}
                    </div>
                    <div class="user-info">
                      <span class="user-name">{user.displayName || user.username}</span>
                      {#if user.username && user.username !== user.displayName}
                        <span class="user-username">@{user.username}</span>
                      {/if}
                    </div>
                  </button>
                </li>
              {/each}
            </ul>
          </div>
        {/if}
        
        {#if filteredChats.length > 0}
          <div class="search-results-section">
            {#if userSearchResults.length > 0}
              <h3 class="search-section-title">Chats</h3>
            {/if}
            <ul class="chat-items">
              {#each filteredChats as chat}
                <li>
                  <button
                    class="chat-item {selectedChat?.id === chat.id ? 'active' : ''}"
                    on:click={() => selectChat(chat)}
                  >
                    <div class="avatar-container">
                      {#if chat.avatar}
                        <img src={chat.avatar} alt={chat.name} class="avatar-image" />
                      {:else}
                        <span class="avatar-placeholder">👤</span>
                      {/if}
                    </div>
                    <div class="chat-info">
                      <div class="chat-header">
                        <span class="chat-name">{chat.name}</span>
                        <span class="chat-time">{chat.lastMessage ? formatTimeAgo(chat.lastMessage.timestamp) : ''}</span>
                      </div>
                      <div class="chat-preview">
                        {chat.lastMessage ? chat.lastMessage.content : 'No messages yet'}
                      </div>
                    </div>
                    {#if chat.unreadCount > 0}
                      <span class="unread-badge">{chat.unreadCount}</span>
                    {/if}
                  </button>
                </li>
              {/each}
            </ul>
          </div>
        {/if}
      {/if}
    </div>
  </div>

  <!-- Right: Chat Content -->
  <div class="right-section">
    {#if selectedChat}
      <!-- Chat header -->
      <div class="chat-header">
        <div class="chat-avatar">
          {#if selectedChat.avatar}
            <img src={selectedChat.avatar} alt={selectedChat.name} class="avatar-image" />
          {:else}
            <span class="avatar-placeholder">👤</span>
          {/if}
        </div>
        <div class="chat-title">
          <h2>{selectedChat.name}</h2>
          <p class="group-info">{selectedChat.type === 'group' ? `${selectedChat.participants.length} members` : ''}</p>
        </div>
      </div>
      
      <!-- Messages -->
      <div class="messages-container">
        {#each selectedChat.messages as message}
          <div class="message-wrapper {message.isOwn ? 'own-message' : 'other-message'}">
            <div class="message-bubble">
              {#if !message.isOwn}
                <div class="sender-name">{message.senderName}</div>
              {/if}
              <div class={message.isDeleted ? 'deleted-message' : ''}>
                {message.isDeleted ? 'This message was deleted' : message.content}
              </div>
              {#if message.attachments.length > 0}
                <div class="attachments">
                  {#each message.attachments as attachment}
                    {#if attachment.type === 'image' || attachment.type === 'gif'}
                      <img src={attachment.url} alt="Attachment" class="attachment-image" />
                    {:else if attachment.type === 'video'}
                      <video src={attachment.url} controls class="attachment-video">
                        <track kind="captions" src="path-to-captions.vtt" label="English" default />
                        Your browser does not support the video tag.
                      </video>
                    {/if}
                  {/each}
                </div>
              {/if}
              <div class="message-meta">
                {formatTimeAgo(message.timestamp)}
                {#if message.isOwn && !message.isDeleted && isWithinTime(message.timestamp)}
                  <button class="unsend-button" on:click={() => unsendMessage(message.id)}>
                    Unsend
                  </button>
                {/if}
              </div>
            </div>
          </div>
        {/each}
      </div>
      
      <!-- Message input -->
      <div class="message-input-container">
        <div class="message-actions">
          <button class="action-button" on:click={() => clearSearch()} aria-label="Clear search">
            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor">
              <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clip-rule="evenodd" />
            </svg>
          </button>
          <button class="action-button" on:click={() => handleAttachment('image')} aria-label="Attach image">
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
            </svg>
          </button>
          <button class="action-button" on:click={() => handleAttachment('gif')} aria-label="Attach GIF">
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z" />
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
          </button>
        </div>
        <input
          type="text"
          bind:value={newMessage}
          placeholder="Start a new message"
          class="message-input"
          on:keydown={(e) => e.key === 'Enter' && sendMessage()}
        />
        <button
          class="send-button"
          on:click={sendMessage}
          disabled={!newMessage.trim()}
          aria-label="Send message"
        >
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8" />
          </svg>
        </button>
      </div>
    {:else}
      <div class="empty-chat">
        <h2>Select a message</h2>
        <p>Choose from your existing conversations, start a new one, or just keep swimming.</p>
        <button class="new-message-button">New message</button>
      </div>
    {/if}
  </div>
</div>

<style>
  /* Main Container */
  .message-container {
    display: grid;
    grid-template-columns: 288px 300px 1fr;
    height: 100vh;
    background-color: var(--background-color, white);
    color: var(--text-color, black);
  }

  /* Dark mode overrides */
  :global(.dark) .message-container {
    --background-color: black;
    --text-color: white;
    --border-color: #2d3748;
    --hover-bg: #1a202c;
    --active-bg: rgba(29, 78, 216, 0.2);
    --message-bg: #2d3748;
    --own-message-bg: #3b82f6;
    --input-bg: #1a202c;
  }

  /* Light mode variables */
  .message-container {
    --border-color: #e2e8f0;
    --hover-bg: #f7fafc;
    --active-bg: rgba(59, 130, 246, 0.1);
    --message-bg: #e2e8f0;
    --own-message-bg: #3b82f6;
    --input-bg: #f7fafc;
  }

  /* Left Sidebar */
  .left-sidebar {
    border-right: 1px solid var(--border-color);
    height: 100%;
    overflow-y: auto;
    min-width: 288px;
  }

  /* Middle Section */
  .middle-section {
    display: flex;
    flex-direction: column;
    border-right: 1px solid var(--border-color);
    height: 100%;
  }

  .section-header {
    padding: 16px;
    border-bottom: 1px solid var(--border-color);
  }

  .section-header h1 {
    font-size: 1.25rem;
    font-weight: bold;
  }

  .search-container {
    padding: 12px;
    border-bottom: 1px solid var(--border-color);
    position: relative;
  }

  .search-input-wrapper {
    position: relative;
    display: flex;
    align-items: center;
  }

  .search-input {
    width: 100%;
    padding: 8px 16px;
    border-radius: 9999px;
    border: 1px solid var(--border-color);
    background-color: var(--input-bg);
    color: var(--text-color);
  }

  .clear-search-button {
    position: absolute;
    right: 8px;
    background: none;
    border: none;
    padding: 4px;
    color: var(--text-secondary, #6c757d);
    cursor: pointer;
    border-radius: 50%;
  }

  .clear-search-button:hover {
    background-color: var(--hover-bg);
  }

  .search-dropdown {
    position: absolute;
    top: calc(100% + 4px);
    left: 12px;
    right: 12px;
    max-height: 350px;
    overflow-y: auto;
    background-color: var(--background-color);
    border: 1px solid var(--border-color);
    border-radius: 8px;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
    z-index: 1000;
  }

  .search-dropdown-section {
    padding: 8px;
  }

  .search-dropdown-section + .search-dropdown-section {
    border-top: 1px solid var(--border-color);
  }

  .search-dropdown-title {
    font-size: 12px;
    color: var(--text-secondary, #6c757d);
    text-transform: uppercase;
    letter-spacing: 0.5px;
    margin: 4px 8px 8px;
  }

  .dropdown-item {
    padding: 8px;
    border-radius: 6px;
    margin-bottom: 2px;
    transition: background-color 0.15s ease;
  }

  .dropdown-item:hover {
    background-color: var(--hover-bg);
  }

  .dropdown-item .avatar-container {
    width: 36px;
    height: 36px;
    margin-right: 10px;
  }

  .dropdown-item .user-name {
    font-weight: 600;
    display: block;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .dropdown-item .user-username,
  .dropdown-item .chat-preview {
    font-size: 12px;
    color: var(--text-secondary, #6c757d);
    display: block;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .chat-info {
    flex: 1;
    min-width: 0;
  }

  .chat-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .search-container {
    padding: 12px;
    border-bottom: 1px solid var(--border-color);
    position: relative;
  }

  .search-input-wrapper {
    position: relative;
    display: flex;
    align-items: center;
  }

  .chat-name {
    font-weight: 600;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .chat-time {
    font-size: 0.75rem;
    color: gray;
    margin-left: 8px;
    flex-shrink: 0;
  }

  .chat-preview {
    color: gray;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    font-size: 0.875rem;
  }

  .unread-badge {
    background-color: #3b82f6;
    color: white;
    font-size: 0.75rem;
    width: 20px;
    height: 20px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    margin-left: 8px;
    flex-shrink: 0;
  }

  /* Right Section */
  .right-section {
    display: flex;
    flex-direction: column;
    height: 100%;
  }

  .chat-header {
    display: flex;
    align-items: center;
    padding: 16px;
    border-bottom: 1px solid var(--border-color);
  }

  .chat-avatar {
    width: 40px;
    height: 40px;
    border-radius: 50%;
    overflow: hidden;
    background-color: #6b7280;
    display: flex;
    align-items: center;
    justify-content: center;
    margin-right: 12px;
  }

  .chat-title h2 {
    font-weight: bold;
    margin: 0;
  }

  .group-info {
    font-size: 0.875rem;
    color: gray;
    margin: 0;
  }

  .messages-container {
    flex: 1;
    overflow-y: auto;
    padding: 16px;
    background-color: #f9fafb;
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  :global(.dark) .messages-container {
    background-color: black;
  }

  .message-wrapper {
    display: flex;
  }

  .own-message {
    justify-content: flex-end;
  }

  .other-message {
    justify-content: flex-start;
  }

  .message-bubble {
    max-width: 70%;
    padding: 12px;
    border-radius: 12px;
    position: relative;
  }

  .own-message .message-bubble {
    background-color: var(--own-message-bg);
    color: white;
    border-top-right-radius: 4px;
  }

  .other-message .message-bubble {
    background-color: var(--message-bg);
    border-top-left-radius: 4px;
  }

  .sender-name {
    font-weight: 600;
    margin-bottom: 4px;
  }

  .deleted-message {
    font-style: italic;
    color: #9ca3af;
  }

  .attachments {
    margin-top: 8px;
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .attachment-image, .attachment-video {
    max-width: 100%;
    border-radius: 8px;
  }

  .message-meta {
    display: flex;
    align-items: center;
    font-size: 0.75rem;
    margin-top: 4px;
    color: rgba(255, 255, 255, 0.7);
  }

  .other-message .message-meta {
    color: gray;
  }

  .unsend-button {
    margin-left: 8px;
    background: none;
    border: none;
    color: inherit;
    text-decoration: underline;
    cursor: pointer;
    padding: 0;
  }

  .message-input-container {
    display: flex;
    align-items: center;
    padding: 12px;
    border-top: 1px solid var(--border-color);
  }

  .message-actions {
    display: flex;
    margin-right: 8px;
  }

  .action-button {
    background: none;
    border: none;
    color: gray;
    width: 36px;
    height: 36px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    margin-right: 4px;
  }

  .action-button:hover {
    background-color: var(--hover-bg);
    color: #3b82f6;
  }

  .action-button svg {
    width: 20px;
    height: 20px;
  }

  .message-input {
    flex: 1;
    padding: 10px 16px;
    border-radius: 9999px;
    border: 1px solid var(--border-color);
    background-color: var(--input-bg);
    color: var(--text-color);
  }

  .send-button {
    width: 40px;
    height: 40px;
    border-radius: 50%;
    background-color: #3b82f6;
    color: white;
    border: none;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    margin-left: 8px;
  }

  .send-button:hover {
    background-color: #2563eb;
  }

  .send-button[disabled] {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .send-button svg {
    width: 20px;
    height: 20px;
  }

  .empty-chat {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 100%;
    text-align: center;
    padding: 0 16px;
  }

  .empty-chat h2 {
    font-size: 1.5rem;
    font-weight: bold;
    margin-bottom: 8px;
  }

  .empty-chat p {
    color: gray;
    margin-bottom: 24px;
    max-width: 400px;
  }

  .new-message-button {
    background-color: #3b82f6;
    color: white;
    font-weight: bold;
    padding: 8px 24px;
    border-radius: 9999px;
    border: none;
    cursor: pointer;
  }

  .new-message-button:hover {
    background-color: #2563eb;
  }

  /* Media Queries for Responsive Layout */
  @media (max-width: 1280px) {
    .message-container {
      grid-template-columns: 288px 250px 1fr;
    }
  }

  @media (max-width: 1024px) {
    .message-container {
      grid-template-columns: 240px 200px 1fr;
    }
  }

  @media (max-width: 768px) {
    .message-container {
      grid-template-columns: 72px 1fr;
    }
    
    .middle-section {
      display: var(--middle-display, flex);
    }
    
    .right-section {
      display: var(--right-display, none);
    }
    
    /* When a chat is selected, show right and hide middle */
    .message-container.chat-selected {
      --middle-display: none;
      --right-display: flex;
    }
  }

  @media (max-width: 640px) {
    .message-container {
      grid-template-columns: 1fr;
    }
    
    .left-sidebar {
      display: none;
    }
  }

  /* Search results styling */
  .search-results-section {
    margin-bottom: 16px;
  }
  
  .search-section-title {
    font-size: 14px;
    color: #6c757d;
    margin: 8px 0;
    padding-left: 8px;
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }
  
  .user-results {
    list-style: none;
    padding: 0;
    margin: 0;
  }
  
  .user-result-item {
    display: flex;
    align-items: center;
    width: 100%;
    padding: 10px;
    border: none;
    background: none;
    text-align: left;
    cursor: pointer;
    border-radius: 8px;
    transition: background-color 0.2s;
  }
  
  .user-result-item:hover {
    background-color: #f0f2f5;
  }
  
  .user-info {
    margin-left: 10px;
    display: flex;
    flex-direction: column;
  }
  
  .user-name {
    font-weight: 500;
  }
  
  .user-username {
    font-size: 12px;
    color: #6c757d;
  }
  
  .dark-theme .user-result-item:hover {
    background-color: #2d3748;
  }
  
  .dark-theme .search-section-title {
    color: #cbd5e0;
  }
  
  .dark-theme .user-username {
    color: #a0aec0;
  }

  /* Make sure the chat list takes remaining space */
  .chat-list {
    flex: 1;
    overflow-y: auto;
  }

  /* Ensure these base styles exist */
  .avatar-container {
    width: 48px;
    height: 48px;
    border-radius: 50%;
    overflow: hidden;
    background-color: #6b7280;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
    margin-right: 12px;
  }
  
  .avatar-image {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }
  
  .avatar-placeholder {
    font-size: 1.25rem;
  }
  
  .user-info {
    flex: 1;
    min-width: 0;
  }
  
  .search-dropdown-list {
    list-style: none;
    padding: 0;
    margin: 0;
  }
  
  .dropdown-item {
    display: flex;
    align-items: center;
    width: 100%;
    padding: 8px;
    border: none;
    background: none;
    text-align: left;
    cursor: pointer;
    border-radius: 6px;
    margin-bottom: 2px;
    transition: background-color 0.15s ease;
  }
</style>
