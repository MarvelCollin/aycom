import { createLoggerWithPrefix } from '../utils/logger';
import appConfig from '../config/appConfig';
import { getAuthToken, getUserId } from '../utils/auth';

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
        temp_id: tempId,
        type: 'text'
      });
    } catch (handlerError) {
      logger.error('Error in message handler:', handlerError);
    }
  }
}

export async function createChat(data: Record<string, any>) {
  try {
    if (!data || (!data.participants && !data.name)) {
      logger.error('Cannot create chat: Missing required data');
      throw new Error('Missing required data: participants or name is required to create a chat');
    }
    
    // Validate participant IDs if present
    if (data.participants && Array.isArray(data.participants)) {
      for (const participantId of data.participants) {
        if (typeof participantId === 'string' && !/^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i.test(participantId)) {
          logger.error(`Cannot create chat: Invalid UUID format for participant ID: ${participantId}`);
          throw new Error(`Invalid participant ID format: ${participantId} must be a valid UUID`);
        }
      }
    }

    const token = getAuthToken();
    logger.debug('Creating chat with data', { data, apiUrl: `${API_BASE_URL}/chats` });

    // Check for individual chat type and see if chat already exists
    if (data.type === 'individual' && data.participants && data.participants.length === 1) {
      const participantId = data.participants[0];
      const currentUserId = getUserId();

      try {
        // Get existing chats
        const existingChats = await getChatHistoryList();
        logger.debug('Checking existing chats for participant', { 
          participantId, 
          currentUserId,
          chatsCount: existingChats.chats?.length || 0 
        });

        // Look for an existing individual chat with exactly these two participants
        const existingChat = existingChats.chats?.find(chat => {
          // Skip group chats
          if (chat.is_group_chat === true) {
            return false;
          }

          // Check if participants array exists
          if (!chat.participants || !Array.isArray(chat.participants)) {
            return false;
          }

          // For individual chats, we should have exactly 2 participants
          if (chat.participants.length !== 2) {
            return false;
          }

          // Check if the participants are the current user and the target participant
          const hasCurrentUser = chat.participants.some(p => 
            (p.id === currentUserId || p.user_id === currentUserId)
          );
          
          const hasTargetUser = chat.participants.some(p => 
            (p.id === participantId || p.user_id === participantId)
          );
          
          // Return true if both users are in this chat
          return hasCurrentUser && hasTargetUser;
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

    // Send the request to create a new chat
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
      } catch (parseError) {
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
    // Validate chat ID first
    if (!chatId || chatId.trim() === '') {
      logger.error('Cannot send message: Invalid or missing chat ID');
      throw new Error('Invalid chat ID: A valid chat ID is required to send messages');
    }
    
    // Validate that the chat ID is a valid UUID
    if (!/^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i.test(chatId)) {
      logger.error(`Cannot send message: Invalid UUID format for chat ID: ${chatId}`);
      throw new Error('Invalid chat ID format: Must be a valid UUID');
    }

    // Validate user ID before sending
    const currentUserId = getUserId();
    if (!currentUserId || !/^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i.test(currentUserId)) {
      logger.error(`Cannot send message: Invalid or missing user ID: ${currentUserId}`);
      throw new Error('Invalid user ID: Please try logging in again');
    }

    // Send the message to the API
    const token = getAuthToken();
    logger.debug(`Sending message to chat ${chatId}`, { content: data.content });

    // Make sure we're sending the right format
    const messageData = {
      content: data.content,
      // Include any attachments if present
      ...(data.attachments && data.attachments.length > 0 ? { attachments: data.attachments } : {})
    };

    try {
    const response = await fetch(`${API_BASE_URL}/chats/${chatId}/messages`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': token ? `Bearer ${token}` : ''
      },
        body: JSON.stringify(messageData),
      credentials: 'include'
    });

    if (!response.ok) {
        let errorMessage = `Server error when sending message: ${response.status} ${response.statusText}`;
        logger.warn(errorMessage);
        
        // Try to get more detailed error information from the response
        try {
          const errorResponse = await response.json();
          if (errorResponse && errorResponse.error) {
            // Handle structured error responses
            if (typeof errorResponse.error === 'string') {
              errorMessage = errorResponse.error;
            } else if (errorResponse.error.message) {
              errorMessage = errorResponse.error.message;
            }
          } else if (errorResponse && errorResponse.message) {
            errorMessage = errorResponse.message;
          }
          
          logger.error(`API error details:`, errorResponse);
        } catch (parseError) {
          // If we can't parse the error response, just use the status code
          logger.error(`Could not parse error response: ${parseError}`);
        }

        // Handle specific error codes
        if (response.status === 404) {
          throw new Error('Chat not found. It may have been deleted or you may not have access.');
        } else if (response.status === 403) {
          throw new Error('You do not have permission to send messages in this chat.');
        } else if (response.status === 401) {
          throw new Error('Authentication required. Please log in again.');
        } else if (response.status === 400) {
          throw new Error(`Invalid message: ${errorMessage}`);
        } else if (response.status === 500) {
          // For server errors related to invalid user ID, suggest logging in again
          if (errorMessage.includes('user ID') || errorMessage.includes('UUID')) {
            logger.error(`User ID validation error detected: ${errorMessage}`);
            throw new Error('Session error: Please log out and log in again.');
          }
          
          // Use a fallback for other server errors
          logger.warn(`Using local fallback due to server error: ${errorMessage}`);
          return createFallbackMessage(chatId, data);
      } else {
          throw new Error(errorMessage);
      }
    }

      // Parse the response
    const responseText = await response.text();
    if (!responseText || responseText.trim() === '') {
      logger.warn(`Empty response for sending message to chat ${chatId}`);
        return createFallbackMessage(chatId, data);
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
        return createFallbackMessage(chatId, data);
      }
    } catch (apiError) {
      // Handle network errors or other exceptions
      if (apiError instanceof Error) {
        logger.error(`API error when sending message to ${chatId}:`, apiError);
        throw apiError;
      } else {
        logger.error(`Unknown API error when sending message to ${chatId}:`, apiError);
        throw new Error('Failed to send message due to a network error');
      }
    }
  } catch (error) {
    logger.error(`Send message to chat ${chatId} failed:`, error);
    throw error;
  }
}

// Helper function to create a fallback message when the API fails
function createFallbackMessage(chatId: string, data: Record<string, any>) {
  const currentUserId = getUserId();
  const messageTimestamp = Date.now();
  const tempMessageId = `temp-${messageTimestamp}`;
  
  // Log the user ID for debugging
  logger.debug(`Creating fallback message with user ID: ${currentUserId}`);
  
  // Generate a proper UUID for sender_id if needed
  let senderId = "";
  if (currentUserId) {
    // Check if user ID is already a valid UUID
    const isValidUUID = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i.test(currentUserId);
    
    if (isValidUUID) {
      senderId = currentUserId;
    } else {
      // If not a valid UUID, log it and use a temporary formatted value
      logger.warn(`Invalid UUID format detected for user ID: ${currentUserId}, using temp ID`);
      // We'll use a placeholder formatted as 'local-{timestamp}' for client-side only
      senderId = `local-${messageTimestamp}`;
    }
  } else {
    // If there's no user ID at all, use a temporary ID
    logger.warn("No user ID available, using temp ID");
    senderId = `local-${messageTimestamp}`;
  }
  
  return {
    message_id: tempMessageId,
    message: {
      id: tempMessageId,
      chat_id: chatId,
      sender_id: senderId,
      content: data.content,
      timestamp: messageTimestamp / 1000,
      is_read: false,
      is_edited: false,
      is_deleted: false,
      is_local: true, // Mark this as a local message that hasn't been saved to the server
    }
  };
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
          logger.error('Error response from message list endpoint', errorData);
          
          // Check for standard error formats
          if (errorData.error && errorData.error.message) {
            // Format: { success: false, error: { code: 'ERROR_CODE', message: 'Error message' } }
            throw new Error(errorData.error.message);
          } else if (errorData.message) {
            // Format: { message: 'Error message' }
            throw new Error(errorData.message);
          } else if (errorData.error && typeof errorData.error === 'string') {
            // Format: { error: 'Error message' }
            throw new Error(errorData.error);
          } else if (errorData.success === false) {
            // Format: { success: false, ... }
            throw new Error('Failed to list messages: Request failed');
          } else {
            throw new Error('Failed to list messages: Unknown error');
          }
        } catch (parseError) {
          if (parseError instanceof Error && parseError.message !== 'Failed to list messages: Unknown error') {
            throw parseError; // Re-throw if it's our custom error with message
          }
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
          hasSuccess: responseData && 'success' in responseData,
          hasMessages: responseData && 'messages' in responseData,
          hasDataObject: responseData && 'data' in responseData,
          isMessagesArray: responseData && 'messages' in responseData && Array.isArray(responseData.messages),
          messagesLength: responseData?.messages?.length || 0,
          responseKeys: responseData ? Object.keys(responseData) : []
        });
        
        // Handle different response formats
        if (responseData && responseData.success && responseData.data && responseData.data.messages) {
          // Format: { success: true, data: { messages: [...] } }
          return responseData.data;
        } else if (responseData && 'messages' in responseData) {
          // Format: { messages: [...] }
          return responseData;
        } else if (responseData && responseData.data && Array.isArray(responseData.data)) {
          // Format: { data: [...] } - array might be messages
          return { messages: responseData.data };
        } else if (Array.isArray(responseData)) {
          // Format: [...] - direct array of messages
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
        
        // Enhanced logging to debug response structure
        logger.debug('API response structure:', { 
          hasData: !!data,
          hasSuccess: data && 'success' in data,
          hasChats: data && 'chats' in data,
          hasDataChats: data && data.data && 'chats' in data.data,
          isChatsArray: data && 'chats' in data && Array.isArray(data.chats),
          isDataChatsArray: data && data.data && 'chats' in data.data && Array.isArray(data.data.chats),
          chatsLength: data?.chats?.length || (data?.data?.chats?.length) || 0,
          responseKeys: data ? Object.keys(data) : []
        });
        
        // Handle different response formats
        if (data && data.data && data.data.chats) {
          // Format: { success: true, data: { chats: [...] } }
          return {
            chats: data.data.chats
          };
        } else if (data && data.chats) {
          // Format: { chats: [...] }
        return data;
        } else if (data && Array.isArray(data)) {
          // Format: [...]
          return {
            chats: data
          };
        } else {
          // Default empty response
          logger.warn('Unexpected response format from chat history endpoint', data);
          return { chats: [] };
        }
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
        // Also prioritize the /chats endpoint even if others work
        if (endpoint === '/api/v1/chats' && apiResponse.ok) {
          successful = true;
          testedEndpoint = endpoint;
          
          try {
            responseData = await apiResponse.json();
            logger.debug(`API response data from ${endpoint}:`, responseData);
          } catch (e) {
            logger.debug(`Could not parse JSON from ${endpoint} response`);
          }
          
          // If we found the chats endpoint working, prioritize it
          break;
        }
        else if ((endpoint === '/api/v1' && status === 404) || apiResponse.ok) {
          successful = true;
          testedEndpoint = endpoint;
          
          if (apiResponse.ok) {
            try {
              responseData = await apiResponse.json();
              logger.debug(`API response data from ${endpoint}:`, responseData);
            } catch (e) {
              logger.debug(`Could not parse JSON from ${endpoint} response`);
            }
          }
          
          // Don't break here if it's the base endpoint with 404
          // Keep looking for a better endpoint unless it's the last one
          if (endpoint !== '/api/v1' || endpointsToTry.indexOf(endpoint) === endpointsToTry.length - 1) {
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

// Adding a function to join a chat as a participant
export async function joinChat(chatId: string) {
  try {
    const token = getAuthToken();
    logger.debug(`Joining chat ${chatId}`);

    const response = await fetch(`${API_BASE_URL}/chats/${chatId}/participants`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': token ? `Bearer ${token}` : ''
      },
      body: JSON.stringify({ is_admin: false }), // Join as a regular participant
      credentials: 'include'
    });

    if (!response.ok) {
      const contentType = response.headers.get('content-type') || '';
      if (contentType && contentType.includes('application/json')) {
        const errorData = await response.json();
        throw new Error(errorData.message || errorData.error?.message || 'Failed to join chat');
      } else {
        throw new Error(`Failed to join chat: ${response.status} ${response.statusText}`);
      }
    }

    const contentType = response.headers.get('content-type') || '';
    if (contentType && contentType.includes('application/json')) {
      try {
        return await response.json();
      } catch (parseError: unknown) {
        logger.error(`Failed to parse JSON response for joining chat ${chatId}:`, parseError);
        return { success: true };
      }
    } else {
      logger.warn(`Non-JSON response for joining chat ${chatId}`);
      return { success: true };
    }
  } catch (error) {
    logger.error(`Join chat ${chatId} failed:`, error);
    throw error;
  }
}