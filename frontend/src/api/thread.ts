import { getAuthToken, getUserId, getAuthData } from "../utils/auth";
import appConfig from "../config/appConfig";
import { uploadMultipleThreadMedia } from "../utils/supabase";
import { createLoggerWithPrefix } from "../utils/logger";
import { useAuth } from "../hooks/useAuth";

const API_BASE_URL = appConfig.api.baseUrl;
const AI_SERVICE_URL = appConfig.api.aiServiceUrl || "http:
const logger = createLoggerWithPrefix("ThreadAPI");

logger.info("Thread API using URL:", API_BASE_URL);

async function handleApiResponse(response: Response, errorMessage: string = "API request failed") {
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


async function ensureValidToken() {
  try {
    const { checkAndRefreshTokenIfNeeded } = useAuth();
    await checkAndRefreshTokenIfNeeded();
  } catch (error) {
    logger.warn("Error ensuring token freshness:", error);
    
    
  }
}

async function makeApiRequest(url: string, method: string, body?: any, errorMessage?: string, timeout: number = 15000) {
  try {
    
    const isPublicReadRequest = method === "GET" && url.includes("/threads");

    try {
      const { checkAndRefreshTokenIfNeeded } = useAuth();
      await checkAndRefreshTokenIfNeeded();
    } catch (error) {
      
      
      if (!isPublicReadRequest) {
        throw error;
      }
    }

    const token = getAuthToken();

    
    logger.debug(`Full request URL: ${url}`);

    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), timeout);

    try {
      
      const origin = typeof window !== "undefined" ? window.location.origin : "http:
      logger.debug(`Using origin: ${origin} for CORS headers`);

      const options: RequestInit = {
        method,
        headers: {
          "Content-Type": "application/json",
          "Authorization": token ? `Bearer ${token}` : "",
          "Accept": "application/json",
          "Origin": origin, 
          "Cache-Control": "no-cache" 
        },
        credentials: "include",
        signal: controller.signal,
        mode: "cors", 
        redirect: "follow" 
      };

      if (body) {
        options.body = JSON.stringify(body);
        logger.debug(`Request body (truncated): ${JSON.stringify(body).substring(0, 100)}...`);
      }

      logger.debug(`Making API request to ${url} with method ${method}`);

      
      if ((method === "POST" || method === "PUT") && !token && !isPublicReadRequest) {
        logger.warn("Attempting to make a write request without auth token");
      }

      
      if (url.includes("/threads") && method === "POST") {
        logger.debug("Critical path detected: Creating thread. Adding extra headers.");
        
        options.headers = {
          ...options.headers,
          "X-Requested-With": "XMLHttpRequest"
        };
      }

      
      const fetchWithRetry = async (fetchUrl: string, fetchOptions: RequestInit, retries = 1): Promise<Response> => {
        try {
          const response = await fetch(fetchUrl, fetchOptions);

          
          logger.debug(`API response status: ${response.status} ${response.statusText} for ${fetchUrl}`);

          
          if ([301, 302, 307, 308].includes(response.status)) {
            const location = response.headers.get("Location");
            if (location && retries > 0) {
              logger.info(`Following redirect to ${location}, retries left: ${retries-1}`);

              
              
              const redirectUrl = new URL(location, fetchUrl);

              return fetchWithRetry(redirectUrl.toString(), fetchOptions, retries - 1);
            }
          }

          return response;
        } catch (error: any) {
          logger.error(`Fetch attempt error for ${fetchUrl}: ${error.message}`);
          throw error;
        }
      };

      
      const response = await fetchWithRetry(url, options, 2);
      clearTimeout(timeoutId);

      
      if (response.status === 401 && isPublicReadRequest && token) {
        logger.warn("Got 401 on public endpoint, retrying without auth");
        const publicOptions = { ...options };
        publicOptions.headers = {
          "Content-Type": "application/json",
          "Accept": "application/json",
          "Origin": origin
        };

        const publicResponse = await fetch(url, publicOptions);
        return await handleApiResponse(publicResponse, errorMessage);
      }

      
      if (url.includes("/threads") && method === "POST" && response.status >= 200 && response.status < 300) {
        logger.debug("Thread creation succeeded with status:", response.status);
      }

      return await handleApiResponse(response, errorMessage);
    } catch (error: any) {
      clearTimeout(timeoutId);
      if (error.name === "AbortError") {
        throw new Error("Request timed out");
      }

      
      if (error.message.includes("CORS") || error.message.includes("cross-origin") || error.message === "Failed to fetch") {
        logger.error(`CORS error for ${url}:`, error.message);

        
        if (url.includes("/threads") && method === "POST") {
          logger.debug("Thread creation CORS issue - attempting fallback approach");

          try {
            
            const fallbackOptions: RequestInit = {
              method,
              headers: {
                "Content-Type": "application/json",
                "Authorization": token ? `Bearer ${token}` : "",
                "Accept": "*/*",
              },
              credentials: "include",
              mode: "cors",
              body: body ? JSON.stringify(body) : undefined
            };

            
            await new Promise(resolve => setTimeout(resolve, 500));

            const fallbackResponse = await fetch(url, fallbackOptions);
            logger.debug(`Fallback approach status: ${fallbackResponse.status}`);

            return await handleApiResponse(fallbackResponse, errorMessage);
          } catch (fallbackError) {
            logger.error("Fallback approach failed:", fallbackError);
            throw new Error("CORS error: The server may not allow requests from this origin. Please check your browser console for more details.");
          }
        } else {
          throw new Error(`CORS error: ${error.message}. The server may not allow requests from this origin.`);
        }
      }

      logger.error(`Fetch error for ${url}: ${error.message}`);
      throw error;
    }
  } catch (error: any) {
    logger.error(`API request failed for ${url}: ${error.message}`);
    throw error;
  }
}

export async function createThread(data: Record<string, any>) {
  try {
    logger.debug("Creating new thread with data:", {
      contentLength: data.content?.length || 0,
      hasMedia: !!data.media,
      hasPoll: !!data.poll,
      whoCanReply: data.who_can_reply
    });

    
    const url = `${API_BASE_URL}/threads`;
    logger.debug(`Using endpoint: ${url}`);

    
    const token = getAuthToken();
    if (!token) {
      logger.warn("No auth token available for createThread. User may need to log in");
      throw new Error("Authentication required. Please log in to post.");
    }

    
    logger.debug(`Using auth token for createThread: ${token.substring(0, 10)}...${token.substring(token.length - 10)}`);

    try {
      
      const origin = typeof window !== "undefined" ? window.location.origin : "http:
      const response = await fetch(url, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "Authorization": `Bearer ${token}`,
          "Accept": "application/json",
          "Origin": origin,
          "Cache-Control": "no-cache"
        },
        credentials: "include",
        mode: "cors",
        body: JSON.stringify(data)
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({ message: `HTTP error ${response.status}` }));
        logger.error("Thread creation failed with status:", response.status, errorData);
        throw new Error(errorData.message || `Failed to create thread: ${response.status}`);
      }

      const result = await response.json();
      logger.info("Thread created successfully");
      return result;
    } catch (apiError: any) {
      
      if (apiError.message?.includes("307") || apiError.message?.includes("Temporary Redirect")) {
        logger.error("Server returned redirect for thread creation. This could be a CORS issue");
        throw new Error("Server redirect occurred. This might be due to CORS configuration issues. Please check your network settings or contact support.");
      }

      throw apiError;
    }
  } catch (error: any) {
    logger.error(`Thread creation failed: ${error.message}`);
    throw error;
  }
}

export async function getThread(id: string) {
  try {
    logger.debug(`Fetching thread with ID: ${id}`);

    const response = await makeApiRequest(
      `${API_BASE_URL}/threads/${id}`,
      "GET",
      null,
      "Failed to fetch thread"
    );

    
    logger.debug(`Thread API response for ID ${id}:`, response);

    if (!response || (typeof response === "object" && Object.keys(response).length === 0)) {
      logger.warn(`Empty thread response for ID ${id}`);
      throw new Error(`Thread with ID ${id} not found or returned empty response`);
    }

    
    const standardizedResponse = {
      id: response.id,
      content: response.content,
      created_at: response.created_at,
      updated_at: response.updated_at,
      user_id: response.user_id,
      username: response.username,
      name: response.name,
      profile_picture_url: response.profile_picture_url,
      likes_count: response.likes_count || 0,
      replies_count: response.replies_count || 0,
      reposts_count: response.reposts_count || 0,
      bookmark_count: response.bookmark_count || 0,
      views_count: response.views_count || 0,
      is_liked: Boolean(response.is_liked),
      is_bookmarked: Boolean(response.is_bookmarked),
      is_reposted: Boolean(response.is_reposted),
      is_pinned: Boolean(response.is_pinned),
      is_verified: Boolean(response.is_verified),
      media: Array.isArray(response.media) ? response.media : []
    };

    logger.debug(`Thread data standardized for ID: ${id}`);
    return standardizedResponse;
  } catch (error) {
    logger.error(`Get thread ${id} failed:`, error);

    
    if (error instanceof Response) {
      try {
        const errorText = await error.text();
        logger.error(`API error response for thread ${id}:`, errorText);
      } catch (e) {
        logger.error("Could not read API error response:", e);
      }
    }

    throw error;
  }
}

export async function getThreadsByUser(userId: string, page: number = 1, limit: number = 10) {
  try {
    let actualUserId = userId;

    if (userId === "me") {
      const currentUserId = getUserId();
      logger.debug("Current user ID from auth:", currentUserId);

      if (!currentUserId) {
        throw new Error("User ID is required");
      }
      actualUserId = currentUserId;
    }

    const endpoint = `${API_BASE_URL}/threads/user/${actualUserId}?page=${page}&limit=${limit}`;
    logger.debug(`Making request to: ${endpoint}`);

    const data = await makeApiRequest(
      endpoint,
      "GET",
      null,
      "Failed to get user threads"
    );

    logger.info(`Received ${data.threads?.length || 0} threads for user ${actualUserId}`);
    logger.debug("Pinned threads:", data.threads?.filter(thread => thread.is_pinned === true).length);

    return data;
  } catch (error) {
    logger.error("Error getting user threads:", error);
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
  community_id?: string;
  community_name?: string;
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
        "GET",
        null,
        "Failed to fetch threads"
      );

      logger.info(`getAllThreads received ${data.threads?.length || 0} threads`);

      if (!data.threads || !Array.isArray(data.threads)) {
        logger.warn("API returned invalid threads data structure", data);
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
      logger.error("Fetch error in getAllThreads:", error);
      throw error;
    }
  } catch (error) {
    logger.error("Get all threads failed:", error);
    throw error;
  }
}

export async function updateThread(id: string, data: Record<string, any>) {
  try {
    return await makeApiRequest(
      `${API_BASE_URL}/threads/${id}`,
      "PUT",
      data,
      "Failed to update thread"
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
      "DELETE",
      null,
      "Failed to delete thread"
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
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "Authorization": token ? `Bearer ${token}` : ""
        },
        body: JSON.stringify({ media_urls: urls }),
        credentials: "include",
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
  formData.append("thread_id", threadId);

  files.forEach((file, index) => {
    formData.append("file", file);
  });

  const response = await fetch(`${API_BASE_URL}/threads/media`, {
    method: "POST",
    headers: {
      "Authorization": token ? `Bearer ${token}` : ""
    },
    body: formData,
    credentials: "include",
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


const likeDebounceMap = new Map();
const DEBOUNCE_DELAY = 500; 

export async function likeThread(threadId: string) {
  try {
    
    if (likeDebounceMap.has(threadId)) {
      logger.warn(`Like operation for thread ${threadId} already in progress, skipping`);
      return { success: false, message: "Operation already in progress" };
    }

    
    likeDebounceMap.set(threadId, true);

    
    setTimeout(() => {
      likeDebounceMap.delete(threadId);
    }, DEBOUNCE_DELAY);

    const token = getAuthToken();

    
    if (!token) {
      logger.error(`Cannot like thread ${threadId}: No auth token available`);
      throw new Error("Authentication required. Please log in again.");
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
      
      if (response.status === 401) {
        logger.error(`Authentication failed when liking thread ${threadId} - token may be invalid`);
        throw new Error("Your session has expired. Please log in again.");
      }

      const errorData = await response.json();
      logger.error(`Error liking thread ${threadId}:`, errorData);
      throw new Error(errorData.message || "Failed to like thread");
    }

    const data = await response.json();
    logger.debug(`Successfully liked thread ${threadId}`);
    return { ...data, success: true };
  } catch (error: any) {
    logger.error(`Like thread ${threadId} failed:`, error);
    throw error;
  } finally {
    
    setTimeout(() => {
      likeDebounceMap.delete(threadId);
    }, 50);
  }
}

export async function unlikeThread(threadId: string) {
  try {
    
    if (likeDebounceMap.has(threadId)) {
      logger.warn(`Unlike operation for thread ${threadId} already in progress, skipping`);
      return { success: false, message: "Operation already in progress" };
    }

    
    likeDebounceMap.set(threadId, true);

    
    setTimeout(() => {
      likeDebounceMap.delete(threadId);
    }, DEBOUNCE_DELAY);

    const token = getAuthToken();

    
    if (!token) {
      logger.error(`Cannot unlike thread ${threadId}: No auth token available`);
      throw new Error("Authentication required. Please log in again.");
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
      
      if (response.status === 401) {
        logger.error(`Authentication failed when unliking thread ${threadId} - token may be invalid`);
        throw new Error("Your session has expired. Please log in again.");
      }

      const errorData = await response.json();
      logger.error(`Error unliking thread ${threadId}:`, errorData);
      throw new Error(errorData.message || "Failed to unlike thread");
    }

    const data = await response.json();
    logger.debug(`Successfully unliked thread ${threadId}`);
    return { ...data, success: true };
  } catch (error) {
    logger.error(`Unlike thread ${threadId} failed:`, error);
    throw error;
  } finally {
    
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
      "POST",
      data,
      "Failed to post reply"
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
      "GET",
      null,
      "Failed to get thread replies"
    );

    let standardizedReplies = [];

    
    if (response && response.replies && Array.isArray(response.replies)) {
      standardizedReplies = response.replies.map(reply => ({
        id: reply.id,
        content: reply.content || "",
        created_at: reply.created_at,
        updated_at: reply.updated_at,
        thread_id: reply.thread_id || threadId,
        parent_id: reply.parent_id || reply.parent_reply_id || null,
        user_id: reply.user_id,
        username: reply.username,
        name: reply.name,
        profile_picture_url: reply.profile_picture_url,
        likes_count: reply.likes_count || 0,
        replies_count: reply.replies_count || 0,
        reposts_count: reply.reposts_count || 0,
        bookmark_count: reply.bookmark_count || 0,
        is_liked: Boolean(reply.is_liked),
        is_bookmarked: Boolean(reply.is_bookmarked),
        is_reposted: Boolean(reply.is_reposted),
        is_pinned: Boolean(reply.is_pinned),
        media: Array.isArray(reply.media) ? reply.media : []
      }));
    }

    return {
      replies: standardizedReplies,
      total: response.total || standardizedReplies.length
    };
  } catch (error) {
    logger.error(`Get replies for thread ${threadId} failed:`, error);
    throw error;
  }
}

export async function repostThread(threadId: string, content = "") {
  try {
    return await makeApiRequest(
      `${API_BASE_URL}/threads/${threadId}/reposts`,
      "POST",
      { content },
      "Failed to repost thread"
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
      "DELETE",
      null,
      "Failed to remove repost"
    );
  } catch (error) {
    logger.error(`Remove repost ${repostId} failed:`, error);
    throw error;
  }
}


const bookmarkDebounceMap = new Map();

export async function bookmarkThread(threadId: string) {
  try {
    
    if (bookmarkDebounceMap.has(threadId)) {
      logger.warn(`Bookmark operation for thread ${threadId} already in progress, skipping`);
      return { success: false, message: "Operation already in progress" };
    }

    
    bookmarkDebounceMap.set(threadId, true);

    
    setTimeout(() => {
      bookmarkDebounceMap.delete(threadId);
    }, DEBOUNCE_DELAY);

    const token = getAuthToken();

    
    if (!token) {
      logger.error(`Cannot bookmark thread ${threadId}: No auth token available`);
      throw new Error("Authentication required. Please log in again.");
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
      
      if (response.status === 401) {
        logger.error(`Authentication failed when bookmarking thread ${threadId} - token may be invalid`);
        throw new Error("Your session has expired. Please log in again.");
      }

      const errorData = await response.json();
      logger.error(`Error bookmarking thread ${threadId}:`, errorData);
      throw new Error(errorData.message || "Failed to bookmark thread");
    }

    const data = await response.json();
    logger.debug(`Successfully bookmarked thread ${threadId}`);
    return { ...data, success: true };
  } catch (error) {
    logger.error(`Bookmark thread ${threadId} failed:`, error);
    throw error;
  } finally {
    
    setTimeout(() => {
      bookmarkDebounceMap.delete(threadId);
    }, 50);
  }
}

export async function removeBookmark(threadId: string) {
  try {
    
    if (bookmarkDebounceMap.has(threadId)) {
      logger.warn(`Remove bookmark operation for thread ${threadId} already in progress, skipping`);
      return { success: false, message: "Operation already in progress" };
    }

    
    bookmarkDebounceMap.set(threadId, true);

    
    setTimeout(() => {
      bookmarkDebounceMap.delete(threadId);
    }, DEBOUNCE_DELAY);

    const token = getAuthToken();

    
    if (!token) {
      logger.error(`Cannot remove bookmark for thread ${threadId}: No auth token available`);
      throw new Error("Authentication required. Please log in again.");
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
      
      if (response.status === 401) {
        logger.error(`Authentication failed when removing bookmark for thread ${threadId} - token may be invalid`);
        throw new Error("Your session has expired. Please log in again.");
      }

      const errorData = await response.json();
      logger.error(`Error removing bookmark for thread ${threadId}:`, errorData);
      throw new Error(errorData.message || "Failed to remove bookmark");
    }

    const data = await response.json();
    logger.debug(`Successfully removed bookmark for thread ${threadId}`);
    return { ...data, success: true };
  } catch (error) {
    logger.error(`Remove bookmark for thread ${threadId} failed:`, error);
    throw error;
  } finally {
    
    setTimeout(() => {
      bookmarkDebounceMap.delete(threadId);
    }, 50);
  }
}

export async function getFollowingThreads(page = 1, limit = 20) {
  try {
    const url = `${API_BASE_URL}/threads/following?page=${page}&limit=${limit}`;

    const response = await makeApiRequest(
      url,
      "GET",
      null,
      "Failed to fetch following threads"
    );

    
    if (response && response.data) {
      
      return {
        success: true,
        threads: response.data.threads || [],
        total_count: response.data.pagination?.total_count || 0,
        pagination: response.data.pagination || { total_count: 0, current_page: page, per_page: limit }
      };
    }

    
    if (response && response.threads) {
      return {
        success: true,
        threads: response.threads,
        total_count: response.total || response.pagination?.total_count || response.threads.length,
        pagination: response.pagination || { total_count: response.threads.length, current_page: page, per_page: limit }
      };
    }

    
    return {
      success: true,
      threads: [],
      total_count: 0,
      pagination: { total_count: 0, current_page: page, per_page: limit }
    };
  } catch (error) {
    logger.error("Get following threads failed:", error);
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

    
    if (options?.filter) {
      params.append("filter", options.filter);
      console.log(`Using filter: ${options.filter}`);
    }
    if (options?.category) params.append("category", options.category);
    if (options?.sort_by) params.append("sort_by", options.sort_by);

    logger.debug(`Searching threads with query: ${query}, filter: ${options?.filter || "all"}`);

    const response = await fetch(`${API_BASE_URL}/threads/search?${params}`, {
      method: "GET",
      headers: {
        "Content-Type": "application/json",
        "Authorization": getAuthToken() ? `Bearer ${getAuthToken()}` : ""
      },
      credentials: "include"
    });

    if (!response.ok) {
      const errorText = await response.text();
      logger.error(`Thread search failed with status ${response.status}: ${errorText}`);
      return {
        threads: [],
        pagination: {
          total_count: 0,
          current_page: page,
          per_page: limit,
          total_pages: 0
        }
      };
    }

    const data = await response.json();
    logger.debug("Search threads response:", data);

    return {
      threads: data.threads || [],
      pagination: data.pagination || {
        total_count: data.threads?.length || 0,
        current_page: page,
        per_page: limit,
        total_pages: Math.ceil((data.threads?.length || 0) / limit)
      }
    };
  } catch (error) {
    logger.error("Search threads failed:", error);

    
    return {
      threads: [],
      pagination: {
        total_count: 0,
        current_page: page,
        per_page: limit,
        total_pages: 0
      }
    };
  }
}

export async function searchThreadsWithMedia(
  query: string,
  page: number = 1,
  limit: number = 10,
  options?: { filter?: string; category?: string }
) {
  try {
    
    const url = new URL(`${API_BASE_URL}/threads/search`);

    url.searchParams.append("q", query);
    url.searchParams.append("page", page.toString());
    url.searchParams.append("limit", limit.toString());
    url.searchParams.append("media_only", "true"); 

    
    if (options?.filter) {
      url.searchParams.append("filter", options.filter);
      console.log(`Using filter for media search: ${options.filter}`);
    }

    if (options?.category) {
      url.searchParams.append("category", options.category);
    }

    const token = getAuthToken();
    console.log(`Searching threads with media: ${url.toString()}`);

    const response = await fetch(url.toString(), {
      method: "GET",
      headers: {
        "Authorization": token ? `Bearer ${token}` : "",
        "Content-Type": "application/json"
      }
    });

    if (!response.ok) {
      
      throw new Error(`Failed to search threads with media: ${response.status}`);
    }

    const data = await response.json();

    
    if (data.threads) {
      data.threads = data.threads.filter(thread => thread.media && thread.media.length > 0);
    }

    return data;
  } catch (error: any) {
    console.error("Error searching threads with media:", error);
    throw error;
  }
}

export async function getThreadsByHashtag(
  hashtag: string,
  page: number = 1,
  limit: number = 10
) {
  try {
    const cleanHashtag = hashtag.startsWith("#") ? hashtag.substring(1) : hashtag;

    const url = new URL(`${API_BASE_URL}/threads/hashtag/${encodeURIComponent(cleanHashtag)}`);

    url.searchParams.append("page", page.toString());
    url.searchParams.append("limit", limit.toString());
    url.searchParams.append("sort_by", "likes");
    url.searchParams.append("sort_order", "desc");

    const token = getAuthToken();

    const response = await fetch(url.toString(), {
      method: "GET",
      headers: {
        "Authorization": token ? `Bearer ${token}` : "",
        "Content-Type": "application/json"
      }
    });

    if (!response.ok) {
      throw new Error(`Failed to get threads by hashtag: ${response.status}`);
    }

    return await response.json();
  } catch (error: any) {
    console.error("Error getting threads by hashtag:", error);
    throw error;
  }
}

export async function getReplyReplies(replyId: string, page = 1, limit = 20): Promise<{ replies: any[], total_count: number, cached?: boolean, error?: string }> {
  try {
    console.log(`Fetching replies for reply ID: ${replyId}`);

    const params = new URLSearchParams({
      page: page.toString(),
      limit: limit.toString()
    });

    const data = await makeApiRequest(
      `${API_BASE_URL}/replies/${replyId}/replies?${params.toString()}`,
      "GET",
      null,
      "Failed to fetch reply replies"
    );

    
    if (data && data.replies && Array.isArray(data.replies)) {
      data.replies = data.replies.map(reply => {
        
        if (reply.replies_count === undefined && reply.repliesCount === undefined) {
          reply.replies_count = 0;
        }

        
        return {
          id: reply.id,
          content: reply.content || "",
          created_at: reply.created_at,
          updated_at: reply.updated_at,
          thread_id: reply.thread_id,
          parent_id: reply.parent_id || null,
          user_id: reply.user_id,
          username: reply.username || "",
          name: reply.name || "",
          profile_picture_url: reply.profile_picture_url || "",
          likes_count: reply.likes_count || 0,
          replies_count: reply.replies_count || reply.repliesCount || 0,
          reposts_count: reply.reposts_count || 0,
          bookmark_count: reply.bookmark_count || 0,
          is_liked: Boolean(reply.is_liked),
          is_bookmarked: Boolean(reply.is_bookmarked),
          is_reposted: Boolean(reply.is_reposted),
          is_pinned: Boolean(reply.is_pinned),
          media: Array.isArray(reply.media) ? reply.media : [],
          
          parent_content: reply.parent_content || null,
          parent_user: reply.parent_user || null
        };
      });
    }

    console.log(`Successfully fetched ${data.replies?.length || 0} replies for reply ${replyId}`);
    return data;
  } catch (error: any) {
    console.error(`Error fetching replies for reply ${replyId}:`, error);
    return { replies: [], total_count: 0, error: error.message || "Network error fetching replies" };
  }
}

export async function suggestThreadCategory(content: string) {
  try {
    console.log("Requesting category suggestion for content:", content.substring(0, 50) + (content.length > 50 ? "..." : ""));

    if (!content || content.trim().length < 10) {
      return {
        category: "general",
        confidence: 0
      };
    }

    const response = await fetch(`${AI_SERVICE_URL}/predict/category`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json"
      },
      body: JSON.stringify({ content })
    });

    if (!response.ok) {
      console.warn("Category suggestion failed:", response.status, response.statusText);
      return {
        category: "general",
        confidence: 0
      };
    }

    const data = await response.json();
    console.log("Received category suggestion:", data);

    return {
      category: data.category || "general",
      confidence: data.confidence || 0
    };
  } catch (error) {
    console.error("Error suggesting thread category:", error);
    return {
      category: "general",
      confidence: 0
    };
  }
}

export async function likeReply(replyId: string) {
  try {
    const token = getAuthToken();

    
    if (!token) {
      logger.error(`Cannot like reply ${replyId}: No auth token available`);
      throw new Error("Authentication required. Please log in again.");
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
        throw new Error("Your session has expired. Please log in again.");
      }

      const errorData = await response.json();
      console.error(`Error liking reply ${replyId}:`, errorData);
      return { success: false, error: errorData.message || "Failed to like reply" };
    }

    const data = await response.json();
    console.log(`Successfully liked reply ${replyId}`);
    return { ...data, success: true };
  } catch (error) {
    console.error(`Error liking reply ${replyId}:`, error);
    return { success: false, error: "Network error liking reply" };
  }
}

export async function unlikeReply(replyId: string) {
  try {
    const token = getAuthToken();

    
    if (!token) {
      logger.error(`Cannot unlike reply ${replyId}: No auth token available`);
      throw new Error("Authentication required. Please log in again.");
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
      
      if (response.status === 401) {
        logger.error(`Authentication failed when unliking reply ${replyId} - token may be invalid`);
        throw new Error("Your session has expired. Please log in again.");
      }

      const errorData = await response.json();
      console.error(`Error unliking reply ${replyId}:`, errorData);
      return { success: false, error: errorData.message || "Failed to unlike reply" };
    }

    const data = await response.json();
    console.log(`Successfully unliked reply ${replyId}`);
    return { ...data, success: true };
  } catch (error) {
    console.error(`Error unliking reply ${replyId}:`, error);
    return { success: false, error: "Network error unliking reply" };
  }
}

function isUuid(str: string): boolean {
  const uuidPattern = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;
  return uuidPattern.test(str);
}

async function resolveUserIdIfNeeded(userId: string): Promise<string> {
  if (!userId) {
    logger.error("resolveUserIdIfNeeded: userId is empty or undefined");
    throw new Error("User ID is required");
  }

  logger.debug(`resolveUserIdIfNeeded: Processing userId "${userId}"`);

  if (isUuid(userId)) {
    logger.debug(`userId "${userId}" is already a valid UUID, using as is`);
    return userId;
  }

  if (userId === "me") {
    const token = getAuthToken();
    if (!token) {
      logger.error("No auth token available to resolve 'me' user ID");
      throw new Error("Authentication required to access your profile");
    }

    try {
      const tokenPayload = JSON.parse(atob(token.split(".")[1]));
      const currentUserId = tokenPayload.sub;
      
      if (!currentUserId) {
        logger.error("Token does not contain a subject (sub) claim");
        throw new Error("Invalid authentication token");
      }
      
      logger.debug(`Resolved 'me' to user ID: ${currentUserId}`);
      return currentUserId;
    } catch (error) {
      logger.error("Failed to parse JWT token:", error);
      throw new Error("Invalid authentication token");
    }
  }

  
  try {
    logger.debug(`Attempting to resolve username: ${userId}`);
    const response = await fetch(`${API_BASE_URL}/users/username/${encodeURIComponent(userId)}`, {
      method: "GET",
      headers: {
        "Content-Type": "application/json",
        "Authorization": `Bearer ${getAuthToken()}`
      }
    });

    if (!response.ok) {
      logger.error(`Failed to resolve username '${userId}': ${response.status} ${response.statusText}`);
      throw new Error(`Failed to resolve username: ${response.status}`);
    }

    const data = await response.json();
    if (!data.user?.id) {
      logger.error(`User ID not found in response for username '${userId}'`, data);
      throw new Error("User not found");
    }

    logger.debug(`Successfully resolved username '${userId}' to ID: ${data.user.id}`);
    return data.user.id;
  } catch (error) {
    logger.error(`Error resolving username '${userId}':`, error);
    throw error;
  }
}

export const getUserThreads = async (userId: string, page = 1, limit = 10): Promise<any> => {
  const maxRetries = 2;
  let retryCount = 0;
  let lastError: Error | null = null;

  while (retryCount <= maxRetries) {
    try {
      if (!userId) {
        logger.error("getUserThreads: userId is undefined or empty");
        return {
          threads: [],
          total: 0,
          error: "User ID is required",
          success: false
        };
      }

      let resolvedUserId;
      try {
        resolvedUserId = await resolveUserIdIfNeeded(userId);
        logger.debug(`getUserThreads: Resolved userId from ${userId} to ${resolvedUserId}`);
      } catch (resolveError: any) {
        logger.error(`getUserThreads: Failed to resolve user ID: ${resolveError.message}`, resolveError);
        return {
          threads: [],
          total: 0,
          error: `Failed to resolve user ID: ${resolveError.message}`,
          success: false
        };
      }

      
      if (!resolvedUserId) {
        logger.error("getUserThreads: Resolved userId is empty");
        return {
          threads: [],
          total: 0,
          error: "Invalid user ID after resolution",
          success: false
        };
      }

      logger.debug(`Fetching threads for user ${resolvedUserId} (original: ${userId}), page: ${page}, limit: ${limit}, attempt: ${retryCount + 1}/${maxRetries + 1}`);

      
      const controller = new AbortController();
      const timeoutId = setTimeout(() => controller.abort(), 15000); 

      try {
        const url = `${API_BASE_URL}/threads/user/${resolvedUserId}?page=${page}&limit=${limit}`;
        logger.debug(`Fetching from URL: ${url}`);

        const response = await fetch(url, {
          method: "GET",
          headers: {
            "Authorization": `Bearer ${getAuthToken()}`,
            "Content-Type": "application/json"
          },
          signal: controller.signal
        });

        
        clearTimeout(timeoutId);

        if (!response.ok) {
          let errorMessage = `Failed to get user threads: ${response.status}`;
          let errorDetails = {};
          
          try {
            const errorData = await response.json();
            logger.error("Error getting user threads:", errorData);
            errorDetails = errorData;
            
            if (errorData.message) {
              errorMessage += ` - ${errorData.message}`;
            }
            if (errorData.error) {
              errorMessage += ` - ${errorData.error}`;
            }
          } catch (parseError) {
            logger.error("Could not parse error response:", parseError);
          }

          
          if (response.status === 400) {
            logger.error(`Bad Request (400) when fetching threads for user ID: ${resolvedUserId}`);
            
            
            if (userId === "me") {
              logger.debug("Attempting to refresh auth token...");
              
            }
          }

          
          if (response.status >= 500) {
            throw new Error(errorMessage);
          } else {
            
            return {
              threads: [],
              total: 0,
              error: errorMessage,
              errorDetails,
              success: false
            };
          }
        }

        const result = await response.json();
        logger.debug(`Successfully fetched threads for user ${resolvedUserId}:`, result);

        
        if (!result.threads && result.success) {
          logger.warn("Response is missing threads array:", result);

          
          if (Array.isArray(result.data)) {
            result.threads = result.data;
            logger.debug("Using data array as threads");
          }
        }

        
        return {
          ...result,
          success: true
        };
      } catch (fetchError) {
        
        clearTimeout(timeoutId);
        throw fetchError;
      }
    } catch (error: any) {
      lastError = error;

      
      const isNetworkError = error.name === "AbortError" ||
                            error.message?.includes("network") ||
                            error.message?.includes("timeout");

      if (isNetworkError || error.message?.includes("500")) {
        logger.warn(`Attempt ${retryCount + 1}/${maxRetries + 1} failed: ${error.message}. Retrying...`);
        retryCount++;

        if (retryCount <= maxRetries) {
          
          await new Promise(resolve => setTimeout(resolve, retryCount * 2000));
          continue;
        }
      } else {
        
        logger.error("Non-retriable error in getUserThreads:", error);
        break;
      }
    }
  }

  
  logger.error("Error in getUserThreads after all retries:", lastError);

  
  return {
    threads: [],
    total: 0,
    error: lastError?.message || "Unknown error fetching threads",
    success: false
  };
};

export const getUserReplies = async (userId: string, page = 1, limit = 10): Promise<any> => {
  const maxRetries = 2;
  let retryCount = 0;
  let lastError: Error | null = null;

  while (retryCount <= maxRetries) {
    try {
      const resolvedUserId = await resolveUserIdIfNeeded(userId);
      logger.debug(`Fetching replies for user ${resolvedUserId} (original: ${userId}), page: ${page}, limit: ${limit}, attempt: ${retryCount + 1}/${maxRetries + 1}`);

      
      const controller = new AbortController();
      const timeoutId = setTimeout(() => controller.abort(), 15000); 

      try {
        const response = await fetch(`${API_BASE_URL}/threads/user/${resolvedUserId}/replies?page=${page}&limit=${limit}`, {
          method: "GET",
          headers: {
            "Authorization": `Bearer ${getAuthToken()}`,
            "Content-Type": "application/json"
          },
          signal: controller.signal
        });

        
        clearTimeout(timeoutId);

        if (!response.ok) {
          let errorMessage = `Failed to get user replies: ${response.status}`;
          try {
            const errorData = await response.json();
            logger.error("Error getting user replies:", errorData);
            if (errorData.message) {
              errorMessage += ` - ${errorData.message}`;
            }
          } catch (parseError) {
            logger.error("Could not parse error response:", parseError);
          }

          
          if (response.status >= 500) {
            throw new Error(errorMessage);
          } else {
            
            return {
              replies: [],
              total: 0,
              error: errorMessage,
              success: false
            };
          }
        }

        const result = await response.json();

        
        return {
          ...result,
          success: true
        };
      } catch (fetchError) {
        
        clearTimeout(timeoutId);
        throw fetchError;
      }
    } catch (error: any) {
      lastError = error;

      
      const isNetworkError = error.name === "AbortError" ||
                             error.message?.includes("network") ||
                             error.message?.includes("timeout");

      if (isNetworkError || error.message?.includes("500")) {
        logger.warn(`Attempt ${retryCount + 1}/${maxRetries + 1} failed: ${error.message}. Retrying...`);
        retryCount++;

        if (retryCount <= maxRetries) {
          
          await new Promise(resolve => setTimeout(resolve, retryCount * 2000));
          continue;
        }
      } else {
        
        logger.error("Non-retriable error in getUserReplies:", error);
        break;
      }
    }
  }

  
  logger.error("Error in getUserReplies after all retries:", lastError);

  
  return {
    replies: [],
    total: 0,
    error: lastError?.message || "Unknown error fetching replies",
    success: false
  };
};

export const getUserLikedThreads = async (userId: string, page: number = 1, limit: number = 10) => {
  const maxRetries = 2;
  let retryCount = 0;
  let lastError: Error | null = null;

  while (retryCount <= maxRetries) {
    try {
      const token = getAuthToken();
      let actualUserId = userId;

      if (userId === "me") {
        actualUserId = await resolveUserIdIfNeeded(userId);
      }

      const endpoint = `${API_BASE_URL}/threads/user/${actualUserId}/likes?page=${page}&limit=${limit}`;
      logger.debug(`Making request to: ${endpoint}, attempt: ${retryCount + 1}/${maxRetries + 1}`);

      
      const controller = new AbortController();
      const timeoutId = setTimeout(() => controller.abort(), 15000); 

      try {
        const response = await fetch(endpoint, {
          method: "GET",
          headers: {
            "Content-Type": "application/json",
            "Authorization": token ? `Bearer ${token}` : ""
          },
          credentials: "include",
          signal: controller.signal
        });

        
        clearTimeout(timeoutId);

        if (!response.ok) {
          let errorMessage = `Failed to get user liked threads: ${response.status}`;
          try {
            const errorData = await response.json();
            errorMessage = errorData.message || errorMessage;
            logger.error("API error response:", errorData);
          } catch (parseError) {
            logger.error("Could not parse error response:", parseError);
          }

          
          if (response.status >= 500) {
            throw new Error(errorMessage);
          } else {
            
            return {
              threads: [],
              total: 0,
              error: errorMessage,
              success: false
            };
          }
        }

        const responseData = await response.json();

        
        return {
          ...responseData,
          success: true
        };
      } catch (fetchError) {
        
        clearTimeout(timeoutId);
        throw fetchError;
      }
    } catch (error: any) {
      lastError = error;

      
      const isNetworkError = error.name === "AbortError" ||
                             error.message?.includes("network") ||
                             error.message?.includes("timeout");

      if (isNetworkError || error.message?.includes("500")) {
        logger.warn(`Attempt ${retryCount + 1}/${maxRetries + 1} failed: ${error.message}. Retrying...`);
        retryCount++;

        if (retryCount <= maxRetries) {
          
          await new Promise(resolve => setTimeout(resolve, retryCount * 2000));
          continue;
        }
      } else {
        
        logger.error("Non-retriable error in getUserLikedThreads:", error);
        break;
      }
    }
  }

  
  logger.error("Error in getUserLikedThreads after all retries:", lastError);

  
  return {
    threads: [],
    total: 0,
    error: lastError?.message || "Unknown error fetching liked threads",
    success: false
  };
};

export const getUserMedia = async (userId: string, page = 1, limit = 10): Promise<any> => {
  const maxRetries = 2;
  let retryCount = 0;
  let lastError: Error | null = null;

  while (retryCount <= maxRetries) {
    try {
      const resolvedUserId = await resolveUserIdIfNeeded(userId);
      logger.debug(`Fetching media for user ${resolvedUserId} (original: ${userId}), page: ${page}, limit: ${limit}, attempt: ${retryCount + 1}/${maxRetries + 1}`);

      
      const controller = new AbortController();
      const timeoutId = setTimeout(() => controller.abort(), 15000); 

      try {
        const response = await fetch(`${API_BASE_URL}/threads/user/${resolvedUserId}/media?page=${page}&limit=${limit}`, {
          method: "GET",
          headers: {
            "Authorization": `Bearer ${getAuthToken()}`,
            "Content-Type": "application/json"
          },
          signal: controller.signal
        });

        
        clearTimeout(timeoutId);

        if (!response.ok) {
          let errorMessage = `Failed to get user media: ${response.status}`;
          try {
            const errorData = await response.json();
            logger.error("Error getting user media:", errorData);
            if (errorData.message) {
              errorMessage += ` - ${errorData.message}`;
            }
          } catch (parseError) {
            logger.error("Could not parse error response:", parseError);
          }

          
          if (response.status >= 500) {
            throw new Error(errorMessage);
          } else {
            
            return {
              media: [],
              total: 0,
              error: errorMessage,
              success: false
            };
          }
        }

        const result = await response.json();

        
        return {
          ...result,
          success: true
        };
      } catch (fetchError) {
        
        clearTimeout(timeoutId);
        throw fetchError;
      }
    } catch (error: any) {
      lastError = error;

      
      const isNetworkError = error.name === "AbortError" ||
                             error.message?.includes("network") ||
                             error.message?.includes("timeout");

      if (isNetworkError || error.message?.includes("500")) {
        logger.warn(`Attempt ${retryCount + 1}/${maxRetries + 1} failed: ${error.message}. Retrying...`);
        retryCount++;

        if (retryCount <= maxRetries) {
          
          await new Promise(resolve => setTimeout(resolve, retryCount * 2000));
          continue;
        }
      } else {
        
        logger.error("Non-retriable error in getUserMedia:", error);
        break;
      }
    }
  }

  
  logger.error("Error in getUserMedia after all retries:", lastError);

  
  return {
    media: [],
    total: 0,
    error: lastError?.message || "Unknown error fetching media",
    success: false
  };
};

export const getUserBookmarks = async (userId: string, page = 1, limit = 10): Promise<any> => {
  try {
    const token = getAuthToken();

    
    if (!token) {
      logger.error("Cannot get user bookmarks: No auth token available");
      throw new Error("Authentication required. Please log in again.");
    }

    const actualUserId = userId === "me" ? getUserId() : userId;

    if (!actualUserId) {
      logger.error("No user ID available, cannot fetch bookmarks");
      throw new Error("User ID is required");
    }

    
    const url = `${API_BASE_URL}/bookmarks?page=${page}&limit=${limit}`;
    logger.debug(`Fetching bookmarks from: ${url} for user ${actualUserId}`);
    logger.debug(`Auth token available: ${!!token}, length: ${token?.length}`);

    const response = await fetch(url, {
      method: "GET",
      headers: {
        "Content-Type": "application/json",
        "Authorization": `Bearer ${token}`
      },
      credentials: "include"
    });

    if (!response.ok) {
      
      if (response.status === 401) {
        logger.error("Authentication failed when getting bookmarks - token may be invalid");
        throw new Error("Your session has expired. Please log in again.");
      }

      logger.error(`Failed to get bookmarks: ${response.status}`);
      throw new Error(`Failed to get bookmarks: ${response.status}`);
    }

    const data = await response.json();
    logger.debug("Bookmarks API returned data:", data);

    
    
    
    const bookmarks = data.data?.bookmarks || data.bookmarks || [];
    logger.debug(`Bookmarks count: ${bookmarks.length}`);

    
    return {
      success: true,
      bookmarks: bookmarks,
      total: data.data?.total || data.total || bookmarks.length,
      pagination: data.data?.pagination || data.pagination || null
    };
  } catch (err) {
    logger.error("Error getting user bookmarks:", err);
    throw err;
  }
};

export const searchBookmarks = async (query: string, page = 1, limit = 10): Promise<any> => {
  try {
    const token = getAuthToken();

    
    if (!token) {
      logger.error("Cannot search bookmarks: No auth token available");
      throw new Error("Authentication required. Please log in again.");
    }

    const url = `${API_BASE_URL}/bookmarks/search?q=${encodeURIComponent(query)}&page=${page}&limit=${limit}`;
    logger.debug(`Searching bookmarks from: ${url}`);

    const response = await fetch(url, {
      method: "GET",
      headers: {
        "Content-Type": "application/json",
        "Authorization": `Bearer ${token}`
      },
      credentials: "include"
    });

    if (!response.ok) {
      
      if (response.status === 401) {
        logger.error("Authentication failed when searching bookmarks - token may be invalid");
        throw new Error("Your session has expired. Please log in again.");
      }

      logger.error(`Failed to search bookmarks: ${response.status}`);
      throw new Error(`Failed to search bookmarks: ${response.status}`);
    }

    const data = await response.json();
    logger.debug("Search bookmarks API returned data:", data);

    
    return {
      success: true,
      bookmarks: data.bookmarks || [],
      total: data.total || 0,
      pagination: data.pagination || null
    };
  } catch (err) {
    logger.error("Error searching bookmarks:", err);
    throw err;
  }
};