import { createLoggerWithPrefix } from './logger';
import appConfig from '../config/appConfig';

const logger = createLoggerWithPrefix('Auth');
const TOKEN_VALIDATION_INTERVAL = 1000 * 60 * 5;
let tokenValidationTimer: number | null = null;

export function getAuthToken(): string {
  try {
    const authData = localStorage.getItem('auth');
    if (authData) {
      const auth = JSON.parse(authData);
      
      if (auth.accessToken) {
        if (auth.expiresAt && Date.now() > auth.expiresAt) {
          logger.warn('Token has expired, will need refresh');
          
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

export function isAuthenticated(): boolean {
  try {
    const authData = localStorage.getItem('auth');
    if (authData) {
      const auth = JSON.parse(authData);
      
      if (auth.accessToken && auth.isAuthenticated === true) {
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
  
  if (localStorage.getItem('aycom_authenticated') === 'true') {
    logger.warn('Simple auth flag is true but auth data validation failed - clearing flag');
    localStorage.setItem('aycom_authenticated', 'false');
  }
  
  return false;
}

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

export function ensureTokenFreshness(): boolean {
  try {
    const authData = getAuthData();
    if (!authData || !authData.expiresAt) return false;
    
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

export function setAuthData(userData: {
  accessToken: string;
  refreshToken?: string;
  userId: string;
  expiresAt?: number;
} | null): void {
  try {
    if (userData === null) {
      clearAuthData();
      logger.info('Auth data cleared');
      return;
    }
    
    const expiresAt = userData.expiresAt || (Date.now() + 3600 * 1000);
    
    const authData = {
      isAuthenticated: true,
      userId: userData.userId,
      accessToken: userData.accessToken,
      refreshToken: userData.refreshToken || null,
      expiresAt: expiresAt
    };
    
    localStorage.setItem('auth', JSON.stringify(authData));
    localStorage.setItem('aycom_authenticated', 'true');
    
    logger.info('Auth data updated successfully');
    logger.debug(`User ID: ${userData.userId}, Token: ${userData.accessToken.substring(0, 10)}..., Expires: ${new Date(expiresAt).toLocaleString()}`);
    
    setupTokenValidation();
  } catch (err) {
    logger.error("Error setting auth data:", err);
  }
}

export function clearAuthData(): void {
  try {
    localStorage.removeItem('auth');
    localStorage.setItem('aycom_authenticated', 'false');
    logger.info('Auth data cleared');
    
    if (tokenValidationTimer !== null) {
      window.clearInterval(tokenValidationTimer);
      tokenValidationTimer = null;
      logger.debug('Token validation interval cleared');
    }
  } catch (err) {
    logger.error("Error clearing auth data:", err);
  }
}

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

function setupTokenValidation() {
  if (tokenValidationTimer !== null) {
    window.clearInterval(tokenValidationTimer);
  }
  
  tokenValidationTimer = window.setInterval(() => {
    logger.debug('Running periodic token validation');
    const isValid = isAuthenticated();
    if (!isValid) {
      logger.warn('Token validation failed during periodic check');
    }
  }, TOKEN_VALIDATION_INTERVAL);
  
  logger.debug('Token validation interval setup');
}

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

/**
 * Gets the user's role from auth data or fetches it from the API
 * @returns User role: 'admin', 'moderator', or 'user' (default)
 */
export async function getUserRole(): Promise<string> {
  try {
    // First check if we have the role cached in auth data
    const authData = getAuthData();
    if (authData && authData.userRole) {
      logger.debug(`Using cached user role: ${authData.userRole}`);
      return authData.userRole;
    }
    
    // If not, fetch from the API
    const userId = getUserId();
    if (!userId) {
      logger.warn('Cannot get user role - not logged in');
      return 'user';
    }
    
    const token = getAuthToken();
    if (!token) {
      logger.warn('Cannot get user role - no auth token');
      return 'user';
    }
    
    const API_BASE_URL = appConfig.api.baseUrl;
    logger.debug('Fetching user role from API');
    
    const response = await fetch(`${API_BASE_URL}/users/${userId}/role`, {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      },
      credentials: 'include'
    });
    
    if (!response.ok) {
      logger.warn(`Failed to get user role: ${response.status} ${response.statusText}`);
      return 'user';
    }
    
    const data = await response.json();
    const role = data.role || 'user';
    
    // Cache the role in auth data
    if (authData) {
      authData.userRole = role;
      localStorage.setItem('auth', JSON.stringify(authData));
      logger.debug(`Cached user role: ${role}`);
    }
    
    return role;
  } catch (error) {
    logger.error('Error getting user role:', error);
    return 'user';
  }
}