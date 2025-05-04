import { apiRequest } from '../utils/apiClient';
import { createLoggerWithPrefix } from '../utils/logger';

const logger = createLoggerWithPrefix('ChatAPI');

export async function createChat(data: Record<string, any>) {
  try {
    const response = await apiRequest('/chats', {
      method: 'POST',
      body: JSON.stringify(data)
    });
    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.message || 'Failed to create chat');
    }
    return response.json();
  } catch (error) {
    logger.error('Create chat failed:', error);
    throw error;
  }
}

export async function listChats() {
  try {
    const response = await apiRequest('/chats', { method: 'GET' });
    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.message || 'Failed to list chats');
    }
    return response.json();
  } catch (error) {
    logger.error('List chats failed:', error);
    throw error;
  }
}

export async function listChatParticipants(chatId: string) {
  try {
    const response = await apiRequest(`/chats/${chatId}/participants`, { method: 'GET' });
    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.message || 'Failed to list chat participants');
    }
    return response.json();
  } catch (error) {
    logger.error(`List participants for chat ${chatId} failed:`, error);
    throw error;
  }
}

export async function addChatParticipant(chatId: string, data: Record<string, any>) {
  try {
    const response = await apiRequest(`/chats/${chatId}/participants`, {
      method: 'POST',
      body: JSON.stringify(data)
    });
    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.message || 'Failed to add chat participant');
    }
    return response.json();
  } catch (error) {
    logger.error(`Add participant to chat ${chatId} failed:`, error);
    throw error;
  }
}

export async function removeChatParticipant(chatId: string, userId: string) {
  try {
    const response = await apiRequest(`/chats/${chatId}/participants/${userId}`, { method: 'DELETE' });
    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.message || 'Failed to remove chat participant');
    }
    return response.json();
  } catch (error) {
    logger.error(`Remove participant ${userId} from chat ${chatId} failed:`, error);
    throw error;
  }
}

export async function sendMessage(chatId: string, data: Record<string, any>) {
  try {
    const response = await apiRequest(`/chats/${chatId}/messages`, {
      method: 'POST',
      body: JSON.stringify(data)
    });
    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.message || 'Failed to send message');
    }
    return response.json();
  } catch (error) {
    logger.error(`Send message to chat ${chatId} failed:`, error);
    throw error;
  }
}

export async function listMessages(chatId: string) {
  try {
    const response = await apiRequest(`/chats/${chatId}/messages`, { method: 'GET' });
    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.message || 'Failed to list messages');
    }
    return response.json();
  } catch (error) {
    logger.error(`List messages for chat ${chatId} failed:`, error);
    throw error;
  }
}

export async function deleteMessage(chatId: string, messageId: string) {
  try {
    const response = await apiRequest(`/chats/${chatId}/messages/${messageId}`, { method: 'DELETE' });
    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.message || 'Failed to delete message');
    }
    return response.json();
  } catch (error) {
    logger.error(`Delete message ${messageId} from chat ${chatId} failed:`, error);
    throw error;
  }
}

export async function unsendMessage(chatId: string, messageId: string) {
  try {
    const response = await apiRequest(`/chats/${chatId}/messages/${messageId}/unsend`, { method: 'POST' });
    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.message || 'Failed to unsend message');
    }
    return response.json();
  } catch (error) {
    logger.error(`Unsend message ${messageId} in chat ${chatId} failed:`, error);
    throw error;
  }
}

export async function searchMessages(chatId: string, query: string) {
  try {
    const response = await apiRequest(`/chats/${chatId}/messages/search?query=${encodeURIComponent(query)}`, { method: 'GET' });
    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.message || 'Failed to search messages');
    }
    return response.json();
  } catch (error) {
    logger.error(`Search messages in chat ${chatId} failed:`, error);
    throw error;
  }
}
