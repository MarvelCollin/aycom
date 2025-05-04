import { getAuthToken, getAuthData, setAuthData } from './auth';
import appConfig from '../config/appConfig';
import { createLoggerWithPrefix } from './logger';

const API_BASE_URL = appConfig.api.baseUrl;
const logger = createLoggerWithPrefix('API');

// Function to refresh token using the refresh token
async function refreshAuthToken(refreshToken: string) {
  try {
    logger.info('Attempting to refresh token');
    const response = await fetch(`${API_BASE_URL}/auth/refresh-token`, {
      method: "POST",
      headers: { 
        "Content-Type": "application/json"
      },
      body: JSON.stringify({ refresh_token: refreshToken }), // Fixed parameter name to match backend expectation
      credentials: "include",
    });
    
    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      logger.error('Token refresh failed with status:', response.status, errorData);
      throw new Error(errorData.message || 'Token refresh failed');
    }
    
    const data = await response.json();
    
    // Update auth data in local storage
    if (data.access_token) {
      const authData = getAuthData() || {};
      setAuthData({
        accessToken: data.access_token,
        refreshToken: data.refresh_token || authData.refreshToken,
        userId: data.user_id || authData.userId,
        expiresAt: Date.now() + (data.expires_in * 1000)
      });
      
      return data.access_token;
    }
    
    throw new Error('No access token in refresh response');
  } catch (error) {
    logger.error('Failed to refresh token:', error);
    throw error;
  }
}

// Enhanced fetch function with automatic token refresh
export async function apiRequest(endpoint: string, options: RequestInit = {}) {
  // Get current auth token
  let token = getAuthToken();
  const authData = getAuthData();
  
  // Prepare headers
  const headers = new Headers(options.headers || {});
  if (!headers.has('Content-Type') && !options.body?.toString().includes('FormData')) {
    headers.set('Content-Type', 'application/json');
  }
  if (token) {
    headers.set('Authorization', `Bearer ${token}`);
  }
  
  // Make the request
  try {
    const response = await fetch(`${API_BASE_URL}${endpoint}`, {
      ...options,
      headers,
      credentials: 'include'
    });
    
    // If unauthorized and we have a refresh token, try to refresh
    if (response.status === 401 && authData?.refreshToken) {
      logger.info('Token expired, attempting to refresh');
      
      try {
        // Get a new token
        token = await refreshAuthToken(authData.refreshToken);
        
        // Update headers with new token
        headers.set('Authorization', `Bearer ${token}`);
        
        // Retry the original request with the new token
        logger.info('Token refreshed, retrying original request');
        const newResponse = await fetch(`${API_BASE_URL}${endpoint}`, {
          ...options,
          headers,
          credentials: 'include'
        });
        
        return newResponse;
      } catch (refreshError) {
        logger.error('Token refresh failed:', refreshError);
        // Clear auth data on refresh failure
        setAuthData(null);
        throw new Error('Authentication failed. Please log in again.');
      }
    }
    
    return response;
  } catch (error) {
    logger.error(`API request failed for ${endpoint}:`, error);
    throw error;
  }
}