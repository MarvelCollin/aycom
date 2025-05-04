import { apiRequest } from '../utils/apiClient';
import { getAuthToken } from '../utils/auth';
import appConfig from '../config/appConfig';
import { createLoggerWithPrefix } from '../utils/logger';

const API_BASE_URL = appConfig.api.baseUrl;
const logger = createLoggerWithPrefix('ThreadAPI');

export async function createThread(data: Record<string, any>) {
  try {
    const response = await apiRequest('/threads', {
      method: "POST",
      body: JSON.stringify(data)
    });
    
    if (!response.ok) {
      try {
        const errorData = await response.json();
        throw new Error(
          errorData.message || 
          errorData.error?.message || 
          `Failed to create thread: ${response.status} ${response.statusText}`
        );
      } catch (parseError) {
        throw new Error(`Failed to create thread: ${response.status} ${response.statusText}`);
      }
    }
    
    return response.json();
  } catch (error) {
    logger.error('Create thread failed:', error);
    throw error;
  }
}

export async function getThread(id: string) {
  try {
    const response = await apiRequest(`/threads/${id}`, {
      method: "GET"
    });
    
    if (!response.ok) {
      try {
        const errorData = await response.json();
        throw new Error(
          errorData.message || 
          errorData.error?.message || 
          `Failed to fetch thread: ${response.status} ${response.statusText}`
        );
      } catch (parseError) {
        throw new Error(`Failed to fetch thread: ${response.status} ${response.statusText}`);
      }
    }
    
    return response.json();
  } catch (error) {
    logger.error(`Get thread ${id} failed:`, error);
    throw error;
  }
}

export async function getThreadsByUser(userId: string) {
  try {
    // Always use /threads/user/{id} format, even when userId is 'me'
    // The backend will interpret 'me' appropriately
    const endpoint = `/threads/user/${userId}`;
      
    logger.debug(`Fetching threads using endpoint: ${endpoint}`);
    
    const response = await apiRequest(endpoint, {
      method: "GET"
    });
    
    if (!response.ok) {
      try {
        const errorData = await response.json();
        throw new Error(
          errorData.message || 
          errorData.error?.message || 
          `Failed to fetch user's threads: ${response.status} ${response.statusText}`
        );
      } catch (parseError) {
        throw new Error(`Failed to fetch user's threads: ${response.status} ${response.statusText}`);
      }
    }
    
    return response.json();
  } catch (error) {
    logger.error(`Get threads for user ${userId} failed:`, error);
    throw error;
  }
}

export async function updateThread(id: string, data: Record<string, any>) {
  try {
    const response = await apiRequest(`/threads/${id}`, {
      method: "PUT",
      body: JSON.stringify(data)
    });
    
    if (!response.ok) {
      try {
        const errorData = await response.json();
        throw new Error(
          errorData.message || 
          errorData.error?.message || 
          `Failed to update thread: ${response.status} ${response.statusText}`
        );
      } catch (parseError) {
        throw new Error(`Failed to update thread: ${response.status} ${response.statusText}`);
      }
    }
    
    return response.json();
  } catch (error) {
    logger.error(`Update thread ${id} failed:`, error);
    throw error;
  }
}

export async function deleteThread(id: string) {
  try {
    const response = await apiRequest(`/threads/${id}`, {
      method: "DELETE"
    });
    
    if (!response.ok) {
      try {
        const errorData = await response.json();
        throw new Error(
          errorData.message || 
          errorData.error?.message || 
          `Failed to delete thread: ${response.status} ${response.statusText}`
        );
      } catch (parseError) {
        throw new Error(`Failed to delete thread: ${response.status} ${response.statusText}`);
      }
    }
    
    return response.json();
  } catch (error) {
    logger.error(`Delete thread ${id} failed:`, error);
    throw error;
  }
}

export async function uploadThreadMedia(threadId: string, files: File[]) {
  try {
    const formData = new FormData();
    formData.append('thread_id', threadId);
    
    files.forEach((file, index) => {
      formData.append(`media_${index}`, file);
    });
    
    // For FormData we need to omit the Content-Type header
    const response = await apiRequest('/threads/media', {
      method: 'POST',
      body: formData
    });
    
    if (!response.ok) {
      try {
        const errorData = await response.json();
        throw new Error(
          errorData.message || 
          errorData.error?.message || 
          `Failed to upload media: ${response.status} ${response.statusText}`
        );
      } catch (parseError) {
        throw new Error(`Failed to upload media: ${response.status} ${response.statusText}`);
      }
    }
    
    return response.json();
  } catch (error) {
    logger.error('Upload thread media failed:', error);
    throw error;
  }
}