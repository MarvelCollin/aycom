import { getAuthToken, getAuthData, setAuthData } from './auth';
import appConfig from '../config/appConfig';
import { createLoggerWithPrefix } from './logger';

const API_BASE_URL = appConfig.api.baseUrl;
const logger = createLoggerWithPrefix('API');

// Track failed refresh attempts to prevent infinite loops
let consecutiveRefreshFailures = 0;
const MAX_REFRESH_FAILURES = 3;

// Function to refresh token using the refresh token
async function refreshAuthToken(refreshToken: string) {
  try {
    // If we've failed too many times in a row, stop trying to prevent infinite loops
    if (consecutiveRefreshFailures >= MAX_REFRESH_FAILURES) {
      logger.error(`Too many consecutive refresh failures (${consecutiveRefreshFailures}), giving up`);
      setAuthData(null); // Clear auth data
      throw new Error('Authentication failed after multiple refresh attempts. Please log in again.');
    }
    
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
      consecutiveRefreshFailures++;
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
      
      // Reset failure counter on success
      consecutiveRefreshFailures = 0;
      
      return data.access_token;
    }
    
    logger.error('Refresh response missing access token', data);
    consecutiveRefreshFailures++;
    throw new Error('No access token in refresh response');
  } catch (error) {
    logger.error('Failed to refresh token:', error);
    consecutiveRefreshFailures++;
    throw error;
  }
}

// Enhanced fetch function with automatic token refresh
export async function apiRequest(endpoint: string, options: RequestInit = {}) {
  // Get current auth token
  let token = getAuthToken();
  const authData = getAuthData();
  
  // Debug token
  if (token) {
    logger.debug(`Using token for request to ${endpoint}: ${token.substring(0, 15)}...`);
    
    // Decode token for debugging
    const parts = token.split('.');
    if (parts.length === 3) {
      try {
        // Decode payload
        const payload = JSON.parse(atob(parts[1].replace(/-/g, '+').replace(/_/g, '/')));
        logger.debug(`Token payload for ${endpoint}:`, payload);
        
        // Check token expiration
        if (payload.exp) {
          const expiresAt = payload.exp * 1000; // Convert to milliseconds
          const now = Date.now();
          logger.debug(`Token expires in ${Math.round((expiresAt - now) / 1000)}s`);
        }
      } catch (error) {
        logger.error('Failed to decode token:', error);
      }
    }
  } else {
    logger.warn(`No token available for request to ${endpoint}`);
  }
  
  // Prepare headers
  const headers = new Headers(options.headers || {});
  if (!headers.has('Content-Type') && !(options.body instanceof FormData)) {
    headers.set('Content-Type', 'application/json');
  }
  if (token) {
    // Ensure token is properly formatted with 'Bearer ' prefix
    if (!token.startsWith('Bearer ')) {
      headers.set('Authorization', `Bearer ${token}`);
      logger.debug(`Setting Authorization header: Bearer ${token.substring(0, 10)}...`);
    } else {
      headers.set('Authorization', token);
      logger.debug(`Setting Authorization header with pre-formatted token`);
    }
  } else {
    logger.debug(`No token available for ${endpoint}`);
  }
  
  // Make the request
  try {
    logger.debug(`Sending ${options.method || 'GET'} request to ${endpoint}`);
    const response = await fetch(`${API_BASE_URL}${endpoint}`, {
      ...options,
      headers,
      credentials: 'include'
    });
    
    // Log the response status
    logger.debug(`Response status for ${endpoint}: ${response.status} ${response.statusText}`);
    
    // If unauthorized and we have a refresh token, try to refresh
    if (response.status === 401 && authData?.refreshToken) {
      logger.info(`Token expired for ${endpoint}, attempting to refresh`);
      
      try {
        // Get a new token
        token = await refreshAuthToken(authData.refreshToken);
        
        // Update headers with new token
        headers.set('Authorization', `Bearer ${token}`);
        
        // Retry the original request with the new token
        logger.info(`Token refreshed, retrying request to ${endpoint}`);
        const newResponse = await fetch(`${API_BASE_URL}${endpoint}`, {
          ...options,
          headers,
          credentials: 'include'
        });
        
        // Log the retry response status
        logger.debug(`Retry response status for ${endpoint}: ${newResponse.status} ${newResponse.statusText}`);
        
        // Check if the retry also failed with 401
        if (newResponse.status === 401) {
          logger.error(`Still getting 401 after token refresh for ${endpoint}`);
          setAuthData(null); // Clear auth data
          window.location.href = '/login?session_expired=true';
          throw new Error('Authentication failed even after token refresh. Please log in again.');
        }
        
        return newResponse;
      } catch (refreshError) {
        logger.error(`Token refresh failed for ${endpoint}:`, refreshError);
        // Clear auth data on refresh failure
        setAuthData(null);
        throw new Error('Authentication failed. Please log in again.');
      }
    }
    
    // Log non-successful responses for debugging
    if (!response.ok) {
      logger.warn(`Request to ${endpoint} failed with status ${response.status}`);
      try {
        const errorBody = await response.clone().json();
        logger.warn(`Error response body:`, errorBody);
      } catch (e) {
        // Ignore JSON parsing errors
      }
    }
    
    return response;
  } catch (error) {
    logger.error(`API request failed for ${endpoint}:`, error);
    throw error;
  }
}