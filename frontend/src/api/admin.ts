import appConfig from '../config/appConfig';
import { getAuthToken } from '../utils/auth';
import { createLoggerWithPrefix } from '../utils/logger';
import { checkAdminStatus } from './user';
import type { IApiResponse, IPagination } from '../interfaces/ICommon';

const API_BASE_URL = appConfig.api.baseUrl;
const logger = createLoggerWithPrefix('AdminAPI');

export interface RequestsResponse {
  success: boolean;
  requests?: any[];
  pagination: IPagination;
}

export interface CategoriesResponse {
  success: boolean;
  categories?: any[];
  pagination: IPagination;
}

export interface StatisticsResponse extends IApiResponse<{
  total_users?: number;
  active_users?: number;
  total_communities?: number;
  total_threads?: number;
  pending_reports?: number;
  new_users_today?: number;
  new_posts_today?: number;
  [key: string]: any;
}> {}

async function apiRequest<T>(url: string, method: string, body?: any): Promise<T> {

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

  logger.info(`Making ${method} request to ${url}`);

  const options: RequestInit = {
    method,
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`,
      'X-Admin-Request': 'true'  
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
      throw error; 
    }
    throw new Error('An unknown error occurred while communicating with the server');
  }
}

export async function getDashboardStatistics(): Promise<StatisticsResponse> {
  return apiRequest<StatisticsResponse>(
    `${API_BASE_URL}/admin/dashboard/statistics`,
    'GET'
  );
}

export async function banUser(userId: string, ban: boolean, reason?: string): Promise<IApiResponse<void>> {
  return apiRequest<IApiResponse<void>>(
    `${API_BASE_URL}/admin/users/${userId}/ban`, 
    'POST',
    { ban, reason }
  );
}

export async function sendNewsletter(subject: string, content: string): Promise<IApiResponse<void>> {
  return apiRequest<IApiResponse<void>>(
    `${API_BASE_URL}/admin/newsletter/send`,
    'POST',
    { subject, content }
  );
}

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

export async function processCommunityRequest(requestId: string, approve: boolean, reason?: string): Promise<IApiResponse<void>> {
  return apiRequest<IApiResponse<void>>(
    `${API_BASE_URL}/admin/community-requests/${requestId}/process`,
    'POST',
    { approve, reason }
  );
}

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

export async function processPremiumRequest(requestId: string, approve: boolean, reason?: string): Promise<IApiResponse<void>> {
  return apiRequest<IApiResponse<void>>(
    `${API_BASE_URL}/admin/premium-requests/${requestId}/process`,
    'POST',
    { approve, reason }
  );
}

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

export async function processReportRequest(requestId: string, approve: boolean, reason?: string): Promise<IApiResponse<void>> {
  return apiRequest<IApiResponse<void>>(
    `${API_BASE_URL}/admin/report-requests/${requestId}/process`,
    'POST',
    { approve, reason }
  );
}

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

export async function createThreadCategory(name: string, description?: string): Promise<IApiResponse<void>> {
  return apiRequest<IApiResponse<void>>(
    `${API_BASE_URL}/admin/thread-categories`,
    'POST',
    { name, description }
  );
}

export async function updateThreadCategory(categoryId: string, name: string, description?: string): Promise<IApiResponse<void>> {
  return apiRequest<IApiResponse<void>>(
    `${API_BASE_URL}/admin/thread-categories/${categoryId}`,
    'PUT',
    { name, description }
  );
}

export async function deleteThreadCategory(categoryId: string): Promise<IApiResponse<void>> {
  return apiRequest<IApiResponse<void>>(
    `${API_BASE_URL}/admin/thread-categories/${categoryId}`,
    'DELETE'
  );
}

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

export async function createCommunityCategory(name: string, description?: string): Promise<IApiResponse<void>> {
  return apiRequest<IApiResponse<void>>(
    `${API_BASE_URL}/admin/community-categories`,
    'POST',
    { name, description }
  );
}

export async function updateCommunityCategory(categoryId: string, name: string, description?: string): Promise<IApiResponse<void>> {
  return apiRequest<IApiResponse<void>>(
    `${API_BASE_URL}/admin/community-categories/${categoryId}`,
    'PUT',
    { name, description }
  );
}

export async function deleteCommunityCategory(categoryId: string): Promise<IApiResponse<void>> {
  return apiRequest<IApiResponse<void>>(
    `${API_BASE_URL}/admin/community-categories/${categoryId}`,
    'DELETE'
  );
}