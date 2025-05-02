import { getAuthToken } from '../utils/auth';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8081';

export async function createThread(data: Record<string, any>) {
  try {
    console.log("Creating thread with data:", data);
    
    // Get the token from auth utility
    const token = getAuthToken();
    console.log("Found auth token:", !!token);
    
    const response = await fetch(`${API_BASE_URL}/threads`, {
      method: "POST",
      headers: { 
        "Content-Type": "application/json",
        "Authorization": token ? `Bearer ${token}` : ''
      },
      body: JSON.stringify(data),
      credentials: "include",
    });
    
    console.log("Thread creation response status:", response.status);
    
    if (!response.ok) {
      // Try to extract detailed error information
      try {
        const errorData = await response.json();
        console.error("Thread creation error response:", errorData);
        throw new Error(
          errorData.message || 
          errorData.error?.message || 
          `Failed to create thread: ${response.status} ${response.statusText}`
        );
      } catch (parseError) {
        console.error("Failed to parse error response:", parseError);
        throw new Error(`Failed to create thread: ${response.status} ${response.statusText}`);
      }
    }
    
    const result = await response.json();
    console.log("Thread creation successful:", result);
    return result;
  } catch (error) {
    console.error("Thread creation exception:", error);
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
  
  // Add each file to the form data
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