import { apiRequest } from '../utils/apiClient';
import { getAuthToken, getUserId } from '../utils/auth';
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
    // Replace 'me' with the actual user ID from auth
    let endpoint;
    
    if (userId === 'me') {
      const actualUserId = getUserId();
      if (actualUserId) {
        logger.debug(`Replacing 'me' with actual user ID: ${actualUserId}`);
        endpoint = `/threads/user/${actualUserId}`;
      } else {
        logger.error('User ID not found in auth data while trying to fetch threads');
        throw new Error('Authentication issue: User ID not found. Please log in again.');
      }
    } else {
      endpoint = `/threads/user/${userId}`;
    }
      
    logger.debug(`Fetching threads using endpoint: ${endpoint}`);
    
    const response = await apiRequest(endpoint, {
      method: "GET"
    });
    
    if (!response.ok) {
      try {
        const errorData = await response.json();
        logger.error(`Failed to fetch threads: ${response.status}`, errorData);
        throw new Error(
          errorData.message || 
          errorData.error?.message || 
          `Failed to fetch user's threads: ${response.status} ${response.statusText}`
        );
      } catch (parseError) {
        throw new Error(`Failed to fetch user's threads: ${response.status} ${response.statusText}`);
      }
    }
    
    const result = await response.json();
    logger.debug(`Successfully fetched ${result.data?.length || 0} threads`);
    return result;
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