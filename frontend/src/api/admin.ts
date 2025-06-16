import appConfig from '../config/appConfig';
import { getAuthToken } from '../utils/auth';
import { createLoggerWithPrefix } from '../utils/logger';
import { checkAdminStatus } from './user';
import type { IApiResponse, IPagination } from '../interfaces/ICommon';
import { 
  standardizeCommunityRequest, 
  standardizePremiumRequest, 
  standardizeReportRequest, 
  standardizePagination,
  standardizeUser
} from '../utils/standardizeApiData';

const API_BASE_URL = appConfig.api.baseUrl;
const logger = createLoggerWithPrefix('AdminAPI');

export type AdminApiResponse = IApiResponse<{
  message: string;
}>;

export interface RequestsResponse {
  success: boolean;
  data: any[];  
  requests?: any[];  
  pagination: IPagination;
}

export interface CategoriesResponse {
  success: boolean;
  data: any[];
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
}> {}

async function apiRequest<T>(url: string, method: string, body?: any): Promise<T> {

  logger.info(`Making ${method} request to ${url}`);

  const headers: Record<string, string> = {
    'Content-Type': 'application/json'
  };

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

export async function getCommunityRequests(page: number = 1, limit: number = 10, status?: string): Promise<RequestsResponse> {
  const response = await apiRequest<any>(
    `${API_BASE_URL}/admin/community-requests?page=${page}&limit=${limit}${status ? `&status=${status}` : ''}`,
    'GET'
  );

  const standardizedResponse: RequestsResponse = {
    success: response.success || false,
    data: [],
    pagination: {
      total_count: response.total_count || 0,
      current_page: response.page || page,
      per_page: limit,
      total_pages: Math.ceil((response.total_count || 0) / limit)
    }
  };

  // Handle the nested data structure
  if (response.data && response.data.requests && Array.isArray(response.data.requests)) {
    standardizedResponse.data = response.data.requests.map(standardizeCommunityRequest);
    standardizedResponse.requests = response.data.requests;
    standardizedResponse.pagination.total_count = response.data.total_count || 0;
  } else if (response.requests && Array.isArray(response.requests)) {
    standardizedResponse.data = response.requests.map(standardizeCommunityRequest);
    standardizedResponse.requests = response.requests;
  } else if (response.data && Array.isArray(response.data)) {
    standardizedResponse.data = response.data.map(standardizeCommunityRequest);
  }

  return standardizedResponse;
}

export async function getPremiumRequests(page: number = 1, limit: number = 10, status?: string): Promise<RequestsResponse> {
  const response = await apiRequest<any>(
    `${API_BASE_URL}/admin/premium-requests?page=${page}&limit=${limit}${status ? `&status=${status}` : ''}`,
    'GET'
  );

  const standardizedResponse: RequestsResponse = {
    success: response.success || false,
    data: [],
    pagination: {
      total_count: response.total_count || 0,
      current_page: response.page || page,
      per_page: limit,
      total_pages: Math.ceil((response.total_count || 0) / limit)
    }
  };

  if (response.requests && Array.isArray(response.requests)) {
    standardizedResponse.data = response.requests.map(standardizePremiumRequest);
  } else if (response.data && Array.isArray(response.data)) {
    standardizedResponse.data = response.data.map(standardizePremiumRequest);
  }

  if (response.requests) {
    standardizedResponse.requests = response.requests;
  }

  return standardizedResponse;
}

export async function getReportRequests(page: number = 1, limit: number = 10, status?: string): Promise<RequestsResponse> {
  const response = await apiRequest<any>(
    `${API_BASE_URL}/admin/report-requests?page=${page}&limit=${limit}${status ? `&status=${status}` : ''}`,
    'GET'
  );

  const standardizedResponse: RequestsResponse = {
    success: response.success || false,
    data: [],
    pagination: {
      total_count: response.total_count || 0,
      current_page: response.page || page,
      per_page: limit,
      total_pages: Math.ceil((response.total_count || 0) / limit)
    }
  };

  if (response.requests && Array.isArray(response.requests)) {
    standardizedResponse.data = response.requests.map(standardizeReportRequest);
  } else if (response.data && Array.isArray(response.data)) {
    standardizedResponse.data = response.data.map(standardizeReportRequest);
  }

  if (response.requests) {
    standardizedResponse.requests = response.requests;
  }

  return standardizedResponse;
}

export async function getDashboardStatistics(): Promise<StatisticsResponse> {
  return apiRequest<StatisticsResponse>(
    `${API_BASE_URL}/admin/dashboard/statistics`,
    'GET'
  );
}

export async function banUser(userId: string, ban: boolean, reason?: string): Promise<AdminApiResponse> {

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
  const response = await apiRequest<any>(
    `${API_BASE_URL}/admin/thread-categories?page=${page}&limit=${limit}`,
    'GET'
  );

  const standardizedResponse: CategoriesResponse = {
    success: response.success || false,
    data: [],
    pagination: {
      total_count: response.total_count || 0,
      current_page: response.page || page,
      per_page: limit,
      total_pages: Math.ceil((response.total_count || 0) / limit)
    }
  };

  if (response.categories && Array.isArray(response.categories)) {
    standardizedResponse.data = response.categories;
    console.log(`Mapped ${response.categories.length} thread categories to data field`);
  }

  return standardizedResponse;
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

  const response = await apiRequest<any>(
    `${API_BASE_URL}/admin/community-categories?${params}`,
    'GET'
  );

  const standardizedResponse: CategoriesResponse = {
    success: response.success || false,
    data: [],
    pagination: {
      total_count: response.total_count || 0,
      current_page: response.page || page,
      per_page: limit,
      total_pages: Math.ceil((response.total_count || 0) / limit)
    }
  };

  if (response.categories && Array.isArray(response.categories)) {
    standardizedResponse.data = response.categories;
    console.log(`Mapped ${response.categories.length} community categories to data field`);
  }

  return standardizedResponse;
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

export async function getNewsletterSubscribers(page: number = 1, limit: number = 10): Promise<RequestsResponse> {
  try {
    const response = await apiRequest<any>(
      `${API_BASE_URL}/admin/newsletter-subscribers?page=${page}&limit=${limit}`,
      'GET'
    );

    const standardizedResponse: RequestsResponse = {
      success: response.success || false,
      data: [],
      pagination: {
        total_count: response.total_count || 0,
        current_page: response.page || page,
        per_page: limit,
        total_pages: Math.ceil((response.total_count || 0) / limit)
      }
    };

    if (response.users && Array.isArray(response.users)) {
      standardizedResponse.data = response.users.map(standardizeUser);
    }

    return standardizedResponse;
  } catch (error) {
    console.error('Error fetching newsletter subscribers:', error);
    return {
      success: false,
      data: [],
      pagination: {
        total_count: 0,
        current_page: page,
        per_page: limit,
        total_pages: 0
      }
    };
  }
}

export async function syncCommunityRequests(): Promise<IApiResponse<{
  total_pending_communities: number;
  pending_community_ids?: string[];
  creator_ids?: string[];
}>> {
  const response = await apiRequest<IApiResponse<{
    total_pending_communities: number;
    pending_community_ids?: string[];
    creator_ids?: string[];
  }>>(
    `${API_BASE_URL}/admin/community-requests/sync`,
    'POST'
  );

  return response;
}