/**
 * Admin-related interfaces
 */

import type { IPagination } from './ICommon';
import type { ICategory } from './ICategory';

/**
 * Ban user request
 */
export interface IBanUserRequest {
  ban: boolean;
  reason?: string;
}

/**
 * Ban user response
 */
export interface IBanUserResponse {
  success: boolean;
  data: {
    message: string;
  };
}

/**
 * Send newsletter request
 */
export interface ISendNewsletterRequest {
  subject: string;
  content: string;
}

/**
 * Send newsletter response
 */
export interface ISendNewsletterResponse {
  success: boolean;
  data: {
    message: string;
    recipients_count: number;
  };
}

/**
 * Community request
 */
export interface ICommunityRequest {
  id: string;
  user_id: string;
  name: string;
  description: string;
  category_id?: string;
  status: string;
  created_at: string;
  updated_at?: string;
  logo_url?: string;
  banner_url?: string;
  requester?: any;
}

/**
 * Premium request
 */
export interface IPremiumRequest {
  id: string;
  user_id: string;
  reason?: string;
  status: string;
  created_at: string;
  updated_at?: string;
  requester?: any;
}

/**
 * Report request
 */
export interface IReportRequest {
  id: string;
  reporter_id: string;
  reported_user_id: string;
  reason: string;
  status: string;
  created_at: string;
  updated_at?: string;
  reporter?: any;
  reported_user?: any;
}

/**
 * Community requests response
 */
export interface ICommunityRequestsResponse {
  success: boolean;
  requests: ICommunityRequest[];
  total_count: number;
  page: number;
  limit: number;
}

/**
 * Premium requests response
 */
export interface IPremiumRequestsResponse {
  success: boolean;
  requests: IPremiumRequest[];
  total_count: number;
  page: number;
  limit: number;
}

/**
 * Report requests response
 */
export interface IReportRequestsResponse {
  success: boolean;
  requests: IReportRequest[];
  total_count: number;
  page: number;
  limit: number;
}

/**
 * Process request (approve/reject)
 */
export interface IProcessRequestRequest {
  approve: boolean;
  reason?: string;
}

/**
 * Process request response
 */
export interface IProcessRequestResponse {
  success: boolean;
  message: string;
}

/**
 * Thread categories response
 */
export interface IThreadCategoriesResponse {
  success: boolean;
  categories: ICategory[];
  total_count: number;
  page: number;
  limit: number;
}

/**
 * Community categories response
 */
export interface ICommunityCategories {
  success: boolean;
  categories: ICategory[];
  total_count: number;
  page: number;
  limit: number;
} 