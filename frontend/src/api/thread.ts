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

export async function getThreadsByUser(userId: string, page: number = 1, limit: number = 10) {
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
      // Handle different error types
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
    
    // Debug output to check pinned status
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
  
  // Append each file with a unique name
  files.forEach((file, index) => {
    formData.append(`file`, file); // Changed from 'media_${index}' to 'file' to match backend expectation
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
    // Build the URL with query parameters
    const url = new URL(`${API_BASE_URL}/threads/search`);
    url.searchParams.append('q', query);
    url.searchParams.append('page', page.toString());
    url.searchParams.append('limit', limit.toString());
    
    // Add optional parameters if they exist
    if (options?.filter) {
      url.searchParams.append('filter', options.filter);
    }
    
    if (options?.category) {
      url.searchParams.append('category', options.category);
    }
    
    if (options?.sortBy) {
      url.searchParams.append('sort', options.sortBy);
    }
    
    // Get the auth token
    const token = getAuthToken();
    
    // Make the request with the token
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
  } catch (error) {
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
    const token = getAuthToken();
    
    // Make request
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
  } catch (error) {
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
    // Remove the hashtag symbol if it exists
    const cleanHashtag = hashtag.startsWith('#') ? hashtag.substring(1) : hashtag;
    
    const url = new URL(`${API_BASE_URL}/threads/hashtag/${encodeURIComponent(cleanHashtag)}`);
    
    // Add pagination
    url.searchParams.append('page', page.toString());
    url.searchParams.append('limit', limit.toString());
    
    // Get token
    const token = getAuthToken();
    
    // Make request
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
  } catch (error) {
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
    
    // Add retry logic with backoff
    let retries = 0;
    const maxRetries = 2;
    
    while (retries <= maxRetries) {
      try {
        const response = await fetch(`${API_BASE_URL}/replies/${replyId}/like`, {
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
            `Failed to like reply: ${response.status} ${response.statusText}`
          );
        } catch (parseError) {
          throw new Error(`Failed to like reply: ${response.status} ${response.statusText}`);
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
    
    throw new Error("Failed to like reply after multiple attempts");
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
    
    // Add retry logic with backoff
    let retries = 0;
    const maxRetries = 2;
    
    while (retries <= maxRetries) {
      try {
        const response = await fetch(`${API_BASE_URL}/replies/${replyId}/like`, {
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
            `Failed to unlike reply: ${response.status} ${response.statusText}`
          );
        } catch (parseError) {
          throw new Error(`Failed to unlike reply: ${response.status} ${response.statusText}`);
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
    
    throw new Error("Failed to unlike reply after multiple attempts");
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
    
    // Add retry logic with backoff
    let retries = 0;
    const maxRetries = 2;
    
    while (retries <= maxRetries) {
      try {
        const response = await fetch(`${API_BASE_URL}/replies/${replyId}/bookmark`, {
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
            `Failed to bookmark reply: ${response.status} ${response.statusText}`
          );
        } catch (parseError) {
          throw new Error(`Failed to bookmark reply: ${response.status} ${response.statusText}`);
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
    
    throw new Error("Failed to bookmark reply after multiple attempts");
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
    
    // Add retry logic with backoff
    let retries = 0;
    const maxRetries = 2;
    
    while (retries <= maxRetries) {
      try {
        const response = await fetch(`${API_BASE_URL}/replies/${replyId}/bookmark`, {
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
            `Failed to remove reply bookmark: ${response.status} ${response.statusText}`
          );
        } catch (parseError) {
          throw new Error(`Failed to remove reply bookmark: ${response.status} ${response.statusText}`);
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
    
    throw new Error("Failed to remove reply bookmark after multiple attempts");
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
      // Handle different error types
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
    
    return await response.json();
  } catch (err) {
    console.error('Error getting user threads:', err);
    throw err;
  }
}

export async function getUserReplies(userId: string, page: number = 1, limit: number = 10) {
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
      // Handle different error types
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
    
    // If 'me' is specified, get the actual user ID
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
      // Handle different error types
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
      throw new Error(`Failed to parse bookmarks response: ${error.message}`);
    }
  } catch (error) {
    console.error('Error in getUserBookmarks:', error);
    // Return empty bookmarks structure rather than throwing
    return { bookmarks: [] };
  }
}