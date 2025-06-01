import { getAuthToken, getUserId } from '../utils/auth';
import appConfig from '../config/appConfig';
import { uploadMultipleThreadMedia } from '../utils/supabase';

const API_BASE_URL = appConfig.api.baseUrl;
const AI_SERVICE_URL = appConfig.api.aiServiceUrl || 'http://localhost:5000';

// Log the API URL being used for debugging
console.log('Thread API using URL:', API_BASE_URL);

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
  } catch (error: any) {
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

export async function getThreadsByUser(userId: string, page: number = 1, limit: number = 10) {
  try {
    const token = getAuthToken();
    let actualUserId = userId;
    
    if (userId === 'me') {
      const currentUserId = getUserId();
      console.log('Current user ID from auth:', currentUserId);
      
      if (!currentUserId) {
        throw new Error('User ID is required');
      }
      actualUserId = currentUserId;
    }
    
    const endpoint = `${API_BASE_URL}/threads/user/${actualUserId}?page=${page}&limit=${limit}`;
    console.log(`Making request to: ${endpoint}`);
    
    const response = await fetch(endpoint, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': token ? `Bearer ${token}` : ''
      },
      credentials: "include"
    });
    
    if (!response.ok) {
      if (response.status === 400) {
        let errorMessage = `Bad request when getting user threads`;
        try {
          const errorData = await response.json();
          errorMessage = errorData.message || errorMessage;
          console.error("API error response:", errorData);
        } catch (parseError) {
          console.error("Could not parse error response:", parseError);
        }
        throw new Error(errorMessage);
      }
      throw new Error(`Failed to get user threads: ${response.status}`);
    }
    
    const data = await response.json();
    
    console.log("Threads received from API:", data.threads);
    console.log("Pinned threads:", data.threads.filter(thread => thread.is_pinned === true).length);
    console.log("Pinned thread IDs:", data.threads.filter(thread => thread.is_pinned === true).map(t => t.id));
    
    return data;
  } catch (err) {
    console.error('Error getting user threads:', err);
    throw err;
  }
}

/**
 * Get all threads with standardized pagination
 * @param page Page number (starts at 1)
 * @param limit Number of items per page
 * @returns Promise with threads data
 */
export async function getAllThreads(page = 1, limit = 10) {
  try {
    const token = getAuthToken();
    
    // Log what URL we're trying to hit
    const url = `${API_BASE_URL}/threads?page=${page}&limit=${limit}`;
    console.log(`[Thread API] Attempting to fetch threads from: ${url}`);
    
    // Create an AbortController for timeout handling
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), 15000); // 15 second timeout
    
    try {
      console.log(`[Thread API] Making fetch request with headers:`, {
        'Content-Type': 'application/json',
        'Authorization': token ? 'Bearer [token]' : 'none'
      });
      
      const response = await fetch(url, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': token ? `Bearer ${token}` : ''
        },
        credentials: 'include',
        signal: controller.signal
      });
      
      // Clear timeout
      clearTimeout(timeoutId);
      
      console.log(`[Thread API] getAllThreads response status: ${response.status}`);
      
      if (!response.ok) {
        try {
          const errorText = await response.text();
          console.error(`[Thread API] API returned error status: ${response.status}`, errorText);
          throw new Error(`Failed to fetch threads (${response.status}): ${errorText.substring(0, 100)}`);
        } catch (textError) {
          console.error(`[Thread API] Failed to read error response text:`, textError);
          throw new Error(`Failed to fetch threads (${response.status})`);
        }
      }
      
      let data;
      try {
        const responseText = await response.text();
        console.log(`[Thread API] Raw response text (first 100 chars): ${responseText.substring(0, 100)}...`);
        
        try {
          data = JSON.parse(responseText);
        } catch (jsonError) {
          console.error(`[Thread API] JSON parse error:`, jsonError);
          console.error(`[Thread API] Response text that failed to parse:`, responseText);
          throw new Error('Failed to parse API response as JSON');
        }
      } catch (textError) {
        console.error(`[Thread API] Error reading response text:`, textError);
        throw new Error('Failed to read API response');
      }
      
      console.log(`[Thread API] getAllThreads received ${data.threads?.length || 0} threads`);
      
      if (!data.threads || !Array.isArray(data.threads)) {
        console.warn('[Thread API] API returned invalid threads data structure', data);
        return {
          success: false,
          error: 'Invalid data format received from API',
          threads: [],
          total_count: 0
        };
      }
      
      return {
        success: true,
        threads: data.threads || [],
        total_count: data.total || 0
      };
    } catch (fetchError: unknown) {
      // Clear timeout
      clearTimeout(timeoutId);
      
      // Check if this was a timeout
      if (fetchError instanceof Error && fetchError.name === 'AbortError') {
        console.error("[Thread API] getAllThreads request timed out after 15 seconds");
        throw new Error("Request timed out. The API server might be overloaded or unavailable.");
      }
      
      throw fetchError;
    }
  } catch (error) {
    console.error('[Thread API] Get all threads failed:', error);
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
  try {
    // First try to upload directly to Supabase
    const urls = await uploadMultipleThreadMedia(files, threadId);
    
    if (urls && urls.length > 0) {
      const token = getAuthToken();
      
      // Update thread with the Supabase media URLs
      const response = await fetch(`${API_BASE_URL}/threads/${threadId}/media/update`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': token ? `Bearer ${token}` : ''
        },
        body: JSON.stringify({ mediaUrls: urls }),
        credentials: 'include',
      });
      
      if (!response.ok) {
        try {
          const errorData = await response.json();
          throw new Error(
            errorData.message || 
            errorData.error?.message || 
            `Failed to update thread with media URLs: ${response.status}`
          );
        } catch (parseError) {
          throw new Error(`Failed to update thread with media URLs: ${response.status}`);
        }
      }
      
      return response.json();
    }
    
    // Fall back to the API if Supabase upload fails
  const token = getAuthToken();
  
  const formData = new FormData();
  formData.append('thread_id', threadId);
  
  files.forEach((file, index) => {
    formData.append(`file`, file);
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
  } catch (error) {
    console.error("Error in uploadThreadMedia:", error);
    throw error;
  }
}

// Social Features

export async function likeThread(threadId: string) {
  try {
    const token = getAuthToken();
    
    console.log(`Attempting to like thread ${threadId}`);
    
    // Create controller for timeout management
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), 5000); // 5 second timeout
    
    const response = await fetch(`${API_BASE_URL}/threads/${threadId}/like`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Authorization": token ? `Bearer ${token}` : ''
      },
      credentials: "include",
      signal: controller.signal
    });
    
    clearTimeout(timeoutId);

    const data = await response.json();
    
    // Special handling for ALREADY_LIKED status - treat as success
    if (data.code === "ALREADY_LIKED") {
      console.log(`Thread ${threadId} is already liked by the user`);
      return {
        success: true,
        message: "Thread already liked",
        data: { thread_id: threadId }
      };
    }

    if (!response.ok) {
      // Handle error response
      const errorMessage = data.message || 
        data.error?.message || 
        `Failed to like thread (${response.status} ${response.statusText})`;
      
      console.error(`Like thread error: ${errorMessage}`, data);
      throw new Error(errorMessage);
    }

    console.log(`Successfully liked thread ${threadId}`, data);
    
    // Add the thread to local storage to track liked status
    try {
      const likedThreads = JSON.parse(localStorage.getItem('likedThreads') || '[]');
      if (!likedThreads.includes(threadId)) {
        likedThreads.push(threadId);
        localStorage.setItem('likedThreads', JSON.stringify(likedThreads));
      }
    } catch (err) {
      console.error('Error saving liked status to localStorage:', err);
      // Continue even if local storage fails
    }
    
    return data;
  } catch (error) {
    if (error instanceof DOMException && error.name === 'AbortError') {
      console.error("Like thread request timed out after 5 seconds");
      throw new Error("Request timed out. Please try again.");
    }
    
    console.error("Error in likeThread:", error);
    throw error;
  }
}

export async function unlikeThread(threadId: string) {
  try {
    const token = getAuthToken();
    
    if (!token) {
      console.error("No auth token available for unliking thread");
      throw new Error("Authentication required to unlike a thread");
    }
    
    console.log(`Attempting to unlike thread ${threadId}`);
    
    // Create controller for timeout management
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), 5000); // 5 second timeout
    
    const response = await fetch(`${API_BASE_URL}/threads/${threadId}/like`, {
      method: "DELETE",
      headers: { 
        "Content-Type": "application/json",
        "Authorization": `Bearer ${token}`
      },
      credentials: "include",
      signal: controller.signal
    });
    
    clearTimeout(timeoutId);
    
    if (!response.ok) {
      // Handle error response
      try {
        const errorData = await response.json();
        const errorMessage = errorData.message || 
          errorData.error?.message || 
          `Failed to unlike thread (${response.status} ${response.statusText})`;
        
        console.error(`Unlike thread error: ${errorMessage}`, errorData);
        throw new Error(errorMessage);
      } catch (parseError) {
        // If the response is not valid JSON
        console.error(`Unlike thread error (non-JSON response): ${response.status} ${response.statusText}`);
        throw new Error(`Failed to unlike thread (${response.status} ${response.statusText})`);
      }
    }
    
    const data = await response.json();
    console.log(`Successfully unliked thread ${threadId}`, data);
    
    // Remove thread from local storage
    try {
      const likedThreads = JSON.parse(localStorage.getItem('likedThreads') || '[]');
      const updatedLikes = likedThreads.filter(id => id !== threadId);
      localStorage.setItem('likedThreads', JSON.stringify(updatedLikes));
    } catch (err) {
      console.error('Error updating liked status in localStorage:', err);
      // Continue even if local storage fails
    }
    
    return data;
  } catch (error) {
    if (error instanceof DOMException && error.name === 'AbortError') {
      console.error("Unlike thread request timed out after 5 seconds");
      throw new Error("Request timed out. Please try again.");
    }
    
    console.error("Error in unlikeThread:", error);
    throw error;
  }
}

export async function replyToThread(threadId: string, data: {
  content: string;
  media?: any[];
  parent_reply_id?: string;
  mentioned_user_ids?: string[];
}) {
  try {
    const token = getAuthToken();
    
    if (!token) {
      console.error("No auth token available for replying to thread");
      throw new Error("Authentication required");
    }
    
    console.log(`Attempting to reply to thread ${threadId}`);
    
    const response = await fetch(`${API_BASE_URL}/threads/${threadId}/replies`, {
      method: "POST",
      headers: { 
        "Content-Type": "application/json",
        "Authorization": `Bearer ${token}`
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
          `Failed to reply to thread: ${response.status} ${response.statusText}`
        );
      } catch (parseError) {
        throw new Error(`Failed to reply to thread: ${response.status} ${response.statusText}`);
      }
    }
    
    return response.json();
  } catch (error) {
    console.error("Error in replyToThread:", error);
    throw error;
  }
}

export async function getThreadReplies(threadId: string) {
  try {
    const token = getAuthToken();
    
    // Set up headers - allow unauthenticated access but add auth if available
    const headers: Record<string, string> = {
      "Content-Type": "application/json",
    };
    
    if (token) {
      headers["Authorization"] = `Bearer ${token}`;
    }
    
    console.log(`Fetching replies for thread ${threadId}`);
    const response = await fetch(`${API_BASE_URL}/threads/${threadId}/replies`, {
      method: "GET",
      headers: headers,
      credentials: "include",
    });
    
    if (response.ok) {
      const data = await response.json();
      console.log(`Thread replies data for thread ${threadId}:`, data);
      
      // Check if we have user data properly included
      if (data.replies && data.replies.length > 0) {
        // Check the first reply's structure
        const firstReply = data.replies[0];
        console.log(`First reply structure from API:`, firstReply);
        
        // Reorganize replies to create proper nesting hierarchy
        // First, identify top-level replies (those without parent_reply_id)
        const topLevelReplies = data.replies.filter(reply => !reply.parent_reply_id);
        
        // Then create a map of parent_reply_id -> child replies
        const nestedRepliesMap = new Map();
        data.replies.forEach(reply => {
          if (reply.parent_reply_id) {
            const parentId = reply.parent_reply_id;
            if (!nestedRepliesMap.has(parentId)) {
              nestedRepliesMap.set(parentId, []);
            }
            nestedRepliesMap.get(parentId).push(reply);
          }
        });
        
        // Log the hierarchy structure for debugging
        console.log(`Organized top-level replies:`, topLevelReplies.length);
        console.log(`Organized nested replies map:`, Array.from(nestedRepliesMap.entries()).map(
          ([parentId, children]) => ({ parentId, childrenCount: children.length })
        ));
        
        // Replace the original flat reply list with top-level replies only
        data.replies = topLevelReplies;
        
        // Pass the nested replies map in the data structure
        data.nestedRepliesMap = Object.fromEntries(nestedRepliesMap);
        
        // Add user data fields if they're missing but can be derived from nested structures
        data.replies = data.replies.map(reply => {
          // Check if reply has a valid user field
          if (reply.user) {
            // Ensure user data is accessible from top level of the reply object as well
            return {
              ...reply,
              author_username: reply.user.username,
              author_name: reply.user.name,
              author_avatar: reply.user.profile_picture_url,
            };
          }
          return reply;
        });
      }
      
      return data;
    }
    
    // If 401 unauthorized, we could return empty results
    if (response.status === 401) {
      console.warn("Unauthorized when fetching replies - returning empty results");
      return { 
        replies: [],
        total_count: 0
      };
    }
    
    // Handle error response
    try {
      const errorData = await response.json();
      throw new Error(
        errorData.message || 
        errorData.error?.message || 
        `Failed to fetch replies: ${response.status} ${response.statusText}`
      );
    } catch (parseError) {
      throw new Error(`Failed to fetch replies: ${response.status} ${response.statusText}`);
    }
  } catch (error) {
    console.error("Error in getThreadReplies:", error);
    throw error;
  }
}

export async function repostThread(threadId: string, content = '') {
  try {
    const token = getAuthToken();
    
    const response = await fetch(`${API_BASE_URL}/threads/${threadId}/repost`, {
      method: "POST",
      headers: { 
        "Content-Type": "application/json",
        "Authorization": token ? `Bearer ${token}` : ''
      },
      body: JSON.stringify({ content }),
      credentials: "include",
    });
    
    if (!response.ok) {
      try {
        const errorData = await response.json();
        throw new Error(
          errorData.message || 
          errorData.error?.message || 
          `Failed to repost thread: ${response.status} ${response.statusText}`
        );
      } catch (parseError) {
        throw new Error(`Failed to repost thread: ${response.status} ${response.statusText}`);
      }
    }
    
    return response.json();
  } catch (error) {
    console.error("Error in repostThread:", error);
    throw error;
  }
}

export async function removeRepost(repostId: string) {
  try {
    const token = getAuthToken();
    
    const response = await fetch(`${API_BASE_URL}/threads/${repostId}/repost`, {
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
          `Failed to remove repost: ${response.status} ${response.statusText}`
        );
      } catch (parseError) {
        throw new Error(`Failed to remove repost: ${response.status} ${response.statusText}`);
      }
    }
    
    return response.json();
  } catch (error) {
    console.error("Error in removeRepost:", error);
    throw error;
  }
}

export async function bookmarkThread(threadId: string) {
  const token = getAuthToken();

  console.log(`bookmarkThread API called for threadId: ${threadId}`);
  
  try {
    console.log(`Attempting to bookmark thread ${threadId}`);
    
    // Create controller for timeout management
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), 5000); // 5 second timeout
    
    const response = await fetch(`${API_BASE_URL}/threads/${threadId}/bookmark`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Authorization": token ? `Bearer ${token}` : ''
      },
      credentials: "include",
      signal: controller.signal
    });
    
    clearTimeout(timeoutId);

    console.log(`Bookmark API response status: ${response.status}`);
    
    const data = await response.json();
    
    // Special handling for ALREADY_BOOKMARKED status - treat as success
    if (data.code === "ALREADY_BOOKMARKED") {
      console.log(`Thread ${threadId} is already bookmarked by the user`);
      return {
        success: true,
        message: "Thread already bookmarked",
        data: { thread_id: threadId }
      };
    }

    // Handle specific HTTP status codes
    if (response.status === 401) {
      console.error('User not authenticated for bookmarking');
      throw new Error('Authentication required to bookmark a thread');
    }
    
    if (response.status === 404) {
      console.error('Thread not found for bookmarking');
      throw new Error('The thread you are trying to bookmark does not exist');
    }

    if (!response.ok) {
      let errorMessage = `Failed to bookmark thread (${response.status})`;
      
      console.error("Bookmark failed with error:", data);
      errorMessage = data.message || 
                    data.error?.message || 
                    errorMessage;
      
      throw new Error(errorMessage);
    }

    console.log(`Successfully bookmarked thread ${threadId}`, data);
    
    // Store bookmark state in localStorage for offline recovery
    try {
      const bookmarkedThreads = JSON.parse(localStorage.getItem('bookmarkedThreads') || '[]');
      if (!bookmarkedThreads.includes(threadId)) {
        bookmarkedThreads.push(threadId);
        localStorage.setItem('bookmarkedThreads', JSON.stringify(bookmarkedThreads));
      }
    } catch (err) {
      console.error('Error saving bookmark state to localStorage:', err);
      // Continue even if local storage fails
    }
    
    return data;
  } catch (error) {
    if (error instanceof DOMException && error.name === 'AbortError') {
      console.error("Bookmark thread request timed out after 5 seconds");
      throw new Error("Request timed out. Please try again.");
    }
    
    console.error(`Error in bookmark API call:`, error);
    throw error;
  }
}

export async function removeBookmark(threadId: string) {
  const token = getAuthToken();

  console.log(`removeBookmark API called for threadId: ${threadId}`);
  
  try {
    console.log(`Attempting to remove bookmark for thread ${threadId}`);
    
    // Create controller for timeout management
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), 5000); // 5 second timeout
    
    const response = await fetch(`${API_BASE_URL}/threads/${threadId}/bookmark`, {
      method: "DELETE",
      headers: {
        "Content-Type": "application/json",
        "Authorization": token ? `Bearer ${token}` : ''
      },
      credentials: "include",
      signal: controller.signal
    });
    
    clearTimeout(timeoutId);

    console.log(`Unbookmark API response status: ${response.status}`);

    // Handle specific status codes
    if (response.status === 401) {
      console.error('User not authenticated for removing bookmark');
      throw new Error('Authentication required to remove a bookmark');
    }

    if (!response.ok) {
      let errorMessage = `Failed to remove bookmark (${response.status})`;
      
      try {
        const errorData = await response.json();
        console.error("Unbookmark failed with error:", errorData);
        errorMessage = errorData.message || 
                      errorData.error?.message || 
                      errorMessage;
      } catch (parseError) {
        // If we can't parse the error, just use the status code message
        console.error("Couldn't parse error response:", parseError);
      }
      
      throw new Error(errorMessage);
    }

    const result = await response.json();
    console.log(`Successfully removed bookmark for thread ${threadId}`, result);
    
    // Update localStorage for offline recovery
    try {
      const bookmarkedThreads = JSON.parse(localStorage.getItem('bookmarkedThreads') || '[]');
      const updatedBookmarks = bookmarkedThreads.filter(id => id !== threadId);
      localStorage.setItem('bookmarkedThreads', JSON.stringify(updatedBookmarks));
    } catch (err) {
      console.error('Error updating bookmark state in localStorage:', err);
      // Continue even if local storage fails
    }
    
    return result;
  } catch (error) {
    if (error instanceof DOMException && error.name === 'AbortError') {
      console.error("Unbookmark thread request timed out after 5 seconds");
      throw new Error("Request timed out. Please try again.");
    }
    
    console.error(`Error in unbookmark API call:`, error);
    throw error;
  }
}

export async function getFollowingThreads(page = 1, limit = 20) {
  try {
    const token = getAuthToken();
    
    if (!token) {
      console.error('No auth token available, cannot fetch following threads');
      throw new Error('Authentication required');
    }
    
    // Log what URL we're trying to hit
    const url = `${API_BASE_URL}/threads/following?page=${page}&limit=${limit}`;
    console.log(`Attempting to fetch following threads from: ${url}`);
    
    // Create an AbortController for timeout handling
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), 15000); // 15 second timeout
    
    try {
      console.log('Fetching following threads...');
      const response = await fetch(url, {
        method: "GET",
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`
        },
        credentials: "include",
        signal: controller.signal
      });
      
      // Clear timeout
      clearTimeout(timeoutId);
      
      console.log('Following threads API response status:', response.status);
      
      if (response.ok) {
        const data = await response.json();
        
        // Log received data for debugging
        if (data && data.threads && data.threads.length > 0) {
          console.log(`Received ${data.threads.length} following threads from API`);
          console.log('First thread data structure:', data.threads[0]);
        } else {
          console.log('API returned no following threads');
          return { 
            success: true,
            threads: [], 
            total_count: 0, 
            page, 
            limit 
          };
        }
        
        return {
          ...data,
          success: true
        };
      }
      
      // Handle API errors
      console.error(`API returned error status: ${response.status}`);
      throw new Error(`Failed to fetch following threads: ${response.status}`);
    } catch (fetchError: unknown) {
      // Clear timeout
      clearTimeout(timeoutId);
      
      // Check if this was a timeout
      if (fetchError instanceof Error && fetchError.name === 'AbortError') {
        console.error("Request timed out after 15 seconds");
        throw new Error("Request timed out after 15 seconds");
      } else {
        console.error("Network error when fetching following threads:", fetchError);
        throw fetchError;
      }
    }
  } catch (error: any) {
    console.error("Unexpected error in getFollowingThreads:", error);
    throw error;
  }
}

// Search threads based on query
export async function searchThreads(
  query: string, 
  page: number = 1, 
  limit: number = 10, 
  options?: { filter?: string; category?: string; sortBy?: string }
) {
  try {
    const url = new URL(`${API_BASE_URL}/threads/search`);
    url.searchParams.append('q', query);
    url.searchParams.append('page', page.toString());
    url.searchParams.append('limit', limit.toString());
    
    if (options?.filter) {
      url.searchParams.append('filter', options.filter);
    }
    
    if (options?.category) {
      url.searchParams.append('category', options.category);
    }
    
    if (options?.sortBy) {
      url.searchParams.append('sort', options.sortBy);
    }
    
    const token = getAuthToken();
    
    const response = await fetch(url.toString(), {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': token ? `Bearer ${token}` : ''
      }
    });
    
    if (!response.ok) {
      throw new Error(`Failed to search threads: ${response.status}`);
    }
    
    const data = await response.json();
    return data;
  } catch (error: any) {
    console.error('Error searching threads:', error);
    throw error;
  }
}

// Search threads with media based on query
export async function searchThreadsWithMedia(
  query: string, 
  page: number = 1, 
  limit: number = 10, 
  options?: { filter?: string; category?: string }
) {
  try {
    const url = new URL(`${API_BASE_URL}/threads/search/media`);
    
    url.searchParams.append('q', query);
    url.searchParams.append('page', page.toString());
    url.searchParams.append('limit', limit.toString());
    
    if (options?.filter) {
      url.searchParams.append('filter', options.filter);
    }
    
    if (options?.category) {
      url.searchParams.append('category', options.category);
    }
    
    const token = getAuthToken();
    
    const response = await fetch(url.toString(), {
      method: 'GET',
      headers: {
        'Authorization': token ? `Bearer ${token}` : '',
        'Content-Type': 'application/json'
      }
    });
    
    if (!response.ok) {
      throw new Error(`Failed to search threads with media: ${response.status}`);
    }
    
    return await response.json();
  } catch (error: any) {
    console.error('Error searching threads with media:', error);
    throw error;
  }
}

// Get threads by hashtag
export async function getThreadsByHashtag(
  hashtag: string, 
  page: number = 1, 
  limit: number = 10
) {
  try {
    const cleanHashtag = hashtag.startsWith('#') ? hashtag.substring(1) : hashtag;
    
    const url = new URL(`${API_BASE_URL}/threads/hashtag/${encodeURIComponent(cleanHashtag)}`);
    
    url.searchParams.append('page', page.toString());
    url.searchParams.append('limit', limit.toString());
    
    const token = getAuthToken();
    
    const response = await fetch(url.toString(), {
      method: 'GET',
      headers: {
        'Authorization': token ? `Bearer ${token}` : '',
        'Content-Type': 'application/json'
      }
    });
    
    if (!response.ok) {
      throw new Error(`Failed to get threads by hashtag: ${response.status}`);
    }
    
    return await response.json();
  } catch (error: any) {
    console.error('Error getting threads by hashtag:', error);
    throw error;
  }
}

export async function getReplyReplies(replyId: string, page = 1, limit = 20): Promise<{ replies: any[], total_count: number, cached?: boolean, error?: string }> {
  try {
    console.log(`Fetching replies for reply ID: ${replyId}`);
    const token = getAuthToken();
    
    // Create query parameters
    const params = new URLSearchParams({
      page: page.toString(),
      limit: limit.toString()
    });
    
    const response = await fetch(`${API_BASE_URL}/replies/${replyId}/replies?${params.toString()}`, {
      method: "GET",
      headers: {
        "Content-Type": "application/json",
        "Authorization": token ? `Bearer ${token}` : ''
      },
      credentials: "include"
    });
    
    if (!response.ok) {
      const errorData = await response.json();
      console.error(`Error fetching replies for reply ${replyId}:`, errorData);
      return { replies: [], total_count: 0, error: errorData.message || 'Failed to fetch replies' };
    }
    
    const data = await response.json();
    console.log(`Successfully fetched ${data.replies?.length || 0} replies for reply ${replyId}`);
    return data;
  } catch (error) {
    console.error(`Error fetching replies for reply ${replyId}:`, error);
    return { replies: [], total_count: 0, error: 'Network error fetching replies' };
  }
}

export async function suggestThreadCategory(content: string) {
  try {
    console.log("Requesting category suggestion for content:", content.substring(0, 50) + (content.length > 50 ? "..." : ""));
    
    // If content is empty or too short, don't make the request
    if (!content || content.trim().length < 10) {
      return { 
        category: 'general',
        confidence: 0
      };
    }
    
    const response = await fetch(`${AI_SERVICE_URL}/predict/category`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({ content })
    });
    
    if (!response.ok) {
      console.warn("Category suggestion failed:", response.status, response.statusText);
      return { 
        category: 'general',
        confidence: 0
      };
    }
    
    const data = await response.json();
    console.log("Received category suggestion:", data);
    
    return {
      category: data.category || 'general',
      confidence: data.confidence || 0
    };
  } catch (error) {
    console.error("Error suggesting thread category:", error);
    return { 
      category: 'general',
      confidence: 0
    };
  }
}

export async function likeReply(replyId: string) {
  try {
    const token = getAuthToken();
    
    const response = await fetch(`${API_BASE_URL}/replies/${replyId}/like`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Authorization": token ? `Bearer ${token}` : ''
      },
      credentials: "include"
    });
    
    if (!response.ok) {
      const errorData = await response.json();
      console.error(`Error liking reply ${replyId}:`, errorData);
      return { success: false, error: errorData.message || 'Failed to like reply' };
    }
    
    const data = await response.json();
    console.log(`Successfully liked reply ${replyId}`);
    return { ...data, success: true };
  } catch (error) {
    console.error(`Error liking reply ${replyId}:`, error);
    return { success: false, error: 'Network error liking reply' };
  }
}

export async function unlikeReply(replyId: string) {
  try {
    const token = getAuthToken();
    
    const response = await fetch(`${API_BASE_URL}/replies/${replyId}/unlike`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Authorization": token ? `Bearer ${token}` : ''
      },
      credentials: "include"
    });
    
    if (!response.ok) {
      const errorData = await response.json();
      console.error(`Error unliking reply ${replyId}:`, errorData);
      return { success: false, error: errorData.message || 'Failed to unlike reply' };
    }
    
    const data = await response.json();
    console.log(`Successfully unliked reply ${replyId}`);
    return { ...data, success: true };
  } catch (error) {
    console.error(`Error unliking reply ${replyId}:`, error);
    return { success: false, error: 'Network error unliking reply' };
  }
}

// Helper function to check if a string is a valid UUID
function isUuid(str: string): boolean {
  const uuidPattern = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;
  return uuidPattern.test(str);
}

// Helper function to get user ID from username if necessary
async function resolveUserIdIfNeeded(userId: string): Promise<string> {
  // If already a valid UUID, return immediately
  if (isUuid(userId)) {
    return userId;
  }
  
  // If it's 'me', get current user ID
  if (userId === 'me') {
    const token = getAuthToken();
    const currentUserId = token ? JSON.parse(atob(token.split('.')[1])).sub : null;
    return currentUserId || userId;
  }
  
  // Otherwise, assume it's a username and resolve it
  try {
    const response = await fetch(`${API_BASE_URL}/users/username/${encodeURIComponent(userId)}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${getAuthToken()}`
      }
    });
    
    if (!response.ok) {
      console.error(`Failed to resolve username: ${response.status}`);
      return userId; // Return original as fallback
    }
    
    const data = await response.json();
    return data.user?.id || userId;
  } catch (error) {
    console.error('Error resolving user ID:', error);
    return userId; // Return original as fallback
  }
}

// User thread functions
export const getUserThreads = async (userId: string, page = 1, limit = 10): Promise<any> => {
  try {
    const resolvedUserId = await resolveUserIdIfNeeded(userId);
    console.log(`Fetching threads for user ${resolvedUserId} (original: ${userId}), page: ${page}, limit: ${limit}`);
    
    const response = await fetch(`${API_BASE_URL}/threads/user/${resolvedUserId}?page=${page}&limit=${limit}`, {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${getAuthToken()}`,
        'Content-Type': 'application/json'
      }
    });

    if (!response.ok) {
      let errorMessage = `Failed to get user threads: ${response.status}`;
      try {
        const errorData = await response.json();
        console.error('Error getting user threads:', errorData);
        if (errorData.message) {
          errorMessage += ` - ${errorData.message}`;
        }
      } catch (parseError) {
        console.error('Could not parse error response:', parseError);
      }
      throw new Error(errorMessage);
    }

    return await response.json();
  } catch (error) {
    console.error('Error in getUserThreads:', error);
    throw error;
  }
};

export const getUserReplies = async (userId: string, page = 1, limit = 10): Promise<any> => {
  try {
    const resolvedUserId = await resolveUserIdIfNeeded(userId);
    console.log(`Fetching replies for user ${resolvedUserId} (original: ${userId}), page: ${page}, limit: ${limit}`);
    
    const response = await fetch(`${API_BASE_URL}/threads/user/${resolvedUserId}/replies?page=${page}&limit=${limit}`, {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${getAuthToken()}`,
        'Content-Type': 'application/json'
      }
    });

    if (!response.ok) {
      let errorMessage = `Failed to get user replies: ${response.status}`;
      try {
        const errorData = await response.json();
        console.error('Error getting user replies:', errorData);
        if (errorData.message) {
          errorMessage += ` - ${errorData.message}`;
        }
      } catch (parseError) {
        console.error('Could not parse error response:', parseError);
      }
      throw new Error(errorMessage);
    }

    return await response.json();
  } catch (error) {
    console.error('Error getting user replies:', error);
    throw error;
  }
};

export const getUserLikedThreads = async (userId: string, page: number = 1, limit: number = 10) => {
  try {
    const token = getAuthToken();
    let actualUserId = userId;
    
    if (userId === 'me') {
      actualUserId = await resolveUserIdIfNeeded(userId);
    }
    
    const endpoint = `${API_BASE_URL}/threads/user/${actualUserId}/likes?page=${page}&limit=${limit}`;
    console.log(`Making request to: ${endpoint}`);
    
    const response = await fetch(endpoint, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': token ? `Bearer ${token}` : ''
      },
      credentials: "include"
    });
    
    if (!response.ok) {
      if (response.status === 400) {
        let errorMessage = `Bad request when getting user liked threads`;
        try {
          const errorData = await response.json();
          errorMessage = errorData.message || errorMessage;
          console.error("API error response:", errorData);
        } catch (parseError) {
          console.error("Could not parse error response:", parseError);
        }
        throw new Error(errorMessage);
      }
      throw new Error(`Failed to get user liked threads: ${response.status}`);
    }
    
    const responseData = await response.json();
    return responseData;
  } catch (err) {
    console.error('Error getting user liked threads:', err);
    throw err;
  }
};

export const getUserMedia = async (userId: string, page = 1, limit = 10): Promise<any> => {
  try {
    const resolvedUserId = await resolveUserIdIfNeeded(userId);
    console.log(`Fetching media for user ${resolvedUserId} (original: ${userId}), page: ${page}, limit: ${limit}`);
    
    const response = await fetch(`${API_BASE_URL}/threads/user/${resolvedUserId}/media?page=${page}&limit=${limit}`, {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${getAuthToken()}`,
        'Content-Type': 'application/json'
      }
    });

    if (!response.ok) {
      let errorMessage = `Failed to get user media: ${response.status}`;
      try {
        const errorData = await response.json();
        console.error('Error getting user media:', errorData);
        if (errorData.message) {
          errorMessage += ` - ${errorData.message}`;
        }
      } catch (parseError) {
        console.error('Could not parse error response:', parseError);
      }
      throw new Error(errorMessage);
    }

    return await response.json();
  } catch (error) {
    console.error('Error getting user media:', error);
    throw error;
  }
};

export const getUserBookmarks = async (userId: string, page = 1, limit = 10): Promise<any> => {
  try {
    const token = getAuthToken();
    const actualUserId = userId === 'me' ? getUserId() : userId;
    
    if (!actualUserId) {
      console.error('No user ID available, cannot fetch bookmarks');
      throw new Error('User ID is required');
    }
    
    // Log what URL we're trying to hit
    const url = `${API_BASE_URL}/bookmarks?page=${page}&limit=${limit}`;
    console.log(`Attempting to fetch bookmarks from: ${url}`);
    
    const response = await fetch(url, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': token ? `Bearer ${token}` : ''
      },
      credentials: "include"
    });
    
    if (!response.ok) {
      console.error(`Failed to get bookmarks: ${response.status}`);
      throw new Error(`Failed to get bookmarks: ${response.status}`);
    }
    
    const data = await response.json();
    console.log('Bookmarks API returned data:', data);
    
    return {
      success: true,
      threads: data.bookmarks || [],
      total: data.total || 0,
      pagination: data.pagination || null
    };
  } catch (err) {
    console.error('Error getting user bookmarks:', err);
    throw err;
  }
}

/**
 * Direct attempt to fetch from the API in case of network issues
 * Provides a fallback method in case normal API calls fail
 */
export async function directFetchThreads(page = 1, limit = 10) {
  try {
    // Try multiple possible URLs to find one that works
    const possibleUrls = [
      // Docker internal network
      'http://api_gateway:8081/api/v1/threads',
      // Host machine mapping
      'http://localhost:8083/api/v1/threads',
      // Config URL
      `${API_BASE_URL}/threads`
    ];
    
    console.log('[Thread API] Attempting direct fetch from multiple URLs');
    
    // Try each URL in sequence
    for (const baseUrl of possibleUrls) {
      try {
        const url = `${baseUrl}?page=${page}&limit=${limit}`;
        console.log(`[Thread API] Trying direct fetch from: ${url}`);
        
        const response = await fetch(url, {
          method: 'GET',
          headers: {
            'Content-Type': 'application/json',
            'Accept': 'application/json'
          },
          mode: 'cors'
        });
        
        if (response.ok) {
          const data = await response.json();
          console.log(`[Thread API] Direct fetch successful from ${url}, got ${data.threads?.length || 0} threads`);
          
          return {
            success: true,
            threads: data.threads || [],
            total_count: data.total || 0,
            source_url: url
          };
        } else {
          console.log(`[Thread API] Direct fetch failed from ${url} with status ${response.status}`);
        }
      } catch (urlError) {
        console.log(`[Thread API] Error with URL ${baseUrl}:`, urlError);
      }
    }
    
    throw new Error('All direct fetch attempts failed');
  } catch (error) {
    console.error('[Thread API] Direct fetch failed:', error);
    throw error;
  }
}