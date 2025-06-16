/**
 * Community-related interfaces
 */

import type { IUser } from "./IUser";
import type { IPagination } from "./ICommon";

/**
 * Community interface
 */
export interface ICommunity {
  id: string;
  name: string;
  description: string;
  logo_url: string;
  banner_url: string;
  creator_id: string;
  is_approved: boolean;
  categories: string[];
  created_at: string;
  member_count?: number;
}

/**
 * Community response
 */
export interface ICommunityResponse {
  success: boolean;
  data: ICommunity;
}

/**
 * Communities list response
 */
export interface ICommunitiesResponse {
  success: boolean;
  data: {
    communities: ICommunity[];
    pagination: IPagination;
  };
}

/**
 * Community creation request
 */
export interface ICreateCommunityRequest {
  name: string;
  description: string;
  logo_url?: string;
  banner_url?: string;
  categories?: string[];
}

/**
 * Community update request
 */
export interface IUpdateCommunityRequest {
  name?: string;
  description?: string;
  logo_url?: string;
  banner_url?: string;
  categories?: string[];
}

/**
 * Community rule
 */
export interface ICommunityRule {
  id: string;
  community_id: string;
  title: string;
  description: string;
  order: number;
}

/**
 * Community rules response
 */
export interface ICommunityRulesResponse {
  success: boolean;
  data: {
    rules: ICommunityRule[];
  };
}

/**
 * Add rule request
 */
export interface IAddRuleRequest {
  rule_text: string;
}

/**
 * Rule response
 */
export interface IRuleResponse {
  success: boolean;
  data: {
    rule: ICommunityRule;
  };
}

/**
 * Community members response
 */
export interface IMembersResponse {
  success: boolean;
  data: {
    members: Array<IUser & {
      role: string;
      joined_at: string;
    }>;
    pagination: IPagination;
  };
}

/**
 * Update member role request
 */
export interface IUpdateMemberRoleRequest {
  role: string;
}

/**
 * Member response
 */
export interface IMemberResponse {
  success: boolean;
  data: {
    member: {
      user_id: string;
      community_id: string;
      role: string;
      joined_at: string;
    };
  };
}

/**
 * Join request
 */
export interface IJoinRequest {
  id: string;
  community_id: string;
  user_id: string;
  status: string;
  created_at: string;
}

/**
 * Join requests response
 */
export interface IJoinRequestsResponse {
  success: boolean;
  data: {
    join_requests: IJoinRequest[];
  };
}

/**
 * Join request response
 */
export interface IJoinRequestResponse {
  success: boolean;
  data: {
    message: string;
    join_request: {
      id: string;
      community_id: string;
      user_id: string;
      status: string;
    };
  };
}

/**
 * Membership status response
 */
export interface IMembershipStatusResponse {
  success: boolean;
  data: {
    status: "member" | "pending" | "none";
  };
}