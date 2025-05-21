import { getAuthToken, getUserId } from '../utils/auth';
import appConfig from '../config/appConfig';
import { uploadMultipleThreadMedia } from '../utils/supabase';

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

export async function getAllThreads(page = 1, limit = 20) {
  try {
  const token = getAuthToken();
  
    // Define headers - allow unauthenticated access but add token if available
    const headers: Record<string, string> = {
      "Content-Type": "application/json",
    };
    
    if (token) {
      headers["Authorization"] = `Bearer ${token}`;
    }
    
    console.log(`Fetching threads - page ${page}, limit ${limit}`);
    console.log(`Auth token present: ${!!token}`);
    
    const response = await fetch(`${API_BASE_URL}/threads?page=${page}&limit=${limit}`, {
      method: "GET",
      headers: headers,
      credentials: "include",
    });
    
    console.log(`API response status: ${response.status}`);
    
    if (response.ok) {
      const data = await response.json();
      
      // Log like status of threads for debugging
      if (data && data.threads && data.threads.length > 0) {
        console.log(`Received ${data.threads.length} threads from API`);
        console.log('Like status of first few threads:', data.threads.slice(0, 3).map(t => ({ 
          id: t.id, 
          is_liked: t.is_liked, 
          isLiked: t.isLiked 
        })));
      }
      
      // Get liked threads from localStorage for client-side verification
      let likedThreads: string[] = [];
      try {
        likedThreads = JSON.parse(localStorage.getItem('likedThreads') || '[]');
      } catch (err) {
        console.error('Error parsing liked threads from localStorage:', err);
      }
      
      // Ensure threads have consistent like status from localStorage
      if (data.threads) {
        data.threads = data.threads.map(thread => {
          // If the thread is in our liked threads localStorage, make sure it shows as liked
          if (likedThreads.includes(thread.id)) {
      return { 
              ...thread,
              is_liked: true,
              isLiked: true
            };
          }
          return thread;
        });
      }
      
      return data;
    }
    
    // If 401 Unauthorized, return empty results instead of throwing
    if (response.status === 401) {
      console.warn("Unauthorized when fetching threads - returning empty results");
      return { threads: [], total_count: 0, page, limit };
    }
    
    // If 500 server error, return empty results with a log message instead of throwing
    if (response.status === 500) {
      console.error("Server error (500) when fetching threads - returning empty results");
      return { threads: [], total_count: 0, page, limit };
    }
    
    // For other errors, try to parse the error message but don't throw
      try {
        const errorData = await response.json();
      console.error(`Error in getAllThreads: ${errorData.message || `Failed to fetch threads: ${response.status}`}`);
      return { threads: [], total_count: 0, page, limit };
      } catch (parseError) {
      console.error(`Error in getAllThreads: Failed to fetch threads: ${response.status}`);
      return { threads: [], total_count: 0, page, limit };
    }
  } catch (error) {
    console.error("Error in getAllThreads:", error);
    // Return empty results instead of throwing to avoid breaking the UI
    return { threads: [], total_count: 0, page, limit };
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
      // Handle error response
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
  const token = getAuthToken();
  
  try {
    console.log(`Fetching followed users threads, page: ${page}, limit: ${limit}`);
    
    const endpoint = `${API_BASE_URL}/threads/following?page=${page}&limit=${limit}`;
    console.log(`Making request to: ${endpoint}`);
    
    if (!token) {
      console.warn("No token available for getFollowingThreads");
      return { 
        threads: [],
        total_count: 0,
        page: page,
        limit: limit
      };
    }
    
    const response = await fetch(endpoint, {
      method: "GET",
      headers: { 
        "Content-Type": "application/json",
        "Authorization": `Bearer ${token}`
      },
      credentials: "include",
    });
    
    if (response.status === 401) {
      console.warn("Unauthorized when fetching following threads - returning empty results");
      return { 
        threads: [],
        total_count: 0,
        page: page,
        limit: limit
      };
    }
    
    if (!response.ok) {
      let errorMessage = `Failed to fetch following threads: ${response.status} ${response.statusText}`;
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
    console.log(`Successfully retrieved ${data.threads?.length || 0} following threads`);
    return data;
  } catch (error) {
    console.error("Error in getFollowingThreads:", error);
    // Return empty results instead of throwing to keep UI working
    return { 
      threads: [],
      total_count: 0,
      page: page,
      limit: limit
    };
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

export async function likeReply(replyId: string) {
  try {
    const token = getAuthToken();
    
    if (!token) {
      console.error("No auth token available for liking reply");
      throw new Error("Authentication required");
    }
    
    const response = await fetch(`${API_BASE_URL}/replies/${replyId}/like`, {
      method: "POST",
      headers: { 
        "Content-Type": "application/json",
        "Authorization": `Bearer ${token}`
      },
      credentials: "include",
    });
    
    if (!response.ok) {
      // Handle error response
      try {
        const errorData = await response.json();
        throw new Error(
          errorData.message || 
          errorData.error?.message || 
          `Failed to like reply: ${response.status} ${response.statusText}`
        );
      } catch (parseError) {
        throw new Error(`Failed to like reply: ${response.status} ${response.statusText}`);
      }
    }
    
    return response.json();
  } catch (error) {
    console.error("Error in likeReply:", error);
    throw error;
  }
}

export async function unlikeReply(replyId: string) {
  try {
    const token = getAuthToken();
    
    if (!token) {
      console.error("No auth token available for unliking reply");
      throw new Error("Authentication required");
    }
    
    const response = await fetch(`${API_BASE_URL}/replies/${replyId}/like`, {
      method: "DELETE",
      headers: { 
        "Content-Type": "application/json",
        "Authorization": `Bearer ${token}`
      },
      credentials: "include",
    });
    
    if (!response.ok) {
      // Handle error response
      try {
        const errorData = await response.json();
        throw new Error(
          errorData.message || 
          errorData.error?.message || 
          `Failed to unlike reply: ${response.status} ${response.statusText}`
        );
      } catch (parseError) {
        throw new Error(`Failed to unlike reply: ${response.status} ${response.statusText}`);
      }
    }
    
    return response.json();
  } catch (error) {
    console.error("Error in unlikeReply:", error);
    throw error;
  }
}

export async function bookmarkReply(replyId: string) {
  try {
    const token = getAuthToken();
    
    if (!token) {
      console.error("No auth token available for bookmarking reply");
      throw new Error("Authentication required");
    }
    
    const response = await fetch(`${API_BASE_URL}/replies/${replyId}/bookmark`, {
      method: "POST",
      headers: { 
        "Content-Type": "application/json",
        "Authorization": `Bearer ${token}`
      },
      credentials: "include",
    });
    
    if (!response.ok) {
      // Handle error response
      try {
        const errorData = await response.json();
        throw new Error(
          errorData.message || 
          errorData.error?.message || 
          `Failed to bookmark reply: ${response.status} ${response.statusText}`
        );
      } catch (parseError) {
        throw new Error(`Failed to bookmark reply: ${response.status} ${response.statusText}`);
      }
    }
    
    return response.json();
  } catch (error) {
    console.error("Error in bookmarkReply:", error);
    throw error;
  }
}

export async function removeReplyBookmark(replyId: string) {
  try {
    const token = getAuthToken();
    
    if (!token) {
      console.error("No auth token available for removing reply bookmark");
      throw new Error("Authentication required");
    }
    
    const response = await fetch(`${API_BASE_URL}/replies/${replyId}/bookmark`, {
      method: "DELETE",
      headers: { 
        "Content-Type": "application/json",
        "Authorization": `Bearer ${token}`
      },
      credentials: "include",
    });
    
    if (!response.ok) {
      // Handle error response
      try {
        const errorData = await response.json();
        throw new Error(
          errorData.message || 
          errorData.error?.message || 
          `Failed to remove reply bookmark: ${response.status} ${response.statusText}`
        );
      } catch (parseError) {
        throw new Error(`Failed to remove reply bookmark: ${response.status} ${response.statusText}`);
      }
    }
    
    return response.json();
  } catch (error) {
    console.error("Error in removeReplyBookmark:", error);
    throw error;
  }
}

// User thread functions

export async function getUserThreads(userId: string, page: number = 1, limit: number = 10) {
  try {
    const token = getAuthToken();
    let actualUserId = userId;
    
    // If 'me' is specified, get the actual user ID
    if (userId === 'me') {
      const currentUserId = getUserId();
      console.log('Current user ID from auth:', currentUserId);
      
      if (!currentUserId) {
        throw new Error('User ID is required');
      }
      actualUserId = currentUserId;
    }
    
    // Set up headers - allow unauthenticated access but add auth if available
    const headers: Record<string, string> = {
      "Content-Type": "application/json",
    };
      
    if (token) {
      headers["Authorization"] = `Bearer ${token}`;
    }
    
    console.log(`Fetching threads for user ${actualUserId}, page: ${page}, limit: ${limit}`);
    
    const response = await fetch(`${API_BASE_URL}/threads/user/${actualUserId}?page=${page}&limit=${limit}`, {
      method: "GET",
      headers: headers,
      credentials: "include",
    });
    
    if (!response.ok) {
      const error = await response.text();
      throw new Error(`Failed to get user threads: ${response.status} - ${error}`);
    }
    
    // Get liked threads from localStorage for client-side verification
    let likedThreads: string[] = [];
    try {
      likedThreads = JSON.parse(localStorage.getItem('likedThreads') || '[]');
    } catch (err) {
      console.error('Error parsing liked threads from localStorage:', err);
    }
    
    const data = await response.json();
    
    // Log thread data to debug like status
    console.log(`Received ${data.threads?.length || 0} threads for user ${actualUserId}`);
    
    // Check for any thread with is_liked = true
    if (data.threads && data.threads.length > 0) {
      const likedCount = data.threads.filter(t => t.is_liked || t.isLiked).length;
      console.log(`Threads with is_liked=true from API: ${likedCount}`);
      
      // Ensure threads have consistent like status from localStorage
      data.threads = data.threads.map(thread => {
        // If the thread is in our liked threads localStorage, make sure it shows as liked
        // This is a failsafe in case the API doesn't return the correct like status
        if (likedThreads.includes(thread.id)) {
          return {
            ...thread,
            is_liked: true,
            isLiked: true
          };
    }
        return thread;
      });
    }
    
    return data;
    
  } catch (error) {
    console.error("Error in getUserThreads:", error);
    throw error;
  }
}

export async function getUserReplies(userId: string, page: number = 1, limit: number = 10) {
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
    
    const endpoint = `${API_BASE_URL}/threads/user/${actualUserId}/replies?page=${page}&limit=${limit}`;
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
        let errorMessage = `Bad request when getting user replies`;
        try {
          const errorData = await response.json();
          errorMessage = errorData.message || errorMessage;
          console.error("API error response:", errorData);
        } catch (parseError) {
          console.error("Could not parse error response:", parseError);
        }
        throw new Error(errorMessage);
      }
      throw new Error(`Failed to get user replies: ${response.status}`);
    }
    
    return await response.json();
  } catch (err) {
    console.error('Error getting user replies:', err);
    throw err;
  }
}

export async function getUserLikedThreads(userId: string, page: number = 1, limit: number = 10) {
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
    
    return await response.json();
  } catch (err) {
    console.error('Error getting user liked threads:', err);
    throw err;
  }
}

export async function getUserMedia(userId: string, page: number = 1, limit: number = 10) {
  try {
    const token = getAuthToken();
    let actualUserId = userId;
    
    // If 'me' is specified, get the actual user ID
    if (userId === 'me') {
      const currentUserId = getUserId();
      console.log('Current user ID from auth:', currentUserId);
      
      if (!currentUserId) {
        throw new Error('User ID is required');
      }
      actualUserId = currentUserId;
    }
    
    const endpoint = `${API_BASE_URL}/threads/user/${actualUserId}/media?page=${page}&limit=${limit}`;
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
      // Handle different error types
      if (response.status === 400) {
        let errorMessage = `Bad request when getting user media`;
        try {
          const errorData = await response.json();
          errorMessage = errorData.message || errorMessage;
          console.error("API error response:", errorData);
        } catch (parseError) {
          console.error("Could not parse error response:", parseError);
        }
        throw new Error(errorMessage);
      }
      throw new Error(`Failed to get user media: ${response.status}`);
    }
    
    return await response.json();
  } catch (err) {
    console.error('Error getting user media:', err);
    throw err;
  }
}

// Bookmarks API functions
export async function getUserBookmarks(page = 1, limit = 20) {
  try {
    const token = getAuthToken();
    
    console.log(`Fetching user bookmarks: page=${page}, limit=${limit}`);
    
    // This endpoint should return the user's bookmarks
    const endpoint = `${API_BASE_URL}/bookmarks?page=${page}&limit=${limit}`;
    console.log(`Making bookmarks request to: ${endpoint}`);
    
    const response = await fetch(endpoint, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': token ? `Bearer ${token}` : ''
      },
      credentials: "include"
    });
    
    console.log(`Bookmarks API response status:`, response.status);
    
    // Debug response
    let responseText;
    try {
      responseText = await response.text();
      console.log(`Bookmarks API raw response:`, responseText);
      
      // Try to parse the text as JSON
      const data = JSON.parse(responseText);
      console.log('Parsed bookmarks data:', data);
      
      if (data && data.bookmarks && Array.isArray(data.bookmarks)) {
        console.log(`Found ${data.bookmarks.length} bookmarks in response`);
      } else {
        console.warn('No bookmarks array found in response or it is not an array');
      }
      
      return data;
    } catch (error) {
      console.error('Error parsing bookmarks response:', error);
      console.error('Response text was:', responseText);
      throw new Error(`Failed to parse bookmarks response: ${(error as Error).message}`);
    }
  } catch (error: unknown) {
    console.error('Error in getUserBookmarks:', error);
    // Return empty bookmarks structure rather than throwing
    return { bookmarks: [] };
  }
}

export async function getReplyReplies(replyId: string, page = 1, limit = 20): Promise<{ replies: any[], total_count: number, cached?: boolean }> {
  try {
    console.log(`getReplyReplies: Fetching replies for reply ${replyId}, page=${page}, limit=${limit}`);
    
    // Check if we have cached replies for this parent
    const cacheKey = `reply_replies_${replyId}_${page}_${limit}`;
    const cachedData = localStorage.getItem(cacheKey);
    let useCache = false;
    
    if (cachedData) {
      try {
        const { data, timestamp } = JSON.parse(cachedData);
        // Use cache if it's less than 30 seconds old
        if (Date.now() - timestamp < 30000) {
          console.log(`getReplyReplies: Using cached replies for ${replyId}`);
          useCache = true;
          return { ...data, cached: true };
        }
      } catch (error) {
        console.error('Error parsing cached reply replies:', error);
        // Continue with API call if cache parsing fails
      }
    }
    
    const token = getAuthToken();
    
    // Set up headers - allow unauthenticated access but add auth if available
    const headers: Record<string, string> = {
      "Content-Type": "application/json",
    };
    
    if (token) {
      headers["Authorization"] = `Bearer ${token}`;
    }
    
    // Create controller for timeout management
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), 8000); // 8 second timeout
    
    console.log(`getReplyReplies: Making API request to GET /replies/${replyId}/replies`);
    const response = await fetch(`${API_BASE_URL}/replies/${replyId}/replies?page=${page}&limit=${limit}`, {
      method: "GET",
      headers: headers,
      credentials: "include",
      signal: controller.signal
    });
    
    clearTimeout(timeoutId);
    
    if (response.ok) {
      const data = await response.json();
      console.log(`getReplyReplies: Got ${data.replies?.length || 0} replies for reply ${replyId}`);
      
      // Process the replies to ensure consistent field names
      if (data.replies && data.replies.length > 0) {
        data.replies = data.replies.map(reply => {
          // Create a standard reply object with consistent field names
          const standardizedReply: any = {
            id: reply.id || reply.reply_id,
            threadId: reply.thread_id,
            content: reply.content || '',
            timestamp: reply.created_at || new Date().toISOString(),
            userId: reply.user_id || reply.author_id,
            username: '',
            displayName: '',
            avatar: null,
            likes: reply.likes_count || reply.like_count || 0,
            replies: reply.replies_count || reply.reply_count || 0,
            bookmarks: reply.bookmarks_count || reply.bookmark_count || 0,
            isLiked: reply.is_liked || false,
            isBookmarked: reply.is_bookmarked || false,
            
            // Additional fields for frontend compatibility
            author_username: '',
            author_name: '',
            author_avatar: null
          };
          
          // Extract user info from all possible locations
          if (reply.user) {
            standardizedReply.username = reply.user.username;
            standardizedReply.displayName = reply.user.name;
            standardizedReply.avatar = reply.user.profile_picture_url;
          }
          
          // For UI convenience, include these fields as well
          standardizedReply.author_username = standardizedReply.username;
          standardizedReply.author_name = standardizedReply.displayName;
          standardizedReply.author_avatar = standardizedReply.avatar;
          
          return standardizedReply;
        });
      }
      
      // Cache the replies for 30 seconds
      try {
        localStorage.setItem(cacheKey, JSON.stringify({
          data,
          timestamp: Date.now()
        }));
      } catch (error) {
        console.error('Error caching reply replies:', error);
        // Continue even if caching fails
      }
      
      return data;
    }
    
    // Handle specific HTTP status codes
    if (response.status === 401) {
      console.warn("Unauthorized when fetching reply replies - returning empty results");
      return { 
        replies: [],
        total_count: 0
      };
    }
    
    if (response.status === 404) {
      console.warn(`Parent reply ${replyId} not found`);
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
        `Failed to fetch reply replies: ${response.status} ${response.statusText}`
      );
    } catch (parseError) {
      throw new Error(`Failed to fetch reply replies: ${response.status} ${response.statusText}`);
    }
  } catch (error) {
    if (error instanceof DOMException && error.name === 'AbortError') {
      console.error(`Request timeout fetching replies for reply ${replyId}`);
      throw new Error("Request timed out. Please try again.");
    }
    
    console.error("Error in getReplyReplies:", error);
    throw error;
  }
}