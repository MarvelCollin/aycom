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

export interface AdminApiResponse {
  success: boolean;
  message?: string;
  [key: string]: any;
}

async function apiRequest<T>(url: string, method: string, body?: any): Promise<T> {
  // For development/demo purposes: no admin check required
  logger.info(`Making ${method} request to ${url}`);

  const headers: Record<string, string> = {
    'Content-Type': 'application/json'
  };
  
  // Add auth token if available, but don't require it
  const token = getAuthToken();
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
    logger.info(`Using auth token for request`);
  }

  const options: RequestInit = {
    method,
    headers,
    credentials: 'include'
  };

  if (body) {
    options.body = JSON.stringify(body);
    logger.info(`Request includes body data`);
  }

  try {
    logger.info(`Sending fetch request to ${url}`);
    const response = await fetch(url, options);
    logger.info(`Received response with status: ${response.status} from ${url}`);

    if (!response.ok) {
      // Try to get error details from response
      let errorMessage = `Request failed with status: ${response.status}`;
      try {
        const errorData = await response.json();
        if (errorData.error && errorData.error.message) {
          errorMessage = errorData.error.message;
        } else if (errorData.message) {
          errorMessage = errorData.message;
        }
      } catch (jsonError) {
        // If JSON parsing fails, use status text
        errorMessage = response.statusText || errorMessage;
      }
      
      logger.error(`Request failed: ${errorMessage}`);
      throw new Error(errorMessage);
    }

    const data = await response.json() as T;
    logger.info(`Successfully parsed response data`);
    return data;
  } catch (error) {
    logger.error('API request error:', error);
    throw error;
  }
}

export async function getDashboardStatistics(): Promise<StatisticsResponse> {
  return apiRequest<StatisticsResponse>(
    `${API_BASE_URL}/admin/dashboard/statistics`,
    'GET'
  );
}

export async function banUser(userId: string, ban: boolean, reason?: string): Promise<AdminApiResponse> {
  // Backend expects a boolean, not a "t" or "f" string
  return apiRequest<AdminApiResponse>(
    `${API_BASE_URL}/admin/users/${userId}/ban`, 
    'POST',
    { ban: ban, reason }
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

export async function processCommunityRequest(requestId: string, approve: boolean, reason?: string): Promise<AdminApiResponse> {
  // Backend expects a boolean, not a "t" or "f" string
  return apiRequest<AdminApiResponse>(
    `${API_BASE_URL}/admin/community-requests/${requestId}/process`,
    'POST',
    { approve: approve, reason }
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

export async function processPremiumRequest(requestId: string, approve: boolean, reason?: string): Promise<AdminApiResponse> {
  return apiRequest<AdminApiResponse>(
    `${API_BASE_URL}/admin/premium-requests/${requestId}/process`,
    'POST',
    { approve: approve, reason }
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

export async function processReportRequest(requestId: string, approve: boolean, reason?: string): Promise<AdminApiResponse> {
  return apiRequest<AdminApiResponse>(
    `${API_BASE_URL}/admin/report-requests/${requestId}/process`,
    'POST',
    { approve: approve, reason }
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