// Auth utility for consistent token management
import { createLoggerWithPrefix } from './logger';

const logger = createLoggerWithPrefix('Auth');
const TOKEN_VALIDATION_INTERVAL = 1000 * 60 * 5; // Check token every 5 minutes
let tokenValidationTimer: number | null = null;

/**
 * Helper to get the authentication token from localStorage
 */
export function getAuthToken(): string {
  try {
    const authData = localStorage.getItem('auth');
    if (authData) {
      const auth = JSON.parse(authData);
      
      // Log token status for debugging
      if (auth.accessToken) {
        // Check expiration before returning token
        if (auth.expiresAt && Date.now() > auth.expiresAt) {
          logger.warn('Token has expired, will need refresh');
          
          // If refresh token exists, we'll still return the token and 
          // let the apiClient handle the refresh
          if (!auth.refreshToken) {
            logger.warn('No refresh token available for expired token');
            return '';
          }
        }
        
        logger.debug(`Found token: ${auth.accessToken.substring(0, 10)}...`);
        return auth.accessToken;
      } else {
        logger.warn('Token exists in auth data but is empty');
      }
    }
  } catch (err) {
    logger.error("Error parsing auth data:", err);
  }
  
  logger.warn('No auth token found');
  return '';
}

/**
 * Helper to check if user is authenticated
 */
export function isAuthenticated(): boolean {
  try {
    const authData = localStorage.getItem('auth');
    if (authData) {
      const auth = JSON.parse(authData);
      
      // Check if token exists and is not expired
      if (auth.accessToken && auth.isAuthenticated === true) {
        // If we have an expiry, check it
        if (auth.expiresAt) {
          const isValid = Date.now() < auth.expiresAt;
          logger.debug(`Token expiry check: now=${Date.now()}, expires=${auth.expiresAt}, isValid=${isValid}`);
          return isValid;
        }
        return true;
      } else {
        logger.warn(`Auth validation failed: token=${!!auth.accessToken}, isAuthenticated=${auth.isAuthenticated}`);
      }
    } else {
      logger.warn('No auth data in localStorage');
    }
  } catch (err) {
    logger.error("Error checking authentication status:", err);
  }
  
  // Remove the simple auth flag if we fail validation
  if (localStorage.getItem('aycom_authenticated') === 'true') {
    logger.warn('Simple auth flag is true but auth data validation failed - clearing flag');
    localStorage.setItem('aycom_authenticated', 'false');
  }
  
  return false;
}

/**
 * Get current user ID
 */
export function getUserId(): string | null {
  try {
    const authData = localStorage.getItem('auth');
    if (authData) {
      const auth = JSON.parse(authData);
      if (auth.userId) {
        logger.debug(`Found user ID: ${auth.userId}`);
        return auth.userId;
      } else {
        logger.warn('Auth data exists but no user ID found');
      }
    }
  } catch (err) {
    logger.error("Error getting user ID:", err);
  }
  
  return null;
}

/**
 * Refresh the existing token if it's near expiration
 */
export function ensureTokenFreshness(): boolean {
  try {
    const authData = getAuthData();
    if (!authData || !authData.expiresAt) return false;
    
    // Check if token will expire in less than 5 minutes
    const fiveMinutesMs = 5 * 60 * 1000;
    const timeUntilExpiry = authData.expiresAt - Date.now();
    const isNearExpiry = timeUntilExpiry < fiveMinutesMs;
    
    if (isNearExpiry) {
      logger.info(`Token is near expiry (expires in ${Math.round(timeUntilExpiry/1000)}s), should be refreshed`);
    }
    
    return isNearExpiry;
  } catch (error) {
    logger.error('Error checking token freshness:', error);
    return false;
  }
}

/**
 * Set auth data to localStorage and mark user as authenticated
 */
export function setAuthData(userData: {
  accessToken: string;
  refreshToken?: string;
  userId: string;
  expiresAt?: number;
} | null): void {
  try {
    if (userData === null) {
      // Clear auth data if null is passed
      clearAuthData();
      logger.info('Auth data cleared');
      return;
    }
    
    // Ensure there's a valid expiry time
    const expiresAt = userData.expiresAt || (Date.now() + 3600 * 1000); // Default 1 hour from now
    
    const authData = {
      isAuthenticated: true,
      userId: userData.userId,
      accessToken: userData.accessToken,
      refreshToken: userData.refreshToken || null,
      expiresAt: expiresAt
    };
    
    // Save to localStorage
    localStorage.setItem('auth', JSON.stringify(authData));
    localStorage.setItem('aycom_authenticated', 'true');
    
    logger.info('Auth data updated successfully');
    logger.debug(`User ID: ${userData.userId}, Token: ${userData.accessToken.substring(0, 10)}..., Expires: ${new Date(expiresAt).toLocaleString()}`);
    
    // Set up token validation interval if not already running
    setupTokenValidation();
  } catch (err) {
    logger.error("Error setting auth data:", err);
  }
}

/**
 * Clear auth data and mark user as logged out
 */
export function clearAuthData(): void {
  try {
    localStorage.removeItem('auth');
    localStorage.setItem('aycom_authenticated', 'false');
    logger.info('Auth data cleared');
    
    // Clear token validation interval
    if (tokenValidationTimer !== null) {
      window.clearInterval(tokenValidationTimer);
      tokenValidationTimer = null;
      logger.debug('Token validation interval cleared');
    }
  } catch (err) {
    logger.error("Error clearing auth data:", err);
  }
}

/**
 * Get the full auth data object
 */
export function getAuthData() {
  try {
    const authData = localStorage.getItem('auth');
    if (authData) {
      return JSON.parse(authData);
    }
  } catch (err) {
    logger.error("Error getting auth data:", err);
  }
  
  return null;
}

/**
 * Setup periodic token validation
 */
function setupTokenValidation() {
  // Clear existing timer if it exists
  if (tokenValidationTimer !== null) {
    window.clearInterval(tokenValidationTimer);
  }
  
  // Set up new validation timer
  tokenValidationTimer = window.setInterval(() => {
    logger.debug('Running periodic token validation');
    const isValid = isAuthenticated();
    if (!isValid) {
      logger.warn('Token validation failed during periodic check');
      // We'll let the hook handle token refresh
    }
  }, TOKEN_VALIDATION_INTERVAL);
  
  logger.debug('Token validation interval setup');
}

/**
 * Update token expiry time
 */
export function updateTokenExpiry(newExpiresAt: number): void {
  try {
    const authData = getAuthData();
    if (authData) {
      authData.expiresAt = newExpiresAt;
      localStorage.setItem('auth', JSON.stringify(authData));
      logger.debug(`Token expiry updated to ${new Date(newExpiresAt).toLocaleString()}`);
    }
  } catch (err) {
    logger.error("Error updating token expiry:", err);
  }
}