import appConfig from '../config/appConfig';
import { getAuthToken } from '../utils/auth';
import { createLoggerWithPrefix } from '../utils/logger';
import { checkAdminStatus } from './user';

const API_BASE_URL = appConfig.api.baseUrl;
const logger = createLoggerWithPrefix('AdminAPI');

// Define response interfaces
export interface BaseResponse {
  success: boolean;
  message?: string;
}

export interface RequestsResponse extends BaseResponse {
  requests?: any[];
  total_count?: number;
  page?: number;
  total_pages?: number;
}

export interface CategoriesResponse extends BaseResponse {
  categories?: any[];
  total_count?: number;
  page?: number;
  total_pages?: number;
}

export interface StatisticsResponse extends BaseResponse {
  totalUsers?: number;
  activeUsers?: number;
  totalCommunities?: number;
  totalThreads?: number;
  pendingReports?: number;
  newUsersToday?: number;
  newPostsToday?: number;
  [key: string]: any;
}

/**
 * Standard API request handler with proper error management
 * @param url API endpoint URL
 * @param method HTTP method
 * @param body Request body (optional)
 * @returns Promise with the response data
 */
async function apiRequest<T>(url: string, method: string, body?: any): Promise<T> {
  // Check if user is admin before making request
  try {
    const isAdmin = await checkAdminStatus();
    if (!isAdmin) {
      logger.error('User does not have admin permissions');
      throw new Error('You do not have permission to access this resource');
    }
  } catch (adminCheckError) {
    logger.error('Error checking admin status:', adminCheckError);
  }

  const token = getAuthToken();
  
  if (!token) {
    logger.error('Missing authentication token for admin API request');
    throw new Error('Authentication required');
  }
  
  // Log request details (without sensitive data)
  logger.info(`Making ${method} request to ${url}`);
  
  const options: RequestInit = {
    method,
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`,
      'X-Admin-Request': 'true'  // Add custom header to indicate admin request
    },
    credentials: 'include'
  };
  
  if (body) {
    options.body = JSON.stringify(body);
    logger.info(`Request includes body data: ${Object.keys(body).join(', ')}`);
  }
  
  try {
    const response = await fetch(url, options);
    logger.info(`Received response with status: ${response.status} from ${url}`);
    
    if (response.status === 403) {
      logger.error(`Access denied (403) for ${method} ${url} - User lacks permission`);
      throw new Error('You do not have permission to access this resource');
    }
    
    if (response.status === 401) {
      logger.error(`Authentication failed (401) for ${method} ${url} - Token may be invalid`);
      throw new Error('Authentication failed - please log in again');
    }
    
    if (!response.ok) {
      let errorMessage = `Request failed with status: ${response.status}`;
      try {
        const errorData = await response.json();
        errorMessage = errorData.message || errorMessage;
        logger.error(`API error response: ${JSON.stringify(errorData)}`);
      } catch (e) {
        // Parsing error response failed, use default message
        logger.error(`Failed to parse error response: ${e}`);
      }
      
      logger.error(`API error: ${errorMessage}`);
      throw new Error(errorMessage);
    }
    
    try {
      const data = await response.json() as T;
      logger.info(`Successfully parsed response data from ${url}`);
      return data;
    } catch (e) {
      logger.error(`Failed to parse success response as JSON from ${url}: ${e}`);
      throw new Error('Invalid response format from server');
    }
  } catch (error) {
    if (error instanceof Error) {
      throw error; // Re-throw if it's already an Error object
    }
    throw new Error('An unknown error occurred while communicating with the server');
  }
}

// Dashboard Statistics API
export async function getDashboardStatistics(): Promise<StatisticsResponse> {
  return apiRequest<StatisticsResponse>(
    `${API_BASE_URL}/admin/dashboard/statistics`,
    'GET'
  );
}

// User Management APIs
export async function banUser(userId: string, ban: boolean, reason?: string): Promise<BaseResponse> {
  return apiRequest<BaseResponse>(
    `${API_BASE_URL}/admin/users/${userId}/ban`, 
    'POST',
    { ban, reason }
  );
}

// Newsletter APIs
export async function sendNewsletter(subject: string, content: string): Promise<BaseResponse> {
  return apiRequest<BaseResponse>(
    `${API_BASE_URL}/admin/newsletter/send`,
    'POST',
    { subject, content }
  );
}

// Community Request APIs
export async function getCommunityRequests(page: number = 1, limit: number = 10, status?: string): Promise<RequestsResponse> {
  const params = new URLSearchParams({
    page: page.toString(),
    limit: limit.toString()
  });
  
  if (status) {
    params.append('status', status);
  }
  
  return apiRequest<RequestsResponse>(
    `${API_BASE_URL}/admin/community-requests?${params}`,
    'GET'
  );
}

export async function processCommunityRequest(requestId: string, approve: boolean, reason?: string): Promise<BaseResponse> {
  return apiRequest<BaseResponse>(
    `${API_BASE_URL}/admin/community-requests/${requestId}/process`,
    'POST',
    { approve, reason }
  );
}

// Premium Request APIs
export async function getPremiumRequests(page: number = 1, limit: number = 10, status?: string): Promise<RequestsResponse> {
  const params = new URLSearchParams({
    page: page.toString(),
    limit: limit.toString()
  });
  
  if (status) {
    params.append('status', status);
  }
  
  return apiRequest<RequestsResponse>(
    `${API_BASE_URL}/admin/premium-requests?${params}`,
    'GET'
  );
}

export async function processPremiumRequest(requestId: string, approve: boolean, reason?: string): Promise<BaseResponse> {
  return apiRequest<BaseResponse>(
    `${API_BASE_URL}/admin/premium-requests/${requestId}/process`,
    'POST',
    { approve, reason }
  );
}

// Report Request APIs
export async function getReportRequests(page: number = 1, limit: number = 10, status?: string): Promise<RequestsResponse> {
  const params = new URLSearchParams({
    page: page.toString(),
    limit: limit.toString()
  });
  
  if (status) {
    params.append('status', status);
  }
  
  return apiRequest<RequestsResponse>(
    `${API_BASE_URL}/admin/report-requests?${params}`,
    'GET'
  );
}

export async function processReportRequest(requestId: string, approve: boolean, reason?: string): Promise<BaseResponse> {
  return apiRequest<BaseResponse>(
    `${API_BASE_URL}/admin/report-requests/${requestId}/process`,
    'POST',
    { approve, reason }
  );
}

// Thread Category APIs
export async function getThreadCategories(page: number = 1, limit: number = 10): Promise<CategoriesResponse> {
  const params = new URLSearchParams({
    page: page.toString(),
    limit: limit.toString()
  });
  
  return apiRequest<CategoriesResponse>(
    `${API_BASE_URL}/admin/thread-categories?${params}`,
    'GET'
  );
}

export async function createThreadCategory(name: string, description?: string): Promise<BaseResponse> {
  return apiRequest<BaseResponse>(
    `${API_BASE_URL}/admin/thread-categories`,
    'POST',
    { name, description }
  );
}

export async function updateThreadCategory(categoryId: string, name: string, description?: string): Promise<BaseResponse> {
  return apiRequest<BaseResponse>(
    `${API_BASE_URL}/admin/thread-categories/${categoryId}`,
    'PUT',
    { name, description }
  );
}

export async function deleteThreadCategory(categoryId: string): Promise<BaseResponse> {
  return apiRequest<BaseResponse>(
    `${API_BASE_URL}/admin/thread-categories/${categoryId}`,
    'DELETE'
  );
}

// Community Category APIs
export async function getCommunityCategories(page: number = 1, limit: number = 10): Promise<CategoriesResponse> {
  const params = new URLSearchParams({
    page: page.toString(),
    limit: limit.toString()
  });
  
  return apiRequest<CategoriesResponse>(
    `${API_BASE_URL}/admin/community-categories?${params}`,
    'GET'
  );
}

export async function createCommunityCategory(name: string, description?: string): Promise<BaseResponse> {
  return apiRequest<BaseResponse>(
    `${API_BASE_URL}/admin/community-categories`,
    'POST',
    { name, description }
  );
}

export async function updateCommunityCategory(categoryId: string, name: string, description?: string): Promise<BaseResponse> {
  return apiRequest<BaseResponse>(
    `${API_BASE_URL}/admin/community-categories/${categoryId}`,
    'PUT',
    { name, description }
  );
}

export async function deleteCommunityCategory(categoryId: string): Promise<BaseResponse> {
  return apiRequest<BaseResponse>(
    `${API_BASE_URL}/admin/community-categories/${categoryId}`,
    'DELETE'
  );
}
