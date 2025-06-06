import axios from 'axios';
import { API_URL } from '../config';

export interface Community {
  id: string;
  name: string;
  description: string;
  logo_url: string;
  banner_url: string;
  creator_id: string;
  is_approved: boolean;
  categories: string[];
  created_at: string;
  member_count: number;
}

export interface CommunityMember {
  id: string;
  user_id: string;
  username: string;
  name: string;
  role: string;
  joined_at: string;
  profile_picture_url: string;
}

export interface CommunityRule {
  id: string;
  community_id: string;
  title: string;
  description: string;
  order: number;
}

export interface Category {
  id: string;
  name: string;
}

export interface JoinRequest {
  id: string;
  community_id: string;
  user_id: string;
  status: string;
}

export interface PaginationData {
  total_count: number;
  current_page: number;
  per_page: number;
  total_pages: number;
}

export interface MembershipStatus {
  status: 'member' | 'pending' | 'none';
}

// Create community request interface
export interface CreateCommunityRequest {
  name: string;
  description: string;
  logo_url: string;
  banner_url: string;
  rules: string;
  categories: string[];
}

// API methods
export const createCommunity = async (communityData: CreateCommunityRequest): Promise<Community> => {
  const response = await axios.post(`${API_URL}/communities`, communityData);
  return response.data.data;
};

export const getCommunityById = async (id: string): Promise<Community> => {
  const response = await axios.get(`${API_URL}/communities/${id}`);
  return response.data.data;
};

export const listCommunities = async (
  page: number = 1,
  limit: number = 25,
  filter: string = 'all',
  query?: string,
  categories?: string[]
): Promise<{ communities: Community[], pagination: PaginationData }> => {
  let url = `${API_URL}/communities?page=${page}&limit=${limit}&filter=${filter}`;
  
  if (query) {
    url += `&q=${query}`;
  }
  
  if (categories && categories.length > 0) {
    categories.forEach(category => {
      url += `&category=${category}`;
    });
  }
  
  const response = await axios.get(url);
  return response.data.data;
};

export const updateCommunity = async (
  id: string,
  communityData: Partial<CreateCommunityRequest>
): Promise<Community> => {
  const response = await axios.put(`${API_URL}/communities/${id}`, communityData);
  return response.data.data;
};

export const deleteCommunity = async (id: string): Promise<void> => {
  await axios.delete(`${API_URL}/communities/${id}`);
};

export const approveCommunity = async (id: string): Promise<Community> => {
  const response = await axios.post(`${API_URL}/communities/${id}/approve`);
  return response.data.data;
};

export const listCommunityMembers = async (
  communityId: string,
  page: number = 1,
  limit: number = 20
): Promise<{ members: CommunityMember[], pagination: PaginationData }> => {
  const response = await axios.get(
    `${API_URL}/communities/${communityId}/members?page=${page}&limit=${limit}`
  );
  return response.data.data;
};

export const addMember = async (communityId: string, userId: string): Promise<CommunityMember> => {
  const response = await axios.post(`${API_URL}/communities/${communityId}/members`, { user_id: userId });
  return response.data.data;
};

export const removeMember = async (communityId: string, userId: string): Promise<void> => {
  await axios.delete(`${API_URL}/communities/${communityId}/members/${userId}`);
};

export const updateMemberRole = async (
  communityId: string,
  userId: string,
  role: string
): Promise<CommunityMember> => {
  const response = await axios.put(
    `${API_URL}/communities/${communityId}/members/${userId}`,
    { role }
  );
  return response.data.data;
};

export const listRules = async (communityId: string): Promise<CommunityRule[]> => {
  const response = await axios.get(`${API_URL}/communities/${communityId}/rules`);
  return response.data.data.rules;
};

export const addRule = async (
  communityId: string,
  ruleText: string
): Promise<CommunityRule> => {
  const response = await axios.post(`${API_URL}/communities/${communityId}/rules`, {
    rule_text: ruleText
  });
  return response.data.data;
};

export const removeRule = async (communityId: string, ruleId: string): Promise<void> => {
  await axios.delete(`${API_URL}/communities/${communityId}/rules/${ruleId}`);
};

export const requestToJoin = async (communityId: string): Promise<JoinRequest> => {
  const response = await axios.post(`${API_URL}/communities/${communityId}/join-requests`);
  return response.data.data.join_request;
};

export const listJoinRequests = async (
  communityId: string,
  page: number = 1,
  limit: number = 20
): Promise<{ join_requests: JoinRequest[], pagination: PaginationData }> => {
  const response = await axios.get(
    `${API_URL}/communities/${communityId}/join-requests?page=${page}&limit=${limit}`
  );
  return response.data.data;
};

export const approveJoinRequest = async (
  communityId: string,
  requestId: string
): Promise<JoinRequest> => {
  const response = await axios.post(
    `${API_URL}/communities/${communityId}/join-requests/${requestId}/approve`
  );
  return response.data.data.join_request;
};

export const rejectJoinRequest = async (
  communityId: string,
  requestId: string
): Promise<JoinRequest> => {
  const response = await axios.post(
    `${API_URL}/communities/${communityId}/join-requests/${requestId}/reject`
  );
  return response.data.data.join_request;
};

export const checkMembershipStatus = async (communityId: string): Promise<MembershipStatus> => {
  const response = await axios.get(`${API_URL}/communities/${communityId}/membership`);
  return response.data.data;
};

export const listCategories = async (): Promise<Category[]> => {
  const response = await axios.get(`${API_URL}/admin/community-categories`);
  return response.data.data.categories;
};

export default {
  createCommunity,
  getCommunityById,
  listCommunities,
  updateCommunity,
  deleteCommunity,
  approveCommunity,
  listCommunityMembers,
  addMember,
  removeMember,
  updateMemberRole,
  listRules,
  addRule,
  removeRule,
  requestToJoin,
  listJoinRequests,
  approveJoinRequest,
  rejectJoinRequest,
  checkMembershipStatus,
  listCategories
};