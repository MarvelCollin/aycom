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
  logger.debug('Processing WebSocket message', { messageType: message.type, messageId: message.message_id });

  if (!message || !message.type) {
    logger.error('Invalid message format', { message });
    return;
  }

  const normalizedMessage: any = {
    type: message.type,
    message_id: message.message_id || message.messageId || message.id,
    chat_id: message.chat_id || message.chatId,
    user_id: message.user_id || message.userId || message.sender_id || message.senderId,
    timestamp: message.timestamp || message.created_at || message.createdAt || new Date().toISOString()
  };

  switch (message.type) {
    case 'text':
      normalizedMessage.content = message.content || message.text || '';
      normalizedMessage.is_edited = message.is_edited || message.isEdited || false;
      normalizedMessage.is_deleted = message.is_deleted || message.isDeleted || false;
      normalizedMessage.attachments = message.attachments || message.media || [];
      break;

    case 'typing':
      normalizedMessage.is_typing = message.is_typing || message.isTyping || true;
      break;

    case 'read':
      normalizedMessage.last_read_id = message.last_read_id || message.lastReadId;
      break;

    case 'edit':
      normalizedMessage.content = message.content || message.text || '';
      normalizedMessage.original_message_id = message.original_message_id || message.originalMessageId;
      break;

    case 'delete':
      normalizedMessage.target_message_id = message.target_message_id || message.targetMessageId;
      break;
  }

  switch (normalizedMessage.type) {
    case 'text':
      const originalTempId = extractTempIdFromMessage(normalizedMessage);

      if (originalTempId && normalizedMessage.message_id && normalizedMessage.message_id !== originalTempId) {
        logger.debug('Updating temp message with server data', { 
          tempId: originalTempId, 
          serverId: normalizedMessage.message_id 
        });

        const timestamp = normalizedMessage.timestamp ? 
          (typeof normalizedMessage.timestamp === 'number' ? 
            new Date(normalizedMessage.timestamp * 1000) : new Date(normalizedMessage.timestamp)) 
          : new Date();

        if (messageHandler) {
          const updateMessage = {
            ...normalizedMessage,
            type: 'update',
            original_temp_id: originalTempId,
            timestamp
          };
          messageHandler(updateMessage);
        }
      } else {
        internalHandleIncomingMessage(normalizedMessage);
      }
      break;

    case 'typing':
    case 'read':
    case 'edit':
    case 'delete':
      internalHandleIncomingMessage(normalizedMessage);
      break;

    default:
      logger.warn('Unknown WebSocket message type', { type: normalizedMessage.type });
  }
}

function extractTempIdFromMessage(message: any): string | null {
  if (message.original_id && message.original_id.startsWith('temp-')) {
    return message.original_id;
  }

  if (message.client_id && message.client_id.startsWith('temp-')) {
    return message.client_id;
  }

  if (message.message_id && message.message_id.startsWith('temp-')) {
    return message.message_id;
  }

  return null;
}

function internalHandleIncomingMessage(message: any) {
  if (!message || !message.chat_id) {
    logger.error('Invalid message for processing', { message });
    return;
  }

  if (messageHandler) {
    messageHandler(message);
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
      let errorMessage = `Failed to create chat: ${response.status} ${response.statusText}`;
      try {
        // Try to read the response once
        const responseText = await response.text();
        try {
          // Try to parse as JSON
          const errorJson = JSON.parse(responseText);
          logger.error('Failed to create chat:', { 
            status: response.status, 
            errorJson,
            url: `${API_BASE_URL}/chats`,
            method: 'POST'
          });
          errorMessage += ` - ${JSON.stringify(errorJson)}`;
        } catch (jsonParseError) {
          // If not JSON, use as text
          logger.error('Failed to create chat:', { 
            status: response.status, 
            response: responseText,
            url: `${API_BASE_URL}/chats`,
            method: 'POST'
          });
          if (responseText) {
            errorMessage += ` - ${responseText}`;
          }
        }
      } catch (readError) {
        logger.error('Failed to read error response:', readError);
      }
      throw new Error(errorMessage);
    }

    let jsonResponse;
    try {
      jsonResponse = await response.json();
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

    const response = await fetch(`${API_BASE_URL}/chats`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': token ? `Bearer ${token}` : ''
      },
      credentials: 'include'
    });

    if (!response.ok) {
      const contentType = response.headers.get('content-type');
      if (contentType && contentType.includes('application/json')) {
        const errorData = await response.json();
        throw new Error(errorData.message || 'Failed to list chats');
      } else {
        throw new Error(`Failed to list chats: ${response.status} ${response.statusText}`);
      }
    }

    const contentType = response.headers.get('content-type');
    if (contentType && contentType.includes('application/json')) {
      try {
        return await response.json();
      } catch (parseError: unknown) {
        logger.error('Failed to parse JSON response for listing chats:', parseError);
        return { chats: [] };
      }
    } else {
      logger.warn('Non-JSON response for listing chats');
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
      const contentType = response.headers.get('content-type');
      if (contentType && contentType.includes('application/json')) {
        const errorData = await response.json();
        throw new Error(errorData.message || 'Failed to list chat participants');
      } else {
        throw new Error(`Failed to list chat participants: ${response.status} ${response.statusText}`);
      }
    }

    const contentType = response.headers.get('content-type');
    if (contentType && contentType.includes('application/json')) {
      try {
        return await response.json();
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
      const contentType = response.headers.get('content-type');
      if (contentType && contentType.includes('application/json')) {
        const errorData = await response.json();
        throw new Error(errorData.message || 'Failed to add chat participant');
      } else {
        throw new Error(`Failed to add chat participant: ${response.status} ${response.statusText}`);
      }
    }

    const contentType = response.headers.get('content-type');
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
      const contentType = response.headers.get('content-type');
      if (contentType && contentType.includes('application/json')) {
        const errorData = await response.json();
        throw new Error(errorData.message || 'Failed to remove chat participant');
      } else {
        throw new Error(`Failed to remove chat participant: ${response.status} ${response.statusText}`);
      }
    }

    const contentType = response.headers.get('content-type');
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
      const contentType = response.headers.get('content-type');
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

    const response = await fetch(`${API_BASE_URL}/chats/${chatId}/messages`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': token ? `Bearer ${token}` : ''
      },
      credentials: 'include'
    });

    if (!response.ok) {
      const contentType = response.headers.get('content-type');
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

    const contentType = response.headers.get('content-type');
    if (contentType && contentType.includes('application/json')) {
      try {
        const data = await response.json();
        logger.debug(`Received ${data.messages?.length || 0} messages for chat ${chatId}`);
        return data;
      } catch (parseError: unknown) {
        logger.error(`Failed to parse JSON response for chat ${chatId}:`, parseError);
        return { messages: [] };
      }
    } else {
      logger.warn(`Non-JSON response for chat ${chatId}`);
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
      const contentType = response.headers.get('content-type');
      if (contentType && contentType.includes('application/json')) {
        const errorData = await response.json();
        throw new Error(errorData.message || 'Failed to delete message');
      } else {
        throw new Error(`Failed to delete message: ${response.status} ${response.statusText}`);
      }
    }

    const contentType = response.headers.get('content-type');
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
      const contentType = response.headers.get('content-type');
      if (contentType && contentType.includes('application/json')) {
        const errorData = await response.json();
        throw new Error(errorData.message || 'Failed to unsend message');
      } else {
        throw new Error(`Failed to unsend message: ${response.status} ${response.statusText}`);
      }
    }

    const contentType = response.headers.get('content-type');
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
      const contentType = response.headers.get('content-type');
      if (contentType && contentType.includes('application/json')) {
        const errorData = await response.json();
        throw new Error(errorData.message || 'Failed to search messages');
      } else {
        throw new Error(`Failed to search messages: ${response.status} ${response.statusText}`);
      }
    }

    const contentType = response.headers.get('content-type');
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
      const contentType = response.headers.get('content-type');
      if (contentType && contentType.includes('application/json')) {
        const errorData = await response.json();
        logger.error('Error response from chat history endpoint', errorData);
        throw new Error(errorData.message || 'Failed to get chat history');
      } else {
        throw new Error(`Failed to get chat history: ${response.status} ${response.statusText}`);
      }
    }

    const contentType = response.headers.get('content-type');
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