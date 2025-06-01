/**
 * Search-related interfaces
 */

import type { IUser } from './IUser';
import type { ITweet } from './ISocialMedia';
import type { ICommunity } from './ICommunity';
import type { IPagination } from './ICommon';

/**
 * Search users request
 */
export interface ISearchUsersRequest {
  q: string;
  filter?: string;
  page?: number;
  limit?: number;
}

/**
 * Search users response
 */
export interface ISearchUsersResponse {
  success: boolean;
  data: {
    users: IUser[];
    pagination: IPagination;
  };
}

/**
 * Search threads request
 */
export interface ISearchThreadsRequest {
  q: string;
  filter?: string;
  page?: number;
  limit?: number;
}

/**
 * Search threads response
 */
export interface ISearchThreadsResponse {
  success: boolean;
  data: {
    threads: ITweet[];
    pagination: IPagination;
  };
}

/**
 * Search communities request
 */
export interface ISearchCommunitiesRequest {
  q: string;
  categories?: string[];
  page?: number;
  limit?: number;
}

/**
 * Search communities response
 */
export interface ISearchCommunitiesResponse {
  success: boolean;
  data: {
    communities: ICommunity[];
    pagination: IPagination;
  };
}

/**
 * User recommendations response
 */
export interface IUserRecommendationsResponse {
  success: boolean;
  data: {
    users: IUser[];
  };
} 