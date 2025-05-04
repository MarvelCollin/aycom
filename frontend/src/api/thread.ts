import { getAuthToken, getUserId } from '../utils/auth';
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
  
  try {
    // If userId is 'me', get the actual user ID from auth
    if (userId === 'me') {
      const currentUserId = getUserId();
      if (!currentUserId) {
        throw new Error('User ID is required');
      }
      userId = currentUserId;
    }
    
    console.log(`Fetching threads for user: ${userId}, token exists: ${!!token}`);
    
    const endpoint = `${API_BASE_URL}/threads/user/${userId}`;
    console.log(`Making request to: ${endpoint}`);
    
    const response = await fetch(endpoint, {
      method: "GET",
      headers: { 
        "Content-Type": "application/json",
        "Authorization": token ? `Bearer ${token}` : ''
      },
      credentials: "include",
    });
    
    if (response.status === 401) {
      console.warn("Authentication error when fetching threads - token may be invalid");
      // Don't throw yet - allow caller to handle auth errors
    }
    
    // If we got a 500 error, it might be due to the proto issues with user data
    if (response.status === 500) {
      console.warn("Encountered 500 error in getThreadsByUser, attempting to process response anyway");
      
      // Even with the 500 error, try to parse the response body
      // The server might have returned partial data we can use
      try {
        const data = await response.json();
        
        // Check if we have at least some threads data we can use
        if (data && data.threads && Array.isArray(data.threads)) {
          console.info(`Retrieved ${data.threads.length} threads despite error status`);
          return data;
        }
      } catch (parseError) {
        console.error("Failed to parse response from error status:", parseError);
        // Continue to error handling below
      }
    }
    
    if (!response.ok) {
      let errorMessage = `Failed to fetch user's threads: ${response.status} ${response.statusText}`;
      try {
        const errorData = await response.json();
        errorMessage = errorData.message || 
                      errorData.error?.message || 
                      errorMessage;
        console.error("API error response:", errorData);
      } catch (parseError) {
        console.error("Could not parse error response:", parseError);
      }
      throw new Error(errorMessage);
    }
    
    const data = await response.json();
    console.log(`Successfully retrieved ${data.threads?.length || 0} threads`);
    return data;
  } catch (error) {
    console.error("Error in getThreadsByUser:", error);
    throw error;
  }
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