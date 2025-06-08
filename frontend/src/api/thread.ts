import { getAuthToken, getUserId, getAuthData } from '../utils/auth';
import appConfig from '../config/appConfig';
import { uploadMultipleThreadMedia } from '../utils/supabase';
import { createLoggerWithPrefix } from '../utils/logger';
import { useAuth } from '../hooks/useAuth';

const API_BASE_URL = appConfig.api.baseUrl;
const AI_SERVICE_URL = appConfig.api.aiServiceUrl || 'http://localhost:5000';
const logger = createLoggerWithPrefix('ThreadAPI');

logger.info('Thread API using URL:', API_BASE_URL);

async function handleApiResponse(response: Response, errorMessage: string = 'API request failed') {
  if (!response.ok) {
    try {
      const errorData = await response.json();
      throw new Error(
        errorData.message || 
        errorData.error?.message || 
        `${errorMessage}: ${response.status} ${response.statusText}`
      );
    } catch (parseError) {
      throw new Error(`${errorMessage}: ${response.status} ${response.statusText}`);
    }
  }

  return response.json();
}

// Check if token needs refresh before making API request
async function ensureValidToken() {
  try {
    const { checkAndRefreshTokenIfNeeded } = useAuth();
    await checkAndRefreshTokenIfNeeded();
  } catch (error) {
    logger.warn('Error ensuring token freshness:', error);
    // Continue with the request even if token refresh fails
    // The server will respond with 401 if necessary
  }
}

async function makeApiRequest(url: string, method: string, body?: any, errorMessage?: string, timeout: number = 15000) {
  try {
    // For GET requests that can work without authentication, we'll try but not fail if auth refresh fails
    const isPublicReadRequest = method === 'GET' && url.includes('/threads');
    
    try {
      const { checkAndRefreshTokenIfNeeded } = useAuth();
      await checkAndRefreshTokenIfNeeded();
    } catch (error) {
      // If this is a public read request, continue without a token
      // Otherwise re-throw the error
      if (!isPublicReadRequest) {
        throw error;
      }
    }
    
    const token = getAuthToken();
    
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), timeout);

    try {
      const options: RequestInit = {
        method,
        headers: {
          'Content-Type': 'application/json',
          'Authorization': token ? `Bearer ${token}` : ''
        },
        credentials: 'include',
        signal: controller.signal
      };

      if (body) {
        options.body = JSON.stringify(body);
      }

      const response = await fetch(url, options);
      clearTimeout(timeoutId);
      
      // For 401 responses on public endpoints, try again without auth token
      if (response.status === 401 && isPublicReadRequest && token) {
        logger.warn('Got 401 on public endpoint, retrying without auth');
        const publicOptions = { ...options };
        publicOptions.headers = { 'Content-Type': 'application/json' };
        
        const publicResponse = await fetch(url, publicOptions);
        return await handleApiResponse(publicResponse, errorMessage);
      }
      
      return await handleApiResponse(response, errorMessage);
    } catch (error: any) {
      clearTimeout(timeoutId);
      if (error.name === 'AbortError') {
        throw new Error('Request timed out');
      }
      throw error;
    }
  } catch (error: any) {
    logger.error(`API request failed: ${error.message}`);
    throw error;
  }
}

export async function createThread(data: Record<string, any>) {
  try {
    return await makeApiRequest(
      `${API_BASE_URL}/threads`, 
      'POST', 
      data, 
      'Failed to create thread'
    );
  } catch (error) {
    logger.error('Create thread failed:', error);
    throw error;
  }
}

export async function getThread(id: string) {
  try {
    return await makeApiRequest(
      `${API_BASE_URL}/threads/${id}`, 
      'GET', 
      null, 
      'Failed to fetch thread'
    );
  } catch (error) {
    logger.error(`Get thread ${id} failed:`, error);
    throw error;
  }
}

export async function getThreadsByUser(userId: string, page: number = 1, limit: number = 10) {
  try {
    let actualUserId = userId;

    if (userId === 'me') {
      const currentUserId = getUserId();
      logger.debug('Current user ID from auth:', currentUserId);

      if (!currentUserId) {
        throw new Error('User ID is required');
      }
      actualUserId = currentUserId;
    }

    const endpoint = `${API_BASE_URL}/threads/user/${actualUserId}?page=${page}&limit=${limit}`;
    logger.debug(`Making request to: ${endpoint}`);

    const data = await makeApiRequest(
      endpoint, 
      'GET', 
      null, 
      'Failed to get user threads'
    );

    logger.info(`Received ${data.threads?.length || 0} threads for user ${actualUserId}`);
    logger.debug("Pinned threads:", data.threads?.filter(thread => thread.is_pinned === true).length);

    return data;
  } catch (error) {
    logger.error('Error getting user threads:', error);
    throw error;
  }
}

interface ThreadMedia {
  id: string;
  url: string;
  type: string;
  thumbnail?: string;
}

interface Thread {
  id: string;
  content: string;
  created_at: string;
  updated_at: string;
  user_id: string;
  username: string;
  name: string;
  profile_picture_url: string;
  likes_count: number;
  replies_count: number;
  reposts_count: number;
  is_liked: boolean;
  is_reposted: boolean;
  is_bookmarked: boolean;
  is_pinned: boolean;
  media: ThreadMedia[];
}

interface ThreadsResponse {
  success: boolean;
  threads: Thread[];
  total_count: number;
  total?: number;
}

export async function getAllThreads(page = 1, limit = 10): Promise<ThreadsResponse> {
  try {
    const url = `${API_BASE_URL}/threads?page=${page}&limit=${limit}`;

    try {
      const data = await makeApiRequest(
        url, 
        'GET', 
        null, 
        'Failed to fetch threads'
      );

      logger.info(`getAllThreads received ${data.threads?.length || 0} threads`);

      if (!data.threads || !Array.isArray(data.threads)) {
        logger.warn('API returned invalid threads data structure', data);
        return {
          success: false,
          threads: [],
          total_count: 0
        };
      }

      const totalCount = data.total_count !== undefined ? data.total_count : data.total || 0;

      return {
        success: true,
        threads: data.threads || [],
        total_count: totalCount
      };
    } catch (error) {
      logger.error('Fetch error in getAllThreads:', error);
      throw error;
    }
  } catch (error) {
    logger.error('Get all threads failed:', error);
    throw error;
  }
}

export async function updateThread(id: string, data: Record<string, any>) {
  try {
    return await makeApiRequest(
      `${API_BASE_URL}/threads/${id}`, 
      'PUT', 
      data, 
      'Failed to update thread'
    );
  } catch (error) {
    logger.error(`Update thread ${id} failed:`, error);
    throw error;
  }
}

export async function deleteThread(id: string) {
  try {
    return await makeApiRequest(
      `${API_BASE_URL}/threads/${id}`, 
      'DELETE', 
      null, 
      'Failed to delete thread'
    );
  } catch (error) {
    logger.error(`Delete thread ${id} failed:`, error);
    throw error;
  }
}

export async function uploadThreadMedia(threadId: string, files: File[]) {
  try {

    const urls = await uploadMultipleThreadMedia(files, threadId);

    if (urls && urls.length > 0) {
      const token = getAuthToken();

      const response = await fetch(`${API_BASE_URL}/threads/${threadId}/media/update`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': token ? `Bearer ${token}` : ''
        },
        body: JSON.stringify({ media_urls: urls }),
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

// Prevent multiple rapid like/unlike requests
let likeDebounceMap = new Map();
const DEBOUNCE_DELAY = 500; // ms

export async function likeThread(threadId: string) {
  try {
    // Check for existing ongoing request
    if (likeDebounceMap.has(threadId)) {
      logger.warn(`Like operation for thread ${threadId} already in progress, skipping`);
      return { success: false, message: 'Operation already in progress' };
    }

    // Set debounce lock
    likeDebounceMap.set(threadId, true);
    
    // Clear lock after delay regardless of outcome
    setTimeout(() => {
      likeDebounceMap.delete(threadId);
    }, DEBOUNCE_DELAY);

    const token = getAuthToken();
    
    // Check if token exists - this is critical
    if (!token) {
      logger.error(`Cannot like thread ${threadId}: No auth token available`);
      throw new Error('Authentication required. Please log in again.');
    }

    logger.debug(`Sending like request for thread ${threadId} with token length: ${token.length}`);
    
    const response = await fetch(`${API_BASE_URL}/threads/${threadId}/like`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Authorization": `Bearer ${token}`
      },
      credentials: "include"
    });

    if (!response.ok) {
      // Check specifically for auth errors
      if (response.status === 401) {
        logger.error(`Authentication failed when liking thread ${threadId} - token may be invalid`);
        throw new Error('Your session has expired. Please log in again.');
      }
      
      const errorData = await response.json();
      logger.error(`Error liking thread ${threadId}:`, errorData);
      throw new Error(errorData.message || 'Failed to like thread');
    }

    const data = await response.json();
    logger.debug(`Successfully liked thread ${threadId}`);
    return { ...data, success: true };
  } catch (error: any) {
    logger.error(`Like thread ${threadId} failed:`, error);
    throw error;
  } finally {
    // Ensure lock is cleared in case of early return
    setTimeout(() => {
      likeDebounceMap.delete(threadId);
    }, 50);
  }
}

export async function unlikeThread(threadId: string) {
  try {
    // Check for existing ongoing request
    if (likeDebounceMap.has(threadId)) {
      logger.warn(`Unlike operation for thread ${threadId} already in progress, skipping`);
      return { success: false, message: 'Operation already in progress' };
    }

    // Set debounce lock
    likeDebounceMap.set(threadId, true);
    
    // Clear lock after delay regardless of outcome
    setTimeout(() => {
      likeDebounceMap.delete(threadId);
    }, DEBOUNCE_DELAY);

    const token = getAuthToken();
    
    // Check if token exists - this is critical
    if (!token) {
      logger.error(`Cannot unlike thread ${threadId}: No auth token available`);
      throw new Error('Authentication required. Please log in again.');
    }
    
    logger.debug(`Sending unlike request for thread ${threadId} with token length: ${token.length}`);

    const response = await fetch(`${API_BASE_URL}/threads/${threadId}/like`, {
      method: "DELETE",
      headers: {
        "Content-Type": "application/json",
        "Authorization": `Bearer ${token}`
      },
      credentials: "include"
    });

    if (!response.ok) {
      // Check specifically for auth errors
      if (response.status === 401) {
        logger.error(`Authentication failed when unliking thread ${threadId} - token may be invalid`);
        throw new Error('Your session has expired. Please log in again.');
      }
      
      const errorData = await response.json();
      logger.error(`Error unliking thread ${threadId}:`, errorData);
      throw new Error(errorData.message || 'Failed to unlike thread');
    }

    const data = await response.json();
    logger.debug(`Successfully unliked thread ${threadId}`);
    return { ...data, success: true };
  } catch (error) {
    logger.error(`Unlike thread ${threadId} failed:`, error);
    throw error;
  } finally {
    // Ensure lock is cleared in case of early return
    setTimeout(() => {
      likeDebounceMap.delete(threadId);
    }, 50);
  }
}

export async function replyToThread(threadId: string, data: {
  content: string;
  media?: any[];
  parent_reply_id?: string;
  mentioned_user_ids?: string[];
}) {
  try {
    return await makeApiRequest(
      `${API_BASE_URL}/threads/${threadId}/replies`, 
      'POST', 
      data, 
      'Failed to post reply'
    );
  } catch (error) {
    logger.error(`Reply to thread ${threadId} failed:`, error);
    throw error;
  }
}

export async function getThreadReplies(threadId: string) {
  try {
    const response = await makeApiRequest(
      `${API_BASE_URL}/threads/${threadId}/replies`, 
      'GET', 
      null, 
      'Failed to get thread replies'
    );
    
    // Process replies to ensure they have replies_count set
    if (response && response.replies && Array.isArray(response.replies)) {
      response.replies = response.replies.map(reply => {
        // Make sure replies_count is set to 0 if not present
        if (reply.replies_count === undefined && reply.repliesCount === undefined) {
          reply.replies_count = 0;
        }
        return reply;
      });
    }
    
    return response;
  } catch (error) {
    logger.error(`Get replies for thread ${threadId} failed:`, error);
    throw error;
  }
}

export async function repostThread(threadId: string, content = '') {
  try {
    return await makeApiRequest(
      `${API_BASE_URL}/threads/${threadId}/reposts`, 
      'POST', 
      { content }, 
      'Failed to repost thread'
    );
  } catch (error) {
    logger.error(`Repost thread ${threadId} failed:`, error);
    throw error;
  }
}

export async function removeRepost(repostId: string) {
  try {
    return await makeApiRequest(
      `${API_BASE_URL}/threads/reposts/${repostId}`, 
      'DELETE', 
      null, 
      'Failed to remove repost'
    );
  } catch (error) {
    logger.error(`Remove repost ${repostId} failed:`, error);
    throw error;
  }
}

// Prevent multiple rapid bookmark/unbookmark requests
let bookmarkDebounceMap = new Map();

export async function bookmarkThread(threadId: string) {
  try {
    // Check for existing ongoing request
    if (bookmarkDebounceMap.has(threadId)) {
      logger.warn(`Bookmark operation for thread ${threadId} already in progress, skipping`);
      return { success: false, message: 'Operation already in progress' };
    }

    // Set debounce lock
    bookmarkDebounceMap.set(threadId, true);
    
    // Clear lock after delay regardless of outcome
    setTimeout(() => {
      bookmarkDebounceMap.delete(threadId);
    }, DEBOUNCE_DELAY);

    const token = getAuthToken();
    
    // Check if token exists - this is critical
    if (!token) {
      logger.error(`Cannot bookmark thread ${threadId}: No auth token available`);
      throw new Error('Authentication required. Please log in again.');
    }
    
    logger.debug(`Sending bookmark request for thread ${threadId} with token length: ${token.length}`);

    const response = await fetch(`${API_BASE_URL}/threads/${threadId}/bookmark`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Authorization": `Bearer ${token}`
      },
      credentials: "include"
    });

    if (!response.ok) {
      // Check specifically for auth errors
      if (response.status === 401) {
        logger.error(`Authentication failed when bookmarking thread ${threadId} - token may be invalid`);
        throw new Error('Your session has expired. Please log in again.');
      }
      
      const errorData = await response.json();
      logger.error(`Error bookmarking thread ${threadId}:`, errorData);
      throw new Error(errorData.message || 'Failed to bookmark thread');
    }

    const data = await response.json();
    logger.debug(`Successfully bookmarked thread ${threadId}`);
    return { ...data, success: true };
  } catch (error) {
    logger.error(`Bookmark thread ${threadId} failed:`, error);
    throw error;
  } finally {
    // Ensure lock is cleared in case of early return
    setTimeout(() => {
      bookmarkDebounceMap.delete(threadId);
    }, 50);
  }
}

export async function removeBookmark(threadId: string) {
  try {
    // Check for existing ongoing request
    if (bookmarkDebounceMap.has(threadId)) {
      logger.warn(`Remove bookmark operation for thread ${threadId} already in progress, skipping`);
      return { success: false, message: 'Operation already in progress' };
    }

    // Set debounce lock
    bookmarkDebounceMap.set(threadId, true);
    
    // Clear lock after delay regardless of outcome
    setTimeout(() => {
      bookmarkDebounceMap.delete(threadId);
    }, DEBOUNCE_DELAY);

    const token = getAuthToken();
    
    // Check if token exists - this is critical
    if (!token) {
      logger.error(`Cannot remove bookmark for thread ${threadId}: No auth token available`);
      throw new Error('Authentication required. Please log in again.');
    }
    
    logger.debug(`Sending remove bookmark request for thread ${threadId} with token length: ${token.length}`);

    const response = await fetch(`${API_BASE_URL}/threads/${threadId}/bookmark`, {
      method: "DELETE",
      headers: {
        "Content-Type": "application/json",
        "Authorization": `Bearer ${token}`
      },
      credentials: "include"
    });

    if (!response.ok) {
      // Check specifically for auth errors
      if (response.status === 401) {
        logger.error(`Authentication failed when removing bookmark for thread ${threadId} - token may be invalid`);
        throw new Error('Your session has expired. Please log in again.');
      }
      
      const errorData = await response.json();
      logger.error(`Error removing bookmark for thread ${threadId}:`, errorData);
      throw new Error(errorData.message || 'Failed to remove bookmark');
    }

    const data = await response.json();
    logger.debug(`Successfully removed bookmark for thread ${threadId}`);
    return { ...data, success: true };
  } catch (error) {
    logger.error(`Remove bookmark for thread ${threadId} failed:`, error);
    throw error;
  } finally {
    // Ensure lock is cleared in case of early return
    setTimeout(() => {
      bookmarkDebounceMap.delete(threadId);
    }, 50);
  }
}

export async function getFollowingThreads(page = 1, limit = 20) {
  try {
    const url = `${API_BASE_URL}/threads/following?page=${page}&limit=${limit}`;

    return await makeApiRequest(
      url, 
      'GET', 
      null, 
      'Failed to fetch following threads'
    );
  } catch (error) {
    logger.error('Get following threads failed:', error);
    throw error;
  }
}

export async function searchThreads(
  query: string, 
  page: number = 1, 
  limit: number = 10, 
  options?: { filter?: string; category?: string; sort_by?: string }
) {
  try {
    const params = new URLSearchParams({
      q: query,
      page: page.toString(),
      limit: limit.toString()
    });

    // Make sure all filters are properly appended to the request
    if (options?.filter) {
      params.append('filter', options.filter);
      console.log(`Using filter: ${options.filter}`);
    }
    if (options?.category) params.append('category', options.category);
    if (options?.sort_by) params.append('sort_by', options.sort_by);

    console.log(`Searching threads with query: ${query}, filter: ${options?.filter || 'all'}, URL: ${API_BASE_URL}/threads/search?${params}`);

    return await makeApiRequest(
      `${API_BASE_URL}/threads/search?${params}`, 
      'GET', 
      null, 
      'Thread search failed'
    );
  } catch (error) {
    logger.error('Search threads failed:', error);
    throw error;
  }
}

export async function searchThreadsWithMedia(
  query: string, 
  page: number = 1, 
  limit: number = 10, 
  options?: { filter?: string; category?: string }
) {
  try {
    // Use the regular thread search endpoint instead, but filter for threads with media
    const url = new URL(`${API_BASE_URL}/threads/search`);

    url.searchParams.append('q', query);
    url.searchParams.append('page', page.toString());
    url.searchParams.append('limit', limit.toString());
    url.searchParams.append('media_only', 'true'); // This will be a hint to the backend

    // Make sure filter is always included if provided
    if (options?.filter) {
      url.searchParams.append('filter', options.filter);
      console.log(`Using filter for media search: ${options.filter}`);
    }

    if (options?.category) {
      url.searchParams.append('category', options.category);
    }

    const token = getAuthToken();
    console.log(`Searching threads with media: ${url.toString()}`);

    const response = await fetch(url.toString(), {
      method: 'GET',
      headers: {
        'Authorization': token ? `Bearer ${token}` : '',
        'Content-Type': 'application/json'
      }
    });

    if (!response.ok) {
      // Don't fall back to empty results, throw an error instead
      throw new Error(`Failed to search threads with media: ${response.status}`);
    }

    const data = await response.json();
    
    // Filter the results to only include threads with media (just in case backend doesn't filter)
    if (data.threads) {
      data.threads = data.threads.filter(thread => thread.media && thread.media.length > 0);
    }

    return data;
  } catch (error: any) {
    console.error('Error searching threads with media:', error);
    throw error;
  }
}

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
    url.searchParams.append('sort_by', 'likes');
    url.searchParams.append('sort_order', 'desc');

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
    
    // Process replies to ensure they have replies_count set
    if (data.replies && Array.isArray(data.replies)) {
      data.replies = data.replies.map(reply => {
        // Make sure replies_count is set to 0 if not present
        if (reply.replies_count === undefined && reply.repliesCount === undefined) {
          reply.replies_count = 0;
        }
        return reply;
      });
    }
    
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

    // Check if token exists
    if (!token) {
      logger.error(`Cannot like reply ${replyId}: No auth token available`);
      throw new Error('Authentication required. Please log in again.');
    }

    const response = await fetch(`${API_BASE_URL}/replies/${replyId}/like`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Authorization": `Bearer ${token}`
      },
      credentials: "include"
    });

    if (!response.ok) {
      if (response.status === 401) {
        logger.error(`Authentication failed when liking reply ${replyId} - token may be invalid`);
        throw new Error('Your session has expired. Please log in again.');
      }
      
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

    // Check if token exists
    if (!token) {
      logger.error(`Cannot unlike reply ${replyId}: No auth token available`);
      throw new Error('Authentication required. Please log in again.');
    }

    const response = await fetch(`${API_BASE_URL}/replies/${replyId}/like`, {
      method: "DELETE",
      headers: {
        "Content-Type": "application/json",
        "Authorization": `Bearer ${token}`
      },
      credentials: "include"
    });

    if (!response.ok) {
      // Check specifically for auth errors
      if (response.status === 401) {
        logger.error(`Authentication failed when unliking reply ${replyId} - token may be invalid`);
        throw new Error('Your session has expired. Please log in again.');
      }
      
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

function isUuid(str: string): boolean {
  const uuidPattern = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;
  return uuidPattern.test(str);
}

async function resolveUserIdIfNeeded(userId: string): Promise<string> {

  if (isUuid(userId)) {
    return userId;
  }

  if (userId === 'me') {
    const token = getAuthToken();
    const currentUserId = token ? JSON.parse(atob(token.split('.')[1])).sub : null;
    return currentUserId || userId;
  }

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
      return userId; 
    }

    const data = await response.json();
    return data.user?.id || userId;
  } catch (error) {
    console.error('Error resolving user ID:', error);
    return userId; 
  }
}

export const getUserThreads = async (userId: string, page = 1, limit = 10): Promise<any> => {
  try {
    const resolvedUserId = await resolveUserIdIfNeeded(userId);
    logger.debug(`Fetching threads for user ${resolvedUserId} (original: ${userId}), page: ${page}, limit: ${limit}`);

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
    logger.debug(`Fetching replies for user ${resolvedUserId} (original: ${userId}), page: ${page}, limit: ${limit}`);

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
    logger.debug(`Making request to: ${endpoint}`);

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
          logger.error("API error response:", errorData);
        } catch (parseError) {
          logger.error("Could not parse error response:", parseError);
        }
        throw new Error(errorMessage);
      }
      throw new Error(`Failed to get user liked threads: ${response.status}`);
    }

    const responseData = await response.json();
    return responseData;
  } catch (err) {
    logger.error('Error getting user liked threads:', err);
    throw err;
  }
};

export const getUserMedia = async (userId: string, page = 1, limit = 10): Promise<any> => {
  try {
    const resolvedUserId = await resolveUserIdIfNeeded(userId);
    logger.debug(`Fetching media for user ${resolvedUserId} (original: ${userId}), page: ${page}, limit: ${limit}`);

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
    
    // Check if token exists - this is critical
    if (!token) {
      logger.error(`Cannot get user bookmarks: No auth token available`);
      throw new Error('Authentication required. Please log in again.');
    }
    
    const actualUserId = userId === 'me' ? getUserId() : userId;

    if (!actualUserId) {
      logger.error('No user ID available, cannot fetch bookmarks');
      throw new Error('User ID is required');
    }

    const url = `${API_BASE_URL}/bookmarks?page=${page}&limit=${limit}`;
    logger.debug(`Fetching bookmarks from: ${url}`);

    const response = await fetch(url, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      credentials: "include"
    });

    if (!response.ok) {
      // Check specifically for auth errors
      if (response.status === 401) {
        logger.error('Authentication failed when getting bookmarks - token may be invalid');
        throw new Error('Your session has expired. Please log in again.');
      }
      
      logger.error(`Failed to get bookmarks: ${response.status}`);
      throw new Error(`Failed to get bookmarks: ${response.status}`);
    }

    const data = await response.json();
    logger.debug('Bookmarks API returned data:', data);

    return {
      success: true,
      threads: data.bookmarks || [],
      total: data.total || 0,
      pagination: data.pagination || null
    };
  } catch (err) {
    logger.error('Error getting user bookmarks:', err);
    throw err;
  }
};

export const searchBookmarks = async (query: string, page = 1, limit = 10): Promise<any> => {
  try {
    const token = getAuthToken();
    
    // Check if token exists - this is critical
    if (!token) {
      logger.error(`Cannot search bookmarks: No auth token available`);
      throw new Error('Authentication required. Please log in again.');
    }
    
    const url = `${API_BASE_URL}/bookmarks/search?q=${encodeURIComponent(query)}&page=${page}&limit=${limit}`;
    logger.debug(`Searching bookmarks from: ${url}`);

    const response = await fetch(url, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      credentials: "include"
    });

    if (!response.ok) {
      // Check specifically for auth errors
      if (response.status === 401) {
        logger.error('Authentication failed when searching bookmarks - token may be invalid');
        throw new Error('Your session has expired. Please log in again.');
      }
      
      logger.error(`Failed to search bookmarks: ${response.status}`);
      throw new Error(`Failed to search bookmarks: ${response.status}`);
    }

    const data = await response.json();
    logger.debug('Search bookmarks API returned data:', data);

    return {
      success: true,
      threads: data.bookmarks || [],
      total: data.total || 0,
      pagination: data.pagination || null
    };
  } catch (err) {
    logger.error('Error searching bookmarks:', err);
    throw err;
  }
};