import { getAuthToken } from '../utils/auth';
import appConfig from '../config/appConfig';

const API_BASE_URL = appConfig.api.baseUrl;

export async function createThread(data: Record<string, any>) {
  try {
    const token = getAuthToken();
    
    const response = await fetch(`${API_BASE_URL}/threads`, {
      method: "POST",
      headers: { 
        "Content-Type": "application/json",
        "Authorization": token ? `Bearer ${token}` : ''
      },
      body: JSON.stringify(data),
      credentials: "include",
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
    throw error;
  }
}

export async function getThread(id: string) {
  const token = getAuthToken();
  
  const response = await fetch(`${API_BASE_URL}/threads/${id}`, {
    method: "GET",
    headers: { 
      "Content-Type": "application/json",
      "Authorization": token ? `Bearer ${token}` : ''
    },
    credentials: "include",
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
}

export async function getThreadsByUser(userId: string) {
  const token = getAuthToken();
  
  const response = await fetch(`${API_BASE_URL}/threads/user/${userId}`, {
    method: "GET",
    headers: { 
      "Content-Type": "application/json",
      "Authorization": token ? `Bearer ${token}` : ''
    },
    credentials: "include",
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
}

export async function updateThread(id: string, data: Record<string, any>) {
  const token = getAuthToken();
  
  const response = await fetch(`${API_BASE_URL}/threads/${id}`, {
    method: "PUT",
    headers: { 
      "Content-Type": "application/json",
      "Authorization": token ? `Bearer ${token}` : ''
    },
    body: JSON.stringify(data),
    credentials: "include",
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
}

export async function deleteThread(id: string) {
  const token = getAuthToken();
  
  const response = await fetch(`${API_BASE_URL}/threads/${id}`, {
    method: "DELETE",
    headers: { 
      "Content-Type": "application/json",
      "Authorization": token ? `Bearer ${token}` : ''
    },
    credentials: "include",
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
}

export async function uploadThreadMedia(threadId: string, files: File[]) {
  const token = getAuthToken();
  
  const formData = new FormData();
  formData.append('thread_id', threadId);
  
  files.forEach((file, index) => {
    formData.append(`media_${index}`, file);
  });
  
  const response = await fetch(`${API_BASE_URL}/threads/media`, {
    method: 'POST',
    headers: {
      "Authorization": token ? `Bearer ${token}` : ''
    },
    body: formData,
    credentials: 'include',
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
} 