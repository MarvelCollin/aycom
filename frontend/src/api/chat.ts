import { createLoggerWithPrefix } from '../utils/logger';
import appConfig from '../config/appConfig';
import { getAuthToken } from '../utils/auth';

const API_BASE_URL = appConfig.api.baseUrl;
const logger = createLoggerWithPrefix('ChatAPI');

let messageHandler: ((message: any) => void) | null = null;

export function setMessageHandler(handler: (message: any) => void) {
  messageHandler = handler;
}

export function processWebSocketMessage(message: any) {
  try {
    logger.debug('Processing WebSocket message:', message);
    
    // Validate the message format
    if (!message || !message.type) {
      logger.warn('Invalid WebSocket message format:', message);
      return;
    }

    // Handle different message types
    switch (message.type) {
      case 'text':
        handleIncomingTextMessage(message);
        break;
      case 'typing':
        // Handle typing indicator
        break;
      case 'read':
        // Handle read receipts
        break;
      case 'delete':
        // Handle message deletion
        break;
      case 'edit':
        // Handle message edit
        break;
      case 'connection_status':
        logger.info('WebSocket connection status:', message.status);
        break;
      default:
        logger.warn('Unknown WebSocket message type:', message.type);
    }
  } catch (error) {
    logger.error('Error processing WebSocket message:', error);
  }
}

// Extract temp ID from message
function extractTempIdFromMessage(message: any): string | null {
  try {
    if (message && message.temp_id) {
      return message.temp_id;
    }
    
    if (message && message.content && typeof message.content === 'string' && message.content.startsWith('temp-')) {
      const match = message.content.match(/temp-(\d+)/);
      if (match && match[1]) {
        return `temp-${match[1]}`;
      }
    }
    
    return null;
  } catch (error) {
    logger.error('Error extracting temp ID from message:', error);
    return null;
  }
}

// Handle incoming text message
function handleIncomingTextMessage(message: any) {
  logger.debug('Handling incoming text message:', message);
  
  // Create a standardized message object
  const processedMessage = {
    id: message.id || message.message_id,
    chat_id: message.chat_id,
    content: message.content,
    sender_id: message.user_id || message.sender_id,
    sender_name: message.sender_name || 'User',
    sender_avatar: message.sender_avatar,
    timestamp: message.timestamp ? new Date(message.timestamp).toISOString() : new Date().toISOString(),
    is_read: message.is_read || false,
    is_edited: message.is_edited || false,
    is_deleted: message.is_deleted || false
  };
  
  // Check if this is a response to a temporary message
  const tempId = extractTempIdFromMessage(message);
  
  // Notify the registered message handler
  if (messageHandler) {
    try {
      messageHandler({
        ...processedMessage,
        temp_id: tempId
      });
    } catch (handlerError) {
      logger.error('Error in message handler:', handlerError);
    }
  }
}

export async function createChat(data: Record<string, any>) {
  try {
    const token = getAuthToken();
    logger.debug('Creating chat with data', { data, apiUrl: `${API_BASE_URL}/chats` });

    // Check for individual chat type and see if chat already exists
    if (data.type === 'individual' && data.participants && data.participants.length === 1) {
      const participantId = data.participants[0];

      try {
        const existingChats = await getChatHistoryList();
        logger.debug('Checking existing chats for participant', { participantId, chatsCount: existingChats.chats?.length || 0 });

        const existingChat = existingChats.chats?.find(chat => {
          if (chat.is_group_chat || !chat.participants || chat.participants.length !== 2) {
            return false;
          }

          return chat.participants.some(p => 
            (p.id === participantId || p.user_id === participantId)
          );
        });

        if (existingChat) {
          logger.debug('Found existing chat with this participant', { chatId: existingChat.id });
          return { 
            success: true,
            chat: existingChat
          };
        }
      } catch (err) {
        logger.warn('Error checking existing chats:', err);
      }
    }

    // Send the request
    const response = await fetch(`${API_BASE_URL}/chats`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      body: JSON.stringify(data)
    });

    if (!response.ok) {
      const contentType = response.headers.get('content-type') || '';
      if (contentType && contentType.includes('application/json')) {
        try {
          const errorData = await response.json();
          throw new Error(errorData.message || 'Failed to create chat');
        } catch (parseError) {
          throw new Error(`Failed to create chat: ${response.status} ${response.statusText}`);
        }
      } else {
        throw new Error(`Failed to create chat: ${response.status} ${response.statusText}`);
      }
    }

    // Check if the response is empty
    const contentLength = response.headers.get('content-length');
    if (contentLength === '0') {
      logger.warn('Empty response received from create chat endpoint');
      return {
        success: true,
        chat: {
          id: `temp-${Date.now()}`,
          name: data.name || 'New Chat',
          is_group_chat: data.type === 'group',
          created_at: new Date().toISOString(),
          participants: data.participants || []
        }
      };
    }

    try {
      const responseText = await response.text();
      if (!responseText || responseText.trim() === '') {
        logger.warn('Empty response body from create chat endpoint');
        return {
          success: true,
          chat: {
            id: `temp-${Date.now()}`,
            name: data.name || 'New Chat',
            is_group_chat: data.type === 'group',
            created_at: new Date().toISOString(),
            participants: data.participants || []
          }
        };
      }
      
      const jsonResponse = JSON.parse(responseText);
      logger.debug('Chat creation response', { jsonResponse });
      
      // Enhanced logging to debug response format
      logger.debug('Chat response analysis', { 
        hasSuccess: jsonResponse && typeof jsonResponse.success === 'boolean',
        successValue: jsonResponse?.success,
        hasData: jsonResponse && jsonResponse.data !== undefined,
        dataType: jsonResponse?.data ? typeof jsonResponse.data : 'undefined', 
        hasChat: jsonResponse && jsonResponse.chat !== undefined,
        chatType: jsonResponse?.chat ? typeof jsonResponse.chat : 'undefined',
        responseKeys: jsonResponse ? Object.keys(jsonResponse) : [] 
      });
      
      // Check for the standardized response format
      if (jsonResponse && jsonResponse.success === true && jsonResponse.data) {
        // Server returns {"success": true, "data": { "chat": { ... } }}
        logger.debug('Using success+data response format');
        return jsonResponse.data;
      } else if (jsonResponse && jsonResponse.chat) {
        // Direct chat object in response
        logger.debug('Using direct chat object response format');
        return jsonResponse;
      } else if (jsonResponse && typeof jsonResponse === 'object') {
        // Just return whatever we got if it's an object
        logger.debug('Using generic object response format');
        return jsonResponse;
      } else {
        // Create a fallback response if format doesn't match any expected pattern
        logger.warn('Unexpected chat response format, creating fallback', { jsonResponse });
        return {
          chat: {
            id: `fallback-${Date.now()}`,
            name: data.name || 'Chat',
            is_group_chat: data.type === 'group',
            created_by: 'current-user',
            created_at: new Date().toISOString(),
            updated_at: new Date().toISOString(),
            participants: data.participants || []
          }
        };
      }
    } catch (parseError) {
      logger.error('Failed to parse JSON response for chat creation:', parseError);
      throw new Error(`Failed to parse response when creating chat: ${response.status} ${response.statusText}`);
    }
  } catch (error) {
    logger.error('Create chat failed:', error);
    throw error;
  }
}

export async function listChats() {
  try {
    const token = getAuthToken();
    logger.debug('Fetching chat list');
    
    // Log API URL to ensure it's correct
    logger.info(`API URL for chats: ${API_BASE_URL}/chats`);

    const response = await fetch(`${API_BASE_URL}/chats`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Accept': 'application/json',
        'Authorization': token ? `Bearer ${token}` : ''
      },
      credentials: 'include'
    });

    // Log the response status
    logger.info(`Chat list response status: ${response.status} ${response.statusText}`);
    
    // Log response headers to debug content-type issues
    const contentTypeHeader = response.headers.get('content-type');
    logger.info(`Response content type: ${contentTypeHeader || 'not provided'}`);

    if (!response.ok) {
      // Try to log the error body for more context
      try {
        const errorText = await response.text();
        logger.error(`Error response body: ${errorText}`);
      } catch (textError) {
        logger.error('Could not read error response body', textError);
      }

      const contentType = response.headers.get('content-type') || '';
      if (contentType && contentType.includes('application/json')) {
        try {
          const errorData = await response.json();
          throw new Error(errorData.message || 'Failed to list chats');
        } catch (parseError) {
          logger.error('Could not parse error response as JSON', parseError);
          throw new Error(`Failed to list chats: ${response.status} ${response.statusText}`);
        }
      } else {
        throw new Error(`Failed to list chats: ${response.status} ${response.statusText}`);
      }
    }

    // Check if the response is empty
    const contentLength = response.headers.get('content-length');
    if (contentLength === '0') {
      logger.warn('Empty response received from list chats endpoint');
      return { chats: [] };
    }

    try {
      const responseText = await response.text();
      if (!responseText || responseText.trim() === '') {
        logger.warn('Empty response body from list chats endpoint');
        return { chats: [] };
      }
      
      try {
        const responseData = JSON.parse(responseText);
        // Log the shape of the response data
        logger.debug('API response structure:', {
          hasData: !!responseData,
          hasChats: responseData && 'chats' in responseData,
          isChatsArray: responseData && 'chats' in responseData && Array.isArray(responseData.chats),
          chatsLength: responseData?.chats?.length || 0,
          responseKeys: responseData ? Object.keys(responseData) : []
        });
        
        // Handle different response formats
        if (responseData && 'chats' in responseData && Array.isArray(responseData.chats)) {
          return responseData;
        } else if (Array.isArray(responseData)) {
          return { chats: responseData };
        } else if (responseData && typeof responseData === 'object') {
          // If the response is a single chat object (has id and name)
          if (responseData.id && (responseData.name || responseData.participants)) {
            return { chats: [responseData] };
          }
          // If it's some other kind of object, try to extract chats
          if (responseData.data && Array.isArray(responseData.data)) {
            return { chats: responseData.data };
          }
          return { chats: [] };
        } else {
          return { chats: [] };
        }
      } catch (parseError: unknown) {
        logger.error('Failed to parse JSON response for listing chats:', parseError);
        // Try to log the raw response for debugging
        if (responseText) {
          logger.error('Raw response text:', responseText.substring(0, 200) + (responseText.length > 200 ? '...' : ''));
        }
        return { chats: [] };
      }
    } catch (textError) {
      logger.error('Could not read response text', textError);
      return { chats: [] };
    }
  } catch (error) {
    logger.error('List chats failed:', error);
    throw error;
  }
}

export async function listChatParticipants(chatId: string) {
  try {
    const token = getAuthToken();

    const response = await fetch(`${API_BASE_URL}/chats/${chatId}/participants`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': token ? `Bearer ${token}` : ''
      },
      credentials: 'include'
    });

    if (!response.ok) {
      const contentType = response.headers.get('content-type') || '';
      if (contentType && contentType.includes('application/json')) {
        const errorData = await response.json();
        throw new Error(errorData.message || 'Failed to list chat participants');
      } else {
        throw new Error(`Failed to list chat participants: ${response.status} ${response.statusText}`);
      }
    }

    const contentType = response.headers.get('content-type') || '';
    if (contentType && contentType.includes('application/json')) {
      try {
        const data = await response.json();
        return data;
      } catch (parseError: unknown) {
        logger.error(`Failed to parse JSON response for listing participants in chat ${chatId}:`, parseError);
        return { participants: [] };
      }
    } else {
      logger.warn(`Non-JSON response for listing participants in chat ${chatId}`);
      return { participants: [] };
    }
  } catch (error) {
    logger.error(`List participants for chat ${chatId} failed:`, error);
    throw error;
  }
}

export async function addChatParticipant(chatId: string, data: Record<string, any>) {
  try {
    const token = getAuthToken();

    const response = await fetch(`${API_BASE_URL}/chats/${chatId}/participants`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': token ? `Bearer ${token}` : ''
      },
      body: JSON.stringify(data),
      credentials: 'include'
    });

    if (!response.ok) {
      const contentType = response.headers.get('content-type') || '';
      if (contentType && contentType.includes('application/json')) {
        const errorData = await response.json();
        throw new Error(errorData.message || 'Failed to add chat participant');
      } else {
        throw new Error(`Failed to add chat participant: ${response.status} ${response.statusText}`);
      }
    }

    const contentType = response.headers.get('content-type') || '';
    if (contentType && contentType.includes('application/json')) {
      try {
        return await response.json();
      } catch (parseError: unknown) {
        logger.error(`Failed to parse JSON response for adding participant to chat ${chatId}:`, parseError);
        return { success: true };
      }
    } else {
      logger.warn(`Non-JSON response for adding participant to chat ${chatId}`);
      return { success: true };
    }
  } catch (error) {
    logger.error(`Add participant to chat ${chatId} failed:`, error);
    throw error;
  }
}

export async function removeChatParticipant(chatId: string, userId: string) {
  try {
    const token = getAuthToken();

    const response = await fetch(`${API_BASE_URL}/chats/${chatId}/participants/${userId}`, {
      method: 'DELETE',
      headers: {
        'Authorization': token ? `Bearer ${token}` : ''
      },
      credentials: 'include'
    });

    if (!response.ok) {
      const contentType = response.headers.get('content-type') || '';
      if (contentType && contentType.includes('application/json')) {
        const errorData = await response.json();
        throw new Error(errorData.message || 'Failed to remove chat participant');
      } else {
        throw new Error(`Failed to remove chat participant: ${response.status} ${response.statusText}`);
      }
    }

    const contentType = response.headers.get('content-type') || '';
    if (contentType && contentType.includes('application/json')) {
      try {
        return await response.json();
      } catch (parseError: unknown) {
        logger.error(`Failed to parse JSON response for removing participant ${userId} from chat ${chatId}:`, parseError);
        return { success: true };
      }
    } else {
      logger.warn(`Non-JSON response for removing participant ${userId} from chat ${chatId}`);
      return { success: true };
    }
  } catch (error) {
    logger.error(`Remove participant ${userId} from chat ${chatId} failed:`, error);
    throw error;
  }
}

export async function sendMessage(chatId: string, data: Record<string, any>) {
  try {
    // TEMPORARY: Return mock data for testing UI
    logger.debug(`TESTING MODE: Returning mock send message response for chat ${chatId}`, { content: data.content });
    
    const mockMessage = {
      message_id: `msg-${Date.now()}`,
      message: {
        id: `msg-${Date.now()}`,
        chat_id: chatId,
        sender_id: "test-user-123",
        content: data.content,
        timestamp: Date.now() / 1000,
        is_read: false,
        is_edited: false,
        is_deleted: false,
      }
    };
    
    return mockMessage;

    const token = getAuthToken();
    logger.debug(`Sending message to chat ${chatId}`, { content: data.content });

    const response = await fetch(`${API_BASE_URL}/chats/${chatId}/messages`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': token ? `Bearer ${token}` : ''
      },
      body: JSON.stringify(data),
      credentials: 'include'
    });

    if (!response.ok) {
      const contentType = response.headers.get('content-type') || '';
      if (contentType && contentType.includes('application/json')) {
        try {
          const errorData = await response.json();
          throw new Error(errorData.message || 'Failed to send message');
        } catch (parseError) {
          throw new Error(`Failed to send message: ${response.status} ${response.statusText}`);
        }
      } else {
        throw new Error(`Failed to send message: ${response.status} ${response.statusText}`);
      }
    }

    const responseText = await response.text();
    if (!responseText || responseText.trim() === '') {
      logger.warn(`Empty response for sending message to chat ${chatId}`);
      return { 
        message: { 
          id: data.message_id || 'temp-' + Date.now(),
          chat_id: chatId,
          original_id: data.message_id,
          timestamp: Date.now() / 1000
        } 
      };
    }

    try {
      const responseData = JSON.parse(responseText);
      logger.debug(`Message sent successfully to chat ${chatId}`, { 
        messageId: responseData.message?.id || responseData.message?.message_id,
        responseData
      });
      return responseData;
    } catch (parseError) {
      logger.error(`Failed to parse response for chat ${chatId}:`, parseError);
      return { 
        success: true,
        message: { 
          id: data.message_id || 'temp-' + Date.now(),
          chat_id: chatId,
          content: data.content,
          timestamp: Date.now() / 1000,
          original_id: data.message_id
        } 
      };
    }
  } catch (error) {
    logger.error(`Send message to chat ${chatId} failed:`, error);
    throw error;
  }
}

export async function listMessages(chatId: string) {
  try {
    if (!chatId) {
      logger.error('Cannot list messages: Chat ID is undefined or empty');
      throw new Error('Invalid chat ID: Chat ID is required');
    }

    const token = getAuthToken();
    logger.debug(`Fetching messages for chat ${chatId}`);

    // Log API URL to ensure it's correct
    logger.info(`API URL for messages: ${API_BASE_URL}/chats/${chatId}/messages`);

    const response = await fetch(`${API_BASE_URL}/chats/${chatId}/messages`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Accept': 'application/json',
        'Authorization': token ? `Bearer ${token}` : ''
      },
      credentials: 'include'
    });

    // Log the response status
    logger.info(`Messages response status: ${response.status} ${response.statusText}`);
    
    // Log response headers to debug content-type issues
    const contentTypeHeader = response.headers.get('content-type');
    logger.info(`Response content type: ${contentTypeHeader || 'not provided'}`);

    if (!response.ok) {
      const contentType = response.headers.get('content-type') || '';
      if (contentType && contentType.includes('application/json')) {
        try {
          const errorData = await response.json();
          throw new Error(errorData.message || 'Failed to list messages');
        } catch (parseError) {
          throw new Error(`Failed to list messages: ${response.status} ${response.statusText}`);
        }
      } else {
        throw new Error(`Failed to list messages: ${response.status} ${response.statusText}`);
      }
    }

    // Check if the response is empty
    const contentLength = response.headers.get('content-length');
    if (contentLength === '0') {
      logger.warn(`Empty response received from messages endpoint for chat ${chatId}`);
      return { messages: [] };
    }

    try {
      const responseText = await response.text();
      if (!responseText || responseText.trim() === '') {
        logger.warn(`Empty response body from messages endpoint for chat ${chatId}`);
        return { messages: [] };
      }
      
      try {
        const responseData = JSON.parse(responseText);
        // Log the shape of the response data
        logger.debug('API messages response structure:', {
          hasData: !!responseData,
          hasMessages: responseData && 'messages' in responseData,
          isMessagesArray: responseData && 'messages' in responseData && Array.isArray(responseData.messages),
          messagesLength: responseData?.messages?.length || 0,
          responseKeys: responseData ? Object.keys(responseData) : []
        });
        
        // Handle different response formats
        if (responseData && 'messages' in responseData) {
          return responseData;
        } else if (Array.isArray(responseData)) {
          return { messages: responseData };
        } else if (responseData && typeof responseData === 'object') {
          // If the response is an object but doesn't have a messages property,
          // it might be a single message or an empty object
          if (responseData.id && responseData.content) {
            return { messages: [responseData] };
          }
          return { messages: [] };
        } else {
          return { messages: [] };
        }
      } catch (parseError: unknown) {
        logger.error(`Failed to parse JSON response for chat ${chatId}:`, parseError);
        // Try to log the raw response for debugging
        logger.error('Raw response text:', responseText.substring(0, 200) + '...');
        return { messages: [] };
      }
    } catch (textError) {
      logger.error(`Could not read response text for chat ${chatId}:`, textError);
      return { messages: [] };
    }
  } catch (error) {
    logger.error(`List messages for chat ${chatId} failed:`, error);
    throw error;
  }
}

export async function deleteMessage(chatId: string, messageId: string) {
  try {
    const token = getAuthToken();

    const response = await fetch(`${API_BASE_URL}/chats/${chatId}/messages/${messageId}`, {
      method: 'DELETE',
      headers: {
        'Authorization': token ? `Bearer ${token}` : ''
      },
      credentials: 'include'
    });

    if (!response.ok) {
      const contentType = response.headers.get('content-type') || '';
      if (contentType && contentType.includes('application/json')) {
        const errorData = await response.json();
        throw new Error(errorData.message || 'Failed to delete message');
      } else {
        throw new Error(`Failed to delete message: ${response.status} ${response.statusText}`);
      }
    }

    const contentType = response.headers.get('content-type') || '';
    if (contentType && contentType.includes('application/json')) {
      try {
        return await response.json();
      } catch (parseError: unknown) {
        logger.error(`Failed to parse JSON response for deleting message ${messageId} in chat ${chatId}:`, parseError);
        return { success: true };
      }
    } else {
      logger.warn(`Non-JSON response for deleting message ${messageId} in chat ${chatId}`);
      return { success: true };
    }
  } catch (error) {
    logger.error(`Delete message ${messageId} from chat ${chatId} failed:`, error);
    throw error;
  }
}

export async function unsendMessage(chatId: string, messageId: string) {
  try {
    const token = getAuthToken();

    const response = await fetch(`${API_BASE_URL}/chats/${chatId}/messages/${messageId}/unsend`, {
      method: 'POST',
      headers: {
        'Authorization': token ? `Bearer ${token}` : ''
      },
      credentials: 'include'
    });

    if (!response.ok) {
      const contentType = response.headers.get('content-type') || '';
      if (contentType && contentType.includes('application/json')) {
        const errorData = await response.json();
        throw new Error(errorData.message || 'Failed to unsend message');
      } else {
        throw new Error(`Failed to unsend message: ${response.status} ${response.statusText}`);
      }
    }

    const contentType = response.headers.get('content-type') || '';
    if (contentType && contentType.includes('application/json')) {
      try {
        return await response.json();
      } catch (parseError: unknown) {
        logger.error(`Failed to parse JSON response for unsending message ${messageId} in chat ${chatId}:`, parseError);
        return { success: true };
      }
    } else {
      logger.warn(`Non-JSON response for unsending message ${messageId} in chat ${chatId}`);
      return { success: true };
    }
  } catch (error) {
    logger.error(`Unsend message ${messageId} in chat ${chatId} failed:`, error);
    throw error;
  }
}

export async function searchMessages(chatId: string, query: string) {
  try {
    const token = getAuthToken();

    const response = await fetch(`${API_BASE_URL}/chats/${chatId}/messages/search?query=${encodeURIComponent(query)}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': token ? `Bearer ${token}` : ''
      },
      credentials: 'include'
    });

    if (!response.ok) {
      const contentType = response.headers.get('content-type') || '';
      if (contentType && contentType.includes('application/json')) {
        const errorData = await response.json();
        throw new Error(errorData.message || 'Failed to search messages');
      } else {
        throw new Error(`Failed to search messages: ${response.status} ${response.statusText}`);
      }
    }

    const contentType = response.headers.get('content-type') || '';
    if (contentType && contentType.includes('application/json')) {
      try {
        return await response.json();
      } catch (parseError: unknown) {
        logger.error(`Failed to parse JSON response for searching messages in chat ${chatId}:`, parseError);
        return { messages: [] };
      }
    } else {
      logger.warn(`Non-JSON response for searching messages in chat ${chatId}`);
      return { messages: [] };
    }
  } catch (error) {
    logger.error(`Search messages in chat ${chatId} failed:`, error);
    throw error;
  }
}

export async function getChatHistoryList() {
  try {
    const token = getAuthToken();
    logger.debug('Fetching chat history list');

    const response = await fetch(`${API_BASE_URL}/chats/history`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': token ? `Bearer ${token}` : ''
      },
      credentials: 'include'
    });

    logger.debug('Chat history response status', { 
      status: response.status, 
      statusText: response.statusText
    });

    if (!response.ok) {
      const contentType = response.headers.get('content-type') || '';
      if (contentType && contentType.includes('application/json')) {
        const errorData = await response.json();
        logger.error('Error response from chat history endpoint', errorData);
        throw new Error(errorData.message || 'Failed to get chat history');
      } else {
        throw new Error(`Failed to get chat history: ${response.status} ${response.statusText}`);
      }
    }

    const contentType = response.headers.get('content-type') || '';
    if (contentType && contentType.includes('application/json')) {
      try {
        const data = await response.json();
        logger.debug('Chat history response data', { 
          hasSuccess: 'success' in data,
          hasChats: 'chats' in data,
          chatsIsArray: Array.isArray(data.chats),
          chatsLength: data.chats ? data.chats.length : 0,
          data: data
        });
        return data;
      } catch (parseError: unknown) {
        logger.error('Failed to parse JSON response for getting chat history:', parseError);
        return { chats: [] };
      }
    } else {
      logger.warn('Non-JSON response for getting chat history');
      return { chats: [] };
    }
  } catch (error) {
    logger.error('Get chat history failed:', error);
    throw error;
  }
}

// Adding API test function to check server connection
export async function testApiConnection() {
  try {
    logger.debug('Testing API connection');
    
    // Check API URL format
    const baseUrlComponents = API_BASE_URL.split('/');
    if (baseUrlComponents.length < 3) {
      logger.error('Invalid API URL format:', API_BASE_URL);
      return { success: false, error: 'Invalid API URL format', url: API_BASE_URL };
    }
    
    const protocol = baseUrlComponents[0];
    const host = baseUrlComponents[2];
    
    logger.debug(`API Protocol: ${protocol}, Host: ${host}`);
    
    // First check if the server is reachable with a basic request
    try {
      const basicResponse = await fetch(`${protocol}//${host}/`, {
        method: 'GET',
        headers: {
          'Accept': 'text/html,application/json'
        }
      });
      
      logger.debug(`Basic server response: ${basicResponse.status}`);
    } catch (err) {
      logger.warn('Basic connectivity check failed:', err);
      // Continue with other tests even if this fails
    }
    
    // Try the authenticated chats endpoint first (most important for this page)
    const token = getAuthToken();
    
    if (token) {
      try {
        logger.debug('Testing authenticated chats endpoint');
        const chatResponse = await fetch(`${API_BASE_URL}/chats`, {
          method: 'GET',
          headers: {
            'Content-Type': 'application/json',
            'Accept': 'application/json',
            'Authorization': `Bearer ${token}`
          }
        });
        
        logger.debug(`Chats endpoint response: ${chatResponse.status}`);
        
        if (chatResponse.ok) {
          // Successfully tested the endpoint we actually need
          return {
            success: true,
            status: chatResponse.status,
            endpoint: '/chats',
            authenticated: true
          };
        }
      } catch (chatErr) {
        logger.debug('Failed to test authenticated chats endpoint:', chatErr);
        // Continue to try other endpoints
      }
    }
    
    // Define endpoints to try in order
    const endpointsToTry = [
      '/api/v1/chats',     // Try chats endpoint first (what we actually need)
      '/api/v1/users/me',  // Try user profile endpoint next
      '/api/v1/trends',    // Try trends endpoint
      '/api/v1/health',    // Try health endpoint
      '/api/v1'            // Try base API path last
    ];
    
    let successful = false;
    let status = 0;
    let responseData = null;
    let errorMessage = '';
    let testedEndpoint = '';
    
    // Try each endpoint until one works
    for (const endpoint of endpointsToTry) {
      try {
        logger.debug(`Testing API endpoint: ${endpoint}`);
        const apiResponse = await fetch(`${protocol}//${host}${endpoint}`, {
          method: 'GET',
          headers: {
            'Content-Type': 'application/json',
            'Accept': 'application/json',
            'Authorization': token ? `Bearer ${token}` : ''
          }
        });
        
        status = apiResponse.status;
        logger.debug(`API endpoint ${endpoint} response: ${status}`);
        
        // For the base /api/v1 endpoint, a 404 is expected since it's not implemented directly
        if ((endpoint === '/api/v1' && status === 404) || apiResponse.ok) {
          successful = true;
          testedEndpoint = endpoint;
          
          // If we got a successful response, try to parse it
          if (apiResponse.ok) {
            try {
              responseData = await apiResponse.json();
              logger.debug(`API response data from ${endpoint}:`, responseData);
            } catch (e) {
              logger.debug(`Could not parse JSON from ${endpoint} response`);
            }
            // Found a working endpoint, stop trying more
            break;
          }
        } else {
          errorMessage = apiResponse.statusText;
        }
      } catch (endpointErr) {
        logger.debug(`Failed to connect to ${endpoint}:`, endpointErr);
      }
    }
    
    if (successful) {
      return {
        success: true,
        status,
        endpoint: testedEndpoint,
        data: responseData
      };
    } else {
      return {
        success: false,
        error: errorMessage || 'Could not connect to any API endpoint',
        testedEndpoints: endpointsToTry
      };
    }
  } catch (error) {
    logger.error('API connection test failed:', error);
    return {
      success: false,
      error: error instanceof Error ? error.message : 'Unknown error'
    };
  }
}

// Adding a function to help diagnose auth token issues
export function logAuthTokenInfo() {
  try {
    const token = getAuthToken();
    if (!token) {
      logger.error('No auth token found');
      return { success: false, error: 'No auth token available' };
    }
    
    logger.info('Auth token found, checking format');
    
    // Check token format
    const parts = token.split('.');
    if (parts.length !== 3) {
      logger.error('Invalid JWT format - expected 3 parts separated by periods');
      return { success: false, error: 'Invalid JWT format', tokenLength: token.length };
    }
    
    // Check header
    try {
      const headerJson = atob(parts[0]);
      const header = JSON.parse(headerJson);
      logger.debug('Token header:', header);
      
      if (!header.alg) {
        logger.warn('Token header missing algorithm');
      }
      
      if (!header.typ) {
        logger.warn('Token header missing type');
      }
    } catch (e) {
      logger.error('Failed to parse token header:', e);
    }
    
    // Check payload
    try {
      const base64Url = parts[1];
      const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
      const payloadJson = decodeURIComponent(atob(base64).split('').map(c => {
        return '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2);
      }).join(''));
      
      const payload = JSON.parse(payloadJson);
      
      // Log useful information but redact sensitive data
      const safePayload = { ...payload };
      delete safePayload.password;
      delete safePayload.secret;
      
      logger.info('Token payload:', safePayload);
      
      // Check for important claims
      if (!payload.exp) {
        logger.warn('Token missing expiration (exp) claim');
      } else {
        const expiry = new Date(payload.exp * 1000);
        const now = new Date();
        
        if (expiry < now) {
          logger.error(`Token expired at ${expiry.toISOString()}`);
        } else {
          logger.info(`Token valid until ${expiry.toISOString()} (${Math.floor((expiry.getTime() - now.getTime()) / 60000)} minutes)`);
        }
      }
      
      if (!payload.sub && !payload.user_id) {
        logger.warn('Token missing subject (sub) or user_id claim');
      }
      
      return {
        success: true,
        isExpired: payload.exp ? new Date(payload.exp * 1000) < new Date() : null,
        userId: payload.sub || payload.user_id,
        expiresAt: payload.exp ? new Date(payload.exp * 1000).toISOString() : null,
        issuer: payload.iss || null
      };
    } catch (e) {
      logger.error('Failed to parse token payload:', e);
      return { success: false, error: 'Failed to parse token payload' };
    }
  } catch (e) {
    logger.error('Error checking auth token:', e);
    return { success: false, error: 'Error analyzing token' };
  }
}