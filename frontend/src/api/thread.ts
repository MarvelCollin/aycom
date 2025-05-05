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
    if (userId === 'me') {
      const currentUserId = getUserId();
      if (!currentUserId) {
        throw new Error('User ID isa required');
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

export async function getAllThreads(page = 1, limit = 20) {
  const token = getAuthToken();
  
  try {
    console.log(`Fetching all threads, page: ${page}, limit: ${limit}`);
    
    // There's a public endpoint for getting all threads
    const endpoint = `${API_BASE_URL}/threads?page=${page}&limit=${limit}`;
    console.log(`Making request to: ${endpoint}`);
    
    // Set up headers - allow unauthenticated access
    const headers: Record<string, string> = {
      "Content-Type": "application/json",
    };
    
    // Add authorization header if token exists
    if (token) {
      headers["Authorization"] = `Bearer ${token}`;
    }
    
    const response = await fetch(endpoint, {
      method: "GET",
      headers: headers,
      credentials: "include",
    });
    
    // For 401 unauthorized, we could still attempt to return mock data
    // This keeps the UI working even if the backend requires auth
    if (response.status === 401) {
      console.warn("Unauthorized when fetching threads - returning empty results");
      return { 
        threads: [],
        total_count: 0,
        page: page,
        limit: limit
      };
    }
    
    if (!response.ok) {
      let errorMessage = `Failed to fetch threads: ${response.status} ${response.statusText}`;
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
    console.error("Error in getAllThreads:", error);
    // Return empty results instead of throwing to keep UI working
    return { 
      threads: [],
      total_count: 0,
      page: page,
      limit: limit
    };
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

// Social Features

export async function likeThread(threadId: string) {
  try {
    const token = getAuthToken();
    
    if (!token) {
      console.error("No auth token available for liking thread");
      throw new Error("Authentication required");
    }
    
    // Add retry logic with backoff
    let retries = 0;
    const maxRetries = 2;
    
    while (retries <= maxRetries) {
      try {
        const response = await fetch(`${API_BASE_URL}/threads/${threadId}/like`, {
          method: "POST",
          headers: { 
            "Content-Type": "application/json",
            "Authorization": `Bearer ${token}`
          },
          credentials: "include",
        });
        
        if (response.ok) {
          return response.json();
        }
        
        // If server error, try again
        if (response.status >= 500 && retries < maxRetries) {
          retries++;
          // Wait before retrying (exponential backoff)
          await new Promise(resolve => setTimeout(resolve, 1000 * Math.pow(2, retries)));
          continue;
        }
        
        // Handle error response
        try {
          const errorData = await response.json();
          throw new Error(
            errorData.message || 
            errorData.error?.message || 
            `Failed to like thread: ${response.status} ${response.statusText}`
          );
        } catch (parseError) {
          throw new Error(`Failed to like thread: ${response.status} ${response.statusText}`);
        }
      } catch (fetchError) {
        // If network error and we can retry, do so
        if (retries < maxRetries) {
          retries++;
          await new Promise(resolve => setTimeout(resolve, 1000 * Math.pow(2, retries)));
          continue;
        }
        throw fetchError;
      }
    }
    
    throw new Error("Failed to like thread after multiple attempts");
  } catch (error) {
    console.error("Error in likeThread:", error);
    throw error;
  }
}

export async function unlikeThread(threadId: string) {
  try {
    const token = getAuthToken();
    
    if (!token) {
      console.error("No auth token available for unliking thread");
      throw new Error("Authentication required");
    }
    
    // Add retry logic with backoff
    let retries = 0;
    const maxRetries = 2;
    
    while (retries <= maxRetries) {
      try {
        const response = await fetch(`${API_BASE_URL}/threads/${threadId}/like`, {
          method: "DELETE",
          headers: { 
            "Content-Type": "application/json",
            "Authorization": `Bearer ${token}`
          },
          credentials: "include",
        });
        
        if (response.ok) {
          return response.json();
        }
        
        // If server error, try again
        if (response.status >= 500 && retries < maxRetries) {
          retries++;
          // Wait before retrying (exponential backoff)
          await new Promise(resolve => setTimeout(resolve, 1000 * Math.pow(2, retries)));
          continue;
        }
        
        // Handle error response
        try {
          const errorData = await response.json();
          throw new Error(
            errorData.message || 
            errorData.error?.message || 
            `Failed to unlike thread: ${response.status} ${response.statusText}`
          );
        } catch (parseError) {
          throw new Error(`Failed to unlike thread: ${response.status} ${response.statusText}`);
        }
      } catch (fetchError) {
        // If network error and we can retry, do so
        if (retries < maxRetries) {
          retries++;
          await new Promise(resolve => setTimeout(resolve, 1000 * Math.pow(2, retries)));
          continue;
        }
        throw fetchError;
      }
    }
    
    throw new Error("Failed to unlike thread after multiple attempts");
  } catch (error) {
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
    
    // Add retry logic with backoff
    let retries = 0;
    const maxRetries = 2;
    
    while (retries <= maxRetries) {
      try {
        const response = await fetch(`${API_BASE_URL}/threads/${threadId}/replies`, {
          method: "POST",
          headers: { 
            "Content-Type": "application/json",
            "Authorization": `Bearer ${token}`
          },
          body: JSON.stringify(data),
          credentials: "include",
        });
        
        if (response.ok) {
          return response.json();
        }
        
        // If server error, try again
        if (response.status >= 500 && retries < maxRetries) {
          retries++;
          // Wait before retrying (exponential backoff)
          await new Promise(resolve => setTimeout(resolve, 1000 * Math.pow(2, retries)));
          continue;
        }
        
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
      } catch (fetchError) {
        // If network error and we can retry, do so
        if (retries < maxRetries) {
          retries++;
          await new Promise(resolve => setTimeout(resolve, 1000 * Math.pow(2, retries)));
          continue;
        }
        throw fetchError;
      }
    }
    
    throw new Error("Failed to reply to thread after multiple attempts");
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
    
    // Add retry logic with backoff
    let retries = 0;
    const maxRetries = 2;
    
    while (retries <= maxRetries) {
      try {
        const response = await fetch(`${API_BASE_URL}/threads/${threadId}/replies`, {
          method: "GET",
          headers: headers,
          credentials: "include",
        });
        
        if (response.ok) {
          return response.json();
        }
        
        // If server error, try again
        if (response.status >= 500 && retries < maxRetries) {
          retries++;
          // Wait before retrying (exponential backoff)
          await new Promise(resolve => setTimeout(resolve, 1000 * Math.pow(2, retries)));
          continue;
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
      } catch (fetchError) {
        // If network error and we can retry, do so
        if (retries < maxRetries) {
          retries++;
          await new Promise(resolve => setTimeout(resolve, 1000 * Math.pow(2, retries)));
          continue;
        }
        throw fetchError;
      }
    }
    
    throw new Error("Failed to fetch replies after multiple attempts");
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
  try {
    const token = getAuthToken();
    
    if (!token) {
      console.error("No auth token available for bookmarking thread");
      throw new Error("Authentication required");
    }
    
    // Add retry logic with backoff
    let retries = 0;
    const maxRetries = 2;
    
    while (retries <= maxRetries) {
      try {
        const response = await fetch(`${API_BASE_URL}/threads/${threadId}/bookmark`, {
          method: "POST",
          headers: { 
            "Content-Type": "application/json",
            "Authorization": `Bearer ${token}`
          },
          credentials: "include",
        });
        
        if (response.ok) {
          return response.json();
        }
        
        // If server error, try again
        if (response.status >= 500 && retries < maxRetries) {
          retries++;
          // Wait before retrying (exponential backoff)
          await new Promise(resolve => setTimeout(resolve, 1000 * Math.pow(2, retries)));
          continue;
        }
        
        // Handle error response
        try {
          const errorData = await response.json();
          throw new Error(
            errorData.message || 
            errorData.error?.message || 
            `Failed to bookmark thread: ${response.status} ${response.statusText}`
          );
        } catch (parseError) {
          throw new Error(`Failed to bookmark thread: ${response.status} ${response.statusText}`);
        }
      } catch (fetchError) {
        // If network error and we can retry, do so
        if (retries < maxRetries) {
          retries++;
          await new Promise(resolve => setTimeout(resolve, 1000 * Math.pow(2, retries)));
          continue;
        }
        throw fetchError;
      }
    }
    
    throw new Error("Failed to bookmark thread after multiple attempts");
  } catch (error) {
    console.error("Error in bookmarkThread:", error);
    throw error;
  }
}

export async function removeBookmark(threadId: string) {
  try {
    const token = getAuthToken();
    
    if (!token) {
      console.error("No auth token available for removing bookmark");
      throw new Error("Authentication required");
    }
    
    console.log(`Attempting to remove bookmark for thread: ${threadId}`);
    
    // Add retry logic with backoff
    let retries = 0;
    const maxRetries = 2;
    
    while (retries <= maxRetries) {
      try {
        // Ensure we're sending DELETE request to the correct endpoint
        const endpoint = `${API_BASE_URL}/threads/${threadId}/bookmark`;
        console.log(`Sending DELETE request to endpoint: ${endpoint}`);
        
        const response = await fetch(endpoint, {
          method: "DELETE",
          headers: { 
            "Content-Type": "application/json",
            "Authorization": `Bearer ${token}`
          },
          credentials: "include",
        });
        
        // Log the response status
        console.log(`Remove bookmark response status: ${response.status}`);
        
        if (response.ok) {
          console.log("Successfully removed bookmark");
          // Try to parse JSON, but handle case where response might be empty
          try {
            return await response.json();
          } catch (jsonError) {
            console.log("Empty response body from remove bookmark");
            return { success: true };
          }
        }
        
        // Try to get response text for better debugging
        let responseText = '';
        try {
          responseText = await response.text();
          console.error(`Error response from server: ${responseText}`);
        } catch (textError) {
          console.error("Could not read response text");
        }
        
        // If server error, try again
        if (response.status >= 500 && retries < maxRetries) {
          console.log(`Server error (${response.status}), retrying...`);
          retries++;
          // Wait before retrying (exponential backoff)
          await new Promise(resolve => setTimeout(resolve, 1000 * Math.pow(2, retries)));
          continue;
        }
        
        // Handle error response
        try {
          let errorMessage = `Failed to remove bookmark: ${response.status} ${response.statusText}`;
          
          // Try to parse the response text as JSON if possible
          if (responseText) {
            try {
              const errorData = JSON.parse(responseText);
              errorMessage = errorData.message || 
                            errorData.error?.message || 
                            errorMessage;
            } catch (e) {
              // If not valid JSON, use the raw text in the error
              errorMessage = `${errorMessage}. Response: ${responseText}`;
            }
          }
          
          throw new Error(errorMessage);
        } catch (parseError) {
          throw new Error(`Failed to remove bookmark: ${response.status} ${response.statusText}`);
        }
      } catch (fetchError) {
        // If network error and we can retry, do so
        if (retries < maxRetries) {
          console.log(`Network error, retrying (${retries + 1}/${maxRetries})...`);
          retries++;
          await new Promise(resolve => setTimeout(resolve, 1000 * Math.pow(2, retries)));
          continue;
        }
        throw fetchError;
      }
    }
    
    throw new Error("Failed to remove bookmark after multiple attempts");
  } catch (error) {
    console.error("Error in removeBookmark:", error);
    throw error;
  }
}

/**
 * Fetches threads from users that the current user follows
 * @param page Page number to fetch (1-based)
 * @param limit Number of threads per page
 * @returns Object containing threads array and pagination info
 */
export async function getFollowingThreads(page = 1, limit = 20) {
  const token = getAuthToken();
  
  try {
    console.log(`Fetching followed users threads, page: ${page}, limit: ${limit}`);
    
    // Endpoint for getting threads from followed users
    const endpoint = `${API_BASE_URL}/threads/following?page=${page}&limit=${limit}`;
    console.log(`Making request to: ${endpoint}`);
    
    // Authorization is required to know which users you follow
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
    
    // For 401 unauthorized, return empty results
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
    const url = new URL(`${import.meta.env.VITE_API_BASE_URL || 'http://localhost:8083/api/v1'}/threads/search`);
    
    // Set query parameters
    url.searchParams.append('q', query);
    url.searchParams.append('page', page.toString());
    url.searchParams.append('limit', limit.toString());
    
    // Add optional filters
    if (options?.filter) {
      url.searchParams.append('filter', options.filter);
    }
    
    // Add category if provided
    if (options?.category) {
      url.searchParams.append('category', options.category);
    }
    
    // Add sorting if provided
    if (options?.sortBy) {
      url.searchParams.append('sort', options.sortBy);
    }
    
    // Get token
    const token = localStorage.getItem('aycom_access_token');
    
    // Make request
    const response = await fetch(url.toString(), {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      }
    });
    
    if (!response.ok) {
      throw new Error(`Failed to search threads: ${response.status}`);
    }
    
    return await response.json();
  } catch (error) {
    console.error('Error searching threads:', error);
    // Mock data for development
    return {
      threads: [
        {
          id: '101',
          content: 'Just had a great time exploring the new features of this platform! #tech #exploration',
          username: 'johndoe',
          display_name: 'John Doe',
          created_at: new Date(Date.now() - 3600000).toISOString(),
          like_count: 42,
          reply_count: 7,
          repost_count: 5,
          media: []
        },
        {
          id: '102',
          content: 'The search functionality is really impressive! #search',
          username: 'janedoe',
          display_name: 'Jane Doe',
          created_at: new Date(Date.now() - 7200000).toISOString(),
          like_count: 23,
          reply_count: 3,
          repost_count: 1,
          media: []
        }
      ]
    };
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
    const url = new URL(`${import.meta.env.VITE_API_BASE_URL || 'http://localhost:8083/api/v1'}/threads/search/media`);
    
    // Set query parameters
    url.searchParams.append('q', query);
    url.searchParams.append('page', page.toString());
    url.searchParams.append('limit', limit.toString());
    
    // Add optional filters
    if (options?.filter) {
      url.searchParams.append('filter', options.filter);
    }
    
    // Add category if provided
    if (options?.category) {
      url.searchParams.append('category', options.category);
    }
    
    // Get token
    const token = localStorage.getItem('aycom_access_token');
    
    // Make request
    const response = await fetch(url.toString(), {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      }
    });
    
    if (!response.ok) {
      throw new Error(`Failed to search threads with media: ${response.status}`);
    }
    
    return await response.json();
  } catch (error) {
    console.error('Error searching threads with media:', error);
    // Mock data for development
    return {
      threads: [
        {
          id: '201',
          content: 'Check out this amazing view! #travel',
          username: 'traveler',
          display_name: 'Travel Enthusiast',
          created_at: new Date(Date.now() - 3600000).toISOString(),
          like_count: 157,
          reply_count: 23,
          repost_count: 42,
          media: [
            {
              media_id: 'm1',
              url: 'https://images.unsplash.com/photo-1506744038136-46273834b3fb',
              type: 'image'
            }
          ]
        },
        {
          id: '202',
          content: 'New tutorial on web development! #webdev #coding',
          username: 'coder',
          display_name: 'Coding Expert',
          created_at: new Date(Date.now() - 7200000).toISOString(),
          like_count: 89,
          reply_count: 12,
          repost_count: 14,
          media: [
            {
              media_id: 'm2',
              url: 'https://images.unsplash.com/photo-1498050108023-c5249f4df085',
              type: 'image'
            }
          ]
        }
      ]
    };
  }
}

// Get threads by hashtag
export async function getThreadsByHashtag(
  hashtag: string, 
  page: number = 1, 
  limit: number = 10
) {
  try {
    // Remove # if it's included
    const tag = hashtag.startsWith('#') ? hashtag.substring(1) : hashtag;
    
    const url = new URL(`${import.meta.env.VITE_API_BASE_URL || 'http://localhost:8083/api/v1'}/threads/hashtag/${tag}`);
    
    // Set query parameters
    url.searchParams.append('page', page.toString());
    url.searchParams.append('limit', limit.toString());
    
    // Get token
    const token = localStorage.getItem('aycom_access_token');
    
    // Make request
    const response = await fetch(url.toString(), {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      }
    });
    
    if (!response.ok) {
      throw new Error(`Failed to get threads by hashtag: ${response.status}`);
    }
    
    return await response.json();
  } catch (error) {
    console.error('Error getting threads by hashtag:', error);
    // Mock data for development
    return {
      threads: [
        {
          id: '301',
          content: `Exploring the ${hashtag} trend! What do you think?`,
          username: 'trendwatcher',
          display_name: 'Trend Watcher',
          created_at: new Date(Date.now() - 3600000).toISOString(),
          like_count: 78,
          reply_count: 14,
          repost_count: 8,
          media: []
        },
        {
          id: '302',
          content: `Let's talk about ${hashtag} and why it's trending today.`,
          username: 'analyzer',
          display_name: 'Trend Analyzer',
          created_at: new Date(Date.now() - 7200000).toISOString(),
          like_count: 45,
          reply_count: 9,
          repost_count: 3,
          media: []
        }
      ]
    };
  }
}