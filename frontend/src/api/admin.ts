import appConfig from '../config/appConfig';
import { getAuthToken } from '../utils/auth';
import { createLoggerWithPrefix } from '../utils/logger';
import { checkAdminStatus } from './user';
import type { IApiResponse, IPagination } from '../interfaces/ICommon';
import { 
  standardizeCommunityRequest, 
  standardizePremiumRequest, 
  standardizeReportRequest, 
  standardizePagination 
} from '../utils/standardizeApiData';

const API_BASE_URL = appConfig.api.baseUrl;
const logger = createLoggerWithPrefix('AdminAPI');

// Standard API response for admin endpoints
export type AdminApiResponse = IApiResponse<{
  message: string;
}>;

// Response for requests endpoints (community, premium, report)
export interface RequestsResponse {
  success: boolean;
  data: any[];
  pagination: IPagination;
}

// Response for category endpoints
export interface CategoriesResponse {
  success: boolean;
  data: any[];
  pagination: IPagination;
}

// Statistics response
export interface StatisticsResponse extends IApiResponse<{
  total_users?: number;
  active_users?: number;
  total_communities?: number;
  total_threads?: number;
  pending_reports?: number;
  new_users_today?: number;
  new_posts_today?: number;
}> {}

/**
 * Make a standardized API request
 */
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
    
    if (!response.ok) {
      logger.error(`API error: ${response.status} ${response.statusText}`);
      const errorData = await response.json();
      throw new Error(errorData.error?.message || `Request failed with status ${response.status}`);
    }
    
    const data = await response.json();
    logger.info(`Request successful, data received`);
    return data;
  } catch (error: unknown) {
    const errorMessage = error instanceof Error ? error.message : 'Unknown error occurred';
    logger.error(`Request failed: ${errorMessage}`);
    throw error;
  }
}

/**
 * Get all community requests with pagination
 */
export async function getCommunityRequests(page: number = 1, limit: number = 10, status?: string): Promise<RequestsResponse> {
  const response = await apiRequest<RequestsResponse>(
    `${API_BASE_URL}/admin/community-requests?page=${page}&limit=${limit}${status ? `&status=${status}` : ''}`,
    'GET'
  );
  
  // Standardize data
  if (response.data && Array.isArray(response.data)) {
    response.data = response.data.map(standardizeCommunityRequest);
  }
  
  if (response.pagination) {
    response.pagination = standardizePagination(response.pagination);
  }
  
  return response;
}

/**
 * Get premium requests with pagination
 */
export async function getPremiumRequests(page: number = 1, limit: number = 10, status?: string): Promise<RequestsResponse> {
  const response = await apiRequest<RequestsResponse>(
    `${API_BASE_URL}/admin/premium-requests?page=${page}&limit=${limit}${status ? `&status=${status}` : ''}`,
    'GET'
  );
  
  // Standardize data
  if (response.data && Array.isArray(response.data)) {
    response.data = response.data.map(standardizePremiumRequest);
  }
  
  if (response.pagination) {
    response.pagination = standardizePagination(response.pagination);
  }
  
  return response;
}

/**
 * Get report requests with pagination
 */
export async function getReportRequests(page: number = 1, limit: number = 10, status?: string): Promise<RequestsResponse> {
  const response = await apiRequest<RequestsResponse>(
    `${API_BASE_URL}/admin/report-requests?page=${page}&limit=${limit}${status ? `&status=${status}` : ''}`,
    'GET'
  );
  
  // Standardize data
  if (response.data && Array.isArray(response.data)) {
    response.data = response.data.map(standardizeReportRequest);
  }
  
  if (response.pagination) {
    response.pagination = standardizePagination(response.pagination);
  }
  
  return response;
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

export async function processCommunityRequest(requestId: string, approve: boolean, reason?: string): Promise<AdminApiResponse> {
  // Backend expects a boolean, not a "t" or "f" string
  return apiRequest<AdminApiResponse>(
    `${API_BASE_URL}/admin/community-requests/${requestId}/process`,
    'POST',
    { approve: approve, reason }
  );
}

export async function processReportRequest(requestId: string, approve: boolean, reason?: string): Promise<AdminApiResponse> {
  return apiRequest<AdminApiResponse>(
    `${API_BASE_URL}/admin/report-requests/${requestId}/process`,
    'POST',
    { approve: approve, reason }
  );
}

export async function processPremiumRequest(requestId: string, approve: boolean, reason?: string): Promise<AdminApiResponse> {
  return apiRequest<AdminApiResponse>(
    `${API_BASE_URL}/admin/premium-requests/${requestId}/process`,
    'POST',
    { approve: approve, reason }
  );
}

export async function getThreadCategories(page: number = 1, limit: number = 10): Promise<CategoriesResponse> {
  const response = await apiRequest<CategoriesResponse>(
    `${API_BASE_URL}/admin/thread-categories?page=${page}&limit=${limit}`,
    'GET'
  );
  
  // Standardize data
  if (response.data && Array.isArray(response.data)) {
    // Apply any specific standardization if needed
  }
  
  if (response.pagination) {
    response.pagination = standardizePagination(response.pagination);
  }
  
  return response;
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