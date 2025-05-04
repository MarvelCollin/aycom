import appConfig from '../config/appConfig';
import { apiRequest } from '../utils/apiClient';
import { createLoggerWithPrefix } from '../utils/logger';

const API_BASE_URL = appConfig.api.baseUrl;
const logger = createLoggerWithPrefix('UserAPI');

export async function getProfile() {
  try {
    const response = await apiRequest('/users/profile', {
      method: "GET"
    });
    
    if (!response.ok) {
      try {
        const errorData = await response.json();
        logger.error('Failed to fetch user profile:', errorData);
        throw new Error(
          errorData.message || 
          errorData.error?.message || 
          `Failed to fetch user profile (${response.status} ${response.statusText}): ${JSON.stringify(errorData)}`
        );
      } catch (parseError) {
        throw new Error(`Failed to fetch user profile. Status: ${response.status}. Details: ${await response.text()}`);
      }
    }
    
    return response.json();
  } catch (error) {
    logger.error('Get profile failed:', error);
    throw error;
  }
}

export async function updateProfile(data: Record<string, any>) {
  try {
    const response = await apiRequest('/users/profile', {
      method: "PUT",
      body: JSON.stringify(data)
    });
    
    if (!response.ok) {
      try {
        const errorData = await response.json();
        logger.error('Failed to update user profile:', errorData);
        throw new Error(
          errorData.message || 
          errorData.error?.message || 
          `Failed to update user profile: ${response.status} ${response.statusText}`
        );
      } catch (parseError) {
        throw new Error(`Failed to update user profile: ${response.status} ${response.statusText}`);
      }
    }
    
    return response.json();
  } catch (error) {
    logger.error('Update profile failed:', error);
    throw error;
  }
}

export async function getUserById(userId: string) {
  try {
    const response = await apiRequest(`/users/${userId}`, {
      method: "GET"
    });
    
    if (!response.ok) {
      try {
        const errorData = await response.json();
        logger.error(`Failed to fetch user ${userId}:`, errorData);
        throw new Error(
          errorData.message || 
          errorData.error?.message || 
          `Failed to fetch user: ${response.status} ${response.statusText}`
        );
      } catch (parseError) {
        throw new Error(`Failed to fetch user: ${response.status} ${response.statusText}`);
      }
    }
    
    return response.json();
  } catch (error) {
    logger.error(`Get user ${userId} failed:`, error);
    throw error;
  }
}